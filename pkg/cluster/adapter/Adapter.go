/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package adapter

import (
	"github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
)

func GetClusterBean(model repository.Cluster) bean.ClusterBean {
	clusterBean := bean.ClusterBean{}
	clusterBean.Id = model.Id
	clusterBean.ClusterName = model.ClusterName
	//clusterBean.Note = model.Note
	clusterBean.ServerUrl = model.ServerUrl
	clusterBean.PrometheusUrl = model.PrometheusEndpoint
	clusterBean.AgentInstallationStage = model.AgentInstallationStage
	clusterBean.Active = model.Active
	clusterBean.Config = model.Config
	clusterBean.K8sVersion = model.K8sVersion
	clusterBean.InsecureSkipTLSVerify = model.InsecureSkipTlsVerify
	clusterBean.IsVirtualCluster = model.IsVirtualCluster
	clusterBean.ErrorInConnecting = model.ErrorInConnecting
	clusterBean.IsProd = model.IsProd
	clusterBean.PrometheusAuth = &bean.PrometheusAuth{
		UserName:      model.PUserName,
		Password:      model.PPassword,
		TlsClientCert: model.PTlsClientCert,
		TlsClientKey:  model.PTlsClientKey,
	}
	return clusterBean
}
