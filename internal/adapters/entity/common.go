package entity

import (
	"time"

	"github.com/guregu/null"
)

type TimeStamps struct {
	CreatedAt time.Time
	CreatedBy null.String
	UpdatedAt null.Time
	UpdatedBy null.String
}
