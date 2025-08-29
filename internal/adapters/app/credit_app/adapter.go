package credit_app

import (
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	port "github.com/etwicaksono/data-transaksi-service/internal/ports/app"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/secondary/db"
	"gorm.io/gorm"
)

// adapter is a concrete implementation of port.CreditAppInterface.
type adapter struct {
	cfg      config.Config
	dbClient *gorm.DB
	creditDb db.CreditDbInterface
}

func NewCreditApp(
	cfg config.Config,
	dbClient *gorm.DB,
	creditDb db.CreditDbInterface,
) port.CreditAppInterface {
	return &adapter{
		cfg:      cfg,
		dbClient: dbClient,
		creditDb: creditDb,
	}
}
