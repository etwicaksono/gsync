package transfer_app

import (
	port "github.com/etwicaksono/data-transaksi-service/internal/ports/app"
)

// adapter is a concrete implementation of port.TransferAppInterface.
type adapter struct {
	// TODO: deps (mongo/sql/etc)
}

func NewTransferApp( /* deps */ ) port.TransferAppInterface {
	return &adapter{} // TODO: implement me
}
