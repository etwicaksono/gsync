package db

import (
	"context"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"gorm.io/gorm"
)

type TransactionDbInterface interface {
	BulkInsert(ctx context.Context, tx *gorm.DB, request []entity.Transaction) (code int, err error)
}
