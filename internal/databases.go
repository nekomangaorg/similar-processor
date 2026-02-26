package internal

import (
	"database/sql"
	"encoding/json"
	"iter"
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

func StreamAllManga() iter.Seq[Manga] {
	return func(yield func(Manga) bool) {
		rows, err := DB.Query("SELECT JSON FROM " + TableManga + " ORDER BY UUID ASC ")
		if err != nil {
			log.Printf("ERROR: failed to query manga: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			manga := Manga{}
			var jsonManga []byte
			err := rows.Scan(&jsonManga)
			if err != nil {
				log.Printf("ERROR: failed to scan manga row: %v", err)
				continue
			}
			err = json.Unmarshal(jsonManga, &manga)
			if err != nil {
				log.Printf("ERROR: Failed to unmarshal manga JSON, skipping. Data: %s, Error: %v", string(jsonManga), err)
				continue
			}
			if !yield(manga) {
				return
			}
		}
		if err := rows.Err(); err != nil {
			log.Printf("ERROR: error iterating manga rows: %v", err)
		}
	}
}

func GetAllManga() []Manga {
	var mangaList []Manga
	for manga := range StreamAllManga() {
		mangaList = append(mangaList, manga)
	}
	return mangaList
}
