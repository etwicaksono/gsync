package entity

type Account struct {
	ID         int64
	Code       string
	Name       string
	Active     string
	Usability  string
	Ownership  string
	Position   *AccountPosition
	TimeStamps TimeStamps
}

type AccountPosition struct {
	ID        Cell
	Code      Cell
	Name      Cell
	Active    Cell
	Usability Cell
	Ownership Cell
}
