package bean

type ChartProxyReqDto struct {
	GitOpsRepoName string `json:"gitOpsRepoName"`
	AppName        string `json:"appName,omitempty"`
	UserId         int32  `json:"-"`
}

type PushChartToGitRequestDTO struct {
	AppName           string
	EnvName           string
	ChartAppStoreName string
	RepoURL           string
	TempChartRefDir   string
	UserId            int32
}
