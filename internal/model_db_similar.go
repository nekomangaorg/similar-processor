package internal

import (
	_ "github.com/mattn/go-sqlite3"
)

type DbSimilar struct {
	Id   string
	JSON string
}
