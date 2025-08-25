package model

import "github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"

type Cell struct {
	Row int    `json:"row"`
	Col int    `json:"col"`
	A1  string `json:"a1"`
}

func (Cell) FromEntity(e entity.Cell) Cell {
	return Cell{
		Row: e.Row,
		Col: e.Col,
		A1:  e.A1,
	}
}

type ReadData struct {
	Val  string `json:"val"`
	Cell Cell   `json:"cell"`
}
