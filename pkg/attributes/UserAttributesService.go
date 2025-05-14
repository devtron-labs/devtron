/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package attributes

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"reflect"
)

type UserAttributesService interface {
	AddUserAttributes(request *UserAttributesDto) (*UserAttributesDto, error)
	UpdateUserAttributes(request *UserAttributesDto) (*UserAttributesDto, error)
	PatchUserAttributes(request *UserAttributesDto) (*UserAttributesDto, error)
	GetUserAttribute(request *UserAttributesDto) (*UserAttributesDto, error)
}

type UserAttributesServiceImpl struct {
	logger               *zap.SugaredLogger
	attributesRepository repository.UserAttributesRepository
}

type UserAttributesDto struct {
	EmailId string `json:"emailId"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	UserId  int32  `json:"-"`
}

func NewUserAttributesServiceImpl(logger *zap.SugaredLogger,
	attributesRepository repository.UserAttributesRepository) *UserAttributesServiceImpl {
	serviceImpl := &UserAttributesServiceImpl{
		logger:               logger,
		attributesRepository: attributesRepository,
	}
	return serviceImpl
}

func (impl UserAttributesServiceImpl) AddUserAttributes(request *UserAttributesDto) (*UserAttributesDto, error) {
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

func (impl UserAttributesServiceImpl) UpdateUserAttributes(request *UserAttributesDto) (*UserAttributesDto, error) {

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

func (impl UserAttributesServiceImpl) PatchUserAttributes(request *UserAttributesDto) (*UserAttributesDto, error) {
	userAttribute, err := impl.GetUserAttribute(request)
	if err != nil {
		impl.logger.Errorw("error while getting user attributes during patch request", "req", request, "error", err)
		return nil, errors.New("error occurred while updating user attributes")
	}
	if userAttribute == nil {
		impl.logger.Info("no data found for request, so going to add instead of update", "req", request)
		attributes, err := impl.AddUserAttributes(request)
		if err != nil {
			impl.logger.Errorw("error in adding new user attributes", "req", request, "error", err)
			return nil, errors.New("error occurred while updating user attributes")
		}
		return attributes, nil
	}

	// Parse existing JSON
	var existingData map[string]interface{}
	if userAttribute.Value != "" {
		err = json.Unmarshal([]byte(userAttribute.Value), &existingData)
		if err != nil {
			impl.logger.Errorw("error parsing existing json value", "value", userAttribute.Value, "error", err)
			return nil, errors.New("error occurred while updating user attributes")
		}
	} else {
		existingData = make(map[string]interface{})
	}

	// Parse new JSON
	var newData map[string]interface{}
	if request.Value != "" {
		err = json.Unmarshal([]byte(request.Value), &newData)
		if err != nil {
			impl.logger.Errorw("error parsing request json value", "value", request.Value, "error", err)
			return nil, errors.New("error occurred while updating user attributes")
		}
	} else {
		newData = make(map[string]interface{})
	}

	// Check if there are any changes
	anyChanges := false

	// Merge the objects (patch style)
	for key, newValue := range newData {
		existingValue, exists := existingData[key]
		if !exists || !reflect.DeepEqual(existingValue, newValue) {
			existingData[key] = newValue
			anyChanges = true
		}
	}

	// If no changes, return the existing data
	if !anyChanges {
		impl.logger.Infow("no change detected, skipping update", "key", request.Key)
		return userAttribute, nil
	}

	// Convert back to JSON string
	mergedJson, err := json.Marshal(existingData)
	if err != nil {
		impl.logger.Errorw("error converting merged data to json", "data", existingData, "error", err)
		return nil, errors.New("error occurred while updating user attributes")
	}

	dao := &repository.UserAttributesDao{
		EmailId: request.EmailId,
		Key:     request.Key,
		Value:   string(mergedJson),
		UserId:  request.UserId,
	}

	err = impl.attributesRepository.UpdateDataValByKey(dao)
	if err != nil {
		impl.logger.Errorw("error in update attributes", "req", dao, "error", err)
		return nil, errors.New("error occurred while updating user attributes")
	}

	// Return the updated data
	result := &UserAttributesDto{
		EmailId: request.EmailId,
		Key:     request.Key,
		Value:   string(mergedJson),
		UserId:  request.UserId,
	}

	return result, nil
}

func (impl UserAttributesServiceImpl) GetUserAttribute(request *UserAttributesDto) (*UserAttributesDto, error) {

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
	resAttrDto := &UserAttributesDto{
		EmailId: request.EmailId,
		Key:     request.Key,
		Value:   modelValue,
	}
	return resAttrDto, nil
}
