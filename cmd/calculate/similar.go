package calculate

import (
	"fmt"
	"github.com/caneroj1/stemmer"
	"github.com/james-bowman/nlp"
	"github.com/james-bowman/nlp/measures/pairwise"
	"github.com/james-bowman/sparse"
	_ "github.com/mattn/go-sqlite3"
	"github.com/similar-manga/similar/internal"
	"github.com/similar-manga/similar/similar"
	"github.com/spf13/cobra"
	"gonum.org/v1/gonum/mat"
	"log"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

var similarCmd = &cobra.Command{
	Use:   "similar",
	Short: "This updates the similar calculations",
	Long:  `\nCalculate and update the similar generations for manga entries`,
	Run:   runSimilar,
}

func init() {
	calculateCmd.AddCommand(similarCmd)
	calculateCmd.PersistentFlags().BoolP("skipped", "s", false, "Print out reason a match was skipped")
	calculateCmd.PersistentFlags().BoolP("debug", "d", false, "Run a set of `debug` entries only.  Printing results to the screen only.")
}
func runSimilar(cmd *cobra.Command, args []string) {
	startProcessing := time.Now()
	fmt.Printf("\nBegin calculating similars\n")

	// Settings
	numSimToGet := 40
	tagScoreRatio := 0.40
	ignoreDescScoreUnder := 0.01
	acceptDescScoreOver := 0.45
	ignoreTagsUnderCount := 2
	minDescriptionWords := 15

	// Loop through all manga and try to get their chapter information for each
	countMangasProcessed := 0

	corpusTag := []string{}
	corpusDesc := []string{}
	corpusDescLength := []int{}

	debugMode, _ := cmd.Flags().GetBool("debug")
	skippedMode, _ := cmd.Flags().GetBool("skipped")

	// Debug check / skip mangas
	debugMangaIds := map[string]bool{"f7888782-0727-49b0-95ec-a3530c70f83b": true, "e56a163f-1a4c-400b-8c1d-6cb98e63ce04": true, "ee0df4ab-1e8d-49b9-9404-da9dcb11a32a": true, "32d76d19-8a05-4db0-9fc2-e0b0648fe9d0": true, "d46d9573-2ad9-45b2-9b6d-45f95452d1c0": true,
		"e78a489b-6632-4d61-b00b-5206f5b8b22b": true, "58bc83a0-1808-484e-88b9-17e167469e23": true, "0fa5dab2-250a-4f69-bd15-9ceea54176fa": true}
	if debugMode {
		fmt.Printf("\nRunning in Debug mode for the following ids:\n")
		for k := range debugMangaIds {
			fmt.Printf("  - %s\n", k)
		}
		fmt.Println("\n")

	} else {
		DeleteSimilarDB()
	}

	mangaList := GetAllManga()

	fmt.Println("Begin loading into corpus")

	for _, manga := range mangaList {
		// Skip if invalid, this should hardily ever occur
		if manga.Title == nil || manga.Description == nil {
			fmt.Printf("!!! Manga with Id %s had nil title or nil description", manga.Id)
			continue
		}

		// Get the tag and description for this manga
		tagText := ""
		for _, tag := range manga.Tags {
			reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
			tagText += reg.ReplaceAllString((*tag.Name)["en"], "") + " "
		}
		descText := similar.CleanTitle((*manga.Title)["en"]) + " "
		for _, altTitle := range manga.AltTitles {
			if val, ok := altTitle["en"]; ok {
				if similar.CleanTitle(val) != "" {
					descText += similar.CleanTitle(val) + " "
				}
			}
		}
		descText += similar.CleanDescription((*manga.Description)["en"])

		// Append to the corpusDesc
		corpusTag = append(corpusTag, tagText)
		corpusDesc = append(corpusDesc, descText)
		corpusDescLength = append(corpusDescLength, len(strings.Split(descText, " ")))
	}
	amountOfMangaToProcess := len(mangaList)

	if debugMode {
		amountOfMangaToProcess = len(debugMangaIds)
	}
	fmt.Printf("\n\nLoaded %d Manga into our corpus\n\n", len(corpusDesc))

	// Create our tf-idf pipeline
	lsiTagVectoriser := nlp.NewCountVectoriser([]string{}...)
	lsiPipelineTag := nlp.NewPipeline(lsiTagVectoriser)
	stopWordsStemmed := append([]string(nil), similar.StopWords...)
	stemmer.StemMultipleMutate(&stopWordsStemmed)
	for i := range stopWordsStemmed {
		stopWordsStemmed[i] = strings.ToLower(stopWordsStemmed[i])
	}
	lsiPipelineDescription := nlp.NewPipeline(nlp.NewCountVectoriser(stopWordsStemmed...), nlp.NewTfidfTransformer())

	// Transform the corpusTag into an LSI fitting the model to the documents in the process
	start := time.Now()
	fmt.Printf("fitting to corpus of tags!\n")
	lsiTag, err := lsiPipelineTag.FitTransform(corpusTag...)
	if err != nil {
		log.Fatalf("ERROR: failed to process documents because\n %v\n", err)
	}
	lsiTagCSC := lsiTag.(sparse.TypeConverter).ToCSC()
	m, n := lsiTag.Dims()
	fmt.Printf("\t- fitted data in %s\n", time.Since(start))
	fmt.Printf("\t- system dim = %d x %d\n\n", m, n)

	// We will now apply our custom weights for tags
	// Each row of this matrix is a tag which we have a weight for
	fmt.Println("Tag Vectoriser Vocabulary:")
	fmt.Println(lsiTagVectoriser.Vocabulary)
	fmt.Println()
	vocabularyInverse := map[int]string{}
	for k, v := range lsiTagVectoriser.Vocabulary {
		vocabularyInverse[v] = k
	}

	// Special weights for tags that should have higher priority over others
	// These are hand tuned and adhoc in nature, but seem to work?
	tagWeights := map[string]float64{
		"sexualviolence": 1.00,
		"gore":           1.00,
		"koma":           1.00,
		"wuxia":          1.00,
		"loli":           0.90,
		"incest":         0.90,
		"sports":         0.90,
		"boyslove":       0.90,
		"girlslove":      0.90,
		"isekai":         0.90,
		"villainess":     0.90,
		"historical":     0.80,
		"horror":         0.80,
		"mecha":          0.80,
		"medical":        0.80,
		"sliceoflife":    0.80,
		"cooking":        0.80,
		"crossdressing":  0.80,
		"genderswap":     0.80,
		"harem":          0.80,
		"reverseharem":   0.80,
		"vampires":       0.80,
		"zombies":        0.80,
	}

	// Loop through the tag weights and set them to our custom ones
	lsiTagCSCWeighted := lsiTag.(sparse.TypeConverter).ToCSC()
	dimR, dimC := lsiTagCSCWeighted.Dims()
	for r := 0; r < dimR; r++ {
		tag := vocabularyInverse[r]
		tagWeight := 0.70
		if val, ok := tagWeights[tag]; ok {
			tagWeight = val
		}
		for c := 0; c < dimC; c++ {
			if lsiTagCSCWeighted.At(r, c) > 0 {
				lsiTagCSCWeighted.Set(r, c, tagWeight)
			}
		}
	}

	// Transform the corpusDesc into an LSI fitting the model to the documents in the process
	start = time.Now()
	fmt.Printf("fitting to corpus of descriptions!\n")
	lsiDesc, err := lsiPipelineDescription.FitTransform(corpusDesc...)
	if err != nil {
		log.Fatalf("ERROR: failed to process documents because\n %v\n", err)
	}
	lsiDescCSC := lsiDesc.(sparse.TypeConverter).ToCSC()
	m, n = lsiDesc.Dims()
	fmt.Printf("\t- fitted data in %s\n", time.Since(start))
	fmt.Printf("\t- system dim = %d x %d\n\n", m, n)

	// Create a "buffer" that is our num of max routines
	// If we can append to it, then we will run a coroutine
	// https://stackoverflow.com/a/25306241/7718197
	// https://downey.io/notes/dev/openmp-parallel-for-in-golang/
	var wg sync.WaitGroup
	wg.Add(len(mangaList))
	maxGoroutines := 1000
	guard := make(chan struct{}, maxGoroutines)

	//	// For each manga we will get the top calculate for tags and description
	//	// We will then combine these into a single score which is then used to rank all manga
	start = time.Now()

	for currentMangaIndex := 0; currentMangaIndex < len(mangaList); currentMangaIndex++ {

		// would block if guard channel is already filled
		guard <- struct{}{}
		go func(currentMangaIndex int) {
			defer wg.Done()

			// This manga we will try to match to
			// NOTE: here we use the weighted tag CSC matrix, so we will multiply this against a one-hot-matrix
			// NOTE: e.g. [0.7 1.0 0.0 0.0 0.9] * [0 1 0 0 1] => 1.9 score value for current against another
			currentManga := mangaList[currentMangaIndex]

			vTagWeighted := lsiTagCSCWeighted.ColView(currentMangaIndex)
			numTags := int(mat.Sum(lsiTagCSC.ColView(currentMangaIndex)))
			vDesc := lsiDescCSC.ColView(currentMangaIndex)

			// Skip this manga if it has no description
			if corpusDescLength[currentMangaIndex] < minDescriptionWords {
				<-guard
				return
			}
			if debugMode {
				if _, ok := debugMangaIds[currentManga.Id]; !ok {
					<-guard
					return
				}
			}

			var sb strings.Builder

			// Perform matching to all the other vectors
			var matches []customMatch
			for mangaMatchCheckIndex := 0; mangaMatchCheckIndex < len(mangaList); mangaMatchCheckIndex++ {

				// Get score for both tags and description
				distTag := pairwise.CosineSimilarity(vTagWeighted, lsiTagCSC.ColView(mangaMatchCheckIndex))
				distDesc := pairwise.CosineSimilarity(vDesc, lsiDescCSC.ColView(mangaMatchCheckIndex))

				// Reject invalid matches
				if math.IsNaN(distTag) || distTag < 1e-4 {
					distTag = 0
				}
				if math.IsNaN(distDesc) || distDesc < 1e-4 {
					distDesc = 0
				}

				// Special reject criteria to try to be robust to small label / description length
				if numTags < ignoreTagsUnderCount {
					distTag = 1
				}
				if distDesc < ignoreDescScoreUnder || corpusDescLength[mangaMatchCheckIndex] < minDescriptionWords {
					distDesc = 0
				}
				if distDesc > acceptDescScoreOver {
					distTag = 1
				}

				// Combine the two
				match := customMatch{}
				match.ID = mangaMatchCheckIndex
				match.Distance = tagScoreRatio*distTag + distDesc
				match.DistanceTag = distTag
				match.DistanceDesc = distDesc
				matches = append(matches, match)

			}
			sort.Slice(matches, func(i, j int) bool {
				return matches[i].Distance > matches[j].Distance
			})

			fmt.Fprintf(&sb, "Manga %d has %d tags -> %s - https://mangadex.org/title/%s\n", currentMangaIndex, numTags, (*currentManga.Title)["en"], currentManga.Id)

			// Create our calculate manga api object which will have our matches in it
			similarMangaData := similar.SimilarManga{}
			similarMangaData.Id = currentManga.Id
			similarMangaData.Title = *currentManga.Title
			similarMangaData.ContentRating = currentManga.ContentRating
			similarMangaData.UpdatedAt = time.Now().UTC().Format("2006-01-02T15:04:05+00:00")

			// Finally loop through all our matches and try to find the best ones!
			var matchesBest []customMatch
			for _, match := range matches {

				matchIndex := match.ID.(int)

				matchManga := mangaList[matchIndex]

				if invalid, reason := invalidForProcessing(match, currentMangaIndex, currentManga, matchManga); invalid {
					if skippedMode {
						fmt.Fprintf(&sb, "  | skipped because %s ->%s - https://mangadex.org/title/%s\n", reason, truncateText((*matchManga.Title)["en"], 30), matchManga.Id)
					}
					continue
				}

				// Otherwise lets append it!
				matchData := similar.SimilarMatch{}
				matchData.Id = matchManga.Id
				matchData.Title = *matchManga.Title
				matchData.ContentRating = matchManga.ContentRating
				matchData.Score = float32(match.Distance) / float32(tagScoreRatio+1.0)
				matchData.Languages = matchManga.AvailableTranslatedLanguages
				similarMangaData.SimilarMatches = append(similarMangaData.SimilarMatches, matchData)
				matchesBest = append(matchesBest, match)

				// Debug error if score is invalid
				if matchData.Score > 1 || matchData.Score < 0 {
					log.Fatalf("\u001B[1;31mINVALID SCORE: %s -> %s gave %.4f\u001B[0m\n", similarMangaData.Id, matchManga.Id, matchData.Score)
				}

				// Exit if we have found enough calculate manga!
				if len(similarMangaData.SimilarMatches) >= numSimToGet {
					break
				}

			}

			// Finally if we have non-zero matches then we should save it!
			if len(similarMangaData.SimilarMatches) > 0 {
				if !debugMode {
					InsertSimilarData(similarMangaData)
				}
			}
			countMangasProcessed++
			avgIterTime := float64(currentMangaIndex+1) / time.Since(start).Seconds()

			for i, match := range matchesBest {
				id := match.ID.(int)
				score := similarMangaData.SimilarMatches[i].Score
				fmt.Fprintf(&sb, "  | matched %d (%.3f tag, %.3f desc, %.3f comb) -> %s - https://mangadex.org/title/%s\n",
					id, match.DistanceTag, match.DistanceDesc, score, truncateText((*mangaList[id].Title)["en"], 30), mangaList[id].Id)
			}
			if !debugMode {
				//This line makes no sense if we are in debug mode
				fmt.Fprintf(&sb, "%d/%d processed at %.2f manga/sec....\n\n", currentMangaIndex+1, amountOfMangaToProcess, avgIterTime)
			}
			fmt.Println(sb.String())

			<-guard
		}(currentMangaIndex)

	}
	wg.Wait()

	fmt.Printf("Calculated simularities for %d Manga in %s\n\n", amountOfMangaToProcess, time.Since(startProcessing))

}

func truncateText(text string, maxLen int) string {
	lastSpaceIx := maxLen
	len := 0
	for i, r := range text {
		if unicode.IsSpace(r) {
			lastSpaceIx = i
		}
		len++
		if len > maxLen {
			return text[:lastSpaceIx] + "..."
		}
	}
	return text
}

func invalidForProcessing(match customMatch, currentMangaIndex int, currentManga internal.Manga, matchManga internal.Manga) (bool, string) {
	// Skip if not a valid score
	if match.Distance <= 0 {
		return true, "Invalid Score"
	}

	// Skip if the same id
	matchMangaIndexId := match.ID.(int)
	if matchMangaIndexId == currentMangaIndex {
		return true, "Same UUID"
	}

	// Skip if no common languages
	// This also enforces that the other manga has at least one chapter a user can read!
	foundCommonLang := false
	for _, lang1 := range currentManga.AvailableTranslatedLanguages {
		for _, lang2 := range matchManga.AvailableTranslatedLanguages {
			if lang1 == lang2 {
				foundCommonLang = true
				break
			}
		}
		if foundCommonLang {
			break
		}
	}
	if !foundCommonLang && len(currentManga.AvailableTranslatedLanguages) > 0 {
		return true, "No Common Languages"
	}

	// Tags / content ratings / demographics we enforce
	// Also enforce that the manga can't be *related* to the match
	if similar.NotValidMatch2(currentManga, matchManga) {
		return true, "Tag Check"
	}

	return false, ""
}

// Type of match which also stores the description
// Modeled after nlp.Match object
type customMatch struct {
	ID           interface{}
	Distance     float64
	DistanceTag  float64
	DistanceDesc float64
}
