package commands

import (
	"fmt"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/app/budgetmate_app"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/appsscript"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/budgetmateapi"
	"github.com/spf13/cobra"
)

var apiPushSkip, apiPushLimit int
var apiPushDebug bool

var budgetmateCmd = &cobra.Command{
	Use:   "budgetmate",
	Short: "Push data from Google Sheet to BudgetMate API",
	Long:  `BudgetMate API commands to manage syncing your data.`,
}

var bmPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push records to BudgetMate",
	Long:  `Push records (transactions, transfers) from Google Sheet to BudgetMate API`,
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

		fmt.Println("Fetching accounts, categories and labels...")
		bmClient.SetDebug(apiPushDebug)
		if _, err := bmClient.FetchAccounts(ctx); err != nil {
			return fmt.Errorf("fetch accounts: %w", err)
		}
		if _, err := bmClient.FetchCategories(ctx); err != nil {
			return fmt.Errorf("fetch categories: %w", err)
		}
		if _, err := bmClient.FetchLabels(ctx); err != nil {
			return fmt.Errorf("fetch labels: %w", err)
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

var bmDebugLabelsCmd = &cobra.Command{
	Use:   "debug-labels",
	Short: "Diagnose why transaction labels are not reaching BudgetMate",
	Long:  `Compare labels coming from the Google Sheet against labels in BudgetMate and preview the outgoing payload`,
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
		return app.DebugLabels(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(budgetmateCmd)
	budgetmateCmd.AddCommand(bmPushCmd)
	bmPushCmd.AddCommand(bmTransactionCmd)
	bmPushCmd.AddCommand(bmTransferCmd)
	budgetmateCmd.AddCommand(bmDebugCmd)
	budgetmateCmd.AddCommand(bmDebugLabelsCmd)

	for _, cmd := range []*cobra.Command{bmTransactionCmd, bmTransferCmd} {
		cmd.Flags().IntVarP(&apiPushSkip, "skip", "s", 0, "Number of records to skip")
		cmd.Flags().IntVarP(&apiPushLimit, "limit", "l", 0, "Number of records to process (0 = all)")
		cmd.Flags().BoolVar(&apiPushDebug, "debug", false, "Log HTTP request and response bodies")
	}
}
