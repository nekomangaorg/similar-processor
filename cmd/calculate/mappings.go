package calculate

import (
	"fmt"
	"github.com/spf13/cobra"
	"go.uber.org/ratelimit"
	"time"
)

var mappingsCmd = &cobra.Command{
	Use:   "mappings",
	Short: "This updates the similar calculations",
	Long:  `\nCalculate the mapping files`,
	Run:   runMappings,
}

func init() {
	calculateCmd.AddCommand(mappingsCmd)
}

func runMappings(cmd *cobra.Command, args []string) {
	start := time.Now()
	calculateMUIds()
	fmt.Printf("\t- Finished in %s\n", time.Since(start))

}

func calculateMUIds() {
	fmt.Println("Calculating New MangaUpdate Ids")
	rateLimiter := ratelimit.New(1)

	file := OpenMappingsFile("mangaupdates_new2mdex")
	defer file.Close()

	fileMap := FileAsMap(file)
	//	fileMapReverse := FileAsMapReverse(file)

	// mangaupdates
	// https://www.mangaupdates.com/series.html?id=`{id}`
	// https://api.mangaupdates.com/#operation/retrieveSeries
	// https://api.mangaupdates.com/v1/series/(base38 encoding of 7char ids)
	// https://api.mangaupdates.com/v1/series/66788345008/rss

	// Loop through all manga and try to get their chapter information for each
	start := time.Now()
	mangaList := GetAllManga()
	totalManga := len(mangaList)
	for index, manga := range mangaList {
		// Our search file
		if _, ok := manga.Links["mu"]; ok {

			muLink := manga.Links["mu"]

			// If the string is 7 long it is likely already the base36 format
			// Thus we should try to directly extract from it the new API id
			if AddAlreadyConvertedId(index, totalManga, manga.Id, muLink, file, fileMap, rateLimiter) {
				continue
			}

			// Else lets try to extract the first int from the string
			// This will be our API id number we will query with
			if CheckAndAddLegacyId(index, totalManga, manga.Id, muLink, file, fileMap, rateLimiter) {
				continue
			}
			fmt.Printf("%d/%d manga %s -> mu invalid %s\n", index+1, totalManga, manga.Id, muLink)
		} else {
			fmt.Printf("%d/%d manga %s -> has no MU id\n", index+1, totalManga, manga.Id)
		}
	}
	fmt.Printf("done processing MangaUpdates2NewUUID (%.2f seconds)!\n", time.Since(start).Seconds())
}
