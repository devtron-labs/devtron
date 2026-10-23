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

package devtronApps

import "testing"

func TestReplaceImageTagWithDigest(t *testing.T) {
	const digest = "sha256:0123456789abcdef"
	tests := []struct {
		name  string
		image string
		want  string
	}{
		{"tag", "nginx:1.25", "nginx@" + digest},
		{"repository path and tag", "team/app:v1", "team/app@" + digest},
		{"registry and tag", "registry.example.com/team/app:v1", "registry.example.com/team/app@" + digest},
		{"registry with port and tag", "registry.local:5000/team/app:v1", "registry.local:5000/team/app@" + digest},
		{"registry with port and no tag", "registry.local:5000/team/app", "registry.local:5000/team/app@" + digest},
		{"no tag", "team/app", "team/app@" + digest},
		{"tag and old digest", "registry.local:5000/team/app:v1@sha256:fedcba", "registry.local:5000/team/app@" + digest},
		{"old digest only", "registry.local:5000/team/app@sha256:fedcba", "registry.local:5000/team/app@" + digest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplaceImageTagWithDigest(tt.image, digest); got != tt.want {
				t.Errorf("ReplaceImageTagWithDigest(%q) = %q, want %q", tt.image, got, tt.want)
			}
		})
	}
}
