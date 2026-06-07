package commands

import (
	"context"
	"fmt"
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

	// Phase 1 — create all transactions in Finance Manager.
	var createdNos []string
	for i, tx := range resp.Data {
		fmt.Printf("[%d/%d] Creating transaction #%s...\n", i+1, len(resp.Data), tx.No)

		// Map account name to ID
		accountID, ok := financeClient.GetAccountIDWithMapping(tx.Account)
		if !ok {
			return fmt.Errorf("account not found: %q (available: %v)", tx.Account, financeClient.GetAccountNames())
		}

		// Map category name to ID
		categoryID, ok := financeClient.GetCategoryIDWithMapping(tx.Category)
		if !ok {
			return fmt.Errorf("category not found: %q (available: %v)", tx.Category, financeClient.GetCategoryNames())
		}

		// Parse amount
		amount, err := strconv.ParseFloat(tx.Amount, 64)
		if err != nil {
			return fmt.Errorf("invalid amount %q: %w", tx.Amount, err)
		}

		// Make amount positive (API expects positive amounts, type indicates expense/income)
		if amount < 0 {
			amount = -amount
		}
		if amount == 0 {
			return fmt.Errorf("amount cannot be 0 for transaction #%s", tx.No)
		}

		// Parse date (DD/MM/YYYY to time.Time)
		dateTime, err := parseSheetDate(tx.Date)
		if err != nil {
			return fmt.Errorf("invalid date %q: %w", tx.Date, err)
		}

		// Create transaction
		now := time.Now()
		createReq := financeapi.CreateTransactionRequest{
			Amount:        amount,
			AccountID:     accountID,
			CategoryID:    categoryID,
			Date:          time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.Local).UTC().Format(time.RFC3339),
			Type:          strings.ToLower(tx.Type),
			PaymentStatus: "Cleared",
			Description:   tx.Note,
		}

		if _, err := financeClient.CreateTransaction(ctx, createReq); err != nil {
			return fmt.Errorf("create transaction #%s failed: %w", tx.No, err)
		}

		fmt.Printf("  created\n")
		createdNos = append(createdNos, tx.No)
	}

	// Phase 2 — update Google Sheet status for every successfully created transaction.
	fmt.Printf("\nUpdating Google Sheet status for %d transaction(s)...\n", len(createdNos))
	var statusFailedNos []string
	var statusRequests []appsscript.UpdateStatusRequest
	for _, no := range createdNos {
		statusRequests = append(statusRequests, appsscript.UpdateStatusRequest{
			No:     no,
			Status: newFinanceInsertedMessage,
			Type:   transactionType,
			Target: financeAppType,
		})
	}
	
	if len(statusRequests) > 0 {
		if _, err := appsClient.UpdateStatusBatch(ctx, statusRequests); err != nil {
			fmt.Printf("  BATCH UPDATE FAILED: %v\n", err)
			statusFailedNos = createdNos // all failed
		} else {
			fmt.Printf("  %d status updated successfully\n", len(createdNos))
		}
	}

	fmt.Printf("\nSummary: %d created", len(createdNos))
	if len(statusFailedNos) == 0 {
		fmt.Printf(", %d status updated\n", len(createdNos))
		return nil
	}

	fmt.Printf(", %d status update(s) FAILED\n", len(statusFailedNos))
	fmt.Printf("The following transactions were created in Finance Manager but their Google Sheet\n")
	fmt.Printf("status could NOT be updated. Re-run with these numbers or update them manually:\n")
	for _, no := range statusFailedNos {
		fmt.Printf("  - #%s\n", no)
	}
	return fmt.Errorf("%d transaction(s) created but status update failed: %s",
		len(statusFailedNos), strings.Join(statusFailedNos, ", "))
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

	// Phase 1 — create all transfers in Finance Manager.
	var createdNos []string
	for i, tf := range resp.Data {
		fmt.Printf("[%d/%d] Creating transfer #%s...\n", i+1, len(resp.Data), tf.No)

		// Map from account name to ID
		fromAccountID, ok := financeClient.GetAccountIDWithMapping(tf.FromAccount)
		if !ok {
			return fmt.Errorf("from account not found: %q (available: %v)", tf.FromAccount, financeClient.GetAccountNames())
		}

		// Map to account name to ID
		toAccountID, ok := financeClient.GetAccountIDWithMapping(tf.ToAccount)
		if !ok {
			return fmt.Errorf("to account not found: %q (available: %v)", tf.ToAccount, financeClient.GetAccountNames())
		}

		// Parse amount
		amount, err := strconv.ParseFloat(tf.Amount, 64)
		if err != nil {
			return fmt.Errorf("invalid amount %q: %w", tf.Amount, err)
		}

		// Make amount positive
		if amount < 0 {
			amount = -amount
		}
		if amount == 0 {
			return fmt.Errorf("amount cannot be 0 for transfer #%s", tf.No)
		}

		// Parse date (DD/MM/YYYY to time.Time)
		dateTime, err := parseSheetDate(tf.Date)
		if err != nil {
			return fmt.Errorf("invalid date %q: %w", tf.Date, err)
		}

		// Create transfer
		now := time.Now()
		createReq := financeapi.CreateTransferRequest{
			Amount:        amount,
			FromAccountID: fromAccountID,
			ToAccountID:   toAccountID,
			Date:          time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.Local).UTC().Format(time.RFC3339),
			PaymentStatus: "Cleared",
			Description:   tf.Note,
		}

		if _, err := financeClient.CreateTransfer(ctx, createReq); err != nil {
			return fmt.Errorf("create transfer #%s failed: %w", tf.No, err)
		}

		fmt.Printf("  created\n")
		createdNos = append(createdNos, tf.No)
	}

	// Phase 2 — update Google Sheet status for every successfully created transfer.
	fmt.Printf("\nUpdating Google Sheet status for %d transfer(s)...\n", len(createdNos))
	var statusFailedNos []string
	var statusRequests []appsscript.UpdateStatusRequest
	for _, no := range createdNos {
		statusRequests = append(statusRequests, appsscript.UpdateStatusRequest{
			No:     no,
			Status: newFinanceInsertedMessage,
			Type:   transferType,
			Target: financeAppType,
		})
	}
	
	if len(statusRequests) > 0 {
		if _, err := appsClient.UpdateStatusBatch(ctx, statusRequests); err != nil {
			fmt.Printf("  BATCH UPDATE FAILED: %v\n", err)
			statusFailedNos = createdNos // all failed
		} else {
			fmt.Printf("  %d status updated successfully\n", len(createdNos))
		}
	}

	fmt.Printf("\nSummary: %d created", len(createdNos))
	if len(statusFailedNos) == 0 {
		fmt.Printf(", %d status updated\n", len(createdNos))
		return nil
	}

	fmt.Printf(", %d status update(s) FAILED\n", len(statusFailedNos))
	fmt.Printf("The following transfers were created in Finance Manager but their Google Sheet\n")
	fmt.Printf("status could NOT be updated. Re-run with these numbers or update them manually:\n")
	for _, no := range statusFailedNos {
		fmt.Printf("  - #%s\n", no)
	}
	return fmt.Errorf("%d transfer(s) created but status update failed: %s",
		len(statusFailedNos), strings.Join(statusFailedNos, ", "))
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
