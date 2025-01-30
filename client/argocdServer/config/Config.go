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

package config

import (
	k8sUtil "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/read"
	"github.com/devtron-labs/devtron/pkg/util"
	util2 "github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
)

type ArgoCDConfigGetter interface {
	GetGRPCConfig() (*bean.ArgoGRPCConfig, error)
	GetK8sConfig() (*bean.ArgoK8sConfig, error)
}

type ArgoCDConfigGetterImpl struct {
	config              *bean.Config
	devtronSecretConfig *util2.DevtronSecretConfig
	ACDAuthConfig       *util.ACDAuthConfig
	clusterReadService  read.ClusterReadService
	logger              *zap.SugaredLogger
	K8sService          k8sUtil.K8sService
}

func NewArgoCDConfigGetter(
	config *bean.Config,
	environmentVariables *util2.EnvironmentVariables,
	ACDAuthConfig *util.ACDAuthConfig,
	clusterReadService read.ClusterReadService,
	logger *zap.SugaredLogger,
	K8sService k8sUtil.K8sService,
) *ArgoCDConfigGetterImpl {
	return &ArgoCDConfigGetterImpl{
		config:              config,
		devtronSecretConfig: environmentVariables.DevtronSecretConfig,
		ACDAuthConfig:       ACDAuthConfig,
		clusterReadService:  clusterReadService,
		logger:              logger,
		K8sService:          K8sService,
	}
}

func (impl *ArgoCDConfigGetterImpl) GetGRPCConfig() (*bean.ArgoGRPCConfig, error) {
	return &bean.ArgoGRPCConfig{
		ConnectionConfig: impl.config,
		AuthConfig: &bean.AcdAuthConfig{
			ClusterId:                 bean2.DefaultClusterId,
			DevtronSecretName:         impl.devtronSecretConfig.DevtronSecretName,
			DevtronDexSecretNamespace: impl.devtronSecretConfig.DevtronDexSecretNamespace,
		},
	}, nil
}

func (impl *ArgoCDConfigGetterImpl) GetK8sConfig() (*bean.ArgoK8sConfig, error) {
	clusterBean, err := impl.clusterReadService.FindOne(bean2.DEFAULT_CLUSTER)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster bean from db", "err", err)
		return nil, err
	}
	cfg := clusterBean.GetClusterConfig()
	restConfig, err := impl.K8sService.GetRestConfigByCluster(cfg)
	if err != nil {
		impl.logger.Errorw("error in getting k8s config", "err", err)
		return nil, err
	}
	k8sConfig := &bean.ArgoK8sConfig{
		RestConfig:       restConfig,
		AcdNamespace:     impl.ACDAuthConfig.ACDConfigMapNamespace,
		AcdConfigMapName: impl.ACDAuthConfig.ACDConfigMapName,
	}
	return k8sConfig, nil
}
