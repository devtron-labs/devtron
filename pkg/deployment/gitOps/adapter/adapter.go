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
	apiBean "github.com/devtron-labs/devtron/api/bean"
	apiGitOpsBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/sql/repository"
)

func GetGitOpsConfigBean(model *repository.GitOpsConfig) *apiGitOpsBean.GitOpsConfigDto {
	return &apiGitOpsBean.GitOpsConfigDto{
		Id:                    model.Id,
		Provider:              model.Provider,
		GitHubOrgId:           model.GitHubOrgId,
		GitLabGroupId:         model.GitLabGroupId,
		Active:                model.Active,
		Token:                 model.Token.String(),
		Host:                  model.Host,
		Username:              model.Username,
		UserId:                model.CreatedBy,
		AzureProjectName:      model.AzureProject,
		BitBucketWorkspaceId:  model.BitBucketWorkspaceId,
		BitBucketProjectKey:   model.BitBucketProjectKey,
		AllowCustomRepository: model.AllowCustomRepository,
		EnableTLSVerification: model.EnableTLSVerification,
		TLSConfig: &apiBean.TLSConfig{
			CaData:      model.CaCert,
			TLSCertData: model.TlsCert,
			TLSKeyData:  model.TlsKey,
		},
	}
}
