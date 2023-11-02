package delete

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"

	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	"github.com/devtron-labs/devtron/pkg/chartRepo"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DeleteService interface {
	DeleteCluster(deleteRequest *cluster.ClusterBean, userId int32) error
	DeleteEnvironment(deleteRequest *cluster.EnvironmentBean, userId int32) error
	DeleteTeam(deleteRequest *team.TeamRequest) error
	DeleteChartRepo(deleteRequest *chartRepo.ChartRepoDto) error
	DeleteVirtualCluster(bean *cluster.VirtualClusterBean, userId int32) error
	DeleteDockerRegistryConfig(deleteRequest *types.DockerArtifactStoreBean) error
	CanDeleteChartRegistryPullConfig(storeId string) bool
}

type DeleteServiceImpl struct {
	logger                   *zap.SugaredLogger
	teamService              team.TeamService
	clusterService           cluster.ClusterService
	environmentService       cluster.EnvironmentService
	chartRepositoryService   chartRepo.ChartRepositoryService
	installedAppRepository   repository.InstalledAppRepository
	dockerRegistryConfig     pipeline.DockerRegistryConfig
	dockerRegistryRepository dockerRegistryRepository.DockerArtifactStoreRepository
}

func NewDeleteServiceImpl(logger *zap.SugaredLogger,
	teamService team.TeamService,
	clusterService cluster.ClusterService,
	environmentService cluster.EnvironmentService,
	chartRepositoryService chartRepo.ChartRepositoryService,
	installedAppRepository repository.InstalledAppRepository,
	dockerRegistryConfig pipeline.DockerRegistryConfig,
	dockerRegistryRepository dockerRegistryRepository.DockerArtifactStoreRepository,
) *DeleteServiceImpl {
	return &DeleteServiceImpl{
		logger:                   logger,
		teamService:              teamService,
		clusterService:           clusterService,
		environmentService:       environmentService,
		chartRepositoryService:   chartRepositoryService,
		installedAppRepository:   installedAppRepository,
		dockerRegistryConfig:     dockerRegistryConfig,
		dockerRegistryRepository: dockerRegistryRepository,
	}
}

func (impl DeleteServiceImpl) DeleteCluster(deleteRequest *cluster.ClusterBean, userId int32) error {
	err := impl.clusterService.DeleteFromDb(deleteRequest, userId)
	if err != nil {
		impl.logger.Errorw("error im deleting cluster", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	return nil
}

func (impl DeleteServiceImpl) DeleteEnvironment(deleteRequest *cluster.EnvironmentBean, userId int32) error {
	err := impl.environmentService.Delete(deleteRequest, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting environment", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	return nil
}
func (impl DeleteServiceImpl) DeleteTeam(deleteRequest *team.TeamRequest) error {
	err := impl.teamService.Delete(deleteRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting team", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	return nil
}

func (impl DeleteServiceImpl) DeleteChartRepo(deleteRequest *chartRepo.ChartRepoDto) error {

	deployedCharts, err := impl.installedAppRepository.GetAllInstalledAppsByChartRepoId(deleteRequest.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in deleting repo", "deleteRequest", deployedCharts)
		return err
	}
	if len(deployedCharts) > 0 {
		impl.logger.Errorw("err in deleting repo, found charts deployed using this repo", "deleteRequest", deployedCharts)
		return fmt.Errorf("cannot delete repo, found charts deployed in this repo")
	}
	err = impl.chartRepositoryService.DeleteChartRepo(deleteRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting chart repo", "err", err, "deleteRequest", deleteRequest)
		return err
	}

	return nil
}

func (impl DeleteServiceImpl) DeleteVirtualCluster(bean *cluster.VirtualClusterBean, userId int32) error {
	err := impl.clusterService.DeleteVirtualClusterFromDb(bean, userId)
	if err != nil {
		impl.logger.Errorw("error im deleting cluster", "err", err, "deleteRequest", bean)
		return err
	}
	return nil
}

func (impl DeleteServiceImpl) DeleteDockerRegistryConfig(deleteRequest *types.DockerArtifactStoreBean) error {
	store, err := impl.dockerRegistryRepository.FindOneWithDeploymentCount(deleteRequest.Id)
	if err != nil {
		impl.logger.Errorw("error in deleting docker registry", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	if store.DeploymentCount > 0 {
		impl.logger.Errorw("err in deleting docker registry, found chart deployments using registry", "dockerRegistry", deleteRequest.Id, "err", err)
		return fmt.Errorf(" Please update all related docker config before deleting this registry")
	}
	err = impl.dockerRegistryConfig.DeleteReg(deleteRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting docker registry", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	return nil
}

func (impl DeleteServiceImpl) CanDeleteChartRegistryPullConfig(storeId string) bool {
	//finding if docker reg chart is used in any deployment, if yes then will not delete
	store, err := impl.dockerRegistryRepository.FindOneWithDeploymentCount(storeId)
	if err != nil {
		impl.logger.Errorw("error in fetching registry chart deployment docker registry", "dockerRegistry", storeId, "err", err)
		return false
	}
	if store.DeploymentCount > 0 {
		return false
	}
	return true
}
