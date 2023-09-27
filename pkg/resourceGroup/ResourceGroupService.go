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

package resourceGroup

import (
	"context"
	"github.com/devtron-labs/devtron/internal/sql/repository/resourceGroup"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strings"
	"time"
)

type ResourceGroupService interface {
	GetActiveResourceGroupList(emailId string, checkAuthBatch func(emailId string, appObject []string, action string) map[string]bool, envId int, groupType ResourceGroupType) ([]*ResourceGroupDto, error)
	//GetApplicationsForResourceGroup(appGroupId int) ([]*ApplicationDto, error)
	GetResourceIdsByResourceGroupId(resourceGroupId int) ([]int, error)
	CreateResourceGroup(request *ResourceGroupDto) (*ResourceGroupDto, error)
	UpdateResourceGroup(request *ResourceGroupDto) (*ResourceGroupDto, error)
	CheckResourceGroupPermissions(request *ResourceGroupDto) (bool, error)
	DeleteResourceGroup(resourceGroupId int, groupType ResourceGroupType, emailId string, checkAuthBatch func(emailId string, appObject []string, action string) map[string]bool) (bool, error)
}
type ResourceGroupServiceImpl struct {
	logger                         *zap.SugaredLogger
	resourceGroupRepository        resourceGroup.ResourceGroupRepository
	resourceGroupMappingRepository resourceGroup.ResourceGroupMappingRepository
	enforcerUtil                   rbac.EnforcerUtil
	devtronResourceService         devtronResource.DevtronResourceService
}

func NewResourceGroupServiceImpl(logger *zap.SugaredLogger, resourceGroupRepository resourceGroup.ResourceGroupRepository,
	resourceGroupMappingRepository resourceGroup.ResourceGroupMappingRepository, enforcerUtil rbac.EnforcerUtil,
	devtronResourceService devtronResource.DevtronResourceService,
) *ResourceGroupServiceImpl {
	return &ResourceGroupServiceImpl{
		logger:                         logger,
		resourceGroupRepository:        resourceGroupRepository,
		resourceGroupMappingRepository: resourceGroupMappingRepository,
		enforcerUtil:                   enforcerUtil,
		devtronResourceService:         devtronResourceService,
	}
}

type ResourceGroupingRequest struct {
	ResourceGroupId   int
	ParentResourceId  int
	ResourceGroupType ResourceGroupType
	ResourceIds       []int                                                                                           `json:"appIds,omitempty"`
	EmailId           string                                                                                          `json:"emailId,omitempty"`
	CheckAuthBatch    func(emailId string, appObject []string, envObject []string) (map[string]bool, map[string]bool) `json:"-"`
	Ctx               context.Context                                                                                 `json:"-"`
	UserId            int32                                                                                           `json:"-"`
}

type ResourceGroupType string

const (
	APP_GROUP ResourceGroupType = "app-group"
	ENV_GROUP ResourceGroupType = "env-group"
)

func (groupType ResourceGroupType) getResourceKey(resourceKeyToId map[bean.DevtronResourceSearchableKeyName]int) int {
	switch groupType {
	case ENV_GROUP:
		return resourceKeyToId[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID]
	case APP_GROUP:
		return resourceKeyToId[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID]
	}
	return 0
}

func (groupType ResourceGroupType) getMappedResourceKey(resourceKeyToId map[bean.DevtronResourceSearchableKeyName]int) int {
	switch groupType {
	case ENV_GROUP:
		return resourceKeyToId[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID]
	case APP_GROUP:
		return resourceKeyToId[bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID]
	}
	return 0
}

type ResourceGroupDto struct {
	Id               int                                                                     `json:"id,omitempty"`
	Name             string                                                                  `json:"name,omitempty" validate:"required,max=30,name-component"`
	Description      string                                                                  `json:"description,omitempty" validate:"max=50"`
	ResourceIds      []int                                                                   `json:"resourceIds"`
	Active           bool                                                                    `json:"active,omitempty"`
	ParentResourceId int                                                                     `json:"parentResourceId"`
	GroupType        ResourceGroupType                                                       `json:"groupType,omitempty"`
	UserId           int32                                                                   `json:"-"`
	EmailId          string                                                                  `json:"-"`
	CheckAuthBatch   func(emailId string, appObject []string, action string) map[string]bool `json:"-"`

	//for backward compatibility
	AppIds        []int `json:"appIds,omitempty"`
	EnvironmentId int   `json:"environmentId,omitempty"`
}

//type ApplicationDto struct {
//	ResourceGroupId int    `json:"appGroupId,omitempty"`
//	AppId           int    `json:"appId,omitempty"`
//	AppName         string `json:"appName,omitempty"`
//	EnvironmentId   int    `json:"environmentId,omitempty"`
//	Description     string `json:"description,omitempty"`
//}

