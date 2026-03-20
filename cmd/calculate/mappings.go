package calculate

import (
	"fmt"
	"github.com/similar-manga/similar/internal"
	"iter"
	"github.com/spf13/cobra"
	"go.uber.org/ratelimit"
	"sync"
	"time"
)

var mappingsCmd = &cobra.Command{
	Use:   "mappings",
	Short: "This updates the external website mapping ids to MangaDex uuids",
	Long:  "This updates the external website mapping ids to MangaDex uuids",
	Run:   runMappings,
}

func init() {
	calculateCmd.AddCommand(mappingsCmd)
}

func runMappings(cmd *cobra.Command, args []string) {
	initialStart := time.Now()

	mangaStream := internal.StreamAllManga()

	type mappingInfo struct {
		name      string
		linkKey   string
		tableName string
		fileName  string
	}

	mappings := []mappingInfo{
		{"AniList", "al", internal.TableAnilist, "anilist2mdex"},
		{"AnimePlanet", "ap", internal.TableAnimePlanet, "animeplanet2mdex"},
		{"BookWalker", "bw", internal.TableBookWalker, "bookwalker2mdex"},
		{"NovelUpdates", "nu", internal.TableNovelUpdates, "novelupdates2mdex"},
		{"Kitsu", "kt", internal.TableKitsu, "kitsu2mdex"},
		{"MyAnimeList", "mal", internal.TableMyanimelist, "myanimelist2mdex"},
		{"MangaUpdates", "mu", internal.TableMangaupdates, "mangaupdates2mdex"},
	}

	fmt.Println("Calculating mappings...")
	tx, err := internal.DB.Begin()
	internal.CheckErr(err)
	defer tx.Rollback()

	for manga := range mangaStream {
		for _, m := range mappings {
			id := manga.Links[m.linkKey]
			if id != "" {
				UpsertGeneric(tx, m.tableName, manga.Id, id)
			}
		}
	}

	err = tx.Commit()
	internal.CheckErr(err)

	fmt.Println("Exporting mapping files...")
	for _, m := range mappings {
		fmt.Printf("Exporting %s mapping file\n", m.name)
		exportMapping(m.tableName, m.fileName)
	}

	totalManga, err := internal.GetMangaCount()
	internal.CheckErr(err)
	calculateMangaUpdatesNewIdMapping(internal.StreamAllManga(), totalManga)

	fmt.Printf("Finished all mappings in %s\n", time.Since(initialStart))

}

func calculateMangaUpdatesNewIdMapping(mangaList iter.Seq[internal.Manga], totalManga int) {
	fmt.Println("Calculating MangaUpdates New Id Mapping")
	rateLimiter := ratelimit.New(1)

	// mangaupdates
	// https://www.mangaupdates.com/series.html?id=`{id}`
	// https://api.mangaupdates.com/#operation/retrieveSeries
	// https://api.mangaupdates.com/v1/series/(base38 encoding of 7char ids)
	// https://api.mangaupdates.com/v1/series/66788345008/rss

	// Loop through all manga and try to get their chapter information for each
	start := time.Now()
	var wg sync.WaitGroup
	maxGoroutines := 1000
	guard := make(chan struct{}, maxGoroutines)

	index := 0
	for manga := range mangaList {
		muLink := manga.Links["mu"]

		// would block if guard channel is already filled
		guard <- struct{}{}

		wg.Add(1)
		go func(index int, totalManga int, uuid string, muLink string, limiter ratelimit.Limiter) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("goroutine paniqued: ", r)
				}
			}()
			// Our search file
			defer wg.Done()
			if muLink != "" {
				if !AddAlreadyConvertedId(index, totalManga, uuid, muLink, rateLimiter) && !CheckAndAddLegacyId(index, totalManga, uuid, muLink, rateLimiter) {
					fmt.Printf("%d/%d manga %s -> mu invalid %s\n", index+1, totalManga, uuid, muLink)
				}
			}
			<-guard
		}(index, totalManga, manga.Id, muLink, rateLimiter)
		index++
	}

	wg.Wait()

	fmt.Println("Exporting MangaUpdates New Ids file")
	exportMapping(internal.TableMangaupdatesNewId, "mangaupdates_new2mdex")

	fmt.Printf("done processing MangaUpdates New Ids (%.2f seconds)!\n", time.Since(start).Seconds())
}
