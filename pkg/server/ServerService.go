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

package server

import (
	"context"
	"errors"
	client "github.com/devtron-labs/devtron/api/helm-app"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	moduleUtil "github.com/devtron-labs/devtron/pkg/module/util"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	serverDataStore "github.com/devtron-labs/devtron/pkg/server/store"
	util2 "github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"time"
)

type ServerService interface {
	GetServerInfo() (*serverBean.ServerInfoDto, error)
	HandleServerAction(userId int32, serverActionRequest *serverBean.ServerActionRequestDto) (*serverBean.ActionResponse, error)
}

type ServerServiceImpl struct {
	logger                         *zap.SugaredLogger
	serverActionAuditLogRepository ServerActionAuditLogRepository
	serverDataStore                *serverDataStore.ServerDataStore
	serverEnvConfig                *serverEnvConfig.ServerEnvConfig
	helmAppService                 client.HelmAppService
	moduleRepository               moduleRepo.ModuleRepository
}

func NewServerServiceImpl(logger *zap.SugaredLogger, serverActionAuditLogRepository ServerActionAuditLogRepository,
	serverDataStore *serverDataStore.ServerDataStore, serverEnvConfig *serverEnvConfig.ServerEnvConfig, helmAppService client.HelmAppService, moduleRepository moduleRepo.ModuleRepository) *ServerServiceImpl {
	return &ServerServiceImpl{
		logger:                         logger,
		serverActionAuditLogRepository: serverActionAuditLogRepository,
		serverDataStore:                serverDataStore,
		serverEnvConfig:                serverEnvConfig,
		helmAppService:                 helmAppService,
		moduleRepository:               moduleRepository,
	}
}

func (impl ServerServiceImpl) GetServerInfo() (*serverBean.ServerInfoDto, error) {
	impl.logger.Debug("getting server info")

	serverInfoDto := &serverBean.ServerInfoDto{
		CurrentVersion:   impl.serverDataStore.CurrentVersion,
		ReleaseName:      impl.serverEnvConfig.DevtronHelmReleaseName,
		Status:           serverBean.ServerStatusUnknown,
		InstallationType: impl.serverEnvConfig.DevtronInstallationType,
	}

	// if installation type is not OSS helm, then return (do not calculate server status)
	if serverInfoDto.InstallationType != serverBean.DevtronInstallationTypeOssHelm {
		return serverInfoDto, nil
	}

	// fetch status of devtron helm app
	devtronHelmAppIdentifier := impl.helmAppService.GetDevtronHelmAppIdentifier()
	devtronAppDetail, err := impl.helmAppService.GetApplicationDetail(context.Background(), devtronHelmAppIdentifier)
	if err != nil {
		impl.logger.Errorw("error in getting devtron helm app release status ", "err", err)
		return nil, err
	}

	helmReleaseStatus := devtronAppDetail.ReleaseStatus.Status
	var serverStatus string

	// for hyperion mode mode i.e. installer object not found - use mapping
	// for full mode  -
	// if installer object status is applied then use mapping
	// if empty or downloaded, then check timeout
	// else if deployed then upgrading
	// else use mapping
	if !impl.serverDataStore.InstallerCrdObjectExists {
		serverStatus = mapServerStatusFromHelmReleaseStatus(helmReleaseStatus)
	} else {
		if impl.serverDataStore.InstallerCrdObjectStatus == serverBean.InstallerCrdObjectStatusApplied {
			serverStatus = mapServerStatusFromHelmReleaseStatus(helmReleaseStatus)
		} else if time.Now().After(devtronAppDetail.GetLastDeployed().AsTime().Add(1 * time.Hour)) {
			serverStatus = serverBean.ServerStatusTimeout
		} else if helmReleaseStatus == serverBean.HelmReleaseStatusDeployed {
			serverStatus = serverBean.ServerStatusUpgrading
		} else {
			serverStatus = mapServerStatusFromHelmReleaseStatus(helmReleaseStatus)
		}
	}

	serverInfoDto.Status = serverStatus
	return serverInfoDto, nil
}

func (impl ServerServiceImpl) HandleServerAction(userId int32, serverActionRequest *serverBean.ServerActionRequestDto) (*serverBean.ActionResponse, error) {
	impl.logger.Debugw("handling server action request", "userId", userId, "payload", serverActionRequest)

	// check if can update server
	if impl.serverEnvConfig.DevtronInstallationType != serverBean.DevtronInstallationTypeOssHelm {
		return nil, errors.New("server up-gradation is not allowed")
	}

	// insert into audit table
	serverActionAuditLog := &ServerActionAuditLog{
		Action:    serverActionRequest.Action,
		Version:   serverActionRequest.Version,
		CreatedOn: time.Now(),
		CreatedBy: userId,
	}
	err := impl.serverActionAuditLogRepository.Save(serverActionAuditLog)
	if err != nil {
		impl.logger.Errorw("error in saving into audit log for server action ", "err", err)
		return nil, err
	}

	// HELM_OPERATION Starts
	devtronHelmAppIdentifier := impl.helmAppService.GetDevtronHelmAppIdentifier()
	chartRepository := &client.ChartRepository{
		Name: impl.serverEnvConfig.DevtronHelmRepoName,
		Url:  impl.serverEnvConfig.DevtronHelmRepoUrl,
	}

	extraValues := make(map[string]interface{})
	extraValues["installer.release"] = serverActionRequest.Version
	alreadyInstalledModuleNames, err := impl.moduleRepository.GetInstalledModuleNames()
	if err != nil {
		impl.logger.Errorw("error in getting modules with installed status ", "err", err)
		return nil, err
	}
	for _, alreadyInstalledModuleName := range alreadyInstalledModuleNames {
		alreadyInstalledModuleEnableKeys := moduleUtil.BuildAllModuleEnableKeys(alreadyInstalledModuleName)
		for _, alreadyInstalledModuleEnableKey := range alreadyInstalledModuleEnableKeys {
			extraValues[alreadyInstalledModuleEnableKey] = true
		}
	}
	extraValuesYamlUrl := util2.BuildDevtronBomUrl(impl.serverEnvConfig.DevtronBomUrl, serverActionRequest.Version)
	updateResponse, err := impl.helmAppService.UpdateApplicationWithChartInfoWithExtraValues(context.Background(), devtronHelmAppIdentifier, chartRepository, extraValues, extraValuesYamlUrl, true)
	if err != nil {
		impl.logger.Errorw("error in updating helm release ", "err", err)
		return nil, err
	}
	if !updateResponse.GetSuccess() {
		return nil, errors.New("success is false from helm")
	}
	// HELM_OPERATION Ends

	return &serverBean.ActionResponse{
		Success: true,
	}, nil
}

func mapServerStatusFromHelmReleaseStatus(helmReleaseStatus string) string {
	var serverStatus string
	switch helmReleaseStatus {
	case serverBean.HelmReleaseStatusDeployed:
		serverStatus = serverBean.ServerStatusHealthy
	case serverBean.HelmReleaseStatusFailed:
		serverStatus = serverBean.ServerStatusUpgradeFailed
	case serverBean.HelmReleaseStatusPendingUpgrade:
		serverStatus = serverBean.ServerStatusUpgrading
	default:
		serverStatus = serverBean.ServerStatusUnknown
	}
	return serverStatus
}
