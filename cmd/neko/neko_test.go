package neko

import (
	"database/sql"
	"testing"

	"github.com/similar-manga/similar/internal"
	_ "github.com/mattn/go-sqlite3"
)

func TestPopulateField(t *testing.T) {
	// Setup in-memory DB
	var err error
	internal.DB, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open memory db: %v", err)
	}
	defer internal.DB.Close()

	// Create a test table
	tableName := "TEST_TABLE"
	_, err = internal.DB.Exec("CREATE TABLE " + tableName + " (UUID TEXT, ID TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	testUUID := "test-uuid-123"
	testID := "test-id-456"
	_, err = internal.DB.Exec("INSERT INTO "+tableName+" (UUID, ID) VALUES (?, ?)", testUUID, testID)
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// Test case 1: Record exists
	var targetField string
	populateField(tableName, testUUID, &targetField)
	if targetField != testID {
		t.Errorf("Expected %s, got %s", testID, targetField)
	}

	// Test case 2: Record does not exist
	targetField = ""
	populateField(tableName, "non-existent-uuid", &targetField)
	if targetField != "" {
		t.Errorf("Expected empty string, got %s", targetField)
	}
}
