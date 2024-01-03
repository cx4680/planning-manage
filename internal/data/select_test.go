package data

import (
	"reflect"
	"testing"
)

type test struct {
	name string `gorm:"column:name"`
	addr string `gorm:"column:addr"`
}

func TestSelect(t *testing.T) {
	type args struct {
		source any
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "show column",
			args: args{source: test{}},
			want: []string{"name", "addr"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectColumn(tt.args.source); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SelectColumn() = %v, want %v", got, tt.want)
			}
		})
	}
}
