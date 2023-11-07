package internal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
)

const TableMangaupdates = "MANGAUPDATES_OLD"
const TableMangaupdatesNewId = "MANGAUPDATES_NEW"
const TableAnilist = "ANILIST"
const TableMyanimelist = "MYANIMELIST"
const TableManga = "MANGA"
const TableSimilar = "SIMILAR"
const TableNovelUpdates = "NOVEL_UPDATES"
const TableKitsu = "KITSU"
const TableBookWalker = "BOOK_WALKER"
const TableAnimePlanet = "ANIME_PLANET"

const TableNekoMappings = "mappings"

var DB *sql.DB

func ConnectDB() {
	db, err := sql.Open("sqlite3", "data/data.db")
	if err != nil {
		panic(err.Error())
	}
	db.SetMaxOpenConns(1)
	DB = db
}

func ConnectNekoDB(name string) *sql.DB {
	db, err := sql.Open("sqlite3", "data/"+name+".db")
	if err != nil {
		panic(err.Error())
	}
	db.SetMaxOpenConns(1)
	return db
}

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func GetAllManga() []Manga {
	rows, err := DB.Query("SELECT JSON FROM " + TableManga + " ORDER BY UUID ASC ")
	defer rows.Close()
	CheckErr(err)

	var mangaList []Manga
	for rows.Next() {
		manga := Manga{}
		var jsonManga []byte
		rows.Scan(&jsonManga)
		err := json.Unmarshal(jsonManga, &manga)
		if err != nil {
			fmt.Printf(string(jsonManga))
		}
		CheckErr(err)
		mangaList = append(mangaList, manga)
	}
	return mangaList
}
