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

package FullMode

import (
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/appStatus"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"go.uber.org/zap"
)

type InstalledAppDBExtendedService interface {
	EAMode.InstalledAppDBService
	UpdateInstalledAppVersionStatus(application *v1alpha1.Application) (bool, error)
}

type InstalledAppDBExtendedServiceImpl struct {
	*EAMode.InstalledAppDBServiceImpl
	appStatusService appStatus.AppStatusService
}

func NewInstalledAppDBExtendedServiceImpl(logger *zap.SugaredLogger,
	installedAppRepository repository2.InstalledAppRepository,
	appRepository app.AppRepository,
	userService user.UserService,
	installedAppRepositoryHistory repository2.InstalledAppVersionHistoryRepository,
	appStatusService appStatus.AppStatusService) *InstalledAppDBExtendedServiceImpl {
	return &InstalledAppDBExtendedServiceImpl{
		InstalledAppDBServiceImpl: &EAMode.InstalledAppDBServiceImpl{
			Logger:                        logger,
			InstalledAppRepository:        installedAppRepository,
			AppRepository:                 appRepository,
			UserService:                   userService,
			InstalledAppRepositoryHistory: installedAppRepositoryHistory,
		},
		appStatusService: appStatusService,
	}
}

func (impl *InstalledAppDBExtendedServiceImpl) UpdateInstalledAppVersionStatus(application *v1alpha1.Application) (bool, error) {
	isHealthy := false
	dbConnection := impl.InstalledAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return isHealthy, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	gitHash := ""
	if application.Operation != nil && application.Operation.Sync != nil {
		gitHash = application.Operation.Sync.Revision
	} else if application.Status.OperationState != nil && application.Status.OperationState.Operation.Sync != nil {
		gitHash = application.Status.OperationState.Operation.Sync.Revision
	}
	versionHistory, err := impl.InstalledAppRepositoryHistory.GetLatestInstalledAppVersionHistoryByGitHash(gitHash)
	if err != nil {
		impl.Logger.Errorw("error while fetching installed version history", "error", err)
		return isHealthy, err
	}
	if versionHistory.Status != (argoApplication.Healthy) {
		versionHistory.Status = string(application.Status.Health.Status)
		versionHistory.UpdatedOn = time.Now()
		versionHistory.UpdatedBy = 1
		impl.InstalledAppRepositoryHistory.UpdateInstalledAppVersionHistory(versionHistory, tx)
	}
	err = tx.Commit()
	if err != nil {
		impl.Logger.Errorw("error while committing transaction to db", "error", err)
		return isHealthy, err
	}

	appId, envId, err := impl.InstalledAppRepositoryHistory.GetAppIdAndEnvIdWithInstalledAppVersionId(versionHistory.InstalledAppVersionId)
	if err == nil {
		err = impl.appStatusService.UpdateStatusWithAppIdEnvId(appId, envId, string(application.Status.Health.Status))
		if err != nil {
			impl.Logger.Errorw("error while updating app status in app_status table", "error", err, "appId", appId, "envId", envId)
		}
	}
	return true, nil
}
