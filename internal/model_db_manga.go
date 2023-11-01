package internal

import (
	_ "github.com/mattn/go-sqlite3"
)

type DB_Manga struct {
	Id           string
	Manga_Json   string
	Similar_Json string
}
