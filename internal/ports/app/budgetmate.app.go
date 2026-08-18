package app

import "context"

type BudgetMateAppInterface interface {
	SyncTransactions(ctx context.Context, skip, limit int) error
	SyncTransfers(ctx context.Context, skip, limit int) error
	DebugMapping(ctx context.Context) error
	DebugLabels(ctx context.Context) error
}
