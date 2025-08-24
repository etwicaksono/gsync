package transaction_app

import (
	port "github.com/etwicaksono/data-transaksi-service/internal/ports/app"
)

// adapter is a concrete implementation of port.TransactionAppInterface.
type adapter struct {
	// TODO: deps (mongo/sql/etc)
}

func NewTransactionApp( /* deps */ ) port.TransactionAppInterface {
	return &adapter{} // TODO: implement me
}
