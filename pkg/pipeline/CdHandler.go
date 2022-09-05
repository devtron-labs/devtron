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
	"bufio"
	"context"
	"errors"
	"fmt"
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"os"
	"strconv"
	"strings"
	"time"
)

type CdHandler interface {
	UpdateWorkflow(workflowStatus v1alpha1.WorkflowStatus) (int, string, error)
	GetCdBuildHistory(appId int, environmentId int, pipelineId int, offset int, size int) ([]pipelineConfig.CdWorkflowWithArtifact, error)
	GetRunningWorkflowLogs(environmentId int, pipelineId int, workflowId int) (*bufio.Reader, func() error, error)
	FetchCdWorkflowDetails(appId int, environmentId int, pipelineId int, buildId int) (WorkflowResponse, error)
	DownloadCdWorkflowArtifacts(pipelineId int, buildId int) (*os.File, error)
	FetchCdPrePostStageStatus(pipelineId int) ([]pipelineConfig.CdWorkflowWithArtifact, error)
	CancelStage(workflowRunnerId int) (int, error)
	FetchAppWorkflowStatusForTriggerView(pipelineId int) ([]*pipelineConfig.CdWorkflowStatus, error)
	CheckHelmAppStatusPeriodicallyAndUpdateInDb(timeForDegradation int) error
	CheckArgoAppStatusPeriodicallyAndUpdateInDb(timeForDegradation int) error
}

type CdHandlerImpl struct {
	Logger                           *zap.SugaredLogger
	cdService                        CdWorkflowService
	cdConfig                         *CdConfig
	ciConfig                         *CiConfig
	userService                      user.UserService
	ciLogService                     CiLogService
	ciArtifactRepository             repository.CiArtifactRepository
	ciPipelineMaterialRepository     pipelineConfig.CiPipelineMaterialRepository
	cdWorkflowRepository             pipelineConfig.CdWorkflowRepository
	envRepository                    repository2.EnvironmentRepository
	pipelineRepository               pipelineConfig.PipelineRepository
	ciWorkflowRepository             pipelineConfig.CiWorkflowRepository
	helmAppService                   client.HelmAppService
	pipelineOverrideRepository       chartConfig.PipelineOverrideRepository
	workflowDagExecutor              WorkflowDagExecutor
	appListingService                app.AppListingService
	appListingRepository             repository.AppListingRepository
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository
	application                      application.ServiceClient
	argoUserService                  argo.ArgoUserService
	deploymentFailureHandler         app.DeploymentFailureHandler
}

func NewCdHandlerImpl(Logger *zap.SugaredLogger, cdConfig *CdConfig, userService user.UserService,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	cdWorkflowService CdWorkflowService,
	ciLogService CiLogService,
	ciArtifactRepository repository.CiArtifactRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	envRepository repository2.EnvironmentRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	ciConfig *CiConfig, helmAppService client.HelmAppService,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository, workflowDagExecutor WorkflowDagExecutor,
	appListingService app.AppListingService, appListingRepository repository.AppListingRepository,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	application application.ServiceClient, argoUserService argo.ArgoUserService,
	deploymentFailureHandler app.DeploymentFailureHandler) *CdHandlerImpl {
	return &CdHandlerImpl{
		Logger:                           Logger,
		cdConfig:                         cdConfig,
		userService:                      userService,
		cdService:                        cdWorkflowService,
		ciLogService:                     ciLogService,
		cdWorkflowRepository:             cdWorkflowRepository,
		ciArtifactRepository:             ciArtifactRepository,
		ciPipelineMaterialRepository:     ciPipelineMaterialRepository,
		envRepository:                    envRepository,
		pipelineRepository:               pipelineRepository,
		ciWorkflowRepository:             ciWorkflowRepository,
		ciConfig:                         ciConfig,
		helmAppService:                   helmAppService,
		pipelineOverrideRepository:       pipelineOverrideRepository,
		workflowDagExecutor:              workflowDagExecutor,
		appListingService:                appListingService,
		appListingRepository:             appListingRepository,
		pipelineStatusTimelineRepository: pipelineStatusTimelineRepository,
		application:                      application,
		argoUserService:                  argoUserService,
		deploymentFailureHandler:         deploymentFailureHandler,
	}
}

