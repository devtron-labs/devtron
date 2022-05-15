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
	"fmt"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"time"
)

type ModuleCronService interface {
}

type ModuleCronServiceImpl struct {
	logger           *zap.SugaredLogger
	cron             *cron.Cron
	moduleEnvConfig  *ModuleEnvConfig
	moduleRepository ModuleRepository
	serverEnvConfig  *serverEnvConfig.ServerEnvConfig
}

func NewModuleCronServiceImpl(logger *zap.SugaredLogger, moduleEnvConfig *ModuleEnvConfig, moduleRepository ModuleRepository, serverEnvConfig *serverEnvConfig.ServerEnvConfig) (*ModuleCronServiceImpl, error) {

	moduleCronServiceImpl := &ModuleCronServiceImpl{
		logger:           logger,
		moduleEnvConfig:  moduleEnvConfig,
		moduleRepository: moduleRepository,
		serverEnvConfig:  serverEnvConfig,
	}

	// if devtron user type is OSS_HELM then only cron to update module timeout status is useful
	if serverEnvConfig.DevtronInstallationType == serverBean.DevtronInstallationTypeOssHelm {
		// cron job to update status as timeout if installing state keeps in more than 1 hour
		// initialise cron
		cron := cron.New(
			cron.WithChain())
		cron.Start()

		// add function into cron
		_, err := cron.AddFunc(fmt.Sprintf("@every %dm", moduleEnvConfig.ModuleTimeoutStatusHandlingCronDurationInMin), moduleCronServiceImpl.HandleModuleTimeoutStatus)
		if err != nil {
			fmt.Println("error in adding cron function into module cron service")
			return nil, err
		}

		moduleCronServiceImpl.cron = cron
	}

	return moduleCronServiceImpl, nil
}

// check modules from DB. if status if installing for 1 hour, mark it as timeout
func (impl *ModuleCronServiceImpl) HandleModuleTimeoutStatus() {
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
		if module.Status != ModuleStatusInstalling || !time.Now().After(module.UpdatedOn.Add(1*time.Hour)) {
			continue
		}

		impl.logger.Debugw("updating module status as timeout", "name", module.Name)
		module.Status = ModuleStatusTimeout
		module.UpdatedOn = time.Now()
		err = impl.moduleRepository.Update(&module)
		if err != nil {
			impl.logger.Errorw("error in updating module status to timeout", "name", module.Name, "err", err)
		}

	}

}
