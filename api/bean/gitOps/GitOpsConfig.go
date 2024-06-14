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

package gitOps

import (
	"github.com/devtron-labs/devtron/api/bean"
	"time"
)

type GitOpsConfigDto struct {
	Id                    int            `json:"id,omitempty"`
	Provider              string         `json:"provider" validate:"oneof=GITLAB GITHUB AZURE_DEVOPS BITBUCKET_CLOUD"`
	Username              string         `json:"username"`
	Token                 string         `json:"token"`
	GitLabGroupId         string         `json:"gitLabGroupId"`
	GitHubOrgId           string         `json:"gitHubOrgId"`
	Host                  string         `json:"host"`
	Active                bool           `json:"active"`
	AzureProjectName      string         `json:"azureProjectName"`
	BitBucketWorkspaceId  string         `json:"bitBucketWorkspaceId"`
	BitBucketProjectKey   string         `json:"bitBucketProjectKey"`
	AllowCustomRepository bool           `json:"allowCustomRepository"`
	EnableTLSVerification bool           `json:"enableTLSVerification"`
	TLSConfig             bean.TLSConfig `json:"tlsConfig"`
	// TODO refactoring: create different struct for internal fields
	GitRepoName string `json:"-"`
	UserEmailId string `json:"-"`
	Description string `json:"-"`
	UserId      int32  `json:"-"`
}

type GitRepoRequestDto struct {
	Host                 string `json:"host"`
	Provider             string `json:"provider"`
	GitRepoName          string `json:"gitRepoName"`
	Username             string `json:"username"`
	UserEmailId          string `json:"userEmailId"`
	Token                string `json:"token"`
	GitLabGroupId        string `json:"gitLabGroupId"`
	GitHubOrgId          string `json:"gitHubOrgId"`
	AzureProjectName     string `json:"azureProjectName"`
	BitBucketWorkspaceId string `json:"bitBucketWorkspaceId"`
	BitBucketProjectKey  string `json:"bitBucketProjectKey"`
}

type DetailedErrorGitOpsConfigResponse struct {
	SuccessfulStages  []string          `json:"successfulStages"`
	StageErrorMap     map[string]string `json:"stageErrorMap"`
	ValidatedOn       time.Time         `json:"validatedOn"`
	DeleteRepoFailed  bool              `json:"deleteRepoFailed"`
	ValidationSkipped bool              `json:"validationSkipped"`
}

const (
	GIT_REPO_DEFAULT        = "Default"
	GIT_REPO_NOT_CONFIGURED = "NOT_CONFIGURED" // The value of the constant has been used in the migration script for `custom_gitops_repo_url`; Need to add another migration script if the value is updated.
)

func IsGitOpsRepoNotConfigured(gitRepoUrl string) bool {
	return len(gitRepoUrl) == 0 || gitRepoUrl == GIT_REPO_NOT_CONFIGURED
}
