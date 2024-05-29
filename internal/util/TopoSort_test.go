/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package util

import (
	"reflect"
	"testing"
)

func TestSort(t *testing.T) {
	tests := []struct {
		name string
		args map[int][]int
		want []int
	}{
		{name: "test1",
			args: map[int][]int{
				1: {2, 3},
				2: {},
				3: {2},
			},
			want: []int{1, 3, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TopoSort(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TopSort() = %v, want %v", got, tt.want)
			}
		})
	}
}
