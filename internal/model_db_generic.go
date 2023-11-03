package internal

import (
	_ "github.com/mattn/go-sqlite3"
)

type DbGeneric struct {
	UUID string
	ID   string
}
