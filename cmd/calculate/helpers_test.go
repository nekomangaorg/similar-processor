package calculate

import (
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nekomangaorg/similar-processor/internal"
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
	resultSeq, err := getAllGenericFromTable(internal.TableAnilist)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []internal.DbGeneric
	for generic := range resultSeq {
		result = append(result, generic)
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

func TestDeleteSimilarDB(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	// Create SIMILAR table
	_, err := db.Exec("CREATE TABLE " + internal.TableSimilar + " (UUID TEXT PRIMARY KEY, JSON TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	// Insert multiple test data rows
	testRows := []struct {
		uuid string
		json string
	}{
		{"uuid1", "{}"},
		{"uuid2", "{}"},
		{"uuid3", "{}"},
	}
	for _, row := range testRows {
		_, err = db.Exec("INSERT INTO "+internal.TableSimilar+" (UUID, JSON) VALUES (?, ?)", row.uuid, row.json)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Verify all data is there
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM " + internal.TableSimilar).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != len(testRows) {
		t.Fatalf("expected %d rows, got %d", len(testRows), count)
	}

	// Call DeleteSimilarDB
	if err := DeleteSimilarDB(); err != nil {
		t.Fatalf("DeleteSimilarDB returned an unexpected error: %v", err)
	}

	// Assert table is empty
	err = db.QueryRow("SELECT COUNT(*) FROM " + internal.TableSimilar).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows after delete, got %d", count)
	}
}
