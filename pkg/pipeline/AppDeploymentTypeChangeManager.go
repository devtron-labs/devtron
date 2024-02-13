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

package pipeline

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appStatus"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	app2 "github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/bean"
	chartService "github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

type AppDeploymentTypeChangeManager interface {
	//ChangeDeploymentType : takes in DeploymentAppTypeChangeRequest struct and
	// deletes all the cd pipelines for that deployment type in all apps that belongs to
	// that environment and updates the db with desired deployment app type
	ChangeDeploymentType(ctx context.Context, request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error)
	//ChangePipelineDeploymentType : takes in DeploymentAppTypeChangeRequest struct and
	// deletes all the cd pipelines for that deployment type in all apps that belongs to
	// that environment and updates the db with desired deployment app type
	ChangePipelineDeploymentType(ctx context.Context, request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error)
	//TriggerDeploymentAfterTypeChange :
	TriggerDeploymentAfterTypeChange(ctx context.Context, request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error)
	//DeleteDeploymentApps : takes in a list of pipelines and delete the applications
	DeleteDeploymentApps(ctx context.Context, pipelines []*pipelineConfig.Pipeline, userId int32) *bean.DeploymentAppTypeChangeResponse
	//DeleteDeploymentAppsForEnvironment : takes in environment id and current deployment app type
	// and deletes all the cd pipelines for that deployment type in all apps that belongs to
	// that environment.
	DeleteDeploymentAppsForEnvironment(ctx context.Context, environmentId int, currentDeploymentAppType bean.DeploymentType, exclusionList []int, includeApps []int, userId int32) (*bean.DeploymentAppTypeChangeResponse, error)
}

type AppDeploymentTypeChangeManagerImpl struct {
	logger              *zap.SugaredLogger
	pipelineRepository  pipelineConfig.PipelineRepository
	workflowDagExecutor WorkflowDagExecutor
	appService          app2.AppService
	appStatusRepository appStatus.AppStatusRepository
	helmAppService      service.HelmAppService
	application         application2.ServiceClient

	appArtifactManager      AppArtifactManager
	cdPipelineConfigService CdPipelineConfigService
	gitOpsConfigReadService config.GitOpsConfigReadService
	chartService            chartService.ChartService
}

func NewAppDeploymentTypeChangeManagerImpl(
	logger *zap.SugaredLogger,
	pipelineRepository pipelineConfig.PipelineRepository,
	workflowDagExecutor WorkflowDagExecutor,
	appService app2.AppService,
	appStatusRepository appStatus.AppStatusRepository,
	helmAppService service.HelmAppService,
	application application2.ServiceClient,
	appArtifactManager AppArtifactManager,
	cdPipelineConfigService CdPipelineConfigService,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	chartService chartService.ChartService) *AppDeploymentTypeChangeManagerImpl {
	return &AppDeploymentTypeChangeManagerImpl{
		logger:                  logger,
		pipelineRepository:      pipelineRepository,
		workflowDagExecutor:     workflowDagExecutor,
		appService:              appService,
		appStatusRepository:     appStatusRepository,
		helmAppService:          helmAppService,
		application:             application,
		appArtifactManager:      appArtifactManager,
		cdPipelineConfigService: cdPipelineConfigService,
		gitOpsConfigReadService: gitOpsConfigReadService,
		chartService:            chartService,
	}
}

