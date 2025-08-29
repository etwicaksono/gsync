package app

import "context"

type AccountAppInterface interface {
	PullAccount(ctx context.Context, truncateFirst bool) (err error)
}
