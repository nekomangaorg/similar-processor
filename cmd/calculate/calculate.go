package calculate

import (
	"github.com/similar-manga/similar/cmd"
	"github.com/spf13/cobra"
	"os"
)

var calculateCmd = &cobra.Command{
	Use:   "calculate",
	Short: "calculate command",
	Long: `
Actions related to the calculation of mappings and similar entries.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	cmd.RootCmd.AddCommand(calculateCmd)
}
