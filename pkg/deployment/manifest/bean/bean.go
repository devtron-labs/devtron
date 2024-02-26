package bean

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
)

const (
	FullnameOverride        = "fullnameOverride"
	NameOverride            = "nameOverride"
	KedaAutoscaling         = "kedaAutoscaling"
	HorizontalPodAutoscaler = "HorizontalPodAutoscaler"
	Enabled                 = "enabled"
	ReplicaCount            = "replicaCount"
)

type ConfigMapAndSecretJsonV2 struct {
	AppId                                 int
	EnvId                                 int
	PipeLineId                            int
	ChartVersion                          string
	DeploymentWithConfig                  bean.DeploymentConfigurationType
	WfrIdForDeploymentWithSpecificTrigger int
	Scope                                 resourceQualifiers.Scope
}
