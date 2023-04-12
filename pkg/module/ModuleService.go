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
	"github.com/caarlos0/env/v6"
	client "github.com/devtron-labs/devtron/api/helm-app"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	moduleUtil "github.com/devtron-labs/devtron/pkg/module/util"
	"github.com/devtron-labs/devtron/pkg/server"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	serverDataStore "github.com/devtron-labs/devtron/pkg/server/store"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"reflect"
	"time"
)

type ModuleService interface {
	GetModuleInfo(name string) (*ModuleInfoDto, error)
	GetModuleConfig(name string) (*ModuleConfigDto, error)
	HandleModuleAction(userId int32, moduleName string, moduleActionRequest *ModuleActionRequestDto) (*ActionResponse, error)
	GetAllModuleInfo() ([]ModuleInfoDto, error)
}

type ModuleServiceImpl struct {
	logger                         *zap.SugaredLogger
	serverEnvConfig                *serverEnvConfig.ServerEnvConfig
	moduleRepository               moduleRepo.ModuleRepository
	moduleActionAuditLogRepository ModuleActionAuditLogRepository
	helmAppService                 client.HelmAppService
	serverDataStore                *serverDataStore.ServerDataStore
	// no need to inject serverCacheService, moduleCacheService and cronService, but not generating in wire_gen (not triggering cache work in constructor) if not injecting. hence injecting
	// serverCacheService should be injected first as it changes serverEnvConfig in its constructor, which is used by moduleCacheService and moduleCronService
	serverCacheService             server.ServerCacheService
	moduleCacheService             ModuleCacheService
	moduleCronService              ModuleCronService
	moduleServiceHelper            ModuleServiceHelper
	moduleResourceStatusRepository moduleRepo.ModuleResourceStatusRepository
}

func NewModuleServiceImpl(logger *zap.SugaredLogger, serverEnvConfig *serverEnvConfig.ServerEnvConfig, moduleRepository moduleRepo.ModuleRepository,
	moduleActionAuditLogRepository ModuleActionAuditLogRepository, helmAppService client.HelmAppService, serverDataStore *serverDataStore.ServerDataStore, serverCacheService server.ServerCacheService, moduleCacheService ModuleCacheService, moduleCronService ModuleCronService,
	moduleServiceHelper ModuleServiceHelper, moduleResourceStatusRepository moduleRepo.ModuleResourceStatusRepository) *ModuleServiceImpl {
	return &ModuleServiceImpl{
		logger:                         logger,
		serverEnvConfig:                serverEnvConfig,
		moduleRepository:               moduleRepository,
		moduleActionAuditLogRepository: moduleActionAuditLogRepository,
		helmAppService:                 helmAppService,
		serverDataStore:                serverDataStore,
		serverCacheService:             serverCacheService,
		moduleCacheService:             moduleCacheService,
		moduleCronService:              moduleCronService,
		moduleServiceHelper:            moduleServiceHelper,
		moduleResourceStatusRepository: moduleResourceStatusRepository,
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
			status, err := impl.handleModuleNotFoundStatus(name)
			if err != nil {
				impl.logger.Errorw("error in handling module not found status ", "name", name, "err", err)
			}
			moduleInfoDto.Status = status
			return moduleInfoDto, err
		}
		// otherwise some error case
		impl.logger.Errorw("error in getting module from DB ", "name", name, "err", err)
		return nil, err
	}

	// now this is the case when data found in DB
	// if module is in installing state, then trigger module status check and override module model
	if module.Status == ModuleStatusInstalling {
		impl.moduleCronService.HandleModuleStatusIfNotInProgress(module.Name)
		// override module model
		module, err = impl.moduleRepository.FindOne(name)
		if err != nil {
			impl.logger.Errorw("error in getting module from DB ", "name", name, "err", err)
			return nil, err
		}
	}

	// send DB status
	moduleInfoDto.Status = module.Status

	// handle module resources status data
	moduleId := module.Id
	moduleResourcesStatusFromDb, err := impl.moduleResourceStatusRepository.FindAllActiveByModuleId(moduleId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting module resources status from DB ", "moduleId", moduleId, "moduleName", name, "err", err)
		return nil, err
	}
	if moduleResourcesStatusFromDb != nil {
		var moduleResourcesStatus []*ModuleResourceStatusDto
		for _, moduleResourceStatusFromDb := range moduleResourcesStatusFromDb {
			moduleResourcesStatus = append(moduleResourcesStatus, &ModuleResourceStatusDto{
				Group:         moduleResourceStatusFromDb.Group,
				Version:       moduleResourceStatusFromDb.Version,
				Kind:          moduleResourceStatusFromDb.Kind,
				Name:          moduleResourceStatusFromDb.Name,
				HealthStatus:  moduleResourceStatusFromDb.HealthStatus,
				HealthMessage: moduleResourceStatusFromDb.HealthMessage,
			})
		}
		moduleInfoDto.ModuleResourcesStatus = moduleResourcesStatus
	}

	return moduleInfoDto, nil
}

