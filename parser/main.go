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
			//printAST(stmt, 0)
			i, err := Evaluate(stmt)
			if err != nil {
				panic(err.Error())
			}
			fmt.Printf("evaluated:%.2f\n", i)
		}
	}
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

type DiceExpression struct {
}

func preformArithmitic(t *ast, op string) (float64, error) {
	var x float64
	x, err := Evaluate(t.children[0])
	if err != nil {
		return 0, err
	}
	for i := 1; i < len(t.children); i++ {
		y, err := Evaluate(t.children[i])
		if err != nil {
			return 0, err
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
			return 0, fmt.Errorf("invalid operator: %s", op)
		}
	}
	return x, nil
}
func Evaluate(t *ast) (float64, error) {
	switch t.sym {
	case "(NUMBER)":
		i, _ := strconv.ParseFloat(t.value, 64)
		return i, nil
	case "d":
		dice := Dice{}
		//d is always binary
		l, err := Evaluate(t.children[0])
		dice.Count = int64(l)
		if err != nil {
			return 0, err
		}
		r, err := Evaluate(t.children[1])
		dice.Sides = int64(r)
		if err != nil {
			return 0, err
		}
		// if K or L
		if len(t.children) > 2 {
			var x float64
			var intx int64
			if len(t.children[2].children) != 0 {
				x, err = Evaluate(t.children[2])
				if err != nil {
					return 0, err
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
				return 0, fmt.Errorf("unsupported dice modifier found: \"%s\"", s)
			}
		}
		//actually roll dice here
		res, err := dice.Roll()
		fmt.Printf("%+v\n", dice)
		return float64(res), err
	case "+", "-", "*", "/", "^", "-H", "-L":
		x, err := preformArithmitic(t, t.sym)
		if err != nil {
			return 0, err
		}
		return x, nil
	case "{":
		var x float64
		for _, c := range t.children {
			y, err := Evaluate(c)
			if err != nil {
				return 0, err
			}
			x += y
		}
		return x, nil
	case "(IDENT)":
		var x float64
		for _, c := range t.children {
			y, err := Evaluate(c)
			if err != nil {
				return 0, err
			}
			x += y
		}
		return x, nil
	case "roll":
		var x float64
		for _, c := range t.children {
			y, err := Evaluate(c)
			if err != nil {
				return 0, err
			}
			x += y
		}
		return x, nil
	case "if":
		res, err := EvaluateBoolean(t.children[0])
		if err != nil {
			return 0, err
		}
		fmt.Print(res, " ")
		var c *ast
		if res {
			c = t.children[1]
		} else {
			if len(t.children) < 3 {
				return 0, nil
			}
			c = t.children[2]
		}
		var x float64
		for i := 0; i < len(c.children); i++ {
			y, err := Evaluate(c.children[i])
			if err != nil {
				return 0, err
			}
			x += y
		}
		return x, nil
	default:
		return 0, fmt.Errorf("Unsupported symbol: %s", t.sym)
	}
	return 0, fmt.Errorf("bad ast")
}
func EvaluateBoolean(t *ast) (bool, error) {
	switch t.sym {
	case ">":
		l, err := Evaluate(t.children[0])
		if err != nil {
			return false, err
		}
		r, err := Evaluate(t.children[1])
		if err != nil {
			return false, err
		}
		if l > r {
			return true, nil
		}
		return false, nil
	}
	return false, fmt.Errorf("Bad bool")
}

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
	if d.H > 0 || d.L > 0 {
		d.Max = (d.H + d.L) * d.Sides
	} else {
		d.Max = d.Count * d.Sides
	}
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
