package util

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/constants"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

/* ===== Sheets service ===== */

func MustSheetsService(ctx context.Context, credsPath string) *sheets.Service {
	if credsPath == "" {
		log.Fatal("missing credentials: set --creds or env GOOGLE_APPLICATION_CREDENTIALS")
	}
	b, err := os.ReadFile(credsPath)
	if err != nil {
		log.Fatalf("read creds: %v", err)
	}
	type saMinimal struct {
		Type        string `json:"type"`
		ClientEmail string `json:"client_email"`
	}
	var sa saMinimal
	_ = json.Unmarshal(b, &sa)
	if sa.Type != "service_account" {
		log.Fatalf("the provided JSON is not a service account key (type=%q). Download a Service Account key JSON.", sa.Type)
	}
	log.Printf("Using service account: %s", sa.ClientEmail)

	srv, err := sheets.NewService(
		ctx,
		option.WithCredentialsFile(credsPath),
		option.WithScopes(sheets.SpreadsheetsScope), // read + write
	)
	if err != nil {
		log.Fatalf("create sheets service: %v", err)
	}
	return srv
}

/* ===== Reading (FIXED range parsing) ===== */

// ReadSheetWithPos reads a range and returns [][]ReadData with absolute positions.
// It correctly handles ranges like "My Sheet!Y5:AL20" by using the upper-left cell (Y5)
// and quotes the sheet name when building A1 addresses.
func ReadSheetWithPos(ctx context.Context, srv *sheets.Service, spreadsheetID, a1Range string, expectedCols int) ([][]entity.ReadData, error) {
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, a1Range).ValueRenderOption("UNFORMATTED_VALUE").Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	sheetTitle, startCol, startRow := ParseA1RangeStart(a1Range) // robust start-cell extraction

	out := make([][]entity.ReadData, len(resp.Values))
	for i, r := range resp.Values {
		// pad to expectedCols
		row := make([]entity.ReadData, expectedCols)
		for j := 0; j < expectedCols; j++ {
			var v interface{} = ""
			if j < len(r) {
				v = r[j] // take actual value when present
			}
			val := fmt.Sprint(v)

			absRow := startRow + i
			absCol := startCol + j
			a1 := fmt.Sprintf("%s!%s%d", QuoteSheet(sheetTitle), ColToLetter(absCol), absRow)

			row[j] = entity.ReadData{
				Val: val,
				Cell: entity.Cell{
					Row: absRow,
					Col: absCol,
					A1:  a1,
				},
			}
		}
		out[i] = row
	}
	return out, nil
}

/* ===== Utilities (range parsing, quoting, etc.) ===== */

// ParseA1RangeStart returns the sheet title and the numeric start column+row
// from an A1 range like "Sheet!Y5:AL20" (→ Y5). It accepts quoted sheet names.
func ParseA1RangeStart(a1 string) (sheet string, col int, row int) {
	sheet, rest := SplitRangeRaw(a1) // e.g. "Categories & Budget Mapping", "Y5:AL20"
	sheet = UnquoteSheet(sheet)

	if rest == "" {
		return sheet, 1, 1
	}

	// Keep upper-left cell if a range provided (before ':')
	first := rest
	if idx := strings.Index(rest, ":"); idx != -1 {
		first = rest[:idx]
	}

	c, r, err := ParseA1Cell(first)
	if err != nil || c <= 0 || r <= 0 {
		// Fallback for column-only/row-only ranges like "Y:AL" or "5:20"
		return sheet, 1, 1
	}
	return sheet, c, r
}

