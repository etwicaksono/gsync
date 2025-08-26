package app

import "context"

type TransferAppInterface interface {
	PullTransfer(ctx context.Context, truncateFirst bool) (err error)
}
