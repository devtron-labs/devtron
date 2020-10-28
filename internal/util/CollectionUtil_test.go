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
