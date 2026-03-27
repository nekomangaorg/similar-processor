package neko

import (
	"bytes"
	"database/sql"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/similar-manga/similar/internal"
)

func TestGetAllMappings(t *testing.T) {
	t.Run("invalid table name", func(t *testing.T) {
		if os.Getenv("BE_CRASHER") == "1" {
			getAllMappings("INVALID_TABLE")
			return
		}
		cmd := exec.Command(os.Args[0], "-test.run=^TestGetAllMappings/invalid_table_name$")
		cmd.Env = append(os.Environ(), "BE_CRASHER=1")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()

		if e, ok := err.(*exec.ExitError); ok && !e.Success() {
			const expectedLog = "getAllMappings: invalid table name"
			if !strings.Contains(stderr.String(), expectedLog) {
				t.Errorf("expected stderr to contain %q, got %q", expectedLog, stderr.String())
			}
			return
		}
		t.Fatalf("process ran with err %v, want exit status 1. stderr: %s", err, stderr.String())
	})
}

func TestProcessMangaList(t *testing.T) {
	// Setup Output DB
	outputDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open output memory db: %v", err)
	}
	defer outputDB.Close()

	// Create 'mappings' table
	_, err = outputDB.Exec("CREATE TABLE " + internal.TableNekoMappings + " (mdex TEXT, al TEXT, ap TEXT, bw TEXT, mu TEXT, mu_new TEXT, nu TEXT, kt TEXT, mal TEXT)")
	if err != nil {
		t.Fatalf("Failed to create mappings table: %v", err)
	}

	// Setup data
	mangaList := []internal.Manga{
		{Id: "uuid-1"},
	}

	mappings := make(map[string]map[string]string)
	mappings[internal.TableAnilist] = map[string]string{"uuid-1": "al-1"}

	tx, err := outputDB.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	// Pass nil for other maps
	processMangaList(tx, slices.Values(mangaList), mappings)
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Verify
	rows, err := outputDB.Query("SELECT mdex, al FROM " + internal.TableNekoMappings + " WHERE mdex = 'uuid-1'")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	if rows.Next() {
		var mdex, al string
		rows.Scan(&mdex, &al)
		if mdex != "uuid-1" {
			t.Errorf("Expected uuid-1, got %s", mdex)
		}
		if al != "al-1" {
			t.Errorf("Expected al-1, got %s", al)
		}
	} else {
		t.Errorf("No rows found")
	}
}
