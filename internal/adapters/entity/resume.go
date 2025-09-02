package entity

import "time"

type Resume struct {
	ID      int64
	Code    string
	Monthly MonthlyList
}

type MonthlyPoint struct {
	YM  string
	Amt float64
}

type MonthlyList []MonthlyPoint

type VResumeDetailListFilter struct {
	StartDate time.Time
	EndDate   time.Time
}
