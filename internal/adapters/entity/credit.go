package entity

type Credit struct {
	ID         int64
	Account    string
	Name       string
	Amount     float64
	Position   *CreditPosition
	TimeStamps TimeStamps
}

type CreditPosition struct {
	ID      Cell
	Account Cell
	Name    Cell
	Amount  Cell
}
