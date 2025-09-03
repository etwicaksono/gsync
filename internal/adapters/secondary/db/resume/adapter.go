package resume

import (
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/secondary/db"
)

type adapter struct {
	cfg config.Config
}

func NewResumeDb(
	cfg config.Config,
) db.ResumeDbInterface {
	return &adapter{
		cfg: cfg,
	}
}
