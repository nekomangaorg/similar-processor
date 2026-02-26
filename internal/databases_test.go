package internal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	_ "github.com/mattn/go-sqlite3"
)

func setupDB(b *testing.B) *sql.DB {
	// Setup in-memory DB
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatal(err)
	}
	DB = db

	// Create table
	_, err = db.Exec("CREATE TABLE " + TableManga + " (UUID TEXT PRIMARY KEY, JSON TEXT)")
	if err != nil {
		b.Fatal(err)
	}

	// Insert dummy data
	stmt, err := db.Prepare("INSERT INTO " + TableManga + " (UUID, JSON) VALUES (?, ?)")
	if err != nil {
		b.Fatal(err)
	}
	defer stmt.Close()

	for i := 0; i < 1000; i++ {
		manga := Manga{
			Id: fmt.Sprintf("uuid-%d", i),
			Title: &map[string]string{"en": fmt.Sprintf("Title %d", i)},
		}
		jsonData, _ := json.Marshal(manga)
		_, err = stmt.Exec(manga.Id, string(jsonData))
		if err != nil {
			b.Fatal(err)
		}
	}
	return db
}

func BenchmarkGetAllManga(b *testing.B) {
	db := setupDB(b)
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list := GetAllManga()
		if len(list) != 1000 {
			b.Fatalf("expected 1000 items, got %d", len(list))
		}
	}
}

func BenchmarkStreamAllManga(b *testing.B) {
	db := setupDB(b)
	defer db.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		for _ = range StreamAllManga() {
			count++
		}
		if count != 1000 {
			b.Fatalf("expected 1000 items, got %d", count)
		}
	}
}
