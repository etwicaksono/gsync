package category_app

import (
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	port "github.com/etwicaksono/data-transaksi-service/internal/ports/app"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/secondary/db"
	"gorm.io/gorm"
)

// adapter is a concrete implementation of port.CategoryAppInterface.
type adapter struct {
	cfg        config.Config
	dbClient   *gorm.DB
	categoryDb db.CategoryDbInterface
}

func NewCategoryApp(
	cfg config.Config,
	dbClient *gorm.DB,
	categoryDb db.CategoryDbInterface,
) port.CategoryAppInterface {
	return &adapter{
		cfg:        cfg,
		dbClient:   dbClient,
		categoryDb: categoryDb,
	}
}