// SplitRangeRaw splits "<Sheet>!<rest>" without validating <rest>.
func SplitRangeRaw(a1 string) (sheetTitle, rest string) {
	parts := strings.SplitN(a1, "!", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func ColToLetter(n int) string {
	res := ""
	for n > 0 {
		n--
		res = string(rune('A'+(n%26))) + res
		n /= 26
	}
	return res
}

var a1CellRe = regexp.MustCompile(`^([A-Za-z]+)(\d+)$`)

func ParseA1Cell(cell string) (col int, row int, err error) {
	if cell == "" {
		return 1, 1, nil
	}
	m := a1CellRe.FindStringSubmatch(cell)
	if m == nil {
		return 0, 0, fmt.Errorf("invalid A1 cell: %q", cell)
	}
	col = 0
	for _, c := range strings.ToUpper(m[1]) {
		col = col*26 + int(c-'A'+1)
	}
	row, err = strconv.Atoi(m[2])
	return
}

// If the sheet title is quoted ('My Sheet'), remove quotes and unescape.
func UnquoteSheet(title string) string {
	if len(title) >= 2 && strings.HasPrefix(title, "'") && strings.HasSuffix(title, "'") {
		inner := title[1 : len(title)-1]
		inner = strings.ReplaceAll(inner, "''", "'")
		return inner
	}
	return title
}

// Quote sheet names with spaces/specials and escape internal single quotes.
func QuoteSheet(title string) string {
	if strings.ContainsAny(title, " '!:,") {
		return "'" + strings.ReplaceAll(title, "'", "''") + "'"
	}
	return title
}

func FromSheetSerialDate(n float64) time.Time {
	// Excel/Sheets epoch starts at 1899-12-30
	base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	return base.Add(time.Duration(n*24) * time.Hour)
}

// helper: ensure required cells are non-empty
func NonEmpty(row []entity.ReadData, idx ...int) bool {
	for _, i := range idx {
		if i >= len(row) || strings.TrimSpace(row[i].Val) == "" {
			return false
		}
	}
	return true
}

func EnsureTargetSheet(ctx context.Context, srv *sheets.Service, spreadsheetID, sheetTitle string, createIfMissing bool) (int64, error) {
	ss, err := srv.Spreadsheets.Get(spreadsheetID).
		Fields("sheets(properties(title,sheetId))").
		Context(ctx).Do()
	if err != nil {
		slog.ErrorContext(ctx, "failed to read target spreadsheet", slog.String(constants.Error, err.Error()))
		return 0, fmt.Errorf("cannot open target spreadsheet (bad ID or no permission): %w", err)
	}
	for _, sh := range ss.Sheets {
		if sh.Properties != nil && sh.Properties.Title == sheetTitle {
			return sh.Properties.SheetId, nil
		}
	}
	if !createIfMissing {
		return 0, fmt.Errorf("sheet/tab %q not found in target", sheetTitle)
	}
	req := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{Title: sheetTitle},
			}},
		},
	}
	resp, err := srv.Spreadsheets.BatchUpdate(spreadsheetID, req).Context(ctx).Do()
	if err != nil {
		slog.ErrorContext(ctx, "failed to create target sheet", slog.String(constants.Error, err.Error()))
		return 0, fmt.Errorf("failed to create sheet %q: %w", sheetTitle, err)
	}
	for _, r := range resp.Replies {
		if r.AddSheet != nil && r.AddSheet.Properties != nil && r.AddSheet.Properties.Title == sheetTitle {
			return r.AddSheet.Properties.SheetId, nil
		}
	}
	return 0, fmt.Errorf("sheet %q creation did not return a sheetId", sheetTitle)
}

// ReadSheetRaw is a simple [][]string reader for places we don't need positions.
func ReadSheetRaw(ctx context.Context, srv *sheets.Service, spreadsheetID, a1Range string) ([][]string, error) {
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, a1Range).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	vals := make([][]string, len(resp.Values))
	for i, row := range resp.Values {
		vals[i] = make([]string, len(row))
		for j, cell := range row {
			vals[i][j] = fmt.Sprint(cell)
		}
	}
	return vals, nil
}

func GridValues(grid [][]entity.ReadData) [][]string {
	out := make([][]string, len(grid))
	for i := range grid {
		out[i] = make([]string, len(grid[i]))
		for j := range grid[i] {
			out[i][j] = grid[i][j].Val
		}
	}
	return out
}

func FindHeaderIndex(header []string, name string) int {
	name = strings.TrimSpace(strings.ToLower(name))
	for i, h := range header {
		if strings.ToLower(strings.TrimSpace(h)) == name {
			return i + 1 // 1-based
		}
	}
	return 0
}

func SafeCell(row []string, oneBasedIdx int) string {
	i := oneBasedIdx - 1
	if i < 0 || i >= len(row) {
		return ""
	}
	return row[i]
}

func MaxRowWidth(rows [][]string) int {
	w := 0
	for _, r := range rows {
		if len(r) > w {
			w = len(r)
		}
	}
	if w == 0 {
		w = 1
	}
	return w
}

func Btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
