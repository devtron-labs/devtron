package bean

type DeployedAppMetricsRequest struct {
	EnableMetrics bool
	AppId         int
	EnvId         int // if not zero then request for override
	ChartRefId    int
	UserId        int32
}
