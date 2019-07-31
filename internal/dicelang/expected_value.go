package dicelang

import "fmt"

//ExportedValue takes a probability mass function(pmf) and returns an expected value(eX)
func ExpectedValue(pmf map[int64]float64) float64 {
	eX := 0.0
	for value, probability := range pmf {
		fmt.Println()
		eX += float64(value) * probability
	}
	return eX
}
