package debt_db

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/model"
	"github.com/etwicaksono/data-transaksi-service/internal/constants"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (a adapter) BulkUpsert(ctx context.Context, tx *gorm.DB, request []entity.Debt, truncateFirst bool) (code int, err error) {
	upsertDataModels := make([]model.Debt, len(request))
	for i, item := range request {
		upsertDataModels[i] = model.Debt{}.FromEntity(item)
	}

	if truncateFirst {
		err = tx.Exec(fmt.Sprintf("DELETE FROM %s", model.Debt{}.TableName())).Error
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("error on truncate %s", model.Debt{}.TableName()), slog.String(constants.Error, err.Error()))
			return
		}
	}

	err = tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"account", "name", "amount", "position", "updated_at", "updated_by"}),
	}).CreateInBatches(upsertDataModels, 1000).Error
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("error in BulkUpsert %s", model.Debt{}.TableName()), slog.String(constants.Error, err.Error()))
		code = http.StatusInternalServerError
		return
	}

	code = http.StatusOK
	return
}
