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

package sso

import (
	"encoding/json"
	"fmt"
	auth "github.com/devtron-labs/devtron/pkg/auth/authorisation/globalConfig"
	"time"

	"github.com/devtron-labs/common-lib-private/utils/k8s"
	"github.com/devtron-labs/devtron/pkg/auth/authentication"

	"github.com/devtron-labs/devtron/api/bean"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
)

type SSOLoginService interface {
	CreateSSOLogin(userInfo *bean.SSOLoginDto) (*bean.SSOLoginDto, error)
	UpdateSSOLogin(userInfo *bean.SSOLoginDto) (*bean.SSOLoginDto, error)
	GetById(id int32) (*bean.SSOLoginDto, error)
	GetAll() ([]*bean.SSOLoginDto, error)
	GetByName(name string) (*bean.SSOLoginDto, error)
}

type SSOLoginServiceImpl struct {
	logger                  *zap.SugaredLogger
	ssoLoginRepository      SSOLoginRepository
	K8sUtil                 *k8s.K8sUtilExtended
	devtronSecretConfig     *util2.DevtronSecretConfig
	userAuthOidcHelper      authentication.UserAuthOidcHelper
	globalAuthConfigService auth.GlobalAuthorisationConfigService
}

type Config struct {
	Id     string                 `json:"id"`
	Type   string                 `json:"type"`
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

const ClientID = "clientID"
const ClientSecret = "clientSecret"

func NewSSOLoginServiceImpl(
	logger *zap.SugaredLogger,
	ssoLoginRepository SSOLoginRepository,
	K8sUtil *k8s.K8sUtilExtended, devtronSecretConfig *util2.DevtronSecretConfig, userAuthOidcHelper authentication.UserAuthOidcHelper,
	globalAuthConfigService auth.GlobalAuthorisationConfigService) *SSOLoginServiceImpl {
	serviceImpl := &SSOLoginServiceImpl{
		logger:                  logger,
		ssoLoginRepository:      ssoLoginRepository,
		K8sUtil:                 K8sUtil,
		devtronSecretConfig:     devtronSecretConfig,
		userAuthOidcHelper:      userAuthOidcHelper,
		globalAuthConfigService: globalAuthConfigService,
	}
	return serviceImpl
}

func (impl SSOLoginServiceImpl) CreateSSOLogin(request *bean.SSOLoginDto) (*bean.SSOLoginDto, error) {
	dbConnection := impl.ssoLoginRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	configDataByte, err := json.Marshal(request.Config)
	if err != nil {
		return nil, err
	}

	existingModel, err := impl.ssoLoginRepository.GetActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in creating new sso login config", "error", err)
		return nil, err
	}
	if existingModel != nil && existingModel.Id > 0 {
		existingModel.Active = false
		existingModel.UpdatedOn = time.Now()
		existingModel.UpdatedBy = request.UserId
		_, err = impl.ssoLoginRepository.Update(existingModel, tx)
		if err != nil {
			impl.logger.Errorw("error in creating new sso login config", "error", err)
			return nil, err
		}
	}
	model := &SSOLoginModel{
		Name:   request.Name,
		Label:  request.Label,
		Config: string(configDataByte),
		Url:    request.Url,
	}
	model.Active = true
	model.CreatedBy = request.UserId
	model.UpdatedBy = request.UserId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	_, err = impl.ssoLoginRepository.Create(model, tx)
	if err != nil {
		impl.logger.Errorw("error in creating new sso login config", "error", err)
		return nil, err
	}
	request.Id = model.Id

	//  updating/creating globalAuthConfig here, doing this here to improve user experience, but altogeather they are different components
	if len(request.GlobalAuthConfigType) > 0 {
		_, err = impl.globalAuthConfigService.CreateOrUpdateGroupClaimsAuthConfig(tx, request.GlobalAuthConfigType, request.UserId)
		if err != nil {
			impl.logger.Errorw("error in CreateOrUpdateGroupClaimsAuthConfig", "err", err)
			return nil, err
		}
	}
	_, err = impl.updateDexConfig(request)
	if err != nil {
		impl.logger.Errorw("error in creating new sso login config", "error", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// update in memory data on sso add-update
	impl.userAuthOidcHelper.UpdateInMemoryDataOnSsoAddUpdate(request.Url)
	// Updating cache for globalAuthConfig
	impl.globalAuthConfigService.ReloadCache()
	return request, nil
}

func (impl SSOLoginServiceImpl) UpdateSSOLogin(request *bean.SSOLoginDto) (*bean.SSOLoginDto, error) {
	dbConnection := impl.ssoLoginRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	configDataByte, err := json.Marshal(request.Config)
	if err != nil {
		return nil, err
	}
	model, err := impl.ssoLoginRepository.GetById(request.Id)
	if err != nil {
		impl.logger.Errorw("error in update new sso login config", "error", err)
		return nil, err
	}

	existingModel, err := impl.ssoLoginRepository.GetActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in creating new sso login config", "error", err)
		return nil, err
	}
	if existingModel != nil && existingModel.Id > 0 {
		if existingModel.Id != model.Id {
			existingModel.Active = false
			existingModel.UpdatedOn = time.Now()
			existingModel.UpdatedBy = request.UserId
			_, err = impl.ssoLoginRepository.Update(existingModel, tx)
			if err != nil {
				impl.logger.Errorw("error in creating new sso login config", "error", err)
				return nil, err
			}
		}
	}
	configString := string(configDataByte)
	var configData Config
	err = json.Unmarshal([]byte(configString), &configData)
	if err != nil {
		impl.logger.Errorw("error while Unmarshalling configString", "error", err)
		return nil, err
	}
	var modelConfigData Config
	err = json.Unmarshal([]byte(model.Config), &modelConfigData)
	if err != nil {
		impl.logger.Errorw("error while Unmarshalling model's config", "error", err)
		return nil, err
	}
	updateSecretFromBase(&configData, &modelConfigData, ClientID)
	updateSecretFromBase(&configData, &modelConfigData, ClientSecret)
	newConfigString, err := json.Marshal(configData)
	if err != nil {
		impl.logger.Errorw("error while Marshaling configData", "error", err)
		return nil, err
	}
	updatedConfig := string(newConfigString)
	model.Label = request.Label
	model.Url = request.Url
	model.Config = updatedConfig
	model.Active = true
	model.UpdatedBy = request.UserId
	model.UpdatedOn = time.Now()
	_, err = impl.ssoLoginRepository.Update(model, tx)
	if err != nil {
		impl.logger.Errorw("error in creating new sso login config", "error", err)
		return nil, err
	}
	request.Config = newConfigString

	//  updating/creating globalAuthConfig here, doing this here to improve user experience, but altogeather they are different components
	if len(request.GlobalAuthConfigType) > 0 {
		_, err = impl.globalAuthConfigService.CreateOrUpdateGroupClaimsAuthConfig(tx, request.GlobalAuthConfigType, request.UserId)
		if err != nil {
			impl.logger.Errorw("error in CreateOrUpdateGroupClaimsAuthConfig", "err", err, "globalAuthConfigType", request.GlobalAuthConfigType)
			return nil, err
		}
	}
	_, err = impl.updateDexConfig(request)
	if err != nil {
		impl.logger.Errorw("error in creating new sso login config", "error", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// update in memory data on sso add-update
	impl.userAuthOidcHelper.UpdateInMemoryDataOnSsoAddUpdate(request.Url)
	// Updating cache for globalAuthConfig
	impl.globalAuthConfigService.ReloadCache()
	return request, nil
}

func (impl SSOLoginServiceImpl) updateDexConfig(request *bean.SSOLoginDto) (bool, error) {
	flag := false
	k8sClient, err := impl.K8sUtil.GetClientForInCluster()
	if err != nil {
		impl.logger.Errorw("exception in fetching client", "error", err)
		return flag, err
	}
	updateSuccess := false
	retryCount := 0
	for !updateSuccess && retryCount < 3 {
		retryCount = retryCount + 1
		secret, err := impl.K8sUtil.GetSecret(impl.devtronSecretConfig.DevtronDexSecretNamespace, impl.devtronSecretConfig.DevtronSecretName, k8sClient)
		if err != nil {
			impl.logger.Errorw("exception in fetching configmap", "error", err)
			return flag, err
		}
		updatedData, err := impl.updateSSODexConfigOnDevtronSecret(request.Config)
		if err != nil {
			impl.logger.Errorw("exception in update configmap sso config", "error", err)
			return flag, err
		}
		data := secret.Data
		if secret.Data == nil {
			data = make(map[string][]byte)
		}
		data["dex.config"] = []byte(updatedData["dex.config"])
		data["url"] = []byte(request.Url)
		secret.Data = data
		_, err = impl.K8sUtil.UpdateSecret(impl.devtronSecretConfig.DevtronDexSecretNamespace, secret, k8sClient)
		if err != nil {
			impl.logger.Warnw("config map update failed for sso config", "err", err)
			continue
		}
		if err == nil {
			impl.logger.Debugw("config map apply succeeded for sso config", "on retryCount", retryCount)
			updateSuccess = true
		}
	}
	if !updateSuccess {
		return flag, fmt.Errorf("resouce version not matched with config map attempted 3 times")
	}

	// TODO - END

	return true, nil
}

func (impl SSOLoginServiceImpl) updateSSODexConfigOnDevtronSecret(config json.RawMessage) (map[string]string, error) {
	connectorConfig := map[string][]json.RawMessage{}
	var connectors []json.RawMessage
	connectors = append(connectors, config)
	connectorConfig["connectors"] = connectors
	connectorsJsonByte, err := json.Marshal(connectorConfig)
	if err != nil {
		impl.logger.Errorw("exception in update configmap sso config", "error", err)
		return nil, err
	}
	connectorsYamlByte, err := yaml.JSONToYAML(connectorsJsonByte)
	if err != nil {
		impl.logger.Errorw("exception in update configmap sso config", "error", err)
		return nil, err
	}
	dexConfig := map[string]string{}
	dexConfig["dex.config"] = string(connectorsYamlByte)
	return dexConfig, nil
}

func (impl SSOLoginServiceImpl) GetById(id int32) (*bean.SSOLoginDto, error) {
	model, err := impl.ssoLoginRepository.GetById(id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in update new sso login config", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return nil, nil
	}
	var config json.RawMessage
	err = json.Unmarshal([]byte(model.Config), &config)
	if err != nil {
		impl.logger.Warnw("error while Unmarshal", "error", err)
	}

	ssoLoginDto := &bean.SSOLoginDto{
		Id:     model.Id,
		Name:   model.Name,
		Label:  model.Label,
		Active: model.Active,
		Config: config,
		Url:    model.Url,
	}
	return ssoLoginDto, nil
}

func (impl SSOLoginServiceImpl) GetAll() ([]*bean.SSOLoginDto, error) {
	ssoConfigs := make([]*bean.SSOLoginDto, 0)

	models, err := impl.ssoLoginRepository.GetAll()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in update new sso login config", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return ssoConfigs, nil
	}
	for _, model := range models {
		var config json.RawMessage
		err = json.Unmarshal([]byte(model.Config), &config)
		if err != nil {
			impl.logger.Warnw("error while Unmarshal", "error", err)
		}

		ssoLoginDto := &bean.SSOLoginDto{
			Id:     model.Id,
			Name:   model.Name,
			Label:  model.Label,
			Active: model.Active,
			Url:    model.Url,
		}
		ssoConfigs = append(ssoConfigs, ssoLoginDto)
	}
	return ssoConfigs, nil
}

func (impl SSOLoginServiceImpl) GetByName(name string) (*bean.SSOLoginDto, error) {
	model, err := impl.ssoLoginRepository.GetByName(name)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in update new sso login config", "error", err)
		return nil, err
	}

	if err == pg.ErrNoRows {
		return nil, nil
	}
	var configData Config
	err = json.Unmarshal([]byte(model.Config), &configData)
	if err != nil {
		impl.logger.Errorw("error while Unmarshalling model's config", "error", err)
		return nil, err
	}
	secureCredentialValue(&configData, ClientID)
	secureCredentialValue(&configData, ClientSecret)
	configString, err := json.Marshal(configData)
	if err != nil {
		impl.logger.Errorw("error while Marshaling configData", "error", err)
		return nil, err
	}
	var config json.RawMessage
	err = json.Unmarshal(configString, &config)
	if err != nil {
		impl.logger.Warnw("error while Unmarshal", "error", err)
	}

	ssoLoginDto := &bean.SSOLoginDto{
		Id:     model.Id,
		Name:   model.Name,
		Label:  model.Label,
		Active: model.Active,
		Config: config,
		Url:    model.Url,
	}
	return ssoLoginDto, nil
}

func updateSecretFromBase(configData *Config, baseConfigData *Config, key string) {
	if configData.Config[key] == "" && baseConfigData.Config[key] != nil {
		configData.Config[key] = baseConfigData.Config[key]
	}
}

func secureCredentialValue(configData *Config, credentialKey string) {
	if configData.Config[credentialKey] != nil {
		configData.Config[credentialKey] = ""
	}
}
