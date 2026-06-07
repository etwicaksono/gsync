package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/app/budgetbakers_app"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/appsscript"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/budgetbakers"
	"github.com/spf13/cobra"
)

var (
	bbSkip  int
	bbLimit int
)

var budgetbakersCmd = &cobra.Command{
	Use:     "budgetbakers",
	Short:   "BudgetBakers REST API sync commands",
	Long:    `Synchronize data from Google Sheets to BudgetBakers via REST API.`,
	Aliases: []string{"bb"},
}

var bbPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push records to BudgetBakers",
	Long:  `Push records (transactions, transfers) from Google Sheet to BudgetBakers API`,
}

var bbTransactionCmd = &cobra.Command{
	Use:   "transaction",
	Short: "Run transaction synchronization",
	Long:  `Run BudgetBakers transaction synchronization to insert transactions from Google Sheets.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBudgetBakersSync()
	},
}

var bbTransferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Run transfer synchronization",
	Long:  `Run BudgetBakers transfer synchronization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("⚠️  Transfers are currently not supported via the BudgetBakers API.")
		fmt.Println("Please create transfers manually or use two separate transactions (expense & income).")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(budgetbakersCmd)
	budgetbakersCmd.AddCommand(bbPushCmd)
	bbPushCmd.AddCommand(bbTransactionCmd)
	bbPushCmd.AddCommand(bbTransferCmd)

	for _, cmd := range []*cobra.Command{bbTransactionCmd, bbTransferCmd} {
		cmd.Flags().IntVarP(&bbSkip, "skip", "s", 0, "Number of records to skip")
		cmd.Flags().IntVarP(&bbLimit, "limit", "l", 0, "Number of records to process (0 = all)")
	}
}

func runBudgetBakersSync() error {
	// Safety check: confirm if limit=0 (process all records)
	if bbLimit == 0 {
		fmt.Println("WARNING: --limit is 0, which means ALL uninserted records will be processed.")
		fmt.Println("Are you sure you want to continue? (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	appsClient := appsscript.NewClient(cfg.AppsScript.URL)
	bbClient := budgetbakers.NewClient(cfg.BudgetBakers.APIToken)
	bbApp := budgetbakers_app.NewBudgetBakersApp(appsClient, bbClient)

	ctx := context.Background()
	if err := bbApp.SyncTransactions(ctx, bbSkip, bbLimit); err != nil {
		return fmt.Errorf("failed to sync transactions: %w", err)
	}

	fmt.Println("Synchronization complete!")
	return nil
}
