package adapter

import (
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/cluster/repository/bean"
)

// NewEnvironmentBean provides a new cluster.EnvironmentBean for the given repository.Environment
// Note: NewEnvironmentBean doesn't include AppCount and AllowedDeploymentTypes
func NewEnvironmentBean(envModel *repository.Environment) *bean.EnvironmentBean {
	envBean := &bean.EnvironmentBean{
		Id:                    envModel.Id,
		Environment:           envModel.Name,
		ClusterId:             envModel.ClusterId,
		Active:                envModel.Active,
		Default:               envModel.Default,
		Namespace:             envModel.Namespace,
		EnvironmentIdentifier: envModel.EnvironmentIdentifier,
		Description:           envModel.Description,
		IsVirtualEnvironment:  envModel.IsVirtualEnvironment,
	}
	if envModel.Cluster != nil {
		envBean.ClusterName = envModel.Cluster.ClusterName
		envBean.PrometheusEndpoint = envModel.Cluster.PrometheusEndpoint
		envBean.CdArgoSetup = envModel.Cluster.CdArgoSetup
		// populate internal use only fields
		envBean.ClusterServerUrl = envModel.Cluster.ServerUrl
		envBean.ErrorInConnecting = envModel.Cluster.ErrorInConnecting
	}
	return envBean
}
