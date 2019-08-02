package dicelang

import (
	"math"
)

// StandardDeviation calculates standard deviation using the expected value (eX) and probability mass function (pmf)
func StandardDeviation(eX float64, pmf map[int64]float64) float64 {
	variance := float64(0.0)
	for value, probability := range pmf {
		sum += (probability / 100) * math.Pow(float64(value) - eX, 2)
	}
	return math.Sqrt(variance)
}
