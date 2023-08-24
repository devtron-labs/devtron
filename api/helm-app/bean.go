package client

import openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"

const (
	DEFAULT_CLUSTER_ID                     = 1
	SOURCE_DEVTRON_APP       SourceAppType = "devtron-app"
	SOURCE_HELM_APP          SourceAppType = "helm-app"
	SOURCE_EXTERNAL_HELM_APP SourceAppType = "external-helm-app"
	SOURCE_UNKNOWN           SourceAppType = "unknown"
)

type SourceAppType string

type UpdateApplicationRequestDto struct {
	*openapi.UpdateReleaseRequest
	SourceAppType SourceAppType `json:"-"`
}

type UpdateApplicationWithChartInfoRequestDto struct {
	*InstallReleaseRequest
	SourceAppType SourceAppType `json:"-"`
}
