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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type AttributesService interface {
	AddAttributes(request *AttributesDto) (*AttributesDto, error)
	UpdateAttributes(request *AttributesDto) (*AttributesDto, error)
	GetById(id int) (*AttributesDto, error)
	GetActiveList() ([]*AttributesDto, error)
	GetByKey(key string) (*AttributesDto, error)
	UpdateKeyValueByOne(key string) error
	AddDeploymentEnforcementConfig(request *AttributesDto) (*AttributesDto, error)
}

const (
	HostUrlKey                     string = "url"
	API_SECRET_KEY                 string = "apiTokenSecret"
	NOTIFICATION_SECRET_KEY        string = "notificationTokenSecret"
	ENFORCE_DEPLOYMENT_TYPE_CONFIG string = "enforceDeploymentTypeConfig"
)

type AttributesDto struct {
	Id     int    `json:"id"`
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
	Active bool   `json:"active"`
	UserId int32  `json:"-"`
}

type AttributesServiceImpl struct {
	logger               *zap.SugaredLogger
	attributesRepository repository.AttributesRepository
}

func NewAttributesServiceImpl(logger *zap.SugaredLogger,
	attributesRepository repository.AttributesRepository) *AttributesServiceImpl {
	serviceImpl := &AttributesServiceImpl{
		logger:               logger,
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

func (impl AttributesServiceImpl) UpdateKeyValueByOne(key string) error {

	model, err := impl.attributesRepository.FindByKey(key)

	dbConnection := impl.attributesRepository.GetConnection()

	tx, err := dbConnection.Begin()

	defer tx.Rollback()

	model.Key = key

	if err != nil {
		return err
	}
	if model.Value == "" {
		model.Value = "1"
		model.Active = true
	} else {
		newValue, _ := strconv.Atoi(model.Value)
		model.Value = strconv.Itoa(newValue + 1)
	}

	err = impl.attributesRepository.Update(model, tx)

	if err == pg.ErrNoRows {
		_, err = impl.attributesRepository.Save(model, tx)
	}

	err = tx.Commit()

	return err

}

func (impl AttributesServiceImpl) AddDeploymentEnforcementConfig(request *AttributesDto) (*AttributesDto, error) {
	newConfig := make(map[string]map[string]bool)
	attributesErr := json.Unmarshal([]byte(request.Value), &newConfig)
	if attributesErr != nil {
		return request, attributesErr
	}
	for environmentId, envConfig := range newConfig {
		AllowedDeploymentAppTypes := 0
		for _, allowed := range envConfig {
			if allowed {
				AllowedDeploymentAppTypes++
			}
		}
		if AllowedDeploymentAppTypes == 0 && len(envConfig) > 0 {
			return request, errors.New(fmt.Sprintf("Received invalid config for environment with id %s, "+
				"at least one deployment app type should be allowed", environmentId))
		}
	}
	dbConnection := impl.attributesRepository.GetConnection()
	tx, terr := dbConnection.Begin()
	if terr != nil {
		return request, terr
	}
	// Rollback tx on error.
	defer tx.Rollback()

	model, err := impl.attributesRepository.FindByKey(ENFORCE_DEPLOYMENT_TYPE_CONFIG)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching deploymentEnforcementConfig from db", "error", err, "key", request.Key)
		return request, err
	}
	if err == pg.ErrNoRows {
		model := &repository.Attributes{
			Key:   ENFORCE_DEPLOYMENT_TYPE_CONFIG,
			Value: request.Value,
		}
		model.Active = true
		model.UpdatedOn = time.Now()
		model.UpdatedBy = request.UserId
		_, err = impl.attributesRepository.Save(model, tx)
		if err != nil {
			return request, err
		}
	} else {

		oldConfig := make(map[string]map[string]bool)
		oldConfigString := model.Value
		//initialConfigString = `{ "1": {"argo_cd": true}}`
		err = json.Unmarshal([]byte(oldConfigString), &oldConfig)
		if err != nil {
			return request, err
		}
		mergedConfig := oldConfig
		for k, v := range newConfig {
			mergedConfig[k] = v
		}
		value, err := json.Marshal(mergedConfig)
		if err != nil {
			return request, err
		}
		model.Value = string(value)
		model.UpdatedOn = time.Now()
		model.UpdatedBy = request.UserId
		model.Active = true
		err = impl.attributesRepository.Update(model, tx)
		if err != nil {
			return request, err
		}
	}
	tx.Commit()
	return request, nil
}
