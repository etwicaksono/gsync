package budgetmate_app

import (
	"fmt"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/appsscript"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/budgetmateapi"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/app"
)

type adapter struct {
	appsClient       *appsscript.Client
	budgetmateClient *budgetmateapi.Client
}

func NewBudgetMateApp(appsClient *appsscript.Client, budgetmateClient *budgetmateapi.Client) app.BudgetMateAppInterface {
	return &adapter{
		appsClient:       appsClient,
		budgetmateClient: budgetmateClient,
	}
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
	for _, format := range formats {
		t, err = time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
