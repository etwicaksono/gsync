package entity

type Debt struct {
	ID         int64
	Account    string
	Name       string
	Amount     float64
	Position   *DebtPosition
	TimeStamps TimeStamps
}

type DebtPosition struct {
	ID      Cell
	Account Cell
	Name    Cell
	Amount  Cell
}
