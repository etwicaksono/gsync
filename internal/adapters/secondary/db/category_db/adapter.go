package category_db

import (
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/secondary/db"
)

type adapter struct {
	cfg config.Config
}

func NewCategoryDb(
	cfg config.Config,
) db.CategoryDbInterface {
	return &adapter{
		cfg: cfg,
	}
}
