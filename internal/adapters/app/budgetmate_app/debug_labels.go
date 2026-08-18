package budgetmate_app

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// DebugLabels compares the labels coming out of the Google Sheet against the
// labels that exist in BudgetMate, and reports the outgoing payload for the
// first labelled transaction. It answers the three questions that a silently
// unlabelled sync raises: does the sheet send labels, does the API return any,
// and do the two sides actually match.
func (a *adapter) DebugLabels(ctx context.Context) error {
	fmt.Println("Fetching transactions from Google Sheet...")
	txResp, err := a.appsClient.FetchTransactions(ctx, 0, 100)
	if err != nil {
		return fmt.Errorf("fetch transactions: %w", err)
	}
	fmt.Printf("Fetched %d transaction(s)\n", len(txResp.Data))

	// Step 1: does the sheet send labels at all?
	labelled := 0
	sheetLabels := make(map[string]int)
	for _, tx := range txResp.Data {
		if len(tx.Labels) == 0 {
			continue
		}
		labelled++
		for _, name := range tx.Labels {
			trimmed := strings.TrimSpace(name)
			if trimmed != "" {
				sheetLabels[trimmed]++
			}
		}
	}

	fmt.Printf("\n=== Step 1: labels arriving from the sheet ===\n")
	fmt.Printf("%d of %d transaction(s) carry at least one label\n", labelled, len(txResp.Data))
	if len(sheetLabels) == 0 {
		fmt.Printf("\nThe sheet sent no labels at all. The payload never contained a label,\n")
		fmt.Printf("so nothing could reach BudgetMate. Most likely cause:\n")
		fmt.Printf("  - The Apps Script deployment is older than the label change, so it\n")
		fmt.Printf("    never adds the labels field, or\n")
		fmt.Printf("  - Column I of the Transaction sheet is empty\n")
		fmt.Printf("\nRedeploy scripts/budgetbakers/google-apps-script.js and re-run this command.\n")
	} else {
		for _, name := range sortedKeys(sheetLabels) {
			fmt.Printf("  - %-24q used by %d transaction(s)\n", name, sheetLabels[name])
		}
	}

	if a.budgetmateClient == nil {
		fmt.Printf("\nBudgetMate credentials not configured, stopping after the sheet check.\n")
		return nil
	}

	// Step 2: what does the API know about?
	fmt.Println("\nLogging in to BudgetMate API...")
	if err := a.budgetmateClient.Login(ctx); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	a.budgetmateClient.SetDebug(true)
	apiLabels, err := a.budgetmateClient.FetchLabels(ctx)
	a.budgetmateClient.SetDebug(false)
	if err != nil {
		return fmt.Errorf("fetch labels: %w", err)
	}

	fmt.Printf("\n=== Step 2: labels available in BudgetMate ===\n")
	if len(apiLabels) == 0 {
		fmt.Printf("None. Create the labels in BudgetMate first, otherwise every labelled\n")
		fmt.Printf("transaction will stop the sync with a \"label not found\" error.\n")
	}
	for _, label := range apiLabels {
		fmt.Printf("  - %-24q id=%s\n", label.Name, label.ID)
	}

	// Step 3: do they line up?
	fmt.Printf("\n=== Step 3: match ===\n")
	unmatched := make([]string, 0, len(sheetLabels))
	for _, name := range sortedKeys(sheetLabels) {
		id, ok := a.budgetmateClient.GetLabelIDWithMapping(name)
		if ok {
			fmt.Printf("  OK       %-24q -> %s\n", name, id)
			continue
		}
		unmatched = append(unmatched, name)
		fmt.Printf("  MISSING  %-24q no label of that name in BudgetMate\n", name)
	}

	if len(unmatched) > 0 {
		fmt.Printf("\n%d sheet label(s) have no counterpart in BudgetMate. The push will stop\n", len(unmatched))
		fmt.Printf("at the first transaction using one of them. Either create them in\n")
		fmt.Printf("BudgetMate or map them with SetLabelMapping.\n")
	}

	// Step 4: show the payload that would actually be sent.
	fmt.Printf("\n=== Step 4: outgoing payload preview ===\n")
	previewed := false
	for _, tx := range txResp.Data {
		if len(tx.Labels) == 0 {
			continue
		}
		ids, missing := a.budgetmateClient.ResolveLabelIDs(tx.Labels)
		fmt.Printf("Transaction #%s\n", tx.No)
		fmt.Printf("  sheet labels : %v\n", tx.Labels)
		fmt.Printf("  resolved ids : %v\n", ids)
		if len(missing) > 0 {
			fmt.Printf("  unresolved   : %v (this transaction would abort the push)\n", missing)
		}
		if len(ids) == 0 && len(missing) == 0 {
			fmt.Printf("  note         : resolved to nothing, so the labels field is omitted\n")
		}
		previewed = true
		break
	}
	if !previewed {
		fmt.Printf("No labelled transaction available to preview.\n")
	}

	if labelled > 0 && len(unmatched) == 0 {
		fmt.Printf("\nSheet and API agree. If labels still do not appear in BudgetMate after a\n")
		fmt.Printf("push, the IDs are being sent but not stored: re-run the push and compare\n")
		fmt.Printf("the logged request body against what the UI shows.\n")
	}

	return nil
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
