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

const (
	newFinanceInsertedMessage   = "INSERTED TO NEW FINANCE APP"
	budgetBakersInsertedMessage = "INSERTED TO BUDGETBAKERS"
	financeAppType              = "financeapp"
	budgetBakersAppType         = "budgetbakers"
	transactionType             = "transaction"
	transferType             = "transfer"
)
