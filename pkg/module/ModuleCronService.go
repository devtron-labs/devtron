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
	"encoding/json"
	"fmt"
	client "github.com/devtron-labs/devtron/api/helm-app"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	moduleDataStore "github.com/devtron-labs/devtron/pkg/module/store"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	"github.com/devtron-labs/devtron/util"
	cron2 "github.com/devtron-labs/devtron/util/cron"
	"github.com/go-pg/pg"
	"github.com/robfig/cron/v3"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"time"
)

type ModuleCronService interface {
	HandleModuleStatusIfNotInProgress(moduleName string)
}

type ModuleCronServiceImpl struct {
	logger                         *zap.SugaredLogger
	cron                           *cron.Cron
	moduleEnvConfig                *ModuleEnvConfig
	moduleRepository               moduleRepo.ModuleRepository
	serverEnvConfig                *serverEnvConfig.ServerEnvConfig
	helmAppService                 client.HelmAppService
	moduleServiceHelper            ModuleServiceHelper
	moduleResourceStatusRepository moduleRepo.ModuleResourceStatusRepository
	moduleDataStore                *moduleDataStore.ModuleDataStore
}

func NewModuleCronServiceImpl(logger *zap.SugaredLogger, moduleEnvConfig *ModuleEnvConfig, moduleRepository moduleRepo.ModuleRepository,
	serverEnvConfig *serverEnvConfig.ServerEnvConfig, helmAppService client.HelmAppService, moduleServiceHelper ModuleServiceHelper, moduleResourceStatusRepository moduleRepo.ModuleResourceStatusRepository,
	moduleDataStore *moduleDataStore.ModuleDataStore, cronLogger *cron2.CronLoggerImpl) (*ModuleCronServiceImpl, error) {

	moduleCronServiceImpl := &ModuleCronServiceImpl{
		logger:                         logger,
		moduleEnvConfig:                moduleEnvConfig,
		moduleRepository:               moduleRepository,
		serverEnvConfig:                serverEnvConfig,
		helmAppService:                 helmAppService,
		moduleServiceHelper:            moduleServiceHelper,
		moduleResourceStatusRepository: moduleResourceStatusRepository,
		moduleDataStore:                moduleDataStore,
	}

	// if devtron user type is OSS_HELM then only cron to update module status is useful
	if serverEnvConfig.DevtronInstallationType == serverBean.DevtronInstallationTypeOssHelm {
		// cron job to update module status
		// initialise cron
		cron := cron.New(
			cron.WithChain(cron.Recover(cronLogger)))
		cron.Start()

		// add function into cron
		_, err := cron.AddFunc(fmt.Sprintf("@every %dm", moduleEnvConfig.ModuleStatusHandlingCronDurationInMin), moduleCronServiceImpl.handleAllModuleStatusIfNotInProgress)
		if err != nil {
			fmt.Println("error in adding cron function into module cron service")
			return nil, err
		}

		moduleCronServiceImpl.cron = cron
	}

	return moduleCronServiceImpl, nil
}

func (impl *ModuleCronServiceImpl) HandleModuleStatusIfNotInProgress(moduleName string) {
	if impl.moduleDataStore.ModuleStatusCronInProgress {
		impl.logger.Warn("module status cron is already in progress, returning.")
		return
	}
	impl.moduleDataStore.ModuleStatusCronInProgress = true
	impl.handleModuleStatus(moduleName)
	impl.moduleDataStore.ModuleStatusCronInProgress = false
}

func (impl *ModuleCronServiceImpl) handleAllModuleStatusIfNotInProgress() {
	impl.HandleModuleStatusIfNotInProgress("")
}

// check modules from DB.
// if status is installing for 1 hour, mark it as timeout
// if status is installing and helm release is healthy then mark as installed
func (impl *ModuleCronServiceImpl) handleModuleStatus(moduleNameInput string) {
	impl.logger.Debug("starting module status check thread")
	defer impl.logger.Debug("stopped module status check thread")

	// fetch all modules from DB
	modules, err := impl.moduleRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("error occurred while fetching all the modules from DB", "err", err)
		return
	}

	// update status timeout if module status is installing for more than 1 hour
	for _, module := range modules {
		if module.Status != ModuleStatusInstalling {
			continue
		}
		if len(moduleNameInput) > 0 && module.Name != moduleNameInput {
			continue
		}
		if time.Now().After(module.UpdatedOn.Add(1 * time.Hour)) {
			// timeout case
			impl.updateModuleStatus(module, ModuleStatusTimeout)
		} else if !util.IsBaseStack() {
			// if module is cicd then insert as installed
			if module.Name == ModuleNameCicd {
				impl.updateModuleStatus(module, ModuleStatusInstalled)
			} else {
				resourceTreeFilter, err := impl.buildResourceTreeFilter(module.Name)
				if err != nil {
					continue
				}
				appIdentifier := client.AppIdentifier{
					ClusterId:   1,
					Namespace:   impl.serverEnvConfig.DevtronHelmReleaseNamespace,
					ReleaseName: impl.serverEnvConfig.DevtronHelmReleaseName,
				}
				appDetail, err := impl.helmAppService.GetApplicationDetailWithFilter(context.Background(), &appIdentifier, resourceTreeFilter)
				if err != nil {
					impl.logger.Errorw("Error occurred while fetching helm application detail to check if module is installed", "moduleName", module.Name, "err", err)
					continue
				} else if appDetail.ApplicationStatus == serverBean.AppHealthStatusHealthy {
					impl.updateModuleStatus(module, ModuleStatusInstalled)
				}

				// save module resources status
				err = impl.saveModuleResourcesStatus(module.Id, appDetail)
				if err != nil {
					continue
				}
			}
		}
	}

}