func (impl *CdHandlerImpl) CheckArgoAppStatusPeriodicallyAndUpdateInDb(timeForDegradation int) error {
	//getting all the progressing status that are stucked since some time
	deploymentStatuses, err := impl.appListingService.GetLastProgressingDeploymentStatusesOfActiveAppsWithActiveEnvs(timeForDegradation)
	if err != nil {
		impl.Logger.Errorw("err in getting latest deployment statuses for argo pipelines", err)
		return err
	}
	impl.Logger.Infow("received deployment statuses for stucked argo cd pipelines", "deploymentStatuses", deploymentStatuses)
	var newDeploymentStatuses []repository.DeploymentStatus
	var newCdWfrs []pipelineConfig.CdWorkflowRunner
	var timelines []pipelineConfig.PipelineStatusTimeline
	for _, deploymentStatus := range deploymentStatuses {
		timelineStatus, appStatus, statusMessage := impl.GetAppStatusByResourceTreeFetchFromArgo(deploymentStatus.AppName)
		newDeploymentStatus := deploymentStatus
		newDeploymentStatus.Id = 0
		newDeploymentStatus.Status = appStatus
		newDeploymentStatus.CreatedOn = time.Now()
		newDeploymentStatus.UpdatedOn = time.Now()
		newDeploymentStatuses = append(newDeploymentStatuses, newDeploymentStatus)
		var cdWfr pipelineConfig.CdWorkflowRunner
		cdWfr, err = impl.cdWorkflowRepository.FindCdWorkflowRunnerByEnvironmentIdAndRunnerType(deploymentStatus.AppId, deploymentStatus.EnvId, bean.CD_WORKFLOW_TYPE_DEPLOY)
		if err != nil {
			//only log this error and continue for next deployment status
			impl.Logger.Errorw("found error, skipping argo apps status update for this trigger", "appId", deploymentStatus.AppId, "envId", deploymentStatus.EnvId, "err", err)
			continue
		}
		cdWfr.Status = appStatus
		newCdWfrs = append(newCdWfrs, cdWfr)
		// creating cd pipeline status timeline for degraded app
		timeline := pipelineConfig.PipelineStatusTimeline{
			CdWorkflowRunnerId: cdWfr.Id,
			Status:             timelineStatus,
			StatusDetail:       statusMessage,
			StatusTime:         time.Now(),
			AuditLog: sql.AuditLog{
				CreatedBy: 1,
				CreatedOn: time.Now(),
				UpdatedBy: 1,
				UpdatedOn: time.Now(),
			},
		}
		timelines = append(timelines, timeline)
		//writing pipeline failure event
		impl.deploymentFailureHandler.WriteCDFailureEvent(cdWfr.CdWorkflow.PipelineId, deploymentStatus.AppId, deploymentStatus.EnvId)
	}

	dbConnection := impl.cdWorkflowRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.Logger.Errorw("error on update status, db get txn failed", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	if len(newDeploymentStatuses) > 0 {
		err = impl.appListingRepository.SaveNewDeploymentsWithTxn(newDeploymentStatuses, tx)
		if err != nil {
			impl.Logger.Errorw("error on saving new deployment statuses for wf", "err", err)
			return err
		}
	}
	if len(newCdWfrs) > 0 {
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunnersWithTxn(newCdWfrs, tx)
		if err != nil {
			impl.Logger.Errorw("error on update cd workflow runners", "cdWfr", newCdWfrs, "err", err)
			return err
		}
	}
	if len(timelines) > 0 {
		err = impl.pipelineStatusTimelineRepository.SaveTimelinesWithTxn(timelines, tx)
		if err != nil {
			impl.Logger.Errorw("error in creating timeline statuses for degraded app", "err", err, "timelines", timelines)
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.Logger.Errorw("error on db transaction commit for", "err", err)
		return err
	}
	return nil
}

func (impl *CdHandlerImpl) GetAppStatusByResourceTreeFetchFromArgo(appName string) (timelineStatus pipelineConfig.TimelineStatus, appStatus, statusMessage string) {
	//this should only be called when we have git-ops configured
	//try fetching status from argo cd
	acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		impl.Logger.Errorw("error in getting acd token", "err", err)
	}
	ctx := context.WithValue(context.Background(), "token", acdToken)
	query := &application2.ResourcesQuery{
		ApplicationName: &appName,
	}
	resp, err := impl.application.ResourceTree(ctx, query)
	if err != nil {
		impl.Logger.Errorw("error in getting resource tree of acd", "err", err, "appName", appName)
		appStatus = WorkflowFailed
		timelineStatus = pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_FAILED
		statusMessage = "Failed to connect to Argo CD to fetch deployment status."
	} else {
		if resp.Status == string(health.HealthStatusHealthy) {
			appStatus = resp.Status
			timelineStatus = pipelineConfig.TIMELINE_STATUS_APP_HEALTHY
			statusMessage = "App is healthy."
		} else if resp.Status == string(health.HealthStatusDegraded) {
			appStatus = resp.Status
			timelineStatus = pipelineConfig.TIMELINE_STATUS_APP_DEGRADED
			statusMessage = "App is degraded."
		} else {
			appStatus = WorkflowFailed
			timelineStatus = pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_FAILED
			statusMessage = "Deployment timed out. Failed to deploy application."
		}
	}
	return timelineStatus, appStatus, statusMessage
}

func (impl *CdHandlerImpl) CheckHelmAppStatusPeriodicallyAndUpdateInDb(timeForDegradation int) error {
	pipelineOverrides, err := impl.pipelineOverrideRepository.FetchHelmTypePipelineOverridesForStatusUpdate()
	if err != nil {
		impl.Logger.Errorw("error on fetching all the recent deployment trigger for helm app type", "err", err)
		return nil
	}
	impl.Logger.Infow("checking helm app status for deployment triggers", "pipelineOverrides", pipelineOverrides)
	for _, pipelineOverride := range pipelineOverrides {
		appIdentifier := &client.AppIdentifier{
			ClusterId:   pipelineOverride.Pipeline.Environment.ClusterId,
			Namespace:   pipelineOverride.Pipeline.Environment.Namespace,
			ReleaseName: fmt.Sprintf("%s-%s", pipelineOverride.Pipeline.App.AppName, pipelineOverride.Pipeline.Environment.Name),
		}
		helmApp, err := impl.helmAppService.GetApplicationDetail(context.Background(), appIdentifier)
		if err != nil {
			impl.Logger.Errorw("error in getting helm app release status ", "appIdentifier", appIdentifier, "err", err)
			//return err
			//skip this error and continue for next workflow status
			impl.Logger.Warnw("found error, skipping helm apps status update for this trigger", "appIdentifier", appIdentifier, "err", err)
			continue
		}
		cdWf, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(pipelineOverride.CdWorkflowId, bean.CD_WORKFLOW_TYPE_DEPLOY)
		if err != nil && err != pg.ErrNoRows {
			impl.Logger.Errorw("err on fetching cd workflow", "CdWorkflowId", pipelineOverride.CdWorkflowId, "err", err)
			//skip this error and continue for next workflow status
			impl.Logger.Warnw("found error, skipping helm apps status update for this trigger", "CdWorkflowId", pipelineOverride.CdWorkflowId, "err", err)
			continue
		}
		if pipelineOverride.CreatedOn.Before(time.Now().Add(-time.Minute * time.Duration(timeForDegradation))) {
			// apps which are still not healthy after DegradeTime, make them "Degraded"
			cdWf.Status = application.Degraded
		} else {
			cdWf.Status = helmApp.ApplicationStatus
		}
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(&cdWf)
		if err != nil {
			impl.Logger.Errorw("error on update cd workflow runner", "cdWf", cdWf, "err", err)
			return err
		}
		impl.Logger.Infow("updating workflow runner status for helm app", "cdWf", cdWf)
		if cdWf.Status == application.Healthy {
			err = impl.workflowDagExecutor.HandleDeploymentSuccessEvent("", pipelineOverride.Id)
			if err != nil {
				impl.Logger.Errorw("error on handling deployment success event", "cdWf", cdWf, "err", err)
				return err
			}
		}
	}
	return nil
}

func (impl *CdHandlerImpl) CancelStage(workflowRunnerId int) (int, error) {
	workflowRunner, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(workflowRunnerId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return 0, err
	}
	if !(string(v1alpha1.NodePending) == workflowRunner.Status || string(v1alpha1.NodeRunning) == workflowRunner.Status) {
		impl.Logger.Info("cannot cancel stage, stage not in progress")
		return 0, errors.New("cannot cancel stage, stage not in progress")
	}
	pipeline, err := impl.pipelineRepository.FindById(workflowRunner.CdWorkflow.PipelineId)
	if err != nil {
		impl.Logger.Errorw("error while fetching cd pipeline", "err", err)
		return 0, err
	}

	env, err := impl.envRepository.FindById(pipeline.EnvironmentId)
	if err != nil {
		impl.Logger.Errorw("could not fetch stage env", "err", err)
		return 0, err
	}

	serverUrl := env.Cluster.ServerUrl
	configMap := env.Cluster.Config
	bearerToken := configMap["bearer_token"]

	var isExtCluster bool
	if workflowRunner.WorkflowType == PRE {
		isExtCluster = pipeline.RunPreStageInEnv
	} else if workflowRunner.WorkflowType == POST {
		isExtCluster = pipeline.RunPostStageInEnv
	}

	runningWf, err := impl.cdService.GetWorkflow(workflowRunner.Name, workflowRunner.Namespace, serverUrl, bearerToken, isExtCluster)
	if err != nil {
		impl.Logger.Errorw("cannot find workflow ", "name", workflowRunner.Name)
		return 0, errors.New("cannot find workflow " + workflowRunner.Name)
	}

	// Terminate workflow
	err = impl.cdService.TerminateWorkflow(runningWf.Name, runningWf.Namespace, serverUrl, bearerToken, isExtCluster)
	if err != nil {
		impl.Logger.Error("cannot terminate wf runner", "err", err)
		return 0, err
	}

	workflowRunner.Status = WorkflowCancel
	err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(workflowRunner)
	if err != nil {
		impl.Logger.Error("cannot update deleted workflow runner status, but wf deleted", "err", err)
		return 0, err
	}
	return workflowRunner.Id, nil
}

func (impl *CdHandlerImpl) UpdateWorkflow(workflowStatus v1alpha1.WorkflowStatus) (int, string, error) {
	wfStatusRs := impl.extractWorkfowStatus(workflowStatus)
	workflowName, status, podStatus, message := wfStatusRs.WorkflowName, wfStatusRs.Status, wfStatusRs.PodStatus, wfStatusRs.Message
	impl.Logger.Debugw("cd update for ", "wf ", workflowName, "status", status)
	if workflowName == "" {
		return 0, "", errors.New("invalid wf name")
	}
	workflowId, err := strconv.Atoi(workflowName[:strings.Index(workflowName, "-")])
	if err != nil {
		impl.Logger.Error("invalid wf status update req", "err", err)
		return 0, "", err
	}

	savedWorkflow, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(workflowId)
	if err != nil {
		impl.Logger.Error("cannot get saved wf", "err", err)
		return 0, "", err
	}

	ciWorkflowConfig, err := impl.cdWorkflowRepository.FindConfigByPipelineId(savedWorkflow.CdWorkflow.PipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("unable to fetch ciWorkflowConfig", "err", err)
		return 0, "", err
	}

	ciArtifactLocationFormat := ciWorkflowConfig.CdArtifactLocationFormat
	if ciArtifactLocationFormat == "" {
		ciArtifactLocationFormat = impl.cdConfig.CdArtifactLocationFormat
	}

	if impl.stateChanged(status, podStatus, message, workflowStatus.FinishedAt.Time, savedWorkflow) {
		if savedWorkflow.Status != WorkflowCancel {
			savedWorkflow.Status = status
		}
		savedWorkflow.PodStatus = podStatus
		savedWorkflow.Message = message
		savedWorkflow.FinishedOn = workflowStatus.FinishedAt.Time
		savedWorkflow.Name = workflowName
		savedWorkflow.LogLocation = wfStatusRs.LogLocation
		impl.Logger.Debugw("updating workflow ", "workflow", savedWorkflow)
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(savedWorkflow)
		if err != nil {
			impl.Logger.Error("update wf failed for id " + strconv.Itoa(savedWorkflow.Id))
			return 0, "", err
		}
		if string(v1alpha1.NodeError) == savedWorkflow.Status || string(v1alpha1.NodeFailed) == savedWorkflow.Status {
			impl.Logger.Warnw("cd stage failed for workflow: ", "wfId", savedWorkflow.Id)
		}
	}
	return savedWorkflow.Id, savedWorkflow.Status, nil
}

func (impl *CdHandlerImpl) extractWorkfowStatus(workflowStatus v1alpha1.WorkflowStatus) *WorkflowStatus {
	workflowName := ""
	status := string(workflowStatus.Phase)
	podStatus := "Pending"
	message := ""
	logLocation := ""
	for k, v := range workflowStatus.Nodes {
		impl.Logger.Debugw("extractWorkflowStatus", "workflowName", k, "v", v)
		if v.TemplateName == CD_WORKFLOW_NAME {
			workflowName = k
			podStatus = string(v.Phase)
			message = v.Message
			if v.Outputs != nil && len(v.Outputs.Artifacts) > 0 && v.Outputs.Artifacts[0].S3 != nil {
				logLocation = v.Outputs.Artifacts[0].S3.Key
			}
			break
		}
	}
	workflowStatusRes := &WorkflowStatus{
		WorkflowName: workflowName,
		Status:       status,
		PodStatus:    podStatus,
		Message:      message,
		LogLocation:  logLocation,
	}
	return workflowStatusRes
}

type WorkflowStatus struct {
	WorkflowName, Status, PodStatus, Message, LogLocation string
}

func (impl *CdHandlerImpl) stateChanged(status string, podStatus string, msg string,
	finishedAt time.Time, savedWorkflow *pipelineConfig.CdWorkflowRunner) bool {
	return savedWorkflow.Status != status || savedWorkflow.PodStatus != podStatus || savedWorkflow.Message != msg || savedWorkflow.FinishedOn != finishedAt
}

func (impl *CdHandlerImpl) GetCdBuildHistory(appId int, environmentId int, pipelineId int, offset int, size int) ([]pipelineConfig.CdWorkflowWithArtifact, error) {

	var cdWorkflowArtifact []pipelineConfig.CdWorkflowWithArtifact
	if pipelineId == 0 {
		wfrList, err := impl.cdWorkflowRepository.FindCdWorkflowMetaByEnvironmentId(appId, environmentId, offset, size)
		if err != nil && err != pg.ErrNoRows {
			return cdWorkflowArtifact, err
		}
		cdWorkflowArtifact = impl.converterWFRList(wfrList)
	} else {
		wfrList, err := impl.cdWorkflowRepository.FindCdWorkflowMetaByPipelineId(pipelineId, offset, size)
		if err != nil && err != pg.ErrNoRows {
			return cdWorkflowArtifact, err
		}
		cdWorkflowArtifact = impl.converterWFRList(wfrList)
	}

	return cdWorkflowArtifact, nil
}

func (impl *CdHandlerImpl) GetRunningWorkflowLogs(environmentId int, pipelineId int, wfrId int) (*bufio.Reader, func() error, error) {
	cdWorkflow, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(wfrId)
	if err != nil {
		impl.Logger.Errorw("error on fetch wf runner", "err", err)
		return nil, nil, err
	}

	env, err := impl.envRepository.FindById(environmentId)
	if err != nil {
		impl.Logger.Errorw("could not fetch stage env", "err", err)
		return nil, nil, err
	}

	pipeline, err := impl.pipelineRepository.FindById(cdWorkflow.CdWorkflow.PipelineId)
	if err != nil {
		impl.Logger.Errorw("error while fetching cd pipeline", "err", err)
		return nil, nil, err
	}

	serverUrl := env.Cluster.ServerUrl
	configMap := env.Cluster.Config
	bearerToken := configMap["bearer_token"]

	var isExtCluster bool
	if cdWorkflow.WorkflowType == PRE {
		isExtCluster = pipeline.RunPreStageInEnv
	} else if cdWorkflow.WorkflowType == POST {
		isExtCluster = pipeline.RunPostStageInEnv
	}
	return impl.getWorkflowLogs(pipelineId, cdWorkflow, bearerToken, serverUrl, isExtCluster)
}

func (impl *CdHandlerImpl) getWorkflowLogs(pipelineId int, cdWorkflow *pipelineConfig.CdWorkflowRunner, token string, host string, runStageInEnv bool) (*bufio.Reader, func() error, error) {
	cdLogRequest := CiLogRequest{
		WorkflowName: cdWorkflow.Name,
		Namespace:    cdWorkflow.Namespace,
	}

	logStream, cleanUp, err := impl.ciLogService.FetchRunningWorkflowLogs(cdLogRequest, token, host, runStageInEnv)
	if logStream == nil || err != nil {
		if string(v1alpha1.NodeSucceeded) == cdWorkflow.Status || string(v1alpha1.NodeError) == cdWorkflow.Status || string(v1alpha1.NodeFailed) == cdWorkflow.Status || cdWorkflow.Status == WorkflowCancel {
			impl.Logger.Debugw("pod is not live ", "err", err)
			return impl.getLogsFromRepository(pipelineId, cdWorkflow)
		}
		impl.Logger.Errorw("err on fetch workflow logs", "err", err)
		return nil, nil, err
	}
	logReader := bufio.NewReader(logStream)
	return logReader, cleanUp, err
}

func (impl *CdHandlerImpl) getLogsFromRepository(pipelineId int, cdWorkflow *pipelineConfig.CdWorkflowRunner) (*bufio.Reader, func() error, error) {
	impl.Logger.Debug("getting historic logs")

	cdConfig, err := impl.cdWorkflowRepository.FindConfigByPipelineId(pipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", err)
		return nil, nil, err
	}

	if cdConfig.LogsBucket == "" {
		cdConfig.LogsBucket = impl.cdConfig.DefaultBuildLogsBucket //TODO -fixme
	}
	if cdConfig.CdCacheRegion == "" {
		cdConfig.CdCacheRegion = impl.cdConfig.DefaultCdLogsBucketRegion
	}

	cdLogRequest := CiLogRequest{
		PipelineId:   cdWorkflow.CdWorkflow.PipelineId,
		WorkflowId:   cdWorkflow.Id,
		WorkflowName: cdWorkflow.Name,
		//AccessKey:    cdConfig,
		//SecretKet:    cdWorkflow.CdPipeline.CiTemplate.DockerRegistry.AWSSecretAccessKey,
		Region:        cdConfig.CdCacheRegion,
		LogsBucket:    cdConfig.LogsBucket,
		LogsFilePath:  cdWorkflow.LogLocation, // impl.cdConfig.DefaultBuildLogsKeyPrefix + "/" + cdWorkflow.Name + "/main.log", //TODO - fixme
		CloudProvider: impl.ciConfig.CloudProvider,
		AzureBlobConfig: &AzureBlobConfig{
			Enabled:            impl.ciConfig.CloudProvider == BLOB_STORAGE_AZURE,
			AccountName:        impl.ciConfig.AzureAccountName,
			BlobContainerCiLog: impl.ciConfig.AzureBlobContainerCiLog,
			AccountKey:         impl.ciConfig.AzureAccountKey,
		},
	}
	if impl.ciConfig.CloudProvider == BLOB_STORAGE_MINIO {
		cdLogRequest.MinioEndpoint = impl.ciConfig.MinioEndpoint
		cdLogRequest.AccessKey = impl.ciConfig.MinioAccessKey
		cdLogRequest.SecretKet = impl.ciConfig.MinioSecretKey
		cdLogRequest.Region = impl.ciConfig.MinioRegion
	}
	impl.Logger.Infow("s3 log req ", "req", cdLogRequest)
	oldLogsStream, cleanUp, err := impl.ciLogService.FetchLogs(cdLogRequest)
	if err != nil {
		impl.Logger.Errorw("err", err)
		return nil, nil, err
	}
	logReader := bufio.NewReader(oldLogsStream)
	return logReader, cleanUp, err
}

func (impl *CdHandlerImpl) FetchCdWorkflowDetails(appId int, environmentId int, pipelineId int, buildId int) (WorkflowResponse, error) {
	workflowR, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(buildId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("err", "err", err)
		return WorkflowResponse{}, err
	} else if err == pg.ErrNoRows {
		return WorkflowResponse{}, nil
	}
	workflow := impl.converterWFR(*workflowR)

	triggeredByUser, err := impl.userService.GetById(workflow.TriggeredBy)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", "err", err)
		return WorkflowResponse{}, err
	}
	if triggeredByUser == nil {
		triggeredByUser = &bean.UserInfo{EmailId: "anonymous"}
	}

	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(workflowR.CdWorkflow.Pipeline.CiPipelineId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return WorkflowResponse{}, err
	}

	var ciMaterialsArr []CiPipelineMaterialResponse
	for _, m := range ciMaterials {
		res := CiPipelineMaterialResponse{
			Id:              m.Id,
			GitMaterialId:   m.GitMaterialId,
			GitMaterialName: m.GitMaterial.Name[strings.Index(m.GitMaterial.Name, "-")+1:],
			Type:            string(m.Type),
			Value:           m.Value,
			Active:          m.Active,
			Url:             m.GitMaterial.Url,
		}
		ciMaterialsArr = append(ciMaterialsArr, res)
	}
	ciWf, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflowByArtifactId(workflow.CiArtifactId)
	if err != nil {
		impl.Logger.Errorw("error in fetching ci wf", "artifactId", workflow.CiArtifactId, "err", err)
		return WorkflowResponse{}, err
	}
	workflowResponse := WorkflowResponse{
		Id:               workflow.Id,
		Name:             workflow.Name,
		Status:           workflow.Status,
		PodStatus:        workflow.PodStatus,
		Message:          workflow.Message,
		StartedOn:        workflow.StartedOn,
		FinishedOn:       workflow.FinishedOn,
		Namespace:        workflow.Namespace,
		CiMaterials:      ciMaterialsArr,
		TriggeredBy:      workflow.TriggeredBy,
		TriggeredByEmail: triggeredByUser.EmailId,
		Artifact:         workflow.Image,
		Stage:            workflow.WorkflowType,
		GitTriggers:      ciWf.GitTriggers,
	}
	return workflowResponse, nil
}

