package roll

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//RollExpression is a collection of Segments
type RollExpression struct {
	InitialText          string        `datastore:",noindex"`
	ExpandedTextTemplate string        `datastore:",noindex"`
	DiceSet              DiceSet       `datastore:",noindex"`
	SegmentHalfs         []SegmentHalf `datastore:",noindex"`
	RollTotals           []Total       `datastore:",noindex"`
	MathTree             MathTree
}
type MathTree struct {
	Root *MathNode
}
type MathNode struct {
	Left     *MathNode
	Right    *MathNode
	Operator string
	Data     MathData
}
type MathData struct {
	SegmentType string
	Value       int64
}

//SegmentHalf is half of a mathmatical expression along it's its evaluation priority
type SegmentHalf struct {
	Operator           string `datastore:",noindex"`
	Number             int64  `datastore:",noindex"`
	SegmentType        string `datastore:",noindex"`
	EvaluationPriority int
}

//Total represents collapsed Segments that have been evaluated
type Total struct {
	RollType   string
	RollResult int64
	Faces      []int64
}

func getHighestPriority(r []SegmentHalf) int {
	highestPriority := 0
	for _, e := range r {
		if e.EvaluationPriority < highestPriority {
			highestPriority = e.EvaluationPriority
		}

	}
	return highestPriority
}

func (r *RollExpression) String() string { return r.InitialText }
func (r *RollExpression) TotalsString() string {
	var buff bytes.Buffer
	rollTotal := int64(0)
	allUnspecified := true
	for _, t := range r.RollTotals {
		if t.RollType != "" {
			allUnspecified = false
		}
		rollTotal += t.RollResult
	}
	if allUnspecified {
		buff.WriteString("You rolled: ")
		buff.WriteString(strconv.FormatInt(rollTotal, 10))
	} else {
		for i, t := range r.RollTotals {
			buff.WriteString(strconv.FormatInt(t.RollResult, 10))
			buff.WriteString(" [")
			if t.RollType == "" {
				buff.WriteString("_Unspecified_")
			} else {
				buff.WriteString(t.RollType)
			}
			buff.WriteString("]")
			if i != len(r.RollTotals) {
				buff.WriteString("\n")
			} else {
				buff.WriteString("\nFor a total of: ")
				buff.WriteString(strconv.FormatInt(rollTotal, 10))
			}
		}
	}
	return buff.String()
}

//FormattedString returns formatted string for human consumption
func (r *RollExpression) FormattedString() string {
	_, err := r.DiceSet.Roll()
	if err != nil {
		return ""
	}
	var fmtString []interface{}
	for _, die := range r.DiceSet.Dice {
		fmtString = append(fmtString, fmt.Sprintf("%dd%d(%s)", die.Count, die.Sides, Int64SliceToCSV(die.Faces...)))
	}
	return fmt.Sprintf(r.ExpandedTextTemplate, fmtString...)
}

func Int64SliceToCSV(s ...int64) string {
	var buff bytes.Buffer
	for i := 0; i < len(s); i++ {
		buff.WriteString(strconv.FormatInt(s[i], 10))
		if i+1 != len(s) {
			buff.WriteString(", ")
		}
	}
	return buff.String()
}

//Total rolls all the dice and populates RollTotals and ExpandedText
func (r *RollExpression) Total() error {
	totalsMap := make(map[string]int64)
	facesMap := make(map[string][]int64)
	rollTotals := []Total{}
	//break segments into their Damage Types
	segmentsPerSegmentType := make(map[string][]SegmentHalf)
	for _, e := range r.SegmentHalfs {
		segmentsPerSegmentType[e.SegmentType] = append(segmentsPerSegmentType[e.SegmentType], e)
	}
	//for each damage type
	for k, remainingSegments := range segmentsPerSegmentType {
		// Establish highest priority (represented as lowest number)
		highestPriority := getHighestPriority(remainingSegments)
		var lastSegment SegmentHalf

		//loop through priorities
		for p := highestPriority; p < 1; p++ {
			for i := 0; i < len(remainingSegments); i++ {
				if !strings.ContainsAny(remainingSegments[i].Operator, "+-*/") {
					return fmt.Errorf("%s is not a valid operator", remainingSegments[i].Operator)
				}
				if remainingSegments[i].EvaluationPriority == p && len(remainingSegments) > 1 && i > 0 {
					var replacementSegment SegmentHalf
					if remainingSegments[i].Operator == "d" {
						d := Dice{Count: lastSegment.Number, Sides: remainingSegments[i].Number}
						result, err := d.Roll()
						if err != nil {
							return err
						}
						r.DiceSet.Dice = append(r.DiceSet.Dice, d)
						replacementSegment = SegmentHalf{Operator: lastSegment.Operator, EvaluationPriority: lastSegment.EvaluationPriority, Number: result}
						for _, face := range d.Faces {
							facesMap[k] = append(facesMap[k], face)
						}
					} else {
						var err error
						replacementSegment, err = doMath(lastSegment, remainingSegments[i])
						if err != nil {
							return err
						}
					}
					remainingSegments = insertAtLocation(deleteAtLocation(remainingSegments, i-1, 2), replacementSegment, i-1)
					lastSegment = replacementSegment
					i--
				} else {
					lastSegment = remainingSegments[i]
				}
			}
		}
		//I have fully collapsed this loop. Add to final result.
		totalsMap[k] += int64(lastSegment.Number)
	}

	//sort it
	var keys []string
	for k := range totalsMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		rollTotals = append(rollTotals, Total{RollType: k, RollResult: totalsMap[k], Faces: facesMap[k]})
	}
	r.DiceSet.Roll()
	r.RollTotals = rollTotals
	return nil
}

func deleteAtLocation(segment []SegmentHalf, location int, numberToDelete int) []SegmentHalf {
	return append(segment[:location], segment[location+numberToDelete:]...)
}
func insertAtLocation(segment []SegmentHalf, segmentToInsert SegmentHalf, location int) []SegmentHalf {
	segment = append(segment, segmentToInsert)
	copy(segment[location+1:], segment[location:])
	segment[location] = segmentToInsert
	return segment
}

// func doMath(leftMod SegmentHalf, rightmod SegmentHalf) (SegmentHalf, error) {
// 	m := SegmentHalf{}
// 	switch rightmod.Operator {
// 	case "*":
// 		m.Number = leftMod.Number * rightmod.Number
// 	case "/":
// 		if rightmod.Number == 0 {
// 			return m, fmt.Errorf("Don't make me break the universe.")
// 		}
// 		m.Number = leftMod.Number / rightmod.Number
// 	case "+":
// 		m.Number = leftMod.Number + rightmod.Number
// 	case "-":
// 		m.Number = leftMod.Number - rightmod.Number
// 	case "d":
// 		_, num, err := roll(leftMod.Number, rightmod.Number, )
// 		m.Number = num
// 		if err != nil {
// 			return m, err
// 		}
// 	}
// 	m.Operator = leftMod.Operator
// 	m.EvaluationPriority = leftMod.EvaluationPriority
// 	return m, nil
// }
