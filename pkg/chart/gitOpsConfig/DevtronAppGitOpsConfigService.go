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
	apiGitOpsBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/client/argocdServer"
	chartService "github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/chart/gitOpsConfig/bean"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/validation"
	bean3 "github.com/devtron-labs/devtron/pkg/deployment/gitOps/validation/bean"
	"net/http"
	"path/filepath"

	//"github.com/devtron-labs/devtron/pkg/pipeline"

	"github.com/devtron-labs/devtron/internal/util"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"go.uber.org/zap"
)

type DevtronAppGitOpConfigService interface {
	SaveAppLevelGitOpsConfiguration(appGitOpsRequest *bean.AppGitOpsConfigRequest, appName string, ctx context.Context) (err error)
	GetAppLevelGitOpsConfiguration(appId int) (*bean.AppGitOpsConfigResponse, error)
}

type DevtronAppGitOpConfigServiceImpl struct {
	logger                   *zap.SugaredLogger
	chartRepository          chartRepoRepository.ChartRepository
	chartService             chartService.ChartService
	gitOpsConfigReadService  config.GitOpsConfigReadService
	gitOpsValidationService  validation.GitOpsValidationService
	argoClientWrapperService argocdServer.ArgoClientWrapperService
}

func NewDevtronAppGitOpConfigServiceImpl(logger *zap.SugaredLogger,
	chartRepository chartRepoRepository.ChartRepository,
	chartService chartService.ChartService,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	gitOpsValidationService validation.GitOpsValidationService,
	argoClientWrapperService argocdServer.ArgoClientWrapperService) *DevtronAppGitOpConfigServiceImpl {
	return &DevtronAppGitOpConfigServiceImpl{
		logger:                   logger,
		chartRepository:          chartRepository,
		chartService:             chartService,
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
		apiErr := &util.ApiError{
			HttpStatusCode:  http.StatusPreconditionFailed,
			UserMessage:     "GitOps integration is not installed/configured. Please install/configure GitOps.",
			InternalMessage: "GitOps integration is not installed/configured. Please install/configure GitOps.",
		}
		return apiErr
	}
	if !gitOpsConfigurationStatus.AllowCustomRepository {
		apiErr := &util.ApiError{
			HttpStatusCode:  http.StatusConflict,
			UserMessage:     "Invalid request! Please configure GitOps with 'Allow changing git repository for application'.",
			InternalMessage: "Invalid request! Custom repository is not valid, as the global configuration is updated",
		}
		return apiErr
	}

	if impl.isGitRepoUrlPresent(appGitOpsRequest.AppId) {
		apiErr := &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			UserMessage:     "Invalid request! GitOps repository is already configured.",
			InternalMessage: "Invalid request! GitOps repository is already configured.",
		}
		return apiErr
	}

	appDeploymentTemplate, err := impl.chartService.FindLatestChartForAppByAppId(appGitOpsRequest.AppId)
	if util.IsErrNoRows(err) {
		impl.logger.Errorw("no base charts configured for app", "appId", appGitOpsRequest.AppId, "err", err)
		apiErr := &util.ApiError{
			HttpStatusCode:  http.StatusPreconditionFailed,
			UserMessage:     "Invalid request! Base deployment chart is not configured for the app",
			InternalMessage: "Invalid request! Base deployment chart is not configured for the app",
		}
		return apiErr
	} else if err != nil {
		impl.logger.Errorw("error in fetching latest chart for app by appId", "appId", appGitOpsRequest.AppId, "err", err)
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
		RepoUrl:       repoUrl,
		ChartLocation: filepath.Join(appDeploymentTemplate.RefChartTemplate, appDeploymentTemplate.LatestChartVersion),
	}
	err = impl.argoClientWrapperService.RegisterGitOpsRepoInArgoWithRetry(ctx, chartGitAttr.RepoUrl, appGitOpsRequest.UserId)
	if err != nil {
		impl.logger.Errorw("error while register git repo in argo", "err", err)
		return err
	}
	isCustomGitOpsRepo := gitOpsConfigurationStatus.AllowCustomRepository && appGitOpsRequest.GitOpsRepoURL != apiGitOpsBean.GIT_REPO_DEFAULT
	err = impl.chartService.ConfigureGitOpsRepoUrl(appGitOpsRequest.AppId, chartGitAttr.RepoUrl, chartGitAttr.ChartLocation, isCustomGitOpsRepo, appGitOpsRequest.UserId)
	if err != nil {
		impl.logger.Errorw("error in updating git repo url in charts", "err", err)
		return err
	}
	return nil
}

func (impl *DevtronAppGitOpConfigServiceImpl) GetAppLevelGitOpsConfiguration(appId int) (*bean.AppGitOpsConfigResponse, error) {
	gitOpsConfigurationStatus, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		impl.logger.Errorw("error in fetching active gitOps config", "err", err)
		return nil, err
	} else if !gitOpsConfigurationStatus.IsGitOpsConfigured {
		apiErr := &util.ApiError{
			HttpStatusCode:  http.StatusPreconditionFailed,
			UserMessage:     "GitOps integration is not installed/configured. Please install/configure GitOps.",
			InternalMessage: "GitOps integration is not installed/configured. Please install/configure GitOps.",
		}
		return nil, apiErr
	} else if !gitOpsConfigurationStatus.AllowCustomRepository {
		apiErr := &util.ApiError{
			HttpStatusCode:  http.StatusConflict,
			UserMessage:     "Invalid request! Please configure GitOps with 'Allow changing git repository for application'.",
			InternalMessage: "Invalid request! Custom repository is not valid, as the global configuration is updated",
		}
		return nil, apiErr
	}
	appDeploymentTemplate, err := impl.chartService.FindLatestChartForAppByAppId(appId)
	if util.IsErrNoRows(err) {
		impl.logger.Errorw("no base charts configured for app", "appId", appId, "err", err)
		apiErr := &util.ApiError{
			HttpStatusCode:  http.StatusPreconditionFailed,
			UserMessage:     "Invalid request! Base deployment chart is not configured for the app",
			InternalMessage: "Invalid request! Base deployment chart is not configured for the app",
		}
		return nil, apiErr
	} else if err != nil {
		impl.logger.Errorw("error in fetching latest chart for app by appId", "appId", appId, "err", err)
		return nil, err
	}
	appGitOpsConfigResponse := &bean.AppGitOpsConfigResponse{
		IsEditable: true,
	}
	isGitOpsRepoConfigured := !apiGitOpsBean.IsGitOpsRepoNotConfigured(appDeploymentTemplate.GitRepoUrl)
	if isGitOpsRepoConfigured {
		appGitOpsConfigResponse.GitRepoURL = appDeploymentTemplate.GitRepoUrl
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
	return !apiGitOpsBean.IsGitOpsRepoNotConfigured(fetchedChart.GitRepoUrl)
}
