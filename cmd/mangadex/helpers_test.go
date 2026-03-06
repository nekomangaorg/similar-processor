package mangadex

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/similar-manga/similar/internal"
	_ "github.com/mattn/go-sqlite3"
)

var testUUIDs []string

func init() {
	testUUIDs = make([]string, 1000)
	for i := 0; i < 1000; i++ {
		testUUIDs[i] = fmt.Sprintf("uuid-%d", i)
	}
}

func setupTestDB() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("CREATE TABLE " + internal.TableManga + " (UUID TEXT PRIMARY KEY, JSON TEXT, DATE TEXT)")
	if err != nil {
		panic(err)
	}

	stmt, err := db.Prepare("INSERT INTO " + internal.TableManga + " (UUID, JSON, DATE) VALUES (?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	for _, uuid := range testUUIDs {
		_, err = stmt.Exec(uuid, "{}", "2023-01-01")
		if err != nil {
			panic(err)
		}
	}

	return db
}

func BenchmarkExistsInDatabase(b *testing.B) {
	db := setupTestDB()
	defer db.Close()
	internal.DB = db

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExistsInDatabase(testUUIDs[i%1000])
	}
}

func BenchmarkGetExistingMangaUUIDs(b *testing.B) {
	db := setupTestDB()
	defer db.Close()
	internal.DB = db

	batchSize := 100
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := (i * batchSize) % 1000
		end := start + batchSize
		if end > 1000 {
			end = 1000
		}
		GetExistingMangaUUIDs(testUUIDs[start:end])
	}
}

func TestExistsInDatabase(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	internal.DB = db

	// Test existing
	if !ExistsInDatabase("uuid-1") {
		t.Error("uuid-1 should exist")
	}

	// Test non-existing
	if ExistsInDatabase("uuid-9999") {
		t.Error("uuid-9999 should not exist")
	}
}

func TestGetExistingMangaUUIDs(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	internal.DB = db

	uuids := []string{"uuid-1", "uuid-2", "uuid-9999"}
	existing := GetExistingMangaUUIDs(uuids)

	if !existing["uuid-1"] {
		t.Error("uuid-1 should exist")
	}
	if !existing["uuid-2"] {
		t.Error("uuid-2 should exist")
	}
	if existing["uuid-9999"] {
		t.Error("uuid-9999 should not exist")
	}
	if len(existing) != 2 {
		t.Errorf("expected 2 existing uuids, got %d", len(existing))
	}
}
