package transaction_db

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/model"
	"github.com/etwicaksono/data-transaksi-service/internal/constants"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (a adapter) BulkUpsert(ctx context.Context, tx *gorm.DB, request []entity.Transaction) (code int, err error) {
	upsertDataModels := make([]model.Transaction, len(request))
	for i, item := range request {
		upsertDataModels[i] = model.Transaction{}.FromEntity(item)
	}

	err = tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"date", "account", "category", "amount", "note", "type", "position", "updated_at", "updated_by"}),
	}).CreateInBatches(upsertDataModels, 1000).Error
	if err != nil {
		slog.ErrorContext(ctx, "error in BulkUpsert", slog.String(constants.Error, err.Error()))
		code = http.StatusInternalServerError
		return
	}

	code = http.StatusOK
	return
}