func (impl *CdHandlerImpl) DownloadCdWorkflowArtifacts(pipelineId int, buildId int) (*os.File, error) {
	wfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(buildId)
	if err != nil {
		impl.Logger.Errorw("unable to fetch ciWorkflow", "err", err)
		return nil, err
	}

	cdConfig, err := impl.cdWorkflowRepository.FindConfigByPipelineId(pipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("unable to fetch ciConfig", "err", err)
		return nil, err
	}

	if cdConfig.LogsBucket == "" {
		cdConfig.LogsBucket = impl.cdConfig.DefaultBuildLogsBucket
	}
	if cdConfig.CdCacheRegion == "" {
		cdConfig.CdCacheRegion = impl.cdConfig.DefaultCdLogsBucketRegion
	}

	item := strconv.Itoa(wfr.Id)
	file, err := os.Create(item)
	if err != nil {
		impl.Logger.Errorw("unable to open file", "err", err)
		return nil, errors.New("unable to open file")
	}

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(cdConfig.CdCacheRegion),
		//Credentials: credentials.NewStaticCredentials(ciWorkflow.CiPipeline.CiTemplate.DockerRegistry.AWSAccessKeyId, ciWorkflow.CiPipeline.CiTemplate.DockerRegistry.AWSSecretAccessKey, ""),
	})

	downloader := s3manager.NewDownloader(sess)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(cdConfig.LogsBucket),
			Key:    aws.String(fmt.Sprintf("%s/"+impl.cdConfig.CdArtifactLocationFormat, impl.cdConfig.DefaultArtifactKeyPrefix, wfr.CdWorkflow.Id, wfr.Id)),
		})
	if err != nil {
		impl.Logger.Errorw("unable to download file from s3", "err", err)
		return nil, err
	}
	impl.Logger.Infow("Downloaded ", "name", file.Name(), "bytes", numBytes)
	return file, nil
}

