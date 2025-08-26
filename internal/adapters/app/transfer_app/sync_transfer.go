package transfer_app

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

func (a adapter) SyncTransfer(ctx context.Context, truncateFirst bool) (err error) {
	// Context with timeout
	ctx, cancel := context.WithTimeout(ctx, a.cfg.CliApp.CliTimeout)
	defer cancel()

	// Sheets service
	srv := util.MustSheetsService(ctx, a.cfg.Gsheet.CredentialPath)

	// Read with positions (FIXED for ranges like Y5:AL20 and quoted sheet names)
	srcGrid, err := util.ReadSheetWithPos(ctx, srv, a.cfg.Gsheet.Id, a.cfg.Gsheet.TransferRange, 6)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read sheet", slog.String(constants.Error, err.Error()))
		return err
	}
	totalCells := 0
	for _, row := range srcGrid {
		totalCells += len(row)
	}
	slog.InfoContext(ctx, fmt.Sprintf("Read %d row(s), %d cell(s) from %s %q", len(srcGrid), totalCells, a.cfg.Gsheet.Id, a.cfg.Gsheet.TransactionRange))

	authData, err := util.GetAuthData(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get auth data", slog.String(constants.Error, err.Error()))
		return
	}
	// save data to db
	var upsertData = make([]entity.Transfer, 0)
	for _, row := range srcGrid {
		if len(row) == 0 {
			continue
		}

		if util.NonEmpty(row, 0, 1, 2, 3, 4) {
			idParsed := row[0]
			dateParsed := row[1]
			fromAccountParsed := row[2]
			toAccountParsed := row[3]
			amountParsed := row[4]
			noteParsed := row[5]

			id, _ := strconv.ParseInt(strings.TrimSpace(idParsed.Val), 10, 64)
			dateFloat, _ := strconv.ParseFloat(strings.TrimSpace(dateParsed.Val), 64)
			amountFloat, _ := strconv.ParseFloat(strings.TrimSpace(amountParsed.Val), 64)

			upsertData = append(upsertData, entity.Transfer{
				ID:          id,
				Date:        util.FromSheetSerialDate(dateFloat),
				FromAccount: strings.TrimSpace(fromAccountParsed.Val),
				ToAccount:   strings.TrimSpace(toAccountParsed.Val),
				Amount:      amountFloat,
				Note:        strings.TrimSpace(noteParsed.Val),
				Position: &entity.TransferPosition{
					ID:          idParsed.Cell,
					Date:        dateParsed.Cell,
					FromAccount: fromAccountParsed.Cell,
					ToAccount:   toAccountParsed.Cell,
					Amount:      amountParsed.Cell,
					Note:        noteParsed.Cell,
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
	_, err = a.transferDb.BulkUpsert(ctx, tx, upsertData, truncateFirst)
	if err != nil {
		slog.ErrorContext(ctx, "failed to bulk upsert transfer", slog.String(constants.Error, err.Error()))
		return err
	}

	if err = tx.Commit().Error; err != nil {
		slog.ErrorContext(ctx, "failed to commit transactions", slog.String(constants.Error, err.Error()))
		return err
	}

	return
}
