package history

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
)

type CiTemplateHistoryService interface {
	SaveHistory(material *pipelineConfig.CiTemplate) error
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

func (impl CiTemplateHistoryServiceImpl) SaveHistory(material *pipelineConfig.CiTemplate) error {

	materialHistory := &repository.CiTemplateHistory{
		Id:                 material.Id,
		AppId:              material.AppId,
		DockerRegistryId:   material.DockerRegistryId,
		DockerRepository:   material.DockerRepository,
		DockerfilePath:     material.DockerfilePath,
		Args:               material.Args,
		TargetPlatform:     material.TargetPlatform,
		BeforeDockerBuild:  material.BeforeDockerBuild,
		AfterDockerBuild:   material.AfterDockerBuild,
		TemplateName:       material.TemplateName,
		Version:            material.Version,
		Active:             material.Active,
		GitMaterialId:      material.GitMaterialId,
		DockerBuildOptions: material.DockerBuildOptions,
		AuditLog:           sql.AuditLog{CreatedOn: material.CreatedOn, CreatedBy: material.CreatedBy, UpdatedBy: material.UpdatedBy, UpdatedOn: material.UpdatedOn},
		App:                material.App,
		DockerRegistry:     material.DockerRegistry,
	}

	err := impl.CiTemplateHistoryRepository.Save(materialHistory)

	if err != nil {
		impl.logger.Errorw("unable to save history for ci template repository")
		return err
	}

	return nil

}
