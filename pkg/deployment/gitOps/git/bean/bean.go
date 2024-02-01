package bean

type ChartProxyReqDto struct {
	GitOpsRepoName string `json:"gitOpsRepoName"`
	AppName        string `json:"appName,omitempty"`
	UserId         int32  `json:"-"`
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
}
