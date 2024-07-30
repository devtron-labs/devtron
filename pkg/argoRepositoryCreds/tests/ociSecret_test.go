package tests

import (
	"github.com/devtron-labs/devtron/pkg/argoRepositoryCreds"
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
			if host, fullRepoPath, err := argoRepositoryCreds.GetHostAndFullRepoPath(tt.args.repositoryURL, tt.args.repositoryName); err != nil || host != tt.expectedHost || fullRepoPath != tt.expectedFullRepoPath {
				t.Errorf("SanitizeRepoNameAndURLForOCIRepo() = repositoryURL: %v , repositoryName: %v, want  %v %v", host, fullRepoPath, tt.expectedHost, tt.expectedFullRepoPath)
			}
		})
	}
}
