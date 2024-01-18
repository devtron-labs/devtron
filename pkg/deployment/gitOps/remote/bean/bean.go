package bean

type ChartProxyReqDto struct {
	GitOpsRepoName string `json:"gitOpsRepoName"`
	AppName        string `json:"appName,omitempty"`
	UserId         int32  `json:"-"`
}
