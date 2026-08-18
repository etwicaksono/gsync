package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/db/account_db"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/db/category_db"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/db/credit_db"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/db/debt_db"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/db/transaction_db"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/db/transfer_db"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/model"
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	"github.com/guregu/null"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

func newAdapter(t *testing.T) *AdapterGorm {
	t.Helper()

	cfg := config.Config{
		CliApp: config.CliApp{App: config.App{Name: "gsync-test", LogLevel: "error"}},
		Db: config.DbConfig{
			Path:                  filepath.Join(t.TempDir(), "gsync_test.db"),
			MaxOpenConnections:    10,
			MaxIdleConnections:    2,
			MaxConnectionLifetime: 60,
			MaxConnectionIdleTime: 20,
		},
	}

	a := NewGormDb(context.Background(), cfg, slog.New(slog.NewTextHandler(io_Discard{}, nil)))
	require.NoError(t, a.Connect())

	t.Cleanup(func() {
		if sqlDB, err := a.client.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	return a
}

// io_Discard avoids pulling io/ioutil noise into the import list.
type io_Discard struct{}

func (io_Discard) Write(p []byte) (int, error) { return len(p), nil }

func stamps() entity.TimeStamps {
	now := time.Now().UTC().Truncate(time.Second)
	return entity.TimeStamps{
		CreatedAt: now,
		CreatedBy: null.StringFrom("tester"),
		UpdatedAt: null.TimeFrom(now),
		UpdatedBy: null.StringFrom("tester"),
	}
}

// Connect must create every table the CLI writes to.
func TestConnectMigratesAllTables(t *testing.T) {
	a := newAdapter(t)

	for _, table := range []string{"accounts", "categories", "credits", "debts", "transactions", "transfers"} {
		assert.Truef(t, a.client.Migrator().HasTable(table), "table %q should exist after Connect", table)
	}
}

// The sheet supplies the primary key, so AutoMigrate must not turn it into an
// AUTOINCREMENT column, otherwise re-pulling a row with id 0 would duplicate it.
func TestPrimaryKeyIsNotAutoIncrement(t *testing.T) {
	a := newAdapter(t)

	var ddl string
	require.NoError(t, a.client.Raw("SELECT sql FROM sqlite_master WHERE name = 'transactions'").Scan(&ddl).Error)
	assert.NotContains(t, ddl, "AUTOINCREMENT", "sheet-provided ids must be stored verbatim")

	db := transaction_db.NewTransactionDb(config.Config{})
	rows := []entity.Transaction{{
		ID: 0, Date: time.Now().UTC(), Account: "BANK", Category: "FOOD",
		Amount: 10, Note: "first", Type: "expense", TimeStamps: stamps(),
	}}

	_, err := db.BulkUpsert(context.Background(), a.client, rows, false)
	require.NoError(t, err)

	rows[0].Note = "second"
	_, err = db.BulkUpsert(context.Background(), a.client, rows, false)
	require.NoError(t, err)

	var stored []model.Transaction
	require.NoError(t, a.client.Find(&stored).Error)
	require.Len(t, stored, 1, "re-pulling the same row must update, not duplicate")
	assert.Equal(t, "second", stored[0].Note)
}

// truncateFirst previously used TRUNCATE, which SQLite does not support.
func TestBulkUpsertTruncateFirst(t *testing.T) {
	a := newAdapter(t)
	db := transaction_db.NewTransactionDb(config.Config{})
	ctx := context.Background()

	first := []entity.Transaction{
		{ID: 1, Date: time.Now().UTC(), Account: "BANK", Category: "FOOD", Amount: 1, Note: "a", Type: "expense", TimeStamps: stamps()},
		{ID: 2, Date: time.Now().UTC(), Account: "BANK", Category: "FOOD", Amount: 2, Note: "b", Type: "expense", TimeStamps: stamps()},
	}
	_, err := db.BulkUpsert(ctx, a.client, first, false)
	require.NoError(t, err)

	second := []entity.Transaction{
		{ID: 9, Date: time.Now().UTC(), Account: "BANK", Category: "FOOD", Amount: 9, Note: "z", Type: "expense", TimeStamps: stamps()},
	}
	_, err = db.BulkUpsert(ctx, a.client, second, true)
	require.NoError(t, err)

	var stored []model.Transaction
	require.NoError(t, a.client.Find(&stored).Error)
	require.Len(t, stored, 1, "truncateFirst must clear previous rows")
	assert.Equal(t, int64(9), stored[0].ID)
}

// The position column is JSON plus the timestamps use null wrappers; both must
// survive a write/read cycle on SQLite.
func TestPositionAndTimestampRoundTrip(t *testing.T) {
	a := newAdapter(t)
	db := transaction_db.NewTransactionDb(config.Config{})

	ts := stamps()
	in := []entity.Transaction{{
		ID: 7, Date: time.Date(2025, 5, 4, 0, 0, 0, 0, time.UTC),
		Account: "BANK", Category: "FOOD", Amount: 12.34, Note: "n", Type: "expense",
		Position: &entity.TransactionPosition{
			ID:     entity.Cell{Row: 5, Col: 2, A1: "B5"},
			Amount: entity.Cell{Row: 5, Col: 6, A1: "F5"},
		},
		TimeStamps: ts,
	}}
	_, err := db.BulkUpsert(context.Background(), a.client, in, false)
	require.NoError(t, err)

	var got model.Transaction
	require.NoError(t, a.client.First(&got, "id = ?", 7).Error)

	require.NotNil(t, got.Position, "json column must round-trip")
	assert.Equal(t, "B5", got.Position.ID.A1)
	assert.Equal(t, 6, got.Position.Amount.Col)
	assert.Equal(t, 12.34, got.Amount)
	assert.Equal(t, "2025-05-04", got.Date.UTC().Format(time.DateOnly))
	assert.True(t, got.UpdatedAt.Valid, "nullable timestamp must be preserved")
	assert.Equal(t, "tester", got.CreatedBy.String)
}

// "gsync finance pull all" runs every target concurrently through errgroup.
// SQLite allows a single writer, so this must not fail with "database is locked".
func TestConcurrentPullAllWrites(t *testing.T) {
	a := newAdapter(t)
	ctx := context.Background()

	accountDb := account_db.NewAccountDb(config.Config{})
	categoryDb := category_db.NewCategoryDb(config.Config{})
	creditDb := credit_db.NewCreditDb(config.Config{})
	debtDb := debt_db.NewDebtDb(config.Config{})
	transactionDb := transaction_db.NewTransactionDb(config.Config{})
	transferDb := transfer_db.NewTransferDb(config.Config{})

	const n = 300

	// each closure mirrors the real pull_*.go flow: Begin -> BulkUpsert -> Commit
	writers := []func() error{
		func() error {
			data := make([]entity.Account, n)
			for i := range data {
				data[i] = entity.Account{ID: int64(i + 1), Code: fmt.Sprintf("A%d", i), Name: "n", Active: "y", Usability: "u", Ownership: "o", TimeStamps: stamps()}
			}
			return commit(ctx, a.client, func(tx *gorm.DB) error {
				_, err := accountDb.BulkUpsert(ctx, tx, data, true)
				return err
			})
		},
		func() error {
			data := make([]entity.Category, n)
			for i := range data {
				data[i] = entity.Category{ID: int64(i + 1), Parent: "", Name: "n", Code: fmt.Sprintf("C%d", i), Nature: "e", Visibility: "v", TimeStamps: stamps()}
			}
			return commit(ctx, a.client, func(tx *gorm.DB) error {
				_, err := categoryDb.BulkUpsert(ctx, tx, data, true)
				return err
			})
		},
		func() error {
			data := make([]entity.Credit, n)
			for i := range data {
				data[i] = entity.Credit{ID: int64(i + 1), Account: "BANK", Name: "n", Amount: 1, TimeStamps: stamps()}
			}
			return commit(ctx, a.client, func(tx *gorm.DB) error {
				_, err := creditDb.BulkUpsert(ctx, tx, data, true)
				return err
			})
		},
		func() error {
			data := make([]entity.Debt, n)
			for i := range data {
				data[i] = entity.Debt{ID: int64(i + 1), Account: "BANK", Name: "n", Amount: 1, TimeStamps: stamps()}
			}
			return commit(ctx, a.client, func(tx *gorm.DB) error {
				_, err := debtDb.BulkUpsert(ctx, tx, data, true)
				return err
			})
		},
		func() error {
			data := make([]entity.Transaction, n)
			for i := range data {
				data[i] = entity.Transaction{ID: int64(i + 1), Date: time.Now().UTC(), Account: "BANK", Category: "FOOD", Amount: 1, Note: "n", Type: "expense", TimeStamps: stamps()}
			}
			return commit(ctx, a.client, func(tx *gorm.DB) error {
				_, err := transactionDb.BulkUpsert(ctx, tx, data, true)
				return err
			})
		},
		func() error {
			data := make([]entity.Transfer, n)
			for i := range data {
				data[i] = entity.Transfer{ID: int64(i + 1), Date: time.Now().UTC(), FromAccount: "A", ToAccount: "B", Amount: 1, Note: "n", TimeStamps: stamps()}
			}
			return commit(ctx, a.client, func(tx *gorm.DB) error {
				_, err := transferDb.BulkUpsert(ctx, tx, data, true)
				return err
			})
		},
	}

	var g errgroup.Group
	for _, w := range writers {
		g.Go(w)
	}
	require.NoError(t, g.Wait(), "concurrent pulls must not hit SQLITE_BUSY")

	for _, tbl := range []string{"accounts", "categories", "credits", "debts", "transactions", "transfers"} {
		var count int64
		require.NoError(t, a.client.Table(tbl).Count(&count).Error)
		assert.Equalf(t, int64(n), count, "table %q should hold every written row", tbl)
	}
}

func commit(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
