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
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository"
	"go.uber.org/zap"
)

type ImageScanDeployInfoReadService interface {
	FetchByAppIdAndEnvId(appId int, envId int, objectType []string) (*repository.ImageScanDeployInfo, error)
}

type ImageScanDeployInfoReadServiceImpl struct {
	logger                        *zap.SugaredLogger
	imageScanDeployInfoRepository repository.ImageScanDeployInfoRepository
}

func NewImageScanDeployInfoReadService(logger *zap.SugaredLogger,
	imageScanDeployInfoRepository repository.ImageScanDeployInfoRepository) *ImageScanDeployInfoReadServiceImpl {
	return &ImageScanDeployInfoReadServiceImpl{
		logger:                        logger,
		imageScanDeployInfoRepository: imageScanDeployInfoRepository,
	}
}

func (impl *ImageScanDeployInfoReadServiceImpl) FetchByAppIdAndEnvId(appId int, envId int, objectType []string) (*repository.ImageScanDeployInfo, error) {
	return impl.imageScanDeployInfoRepository.FetchByAppIdAndEnvId(appId, envId, objectType)
}
