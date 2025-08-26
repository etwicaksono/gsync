package app

import "context"

type TransferAppInterface interface {
	SyncTransfer(ctx context.Context, truncateFirst bool) (err error)
}
