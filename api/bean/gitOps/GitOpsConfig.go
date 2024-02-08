package gitOps

import "time"

type GitOpsConfigDto struct {
	Id                    int    `json:"id,omitempty"`
	Provider              string `json:"provider"`
	Username              string `json:"username"`
	Token                 string `json:"token"`
	GitLabGroupId         string `json:"gitLabGroupId"`
	GitHubOrgId           string `json:"gitHubOrgId"`
	Host                  string `json:"host"`
	Active                bool   `json:"active"`
	AzureProjectName      string `json:"azureProjectName"`
	BitBucketWorkspaceId  string `json:"bitBucketWorkspaceId"`
	BitBucketProjectKey   string `json:"bitBucketProjectKey"`
	AllowCustomRepository bool   `json:"allowCustomRepository"`

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
