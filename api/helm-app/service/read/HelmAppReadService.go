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

package read

import (
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/pkg/cluster/read"
	"go.uber.org/zap"
)

type HelmAppReadServiceImpl struct {
	logger             *zap.SugaredLogger
	clusterReadService read.ClusterReadService
}

type HelmAppReadService interface {
	GetClusterConf(clusterId int) (*gRPC.ClusterConfig, error)
}

func NewHelmAppReadServiceImpl(logger *zap.SugaredLogger,
	clusterReadService read.ClusterReadService,
) *HelmAppReadServiceImpl {
	return &HelmAppReadServiceImpl{
		logger:             logger,
		clusterReadService: clusterReadService,
	}
}

func (impl *HelmAppReadServiceImpl) GetClusterConf(clusterId int) (*gRPC.ClusterConfig, error) {
	cluster, err := impl.clusterReadService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "err", err)
		return nil, err
	}
	config := &gRPC.ClusterConfig{
		ApiServerUrl:          cluster.ServerUrl,
		Token:                 cluster.Config[commonBean.BearerToken],
		ClusterId:             int32(cluster.Id),
		ClusterName:           cluster.ClusterName,
		InsecureSkipTLSVerify: cluster.InsecureSkipTLSVerify,
	}
	if cluster.InsecureSkipTLSVerify == false {
		config.KeyData = cluster.Config[commonBean.TlsKey]
		config.CertData = cluster.Config[commonBean.CertData]
		config.CaData = cluster.Config[commonBean.CertificateAuthorityData]
	}
	return config, nil
}
