package read

import (
	"github.com/devtron-labs/devtron/pkg/team/adapter"
	"github.com/devtron-labs/devtron/pkg/team/bean"
	"github.com/devtron-labs/devtron/pkg/team/repository"
	"go.uber.org/zap"
)

type TeamReadService interface {
	FindAllActive() ([]bean.TeamRequest, error)
	FindOne(id int) (*bean.TeamRequest, error)
	FindByTeamName(name string) (*bean.TeamRequest, error)
	FindByIds(ids []*int) ([]*bean.TeamRequest, error)
	FindAllActiveTeamNames() ([]string, error)
}

type TeamReadServiceImpl struct {
	logger         *zap.SugaredLogger
	teamRepository repository.TeamRepository
}

func NewTeamReadService(logger *zap.SugaredLogger,
	teamRepository repository.TeamRepository) *TeamReadServiceImpl {
	return &TeamReadServiceImpl{
		logger:         logger,
		teamRepository: teamRepository,
	}
}

func (impl TeamReadServiceImpl) FindAllActive() ([]bean.TeamRequest, error) {
	impl.logger.Debug("fetch all team from db")
	teams, err := impl.teamRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetch all team", "err", err)
		return nil, err
	}
	var teamRequests []bean.TeamRequest
	for _, teamDBObj := range teams {
		providerRes := adapter.ConvertTeamDBObjToDTO(&teamDBObj)
		teamRequests = append(teamRequests, *providerRes)
	}
	return teamRequests, nil
}

func (impl TeamReadServiceImpl) FindOne(id int) (*bean.TeamRequest, error) {
	impl.logger.Debug("fetch team by ID from db")
	team, err := impl.teamRepository.FindOne(id)
	if err != nil {
		impl.logger.Errorw("error in fetch all team", "err", err)
		return nil, err
	}
	teamDTO := adapter.ConvertTeamDBObjToDTO(&team)
	return teamDTO, nil
}

func (impl TeamReadServiceImpl) FindByTeamName(name string) (*bean.TeamRequest, error) {
	impl.logger.Debug("fetch team by ID from db")
	team, err := impl.teamRepository.FindByTeamName(name)
	if err != nil {
		impl.logger.Errorw("error in fetch all team", "err", err)
		return nil, err
	}
	teamDTO := adapter.ConvertTeamDBObjToDTO(&team)
	return teamDTO, nil
}

func (impl TeamReadServiceImpl) FindByIds(ids []*int) ([]*bean.TeamRequest, error) {
	teams, err := impl.teamRepository.FindByIds(ids)
	if err != nil {
		impl.logger.Errorw("error in fetch all team", "err", err)
		return nil, err
	}
	var teamRequests []*bean.TeamRequest
	for _, team := range teams {
		teamDTO := adapter.ConvertTeamDBObjToDTO(team)
		teamRequests = append(teamRequests, teamDTO)
	}
	return teamRequests, err
}

func (impl TeamReadServiceImpl) FindAllActiveTeamNames() ([]string, error) {
	teamNames, err := impl.teamRepository.FindAllActiveTeamNames()
	if err != nil {
		impl.logger.Errorw("error in fetch all team", "err", err)
		return nil, err
	}
	var allNames []string
	for _, name := range teamNames {
		allNames = append(allNames, name)
	}
	return allNames, err
}
