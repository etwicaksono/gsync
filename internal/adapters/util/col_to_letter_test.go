package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColToLetter(t *testing.T) {
	t.Parallel()

	tests := map[int]string{
		0:     "", // non-positive → empty (matches current function behavior)
		-1:    "",
		1:     "A",
		2:     "B",
		25:    "Y",
		26:    "Z",
		27:    "AA",
		28:    "AB",
		52:    "AZ",
		53:    "BA",
		54:    "BB",
		701:   "ZY",
		702:   "ZZ",
		703:   "AAA",
		704:   "AAB",
		18278: "ZZZ",
	}

	for in, want := range tests {
		in, want := in, want
		t.Run(want, func(t *testing.T) {
			t.Parallel()
			got := ColToLetter(in)
			fmt.Printf("input: %v result: %v\n", in, got)
			assert.Equal(t, want, got, "ColToLetter(%d)", in)
		})
	}
}
