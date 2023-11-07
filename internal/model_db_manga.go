package internal

import (
	_ "github.com/mattn/go-sqlite3"
)

type DbManga struct {
	Id   string
	JSON string
	DATE string
}
