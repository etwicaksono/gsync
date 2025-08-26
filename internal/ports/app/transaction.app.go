package app

import "context"

type TransactionAppInterface interface {
	PullTransaction(ctx context.Context, truncateFirst bool) (err error)
}
