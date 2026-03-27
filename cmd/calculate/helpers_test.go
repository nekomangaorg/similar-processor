package calculate

import (
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/similar-manga/similar/internal"
)

func setupTestDB(t *testing.T) (db *sql.DB, teardown func()) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	originalDB := internal.DB
	internal.DB = db

	teardown = func() {
		internal.DB = originalDB
		db.Close()
	}

	return db, teardown
}

func TestGetAllGenericFromTable(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	// Create an allowed table
	_, err := db.Exec("CREATE TABLE " + internal.TableAnilist + " (UUID TEXT PRIMARY KEY, ID TEXT)")
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
	result, err := getAllGenericFromTable(internal.TableAnilist)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(result, testData) {
		t.Errorf("expected %v, got %v", testData, result)
	}
}

func TestGetAllGenericFromTable_InvalidTable(t *testing.T) {
	_, err := getAllGenericFromTable("INVALID_TABLE")
	if err == nil {
		t.Fatal("expected an error for invalid table name, but got nil")
	}

	expectedError := "getAllGenericFromTable: invalid table name INVALID_TABLE"
	if err.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, err.Error())
	}
}