func (impl *CdHandlerImpl) converterWFR(wfr pipelineConfig.CdWorkflowRunner) pipelineConfig.CdWorkflowWithArtifact {
	workflow := pipelineConfig.CdWorkflowWithArtifact{}
	if wfr.Id > 0 {
		workflow.Name = wfr.Name
		workflow.Id = wfr.Id
		workflow.Namespace = wfr.Namespace
		workflow.Status = wfr.Status
		workflow.Message = wfr.Message
		workflow.PodStatus = wfr.PodStatus
		workflow.FinishedOn = wfr.FinishedOn
		workflow.TriggeredBy = wfr.TriggeredBy
		workflow.StartedOn = wfr.StartedOn
		workflow.WorkflowType = string(wfr.WorkflowType)
		workflow.CdWorkflowId = wfr.CdWorkflowId
		workflow.Image = wfr.CdWorkflow.CiArtifact.Image
		workflow.PipelineId = wfr.CdWorkflow.PipelineId
		workflow.CiArtifactId = wfr.CdWorkflow.CiArtifactId

	}
	return workflow
}

func (impl *CdHandlerImpl) converterWFRList(wfrList []pipelineConfig.CdWorkflowRunner) []pipelineConfig.CdWorkflowWithArtifact {
	var workflowList []pipelineConfig.CdWorkflowWithArtifact
	var results []pipelineConfig.CdWorkflowWithArtifact
	var ids []int32
	for _, item := range wfrList {
		ids = append(ids, item.TriggeredBy)
		workflowList = append(workflowList, impl.converterWFR(item))
	}
	userEmails := make(map[int32]string)
	users, err := impl.userService.GetByIds(ids)
	if err != nil {
		impl.Logger.Errorw("unable to find user", "err", err)
	}
	for _, item := range users {
		userEmails[item.Id] = item.EmailId
	}
	for _, item := range workflowList {
		item.EmailId = userEmails[item.TriggeredBy]
		results = append(results, item)
	}
	return results
}