func (impl *AppDeploymentTypeChangeManagerImpl) ChangeDeploymentType(ctx context.Context,
	request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error) {

	var response *bean.DeploymentAppTypeChangeResponse
	var deleteDeploymentType bean.DeploymentType
	var err error

	if request.DesiredDeploymentType == bean.ArgoCd {
		deleteDeploymentType = bean.Helm
	} else {
		deleteDeploymentType = bean.ArgoCd
	}

	// Force delete apps
	response, err = impl.DeleteDeploymentAppsForEnvironment(ctx,
		request.EnvId, deleteDeploymentType, request.ExcludeApps, request.IncludeApps, request.UserId)

	if err != nil {
		return nil, err
	}

	// Updating the env id and desired deployment app type received from request in the response
	response.EnvId = request.EnvId
	response.DesiredDeploymentType = request.DesiredDeploymentType
	response.TriggeredPipelines = make([]*bean.CdPipelineTrigger, 0)

	// Update the deployment app type to Helm and toggle deployment_app_created to false in db
	var cdPipelineIds []int
	for _, item := range response.SuccessfulPipelines {
		cdPipelineIds = append(cdPipelineIds, item.PipelineId)
	}

	// If nothing to update in db
	if len(cdPipelineIds) == 0 {
		return response, nil
	}

	// Update in db
	err = impl.pipelineRepository.UpdateCdPipelineDeploymentAppInFilter(string(request.DesiredDeploymentType),
		cdPipelineIds, request.UserId, false, true)

	if err != nil {
		impl.logger.Errorw("failed to update deployment app type in db",
			"pipeline ids", cdPipelineIds,
			"desired deployment type", request.DesiredDeploymentType,
			"err", err)

		return response, nil
	}

	if !request.AutoTriggerDeployment {
		return response, nil
	}

	// Bulk trigger all the successfully changed pipelines (async)
	bulkTriggerRequest := make([]*BulkTriggerRequest, 0)

	pipelineIds := make([]int, 0, len(response.SuccessfulPipelines))
	for _, item := range response.SuccessfulPipelines {
		pipelineIds = append(pipelineIds, item.PipelineId)
	}

	// Get all pipelines
	pipelines, err := impl.pipelineRepository.FindByIdsIn(pipelineIds)
	if err != nil {
		impl.logger.Errorw("failed to fetch pipeline details",
			"ids", pipelineIds,
			"err", err)

		return response, nil
	}

	for _, pipeline := range pipelines {

		artifactDetails, err := impl.appArtifactManager.RetrieveArtifactsByCDPipeline(pipeline, "DEPLOY")

		if err != nil {
			impl.logger.Errorw("failed to fetch artifact details for cd pipeline",
				"pipelineId", pipeline.Id,
				"appId", pipeline.AppId,
				"envId", pipeline.EnvironmentId,
				"err", err)

			return response, nil
		}

		if artifactDetails.LatestWfArtifactId == 0 || artifactDetails.LatestWfArtifactStatus == "" {
			continue
		}

		bulkTriggerRequest = append(bulkTriggerRequest, &BulkTriggerRequest{
			CiArtifactId: artifactDetails.LatestWfArtifactId,
			PipelineId:   pipeline.Id,
		})
		response.TriggeredPipelines = append(response.TriggeredPipelines, &bean.CdPipelineTrigger{
			CiArtifactId: artifactDetails.LatestWfArtifactId,
			PipelineId:   pipeline.Id,
		})
	}

	// pg panics if empty slice is passed as an argument
	if len(bulkTriggerRequest) == 0 {
		return response, nil
	}

	// Trigger
	_, err = impl.workflowDagExecutor.TriggerBulkDeploymentAsync(bulkTriggerRequest, request.UserId)

	if err != nil {
		impl.logger.Errorw("failed to bulk trigger cd pipelines with error: "+err.Error(),
			"err", err)
	}
	return response, nil
}

