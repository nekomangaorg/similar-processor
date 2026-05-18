package calculate

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/nekomangaorg/similar-processor/internal"
	"go.uber.org/ratelimit"
	"sync"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

var legacyIdRegex = regexp.MustCompile(`\d+`)

const muUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36"

var (
	muEntryExistsStmt *sql.Stmt
	upsertMuStmt      *sql.Stmt
	mappingStmtMu     sync.RWMutex
	muEntryExistsOnce sync.Once
	upsertMuOnce      sync.Once
)

func resetMappingStmts() {
	mappingStmtMu.Lock()
	defer mappingStmtMu.Unlock()
	if muEntryExistsStmt != nil {
		_ = muEntryExistsStmt.Close()
		muEntryExistsStmt = nil
	}
	if upsertMuStmt != nil {
		_ = upsertMuStmt.Close()
		upsertMuStmt = nil
	}
	muEntryExistsOnce = sync.Once{}
	upsertMuOnce = sync.Once{}
}

func muGet(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", muUserAgent)
	return httpClient.Do(req)
}

func muEntryExistsInNewIDDatabase(uuid string) (bool, error) {
	muEntryExistsOnce.Do(func() {
		mappingStmtMu.Lock()
		defer mappingStmtMu.Unlock()
		var err error
		muEntryExistsStmt, err = internal.DB.Prepare("SELECT 1 FROM " + internal.TableMangaupdatesNewId + " WHERE UUID= ?")
		if err != nil {
			fmt.Printf("ERROR: failed to prepare muEntryExistsInNewIDDatabase statement: %v\n", err)
		}
	})

	mappingStmtMu.RLock()
	stmt := muEntryExistsStmt
	mappingStmtMu.RUnlock()

	if stmt == nil {
		return false, fmt.Errorf("muEntryExistsInNewIDDatabase: statement not initialized")
	}

	var exists int
	err := stmt.QueryRow(uuid).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func upsertNewMuId(uuid string, id string) error {
	upsertMuOnce.Do(func() {
		mappingStmtMu.Lock()
		defer mappingStmtMu.Unlock()
		var err error
		upsertMuStmt, err = internal.DB.Prepare("INSERT INTO " + internal.TableMangaupdatesNewId + " (UUID, ID) VALUES (?, ?) ON CONFLICT (UUID) DO UPDATE SET ID=excluded.ID")
		if err != nil {
			fmt.Printf("ERROR: failed to prepare upsertNewMuId statement: %v\n", err)
		}
	})

	mappingStmtMu.RLock()
	stmt := upsertMuStmt
	mappingStmtMu.RUnlock()

	if stmt == nil {
		return fmt.Errorf("upsertNewMuId: statement not initialized")
	}

	_, err := stmt.Exec(uuid, id)
	return err
}

func UpsertGeneric(tx *sql.Tx, table string, uuid string, id string) error {
	if !internal.IsValidMappingTable(table) {
		return fmt.Errorf("UpsertGeneric: invalid table name %s", table)
	}
	_, err := tx.Exec("INSERT INTO "+table+" (UUID, ID) VALUES (?, ?) ON CONFLICT (UUID) DO UPDATE SET ID=excluded.ID", uuid, id)
	return err
}

func AddAlreadyConvertedId(ctx context.Context, index int, total int, uuid string, muLink string, rateLimiter ratelimit.Limiter) bool {
	if len(muLink) == 7 {
		// Encode from base36 format
		idEncoded := int64(internal.Decode(muLink))
		base10Id := strconv.FormatInt(idEncoded, 10)

		// Try the new id!
		rateLimiter.Take()
		resp2, err := muGet(ctx, "https://api.mangaupdates.com/v1/series/"+url.PathEscape(base10Id))
		if err != nil {
			fmt.Printf("\u001B[1;31m %s EXTERNAL MU: failed to get new id %s: %v\u001B[0m\n", uuid, base10Id, err)
			return false
		}
		defer drainAndClose(resp2)

		// Save if good!
		if resp2.StatusCode == 200 {
			fmt.Printf("%d/%d manga %s -> mu id %s encoded into %s -> is new MU id!\n", index+1, total, uuid, muLink, base10Id)
			if err := upsertNewMuId(uuid, base10Id); err != nil {
				fmt.Printf("\u001B[1;31m %s EXTERNAL MU: failed to save new id %s: %v\u001B[0m\n", uuid, base10Id, err)
			}
			return true
		}
	}
	return false
}

func CheckAndAddLegacyId(ctx context.Context, index int, total int, uuid string, muLink string, rateLimiter ratelimit.Limiter) bool {
	// For our ID conversion
	// https://www.unitconverters.net/numbers/base-36-to-decimal.htm

	ints := legacyIdRegex.FindAllString(muLink, -1)
	if len(ints) < 1 {
		return false
	}
	idOriginal, err := strconv.Atoi(ints[0])
	if err == nil {
		convertedId := strconv.Itoa(idOriginal)

		rateLimiter.Take()
		// Try the existing as the id (not likely since mangadex won't have updated..)
		resp1, err1 := muGet(ctx, "https://api.mangaupdates.com/v1/series/"+url.PathEscape(convertedId))

		if err1 == nil && resp1.StatusCode == 200 {
			drainAndClose(resp1)

			fmt.Printf("%d/%d manga %s -> mu id of %d -> is old MU id...\n", index+1, total, uuid, idOriginal)
			if err := upsertNewMuId(uuid, convertedId); err != nil {
				fmt.Printf("\u001B[1;31m %s EXTERNAL MU: failed to save legacy id %s: %v\u001B[0m\n", uuid, convertedId, err)
			}
			return true
		} else {
			if err1 != nil {
				fmt.Printf("\u001B[1;31m %s EXTERNAL MU: failed to get legacy id %s: %v\u001B[0m\n", uuid, convertedId, err1)
			}
			if resp1 != nil {
				drainAndClose(resp1)
			}

			// We have a couple retires here
			counterMax := 5
			for counter := 1; counter < counterMax; counter++ {
				rateLimiter.Take()

				// If invalid, then try to get the page and parse it!
				// Query and get our html... (no api to get this...)
				muUrl := "https://www.mangaupdates.com/series.html?id=" + url.QueryEscape(convertedId)
				resp, err := muGet(ctx, muUrl)

				// Sleep if we get a warning, otherwise we don't retry again!
				if err == nil && resp.StatusCode == 429 {
					fmt.Printf("\u001B[1;31m %s EXTERNAL MU: http code %d (try %d of %d)\u001B[0m\n", uuid, resp.StatusCode, counter, counterMax)

					drainAndClose(resp)

					select {
					case <-time.After(2 * time.Second):
					case <-ctx.Done():
						return false
					}
				} else if err == nil && resp.StatusCode != 200 {
					if resp.StatusCode == 503 {
						//this is a bad id on Dex's side write to debug file
						_ = WriteLineToDebugFile("BadMUIds", "https://mangadex.org/title/"+uuid)

						drainAndClose(resp)

						return false
					} else {
						fmt.Printf("\u001B[1;31m %s EXTERNAL MU %s: http code %d (try %d of %d)\u001B[0m\n", uuid, muUrl, resp.StatusCode, counter, counterMax)

						drainAndClose(resp)

						select {
						case <-time.After(2 * time.Second):
						case <-ctx.Done():
							return false
						}
					}

				} else if err == nil && resp.StatusCode == 200 {
					// Load the HTML document
					// Logic found using google chrome (right click in inspector and copy "selector")
					doc, err := goquery.NewDocumentFromReader(resp.Body)
					drainAndClose(resp)

					if err != nil {
						fmt.Printf("\u001B[1;31m %s EXTERNAL MU: failed to parse HTML for %s: %v\u001B[0m\n", uuid, muUrl, err)
						continue
					}

					rssUrl := doc.Find("#main_content > div:nth-child(2) > div.row.no-gutters > div.col-12.p-2 > a").AttrOr("href", "")
					paths := strings.Split(rssUrl, "/")
					if len(paths) > 3 {
						rssId := paths[len(paths)-2]
						fmt.Printf("%d/%d manga %s -> mu id of %d | RSS URL IS %s | %s id found\n", index+1, total, uuid, idOriginal, rssUrl, rssId)
						if err := upsertNewMuId(uuid, convertedId); err != nil {
							fmt.Printf("\u001B[1;31m %s EXTERNAL MU: failed to save scraped id %s: %v\u001B[0m\n", uuid, convertedId, err)
						}
						return true
					}
				} else {
					if err != nil {
						fmt.Printf("\u001B[1;31m %s EXTERNAL MU: request failed for %s (try %d of %d): %v\u001B[0m\n", uuid, muUrl, counter, counterMax, err)
					}
					// Catch all case to ensure body is closed
					if resp != nil {
						drainAndClose(resp)
					}
				}
			}
		}
	}
	return false

}

func drainAndClose(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}
