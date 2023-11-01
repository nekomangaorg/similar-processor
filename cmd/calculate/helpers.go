package calculate

import (
	"encoding/json"
	"github.com/similar-manga/similar/internal"
)

func GetAllManga() []internal.Manga {
	rows, err := internal.MangaDB.Query("SELECT MANGA_JSON FROM MANGA")
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
