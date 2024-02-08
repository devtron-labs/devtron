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

package gitOpsConfig

import (
	"context"
	"fmt"
	apiBean "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/pkg/chart/gitOpsConfig/bean"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/validation"
	bean3 "github.com/devtron-labs/devtron/pkg/deployment/gitOps/validation/bean"
	"github.com/devtron-labs/devtron/util/ChartsUtil"
	"net/http"

	//"github.com/devtron-labs/devtron/pkg/pipeline"

	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"time"

	"github.com/devtron-labs/devtron/internal/util"
	"go.uber.org/zap"
)

type DevtronAppGitOpConfigService interface {
	SaveAppLevelGitOpsConfiguration(appGitOpsRequest *bean.AppGitOpsConfigRequest, appName string, ctx context.Context) (err error)
	GetAppLevelGitOpsConfiguration(appId int) (*bean.AppGitOpsConfigResponse, error)
}

type DevtronAppGitOpConfigServiceImpl struct {
	logger                   *zap.SugaredLogger
	chartRepository          chartRepoRepository.ChartRepository
	gitOpsConfigReadService  config.GitOpsConfigReadService
	gitOpsValidationService  validation.GitOpsValidationService
	argoClientWrapperService argocdServer.ArgoClientWrapperService
}

func NewDevtronAppGitOpConfigServiceImpl(logger *zap.SugaredLogger,
	chartRepository chartRepoRepository.ChartRepository,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	gitOpsValidationService validation.GitOpsValidationService,
	argoClientWrapperService argocdServer.ArgoClientWrapperService) *DevtronAppGitOpConfigServiceImpl {
	return &DevtronAppGitOpConfigServiceImpl{
		logger:                   logger,
		chartRepository:          chartRepository,
		gitOpsConfigReadService:  gitOpsConfigReadService,
		gitOpsValidationService:  gitOpsValidationService,
		argoClientWrapperService: argoClientWrapperService,
	}
}

func (impl *DevtronAppGitOpConfigServiceImpl) SaveAppLevelGitOpsConfiguration(appGitOpsRequest *bean.AppGitOpsConfigRequest, appName string, ctx context.Context) (err error) {
	gitOpsConfigurationStatus, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		impl.logger.Errorw("error in fetching active gitOps config", "err", err)
		return err
	}
	if !gitOpsConfigurationStatus.IsGitOpsConfigured {
		return fmt.Errorf("Gitops integration is not installed/configured. Please install/configure gitops.")
	}
	if !gitOpsConfigurationStatus.AllowCustomRepository {
		apiErr := &util.ApiError{
			HttpStatusCode:  http.StatusConflict,
			UserMessage:     "Invalid request! Please configure Gitops with 'Allow changing git repository for application'.",
			InternalMessage: "Invalid request! Custom repository is not valid, as the global configuration is updated",
		}
		return apiErr
	}

	if impl.isGitRepoUrlPresent(appGitOpsRequest.AppId) {
		return fmt.Errorf("Invalid request! GitOps repository is already configured.")
	}

	charts, err := impl.chartRepository.FindActiveChartsByAppId(appGitOpsRequest.AppId)
	if err != nil {
		impl.logger.Errorw("Error in fetching charts for app", "err", err, "appId", appGitOpsRequest.AppId)
		return err
	}

	validateCustomGitRepoURLRequest := bean3.ValidateCustomGitRepoURLRequest{
		GitRepoURL:     appGitOpsRequest.GitOpsRepoURL,
		UserId:         appGitOpsRequest.UserId,
		AppName:        appName,
		GitOpsProvider: gitOpsConfigurationStatus.Provider,
	}
	repoUrl, _, validationErr := impl.gitOpsValidationService.ValidateCustomGitRepoURL(validateCustomGitRepoURLRequest)
	if validationErr != nil {
		apiErr := &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			UserMessage:     validationErr.Error(),
			InternalMessage: validationErr.Error(),
		}
		return apiErr
	}
	// ValidateCustomGitRepoURL returns sanitized repo url after validation
	appGitOpsRequest.GitOpsRepoURL = repoUrl
	chartGitAttr := &commonBean.ChartGitAttribute{
		RepoUrl: repoUrl,
	}
	err = impl.argoClientWrapperService.RegisterGitOpsRepoInArgo(ctx, chartGitAttr.RepoUrl, appGitOpsRequest.UserId)
	if err != nil {
		impl.logger.Errorw("error while register git repo in argo", "err", err)
		return err
	}

	for _, chart := range charts {
		chart.GitRepoUrl = repoUrl
		chart.UpdatedOn = time.Now()
		chart.UpdatedBy = appGitOpsRequest.UserId
		chart.IsCustomGitRepository = gitOpsConfigurationStatus.AllowCustomRepository && appGitOpsRequest.GitOpsRepoURL != apiBean.GIT_REPO_DEFAULT
		err = impl.chartRepository.Update(chart)
		if err != nil {
			impl.logger.Errorw("error in updating git repo url in charts while saving git repo url", "err", err, "appGitOpsRequest", appGitOpsRequest)
			return err
		}
	}
	return nil
}

func (impl *DevtronAppGitOpConfigServiceImpl) GetAppLevelGitOpsConfiguration(appId int) (*bean.AppGitOpsConfigResponse, error) {
	gitOpsConfigurationStatus, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		impl.logger.Errorw("error in fetching active gitOps config", "err", err)
		return nil, err
	}
	if !gitOpsConfigurationStatus.IsGitOpsConfigured {
		return nil, fmt.Errorf("Gitops integration is not installed/configured. Please install/configure gitops.")
	}
	if !gitOpsConfigurationStatus.AllowCustomRepository {
		apiErr := &util.ApiError{
			HttpStatusCode:  http.StatusConflict,
			UserMessage:     "Invalid request! Please configure Gitops with 'Allow changing git repository for application'.",
			InternalMessage: "Invalid request! Custom repository is not valid, as the global configuration is updated",
		}
		return nil, apiErr
	}
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching latest chart for app by appId", "err", err, "appId", appId)
		return nil, err
	}
	appGitOpsConfigResponse := &bean.AppGitOpsConfigResponse{
		IsEditable: true,
	}
	isGitOpsRepoConfigured := !ChartsUtil.IsGitOpsRepoNotConfigured(chart.GitRepoUrl)
	if isGitOpsRepoConfigured {
		appGitOpsConfigResponse.GitRepoURL = chart.GitRepoUrl
		appGitOpsConfigResponse.IsEditable = false
		return appGitOpsConfigResponse, nil
	}
	return appGitOpsConfigResponse, nil
}

func (impl *DevtronAppGitOpConfigServiceImpl) isGitRepoUrlPresent(appId int) bool {
	fetchedChart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error fetching git repo url from the latest chart")
		return false
	}
	if ChartsUtil.IsGitOpsRepoNotConfigured(fetchedChart.GitRepoUrl) {
		return false
	}
	return true
}
