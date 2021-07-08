package utils

import "testing"

func TestMatch(t *testing.T) {
	type args struct {
		format string
		value  string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			args: args{"", "x"},
			want: false,
		},
		{
			args: args{"*", "x"},
			want: true,
		},
		{
			args: args{"^*", "x"},
			want: false,
		},
		{
			args: args{"x", "x"},
			want: true,
		},
		{
			args: args{"x*", "xy"},
			want: true,
		},
		{
			args: args{"*x", "yx"},
			want: true,
		},
		{
			args: args{"*x*", "yxy"},
			want: true,
		},
		{
			args: args{"x", "y"},
			want: false,
		},
		{
			args: args{"x*,^xy", "xy"},
			want: false,
		},
		{
			args: args{"xy,^x*", "xy"},
			want: true,
		},
		{
			args: args{"^z", "xy"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Match(tt.args.format, tt.args.value); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
