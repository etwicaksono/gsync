package finance_app

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/util"
	"github.com/etwicaksono/data-transaksi-service/internal/constants"
	"google.golang.org/api/sheets/v4"
)

func (a adapter) PushResumeDetail(ctx context.Context, resumeStartColRange string, filter entity.VResumeDetailListFilter) (err error) {
	// Context with timeout
	ctx, cancel := context.WithTimeout(ctx, a.cfg.CliApp.CliTimeout)
	defer cancel()

	tx := a.dbClient.WithContext(ctx)
	resumeDetailList, err := a.vResumeDetailDb.ListResumeDetail(ctx, tx, filter)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list all vResumeDb", slog.String(constants.Error, err.Error()))
		return err
	}

	var header = []string{"ID", "Code"} // setup first row as header
	for _, v := range resumeDetailList[0].Monthly {
		ymLong, err := util.YMToLong(v.YM)
		if err != nil {
			ymLong = v.YM
		}
		header = append(header, ymLong)
	}

	var rowData = [][]string{
		header,
	}
	for _, v := range resumeDetailList {
		var row = []string{strconv.Itoa(int(v.ID)), v.Code}
		for _, m := range v.Monthly {
			row = append(row, fmt.Sprintf("%.2f", m.Amt))
		}
		if len(row) != len(header) {
			slog.ErrorContext(ctx, "malformed row", slog.String(constants.Error, "malformed row"), slog.Any("row", row))
			continue
		}

		rowData = append(rowData, row)
	}

	spreadsheetID := a.cfg.GsheetRekapKeuangan.Id
	sheetTitle, firstCol, firstRow := util.ParseA1RangeStart(resumeStartColRange)
	startColLetter := util.ColToLetter(firstCol)
	endColLetter := util.ColToLetter(len(rowData[0]) + firstCol - 1)

	srv := util.MustSheetsService(ctx, a.cfg.GsheetRekapKeuangan.CredentialPath)
	if _, err = util.EnsureTargetSheet(ctx, srv, spreadsheetID, sheetTitle, true); err != nil {
		slog.ErrorContext(ctx, "failed to ensure target sheet", slog.String(constants.Error, err.Error()))
		return fmt.Errorf("ensure target sheet: %w", err)
	}

	// 2) Build batch update requests
	var data = []*sheets.ValueRange{
		{
			Range:  fmt.Sprintf("%s!A1", util.QuoteSheet(sheetTitle)),
			Values: [][]interface{}{{fmt.Sprintf("Last Update: %s", time.Now().Format(time.DateTime))}},
		},
	}
	row := firstRow
	for _, c := range rowData {
		if len(c) < len(header) {
			continue // skip malformed rows
		}
		rng := fmt.Sprintf("%s!%s%d:%s%d",
			util.QuoteSheet(sheetTitle), startColLetter, row, endColLetter, row)
		rowVals := make([]interface{}, len(c))
		for i, v := range c {
			rowVals[i] = v
		}
		data = append(data, &sheets.ValueRange{
			Range:          rng,
			Values:         [][]interface{}{rowVals},
			MajorDimension: "ROWS",
		})
		row++
	}

	// 3) Execute batch update
	if len(data) > 0 {
		_, err = srv.Spreadsheets.Values.BatchUpdate(spreadsheetID, &sheets.BatchUpdateValuesRequest{
			ValueInputOption: "USER_ENTERED",
			Data:             data,
		}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("batch update: %w", err)
		}
		log.Printf("Upserted %d rows", len(data))
	} else {
		log.Println("No data to write.")
	}

	return nil
}
