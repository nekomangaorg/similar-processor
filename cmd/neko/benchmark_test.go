package neko

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/similar-manga/similar/internal"
	_ "github.com/mattn/go-sqlite3"
)

func BenchmarkNekoExport(b *testing.B) {
	// 1. Setup internal.DB
	var err error
	internal.DB, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("Failed to open internal memory db: %v", err)
	}
	defer internal.DB.Close()

	tables := []string{
		internal.TableAnilist,
		internal.TableAnimePlanet,
		internal.TableBookWalker,
		internal.TableKitsu,
		internal.TableMyanimelist,
		internal.TableMangaupdates,
		internal.TableMangaupdatesNewId,
		internal.TableNovelUpdates,
	}

	for _, table := range tables {
		_, err = internal.DB.Exec("CREATE TABLE " + table + " (UUID TEXT, ID TEXT)")
		if err != nil {
			b.Fatalf("Failed to create table %s: %v", table, err)
		}
	}

	// 2. Insert sample data
	numManga := 100 // Simulate 100 manga for the benchmark unit
	mangaList := make([]internal.Manga, numManga)

	stmtMap := make(map[string]*sql.Stmt)
	for _, table := range tables {
		stmt, err := internal.DB.Prepare("INSERT INTO " + table + " (UUID, ID) VALUES (?, ?)")
		if err != nil {
			b.Fatalf("Failed to prepare statement for table %s: %v", table, err)
		}
		stmtMap[table] = stmt
		defer stmt.Close()
	}

	for i := 0; i < numManga; i++ {
		uuid := fmt.Sprintf("uuid-%d", i)
		mangaList[i] = internal.Manga{Id: uuid}

		// Insert mappings for each table
		for _, table := range tables {
			_, err = stmtMap[table].Exec(uuid, fmt.Sprintf("id-%s-%d", table, i))
			if err != nil {
				b.Fatalf("Failed to insert data: %v", err)
			}
		}
	}

	// 3. Setup Output DB
	outputDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("Failed to open output memory db: %v", err)
	}
	defer outputDB.Close()

	// Create 'mappings' table
	// Schema inferred from insertNekoEntry: (mdex, al, ap, bw, mu, mu_new, nu, kt , mal)
	_, err = outputDB.Exec("CREATE TABLE " + internal.TableNekoMappings + " (mdex TEXT, al TEXT, ap TEXT, bw TEXT, mu TEXT, mu_new TEXT, nu TEXT, kt TEXT, mal TEXT)")
	if err != nil {
		b.Fatalf("Failed to create mappings table: %v", err)
	}

	// Load mappings for benchmark
	mappings := make(map[string]map[string]string)
	mappings[internal.TableAnilist] = getAllMappings(internal.TableAnilist)
	mappings[internal.TableAnimePlanet] = getAllMappings(internal.TableAnimePlanet)
	mappings[internal.TableBookWalker] = getAllMappings(internal.TableBookWalker)
	mappings[internal.TableKitsu] = getAllMappings(internal.TableKitsu)
	mappings[internal.TableMyanimelist] = getAllMappings(internal.TableMyanimelist)
	mappings[internal.TableMangaupdates] = getAllMappings(internal.TableMangaupdates)
	mappings[internal.TableMangaupdatesNewId] = getAllMappings(internal.TableMangaupdatesNewId)
	mappings[internal.TableNovelUpdates] = getAllMappings(internal.TableNovelUpdates)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, err := outputDB.Begin()
		if err != nil {
			b.Fatalf("Failed to begin transaction: %v", err)
		}
		processMangaList(tx, mangaList, mappings)
		tx.Rollback()
	}
}
