package internal

import (
	"database/sql"
	"log"
)

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
