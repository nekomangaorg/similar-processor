package mangadex

import (
	"bufio"
	"context"
	"fmt"
	"github.com/antihax/optional"
	_ "github.com/mattn/go-sqlite3"
	"github.com/similar-manga/similar/internal"
	"github.com/similar-manga/similar/mangadex"
	"github.com/spf13/cobra"
	"go.uber.org/ratelimit"
	"os"
	"strconv"
	"strings"
	"time"
)

var metadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "This queries every manga uuid and updates the metadata",
	Long:  `Query MangaDex for every given manga and mangadex the json metadata in the database`,
	Run:   runMetadata,
}

func init() {
	mangadexCmd.AddCommand(metadataCmd)
	metadataCmd.Flags().BoolP("all", "a", false, "queries and updates the entire database")
	metadataCmd.Flags().StringP("id", "i", "", "update metadata for a specific uuid in the database")

}

func runMetadata(cmd *cobra.Command, args []string) {
	start := time.Now()

	updateAll, _ := cmd.Flags().GetBool("all")
	updateId, _ := cmd.Flags().GetString("id")

	client := CreateMangaDexClient()
	ctx := context.Background()

	if updateAll {
		fmt.Printf("Getting mangadex metadata for all entries\n")

		rateLimiter := ratelimit.New(1)

		mangaIdArray := collectAllMangaIds()

		for index, ids := range mangaIdArray {

			opts := mangadex.MangaApiGetSearchMangaOpts{}
			opts.OrderCreatedAt = optional.NewString("desc")
			opts.Limit = optional.NewInt32(100)
			opts.Ids = optional.NewInterface(ids)

			printProgress(index+1, len(mangaIdArray))

			mangaList := SearchMangaDex(rateLimiter, client, ctx, opts)

			for _, apiManga := range mangaList.Data {
				UpsertManga(apiManga)
			}
		}
		fmt.Println()

	} else if updateId != "" {
		fmt.Printf("Updating MangaDex metadata for %s\n", updateId)
		rateLimiter := ratelimit.New(1)

		opts := mangadex.MangaApiGetSearchMangaOpts{}
		opts.OrderCreatedAt = optional.NewString("desc")
		opts.Limit = optional.NewInt32(1)
		opts.Ids = optional.NewInterface([]string{updateId})
		mangaList := SearchMangaDex(rateLimiter, client, ctx, opts)
		for _, apiManga := range mangaList.Data {
			UpsertManga(apiManga)
		}

	} else {
		rateLimiter := ratelimit.New(1, ratelimit.Per(2*time.Second))

		readFile, err := os.Open("data/last_metadata_update.txt")
		internal.CheckErr(err)
		fileScanner := bufio.NewScanner(readFile)

		fileScanner.Split(bufio.ScanLines)

		var lastUpdatedTime string
		for fileScanner.Scan() {
			lastUpdatedTime = fileScanner.Text()
		}

		fmt.Printf("Getting mangadex metadata since last updated time -> %s\n", lastUpdatedTime)

		readFile.Close()

		currentLimit := int32(100)
		maxOffset := int32(10000)
		done := false

		for currentOffset := int32(0); currentOffset < maxOffset && done == false; currentOffset += currentLimit {

			opts := mangadex.MangaApiGetSearchMangaOpts{}
			opts.UpdatedAtSince = optional.NewString(lastUpdatedTime)
			opts.Limit = optional.NewInt32(currentLimit)
			opts.Offset = optional.NewInt32(currentOffset)
			fmt.Printf("\rGetting mangadex metadata for offset %d   ", currentOffset)

			mangaList := SearchMangaDex(rateLimiter, client, ctx, opts)

			if len(mangaList.Data) != 0 {
				for _, apiManga := range mangaList.Data {
					UpsertManga(apiManga)
				}
			} else {
				done = true
			}
		}
		fmt.Println()

	}

	metadataFile, err := os.OpenFile("data/last_metadata_update.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	internal.CheckErr(err)
	_, err = metadataFile.WriteString(strings.Split(time.Now().UTC().Format(time.RFC3339), "Z")[0])
	internal.CheckErr(err)
	metadataFile.Close()

	ExportManga()

	fmt.Printf("\t- Finished in %s\n", time.Since(start))
}

func printProgress(current, total int) {
	if total <= 0 {
		return
	}
	width := 50
	percent := float64(current) / float64(total) * 100
	completed := int(float64(width) * (float64(current) / float64(total)))
	if completed > width {
		completed = width
	}

	bar := strings.Repeat("=", completed) + strings.Repeat("-", width-completed)
	fmt.Printf("\r[%s] %.2f%% (%d/%d)", bar, percent, current, total)
}

func collectAllMangaIds() [][]string {
	var mangaIdArray [][]string
	processing := true
	dbOffset := 0

	for processing {
		rows, _ := internal.DB.Query("SELECT UUID FROM " + internal.TableManga + " ORDER BY UUID LIMIT 100 OFFSET " + strconv.Itoa(dbOffset))
		var mangaIds []string
		for rows.Next() {
			var uuid string
			rows.Scan(&uuid)
			mangaIds = append(mangaIds, uuid)
		}

		if len(mangaIds) == 0 {
			processing = false
			break
		}

		mangaIdArray = append(mangaIdArray, mangaIds)
		dbOffset = dbOffset + 100
		rows.Close()
	}
	return mangaIdArray
}
