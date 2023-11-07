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
	Short: "neko command",
	Long: `
Generate the neko mapping.db`,
	Run: runNeko,
}

func init() {
	cmd.RootCmd.AddCommand(nekoCmd)
}

func runNeko(command *cobra.Command, args []string) {
	initialStart := time.Now()

	nekoDb := createNekoMappingDB()
	fmt.Println("Starting neko export")
	mangaList := internal.GetAllManga()
	tx, _ := nekoDb.Begin()

	for _, manga := range mangaList {
		nekoEntry := internal.DbNeko{}
		nekoEntry.UUID = manga.Id
		generic := getGeneric(internal.TableAnilist, manga.Id)
		if generic.UUID != "" {
			nekoEntry.ANILIST = generic.ID
		}

		generic = getGeneric(internal.TableAnimePlanet, manga.Id)
		if generic.UUID != "" {
			nekoEntry.ANIMEPLANET = generic.ID
		}

		generic = getGeneric(internal.TableBookWalker, manga.Id)
		if generic.UUID != "" {
			nekoEntry.BOOKWALKER = generic.ID
		}

		generic = getGeneric(internal.TableKitsu, manga.Id)
		if generic.UUID != "" {
			nekoEntry.KITSU = generic.ID
		}

		generic = getGeneric(internal.TableMyanimelist, manga.Id)
		if generic.UUID != "" {
			nekoEntry.MYANIMELIST = generic.ID
		}

		generic = getGeneric(internal.TableMangaupdates, manga.Id)
		if generic.UUID != "" {
			nekoEntry.MANGAUPDATES = generic.ID
		}

		generic = getGeneric(internal.TableMangaupdatesNewId, manga.Id)
		if generic.UUID != "" {
			nekoEntry.MANGAUPDATES_NEW = generic.ID
		}

		generic = getGeneric(internal.TableNovelUpdates, manga.Id)
		if generic.UUID != "" {
			nekoEntry.NOVEL_UPDATES = generic.ID
		}

		insertNekoEntry(tx, nekoEntry)
	}

	err := tx.Commit()
	internal.CheckErr(err)
	fmt.Printf("Finished neko export in %s\n", time.Since(initialStart))
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

func getGeneric(table string, uuid string) internal.DbGeneric {
	rows, err := internal.DB.Query("SELECT UUID, ID FROM " + table + " WHERE UUID = '" + uuid + "'")
	defer rows.Close()
	internal.CheckErr(err)
	generic := internal.DbGeneric{}
	if rows.Next() {
		rows.Scan(&generic.UUID, &generic.ID)
	} else {
	}
	return generic
}

func insertNekoEntry(tx *sql.Tx, nekoEntry internal.DbNeko) {
	_, err := tx.Exec("INSERT INTO "+internal.TableNekoMappings+" (mdex, al, ap, bw, mu, mu_new, nu, kt , mal) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", nekoEntry.UUID, nekoEntry.ANILIST, nekoEntry.ANIMEPLANET, nekoEntry.BOOKWALKER, nekoEntry.MANGAUPDATES, nekoEntry.MANGAUPDATES_NEW, nekoEntry.NOVEL_UPDATES, nekoEntry.KITSU, nekoEntry.MYANIMELIST)
	internal.CheckErr(err)
}
