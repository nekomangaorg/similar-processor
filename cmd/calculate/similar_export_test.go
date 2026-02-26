package calculate

import (
	"database/sql"
	"fmt"
	"github.com/similar-manga/similar/internal"
	"os"
	"strings"
	"testing"
	_ "github.com/mattn/go-sqlite3"
)

func TestExportSimilar(t *testing.T) {
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

	_, err = db.Exec("CREATE TABLE SIMILAR (UUID TEXT PRIMARY KEY, JSON BLOB)")
	if err != nil {
		t.Fatal(err)
	}

	stmt, err := db.Prepare("INSERT INTO SIMILAR (UUID, JSON) VALUES (?, ?)")
	if err != nil {
		t.Fatal(err)
	}

	// Insert test data
	testData := []struct {
		UUID string
		JSON string
	}{
		{"12345", `{"id": "12345"}`},
		{"12346", `{"id": "12346"}`},
		{"12445", `{"id": "12445"}`}, // different suffix
		{"13345", `{"id": "13345"}`}, // different folder
	}

	for _, d := range testData {
		if _, err := stmt.Exec(d.UUID, d.JSON); err != nil {
			t.Fatal(err)
		}
	}
	stmt.Close()

	exportSimilar()

	// Verify files
	expectedFiles := map[string][]string{
		"data/similar/12/123.html": {"12345", "12346"},
		"data/similar/12/124.html": {"12445"},
		"data/similar/13/133.html": {"13345"},
	}

	for path, expectedUUIDs := range expectedFiles {
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("File %s not found: %v", path, err)
			continue
		}
		strContent := string(contentBytes)
		for _, uuid := range expectedUUIDs {
			expectedLinePart := uuid + ":::||@!@||:::"
			if !strings.Contains(strContent, expectedLinePart) {
				t.Errorf("File %s missing uuid %s", path, uuid)
			}
		}
	}

	// Check specifically that 123.html contains BOTH
	content, _ := os.ReadFile("data/similar/12/123.html")
	if !strings.Contains(string(content), "12345") || !strings.Contains(string(content), "12346") {
		t.Error("123.html should contain both 12345 and 12346")
	}

    // Cleanup
    os.RemoveAll("data/similar/")
}

func BenchmarkExportSimilar(b *testing.B) {
	// Setup DB
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	// Swap global DB
	originalDB := internal.DB
	internal.DB = db
	defer func() { internal.DB = originalDB }()

	_, err = db.Exec("CREATE TABLE SIMILAR (UUID TEXT PRIMARY KEY, JSON BLOB)")
	if err != nil {
		b.Fatal(err)
	}

	// Insert data
	stmt, err := db.Prepare("INSERT INTO SIMILAR (UUID, JSON) VALUES (?, ?)")
	if err != nil {
		b.Fatal(err)
	}

	// Generate 1000 items
	// 10 folders (00..09), 10 suffixes per folder (000..009), 10 items per suffix.
	// Total 1000 items.
	for f := 0; f < 10; f++ {
		for s := 0; s < 10; s++ {
			suffixStr := fmt.Sprintf("%02d%d", f, s) // e.g., "000", "001" ... "099"
			for i := 0; i < 10; i++ {
				uuid := fmt.Sprintf("%s%05d", suffixStr, i)
				json := fmt.Sprintf(`{"id": "%s", "data": "test"}`, uuid)
				_, err = stmt.Exec(uuid, json)
				if err != nil {
					b.Fatal(err)
				}
			}
		}
	}
	stmt.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exportSimilar()
	}
	b.StopTimer()
	os.RemoveAll("data/similar/")
}
