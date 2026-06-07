package budgetbakers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/appsscript"
)

// SyncManager coordinates syncing between Google Sheets and BudgetBakers
type SyncManager struct {
	AppsClient *appsscript.Client
	APIClient  *Client
	Accounts   []Account
	Categories []Category
}

// NewSyncManager creates a new sync manager
func NewSyncManager(appsClient *appsscript.Client, token string) (*SyncManager, error) {
	if token == "" {
		return nil, fmt.Errorf("budgetbakers API token is required")
	}

	return &SyncManager{
		AppsClient: appsClient,
		APIClient:  NewClient(token),
	}, nil
}

// LoadReferenceData fetches accounts and categories from BudgetBakers
func (s *SyncManager) LoadReferenceData() error {
	var err error
	s.Accounts, err = s.APIClient.GetAccounts()
	if err != nil {
		return fmt.Errorf("failed to load accounts: %w", err)
	}

	s.Categories, err = s.APIClient.GetCategories()
	if err != nil {
		return fmt.Errorf("failed to load categories: %w", err)
	}

	return nil
}

// FindAccountID finds the BudgetBakers account ID by name
func (s *SyncManager) FindAccountID(name string) (string, error) {
	name = strings.TrimSpace(name)
	
	// Exact match
	for _, acc := range s.Accounts {
		if strings.EqualFold(strings.TrimSpace(acc.Name), name) {
			return acc.ID, nil
		}
	}

	// Contains match
	for _, acc := range s.Accounts {
		if strings.Contains(strings.ToLower(acc.Name), strings.ToLower(name)) {
			return acc.ID, nil
		}
	}

	return "", fmt.Errorf("account not found: %s", name)
}

// FindCategoryID finds the BudgetBakers category ID by name
func (s *SyncManager) FindCategoryID(name string) (string, error) {
	name = strings.TrimSpace(name)
	
	// Manual mappings for categories that differ significantly between UI and API
	// Key: Google Sheets/UI name, Value: BudgetBakers API name
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

	// Create variations to handle "&" vs "," mismatches
	nameVar1 := strings.ReplaceAll(nameLower, ",", " &")
	nameVar2 := strings.ReplaceAll(nameLower, "&", ",")
	nameVar3 := strings.ReplaceAll(nameLower, ", ", " & ")

	// Exact match
	for _, cat := range s.Categories {
		catLower := strings.ToLower(strings.TrimSpace(cat.Name))
		if catLower == nameLower || catLower == nameVar1 || catLower == nameVar2 || catLower == nameVar3 {
			return cat.ID, nil
		}
	}

	// Contains match
	for _, cat := range s.Categories {
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

// ProcessOptions holds configuration for processing transactions
type ProcessOptions struct {
	TargetApp     string
	SuccessStatus string
	ItemType      string
}

// ProcessTransactions converts and sends a batch of transactions
func (s *SyncManager) ProcessTransactions(ctx context.Context, transactions []appsscript.SheetTransaction, opts ProcessOptions) error {
	if len(transactions) == 0 {
		return nil
	}

	var records []Record
	var recordToTxMap = make(map[int]appsscript.SheetTransaction)

	for _, tx := range transactions {
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

		accID, err := s.FindAccountID(tx.Account)
		if err != nil {
			fmt.Printf("Warning: %v\n", err)
			continue
		}

		catID, err := s.FindCategoryID(tx.Category)
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

		record := Record{
			AccountID:   accID,
			CategoryID:  catID,
			Amount:      AmountWithCurrency{Value: amount},
			PaymentType: "cash", // Or "debit_card", maybe default to cash? The js script didn't select this.
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

		resp, err := s.APIClient.CreateRecords(batch)
		if err != nil {
			return fmt.Errorf("failed to send batch: %w", err)
		}

		if resp.Summary.Succeeded > 0 {
			fmt.Printf("Batch success: %d/%d (API Quota Remaining: %d)\n", resp.Summary.Succeeded, resp.Summary.Total, s.APIClient.RateLimitRemaining)
			
			var statusRequests []appsscript.UpdateStatusRequest
			for j, res := range resp.Results {
				tx := recordToTxMap[i+j]
				if res.Success {
					fmt.Printf("✓ [#%s] %s: Rp%.2f\n", tx.No, tx.Account, batch[j].Amount.Value)
					statusRequests = append(statusRequests, appsscript.UpdateStatusRequest{
						No:     tx.No,
						Status: opts.SuccessStatus,
						Type:   opts.ItemType,
						Target: opts.TargetApp,
					})
				} else {
					fmt.Printf("✗ [#%s] %s: Failed - %s (%s)\n", tx.No, tx.Account, res.Error, res.ErrorType)
				}
			}

			if len(statusRequests) > 0 {
				if _, err := s.AppsClient.UpdateStatusBatch(ctx, statusRequests); err != nil {
					fmt.Printf("  Failed to update sheet statuses: %v\n", err)
				}
			}
		} else {
			fmt.Printf("Batch entirely failed. Client Errors: %d, Server Errors: %d (API Quota Remaining: %d)\n", resp.Summary.ClientErrors, resp.Summary.ServerErrors, s.APIClient.RateLimitRemaining)
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
