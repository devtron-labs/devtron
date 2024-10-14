package util

import "testing"

func TestIsValidUrlSubPath(t *testing.T) {
	type args struct {
		subPath string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Test Case 1: Invalid URL", args: args{subPath: "%!d(string=)/5.zip%!(EXTRA int=5)"}, want: false},
		{name: "Test Case 2: Valid URL", args: args{subPath: "5/5.zip"}, want: true},
		{name: "Test Case 3: Valid URL", args: args{subPath: "prefix/1/5.zip"}, want: true},
		{name: "Test Case 5: Valid URL", args: args{subPath: "/prefix/1/5.zip"}, want: true},
		{name: "Test Case 6: Valid URL", args: args{subPath: "//1/5.zip"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidUrlSubPath(tt.args.subPath); got != tt.want {
				t.Errorf("IsValidUrlSubPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
