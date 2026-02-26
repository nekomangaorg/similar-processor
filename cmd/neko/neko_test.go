package neko

import (
	"database/sql"
	"testing"

	"github.com/similar-manga/similar/internal"
	_ "github.com/mattn/go-sqlite3"
)

func TestExportNeko(t *testing.T) {
	// 1. Setup Internal DB
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open internal memory db: %v", err)
	}
	internal.DB = db
	defer db.Close()

	// Create tables
	_, err = db.Exec("CREATE TABLE " + internal.TableManga + " (UUID TEXT, JSON TEXT, DATE TEXT)")
	if err != nil {
		t.Fatalf("Failed to create MANGA table: %v", err)
	}

	for _, table := range mappingTables {
		_, err = db.Exec("CREATE TABLE " + table + " (UUID TEXT, ID TEXT)")
		if err != nil {
			t.Fatalf("Failed to create table %s: %v", table, err)
		}
	}

	// Insert data
	_, err = db.Exec("INSERT INTO " + internal.TableManga + " (UUID, JSON, DATE) VALUES ('uuid-1', '{}', '2023-01-01')")
	if err != nil {
		t.Fatalf("Failed to insert manga: %v", err)
	}
	_, err = db.Exec("INSERT INTO " + internal.TableAnilist + " (UUID, ID) VALUES ('uuid-1', 'al-1')")
	if err != nil {
		t.Fatalf("Failed to insert mapping: %v", err)
	}

	// 2. Setup Output DB
	outputDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open output memory db: %v", err)
	}
	defer outputDB.Close()

	_, err = outputDB.Exec("CREATE TABLE " + internal.TableNekoMappings + " (mdex TEXT, al TEXT, ap TEXT, bw TEXT, mu TEXT, mu_new TEXT, nu TEXT, kt TEXT, mal TEXT)")
	if err != nil {
		t.Fatalf("Failed to create mappings table: %v", err)
	}

	// 3. Run exportNeko
	tx, err := outputDB.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	exportNeko(tx)

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// 4. Verify
	rows, err := outputDB.Query("SELECT mdex, al FROM "+internal.TableNekoMappings+" WHERE mdex = 'uuid-1'")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	if rows.Next() {
		var mdex string
		var al sql.NullString // Expecting NullString or string depending on driver/scan
		// Wait, scan target must match type or implement Scanner.
		// In test verification, I can scan into string if it's not null.

		if err := rows.Scan(&mdex, &al); err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		if mdex != "uuid-1" {
			t.Errorf("Expected uuid-1, got %s", mdex)
		}
		if al.String != "al-1" {
			t.Errorf("Expected al-1, got %s", al.String)
		}
	} else {
		t.Errorf("No rows found")
	}
}
