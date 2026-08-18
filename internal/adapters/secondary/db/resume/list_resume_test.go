package resume

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/model"
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "resume_test.db")

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Discard,
	})
	require.NoError(t, err)

	// Windows keeps the file locked until the pool is closed, which would make
	// t.TempDir cleanup fail.
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	require.NoError(t, db.AutoMigrate(&model.Category{}, &model.Transaction{}))
	return db
}

func date(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.DateOnly, s)
	require.NoError(t, err)
	return parsed
}

func seed(t *testing.T, db *gorm.DB) {
	t.Helper()

	categories := []model.Category{
		{ID: 1, Parent: "", Name: "Food", Code: "FOOD", Nature: "expense", Visibility: "visible"},
		{ID: 2, Parent: "Food", Name: "Grocery", Code: "GROCERY", Nature: "expense", Visibility: "visible"},
		{ID: 3, Parent: "", Name: "Salary", Code: "SALARY", Nature: "income", Visibility: "visible"},
	}
	require.NoError(t, db.Create(&categories).Error)

	transactions := []model.Transaction{
		{ID: 1, Date: date(t, "2025-01-10"), Account: "BANK", Category: "FOOD", Amount: 100, Note: "n", Type: "expense"},
		{ID: 2, Date: date(t, "2025-01-20"), Account: "BANK", Category: "GROCERY", Amount: 50, Note: "n", Type: "expense"},
		{ID: 3, Date: date(t, "2025-02-05"), Account: "BANK", Category: "FOOD", Amount: 25, Note: "n", Type: "expense"},
		{ID: 4, Date: date(t, "2025-03-01"), Account: "BANK", Category: "SALARY", Amount: 900, Note: "n", Type: "income"},
		// out of range, must be excluded
		{ID: 5, Date: date(t, "2025-04-01"), Account: "BANK", Category: "FOOD", Amount: 777, Note: "n", Type: "expense"},
	}
	require.NoError(t, db.Create(&transactions).Error)
}

func newAdapter() adapter {
	return adapter{cfg: config.Config{}}
}

func toMap(points []entity.MonthlyPoint) map[string]float64 {
	out := make(map[string]float64, len(points))
	for _, p := range points {
		out[p.YM] = p.Amt
	}
	return out
}

func TestListResumeDetail(t *testing.T) {
	db := newTestDB(t)
	seed(t, db)

	got, err := newAdapter().ListResumeDetail(context.Background(), db, entity.VResumeDetailListFilter{
		StartDate: date(t, "2025-01-01"),
		EndDate:   date(t, "2025-04-01"),
	})
	require.NoError(t, err)
	require.Len(t, got, 3, "one row per category")

	byCode := map[string]entity.Resume{}
	for _, r := range got {
		byCode[r.Code] = r
		// every category must expose exactly the 3 months in range, in order
		require.Len(t, r.Monthly, 3, "code=%s", r.Code)
		assert.Equal(t, []string{"2025-01", "2025-02", "2025-03"},
			[]string{r.Monthly[0].YM, r.Monthly[1].YM, r.Monthly[2].YM},
			"months must be chronologically ordered for %s", r.Code)
	}

	// detail keeps child categories separate (no parent rollup)
	assert.Equal(t, map[string]float64{"2025-01": 100, "2025-02": 25, "2025-03": 0}, toMap(byCode["FOOD"].Monthly))
	assert.Equal(t, map[string]float64{"2025-01": 50, "2025-02": 0, "2025-03": 0}, toMap(byCode["GROCERY"].Monthly))
	assert.Equal(t, map[string]float64{"2025-01": 0, "2025-02": 0, "2025-03": 900}, toMap(byCode["SALARY"].Monthly))
}

func TestListResumeRecap(t *testing.T) {
	db := newTestDB(t)
	seed(t, db)

	got, err := newAdapter().ListResumeRecap(context.Background(), db, entity.VResumeDetailListFilter{
		StartDate: date(t, "2025-01-01"),
		EndDate:   date(t, "2025-04-01"),
	})
	require.NoError(t, err)

	byCode := map[string]entity.Resume{}
	for _, r := range got {
		byCode[r.Code] = r
		assert.Equal(t, []string{"2025-01", "2025-02", "2025-03"},
			[]string{r.Monthly[0].YM, r.Monthly[1].YM, r.Monthly[2].YM},
			"months must be chronologically ordered for %s", r.Code)
	}

	// GROCERY rolls up into its parent FOOD, so it must not appear on its own
	assert.NotContains(t, byCode, "GROCERY", "child category should be grouped under its parent")

	// FOOD = own 100 + child GROCERY 50 in January
	assert.Equal(t, map[string]float64{"2025-01": 150, "2025-02": 25, "2025-03": 0}, toMap(byCode["FOOD"].Monthly))
	assert.Equal(t, map[string]float64{"2025-01": 0, "2025-02": 0, "2025-03": 900}, toMap(byCode["SALARY"].Monthly))
}

func TestListResumeEmptyDatabase(t *testing.T) {
	db := newTestDB(t)

	detail, err := newAdapter().ListResumeDetail(context.Background(), db, entity.VResumeDetailListFilter{
		StartDate: date(t, "2025-01-01"),
		EndDate:   date(t, "2025-03-01"),
	})
	require.NoError(t, err)
	assert.Empty(t, detail)

	recap, err := newAdapter().ListResumeRecap(context.Background(), db, entity.VResumeDetailListFilter{
		StartDate: date(t, "2025-01-01"),
		EndDate:   date(t, "2025-03-01"),
	})
	require.NoError(t, err)
	assert.Empty(t, recap)
}

func TestListResumeSpansYearBoundary(t *testing.T) {
	db := newTestDB(t)
	seed(t, db)

	got, err := newAdapter().ListResumeDetail(context.Background(), db, entity.VResumeDetailListFilter{
		StartDate: date(t, "2024-12-01"),
		EndDate:   date(t, "2025-02-01"),
	})
	require.NoError(t, err)
	require.NotEmpty(t, got)

	for _, r := range got {
		assert.Equal(t, []string{"2024-12", "2025-01"},
			[]string{r.Monthly[0].YM, r.Monthly[1].YM},
			"recursive month series must cross the year boundary for %s", r.Code)
	}
}