func (impl *AppDeploymentTypeChangeManagerImpl) ChangePipelineDeploymentType(ctx context.Context,
	request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error) {

	response := &bean.DeploymentAppTypeChangeResponse{
		EnvId:                 request.EnvId,
		DesiredDeploymentType: request.DesiredDeploymentType,
		TriggeredPipelines:    make([]*bean.CdPipelineTrigger, 0),
	}

	var deleteDeploymentType bean.DeploymentType

	if request.DesiredDeploymentType == bean.ArgoCd {
		deleteDeploymentType = bean.Helm
	} else {
		deleteDeploymentType = bean.ArgoCd
	}

	pipelines, err := impl.pipelineRepository.FindActiveByEnvIdAndDeploymentType(request.EnvId,
		string(deleteDeploymentType), request.ExcludeApps, request.IncludeApps)

	if err != nil {
		impl.logger.Errorw("Error fetching cd pipelines",
			"environmentId", request.EnvId,
			"currentDeploymentAppType", string(deleteDeploymentType),
			"err", err)
		return response, err
	}

	var pipelineIds []int
	for _, item := range pipelines {
		pipelineIds = append(pipelineIds, item.Id)
	}

	if len(pipelineIds) == 0 {
		return response, nil
	}

	err = impl.pipelineRepository.UpdateCdPipelineDeploymentAppInFilter(string(request.DesiredDeploymentType),
		pipelineIds, request.UserId, false, true)

	if err != nil {
		impl.logger.Errorw("failed to update deployment app type in db",
			"pipeline ids", pipelineIds,
			"desired deployment type", request.DesiredDeploymentType,
			"err", err)

		return response, nil
	}
	deleteResponse := impl.DeleteDeploymentApps(ctx, pipelines, request.UserId)

	response.SuccessfulPipelines = deleteResponse.SuccessfulPipelines
	response.FailedPipelines = deleteResponse.FailedPipelines

	var cdPipelineIds []int
	for _, item := range response.FailedPipelines {
		cdPipelineIds = append(cdPipelineIds, item.PipelineId)
	}

	if len(cdPipelineIds) == 0 {
		return response, nil
	}

	err = impl.pipelineRepository.UpdateCdPipelineDeploymentAppInFilter(string(deleteDeploymentType),
		cdPipelineIds, request.UserId, true, false)

	if err != nil {
		impl.logger.Errorw("failed to update deployment app type in db",
			"pipeline ids", cdPipelineIds,
			"desired deployment type", request.DesiredDeploymentType,
			"err", err)

		return response, nil
	}

	return response, nil
}

