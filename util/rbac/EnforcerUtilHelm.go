package rbac

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/team"
	"go.uber.org/zap"
	"strings"
)

type EnforcerUtilHelm interface {
	GetHelmObjectByClusterId(clusterId int, namespace string, appName string) string
	GetHelmObjectByTeamIdAndClusterId(teamId int, clusterId int, namespace string, appName string) string
}
type EnforcerUtilHelmImpl struct {
	logger            *zap.SugaredLogger
	clusterRepository repository.ClusterRepository
	teamRepository    team.TeamRepository
}

func NewEnforcerUtilHelmImpl(logger *zap.SugaredLogger,
	clusterRepository repository.ClusterRepository, teamRepository team.TeamRepository,
) *EnforcerUtilHelmImpl {
	return &EnforcerUtilHelmImpl{
		logger:            logger,
		clusterRepository: clusterRepository,
		teamRepository:    teamRepository,
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
