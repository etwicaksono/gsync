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

func TestParseA1RangeStart(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input     string
		wantSheet string
		wantCol   int
		wantRow   int
	}

	tests := []testCase{
		{
			input:     "Sheet3!A2",
			wantSheet: "Sheet3",
			wantCol:   1,
			wantRow:   2,
		},
		{
			input:     "Sheet3!B2:F2",
			wantSheet: "Sheet3",
			wantCol:   2,
			wantRow:   2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			sheet, col, row := ParseA1RangeStart(tc.input)
			assert.Equal(t, tc.wantSheet, sheet, "ParseA1RangeStart(%d)", tc.input)
			assert.Equal(t, tc.wantCol, col, "ParseA1RangeStart(%d)", tc.input)
			assert.Equal(t, tc.wantRow, row, "ParseA1RangeStart(%d)", tc.input)
		})
	}
}
