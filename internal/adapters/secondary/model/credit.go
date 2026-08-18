package model

import (
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
)

type Credit struct {
	ID       int64           `gorm:"column:id;type:bigint;primaryKey;autoIncrement:false"`
	Account  string          `gorm:"column:account;type:varchar(64);not null"`
	Name     string          `gorm:"column:name;type:varchar(64);not null"`
	Amount   float64         `gorm:"column:amount;float(64);not null"`
	Position *CreditPosition `gorm:"column:position;type:json;serializer:json;default:null"`
	TimeStamps
}

func (Credit) TableName() string {
	return "credits"
}

func (Credit) FromEntity(e entity.Credit) Credit {
	var position *CreditPosition
	if e.Position != nil {
		temp := CreditPosition{}.FromEntity(*e.Position)
		position = &temp
	}

	return Credit{
		ID:         e.ID,
		Account:    e.Account,
		Name:       e.Name,
		Amount:     e.Amount,
		Position:   position,
		TimeStamps: TimeStamps{}.FromEntity(e.TimeStamps),
	}
}

type CreditPosition struct {
	ID      Cell `json:"id"`
	Account Cell `json:"account"`
	Name    Cell `json:"name"`
	Amount  Cell `json:"amount"`
}

func (CreditPosition) FromEntity(e entity.CreditPosition) CreditPosition {
	return CreditPosition{
		ID:      Cell{}.FromEntity(e.ID),
		Account: Cell{}.FromEntity(e.Account),
		Name:    Cell{}.FromEntity(e.Name),
		Amount:  Cell{}.FromEntity(e.Amount),
	}
}
