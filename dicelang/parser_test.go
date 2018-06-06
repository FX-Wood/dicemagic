package dicelang

import (
	"reflect"
	"testing"
)

func TestParser_Statements(t *testing.T) {
	tests := []struct {
		name    string
		parse   *Parser
		want    []*AST
		wantErr bool
	}{
		{name: "roll 1d20",
			parse:   NewParser("roll 1d20"),
			want:    nil,
			wantErr: false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.parse.Statements()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.Statements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.Statements() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
