package app

import "context"

type TransactionAppInterface interface {
	Sync(ctx context.Context) (err error)
}
