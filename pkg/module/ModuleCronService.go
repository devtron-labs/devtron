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
	"fmt"
	client "github.com/devtron-labs/devtron/api/helm-app"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	"github.com/devtron-labs/devtron/util"
	"github.com/robfig/cron/v3"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"time"
)

type ModuleCronService interface {
}

type ModuleCronServiceImpl struct {
	logger           *zap.SugaredLogger
	cron             *cron.Cron
	moduleEnvConfig  *ModuleEnvConfig
	moduleRepository moduleRepo.ModuleRepository
	serverEnvConfig  *serverEnvConfig.ServerEnvConfig
	helmAppService   client.HelmAppService
	moduleService    ModuleService
}

func NewModuleCronServiceImpl(logger *zap.SugaredLogger, moduleEnvConfig *ModuleEnvConfig, moduleRepository moduleRepo.ModuleRepository,
	serverEnvConfig *serverEnvConfig.ServerEnvConfig, helmAppService client.HelmAppService, moduleService ModuleService) (*ModuleCronServiceImpl, error) {

	moduleCronServiceImpl := &ModuleCronServiceImpl{
		logger:           logger,
		moduleEnvConfig:  moduleEnvConfig,
		moduleRepository: moduleRepository,
		serverEnvConfig:  serverEnvConfig,
		helmAppService:   helmAppService,
		moduleService:    moduleService,
	}

	// if devtron user type is OSS_HELM then only cron to update module status is useful
	if serverEnvConfig.DevtronInstallationType == serverBean.DevtronInstallationTypeOssHelm {
		// cron job to update module status
		// initialise cron
		cron := cron.New(
			cron.WithChain())
		cron.Start()

		// add function into cron
		_, err := cron.AddFunc(fmt.Sprintf("@every %dm", moduleEnvConfig.ModuleStatusHandlingCronDurationInMin), moduleCronServiceImpl.HandleModuleStatus)
		if err != nil {
			fmt.Println("error in adding cron function into module cron service")
			return nil, err
		}

		moduleCronServiceImpl.cron = cron
	}

	return moduleCronServiceImpl, nil
}

// check modules from DB.
//if status is installing for 1 hour, mark it as timeout
// if status is installing and helm release is healthy then mark as installed
func (impl *ModuleCronServiceImpl) HandleModuleStatus() {
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
		if time.Now().After(module.UpdatedOn.Add(1 * time.Hour)) {
			// timeout case
			impl.updateModuleStatus(module, ModuleStatusTimeout)
		} else if !util.IsBaseStack() {
			// check if helm release is healthy or not

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
			} else if appDetail.ApplicationStatus == serverBean.AppHealthStatusHealthy {
				impl.updateModuleStatus(module, ModuleStatusInstalled)
			}
		}
	}

}

func (impl *ModuleCronServiceImpl) buildResourceTreeFilter(moduleName string) (*client.ResourceTreeFilter, error) {
	moduleMetaData, err := impl.moduleService.GetModuleMetadata(moduleName)
	if err != nil {
		impl.logger.Errorw("Error in getting module metadata", "moduleName", moduleName, "err", err)
		return nil, err
	}

	moduleMetaDataStr := string(moduleMetaData)
	resourceIdentifiersIface := gjson.Get(moduleMetaDataStr, "result.resourceIdentifiers").Value()

	if resourceIdentifiersIface == nil {
		return nil, nil
	}
	resourceIdentifiersIfaceValues, ok := resourceIdentifiersIface.(map[schema.GroupVersionKind]ResourceFilter)
	if !ok {
		return nil, nil
	}

	var resourceTreeFilter *client.ResourceTreeFilter
	var resourceFilters []*client.ResourceFilter

	for gvk, resourceIdentifiersIfaceValue := range resourceIdentifiersIfaceValues {
		resourceFilters = append(resourceFilters, &client.ResourceFilter{
			Gvk: &client.Gvk{
				Group:   gvk.Group,
				Version: gvk.Version,
				Kind:    gvk.Kind,
			},
			Labels: resourceIdentifiersIfaceValue.Labels,
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
