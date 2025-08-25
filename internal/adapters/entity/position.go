package entity

type Cell struct {
	Row int
	Col int
	A1  string
}

type ReadData struct {
	Val  string
	Cell Cell
}
