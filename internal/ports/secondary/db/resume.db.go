package db

import (
	"context"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"gorm.io/gorm"
)

type ResumeDbInterface interface {
	ListResumeDetail(ctx context.Context, tx *gorm.DB, filter entity.VResumeDetailListFilter) (result []entity.Resume, err error)
	ListResumeRecap(ctx context.Context, tx *gorm.DB, filter entity.VResumeDetailListFilter) (result []entity.Resume, err error)
}
