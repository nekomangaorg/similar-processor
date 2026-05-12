package calculate

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nekomangaorg/similar-processor/internal"
	"go.uber.org/ratelimit"
)

type mockTransport struct {
	RoundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func setupTestDBWithTable(t *testing.T) func() {
	t.Helper()
	_, teardown := setupTestDB(t)
	resetMappingStmts()

	// Ensure the table exists
	_, err := internal.DB.Exec("CREATE TABLE IF NOT EXISTS " + internal.TableMangaupdatesNewId + " (UUID TEXT PRIMARY KEY, ID TEXT)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	return teardown
}

func TestCheckAndAddLegacyId(t *testing.T) {
	teardown := setupTestDBWithTable(t)
	defer teardown()

	// Save original client and restore after tests
	oldClient := httpClient
	defer func() { httpClient = oldClient }()

	tests := []struct {
		name           string
		uuid           string
		muLink         string
		setupDB        func()
		mockResp       func(req *http.Request) (*http.Response, error)
		expectedResult bool
		verify         func(t *testing.T)
	}{
		{
			name:   "Invalid muLink (no digits)",
			uuid:   "uuid1",
			muLink: "no-digits-here",
			mockResp: func(req *http.Request) (*http.Response, error) {
				return nil, nil // Should not be called
			},
			expectedResult: false,
		},
		{
			name:   "Entry already exists in DB",
			uuid:   "uuid2",
			muLink: "https://www.mangaupdates.com/series.html?id=12345",
			setupDB: func() {
				_, _ = internal.DB.Exec("INSERT INTO "+internal.TableMangaupdatesNewId+" (UUID, ID) VALUES (?, ?)", "uuid2", "12345")
			},
			mockResp: func(req *http.Request) (*http.Response, error) {
				return nil, nil // Should not be called
			},
			expectedResult: true,
		},
		{
			name:   "API returns 200 (Success)",
			uuid:   "uuid3",
			muLink: "https://www.mangaupdates.com/series.html?id=67890",
			mockResp: func(req *http.Request) (*http.Response, error) {
				if strings.Contains(req.URL.Path, "/v1/series/67890") {
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
					}, nil
				}
				return &http.Response{StatusCode: 404, Body: io.NopCloser(bytes.NewBufferString(""))}, nil
			},
			expectedResult: true,
			verify: func(t *testing.T) {
				var id string
				err := internal.DB.QueryRow("SELECT ID FROM "+internal.TableMangaupdatesNewId+" WHERE UUID = ?", "uuid3").Scan(&id)
				if err != nil {
					t.Errorf("failed to find entry in DB: %v", err)
				}
				if id != "67890" {
					t.Errorf("expected ID 67890, got %s", id)
				}
			},
		},
		{
			name:   "API fails, Web scraping succeeds",
			uuid:   "uuid4",
			muLink: "https://www.mangaupdates.com/series.html?id=11111",
			mockResp: func(req *http.Request) (*http.Response, error) {
				if strings.Contains(req.URL.Path, "/v1/series/11111") {
					return &http.Response{
						StatusCode: 404,
						Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
					}, nil
				}
				if strings.Contains(req.URL.Path, "series.html") && req.URL.Query().Get("id") == "11111" {
					html := `<html><body><div id="main_content"><div></div><div><div class="row no-gutters"><div class="col-12 p-2"><a href="https://api.mangaupdates.com/v1/series/22222/rss">RSS</a></div></div></div></div></body></html>`
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewBufferString(html)),
					}, nil
				}
				return &http.Response{StatusCode: 404, Body: io.NopCloser(bytes.NewBufferString(""))}, nil
			},
			expectedResult: true,
			verify: func(t *testing.T) {
				var id string
				err := internal.DB.QueryRow("SELECT ID FROM "+internal.TableMangaupdatesNewId+" WHERE UUID = ?", "uuid4").Scan(&id)
				if err != nil {
					t.Errorf("failed to find entry in DB: %v", err)
				}
				if id != "11111" {
					t.Errorf("expected ID 11111, got %s", id)
				}
			},
		},
		{
			name:   "API returns 503 (Bad ID)",
			uuid:   "uuid5",
			muLink: "https://www.mangaupdates.com/series.html?id=55555",
			mockResp: func(req *http.Request) (*http.Response, error) {
				if strings.Contains(req.URL.Path, "/v1/series/55555") {
					return &http.Response{
						StatusCode: 404,
						Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
					}, nil
				}
				if strings.Contains(req.URL.Path, "series.html") && req.URL.Query().Get("id") == "55555" {
					return &http.Response{
						StatusCode: 503,
						Body:       io.NopCloser(bytes.NewBufferString("Service Unavailable")),
					}, nil
				}
				return &http.Response{StatusCode: 404, Body: io.NopCloser(bytes.NewBufferString(""))}, nil
			},
			expectedResult: false,
			verify: func(t *testing.T) {
				CloseDebugFiles()
				debugFile := filepath.Join("debug", "BadMUIds.txt")
				t.Cleanup(func() { _ = os.Remove(debugFile) })

				content, err := os.ReadFile(debugFile)
				if err != nil {
					t.Errorf("expected debug file %s to exist and be readable: %v", debugFile, err)
					return
				}
				if !strings.Contains(string(content), "https://mangadex.org/title/uuid5") {
					t.Errorf("debug file does not contain expected link")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear table for each test
			_, _ = internal.DB.Exec("DELETE FROM " + internal.TableMangaupdatesNewId)
			if tt.setupDB != nil {
				tt.setupDB()
			}

			httpClient = &http.Client{
				Transport: &mockTransport{
					RoundTripFunc: tt.mockResp,
				},
			}

			result := CheckAndAddLegacyId(context.Background(), 0, 1, tt.uuid, tt.muLink, ratelimit.NewUnlimited())

			if result != tt.expectedResult {
				t.Errorf("expected result %v, got %v", tt.expectedResult, result)
			}

			if tt.verify != nil {
				tt.verify(t)
			}
		})
	}
}

func TestCheckAndAddLegacyId_Retry429(t *testing.T) {
	teardown := setupTestDBWithTable(t)
	defer teardown()

	oldClient := httpClient
	defer func() { httpClient = oldClient }()

	callCount := 0
	httpClient = &http.Client{
		Transport: &mockTransport{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				if strings.Contains(req.URL.Path, "/v1/series/99999") {
					return &http.Response{
						StatusCode: 404,
						Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
					}, nil
				}
				if strings.Contains(req.URL.Path, "series.html") {
					callCount++
					if callCount == 1 {
						return &http.Response{
							StatusCode: 429,
							Body:       io.NopCloser(bytes.NewBufferString("Too Many Requests")),
						}, nil
					}
					// Return success on second try to avoid long wait in tests
					html := `<html><body><div id="main_content"><div></div><div><div class="row no-gutters"><div class="col-12 p-2"><a href="https://api.mangaupdates.com/v1/series/99999/rss">RSS</a></div></div></div></div></body></html>`
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewBufferString(html)),
					}, nil
				}
				return &http.Response{StatusCode: 404, Body: io.NopCloser(bytes.NewBufferString(""))}, nil
			},
		},
	}

	result := CheckAndAddLegacyId(context.Background(), 0, 1, "uuid_retry", "https://www.mangaupdates.com/series.html?id=99999", ratelimit.NewUnlimited())

	if !result {
		t.Errorf("expected success after retry")
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 calls due to 429 retry, got %d", callCount)
	}
}
