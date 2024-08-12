package appStoreDeploymentCommon

import (
	"testing"
)

func Test_checkDependencyValuesForOCIRepo(t *testing.T) {
	type args struct {
		repositoryURL  string
		repositoryName string
	}
	tests := []struct {
		name                   string
		args                   args
		expectedRepositoryURL  string
		expectedRepositoryName string
	}{
		{
			name: "case 1",
			args: args{
				repositoryURL:  "docker.io/bitnamicharts",
				repositoryName: "bitnami",
			},
			expectedRepositoryURL:  "oci://docker.io/bitnamicharts",
			expectedRepositoryName: "bitnami",
		},
		{
			name: "case 2",
			args: args{
				repositoryURL:  "docker.io",
				repositoryName: "bitnamicharts/bitnami",
			},
			expectedRepositoryURL:  "oci://docker.io/bitnamicharts",
			expectedRepositoryName: "bitnami",
		},
		{
			name: "case 3",
			args: args{
				repositoryURL:  "oci://docker.io",
				repositoryName: "bitnamicharts/bitnami",
			},
			expectedRepositoryURL:  "oci://docker.io/bitnamicharts",
			expectedRepositoryName: "bitnami",
		},
		{
			name: "case 4",
			args: args{
				repositoryURL:  "https://4.123.13.1/foo/bar",
				repositoryName: "chart",
			},
			expectedRepositoryURL:  "oci://4.123.13.1/foo/bar",
			expectedRepositoryName: "chart",
		},
		{
			name: "case 5",
			args: args{
				repositoryURL:  "http://4.123.13.1/foo/bar",
				repositoryName: "chart",
			},
			expectedRepositoryURL:  "oci://4.123.13.1/foo/bar",
			expectedRepositoryName: "chart",
		},
		{
			name: "case 6",
			args: args{
				repositoryURL:  "https://4.123.13.1",
				repositoryName: "foo/bar/chart",
			},
			expectedRepositoryURL:  "oci://4.123.13.1/foo/bar",
			expectedRepositoryName: "chart",
		},
		{
			name: "case 7",
			args: args{
				repositoryURL:  "http://4.123.13.1",
				repositoryName: "foo/bar/chart",
			},
			expectedRepositoryURL:  "oci://4.123.13.1/foo/bar",
			expectedRepositoryName: "chart",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if repositoryUrl, repositoryName, err := sanitizeRepoNameAndURLForOCIRepo(tt.args.repositoryURL, tt.args.repositoryName); err != nil || repositoryUrl != tt.expectedRepositoryURL || repositoryName != tt.expectedRepositoryName {
				t.Errorf("SanitizeRepoNameAndURLForOCIRepo() = repositoryURL: %v, repositoryName: %v, want  %v %v", repositoryUrl, repositoryName, tt.expectedRepositoryURL, tt.expectedRepositoryName)
			}
		})
	}
}
