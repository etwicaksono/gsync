package model

import (
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
)

type Transaction struct {
	ID       int64                `gorm:"column:id;type:bigint;primaryKey;autoIncrement:false"`
	Date     time.Time            `gorm:"column:date;type:datetime;not null;index:ix_transactions_date"`
	Account  string               `gorm:"column:account;type:varchar(64);not null"`
	Category string               `gorm:"column:category;type:varchar(64);not null;index:ix_transactions_category"`
	Amount   float64              `gorm:"column:amount;float(64);not null;index:ix_transactions_amount"`
	Note     string               `gorm:"column:note;type:text;not null"`
	Type     string               `gorm:"column:type;type:varchar(16);not null"`
	Position *TransactionPosition `gorm:"column:position;type:json;serializer:json;default:null"`
	TimeStamps
}

func (Transaction) TableName() string {
	return "transactions"
}

func (Transaction) FromEntity(e entity.Transaction) Transaction {
	var position *TransactionPosition
	if e.Position != nil {
		temp := TransactionPosition{}.FromEntity(*e.Position)
		position = &temp
	}

	return Transaction{
		ID:         e.ID,
		Date:       e.Date,
		Account:    e.Account,
		Category:   e.Category,
		Amount:     e.Amount,
		Note:       e.Note,
		Type:       e.Type,
		Position:   position,
		TimeStamps: TimeStamps{}.FromEntity(e.TimeStamps),
	}
}

type TransactionPosition struct {
	ID       Cell `json:"id"`
	Date     Cell `json:"date"`
	Account  Cell `json:"account"`
	Category Cell `json:"category"`
	Amount   Cell `json:"amount"`
	Note     Cell `json:"note"`
	Type     Cell `json:"type"`
}

func (TransactionPosition) FromEntity(e entity.TransactionPosition) TransactionPosition {
	return TransactionPosition{
		ID:       Cell{}.FromEntity(e.ID),
		Date:     Cell{}.FromEntity(e.Date),
		Account:  Cell{}.FromEntity(e.Account),
		Category: Cell{}.FromEntity(e.Category),
		Amount:   Cell{}.FromEntity(e.Amount),
		Note:     Cell{}.FromEntity(e.Note),
		Type:     Cell{}.FromEntity(e.Type),
	}
}
