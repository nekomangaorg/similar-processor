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

var mappingTables = []string{
	internal.TableAnilist,
	internal.TableAnimePlanet,
	internal.TableBookWalker,
	internal.TableKitsu,
	internal.TableMyanimelist,
	internal.TableMangaupdates,
	internal.TableMangaupdatesNewId,
	internal.TableNovelUpdates,
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
	for _, table := range mappingTables {
		mappings[table] = getAllMappings(table)
	}

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

		for table, mapping := range mappings {
			if val, ok := mapping[manga.Id]; ok {
				setNekoField(&nekoEntry, table, val)
			}
		}

		insertNekoEntry(tx, nekoEntry)
	}
}

func setNekoField(nekoEntry *internal.DbNeko, table, value string) {
	switch table {
	case internal.TableAnilist:
		nekoEntry.ANILIST = value
	case internal.TableAnimePlanet:
		nekoEntry.ANIMEPLANET = value
	case internal.TableBookWalker:
		nekoEntry.BOOKWALKER = value
	case internal.TableKitsu:
		nekoEntry.KITSU = value
	case internal.TableMyanimelist:
		nekoEntry.MYANIMELIST = value
	case internal.TableMangaupdates:
		nekoEntry.MANGAUPDATES = value
	case internal.TableMangaupdatesNewId:
		nekoEntry.MANGAUPDATES_NEW = value
	case internal.TableNovelUpdates:
		nekoEntry.NOVEL_UPDATES = value
	default:
		fmt.Fprintf(os.Stderr, "Warning: unhandled table in setNekoField: %s\n", table)
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