func (impl ModuleServiceImpl) GetModuleConfig(name string) (*ModuleConfigDto, error) {
	moduleConfig := &ModuleConfigDto{}
	if name == BlobStorage {
		blobStorageConfig := &BlobStorageConfig{}
		env.Parse(blobStorageConfig)
		moduleConfig.Enabled = blobStorageConfig.Enabled
	}
	return moduleConfig, nil
}

func (impl ModuleServiceImpl) handleModuleNotFoundStatus(moduleName string) (ModuleStatus, error) {
	// if entry is not found in database, then check if that module is legacy or not
	// if enterprise user -> if legacy -> then mark as installed in db and return as installed, if not legacy -> return as not installed
	// if non-enterprise user->  fetch helm release enable Key. if true -> then mark as installed in db and return as installed. if false ->
	//// (continuation of above line) if legacy -> check if cicd is installed with <= 0.5.3 from DB and moduleName != argo-cd -> then mark as installed in db and return as installed. otherwise return as not installed

	// central-api call
	moduleMetaData, err := impl.moduleServiceHelper.GetModuleMetadata(moduleName)
	if err != nil {
		impl.logger.Errorw("Error in getting module metadata", "moduleName", moduleName, "err", err)
		return ModuleStatusNotInstalled, err
	}
	moduleMetaDataStr := string(moduleMetaData)
	isLegacyModule := gjson.Get(moduleMetaDataStr, "result.isIncludedInLegacyFullPackage").Bool()
	baseMinVersionSupported := gjson.Get(moduleMetaDataStr, "result.baseMinVersionSupported").String()

	// for enterprise user
	if impl.serverEnvConfig.DevtronInstallationType == serverBean.DevtronInstallationTypeEnterprise {
		if isLegacyModule {
			return impl.saveModuleAsInstalled(moduleName)
		}
		return ModuleStatusNotInstalled, nil
	}

	// for non-enterprise user
	devtronHelmAppIdentifier := impl.helmAppService.GetDevtronHelmAppIdentifier()
	releaseInfo, err := impl.helmAppService.GetValuesYaml(context.Background(), devtronHelmAppIdentifier)
	if err != nil {
		impl.logger.Errorw("Error in getting values yaml for devtron operator helm release", "moduleName", moduleName, "err", err)
		return ModuleStatusNotInstalled, err
	}
	releaseValues := releaseInfo.MergedValues

	// if check non-cicd module status
	if moduleName != ModuleNameCicd {
		isEnabled := gjson.Get(releaseValues, moduleUtil.BuildModuleEnableKey(moduleName)).Bool()
		if isEnabled {
			return impl.saveModuleAsInstalled(moduleName)
		}
	} else if util2.IsBaseStack() {
		// check if cicd is in installing state
		// if devtron is installed with cicd module, then cicd module should be shown as installing
		installerModulesIface := gjson.Get(releaseValues, INSTALLER_MODULES_HELM_KEY).Value()
		if installerModulesIface != nil {
			installerModulesIfaceKind := reflect.TypeOf(installerModulesIface).Kind()
			if installerModulesIfaceKind == reflect.Slice {
				installerModules := installerModulesIface.([]interface{})
				for _, installerModule := range installerModules {
					if installerModule == moduleName {
						return impl.saveModule(moduleName, ModuleStatusInstalling)
					}
				}
			} else {
				impl.logger.Warnw("Invalid installerModulesIfaceKind expected slice", "installerModulesIfaceKind", installerModulesIfaceKind, "val", installerModulesIface)
			}
		}
	}

	// if module not enabled in helm for non enterprise-user
	if isLegacyModule && moduleName != ModuleNameCicd {
		for _, firstReleaseModuleName := range SupportedModuleNamesListFirstReleaseExcludingCicd {
			if moduleName != firstReleaseModuleName {
				cicdModule, err := impl.moduleRepository.FindOne(ModuleNameCicd)
				if err != nil {
					if err == pg.ErrNoRows {
						return ModuleStatusNotInstalled, nil
					} else {
						impl.logger.Errorw("Error in getting cicd module from DB", "err", err)
						return ModuleStatusNotInstalled, err
					}
				}
				cicdVersion := cicdModule.Version
				// if cicd was installed and any module/integration comes after that then mark that module installed only if cicd was installed before that module introduction
				if len(baseMinVersionSupported) > 0 && cicdVersion < baseMinVersionSupported {
					return impl.saveModuleAsInstalled(moduleName)
				}
				break
			}
		}
	}

	return ModuleStatusNotInstalled, nil

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
			module = &moduleRepo.Module{
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
	extraValues[INSTALLER_MODULES_HELM_KEY] = []interface{}{moduleName}
	alreadyInstalledModuleNames, err := impl.moduleRepository.GetInstalledModuleNames()
	if err != nil {
		impl.logger.Errorw("error in getting modules with installed status ", "err", err)
		return nil, err
	}
	moduleEnableKeys := moduleUtil.BuildAllModuleEnableKeys(moduleName)
	for _, moduleEnableKey := range moduleEnableKeys {
		extraValues[moduleEnableKey] = true
	}
	for _, alreadyInstalledModuleName := range alreadyInstalledModuleNames {
		if alreadyInstalledModuleName != moduleName {
			alreadyInstalledModuleEnableKeys := moduleUtil.BuildAllModuleEnableKeys(alreadyInstalledModuleName)
			for _, alreadyInstalledModuleEnableKey := range alreadyInstalledModuleEnableKeys {
				extraValues[alreadyInstalledModuleEnableKey] = true
			}
		}
	}
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

func (impl ModuleServiceImpl) saveModuleAsInstalled(moduleName string) (ModuleStatus, error) {
	return impl.saveModule(moduleName, ModuleStatusInstalled)
}

func (impl ModuleServiceImpl) saveModule(moduleName string, moduleStatus ModuleStatus) (ModuleStatus, error) {
	module := &moduleRepo.Module{
		Name:      moduleName,
		Version:   impl.serverDataStore.CurrentVersion,
		Status:    moduleStatus,
		UpdatedOn: time.Now(),
	}
	err := impl.moduleRepository.Save(module)
	if err != nil {
		impl.logger.Errorw("error in saving module status ", "moduleName", moduleName, "moduleStatus", moduleStatus, "err", err)
		return ModuleStatusNotInstalled, err
	}
	return moduleStatus, nil
}

func (impl ModuleServiceImpl) GetAllModuleInfo() ([]ModuleInfoDto, error) {
	// fetch from DB
	modules, err := impl.moduleRepository.FindAll()
	if err != nil {
		if err == pg.ErrNoRows {
			impl.logger.Errorw("no installed modules found ", "err", err)
			return nil, err
		}
		// otherwise some error case
		impl.logger.Errorw("error in getting modules from DB ", "err", err)
		return nil, err
	}
	var installedModules []ModuleInfoDto
	// now this is the case when data found in DB
	for _, module := range modules {
		moduleInfoDto := ModuleInfoDto{
			Name:   module.Name,
			Status: module.Status,
		}
		moduleId := module.Id
		moduleResourcesStatusFromDb, err := impl.moduleResourceStatusRepository.FindAllActiveByModuleId(moduleId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting module resources status from DB ", "moduleId", moduleId, "moduleName", module.Name, "err", err)
			return nil, err
		}
		if moduleResourcesStatusFromDb != nil {
			var moduleResourcesStatus []*ModuleResourceStatusDto
			for _, moduleResourceStatusFromDb := range moduleResourcesStatusFromDb {
				moduleResourcesStatus = append(moduleResourcesStatus, &ModuleResourceStatusDto{
					Group:         moduleResourceStatusFromDb.Group,
					Version:       moduleResourceStatusFromDb.Version,
					Kind:          moduleResourceStatusFromDb.Kind,
					Name:          moduleResourceStatusFromDb.Name,
					HealthStatus:  moduleResourceStatusFromDb.HealthStatus,
					HealthMessage: moduleResourceStatusFromDb.HealthMessage,
				})
			}
			moduleInfoDto.ModuleResourcesStatus = moduleResourcesStatus
		}
		installedModules = append(installedModules, moduleInfoDto)
	}

	return installedModules, nil
}
