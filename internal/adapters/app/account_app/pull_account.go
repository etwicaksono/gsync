package account_app

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

func (a adapter) PullAccount(ctx context.Context, truncateFirst bool) (err error) {
	// Context with timeout
	ctx, cancel := context.WithTimeout(ctx, a.cfg.CliApp.CliTimeout)
	defer cancel()

	// Sheets service
	srv := util.MustSheetsService(ctx, a.cfg.Gsheet.CredentialPath)

	// Read with positions (FIXED for ranges like Y5:AL20 and quoted sheet names)
	srcGrid, err := util.ReadSheetWithPos(ctx, srv, a.cfg.Gsheet.Id, a.cfg.Gsheet.AccountRange, 12)
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
	var upsertData = make([]entity.Account, 0)
	for _, row := range srcGrid {
		if len(row) == 0 {
			continue
		}

		if util.NonEmpty(row, 0, 1, 2, 3) {
			idParsed := row[0]
			usabilityParsed := row[1]
			codeParsed := row[2]
			nameParsed := row[3]
			ownershipParsed := row[10]
			activeParsed := row[11]

			id, _ := strconv.ParseInt(strings.TrimSpace(idParsed.Val), 10, 64)

			upsertData = append(upsertData, entity.Account{
				ID:        id,
				Code:      strings.TrimSpace(codeParsed.Val),
				Name:      strings.TrimSpace(nameParsed.Val),
				Active:    strings.TrimSpace(activeParsed.Val),
				Usability: strings.TrimSpace(usabilityParsed.Val),
				Ownership: strings.TrimSpace(ownershipParsed.Val),
				Position: &entity.AccountPosition{
					ID:        idParsed.Cell,
					Code:      codeParsed.Cell,
					Name:      nameParsed.Cell,
					Active:    activeParsed.Cell,
					Usability: usabilityParsed.Cell,
					Ownership: ownershipParsed.Cell,
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
	_, err = a.accountDb.BulkUpsert(ctx, tx, upsertData, truncateFirst)
	if err != nil {
		slog.ErrorContext(ctx, "failed to bulk upsert account", slog.String(constants.Error, err.Error()))
		return err
	}

	if err = tx.Commit().Error; err != nil {
		slog.ErrorContext(ctx, "failed to commit transactions", slog.String(constants.Error, err.Error()))
		return err
	}

	return
}
