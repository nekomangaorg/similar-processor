package internal

import (
	"database/sql"
	"log"
)

var DB *sql.DB

func ConnectMangaDB() {
	db, err := sql.Open("sqlite3", "data/manga.db")
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
