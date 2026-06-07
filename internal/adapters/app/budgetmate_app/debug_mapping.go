package budgetmate_app

import (
	"context"
	"fmt"
)

func (a *adapter) DebugMapping(ctx context.Context) error {
	// Fetch transactions
	fmt.Println("Fetching transactions from Google Sheet...")
	txResp, err := a.appsClient.FetchTransactions(ctx, 0, 100)
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
	if a.budgetmateClient != nil {
		fmt.Println("\nLogging in to BudgetMate API...")
		if err := a.budgetmateClient.Login(ctx); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		accounts, _ := a.budgetmateClient.FetchAccounts(ctx)
		categories, _ := a.budgetmateClient.FetchCategories(ctx)

		fmt.Printf("\n=== Accounts in BudgetMate ===\n")
		for _, acc := range accounts {
			fmt.Printf("  - %s (ID: %s)\n", acc.Name, acc.ID)
		}

		fmt.Printf("\n=== Categories in BudgetMate ===\n")
		for _, cat := range categories {
			fmt.Printf("  - %s (ID: %s)\n", cat.Name, cat.ID)
		}
	}

	return nil
}
