package calculate

import (
	"database/sql"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nekomangaorg/similar-processor/internal"
)

func BenchmarkInsertSimilarData(b *testing.B) {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	originalDB := internal.DB
	internal.DB = db
	defer func() {
		internal.DB = originalDB
		resetSimilarInsertStmt()
	}()

	_, err = db.Exec("CREATE TABLE " + internal.TableSimilar + " (UUID TEXT PRIMARY KEY, JSON BLOB)")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	now := time.Now().UTC().Format(time.RFC3339)
	var counter uint64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := atomic.AddUint64(&counter, 1)
			InsertSimilarData(internal.SimilarManga{
				Id:        fmt.Sprintf("uuid-%d", id),
				Title:     map[string]string{"en": "Test"},
				UpdatedAt: now,
			})
		}
	})
}

func BenchmarkBatchInsertSimilarData(b *testing.B) {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	originalDB := internal.DB
	internal.DB = db
	defer func() {
		internal.DB = originalDB
	}()

	_, err = db.Exec("CREATE TABLE " + internal.TableSimilar + " (UUID TEXT PRIMARY KEY, JSON BLOB)")
	if err != nil {
		b.Fatal(err)
	}

	now := time.Now().UTC().Format(time.RFC3339)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		_, err = db.Exec("DELETE FROM " + internal.TableSimilar)
		if err != nil {
			b.Fatal(err)
		}

		results := make(chan internal.SimilarManga, 1000)
		go func() {
			for j := 0; j < 1000; j++ {
				results <- internal.SimilarManga{
					Id:        fmt.Sprintf("uuid-%d-%d", i, j),
					Title:     map[string]string{"en": "Test"},
					UpdatedAt: now,
				}
			}
			close(results)
		}()

		b.StartTimer()
		BatchInsertSimilarData(results)
	}
}
