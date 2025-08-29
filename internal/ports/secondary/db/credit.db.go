package db

import (
	"context"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"gorm.io/gorm"
)

type CreditDbInterface interface {
	BulkUpsert(ctx context.Context, tx *gorm.DB, request []entity.Credit, truncateFirst bool) (code int, err error)
}
