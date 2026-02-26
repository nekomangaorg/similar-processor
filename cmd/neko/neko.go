package neko

import (
	"database/sql"
	"fmt"
	"github.com/similar-manga/similar/cmd"
	"github.com/similar-manga/similar/internal"
	"github.com/spf13/cobra"
	"io"
	"os"
	"time"
)

var nekoCmd = &cobra.Command{
	Use:   "neko",
	Short: "Generate a neko mapping database",
	Long:  `Generate the neko mapping database file`,
	Run:   runNeko,
}

func init() {
	cmd.RootCmd.AddCommand(nekoCmd)
}

func runNeko(command *cobra.Command, args []string) {
	initialStart := time.Now()

	nekoDb := createNekoMappingDB()
	fmt.Println("Starting neko export")
	mangaList := internal.GetAllManga()

	mappings := make(map[string]map[string]string)
	mappings[internal.TableAnilist] = getAllMappings(internal.TableAnilist)
	mappings[internal.TableAnimePlanet] = getAllMappings(internal.TableAnimePlanet)
	mappings[internal.TableBookWalker] = getAllMappings(internal.TableBookWalker)
	mappings[internal.TableKitsu] = getAllMappings(internal.TableKitsu)
	mappings[internal.TableMyanimelist] = getAllMappings(internal.TableMyanimelist)
	mappings[internal.TableMangaupdates] = getAllMappings(internal.TableMangaupdates)
	mappings[internal.TableMangaupdatesNewId] = getAllMappings(internal.TableMangaupdatesNewId)
	mappings[internal.TableNovelUpdates] = getAllMappings(internal.TableNovelUpdates)

	tx, err := nekoDb.Begin()
	internal.CheckErr(err)

	processMangaList(tx, mangaList, mappings)

	err = tx.Commit()
	internal.CheckErr(err)
	fmt.Printf("Finished neko export in %s\n", time.Since(initialStart))
}

func processMangaList(tx *sql.Tx, mangaList []internal.Manga, mappings map[string]map[string]string) {
	for _, manga := range mangaList {
		nekoEntry := internal.DbNeko{}
		nekoEntry.UUID = manga.Id

		if val, ok := mappings[internal.TableAnilist][manga.Id]; ok {
			nekoEntry.ANILIST = val
		}
		if val, ok := mappings[internal.TableAnimePlanet][manga.Id]; ok {
			nekoEntry.ANIMEPLANET = val
		}
		if val, ok := mappings[internal.TableBookWalker][manga.Id]; ok {
			nekoEntry.BOOKWALKER = val
		}
		if val, ok := mappings[internal.TableKitsu][manga.Id]; ok {
			nekoEntry.KITSU = val
		}
		if val, ok := mappings[internal.TableMyanimelist][manga.Id]; ok {
			nekoEntry.MYANIMELIST = val
		}
		if val, ok := mappings[internal.TableMangaupdates][manga.Id]; ok {
			nekoEntry.MANGAUPDATES = val
		}
		if val, ok := mappings[internal.TableMangaupdatesNewId][manga.Id]; ok {
			nekoEntry.MANGAUPDATES_NEW = val
		}
		if val, ok := mappings[internal.TableNovelUpdates][manga.Id]; ok {
			nekoEntry.NOVEL_UPDATES = val
		}

		insertNekoEntry(tx, nekoEntry)
	}
}

func getAllMappings(table string) map[string]string {
	rows, err := internal.DB.Query("SELECT UUID, ID FROM " + table)
	internal.CheckErr(err)
	defer rows.Close()

	mapping := make(map[string]string)
	for rows.Next() {
		var uuid, id string
		if err := rows.Scan(&uuid, &id); err == nil {
			mapping[uuid] = id
		} else {
			fmt.Printf("Warning: failed to scan row in table %s: %v\n", table, err)
		}
	}
	internal.CheckErr(rows.Err())
	return mapping
}

func createNekoMappingDB() *sql.DB {
	fmt.Println("Creating neko_mapping.db")
	src, err := os.Open("data/default_empty_neko_mapping.db")
	currentTime := time.Now().Format(time.DateOnly)
	internal.CheckErr(err)
	dbName := currentTime + "_neko_mapping"
	defer src.Close()
	dst, err := os.Create("data/" + dbName + ".db")
	internal.CheckErr(err)
	defer dst.Close()

	buf := make([]byte, 1024)
	for {
		n, err := src.Read(buf)
		if err != nil && err != io.EOF {
			internal.CheckErr(err)
			break
		}

		if n == 0 {
			break
		}

		if _, err := dst.Write(buf[:n]); err != nil {
			internal.CheckErr(err)
			break
		}
	}

	return internal.ConnectNekoDB(dbName)
}

func insertNekoEntry(tx *sql.Tx, nekoEntry internal.DbNeko) {
	_, err := tx.Exec("INSERT INTO "+internal.TableNekoMappings+" (mdex, al, ap, bw, mu, mu_new, nu, kt , mal) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", nekoEntry.UUID, nekoEntry.ANILIST, nekoEntry.ANIMEPLANET, nekoEntry.BOOKWALKER, nekoEntry.MANGAUPDATES, nekoEntry.MANGAUPDATES_NEW, nekoEntry.NOVEL_UPDATES, nekoEntry.KITSU, nekoEntry.MYANIMELIST)
	internal.CheckErr(err)
}
