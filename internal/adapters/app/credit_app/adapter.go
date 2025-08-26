package credit_app

import (
	port "github.com/etwicaksono/data-transaksi-service/internal/ports/app"
)

// adapter is a concrete implementation of port.CreditAppInterface.
type adapter struct {
	// TODO: deps (mongo/sql/etc)
}

func NewCreditApp( /* deps */ ) port.CreditAppInterface {
	return &adapter{} // TODO: implement me
}
