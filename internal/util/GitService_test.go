package util

import (
	"testing"
)

func getTestGithubClient() GitHubClient {
	logger, err := NewSugardLogger()
	gitCliUtl := NewGitCliUtil(logger)
	gitService := NewGitServiceImpl(&GitConfig{GitToken: "", GitUserName: "nishant"}, logger, gitCliUtl)

	githubClient, err := NewGithubClient("", "", "test-org", logger, gitService)
	if err != nil {
		panic(err)
	}
	return githubClient
}

func TestGitHubClient_CreateRepository(t *testing.T) {

	type args struct {
		name                 string
		description          string
		bitbucketWorkspaceId string
		bitbucketProjectKey  string
	}
	tests := []struct {
		name      string
		args      args
		wantIsNew bool
	}{{"test_create", args{
		name:                 "testn3",
		description:          "desc2",
		bitbucketWorkspaceId: "",
		bitbucketProjectKey:  "",
	}, true}} // TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := getTestGithubClient()
			_, gotIsNew, _ := impl.CreateRepository(tt.args.name, tt.args.description, tt.args.bitbucketWorkspaceId, tt.args.bitbucketProjectKey)

			if gotIsNew != tt.wantIsNew {
				t.Errorf("CreateRepository() gotIsNew = %v, want %v", gotIsNew, tt.wantIsNew)
			}

		})
	}
}
