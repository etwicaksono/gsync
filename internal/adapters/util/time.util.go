package util

import "time"

func YMToLong(s string) (string, error) {
	t, err := time.Parse("2006-01", s) // "2022-12" → 2022-12-01
	if err != nil {
		return "", err
	}
	return t.Format("2006 January"), nil // "2022 December"
}
