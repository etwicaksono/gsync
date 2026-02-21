package commands

import "github.com/spf13/cobra"

var financeCmd = &cobra.Command{
	Use:   "finance",
	Short: "Finance commands",
	Long:  `Finance commands`,
}

func init() {
	rootCmd.AddCommand(financeCmd)
}

var (
	newFinanceInsertedMessage   = "INSERTED TO NEW FINANCE APP"
	budgetBakersInsertedMessage = "INSERTED TO BUDGETBAKERS"
)
