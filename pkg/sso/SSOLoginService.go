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
	"github.com/argoproj/argo-cd/util/session"
	"github.com/devtron-labs/devtron/api/bean"
	session2 "github.com/devtron-labs/devtron/client/argocdServer/session"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/ghodss/yaml"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type SSOLoginService interface {
	CreateSSOLogin(userInfo *bean.SSOLoginDto) (*bean.SSOLoginDto, error)
	UpdateSSOLogin(userInfo *bean.SSOLoginDto) (*bean.SSOLoginDto, error)
	GetById(id int32) (*bean.SSOLoginDto, error)
	GetAll() ([]*bean.SSOLoginDto, error)
	GetByName(name string) (*bean.SSOLoginDto, error)
}

type SSOLoginServiceImpl struct {
	sessionManager      *session.SessionManager
	userAuthRepository  repository.UserAuthRepository
	sessionClient       session2.ServiceClient
	logger              *zap.SugaredLogger
	userRepository      repository.UserRepository
	roleGroupRepository repository.RoleGroupRepository
	ssoLoginRepository  repository.SSOLoginRepository
	K8sUtil             *util.K8sUtil
	clusterService      cluster.ClusterService
	envService          cluster.EnvironmentService
	aCDAuthConfig       *user.ACDAuthConfig
}

func NewSSOLoginServiceImpl(userAuthRepository repository.UserAuthRepository, sessionManager *session.SessionManager,
	client session2.ServiceClient, logger *zap.SugaredLogger, userRepository repository.UserRepository,
	userGroupRepository repository.RoleGroupRepository, ssoLoginRepository repository.SSOLoginRepository,
	K8sUtil *util.K8sUtil, clusterService cluster.ClusterService, envService cluster.EnvironmentService,
	aCDAuthConfig *user.ACDAuthConfig) *SSOLoginServiceImpl {
	serviceImpl := &SSOLoginServiceImpl{
		userAuthRepository:  userAuthRepository,
		sessionManager:      sessionManager,
		sessionClient:       client,
		logger:              logger,
		userRepository:      userRepository,
		roleGroupRepository: userGroupRepository,
		ssoLoginRepository:  ssoLoginRepository,
		K8sUtil:             K8sUtil,
		clusterService:      clusterService,
		envService:          envService,
		aCDAuthConfig:       aCDAuthConfig,
	}
	return serviceImpl
}

func (impl SSOLoginServiceImpl) CreateSSOLogin(request *bean.SSOLoginDto) (*bean.SSOLoginDto, error) {
	dbConnection := impl.userRepository.GetConnection()
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
	model := &repository.SSOLoginModel{
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
	_, err = impl.updateArgocdConfigMapForDexConfig(request)
	if err != nil {
		impl.logger.Errorw("error in creating new sso login config", "error", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (impl SSOLoginServiceImpl) UpdateSSOLogin(request *bean.SSOLoginDto) (*bean.SSOLoginDto, error) {
	dbConnection := impl.userRepository.GetConnection()
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
	model.Label = request.Label
	model.Url = request.Url
	model.Config = string(configDataByte)
	model.Active = true
	model.UpdatedBy = request.UserId
	model.UpdatedOn = time.Now()
	_, err = impl.ssoLoginRepository.Update(model, tx)
	if err != nil {
		impl.logger.Errorw("error in creating new sso login config", "error", err)
		return nil, err
	}

	_, err = impl.updateArgocdConfigMapForDexConfig(request)
	if err != nil {
		impl.logger.Errorw("error in creating new sso login config", "error", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (impl SSOLoginServiceImpl) updateArgocdConfigMapForDexConfig(request *bean.SSOLoginDto) (bool, error) {

	//TODO- update argocd-cm
	flag := false
	clusterBean, err := impl.clusterService.FindOne(cluster.ClusterName)
	if err != nil {
		return flag, err
	}
	cfg, err := impl.clusterService.GetClusterConfig(clusterBean)
	if err != nil {
		return flag, err
	}

	client, err := impl.K8sUtil.GetClient(cfg)
	if err != nil {
		return flag, err
	}
	updateSuccess := false
	retryCount := 0
	for !updateSuccess && retryCount < 3 {
		retryCount = retryCount + 1

		cm, err := impl.K8sUtil.GetConfigMapFast(impl.aCDAuthConfig.ACDConfigMapNamespace, impl.aCDAuthConfig.ACDConfigMapName, client)
		if err != nil {
			return flag, err
		}
		updatedData, err := impl.updateSSODexConfigOnAcdConfigMap(request.Config)
		if err != nil {
			return flag, err
		}
		data := cm.Data
		data["dex.config"] = updatedData["dex.config"]
		data["url"] = request.Url
		cm.Data = data
		_, err = impl.K8sUtil.UpdateConfigMapFast(impl.aCDAuthConfig.ACDConfigMapNamespace, cm, client)
		if err != nil {
			impl.logger.Warnw("config map failed", "err", err)
			continue
		}
		if err == nil {
			impl.logger.Debugw("config map apply succeeded", "on retryCount", retryCount)
			updateSuccess = true
		}
	}
	if !updateSuccess {
		return flag, fmt.Errorf("resouce version not matched with config map attemped 3 times")
	}

	// TODO - END

	return true, nil
}

func (impl SSOLoginServiceImpl) updateSSODexConfigOnAcdConfigMap(config json.RawMessage) (map[string]string, error) {
	connectorConfig := map[string][]json.RawMessage{}
	var connectors []json.RawMessage
	connectors = append(connectors, config)
	connectorConfig["connectors"] = connectors
	connectorsJsonByte, err := json.Marshal(connectorConfig)
	if err != nil {
		panic(err)
	}
	connectorsYamlByte, err := yaml.JSONToYAML(connectorsJsonByte)
	if err != nil {
		panic(err)
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
