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

type ImageScanResultReadService interface {
	FindByImageDigests(digest []string) ([]*repository.ImageScanExecutionResult, error)
}
type ImageScanResultReadServiceImpl struct {
	logger                    *zap.SugaredLogger
	ImageScanResultRepository repository.ImageScanResultRepository
}

func NewImageScanResultReadServiceImpl(logger *zap.SugaredLogger,
	ImageScanResultRepository repository.ImageScanResultRepository) *ImageScanResultReadServiceImpl {
	return &ImageScanResultReadServiceImpl{
		logger:                    logger,
		ImageScanResultRepository: ImageScanResultRepository,
	}
}

func (impl *ImageScanResultReadServiceImpl) FindByImageDigests(digest []string) ([]*repository.ImageScanExecutionResult, error) {
	return impl.ImageScanResultRepository.FindByImageDigests(digest)
}
