package calculate

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/similar-manga/similar/internal"
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

// InitSimilarInsertStmt initializes the prepared statement for InsertSimilarData.
// It is exposed for testing purposes to allow resetting the global state.
func InitSimilarInsertStmt() {
	var err error
	similarInsertStmt, err = internal.DB.Prepare("INSERT INTO " + internal.TableSimilar + " (UUID, JSON) VALUES (?, ?)")
	internal.CheckErr(err)
}

// ResetSimilarInsertStmt closes the prepared statement and resets the once flag.
// This is strictly used for testing and benching.
func ResetSimilarInsertStmt() {
	similarInsertOnce = sync.Once{}
	if similarInsertStmt != nil {
		similarInsertStmt.Close()
		similarInsertStmt = nil
	}
}

func InsertSimilarData(similarData internal.SimilarManga) {
	similarInsertOnce.Do(InitSimilarInsertStmt)

	dst := &bytes.Buffer{}
	jsonSimilar, _ := json.Marshal(similarData)
	err := json.Compact(dst, jsonSimilar)
	internal.CheckErr(err)

	_, err = similarInsertStmt.Exec(similarData.Id, dst.Bytes())
	internal.CheckErr(err)
}

func getDBSimilar() []internal.DbSimilar {
	rows, err := internal.DB.Query("SELECT UUID, JSON FROM SIMILAR ORDER BY UUID ASC")
	internal.CheckErr(err)
	defer rows.Close()

	var similarList []internal.DbSimilar
	for rows.Next() {
		similar := internal.DbSimilar{}
		err = rows.Scan(&similar.Id, &similar.JSON)
		internal.CheckErr(err)
		similarList = append(similarList, similar)
	}
	internal.CheckErr(rows.Err())
	return similarList
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
	genericList := getAllGenericFromTable(tableName)
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

func getAllGenericFromTable(tableName string) []internal.DbGeneric {
	switch tableName {
	case internal.TableAnilist, internal.TableAnimePlanet, internal.TableBookWalker, internal.TableKitsu, internal.TableMyanimelist, internal.TableMangaupdates, internal.TableMangaupdatesNewId, internal.TableNovelUpdates:
		// OK
	default:
		log.Fatalf("getAllGenericFromTable: invalid table name %s", tableName)
	}

	rows, err := internal.DB.Query("SELECT UUID, ID FROM " + tableName + " ORDER BY UUID asc ")
	internal.CheckErr(err)
	defer rows.Close()

	var genericList []internal.DbGeneric
	for rows.Next() {
		generic := internal.DbGeneric{}
		err = rows.Scan(&generic.UUID, &generic.ID)
		internal.CheckErr(err)
		genericList = append(genericList, generic)
	}
	internal.CheckErr(rows.Err())
	return genericList
}

func CreateMappingsFile(fileName string) *os.File {
	file, err := os.Create("data/mappings/" + fileName + ".txt")
	if err != nil {
		log.Fatal(err)
	}
	return file
}
