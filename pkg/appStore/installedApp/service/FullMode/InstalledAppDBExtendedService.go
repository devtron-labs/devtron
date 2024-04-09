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
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/devtron/pkg/appStatus"
)

type InstalledAppDBExtendedService interface {
	EAMode.InstalledAppDBService
	UpdateInstalledAppVersionStatus(application *v1alpha1.Application) (bool, error)
	IsGitOpsRepoAlreadyRegistered(repoUrl string) (bool, error)
}

type InstalledAppDBExtendedServiceImpl struct {
	*EAMode.InstalledAppDBServiceImpl
	appStatusService        appStatus.AppStatusService
	gitOpsConfigReadService config.GitOpsConfigReadService
}

func NewInstalledAppDBExtendedServiceImpl(
	installedAppDBServiceImpl *EAMode.InstalledAppDBServiceImpl,
	appStatusService appStatus.AppStatusService,
	gitOpsConfigReadService config.GitOpsConfigReadService) *InstalledAppDBExtendedServiceImpl {
	return &InstalledAppDBExtendedServiceImpl{
		InstalledAppDBServiceImpl: installedAppDBServiceImpl,
		appStatusService:          appStatusService,
		gitOpsConfigReadService:   gitOpsConfigReadService,
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

func (impl *InstalledAppDBExtendedServiceImpl) IsGitOpsRepoAlreadyRegistered(repoUrl string) (bool, error) {
	repoName := impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(repoUrl)
	installedAppModel, err := impl.InstalledAppRepository.GetInstalledAppByGitRepoUrl(repoName, repoUrl)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("error in fetching chart", "repoUrl", repoUrl, "err", err)
		return false, err
	}
	if util.IsErrNoRows(err) {
		return false, nil
	}
	impl.Logger.Warnw("repository is already in use for helm app", "repoUrl", repoUrl, "appId", installedAppModel.AppId)
	return true, nil
}
