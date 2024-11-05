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
	bean2 "github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
)

func GetClusterBean(model repository.Cluster) bean2.ClusterBean {
	bean := bean2.ClusterBean{}
	bean.Id = model.Id
	bean.ClusterName = model.ClusterName
	//bean.Note = model.Note
	bean.ServerUrl = model.ServerUrl
	bean.PrometheusUrl = model.PrometheusEndpoint
	bean.AgentInstallationStage = model.AgentInstallationStage
	bean.Active = model.Active
	bean.Config = model.Config
	bean.K8sVersion = model.K8sVersion
	bean.InsecureSkipTLSVerify = model.InsecureSkipTlsVerify
	bean.IsVirtualCluster = model.IsVirtualCluster
	bean.ErrorInConnecting = model.ErrorInConnecting
	bean.PrometheusAuth = &bean2.PrometheusAuth{
		UserName:      model.PUserName,
		Password:      model.PPassword,
		TlsClientCert: model.PTlsClientCert,
		TlsClientKey:  model.PTlsClientKey,
	}
	return bean
}
