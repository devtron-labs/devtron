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

package history

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/adapter"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type CiPipelineHistoryService interface {
	SaveHistory(pipeline *pipelineConfig.CiPipeline, CiPipelineMaterial []*pipelineConfig.CiPipelineMaterial, ciTemplateBean *bean.CiTemplateBean, Trigger string) error
}

type CiPipelineHistoryServiceImpl struct {
	CiPipelineHistoryRepository repository.CiPipelineHistoryRepository
	logger                      *zap.SugaredLogger
	ciPipelineRepository        pipelineConfig.CiPipelineRepository
}

func NewCiPipelineHistoryServiceImpl(CiPipelineHistoryRepository repository.CiPipelineHistoryRepository,
	logger *zap.SugaredLogger, ciPipelineRepository pipelineConfig.CiPipelineRepository) *CiPipelineHistoryServiceImpl {
	return &CiPipelineHistoryServiceImpl{
		CiPipelineHistoryRepository: CiPipelineHistoryRepository,
		logger:                      logger,
		ciPipelineRepository:        ciPipelineRepository,
	}
}

func (impl *CiPipelineHistoryServiceImpl) SaveHistory(pipeline *pipelineConfig.CiPipeline, CiPipelineMaterial []*pipelineConfig.CiPipelineMaterial, CiTemplateBean *bean.CiTemplateBean, Trigger string) error {

	CiPipelineMaterialJson, _ := json.Marshal(CiPipelineMaterial)

	var CiPipelineHistory repository.CiPipelineHistory
	var CiTemplateOverride repository.CiPipelineTemplateOverrideHistoryDTO

	IsDockerConfigOverriden := pipeline.IsDockerConfigOverridden

	if IsDockerConfigOverriden {
		ciTemplateId := 0
		ciTemplateOverrideId := 0
		CiTemplateOverride = repository.CiPipelineTemplateOverrideHistoryDTO{
			DockerRegistryId:      CiTemplateBean.CiTemplateOverride.DockerRegistryId,
			DockerRepository:      CiTemplateBean.CiTemplateOverride.DockerRepository,
			DockerfilePath:        CiTemplateBean.CiTemplateOverride.DockerfilePath,
			Active:                CiTemplateBean.CiTemplateOverride.Active,
			AuditLog:              CiTemplateBean.CiTemplateOverride.AuditLog,
			IsCiTemplateOverriden: true,
		}
		if CiTemplateBean.CiBuildConfig != nil {
			CiBuildConfigDbEntity, _ := adapter.ConvertBuildConfigBeanToDbEntity(ciTemplateId, ciTemplateOverrideId, CiTemplateBean.CiBuildConfig, CiTemplateBean.UserId)
			CiTemplateOverride.CiBuildConfigId = CiBuildConfigDbEntity.Id
			CiTemplateOverride.BuildMetaDataType = CiBuildConfigDbEntity.Type
			CiTemplateOverride.BuildMetadata = CiBuildConfigDbEntity.BuildMetadata
		}
	} else {

		CiTemplateOverride = repository.CiPipelineTemplateOverrideHistoryDTO{
			DockerRegistryId:  "",
			DockerRepository:  "",
			DockerfilePath:    "",
			Active:            false,
			CiBuildConfigId:   0,
			BuildMetaDataType: "",
			BuildMetadata:     "",
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: CiTemplateBean.UserId,
				UpdatedOn: time.Now(),
				UpdatedBy: CiTemplateBean.UserId,
			},
			IsCiTemplateOverriden: false,
		}

	}
	CiTemplateOverrideJson, _ := json.Marshal(CiTemplateOverride)

	CiPipelineHistory = repository.CiPipelineHistory{
		CiPipelineId:              pipeline.Id,
		CiTemplateOverrideHistory: string(CiTemplateOverrideJson),
		CiPipelineMaterialHistory: string(CiPipelineMaterialJson),
		Trigger:                   Trigger,
		ScanEnabled:               pipeline.ScanEnabled,
		Manual:                    pipeline.IsManual,
	}

	err := impl.CiPipelineHistoryRepository.Save(&CiPipelineHistory)
	if err != nil {
		impl.logger.Errorw("error in saving history of ci pipeline")
		return err
	}
	ciEnvMapping, err := impl.ciPipelineRepository.FindCiEnvMappingByCiPipelineId(pipeline.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching ciEnvMapping", "ciPipelineId ", pipeline.Id, "err", err)
		return err
	}

	if ciEnvMapping.Id > 0 {
		CiEnvMappingHistory := &repository.CiEnvMappingHistory{
			EnvironmentId: ciEnvMapping.EnvironmentId,
			CiPipelineId:  ciEnvMapping.CiPipelineId,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: ciEnvMapping.CreatedBy,
				UpdatedOn: time.Now(),
				UpdatedBy: ciEnvMapping.UpdatedBy,
			},
		}
		err := impl.CiPipelineHistoryRepository.SaveCiEnvMappingHistory(CiEnvMappingHistory)
		if err != nil {
			impl.logger.Errorw("error in saving history of ci Env Mapping", "err", err)
			return err
		}
	}

	return nil

}
