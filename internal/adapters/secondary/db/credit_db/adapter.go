package credit_db

import (
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/secondary/db"
)

type adapter struct {
	cfg config.Config
}

func NewCreditDb(
	cfg config.Config,
) db.CreditDbInterface {
	return &adapter{
		cfg: cfg,
	}
}
