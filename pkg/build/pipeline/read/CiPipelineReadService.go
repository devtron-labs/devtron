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

package read

import (
	"errors"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CiPipelineConfigReadService interface {
	FindLinkedCiCount(ciPipelineId int) (int, error)
	FindNumberOfAppsWithCiPipeline(appIds []int) (count int, err error)
	FindAllPipelineCreatedCountInLast24Hour() (pipelineCount int, err error)
	FindAllDeletedPipelineCountInLast24Hour() (pipelineCount int, err error)
	GetChildrenCiCount(parentCiPipelineId int) (int, error)
}

type CiPipelineConfigReadServiceImpl struct {
	logger               *zap.SugaredLogger
	ciPipelineRepository pipelineConfig.CiPipelineRepository
}

func NewCiPipelineConfigReadServiceImpl(
	logger *zap.SugaredLogger,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
) *CiPipelineConfigReadServiceImpl {
	return &CiPipelineConfigReadServiceImpl{
		logger:               logger,
		ciPipelineRepository: ciPipelineRepository,
	}
}

func (impl *CiPipelineConfigReadServiceImpl) FindLinkedCiCount(ciPipelineId int) (int, error) {
	return impl.ciPipelineRepository.FindLinkedCiCount(ciPipelineId)
}

func (impl *CiPipelineConfigReadServiceImpl) FindNumberOfAppsWithCiPipeline(appIds []int) (count int, err error) {
	return impl.ciPipelineRepository.FindNumberOfAppsWithCiPipeline(appIds)
}

func (impl *CiPipelineConfigReadServiceImpl) FindAllPipelineCreatedCountInLast24Hour() (pipelineCount int, err error) {
	return impl.ciPipelineRepository.FindAllPipelineCreatedCountInLast24Hour()
}

func (impl *CiPipelineConfigReadServiceImpl) FindAllDeletedPipelineCountInLast24Hour() (pipelineCount int, err error) {
	return impl.ciPipelineRepository.FindAllDeletedPipelineCountInLast24Hour()
}

func (impl *CiPipelineConfigReadServiceImpl) GetChildrenCiCount(parentCiPipelineId int) (int, error) {
	count, err := impl.ciPipelineRepository.GetChildrenCiCount(parentCiPipelineId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("failed to get children ci count", "parentCiPipelineId", parentCiPipelineId, "error", err)
		return 0, err
	} else if errors.Is(err, pg.ErrNoRows) {
		impl.logger.Debugw("no children ci found", "parentCiPipelineId", parentCiPipelineId)
		return 0, nil
	}
	return count, nil
}
