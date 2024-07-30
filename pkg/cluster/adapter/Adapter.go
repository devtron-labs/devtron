/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
