/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package util

import "testing"

func TestCompareUnOrdered(t *testing.T) {
	type args struct {
		a []int
		b []int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test ordered",
			args: args{
				a: []int{2, 3, 5},
				b: []int{2, 3, 5},
			},
			want: true,
		},
		{name: "test unordered",
			args: args{
				a: []int{2, 3, 5},
				b: []int{2, 5, 3},
			},
			want: true,
		},
		{name: "test unequal",
			args: args{
				a: []int{2, 3, 0},
				b: []int{2, 5, 3},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareUnOrdered(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("CompareUnorderd() = %v, want %v", got, tt.want)
			}
		})
	}
}
