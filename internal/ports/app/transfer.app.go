package app

import "context"

type TransferAppInterface interface {
	SyncTransfer(ctx context.Context) (err error)
}
