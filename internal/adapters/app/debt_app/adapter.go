package debt_app

import (
	port "github.com/etwicaksono/data-transaksi-service/internal/ports/app"
)

// adapter is a concrete implementation of port.DebtAppInterface.
type adapter struct {
	// TODO: deps (mongo/sql/etc)
}

func NewDebtApp( /* deps */ ) port.DebtAppInterface {
	return &adapter{} // TODO: implement me
}
