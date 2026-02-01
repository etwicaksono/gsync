 package commands
 
 import (
 	"context"
 	"fmt"
 	"os"
 	"strconv"
 	"strings"
 	"time"
 
 	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/appsscript"
 	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/financeapi"
 	"github.com/spf13/cobra"
 )
 
 var apiPushSkip, apiPushLimit int
 
 var financeApiPushCmd = &cobra.Command{
 	Use:   "api-push",
 	Short: "Push data from Google Sheet to Finance Manager API",
 	Long:  `Push data from Google Sheet to Finance Manager API via REST endpoints`,
 }
 
 var financeApiPushTransactionCmd = &cobra.Command{
 	Use:   "transaction",
 	Short: "Push transactions from Google Sheet to Finance Manager API",
 	Long:  `Fetch transactions from Google Sheet via Apps Script and push to Finance Manager API`,
 	RunE: func(cmd *cobra.Command, args []string) error {
 		return runApiPushTransaction(cmd.Context())
 	},
 }
 
 var financeApiPushTransferCmd = &cobra.Command{
 	Use:   "transfer",
 	Short: "Push transfers from Google Sheet to Finance Manager API",
 	Long:  `Fetch transfers from Google Sheet via Apps Script and push to Finance Manager API`,
 	RunE: func(cmd *cobra.Command, args []string) error {
 		return runApiPushTransfer(cmd.Context())
 	},
 }
 
var financeApiPushDebugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Show unique account/category names from Google Sheet",
	Long:  `Fetch data from Google Sheet and display unique account/category names for mapping`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runApiPushDebug(cmd.Context())
	},
}

 func init() {
 	financeCmd.AddCommand(financeApiPushCmd)
 	financeApiPushCmd.AddCommand(financeApiPushTransactionCmd)
 	financeApiPushCmd.AddCommand(financeApiPushTransferCmd)
	financeApiPushCmd.AddCommand(financeApiPushDebugCmd)
 
 	financeApiPushCmd.PersistentFlags().IntVarP(&apiPushSkip, "skip", "s", 0, "Number of records to skip")
 	financeApiPushCmd.PersistentFlags().IntVarP(&apiPushLimit, "limit", "l", 0, "Number of records to process (0 = all)")
 }
 
 func runApiPushTransaction(ctx context.Context) error {
 	// Validate config
 	if cfg.AppsScript.URL == "" {
 		return fmt.Errorf("APPS_SCRIPT_URL is not configured")
 	}
 	if cfg.FinanceApi.Email == "" || cfg.FinanceApi.Password == "" {
 		return fmt.Errorf("FINANCE_API_EMAIL and FINANCE_API_PASSWORD are required")
 	}
 
 	// Create clients
 	appsClient := appsscript.NewClient(cfg.AppsScript.URL)
 	financeClient, err := financeapi.NewClient(cfg.FinanceApi.URL, cfg.FinanceApi.Email, cfg.FinanceApi.Password)
 	if err != nil {
 		return fmt.Errorf("create finance client: %w", err)
 	}
 
 	// Login to Finance Manager
 	fmt.Println("Logging in to Finance Manager...")
 	if err := financeClient.Login(ctx); err != nil {
 		return fmt.Errorf("login failed: %w", err)
 	}
 	fmt.Println("Login successful")
 
 	// Fetch master data
 	fmt.Println("Fetching accounts and categories...")
	accounts, err := financeClient.FetchAccounts(ctx)
	if err != nil {
 		return fmt.Errorf("fetch accounts: %w", err)
 	}
	categories, err := financeClient.FetchCategories(ctx)
	if err != nil {
 		return fmt.Errorf("fetch categories: %w", err)
 	}
	fmt.Printf("Master data loaded: %d accounts, %d categories\n", len(accounts), len(categories))
 
	// Apply default mappings
	financeClient.SetAccountMapping(financeapi.DefaultAccountMapping)
	financeClient.SetCategoryMapping(financeapi.DefaultCategoryMapping)

 	// Fetch transactions from Google Sheet
 	fmt.Printf("Fetching transactions (skip=%d, limit=%d)...\n", apiPushSkip, apiPushLimit)
 	resp, err := appsClient.FetchTransactions(ctx, apiPushSkip, apiPushLimit)
 	if err != nil {
 		return fmt.Errorf("fetch transactions: %w", err)
 	}
 	fmt.Printf("Found %d transactions (total: %d)\n", len(resp.Data), resp.Meta.TotalData)
 
 	if len(resp.Data) == 0 {
 		fmt.Println("No transactions to process")
 		return nil
 	}
 
 	// Process each transaction
 	var successCount, failCount int
 	for i, tx := range resp.Data {
 		fmt.Printf("[%d/%d] Processing transaction #%s...\n", i+1, len(resp.Data), tx.No)
 
 		// Map account name to ID
		accountID, ok := financeClient.GetAccountIDWithMapping(tx.Account)
 		if !ok {
			fmt.Fprintf(os.Stderr, "  ERROR: Account not found: %q (available: %v)\n", tx.Account, financeClient.GetAccountNames())
 			failCount++
 			continue
 		}
 
 		// Map category name to ID
		categoryID, ok := financeClient.GetCategoryIDWithMapping(tx.Category)
 		if !ok {
			fmt.Fprintf(os.Stderr, "  ERROR: Category not found: %q (available: %v)\n", tx.Category, financeClient.GetCategoryNames())
 			failCount++
 			continue
 		}
 
 		// Parse amount
 		amount, err := strconv.ParseFloat(tx.Amount, 64)
 		if err != nil {
 			fmt.Fprintf(os.Stderr, "  ERROR: Invalid amount: %s\n", tx.Amount)
 			failCount++
 			continue
 		}

		// Make amount positive (API expects positive amounts, type indicates expense/income)
		if amount < 0 {
			amount = -amount
		}
		if amount == 0 {
			fmt.Fprintf(os.Stderr, "  ERROR: Amount cannot be 0\n")
			failCount++
			continue
		}
 
 		// Parse date (DD/MM/YYYY to time.Time)
 		dateTime, err := parseSheetDate(tx.Date)
 		if err != nil {
 			fmt.Fprintf(os.Stderr, "  ERROR: Invalid date: %s\n", tx.Date)
 			failCount++
 			continue
 		}
 
 		// Create transaction
 		req := financeapi.CreateTransactionRequest{
 			Amount:        amount,
 			AccountID:     accountID,
 			CategoryID:    categoryID,
			Date:          dateTime.Format(time.RFC3339),
 			Type:          strings.ToLower(tx.Type),
 			PaymentStatus: "Cleared",
 			Description:   tx.Note,
 		}
 
 		if _, err := financeClient.CreateTransaction(ctx, req); err != nil {
 			fmt.Fprintf(os.Stderr, "  ERROR: %v\n", err)
 			failCount++
 			continue
 		}
 
 		// Update status in Google Sheet
 		statusReq := appsscript.UpdateStatusRequest{
 			No:     tx.No,
 			Status: "INSERTED",
 			Type:   "transaction",
 		}
 		if _, err := appsClient.UpdateStatus(ctx, statusReq); err != nil {
 			fmt.Fprintf(os.Stderr, "  WARNING: Failed to update status: %v\n", err)
 		}
 
 		fmt.Printf("  SUCCESS: Created transaction\n")
 		successCount++
 	}
 
 	fmt.Printf("\nSummary: %d success, %d failed\n", successCount, failCount)
 	return nil
 }
 
 func runApiPushTransfer(ctx context.Context) error {
 	// Validate config
 	if cfg.AppsScript.URL == "" {
 		return fmt.Errorf("APPS_SCRIPT_URL is not configured")
 	}
 	if cfg.FinanceApi.Email == "" || cfg.FinanceApi.Password == "" {
 		return fmt.Errorf("FINANCE_API_EMAIL and FINANCE_API_PASSWORD are required")
 	}
 
 	// Create clients
 	appsClient := appsscript.NewClient(cfg.AppsScript.URL)
 	financeClient, err := financeapi.NewClient(cfg.FinanceApi.URL, cfg.FinanceApi.Email, cfg.FinanceApi.Password)
 	if err != nil {
 		return fmt.Errorf("create finance client: %w", err)
 	}
 
 	// Login to Finance Manager
 	fmt.Println("Logging in to Finance Manager...")
 	if err := financeClient.Login(ctx); err != nil {
 		return fmt.Errorf("login failed: %w", err)
 	}
 	fmt.Println("Login successful")
 
 	// Fetch master data (accounts only for transfers)
 	fmt.Println("Fetching accounts...")
 	if _, err := financeClient.FetchAccounts(ctx); err != nil {
 		return fmt.Errorf("fetch accounts: %w", err)
 	}
 	fmt.Println("Master data loaded")
 
	// Apply default mappings
	financeClient.SetAccountMapping(financeapi.DefaultAccountMapping)

 	// Fetch transfers from Google Sheet
 	fmt.Printf("Fetching transfers (skip=%d, limit=%d)...\n", apiPushSkip, apiPushLimit)
 	resp, err := appsClient.FetchTransfers(ctx, apiPushSkip, apiPushLimit)
 	if err != nil {
 		return fmt.Errorf("fetch transfers: %w", err)
 	}
 	fmt.Printf("Found %d transfers (total: %d)\n", len(resp.Data), resp.Meta.TotalData)
 
 	if len(resp.Data) == 0 {
 		fmt.Println("No transfers to process")
 		return nil
 	}
 
 	// Process each transfer
 	var successCount, failCount int
 	for i, tf := range resp.Data {
 		fmt.Printf("[%d/%d] Processing transfer #%s...\n", i+1, len(resp.Data), tf.No)
 
 		// Map from account name to ID
		fromAccountID, ok := financeClient.GetAccountIDWithMapping(tf.FromAccount)
 		if !ok {
 			fmt.Fprintf(os.Stderr, "  ERROR: From account not found: %s\n", tf.FromAccount)
 			failCount++
 			continue
 		}
 
 		// Map to account name to ID
		toAccountID, ok := financeClient.GetAccountIDWithMapping(tf.ToAccount)
 		if !ok {
 			fmt.Fprintf(os.Stderr, "  ERROR: To account not found: %s\n", tf.ToAccount)
 			failCount++
 			continue
 		}
 
 		// Parse amount
 		amount, err := strconv.ParseFloat(tf.Amount, 64)
 		if err != nil {
 			fmt.Fprintf(os.Stderr, "  ERROR: Invalid amount: %s\n", tf.Amount)
 			failCount++
 			continue
 		}

		// Make amount positive
		if amount < 0 {
			amount = -amount
		}
		if amount == 0 {
			fmt.Fprintf(os.Stderr, "  ERROR: Amount cannot be 0\n")
			failCount++
			continue
		}
 
 		// Parse date (DD/MM/YYYY to time.Time)
 		dateTime, err := parseSheetDate(tf.Date)
 		if err != nil {
 			fmt.Fprintf(os.Stderr, "  ERROR: Invalid date: %s\n", tf.Date)
 			failCount++
 			continue
 		}
 
 		// Create transfer
 		req := financeapi.CreateTransferRequest{
 			Amount:        amount,
 			FromAccountID: fromAccountID,
 			ToAccountID:   toAccountID,
			Date:          dateTime.Format(time.RFC3339),
 			PaymentStatus: "Cleared",
 			Description:   tf.Note,
 		}
 
 		if _, err := financeClient.CreateTransfer(ctx, req); err != nil {
 			fmt.Fprintf(os.Stderr, "  ERROR: %v\n", err)
 			failCount++
 			continue
 		}
 
 		// Update status in Google Sheet
 		statusReq := appsscript.UpdateStatusRequest{
 			No:     tf.No,
 			Status: "INSERTED",
 			Type:   "transfer",
 		}
 		if _, err := appsClient.UpdateStatus(ctx, statusReq); err != nil {
 			fmt.Fprintf(os.Stderr, "  WARNING: Failed to update status: %v\n", err)
 		}
 
 		fmt.Printf("  SUCCESS: Created transfer\n")
 		successCount++
 	}
 
 	fmt.Printf("\nSummary: %d success, %d failed\n", successCount, failCount)
 	return nil
 }
 
 // parseSheetDate parses a date string in DD/MM/YYYY format
 func parseSheetDate(dateStr string) (time.Time, error) {
 	// Try DD/MM/YYYY format
 	t, err := time.Parse("02/01/2006", dateStr)
 	if err == nil {
 		return t, nil
 	}
 
 	// Try other common formats
 	formats := []string{
 		"2006-01-02",
 		"02-01-2006",
 		"01/02/2006",
 	}
 	for _, fmt := range formats {
 		t, err = time.Parse(fmt, dateStr)
 		if err == nil {
 			return t, nil
 		}
 	}
 
 	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
 }

