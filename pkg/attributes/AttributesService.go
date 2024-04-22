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
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/attributes/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type AttributesService interface {
	AddAttributes(request *bean.AttributesDto) (*bean.AttributesDto, error)
	UpdateAttributes(request *bean.AttributesDto) (*bean.AttributesDto, error)
	GetById(id int) (*bean.AttributesDto, error)
	GetActiveList() ([]*bean.AttributesDto, error)
	GetByKey(key string) (*bean.AttributesDto, error)
	UpdateKeyValueByOne(key string) error
	AddDeploymentEnforcementConfig(request *bean.AttributesDto) (*bean.AttributesDto, error)
	// GetDeploymentEnforcementConfig : Retrieves the deployment config values from the attributes table
	GetDeploymentEnforcementConfig(environmentId int) (map[string]bool, error)
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

func (impl AttributesServiceImpl) AddAttributes(request *bean.AttributesDto) (*bean.AttributesDto, error) {
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

func (impl AttributesServiceImpl) UpdateAttributes(request *bean.AttributesDto) (*bean.AttributesDto, error) {
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

func (impl AttributesServiceImpl) GetById(id int) (*bean.AttributesDto, error) {
	model, err := impl.attributesRepository.FindById(id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching attributes", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return nil, nil
	}
	ssoLoginDto := &bean.AttributesDto{
		Id:     model.Id,
		Active: model.Active,
		Key:    model.Key,
		Value:  model.Value,
	}
	return ssoLoginDto, nil
}

func (impl AttributesServiceImpl) GetActiveList() ([]*bean.AttributesDto, error) {
	results := make([]*bean.AttributesDto, 0)
	models, err := impl.attributesRepository.FindActiveList()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching attributes", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return results, nil
	}
	for _, model := range models {
		dto := &bean.AttributesDto{
			Id:     model.Id,
			Active: model.Active,
			Key:    model.Key,
			Value:  model.Value,
		}
		results = append(results, dto)
	}
	return results, nil
}

func (impl AttributesServiceImpl) GetByKey(key string) (*bean.AttributesDto, error) {
	model, err := impl.attributesRepository.FindByKey(key)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching attributes", "error", err, "key", key)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return nil, nil
	}
	dto := &bean.AttributesDto{
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

func (impl AttributesServiceImpl) AddDeploymentEnforcementConfig(request *bean.AttributesDto) (*bean.AttributesDto, error) {
	newConfig := make(map[string]map[string]bool)
	attributesErr := json.Unmarshal([]byte(request.Value), &newConfig)
	if attributesErr != nil {
		impl.logger.Errorw("error in unmarshalling", "value", request.Value, "err", attributesErr)
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
		impl.logger.Errorw("error in initiating db transaction")
		return request, terr
	}
	// Rollback tx on error.
	defer tx.Rollback()

	model, err := impl.attributesRepository.FindByKey(bean.ENFORCE_DEPLOYMENT_TYPE_CONFIG)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching deploymentEnforcementConfig from db", "error", err, "key", request.Key)
		return request, err
	}
	if err == pg.ErrNoRows {
		model := &repository.Attributes{
			Key:   bean.ENFORCE_DEPLOYMENT_TYPE_CONFIG,
			Value: request.Value,
		}
		model.Active = true
		model.UpdatedOn = time.Now()
		model.UpdatedBy = request.UserId
		_, err = impl.attributesRepository.Save(model, tx)
		if err != nil {
			impl.logger.Errorw("error in saving attributes", "model", model, "err", err)
			return request, err
		}
	} else {

		oldConfig := make(map[string]map[string]bool)
		oldConfigString := model.Value
		//initialConfigString = `{ "1": {"argo_cd": true}}`
		err = json.Unmarshal([]byte(oldConfigString), &oldConfig)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling", "oldConfigString", oldConfigString, "err", attributesErr)
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
			impl.logger.Errorw("error in updating attributes", "model", model, "err", err)
			return request, err
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error while commit db transaction to db", "error", err)
		return request, err
	}
	return request, nil
}

func (impl AttributesServiceImpl) GetDeploymentEnforcementConfig(environmentId int) (map[string]bool, error) {
	var deploymentConfig map[string]map[string]bool
	var deploymentConfigEnv map[string]bool
	deploymentConfigValues, err := impl.attributesRepository.FindByKey(bean.ENFORCE_DEPLOYMENT_TYPE_CONFIG)
	if util.IsErrNoRows(err) {
		return deploymentConfigEnv, nil
	}
	//if empty config received(doesn't exist in table) which can't be parsed
	if deploymentConfigValues.Value != "" {
		if err := json.Unmarshal([]byte(deploymentConfigValues.Value), &deploymentConfig); err != nil {
			apiError := &util.ApiError{
				HttpStatusCode:  http.StatusInternalServerError,
				InternalMessage: err.Error(),
				UserMessage:     "Failed to fetch deployment config values from the attributes table",
			}
			return deploymentConfigEnv, apiError
		}
		deploymentConfigEnv, _ = deploymentConfig[fmt.Sprintf("%d", environmentId)]
	}
	return deploymentConfigEnv, nil
}
