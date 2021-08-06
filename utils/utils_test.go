package utils

import (
	"testing"
)

func TestMatch(t *testing.T) {
	type args struct {
		pattern []string
		value   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			args: args{[]string{""}, "x"},
			want: false,
		},
		{
			args: args{[]string{"*"}, "x"},
			want: true,
		},
		{
			args: args{[]string{"^*"}, "x"},
			want: false,
		},
		{
			args: args{[]string{"x"}, "x"},
			want: true,
		},
		{
			args: args{[]string{"x*"}, "xy"},
			want: true,
		},
		{
			args: args{[]string{"*x"}, "yx"},
			want: true,
		},
		{
			args: args{[]string{"*x*"}, "yxy"},
			want: true,
		},
		{
			args: args{[]string{"x"}, "y"},
			want: false,
		},
		{
			args: args{[]string{"x*", "^xy"}, "xy"},
			want: false,
		},
		{
			args: args{[]string{"xy", "^x*"}, "xy"},
			want: true,
		},
		{
			args: args{[]string{"xy", "xz"}, "xz"},
			want: true,
		},
		{
			args: args{[]string{"^z"}, "xy"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Match(tt.args.pattern, tt.args.value); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
