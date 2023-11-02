package calculate

import (
	"bytes"
	"encoding/json"
	"github.com/similar-manga/similar/internal"
	"github.com/similar-manga/similar/similar"
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