func runApiPushDebug(ctx context.Context) error {
	if cfg.AppsScript.URL == "" {
		return fmt.Errorf("APPS_SCRIPT_URL is not configured")
	}

	appsClient := appsscript.NewClient(cfg.AppsScript.URL)

	// Fetch transactions
	fmt.Println("Fetching transactions from Google Sheet...")
	txResp, err := appsClient.FetchTransactions(ctx, 0, 100)
	if err != nil {
		return fmt.Errorf("fetch transactions: %w", err)
	}

	// Collect unique accounts and categories
	accountSet := make(map[string]bool)
	categorySet := make(map[string]bool)
	for _, tx := range txResp.Data {
		accountSet[tx.Account] = true
		categorySet[tx.Category] = true
	}

	fmt.Printf("\n=== Unique Accounts in Google Sheet (first 100 transactions) ===\n")
	for acc := range accountSet {
		fmt.Printf("  - %s\n", acc)
	}

	fmt.Printf("\n=== Unique Categories in Google Sheet (first 100 transactions) ===\n")
	for cat := range categorySet {
		fmt.Printf("  - %s\n", cat)
	}

	// Fetch from Finance Manager if credentials available
	if cfg.FinanceApi.Email != "" && cfg.FinanceApi.Password != "" {
		financeClient, err := financeapi.NewClient(cfg.FinanceApi.URL, cfg.FinanceApi.Email, cfg.FinanceApi.Password)
		if err != nil {
			return fmt.Errorf("create finance client: %w", err)
		}

		fmt.Println("\nLogging in to Finance Manager...")
		if err := financeClient.Login(ctx); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		accounts, _ := financeClient.FetchAccounts(ctx)
		categories, _ := financeClient.FetchCategories(ctx)

		fmt.Printf("\n=== Accounts in Finance Manager ===\n")
		for _, acc := range accounts {
			fmt.Printf("  - %s (ID: %s)\n", acc.Name, acc.ID)
		}

		fmt.Printf("\n=== Categories in Finance Manager ===\n")
		for _, cat := range categories {
			fmt.Printf("  - %s (ID: %s)\n", cat.Name, cat.ID)
		}
	}

	return nil
}
