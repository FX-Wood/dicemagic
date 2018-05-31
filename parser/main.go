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
			v, ds := stmt.GetDiceSet()
			//printAST(stmt, 0)
			for _, v := range ds.Dice {
				fmt.Printf("%+v\n", v)
			}
			fmt.Printf("for a total of: %+v\n", v)
			fmt.Println("----------")
		}
	}
}

func (t *ast) GetDiceSet() (float64, DiceSet) {
	v, ret, _ := t.eval(&DiceSet{})
	ret.Dice = append(ret.Dice, t.Dice...)
	return v, *ret
}

func (t *ast) eval(ds *DiceSet) (float64, *DiceSet, error) {
	switch t.sym {
	case "(NUMBER)":
		i, _ := strconv.ParseFloat(t.value, 64)
		if len(t.children) > 0 {
			t.children[0].eval(ds)
		}
		return i, ds, nil
	case "d":
		dice := Dice{}
		//d is always binary
		l, _, err := t.children[0].eval(ds)
		dice.Count = int64(l)
		if err != nil {
			return 0, nil, err
		}
		r, _, err := t.children[1].eval(ds)
		dice.Sides = int64(r)
		if err != nil {
			return 0, nil, err
		}
		// if K or L
		if len(t.children) > 2 {
			var x float64
			var intx int64
			if len(t.children[2].children) != 0 {
				x, _, err = t.children[2].eval(ds)
				if err != nil {
					return 0, ds, err
				}
				intx = int64(x)
			} else {
				intx = dice.Count - 1
			}
			switch s := t.children[2].sym; s {
			case "-H":
				dice.H = intx
			case "-L":
				dice.L = intx
			default:
				return 0, ds, fmt.Errorf("unsupported dice modifier found: \"%s\"", s)
			}
		}
		//actually roll dice here
		res, err := dice.Roll()
		ds.Add(dice)
		ds.color = ""
		return float64(res), ds, err
	case "+", "-", "*", "/", "^", "-H", "-L":
		x, ds, err := t.preformArithmitic(ds, t.sym)

		if err != nil {
			return 0, ds, err
		}
		if len(ds.Dice) > 0 {
			ds.Dice[len(ds.Dice)-1].Color = ds.color
		}
		return x, ds, nil
	case "{", "and", ",":
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
		ds.color = t.value
		return 0, ds, nil
	case "roll":
		var x float64
		for _, c := range t.children {
			y, ds, err := c.eval(ds)
			if err != nil {
				return 0, ds, err
			}
			x += y
		}
		return x, ds, nil
	case "if":
		res, ds, err := t.evaluateBoolean(ds)
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
		for i := 0; i < len(c.children); i++ {
			y, ds, err := c.children[i].eval(ds)
			if err != nil {
				return 0, ds, err
			}
			x += y
		}
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
	var x float64
	x, ds, err := t.children[0].eval(ds)
	if err != nil {
		return 0, ds, err
	}
	for i := 1; i < len(t.children); i++ {
		y, ds, err := t.children[i].eval(ds)
		if err != nil {
			return 0, ds, err
		}
		switch op {
		case "+", "-K", "-L":
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
	}
	return x, ds, nil
}

func (t *ast) evaluateBoolean(ds *DiceSet) (bool, *DiceSet, error) {
	switch t.sym {
	case ">":
		l, ds, err := t.children[0].eval(ds)
		if err != nil {
			return false, ds, err
		}
		r, ds, err := t.children[1].eval(ds)
		if err != nil {
			return false, ds, err
		}
		if l > r {
			return false, ds, nil
		}
		return false, ds, nil
	}
	return false, ds, fmt.Errorf("Bad bool")
}

// func sumMaps(maps ...map[string]float64) map[string]float64 {
// 	ret := make(map[string]float64)
// 	for _, m := range maps {
// 		for k, v := range m {
// 			ret[k] += v
// 		}
// 	}
// 	return ret
// }

// func Results(t *ast) (DiceSet, error) {
// 	ret := DiceSet{}

// switch sym := t.sym; sym {
// case "and", ",", "roll", "{":
// 	for _, c := range t.children {
// 		r, err := Results(c)
// 		if err != nil {
// 			return ret, err
// 		}
// 		ret.Dice = append(ret.Dice, r.Dice...)
// 	}
// 	return ret, nil
// case "if":
// 	res, err := EvaluateBoolean(t.children[0])
// 	if err != nil {
// 		return ret, err
// 	}
// 	fmt.Print(res, " ")
// 	var c *ast
// 	if res {
// 		c = t.children[1]
// 	} else {
// 		if len(t.children) < 3 {
// 			return ret, nil
// 		}
// 		c = t.children[2]
// 	}
// 	for i := 0; i < len(c.children); i++ {
// 		r, err := Results(c.children[i])
// 		if err != nil {
// 			return ret, err
// 		}
// 		ret.Dice = append(ret.Dice, r.Dice...)
// 	}
// 	return ret, nil
// case "+", "-", "*", "/", "^", "-H", "-L", "(NUMBER)", "(IDENT)", "d":
// 	r, b, _, _, err := EvaluateExpression(t)
// 	if err != nil {
// 		return ret, err
// 	}
// 	if b {
// 		ret.Dice = append(ret.Dice, r)
// 	}
// 	return ret, nil
// default:
// 	return ret, fmt.Errorf("Unsupported symbol: %s", t.sym)
// }
// }
// func EvaluateExpression(t *ast) (Dice, bool, float64, string, error) {
// 	var (
// 		ret   float64
// 		color string
// 		err   error
// 		d     Dice
// 	)
// 	switch t.sym {
// 	case "(NUMBER)":
// 		ret, err := strconv.ParseFloat(t.value, 64)
// 		if err != nil {
// 			return d, false, 0, "", err
// 		}
// 		if len(t.children) > 0 {
// 			_, _, _, color, err = EvaluateExpression(t.children[0])
// 			if err != nil {
// 				return d, false, 0, "", err
// 			}
// 		}
// 		return d, false, ret, color, nil

// 	case "(IDENT)":
// 		return d, false, 0, t.value, nil
// 	case "+", "-", "*", "/", "^", "-H", "-L":
// 		ret, color, err = preformArithmitic(t, t.sym)
// 		if err != nil {
// 			return d, false, 0, "", err
// 		}
// 		return d, false, ret, color, nil
// 	case "d":
// 		d, err = EvaluateDice(t)
// 		if err != nil {
// 			return d, false, 0, "", err
// 		}
// 		return d, true, float64(d.resultTotal), d.Color, nil
// 	}
// 	return d, false, ret, color, fmt.Errorf("Invalid Expression: %s:%s", t.sym, t.value)
// }

// func EvaluateDice(t *ast) (Dice, error) {
// 	var (
// 		color  string
// 		lColor string
// 		rColor string
// 		err    error
// 		left   float64
// 		right  float64
// 	)
// 	dice := Dice{}
// 	if len(t.children) > 3 {
// 		return dice, fmt.Errorf("dice has more than 3 children")
// 	}
// 	_, _, left, lColor, err = EvaluateExpression(t.children[0])
// 	dice.Count = int64(left)
// 	if err != nil {
// 		return dice, err
// 	}
// 	_, _, right, rColor, err = EvaluateExpression(t.children[1])
// 	dice.Sides = int64(right)
// 	if err != nil {
// 		return dice, err
// 	}
// 	color = rColor + lColor
// 	dice.Color = color
// 	// if K or L
// 	if len(t.children) > 2 {
// 		var x float64
// 		var intx int64
// 		if len(t.children[2].children) != 0 {
// 			_, _, x, _, err = EvaluateExpression(t.children[2])
// 			if err != nil {
// 				return dice, err
// 			}
// 			intx = int64(x)
// 		} else {
// 			intx = dice.Count - 1
// 		}
// 		switch s := t.children[2].sym; s {
// 		case "-H":
// 			dice.H = intx
// 		case "-L":
// 			dice.L = intx
// 		default:
// 			return dice, fmt.Errorf("unsupported dice modifier found: \"%s\"", s)
// 		}
// 	}
// 	//actually roll dice here
// 	_, err = dice.Roll()
// 	fmt.Printf("%+v\n", dice)
// 	if err != nil {
// 		return dice, err
// 	}
// 	return dice, err
// }

// func EvaluateBoolean(t *ast) (bool, error) {
// 	_, _, l, _, err := EvaluateExpression(t.children[0])
// 	if err != nil {
// 		return false, err
// 	}
// 	_, _, r, _, err := EvaluateExpression(t.children[1])
// 	if err != nil {
// 		return false, err
// 	}
// 	switch t.sym {
// 	case ">":
// 		if l > r {
// 			return true, nil
// 		}
// 		return false, nil
// 	case "<":
// 		if l < r {
// 			return true, nil
// 		}
// 		return false, nil
// 	}
// 	return false, fmt.Errorf("Bad boolean operator: %s", t.sym)
// }

type Dice struct {
	Count       int64
	Sides       int64
	resultTotal int64
	Faces       []int64
	Max         int64
	Min         int64
	H           int64
	L           int64
	Color       string
	Expression  *ast
}
type DiceSet struct {
	Dice            []Dice
	ResultTotals    []int64
	DiceType        string
	Min             int64
	Max             int64
	color           string
	UnaryExpression float64
}

func (d *DiceSet) Add(dice Dice) {
	dice.Color = d.color
	d.Dice = append(d.Dice, dice)
}

// func (d *DiceSet) Roll() ([]int64, error) {
// 	d.Min, d.Max = 0, 0
// 	for _, ds := range d.Dice {
// 		result, err := ds.Roll()
// 		if err != nil {
// 			return nil, err
// 		}
// 		d.Min += ds.Count
// 		d.Max += (ds.Count - (ds.H + ds.L)) * ds.Sides
// 		d.ResultTotals = append(d.ResultTotals, result)
// 	}
// 	return d.ResultTotals, nil
// }
func (d *Dice) Roll() (int64, error) {
	if d.resultTotal != 0 {
		return d.resultTotal, nil
	}
	faces, result, err := roll(d.Count, d.Sides, d.H, d.L)
	if err != nil {
		return 0, err
	}
	d.Min = d.Count
	if d.H > 0 || d.L > 0 {
		d.Max = (d.H + d.L) * d.Sides
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
			f := faces[int64(len(faces))-H:]
			total = sumFaces(f)
		} else if L > 0 {
			f := faces[:L]
			total = sumFaces(f)
		}
		return faces, total, nil
	}
}
func sumFaces(faces []int64) int64 {
	var total int64
	for _, f := range faces {
		total += f
	}
	return total
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
