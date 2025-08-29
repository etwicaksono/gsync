package entity

import "time"

type Transaction struct {
	ID         int64
	Date       time.Time
	Account    string
	Category   string
	Amount     float64
	Note       string
	Type       string
	Position   *TransactionPosition
	TimeStamps TimeStamps
}

type TransactionPosition struct {
	ID       Cell
	Date     Cell
	Account  Cell
	Category Cell
	Amount   Cell
	Note     Cell
	Type     Cell
}
