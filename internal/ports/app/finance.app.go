package app

import (
	"context"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
)

type FinanceAppInterface interface {
	PullAccount(ctx context.Context, truncateFirst bool) (err error)
	PullCategory(ctx context.Context, truncateFirst bool) (err error)
	PullCredit(ctx context.Context, truncateFirst bool) (err error)
	PullDebt(ctx context.Context, truncateFirst bool) (err error)
	PullTransaction(ctx context.Context, truncateFirst bool) (err error)
	PullTransfer(ctx context.Context, truncateFirst bool) (err error)
	PushResumeDetail(ctx context.Context, resumeStartColRange string, filter entity.VResumeDetailListFilter) (err error)
	PushResumeRecap(ctx context.Context, resumeStartColRange string, filter entity.VResumeDetailListFilter) (err error)
}
