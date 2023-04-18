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
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/appGroup"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type AppGroupService interface {
	GetActiveAppGroupList(emailId string, checkAuthBatch func(emailId string, appObject []string) map[string]bool) ([]*AppGroupDto, error)
	GetApplicationsForAppGroup(appGroupId int) ([]*ApplicationDto, error)
	GetAppIdsByAppGroupId(appGroupId int) ([]int, error)
	CreateAppGroup(request *AppGroupDto) (*AppGroupDto, error)
	UpdateAppGroup(request *AppGroupDto) (*AppGroupDto, error)
	DeleteAppGroup(appGroupId int) (bool, error)
}
type AppGroupServiceImpl struct {
	logger                    *zap.SugaredLogger
	appGroupRepository        appGroup.AppGroupRepository
	appGroupMappingRepository appGroup.AppGroupMappingRepository
	enforcerUtil              rbac.EnforcerUtil
}

func NewAppGroupServiceImpl(logger *zap.SugaredLogger, appGroupRepository appGroup.AppGroupRepository,
	appGroupMappingRepository appGroup.AppGroupMappingRepository, enforcerUtil rbac.EnforcerUtil) *AppGroupServiceImpl {
	return &AppGroupServiceImpl{
		logger:                    logger,
		appGroupRepository:        appGroupRepository,
		appGroupMappingRepository: appGroupMappingRepository,
		enforcerUtil:              enforcerUtil,
	}
}

type AppGroupingRequest struct {
	EnvId          int                                                                                             `json:"envId,omitempty"`
	AppGroupId     int                                                                                             `json:"appGroupId,omitempty"`
	AppIds         []int                                                                                           `json:"appIds,omitempty"`
	EmailId        string                                                                                          `json:"emailId,omitempty"`
	CheckAuthBatch func(emailId string, appObject []string, envObject []string) (map[string]bool, map[string]bool) `json:"-"`
	Ctx            context.Context                                                                                 `json:"-"`
	UserId         int32                                                                                           `json:"-"`
}

type AppGroupDto struct {
	Id          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty" validate:"required,max=50,name-component"`
	Description string `json:"description,omitempty" validate:"max=50"`
	AppIds      []int  `json:"appIds,omitempty"`
	Active      bool   `json:"active,omitempty"`
	UserId      int32  `json:"-"`
}

