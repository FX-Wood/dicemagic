package main

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"sort"
	"strconv"
)

func main() {
	args := os.Args
	if len(args) > 1 {
		buf, _ := ioutil.ReadFile(args[1])
		source := string(buf)

		l := tdopLexer(source)
		p := parser{lexer: l}

		stmts := p.statements()
		fmt.Println(source)
		fmt.Println("----------")
		for _, stmt := range stmts {
			printAST(stmt, 0)
			fmt.Println()
			v, ds, err := stmt.GetDiceSet()
			if err != nil {
				fmt.Printf("Could not parse input: %v\n", err)
				return
			}
			for _, v := range ds.Dice {
				fmt.Printf("%+v\n", v)
			}
			fmt.Printf("for a total of: %+v\n", v)
			fmt.Printf("map: %+v\n", ds.TotalsByColor)
			fmt.Println("----------")
		}
	}
}

func (t *ast) GetDiceSet() (float64, DiceSet, error) {
	v, ret, err := t.eval(&DiceSet{})
	return v, *ret, err
}

func (t *ast) eval(ds *DiceSet) (float64, *DiceSet, error) {
	switch t.sym {
	case "(NUMBER)":
		i, _ := strconv.ParseFloat(t.value, 64)
		if len(t.children) > 0 {
			//grab any color below, get it on ds
			t.children[0].eval(ds)
		}
		return i, ds, nil
	case "-H", "-L":
		var intx int64
		var e float64
		if len(t.children) != 0 {
			var (
				err error
				x   float64
			)
			for _, c := range t.children {
				x, ds, err = c.eval(ds)
				if err != nil {
					return 0, ds, err
				}
				e += x
			}
		}
		intx = int64(math.Max(1, e))
		switch t.sym {
		case "-H":
			ds.keepHighest = intx
		case "-L":
			ds.leepLowest = intx
		}
		return 0, ds, nil
	case "d":
		dice := Dice{}
		//sub d's don't get added to totals
		//ds.Pause()
		//ds.ColorSetDepth++
		var nums []int64
		for i := 0; i < len(t.children); i++ {
			var num float64
			var err error
			num, ds, err = t.children[i].eval(ds)
			if err != nil {
				return 0, nil, err
			}
			if num != 0 {
				nums = append(nums, int64(num))
			}
		}
		//ds.UnPause()
		//ds.ColorSetDepth--
		dice.Count = nums[0]
		dice.Sides = nums[1]
		//actually roll dice here
		res, err := ds.PushAndRoll(dice)
		return float64(res), ds, err
	case "+", "-", "*", "/", "^":
		x, ds, err := t.preformArithmitic(ds, t.sym)
		if err != nil {
			return 0, ds, err
		}
		return x, ds, nil
	case "{", "roll":
		var x float64
		for _, c := range t.children {
			y, ds, err := c.eval(ds)
			if err != nil {
				return 0, ds, err
			}
			x += y
		}
		return x, ds, nil
	case "(IDENT)":
		ds.PushColor(t.value)
		return 0, ds, nil
	case "if":
		res, ds, err := t.children[0].evaluateBoolean(ds)
		if err != nil {
			return 0, ds, err
		}
		fmt.Print(res, " ")
		var c *ast
		if res {
			c = t.children[1]
		} else {
			if len(t.children) < 3 {
				return 0, ds, nil
			}
			c = t.children[2]
		}
		var x float64
		//Evaluate chosen child
		y, ds, err := c.eval(ds)
		if err != nil {
			return 0, ds, err
		}
		x += y
		return x, ds, nil
	default:
		return 0, ds, fmt.Errorf("Unsupported symbol: %s", t.sym)
	}
	return 0, ds, fmt.Errorf("bad ast")
}

func printAST(t *ast, identation int) {
	fmt.Println()
	for i := 0; i < identation; i++ {
		fmt.Print(" ")
	}
	fmt.Print("(")
	fmt.Print(t.sym, ":", t.value)
	if len(t.children) > 0 {
		for _, c := range t.children {
			fmt.Print(" ")
			printAST(c, identation+4)
		}
	}
	fmt.Print(")")
}
func (t *ast) preformArithmitic(ds *DiceSet, op string) (float64, *DiceSet, error) {
	//arithmitic is always binary
	var x float64
	//...except for the "-" unary operator
	if len(t.children) < 2 {

	}
	diceCount := len(ds.Dice)
	for _, c := range t.children {
		ds.colorDepth++
		x, ds, err := t.children[0].eval(ds)
		if err != nil {
			return 0, ds, err
		}
		y, ds, err := t.children[1].eval(ds)
		if err != nil {
			return 0, ds, err
		}
		ds.colorDepth--
	}
	newDice := len(ds.Dice) - diceCount
	switch op {
	case "+":
		x += y
	case "-":
		x -= y
	case "*":
		x *= y
	case "/":
		x /= y
	case "^":
		x = math.Pow(x, y)
	default:
		return 0, ds, fmt.Errorf("invalid operator: %s", op)
	}
	if len(ds.colors) > 1 {
		return 0, ds, fmt.Errorf("cannot preform aritimitic on different color dice, try \",\" or \"and\" instead")
	}
	color := ds.PopColor()
	for i := 0; i < newDice; i++ {
		ds.Top(i).Color = color
	}
	ds.AddToColor(color, x)
	return x, ds, nil
}

