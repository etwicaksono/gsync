package debt_db

import (
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/secondary/db"
)

type adapter struct {
	cfg config.Config
}

func NewDebtDb(
	cfg config.Config,
) db.DebtDbInterface {
	return &adapter{
		cfg: cfg,
	}
}
