package app

import "context"

type DebtAppInterface interface {
	PullDebt(ctx context.Context, truncateFirst bool) (err error)
}
