package app

import "context"

type TransactionAppInterface interface {
	SyncTransaction(ctx context.Context, truncateFirst bool) (err error)
}
