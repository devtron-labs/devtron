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
	"time"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/constants"
	"go.uber.org/zap/zapcore"
)

type GitOpsConfigDto struct {
	Id                    int             `json:"id,omitempty"`
	Provider              string          `json:"provider" validate:"oneof=GITLAB GITHUB AZURE_DEVOPS BITBUCKET_CLOUD"`
	Username              string          `json:"username"`
	Token                 string          `json:"token"`
	GitLabGroupId         string          `json:"gitLabGroupId"`
	GitHubOrgId           string          `json:"gitHubOrgId"`
	Host                  string          `json:"host"`
	Active                bool            `json:"active"`
	AzureProjectName      string          `json:"azureProjectName"`
	BitBucketWorkspaceId  string          `json:"bitBucketWorkspaceId"`
	BitBucketProjectKey   string          `json:"bitBucketProjectKey"`
	AllowCustomRepository bool            `json:"allowCustomRepository"`
	EnableTLSVerification bool            `json:"enableTLSVerification"`
	TLSConfig             *bean.TLSConfig `json:"tlsConfig"`
	AuthMode              AuthMode        `json:"authMode"`

	IsCADataPresent      bool `json:"isCADataPresent"`
	IsTLSCertDataPresent bool `json:"isTLSCertDataPresent"`
	IsTLSKeyDataPresent  bool `json:"isTLSKeyDataPresent"`

	// TODO refactoring: create different struct for internal fields
	GitRepoName    string `json:"-"`
	TargetRevision string `json:"-"`
	UserEmailId    string `json:"-"`
	Description    string `json:"-"`
	UserId         int32  `json:"-"`
}

// MarshalLogObject decides what zap writes when the DTO is logged as a field,
// e.g. logger.Infow("...", "payload", dto). Without it zap writes every JSON
// field, including credentials such as the token and TLS data.
//
// Only safe fields are listed, so a field added later stays out of logs by default.
// Host is left out because a URL can contain credentials (https://user:token@host).
// The value receiver covers both dto and &dto. A nil *GitOpsConfigDto panics in zap,
// so never log a pointer that can be nil. API JSON is not affected.
func (dto GitOpsConfigDto) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("id", dto.Id)
	enc.AddString("provider", dto.Provider)
	enc.AddBool("active", dto.Active)
	enc.AddString("authMode", string(dto.AuthMode))
	enc.AddBool("allowCustomRepository", dto.AllowCustomRepository)
	enc.AddBool("enableTLSVerification", dto.EnableTLSVerification)
	return nil
}

type AuthMode string

const (
	PASSWORD     AuthMode = "PASSWORD"
	ACCESS_TOKEN AuthMode = "ACCESS_TOKEN"
	API_TOKEN    AuthMode = "API_TOKEN"
)

func (auth AuthMode) IsPassword() bool {
	return auth == PASSWORD || auth == "" //empty for backward compatibility
}

func (auth AuthMode) IsAccessToken() bool {
	return auth == ACCESS_TOKEN
}

func (auth AuthMode) IsApiToken() bool {
	return auth == API_TOKEN
}

func (auth AuthMode) ToInternalAuthMode() constants.AuthMode {

	switch auth {
	case PASSWORD:
		return constants.AUTH_MODE_USERNAME_PASSWORD
	case ACCESS_TOKEN:
		return constants.AUTH_MODE_ACCESS_TOKEN
	case API_TOKEN:
		return constants.AUTH_MODE_API_TOKEN
	}
	return ""
}

func (dto GitOpsConfigDto) GetHostUrl() string {
	return dto.Host
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

func IsGitOpsRepoConfigured(gitRepoUrl string) bool {
	return !IsGitOpsRepoNotConfigured(gitRepoUrl)
}
