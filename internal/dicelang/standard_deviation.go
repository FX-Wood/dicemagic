package dicelang

import (
	"math"
	"fmt"
)

// StandardDeviation calculates standard deviation using the expected value (eX) and probability mass function (pmf)
func StandardDeviation(eX float64, pmf map[int64]float64) float64 {
	// calculate variance
	populationSize := 0.0
	sum := float64(0.0)
	for value, probability := range pmf {
		sum += (probability / 100) * math.Pow(float64(value), 2) - math.Pow(eX, 2)
		populationSize++
	}
	variance := sum / populationSize
	fmt.Println(variance)
	return math.Sqrt(variance)
}
