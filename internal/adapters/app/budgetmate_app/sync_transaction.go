package budgetmate_app

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/appsscript"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/budgetmateapi"
	"github.com/etwicaksono/data-transaksi-service/internal/constants"
)

func (a *adapter) SyncTransactions(ctx context.Context, skip, limit int) error {
	// Apply default mappings
	fmt.Printf("Fetching transactions (skip=%d, limit=%d)...\n", skip, limit)
	resp, err := a.appsClient.FetchTransactions(ctx, skip, limit)
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
		accountID, ok := a.budgetmateClient.GetAccountIDWithMapping(tx.Account)
		if !ok {
			return fmt.Errorf("account not found: %q (available: %v)", tx.Account, a.budgetmateClient.GetAccountNames())
		}

		// Map category name to ID
		categoryID, ok := a.budgetmateClient.GetCategoryIDWithMapping(tx.Category)
		if !ok {
			return fmt.Errorf("category not found: %q (available: %v)", tx.Category, a.budgetmateClient.GetCategoryNames())
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
		createReq := budgetmateapi.CreateTransactionRequest{
			Amount:        amount,
			AccountID:     accountID,
			CategoryID:    categoryID,
			Date:          time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.Local).UTC().Format(time.RFC3339),
			Type:          strings.ToLower(tx.Type),
			PaymentStatus: "Cleared",
			Description:   tx.Note,
		}

		if _, err := a.budgetmateClient.CreateTransaction(ctx, createReq); err != nil {
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
			Status: constants.SheetStatusBudgetMateInserted,
			Type:   constants.SyncTypeTransactionLower,
			Target: constants.SyncTargetBudgetMate,
		})
	}
	
	if len(statusRequests) > 0 {
		if _, err := a.appsClient.UpdateStatusBatch(ctx, statusRequests); err != nil {
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
	fmt.Printf("The following transactions were created in BudgetMate but their Google Sheet\n")
	fmt.Printf("status could NOT be updated. Re-run with these numbers or update them manually:\n")
	for _, no := range statusFailedNos {
		fmt.Printf("  - #%s\n", no)
	}
	return fmt.Errorf("%d transaction(s) created but status update failed: %s",
		len(statusFailedNos), strings.Join(statusFailedNos, ", "))
}
