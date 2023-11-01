package mangadex

import (
	"github.com/similar-manga/similar/cmd"
	"github.com/spf13/cobra"
	"os"
)

var mangadexCmd = &cobra.Command{
	Use:   "mangadex",
	Short: "mangadex command",
	Long: `
Actions related to the querying, and parsing of the MangaDex api.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	cmd.RootCmd.AddCommand(mangadexCmd)
}
