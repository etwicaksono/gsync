package model

import (
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
)

type Debt struct {
	ID       int64         `gorm:"column:id;type:bigint;primaryKey;autoIncrement:false"`
	Account  string        `gorm:"column:account;type:varchar(64);not null"`
	Name     string        `gorm:"column:name;type:varchar(64);not null"`
	Amount   float64       `gorm:"column:amount;float(64);not null"`
	Position *DebtPosition `gorm:"column:position;type:json;serializer:json;default:null"`
	TimeStamps
}

func (Debt) TableName() string {
	return "debts"
}

func (Debt) FromEntity(e entity.Debt) Debt {
	var position *DebtPosition
	if e.Position != nil {
		temp := DebtPosition{}.FromEntity(*e.Position)
		position = &temp
	}

	return Debt{
		ID:         e.ID,
		Account:    e.Account,
		Name:       e.Name,
		Amount:     e.Amount,
		Position:   position,
		TimeStamps: TimeStamps{}.FromEntity(e.TimeStamps),
	}
}

type DebtPosition struct {
	ID      Cell `json:"id"`
	Account Cell `json:"account"`
	Name    Cell `json:"name"`
	Amount  Cell `json:"amount"`
}

func (DebtPosition) FromEntity(e entity.DebtPosition) DebtPosition {
	return DebtPosition{
		ID:      Cell{}.FromEntity(e.ID),
		Account: Cell{}.FromEntity(e.Account),
		Name:    Cell{}.FromEntity(e.Name),
		Amount:  Cell{}.FromEntity(e.Amount),
	}
}
