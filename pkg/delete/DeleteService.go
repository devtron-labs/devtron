package delete

import (
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/team"
	"go.uber.org/zap"
)

type DeleteService interface {
	DeleteCluster(deleteRequest *cluster.ClusterBean, userId int32) error
	DeleteEnvironment(deleteRequest *cluster.EnvironmentBean, userId int32) error
	DeleteTeam(deleteRequest *team.TeamRequest) error
}

type DeleteServiceImpl struct {
	logger             *zap.SugaredLogger
	teamService        team.TeamService
	clusterService     cluster.ClusterService
	environmentService cluster.EnvironmentService
}

func NewDeleteServiceImpl(logger *zap.SugaredLogger,
	teamService team.TeamService,
	clusterService cluster.ClusterService,
	environmentService cluster.EnvironmentService,
) *DeleteServiceImpl {
	return &DeleteServiceImpl{
		logger:             logger,
		teamService:        teamService,
		clusterService:     clusterService,
		environmentService: environmentService,
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
