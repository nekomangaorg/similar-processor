package calculate

import (
	"fmt"
	"github.com/similar-manga/similar/internal"
	"github.com/spf13/cobra"
	"go.uber.org/ratelimit"
	"iter"
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

	type upsertData struct {
		tableName string
		uuid      string
		id        string
	}
	var upserts []upsertData
	for manga := range mangaStream {
		for _, m := range mappings {
			id := manga.Links[m.linkKey]
			if id != "" {
				upserts = append(upserts, upsertData{
					tableName: m.tableName,
					uuid:      manga.Id,
					id:        id,
				})
			}
		}
	}

	const batchSize = 1000

	processBatch := func(batch []upsertData) {
		if len(batch) == 0 {
			return
		}
		tx, err := internal.DB.Begin()
		if err != nil {
			fmt.Printf("failed to begin transaction: %v\n", err)
			return
		}
		defer tx.Rollback()

		for _, u := range batch {
			if err := UpsertGeneric(tx, u.tableName, u.uuid, u.id); err != nil {
				fmt.Printf("failed to upsert item %s: %v\n", u.uuid, err)
				continue
			}
		}

		if err := tx.Commit(); err != nil {
			fmt.Printf("failed to commit transaction: %v\n", err)
		}
	}

	for i := 0; i < len(upserts); i += batchSize {
		end := i + batchSize
		if end > len(upserts) {
			end = len(upserts)
		}
		processBatch(upserts[i:end])
	}

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

	// First collect muLinks to avoid locking the database connection
	type muData struct {
		uuid   string
		muLink string
	}
	var muLinks []muData
	for manga := range mangaList {
		muLink := manga.Links["mu"]
		if muLink != "" {
			muLinks = append(muLinks, muData{
				uuid:   manga.Id,
				muLink: muLink,
			})
		}
	}

	// Loop through all manga and try to get their chapter information for each
	start := time.Now()
	var wg sync.WaitGroup
	maxGoroutines := 1000
	guard := make(chan struct{}, maxGoroutines)

	for index, data := range muLinks {
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
			if !AddAlreadyConvertedId(index, totalManga, uuid, muLink, rateLimiter) && !CheckAndAddLegacyId(index, totalManga, uuid, muLink, rateLimiter) {
				fmt.Printf("%d/%d manga %s -> mu invalid %s\n", index+1, totalManga, uuid, muLink)
			}
			<-guard
		}(index, totalManga, data.uuid, data.muLink, rateLimiter)
	}

	wg.Wait()
	CloseDebugFiles()

	fmt.Println("Exporting MangaUpdates New Ids file")
	exportMapping(internal.TableMangaupdatesNewId, "mangaupdates_new2mdex")

	fmt.Printf("done processing MangaUpdates New Ids (%.2f seconds)!\n", time.Since(start).Seconds())
}
