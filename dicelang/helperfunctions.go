package dicelang

func MergeDiceTotalMaps(mapsToMerge ...map[string]float64) map[string]float64 {
	retMap := make(map[string]float64)
	for _, mp := range mapsToMerge {
		for k, v := range mp {
			retMap[k] += v
		}
	}
	return retMap
}

//GetDiceSet returns the sum of an AST, a DiceSet, and an error
func (t *AST) GetDiceSet() (float64, DiceSet, error) {
	v, ret, err := t.eval(&DiceSet{})
	return v, *ret, err
}

//GetDiceSets merges all statements in ...*AST and returns a merged diceTotalMap and all rolled dice.
func GetDiceSets(stmts ...*AST) (map[string]float64, []Dice, error) {
	var maps []map[string]float64
	var dice []Dice

	for i := 0; i < len(stmts); i++ {
		_, ds, err := stmts[i].GetDiceSet()
		if err != nil {
			return nil, dice, err
		}
		maps = append(maps, ds.TotalsByColor)
		for _, d := range ds.Dice {
			dice = append(dice, d)
		}
	}
	return MergeDiceTotalMaps(maps...), dice, nil
}
