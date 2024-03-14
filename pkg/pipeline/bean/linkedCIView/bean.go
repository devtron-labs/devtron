package linkedCIView

import "github.com/devtron-labs/devtron/util/response/pagination"

type LinkedCiInfoFilters struct {
	pagination.QueryParams
	EnvName string `json:"envName"`
}

type LinkedCIDetailsRes struct {
	AppName          string `json:"appName"`
	AppId            int    `json:"appId"`
	EnvironmentName  string `json:"environmentName"`
	EnvironmentId    int    `json:"environmentId"`
	TriggerMode      string `json:"triggerMode"`
	DeploymentStatus string `json:"deploymentStatus"`
}
