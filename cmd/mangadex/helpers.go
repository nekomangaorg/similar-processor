package mangadex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/similar-manga/similar/internal"
	"github.com/similar-manga/similar/mangadex"
	"github.com/similar-manga/similar/similar"
	"go.uber.org/ratelimit"
	"net/http"
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

	jsonManga, _ := json.MarshalIndent(manga, "", " ")
	return jsonManga
}

func CreateMangaDexClient() *mangadex.APIClient {
	config := mangadex.NewConfiguration()
	config.UserAgent = "calculate-manga v3.0"
	config.HTTPClient = &http.Client{
		Timeout: 60 * time.Second,
	}
	return mangadex.NewAPIClient(config)
}

func SearchMangaDex(rateLimiter ratelimit.Limiter, client *mangadex.APIClient, ctx context.Context, opts mangadex.MangaApiGetSearchMangaOpts2) mangadex.MangaList {
	maxRetries := 10
	mangaList := mangadex.MangaList{}
	resp := &http.Response{}
	err := errors.New("startup")

	for retryCount := 0; retryCount <= maxRetries && err != nil; retryCount++ {
		rateLimiter.Take()
		mangaList, resp, err = client.MangaApi.GetSearchManga2(ctx, &opts)
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
	rows, err := internal.MangaDB.Query("SELECT UUID FROM MANGA WHERE UUID= '" + uuid + "'")
	defer rows.Close()
	internal.CheckErr(err)
	return rows.Next()
}

func InsertManga(apiManga mangadex.Manga) {
	fmt.Printf("Inserting manga with ID: %s\n", apiManga.Id)
	jsonManga := ApiMangaToJson(apiManga)
	stmt, _ := internal.MangaDB.Prepare("INSERT INTO manga (uuid, manga_json) VALUES (?, ?)")
	defer stmt.Close()
	stmt.Exec(apiManga.Id, jsonManga)
}

func UpdateManga(apiManga mangadex.Manga) {
	jsonManga := ApiMangaToJson(apiManga)
	stmt, err := internal.MangaDB.Prepare("UPDATE MANGA set UUID = ?, MANGA_JSON = ? where UUID = ?")
	defer stmt.Close()
	internal.CheckErr(err)
	_, err = stmt.Exec(apiManga.Id, jsonManga, apiManga.Id)
	internal.CheckErr(err)
}

func UpdateMangaSimilarData(similarData similar.SimilarManga) {
	jsonSimilar, _ := json.MarshalIndent(similarData, "", " ")
	stmt, err := internal.MangaDB.Prepare("UPDATE MANGA set  SIMILAR_JSON = ? where UUID = ?")
	defer stmt.Close()
	internal.CheckErr(err)
	_, err = stmt.Exec(jsonSimilar, similarData.Id)
	internal.CheckErr(err)
}

func UpsertManga(apiManga mangadex.Manga) {
	if !ExistsInDatabase(apiManga.Id) {
		InsertManga(apiManga)
	} else {
		UpdateManga(apiManga)
	}
}
