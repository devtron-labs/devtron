package util

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"testing"
)

func getTestGithubClient() git.GitHubClient {
	logger, err := NewSugardLogger()
	gitCliUtl := git.NewGitCliUtil(logger)
	gitService := git.NewGitServiceImpl(&git.GitConfig{GitToken: "", GitUserName: "nishant"}, logger, gitCliUtl)

	githubClient, err := git.NewGithubClient("", "", "test-org", logger, gitService, nil)
	if err != nil {
		panic(err)
	}
	return githubClient
}

func TestGitHubClient_CreateRepository(t *testing.T) {
	t.SkipNow()
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
			gitOpsConfigDTO := &bean.GitOpsConfigDto{
				Username:             tt.args.name,
				Description:          tt.args.description,
				BitBucketWorkspaceId: tt.args.bitbucketWorkspaceId,
				BitBucketProjectKey:  tt.args.bitbucketProjectKey,
			}
			_, gotIsNew, _ := impl.CreateRepository(gitOpsConfigDTO)

			if gotIsNew != tt.wantIsNew {
				t.Errorf("CreateRepository() gotIsNew = %v, want %v", gotIsNew, tt.wantIsNew)
			}

		})
	}
}
