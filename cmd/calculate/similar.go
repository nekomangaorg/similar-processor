package calculate

import (
	"bufio"
	"container/heap"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/caneroj1/stemmer"
	"github.com/james-bowman/nlp"
	"github.com/james-bowman/sparse"
	_ "github.com/mattn/go-sqlite3"
	similar "github.com/similar-manga/similar/cmd/calculate/similar_helpers"
	"github.com/similar-manga/similar/internal"
	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/mat"
)

const (
	NumSimToGet          = 20
	TagScoreRatio        = 0.40
	IgnoreDescScoreUnder = 0.01
	AcceptDescScoreOver  = 0.45
	IgnoreTagsUnderCount = 2
	MinDescriptionWords  = 15
	DefaultTagWeight     = 0.70
	SimilarityThreshold  = 1e-4
)

var (
	similarCmd = &cobra.Command{
		Use:   "similar",
		Short: "This updates the similar calculations",
		Long:  `Calculate and update the similar generations for manga entries`,
		Run:   runSimilar,
	}
	cachedStopWords []string
)

func init() {
	calculateCmd.AddCommand(similarCmd)
	similarCmd.Flags().BoolP("skipped", "s", false, "Print out reason a match was skipped")
	similarCmd.Flags().BoolP("debug", "d", false, "Run a set of debug entries only.")
	similarCmd.Flags().BoolP("export", "e", false, "Only export results, don't recalculate similar.")
	similarCmd.Flags().IntP("threads", "t", 1000, "Change the batch processing amount")
	similarCmd.Flags().BoolP("verbose", "v", false, "Print detailed match information")

	// Pre-process stop words once
	cachedStopWords = append([]string(nil), similar.StopWords...)
	stemmer.StemMultipleMutate(&cachedStopWords)
	for i := range cachedStopWords {
		cachedStopWords[i] = strings.ToLower(cachedStopWords[i])
	}
}

func runSimilar(cmd *cobra.Command, args []string) {
	debugMode, _ := cmd.Flags().GetBool("debug")
	skippedMode, _ := cmd.Flags().GetBool("skipped")
	exportOnly, _ := cmd.Flags().GetBool("export")
	threads, _ := cmd.Flags().GetInt("threads")
	verbose, _ := cmd.Flags().GetBool("verbose")

	if !exportOnly {
		fmt.Printf("\nBegin calculating similars\n")
		calculateSimilars(debugMode, skippedMode, threads, verbose)
	}

	if !debugMode {
		startProcessing := time.Now()
		fmt.Printf("Exporting All Similar to txt files\n")
		exportSimilar()
		fmt.Printf("Exporting similarities took %s\n\n", time.Since(startProcessing))
	}
}

func calculateSimilars(debugMode bool, skippedMode bool, threads int, verbose bool) {
	startProcessing := time.Now()
	allManga := internal.GetAllManga()

	if !debugMode {
		DeleteSimilarDB()
	}

	data, err := prepareSimilarityData(allManga)
	if err != nil {
		fmt.Printf("Preparation failed: %v\n", err)
		return
	}

	config := processingConfig{
		debugMode:     debugMode,
		skippedMode:   skippedMode,
		verbose:       verbose,
		debugMangaIds: getDebugMangaIds(),
		threads:       threads,
	}

	runConcurrentProcessing(data, config)

	fmt.Printf("\nCalculated similarities for %d Manga in %s\n\n", len(data.MangaList), time.Since(startProcessing))
}

