package calculate

import (
	"github.com/similar-manga/similar/cmd"
	"github.com/spf13/cobra"
)

var calculateCmd = &cobra.Command{
	Use:   "calculate",
	Short: "calculate mappings or similar data",
	Long: `
Actions related to the calculation of mappings and similar entries.`,
}

func init() {
	cmd.RootCmd.AddCommand(calculateCmd)
}
