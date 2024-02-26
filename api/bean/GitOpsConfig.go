package bean

type GitOpsConfigDto struct {
	Id                   int    `json:"id,omitempty"`
	Provider             string `json:"provider"`
	Username             string `json:"username"`
	Token                string `json:"token"`
	GitLabGroupId        string `json:"gitLabGroupId"`
	GitHubOrgId          string `json:"gitHubOrgId"`
	Host                 string `json:"host"`
	Active               bool   `json:"active"`
	AzureProjectName     string `json:"azureProjectName"`
	BitBucketWorkspaceId string `json:"bitBucketWorkspaceId"`
	BitBucketProjectKey  string `json:"bitBucketProjectKey"`

	// TODO refactoring: create different struct for internal fields
	GitRepoName string `json:"gitRepoName"`
	UserEmailId string `json:"userEmailId"`
	Description string `json:"description"`
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