func prepareSimilarityData(allManga []internal.Manga) (*SimilarityData, error) {
	fmt.Println("Begin loading into corpus")
	corpus := filterAndBuildCorpus(allManga)
	mangaCount := len(corpus.MangaList)

	fmt.Println("Fitting models...")
	lsiTagCSCWeighted, err := buildWeightedTagVectors(corpus.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to build tag vectors: %w", err)
	}
	lsiDescCSC, err := buildDescriptionVectors(corpus.Descriptions)
	if err != nil {
		return nil, fmt.Errorf("failed to build description vectors: %w", err)
	}

	fmt.Println("Caching vectors...")
	tagVectors, descVectors, tagNorms, descNorms := calculateNorms(mangaCount, lsiTagCSCWeighted, lsiDescCSC)

	langMasks := calculateLanguageMasks(corpus.MangaList)

	return &SimilarityData{
		MangaList:        corpus.MangaList,
		TagVectors:       tagVectors,
		DescVectors:      descVectors,
		TagNorms:         tagNorms,
		DescNorms:        descNorms,
		CorpusDescLength: corpus.DescriptionLens,
		LangMasks:        langMasks,
	}, nil
}

func getDebugMangaIds() map[string]bool {
	return map[string]bool{
		"f7888782-0727-49b0-95ec-a3530c70f83b": true,
		"e56a163f-1a4c-400b-8c1d-6cb98e63ce04": true,
		"ee0df4ab-1e8d-49b9-9404-da9dcb11a32a": true,
		"32d76d19-8a05-4db0-9fc2-e0b0648fe9d0": true,
		"d46d9573-2ad9-45b2-9b6d-45f95452d1c0": true,
		"e78a489b-6632-4d61-b00b-5206f5b8b22b": true,
		"58bc83a0-1808-484e-88b9-17e167469e23": true,
		"0fa5dab2-250a-4f69-bd15-9ceea54176fa": true,
	}
}

type CorpusData struct {
	MangaList       []internal.Manga
	Tags            []string
	Descriptions    []string
	DescriptionLens []int
}

func filterAndBuildCorpus(allManga []internal.Manga) *CorpusData {
	// Pre-allocate with max possible capacity to avoid reallocations
	maxSize := len(allManga)
	mangaList := make([]internal.Manga, 0, maxSize)
	corpusTag := make([]string, 0, maxSize)
	corpusDesc := make([]string, 0, maxSize)
	corpusDescLength := make([]int, 0, maxSize)

	for _, manga := range allManga {
		if manga.Title == nil || manga.Description == nil {
			continue
		}

		mangaList = append(mangaList, manga)

		var tagTextBuilder strings.Builder
		for _, tag := range manga.Tags {
			if tag.Name != nil {
				cleanTag((*tag.Name)["en"], &tagTextBuilder)
				tagTextBuilder.WriteByte(' ')
			}
		}
		tagText := tagTextBuilder.String()

		descText := similar.CleanTitle((*manga.Title)["en"]) + " "
		for _, altTitle := range manga.AltTitles {
			if val, ok := altTitle["en"]; ok {
				if cleaned := similar.CleanTitle(val); cleaned != "" {
					descText += cleaned + " "
				}
			}
		}
		descText += similar.CleanDescription((*manga.Description)["en"])

		corpusTag = append(corpusTag, tagText)
		corpusDesc = append(corpusDesc, descText)
		corpusDescLength = append(corpusDescLength, len(strings.Split(descText, " ")))
	}
	return &CorpusData{
		MangaList:       mangaList,
		Tags:            corpusTag,
		Descriptions:    corpusDesc,
		DescriptionLens: corpusDescLength,
	}
}

