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
	"errors"
	"github.com/devtron-labs/devtron/pkg/server"
	util2 "github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"time"
)

type ModuleService interface {
	GetModuleInfo(name string) (*ModuleInfoDto, error)
	HandleModuleAction(userId int32, moduleName string, moduleActionRequest *ModuleActionRequestDto) (*ActionResponse, error)
}

type ModuleServiceImpl struct {
	logger                         *zap.SugaredLogger
	serverEnvConfig                *server.ServerEnvConfig
	moduleRepository               ModuleRepository
	moduleActionAuditLogRepository ModuleActionAuditLogRepository
}

func NewModuleServiceImpl(logger *zap.SugaredLogger, serverEnvConfig *server.ServerEnvConfig, moduleRepository ModuleRepository,
	moduleActionAuditLogRepository ModuleActionAuditLogRepository) *ModuleServiceImpl {
	return &ModuleServiceImpl{
		logger:                         logger,
		serverEnvConfig:                serverEnvConfig,
		moduleRepository:               moduleRepository,
		moduleActionAuditLogRepository: moduleActionAuditLogRepository,
	}
}

func (impl ModuleServiceImpl) GetModuleInfo(name string) (*ModuleInfoDto, error) {
	impl.logger.Debugw("getting module info", "name", name)
	if name != ModuleCiCdName {
		return nil, errors.New("supplied module name is not supported yet")
	}

	moduleInfoDto := &ModuleInfoDto{
		Name: name,
	}

	// if server mod is full then treat it as cicd installed
	if util2.GetDevtronVersion().ServerMode == util2.SERVER_MODE_FULL {
		moduleInfoDto.Status = ModuleStatusInstalled
		return moduleInfoDto, nil
	}

	// assume it as EA_ONLY
	// fetch from DB
	module, err := impl.moduleRepository.FindOne(name)
	if err != nil {
		impl.logger.Errorw("error in getting module from DB ", "err", err)
		return nil, err
	}

	// if status is "unknown" then treat it as "notInstalled"
	if module.Status == ModuleStatusUnknown {
		moduleInfoDto.Status = ModuleStatusNotInstalled
		return moduleInfoDto, nil
	}

	// otherwise send DB status
	moduleInfoDto.Status = module.Status
	return moduleInfoDto, nil
}

func (impl ModuleServiceImpl) HandleModuleAction(userId int32, moduleName string, moduleActionRequest *ModuleActionRequestDto) (*ActionResponse, error) {
	impl.logger.Debugw("handling module action request", "moduleName", moduleName, "userId", userId, "payload", moduleActionRequest)

	// check if can update server
	if !impl.serverEnvConfig.CanServerUpdate {
		return nil, errors.New("module installation is not allowed")
	}
	if moduleName != ModuleCiCdName {
		return nil, errors.New("supplied module name is not supported yet")
	}

	// for full mode, this operation can not be performed, hence throw error
	if util2.GetDevtronVersion().ServerMode == util2.SERVER_MODE_FULL {
		return nil, errors.New("module installation is not allowed in full mode")
	}

	// get module by name
	module, err := impl.moduleRepository.FindOne(moduleName)
	if err != nil {
		impl.logger.Errorw("error in getting module ", "moduleName", moduleName, "err", err)
		return nil, err
	}

	// check if module is already installed or installing
	currentModuleStatus := module.Status
	if currentModuleStatus == ModuleStatusInstalling || currentModuleStatus == ModuleStatusInstalled {
		return nil, errors.New("module is already in installing/installed state")
	}

	// insert into audit table
	moduleActionAuditLog := &ModuleActionAuditLog{
		ModuleId:  module.Id,
		Version:   moduleActionRequest.Version,
		Action:    moduleActionRequest.Action,
		CreatedOn: time.Now(),
		CreatedBy: userId,
	}
	err = impl.moduleActionAuditLogRepository.Save(moduleActionAuditLog)
	if err != nil {
		impl.logger.Errorw("error in saving into audit log for module action ", "err", err)
		return nil, err
	}

	// since the request can only come for install, hence update the DB with installing status
	module.Status = ModuleStatusInstalling
	module.Version = moduleActionRequest.Version
	module.UpdatedOn = time.Now()
	err = impl.moduleRepository.Update(module)
	if err != nil {
		impl.logger.Errorw("error in updating module status as installing in DB ", "err", err)
		return nil, err
	}

	// TODO : call kubelink service
	// TODO : manish
	// on error - mark it as install failed
	if err != nil {
		module.Status = ModuleStatusInstallFailed
		impl.moduleRepository.Update(module)
		return nil, err
	}

	return &ActionResponse{
		Success: true,
	}, nil
}
