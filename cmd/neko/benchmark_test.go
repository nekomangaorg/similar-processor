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
		stmt, _ := internal.DB.Prepare("INSERT INTO " + table + " (UUID, ID) VALUES (?, ?)")
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
	anilistMap := getAllMappings(internal.TableAnilist)
	animePlanetMap := getAllMappings(internal.TableAnimePlanet)
	bookwalkerMap := getAllMappings(internal.TableBookWalker)
	kitsuMap := getAllMappings(internal.TableKitsu)
	malMap := getAllMappings(internal.TableMyanimelist)
	mangaupdatesMap := getAllMappings(internal.TableMangaupdates)
	mangaupdatesNewMap := getAllMappings(internal.TableMangaupdatesNewId)
	novelUpdatesMap := getAllMappings(internal.TableNovelUpdates)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx, _ := outputDB.Begin()
		processMangaList(tx, mangaList, anilistMap, animePlanetMap, bookwalkerMap, kitsuMap, malMap, mangaupdatesMap, mangaupdatesNewMap, novelUpdatesMap)
		tx.Rollback()
	}
}
