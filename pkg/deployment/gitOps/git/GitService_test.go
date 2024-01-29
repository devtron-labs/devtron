package git

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/bean"
	"testing"
)

func getTestGithubClient() GitHubClient {
	logger, err := util.NewSugardLogger()
	gitCliUtl := NewGitCliUtil(logger)
	gitService := NewGitServiceImpl(&bean2.GitConfig{GitToken: "", GitUserName: "nishant"}, logger, gitCliUtl)

	githubClient, err := NewGithubClient("", "", "test-org", logger, gitService)
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
