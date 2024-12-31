/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package team

import (
	"errors"
	bean2 "github.com/devtron-labs/devtron/pkg/team/bean"
	"github.com/devtron-labs/devtron/pkg/team/read"
	"github.com/devtron-labs/devtron/pkg/team/repository"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
)

type TeamService interface {
	Create(request *bean2.TeamRequest) (*bean2.TeamRequest, error)
	FetchAllActive() ([]bean2.TeamRequest, error)
	FetchOne(id int) (*bean2.TeamRequest, error)
	Update(request *bean2.TeamRequest) (*bean2.TeamRequest, error)
	Delete(request *bean2.TeamRequest) error
	FetchForAutocomplete() ([]bean2.TeamRequest, error)
}
type TeamServiceImpl struct {
	logger          *zap.SugaredLogger
	teamRepository  repository.TeamRepository
	userAuthService user.UserAuthService
	teamReadService read.TeamReadService
}

func NewTeamServiceImpl(logger *zap.SugaredLogger, teamRepository repository.TeamRepository,
	userAuthService user.UserAuthService,
	teamReadService read.TeamReadService) *TeamServiceImpl {
	return &TeamServiceImpl{
		logger:          logger,
		teamRepository:  teamRepository,
		userAuthService: userAuthService,
		teamReadService: teamReadService,
	}
}

func (impl TeamServiceImpl) Create(request *bean2.TeamRequest) (*bean2.TeamRequest, error) {
	impl.logger.Debugw("team create request", "req", request)
	if len(request.Name) < 3 || strings.Contains(request.Name, " ") {
		impl.logger.Errorw("name should not contain white spaces and should have min 3 chars ")
		err := errors.New("name should not contain white spaces and should have min 3 chars")
		return nil, err
	}
	t := &repository.Team{
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

func (impl TeamServiceImpl) FetchAllActive() ([]bean2.TeamRequest, error) {
	impl.logger.Debug("fetch all team from db")
	teams, err := impl.teamReadService.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetch all team", "err", err)
		return nil, err
	}
	return teams, err
}

func (impl TeamServiceImpl) FetchOne(teamId int) (*bean2.TeamRequest, error) {
	impl.logger.Debug("fetch team by ID from db")
	team, err := impl.teamReadService.FindOne(teamId)
	if err != nil {
		impl.logger.Errorw("error in fetch all team", "err", err)
		return nil, err
	}
	return team, err
}

func (impl TeamServiceImpl) Update(request *bean2.TeamRequest) (*bean2.TeamRequest, error) {
	impl.logger.Debugw("team update request", "req", request)
	existingProvider, err0 := impl.teamReadService.FindOne(request.Id)
	if err0 != nil {
		impl.logger.Errorw("No matching entry found for update.", "id", request.Id)
		err0 = &util.ApiError{
			Code:            constants.GitProviderUpdateProviderNotExists,
			InternalMessage: "team update failed, does not exist",
			UserMessage:     "team update failed, does not exist",
		}
		return nil, err0
	}
	team := &repository.Team{
		Name:     request.Name,
		Id:       request.Id,
		Active:   request.Active,
		AuditLog: sql.AuditLog{CreatedBy: existingProvider.UserId, CreatedOn: existingProvider.CreatedOn, UpdatedOn: time.Now(), UpdatedBy: request.UserId},
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

func (impl TeamServiceImpl) Delete(deleteRequest *bean2.TeamRequest) error {
	existingTeam, err := impl.teamReadService.FindOne(deleteRequest.Id)
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
	deleteReq := &repository.Team{
		Name:     deleteRequest.Name,
		Id:       deleteRequest.Id,
		Active:   deleteRequest.Active,
		AuditLog: sql.AuditLog{CreatedBy: existingTeam.UserId, CreatedOn: existingTeam.CreatedOn, UpdatedOn: time.Now(), UpdatedBy: deleteRequest.UserId},
	}
	err = impl.teamRepository.MarkTeamDeleted(deleteReq, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting team", "teamId", deleteReq.Id, "teamName", deleteReq.Name)
		return err
	}
	//deleting auth roles entries for this project
	err = impl.userAuthService.DeleteRoles(bean.PROJECT_TYPE, deleteRequest.Name, tx, "", "")
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

func (impl TeamServiceImpl) FetchForAutocomplete() ([]bean2.TeamRequest, error) {
	impl.logger.Debug("fetch all team from db")
	teams, err := impl.teamReadService.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetch all team", "err", err)
		return nil, err
	}
	return teams, err
}
