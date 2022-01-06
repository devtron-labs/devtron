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
}
type EnforcerUtilHelmImpl struct {
	logger            *zap.SugaredLogger
	clusterRepository repository.ClusterRepository
}

func NewEnforcerUtilHelmImpl(logger *zap.SugaredLogger,
	clusterRepository repository.ClusterRepository) *EnforcerUtilHelmImpl {
	return &EnforcerUtilHelmImpl{
		logger:            logger,
		clusterRepository: clusterRepository,
	}
}

func (impl EnforcerUtilHelmImpl) GetHelmObjectByClusterId(clusterId int, namespace string, appName string) string {
	cluster, err := impl.clusterRepository.FindById(clusterId)
	if err != nil {
		return fmt.Sprintf("%s/%s/%s", "", "", "")
	}
	return fmt.Sprintf("%s/%s__%s/%s", team.UNASSIGNED_PROJECT, cluster.ClusterName, namespace, strings.ToLower(appName))
}
