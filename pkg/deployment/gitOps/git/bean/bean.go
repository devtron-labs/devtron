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

package bean

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/constants"
	git "github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/commandManager"
)

type GitConfig struct {
	GitlabGroupId        string //local
	GitlabGroupPath      string //local
	GitToken             string //not null  // public
	GitUserName          string //not null  // public
	GithubOrganization   string
	GitProvider          string // SUPPORTED VALUES  GITHUB, GITLAB
	GitHost              string
	AzureToken           string
	AzureProject         string
	BitbucketWorkspaceId string
	BitbucketProjectKey  string

	IsActiveConfig bool //flag to check if the gitOps config is active
	AuthMode       constants.AuthMode

	EnableTLSVerification bool
	CaCert                string
	TLSCert               string
	TLSKey                string
}

type PushChartToGitRequestDTO struct {
	AppName           string
	EnvName           string
	ChartAppStoreName string
	RepoURL           string
	TargetRevision    string
	TempChartRefDir   string
	UserId            int32
}

func bitBucketGitOpsHelperClient(cfg GitConfig) *git.BasicAuth {
	username := cfg.GitUserName

	// Bitbucket Cloud tokens use a fixed git-over-HTTPS username, not the account name:
	//   - access token  -> `x-token-auth:<token>`              (REST API uses Bearer)
	//   - API token      -> `x-bitbucket-api-token-auth:<token>` (REST API uses Basic email:token)
	switch cfg.AuthMode {
	case constants.AUTH_MODE_ACCESS_TOKEN:
		username = BITBUCKET_ACCESS_TOKEN_USERNAME
	case constants.AUTH_MODE_API_TOKEN:
		username = BITBUCKET_API_TOKEN_USERNAME
	}

	return &git.BasicAuth{
		Username: username,
		Password: cfg.GitToken,
		AuthMode: cfg.AuthMode,
	}
}

func (cfg GitConfig) GetAuth() *git.BasicAuth {
	username := cfg.GitUserName

	if cfg.GitProvider == BITBUCKET_PROVIDER {
		return bitBucketGitOpsHelperClient(cfg)
	}

	return &git.BasicAuth{
		Username: username,
		Password: cfg.GitToken,
		AuthMode: cfg.AuthMode,
	}

}

func (cfg GitConfig) GetTLSConfig() *bean.TLSConfig {
	return &bean.TLSConfig{
		CaData:      cfg.CaCert,
		TLSCertData: cfg.TLSCert,
		TLSKeyData:  cfg.TLSKey,
	}
}
