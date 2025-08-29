package entity

type Category struct {
	ID         int64
	Parent     string
	Name       string
	Code       string
	Nature     string
	Visibility string
	Position   *CategoryPosition
	TimeStamps TimeStamps
}

type CategoryPosition struct {
	ID         Cell
	Parent     Cell
	Name       Cell
	Code       Cell
	Nature     Cell
	Visibility Cell
}
