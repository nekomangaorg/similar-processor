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
	os.MkdirAll("debug", 0777)
	file, err := os.OpenFile("debug/"+fileName+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	internal.CheckErr(err)
	file.WriteString(line + "\n")
	file.Close()
}

func ExportAniList() {
	genericList := getAllGenericFromTable(internal.TableAnilist)
	exportGeneric("anilist2mdex", genericList)
}

func ExportAnimePlanet() {
	genericList := getAllGenericFromTable(internal.TableAnimePlanet)
	exportGeneric("animeplanet2mdex", genericList)
}
func ExportBookWalker() {
	genericList := getAllGenericFromTable(internal.TableBookWalker)
	exportGeneric("bookwalker2mdex", genericList)
}

func ExportMangaUpdates() {
	genericList := getAllGenericFromTable(internal.TableMangaupdates)
	exportGeneric("mangaupdates2mdex", genericList)
}

func ExportNovelUpdates() {
	genericList := getAllGenericFromTable(internal.TableNovelUpdates)
	exportGeneric("novelupdates2mdex", genericList)
}

func ExportKitsu() {
	genericList := getAllGenericFromTable(internal.TableKitsu)
	exportGeneric("kitsu2mdex", genericList)
}

func ExportMyAnimeList() {
	genericList := getAllGenericFromTable(internal.TableMyanimelist)
	exportGeneric("myanimelist2mdex", genericList)
}

func ExportMangaUpdatesNewIds() {
	genericList := getAllGenericFromTable(internal.TableMangaupdatesNewId)
	exportGeneric("mangaupdates_new2mdex", genericList)
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
