package calculate

import (
	"database/sql"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/similar-manga/similar/internal"
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
		ResetSimilarInsertStmt()
	}()

	_, err = db.Exec("CREATE TABLE " + internal.TableSimilar + " (UUID TEXT PRIMARY KEY, JSON BLOB)")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	var counter uint64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := atomic.AddUint64(&counter, 1)
			InsertSimilarData(internal.SimilarManga{
				Id: fmt.Sprintf("uuid-%d", id),
				Title: map[string]string{"en":"Test"},
				UpdatedAt: time.Now().UTC().Format(time.RFC3339),
			})
		}
	})
}
