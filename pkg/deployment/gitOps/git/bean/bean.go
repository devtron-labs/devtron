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
	"strings"

	"github.com/devtron-labs/devtron/api/bean"
	apiGitOpsBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/sql/constants"
	git "github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/commandManager"
)

// ResolveBitbucketCloudAuthMode infers the Bitbucket Cloud token-auth flow from the username the
// user supplied, so the existing UI (which only collects username + token for Bitbucket Cloud) can
// select app password / access token / API token without any frontend change. An explicit auth mode
// (e.g. set directly via the API) always wins.
//
//	username == "x-token-auth"  -> access token (REST: Bearer, git: x-token-auth)
//	username is an email (has @) -> API token   (REST: email:token, git: x-bitbucket-api-token-auth)
//	otherwise                    -> app password (unchanged)
func ResolveBitbucketCloudAuthMode(dto *apiGitOpsBean.GitOpsConfigDto) {
	if dto == nil || strings.ToUpper(dto.Provider) != BITBUCKET_PROVIDER {
		return
	}
	dto.AuthMode = bitbucketCloudAuthMode(dto.Username, dto.AuthMode)
}

// bitbucketCloudAuthMode derives the effective Bitbucket Cloud auth mode. An explicit token mode
// (set via the API, or stored from a prior save) always wins; otherwise it is inferred from the
// username. Keeping the inference here means GitOpsRepoGitUsername does NOT depend on
// ResolveBitbucketCloudAuthMode having run upstream — a legacy config stored as PASSWORD with an
// email/x-token-auth username still resolves correctly on read paths (e.g. ArgoCD/FluxCD).
func bitbucketCloudAuthMode(username string, current apiGitOpsBean.AuthMode) apiGitOpsBean.AuthMode {
	if !current.IsPassword() {
		return current
	}
	switch {
	case username == BITBUCKET_ACCESS_TOKEN_USERNAME:
		return apiGitOpsBean.ACCESS_TOKEN
	case strings.Contains(username, "@"):
		return apiGitOpsBean.API_TOKEN
	}
	return current
}

// GitOpsRepoGitUsername returns the username to use for git-over-HTTPS auth against the GitOps repo.
// This is the single source of truth for git credentials consumed by Devtron's own git operations,
// the ArgoCD credential template + secret, and the FluxCD git secret. Bitbucket Cloud token modes
// use a fixed git username that differs from the stored/REST username (the account email):
//
//	access token -> x-token-auth   |   API token -> x-bitbucket-api-token-auth
func GitOpsRepoGitUsername(dto *apiGitOpsBean.GitOpsConfigDto) string {
	if dto == nil {
		return ""
	}
	if strings.ToUpper(dto.Provider) == BITBUCKET_PROVIDER {
		switch bitbucketCloudAuthMode(dto.Username, dto.AuthMode) {
		case apiGitOpsBean.ACCESS_TOKEN:
			return BITBUCKET_ACCESS_TOKEN_USERNAME
		case apiGitOpsBean.API_TOKEN:
			return BITBUCKET_API_TOKEN_USERNAME
		}
	}
	return dto.Username
}

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

func (cfg GitConfig) GetAuth() *git.BasicAuth {
	username := cfg.GitUserName
	// Bitbucket Cloud tokens use a fixed git-over-HTTPS username, not the account name:
	//   - access token  -> `x-token-auth:<token>`              (REST API uses Bearer)
	//   - API token      -> `x-bitbucket-api-token-auth:<token>` (REST API uses Basic email:token)
	if cfg.GitProvider == BITBUCKET_PROVIDER {
		switch cfg.AuthMode {
		case constants.AUTH_MODE_ACCESS_TOKEN:
			username = BITBUCKET_ACCESS_TOKEN_USERNAME
		case constants.AUTH_MODE_API_TOKEN:
			username = BITBUCKET_API_TOKEN_USERNAME
		}
	}
	return &git.BasicAuth{
		Username: username,
		Password: cfg.GitToken,
	}
}

func (cfg GitConfig) GetTLSConfig() *bean.TLSConfig {
	return &bean.TLSConfig{
		CaData:      cfg.CaCert,
		TLSCertData: cfg.TLSCert,
		TLSKeyData:  cfg.TLSKey,
	}
}
