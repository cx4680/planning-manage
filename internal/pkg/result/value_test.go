package result

import (
	"testing"
)

func Test_IsNilData(t *testing.T) {
	type args struct {
		data interface{}
	}
	var a interface{}

	structFunc := func() interface{} {
		return args{data: "data"}
	}

	nilFunc := func() interface{} {
		return nil
	}

	mapFunc := func() interface{} {
		return map[string]string{}
	}

	pointerFunc := func() interface{} {
		var a *args
		return a
	}

	arrayFunc := func() interface{} {
		var a [2]string
		return a
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil interface",
			args: args{data: a},
			want: true,
		},
		{
			name: "default struct",
			args: args{data: args{}},
			want: false,
		},
		{
			name: "func",
			args: args{data: func() {

			}},
			want: false,
		},
		{
			name: "struct func",
			args: args{data: structFunc()},
			want: false,
		},
		{
			name: "nilFunc",
			args: args{data: nilFunc()},
			want: true,
		},
		{
			name: "map",
			args: args{data: map[string]string{}},
			want: false,
		},
		{
			name: "mapFunc",
			args: args{data: mapFunc()},
			want: false,
		},
		{
			name: "pointerFunc",
			args: args{data: pointerFunc()},
			want: true,
		},
		{
			name: "arrayFunc",
			args: args{data: arrayFunc()},
			want: false,
		},
		{
			name: "slice",
			args: args{data: []string{"a", "b"}},
			want: false,
		},
		{
			name: "array",
			args: args{data: [2]interface{}{nil, "a"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNil(tt.args.data); got != tt.want {
				t.Errorf("nilData() = %v, want %v", got, tt.want)
			}
		})
	}
}
