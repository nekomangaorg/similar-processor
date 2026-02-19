package mangadex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/similar-manga/similar/internal"
	"github.com/similar-manga/similar/mangadex"
	"go.uber.org/ratelimit"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func ApiMangaToJson(apiManga mangadex.Manga) []byte {
	tags := make([]internal.Tag, 0, len(apiManga.Attributes.Tags))
	for _, r := range apiManga.Attributes.Tags {
		tags = append(tags, internal.Tag{
			Id:   r.Id,
			Name: r.Attributes.Name,
		})
	}
	var relatedIds []string
	for _, r := range apiManga.Relationships {
		if r.Related != "" {
			relatedIds = append(relatedIds, r.Id)
		}
	}

	manga := internal.Manga{
		Id:                           apiManga.Id,
		Title:                        apiManga.Attributes.Title,
		AltTitles:                    apiManga.Attributes.AltTitles,
		Description:                  apiManga.Attributes.Description,
		LastChapter:                  apiManga.Attributes.LastChapter,
		AvailableTranslatedLanguages: apiManga.Attributes.AvailableTranslatedLanguages,
		RelatedIds:                   relatedIds,
		Links:                        apiManga.Attributes.Links,
		OriginalLanguage:             apiManga.Attributes.OriginalLanguage,
		PublicationDemographic:       apiManga.Attributes.PublicationDemographic,
		ContentRating:                apiManga.Attributes.ContentRating,
		Tags:                         tags,
	}

	dst := &bytes.Buffer{}
	jsonManga, _ := json.Marshal(manga)
	err := json.Compact(dst, jsonManga)
	internal.CheckErr(err)
	return dst.Bytes()
}

func CreateMangaDexClient() *mangadex.APIClient {
	config := mangadex.NewConfiguration()
	config.UserAgent = "similar-manga v3.0"
	config.HTTPClient = &http.Client{
		Timeout: 30 * time.Second,
	}
	return mangadex.NewAPIClient(config)
}

func SearchMangaDex(rateLimiter ratelimit.Limiter, client *mangadex.APIClient, ctx context.Context, opts mangadex.MangaApiGetSearchMangaOpts) mangadex.MangaList {
	maxRetries := 10
	mangaList := mangadex.MangaList{}
	resp := &http.Response{}
	err := errors.New("startup")

	for retryCount := 0; retryCount <= maxRetries && err != nil; retryCount++ {
		rateLimiter.Take()
		mangaList, resp, err = client.MangaApi.GetSearchManga(ctx, &opts)
		if err != nil {
			fmt.Printf("\u001B[1;31mMANGA ERROR (%d of %d): Status Code %v : %v\u001B[0m\n", retryCount, maxRetries, resp.StatusCode, err)
			if err.Error() == "undefined response type text/html; charset=utf-8" {
				fmt.Println("Sleeping 5 secs since we likely hit the soft rate limit")
				time.Sleep(5 * time.Second)
			}
		} else if resp == nil {
			err = errors.New("invalid response object")
			fmt.Printf("\u001B[1;31mMANGA ERROR (%d of %d): respose object is nil\u001B[0m\n", retryCount, maxRetries)
			continue
		} else if resp.StatusCode != 200 && resp.StatusCode != 204 {
			err = errors.New("invalid http error code")
			fmt.Printf("\u001B[1;31mMANGA ERROR (%d of %d): http code %d\u001B[0m\n", retryCount, maxRetries, resp.StatusCode)
		} else if resp.StatusCode == 429 {
			err = errors.New("rate limited")
			fmt.Printf("\u001B[1;31mRate Limited!! Sleeping. (%d of %d): http code %d\u001B[0m\n", retryCount, maxRetries, resp.StatusCode)
			time.Sleep(time.Duration(int64(500)) * time.Millisecond)

		}

		if err == nil {
			//ignore the error if it fails to close
			_ = resp.Body.Close()
		}
	}
	return mangaList

}

func ExistsInDatabase(uuid string) bool {
	rows, err := internal.DB.Query("SELECT UUID FROM "+internal.TableManga+" WHERE UUID= ?", uuid)
	internal.CheckErr(err)
	defer rows.Close()
	return rows.Next()
}

func UpsertManga(apiManga mangadex.Manga) {
	jsonManga := ApiMangaToJson(apiManga)
	currentDate := strings.Split(time.Now().UTC().Format(time.RFC3339), "Z")[0]
	_, err := internal.DB.Exec("INSERT INTO "+internal.TableManga+" (UUID, JSON, DATE) VALUES (?, ?, ?) ON CONFLICT (UUID) DO UPDATE SET JSON=excluded.JSON", apiManga.Id, jsonManga, currentDate)
	internal.CheckErr(err)
}

func getDBManga() []internal.DbManga {
	rows, err := internal.DB.Query("SELECT UUID, JSON, DATE FROM " + internal.TableManga + " ORDER BY DATE ASC")
	defer rows.Close()
	internal.CheckErr(err)

	var mangaList []internal.DbManga
	for rows.Next() {
		manga := internal.DbManga{}
		rows.Scan(&manga.Id, &manga.JSON, &manga.DATE)
		internal.CheckErr(err)
		mangaList = append(mangaList, manga)
	}
	return mangaList
}

func ExportManga() {
	fmt.Printf("Exporting All Manga to txt files\n")
	os.RemoveAll("data/manga/")
	os.MkdirAll("data/manga/", 0777)
	mangaList := getDBManga()
	suffix := 1
	file := createMangaFile(suffix)
	for index, manga := range mangaList {
		if index > 0 && index%1000 == 0 {
			suffix++
			file.Close()
			file = createMangaFile(suffix)
		}

		file.WriteString(manga.Id + ":::||@!@||:::" + manga.DATE + ":::||@!@||:::" + manga.JSON + "\n")

	}

	file.Close()

}

func createMangaFile(number int) *os.File {
	file, err := os.Create("data/manga/manga_" + fmt.Sprintf("%04d", number) + ".txt")
	if err != nil {
		log.Fatal(err)
	}
	return file
}
