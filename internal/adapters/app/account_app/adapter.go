package account_app

import (
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	port "github.com/etwicaksono/data-transaksi-service/internal/ports/app"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/secondary/db"
	"gorm.io/gorm"
)

// adapter is a concrete implementation of port.AccountAppInterface.
type adapter struct {
	cfg       config.Config
	dbClient  *gorm.DB
	accountDb db.AccountDbInterface
}

func NewAccountApp(
	cfg config.Config,
	dbClient *gorm.DB,
	accountDb db.AccountDbInterface,
) port.AccountAppInterface {
	return &adapter{
		cfg:       cfg,
		dbClient:  dbClient,
		accountDb: accountDb,
	}
}
