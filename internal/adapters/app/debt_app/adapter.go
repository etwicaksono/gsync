package debt_app

import (
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	port "github.com/etwicaksono/data-transaksi-service/internal/ports/app"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/secondary/db"
	"gorm.io/gorm"
)

// adapter is a concrete implementation of port.DebtAppInterface.
type adapter struct {
	cfg      config.Config
	dbClient *gorm.DB
	debtDb   db.DebtDbInterface
}

func NewDebtApp(
	cfg config.Config,
	dbClient *gorm.DB,
	debtDb db.DebtDbInterface,
) port.DebtAppInterface {
	return &adapter{
		cfg:      cfg,
		dbClient: dbClient,
		debtDb:   debtDb,
	}
}
