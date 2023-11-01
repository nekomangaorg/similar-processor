package mangadex

import (
	"context"
	"fmt"
	"github.com/antihax/optional"
	_ "github.com/mattn/go-sqlite3"
	"github.com/similar-manga/similar/mangadex"
	"github.com/spf13/cobra"
	"go.uber.org/ratelimit"
	"time"
)

// addCmd represents the new command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "queries and adds all the new manga UUID's to csv",
	Long:  `This searches for manga ordered by date added (newest first) and adds the Manga UUID to the csv until an existing UUID is found`,
	Run:   runAdd,
}

func init() {
	mangadexCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) {
	rateLimiter := ratelimit.New(1, ratelimit.Per(2*time.Second))

	client := CreateMangaDexClient()
	ctx := context.Background()

	currentLimit := int32(100)
	maxOffset := int32(10000)
	done := false

	for currentOffset := int32(0); currentOffset < maxOffset && done == false; currentOffset += currentLimit {

		opts := mangadex.MangaApiGetSearchMangaOpts2{}
		opts.OrderCreatedAt = optional.NewString("desc")
		opts.Limit = optional.NewInt32(currentLimit)
		opts.Offset = optional.NewInt32(currentOffset)

		mangaList := SearchMangaDex(rateLimiter, client, ctx, opts)

		for _, apiManga := range mangaList.Data {

			if !ExistsInDatabase(apiManga.Id) {
				InsertManga(apiManga)
			} else {
				fmt.Printf("\nFound manga in DB with ID: %s.\nFinished processing.\n", apiManga.Id)
				done = true
				break
			}

		}
	}
}
