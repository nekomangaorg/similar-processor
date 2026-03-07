package internal

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestCheckErr(t *testing.T) {
	// Test nil error
	t.Run("nil error", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("CheckErr(nil) panicked: %v", r)
			}
		}()
		CheckErr(nil)
	})

	// Test non-nil error using a subprocess
	t.Run("non-nil error", func(t *testing.T) {
		if os.Getenv("BE_CRASHER") == "1" {
			CheckErr(errors.New("test error"))
			return
		}
		cmd := exec.Command(os.Args[0], "-test.run=^TestCheckErr/non-nil_error$")
		cmd.Env = append(os.Environ(), "BE_CRASHER=1")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()

		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			const expectedLog = "test error"
			if !strings.Contains(stderr.String(), expectedLog) {
				t.Errorf("expected stderr to contain %q, got %q", expectedLog, stderr.String())
			}
			return // Test passed
		}
		t.Fatalf("process ran with err %v, want exit status 1. stderr: %s", err, stderr.String())
	})
}

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
