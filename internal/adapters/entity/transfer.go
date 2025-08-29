package entity

import "time"

type Transfer struct {
	ID          int64
	Date        time.Time
	FromAccount string
	ToAccount   string
	Amount      float64
	Note        string
	Position    *TransferPosition
	TimeStamps  TimeStamps
}

type TransferPosition struct {
	ID          Cell
	Date        Cell
	FromAccount Cell
	ToAccount   Cell
	Amount      Cell
	Note        Cell
}
