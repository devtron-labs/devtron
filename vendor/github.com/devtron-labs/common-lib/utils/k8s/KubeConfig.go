/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package k8s

import (
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"go.uber.org/zap"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
)

type KubeConfigImpl struct {
	logger              *zap.SugaredLogger
	runTimeConfig       *RuntimeConfig
	kubeconfig          *string
	httpTransportConfig HttpTransportInterface
	kubeConfigBuilder   KubeConfigBuilderInterface
}

func NewKubeConfigImpl(
	logger *zap.SugaredLogger,
	runTimeConfig *RuntimeConfig,
	kubeconfig *string,
	httpTransportConfig HttpTransportInterface,
	kubeConfigBuilder KubeConfigBuilderInterface) *KubeConfigImpl {
	return &KubeConfigImpl{
		logger:              logger,
		runTimeConfig:       runTimeConfig,
		kubeconfig:          kubeconfig,
		httpTransportConfig: httpTransportConfig,
		kubeConfigBuilder:   kubeConfigBuilder,
	}
}

type KubeConfigInterface interface {
	GetK8sInClusterRestConfig() (*rest.Config, error)
	GetK8sConfigAndClients(clusterConfig *ClusterConfig) (*rest.Config, *http.Client, *kubernetes.Clientset, error)
	GetK8sInClusterConfigAndDynamicClients() (*rest.Config, *http.Client, dynamic.Interface, error)
	GetK8sInClusterConfigAndClients() (*rest.Config, *http.Client, *kubernetes.Clientset, error)
	GetRestConfigByCluster(clusterConfig *ClusterConfig) (*rest.Config, error)
	OverrideRestConfigWithCustomTransport(restConfig *rest.Config) (*rest.Config, error)
	GetK8sConfigAndClientsByRestConfig(restConfig *rest.Config) (*http.Client, *kubernetes.Clientset, error)
}

func (impl *KubeConfigImpl) GetK8sInClusterRestConfig() (*rest.Config, error) {
	impl.logger.Debug("getting k8s rest config")
	if impl.runTimeConfig.LocalDevMode {
		restConfig, err := clientcmd.BuildConfigFromFlags("", *impl.kubeconfig)
		if err != nil {
			impl.logger.Errorw("Error while building config from flags", "error", err)
			return nil, err
		}
		return impl.httpTransportConfig.OverrideConfigWithCustomTransport(restConfig)
	} else {
		clusterConfig, err := rest.InClusterConfig()
		if err != nil {
			impl.logger.Errorw("error in fetch default cluster config", "err", err)
			return nil, err
		}
		return impl.httpTransportConfig.OverrideConfigWithCustomTransport(clusterConfig)
	}
}

func (impl *KubeConfigImpl) GetK8sConfigAndClients(clusterConfig *ClusterConfig) (*rest.Config, *http.Client, *kubernetes.Clientset, error) {
	restConfig, err := impl.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster", "err", err, "clusterName", clusterConfig.ClusterName)
		return nil, nil, nil, err
	}

	k8sHttpClient, k8sClientSet, err := impl.GetK8sConfigAndClientsByRestConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting client set by rest config", "err", err, "clusterName", clusterConfig.ClusterName)
		return nil, nil, nil, err
	}
	return restConfig, k8sHttpClient, k8sClientSet, nil
}

func (impl *KubeConfigImpl) GetK8sInClusterConfigAndDynamicClients() (*rest.Config, *http.Client, dynamic.Interface, error) {
	restConfig, err := impl.GetK8sInClusterRestConfig()
	if err != nil {
		impl.logger.Errorw("error in getting rest config for in cluster", "err", err)
		return nil, nil, nil, err
	}

	k8sHttpClient, err := OverrideK8sHttpClientWithTracer(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting k8s http client set by rest config for in cluster", "err", err)
		return nil, nil, nil, err
	}
	dynamicClientSet, err := dynamic.NewForConfigAndClient(restConfig, k8sHttpClient)
	if err != nil {
		impl.logger.Errorw("error in getting client set by rest config for in cluster", "err", err)
		return nil, nil, nil, err
	}
	return restConfig, k8sHttpClient, dynamicClientSet, nil
}

func (impl *KubeConfigImpl) GetK8sInClusterConfigAndClients() (*rest.Config, *http.Client, *kubernetes.Clientset, error) {
	restConfig, err := impl.GetK8sInClusterRestConfig()
	if err != nil {
		impl.logger.Errorw("error in getting rest config for in cluster", "err", err)
		return nil, nil, nil, err
	}

	k8sHttpClient, k8sClientSet, err := impl.GetK8sConfigAndClientsByRestConfig(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting client set by rest config for in cluster", "err", err)
		return nil, nil, nil, err
	}
	return restConfig, k8sHttpClient, k8sClientSet, nil
}

func (impl *KubeConfigImpl) GetRestConfigByCluster(clusterConfig *ClusterConfig) (*rest.Config, error) {
	var restConfig *rest.Config
	var err error
	if clusterConfig.Host == commonBean.DefaultClusterUrl && len(clusterConfig.BearerToken) == 0 {
		return impl.GetK8sInClusterRestConfig()
	}
	restConfig, err = impl.kubeConfigBuilder.BuildKubeConfigForCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config for cluster", "err", err, "clusterName", clusterConfig.ClusterName)
		return nil, err
	}
	return impl.OverrideRestConfigWithCustomTransport(restConfig)
}

func (impl *KubeConfigImpl) OverrideRestConfigWithCustomTransport(restConfig *rest.Config) (*rest.Config, error) {
	var err error
	restConfig, err = impl.httpTransportConfig.OverrideConfigWithCustomTransport(restConfig)
	if err != nil {
		impl.logger.Errorw("error in overriding rest config with custom transport configurations", "err", err)
		return nil, err
	}
	return restConfig, nil
}

func (impl *KubeConfigImpl) GetK8sConfigAndClientsByRestConfig(restConfig *rest.Config) (*http.Client, *kubernetes.Clientset, error) {
	k8sHttpClient, err := OverrideK8sHttpClientWithTracer(restConfig)
	if err != nil {
		impl.logger.Errorw("error in getting k8s http client set by rest config", "err", err)
		return nil, nil, err
	}
	k8sClientSet, err := kubernetes.NewForConfigAndClient(restConfig, k8sHttpClient)
	if err != nil {
		impl.logger.Errorw("error in getting client set by rest config", "err", err)
		return nil, nil, err
	}
	return k8sHttpClient, k8sClientSet, nil
}

type KubeConfigBuilder struct{}

type KubeConfigBuilderInterface interface {
	BuildKubeConfigForCluster(clusterConfig *ClusterConfig) (*rest.Config, error)
}

func NewKubeConfigBuilder() *KubeConfigBuilder {
	return &KubeConfigBuilder{}
}

// BuildKubeConfigForCluster builds a kubeconfig for the given cluster configuration.
// This function is used in KubeConfigExtended for extended implementation.
func (impl *KubeConfigBuilder) BuildKubeConfigForCluster(clusterConfig *ClusterConfig) (*rest.Config, error) {
	restConfig := &rest.Config{Host: clusterConfig.Host, BearerToken: clusterConfig.BearerToken}
	clusterConfig.PopulateTlsConfigurationsInto(restConfig)
	return restConfig, nil
}
