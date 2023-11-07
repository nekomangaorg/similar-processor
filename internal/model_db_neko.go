package internal

import (
	_ "github.com/mattn/go-sqlite3"
)

type DbNeko struct {
	UUID             string
	ANILIST          string
	ANIMEPLANET      string
	BOOKWALKER       string
	MANGAUPDATES     string
	MANGAUPDATES_NEW string
	NOVEL_UPDATES    string
	KITSU            string
	MYANIMELIST      string
}
