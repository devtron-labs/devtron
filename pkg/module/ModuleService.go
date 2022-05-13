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

package module

import (
	"context"
	"errors"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/pkg/server"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ModuleService interface {
	GetModuleInfo(name string) (*ModuleInfoDto, error)
	HandleModuleAction(userId int32, moduleName string, moduleActionRequest *ModuleActionRequestDto) (*ActionResponse, error)
}

type ModuleServiceImpl struct {
	logger                         *zap.SugaredLogger
	serverEnvConfig                *serverEnvConfig.ServerEnvConfig
	moduleRepository               ModuleRepository
	moduleActionAuditLogRepository ModuleActionAuditLogRepository
	helmAppService                 client.HelmAppService
	// no need to inject serverCacheService, moduleCacheService and cronService, but not generating in wire_gen (not triggering cache work in constructor) if not injecting. hence injecting
	// serverCacheService should be injected first as it changes serverEnvConfig in its constructor, which is used by moduleCacheService and moduleCronService
	serverCacheService server.ServerCacheService
	moduleCacheService ModuleCacheService
	moduleCronService  ModuleCronService
}

func NewModuleServiceImpl(logger *zap.SugaredLogger, serverEnvConfig *serverEnvConfig.ServerEnvConfig, moduleRepository ModuleRepository,
	moduleActionAuditLogRepository ModuleActionAuditLogRepository, helmAppService client.HelmAppService, serverCacheService server.ServerCacheService, moduleCacheService ModuleCacheService, moduleCronService ModuleCronService) *ModuleServiceImpl {
	return &ModuleServiceImpl{
		logger:                         logger,
		serverEnvConfig:                serverEnvConfig,
		moduleRepository:               moduleRepository,
		moduleActionAuditLogRepository: moduleActionAuditLogRepository,
		helmAppService:                 helmAppService,
		serverCacheService:             serverCacheService,
		moduleCacheService:             moduleCacheService,
		moduleCronService:              moduleCronService,
	}
}

func (impl ModuleServiceImpl) GetModuleInfo(name string) (*ModuleInfoDto, error) {
	impl.logger.Debugw("getting module info", "name", name)

	moduleInfoDto := &ModuleInfoDto{
		Name: name,
	}

	// fetch from DB
	module, err := impl.moduleRepository.FindOne(name)
	if err != nil {
		if err == pg.ErrNoRows {
			// if entry is not found in database, then treat it as "notInstalled"
			moduleInfoDto.Status = ModuleStatusNotInstalled
			return moduleInfoDto, nil
		}
		// otherwise some error case
		impl.logger.Errorw("error in getting module from DB ", "err", err)
		return nil, err
	}

	// otherwise send DB status
	moduleInfoDto.Status = module.Status
	return moduleInfoDto, nil
}

func (impl ModuleServiceImpl) HandleModuleAction(userId int32, moduleName string, moduleActionRequest *ModuleActionRequestDto) (*ActionResponse, error) {
	impl.logger.Debugw("handling module action request", "moduleName", moduleName, "userId", userId, "payload", moduleActionRequest)

	// check if can update server
	if impl.serverEnvConfig.DevtronInstallationType != serverBean.DevtronInstallationTypeOssHelm {
		return nil, errors.New("module installation is not allowed")
	}

	// insert into audit table
	moduleActionAuditLog := &ModuleActionAuditLog{
		ModuleName: moduleName,
		Version:    moduleActionRequest.Version,
		Action:     moduleActionRequest.Action,
		CreatedOn:  time.Now(),
		CreatedBy:  userId,
	}
	err := impl.moduleActionAuditLogRepository.Save(moduleActionAuditLog)
	if err != nil {
		impl.logger.Errorw("error in saving into audit log for module action ", "err", err)
		return nil, err
	}

	// get module by name
	// if error, throw error
	// if module not found, then insert entry
	// if module found, then update entry
	module, err := impl.moduleRepository.FindOne(moduleName)
	moduleFound := true
	if err != nil {
		// either error or no data found
		if err == pg.ErrNoRows {
			// in case of entry not found, update variable
			moduleFound = false
			// initialise module to save in DB
			module = &Module{
				Name: moduleName,
			}
		} else {
			// otherwise some error case
			impl.logger.Errorw("error in getting module ", "moduleName", moduleName, "err", err)
			return nil, err
		}
	} else {
		// case of data found from DB
		// check if module is already installed or installing
		currentModuleStatus := module.Status
		if currentModuleStatus == ModuleStatusInstalling || currentModuleStatus == ModuleStatusInstalled {
			return nil, errors.New("module is already in installing/installed state")
		}

	}

	// since the request can only come for install, hence update the DB with installing status
	module.Status = ModuleStatusInstalling
	module.Version = moduleActionRequest.Version
	module.UpdatedOn = time.Now()
	if moduleFound {
		err = impl.moduleRepository.Update(module)
	} else {
		err = impl.moduleRepository.Save(module)
	}
	if err != nil {
		impl.logger.Errorw("error in saving/updating module ", "moduleName", moduleName, "err", err)
		return nil, err
	}

	// HELM_OPERATION Starts
	devtronHelmAppIdentifier := impl.helmAppService.GetDevtronHelmAppIdentifier()
	chartRepository := &client.ChartRepository{
		Name: impl.serverEnvConfig.DevtronHelmRepoName,
		Url:  impl.serverEnvConfig.DevtronHelmRepoUrl,
	}

	extraValues := make(map[string]interface{})
	extraValues["installer.release"] = moduleActionRequest.Version
	extraValues["installer.modules"] = []interface{}{moduleName}
	extraValuesYamlUrl := util2.BuildDevtronBomUrl(impl.serverEnvConfig.DevtronBomUrl, moduleActionRequest.Version)

	updateResponse, err := impl.helmAppService.UpdateApplicationWithChartInfoWithExtraValues(context.Background(), devtronHelmAppIdentifier, chartRepository, extraValues, extraValuesYamlUrl, true)
	if err != nil {
		impl.logger.Errorw("error in updating helm release ", "err", err)
		module.Status = ModuleStatusInstallFailed
		impl.moduleRepository.Update(module)
		return nil, err
	}
	if !updateResponse.GetSuccess() {
		module.Status = ModuleStatusInstallFailed
		impl.moduleRepository.Update(module)
		return nil, errors.New("success is false from helm")
	}
	// HELM_OPERATION Ends

	return &ActionResponse{
		Success: true,
	}, nil
}
