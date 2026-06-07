package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

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
	budgetbakersCmd.AddCommand(bbTransactionCmd)
	budgetbakersCmd.AddCommand(bbTransferCmd)

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
	sm, err := budgetbakers.NewSyncManager(appsClient, cfg.BudgetBakers.APIToken)
	if err != nil {
		return fmt.Errorf("failed to initialize sync manager: %w", err)
	}

	fmt.Println("Loading reference data (Accounts & Categories) from BudgetBakers...")
	if err := sm.LoadReferenceData(); err != nil {
		return fmt.Errorf("failed to load reference data: %w", err)
	}
	fmt.Printf("Loaded %d accounts and %d categories.\n", len(sm.Accounts), len(sm.Categories))

	fmt.Printf("Fetching transactions from Google Sheets (skip=%d, limit=%d)...\n", bbSkip, bbLimit)
	ctx := context.Background()
	transactionsResp, err := appsClient.FetchTransactions(ctx, bbSkip, bbLimit)
	if err != nil {
		return fmt.Errorf("failed to fetch transactions: %w", err)
	}

	if len(transactionsResp.Data) == 0 {
		fmt.Println("No transactions to process.")
		return nil
	}

	fmt.Printf("Processing %d transactions...\n", len(transactionsResp.Data))
	opts := budgetbakers.ProcessOptions{
		TargetApp:     budgetBakersAppType,
		SuccessStatus: budgetBakersInsertedMessage,
		ItemType:      transactionType,
	}
	if err := sm.ProcessTransactions(ctx, transactionsResp.Data, opts); err != nil {
		return fmt.Errorf("failed to process transactions: %w", err)
	}

	fmt.Println("Synchronization complete!")
	return nil
}
