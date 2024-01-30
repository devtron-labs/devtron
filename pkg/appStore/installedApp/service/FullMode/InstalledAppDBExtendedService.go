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
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
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
	pubSubClient     *pubsub.PubSubClientServiceImpl
}

func NewInstalledAppDBExtendedServiceImpl(logger *zap.SugaredLogger,
	installedAppRepository repository2.InstalledAppRepository,
	appRepository app.AppRepository,
	userService user.UserService,
	installedAppRepositoryHistory repository2.InstalledAppVersionHistoryRepository,
	appStatusService appStatus.AppStatusService,
	pubSubClient *pubsub.PubSubClientServiceImpl) (*InstalledAppDBExtendedServiceImpl, error) {
	impl := &InstalledAppDBExtendedServiceImpl{
		InstalledAppDBServiceImpl: &EAMode.InstalledAppDBServiceImpl{
			Logger:                        logger,
			InstalledAppRepository:        installedAppRepository,
			AppRepository:                 appRepository,
			UserService:                   userService,
			InstalledAppRepositoryHistory: installedAppRepositoryHistory,
		},
		appStatusService: appStatusService,
		pubSubClient:     pubSubClient,
	}
	err := impl.subscribeHelmInstallStatus()
	if err != nil {
		return nil, err
	}
	return impl, nil
}

func (impl *InstalledAppDBExtendedServiceImpl) subscribeHelmInstallStatus() error {

	callback := func(msg *model.PubSubMsg) {

		helmInstallNatsMessage := &appStoreBean.HelmReleaseStatusConfig{}
		err := json.Unmarshal([]byte(msg.Data), helmInstallNatsMessage)
		if err != nil {
			impl.Logger.Errorw("error in unmarshalling helm install status nats message", "err", err)
			return
		}

		installedAppVersionHistory, err := impl.InstalledAppRepositoryHistory.GetInstalledAppVersionHistory(helmInstallNatsMessage.InstallAppVersionHistoryId)
		if err != nil {
			impl.Logger.Errorw("error in fetching installed app by installed app id in subscribe helm status callback", "err", err)
			return
		}
		if helmInstallNatsMessage.ErrorInInstallation {
			installedAppVersionHistory.Status = pipelineConfig.WorkflowFailed
		} else {
			installedAppVersionHistory.Status = pipelineConfig.WorkflowSucceeded
		}
		installedAppVersionHistory.HelmReleaseStatusConfig = msg.Data
		_, err = impl.InstalledAppRepositoryHistory.UpdateInstalledAppVersionHistory(installedAppVersionHistory, nil)
		if err != nil {
			impl.Logger.Errorw("error in updating helm release status data in installedAppVersionHistoryRepository", "err", err)
			return
		}
	}
	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		helmInstallNatsMessage := &appStoreBean.HelmReleaseStatusConfig{}
		err := json.Unmarshal([]byte(msg.Data), helmInstallNatsMessage)
		if err != nil {
			return "error in unmarshalling helm install status nats message", []interface{}{"err", err}
		}
		return "got nats msg for helm chart install status", []interface{}{"InstallAppVersionHistoryId", helmInstallNatsMessage.InstallAppVersionHistoryId, "ErrorInInstallation", helmInstallNatsMessage.ErrorInInstallation, "IsReleaseInstalled", helmInstallNatsMessage.IsReleaseInstalled}
	}

	err := impl.pubSubClient.Subscribe(pubsub.HELM_CHART_INSTALL_STATUS_TOPIC, callback, loggerFunc)
	if err != nil {
		impl.Logger.Error(err)
		return err
	}
	return nil
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
	if versionHistory.Status != (application2.Healthy) {
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
