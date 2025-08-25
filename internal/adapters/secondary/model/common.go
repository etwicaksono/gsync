package model

import (
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/guregu/null"
	"time"
)

type TimeStamps struct {
	CreatedAt time.Time   `gorm:"column:created_at;type:datetime;not null"`
	CreatedBy null.String `gorm:"column:created_by;type:varchar(64)"`
	UpdatedAt null.Time   `gorm:"column:updated_at;type:datetime;default:NULL"`
	UpdatedBy null.String `gorm:"column:updated_by;type:varchar(64);default:NULL"`
}

func (TimeStamps) FromEntity(e entity.TimeStamps) TimeStamps {
	return TimeStamps(e)
}
