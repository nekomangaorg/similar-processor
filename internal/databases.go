package internal

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
const TableMangaBaka = "MANGABAKA"

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

// StreamAllManga returns an iterator over all manga in the database.
// This allows processing manga one by one without loading the entire dataset into memory.
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

// GetAllManga loads all manga into memory.
// Deprecated: Use StreamAllManga where possible to reduce memory usage.
func GetAllManga() []Manga {
	var mangaList []Manga
	for manga := range StreamAllManga() {
		mangaList = append(mangaList, manga)
	}
	return mangaList
}

// GetMangaCount returns the total number of manga in the database.
func GetMangaCount() (int, error) {
	var count int
	if err := DB.QueryRow("SELECT COUNT(*) FROM " + TableManga).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// IsValidMappingTable returns true if the table name is a valid mapping table.
// This is used to prevent SQL injection when table names are used in queries.
func IsValidMappingTable(table string) bool {
	switch table {
	case TableAnilist, TableAnimePlanet, TableBookWalker, TableKitsu, TableMyanimelist, TableMangaupdates, TableMangaupdatesNewId, TableNovelUpdates, TableMangaBaka:
		return true
	}
	return false
}

// GetExistingUUIDs returns a map of UUIDs that exist in the specified table.
// It uses SQLite's json_each for efficient bulk lookup.
func GetExistingUUIDs(table string, uuids []string) (map[string]bool, error) {
	if len(uuids) == 0 {
		return make(map[string]bool), nil
	}

	// Validate table name to prevent SQL injection
	isValid := IsValidMappingTable(table)
	if !isValid {
		switch table {
		case TableManga, TableSimilar:
			isValid = true
		}
	}

	if !isValid {
		return nil, fmt.Errorf("GetExistingUUIDs: invalid table name %s", table)
	}

	existing := make(map[string]bool, len(uuids))

	jsonUUIDs, err := json.Marshal(uuids)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT UUID FROM %s WHERE UUID IN (SELECT value FROM json_each(?))", table)
	rows, err := DB.Query(query, string(jsonUUIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var uuid string
		if err := rows.Scan(&uuid); err != nil {
			return nil, err
		}
		existing[uuid] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return existing, nil
}
