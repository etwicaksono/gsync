package db

import (
	"context"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"gorm.io/gorm"
)

type AccountDbInterface interface {
	BulkUpsert(ctx context.Context, tx *gorm.DB, request []entity.Account, truncateFirst bool) (code int, err error)
}
