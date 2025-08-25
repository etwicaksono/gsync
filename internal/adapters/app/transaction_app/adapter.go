package transaction_app

import (
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	port "github.com/etwicaksono/data-transaksi-service/internal/ports/app"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/secondary/db"
	"gorm.io/gorm"
)

// adapter is a concrete implementation of port.TransactionAppInterface.
type adapter struct {
	cfg           config.Config
	dbClient      *gorm.DB
	transactionDb db.TransactionDbInterface
}

func NewTransactionApp(
	cfg config.Config,
	db db.TransactionDbInterface,
	dbClient *gorm.DB,
) port.TransactionAppInterface {
	return &adapter{
		cfg:           cfg,
		transactionDb: db,
	}
}
