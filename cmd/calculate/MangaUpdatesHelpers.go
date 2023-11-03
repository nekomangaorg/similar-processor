package calculate

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/similar-manga/similar/external"
	"github.com/similar-manga/similar/internal"
	"go.uber.org/ratelimit"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func AddAlreadyConvertedId(index int, total int, uuid string, muLink string, file *os.File, fileMap map[string]string, rateLimiter ratelimit.Limiter) bool {
	if len(muLink) == 7 {
		// Encode from base36 format
		idEncoded := int64(external.Decode(muLink))
		base10Id := strconv.FormatInt(idEncoded, 10)

		//check if the mappings file already has the entry
		if fileMap[base10Id] != "" {
			fmt.Printf("%d/%d manga %s -> mu id %s encoded into %s -> is new MU id and Already exists in File\n", index+1, total, uuid, muLink, base10Id)
			return true
		}

		// Try the new id!
		rateLimiter.Take()
		resp2, err := http.Get("https://api.mangaupdates.com/v1/series/" + base10Id)
		internal.CheckErr(err)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(resp2.Body)

		// Save if good!
		if err == nil && resp2.StatusCode == 200 {
			fmt.Printf("%d/%d manga %s -> mu id %s encoded into %s -> is new MU id!\n", index+1, total, uuid, muLink, base10Id)
			_, err := io.WriteString(file, base10Id+":::"+uuid+"\n")
			internal.CheckErr(err)
			return true
		}
	}
	return false
}

func CheckAndAddLegacyId(index int, total int, uuid string, muLink string, file *os.File, fileMap map[string]string, rateLimiter ratelimit.Limiter) bool {
	// For our ID conversion
	// https://www.unitconverters.net/numbers/base-36-to-decimal.htm
	re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)

	ints := re.FindAllString(muLink, -1)
	if len(ints) < 1 {
		return false
	}
	idOriginal, err := strconv.Atoi(ints[0])
	if err == nil {
		convertedId := strconv.Itoa(idOriginal)
		//check if id already exists by looping through the maps values
		//check if the mappings file already has the entry
		/*if fileMapReverse[uuid] != "" {
			fmt.Printf("%d/%d manga %s -> mu id %s encoded into %s -> is new MU id and Already exists in File\n", index+1, total, uuid, muLink, fileMapReverse[uuid])
			return true
		}*/

		for key, value := range fileMap {
			if value == uuid {
				fmt.Printf("%d/%d manga %s -> mu id of %d -> is old MU id... but was already converted to %s and Already exists in file.\n", index+1, total, uuid, idOriginal, key)
				return true
			}
		}

		rateLimiter.Take()
		// Try the existing as the id (not likely since mangadex won't have updated..)
		resp1, err1 := http.Get("https://api.mangaupdates.com/v1/series/" + convertedId)
		internal.CheckErr(err1)
		defer resp1.Body.Close()

		if err1 == nil && resp1.StatusCode == 200 {
			fmt.Printf("%d/%d manga %s -> mu id of %d -> is old MU id...\n", index+1, total, uuid, idOriginal)
			_, err := io.WriteString(file, convertedId+":::"+uuid+"\n")
			internal.CheckErr(err)

		} else {

			// We have a couple retires here
			ctr := 0
			ctrMax := 5
			for ctr < ctrMax {
				rateLimiter.Take()

				// If invalid, then try to get the page and parse it!
				// Query and get our html... (no api to get this...)
				url := "https://www.mangaupdates.com/series.html?id=" + convertedId
				client := &http.Client{}
				req, err := http.NewRequest("GET", url, nil)
				internal.CheckErr(err)
				req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")
				resp, err := client.Do(req)
				internal.CheckErr(err)
				defer resp.Body.Close()

				// Sleep if we get a warning, otherwise we don't retry again!
				if err == nil && resp.StatusCode == 429 {
					fmt.Printf("\u001B[1;31mEXTERNAL MU: http code %d (try %d of %d)\u001B[0m\n", resp.StatusCode, ctr, ctrMax)
					time.Sleep(2.0 * time.Second)
				}
				if err == nil && resp.StatusCode != 200 {
					fmt.Printf("\u001B[1;31mEXTERNAL MU: http code %d (try %d of %d)\u001B[0m\n", resp.StatusCode, ctr, ctrMax)
					time.Sleep(1.0 * time.Second)
				}

				// Load the HTML document
				// Logic found using google chrome (right click in inspector and copy "selector")
				if err == nil && resp.StatusCode == 200 {
					doc, err := goquery.NewDocumentFromReader(resp.Body)
					internal.CheckErr(err)

					if err == nil {
						rssUrl := doc.Find("#main_content > div:nth-child(2) > div.row.no-gutters > div.col-12.p-2 > a").AttrOr("href", "")
						paths := strings.Split(rssUrl, "/")
						if len(paths) > 3 {
							rssId := paths[len(paths)-2]
							fmt.Printf("%d/%d manga %s -> mu id of %d | RSS URL IS %s | %s id found\n", index+1, total, uuid, idOriginal, rssUrl, rssId)
							_, err := io.WriteString(file, convertedId+":::"+uuid+"\n")
							internal.CheckErr(err)
							return true
						}
					}
				}
				ctr += 1
			}
		}
	}
	return false

}
