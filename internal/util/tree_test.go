/*
 * Copyright (c) 2024. Devtron Inc.
 */

package util

import (
	"reflect"
	"testing"
)

var tree = map[int][]int{
	1: {2, 3},
	2: {4, 5, 6},
	3: {7},
	7: {8},
	8: {9},
}

func TestIsAncestor(t *testing.T) {
	tests := []struct {
		name  string
		tree  map[int][]int
		nodeA int
		nodeB int
		want  bool
	}{
		{name: "test1",
			tree:  tree,
			nodeA: 1,
			nodeB: 9,
			want:  true,
		},
		{
			name:  "test2",
			tree:  tree,
			nodeA: 2,
			nodeB: 8,
			want:  false,
		},
		{
			name:  "test2",
			tree:  tree,
			nodeA: 2,
			nodeB: 6,
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAncestor(tt.tree, tt.nodeA, tt.nodeB); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IsAncestor() = %v, want %v", got, tt.want)
			}
		})
	}
}
