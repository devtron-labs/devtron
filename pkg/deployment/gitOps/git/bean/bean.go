package bean

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

type PushChartToGitRequestDTO struct {
	AppName           string
	EnvName           string
	ChartAppStoreName string
	RepoURL           string
	TempChartRefDir   string
	UserId            int32
}
