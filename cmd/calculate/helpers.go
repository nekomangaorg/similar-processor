package calculate

import (
	"bytes"
	"encoding/json"
	"github.com/similar-manga/similar/internal"
	"log"
	"os"
)

func DeleteSimilarDB() {
	_, err := internal.DB.Exec("DELETE FROM " + internal.TableSimilar)
	internal.CheckErr(err)
}

func InsertSimilarData(similarData internal.SimilarManga) {
	dst := &bytes.Buffer{}
	jsonSimilar, _ := json.Marshal(similarData)
	err := json.Compact(dst, jsonSimilar)
	internal.CheckErr(err)
	stmt, err := internal.DB.Prepare("INSERT INTO " + internal.TableSimilar + " (UUID, JSON) VALUES (?, ?)")
	internal.CheckErr(err)
	defer stmt.Close()
	_, err = stmt.Exec(similarData.Id, dst.Bytes())
	internal.CheckErr(err)
}

func getDBSimilar() []internal.DbSimilar {
	rows, err := internal.DB.Query("SELECT UUID, JSON FROM SIMILAR ORDER BY UUID ASC")
	defer rows.Close()
	internal.CheckErr(err)

	var similarList []internal.DbSimilar
	for rows.Next() {
		similar := internal.DbSimilar{}
		rows.Scan(&similar.Id, &similar.JSON)
		internal.CheckErr(err)
		similarList = append(similarList, similar)
	}
	return similarList
}

func WriteLineToDebugFile(fileName string, line string) {
	os.MkdirAll("debug", 0700)
	file, err := os.OpenFile("debug/"+fileName+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	internal.CheckErr(err)
	file.WriteString(line + "\n")
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
	rows, err := internal.DB.Query("SELECT UUID, ID FROM " + tableName + " ORDER BY UUID asc ")
	defer rows.Close()
	internal.CheckErr(err)

	var genericList []internal.DbGeneric
	for rows.Next() {
		similar := internal.DbGeneric{}
		rows.Scan(&similar.UUID, &similar.ID)
		internal.CheckErr(err)
		genericList = append(genericList, similar)
	}
	return genericList
}

func CreateMappingsFile(fileName string) *os.File {
	file, err := os.Create("data/mappings/" + fileName + ".txt")
	if err != nil {
		log.Fatal(err)
	}
	return file
}