func (impl *CdHandlerImpl) FetchCdPrePostStageStatus(pipelineId int) ([]pipelineConfig.CdWorkflowWithArtifact, error) {
	var results []pipelineConfig.CdWorkflowWithArtifact
	wfrPre, err := impl.cdWorkflowRepository.FindLastStatusByPipelineIdAndRunnerType(pipelineId, bean.CD_WORKFLOW_TYPE_PRE)
	if err != nil && err != pg.ErrNoRows {
		return results, err
	}
	if wfrPre.Id > 0 {
		workflowPre := impl.converterWFR(wfrPre)
		results = append(results, workflowPre)
	} else {
		workflowPre := pipelineConfig.CdWorkflowWithArtifact{Status: "Notbuilt", WorkflowType: string(bean.CD_WORKFLOW_TYPE_PRE), PipelineId: pipelineId}
		results = append(results, workflowPre)
	}

	wfrPost, err := impl.cdWorkflowRepository.FindLastStatusByPipelineIdAndRunnerType(pipelineId, bean.CD_WORKFLOW_TYPE_POST)
	if err != nil && err != pg.ErrNoRows {
		return results, err
	}
	if wfrPost.Id > 0 {
		workflowPost := impl.converterWFR(wfrPost)
		results = append(results, workflowPost)
	} else {
		workflowPost := pipelineConfig.CdWorkflowWithArtifact{Status: "Notbuilt", WorkflowType: string(bean.CD_WORKFLOW_TYPE_POST), PipelineId: pipelineId}
		results = append(results, workflowPost)
	}
	return results, nil

}

