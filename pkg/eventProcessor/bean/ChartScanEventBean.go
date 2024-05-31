/*
 * Copyright (c) 2024. Devtron Inc.
 */

package bean

import appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"

type ChartScanEventBean struct {
	AppVersionDto *appStoreBean.InstallAppVersionDTO `json:"appVersionDto"`
	DevtronAppDto *DevtronAppDto                     `json:"devtronAppDto"`
}

type DevtronAppDto struct {
	ChartContent []byte `json:"chartContent"`
	ChartVersion string `json:"chartVersion"`
	ChartName    string `json:"chartName"`
	ValuesYaml   string `json:"valuesYaml"`
	CdWorkflowId int    `json:"cdWorkflowRunnerId"`
}
