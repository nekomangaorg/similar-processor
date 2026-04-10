package mangadex

import (
	"encoding/json"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/similar-manga/similar/internal"
	md "github.com/similar-manga/similar/mangadex"
)

var testUUIDs []string

func init() {
	testUUIDs = make([]string, 1000)
	for i := 0; i < 1000; i++ {
		testUUIDs[i] = fmt.Sprintf("uuid-%d", i)
	}
}

func setupTestDB() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("CREATE TABLE " + internal.TableManga + " (UUID TEXT PRIMARY KEY, JSON TEXT, DATE TEXT)")
	if err != nil {
		panic(err)
	}

	stmt, err := db.Prepare("INSERT INTO " + internal.TableManga + " (UUID, JSON, DATE) VALUES (?, ?, ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	for _, uuid := range testUUIDs {
		_, err = stmt.Exec(uuid, "{}", "2023-01-01")
		if err != nil {
			panic(err)
		}
	}

	return db
}

func BenchmarkExistsInDatabase(b *testing.B) {
	db := setupTestDB()
	defer db.Close()
	internal.DB = db

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExistsInDatabase(testUUIDs[i%1000])
	}
}

func BenchmarkGetExistingMangaUUIDs(b *testing.B) {
	db := setupTestDB()
	defer db.Close()
	internal.DB = db

	batchSize := 100
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := (i * batchSize) % 1000
		end := start + batchSize
		if end > 1000 {
			end = 1000
		}
		GetExistingMangaUUIDs(testUUIDs[start:end])
	}
}

func TestExistsInDatabase(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	internal.DB = db

	// Test existing
	if !ExistsInDatabase("uuid-1") {
		t.Error("uuid-1 should exist")
	}

	// Test non-existing
	if ExistsInDatabase("uuid-9999") {
		t.Error("uuid-9999 should not exist")
	}
}

func TestGetExistingMangaUUIDs(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	internal.DB = db

	uuids := []string{"uuid-1", "uuid-2", "uuid-9999"}
	existing := GetExistingMangaUUIDs(uuids)

	if !existing["uuid-1"] {
		t.Error("uuid-1 should exist")
	}
	if !existing["uuid-2"] {
		t.Error("uuid-2 should exist")
	}
	if existing["uuid-9999"] {
		t.Error("uuid-9999 should not exist")
	}
	if len(existing) != 2 {
		t.Errorf("expected 2 existing uuids, got %d", len(existing))
	}
}

func TestGetExistingMangaUUIDs_Chunking(t *testing.T) {
	db := setupTestDB()
	defer db.Close()
	internal.DB = db

	// setupTestDB adds 1000 UUIDs (uuid-0 to uuid-999)
	// We pass 1100 UUIDs to test chunking (chunk size is 900)
	uuids := make([]string, 1100)
	for i := 0; i < 1100; i++ {
		uuids[i] = fmt.Sprintf("uuid-%d", i)
	}

	existing := GetExistingMangaUUIDs(uuids)

	// uuid-0 to uuid-999 should exist (1000)
	// uuid-1000 to uuid-1099 should not exist (100)
	if len(existing) != 1000 {
		t.Errorf("expected 1000 existing uuids, got %d", len(existing))
	}

	if !existing["uuid-0"] {
		t.Error("uuid-0 should exist")
	}
	if !existing["uuid-899"] {
		t.Error("uuid-899 should exist")
	}
	if !existing["uuid-900"] {
		t.Error("uuid-900 should exist (start of second chunk)")
	}
	if !existing["uuid-999"] {
		t.Error("uuid-999 should exist")
	}
	if existing["uuid-1000"] {
		t.Error("uuid-1000 should not exist")
	}
}

