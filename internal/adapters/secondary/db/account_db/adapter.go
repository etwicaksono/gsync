package account_db

import (
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/secondary/db"
)

type adapter struct {
	cfg config.Config
}

func NewAccountDb(
	cfg config.Config,
) db.AccountDbInterface {
	return &adapter{
		cfg: cfg,
	}
}
