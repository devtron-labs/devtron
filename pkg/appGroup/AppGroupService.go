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

package appGroup

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/appGroup"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type AppGroupService interface {
	GetActiveAppGroupList() ([]*AppGroupDto, error)
	GetApplicationsForAppGroup(appGroupId int) ([]*ApplicationDto, error)
	CreateAppGroup(request *AppGroupDto) (*AppGroupDto, error)
	UpdateAppGroup(request *AppGroupDto) (*AppGroupDto, error)
}
type AppGroupServiceImpl struct {
	logger                    *zap.SugaredLogger
	appGroupRepository        appGroup.AppGroupRepository
	appGroupMappingRepository appGroup.AppGroupMappingRepository
}

func NewAppGroupServiceImpl(logger *zap.SugaredLogger, appGroupRepository appGroup.AppGroupRepository,
	appGroupMappingRepository appGroup.AppGroupMappingRepository) *AppGroupServiceImpl {
	return &AppGroupServiceImpl{
		logger:                    logger,
		appGroupRepository:        appGroupRepository,
		appGroupMappingRepository: appGroupMappingRepository,
	}
}

type AppGroupDto struct {
	Id     int    `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	AppIds []int  `json:"appIds,omitempty"`
	Active bool   `json:"active"`
	UserId int32  `json:"-"`
}

type ApplicationDto struct {
	AppGroupId int    `json:"appGroupId,omitempty"`
	AppId      int    `json:"appId,omitempty"`
	AppName    string `json:"appName,omitempty"`
}

func (impl *AppGroupServiceImpl) CreateAppGroup(request *AppGroupDto) (*AppGroupDto, error) {
	dbConnection := impl.appGroupRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	model := &appGroup.AppGroup{
		Name:     request.Name,
		Active:   true,
		AuditLog: sql.AuditLog{CreatedOn: time.Now(), CreatedBy: request.UserId, UpdatedBy: request.UserId, UpdatedOn: time.Now()},
	}
	_, err = impl.appGroupRepository.Save(model, tx)
	if err != nil {
		impl.logger.Errorw("error in creating new attributes", "error", err)
		return nil, err
	}

	for _, appId := range request.AppIds {
		mapping := &appGroup.AppGroupMapping{
			AppGroupId: model.Id,
			AppId:      appId,
			AuditLog:   sql.AuditLog{CreatedOn: time.Now(), CreatedBy: request.UserId, UpdatedBy: request.UserId, UpdatedOn: time.Now()},
		}
		_, err = impl.appGroupMappingRepository.Save(mapping, tx)
		if err != nil {
			impl.logger.Errorw("error in creating new attributes", "error", err)
			return nil, err
		}
	}

	existingModel, err := impl.appGroupRepository.FindById(request.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in update new attributes", "error", err)
		return nil, err
	}
	if existingModel != nil && existingModel.Id > 0 {
		existingModel.Active = false
		err = impl.appGroupRepository.Update(existingModel, tx)
		if err != nil {
			impl.logger.Errorw("error in creating new attributes", "error", err)
			return nil, err
		}
	}

	request.Id = model.Id
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (impl *AppGroupServiceImpl) UpdateAppGroup(request *AppGroupDto) (*AppGroupDto, error) {
	dbConnection := impl.appGroupRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	existingModel, err := impl.appGroupRepository.FindById(request.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in update new attributes", "error", err)
		return nil, err
	}
	if existingModel != nil && existingModel.Id > 0 {
		existingModel.Name = request.Name
		existingModel.Active = false
		err = impl.appGroupRepository.Update(existingModel, tx)
		if err != nil {
			impl.logger.Errorw("error in creating new attributes", "error", err)
			return nil, err
		}
	}

	mappings, err := impl.appGroupMappingRepository.FindByAppGroupId(request.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in update new attributes", "error", err)
		return nil, err
	}
	existingAppIds := make(map[int]int)
	for _, mapping := range mappings {
		existingAppIds[mapping.AppId] = mapping.AppId
	}
	for _, appId := range request.AppIds {
		if _, ok := existingAppIds[appId]; ok {
			// app already added in mapping
			continue
		}
		mapping := &appGroup.AppGroupMapping{
			AppGroupId: existingModel.Id,
			AppId:      appId,
			AuditLog:   sql.AuditLog{CreatedOn: time.Now(), CreatedBy: request.UserId, UpdatedBy: request.UserId, UpdatedOn: time.Now()},
		}
		_, err = impl.appGroupMappingRepository.Save(mapping, tx)
		if err != nil {
			impl.logger.Errorw("error in creating new attributes", "error", err)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (impl *AppGroupServiceImpl) GetActiveAppGroupList() ([]*AppGroupDto, error) {
	var appGroupsDto []*AppGroupDto
	appGroups, err := impl.appGroupRepository.FindActiveList()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in update new attributes", "error", err)
		return nil, err
	}
	for _, appGroup := range appGroups {
		appGroupDto := &AppGroupDto{
			Id:     appGroup.Id,
			Name:   appGroup.Name,
			Active: appGroup.Active,
		}
		appGroupsDto = append(appGroupsDto, appGroupDto)
	}
	return appGroupsDto, nil
}

func (impl *AppGroupServiceImpl) GetApplicationsForAppGroup(appGroupId int) ([]*ApplicationDto, error) {
	var applications []*ApplicationDto
	appGroups, err := impl.appGroupMappingRepository.FindByAppGroupId(appGroupId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in update new attributes", "error", err)
		return nil, err
	}
	for _, appGroup := range appGroups {
		appGroupDto := &ApplicationDto{
			AppId:   appGroup.AppId,
			AppName: appGroup.App.AppName,
		}
		applications = append(applications, appGroupDto)
	}
	return applications, nil
}
