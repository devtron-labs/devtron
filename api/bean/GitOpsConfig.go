package bean

type GitOpsConfigDto struct {
	Id               int    `json:"id,omitempty"`
	Provider         string `json:"provider"`
	Username         string `json:"username"`
	Token            string `json:"token"`
	GitLabGroupId    string `json:"gitLabGroupId"`
	GitHubOrgId      string `json:"gitHubOrgId"`
	Host             string `json:"host"`
	Active           bool   `json:"active"`
	AzureProjectName string `json:"azureProjectName"`
	UserId           int32  `json:"-"`
}
