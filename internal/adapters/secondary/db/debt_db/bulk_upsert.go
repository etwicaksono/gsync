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

func (a adapter) BulkUpsert(ctx context.Context, tx *gorm.DB, request []entity.Transfer, truncateFirst bool) (code int, err error) {
	upsertDataModels := make([]model.Transfer, len(request))
	for i, item := range request {
		upsertDataModels[i] = model.Transfer{}.FromEntity(item)
	}

	if truncateFirst {
		err = tx.Exec(fmt.Sprintf("TRUNCATE TABLE %s", model.Transfer{}.TableName())).Error
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("error on truncate %s", model.Transfer{}.TableName()), slog.String(constants.Error, err.Error()))
			return
		}
	}

	err = tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"date", "from_account", "to_account", "amount", "note", "position", "updated_at", "updated_by"}),
	}).CreateInBatches(upsertDataModels, 1000).Error
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("error in BulkUpsert %s", model.Transfer{}.TableName()), slog.String(constants.Error, err.Error()))
		code = http.StatusInternalServerError
		return
	}

	code = http.StatusOK
	return
}
