package util

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
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
func ReadSheetWithPos(ctx context.Context, srv *sheets.Service, spreadsheetID, a1Range string) ([][]entity.ReadData, error) {
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, a1Range).ValueRenderOption("UNFORMATTED_VALUE").Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	sheetTitle, startCol, startRow := ParseA1RangeStart(a1Range) // robust start-cell extraction

	out := make([][]entity.ReadData, len(resp.Values))
	for i, r := range resp.Values {
		row := make([]entity.ReadData, len(r))
		for j, v := range r {
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
		res = string('A'+(n%26)) + res
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
