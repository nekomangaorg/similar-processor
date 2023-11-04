package calculate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/similar-manga/similar/internal"
	"log"
	"os"
)

func GetAllManga() []internal.Manga {
	rows, err := internal.DB.Query("SELECT JSON FROM MANGA ORDER BY UUID ASC ")
	defer rows.Close()
	internal.CheckErr(err)

	var mangaList []internal.Manga
	for rows.Next() {
		manga := internal.Manga{}
		var jsonManga []byte
		rows.Scan(&jsonManga)
		err := json.Unmarshal(jsonManga, &manga)
		if err != nil {
			fmt.Printf(string(jsonManga))
		}
		internal.CheckErr(err)
		mangaList = append(mangaList, manga)
	}
	return mangaList
}

func DeleteSimilarDB() {
	_, err := internal.DB.Exec("DELETE FROM SIMILAR")
	internal.CheckErr(err)
}

func InsertSimilarData(similarData internal.SimilarManga) {
	dst := &bytes.Buffer{}
	jsonSimilar, _ := json.Marshal(similarData)
	err := json.Compact(dst, jsonSimilar)
	internal.CheckErr(err)
	stmt, err := internal.DB.Prepare("INSERT INTO SIMILAR (UUID, JSON) VALUES (?, ?)")
	internal.CheckErr(err)
	defer stmt.Close()
	_, err = stmt.Exec(similarData.Id, dst.Bytes())
	internal.CheckErr(err)
}

func getDBSimilar() []internal.DbSimilar {
	rows, err := internal.DB.Query("SELECT UUID, JSON FROM SIMILAR")
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

func ExportSimilar() {
	fmt.Printf("Exporting All Similar to csv files\n")
	os.RemoveAll("data/similar/")
	os.MkdirAll("data/similar/", 0777)
	similarList := getDBSimilar()
	for _, similar := range similarList {
		suffix := similar.Id[0:3]
		file, err := os.OpenFile("data/similar/"+suffix+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			log.Fatal(err)
		}
		file.WriteString(similar.Id + ":::||@!@||:::" + similar.JSON + "\n")
		file.Close()

	}
}

func ExportMangaUpdatesNewIds() {
	fmt.Printf("Exporting MangaUpdate new Ids\n")
	file, err := os.Create("data/mappings/mangaupdates_new2mdex.csv")
	internal.CheckErr(err)
	genericList := getMangaUpdatesNewDB()
	for _, entry := range genericList {
		file.WriteString(entry.ID + ":::||@!@||:::" + entry.UUID + "\n")
	}
	file.Close()
}

func getMangaUpdatesNewDB() []internal.DbGeneric {
	rows, err := internal.DB.Query("SELECT UUID, ID FROM MANGAUPDATES_NEW ORDER BY UUID asc ")
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
	file, err := os.Create("data/mappings/" + fileName + ".csv")
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func OpenMappingsFile(fileName string) *os.File {
	file, err := os.OpenFile("data/mappings/"+fileName+".csv", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal(err)
	}
	return file
}
