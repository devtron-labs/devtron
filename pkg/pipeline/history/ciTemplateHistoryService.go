package history

import (
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
)

type CiTemplateHistoryService interface {
	SaveHistory(material *bean.CiTemplateBean, trigger string) error
}

type CiTemplateHistoryServiceImpl struct {
	CiTemplateHistoryRepository repository.CiTemplateHistoryRepository
	logger                      *zap.SugaredLogger
}

func NewCiTemplateHistoryServiceImpl(CiTemplateHistoryRepository repository.CiTemplateHistoryRepository,
	logger *zap.SugaredLogger) *CiTemplateHistoryServiceImpl {

	return &CiTemplateHistoryServiceImpl{
		CiTemplateHistoryRepository: CiTemplateHistoryRepository,
		logger:                      logger,
	}
}

func (impl CiTemplateHistoryServiceImpl) SaveHistory(ciTemplateBean *bean.CiTemplateBean, trigger string) error {

	ciTemplate := ciTemplateBean.CiTemplate
	ciBuildConfig := ciTemplateBean.CiBuildConfig

	ciTemplateId := 0
	ciTemplateOverrideId := 0

	ciBuildConfigDbEntity, err := bean.ConvertBuildConfigBeanToDbEntity(ciTemplateId, ciTemplateOverrideId, ciBuildConfig, ciTemplateBean.UserId)

	materialHistory := &repository.CiTemplateHistory{
		CiTemplateId:       ciTemplate.Id,
		AppId:              ciTemplate.AppId,
		DockerRegistryId:   ciTemplate.DockerRegistryId,
		DockerRepository:   ciTemplate.DockerRepository,
		DockerfilePath:     ciTemplate.DockerfilePath, //in
		Args:               ciTemplate.Args,
		TargetPlatform:     ciTemplate.TargetPlatform,
		BeforeDockerBuild:  ciTemplate.BeforeDockerBuild,
		AfterDockerBuild:   ciTemplate.AfterDockerBuild,
		TemplateName:       ciTemplate.TemplateName,
		Version:            ciTemplate.Version,
		Active:             ciTemplate.Active,
		GitMaterialId:      ciTemplate.GitMaterialId,
		DockerBuildOptions: ciTemplate.DockerBuildOptions,
		App:                ciTemplate.App,
		DockerRegistry:     ciTemplate.DockerRegistry,
		CiBuildConfigId:    ciBuildConfigDbEntity.Id,
		BuildMetaDataType:  ciBuildConfigDbEntity.Type,
		BuildMetadata:      ciBuildConfigDbEntity.BuildMetadata,
		Trigger:            trigger,
		AuditLog:           sql.AuditLog{CreatedOn: ciTemplate.CreatedOn, CreatedBy: ciTemplate.CreatedBy, UpdatedBy: ciTemplate.UpdatedBy, UpdatedOn: ciTemplate.UpdatedOn},
	}

	err = impl.CiTemplateHistoryRepository.Save(materialHistory)

	if err != nil {
		impl.logger.Errorw("unable to save history for ci template repository", "error", err)
		return err
	}

	return nil

}
