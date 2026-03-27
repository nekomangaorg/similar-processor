package calculate

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/similar-manga/similar/internal"
)

func TestGetAllGenericFromTable(t *testing.T) {
	// Setup DB
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Swap global DB
	originalDB := internal.DB
	internal.DB = db
	defer func() { internal.DB = originalDB }()

	// Create an allowed table
	_, err = db.Exec("CREATE TABLE " + internal.TableAnilist + " (UUID TEXT PRIMARY KEY, ID TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	// Insert test data
	testData := []internal.DbGeneric{
		{UUID: "uuid1", ID: "id1"},
		{UUID: "uuid2", ID: "id2"},
	}

	stmt, err := db.Prepare("INSERT INTO " + internal.TableAnilist + " (UUID, ID) VALUES (?, ?)")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	for _, d := range testData {
		if _, err := stmt.Exec(d.UUID, d.ID); err != nil {
			t.Fatal(err)
		}
	}

	// Test happy path
	result := getAllGenericFromTable(internal.TableAnilist)

	if !reflect.DeepEqual(result, testData) {
		t.Errorf("expected %v, got %v", testData, result)
	}
}

func TestGetAllGenericFromTable_InvalidTable(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		getAllGenericFromTable("INVALID_TABLE")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestGetAllGenericFromTable_InvalidTable$")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		expectedLog := "getAllGenericFromTable: invalid table name INVALID_TABLE"
		if !strings.Contains(stderr.String(), expectedLog) {
			t.Errorf("expected stderr to contain %q, got %q", expectedLog, stderr.String())
		}
		return // Test passed
	}
	t.Fatalf("process ran with err %v, want exit status 1. stderr: %s", err, stderr.String())
}
