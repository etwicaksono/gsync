package finance_app

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/util"
	"github.com/etwicaksono/data-transaksi-service/internal/constants"
	"github.com/guregu/null"
)

func (a adapter) PullCategory(ctx context.Context, truncateFirst bool) (err error) {
	// Context with timeout
	ctx, cancel := context.WithTimeout(ctx, a.cfg.CliApp.CliTimeout)
	defer cancel()

	// Sheets service
	srv := util.MustSheetsService(ctx, a.cfg.GsheetDataTransaksi.CredentialPath)

	// Read with positions (FIXED for ranges like Y5:AL20 and quoted sheet names)
	srcGrid, err := util.ReadSheetWithPos(ctx, srv, a.cfg.GsheetDataTransaksi.Id, a.cfg.GsheetDataTransaksi.CategoryRange, 6)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read sheet", slog.String(constants.Error, err.Error()))
		return err
	}
	totalCells := 0
	for _, row := range srcGrid {
		totalCells += len(row)
	}
	slog.InfoContext(ctx, fmt.Sprintf("Read %d row(s), %d cell(s) from %s %q", len(srcGrid), totalCells, a.cfg.GsheetDataTransaksi.Id, a.cfg.GsheetDataTransaksi.TransactionRange))

	authData, err := util.GetAuthData(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get auth data", slog.String(constants.Error, err.Error()))
		return
	}
	// save data to db
	var upsertData = make([]entity.Category, 0)
	for _, row := range srcGrid {
		if len(row) == 0 {
			continue
		}

		if util.NonEmpty(row, 0, 1, 2, 3, 4, 5) {
			idParsed := row[0]
			parentParsed := row[1]
			nameParsed := row[2]
			codeParsed := row[3]
			natureParsed := row[4]
			visibilityParsed := row[5]

			id, _ := strconv.ParseInt(strings.TrimSpace(idParsed.Val), 10, 64)

			upsertData = append(upsertData, entity.Category{
				ID:         id,
				Parent:     strings.TrimSpace(parentParsed.Val),
				Name:       strings.TrimSpace(nameParsed.Val),
				Code:       strings.TrimSpace(codeParsed.Val),
				Nature:     strings.TrimSpace(natureParsed.Val),
				Visibility: strings.TrimSpace(visibilityParsed.Val),
				Position: &entity.CategoryPosition{
					ID:         idParsed.Cell,
					Parent:     parentParsed.Cell,
					Name:       nameParsed.Cell,
					Code:       codeParsed.Cell,
					Nature:     natureParsed.Cell,
					Visibility: visibilityParsed.Cell,
				},
				TimeStamps: entity.TimeStamps{
					CreatedAt: time.Now(),
					CreatedBy: null.StringFrom(authData.Name),
					UpdatedAt: null.TimeFrom(time.Now()),
					UpdatedBy: null.StringFrom(authData.Name),
				},
			})
		}
	}

	tx := a.dbClient.WithContext(ctx).Begin()
	defer tx.Rollback()
	_, err = a.categoryDb.BulkUpsert(ctx, tx, upsertData, truncateFirst)
	if err != nil {
		slog.ErrorContext(ctx, "failed to bulk upsert category", slog.String(constants.Error, err.Error()))
		return err
	}

	if err = tx.Commit().Error; err != nil {
		slog.ErrorContext(ctx, "failed to commit transactions", slog.String(constants.Error, err.Error()))
		return err
	}

	return
}
