package cluster

import (
	"errors"
	"strings"

	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"go.uber.org/zap"
)

type ClusterRbacService interface {
	CheckAuthorization(clusterName string, clusterId int, token string, userId int32, rbacForClusterMappingsAlso bool) (bool, error)
}

type ClusterRbacServiceImpl struct {
	logger             *zap.SugaredLogger
	environmentService EnvironmentService
	enforcer           casbin.Enforcer
	clusterService     ClusterService
	userService        user.UserService
}

func NewClusterRbacServiceImpl(environmentService EnvironmentService,
	enforcer casbin.Enforcer,
	clusterService ClusterService,
	logger *zap.SugaredLogger,
	userService user.UserService) *ClusterRbacServiceImpl {
	clusterRbacService := &ClusterRbacServiceImpl{
		logger:             logger,
		environmentService: environmentService,
		enforcer:           enforcer,
		clusterService:     clusterService,
		userService:        userService,
	}

	return clusterRbacService
}

func (impl *ClusterRbacServiceImpl) CheckAuthorization(clusterName string, clusterId int, token string, userId int32, rbacForClusterMappingsAlso bool) (authenticated bool, err error) {
	if rbacForClusterMappingsAlso {
		allowedClusterMap, err := impl.FetchAllowedClusterMap(userId, token)

		if err != nil {
			impl.logger.Errorw("error in fetching allowedClusterMap ", "err", err, "clusterName", clusterName)
			return false, err
		}
		if allowedClusterMap[clusterName] {
			return true, nil
		}
	}
	//getting all environments for this cluster
	envs, err := impl.environmentService.GetByClusterId(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting environments by clusterId", "err", err, "clusterId", clusterId)
		return false, err
	}
	if len(envs) == 0 {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			return false, nil
		}
		return true, nil
	}

	var envIdentifierList []string
	envIdentifierMap := make(map[string]bool)
	for _, env := range envs {
		envIdentifier := strings.ToLower(env.EnvironmentIdentifier)
		envIdentifierList = append(envIdentifierList, envIdentifier)
		envIdentifierMap[envIdentifier] = true
	}
	if len(envIdentifierList) == 0 {
		return false, errors.New("environment identifier list for rbac batch enforcing contains zero environments")
	}
	// RBAC enforcer applying
	rbacResultMap := impl.enforcer.EnforceInBatch(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, envIdentifierList)
	for envIdentifier, _ := range envIdentifierMap {
		if rbacResultMap[envIdentifier] {
			//if user has view permission to even one environment of this cluster, authorise the request
			return true, nil
		}
	}

	return false, nil
}
func (impl *ClusterRbacServiceImpl) FetchAllowedClusterMap(userId int32, token string) (map[string]bool, error) {
	allowedClustersMap := make(map[string]bool)
	roles, err := impl.clusterService.FetchRolesFromGroup(userId, token)
	if err != nil {
		impl.logger.Errorw("error while fetching user roles from db", "error", err)
		return nil, err
	}
	for _, role := range roles {
		allowedClustersMap[role.Cluster] = true
	}
	return allowedClustersMap, err

}
