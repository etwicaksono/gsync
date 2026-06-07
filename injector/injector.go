//go:build wireinject
// +build wireinject

package injector

import (
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/app/gsheet_finance_app"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/db/account_db"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/db/category_db"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/db/credit_db"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/db/debt_db"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/db/resume"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/db/transaction_db"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/db/transfer_db"
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/app"
	"github.com/google/wire"
	"gorm.io/gorm"
)

var dbSet = wire.NewSet(
	transaction_db.NewTransactionDb,
	transfer_db.NewTransferDb,
	debt_db.NewDebtDb,
	credit_db.NewCreditDb,
	account_db.NewAccountDb,
	category_db.NewCategoryDb,
	gsheet_finance_app.NewFinanceApp,
	resume.NewResumeDb,
)

func GsheetFinanceAppProvider(config config.Config, db *gorm.DB) app.GsheetFinanceAppInterface {
	wire.Build(dbSet)
	return nil
}
