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

package attributes

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/attributes/adapter"
	"github.com/devtron-labs/devtron/pkg/attributes/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"reflect"
)

type UserAttributesService interface {
	AddUserAttributes(request *bean.UserAttributesDto) (*bean.UserAttributesDto, error)
	UpdateUserAttributes(request *bean.UserAttributesDto) (*bean.UserAttributesDto, error)
	PatchUserAttributes(request *bean.UserAttributesDto) (*bean.UserAttributesDto, error)
	GetUserAttribute(request *bean.UserAttributesDto) (*bean.UserAttributesDto, error)
}

type UserAttributesServiceImpl struct {
	logger               *zap.SugaredLogger
	attributesRepository repository.UserAttributesRepository
}

func NewUserAttributesServiceImpl(logger *zap.SugaredLogger,
	attributesRepository repository.UserAttributesRepository) *UserAttributesServiceImpl {
	serviceImpl := &UserAttributesServiceImpl{
		logger:               logger,
		attributesRepository: attributesRepository,
	}
	return serviceImpl
}

func (impl UserAttributesServiceImpl) AddUserAttributes(request *bean.UserAttributesDto) (*bean.UserAttributesDto, error) {
	dao := &repository.UserAttributesDao{
		EmailId: request.EmailId,
		Key:     request.Key,
		Value:   request.Value,
		UserId:  request.UserId,
	}
	_, err := impl.attributesRepository.AddUserAttribute(dao)
	if err != nil {
		impl.logger.Errorw("error in creating new user attributes for req", "req", request, "error", err)
		return nil, errors.New("error occurred while creating user attributes")
	}
	return request, nil
}

func (impl UserAttributesServiceImpl) UpdateUserAttributes(request *bean.UserAttributesDto) (*bean.UserAttributesDto, error) {

	userAttribute, err := impl.GetUserAttribute(request)
	if err != nil {
		impl.logger.Errorw("error while getting user attributes during update request", "req", request, "error", err)
		return nil, errors.New("error occurred while updating user attributes")
	}
	if userAttribute == nil {
		impl.logger.Info("not data found for request, so going to add instead of update", "req", request)
		attributes, err := impl.AddUserAttributes(request)
		if err != nil {
			impl.logger.Errorw("error in adding new user attributes", "req", request, "error", err)
			return nil, errors.New("error occurred while updating user attributes")
		}
		return attributes, nil
	}
	dao := &repository.UserAttributesDao{
		EmailId: request.EmailId,
		Key:     request.Key,
		Value:   request.Value,
		UserId:  request.UserId,
	}
	err = impl.attributesRepository.UpdateDataValByKey(dao)
	if err != nil {
		impl.logger.Errorw("error in update new attributes", "req", request, "error", err)
		return nil, errors.New("error occurred while updating user attributes")
	}
	return request, nil
}

func (impl UserAttributesServiceImpl) PatchUserAttributes(request *bean.UserAttributesDto) (*bean.UserAttributesDto, error) {
	existingAttribute, err := impl.GetUserAttribute(request)
	if err != nil {
		impl.logger.Errorw("error while getting user attributes during patch request", "req", request, "error", err)
		return nil, errors.New("error occurred while getting user attributes")
	}

	if existingAttribute == nil {
		impl.logger.Info("no data found for request, so going to add instead of update", "req", request)
		newAttribute, err := impl.AddUserAttributes(request)
		if err != nil {
			impl.logger.Errorw("error in adding new user attributes", "req", request, "error", err)
			return nil, errors.New("error occurred while adding user attributes")
		}
		return newAttribute, nil
	}

	existingData, err := impl.parseJSONValue(existingAttribute.Value, "existing")
	if err != nil {
		impl.logger.Errorw("error in parsing json", "existingAttribute.Value", existingAttribute.Value, "error", err)
		return nil, err
	}

	newData, err := impl.parseJSONValue(request.Value, "request")
	if err != nil {
		impl.logger.Errorw("error in parsing request json", "request.Value", request.Value, "error", err)
		return nil, err
	}

	hasChanges := impl.mergeUserAttributesData(existingData, newData)
	if !hasChanges {
		impl.logger.Infow("no changes detected, skipping update", "key", request.Key)
		return existingAttribute, nil
	}

	return impl.updateAttributeInDatabase(request, existingData)
}

// parseJSONValue parses a JSON string into a map, with proper error handling
func (impl UserAttributesServiceImpl) parseJSONValue(jsonValue, context string) (map[string]interface{}, error) {
	var data map[string]interface{}

	if jsonValue == "" {
		return make(map[string]interface{}), nil
	}

	err := json.Unmarshal([]byte(jsonValue), &data)
	if err != nil {
		impl.logger.Errorw("error parsing JSON value", "context", context, "value", jsonValue, "error", err)
		return nil, errors.New("error occurred while parsing user attributes data")
	}

	return data, nil
}

