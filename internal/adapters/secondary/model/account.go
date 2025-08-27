package model

import (
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
)

type Account struct {
	ID        int64            `gorm:"column:id;int(64);primaryKey"`
	Code      string           `gorm:"column:code;type:varchar(64);not null"`
	Name      string           `gorm:"column:name;type:varchar(64);not null"`
	Active    string           `gorm:"column:active;type:varchar(64);not null"`
	Usability string           `gorm:"column:usability;type:varchar(64);not null"`
	Ownership string           `gorm:"column:ownership;type:varchar(64);not null"`
	Position  *AccountPosition `gorm:"column:position;type:json;serializer:json;default:null"`
	TimeStamps
}

func (Account) TableName() string {
	return "accounts"
}

func (Account) FromEntity(e entity.Account) Account {
	var position *AccountPosition
	if e.Position != nil {
		temp := AccountPosition{}.FromEntity(*e.Position)
		position = &temp
	}

	return Account{
		ID:         e.ID,
		Position:   position,
		Code:       e.Code,
		Name:       e.Name,
		Active:     e.Active,
		Usability:  e.Usability,
		Ownership:  e.Ownership,
		TimeStamps: TimeStamps{}.FromEntity(e.TimeStamps),
	}
}

type AccountPosition struct {
	ID        Cell `json:"id"`
	Code      Cell `json:"code"`
	Name      Cell `json:"name"`
	Active    Cell `json:"active"`
	Usability Cell `json:"usability"`
	Ownership Cell `json:"ownership"`
}

func (AccountPosition) FromEntity(e entity.AccountPosition) AccountPosition {
	return AccountPosition{
		ID:        Cell{}.FromEntity(e.ID),
		Code:      Cell{}.FromEntity(e.Code),
		Name:      Cell{}.FromEntity(e.Name),
		Active:    Cell{}.FromEntity(e.Active),
		Usability: Cell{}.FromEntity(e.Usability),
		Ownership: Cell{}.FromEntity(e.Ownership),
	}
}
