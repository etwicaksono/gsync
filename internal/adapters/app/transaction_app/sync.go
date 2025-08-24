package transaction_app

import (
	"context"
	"log"
)

/* ===== Types ===== */

func (a adapter) Sync(ctx context.Context) (err error) {
	// Context with timeout
	ctx, cancel := context.WithTimeout(ctx, a.cfg.CliApp.CliTimeout)
	defer cancel()

	// Sheets service
	srv := MustSheetsService(ctx, a.cfg.Gsheet.CredentialPath)

	// Read with positions (FIXED for ranges like Y5:AL20 and quoted sheet names)
	srcGrid, err := ReadSheetWithPos(ctx, srv, a.cfg.Gsheet.Id, a.cfg.Gsheet.TransactionRange)
	if err != nil {
		log.Fatalf("failed to read source sheet: %v", err)
	}
	totalCells := 0
	for _, row := range srcGrid {
		totalCells += len(row)
	}
	log.Printf("Read %d row(s), %d cell(s) from %s %q", len(srcGrid), totalCells, a.cfg.Gsheet.Id, a.cfg.Gsheet.TransactionRange)

	// TODO: implement to save data to db

	return
}
