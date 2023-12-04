package data

import (
	"reflect"
	"testing"
)

func TestSplitCommaString(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty string",
			args: args{str: ""},
			want: nil,
		},
		{
			name: "plain string",
			args: args{str: "str"},
			want: []string{"str"},
		},
		{
			name: "comma string",
			args: args{str: "stra,stringb"},
			want: []string{"stra", "stringb"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitCommaString(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitCommaString() = %v, want %v", got, tt.want)
			}
		})
	}
}
