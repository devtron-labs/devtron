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

package pipeline

import (
	"github.com/devtron-labs/devtron/pkg/variables"
	variableEntity "github.com/devtron-labs/devtron/pkg/variables/repository"
	"go.uber.org/zap"
)

// GitMaterialVariableSyncService re-syncs every Git Material whose URL references a scoped
// variable to git-sensor. Called after the variable set changes (CreateVariables does a full
// delete-then-recreate of all variables, so any of them could be the one a material uses).
type GitMaterialVariableSyncService interface {
	ResyncGitMaterialsForVariableChange() error
}

type GitMaterialVariableSyncServiceImpl struct {
	logger                       *zap.SugaredLogger
	variableEntityMappingService variables.VariableEntityMappingService
	ciCdPipelineOrchestrator     CiCdPipelineOrchestrator
}

func NewGitMaterialVariableSyncServiceImpl(logger *zap.SugaredLogger,
	variableEntityMappingService variables.VariableEntityMappingService,
	ciCdPipelineOrchestrator CiCdPipelineOrchestrator) *GitMaterialVariableSyncServiceImpl {
	return &GitMaterialVariableSyncServiceImpl{
		logger:                       logger,
		variableEntityMappingService: variableEntityMappingService,
		ciCdPipelineOrchestrator:     ciCdPipelineOrchestrator,
	}
}

func (impl *GitMaterialVariableSyncServiceImpl) ResyncGitMaterialsForVariableChange() error {
	materialIds, err := impl.variableEntityMappingService.GetDistinctEntityIdsByType(variableEntity.EntityTypeGitMaterial)
	if err != nil {
		impl.logger.Errorw("error in finding git materials referencing scoped variables", "err", err)
		return err
	}
	for _, materialId := range materialIds {
		err = impl.ciCdPipelineOrchestrator.ResyncGitMaterialToSensor(materialId)
		if err != nil {
			// keep re-syncing the rest - one bad/stale mapping shouldn't block the others
			impl.logger.Errorw("error re-syncing git material after variable change", "materialId", materialId, "err", err)
		}
	}
	return nil
}
