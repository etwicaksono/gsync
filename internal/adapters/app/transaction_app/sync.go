package transaction_app

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/constants"
	"github.com/guregu/null"
)

func (a adapter) Sync(ctx context.Context) (err error) {
	// Context with timeout
	ctx, cancel := context.WithTimeout(ctx, a.cfg.CliApp.CliTimeout)
	defer cancel()

	// Sheets service
	srv := MustSheetsService(ctx, a.cfg.Gsheet.CredentialPath)

	// Read with positions (FIXED for ranges like Y5:AL20 and quoted sheet names)
	srcGrid, err := ReadSheetWithPos(ctx, srv, a.cfg.Gsheet.Id, a.cfg.Gsheet.TransactionRange)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read sheet", slog.String(constants.Error, err.Error()))
		return err
	}
	totalCells := 0
	for _, row := range srcGrid {
		totalCells += len(row)
	}
	slog.InfoContext(ctx, fmt.Sprintf("Read %d row(s), %d cell(s) from %s %q", len(srcGrid), totalCells, a.cfg.Gsheet.Id, a.cfg.Gsheet.TransactionRange))

	// save data to db
	var upsertData = make([]entity.Transaction, 0)
	for _, row := range srcGrid {
		if len(row) == 0 {
			continue
		}

		if nonEmpty(row, 0, 1, 2, 3, 4, 6) {
			idParsed := row[0]
			dateParsed := row[1]
			accountParsed := row[2]
			categoryParsed := row[3]
			amountParsed := row[4]
			noteParsed := row[5]
			typeParsed := row[6]

			id, _ := strconv.ParseInt(strings.TrimSpace(idParsed.Val), 10, 64)
			dateFloat, _ := strconv.ParseFloat(strings.TrimSpace(dateParsed.Val), 64)
			amountFloat, _ := strconv.ParseFloat(strings.TrimSpace(amountParsed.Val), 64)

			upsertData = append(upsertData, entity.Transaction{
				ID:       id,
				Date:     fromSheetSerialDate(dateFloat),
				Account:  strings.TrimSpace(accountParsed.Val),
				Category: strings.TrimSpace(categoryParsed.Val),
				Amount:   amountFloat,
				Note:     strings.TrimSpace(noteParsed.Val),
				Type:     strings.TrimSpace(typeParsed.Val),
				Position: &entity.TransactionPosition{
					ID:       idParsed.Cell,
					Date:     dateParsed.Cell,
					Account:  accountParsed.Cell,
					Category: categoryParsed.Cell,
					Amount:   amountParsed.Cell,
					Note:     noteParsed.Cell,
					Type:     typeParsed.Cell,
				},
				TimeStamps: entity.TimeStamps{
					CreatedAt: time.Now(),
					CreatedBy: null.StringFrom("cli"), // TODO: retrieve from context
					UpdatedAt: null.TimeFrom(time.Now()),
					UpdatedBy: null.StringFrom("cli"), // TODO: retrieve from context
				},
			})
		}
	}

	tx := a.dbClient.WithContext(ctx).Begin()
	defer tx.Rollback()
	_, err = a.transactionDb.BulkUpsert(ctx, tx, upsertData)
	if err != nil {
		slog.ErrorContext(ctx, "failed to bulk upsert transactions: %v", err)
		return err
	}

	if err = tx.Commit().Error; err != nil {
		slog.ErrorContext(ctx, "failed to commit transactions: %v", err)
		return err
	}

	return
}

func fromSheetSerialDate(n float64) time.Time {
	// Excel/Sheets epoch starts at 1899-12-30
	base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	return base.Add(time.Duration(n*24) * time.Hour)
}

// helper: ensure required cells are non-empty
func nonEmpty(row []entity.ReadData, idx ...int) bool {
	for _, i := range idx {
		if i >= len(row) || strings.TrimSpace(row[i].Val) == "" {
			return false
		}
	}
	return true
}
