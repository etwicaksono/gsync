package app

import "context"

type CreditAppInterface interface {
	PullCredit(ctx context.Context, truncateFirst bool) (err error)
}
