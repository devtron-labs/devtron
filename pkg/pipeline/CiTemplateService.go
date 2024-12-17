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
	"go.uber.org/zap"
)

type CiTemplateService interface {
	Save(ciTemplateBean *bean.CiTemplateBean) error
	Update(ciTemplateBean *bean.CiTemplateBean) error
}
type CiTemplateServiceImpl struct {
	Logger                       *zap.SugaredLogger
	CiBuildConfigService         CiBuildConfigService
	CiTemplateRepository         pipelineConfig.CiTemplateRepository
	CiTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository
}

func NewCiTemplateServiceImpl(logger *zap.SugaredLogger, ciBuildConfigService CiBuildConfigService,
	ciTemplateRepository pipelineConfig.CiTemplateRepository, ciTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository) *CiTemplateServiceImpl {
	return &CiTemplateServiceImpl{
		Logger:                       logger,
		CiBuildConfigService:         ciBuildConfigService,
		CiTemplateRepository:         ciTemplateRepository,
		CiTemplateOverrideRepository: ciTemplateOverrideRepository,
	}
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
	} else {
		ciTemplateOverride.CiBuildConfigId = buildConfig.Id
		_, err := impl.CiTemplateOverrideRepository.Save(ciTemplateOverride)
		if err != nil {
			impl.Logger.Errorw("error in saving template override", "err", err, "templateOverrideConfig", ciTemplateOverride)
			return err
		}
		ciTemplateOverrideId = ciTemplateOverride.Id
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
	} else {
		ciTemplateOverride.CiBuildConfigId = ciBuildConfig.Id
		_, err := impl.CiTemplateOverrideRepository.Update(ciTemplateOverride)
		if err != nil {
			impl.Logger.Errorw("error in updating template override", "err", err, "templateOverrideConfig", ciTemplateOverride)
			return err
		}
	}
	return err
}
