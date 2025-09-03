package finance_app

import (
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	port "github.com/etwicaksono/data-transaksi-service/internal/ports/app"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/secondary/db"
	"gorm.io/gorm"
)

// adapter is a concrete implementation of port.FinanceAppInterface.
type adapter struct {
	cfg             config.Config
	dbClient        *gorm.DB
	accountDb       db.AccountDbInterface
	categoryDb      db.CategoryDbInterface
	creditDb        db.CreditDbInterface
	debtDb          db.DebtDbInterface
	transactionDb   db.TransactionDbInterface
	transferDb      db.TransferDbInterface
	vResumeDetailDb db.ResumeDbInterface
}

func NewFinanceApp(
	cfg config.Config,
	dbClient *gorm.DB,
	accountDb db.AccountDbInterface,
	categoryDb db.CategoryDbInterface,
	creditDb db.CreditDbInterface,
	debtDb db.DebtDbInterface,
	transactionDb db.TransactionDbInterface,
	transferDb db.TransferDbInterface,
	vResumeDb db.ResumeDbInterface,
) port.FinanceAppInterface {
	return &adapter{
		cfg:             cfg,
		dbClient:        dbClient,
		accountDb:       accountDb,
		categoryDb:      categoryDb,
		creditDb:        creditDb,
		debtDb:          debtDb,
		transactionDb:   transactionDb,
		transferDb:      transferDb,
		vResumeDetailDb: vResumeDb,
	}
}
