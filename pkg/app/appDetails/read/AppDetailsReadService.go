/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package read

import (
	"context"
	"github.com/devtron-labs/devtron/api/bean/AppView"
	"github.com/devtron-labs/devtron/api/bean/gitOps"
	internalRepo "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/deployment/common/read"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type AppDetailsReadService interface {
	FetchAppStageStatus(appId int, appType int) ([]AppView.AppStageStatus, error)
	FetchAppDetail(ctx context.Context, appId int, envId int) (AppView.AppDetailContainer, error)
}

type AppDetailsReadServiceImpl struct {
	dbConnection                *pg.DB
	Logger                      *zap.SugaredLogger
	gitOpsConfigReadService     config.GitOpsConfigReadService
	deploymentConfigReadService read.DeploymentConfigReadService
	appWorkflowRepository       appWorkflow.AppWorkflowRepository
	appListingRepository        internalRepo.AppListingRepository
}

func NewAppDetailsReadServiceImpl(
	dbConnection *pg.DB,
	Logger *zap.SugaredLogger,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	deploymentConfigReadService read.DeploymentConfigReadService,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	appListingRepository internalRepo.AppListingRepository,
) *AppDetailsReadServiceImpl {
	return &AppDetailsReadServiceImpl{
		dbConnection:                dbConnection,
		Logger:                      Logger,
		gitOpsConfigReadService:     gitOpsConfigReadService,
		deploymentConfigReadService: deploymentConfigReadService,
		appWorkflowRepository:       appWorkflowRepository,
		appListingRepository:        appListingRepository,
	}
}

func (impl *AppDetailsReadServiceImpl) FetchAppStageStatus(appId int, appType int) ([]AppView.AppStageStatus, error) {
	impl.Logger.Debugw("FetchAppStageStatus request received", "appId", appId, "appType", appType)
	var appStageStatus []AppView.AppStageStatus
	stages, err := impl.appListingRepository.FetchAppStage(appId, appType)
	if err != nil {
		impl.Logger.Errorw("error while fetching app stage", "appId", appId, "err", err)
		return appStageStatus, err
	}
	isCustomGitopsRepoUrl := false
	gitOpsConfigStatus, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		impl.Logger.Errorw("error while checking IsGitOpsConfigured", "err", err)
		return nil, err
	}
	if gitOpsConfigStatus.IsArgoCdInstalled && gitOpsConfigStatus.AllowCustomRepository {
		isCustomGitopsRepoUrl = true
	}

	deploymentConfigMin, err := impl.deploymentConfigReadService.GetDeploymentConfigMinForAppAndEnv(appId, 0)
	if err != nil {
		impl.Logger.Errorw("error while getting deploymentConfig", "appId", appId, "err", err)
		return appStageStatus, err
	}

	if deploymentConfigMin != nil {
		stages.DeploymentConfigRepoURL = deploymentConfigMin.GitRepoUrl
	}

	if (gitOps.IsGitOpsRepoNotConfigured(stages.ChartGitRepoUrl) &&
		gitOps.IsGitOpsRepoNotConfigured(stages.DeploymentConfigRepoURL)) &&
		stages.CiPipelineId == 0 {

		stages.ChartGitRepoUrl = ""
		stages.DeploymentConfigRepoURL = ""
	}
	appStageStatus = append(appStageStatus, impl.makeAppStageStatus(0, "APP", stages.AppId, true),
		impl.makeAppStageStatus(1, "MATERIAL", stages.GitMaterialExists, true),
		impl.makeAppStageStatus(2, "TEMPLATE", stages.CiTemplateId, true),
		impl.makeAppStageStatus(3, "CI_PIPELINE", stages.CiPipelineId, true),
		impl.makeAppStageStatus(4, "CHART", stages.ChartId, true),
		impl.makeAppStageStatus(5, "GITOPS_CONFIG", len(stages.ChartGitRepoUrl)+len(stages.DeploymentConfigRepoURL), isCustomGitopsRepoUrl),
		impl.makeAppStageStatus(6, "CD_PIPELINE", stages.PipelineId, true),
		impl.makeAppStageChartEnvConfigStatus(7, "CHART_ENV_CONFIG", stages.YamlStatus == 3 && stages.YamlReviewed),
	)
	return appStageStatus, nil
}

func (impl *AppDetailsReadServiceImpl) FetchAppDetail(ctx context.Context, appId int, envId int) (AppView.AppDetailContainer, error) {
	impl.Logger.Debugw("FetchAppDetail request received", "appId", appId, "envId", envId)
	var appDetailContainer AppView.AppDetailContainer
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "AppDetailsReadServiceImpl.FetchAppDetail")
	defer span.End()
	// Fetch deployment detail of cd pipeline latest triggered within env of any App.
	deploymentDetail, err := impl.deploymentDetailsByAppIdAndEnvId(newCtx, appId, envId)
	if err != nil {
		impl.Logger.Warn("unable to fetch deployment detail for app")
	}
	if deploymentDetail.PcoId > 0 {
		deploymentDetail.IsPipelineTriggered = true
	}
	appWfMapping, _ := impl.appWorkflowRepository.FindWFCDMappingByCDPipelineId(deploymentDetail.CdPipelineId)
	if appWfMapping.ParentType == appWorkflow.CDPIPELINE {
		parentEnvironmentName, _ := impl.appListingRepository.GetEnvironmentNameFromPipelineId(appWfMapping.ParentId)
		deploymentDetail.ParentEnvironmentName = parentEnvironmentName
	}
	appDetailContainer.DeploymentDetailContainer = deploymentDetail
	return appDetailContainer, nil
}

// deploymentDetailsByAppIdAndEnvId It will return the deployment detail of any cd pipeline which is latest triggered for Environment of any App
func (impl *AppDetailsReadServiceImpl) deploymentDetailsByAppIdAndEnvId(ctx context.Context, appId int, envId int) (AppView.DeploymentDetailContainer, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "AppDetailsReadServiceImpl.deploymentDetailsByAppIdAndEnvId")
	defer span.End()
	deploymentDetail, err := impl.appListingRepository.GetDeploymentDetailsByAppIdAndEnvId(appId, envId)
	if err != nil {
		impl.Logger.Errorw("error in getting deployment details by appId and envId", "appId", appId, "envId", envId, "err", err)
		return deploymentDetail, err
	}
	deploymentDetail.EnvironmentId = envId
	deploymentConfigMin, err := impl.deploymentConfigReadService.GetDeploymentConfigMinForAppAndEnv(appId, envId)
	if err != nil {
		impl.Logger.Errorw("error in getting deployment config by appId and envId", "appId", appId, "envId", envId, "err", err)
		return deploymentDetail, err
	}
	deploymentDetail.DeploymentAppType = deploymentConfigMin.DeploymentAppType
	deploymentDetail.ReleaseMode = deploymentConfigMin.ReleaseMode
	return deploymentDetail, nil
}

func (impl *AppDetailsReadServiceImpl) makeAppStageChartEnvConfigStatus(stage int, stageName string, status bool) AppView.AppStageStatus {
	return AppView.AppStageStatus{Stage: stage, StageName: stageName, Status: status, Required: true}
}

func (impl *AppDetailsReadServiceImpl) makeAppStageStatus(stage int, stageName string, id int, isRequired bool) AppView.AppStageStatus {
	return AppView.AppStageStatus{
		Stage:     stage,
		StageName: stageName,
		Status: func() bool {
			if id > 0 {
				return true
			} else {
				return false
			}
		}(),
		Required: isRequired,
	}
}
