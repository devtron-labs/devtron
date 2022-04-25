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
	util2 "github.com/devtron-labs/devtron/util"
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
}

func NewModuleCronServiceImpl(logger *zap.SugaredLogger, moduleEnvConfig *ModuleEnvConfig, moduleRepository ModuleRepository) (*ModuleCronServiceImpl, error) {

	moduleCronServiceImpl := &ModuleCronServiceImpl{
		logger:           logger,
		moduleRepository: moduleRepository,
	}

	// cron job to update status as timeout if installing state keeps in more than 1 hour
	// do this only if mode is hyperion (as for hyperion only module can be installed, for full, module is being treated as already installed)
	if util2.GetDevtronVersion().ServerMode == util2.SERVER_MODE_HYPERION {
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

	// fetch ciCd module from DB
	ciCdModule, err := impl.moduleRepository.FindOne(ModuleCiCdName)
	if err == nil {
		impl.logger.Errorw("error occurred while fetching ciCd module", "err", err)
		return
	}

	// update status timeout if module status is installing for more than 1 hour
	if ciCdModule.Status == ModuleStatusInstalling && time.Now().After(ciCdModule.UpdatedOn.Add(1*time.Hour)) {
		impl.logger.Debugw("updating module status as timeout", "name", ciCdModule.Name)
		ciCdModule.Status = ModuleStatusTimeout
		ciCdModule.UpdatedOn = time.Now()
		err = impl.moduleRepository.Update(ciCdModule)
		if err != nil {
			impl.logger.Errorw("error in updating module status to timeout", "name", ciCdModule.Name, "err", err)
		}
	}

}
