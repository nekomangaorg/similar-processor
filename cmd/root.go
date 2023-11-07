package cmd

import (
	"github.com/similar-manga/similar/internal"
	"github.com/spf13/cobra"
	"os"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "calculate",
	Short: "Find recommendations between all MangaDex manga",
	Long: `
 A scraping and matching utility to find manga which are close in content to other manga.

 If you are trying to do a full run from scratch
 1. calculate mangadex add
 2. calculate mangadex metadata --all
 3. calculate calculate 
`,
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	internal.ConnectDB()

}
