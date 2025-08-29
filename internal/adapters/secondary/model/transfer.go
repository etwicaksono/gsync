package model

import (
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
)

type Transfer struct {
	ID          int64             `gorm:"column:id;int(64);primaryKey"`
	Date        time.Time         `gorm:"column:date;type:datetime;not null"`
	FromAccount string            `gorm:"column:from_account;type:varchar(64);not null"`
	ToAccount   string            `gorm:"column:to_account;type:varchar(64);not null"`
	Amount      float64           `gorm:"column:amount;float(64);not null"`
	Note        string            `gorm:"column:note;type:text;not null"`
	Position    *TransferPosition `gorm:"column:position;type:json;serializer:json;default:null"`
	TimeStamps
}

func (Transfer) TableName() string {
	return "transfers"
}

func (Transfer) FromEntity(e entity.Transfer) Transfer {
	var position *TransferPosition
	if e.Position != nil {
		temp := TransferPosition{}.FromEntity(*e.Position)
		position = &temp
	}

	return Transfer{
		ID:          e.ID,
		Date:        e.Date,
		FromAccount: e.FromAccount,
		ToAccount:   e.ToAccount,
		Amount:      e.Amount,
		Note:        e.Note,
		Position:    position,
		TimeStamps:  TimeStamps{}.FromEntity(e.TimeStamps),
	}
}

type TransferPosition struct {
	ID          Cell `json:"id"`
	Date        Cell `json:"date"`
	FromAccount Cell `json:"from_account"`
	ToAccount   Cell `json:"to_account"`
	Amount      Cell `json:"amount"`
	Note        Cell `json:"note"`
}

func (TransferPosition) FromEntity(e entity.TransferPosition) TransferPosition {
	return TransferPosition{
		ID:          Cell{}.FromEntity(e.ID),
		Date:        Cell{}.FromEntity(e.Date),
		FromAccount: Cell{}.FromEntity(e.FromAccount),
		ToAccount:   Cell{}.FromEntity(e.ToAccount),
		Amount:      Cell{}.FromEntity(e.Amount),
		Note:        Cell{}.FromEntity(e.Note),
	}
}
