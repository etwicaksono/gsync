package commands

import "github.com/spf13/cobra"

var gsheetFinanceCmd = &cobra.Command{
	Use:   "gsheet-finance",
	Short: "GSheet Finance commands",
	Long:  `GSheet Finance commands`,
}

func init() {
	rootCmd.AddCommand(gsheetFinanceCmd)
}