func buildWeightedTagVectors(corpusTag []string) (*sparse.CSC, error) {
	lsiTagVectoriser := nlp.NewCountVectoriser([]string{}...)
	lsiPipelineTag := nlp.NewPipeline(lsiTagVectoriser)

	lsiTag, err := lsiPipelineTag.FitTransform(corpusTag...)
	if err != nil {
		return nil, fmt.Errorf("failed to fit/transform tag corpus: %w", err)
	}

	vocabularyInverse := map[int]string{}
	for k, v := range lsiTagVectoriser.Vocabulary {
		vocabularyInverse[v] = k
	}

	tagWeights := map[string]float64{
		"sexualviolence": 1.0, "gore": 1.0, "koma": 1.0, "wuxia": 1.0,
		"isekai": 0.9, "villainess": 0.9, "historical": 0.8, "horror": 0.8,
	}

	lsiTagCSCWeighted := lsiTag.(sparse.TypeConverter).ToCSC()
	dimR, dimC := lsiTagCSCWeighted.Dims()
	for r := 0; r < dimR; r++ {
		tag := vocabularyInverse[r]
		weight := DefaultTagWeight
		if val, ok := tagWeights[tag]; ok {
			weight = val
		}
		for c := 0; c < dimC; c++ {
			if lsiTagCSCWeighted.At(r, c) > 0 {
				lsiTagCSCWeighted.Set(r, c, weight)
			}
		}
	}
	return lsiTagCSCWeighted, nil
}

func buildDescriptionVectors(corpusDesc []string) (*sparse.CSC, error) {
	lsiPipelineDescription := nlp.NewPipeline(nlp.NewCountVectoriser(cachedStopWords...), nlp.NewTfidfTransformer())

	lsiDesc, err := lsiPipelineDescription.FitTransform(corpusDesc...)
	if err != nil {
		return nil, fmt.Errorf("failed to fit/transform description corpus: %w", err)
	}

	lsiDescCSC := lsiDesc.(sparse.TypeConverter).ToCSC()
	return lsiDescCSC, nil
}

func calculateNorms(mangaCount int, lsiTagCSCWeighted, lsiDescCSC *sparse.CSC) ([]*sparse.Vector, []*sparse.Vector, []float64, []float64) {
	tagVectors := make([]*sparse.Vector, mangaCount)
	descVectors := make([]*sparse.Vector, mangaCount)
	tagNorms := make([]float64, mangaCount)
	descNorms := make([]float64, mangaCount)

	for i := 0; i < mangaCount; i++ {
		tv, ok := lsiTagCSCWeighted.ColView(i).(*sparse.Vector)
		if !ok {
			fmt.Printf("Warning: Type assertion failed for tag vector %d\n", i)
			continue
		}
		tagVectors[i] = tv

		dv, ok := lsiDescCSC.ColView(i).(*sparse.Vector)
		if !ok {
			fmt.Printf("Warning: Type assertion failed for desc vector %d\n", i)
			continue
		}
		descVectors[i] = dv

		tagNorms[i] = mat.Norm(tagVectors[i], 2)
		descNorms[i] = mat.Norm(descVectors[i], 2)
	}
	return tagVectors, descVectors, tagNorms, descNorms
}

func calculateLanguageMasks(mangaList []internal.Manga) []uint64 {
	uniqueLangs := make(map[string]uint64)
	nextBit := 0
	for _, m := range mangaList {
		for _, l := range m.AvailableTranslatedLanguages {
			if _, exists := uniqueLangs[l]; !exists {
				if nextBit < 63 {
					uniqueLangs[l] = 1 << nextBit
					nextBit++
				} else {
					uniqueLangs[l] = 1 << 63
				}
			}
		}
	}

	langMasks := make([]uint64, len(mangaList))
	for i, m := range mangaList {
		var mask uint64
		for _, l := range m.AvailableTranslatedLanguages {
			if val, ok := uniqueLangs[l]; ok {
				mask |= val
			}
		}
		langMasks[i] = mask
	}
	return langMasks
}

type SimilarityData struct {
	MangaList        []internal.Manga
	TagVectors       []*sparse.Vector
	DescVectors      []*sparse.Vector
	TagNorms         []float64
	DescNorms        []float64
	CorpusDescLength []int
	LangMasks        []uint64
}

type processingConfig struct {
	debugMode     bool
	skippedMode   bool
	verbose       bool
	debugMangaIds map[string]bool
	threads       int
}