func (impl *AppDeploymentTypeChangeManagerImpl) TriggerDeploymentAfterTypeChange(ctx context.Context,
	request *bean.DeploymentAppTypeChangeRequest) (*bean.DeploymentAppTypeChangeResponse, error) {

	response := &bean.DeploymentAppTypeChangeResponse{
		EnvId:                 request.EnvId,
		DesiredDeploymentType: request.DesiredDeploymentType,
		TriggeredPipelines:    make([]*bean.CdPipelineTrigger, 0),
	}
	var err error

	cdPipelines, err := impl.pipelineRepository.FindActiveByEnvIdAndDeploymentType(request.EnvId,
		string(request.DesiredDeploymentType), request.ExcludeApps, request.IncludeApps)

	if err != nil {
		impl.logger.Errorw("Error fetching cd pipelines",
			"environmentId", request.EnvId,
			"desiredDeploymentAppType", string(request.DesiredDeploymentType),
			"err", err)
		return response, err
	}

	var cdPipelineIds []int
	for _, item := range cdPipelines {
		cdPipelineIds = append(cdPipelineIds, item.Id)
	}

	if len(cdPipelineIds) == 0 {
		return response, nil
	}

	deleteResponse := impl.FetchDeletedApp(ctx, cdPipelines)

	response.SuccessfulPipelines = deleteResponse.SuccessfulPipelines
	response.FailedPipelines = deleteResponse.FailedPipelines

	var successPipelines []int
	for _, item := range response.SuccessfulPipelines {
		successPipelines = append(successPipelines, item.PipelineId)
	}

	bulkTriggerRequest := make([]*BulkTriggerRequest, 0)

	pipelineIds := make([]int, 0, len(response.SuccessfulPipelines))
	for _, item := range response.SuccessfulPipelines {
		pipelineIds = append(pipelineIds, item.PipelineId)
	}

	pipelines, err := impl.pipelineRepository.FindByIdsIn(pipelineIds)
	if err != nil {
		impl.logger.Errorw("failed to fetch pipeline details",
			"ids", pipelineIds,
			"err", err)

		return response, nil
	}

	for _, pipeline := range pipelines {

		artifactDetails, err := impl.appArtifactManager.RetrieveArtifactsByCDPipeline(pipeline, "DEPLOY")

		if err != nil {
			impl.logger.Errorw("failed to fetch artifact details for cd pipeline",
				"pipelineId", pipeline.Id,
				"appId", pipeline.AppId,
				"envId", pipeline.EnvironmentId,
				"err", err)

			return response, nil
		}

		if artifactDetails.LatestWfArtifactId == 0 || artifactDetails.LatestWfArtifactStatus == "" {
			continue
		}

		bulkTriggerRequest = append(bulkTriggerRequest, &BulkTriggerRequest{
			CiArtifactId: artifactDetails.LatestWfArtifactId,
			PipelineId:   pipeline.Id,
		})
		response.TriggeredPipelines = append(response.TriggeredPipelines, &bean.CdPipelineTrigger{
			CiArtifactId: artifactDetails.LatestWfArtifactId,
			PipelineId:   pipeline.Id,
		})
	}

	if len(bulkTriggerRequest) == 0 {
		return response, nil
	}

	_, err = impl.workflowDagExecutor.TriggerBulkDeploymentAsync(bulkTriggerRequest, request.UserId)

	if err != nil {
		impl.logger.Errorw("failed to bulk trigger cd pipelines with error: "+err.Error(),
			"err", err)
	}

	err = impl.pipelineRepository.UpdateCdPipelineAfterDeployment(string(request.DesiredDeploymentType),
		successPipelines, request.UserId, false)

	if err != nil {
		impl.logger.Errorw("failed to update cd pipelines with error: : "+err.Error(),
			"err", err)
	}

	return response, nil
}

