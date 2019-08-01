package dicelang

import (
	"testing"
	"fmt"
	"strconv"
	"math"
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
				1: 0.25 * 100, // convert to percentages
				2: 0.25 * 100,
				3: 0.25 * 100,
				4: 0.25 * 100,
			},
			output: float64(2.5),
		},
		{
			id: "2d6",
			input: pmf{
				2:  (float64(1)/36) * 100, 
				3:  (float64(2)/36) * 100,
				4:  (float64(3)/36) * 100,
				5:  (float64(4)/36) * 100,
				6:  (float64(5)/36) * 100,
				7:  (float64(6)/36) * 100,
				8:  (float64(5)/36) * 100,
				9:  (float64(4)/36) * 100,
				10: (float64(3)/36) * 100,
				11: (float64(2)/36) * 100,
				12: (float64(1)/36) * 100,
			},
			output: float64(7),
		},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			fmt.Printf("running test %v", tt)
			got := ExpectedValue(tt.input)
			// equality must account for floating point error
			if math.Abs(got - tt.output) > 1e-9  {
				t.Errorf("ExpectedValue(%v) = %v; want %v ± ", tt.id, strconv.FormatFloat(got, 'f', -1, 64), tt.output)
			}

		})
	}
}