func (t *ast) evaluateBoolean(ds *DiceSet) (bool, *DiceSet, error) {
	left, ds, err := t.children[0].eval(ds)
	if err != nil {
		return false, ds, err
	}
	right, ds, err := t.children[1].eval(ds)
	if err != nil {
		return false, ds, err
	}
	switch t.sym {
	case ">":
		return left > right, ds, nil
	case "<":
		return left < right, ds, nil
	case "<=":
		return left <= right, ds, nil
	case ">=":
		return left >= right, ds, nil
	case "==":
		return left == right, ds, nil
	case "!=":
		return left != right, ds, nil
	}
	return false, ds, fmt.Errorf("Bad bool")
}

type Dice struct {
	Count       int64
	Sides       int64
	resultTotal int64
	Faces       []int64
	Max         int64
	Min         int64
	KeepHighest int64
	KeepLowest  int64
	Color       string
}
type DiceSet struct {
	Dice          []Dice
	TotalsByColor map[string]float64
	keepHighest   int64
	leepLowest    int64
	colors        []string
	colorDepth    int
}

//PushAndRoll adds a dice roll to the "stack" applying any values from the set
func (d *DiceSet) PushAndRoll(dice Dice) (int64, error) {
	if d.colorDepth == 0 {
		dice.Color = d.PopColor()
	} else {
		dice.Color = d.PeekColor()
	}
	dice.KeepHighest = d.keepHighest
	dice.KeepLowest = d.leepLowest
	d.leepLowest = 0
	d.keepHighest = 0
	res, err := dice.Roll()
	if err != nil {
		return 0, err
	}
	d.Dice = append(d.Dice, dice)
	d.AddToColor(dice.Color, float64(res))
	return res, nil
}

//PushColor pushes a color to the "stack"
func (d *DiceSet) PushColor(color string) {
	d.colors = append(d.colors, color)
}

//PeekColor returns the most recently added color from the "stack"
func (d *DiceSet) PeekColor() string {
	if len(d.colors) > 0 {
		color := d.colors[len(d.colors)-1]
		return color
	}
	return ""
}

//PopColor pops a color from the "stack"
func (d *DiceSet) PopColor() string {

	if len(d.colors) > 0 {
		color := d.colors[len(d.colors)-1]
		d.colors = d.colors[:len(d.colors)-1]
		return color
	}
	return ""
}

//Top returns a pointer to the most recently added dice roll
func (d *DiceSet) Top(loc int) *Dice {
	if len(d.Dice) > 0 {
		return &d.Dice[len(d.Dice)-loc-1]
	}
	return nil
}

//AddToColor increments the total result for a given color
func (d *DiceSet) AddToColor(color string, value float64) {
	if d.TotalsByColor == nil {
		d.TotalsByColor = make(map[string]float64)
	}
	if d.colorDepth == 0 {
		d.TotalsByColor[color] += value
	}
}

func (d *Dice) Roll() (int64, error) {
	if d.resultTotal != 0 {
		return d.resultTotal, nil
	}
	faces, result, err := roll(d.Count, d.Sides, d.KeepHighest, d.KeepLowest)
	if err != nil {
		return 0, err
	}
	d.Min = d.Count
	if d.KeepHighest > 0 || d.KeepLowest > 0 {
		d.Max = (d.KeepHighest + d.KeepLowest) * d.Sides
	} else {
		d.Max = d.Count * d.Sides
	}
	d.Faces = faces
	d.resultTotal = result
	return result, nil
}

// func (d *Dice) Probability() float64 {
// 	if d.resultTotal > 0 {
// 		return diceProbability(d.Count, d.Sides, d.resultTotal)
// 	}
// 	return 0
// }

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
		if H > 0 {
			keptFaces := faces[:H]
			total = sumInt64(keptFaces...)
		} else if L > 0 {
			keptFaces := faces[L:]
			total = sumInt64(keptFaces...)
		}
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

func diceProbability(numberOfDice int64, sides int64, target int64, H int64, L int64) float64 {
	rollAmount := math.Pow(float64(sides), float64(numberOfDice))
	targetAmount := float64(0)
	var possibilities []int64
	for i := int64(1); i <= sides; i++ {
		possibilities = append(possibilities, i)
	}
	c := make(chan []int64)
	go generateProducts(c, possibilities, numberOfDice)
	for product := range c {
		if H > 0 {
			sort.Slice(product, func(i, j int) bool { return product[i] < product[j] })
			product = product[:H]
		} else if L > 0 {
			sort.Slice(product, func(i, j int) bool { return product[i] < product[j] })
			product = product[L:]
		}
		if sumInt64(product...) == target {
			targetAmount++
		}
	}
	return (targetAmount / rollAmount)
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
