/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commonService

import (
	"errors"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/module/bean"
	moduleRead "github.com/devtron-labs/devtron/pkg/module/read"
	moduleErr "github.com/devtron-labs/devtron/pkg/module/read/error"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"net/http"
)

type CommonBaseServiceImpl struct {
	logger             *zap.SugaredLogger
	globalEnvVariables *util.EnvironmentVariables
	moduleReadService  moduleRead.ModuleReadService
}

func NewCommonBaseServiceImpl(logger *zap.SugaredLogger, envVariables *util.EnvironmentVariables,
	moduleReadService moduleRead.ModuleReadService) *CommonBaseServiceImpl {
	return &CommonBaseServiceImpl{
		logger:             logger,
		globalEnvVariables: envVariables,
		moduleReadService:  moduleReadService,
	}
}

func (impl *CommonBaseServiceImpl) isGitOpsEnable() (bool, error) {
	if !impl.globalEnvVariables.DeploymentServiceTypeConfig.EnableMigrateArgoCdApplication {
		argoModule, err := impl.moduleReadService.GetModuleInfoByName(bean.ModuleNameArgoCd)
		if err != nil && !errors.Is(err, moduleErr.ModuleNotFoundError) {
			impl.logger.Errorw("error in getting argo module", "error", err)
			return false, err
		}
		return argoModule.IsInstalled(), nil
	}
	ciCdModule, err := impl.moduleReadService.GetModuleInfoByName(bean.ModuleNameCiCd)
	if err != nil && !errors.Is(err, moduleErr.ModuleNotFoundError) {
		impl.logger.Errorw("error in getting ci cd module", "error", err)
		return false, err
	}
	return ciCdModule.IsInstalled(), nil
}

func (impl *CommonBaseServiceImpl) EnvironmentVariableList() (*EnvironmentVariableList, error) {
	environmentVariableList := &EnvironmentVariableList{}
	isGitOpsEnabled, err := impl.isGitOpsEnable()
	if err != nil {
		impl.logger.Errorw("error in getting gitops enabled", "error", err)
		return environmentVariableList, err
	}
	environmentVariableList.IsGitOpsEnabled = isGitOpsEnabled
	return environmentVariableList, nil
}

func (impl *CommonBaseServiceImpl) GlobalChecklist() (*GlobalChecklist, error) {
	return nil, util2.DefaultApiError().WithHttpStatusCode(http.StatusNotFound).WithInternalMessage(util.NotSupportedErr).WithUserMessage(util.NotSupportedErr)
}

func (impl *CommonBaseServiceImpl) FetchLatestChartVersion(appId int, envId int) (string, error) {
	return "", util2.DefaultApiError().WithHttpStatusCode(http.StatusNotFound).WithInternalMessage(util.NotSupportedErr).WithUserMessage(util.NotSupportedErr)
}
