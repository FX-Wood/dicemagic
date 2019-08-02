package dicelang

import (
	"testing"
	"fmt"
	"math"
	"strconv"
)

func TestStandardDeviation(t *testing.T) {
	type pmf map[int64]float64
	type inputs struct {
		a float64
		b pmf
	}
	type testCase struct {
		id string
		input inputs
		output float64
	}
	tests := []testCase {
		{
			id: "1d4",
			input: inputs {
				a: 2.5,
				b: pmf{
					1: 25.0,
					2: 25.0,
					3: 25.0,
					4: 25.0,
				},
			},
			output: 1.118033988749895,
		},
		{
			id: "6d4",
			input: inputs {
				a: 15,
				b: pmf{
					6:  0.0244140625,
					7:  0.146484375,
					8:  0.5126953125,
					9:  1.3671875,
					10: 2.9296875,
					11: 5.2734375,
					12: 8.203125,
					13: 11.1328125,
					14: 13.330078125,
					15: 14.16015625,
					16: 13.330078125,
					17: 11.1328125,
					18: 8.203125,
					19: 5.2734375,
					20: 2.9296875,
					21: 1.3671875,
					22: 0.5126953125,
					23: 0.146484375,
					24: 0.0244140625,
				},
			},
			output: 2.7386127875258306,
		},
		{
			id: "3d6 + 1d4",
			input: inputs {
				a: 13.0,
				b: pmf{
					4:   0.115740740741,
					5:   0.462962962963,
					6:   1.15740740741,
					7:   2.31481481481,
					8:   3.93518518519,
					9:   6.01851851852,
					10:  8.21759259259,
					11: 10.1851851852,
					12: 11.5740740741,
					13: 12.037037037,
					14: 11.5740740741,
					15: 10.1851851852,
					16:  8.21759259259,
					17:  6.01851851852,
					18:  3.93518518519,
					19:  2.31481481481,
					20:  1.15740740741,
					21:  0.462962962963,
					22:  0.115740740741,
				},
			},
			output: 3.162277660168958,
		},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			fmt.Printf("running test %v", tt)
			got := StandardDeviation(tt.input.a, tt.input.b)
			// equality must account for floating point error
			margin := 1e-9
			if math.Abs(got - tt.output) > margin {
				t.Errorf("StandardDeviation(%v) = %v; want %v ± %v", tt.id, strconv.FormatFloat(got, 'f', -1, 64), tt.output, margin)
			}
		})
	}
}