func runConcurrentProcessing(data *SimilarityData, config processingConfig) {
	mangaCount := len(data.MangaList)
	jobs := make(chan int, mangaCount)
	progressChan := make(chan struct{}, mangaCount)
	var wg sync.WaitGroup

	for w := 0; w < config.threads; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				processManga(idx, data, config, progressChan)
			}
		}()
	}

	go func() {
		processed := 0
		startTime := time.Now()
		for range progressChan {
			processed++
			if processed%10 == 0 || processed == mangaCount {
				elapsed := time.Since(startTime).Seconds()
				rate := 0.0
				if elapsed > 0 {
					rate = float64(processed) / elapsed
				}
				fmt.Printf("\rProcessing: %d/%d (%.1f%%) - %.2f manga/sec",
					processed, mangaCount, float64(processed)/float64(mangaCount)*100, rate)
			}
		}
		fmt.Println()
	}()

	for i := 0; i < mangaCount; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	close(progressChan)
}

func processManga(idx int, data *SimilarityData, config processingConfig, progress chan<- struct{}) {
	defer func() { progress <- struct{}{} }()

	current := data.MangaList[idx]
	if config.debugMode {
		if _, ok := config.debugMangaIds[current.Id]; !ok {
			return
		}
	}
	if data.CorpusDescLength[idx] < MinDescriptionWords {
		return
	}

	vTag := data.TagVectors[idx]
	vDesc := data.DescVectors[idx]
	h := &MatchMinHeap{}
	heap.Init(h)
	currentMask := data.LangMasks[idx]

	vTagNorm := data.TagNorms[idx]
	vDescNorm := data.DescNorms[idx]

	for i := 0; i < len(data.MangaList); i++ {
		if i == idx {
			continue
		}

		// Performance Optimization:
		// If the current manga has languages specified, the target manga MUST share at least one language
		// to be considered similar (as per existing invalidForProcessing logic).
		// We use a pre-calculated bitmask to quickly skip pairs with no common languages.
		// If currentMask is 0 (no languages), we don't skip because existing logic allows it.
		// If (currentMask & targetMask) == 0, it means no common languages (or overflow bits match, which is safe).
		if currentMask != 0 && (currentMask&data.LangMasks[i]) == 0 {
			continue
		}

		var dTag float64
		if vTagNorm > 0 && data.TagNorms[i] > 0 {
			dTag = dotProductSparse(vTag, data.TagVectors[i]) / (vTagNorm * data.TagNorms[i])
		}

		var dDesc float64
		if vDescNorm > 0 && data.DescNorms[i] > 0 {
			dDesc = dotProductSparse(vDesc, data.DescVectors[i]) / (vDescNorm * data.DescNorms[i])
		}

		if math.IsNaN(dTag) || dTag < SimilarityThreshold {
			dTag = 0
		}
		if math.IsNaN(dDesc) || dDesc < SimilarityThreshold {
			dDesc = 0
		}

		if dDesc > AcceptDescScoreOver {
			dTag = 1
		}

		score := TagScoreRatio*dTag + dDesc
		if score <= 0 {
			continue
		}

		match := customMatch{ID: i, Distance: score, DistanceTag: dTag, DistanceDesc: dDesc}

		if h.Len() < NumSimToGet {
			if invalid, _ := invalidForProcessing(match, idx, current, data.MangaList[i]); !invalid {
				heap.Push(h, match)
			}
		} else if score > (*h)[0].Distance {
			if invalid, _ := invalidForProcessing(match, idx, current, data.MangaList[i]); !invalid {
				heap.Pop(h)
				heap.Push(h, match)
			}
		}
	}

	if h.Len() == 0 {
		return
	}

	simData := internal.SimilarManga{
		Id: current.Id, Title: *current.Title, ContentRating: current.ContentRating,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	for h.Len() > 0 {
		m := heap.Pop(h).(customMatch)
		target := data.MangaList[m.ID]
		simData.SimilarMatches = append(simData.SimilarMatches, internal.SimilarMatch{
			Id: target.Id, Title: *target.Title, Score: float32(m.Distance / (TagScoreRatio + 1.0)),
			Languages: target.AvailableTranslatedLanguages,
		})
	}

	if !config.debugMode {
		InsertSimilarData(simData)
	}
}

func invalidForProcessing(match customMatch, currentIdx int, current, target internal.Manga) (bool, string) {
	if match.Distance <= 0 {
		return true, "Invalid Score"
	}
	if match.ID == currentIdx {
		return true, "Same UUID"
	}

	common := false
	for _, l1 := range current.AvailableTranslatedLanguages {
		for _, l2 := range target.AvailableTranslatedLanguages {
			if l1 == l2 {
				common = true
				break
			}
		}
		if common {
			break
		}
	}
	if !common && len(current.AvailableTranslatedLanguages) > 0 {
		return true, "No Common Languages"
	}

	if similar.NotValidMatch(current, target) {
		return true, "Tag Check"
	}
	return false, ""
}

type customMatch struct {
	ID                                  int
	Distance, DistanceTag, DistanceDesc float64
}

type MatchMinHeap []customMatch

func (h MatchMinHeap) Len() int            { return len(h) }
func (h MatchMinHeap) Less(i, j int) bool  { return h[i].Distance < h[j].Distance }
func (h MatchMinHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *MatchMinHeap) Push(x interface{}) { *h = append(*h, x.(customMatch)) }
func (h *MatchMinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func exportSimilar() {
	if err := os.RemoveAll("data/similar/"); err != nil {
		log.Printf("Warning: failed to remove similar dir: %v", err)
	}
	if err := os.MkdirAll("data/similar/", 0755); err != nil {
		log.Fatal(err)
	}
	similarList := getDBSimilar()

	var currentFile *os.File
	var writer *bufio.Writer
	var currentSuffix string
	var currentFolder string

	for _, sim := range similarList {
		if len(sim.Id) < 3 {
			continue
		}
		folder := "data/similar/" + sim.Id[0:2]
		suffix := sim.Id[0:3]

		if folder != currentFolder {
			if err := os.MkdirAll(folder, 0755); err != nil {
				log.Fatal(err)
			}
			currentFolder = folder
		}

		if suffix != currentSuffix {
			if writer != nil {
				if err := writer.Flush(); err != nil {
					log.Fatal(err)
				}
			}
			if currentFile != nil {
				currentFile.Close()
			}
			f, err := os.OpenFile(folder+"/"+suffix+".html", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			currentFile = f
			writer = bufio.NewWriter(currentFile)
			currentSuffix = suffix
		}
		if _, err := writer.WriteString(sim.Id + ":::||@!@||:::" + sim.JSON + "\n"); err != nil {
			log.Fatal(err)
		}
	}
	if writer != nil {
		if err := writer.Flush(); err != nil {
			log.Fatal(err)
		}
	}
	if currentFile != nil {
		currentFile.Close()
	}
}

// dotProductSparse calculates the dot product of two sparse vectors.
// It assumes the underlying indices are sorted (which is standard for sparse.Vector).
// Accessing RawVector() avoids interface overhead and allows O(NNZ) intersection.
func dotProductSparse(v1, v2 *sparse.Vector) float64 {
	if v1 == nil || v2 == nil {
		return 0
	}
	d1, i1 := v1.RawVector()
	d2, i2 := v2.RawVector()

	var dot float64
	k1, k2 := 0, 0
	n1, n2 := len(i1), len(i2)

	for k1 < n1 && k2 < n2 {
		idx1 := i1[k1]
		idx2 := i2[k2]

		if idx1 < idx2 {
			k1++
		} else if idx1 > idx2 {
			k2++
		} else {
			dot += d1[k1] * d2[k2]
			k1++
			k2++
		}
	}
	return dot
}

func cleanTag(s string, b *strings.Builder) {
	for _, char := range s {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			b.WriteByte(byte(char))
		}
	}
}
