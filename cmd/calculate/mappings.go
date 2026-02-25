package calculate

import (
	"fmt"
	"github.com/similar-manga/similar/internal"
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

	mangaList := internal.GetAllManga()

	calculateGenericMapping(mangaList, "AniList", "al", internal.TableAnilist, "anilist2mdex")
	calculateGenericMapping(mangaList, "AnimePlanet", "ap", internal.TableAnimePlanet, "animeplanet2mdex")
	calculateGenericMapping(mangaList, "BookWalker", "bw", internal.TableBookWalker, "bookwalker2mdex")
	calculateGenericMapping(mangaList, "NovelUpdates", "nu", internal.TableNovelUpdates, "novelupdates2mdex")
	calculateGenericMapping(mangaList, "Kitsu", "kt", internal.TableKitsu, "kitsu2mdex")
	calculateGenericMapping(mangaList, "MyAnimeList", "mal", internal.TableMyanimelist, "myanimelist2mdex")
	calculateGenericMapping(mangaList, "MangaUpdates", "mu", internal.TableMangaupdates, "mangaupdates2mdex")

	calculateMangaUpdatesNewIdMapping(mangaList)

	fmt.Printf("Finished all mappings in %s\n", time.Since(initialStart))

}

func calculateGenericMapping(mangaList []internal.Manga, name, linkKey, tableName, fileName string) {
	fmt.Printf("Calculating %s Mapping\n", name)
	tx, err := internal.DB.Begin()
	internal.CheckErr(err)
	for _, manga := range mangaList {
		id := manga.Links[linkKey]
		if id != "" {
			UpsertGeneric(tx, tableName, manga.Id, id)
		}
	}
	err = tx.Commit()
	internal.CheckErr(err)

	fmt.Printf("Exporting %s mapping file\n", name)
	exportMapping(tableName, fileName)
}

func calculateMangaUpdatesNewIdMapping(mangaList []internal.Manga) {
	fmt.Println("Calculating MangaUpdates New Id Mapping")
	rateLimiter := ratelimit.New(1)

	// mangaupdates
	// https://www.mangaupdates.com/series.html?id=`{id}`
	// https://api.mangaupdates.com/#operation/retrieveSeries
	// https://api.mangaupdates.com/v1/series/(base38 encoding of 7char ids)
	// https://api.mangaupdates.com/v1/series/66788345008/rss

	// Loop through all manga and try to get their chapter information for each
	start := time.Now()
	totalManga := len(mangaList)
	var wg sync.WaitGroup
	wg.Add(totalManga)
	maxGoroutines := 1000
	guard := make(chan struct{}, maxGoroutines)

	for index, manga := range mangaList {
		muLink := manga.Links["mu"]

		// would block if guard channel is already filled
		guard <- struct{}{}

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
	}

	wg.Wait()

	fmt.Println("Exporting MangaUpdates New Ids file")
	exportMapping(internal.TableMangaupdatesNewId, "mangaupdates_new2mdex")

	fmt.Printf("done processing MangaUpdates New Ids (%.2f seconds)!\n", time.Since(start).Seconds())
}