func (impl *CdHandlerImpl) FetchAppWorkflowStatusForTriggerView(appId int) ([]*pipelineConfig.CdWorkflowStatus, error) {
	var cdWorkflowStatus []*pipelineConfig.CdWorkflowStatus

	pipelines, err := impl.pipelineRepository.FindActiveByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		return cdWorkflowStatus, err
	}
	pipelineIds := make([]int, 0)
	//pipelineIdsMap := make(map[int]int)
	for _, pipeline := range pipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
	}

	if len(pipelineIds) == 0 {
		return cdWorkflowStatus, nil
	}

	cdMap := make(map[int]*pipelineConfig.CdWorkflowStatus)
	result, err := impl.cdWorkflowRepository.FetchAllCdStagesLatestEntity(pipelineIds)
	if err != nil {
		return cdWorkflowStatus, err
	}
	var wfrIds []int
	for _, item := range result {
		wfrIds = append(wfrIds, item.WfrId)
	}

	statusMap := make(map[int]string)
	if len(wfrIds) > 0 {
		wfrList, err := impl.cdWorkflowRepository.FetchAllCdStagesLatestEntityStatus(wfrIds)
		if err != nil && !util.IsErrNoRows(err) {
			return cdWorkflowStatus, err
		}
		for _, item := range wfrList {
			statusMap[item.Id] = item.Status
		}
	}

	for _, item := range result {
		if _, ok := cdMap[item.PipelineId]; !ok {
			cdWorkflowStatus := &pipelineConfig.CdWorkflowStatus{}
			cdWorkflowStatus.PipelineId = item.PipelineId
			cdWorkflowStatus.CiPipelineId = item.CiPipelineId
			if item.WorkflowType == "PRE" {
				cdWorkflowStatus.PreStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == "DEPLOY" {
				cdWorkflowStatus.DeployStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == "POST" {
				cdWorkflowStatus.PostStatus = statusMap[item.WfrId]
			}
			cdMap[item.PipelineId] = cdWorkflowStatus
		} else {
			cdWorkflowStatus := cdMap[item.PipelineId]
			cdWorkflowStatus.PipelineId = item.PipelineId
			cdWorkflowStatus.CiPipelineId = item.CiPipelineId
			if item.WorkflowType == "PRE" {
				cdWorkflowStatus.PreStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == "DEPLOY" {
				cdWorkflowStatus.DeployStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == "POST" {
				cdWorkflowStatus.PostStatus = statusMap[item.WfrId]
			}
			cdMap[item.PipelineId] = cdWorkflowStatus
		}
	}

	for _, item := range cdMap {
		if item.PreStatus == "" {
			item.PreStatus = "Not Triggered"
		}
		if item.DeployStatus == "" {
			item.DeployStatus = "Not Deployed"
		}
		if item.PostStatus == "" {
			item.PostStatus = "Not Triggered"
		}
		cdWorkflowStatus = append(cdWorkflowStatus, item)
	}

	if len(cdWorkflowStatus) == 0 {
		for _, item := range pipelineIds {
			cdWs := &pipelineConfig.CdWorkflowStatus{}
			cdWs.PipelineId = item
			cdWs.PreStatus = "Not Triggered"
			cdWs.DeployStatus = "Not Deployed"
			cdWs.PostStatus = "Not Triggered"
			cdWorkflowStatus = append(cdWorkflowStatus, cdWs)
		}
	} else {
		for _, item := range pipelineIds {
			if _, ok := cdMap[item]; !ok {
				cdWs := &pipelineConfig.CdWorkflowStatus{}
				cdWs.PipelineId = item
				cdWs.PreStatus = "Not Triggered"
				cdWs.DeployStatus = "Not Deployed"
				cdWs.PostStatus = "Not Triggered"
				cdWorkflowStatus = append(cdWorkflowStatus, cdWs)
			}
		}
	}

	return cdWorkflowStatus, err
}
