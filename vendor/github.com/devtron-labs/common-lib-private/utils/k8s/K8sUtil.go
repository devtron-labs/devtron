/*
 * Copyright (c) 2020 Devtron Labs
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
 *
 */

package k8s

import (
	"fmt"
	"github.com/devtron-labs/authenticator/client"
	"github.com/devtron-labs/common-lib-private/sshTunnel/bean"
	_ "github.com/devtron-labs/common-lib/utils/k8s"
	k8s2 "github.com/devtron-labs/common-lib/utils/k8s"
	"go.uber.org/zap"
	restclient "k8s.io/client-go/rest"
	"net/http"
	"net/url"
)

type K8sUtilExtended struct {
	k8s2.K8sService
	sshTunnelWrapperService SSHTunnelWrapperService
	logger                  *zap.SugaredLogger
}

func NewK8sUtilExtended(logger *zap.SugaredLogger, runTimeConfig *client.RuntimeConfig,
	sshTunnelWrapperService SSHTunnelWrapperService) *K8sUtilExtended {
	return &K8sUtilExtended{
		K8sService:              k8s2.NewK8sUtil(logger, runTimeConfig),
		sshTunnelWrapperService: sshTunnelWrapperService,
		logger:                  logger,
	}
}

func (impl K8sUtilExtended) GetRestConfigByCluster(clusterConfig *k8s2.ClusterConfig) (*restclient.Config, error) {
	// Call GetRestConfigByCluster for the common configuration
	restConfig, err := impl.K8sService.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		return nil, err
	}

	if clusterConfig.ToConnectWithSSHTunnel {
		hostUrl, err := impl.GetHostUrlForSSHTunnelConfiguredCluster(clusterConfig)
		if err != nil {
			impl.logger.Errorw("error in getting hostUrl for ssh configured cluster", "err", err, "clusterId", clusterConfig.ClusterId)
			return nil, err
		}
		// Override the server URL with the localhost URL where the SSH tunnel is hosted
		restConfig.Host = hostUrl
	} else if len(clusterConfig.ProxyUrl) > 0 {
		proxy, err := url.Parse(clusterConfig.ProxyUrl)
		if err != nil {
			impl.logger.Errorw("error in parsing proxy url", "err", err, "proxyUrl", clusterConfig.ProxyUrl)
			return nil, err
		}
		restConfig.Proxy = http.ProxyURL(proxy)
	}

	return restConfig, nil
}

func (impl K8sUtilExtended) GetHostUrlForSSHTunnelConfiguredCluster(clusterConfig *k8s2.ClusterConfig) (string, error) {
	var sshTunnelUrl string
	//getting port
	port, err := impl.sshTunnelWrapperService.GetPortUsedForACluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting port of ssh tunnel connected cluster", "err", err, "clusterId", clusterConfig.ClusterId)
		return sshTunnelUrl, err
	}
	sshTunnelUrl = fmt.Sprintf("https://%s:%d", bean.LocalHostAddress, port)
	return sshTunnelUrl, nil
}

func (impl K8sUtilExtended) CleanupForClusterUsedForVerification(config *k8s2.ClusterConfig) {
	//cleanup for ssh tunnel, as other methods do not require cleanup
	if config.ToConnectWithSSHTunnel {
		impl.sshTunnelWrapperService.CleanupForVerificationCluster(config.ClusterName)
	}
}
