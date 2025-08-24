package entity

type Cell struct {
	Row int    `json:"row"`
	Col int    `json:"col"`
	A1  string `json:"a1"`
}

type ReadData struct {
	Val  string `json:"val"`
	Cell Cell   `json:"cell"`
}
