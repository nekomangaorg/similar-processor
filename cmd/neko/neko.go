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

	tx, err := nekoDb.Begin()
	internal.CheckErr(err)

	exportNeko(tx)

	err = tx.Commit()
	internal.CheckErr(err)
	fmt.Printf("Finished neko export in %s\n", time.Since(initialStart))
}

func exportNeko(tx *sql.Tx) {
	query := fmt.Sprintf(`SELECT
		m.UUID,
		al.ID,
		ap.ID,
		bw.ID,
		mu.ID,
		mun.ID,
		nu.ID,
		kt.ID,
		mal.ID
	FROM %s m
	LEFT JOIN %s al ON m.UUID = al.UUID
	LEFT JOIN %s ap ON m.UUID = ap.UUID
	LEFT JOIN %s bw ON m.UUID = bw.UUID
	LEFT JOIN %s mu ON m.UUID = mu.UUID
	LEFT JOIN %s mun ON m.UUID = mun.UUID
	LEFT JOIN %s nu ON m.UUID = nu.UUID
	LEFT JOIN %s kt ON m.UUID = kt.UUID
	LEFT JOIN %s mal ON m.UUID = mal.UUID
	ORDER BY m.UUID`,
		internal.TableManga,
		internal.TableAnilist,
		internal.TableAnimePlanet,
		internal.TableBookWalker,
		internal.TableMangaupdates,
		internal.TableMangaupdatesNewId,
		internal.TableNovelUpdates,
		internal.TableKitsu,
		internal.TableMyanimelist)

	rows, err := internal.DB.Query(query)
	internal.CheckErr(err)
	defer rows.Close()

	for rows.Next() {
		var uuid string
		var al, ap, bw, mu, mun, nu, kt, mal sql.NullString

		if err := rows.Scan(&uuid, &al, &ap, &bw, &mu, &mun, &nu, &kt, &mal); err != nil {
			fmt.Printf("Warning: failed to scan row: %v\n", err)
			continue
		}

		nekoEntry := internal.DbNeko{
			UUID:             uuid,
			ANILIST:          al.String,
			ANIMEPLANET:      ap.String,
			BOOKWALKER:       bw.String,
			MANGAUPDATES:     mu.String,
			MANGAUPDATES_NEW: mun.String,
			NOVEL_UPDATES:    nu.String,
			KITSU:            kt.String,
			MYANIMELIST:      mal.String,
		}

		insertNekoEntry(tx, nekoEntry)
	}
	internal.CheckErr(rows.Err())
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
