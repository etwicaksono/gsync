package budgetbakers_app

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/appsscript"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/budgetbakers"
	"github.com/etwicaksono/data-transaksi-service/internal/constants"
)

// SyncTransactions fetches transactions from Google Sheets and pushes them to BudgetBakers
func (a *adapter) SyncTransactions(ctx context.Context, skip int, limit int) error {
	fmt.Println("Loading reference data (Accounts & Categories) from BudgetBakers...")
	var err error
	a.accounts, err = a.bbClient.GetAccounts()
	if err != nil {
		return fmt.Errorf("failed to load accounts: %w", err)
	}
	
	a.categories, err = a.bbClient.GetCategories()
	if err != nil {
		return fmt.Errorf("failed to load categories: %w", err)
	}
	fmt.Printf("Loaded %d accounts and %d categories.\n", len(a.accounts), len(a.categories))

	fmt.Printf("Fetching transactions from Google Sheets (skip=%d, limit=%d)...\n", skip, limit)
	transactionsResp, err := a.appsClient.FetchTransactions(ctx, skip, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch transactions: %w", err)
	}

	if len(transactionsResp.Data) == 0 {
		fmt.Println("No transactions to process.")
		return nil
	}

	fmt.Printf("Processing %d transactions...\n", len(transactionsResp.Data))
	
	var records []budgetbakers.Record
	var recordToTxMap = make(map[int]appsscript.SheetTransaction)

	for _, tx := range transactionsResp.Data {
		amount, err := strconv.ParseFloat(tx.Amount, 64)
		if err != nil {
			fmt.Printf("Warning: failed to parse amount for transaction %s: %v\n", tx.No, err)
			continue
		}

		// Ensure base amount is absolute before applying type logic
		if amount < 0 {
			amount = -amount
		}

		// Round to 2 decimal places to avoid API precision errors (e.g. 19,2)
		amount = float64(int(amount*100+0.5)) / 100

		if amount == 0 {
			fmt.Printf("Warning: amount cannot be 0 for transaction #%s\n", tx.No)
			continue
		}

		// BudgetBakers API: negative for expense, positive for income
		if strings.ToLower(tx.Type) == "expense" {
			amount = -amount
		}

		accID, err := a.findAccountID(tx.Account)
		if err != nil {
			fmt.Printf("Warning: %v\n", err)
			continue
		}

		catID, err := a.findCategoryID(tx.Category)
		if err != nil {
			fmt.Printf("Warning: %v\n", err)
			continue
		}

		// Parse date DD/MM/YYYY
		t, err := time.Parse("02/01/2006", tx.Date)
		if err != nil {
			fmt.Printf("Warning: invalid date format %s for tx %s: %v\n", tx.Date, tx.No, err)
			continue
		}

		// Combine with current time so it doesn't default to midnight UTC
		now := time.Now()
		t = time.Date(t.Year(), t.Month(), t.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.Local).UTC()

		note := tx.Note
		if note != "" && len(note) > 255 {
			note = note[:255]
		}

		record := budgetbakers.Record{
			AccountID:   accID,
			CategoryID:  catID,
			Amount:      budgetbakers.AmountWithCurrency{Value: amount},
			PaymentType: "cash",
			RecordDate:  t,
			Note:        note,
		}

		records = append(records, record)
		recordToTxMap[len(records)-1] = tx
	}

	// Batch size is up to 20 for BudgetBakers POST /records
	batchSize := 20
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		fmt.Printf("Sending batch of %d records to BudgetBakers...\n", len(batch))

		resp, err := a.bbClient.CreateRecords(batch)
		if err != nil {
			return fmt.Errorf("failed to send batch: %w", err)
		}

		if resp.Summary.Succeeded > 0 {
			fmt.Printf("Batch success: %d/%d (API Quota Remaining: %d)\n", resp.Summary.Succeeded, resp.Summary.Total, a.bbClient.RateLimitRemaining)
			
			var statusRequests []appsscript.UpdateStatusRequest
			for j, res := range resp.Results {
				tx := recordToTxMap[i+j]
				if res.Success {
					fmt.Printf("✓ [#%s] %s: Rp%.2f\n", tx.No, tx.Account, batch[j].Amount.Value)
					statusRequests = append(statusRequests, appsscript.UpdateStatusRequest{
						No:     tx.No,
						Status: constants.SheetStatusBudgetBakersInserted,
						Type:   constants.SyncTypeTransactionTitle,
						Target: constants.SyncTargetBudgetBakers,
					})
				} else {
					fmt.Printf("✗ [#%s] %s: Failed - %s (%s)\n", tx.No, tx.Account, res.Error, res.ErrorType)
				}
			}

			if len(statusRequests) > 0 {
				if _, err := a.appsClient.UpdateStatusBatch(ctx, statusRequests); err != nil {
					fmt.Printf("  Failed to update sheet statuses: %v\n", err)
				}
			}
		} else {
			fmt.Printf("Batch entirely failed. Client Errors: %d, Server Errors: %d (API Quota Remaining: %d)\n", resp.Summary.ClientErrors, resp.Summary.ServerErrors, a.bbClient.RateLimitRemaining)
			for j, res := range resp.Results {
				tx := recordToTxMap[i+j]
				if !res.Success {
					fmt.Printf("✗ [#%s] %s: Failed - %s (%s)\n", tx.No, tx.Account, res.Error, res.ErrorType)
				}
			}
		}
	}

	return nil
}

func (a *adapter) findAccountID(name string) (string, error) {
	name = strings.TrimSpace(name)
	
	for _, acc := range a.accounts {
		if strings.EqualFold(strings.TrimSpace(acc.Name), name) {
			return acc.ID, nil
		}
	}

	for _, acc := range a.accounts {
		if strings.Contains(strings.ToLower(acc.Name), strings.ToLower(name)) {
			return acc.ID, nil
		}
	}

	return "", fmt.Errorf("account not found: %s", name)
}

func (a *adapter) findCategoryID(name string) (string, error) {
	name = strings.TrimSpace(name)
	
	categoryMapping := map[string]string{
		"drug-store, chemist":         "drugstore",
		"restaurant, fast-food":       "restaurants & fast food",
		"clothes & footwear":          "clothes & shoes",
		"holiday, trips, hotels":      "holidays, trips, hotels",
		"refunds (tax, purchase)":     "refunds",
		"books, audio, subscriptions": "books, audio, subscription",
	}

	nameLower := strings.ToLower(name)
	if mappedName, exists := categoryMapping[nameLower]; exists {
		nameLower = mappedName
	}

	nameVar1 := strings.ReplaceAll(nameLower, ",", " &")
	nameVar2 := strings.ReplaceAll(nameLower, "&", ",")
	nameVar3 := strings.ReplaceAll(nameLower, ", ", " & ")

	for _, cat := range a.categories {
		catLower := strings.ToLower(strings.TrimSpace(cat.Name))
		if catLower == nameLower || catLower == nameVar1 || catLower == nameVar2 || catLower == nameVar3 {
			return cat.ID, nil
		}
	}

	for _, cat := range a.categories {
		catLower := strings.ToLower(cat.Name)
		if strings.Contains(catLower, nameLower) || 
		   strings.Contains(catLower, nameVar1) || 
		   strings.Contains(catLower, nameVar2) || 
		   strings.Contains(catLower, nameVar3) {
			return cat.ID, nil
		}
	}

	return "", fmt.Errorf("category not found: %s", name)
}
