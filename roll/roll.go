package roll

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"sort"
)

type Dice struct {
	Count       int64
	Sides       int64
	resultTotal int64
	Faces       []int64
	Max         int64
	Min         int64
	H           int64
	L           int64
}
type DiceSet struct {
	Dice         []Dice
	ResultTotals []int64
	DiceType     string
	Min          int64
	Max          int64
}

func (d *DiceSet) Roll() ([]int64, error) {
	d.Min, d.Max = 0, 0
	for _, ds := range d.Dice {
		result, err := ds.Roll()
		if err != nil {
			return nil, err
		}
		d.Min += ds.Count
		d.Max += (ds.Count - (ds.H + ds.L)) * ds.Sides
		d.ResultTotals = append(d.ResultTotals, result)
	}
	return d.ResultTotals, nil
}
func (d *Dice) Roll() (int64, error) {
	if d.resultTotal != 0 {
		return d.resultTotal, nil
	}
	faces, result, err := roll(d.Count, d.Sides, d.H, d.L)
	if err != nil {
		return 0, err
	}
	d.Min = d.Count
	d.Max = (d.Count - (d.H + d.L)) * d.Sides
	d.Faces = faces
	d.resultTotal = result
	return result, nil
}
func (d *Dice) Probability() float64 {
	if d.resultTotal > 0 {
		return diceProbability(d.Count, d.Sides, d.resultTotal)
	}
	return 0
}

//Roll creates a random number that represents the roll of
//some dice
func roll(numberOfDice int64, sides int64, H int64, L int64) ([]int64, int64, error) {
	var faces []int64
	if numberOfDice > 1000 {
		err := fmt.Errorf("I can't hold that many dice")
		return faces, 0, err
	} else if sides > 1000 {
		err := fmt.Errorf("A die with that many sides is basically round")
		return faces, 0, err
	} else if sides < 1 {
		err := fmt.Errorf("/me ponders the meaning of a zero sided die")
		return faces, 0, err
	} else {
		total := int64(0)
		for i := int64(0); i < numberOfDice; i++ {
			face, err := generateRandomInt(1, int64(sides))
			if err != nil {
				return faces, 0, err
			}
			faces = append(faces, face)
			total += face
		}
		sort.Slice(faces, func(i, j int) bool { return faces[i] < faces[j] })
		faces = faces[H:]
		faces = faces[:L]
		return faces, total, nil
	}
}

func generateRandomInt(min int64, max int64) (int64, error) {
	if max <= 0 || min < 0 {
		err := fmt.Errorf("Cannot make a random int of size zero")
		return 0, err
	}
	size := max - min
	if size == 0 {
		return 1, nil
	}
	//rand.Int does not return the max value, add 1
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(size+1)))
	if err != nil {
		err = fmt.Errorf("Couldn't make a random number. Out of entropy?")
		return 0, err
	}
	n := nBig.Int64()
	return n + int64(min), nil
}

func diceProbability(numberOfDice int64, sides int64, target int64) float64 {
	rollAmount := math.Pow(float64(sides), float64(numberOfDice))
	targetAmount := float64(0)
	var possibilities []int64
	for i := int64(1); i <= sides; i++ {
		possibilities = append(possibilities, i)
	}
	c := make(chan []int64)
	go generateProducts(c, possibilities, numberOfDice)
	for product := range c {
		if sumInt64(product...) == target {
			targetAmount++
		}
	}
	p := (targetAmount / rollAmount)
	return p
}

func generateProducts(c chan []int64, possibilities []int64, numberOfDice int64) {
	lens := int64(len(possibilities))
	for ix := make([]int64, numberOfDice); ix[0] < lens; nextIndex(ix, lens) {
		r := make([]int64, numberOfDice)
		for i, j := range ix {
			r[i] = possibilities[j]
		}
		c <- r
	}
	close(c)
}
func nextIndex(ix []int64, lens int64) {
	for j := len(ix) - 1; j >= 0; j-- {
		ix[j]++
		if j == 0 || ix[j] < lens {
			return
		}
		ix[j] = 0
	}
}
func sumInt64(nums ...int64) int64 {
	r := int64(0)
	for _, n := range nums {
		r += n
	}
	return r
}
