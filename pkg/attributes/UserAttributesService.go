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

package attributes

import (
	"errors"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type UserAttributesService interface {
	AddUserAttributes(request *UserAttributesDto) (*UserAttributesDto, error)
	UpdateUserAttributes(request *UserAttributesDto) (*UserAttributesDto, error)
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
