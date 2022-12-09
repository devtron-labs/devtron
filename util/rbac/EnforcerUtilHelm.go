package rbac

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strings"
)

type EnforcerUtilHelm interface {
	GetHelmObjectByClusterId(clusterId int, namespace string, appName string) string
	GetHelmObjectByTeamIdAndClusterId(teamId int, clusterId int, namespace string, appName string) string
	GetHelmObjectForEAMode(appName string, clusterId int, namespace string) string
}
type EnforcerUtilHelmImpl struct {
	logger            *zap.SugaredLogger
	clusterRepository repository.ClusterRepository
	teamRepository    team.TeamRepository
	appRepository     app.AppRepository
}

func NewEnforcerUtilHelmImpl(logger *zap.SugaredLogger,
	clusterRepository repository.ClusterRepository, teamRepository team.TeamRepository, appRepository app.AppRepository,
) *EnforcerUtilHelmImpl {
	return &EnforcerUtilHelmImpl{
		logger:            logger,
		clusterRepository: clusterRepository,
		teamRepository:    teamRepository,
		appRepository:     appRepository,
	}
}

func (impl EnforcerUtilHelmImpl) GetHelmObjectByClusterId(clusterId int, namespace string, appName string) string {
	cluster, err := impl.clusterRepository.FindById(clusterId)
	if err != nil {
		return fmt.Sprintf("%s/%s/%s", "", "", "")
	}
	return fmt.Sprintf("%s/%s__%s/%s", team.UNASSIGNED_PROJECT, cluster.ClusterName, namespace, strings.ToLower(appName))
}

func (impl EnforcerUtilHelmImpl) GetHelmObjectByTeamIdAndClusterId(teamId int, clusterId int, namespace string, appName string) string {

	cluster, err := impl.clusterRepository.FindById(clusterId)

	teamObj, err := impl.teamRepository.FindOne(teamId)

	if err != nil {
		return fmt.Sprintf("%s/%s/%s", "", "", "")
	}
	return fmt.Sprintf("%s/%s__%s/%s", teamObj.Name, cluster.ClusterName, namespace, strings.ToLower(appName))
}

func (impl EnforcerUtilHelmImpl) GetHelmObjectForEAMode(appName string, clusterId int, namespace string) string {

	application, err := impl.appRepository.FindAppAndProjectByAppName(appName)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error on fetching data for rbac object from app repository", "err", err)
		return ""
	}
	cluster, err := impl.clusterRepository.FindById(clusterId)

	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error on fetching data for rbac object from cluster repository", "err", err)
		return ""
	}

	if application.TeamId == 0 {
		return fmt.Sprintf("%s/%s__%s/%s", team.UNASSIGNED_PROJECT, cluster.ClusterName, namespace, strings.ToLower(appName))
	} else {
		return fmt.Sprintf("%s/%s__%s/%s", application.Team, cluster.ClusterName, namespace, strings.ToLower(appName))
	}

}
