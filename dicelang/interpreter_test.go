package dicelang

import (
	"reflect"
	"testing"
)

func TestAST_GetDiceSet(t *testing.T) {
	roll1d20AST, _ := NewParser("roll 20d1 mundane").Statements()
	tests := []struct {
		name    string
		t       *AST
		want    float64
		want1   DiceSet
		wantErr bool
	}{
		{name: "roll 1d20 mundane",
			t:    roll1d20AST[0],
			want: 20,
			want1: DiceSet{
				Dice: []Dice{
					Dice{
						Count:       20,
						Sides:       1,
						Total:       20,
						Faces:       []int64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
						Max:         20,
						Min:         20,
						DropHighest: 0,
						DropLowest:  0,
						Color:       "Mundane"}},
				TotalsByColor: map[string]float64{"Mundane": float64(20)},
				dropHighest:   0,
				dropLowest:    0,
				colors:        []string{},
				colorDepth:    0},
			wantErr: false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.t.GetDiceSet()
			if (err != nil) != tt.wantErr {
				t.Errorf("AST.GetDiceSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AST.GetDiceSet() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("AST.GetDiceSet() got1 = %+v, want %+v", got1, tt.want1)
			}
		})
	}
}