type ApplicationDto struct {
	AppGroupId  int    `json:"appGroupId,omitempty"`
	AppId       int    `json:"appId,omitempty"`
	AppName     string `json:"appName,omitempty"`
	Description string `json:"description,omitempty"`
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
		Name:        request.Name,
		Description: request.Description,
		Active:      true,
		AuditLog:    sql.AuditLog{CreatedOn: time.Now(), CreatedBy: request.UserId, UpdatedBy: request.UserId, UpdatedOn: time.Now()},
	}
	_, err = impl.appGroupRepository.Save(model, tx)
	if err != nil {
		impl.logger.Errorw("error in creating app group", "error", err)
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
			impl.logger.Errorw("error in creating app group", "error", err)
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
		impl.logger.Errorw("error in update app group", "error", err)
		return nil, err
	}
	if existingModel != nil && existingModel.Id > 0 {
		existingModel.Name = request.Name
		existingModel.Description = request.Description
		existingModel.Active = true
		existingModel.UpdatedOn = time.Now()
		existingModel.UpdatedBy = request.UserId
		err = impl.appGroupRepository.Update(existingModel, tx)
		if err != nil {
			impl.logger.Errorw("error in creating app group", "error", err)
			return nil, err
		}
	}

	mappings, err := impl.appGroupMappingRepository.FindByAppGroupId(request.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in update app group", "error", err)
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
			impl.logger.Errorw("error in creating app group", "error", err)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (impl *AppGroupServiceImpl) GetActiveAppGroupList(emailId string, checkAuthBatch func(emailId string, appObject []string) map[string]bool) ([]*AppGroupDto, error) {
	appGroupsDto := make([]*AppGroupDto, 0)
	appGroupMappings, err := impl.appGroupMappingRepository.FindAll()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in update app group", "error", err)
		return nil, err
	}
	var appIds []int
	authorizedAppIds := make(map[int]bool)
	for _, appGroupMapping := range appGroupMappings {
		appIds = append(appIds, appGroupMapping.AppId)
	}

	//authorization block starts here
	var appObjectArr []string
	objects := impl.enforcerUtil.GetRbacObjectsByAppIds(appIds)
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object)
	}
	appResults := checkAuthBatch(emailId, appObjectArr)
	for _, appId := range appIds {
		appObject := objects[appId]
		if !appResults[appObject] {
			//if user unauthorized, skip items
			continue
		}
		authorizedAppIds[appId] = true
	}
	//authorization block ends here

	appIdsMap := make(map[int][]int)
	for _, appGroupMapping := range appGroupMappings {
		// if this app from the group have the permission add in the result set
		if _, ok := authorizedAppIds[appGroupMapping.AppId]; ok {
			appIdsMap[appGroupMapping.AppGroupId] = append(appIdsMap[appGroupMapping.AppGroupId], appGroupMapping.AppId)
		}
	}
	appGroups, err := impl.appGroupRepository.FindActiveList()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in update app group", "error", err)
		return nil, err
	}
	for _, appGroup := range appGroups {
		appIds := appIdsMap[appGroup.Id]
		appGroupDto := &AppGroupDto{
			Id:          appGroup.Id,
			Name:        appGroup.Name,
			Description: appGroup.Description,
			AppIds:      appIds,
		}
		appGroupsDto = append(appGroupsDto, appGroupDto)
	}
	return appGroupsDto, nil
}

func (impl *AppGroupServiceImpl) GetApplicationsForAppGroup(appGroupId int) ([]*ApplicationDto, error) {
	applications := make([]*ApplicationDto, 0)
	appGroups, err := impl.appGroupMappingRepository.FindByAppGroupId(appGroupId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in update app group", "error", err)
		return nil, err
	}
	for _, appGroup := range appGroups {
		appGroupDto := &ApplicationDto{
			AppId:       appGroup.AppId,
			AppName:     appGroup.App.AppName,
			Description: appGroup.AppGroup.Description,
		}
		applications = append(applications, appGroupDto)
	}
	return applications, nil
}

func (impl *AppGroupServiceImpl) GetAppIdsByAppGroupId(appGroupId int) ([]int, error) {
	appIds := make([]int, 0)
	appGroups, err := impl.appGroupMappingRepository.FindByAppGroupId(appGroupId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in update app group", "error", err)
		return nil, err
	}
	for _, appGroup := range appGroups {
		appIds = append(appIds, appGroup.AppId)
	}
	return appIds, nil
}

func (impl *AppGroupServiceImpl) DeleteAppGroup(appGroupId int) (bool, error) {
	dbConnection := impl.appGroupRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return false, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	appGroup, err := impl.appGroupRepository.FindById(appGroupId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in update app group", "error", err)
		return false, err
	}

	appGroupMappings, err := impl.appGroupMappingRepository.FindByAppGroupId(appGroupId)
	if err != nil && err != pg.ErrNoRows {
		return false, err
	}
	for _, appGroupMapping := range appGroupMappings {
		err = impl.appGroupMappingRepository.Delete(appGroupMapping, tx)
		if err != nil {
			return false, err
		}
	}
	appGroup.Name = fmt.Sprintf("%s-deleted", appGroup.Name)
	appGroup.Active = false
	appGroup.UpdatedOn = time.Now()
	appGroup.UpdatedBy = 1
	err = impl.appGroupRepository.Update(appGroup, tx)
	if err != nil {
		return false, err
	}

	err = tx.Commit()
	if err != nil {
		return false, err
	}
	return true, nil
}
