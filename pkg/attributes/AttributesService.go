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
	"github.com/argoproj/argo-cd/util/session"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type AttributesService interface {
	AddAttributes(request *AttributesDto) (*AttributesDto, error)
	UpdateAttributes(request *AttributesDto) (*AttributesDto, error)
	GetById(id int) (*AttributesDto, error)
	GetActiveList() ([]*AttributesDto, error)
	GetByKey(key string) (*AttributesDto, error)
}

const HostUrlKey string = "url"

type AttributesDto struct {
	Id     int    `json:"id"`
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
	Active bool   `json:"active"`
	UserId int32  `json:"-"`
}

type AttributesServiceImpl struct {
	logger               *zap.SugaredLogger
	sessionManager       *session.SessionManager
	attributesRepository repository.AttributesRepository
}

func NewAttributesServiceImpl(logger *zap.SugaredLogger, sessionManager *session.SessionManager,
	attributesRepository repository.AttributesRepository) *AttributesServiceImpl {
	serviceImpl := &AttributesServiceImpl{
		logger:               logger,
		sessionManager:       sessionManager,
		attributesRepository: attributesRepository,
	}
	return serviceImpl
}

func (impl AttributesServiceImpl) AddAttributes(request *AttributesDto) (*AttributesDto, error) {
	dbConnection := impl.attributesRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	existingModel, err := impl.attributesRepository.FindByKey(request.Key)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in update new attributes", "error", err)
		return nil, err
	}
	if existingModel != nil && existingModel.Id > 0 {
		existingModel.Active = false
		err = impl.attributesRepository.Update(existingModel, tx)
		if err != nil {
			impl.logger.Errorw("error in creating new attributes", "error", err)
			return nil, err
		}
	}

	model := &repository.Attributes{
		Key:   request.Key,
		Value: request.Value,
	}
	model.Active = true
	model.CreatedBy = request.UserId
	model.UpdatedBy = request.UserId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	_, err = impl.attributesRepository.Save(model, tx)
	if err != nil {
		impl.logger.Errorw("error in creating new attributes", "error", err)
		return nil, err
	}
	request.Id = model.Id
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (impl AttributesServiceImpl) UpdateAttributes(request *AttributesDto) (*AttributesDto, error) {
	dbConnection := impl.attributesRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	model, err := impl.attributesRepository.FindById(request.Id)
	if err != nil {
		impl.logger.Errorw("error in update new host url", "error", err)
		return nil, err
	}

	model.Key = request.Key
	model.Value = request.Value
	model.Active = true
	model.CreatedBy = request.UserId
	model.UpdatedBy = request.UserId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	err = impl.attributesRepository.Update(model, tx)
	if err != nil {
		impl.logger.Errorw("error in update new attributes", "error", err)
		return nil, err
	}
	request.Id = model.Id
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (impl AttributesServiceImpl) GetById(id int) (*AttributesDto, error) {
	model, err := impl.attributesRepository.FindById(id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching attributes", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return nil, nil
	}
	ssoLoginDto := &AttributesDto{
		Id:     model.Id,
		Active: model.Active,
		Key:    model.Key,
		Value:  model.Value,
	}
	return ssoLoginDto, nil
}

func (impl AttributesServiceImpl) GetActiveList() ([]*AttributesDto, error) {
	results := make([]*AttributesDto, 0)
	models, err := impl.attributesRepository.FindActiveList()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching attributes", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return results, nil
	}
	for _, model := range models {
		dto := &AttributesDto{
			Id:     model.Id,
			Active: model.Active,
			Key:    model.Key,
			Value:  model.Value,
		}
		results = append(results, dto)
	}
	return results, nil
}

func (impl AttributesServiceImpl) GetByKey(key string) (*AttributesDto, error) {
	model, err := impl.attributesRepository.FindByKey(key)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching attributes", "error", err, "key", key)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return nil, nil
	}
	dto := &AttributesDto{
		Id:     model.Id,
		Active: model.Active,
		Key:    model.Key,
		Value:  model.Value,
	}
	return dto, nil
}
