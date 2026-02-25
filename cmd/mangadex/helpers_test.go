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

func setupTestDB(tb testing.TB) *sql.DB {
	tb.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		tb.Fatalf("failed to open database: %v", err)
	}
	_, err = db.Exec("CREATE TABLE " + internal.TableManga + " (UUID TEXT PRIMARY KEY, JSON TEXT, DATE TEXT)")
	if err != nil {
		tb.Fatalf("failed to create table: %v", err)
	}

	stmt, err := db.Prepare("INSERT INTO " + internal.TableManga + " (UUID, JSON, DATE) VALUES (?, ?, ?)")
	if err != nil {
		tb.Fatalf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for _, uuid := range testUUIDs {
		_, err = stmt.Exec(uuid, "{}", "2023-01-01")
		if err != nil {
			tb.Fatalf("failed to insert data: %v", err)
		}
	}

	return db
}

func BenchmarkExistsInDatabase(b *testing.B) {
	originalDB := internal.DB
	defer func() { internal.DB = originalDB }()

	db := setupTestDB(b)
	defer db.Close()
	internal.DB = db

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExistsInDatabase(testUUIDs[i%1000])
	}
}

func TestExistsInDatabase(t *testing.T) {
	originalDB := internal.DB
	defer func() { internal.DB = originalDB }()

	db := setupTestDB(t)
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
