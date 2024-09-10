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

package adapter

import (
	bean2 "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/bean"
)

func ConvertGitOpsConfigToGitConfig(dto *bean2.GitOpsConfigDto) *bean.GitConfig {
	config := &bean.GitConfig{
		GitlabGroupId:         dto.GitLabGroupId,
		GitToken:              dto.Token,
		GitUserName:           dto.Username,
		GithubOrganization:    dto.GitHubOrgId,
		GitProvider:           dto.Provider,
		GitHost:               dto.Host,
		AzureToken:            dto.Token,
		AzureProject:          dto.AzureProjectName,
		BitbucketWorkspaceId:  dto.BitBucketWorkspaceId,
		BitbucketProjectKey:   dto.BitBucketProjectKey,
		EnableTLSVerification: dto.EnableTLSVerification,
	}
	if dto.TLSConfig != nil {
		config.CaCert = dto.TLSConfig.CaData
		config.TLSCert = dto.TLSConfig.TLSCertData
		config.TLSKey = dto.TLSConfig.TLSKeyData
	}
	return config
}
