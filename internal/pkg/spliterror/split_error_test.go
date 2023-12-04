package spliterror

import (
	"testing"
)

func TestSplitError(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "init log",
			args: args{str: "JobM.InvalidData"},
			want: "InvalidData",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitError(tt.args.str); got != tt.want {
				t.Errorf("SplitError() = %v, want %v", got, tt.want)
			}
		})
	}
}
