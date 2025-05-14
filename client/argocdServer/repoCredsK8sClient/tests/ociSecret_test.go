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

package tests

import (
	"github.com/devtron-labs/devtron/client/argocdServer/repoCredsK8sClient"
	"testing"
)

func Test_OCIArgoSecretRepoPathAndHostParseLogic(t *testing.T) {
	type args struct {
		repositoryURL  string
		repositoryName string
	}
	tests := []struct {
		name                 string
		args                 args
		expectedHost         string
		expectedFullRepoPath string
	}{
		{
			name: "case 1",
			args: args{
				repositoryURL:  "docker.io/bitnamicharts",
				repositoryName: "bitnami",
			},
			expectedHost:         "docker.io",
			expectedFullRepoPath: "bitnamicharts/bitnami",
		},
		{
			name: "case 2",
			args: args{
				repositoryURL:  "docker.io",
				repositoryName: "bitnamicharts/bitnami",
			},
			expectedHost:         "docker.io",
			expectedFullRepoPath: "bitnamicharts/bitnami",
		},
		{
			name: "case 3",
			args: args{
				repositoryURL:  "oci://docker.io",
				repositoryName: "bitnamicharts/bitnami",
			},
			expectedHost:         "docker.io",
			expectedFullRepoPath: "bitnamicharts/bitnami",
		},
		{
			name: "case 4",
			args: args{
				repositoryURL:  "https://4.123.13.1/foo/bar",
				repositoryName: "chart",
			},
			expectedHost:         "4.123.13.1",
			expectedFullRepoPath: "foo/bar/chart",
		},
		{
			name: "case 5",
			args: args{
				repositoryURL:  "http://4.123.13.1/foo/bar",
				repositoryName: "chart",
			},
			expectedHost:         "4.123.13.1",
			expectedFullRepoPath: "foo/bar/chart",
		},
		{
			name: "case 6",
			args: args{
				repositoryURL:  "https://4.123.13.1",
				repositoryName: "foo/bar/chart",
			},
			expectedHost:         "4.123.13.1",
			expectedFullRepoPath: "foo/bar/chart",
		},
		{
			name: "case 7",
			args: args{
				repositoryURL:  "http://4.123.13.1",
				repositoryName: "foo/bar/chart",
			},
			expectedHost:         "4.123.13.1",
			expectedFullRepoPath: "foo/bar/chart",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if host, fullRepoPath, err := repoCredsK8sClient.GetHostAndFullRepoPath(tt.args.repositoryURL, tt.args.repositoryName); err != nil || host != tt.expectedHost || fullRepoPath != tt.expectedFullRepoPath {
				t.Errorf("SanitizeRepoNameAndURLForOCIRepo() = repositoryURL: %v , repositoryName: %v, want  %v %v", host, fullRepoPath, tt.expectedHost, tt.expectedFullRepoPath)
			}
		})
	}
}
