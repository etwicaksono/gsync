package transfer_db

import (
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/secondary/db"
)

type adapter struct {
	cfg config.Config
}

func NewTransferDb(
	cfg config.Config,
) db.TransferDbInterface {
	return &adapter{
		cfg: cfg,
	}
}
