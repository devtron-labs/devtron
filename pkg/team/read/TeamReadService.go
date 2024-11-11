package read

import (
	bean2 "github.com/devtron-labs/devtron/pkg/team/bean"
	"github.com/devtron-labs/devtron/pkg/team/repository"
	"go.uber.org/zap"
)

type TeamReadService interface {
	FetchAllActive() ([]bean2.TeamRequest, error)
}

type TeamReadServiceImpl struct {
	logger         *zap.SugaredLogger
	teamRepository repository.TeamRepository
}

func NewTeamReadServiceImpl(logger *zap.SugaredLogger,
	teamRepository repository.TeamRepository) *TeamReadServiceImpl {
	return &TeamReadServiceImpl{
		logger:         logger,
		teamRepository: teamRepository,
	}
}

func (impl *TeamReadServiceImpl) FetchAllActive() ([]bean2.TeamRequest, error) {
	impl.logger.Debug("fetch all team from db")
	teams, err := impl.teamRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetch all team", "err", err)
		return nil, err
	}
	var teamRequests []bean2.TeamRequest
	for _, team := range teams {
		providerRes := bean2.TeamRequest{
			Id:     team.Id,
			Name:   team.Name,
			Active: team.Active,
			UserId: team.CreatedBy,
		}
		teamRequests = append(teamRequests, providerRes)
	}
	return teamRequests, err
}
