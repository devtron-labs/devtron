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
	"crypto/tls"
	"github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/bean"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type GitOpsClient interface {
	CreateRepository(ctx context.Context, config *gitOps.GitOpsConfigDto) (url string, isNew bool, isEmpty bool, detailedErrorGitOpsConfigActions DetailedErrorGitOpsConfigActions)
	CommitValues(ctx context.Context, config *ChartConfig, gitOpsConfig *gitOps.GitOpsConfigDto, publishStatusConflictError bool) (commitHash string, commitTime time.Time, err error)
	GetRepoUrl(config *gitOps.GitOpsConfigDto) (repoUrl string, isRepoEmpty bool, err error)
	DeleteRepository(config *gitOps.GitOpsConfigDto) error
	CreateReadme(ctx context.Context, config *gitOps.GitOpsConfigDto) (string, error)
	// CreateFirstCommitOnHead creates a commit on the HEAD of the repository, used for initializing the repository.
	// It is used when the repository is empty and needs an initial commit.
	CreateFirstCommitOnHead(ctx context.Context, config *gitOps.GitOpsConfigDto) (string, error)
}

func GetGitConfigAll(gitOpsConfigReadService config.GitOpsConfigReadService) ([]*bean.GitConfig, error) {
	gitOpsConfigs, err := gitOpsConfigReadService.GetAllGitOpsConfig()
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	} else if err == pg.ErrNoRows {
		return nil, nil
	}
	cfgs := make([]*bean.GitConfig, 0, len(gitOpsConfigs))
	for _, gitOpsConfig := range gitOpsConfigs {
		cfgs = append(cfgs, &bean.GitConfig{
			GitlabGroupId:         gitOpsConfig.GitLabGroupId,
			GitToken:              gitOpsConfig.Token,
			GitUserName:           gitOpsConfig.Username,
			GithubOrganization:    gitOpsConfig.GitHubOrgId,
			GitProvider:           gitOpsConfig.Provider,
			GitHost:               gitOpsConfig.Host,
			AzureToken:            gitOpsConfig.Token,
			AzureProject:          gitOpsConfig.AzureProjectName,
			BitbucketWorkspaceId:  gitOpsConfig.BitBucketWorkspaceId,
			BitbucketProjectKey:   gitOpsConfig.BitBucketProjectKey,
			IsActiveConfig:        gitOpsConfig.Active,
			CaCert:                gitOpsConfig.TLSConfig.CaData,
			TLSCert:               gitOpsConfig.TLSConfig.TLSCertData,
			TLSKey:                gitOpsConfig.TLSConfig.TLSKeyData,
			EnableTLSVerification: gitOpsConfig.EnableTLSVerification,
		})
	}
	return cfgs, nil
}

func NewGitOpsClient(config *bean.GitConfig, logger *zap.SugaredLogger, gitOpsHelper *GitOpsHelper) (GitOpsClient, error) {

	var tlsConfig *tls.Config
	var err error
	if config.EnableTLSVerification {
		tlsConfig, err = util.GetTlsConfig(config.TLSKey, config.TLSCert, config.CaCert, bean.GIT_TLS_DIR)
		if err != nil {
			logger.Errorw("error in getting tls config", "err", err)
			return nil, err
		}
	}

	if config.GitProvider == bean.GITLAB_PROVIDER {
		gitLabClient, err := NewGitLabClient(config, logger, gitOpsHelper, tlsConfig)
		return gitLabClient, err
	} else if config.GitProvider == bean.GITHUB_PROVIDER {
		gitHubClient, err := NewGithubClient(config.GitHost, config.GitToken, config.GithubOrganization, logger, gitOpsHelper, tlsConfig)
		return gitHubClient, err
	} else if config.GitProvider == bean.AZURE_DEVOPS_PROVIDER {
		gitAzureClient, err := NewGitAzureClient(config.AzureToken, config.GitHost, config.AzureProject, logger, gitOpsHelper, tlsConfig)
		return gitAzureClient, err
	} else if config.GitProvider == bean.BITBUCKET_PROVIDER {
		gitBitbucketClient := NewGitBitbucketClient(config.GitUserName, config.GitToken, config.GitHost, logger, gitOpsHelper, tlsConfig)
		return gitBitbucketClient, nil
	} else {
		logger.Warn("no gitops config provided, gitops will not work")
		return &UnimplementedGitOpsClient{}, nil
	}
}
