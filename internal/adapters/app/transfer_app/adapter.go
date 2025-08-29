package transfer_app

import (
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	port "github.com/etwicaksono/data-transaksi-service/internal/ports/app"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/secondary/db"
	"gorm.io/gorm"
)

// adapter is a concrete implementation of port.TransferAppInterface.
type adapter struct {
	cfg        config.Config
	dbClient   *gorm.DB
	transferDb db.TransferDbInterface
}

func NewTransferApp(
	cfg config.Config,
	db db.TransferDbInterface,
	dbClient *gorm.DB,
) port.TransferAppInterface {
	return &adapter{
		cfg:        cfg,
		dbClient:   dbClient,
		transferDb: db,
	}
}
