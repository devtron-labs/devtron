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
	serverDataStore "github.com/devtron-labs/devtron/pkg/server/store"
	"go.uber.org/zap"
	"time"
)

type ServerService interface {
	GetServerInfo() (*ServerInfoDto, error)
	HandleServerAction(userId int32, serverActionRequest *ServerActionRequestDto) (*ActionResponse, error)
}

type ServerServiceImpl struct {
	logger                         *zap.SugaredLogger
	serverActionAuditLogRepository ServerActionAuditLogRepository
	serverDataStore                *serverDataStore.ServerDataStore
	serverEnvConfig                *ServerEnvConfig
	helmAppService                 client.HelmAppService
}

func NewServerServiceImpl(logger *zap.SugaredLogger, serverActionAuditLogRepository ServerActionAuditLogRepository,
	serverDataStore *serverDataStore.ServerDataStore, serverEnvConfig *ServerEnvConfig, helmAppService client.HelmAppService) *ServerServiceImpl {
	return &ServerServiceImpl{
		logger:                         logger,
		serverActionAuditLogRepository: serverActionAuditLogRepository,
		serverDataStore:                serverDataStore,
		serverEnvConfig:                serverEnvConfig,
		helmAppService:                 helmAppService,
	}
}

func (impl ServerServiceImpl) GetServerInfo() (*ServerInfoDto, error) {
	impl.logger.Debug("getting server info")

	// fetch status of devtron helm app
	appIdentifier := client.AppIdentifier{
		ClusterId:   1,
		Namespace:   impl.serverEnvConfig.DevtronHelmReleaseNamespace,
		ReleaseName: impl.serverEnvConfig.DevtronHelmReleaseName,
	}
	appDetail, err := impl.helmAppService.GetApplicationDetail(context.Background(), &appIdentifier)
	if err != nil {
		impl.logger.Errorw("error in getting devtron helm app release status ", "err", err)
		return nil, err
	}

	serverInfoDto := ServerInfoDto{
		CurrentVersion: impl.serverDataStore.CurrentVersion,
		ReleaseName:    impl.serverEnvConfig.DevtronHelmReleaseName,
		Status:         appDetail.ApplicationStatus,
	}

	return &serverInfoDto, nil
}

func (impl ServerServiceImpl) HandleServerAction(userId int32, serverActionRequest *ServerActionRequestDto) (*ActionResponse, error) {
	impl.logger.Debugw("handling server action request", "userId", userId, "payload", serverActionRequest)

	// check if can update server
	if !impl.serverEnvConfig.CanServerUpdate {
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

	// TODO : call kubelink service
	// TODO : manish


	return &ActionResponse{
		Success: true,
	}, nil
}
