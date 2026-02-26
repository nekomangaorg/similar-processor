package neko

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/similar-manga/similar/internal"
	_ "github.com/mattn/go-sqlite3"
)

func setupBenchmarkDB(b *testing.B, numManga int) (*sql.DB, []internal.Manga) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("Failed to open internal memory db: %v", err)
	}

	// Create mapping tables
	for _, table := range mappingTables {
		_, err = db.Exec("CREATE TABLE " + table + " (UUID TEXT, ID TEXT)")
		if err != nil {
			b.Fatalf("Failed to create table %s: %v", table, err)
		}
	}

	// Create MANGA table
	_, err = db.Exec("CREATE TABLE " + internal.TableManga + " (UUID TEXT, JSON TEXT, DATE TEXT)")
	if err != nil {
		b.Fatalf("Failed to create table %s: %v", internal.TableManga, err)
	}

	mangaList := make([]internal.Manga, numManga)
	stmtMap := make(map[string]*sql.Stmt)
	for _, table := range mappingTables {
		stmt, err := db.Prepare("INSERT INTO " + table + " (UUID, ID) VALUES (?, ?)")
		if err != nil {
			b.Fatalf("Failed to prepare statement for table %s: %v", table, err)
		}
		stmtMap[table] = stmt
	}

	mangaStmt, err := db.Prepare("INSERT INTO " + internal.TableManga + " (UUID, JSON, DATE) VALUES (?, ?, ?)")
	if err != nil {
		b.Fatalf("Failed to prepare statement for table %s: %v", internal.TableManga, err)
	}
	defer mangaStmt.Close()

	for i := 0; i < numManga; i++ {
		uuid := fmt.Sprintf("uuid-%d", i)
		jsonStr := fmt.Sprintf(`{"id": "%s", "title": {"en": "Title %d"}}`, uuid, i)
		mangaList[i] = internal.Manga{Id: uuid}

		_, err = mangaStmt.Exec(uuid, jsonStr, "2023-01-01")
		if err != nil {
			b.Fatalf("Failed to insert into MANGA: %v", err)
		}

		// Insert mappings for each table
		for _, table := range mappingTables {
			_, err = stmtMap[table].Exec(uuid, fmt.Sprintf("id-%s-%d", table, i))
			if err != nil {
				b.Fatalf("Failed to insert data: %v", err)
			}
		}
	}

	for _, stmt := range stmtMap {
		stmt.Close()
	}

	return db, mangaList
}

func BenchmarkNekoExport(b *testing.B) {
	// Setup DB with data
	numManga := 100
	db, _ := setupBenchmarkDB(b, numManga)
	defer db.Close()
	internal.DB = db

	outputDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("Failed to open output memory db: %v", err)
	}
	defer outputDB.Close()

	_, err = outputDB.Exec("CREATE TABLE " + internal.TableNekoMappings + " (mdex TEXT, al TEXT, ap TEXT, bw TEXT, mu TEXT, mu_new TEXT, nu TEXT, kt TEXT, mal TEXT)")
	if err != nil {
		b.Fatalf("Failed to create mappings table: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, err := outputDB.Begin()
		if err != nil {
			b.Fatalf("Failed to begin transaction: %v", err)
		}

		exportNeko(tx)

		tx.Rollback()
	}
}