func (impl *AppDeploymentTypeChangeManagerImpl) DeleteDeploymentApps(ctx context.Context,
	pipelines []*pipelineConfig.Pipeline, userId int32) *bean.DeploymentAppTypeChangeResponse {

	successfulPipelines := make([]*bean.DeploymentChangeStatus, 0)
	failedPipelines := make([]*bean.DeploymentChangeStatus, 0)

	isGitOpsConfigured, gitOpsConfigErr := impl.gitOpsConfigReadService.IsGitOpsConfigured()

	// Iterate over all the pipelines in the environment for given deployment app type
	for _, pipeline := range pipelines {

		var isValid bool
		// check if pipeline info like app name and environment is empty or not
		if failedPipelines, isValid = impl.isPipelineInfoValid(pipeline, failedPipelines); !isValid {
			continue
		}

		var healthChkErr error
		// check health of the app if it is argocd deployment type
		if _, healthChkErr = impl.handleNotDeployedAppsIfArgoDeploymentType(pipeline, failedPipelines); healthChkErr != nil {

			// cannot delete unhealthy app
			continue
		}

		deploymentAppName := fmt.Sprintf("%s-%s", pipeline.App.AppName, pipeline.Environment.Name)
		var err error

		// delete request
		if pipeline.DeploymentAppType == bean.ArgoCd {
			err = impl.deleteArgoCdApp(ctx, pipeline, deploymentAppName, true)

		} else {

			// For converting from Helm to ArgoCD, GitOps should be configured
			if gitOpsConfigErr != nil || !isGitOpsConfigured {
				err = errors.New("GitOps not configured or unable to fetch GitOps configuration")

			} else {
				// Register app in ACD
				var AcdRegisterErr, RepoURLUpdateErr error
				gitopsRepoName, chartGitAttr, createGitRepoErr := impl.appService.CreateGitopsRepo(&app.App{Id: pipeline.AppId, AppName: pipeline.App.AppName}, userId)
				if createGitRepoErr != nil {
					impl.logger.Errorw("error increating git repo", "err", err)
				}
				if createGitRepoErr == nil {
					AcdRegisterErr = impl.cdPipelineConfigService.RegisterInACD(gitopsRepoName,
						chartGitAttr,
						userId,
						ctx)
					if AcdRegisterErr != nil {
						impl.logger.Errorw("error in registering acd app", "err", err)
					}
					if AcdRegisterErr == nil {
						RepoURLUpdateErr = impl.chartService.UpdateGitRepoUrlInCharts(pipeline.AppId, chartGitAttr.RepoUrl, chartGitAttr.ChartLocation, userId)
						if RepoURLUpdateErr != nil {
							impl.logger.Errorw("error in updating git repo url in charts", "err", err)
						}
					}
				}
				if createGitRepoErr != nil {
					err = createGitRepoErr
				} else if AcdRegisterErr != nil {
					err = AcdRegisterErr
				} else if RepoURLUpdateErr != nil {
					err = RepoURLUpdateErr
				}
			}
			if err != nil {
				impl.logger.Errorw("error registering app on ACD with error: "+err.Error(),
					"deploymentAppName", deploymentAppName,
					"envId", pipeline.EnvironmentId,
					"appId", pipeline.AppId,
					"err", err)

				// deletion failed, append to the list of failed pipelines
				failedPipelines = impl.handleFailedDeploymentAppChange(pipeline, failedPipelines,
					"failed to register app on ACD with error: "+err.Error())

				continue
			}
			err = impl.deleteHelmApp(ctx, pipeline)
		}

		if err != nil {
			impl.logger.Errorw("error deleting app on "+pipeline.DeploymentAppType,
				"deployment app name", deploymentAppName,
				"err", err)

			// deletion failed, append to the list of failed pipelines
			failedPipelines = impl.handleFailedDeploymentAppChange(pipeline, failedPipelines,
				"error deleting app with error: "+err.Error())

			continue
		}

		// deletion successful, append to the list of successful pipelines
		successfulPipelines = impl.appendToDeploymentChangeStatusList(
			successfulPipelines,
			pipeline,
			"",
			bean.INITIATED)
	}

	return &bean.DeploymentAppTypeChangeResponse{
		SuccessfulPipelines: successfulPipelines,
		FailedPipelines:     failedPipelines,
	}
}

func (impl *AppDeploymentTypeChangeManagerImpl) DeleteDeploymentAppsForEnvironment(ctx context.Context, environmentId int,
	currentDeploymentAppType bean.DeploymentType, exclusionList []int, includeApps []int, userId int32) (*bean.DeploymentAppTypeChangeResponse, error) {

	// fetch active pipelines from database for the given environment id and current deployment app type
	pipelines, err := impl.pipelineRepository.FindActiveByEnvIdAndDeploymentType(environmentId,
		string(currentDeploymentAppType), exclusionList, includeApps)

	if err != nil {
		impl.logger.Errorw("Error fetching cd pipelines",
			"environmentId", environmentId,
			"currentDeploymentAppType", currentDeploymentAppType,
			"err", err)

		return &bean.DeploymentAppTypeChangeResponse{
			EnvId:               environmentId,
			SuccessfulPipelines: []*bean.DeploymentChangeStatus{},
			FailedPipelines:     []*bean.DeploymentChangeStatus{},
		}, err
	}

	// Currently deleting apps only in argocd is supported
	return impl.DeleteDeploymentApps(ctx, pipelines, userId), nil
}