// updateAttributeInDatabase updates the merged data in the database
func (impl UserAttributesServiceImpl) updateAttributeInDatabase(request *bean.UserAttributesDto, mergedData map[string]interface{}) (*bean.UserAttributesDto, error) {
	mergedJSON, err := json.Marshal(mergedData)
	if err != nil {
		impl.logger.Errorw("error converting merged data to JSON", "data", mergedData, "error", err)
		return nil, errors.New("error occurred while processing user attributes")
	}

	dao := &repository.UserAttributesDao{
		EmailId: request.EmailId,
		Key:     request.Key,
		Value:   string(mergedJSON),
		UserId:  request.UserId,
	}

	// Update in database
	err = impl.attributesRepository.UpdateDataValByKey(dao)
	if err != nil {
		impl.logger.Errorw("error updating user attributes in database", "dao", dao, "error", err)
		return nil, errors.New("error occurred while updating user attributes")
	}

	// Build and return response
	return adapter.BuildResponseDTO(request, string(mergedJSON)), nil
}

// mergeUserAttributesData merges newData into existingData with special handling for resources
func (impl UserAttributesServiceImpl) mergeUserAttributesData(existingData, newData map[string]interface{}) bool {
	hasChanges := false

	for key, newValue := range newData {
		if key == bean.UserPreferencesResourcesKey {
			// Special handling for resources - merge nested structure
			if impl.mergeResourcesData(existingData, newValue) {
				hasChanges = true
			}
		} else {
			if impl.mergeStandardAttribute(existingData, key, newValue) {
				hasChanges = true
			}
		}
	}

	return hasChanges
}

// mergeStandardAttribute merges a standard (non-resource) attribute
func (impl UserAttributesServiceImpl) mergeStandardAttribute(existingData map[string]interface{}, key string, newValue interface{}) bool {
	existingValue, exists := existingData[key]
	if !exists || !reflect.DeepEqual(existingValue, newValue) {
		existingData[key] = newValue
		return true
	}
	return false
}

// mergeResourcesData handles the special merging logic for the resources object
func (impl UserAttributesServiceImpl) mergeResourcesData(existingData map[string]interface{}, newResourcesValue interface{}) bool {
	impl.ensureResourcesStructureExists(existingData)

	existingResources, ok := existingData[bean.UserPreferencesResourcesKey].(map[string]interface{})
	if !ok {
		existingData[bean.UserPreferencesResourcesKey] = newResourcesValue
		return true
	}

	newResources, ok := newResourcesValue.(map[string]interface{})
	if !ok {
		existingData[bean.UserPreferencesResourcesKey] = newResourcesValue
		return true
	}

	return impl.mergeResourceTypes(existingResources, newResources)
}

// ensureResourcesStructureExists initializes the resources structure if it doesn't exist
func (impl UserAttributesServiceImpl) ensureResourcesStructureExists(existingData map[string]interface{}) {
	if existingData[bean.UserPreferencesResourcesKey] == nil {
		existingData[bean.UserPreferencesResourcesKey] = make(map[string]interface{})
	}
}

// mergeResourceTypes merges individual resource types from new resources into existing resources
func (impl UserAttributesServiceImpl) mergeResourceTypes(existingResources, newResources map[string]interface{}) bool {
	hasChanges := false

	// Merge each resource type from newResources
	for resourceType, newResourceData := range newResources {
		existingResourceData, exists := existingResources[resourceType]
		if !exists || !reflect.DeepEqual(existingResourceData, newResourceData) {
			existingResources[resourceType] = newResourceData
			hasChanges = true
		}
	}

	return hasChanges
}

func (impl UserAttributesServiceImpl) GetUserAttribute(request *bean.UserAttributesDto) (*bean.UserAttributesDto, error) {

	dao := &repository.UserAttributesDao{
		EmailId: request.EmailId,
		Key:     request.Key,
		Value:   request.Value,
		UserId:  request.UserId,
	}
	modelValue, err := impl.attributesRepository.GetDataValueByKey(dao)
	if err == pg.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		impl.logger.Errorw("error in fetching user attributes", "req", request, "error", err)
		return nil, errors.New("error occurred while getting user attributes")
	}
	resAttrDto := &bean.UserAttributesDto{
		EmailId: request.EmailId,
		Key:     request.Key,
		Value:   modelValue,
	}
	return resAttrDto, nil
}
