package bean

import appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"

type ChartScanEventBean struct {
	AppVersionDto *appStoreBean.InstallAppVersionDTO
	DevtronAppDto *DevtronAppDto
}

type DevtronAppDto struct {
	ChartContent       []byte
	ChartVersion       string
	ChartName          string
	ValuesYaml         string
	CdWorkflowRunnerId int
}
