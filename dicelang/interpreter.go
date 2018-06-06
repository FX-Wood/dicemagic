package dicelang

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"
)

//Dice represents a a throw of a single type of die
type Dice struct {
	Count       int64
	Sides       int64
	Total       int64
	Faces       []int64
	Max         int64
	Min         int64
	DropHighest int64
	DropLowest  int64
	Color       string
}

//DiceSet represents a collection of Dice and their totals by type
type DiceSet struct {
	Dice          []Dice
	TotalsByColor map[string]float64
	dropHighest   int64
	dropLowest    int64
	colors        []string
	colorDepth    int
}

//PrintAST prints a formatted version of the ast to StdOut
func PrintAST(t *AST, identation int) {
	fmt.Println()
	for i := 0; i < identation; i++ {
		fmt.Print(" ")
	}
	fmt.Print("(")
	fmt.Print(t.sym, ":", t.value)
	if len(t.children) > 0 {
		for _, c := range t.children {
			fmt.Print(" ")
			PrintAST(c, identation+4)
		}
	}
	fmt.Print(")")
}

//GetDiceSet returns the sum of an AST, a DiceSet, and an error
func (t *AST) GetDiceSet() (float64, DiceSet, error) {
	v, ret, err := t.eval(&DiceSet{})
	return v, *ret, err
}

func (t *AST) eval(ds *DiceSet) (float64, *DiceSet, error) {
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
			ds.dropHighest = intx
		case "-L":
			ds.dropLowest = intx
		}
		return 0, ds, nil
	case "d":
		dice := Dice{}
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
		var c *AST
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
}

func (t *AST) preformArithmitic(ds *DiceSet, op string) (float64, *DiceSet, error) {
	//arithmitic is always binary
	//...except for the "-" unary operator
	if len(t.children) < 2 {

	}
	diceCount := len(ds.Dice)
	var nums []float64
	ds.colorDepth++
	for _, c := range t.children {
		switch c.sym {
		case "(IDENT)":
			_, ds, err := c.eval(ds)
			if err != nil {
				return 0, ds, err
			}
		default:
			x, ds, err := c.eval(ds)
			if err != nil {
				return 0, ds, err
			}
			nums = append(nums, x)
		}
	}
	ds.colorDepth--
	newDice := len(ds.Dice) - diceCount
	var x float64
	switch op {
	case "+":
		for _, y := range nums {
			x += y
		}
	case "-":
		if len(nums) < 2 {
			x = -nums[0]
		} else {
			x = nums[0]
			for i := 1; i < len(nums); i++ {
				x -= nums[i]
			}
		}
	case "*":
		x = nums[0]
		for i := 1; i < len(nums); i++ {
			x *= nums[i]
		}
	case "/":
		x = nums[0]
		for i := 1; i < len(nums); i++ {
			x /= nums[i]
		}
	case "^":
		x = nums[0]
		for i := 1; i < len(nums); i++ {
			x = math.Pow(x, nums[i])
		}
	default:
		return 0, ds, fmt.Errorf("invalid operator: %s", op)
	}
	if len(ds.colors) > 1 {
		return 0, ds, fmt.Errorf("cannot preform aritimitic on different color dice, try \",\" or \"and\" instead")
	}
	if ds.colorDepth == 0 {
		color := ds.PopColor()
		for i := 0; i < newDice; i++ {
			ds.Top(i).Color = color
		}
		ds.AddToColor(color, x)
	}

	return x, ds, nil
}
func (t *AST) evaluateBoolean(ds *DiceSet) (bool, *DiceSet, error) {
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

//PushAndRoll adds a dice roll to the "stack" applying any values from the set
func (d *DiceSet) PushAndRoll(dice Dice) (int64, error) {
	if d.colorDepth == 0 {
		dice.Color = d.PopColor()
	}
	dice.DropHighest = d.dropHighest
	dice.DropLowest = d.dropLowest
	d.dropLowest = 0
	d.dropHighest = 0
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

//Roll rolls the dice, sets Min, Max, and Faces. Returns the total. Can be called multiple times and returns the same value each time.
func (d *Dice) Roll() (int64, error) {
	if d.Total != 0 {
		return d.Total, nil
	}
	faces, result, err := roll(d.Count, d.Sides, d.DropHighest, d.DropLowest)
	if err != nil {
		return 0, err
	}
	d.Min = d.Count
	if d.DropHighest > 0 || d.DropLowest > 0 {
		d.Max = (d.Count - (d.DropHighest + d.DropLowest)) * d.Sides
	} else {
		d.Max = d.Count * d.Sides
	}
	d.Faces = faces
	d.Total = result
	return result, nil
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
			keptFaces := faces[:int64(len(faces))-H]
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

func sumInt64(nums ...int64) int64 {
	r := int64(0)
	for _, n := range nums {
		r += n
	}
	return r
}
