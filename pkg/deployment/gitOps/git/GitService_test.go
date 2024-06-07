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

package git

import (
	"github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/util"
	git "github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/commandManager"
	"testing"
)

func getTestGithubClient() GitHubClient {
	logger, err := util.NewSugardLogger()
	gitService := NewGitOpsHelperImpl(
		&git.BasicAuth{
			Username: "nishant",
			Password: "",
		}, logger)

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
			gitOpsConfigDTO := &gitOps.GitOpsConfigDto{
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