func TestApiMangaToJson(t *testing.T) {
	title := map[string]string{"en": "Test Title"}
	altTitles := []map[string]string{{"ja": "テストタイトル"}}
	description := map[string]string{"en": "Test Description"}
	links := map[string]string{"amz": "https://amazon.com"}
	tagName := map[string]string{"en": "Action"}

	apiManga := md.Manga{
		Id: "manga-uuid",
		Attributes: &md.MangaAttributes{
			Title:                        &title,
			AltTitles:                    altTitles,
			Description:                  &description,
			LastChapter:                  "100",
			AvailableTranslatedLanguages: []string{"en", "ja"},
			Links:                        links,
			OriginalLanguage:             "ja",
			PublicationDemographic:       "shounen",
			ContentRating:                "safe",
			Tags: []md.Tag{
				{
					Id: "tag-uuid",
					Attributes: &md.TagAttributes{
						Name: &tagName,
					},
				},
			},
		},
		Relationships: []md.Relationship{
			{
				Id:      "related-uuid-1",
				Related: "monochrome",
			},
			{
				Id:      "related-uuid-2",
				Related: "", // Should be filtered out
			},
		},
	}

	jsonBytes := ApiMangaToJson(apiManga)

	var result internal.Manga
	err := json.Unmarshal(jsonBytes, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result.Id != apiManga.Id {
		t.Errorf("Expected Id %s, got %s", apiManga.Id, result.Id)
	}
	if result.Title == nil || (*result.Title)["en"] != (*apiManga.Attributes.Title)["en"] {
		t.Errorf("Expected Title %v, got %v", apiManga.Attributes.Title, result.Title)
	}
	if len(result.AltTitles) != len(apiManga.Attributes.AltTitles) || result.AltTitles[0]["ja"] != apiManga.Attributes.AltTitles[0]["ja"] {
		t.Errorf("Expected AltTitles %v, got %v", apiManga.Attributes.AltTitles, result.AltTitles)
	}
	if result.Description == nil || (*result.Description)["en"] != (*apiManga.Attributes.Description)["en"] {
		t.Errorf("Expected Description %v, got %v", apiManga.Attributes.Description, result.Description)
	}
	if result.LastChapter != apiManga.Attributes.LastChapter {
		t.Errorf("Expected LastChapter %s, got %s", apiManga.Attributes.LastChapter, result.LastChapter)
	}
	if len(result.AvailableTranslatedLanguages) != len(apiManga.Attributes.AvailableTranslatedLanguages) || result.AvailableTranslatedLanguages[0] != apiManga.Attributes.AvailableTranslatedLanguages[0] {
		t.Errorf("Expected AvailableTranslatedLanguages %v, got %v", apiManga.Attributes.AvailableTranslatedLanguages, result.AvailableTranslatedLanguages)
	}
	if result.Links["amz"] != apiManga.Attributes.Links["amz"] {
		t.Errorf("Expected Links %v, got %v", apiManga.Attributes.Links, result.Links)
	}
	if result.OriginalLanguage != apiManga.Attributes.OriginalLanguage {
		t.Errorf("Expected OriginalLanguage %s, got %s", apiManga.Attributes.OriginalLanguage, result.OriginalLanguage)
	}
	if result.PublicationDemographic != apiManga.Attributes.PublicationDemographic {
		t.Errorf("Expected PublicationDemographic %s, got %s", apiManga.Attributes.PublicationDemographic, result.PublicationDemographic)
	}
	if result.ContentRating != apiManga.Attributes.ContentRating {
		t.Errorf("Expected ContentRating %s, got %s", apiManga.Attributes.ContentRating, result.ContentRating)
	}

	// Tags
	if len(result.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(result.Tags))
	} else {
		if result.Tags[0].Id != apiManga.Attributes.Tags[0].Id {
			t.Errorf("Expected Tag Id %s, got %s", apiManga.Attributes.Tags[0].Id, result.Tags[0].Id)
		}
		if result.Tags[0].Name == nil || (*result.Tags[0].Name)["en"] != (*apiManga.Attributes.Tags[0].Attributes.Name)["en"] {
			t.Errorf("Expected Tag Name %v, got %v", apiManga.Attributes.Tags[0].Attributes.Name, result.Tags[0].Name)
		}
	}

	// RelatedIds
	if len(result.RelatedIds) != 1 {
		t.Errorf("Expected 1 related ID, got %d", len(result.RelatedIds))
	} else if result.RelatedIds[0] != "related-uuid-1" {
		t.Errorf("Expected related ID related-uuid-1, got %s", result.RelatedIds[0])
	}
}
