/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pipeline

import (
	"testing"

	"github.com/devtron-labs/devtron/internal/sql/repository"
)

func TestValidateAndConstructTag(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		counter   int
		sha       string
		want      string
		wantError bool
	}{
		{name: "counter only", pattern: "dev-{x}", counter: 5, want: "dev-4"},
		{name: "sha and counter", pattern: "dev-{sha}-{x}", counter: 18, sha: "a1b2c3d4", want: "dev-a1b2c3d4-17"},
		{name: "sha without counter", pattern: "dev-{sha}", counter: 1, sha: "a1b2c3d4", wantError: true},
		{name: "two sha variables", pattern: "{SHA}-{sha}-{x}", counter: 1, sha: "a1b2c3d4", wantError: true},
		{name: "sha missing at trigger", pattern: "dev-{sha}-{x}", counter: 1, sha: "", wantError: true},
		{name: "sha unused when absent from pattern", pattern: "dev-{x}", counter: 1, sha: "", want: "dev-0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateAndConstructTag(&repository.CustomTag{TagPattern: tt.pattern, AutoIncreasingNumber: tt.counter}, tt.sha)
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error, got tag %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
