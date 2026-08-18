package budgetbakers_app

import (
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/appsscript"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/budgetbakers"
	port "github.com/etwicaksono/data-transaksi-service/internal/ports/app"
)

type adapter struct {
	appsClient *appsscript.Client
	bbClient   *budgetbakers.Client

	// In-memory cache for reference data
	accounts   []budgetbakers.Account
	categories []budgetbakers.Category
	labels     []budgetbakers.Label
}

func NewBudgetBakersApp(appsClient *appsscript.Client, bbClient *budgetbakers.Client) port.BudgetBakersAppInterface {
	return &adapter{
		appsClient: appsClient,
		bbClient:   bbClient,
	}
}
