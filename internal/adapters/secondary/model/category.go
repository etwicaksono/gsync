package model

import (
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
)

type Category struct {
	ID         int64             `gorm:"column:id;type:bigint;primaryKey;autoIncrement:false"`
	Parent     string            `gorm:"column:parent;type:varchar(64);not null"`
	Name       string            `gorm:"column:name;type:varchar(64);not null"`
	Code       string            `gorm:"column:code;type:varchar(64);not null"`
	Nature     string            `gorm:"column:nature;type:varchar(64);not null"`
	Visibility string            `gorm:"column:visibility;type:varchar(64);not null"`
	Position   *CategoryPosition `gorm:"column:position;type:json;serializer:json;default:null"`
	TimeStamps
}

func (Category) TableName() string {
	return "categories"
}

func (Category) FromEntity(e entity.Category) Category {
	var position *CategoryPosition
	if e.Position != nil {
		temp := CategoryPosition{}.FromEntity(*e.Position)
		position = &temp
	}

	return Category{
		ID:         e.ID,
		Parent:     e.Parent,
		Name:       e.Name,
		Code:       e.Code,
		Nature:     e.Nature,
		Visibility: e.Visibility,
		Position:   position,
		TimeStamps: TimeStamps{}.FromEntity(e.TimeStamps),
	}
}

type CategoryPosition struct {
	ID         Cell `json:"id"`
	Parent     Cell `json:"parent"`
	Name       Cell `json:"name"`
	Code       Cell `json:"code"`
	Nature     Cell `json:"nature"`
	Visibility Cell `json:"visibility"`
}

func (CategoryPosition) FromEntity(e entity.CategoryPosition) CategoryPosition {
	return CategoryPosition{
		ID:         Cell{}.FromEntity(e.ID),
		Parent:     Cell{}.FromEntity(e.Parent),
		Name:       Cell{}.FromEntity(e.Name),
		Code:       Cell{}.FromEntity(e.Code),
		Nature:     Cell{}.FromEntity(e.Nature),
		Visibility: Cell{}.FromEntity(e.Visibility),
	}
}
