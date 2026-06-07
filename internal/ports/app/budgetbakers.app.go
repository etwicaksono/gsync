package app

import (
	"context"
)

type BudgetBakersAppInterface interface {
	SyncTransactions(ctx context.Context, skip int, limit int) error
}
