package history

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
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
			CiBuildConfigDbEntity, _ := bean.ConvertBuildConfigBeanToDbEntity(ciTemplateId, ciTemplateOverrideId, CiTemplateBean.CiBuildConfig, CiTemplateBean.UserId)
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
			Id:            ciEnvMapping.Id,
			EnvironmentId: ciEnvMapping.EnvironmentId,
			CiPipelineId:  ciEnvMapping.CiPipelineId,
		}
		err := impl.CiPipelineHistoryRepository.SaveCiEnvMapping(CiEnvMappingHistory)
		if err != nil {
			impl.logger.Errorw("error in saving history of ci Env Mapping")
			return err
		}
	}

	return nil

}