func (impl *AppDeploymentTypeChangeManagerImpl) isPipelineInfoValid(pipeline *pipelineConfig.Pipeline,
	failedPipelines []*bean.DeploymentChangeStatus) ([]*bean.DeploymentChangeStatus, bool) {

	if len(pipeline.App.AppName) == 0 || len(pipeline.Environment.Name) == 0 {
		impl.logger.Errorw("app name or environment name is not present",
			"pipeline id", pipeline.Id)

		failedPipelines = impl.handleFailedDeploymentAppChange(pipeline, failedPipelines,
			"could not fetch app name or environment name")

		return failedPipelines, false
	}
	return failedPipelines, true
}

func (impl *AppDeploymentTypeChangeManagerImpl) handleNotHealthyAppsIfArgoDeploymentType(pipeline *pipelineConfig.Pipeline,
	failedPipelines []*bean.DeploymentChangeStatus) ([]*bean.DeploymentChangeStatus, error) {

	if pipeline.DeploymentAppType == bean.ArgoCd {
		// check if app status is Healthy
		status, err := impl.appStatusRepository.Get(pipeline.AppId, pipeline.EnvironmentId)

		// case: missing status row in db
		if len(status.Status) == 0 {
			return failedPipelines, nil
		}

		// cannot delete the app from argocd if app status is Progressing
		if err != nil || status.Status == "Progressing" {

			healthCheckErr := errors.New("unable to fetch app status or app status is progressing")

			impl.logger.Errorw(healthCheckErr.Error(),
				"appId", pipeline.AppId,
				"environmentId", pipeline.EnvironmentId,
				"err", err)

			failedPipelines = impl.handleFailedDeploymentAppChange(pipeline, failedPipelines, healthCheckErr.Error())

			return failedPipelines, healthCheckErr
		}
		return failedPipelines, nil
	}
	return failedPipelines, nil
}

func (impl *AppDeploymentTypeChangeManagerImpl) handleNotDeployedAppsIfArgoDeploymentType(pipeline *pipelineConfig.Pipeline,
	failedPipelines []*bean.DeploymentChangeStatus) ([]*bean.DeploymentChangeStatus, error) {

	if pipeline.DeploymentAppType == string(bean.ArgoCd) {
		// check if app status is Healthy
		status, err := impl.appStatusRepository.Get(pipeline.AppId, pipeline.EnvironmentId)

		// case: missing status row in db
		if len(status.Status) == 0 {
			return failedPipelines, nil
		}

		// cannot delete the app from argocd if app status is Progressing
		if err != nil {

			healthCheckErr := errors.New("unable to fetch app status")

			impl.logger.Errorw(healthCheckErr.Error(),
				"appId", pipeline.AppId,
				"environmentId", pipeline.EnvironmentId,
				"err", err)

			failedPipelines = impl.handleFailedDeploymentAppChange(pipeline, failedPipelines, healthCheckErr.Error())

			return failedPipelines, healthCheckErr
		}
		return failedPipelines, nil
	}
	return failedPipelines, nil
}

func (impl *AppDeploymentTypeChangeManagerImpl) handleFailedDeploymentAppChange(pipeline *pipelineConfig.Pipeline,
	failedPipelines []*bean.DeploymentChangeStatus, err string) []*bean.DeploymentChangeStatus {

	return impl.appendToDeploymentChangeStatusList(
		failedPipelines,
		pipeline,
		err,
		bean.Failed)
}

