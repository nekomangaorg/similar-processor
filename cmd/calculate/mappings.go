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
	Short: "This updates the similar calculations",
	Long:  `\nCalculate the mapping files`,
	Run:   runMappings,
}

func init() {
	calculateCmd.AddCommand(mappingsCmd)
}

func runMappings(cmd *cobra.Command, args []string) {
	initialStart := time.Now()

	mangaList := GetAllManga()

	calculateAniListMapping(mangaList)
	calculateAnimePlanetMapping(mangaList)
	calculateBookWalkerMapping(mangaList)
	calculateNovelUpdatesMapping(mangaList)
	calculateKitsuMapping(mangaList)
	calculateMyAnimeListMapping(mangaList)
	calculateMangaUpdatesMapping(mangaList)
	calculateMangaUpdatesNewIdMapping(mangaList)

	fmt.Printf("Finished all mappings in %s\n", time.Since(initialStart))

}

func calculateAniListMapping(mangaList []internal.Manga) {
	fmt.Println("Calculating AniList Mapping")
	tx, err := internal.DB.Begin()
	internal.CheckErr(err)
	for _, manga := range mangaList {
		id := manga.Links["al"]
		if id != "" {
			UpsertGeneric(tx, "ANILIST", manga.Id, id)
		}
	}
	err = tx.Commit()
	internal.CheckErr(err)

	fmt.Println("Exporting AniList mapping file")
	ExportAniList()
}

func calculateAnimePlanetMapping(mangaList []internal.Manga) {
	fmt.Println("Calculating AnimePlanet Mapping")
	tx, err := internal.DB.Begin()
	internal.CheckErr(err)
	for _, manga := range mangaList {
		id := manga.Links["ap"]
		if id != "" {
			UpsertGeneric(tx, "ANIME_PLANET", manga.Id, id)
		}
	}
	err = tx.Commit()
	internal.CheckErr(err)

	fmt.Println("Exporting Anime Planet mapping file")
	ExportAnimePlanet()
}

func calculateBookWalkerMapping(mangaList []internal.Manga) {
	fmt.Println("Calculating BookWalker Mapping")

	tx, err := internal.DB.Begin()
	internal.CheckErr(err)

	for _, manga := range mangaList {
		id := manga.Links["bw"]
		if id != "" {
			UpsertGeneric(tx, "BOOK_WALKER", manga.Id, id)
		}
	}

	err = tx.Commit()
	internal.CheckErr(err)

	fmt.Println("Exporting Book Walker mapping file")
	ExportBookWalker()
}

func calculateNovelUpdatesMapping(mangaList []internal.Manga) {
	fmt.Println("Calculating NovelUpdates Mapping")

	tx, err := internal.DB.Begin()
	internal.CheckErr(err)

	for _, manga := range mangaList {
		id := manga.Links["nu"]
		if id != "" {
			UpsertGeneric(tx, "NOVEL_UPDATES", manga.Id, id)
		}
	}

	err = tx.Commit()
	internal.CheckErr(err)

	fmt.Println("Exporting NovelUpdates mapping file")
	ExportNovelUpdates()
}

func calculateKitsuMapping(mangaList []internal.Manga) {
	fmt.Println("Calculating Kitsu Mapping")

	tx, err := internal.DB.Begin()
	internal.CheckErr(err)

	for _, manga := range mangaList {
		id := manga.Links["kt"]
		if id != "" {
			UpsertGeneric(tx, "KITSU", manga.Id, id)
		}
	}

	err = tx.Commit()
	internal.CheckErr(err)

	fmt.Println("Exporting Kitsu mapping file")
	ExportKitsu()
}

func calculateMyAnimeListMapping(mangaList []internal.Manga) {
	fmt.Println("Calculating MyAnimeList Mapping")

	tx, err := internal.DB.Begin()
	internal.CheckErr(err)

	for _, manga := range mangaList {
		id := manga.Links["mal"]
		if id != "" {
			UpsertGeneric(tx, "MYANIMELIST", manga.Id, id)
		}
	}

	err = tx.Commit()
	internal.CheckErr(err)

	fmt.Println("Exporting MyAnimeList New Ids file")
	ExportMyAnimeList()
}

func calculateMangaUpdatesMapping(mangaList []internal.Manga) {
	fmt.Println("Calculating MangaUpdates Mapping")

	tx, err := internal.DB.Begin()
	internal.CheckErr(err)

	for _, manga := range mangaList {
		id := manga.Links["mu"]
		if id != "" {
			UpsertGeneric(tx, "MANGAUPDATES_OLD", manga.Id, id)
		}
	}

	err = tx.Commit()
	internal.CheckErr(err)

	fmt.Println("Exporting MangaUpdates mapping file")
	ExportMangaUpdates()

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
	ExportMangaUpdatesNewIds()

	fmt.Printf("done processing MangaUpdates New Ids (%.2f seconds)!\n", time.Since(start).Seconds())
}
