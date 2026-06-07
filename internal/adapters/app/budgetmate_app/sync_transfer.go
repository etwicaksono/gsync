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

func (a *adapter) SyncTransfers(ctx context.Context, skip, limit int) error {
	// Apply default mappings
	fmt.Printf("Fetching transfers (skip=%d, limit=%d)...\n", skip, limit)
	resp, err := a.appsClient.FetchTransfers(ctx, skip, limit)
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
		fromAccountID, ok := a.budgetmateClient.GetAccountIDWithMapping(tf.FromAccount)
		if !ok {
			return fmt.Errorf("from account not found: %q (available: %v)", tf.FromAccount, a.budgetmateClient.GetAccountNames())
		}

		// Map to account name to ID
		toAccountID, ok := a.budgetmateClient.GetAccountIDWithMapping(tf.ToAccount)
		if !ok {
			return fmt.Errorf("to account not found: %q (available: %v)", tf.ToAccount, a.budgetmateClient.GetAccountNames())
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
		createReq := budgetmateapi.CreateTransferRequest{
			Amount:        amount,
			FromAccountID: fromAccountID,
			ToAccountID:   toAccountID,
			Date:          time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.Local).UTC().Format(time.RFC3339),
			PaymentStatus: "Cleared",
			Description:   tf.Note,
		}

		if _, err := a.budgetmateClient.CreateTransfer(ctx, createReq); err != nil {
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
			Status: constants.SheetStatusBudgetMateInserted,
			Type:   constants.SyncTypeTransfer,
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
	fmt.Printf("The following transfers were created in BudgetMate but their Google Sheet\n")
	fmt.Printf("status could NOT be updated. Re-run with these numbers or update them manually:\n")
	for _, no := range statusFailedNos {
		fmt.Printf("  - #%s\n", no)
	}
	return fmt.Errorf("%d transfer(s) created but status update failed: %s",
		len(statusFailedNos), strings.Join(statusFailedNos, ", "))
}
