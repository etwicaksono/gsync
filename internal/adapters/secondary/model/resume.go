package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
)

type Resume struct {
	ID      int64       `gorm:"column:id"`
	Code    string      `gorm:"column:code"`
	Monthly MonthlyList `gorm:"column:monthly"`
}

func (vrd Resume) ToEntity() entity.Resume {
	// If your entity expects an ordered list:
	points := make([]entity.MonthlyPoint, 0, len(vrd.Monthly))
	for _, p := range vrd.Monthly {
		points = append(points, entity.MonthlyPoint{YM: p.YM, Amt: p.Amt})
	}
	return entity.Resume{
		ID:      vrd.ID,
		Code:    vrd.Code,
		Monthly: points,
	}
}

type MonthlyPoint struct {
	YM  string  `json:"ym"`
	Amt float64 `json:"amt"`
}

type MonthlyList []MonthlyPoint

func (ml *MonthlyList) Scan(v any) error {
	switch x := v.(type) {
	case []byte:
		if len(x) == 0 {
			*ml = MonthlyList{}
			return nil
		}
		return json.Unmarshal(x, ml)
	case string:
		if x == "" {
			*ml = MonthlyList{}
			return nil
		}
		return json.Unmarshal([]byte(x), ml)
	default:
		return fmt.Errorf("MonthlyList.Scan: expected []byte or string, got %T", v)
	}
}

func (ml *MonthlyList) Value() (driver.Value, error) {
	if ml == nil {
		return nil, nil
	}
	return json.Marshal(ml)
}
