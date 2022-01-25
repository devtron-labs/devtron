package delete

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appstore"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	appstore2 "github.com/devtron-labs/devtron/pkg/appstore"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DeleteService interface {
	DeleteCluster(deleteRequest *cluster.ClusterBean, userId int32) error
	DeleteEnvironment(deleteRequest *cluster.EnvironmentBean, userId int32) error
	DeleteTeam(deleteRequest *team.TeamRequest) error
	DeleteGitProvider(deleteRequest *pipeline.GitRegistry) error
	DeleteDockerRegistryConfig(deleteRequest *pipeline.DockerArtifactStoreBean) error
	DeleteChartRepo(deleteRequest *appstore2.ChartRepoDto) error
}
type DeleteServiceImpl struct {
	logger                 *zap.SugaredLogger
	teamService            team.TeamService
	appRepository          app.AppRepository
	gitMaterialRepository  pipelineConfig.MaterialRepository
	gitRegistryConfig      pipeline.GitRegistryConfig
	environmentRepository  repository.EnvironmentRepository
	clusterService         cluster.ClusterService
	pipelineRepository     pipelineConfig.PipelineRepository
	environmentService     cluster.EnvironmentService
	ciTemplateRepository   pipelineConfig.CiTemplateRepository
	dockerRegistryConfig   pipeline.DockerRegistryConfig
	installedAppRepository appstore.InstalledAppRepository
	appStoreService        appstore2.AppStoreService
}

func NewDeleteServiceImpl(logger *zap.SugaredLogger,
	teamService team.TeamService, appRepository app.AppRepository,
	gitMaterialRepository pipelineConfig.MaterialRepository,
	gitRegistryConfig pipeline.GitRegistryConfig,
	environmentRepository repository.EnvironmentRepository,
	clusterService cluster.ClusterService,
	pipelineRepository pipelineConfig.PipelineRepository,
	environmentService cluster.EnvironmentService,
	ciTemplateRepository pipelineConfig.CiTemplateRepository,
	dockerRegistryConfig pipeline.DockerRegistryConfig,
	installedAppRepository appstore.InstalledAppRepository,
	appStoreService appstore2.AppStoreService) *DeleteServiceImpl {
	return &DeleteServiceImpl{
		logger:                 logger,
		teamService:            teamService,
		appRepository:          appRepository,
		gitMaterialRepository:  gitMaterialRepository,
		gitRegistryConfig:      gitRegistryConfig,
		environmentRepository:  environmentRepository,
		clusterService:         clusterService,
		pipelineRepository:     pipelineRepository,
		environmentService:     environmentService,
		ciTemplateRepository:   ciTemplateRepository,
		dockerRegistryConfig:   dockerRegistryConfig,
		installedAppRepository: installedAppRepository,
		appStoreService:        appStoreService,
	}
}

func (impl DeleteServiceImpl) DeleteCluster(deleteRequest *cluster.ClusterBean, userId int32) error {
	//finding if there are env in this cluster or not, if yes then will not delete
	env, err := impl.environmentRepository.FindByClusterId(deleteRequest.Id)
	if !(env == nil && err == pg.ErrNoRows) {
		impl.logger.Errorw("err in deleting cluster, found env in this cluster", "clusterName", deleteRequest.ClusterName, "err", err)
		return fmt.Errorf(" Please delete all related environments before deleting this cluster : %w", err)
	}
	err = impl.clusterService.DeleteFromDb(deleteRequest, userId)
	if err != nil {
		impl.logger.Errorw("error im deleting cluster", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	return nil
}

func (impl DeleteServiceImpl) DeleteEnvironment(deleteRequest *cluster.EnvironmentBean, userId int32) error {
	//finding if this env is used in any cd pipelines, if yes then will not delete
	pipelines, err := impl.pipelineRepository.FindActiveByEnvId(deleteRequest.Id)
	if !(pipelines == nil && err == pg.ErrNoRows) {
		impl.logger.Errorw("err in deleting env, found pipelines in this env", "envName", deleteRequest.Environment, "err", err)
		return fmt.Errorf(" Please delete all related pipelines before deleting this environment : %w", err)
	}
	err = impl.environmentService.Delete(deleteRequest, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting environment", "err", err, "deleteRequest", deleteRequest)
	}
	return nil
}
func (impl DeleteServiceImpl) DeleteTeam(deleteRequest *team.TeamRequest) error {
	//finding if this project is used in some app; if yes, will not perform delete operation
	apps, err := impl.appRepository.FindAppsByTeamName(deleteRequest.Name)
	if !(apps == nil && err == pg.ErrNoRows) {
		impl.logger.Errorw("err in deleting team, found apps in team", "teamName", deleteRequest.Name, "err", err)
		return fmt.Errorf(" Please delete all apps in this project before deleting this project : %w", err)
	}
	err = impl.teamService.Delete(deleteRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting team", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	return nil
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
