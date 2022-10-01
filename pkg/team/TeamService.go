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
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	"go.uber.org/zap"
	"time"
)

type TeamService interface {
	Create(request *TeamRequest) (*TeamRequest, error)
	FetchAllActive() ([]TeamRequest, error)
	FetchOne(id int) (*TeamRequest, error)
	Update(request *TeamRequest) (*TeamRequest, error)
	Delete(request *TeamRequest) error
	FetchForAutocomplete() ([]TeamRequest, error)
	FindByIds(ids []*int) ([]*TeamBean, error)
	FindByTeamName(teamName string) (*TeamRequest, error)
}
type TeamServiceImpl struct {
	logger          *zap.SugaredLogger
	teamRepository  TeamRepository
	userAuthService user.UserAuthService
}

type TeamRequest struct {
	Id        int       `json:"id,omitempty" validate:"number"`
	Name      string    `json:"name,omitempty" validate:"required"`
	Active    bool      `json:"active"`
	UserId    int32     `json:"-"`
	CreatedOn time.Time `json:"-"`
}

func NewTeamServiceImpl(logger *zap.SugaredLogger, teamRepository TeamRepository,
	userAuthService user.UserAuthService) *TeamServiceImpl {
	return &TeamServiceImpl{
		logger:          logger,
		teamRepository:  teamRepository,
		userAuthService: userAuthService,
	}
}

func (impl TeamServiceImpl) Create(request *TeamRequest) (*TeamRequest, error) {
	impl.logger.Debugw("team create request", "req", request)
	t := &Team{
		Name:     request.Name,
		Id:       request.Id,
		Active:   request.Active,
		AuditLog: sql.AuditLog{CreatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: request.UserId},
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

func (impl TeamServiceImpl) FetchAllActive() ([]TeamRequest, error) {
	impl.logger.Debug("fetch all team from db")
	teams, err := impl.teamRepository.FindAllActive()
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
		Id:        team.Id,
		Name:      team.Name,
		Active:    team.Active,
		UserId:    team.CreatedBy,
		CreatedOn: team.CreatedOn,
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
	team := &Team{
		Name:     request.Name,
		Id:       request.Id,
		Active:   request.Active,
		AuditLog: sql.AuditLog{CreatedBy: existingProvider.CreatedBy, CreatedOn: existingProvider.CreatedOn, UpdatedOn: time.Now(), UpdatedBy: request.UserId},
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

func (impl TeamServiceImpl) Delete(deleteRequest *TeamRequest) error {
	existingTeam, err := impl.teamRepository.FindOne(deleteRequest.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete.", "id", deleteRequest.Id)
		return err
	}
	dbConnection := impl.teamRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in establishing connection", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	deleteReq := &Team{
		Name:     deleteRequest.Name,
		Id:       deleteRequest.Id,
		Active:   deleteRequest.Active,
		AuditLog: sql.AuditLog{CreatedBy: existingTeam.CreatedBy, CreatedOn: existingTeam.CreatedOn, UpdatedOn: time.Now(), UpdatedBy: deleteRequest.UserId},
	}
	err = impl.teamRepository.MarkTeamDeleted(deleteReq, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting team", "teamId", deleteReq.Id, "teamName", deleteReq.Name)
		return err
	}
	//deleting auth roles entries for this project
	err = impl.userAuthService.DeleteRoles(repository.PROJECT_TYPE, deleteRequest.Name, tx, "")
	if err != nil {
		impl.logger.Errorw("error in deleting auth roles", "err", err)
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (impl TeamServiceImpl) FetchForAutocomplete() ([]TeamRequest, error) {
	impl.logger.Debug("fetch all team from db")
	teams, err := impl.teamRepository.FindAllActive()
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

func (impl TeamServiceImpl) FindByTeamName(teamName string) (*TeamRequest, error) {
	impl.logger.Debug("fetch team by name from db")
	team, err := impl.teamRepository.FindByTeamName(teamName)
	if err != nil {
		impl.logger.Errorw("error in fetch team", "err", err)
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

type TeamBean struct {
	Id   int    `json:"id"`
	Name string `json:"name,notnull"`
}
