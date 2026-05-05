package calculate

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"iter"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/nekomangaorg/similar-processor/internal"
)

func DeleteSimilarDB() error {
	_, err := internal.DB.Exec("DELETE FROM " + internal.TableSimilar)
	return err
}

var (
	similarInsertStmt *sql.Stmt
	similarInsertOnce sync.Once
)

// initSimilarInsertStmt initializes the prepared statement for InsertSimilarData.
// It is exposed for testing purposes to allow resetting the global state.
func initSimilarInsertStmt() {
	var err error
	similarInsertStmt, err = internal.DB.Prepare("INSERT INTO " + internal.TableSimilar + " (UUID, JSON) VALUES (?, ?)")
	internal.CheckErr(err)
}

// resetSimilarInsertStmt closes the prepared statement and resets the once flag.
// This is strictly used for testing and benching.
func resetSimilarInsertStmt() {
	similarInsertOnce = sync.Once{}
	if similarInsertStmt != nil {
		similarInsertStmt.Close()
		similarInsertStmt = nil
	}
}

func InsertSimilarData(similarData internal.SimilarManga) {
	similarInsertOnce.Do(initSimilarInsertStmt)

	jsonSimilar, err := json.Marshal(similarData)
	internal.CheckErr(err)

	_, err = similarInsertStmt.Exec(similarData.Id, jsonSimilar)
	internal.CheckErr(err)
}

// getDBSimilar streams similar manga entries from the database using an iterator.
// Overclock optimization: Previously, this loaded the entire 'SIMILAR' table into memory
// as a slice before returning, causing an O(n) memory spike and large GC pressure on large datasets.
// By returning an iter.Seq, we turn the batch load into an O(1) streaming pipeline, allowing
// the caller to process and export records as they are yielded directly from the database connection.
func getDBSimilar() iter.Seq[internal.DbSimilar] {
	return func(yield func(internal.DbSimilar) bool) {
		rows, err := internal.DB.Query("SELECT UUID, JSON FROM SIMILAR ORDER BY UUID ASC")
		if err != nil {
			log.Printf("ERROR: failed to query similar: %v", err)
			return
		}
		defer rows.Close() // Safely clean up the database rows iterator

		for rows.Next() {
			similar := internal.DbSimilar{}
			if err := rows.Scan(&similar.Id, &similar.JSON); err != nil {
				log.Printf("ERROR: failed to scan similar row: %v", err)
				continue
			}

			// Yield the row to the caller's loop. If the caller breaks out early,
			// yield returns false and we safely terminate, triggering the defer rows.Close()
			if !yield(similar) {
				return
			}
		}
		if err := rows.Err(); err != nil {
			log.Printf("ERROR: error iterating similar rows: %v", err)
		}
	}
}

var (
	debugFiles sync.Map
)

type debugWriter struct {
	file *os.File
	mu   sync.Mutex
	buf  *bufio.Writer
}

func WriteLineToDebugFile(fileName string, line string) error {
	actualName := filepath.Base(fileName)
	dw, ok := debugFiles.Load(actualName)
	if !ok {
		if err := os.MkdirAll("debug", 0700); err != nil {
			return err
		}
		file, err := os.OpenFile(filepath.Join("debug", actualName+".txt"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		dw = &debugWriter{
			file: file,
			buf:  bufio.NewWriter(file),
		}
		actual, loaded := debugFiles.LoadOrStore(actualName, dw)
		if loaded {
			file.Close()
			dw = actual
		}
	}

	d := dw.(*debugWriter)
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, err := d.buf.WriteString(line); err != nil {
		return err
	}
	return d.buf.WriteByte('\n')
}

// CloseDebugFiles flushes and closes all open debug file handles.
func CloseDebugFiles() {
	debugFiles.Range(func(key, value any) bool {
		d := value.(*debugWriter)
		d.mu.Lock()
		defer d.mu.Unlock()
		_ = d.buf.Flush()
		_ = d.file.Close()
		debugFiles.Delete(key)
		return true
	})
}

func exportMapping(tableName string, fileName string) {
	genericList, err := getAllGenericFromTable(tableName)
	internal.CheckErr(err)
	exportGeneric(fileName, genericList)
}

func exportGeneric(fileName string, genericList iter.Seq[internal.DbGeneric]) {
	file, err := os.Create("data/mappings/" + fileName + ".txt")
	internal.CheckErr(err)
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	for entry := range genericList {
		_, _ = writer.WriteString(entry.ID + ":::||@!@||:::" + entry.UUID + "\n")
	}
}

func getAllGenericFromTable(tableName string) (iter.Seq[internal.DbGeneric], error) {
	switch tableName {
	case internal.TableAnilist, internal.TableAnimePlanet, internal.TableBookWalker, internal.TableKitsu, internal.TableMyanimelist, internal.TableMangaupdates, internal.TableMangaupdatesNewId, internal.TableNovelUpdates:
		// OK
	default:
		return nil, fmt.Errorf("getAllGenericFromTable: invalid table name %s", tableName)
	}

	rows, err := internal.DB.Query("SELECT UUID, ID FROM " + tableName + " ORDER BY UUID asc ")
	if err != nil {
		return nil, err
	}

	return func(yield func(internal.DbGeneric) bool) {
		defer rows.Close()

		for rows.Next() {
			generic := internal.DbGeneric{}
			err = rows.Scan(&generic.UUID, &generic.ID)
			if err != nil {
				log.Printf("ERROR: failed to scan generic row: %v", err)
				continue
			}
			if !yield(generic) {
				return
			}
		}
		if err := rows.Err(); err != nil {
			log.Printf("ERROR: error iterating generic rows: %v", err)
		}
	}, nil
}

func CreateMappingsFile(fileName string) *os.File {
	file, err := os.Create("data/mappings/" + fileName + ".txt")
	if err != nil {
		log.Fatal(err)
	}
	return file
}
