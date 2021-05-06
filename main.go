package main

import (
	"./swagger"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/antihax/optional"
	"io"
	"io/ioutil"
	"log"
	_ "math"
	"net/http"
	"os"
	"time"
)

func checkLoginStatus(client *swagger.APIClient, ctx context.Context) swagger.CheckResponse {
	authResp, resp, err := client.AuthApi.GetAuthCheck(ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("%v", resp)
	}
	return authResp
}

func reportToMangadexNetwork(url string, filename string, start time.Time, success bool, cached bool) {

	// Create default
	values := make(map[string]interface{})
	values["url"] = url
	values["success"] = success
	values["bytes"] = 0
	values["duration"] = time.Since(start).Milliseconds()
	values["cached"] = cached

	// If failed directly report
	if !success {
		jsonValue, _ := json.Marshal(values)
		_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
		if err != nil {
			log.Fatalf("%v", err)
		}
		return
	}

	// If file does not exists then we have already failed
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		values["success"] = false
		jsonValue, _ := json.Marshal(values)
		_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
		if err != nil {
			log.Fatalf("%v", err)
		}
		return
	}

	// Finally report the downloaded image to mangadex @ home network report
	fi, _ := os.Stat(filename)
	values["bytes"] = fi.Size()
	jsonValue, _ := json.Marshal(values)
	//fmt.Println(string(jsonValue))
	_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Fatalf("%v", err)
	}

}

