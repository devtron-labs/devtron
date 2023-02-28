package delete

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	"github.com/devtron-labs/devtron/pkg/chartRepo"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/team"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DeleteService interface {
	DeleteCluster(deleteRequest *cluster.ClusterBean, userId int32) error
	DeleteEnvironment(deleteRequest *cluster.EnvironmentBean, userId int32) error
	DeleteTeam(deleteRequest *team.TeamRequest) error
	DeleteChartRepo(deleteRequest *chartRepo.ChartRepoDto) error
}

type DeleteServiceImpl struct {
	logger                 *zap.SugaredLogger
	teamService            team.TeamService
	clusterService         cluster.ClusterService
	environmentService     cluster.EnvironmentService
	chartRepositoryService chartRepo.ChartRepositoryService
	installedAppRepository repository.InstalledAppRepository
	K8sUtil                util.K8sUtil
	aCDAuthConfig          *util2.ACDAuthConfig
}

func NewDeleteServiceImpl(logger *zap.SugaredLogger,
	teamService team.TeamService,
	clusterService cluster.ClusterService,
	environmentService cluster.EnvironmentService, chartRepositoryService chartRepo.ChartRepositoryService,
	installedAppRepository repository.InstalledAppRepository, K8sUtil util.K8sUtil, aCDAuthConfig *util2.ACDAuthConfig) *DeleteServiceImpl {
	return &DeleteServiceImpl{
		logger:                 logger,
		teamService:            teamService,
		clusterService:         clusterService,
		environmentService:     environmentService,
		chartRepositoryService: chartRepositoryService,
		installedAppRepository: installedAppRepository,
		K8sUtil:                K8sUtil,
		aCDAuthConfig:          aCDAuthConfig,
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

func (impl DeleteServiceImpl) DeleteChartSecret(secretName string) error {
	clusterBean, err := impl.clusterService.FindOne(cluster.DefaultClusterName)
	if err != nil {
		return err
	}
	cfg, err := impl.clusterService.GetClusterConfig(clusterBean)
	if err != nil {
		return err
	}
	client, err := impl.K8sUtil.GetClient(cfg)
	if err != nil {
		return err
	}
	err = impl.K8sUtil.DeleteSecret(impl.aCDAuthConfig.ACDConfigMapNamespace, secretName, client)
	return err
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

	err = impl.DeleteChartSecret(deleteRequest.Name)
	if err != nil {
		impl.logger.Errorw("Error in deleting secret for chart repo", "Chart Name", deleteRequest.Name, "err", err)
	}

	return nil
}
