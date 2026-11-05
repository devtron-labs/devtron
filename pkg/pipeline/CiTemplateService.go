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
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/variables"
	variableEntity "github.com/devtron-labs/devtron/pkg/variables/repository"
	"go.uber.org/zap"
)

type CiTemplateService interface {
	Save(ciTemplateBean *bean.CiTemplateBean) error
	Update(ciTemplateBean *bean.CiTemplateBean) error
	// TrackContainerRepositoryUsage records (or clears) which scoped variable a Container
	// Repository field references. Exported for write paths that update ci_template /
	// ci_template_override directly instead of going through Save/Update - e.g. the CI_JOB
	// override sync in BuildPipelineConfigService.
	TrackContainerRepositoryUsage(entityType variableEntity.EntityType, entityId int, dockerRepository string, userId int32) error
}
type CiTemplateServiceImpl struct {
	Logger                       *zap.SugaredLogger
	CiBuildConfigService         CiBuildConfigService
	CiTemplateRepository         pipelineConfig.CiTemplateRepository
	CiTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository
	variableEntityMappingService variables.VariableEntityMappingService
}

func NewCiTemplateServiceImpl(logger *zap.SugaredLogger, ciBuildConfigService CiBuildConfigService,
	ciTemplateRepository pipelineConfig.CiTemplateRepository, ciTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository,
	variableEntityMappingService variables.VariableEntityMappingService) *CiTemplateServiceImpl {
	return &CiTemplateServiceImpl{
		Logger:                       logger,
		CiBuildConfigService:         ciBuildConfigService,
		CiTemplateRepository:         ciTemplateRepository,
		CiTemplateOverrideRepository: ciTemplateOverrideRepository,
		variableEntityMappingService: variableEntityMappingService,
	}
}

// TrackContainerRepositoryUsage records (or clears) which scoped variable a Container
// Repository field references, so deletion protection can see it as in-use. Container
// Repository supports only a single, full @{{VAR_NAME}} token (see
// RepoFieldVariableResolver), so at most one variable name is ever tracked per entity.
func (impl CiTemplateServiceImpl) TrackContainerRepositoryUsage(entityType variableEntity.EntityType, entityId int, dockerRepository string, userId int32) error {
	var variableNames []string
	if variableName, ok := variables.ExtractVariableName(dockerRepository); ok {
		variableNames = []string{variableName}
	}
	entity := variableEntity.Entity{EntityType: entityType, EntityId: entityId}
	err := impl.variableEntityMappingService.UpdateVariablesForEntity(variableNames, entity, userId, nil)
	if err != nil {
		impl.Logger.Errorw("error in tracking container repository variable mapping", "entityType", entityType, "entityId", entityId, "err", err)
		return err
	}
	return nil
}

func (impl CiTemplateServiceImpl) Save(ciTemplateBean *bean.CiTemplateBean) error {
	ciTemplate := ciTemplateBean.CiTemplate
	ciTemplateOverride := ciTemplateBean.CiTemplateOverride
	ciTemplateId := 0
	ciTemplateOverrideId := 0

	buildConfig := ciTemplateBean.CiBuildConfig
	err := impl.CiBuildConfigService.Save(ciTemplateId, ciTemplateOverrideId, buildConfig, ciTemplateBean.UserId)
	if err != nil {
		impl.Logger.Errorw("error occurred while saving ci build config", "config", buildConfig, "err", err)
	}
	if ciTemplateOverride == nil {
		ciTemplate.CiBuildConfigId = buildConfig.Id
		err := impl.CiTemplateRepository.Save(ciTemplate)
		if err != nil {
			impl.Logger.Errorw("error in saving ci template in db ", "template", ciTemplate, "err", err)
			//TODO delete template from gocd otherwise dangling+ no create in future
			return err
		}
		ciTemplateId = ciTemplate.Id
		err = impl.TrackContainerRepositoryUsage(variableEntity.EntityTypeContainerRepositoryAppLevel, ciTemplate.Id, ciTemplate.DockerRepository, ciTemplateBean.UserId)
		if err != nil {
			return err
		}
	} else {
		ciTemplateOverride.CiBuildConfigId = buildConfig.Id
		_, err := impl.CiTemplateOverrideRepository.Save(ciTemplateOverride)
		if err != nil {
			impl.Logger.Errorw("error in saving template override", "err", err, "templateOverrideConfig", ciTemplateOverride)
			return err
		}
		ciTemplateOverrideId = ciTemplateOverride.Id
		err = impl.TrackContainerRepositoryUsage(variableEntity.EntityTypeContainerRepositoryPipelineLevel, ciTemplateOverride.Id, ciTemplateOverride.DockerRepository, ciTemplateBean.UserId)
		if err != nil {
			return err
		}
	}

	return err
}

func (impl CiTemplateServiceImpl) Update(ciTemplateBean *bean.CiTemplateBean) error {
	ciTemplate := ciTemplateBean.CiTemplate
	ciTemplateOverride := ciTemplateBean.CiTemplateOverride
	ciTemplateId := 0
	ciTemplateOverrideId := 0
	ciBuildConfig := ciTemplateBean.CiBuildConfig
	if ciTemplateOverride == nil {
		ciTemplateId = ciTemplate.Id
	} else {
		ciTemplateOverrideId = ciTemplateOverride.Id
	}
	_, err := impl.CiBuildConfigService.UpdateOrSave(ciTemplateId, ciTemplateOverrideId, ciBuildConfig, ciTemplateBean.UserId)
	if err != nil {
		impl.Logger.Errorw("error in updating ci build config in db", "ciBuildConfig", ciBuildConfig, "err", err)
	}
	if ciTemplateOverride == nil {
		ciTemplate.CiBuildConfigId = ciBuildConfig.Id
		err := impl.CiTemplateRepository.Update(ciTemplate)
		if err != nil {
			impl.Logger.Errorw("error in updating ci template in db", "template", ciTemplate, "err", err)
			return err
		}
		err = impl.TrackContainerRepositoryUsage(variableEntity.EntityTypeContainerRepositoryAppLevel, ciTemplate.Id, ciTemplate.DockerRepository, ciTemplateBean.UserId)
		if err != nil {
			return err
		}
	} else {
		ciTemplateOverride.CiBuildConfigId = ciBuildConfig.Id
		_, err := impl.CiTemplateOverrideRepository.Update(ciTemplateOverride)
		if err != nil {
			impl.Logger.Errorw("error in updating template override", "err", err, "templateOverrideConfig", ciTemplateOverride)
			return err
		}
		err = impl.TrackContainerRepositoryUsage(variableEntity.EntityTypeContainerRepositoryPipelineLevel, ciTemplateOverride.Id, ciTemplateOverride.DockerRepository, ciTemplateBean.UserId)
		if err != nil {
			return err
		}
	}
	return err
}
