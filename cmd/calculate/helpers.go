package calculate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/similar-manga/similar/internal"
	"github.com/similar-manga/similar/similar"
	"log"
	"os"
)

func GetAllManga() []internal.Manga {
	rows, err := internal.DB.Query("SELECT JSON FROM MANGA")
	defer rows.Close()
	internal.CheckErr(err)

	var mangaList []internal.Manga
	for rows.Next() {
		manga := internal.Manga{}
		var jsonManga []byte
		rows.Scan(&jsonManga)
		err := json.Unmarshal(jsonManga, &manga)
		internal.CheckErr(err)
		mangaList = append(mangaList, manga)
	}
	return mangaList
}

func DeleteSimilarDB() {
	_, err := internal.DB.Exec("DELETE FROM SIMILAR")
	internal.CheckErr(err)
}

func InsertSimilarData(similarData similar.SimilarManga) {
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
		file.WriteString(similar.Id + ":::" + similar.JSON + "\n")
		file.Close()

	}

}
