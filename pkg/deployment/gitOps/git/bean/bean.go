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
	TempChartRefDir   string
	UserId            int32
}

func (cfg GitConfig) GetAuth() *git.BasicAuth {
	return &git.BasicAuth{
		Username: cfg.GitUserName,
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
