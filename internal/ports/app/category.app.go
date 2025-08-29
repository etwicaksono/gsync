package app

import "context"

type CategoryAppInterface interface {
	PullCategory(ctx context.Context, truncateFirst bool) (err error)
}
