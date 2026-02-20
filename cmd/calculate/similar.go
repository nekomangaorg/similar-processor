package calculate

import (
	"container/heap"
	"fmt"
	"github.com/caneroj1/stemmer"
	"github.com/james-bowman/nlp"
	"github.com/james-bowman/nlp/measures/pairwise"
	"github.com/james-bowman/sparse"
	_ "github.com/mattn/go-sqlite3"
	similar "github.com/similar-manga/similar/cmd/calculate/similar_helpers"
	"github.com/similar-manga/similar/internal"
	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/mat"
	"math"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var tagRegex = regexp.MustCompile("[^a-zA-Z0-9]+")

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

var similarCmd = &cobra.Command{
	Use:   "similar",
	Short: "This updates the similar calculations",
	Long:  `Calculate and update the similar generations for manga entries`,
	Run:   runSimilar,
}

func init() {
	calculateCmd.AddCommand(similarCmd)
	similarCmd.Flags().BoolP("skipped", "s", false, "Print out reason a match was skipped")
	similarCmd.Flags().BoolP("debug", "d", false, "Run a set of debug entries only.")
	similarCmd.Flags().BoolP("export", "e", false, "Only export results, don't recalculate similar.")
	similarCmd.Flags().IntP("threads", "t", 1000, "Change the batch processing amount")
	similarCmd.Flags().BoolP("verbose", "v", false, "Print detailed match information")
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
	mangaList := internal.GetAllManga()

	var corpusTag []string
	var corpusDesc []string
	var corpusDescLength []int

	debugMangaIds := map[string]bool{
		"f7888782-0727-49b0-95ec-a3530c70f83b": true,
		"e56a163f-1a4c-400b-8c1d-6cb98e63ce04": true,
		"ee0df4ab-1e8d-49b9-9404-da9dcb11a32a": true,
		"32d76d19-8a05-4db0-9fc2-e0b0648fe9d0": true,
		"d46d9573-2ad9-45b2-9b6d-45f95452d1c0": true,
		"e78a489b-6632-4d61-b00b-5206f5b8b22b": true,
		"58bc83a0-1808-484e-88b9-17e167469e23": true,
		"0fa5dab2-250a-4f69-bd15-9ceea54176fa": true,
	}

	if !debugMode {
		DeleteSimilarDB()
	}

	fmt.Println("Begin loading into corpus")
	for _, manga := range mangaList {
		if manga.Title == nil || manga.Description == nil {
			continue
		}

		tagText := ""
		for _, tag := range manga.Tags {
			if tag.Name != nil {
				tagText += tagRegex.ReplaceAllString((*tag.Name)["en"], "") + " "
			}
		}

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

	lsiTagVectoriser := nlp.NewCountVectoriser([]string{}...)
	lsiPipelineTag := nlp.NewPipeline(lsiTagVectoriser)

	stopWordsStemmed := append([]string(nil), similar.StopWords...)
	stemmer.StemMultipleMutate(&stopWordsStemmed)
	for i := range stopWordsStemmed {
		stopWordsStemmed[i] = strings.ToLower(stopWordsStemmed[i])
	}
	lsiPipelineDescription := nlp.NewPipeline(nlp.NewCountVectoriser(stopWordsStemmed...), nlp.NewTfidfTransformer())

	fmt.Println("Fitting models...")
	lsiTag, _ := lsiPipelineTag.FitTransform(corpusTag...)

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

	lsiDesc, _ := lsiPipelineDescription.FitTransform(corpusDesc...)
	lsiDescCSC := lsiDesc.(sparse.TypeConverter).ToCSC()

	fmt.Println("Caching vectors...")
	mangaCount := len(mangaList)
	tagVectors := make([]mat.Vector, mangaCount)
	descVectors := make([]mat.Vector, mangaCount)
	for i := 0; i < mangaCount; i++ {
		tagVectors[i] = lsiTagCSCWeighted.ColView(i)
		descVectors[i] = lsiDescCSC.ColView(i)
	}

	jobs := make(chan int, mangaCount)
	progressChan := make(chan struct{}, mangaCount)
	var wg sync.WaitGroup

	for w := 0; w < threads; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				processManga(idx, mangaList, tagVectors, descVectors, corpusDescLength, debugMode, skippedMode, verbose, debugMangaIds, progressChan)
			}
		}()
	}

	// --- FIXED PROGRESS BAR ---
	go func() {
		processed := 0
		startTime := time.Now()
		for range progressChan {
			processed++
			// Update every 10 items or at the very end
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

	fmt.Printf("\nCalculated similarities for %d Manga in %s\n\n", mangaCount, time.Since(startProcessing))
}

func processManga(idx int, list []internal.Manga, tagVecs, descVecs []mat.Vector, descLens []int, debug, skipped, verbose bool, debugIds map[string]bool, progress chan<- struct{}) {
	defer func() { progress <- struct{}{} }()

	current := list[idx]
	if debug {
		if _, ok := debugIds[current.Id]; !ok {
			return
		}
	}
	if descLens[idx] < MinDescriptionWords {
		return
	}

	vTag := tagVecs[idx]
	vDesc := descVecs[idx]
	h := &MatchMinHeap{}
	heap.Init(h)

	for i := 0; i < len(list); i++ {
		if i == idx {
			continue
		}

		dTag := pairwise.CosineSimilarity(vTag, tagVecs[i])
		dDesc := pairwise.CosineSimilarity(vDesc, descVecs[i])

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
			if invalid, _ := invalidForProcessing(match, idx, current, list[i]); !invalid {
				heap.Push(h, match)
			}
		} else if score > (*h)[0].Distance {
			if invalid, _ := invalidForProcessing(match, idx, current, list[i]); !invalid {
				heap.Pop(h)
				heap.Push(h, match)
			}
		}
	}

	if h.Len() == 0 {
		return
	}

	data := internal.SimilarManga{
		Id: current.Id, Title: *current.Title, ContentRating: current.ContentRating,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	for h.Len() > 0 {
		m := heap.Pop(h).(customMatch)
		target := list[m.ID]
		data.SimilarMatches = append(data.SimilarMatches, internal.SimilarMatch{
			Id: target.Id, Title: *target.Title, Score: float32(m.Distance / (TagScoreRatio + 1.0)),
			Languages: target.AvailableTranslatedLanguages,
		})
	}

	if !debug {
		InsertSimilarData(data)
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
	os.RemoveAll("data/similar/")
	os.MkdirAll("data/similar/", 0755)
	similarList := getDBSimilar()
	var currentFile *os.File
	var currentSuffix string

	for _, sim := range similarList {
		folder := "data/similar/" + sim.Id[0:2]
		suffix := sim.Id[0:3]
		os.MkdirAll(folder, 0755)

		if suffix != currentSuffix {
			if currentFile != nil {
				currentFile.Close()
			}
			currentFile, _ = os.OpenFile(folder+"/"+suffix+".html", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			currentSuffix = suffix
		}
		currentFile.WriteString(sim.Id + ":::||@!@||:::" + sim.JSON + "\n")
	}
	if currentFile != nil {
		currentFile.Close()
	}
}
