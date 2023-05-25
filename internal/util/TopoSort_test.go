/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
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
