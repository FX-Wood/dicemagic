package dicelang

import (
	"testing"
	"fmt"
	"strconv"
)

type pmf map[int64]float64

type testCase struct {
	id string
	input pmf
	output float64
}

func TestExpectedValue(t *testing.T) {
	tests := []testCase {
		{
			id: "1d4",
			input: pmf{
				1: 0.25,
				2: 0.25,
				3: 0.25,
				4: 0.25,
			},
			output: float64(2.5),
		},
		{
			id: "2d6",
			input: pmf{
				2:  float64(1)/36,
				3:  float64(2)/36,
				4:  float64(3)/36,
				5:  float64(4)/36,
				6:  float64(5)/36,
				7:  float64(6)/36,
				8:  float64(5)/36,
				9:  float64(4)/36,
				10: float64(3)/36,
				11: float64(2)/36,
				12: float64(1)/36,
			},
			output: float64(6.999999999999999),
		},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			fmt.Printf("running test %v", tt)
			got := ExpectedValue(tt.input)
			if got != tt.output {
				t.Errorf("ExpectedValue(%v) = %v; want %v", tt.id, strconv.FormatFloat(got, 'f', -1, 64), tt.output)
			}

		})
	}
}