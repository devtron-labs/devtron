package delete

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/appstore"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	appstore2 "github.com/devtron-labs/devtron/pkg/appstore"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DeleteService interface {
	DeleteGitProvider(deleteRequest *pipeline.GitRegistry) error
	DeleteDockerRegistryConfig(deleteRequest *pipeline.DockerArtifactStoreBean) error
	DeleteChartRepo(deleteRequest *appstore2.ChartRepoDto) error
}
type DeleteServiceImpl struct {
	logger                 *zap.SugaredLogger
	gitMaterialRepository  pipelineConfig.MaterialRepository
	gitRegistryConfig      pipeline.GitRegistryConfig
	ciTemplateRepository   pipelineConfig.CiTemplateRepository
	dockerRegistryConfig   pipeline.DockerRegistryConfig
	installedAppRepository appstore.InstalledAppRepository
	appStoreService        appstore2.AppStoreService
}

func NewDeleteServiceImpl(logger *zap.SugaredLogger,
	gitMaterialRepository pipelineConfig.MaterialRepository,
	gitRegistryConfig pipeline.GitRegistryConfig,
	ciTemplateRepository pipelineConfig.CiTemplateRepository,
	dockerRegistryConfig pipeline.DockerRegistryConfig,
	installedAppRepository appstore.InstalledAppRepository,
	appStoreService appstore2.AppStoreService) *DeleteServiceImpl {
	return &DeleteServiceImpl{
		logger:                 logger,
		gitMaterialRepository:  gitMaterialRepository,
		gitRegistryConfig:      gitRegistryConfig,
		ciTemplateRepository:   ciTemplateRepository,
		dockerRegistryConfig:   dockerRegistryConfig,
		installedAppRepository: installedAppRepository,
		appStoreService:        appStoreService,
	}
}

func (impl DeleteServiceImpl) DeleteGitProvider(deleteRequest *pipeline.GitRegistry) error {
	//finding if this git account is used in any git material, if yes then will not delete
	materials, err := impl.gitMaterialRepository.FindByGitProviderId(deleteRequest.Id)
	if !(materials == nil && err == pg.ErrNoRows) {
		impl.logger.Errorw("err in deleting git provider, found git materials using provider", "gitProvider", deleteRequest.Name, "err", err)
		return fmt.Errorf(" Please delete all related git materials before deleting this git account : %w", err)
	}
	err = impl.gitRegistryConfig.Delete(deleteRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting git account", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	return nil
}

func (impl DeleteServiceImpl) DeleteDockerRegistryConfig(deleteRequest *pipeline.DockerArtifactStoreBean) error {
	//finding if docker reg is used in any app, if yes then will not delete
	ciTemplates, err := impl.ciTemplateRepository.FindByDockerRegistryId(deleteRequest.Id)
	if !(ciTemplates == nil && err == pg.ErrNoRows) {
		impl.logger.Errorw("err in deleting docker registry, found docker build config using registry", "dockerRegistry", deleteRequest.Id, "err", err)
		return fmt.Errorf(" Please update all related docker config before deleting this registry : %w", err)
	}
	err = impl.dockerRegistryConfig.DeleteReg(deleteRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting docker registry", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	return nil
}

func (impl DeleteServiceImpl) DeleteChartRepo(deleteRequest *appstore2.ChartRepoDto) error {
	//finding if any charts is deployed using this repo, if yes then will not delete
	deployedCharts, err := impl.installedAppRepository.GetAllInstalledAppsByChartRepoId(deleteRequest.Id)
	if !(deployedCharts == nil && err == pg.ErrNoRows) {
		impl.logger.Errorw("err in deleting repo, found charts deployed using this repo", "deleteRequest", deployedCharts)
		return fmt.Errorf("cannot delete repo, found charts deployed in this repo: %w", err)
	}
	err = impl.appStoreService.DeleteChartRepo(deleteRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting chart repo", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	return nil
}
