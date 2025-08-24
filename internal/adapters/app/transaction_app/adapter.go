package transaction_app

import (
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	port "github.com/etwicaksono/data-transaksi-service/internal/ports/app"
)

// adapter is a concrete implementation of port.TransactionAppInterface.
type adapter struct {
	cfg config.Config
}

func NewTransactionApp(cfg config.Config) port.TransactionAppInterface {
	return &adapter{
		cfg: cfg,
	}
}
