package cmd

import (
	"github.com/similar-manga/similar/internal"
	"github.com/spf13/cobra"
	"os"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Short: "Find recommendations between all MangaDex manga",
	Long: `
 A scraping and matching utility to find manga which are close in content to other manga.

 If you are trying to do a full run from scratch
 1. ./similar init
 2. ./similar mangadex add
 3. ./similar mangadex metadata
 3. ./similar calculate mappings
 4. ./similar calculate similar
 
 After all the files you can generate the neko mapping db also.
  ./similar neko

 If you are running again after a while make sure you pull the latest from git, then rerun from scratch as the manga mappings and 
 manga update mappings are updated frequently.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
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
