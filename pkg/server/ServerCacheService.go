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
	client "github.com/devtron-labs/devtron/api/helm-app"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	serverDataStore "github.com/devtron-labs/devtron/pkg/server/store"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"log"
)

type ServerCacheService interface {
}

type ServerCacheServiceImpl struct {
	logger          *zap.SugaredLogger
	serverEnvConfig *serverEnvConfig.ServerEnvConfig
	serverDataStore *serverDataStore.ServerDataStore
	helmAppService  client.HelmAppService
}

func NewServerCacheServiceImpl(logger *zap.SugaredLogger, serverEnvConfig *serverEnvConfig.ServerEnvConfig, serverDataStore *serverDataStore.ServerDataStore, helmAppService client.HelmAppService) *ServerCacheServiceImpl {
	impl := &ServerCacheServiceImpl{
		logger:          logger,
		serverEnvConfig: serverEnvConfig,
		serverDataStore: serverDataStore,
		helmAppService:  helmAppService,
	}

	// fetch current version from helm release
	appIdentifier := client.AppIdentifier{
		ClusterId:   1,
		Namespace:   impl.serverEnvConfig.DevtronHelmReleaseNamespace,
		ReleaseName: impl.serverEnvConfig.DevtronHelmReleaseName,
	}
	releaseInfo, err := impl.helmAppService.GetValuesYaml(context.Background(), &appIdentifier)
	if err != nil {
		log.Fatalln("got error in fetching devtron helm release values.", "error", err)
	}
	currentVersion := gjson.Get(releaseInfo.GetMergedValues(), impl.serverEnvConfig.DevtronVersionIdentifierInHelmValues).String()
	if len(currentVersion) == 0 {
		log.Fatalln("current devtron version found empty")
	}

	// store current version in-memory
	impl.serverDataStore.CurrentVersion = currentVersion

	return impl
}
