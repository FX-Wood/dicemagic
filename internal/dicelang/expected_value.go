package dicelang

//ExpectedValue takes a probability mass function(pmf) and returns an expected value(eX)
func ExpectedValue(pmf map[int64]float64) float64 {
	eX := 0.0
	for value, probability := range pmf {
		eX += (float64(value) * probability) / 100 // in our case the pmf maps the integer result to probability as a percentage
	}
	return eX
}