func (impl *ResourceGroupServiceImpl) CreateResourceGroup(request *ResourceGroupDto) (*ResourceGroupDto, error) {

	resourceKeyToId := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()
	err := impl.checkAuthForResourceGroup(request, casbin.ActionCreate)
	if err != nil {
		return nil, err
	}

	dbConnection := impl.resourceGroupRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	existingModel, err := impl.resourceGroupRepository.FindByNameAndParentResource(request.Name, request.ParentResourceId, request.GroupType.getResourceKey(resourceKeyToId))
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting group", "error", err)
		return nil, err
	}

	if err == nil && existingModel.Id > 0 {
		if existingModel.Name == request.Name {
			err = &util.ApiError{Code: "409", HttpStatusCode: 409, UserMessage: "group name already exists"}
			return nil, err
		}
	}

	model := &resourceGroup.ResourceGroup{
		Name:        request.Name,
		Description: request.Description,
		ResourceId:  request.ParentResourceId,
		ResourceKey: request.GroupType.getResourceKey(resourceKeyToId),
		Active:      true,
		AuditLog:    sql.AuditLog{CreatedOn: time.Now(), CreatedBy: request.UserId, UpdatedBy: request.UserId, UpdatedOn: time.Now()},
	}

	_, err = impl.resourceGroupRepository.Save(model, tx)
	if err != nil {
		impl.logger.Errorw("error in creating resource group", "error", err)
		return nil, err
	}

	for _, resourceId := range request.ResourceIds {
		mapping := &resourceGroup.ResourceGroupMapping{
			ResourceGroupId: model.Id,
			ResourceId:      resourceId,
			ResourceKey:     request.GroupType.getMappedResourceKey(resourceKeyToId), //resourceKeyToId[mappedResource.ResourceKey],
			AuditLog:        sql.AuditLog{CreatedOn: time.Now(), CreatedBy: request.UserId, UpdatedBy: request.UserId, UpdatedOn: time.Now()},
		}
		_, err = impl.resourceGroupMappingRepository.Save(mapping, tx)
		if err != nil {
			impl.logger.Errorw("error in creating resource group", "error", err)
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

func (impl *ResourceGroupServiceImpl) UpdateResourceGroup(request *ResourceGroupDto) (*ResourceGroupDto, error) {

	resourceKeyToId := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()

	// fetching existing resourceIds in resource group
	mappings, err := impl.resourceGroupMappingRepository.FindByResourceGroupId(request.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting group mappings", "error", err, "id", request.Id)
		return nil, err
	}

	dbConnection := impl.resourceGroupRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	existingModel, err := impl.resourceGroupRepository.FindById(request.Id)
	if err != nil {
		impl.logger.Errorw("error in getting resource group", "error", err, "id", request.Id)
		return nil, err
	}
	request.ParentResourceId = existingModel.ResourceId
	err = impl.checkAuthForResourceGroup(request, casbin.ActionUpdate)
	if err != nil {
		return nil, err
	}

	if existingModel != nil && existingModel.Id > 0 {
		existingModel.Description = request.Description
		existingModel.Active = true
		existingModel.UpdatedOn = time.Now()
		existingModel.UpdatedBy = request.UserId
		err = impl.resourceGroupRepository.Update(existingModel, tx)
		if err != nil {
			impl.logger.Errorw("error in updating app group", "error", err)
			return nil, err
		}
	}
	requestedResourceIds := make(map[int]int)
	for _, resourceId := range request.ResourceIds {
		requestedResourceIds[resourceId] = resourceId
		//appIdsForAuthorization[appId] = appId
	}
	requestedToEliminate := make(map[int]*resourceGroup.ResourceGroupMapping)
	existingResourceIds := make(map[int]int)
	for _, mapping := range mappings {
		existingResourceIds[mapping.ResourceId] = mapping.ResourceId
		if _, ok := requestedResourceIds[mapping.ResourceId]; !ok {
			//this resource is not in request, need to eliminate
			requestedToEliminate[mapping.ResourceId] = mapping
		}
		//appIdsForAuthorization[mapping.ResourceId] = mapping.ResourceId
	}

	for _, resourceId := range request.ResourceIds {
		if _, ok := existingResourceIds[resourceId]; ok {
			// app already added in mapping
			continue
		}
		mapping := &resourceGroup.ResourceGroupMapping{
			ResourceGroupId: existingModel.Id,
			ResourceId:      resourceId,
			ResourceKey:     request.GroupType.getMappedResourceKey(resourceKeyToId), //resourceKeyToId[mappedResource.ResourceKey],
			AuditLog:        sql.AuditLog{CreatedOn: time.Now(), CreatedBy: request.UserId, UpdatedBy: request.UserId, UpdatedOn: time.Now()},
		}
		_, err = impl.resourceGroupMappingRepository.Save(mapping, tx)
		if err != nil {
			impl.logger.Errorw("error in creating resource group", "error", err)
			return nil, err
		}
	}

	for _, resourceGroupMapping := range requestedToEliminate {
		err = impl.resourceGroupMappingRepository.Delete(resourceGroupMapping, tx)
		if err != nil {
			impl.logger.Errorw("error in deleting resource group mapping", "error", err)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (impl *ResourceGroupServiceImpl) GetActiveResourceGroupList(emailId string, checkAuthBatch func(emailId string, appObject []string, action string) map[string]bool, parentResourceId int, groupType ResourceGroupType) ([]*ResourceGroupDto, error) {
	resourceKeyToId := impl.devtronResourceService.GetAllSearchableKeyNameIdMap()
	resourceGroupDtos := make([]*ResourceGroupDto, 0)
	var resourceGroupIds []int
	resourceGroups, err := impl.resourceGroupRepository.FindActiveListByParentResource(parentResourceId, groupType.getResourceKey(resourceKeyToId))

	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting resource group", "error", err, "id", parentResourceId)
		return nil, err
	}
	for _, resourceGroup := range resourceGroups {
		resourceGroupIds = append(resourceGroupIds, resourceGroup.Id)
	}
	if len(resourceGroupIds) == 0 {
		return resourceGroupDtos, nil
	}
	resourceGroupMappings, err := impl.resourceGroupMappingRepository.FindByResourceGroupIds(resourceGroupIds)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting group mappings", "error", err, "ids", resourceGroupIds)
		return nil, err
	}

	resourceIdsMap := make(map[int][]int)

	var resourceIds []int
	authorizedResources := make(map[int]bool)
	for _, resourceGroupMapping := range resourceGroupMappings {
		resourceIds = append(resourceIds, resourceGroupMapping.ResourceId)
	}

	//authorization block starts here
	var rbacObjectArr []string
	var objects map[int]string
	if groupType == APP_GROUP {
		objects = impl.enforcerUtil.GetRbacObjectsByAppIds(resourceIds)
	} else if groupType == ENV_GROUP {
		objects = impl.enforcerUtil.GetRbacObjectsByEnvIdsAndAppId(resourceIds, parentResourceId)
	}
	for _, object := range objects {
		rbacObjectArr = append(rbacObjectArr, object)
	}
	results := checkAuthBatch(emailId, rbacObjectArr, casbin.ActionGet)
	for _, resourceId := range resourceIds {
		appObject := objects[resourceId]
		if !results[appObject] {
			//if user unauthorized, skip items
			continue
		}
		authorizedResources[resourceId] = true
	}
	//authorization block ends here

	//resourceIdsMap := make(map[int][]int)
	for _, resourceGroupMapping := range resourceGroupMappings {
		// if this resource from the group have the permission add in the result set
		if _, ok := authorizedResources[resourceGroupMapping.ResourceId]; ok {
			resourceIdsMap[resourceGroupMapping.ResourceGroupId] = append(resourceIdsMap[resourceGroupMapping.ResourceGroupId], resourceGroupMapping.ResourceId)
		}
	}

	for _, resourceGroup := range resourceGroups {
		resourceIDs := resourceIdsMap[resourceGroup.Id]
		if len(resourceIDs) > 0 {
			resourceGroupDto := &ResourceGroupDto{
				Id:               resourceGroup.Id,
				Name:             resourceGroup.Name,
				Description:      resourceGroup.Description,
				ResourceIds:      resourceIDs,
				GroupType:        groupType,
				ParentResourceId: resourceGroup.ResourceId,
			}
			//backward compatibility
			if groupType == APP_GROUP {
				resourceGroupDto.AppIds = resourceIDs
				resourceGroupDto.EnvironmentId = parentResourceId
			}

			resourceGroupDtos = append(resourceGroupDtos, resourceGroupDto)
		}
	}
	return resourceGroupDtos, nil
}

//func (impl *ResourceGroupServiceImpl) GetApplicationsForResourceGroup(appGroupId int) ([]*ApplicationDto, error) {
//	applications := make([]*ApplicationDto, 0)
//	appGroups, err := impl.resourceGroupMappingRepository.FindByResourceGroupId(appGroupId)
//	if err != nil && err != pg.ErrNoRows {
//		impl.logger.Errorw("error in update app group", "error", err)
//		return nil, err
//	}
//	for _, appGroup := range appGroups {
//		appGroupDto := &ApplicationDto{
//			AppId:       appGroup.ResourceId,
//			Description: appGroup.ResourceGroup.Description,
//		}
//		applications = append(applications, appGroupDto)
//	}
//	return applications, nil
//}

func (impl *ResourceGroupServiceImpl) GetResourceIdsByResourceGroupId(resourceGroupId int) ([]int, error) {
	resourceIds := make([]int, 0)
	resourceGroups, err := impl.resourceGroupMappingRepository.FindByResourceGroupId(resourceGroupId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting resource groups", "error", err, "id", resourceGroupId)
		return nil, err
	}
	for _, resourceGroup := range resourceGroups {
		resourceIds = append(resourceIds, resourceGroup.ResourceId)
	}
	return resourceIds, nil
}

func (impl *ResourceGroupServiceImpl) DeleteResourceGroup(resourceGroupId int, groupType ResourceGroupType, emailId string, checkAuthBatch func(emailId string, appObject []string, action string) map[string]bool) (bool, error) {
	dbConnection := impl.resourceGroupRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return false, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	resourceIdsForAuthorization := make(map[int]int)
	mappings, err := impl.resourceGroupMappingRepository.FindByResourceGroupId(resourceGroupId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetch group mappings", "error", err, "id", resourceGroupId)
		return false, err
	}

	savedResourceGroup, err := impl.resourceGroupRepository.FindById(resourceGroupId)
	if err != nil {
		impl.logger.Errorw("error in getting resource group", "error", err, "id", resourceGroupId)
		return false, err
	}

	for _, mapping := range mappings {
		resourceIdsForAuthorization[mapping.ResourceId] = mapping.ResourceId
	}
	unauthorizedResources, err := impl.checkAuthForResources(resourceIdsForAuthorization, emailId, checkAuthBatch, casbin.ActionDelete, groupType, savedResourceGroup.ResourceId)
	if err != nil {
		return false, err
	}
	if len(unauthorizedResources) > 0 {
		userMessage := make(map[string]interface{})
		userMessage["message"] = "unauthorized for few requested resources"
		userMessage["unauthorizedApps"] = unauthorizedResources
		userMessage["unauthorizedResources"] = unauthorizedResources
		err = &util.ApiError{Code: "403", HttpStatusCode: 403, UserMessage: userMessage}
		return false, err
	}

	savedResourceGroup.Active = false
	savedResourceGroup.UpdatedOn = time.Now()
	savedResourceGroup.UpdatedBy = 1
	err = impl.resourceGroupRepository.Update(savedResourceGroup, tx)
	if err != nil {
		return false, err
	}

	err = tx.Commit()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (impl *ResourceGroupServiceImpl) checkAuthForResources(resourceIdsForAuthorization map[int]int, emailId string,
	checkAuthBatch func(emailId string, appObject []string, action string) map[string]bool, action string, groupType ResourceGroupType, parentResourceId int) ([]string, error) {
	//authorization block starts here
	unauthorizedResources := make([]string, 0)
	var rbacObjectArr []string
	var resourceIds []int
	for resourceId, _ := range resourceIdsForAuthorization {
		resourceIds = append(resourceIds, resourceId)
	}

	var objects map[int]string
	if groupType == APP_GROUP {
		objects = impl.enforcerUtil.GetRbacObjectsByAppIds(resourceIds)
	} else if groupType == ENV_GROUP {
		objects = impl.enforcerUtil.GetRbacObjectsByEnvIdsAndAppId(resourceIds, parentResourceId)
	}

	for _, object := range objects {
		rbacObjectArr = append(rbacObjectArr, object)
	}

	results := checkAuthBatch(emailId, rbacObjectArr, action)
	for _, resourceId := range resourceIds {
		resourceObject := objects[resourceId]
		if !results[resourceObject] {
			//if user unauthorized
			unauthorizedResources = append(unauthorizedResources, strings.Split(resourceObject, "/")[1])
		}
	}
	//authorization block ends here
	return unauthorizedResources, nil
}

func (impl *ResourceGroupServiceImpl) checkAuthForResourceGroup(request *ResourceGroupDto, action string) error {
	resourceIdsForAuthorization := make(map[int]int)
	for _, resourceId := range request.ResourceIds {
		resourceIdsForAuthorization[resourceId] = resourceId
	}
	unauthorizedResources, err := impl.checkAuthForResources(resourceIdsForAuthorization, request.EmailId, request.CheckAuthBatch, action, request.GroupType, request.ParentResourceId)
	if err != nil {
		return err
	}
	if len(unauthorizedResources) > 0 {
		userMessage := make(map[string]interface{})
		userMessage["message"] = "unauthorized for few requested apps"
		userMessage["unauthorizedApps"] = unauthorizedResources
		userMessage["unauthorizedResources"] = unauthorizedResources
		err = &util.ApiError{Code: "403", HttpStatusCode: 403, UserMessage: userMessage}
		return err
	}
	return nil
}

func (impl *ResourceGroupServiceImpl) CheckResourceGroupPermissions(request *ResourceGroupDto) (bool, error) {
	err := impl.checkAuthForResourceGroup(request, casbin.ActionCreate)
	if err != nil {
		return false, err
	}
	return true, nil
}