func (impl *AppDeploymentTypeChangeManagerImpl) FetchDeletedApp(ctx context.Context,
	pipelines []*pipelineConfig.Pipeline) *bean.DeploymentAppTypeChangeResponse {

	successfulPipelines := make([]*bean.DeploymentChangeStatus, 0)
	failedPipelines := make([]*bean.DeploymentChangeStatus, 0)
	// Iterate over all the pipelines in the environment for given deployment app type
	for _, pipeline := range pipelines {

		deploymentAppName := fmt.Sprintf("%s-%s", pipeline.App.AppName, pipeline.Environment.Name)
		var err error
		if pipeline.DeploymentAppType == string(bean.ArgoCd) {
			appIdentifier := &service.AppIdentifier{
				ClusterId:   pipeline.Environment.ClusterId,
				ReleaseName: pipeline.DeploymentAppName,
				Namespace:   pipeline.Environment.Namespace,
			}
			_, err = impl.helmAppService.GetApplicationDetail(ctx, appIdentifier)
		} else {
			req := &application.ApplicationQuery{
				Name: &deploymentAppName,
			}
			_, err = impl.application.Get(ctx, req)
		}
		if err != nil {
			impl.logger.Errorw("error in getting application detail", "err", err, "deploymentAppName", deploymentAppName)
		}

		if err != nil && checkAppReleaseNotExist(err) {
			successfulPipelines = impl.appendToDeploymentChangeStatusList(
				successfulPipelines,
				pipeline,
				"",
				bean.Success)
		} else {
			failedPipelines = impl.appendToDeploymentChangeStatusList(
				failedPipelines,
				pipeline,
				"App Not Yet Deleted.",
				bean.NOT_YET_DELETED)
		}
	}

	return &bean.DeploymentAppTypeChangeResponse{
		SuccessfulPipelines: successfulPipelines,
		FailedPipelines:     failedPipelines,
	}
}

// deleteArgoCdApp takes context and deployment app name used in argo cd and deletes
// the application in argo cd.
func (impl *AppDeploymentTypeChangeManagerImpl) deleteArgoCdApp(ctx context.Context, pipeline *pipelineConfig.Pipeline, deploymentAppName string,
	cascadeDelete bool) error {

	if !pipeline.DeploymentAppCreated {
		return nil
	}

	// building the argocd application delete request
	req := &application.ApplicationDeleteRequest{
		Name:    &deploymentAppName,
		Cascade: &cascadeDelete,
	}

	_, err := impl.application.Delete(ctx, req)

	if err != nil {
		impl.logger.Errorw("error in deleting argocd application", "err", err)
		// Possible that argocd app got deleted but db updation failed
		if strings.Contains(err.Error(), "code = NotFound") {
			return nil
		}
		return err
	}
	return nil
}

func (impl *AppDeploymentTypeChangeManagerImpl) appendToDeploymentChangeStatusList(pipelines []*bean.DeploymentChangeStatus,
	pipeline *pipelineConfig.Pipeline, error string, status bean.Status) []*bean.DeploymentChangeStatus {

	return append(pipelines, &bean.DeploymentChangeStatus{
		PipelineId: pipeline.Id,
		AppId:      pipeline.AppId,
		AppName:    pipeline.App.AppName,
		EnvId:      pipeline.EnvironmentId,
		EnvName:    pipeline.Environment.Name,
		Error:      error,
		Status:     status,
	})
}

// deleteHelmApp takes in context and pipeline object and deletes the release in helm
func (impl *AppDeploymentTypeChangeManagerImpl) deleteHelmApp(ctx context.Context, pipeline *pipelineConfig.Pipeline) error {

	if !pipeline.DeploymentAppCreated {
		return nil
	}

	// validation
	if !util.IsHelmApp(pipeline.DeploymentAppType) {
		return errors.New("unable to delete pipeline with id: " + strconv.Itoa(pipeline.Id) + ", not a helm app")
	}

	// create app identifier
	appIdentifier := &service.AppIdentifier{
		ClusterId:   pipeline.Environment.ClusterId,
		ReleaseName: pipeline.DeploymentAppName,
		Namespace:   pipeline.Environment.Namespace,
	}

	// call for delete resource
	deleteResponse, err := impl.helmAppService.DeleteApplication(ctx, appIdentifier)

	if err != nil {
		impl.logger.Errorw("error in deleting helm application", "error", err, "appIdentifier", appIdentifier)
		return err
	}

	if deleteResponse == nil || !deleteResponse.GetSuccess() {
		return errors.New("helm delete application response unsuccessful")
	}
	return nil
}
