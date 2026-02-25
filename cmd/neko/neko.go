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
	tx, _ := nekoDb.Begin()

	for _, manga := range mangaList {
		nekoEntry := internal.DbNeko{}
		nekoEntry.UUID = manga.Id
		populateField(internal.TableAnilist, manga.Id, &nekoEntry.ANILIST)
		populateField(internal.TableAnimePlanet, manga.Id, &nekoEntry.ANIMEPLANET)
		populateField(internal.TableBookWalker, manga.Id, &nekoEntry.BOOKWALKER)
		populateField(internal.TableKitsu, manga.Id, &nekoEntry.KITSU)
		populateField(internal.TableMyanimelist, manga.Id, &nekoEntry.MYANIMELIST)
		populateField(internal.TableMangaupdates, manga.Id, &nekoEntry.MANGAUPDATES)
		populateField(internal.TableMangaupdatesNewId, manga.Id, &nekoEntry.MANGAUPDATES_NEW)
		populateField(internal.TableNovelUpdates, manga.Id, &nekoEntry.NOVEL_UPDATES)

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
	rows, err := internal.DB.Query("SELECT UUID, ID FROM "+table+" WHERE UUID = ?", uuid)
	internal.CheckErr(err)
	defer rows.Close()
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

func populateField(table string, uuid string, target *string) {
	generic := getGeneric(table, uuid)
	if generic.UUID != "" {
		*target = generic.ID
	}
}