func main() {

	// Directory configuration
	fileSession := "data/session.json"
	//dirMangas := "data/manga/"
	dirChapters := "data/chapter/"
	dirImages := "data/images/"
	userUsername := ""
	userPassword := ""

	// Create client
	config := swagger.NewConfiguration()
	config.UserAgent = "similar-manga v2.0"
	client := swagger.NewAPIClient(config)
	ctx := context.Background()

	// Left first try to login
	token := swagger.LoginResponseToken{}
	if _, err := os.Stat(fileSession); err == nil {
		fileManga, _ := ioutil.ReadFile(fileSession)
		_ = json.Unmarshal([]byte(fileManga), &token)
		config.AddDefaultHeader("Authorization", "Bearer "+token.Session)
	}
	authResp := checkLoginStatus(client, ctx)
	fmt.Println(authResp)

	// On first ever start we will get our session!
	if !authResp.IsAuthenticated && len(token.Session) == 0 {
		fmt.Println("Performing first time login!")
		bodyData := map[string]string{
			"username": userUsername,
			"password": userPassword,
		}
		opts := swagger.AuthApiPostAuthLoginOpts{}
		opts.Body = optional.NewInterface(bodyData)
		authResp, resp, err := client.AuthApi.PostAuthLogin(ctx, &opts)
		if err != nil {
			log.Fatalf("%v\n%v", resp, err)
		}
		if resp.StatusCode != 200 {
			log.Fatalf("%v\n%v", resp, err)
		}
		file, _ := json.MarshalIndent(authResp.Token, "", " ")
		_ = ioutil.WriteFile(fileSession, file, 0644)
		token = *authResp.Token
		config.AddDefaultHeader("Authorization", "Bearer "+token.Session)
	} else if !authResp.IsAuthenticated {
		fmt.Println("Performing session refresh!")
		bodyData := map[string]string{
			"token": token.Refresh,
		}
		opts := swagger.AuthApiPostAuthRefreshOpts{}
		opts.Body = optional.NewInterface(bodyData)
		authResp, resp, err := client.AuthApi.PostAuthRefresh(ctx, &opts)
		if err != nil {
			log.Fatalf("%v\n%v", resp, err)
		}
		if resp.StatusCode != 200 {
			log.Fatalf("%v\n%v", resp, err)
		}
		file, _ := json.MarshalIndent(authResp.Token, "", " ")
		_ = ioutil.WriteFile(fileSession, file, 0644)
		token = *authResp.Token
		config.AddDefaultHeader("Authorization", "Bearer "+token.Session)
	}

	// Specify our max limit and loop through the entire API to get all manga
	//currentLimit := int32(100)
	//maxOffset := int32(100000)
	//for currentOffset := int32(0); currentOffset < maxOffset; currentOffset += currentLimit {
	//
	//	// Perform our api search call to get the response
	//	opts := swagger.MangaApiGetSearchMangaOpts{}
	//	opts.Limit = optional.NewInt32(currentLimit)
	//	opts.Offset = optional.NewInt32(currentOffset)
	//	mangaList, resp, err := client.MangaApi.GetSearchManga(ctx, &opts)
	//	if err != nil {
	//		log.Fatalf("%v", err)
	//	}
	//	if resp.StatusCode != 200 {
	//		fmt.Println("HTTP ERROR CODE %d", resp.StatusCode)
	//		break
	//	}
	//
	//	// Loop through all manga and print their ids
	//	for i, manga := range mangaList.Results {
	//		fmt.Printf("%d/%d -> %s\n", currentOffset+int32(i), maxOffset, manga.Data.Id)
	//		file, _ := json.MarshalIndent(manga.Data, "", " ")
	//		_ = ioutil.WriteFile(dirMangas+manga.Data.Id+".json", file, 0644)
	//	}
	//
	//	// Update our current limit
	//	maxOffset = mangaList.Total
	//	currentLimit = int32(math.Min(float64(currentLimit), float64(maxOffset-currentOffset)))
	//
	//}

	// Loop through all manga and try to get their chapter information for each
	//itemsManga, _ := ioutil.ReadDir(dirMangas)
	//for i, file := range itemsManga {
	//
	//	// Skip if a directory
	//	if file.IsDir() {
	//		continue
	//	}
	//
	//	// Load the json from file into our manga struct
	//	manga := swagger.Manga{}
	//	fmt.Println(dirMangas + file.Name())
	//	fileManga, _ := ioutil.ReadFile(dirMangas + file.Name())
	//	_ = json.Unmarshal([]byte(fileManga), &manga)
	//
	//	// Perform our api search call to get the response
	//	opts := swagger.MangaApiGetMangaIdFeedOpts{}
	//	opts.Limit = optional.NewInt32(500)
	//	chapterList, resp, err := client.MangaApi.GetMangaIdFeed(ctx, manga.Id, &opts)
	//	if resp != nil && resp.StatusCode == 404 {
	//		fmt.Printf("CHAPTER FEED GAVE %d (no chapter?!)\n", resp.StatusCode)
	//		continue
	//	}
	//	if err != nil {
	//		log.Fatalf("%v", err)
	//	}
	//	if resp.StatusCode != 200 {
	//		fmt.Printf("HTTP ERROR CODE %d\n", resp.StatusCode)
	//		continue
	//	}
	//
	//	// Loop through all chapter for this manga and save to disk
	//	for c, chapter := range chapterList.Results {
	//		fmt.Printf("%d/%d (chapter %d) -> %s\n", i, len(itemsManga), c+1, chapter.Data.Id)
	//		fileChapter, _ := json.MarshalIndent(chapter.Data, "", " ")
	//		_ = ioutil.WriteFile(dirChapters+chapter.Data.Id+".json", fileChapter, 0644)
	//	}
	//	fmt.Println()
	//
	//}

	// Loop through all manga and download each chapter's images!
	itemsChapters, _ := ioutil.ReadDir(dirChapters)
	for i, file := range itemsChapters {

		// Skip if a directory
		if file.IsDir() {
			continue
		}

		// Load the json from file into our chapter struct
		chapter := swagger.Chapter{}
		fmt.Println(dirChapters + file.Name())
		fileManga, _ := ioutil.ReadFile(dirChapters + file.Name())
		_ = json.Unmarshal([]byte(fileManga), &chapter)

		// Skip if not in english
		if chapter.Attributes.TranslatedLanguage != "en" {
			continue
		}

		// Create our save folder path
		chapterPath := dirImages + chapter.Id + "/"
		err := os.MkdirAll(chapterPath, os.ModePerm)
		if err != nil {
			log.Fatalf("%v", err)
		}

		// Get the mangadex@home url we will download the images from
		opts := swagger.AtHomeApiGetAtHomeServerChapterIdOpts{}
		mdexAtHome, resp, err := client.AtHomeApi.GetAtHomeServerChapterId(ctx, chapter.Id, &opts)
		if err != nil {
			log.Fatalf("%v", err)
		}
		if resp.StatusCode != 200 {
			fmt.Printf("HTTP ERROR CODE %d\n", resp.StatusCode)
			continue
		}

		// Loop through all chapter for this manga and save to disk
		for c, image := range chapter.Attributes.Data {

			// Create the url we will download
			start := time.Now()
			filename := chapterPath + image
			url := mdexAtHome.BaseUrl + "/data/" + chapter.Attributes.Hash + "/" + image
			fmt.Printf("%d/%d (image %d/%d) -> %s\n", i, len(itemsChapters), c+1, len(chapter.Attributes.Data), url)

			// Try to download
			imgResp, err := http.Get(url)
			if err != nil {
				fmt.Printf("%v\n", err)
				reportToMangadexNetwork(url, filename, start, false, false)
				continue
			}
			cacheHit := imgResp.Header.Get("X-Cache")

			// Open a file for writing and write the response!
			file, err := os.Create(filename)
			if err != nil {
				fmt.Printf("%v\n", err)
				reportToMangadexNetwork(url, filename, start, false, cacheHit == "HIT")
				continue
			}
			_, err = io.Copy(file, imgResp.Body)
			if err != nil {
				fmt.Printf("%v\n", err)
				reportToMangadexNetwork(url, filename, start, false, cacheHit == "HIT")
				continue
			}

			// Report to mangadex @ home network!
			reportToMangadexNetwork(url, filename, start, true, cacheHit == "HIT")

		}
		fmt.Println()

	}

}