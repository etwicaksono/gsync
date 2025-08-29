package account_db

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

func (a adapter) BulkUpsert(ctx context.Context, tx *gorm.DB, request []entity.Account, truncateFirst bool) (code int, err error) {
	upsertDataModels := make([]model.Account, len(request))
	for i, item := range request {
		upsertDataModels[i] = model.Account{}.FromEntity(item)
	}

	if truncateFirst {
		err = tx.Exec(fmt.Sprintf("TRUNCATE TABLE %s", model.Account{}.TableName())).Error
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("error on truncate %s", model.Account{}.TableName()), slog.String(constants.Error, err.Error()))
			return
		}
	}

	err = tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"code", "name", "active", "usability", "ownership", "position", "updated_at", "updated_by"}),
	}).CreateInBatches(upsertDataModels, 1000).Error
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("error in BulkUpsert %s", model.Account{}.TableName()), slog.String(constants.Error, err.Error()))
		code = http.StatusInternalServerError
		return
	}

	code = http.StatusOK
	return
}
