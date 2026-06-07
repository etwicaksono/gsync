package commands

import (
	"fmt"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/app/budgetmate_app"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/appsscript"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/budgetmateapi"
	"github.com/spf13/cobra"
)

var apiPushSkip, apiPushLimit int

var budgetmateCmd = &cobra.Command{
	Use:     "budgetmate",
	Short:   "Push data from Google Sheet to BudgetMate API",
	Long:    `Push data from Google Sheet to BudgetMate API via REST endpoints`,
	Aliases: []string{"api-push"},
}

var bmTransactionCmd = &cobra.Command{
	Use:   "transaction",
	Short: "Push transactions from Google Sheet to BudgetMate API",
	Long:  `Fetch transactions from Google Sheet via Apps Script and push to BudgetMate API`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate config
		if cfg.AppsScript.URL == "" {
			return fmt.Errorf("APPS_SCRIPT_URL is not configured")
		}
		if cfg.BudgetMateApi.Email == "" || cfg.BudgetMateApi.Password == "" {
			return fmt.Errorf("BUDGETMATE_API_EMAIL and BUDGETMATE_API_PASSWORD are required")
		}

		appsClient := appsscript.NewClient(cfg.AppsScript.URL)
		bmClient, err := budgetmateapi.NewClient(cfg.BudgetMateApi.URL, cfg.BudgetMateApi.Email, cfg.BudgetMateApi.Password)
		if err != nil {
			return fmt.Errorf("create budgetmate client: %w", err)
		}
		
		ctx := cmd.Context()
		
		fmt.Println("Logging in to BudgetMate API...")
		if err := bmClient.Login(ctx); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
		fmt.Println("Login successful")

		fmt.Println("Fetching accounts and categories...")
		if _, err := bmClient.FetchAccounts(ctx); err != nil {
			return fmt.Errorf("fetch accounts: %w", err)
		}
		if _, err := bmClient.FetchCategories(ctx); err != nil {
			return fmt.Errorf("fetch categories: %w", err)
		}

		app := budgetmate_app.NewBudgetMateApp(appsClient, bmClient)
		return app.SyncTransactions(ctx, apiPushSkip, apiPushLimit)
	},
}

var bmTransferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Push transfers from Google Sheet to BudgetMate API",
	Long:  `Fetch transfers from Google Sheet via Apps Script and push to BudgetMate API`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate config
		if cfg.AppsScript.URL == "" {
			return fmt.Errorf("APPS_SCRIPT_URL is not configured")
		}
		if cfg.BudgetMateApi.Email == "" || cfg.BudgetMateApi.Password == "" {
			return fmt.Errorf("BUDGETMATE_API_EMAIL and BUDGETMATE_API_PASSWORD are required")
		}

		appsClient := appsscript.NewClient(cfg.AppsScript.URL)
		bmClient, err := budgetmateapi.NewClient(cfg.BudgetMateApi.URL, cfg.BudgetMateApi.Email, cfg.BudgetMateApi.Password)
		if err != nil {
			return fmt.Errorf("create budgetmate client: %w", err)
		}

		ctx := cmd.Context()

		fmt.Println("Logging in to BudgetMate API...")
		if err := bmClient.Login(ctx); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
		fmt.Println("Login successful")

		fmt.Println("Fetching accounts...")
		if _, err := bmClient.FetchAccounts(ctx); err != nil {
			return fmt.Errorf("fetch accounts: %w", err)
		}

		app := budgetmate_app.NewBudgetMateApp(appsClient, bmClient)
		return app.SyncTransfers(ctx, apiPushSkip, apiPushLimit)
	},
}

var bmDebugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Show unique account/category names from Google Sheet",
	Long:  `Fetch data from Google Sheet and display unique account/category names for mapping`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.AppsScript.URL == "" {
			return fmt.Errorf("APPS_SCRIPT_URL is not configured")
		}

		appsClient := appsscript.NewClient(cfg.AppsScript.URL)
		var bmClient *budgetmateapi.Client
		
		if cfg.BudgetMateApi.Email != "" && cfg.BudgetMateApi.Password != "" {
			var err error
			bmClient, err = budgetmateapi.NewClient(cfg.BudgetMateApi.URL, cfg.BudgetMateApi.Email, cfg.BudgetMateApi.Password)
			if err != nil {
				return fmt.Errorf("create budgetmate client: %w", err)
			}
		}

		app := budgetmate_app.NewBudgetMateApp(appsClient, bmClient)
		return app.DebugMapping(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(budgetmateCmd)
	budgetmateCmd.AddCommand(bmTransactionCmd)
	budgetmateCmd.AddCommand(bmTransferCmd)
	budgetmateCmd.AddCommand(bmDebugCmd)

	for _, cmd := range []*cobra.Command{bmTransactionCmd, bmTransferCmd} {
		cmd.Flags().IntVarP(&apiPushSkip, "skip", "s", 0, "Number of records to skip")
		cmd.Flags().IntVarP(&apiPushLimit, "limit", "l", 0, "Number of records to process (0 = all)")
	}
}
