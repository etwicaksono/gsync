package gsheet_finance_app

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

func (a adapter) PullDebt(ctx context.Context, truncateFirst bool) (err error) {
	// Context with timeout
	ctx, cancel := context.WithTimeout(ctx, a.cfg.CliApp.CliTimeout)
	defer cancel()

	// Sheets service
	srv := util.MustSheetsService(ctx, a.cfg.GsheetDataTransaksi.CredentialPath)

	// Read with positions (FIXED for ranges like Y5:AL20 and quoted sheet names)
	srcGrid, err := util.ReadSheetWithPos(ctx, srv, a.cfg.GsheetDataTransaksi.Id, a.cfg.GsheetDataTransaksi.DebtRange, 4)
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
	var upsertData = make([]entity.Debt, 0)
	for _, row := range srcGrid {
		if len(row) == 0 {
			continue
		}

		if util.NonEmpty(row, 0, 1, 2, 3) {
			idParsed := row[0]
			accountParsed := row[1]
			nameParsed := row[2]
			amountParsed := row[3]

			id, _ := strconv.ParseInt(strings.TrimSpace(idParsed.Val), 10, 64)
			amountFloat, _ := strconv.ParseFloat(strings.TrimSpace(amountParsed.Val), 64)

			upsertData = append(upsertData, entity.Debt{
				ID:      id,
				Account: strings.TrimSpace(accountParsed.Val),
				Name:    strings.TrimSpace(nameParsed.Val),
				Amount:  amountFloat,
				Position: &entity.DebtPosition{
					ID:      idParsed.Cell,
					Account: accountParsed.Cell,
					Name:    nameParsed.Cell,
					Amount:  amountParsed.Cell,
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
	_, err = a.debtDb.BulkUpsert(ctx, tx, upsertData, truncateFirst)
	if err != nil {
		slog.ErrorContext(ctx, "failed to bulk upsert debt", slog.String(constants.Error, err.Error()))
		return err
	}

	if err = tx.Commit().Error; err != nil {
		slog.ErrorContext(ctx, "failed to commit transactions", slog.String(constants.Error, err.Error()))
		return err
	}

	return
}
