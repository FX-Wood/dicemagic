package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aasmall/dicemagic/roll"
)

func Test_stringToColor(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{name: "random color",
		args: args{input: "random"},
		want: "#2F164D"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringToColor(tt.args.input); got != tt.want {
				t.Errorf("stringToColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rollExpressionToSlackAttachment(t *testing.T) {
	type args struct {
		expression *roll.RollExpression
	}
	tests := []struct {
		name    string
		args    args
		want    Attachment
		wantErr bool
	}{

		{name: "multi-d",
			args:    args{expression: roll.NewParser(strings.NewReader("Roll 3d3d10")).MustParseTotal()},
			want:    Attachment{},
			wantErr: false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rollExpressionToSlackAttachment(tt.args.expression)
			if (err != nil) != tt.wantErr {
				t.Errorf("rollExpressionToSlackAttachment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rollExpressionToSlackAttachment() = %v, want %v", got, tt.want)
			}
		})
	}
}