func (impl *ModuleCronServiceImpl) saveModuleResourcesStatus(moduleId int, appDetail *client.AppDetail) error {
	impl.logger.Infow("updating module resources status", "moduleId", moduleId)
	if appDetail == nil || appDetail.ResourceTreeResponse == nil {
		return nil
	}
	moduleResourcesStatus, err := impl.moduleResourceStatusRepository.FindAllActiveByModuleId(moduleId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("Error in getting module statues from DB", "moduleId", moduleId, "err", err)
		return err
	}

	// build new data to save
	var moduleResourcesStatusToSave []*moduleRepo.ModuleResourceStatus
	nodes := appDetail.ResourceTreeResponse.Nodes
	if nodes != nil {
		for _, node := range nodes {
			moduleResourceStatusToSave := &moduleRepo.ModuleResourceStatus{
				ModuleId:  moduleId,
				Group:     node.Group,
				Version:   node.Version,
				Kind:      node.Kind,
				Name:      node.Name,
				Active:    true,
				CreatedOn: time.Now(),
			}
			nodeHealth := node.Health
			if nodeHealth == nil || len(nodeHealth.Status) == 0 {
				continue
			}
			moduleResourceStatusToSave.HealthStatus = nodeHealth.Status
			moduleResourceStatusToSave.HealthMessage = nodeHealth.Message
			moduleResourcesStatusToSave = append(moduleResourcesStatusToSave, moduleResourceStatusToSave)
		}
	}

	// initiate tx
	dbConnection := impl.moduleResourceStatusRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in initiating db tx", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// mark inactive if length > 0
	if len(moduleResourcesStatus) > 0 {
		for _, moduleResourceStatus := range moduleResourcesStatus {
			moduleResourceStatus.Active = false
			moduleResourceStatus.UpdatedOn = time.Now()
			err = impl.moduleResourceStatusRepository.Update(moduleResourceStatus, tx)
			if err != nil {
				impl.logger.Errorw("error in updating module resources status in DB", "err", err)
				return err
			}
		}
	}

	// insert if length > 0
	if len(moduleResourcesStatusToSave) > 0 {
		err = impl.moduleResourceStatusRepository.Save(moduleResourcesStatusToSave, tx)
		if err != nil {
			impl.logger.Errorw("error in saving module resources status in DB", "err", err)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing db tx", "err", err)
		return err
	}
	return nil
}

func (impl *ModuleCronServiceImpl) buildResourceTreeFilter(moduleName string) (*client.ResourceTreeFilter, error) {
	moduleMetaData, err := impl.moduleServiceHelper.GetModuleMetadata(moduleName)
	if err != nil {
		impl.logger.Errorw("Error in getting module metadata", "moduleName", moduleName, "err", err)
		return nil, err
	}

	moduleMetaDataStr := string(moduleMetaData)
	resourceFilterIface := gjson.Get(moduleMetaDataStr, "result.resourceFilter").String()

	if len(resourceFilterIface) == 0 {
		return nil, nil
	}

	resourceFilterIfaceValue := ResourceFilter{}
	err = json.Unmarshal([]byte(resourceFilterIface), &resourceFilterIfaceValue)
	if err != nil {
		impl.logger.Errorw("Error while unmarshalling resourceFilterIface", "resourceFilterIface", resourceFilterIface, "err", err)
		return nil, err
	}

	var resourceTreeFilter *client.ResourceTreeFilter

	// handle global filter
	globalFilter := resourceFilterIfaceValue.GlobalFilter
	if globalFilter != nil {
		resourceTreeFilter = &client.ResourceTreeFilter{
			GlobalFilter: &client.ResourceIdentifier{
				Labels: globalFilter.Labels,
			},
		}
		return resourceTreeFilter, nil
	}

	// otherwise handle gvk level
	var resourceFilters []*client.ResourceFilter
	for _, gvkLevelFilters := range resourceFilterIfaceValue.GvkLevelFilters {
		gvk := gvkLevelFilters.Gvk
		resourceFilters = append(resourceFilters, &client.ResourceFilter{
			Gvk: &client.Gvk{
				Group:   gvk.Group,
				Version: gvk.Version,
				Kind:    gvk.Kind,
			},
			ResourceIdentifier: &client.ResourceIdentifier{
				Labels: gvkLevelFilters.ResourceIdentifier.Labels,
			},
		})
	}
	resourceTreeFilter = &client.ResourceTreeFilter{
		ResourceFilters: resourceFilters,
	}
	return resourceTreeFilter, nil
}

func (impl *ModuleCronServiceImpl) updateModuleStatus(module moduleRepo.Module, status ModuleStatus) {
	impl.logger.Debugw("updating module status", "name", module.Name, "status", status)
	module.Status = status
	module.UpdatedOn = time.Now()
	err := impl.moduleRepository.Update(&module)
	if err != nil {
		impl.logger.Errorw("error in updating module status", "name", module.Name, "status", status, "err", err)
	}
}
