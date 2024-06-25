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
	"context"
	"github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type GitOpsClient interface {
	CreateRepository(ctx context.Context, config *gitOps.GitOpsConfigDto) (url string, isNew bool, detailedErrorGitOpsConfigActions DetailedErrorGitOpsConfigActions)
	CommitValues(ctx context.Context, config *ChartConfig, gitOpsConfig *gitOps.GitOpsConfigDto) (commitHash string, commitTime time.Time, err error)
	GetRepoUrl(config *gitOps.GitOpsConfigDto) (repoUrl string, err error)
	DeleteRepository(config *gitOps.GitOpsConfigDto) error
	CreateReadme(ctx context.Context, config *gitOps.GitOpsConfigDto) (string, error)
}

func GetGitConfig(gitOpsConfigReadService config.GitOpsConfigReadService) (*bean.GitConfig, error) {
	gitOpsConfig, err := gitOpsConfigReadService.GetGitOpsConfigActive()
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	} else if err == pg.ErrNoRows {
		// adding this block for backward compatibility,TODO: remove in next  iteration
		// cfg := &GitConfig{}
		// err := env.Parse(cfg)
		// return cfg, err
		return &bean.GitConfig{}, nil
	}

	if gitOpsConfig == nil || gitOpsConfig.Id == 0 {
		return nil, err
	}
	cfg := &bean.GitConfig{
		GitlabGroupId:        gitOpsConfig.GitLabGroupId,
		GitToken:             gitOpsConfig.Token,
		GitUserName:          gitOpsConfig.Username,
		GithubOrganization:   gitOpsConfig.GitHubOrgId,
		GitProvider:          gitOpsConfig.Provider,
		GitHost:              gitOpsConfig.Host,
		AzureToken:           gitOpsConfig.Token,
		AzureProject:         gitOpsConfig.AzureProjectName,
		BitbucketWorkspaceId: gitOpsConfig.BitBucketWorkspaceId,
		BitbucketProjectKey:  gitOpsConfig.BitBucketProjectKey,
	}
	return cfg, err
}

func NewGitOpsClient(config *bean.GitConfig, logger *zap.SugaredLogger, gitOpsHelper *GitOpsHelper) (GitOpsClient, error) {
	if config.GitProvider == GITLAB_PROVIDER {
		gitLabClient, err := NewGitLabClient(config, logger, gitOpsHelper)
		return gitLabClient, err
	} else if config.GitProvider == GITHUB_PROVIDER {
		gitHubClient, err := NewGithubClient(config.GitHost, config.GitToken, config.GithubOrganization, logger, gitOpsHelper)
		return gitHubClient, err
	} else if config.GitProvider == AZURE_DEVOPS_PROVIDER {
		gitAzureClient, err := NewGitAzureClient(config.AzureToken, config.GitHost, config.AzureProject, logger, gitOpsHelper)
		return gitAzureClient, err
	} else if config.GitProvider == BITBUCKET_PROVIDER {
		gitBitbucketClient := NewGitBitbucketClient(config.GitUserName, config.GitToken, config.GitHost, logger, gitOpsHelper)
		return gitBitbucketClient, nil
	} else {
		logger.Errorw("no gitops config provided, gitops will not work ")
		return nil, nil
	}
}
