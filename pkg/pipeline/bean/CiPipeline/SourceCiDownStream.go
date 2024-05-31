/*
 * Copyright (c) 2024. Devtron Inc.
 */

package CiPipeline

import "github.com/devtron-labs/devtron/util/response/pagination"

type SourceCiDownStreamFilters struct {
	pagination.QueryParams
	EnvName string `json:"envName"`
}

type SourceCiDownStreamResponse struct {
	AppName          string `json:"appName"`
	AppId            int    `json:"appId"`
	EnvironmentName  string `json:"environmentName"`
	EnvironmentId    int    `json:"environmentId"`
	TriggerMode      string `json:"triggerMode"`
	DeploymentStatus string `json:"deploymentStatus"`
}

type SourceCiDownStreamEnv struct {
	EnvNames []string `json:"envNames"`
}
