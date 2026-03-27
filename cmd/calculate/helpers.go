package calculate

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/similar-manga/similar/internal"
	"iter"
	"log"
	"os"
	"path/filepath"
	"sync"
)

func DeleteSimilarDB() {
	_, err := internal.DB.Exec("DELETE FROM " + internal.TableSimilar)
	internal.CheckErr(err)
}

var (
	similarInsertStmt *sql.Stmt
	similarInsertOnce sync.Once
)

// initSimilarInsertStmt initializes the prepared statement for InsertSimilarData.
// It is exposed for testing purposes to allow resetting the global state.
func initSimilarInsertStmt() {
	var err error
	similarInsertStmt, err = internal.DB.Prepare("INSERT INTO " + internal.TableSimilar + " (UUID, JSON) VALUES (?, ?)")
	internal.CheckErr(err)
}

// resetSimilarInsertStmt closes the prepared statement and resets the once flag.
// This is strictly used for testing and benching.
func resetSimilarInsertStmt() {
	similarInsertOnce = sync.Once{}
	if similarInsertStmt != nil {
		similarInsertStmt.Close()
		similarInsertStmt = nil
	}
}

func InsertSimilarData(similarData internal.SimilarManga) {
	similarInsertOnce.Do(initSimilarInsertStmt)

	jsonSimilar, err := json.Marshal(similarData)
	internal.CheckErr(err)

	_, err = similarInsertStmt.Exec(similarData.Id, jsonSimilar)
	internal.CheckErr(err)
}

// getDBSimilar streams similar manga entries from the database using an iterator.
// Overclock optimization: Previously, this loaded the entire 'SIMILAR' table into memory
// as a slice before returning, causing an O(n) memory spike and large GC pressure on large datasets.
// By returning an iter.Seq, we turn the batch load into an O(1) streaming pipeline, allowing
// the caller to process and export records as they are yielded directly from the database connection.
func getDBSimilar() iter.Seq[internal.DbSimilar] {
	return func(yield func(internal.DbSimilar) bool) {
		rows, err := internal.DB.Query("SELECT UUID, JSON FROM SIMILAR ORDER BY UUID ASC")
		if err != nil {
			log.Printf("ERROR: failed to query similar: %v", err)
			return
		}
		defer rows.Close() // Safely clean up the database rows iterator

		for rows.Next() {
			similar := internal.DbSimilar{}
			if err := rows.Scan(&similar.Id, &similar.JSON); err != nil {
				log.Printf("ERROR: failed to scan similar row: %v", err)
				continue
			}

			// Yield the row to the caller's loop. If the caller breaks out early,
			// yield returns false and we safely terminate, triggering the defer rows.Close()
			if !yield(similar) {
				return
			}
		}
		if err := rows.Err(); err != nil {
			log.Printf("ERROR: error iterating similar rows: %v", err)
		}
	}
}

func WriteLineToDebugFile(fileName string, line string) {
	os.MkdirAll("debug", 0700)
	file, err := os.OpenFile(filepath.Join("debug", filepath.Base(fileName)+".txt"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	internal.CheckErr(err)
	_, err = file.WriteString(line + "\n")
	internal.CheckErr(err)
	file.Close()
}

func exportMapping(tableName string, fileName string) {
	genericList, err := getAllGenericFromTable(tableName)
	internal.CheckErr(err)
	exportGeneric(fileName, genericList)
}

func exportGeneric(fileName string, genericList []internal.DbGeneric) {
	file, err := os.Create("data/mappings/" + fileName + ".txt")
	internal.CheckErr(err)
	for _, entry := range genericList {
		file.WriteString(entry.ID + ":::||@!@||:::" + entry.UUID + "\n")
	}
	file.Close()
}

func getAllGenericFromTable(tableName string) ([]internal.DbGeneric, error) {
	switch tableName {
	case internal.TableAnilist, internal.TableAnimePlanet, internal.TableBookWalker, internal.TableKitsu, internal.TableMyanimelist, internal.TableMangaupdates, internal.TableMangaupdatesNewId, internal.TableNovelUpdates:
		// OK
	default:
		return nil, fmt.Errorf("getAllGenericFromTable: invalid table name %s", tableName)
	}

	rows, err := internal.DB.Query("SELECT UUID, ID FROM " + tableName + " ORDER BY UUID asc ")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genericList []internal.DbGeneric
	for rows.Next() {
		generic := internal.DbGeneric{}
		err = rows.Scan(&generic.UUID, &generic.ID)
		if err != nil {
			return nil, err
		}
		genericList = append(genericList, generic)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return genericList, nil
}

func CreateMappingsFile(fileName string) *os.File {
	file, err := os.Create("data/mappings/" + fileName + ".txt")
	if err != nil {
		log.Fatal(err)
	}
	return file
}
