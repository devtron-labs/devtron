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

package dockerRegistry

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

type DockerRegistryIpsConfigService interface {
	IsImagePullSecretAccessProvided(dockerRegistryId string, clusterId int) (bool, error)
}

type DockerRegistryIpsConfigServiceImpl struct {
	logger                            *zap.SugaredLogger
	dockerRegistryIpsConfigRepository repository.DockerRegistryIpsConfigRepository
}

func NewDockerRegistryIpsConfigServiceImpl(logger *zap.SugaredLogger, dockerRegistryIpsConfigRepository repository.DockerRegistryIpsConfigRepository) *DockerRegistryIpsConfigServiceImpl {
	return &DockerRegistryIpsConfigServiceImpl{
		logger:                            logger,
		dockerRegistryIpsConfigRepository: dockerRegistryIpsConfigRepository,
	}
}

const ALL_CLUSTER_ID string = "-1"

func (impl DockerRegistryIpsConfigServiceImpl) IsImagePullSecretAccessProvided(dockerRegistryId string, clusterId int) (bool, error) {
	impl.logger.Infow("checking if Ips access provided", "dockerRegistryId", dockerRegistryId, "clusterId", clusterId)

	ipsConfig, err := impl.dockerRegistryIpsConfigRepository.FindByDockerRegistryId(dockerRegistryId)
	if err != nil {
		impl.logger.Errorw("Error while getting docker registry ips config", "dockerRegistryId", dockerRegistryId, "err", err)
		return false, err
	}

	clusterIdStr := strconv.Itoa(clusterId)

	if len(ipsConfig.AppliedClusterIdsCsv) > 0 {
		appliedClusterIds := strings.Split(ipsConfig.AppliedClusterIdsCsv, ",")
		for _, appliedClusterId := range appliedClusterIds {
			if appliedClusterId == ALL_CLUSTER_ID || appliedClusterId == clusterIdStr {
				return true, nil
			}
		}
		return false, nil
	}
	if len(ipsConfig.IgnoredClusterIdsCsv) > 0 {
		ignoredClusterIds := strings.Split(ipsConfig.IgnoredClusterIdsCsv, ",")
		for _, ignoredClusterId := range ignoredClusterIds {
			if ignoredClusterId == ALL_CLUSTER_ID || ignoredClusterId == clusterIdStr {
				return false, nil
			}
		}
		return true, nil
	}

	return false, nil

}
