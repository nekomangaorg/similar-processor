package internal

import (
	"database/sql"
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

var DB *sql.DB

func ConnectDB() {
	db, err := sql.Open("sqlite3", "data/data.db")
	if err != nil {
		panic(err.Error())
	}
	db.SetMaxOpenConns(1)
	DB = db
}

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
