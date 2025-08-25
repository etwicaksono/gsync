package transaction_db

import (
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/secondary/db"
)

type adapter struct {
	cfg config.Config
}

func NewTransactionDb(
	cfg config.Config,
) db.TransactionDbInterface {
	return &adapter{
		cfg: cfg,
	}
}
