/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

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
