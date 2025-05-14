/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package dbMigration

import (
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	installedAppReader "github.com/devtron-labs/devtron/pkg/appStore/installedApp/read"
	"go.uber.org/zap"
	"time"
)

type DbMigration interface {
	FixMultipleAppsForInstalledApp(appNameUniqueIdentifier string) (*appRepository.App, error)
}

type DbMigrationServiceImpl struct {
	logger                  *zap.SugaredLogger
	appRepository           appRepository.AppRepository
	installedAppReadService installedAppReader.InstalledAppReadServiceEA
}

func NewDbMigrationServiceImpl(
	logger *zap.SugaredLogger, appRepository appRepository.AppRepository,
	installedAppReadService installedAppReader.InstalledAppReadServiceEA,
) *DbMigrationServiceImpl {
	impl := &DbMigrationServiceImpl{
		logger:                  logger,
		appRepository:           appRepository,
		installedAppReadService: installedAppReadService,
	}
	return impl
}

func (impl DbMigrationServiceImpl) FixMultipleAppsForInstalledApp(appNameUniqueIdentifier string) (*appRepository.App, error) {
	installedApp, err := impl.installedAppReadService.GetInstalledAppByAppName(appNameUniqueIdentifier)
	if err != nil {
		impl.logger.Errorw("error in fetching installed app by unique identifier", "appNameUniqueIdentifier", appNameUniqueIdentifier, "err", err)
		return nil, err
	}
	validAppId := installedApp.AppId
	allActiveApps, err := impl.appRepository.FindAllActiveByName(appNameUniqueIdentifier)
	if err != nil {
		impl.logger.Errorw("error in fetching all active apps by name", "appName", appNameUniqueIdentifier, "err", err)
		return nil, err
	}
	var validApp *appRepository.App
	for _, activeApp := range allActiveApps {
		if activeApp.Id != validAppId {
			impl.logger.Info("duplicate entries found for app, marking app inactive ", "appName", appNameUniqueIdentifier)
			activeApp.Active = false
			activeApp.UpdatedOn = time.Now()
			activeApp.UpdatedBy = 1
			err := impl.appRepository.Update(activeApp)
			if err != nil {
				impl.logger.Errorw("error in marking app inactive", "name", activeApp.AppName, "err", err)
				return nil, err
			}
		} else {
			validApp = activeApp
		}
	}
	return validApp, nil
}
