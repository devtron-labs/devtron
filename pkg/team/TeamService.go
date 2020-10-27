/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package team

import (
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/team"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/user"
	"go.uber.org/zap"
	"time"
)

type TeamService interface {
	Create(request *TeamRequest) (*TeamRequest, error)
	FetchAll() ([]TeamRequest, error)
	FetchOne(id int) (*TeamRequest, error)
	FindTeamsByUser(userId int32) ([]team.Team, error)
	Update(request *TeamRequest) (*TeamRequest, error)
	FindTeamByAppId(appId int) (*TeamBean, error)
	FindTeamByAppName(appName string) (*TeamBean, error)
	FetchForAutocomplete() ([]TeamRequest, error)
	FindByIds(ids []*int) ([]*TeamBean, error)
}
type TeamServiceImpl struct {
	logger          *zap.SugaredLogger
	userService     user.UserService
	teamRepository  team.TeamRepository
	pipelineBuilder pipeline.PipelineBuilder
	envService      cluster.EnvironmentService
}

type TeamRequest struct {
	Id     int    `json:"id,omitempty" validate:"number"`
	Name   string `json:"name,omitempty" validate:"required"`
	Active bool   `json:"active"`
	UserId int32  `json:"-"`
}

func NewTeamServiceImpl(logger *zap.SugaredLogger, teamRepository team.TeamRepository,
	pipelineBuilder pipeline.PipelineBuilder, envService cluster.EnvironmentService, userService user.UserService) *TeamServiceImpl {
	return &TeamServiceImpl{
		logger:          logger,
		userService:     userService,
		teamRepository:  teamRepository,
		pipelineBuilder: pipelineBuilder,
		envService:      envService,
	}
}

func (impl TeamServiceImpl) Create(request *TeamRequest) (*TeamRequest, error) {
	impl.logger.Debugw("team create request", "req", request)
	t := &team.Team{
		Name:     request.Name,
		Id:       request.Id,
		Active:   request.Active,
		AuditLog: models.AuditLog{CreatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: request.UserId},
	}
	err := impl.teamRepository.Save(t)
	if err != nil {
		impl.logger.Errorw("error in saving team", "data", t, "err", err)
		err = &util.ApiError{
			InternalMessage: "team failed to create in db",
			UserMessage:     "team failed to create in db",
		}
		return nil, err
	}
	request.Id = t.Id
	return request, nil
}

func (impl TeamServiceImpl) FetchAll() ([]TeamRequest, error) {
	impl.logger.Debug("fetch all team from db")
	teams, err := impl.teamRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("error in fetch all team", "err", err)
		return nil, err
	}
	var teamRequests []TeamRequest
	for _, team := range teams {
		providerRes := TeamRequest{
			Id:     team.Id,
			Name:   team.Name,
			Active: team.Active,
			UserId: team.CreatedBy,
		}
		teamRequests = append(teamRequests, providerRes)
	}
	return teamRequests, err
}

func (impl TeamServiceImpl) FetchOne(teamId int) (*TeamRequest, error) {
	impl.logger.Debug("fetch team by ID from db")
	team, err := impl.teamRepository.FindOne(teamId)
	if err != nil {
		impl.logger.Errorw("error in fetch all team", "err", err)
		return nil, err
	}

	teamRes := &TeamRequest{
		Id:     team.Id,
		Name:   team.Name,
		Active: team.Active,
		UserId: team.CreatedBy,
	}

	return teamRes, err
}

func (impl TeamServiceImpl) Update(request *TeamRequest) (*TeamRequest, error) {
	impl.logger.Debugw("team update request", "req", request)
	existingProvider, err0 := impl.teamRepository.FindOne(request.Id)
	if err0 != nil {
		impl.logger.Errorw("No matching entry found for update.", "id", request.Id)
		err0 = &util.ApiError{
			Code:            constants.GitProviderUpdateProviderNotExists,
			InternalMessage: "team update failed, does not exist",
			UserMessage:     "team update failed, does not exist",
		}
		return nil, err0
	}
	team := &team.Team{
		Name:     request.Name,
		Id:       request.Id,
		Active:   request.Active,
		AuditLog: models.AuditLog{CreatedBy: existingProvider.CreatedBy, CreatedOn: existingProvider.CreatedOn, UpdatedOn: time.Now(), UpdatedBy: request.UserId},
	}
	err := impl.teamRepository.Update(team)
	if err != nil {
		impl.logger.Errorw("error in updating team", "data", team, "err", err)
		err = &util.ApiError{
			Code:            constants.GitProviderUpdateFailedInDb,
			InternalMessage: "team failed to update in db",
			UserMessage:     "team failed to update in db",
		}
		return nil, err
	}
	request.Id = team.Id
	return request, nil
}

func (impl TeamServiceImpl) FindTeamByAppId(appId int) (*TeamBean, error) {
	team, err := impl.teamRepository.FindTeamByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error while fetching team", "err", err)
		return nil, err
	}
	teamBean := &TeamBean{Id: team.Id, Name: team.Name}

	return teamBean, err
}

func (impl TeamServiceImpl) FindTeamByAppName(appName string) (*TeamBean, error) {
	team, err := impl.teamRepository.FindTeamByAppName(appName)
	if err != nil {
		impl.logger.Errorw("error while fetching team", "err", err)
		return nil, err
	}
	teamBean := &TeamBean{Id: team.Id, Name: team.Name}
	return teamBean, err
}

func (impl TeamServiceImpl) FetchForAutocomplete() ([]TeamRequest, error) {
	impl.logger.Debug("fetch all team from db")
	teams, err := impl.teamRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("error in fetch all team", "err", err)
		return nil, err
	}
	var teamRequests []TeamRequest
	for _, team := range teams {
		providerRes := TeamRequest{
			Id:   team.Id,
			Name: team.Name,
		}
		teamRequests = append(teamRequests, providerRes)
	}
	return teamRequests, err
}

func (impl TeamServiceImpl) FindTeamsByUser(userId int32) ([]team.Team, error) {
	teamsForUser := make(map[string]bool)
	activeUser, err := impl.userService.GetById(int32(userId))
	for _, r := range activeUser.RoleFilters {
		if r.Team != "" {
			teamsForUser[r.Team] = true
		}
	}
	var teams []team.Team
	for t := range teamsForUser {
		// TODO: Get team id for current team
		team, err := impl.teamRepository.FindByTeamName(t)
		if err != nil {
			impl.logger.Errorw("err", err)
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, err
}

func (impl TeamServiceImpl) FindByIds(ids []*int) ([]*TeamBean, error) {
	teams, err := impl.teamRepository.FindByIds(ids)
	if err != nil {
		impl.logger.Errorw("error in fetch all team", "err", err)
		return nil, err
	}
	var teamRequests []*TeamBean
	for _, team := range teams {
		team := &TeamBean{
			Id:   team.Id,
			Name: team.Name,
		}
		teamRequests = append(teamRequests, team)
	}
	return teamRequests, err
}

type TeamBean struct {
	Id   int    `json:"id"`
	Name string `json:"name,notnull"`
}
