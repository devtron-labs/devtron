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
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
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

	// if devtron user type is enterprise, then don't do anything to get current devtron version
	// if not enterprise, then fetch devtron helm release -
	// if devtron helm release is found, treat it as OSS Helm user otherwise OSS kubectl user
	if serverEnvConfig.DevtronInstallationType == serverBean.DevtronInstallationTypeEnterprise {
		return impl
	}

	// devtron helm release identifier
	appIdentifier := client.AppIdentifier{
		ClusterId:   1,
		Namespace:   impl.serverEnvConfig.DevtronHelmReleaseNamespace,
		ReleaseName: impl.serverEnvConfig.DevtronHelmReleaseName,
	}

	// check if the release is installed or not
	isDevtronHelmReleaseInstalled, err := impl.helmAppService.IsReleaseInstalled(context.Background(), &appIdentifier)
	if err != nil {
		log.Fatalln("not able to check if the devtron helm release exists or not.", "error", err)
	}

	// if not installed, treat it as OSS kubectl user
	// if installed, treat it as OSS helm user and fetch current version
	if isDevtronHelmReleaseInstalled {
		serverEnvConfig.DevtronInstallationType = serverBean.DevtronInstallationTypeOssHelm

		// fetch current version from helm release
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
	} else {
		serverEnvConfig.DevtronInstallationType = serverBean.DevtronInstallationTypeOssKubectl
	}

	return impl
}
