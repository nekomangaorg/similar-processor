package calculate

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"iter"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nekomangaorg/similar-processor/internal"
	"github.com/spf13/cobra"
	"go.uber.org/ratelimit"
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
	calculateMangaUpdatesNewIdMapping(cmd.Context(), internal.StreamAllManga(), totalManga)

	syncMangaBakaFromSeries()

	fmt.Printf("Finished all mappings in %s\n", time.Since(initialStart))

}

func calculateMangaUpdatesNewIdMapping(ctx context.Context, mangaList iter.Seq[internal.Manga], totalManga int) {
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
		exists, err := muEntryExistsInNewIDDatabase(data.uuid)
		if err != nil {
			fmt.Printf("failed to check if entry exists for %s: %v\n", data.uuid, err)
			continue
		}
		if exists {
			continue
		}

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
			if !AddAlreadyConvertedId(ctx, index, totalManga, uuid, muLink, limiter) && !CheckAndAddLegacyId(ctx, index, totalManga, uuid, muLink, limiter) {
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

func syncMangaBakaFromSeries() {
	fmt.Println("\nSyncing missing MangaBaka entries from remote SQLite API...")

	fmt.Println("Loading mapping tables into memory...")
	anilistMap := loadMappingIntoMap(internal.TableAnilist)
	malMap := loadMappingIntoMap(internal.TableMyanimelist)
	kitsuMap := loadMappingIntoMap(internal.TableKitsu)
	animePlanetMap := loadMappingIntoMap(internal.TableAnimePlanet)
	muOldMap := loadMappingIntoMap(internal.TableMangaupdates)
	muNewMap := loadMappingIntoMap(internal.TableMangaupdatesNewId)
	mangaBakaMap := loadMappingIntoMap(internal.TableMangaBaka)

	downloadUrl := "https://api.mangabaka.dev/v1/database/series.sqlite.tar.gz"
	fmt.Printf("Downloading and extracting %s...\n", downloadUrl)
	resp, err := http.Get(downloadUrl)
	internal.CheckErr(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Received HTTP %d from MangaBaka API\n", resp.StatusCode)
		return
	}

	gzReader, err := gzip.NewReader(resp.Body)
	internal.CheckErr(err)
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	tempDbPath := "data/temp_remote_series.sqlite"
	foundDb := false

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		internal.CheckErr(err)

		if strings.HasSuffix(header.Name, ".sqlite") || strings.HasSuffix(header.Name, ".db") {
			fmt.Printf("Found %s in archive, writing to temporary file...\n", header.Name)

			outFile, err := os.Create(tempDbPath)
			internal.CheckErr(err)

			_, err = io.Copy(outFile, tarReader)
			outFile.Close()
			internal.CheckErr(err)

			foundDb = true
			break
		}
	}

	if !foundDb {
		fmt.Println("Error: Could not find a .sqlite file inside the tar archive.")
		return
	}

	defer os.Remove(tempDbPath)

	fmt.Println("Connecting to remote database and processing series...")
	remoteDB, err := sql.Open("sqlite3", tempDbPath)
	internal.CheckErr(err)
	defer remoteDB.Close()

	rows, err := remoteDB.Query(`
		SELECT id, source_anilist_id, source_my_anime_list_id, source_kitsu_id, source_anime_planet_id, source_manga_updates_id 
		FROM series 
		WHERE merged_with IS NULL`)
	internal.CheckErr(err)
	defer rows.Close()

	tx, err := internal.DB.Begin()
	internal.CheckErr(err)
	defer tx.Rollback()

	newEntriesCount := 0
	processedCount := 0

	// NEW: Audit trackers
	noExternalIdCount := 0
	noBridgeFoundCount := 0
	sqlErrorCount := 0

	// Helper to safely extract and trim whitespace from pointers
	formatID := func(id *string) string {
		if id == nil {
			return ""
		}
		return strings.TrimSpace(*id)
	}

	for rows.Next() {
		var id int
		var alID, malID, ktID, apID, muID *string

		err := rows.Scan(&id, &alID, &malID, &ktID, &apID, &muID)
		internal.CheckErr(err)

		processedCount++
		if processedCount%50000 == 0 {
			fmt.Printf("Processed %d series...\n", processedCount)
		}

		al := formatID(alID)
		mal := formatID(malID)
		kt := formatID(ktID)
		ap := formatID(apID)
		mu := formatID(muID)

		var mdexUUID string

		if al != "" && anilistMap[al] != "" {
			mdexUUID = anilistMap[al]
		} else if mal != "" && malMap[mal] != "" {
			mdexUUID = malMap[mal]
		} else if kt != "" && kitsuMap[kt] != "" {
			mdexUUID = kitsuMap[kt]
		} else if ap != "" && animePlanetMap[ap] != "" {
			mdexUUID = animePlanetMap[ap]
		} else if mu != "" && muNewMap[mu] != "" {
			mdexUUID = muNewMap[mu]
		} else if mu != "" && muOldMap[mu] != "" {
			mdexUUID = muOldMap[mu]
		}

		if mdexUUID != "" {
			strId := fmt.Sprintf("%d", id)

			if mangaBakaMap[strId] == "" {
				// NEW: Re-added ON CONFLICT DO UPDATE so duplicate UUIDs don't crash the insert silently
				_, err := tx.Exec("INSERT INTO MANGABAKA(UUID, ID) VALUES (?, ?) ON CONFLICT (UUID) DO UPDATE SET ID=excluded.ID", mdexUUID, id)
				if err == nil {
					mangaBakaMap[strId] = mdexUUID
					newEntriesCount++
				} else {
					sqlErrorCount++
					fmt.Printf(">> SQL Error on ID %d: %v\n", id, err)
				}
			}
		} else {
			// If we didn't find a UUID, figure out why for the audit
			if al == "" && mal == "" && kt == "" && ap == "" && mu == "" {
				noExternalIdCount++
			} else {
				noBridgeFoundCount++
			}
		}
	}

	internal.CheckErr(tx.Commit())

	fmt.Printf("\n--- MangaBaka Sync Audit ---\n")
	fmt.Printf("Total Series Processed: %d\n", processedCount)
	fmt.Printf("New Entries Added: %d\n", newEntriesCount)
	fmt.Printf("Skipped (No External IDs in MangaBaka): %d\n", noExternalIdCount)
	fmt.Printf("Skipped (External IDs found, but no match in local MangaDex maps): %d\n", noBridgeFoundCount)
	fmt.Printf("Failed (SQL Errors): %d\n", sqlErrorCount)
	fmt.Println("----------------------------")
}

func loadMappingIntoMap(tableName string) map[string]string {
	m := make(map[string]string)

	if !internal.IsValidMappingTable(tableName) {
		fmt.Printf("loadMappingIntoMap: invalid table name %s\n", tableName)
		return m
	}

	query := fmt.Sprintf("SELECT ID, UUID FROM %s", tableName)
	rows, err := internal.DB.Query(query)

	if err != nil {
		return m
	}
	defer rows.Close()

	for rows.Next() {
		var id, uuid string
		if err := rows.Scan(&id, &uuid); err == nil {
			m[id] = uuid
		}
	}
	return m
}
