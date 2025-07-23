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

package dag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/workflow/workflowStatusLatest"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/common-lib/async"
	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/common-lib/utils/workFlow"
	bean6 "github.com/devtron-labs/devtron/api/helm-app/bean"
	client2 "github.com/devtron-labs/devtron/api/helm-app/service"
	helmBean "github.com/devtron-labs/devtron/api/helm-app/service/bean"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/adapter/cdWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow"
	cdWorkflow2 "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	"github.com/devtron-labs/devtron/pkg/app/status"
	bean7 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/build/artifacts"
	bean5 "github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	buildCommonBean "github.com/devtron-labs/devtron/pkg/build/pipeline/bean/common"
	"github.com/devtron-labs/devtron/pkg/build/trigger"
	"github.com/devtron-labs/devtron/pkg/cluster/adapter"
	repository5 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	common2 "github.com/devtron-labs/devtron/pkg/deployment/common"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps"
	triggerAdapter "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/adapter"
	triggerBean "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/userDeploymentRequest/service"
	eventProcessorBean "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/devtron-labs/devtron/pkg/executor"
	"github.com/devtron-labs/devtron/pkg/fluxApplication"
	bean8 "github.com/devtron-labs/devtron/pkg/fluxApplication/bean"
	k8sPkg "github.com/devtron-labs/devtron/pkg/k8s"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	constants2 "github.com/devtron-labs/devtron/pkg/pipeline/constants"
	repository2 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning"
	repository3 "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	bean4 "github.com/devtron-labs/devtron/pkg/workflow/cd/bean"
	"github.com/devtron-labs/devtron/pkg/workflow/dag/adaptor"
	bean2 "github.com/devtron-labs/devtron/pkg/workflow/dag/bean"
	"github.com/devtron-labs/devtron/pkg/workflow/dag/helper"
	auditService "github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/service"
	error2 "github.com/devtron-labs/devtron/util/error"
	util2 "github.com/devtron-labs/devtron/util/event"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	util4 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type WorkflowDagExecutor interface {
	UpdateCiWorkflowStatusFailure(timeoutForFailureCiBuild int) error
	UpdateCiWorkflowForCiSuccess(request *bean2.CiArtifactWebhookRequest) (err error)
	HandleCiSuccessEvent(triggerContext triggerBean.TriggerContext, ciPipelineId int, request *bean2.CiArtifactWebhookRequest, imagePushedAt time.Time) (id int, err error)
	HandlePreStageSuccessEvent(triggerContext triggerBean.TriggerContext, cdStageCompleteEvent eventProcessorBean.CdStageCompleteEvent) error
	HandleDeploymentSuccessEvent(triggerContext triggerBean.TriggerContext, pipelineOverride *chartConfig.PipelineOverride) error
	HandlePostStageSuccessEvent(triggerContext triggerBean.TriggerContext, wfr *bean4.CdWorkflowRunnerDto, cdWorkflowId int, cdPipelineId int, triggeredBy int32, pluginRegistryImageDetails map[string][]string) error
	HandleCdStageReTrigger(runner *pipelineConfig.CdWorkflowRunner) error
	HandleCiStepFailedEvent(ciPipelineId int, request *bean2.CiArtifactWebhookRequest) (err error)
	HandleExternalCiWebhook(externalCiId int, request *bean2.CiArtifactWebhookRequest,
		auth func(token string, projectObject string, envObject string) bool, token string) (id int, err error)

	ProcessDevtronAsyncInstallRequest(cdAsyncInstallReq *eventProcessorBean.UserDeploymentRequest, ctx context.Context) error

	UpdateWorkflowRunnerStatusForDeployment(appIdentifier *helmBean.AppIdentifier, wfr *pipelineConfig.CdWorkflowRunner, skipReleaseNotFound bool) bool

	BuildCiArtifactRequestForWebhook(event pipeline.ExternalCiWebhookDto) (*bean2.CiArtifactWebhookRequest, error)
	UpdateWorkflowRunnerStatusForFluxDeployment(appIdentifier *bean8.FluxAppIdentifier, wfr *pipelineConfig.CdWorkflowRunner, pipelineOverride *chartConfig.PipelineOverride) bool
}

type WorkflowDagExecutorImpl struct {
	logger                        *zap.SugaredLogger
	pipelineRepository            pipelineConfig.PipelineRepository
	pipelineOverrideRepository    chartConfig.PipelineOverrideRepository
	cdWorkflowRepository          pipelineConfig.CdWorkflowRepository
	ciArtifactRepository          repository.CiArtifactRepository
	enforcerUtil                  rbac.EnforcerUtil
	appWorkflowRepository         appWorkflow.AppWorkflowRepository
	pipelineStageService          pipeline.PipelineStageService
	ciWorkflowRepository          pipelineConfig.CiWorkflowRepository
	ciPipelineRepository          pipelineConfig.CiPipelineRepository
	pipelineStageRepository       repository4.PipelineStageRepository
	globalPluginRepository        repository2.GlobalPluginRepository
	config                        *types.CdConfig
	ciConfig                      *types.CiConfig
	appServiceConfig              *app.AppServiceConfig
	eventClient                   client.EventClient
	eventFactory                  client.EventFactory
	customTagService              pipeline.CustomTagService
	pipelineStatusTimelineService status.PipelineStatusTimelineService
	cdWorkflowRunnerService       cd.CdWorkflowRunnerService
	ciService                     pipeline.CiService

	helmAppService client2.HelmAppService

	cdWorkflowCommonService      cd.CdWorkflowCommonService
	cdHandlerService             devtronApps.HandlerService
	userDeploymentRequestService service.UserDeploymentRequestService

	manifestCreationService manifest.ManifestCreationService
	commonArtifactService   artifacts.CommonArtifactService
	deploymentConfigService common2.DeploymentConfigService
	asyncRunnable           *async.Runnable
	scanHistoryRepository   repository3.ImageScanHistoryRepository
	imageScanService        imageScanning.ImageScanService

	K8sUtil                     *k8s.K8sServiceImpl
	envRepository               repository5.EnvironmentRepository
	k8sCommonService            k8sPkg.K8sCommonService
	workflowService             executor.WorkflowService
	ciHandlerService            trigger.HandlerService
	workflowTriggerAuditService auditService.WorkflowTriggerAuditService
	fluxApplicationService      fluxApplication.FluxApplicationService
	workflowStatusUpdateService workflowStatusLatest.WorkflowStatusUpdateService
}

func NewWorkflowDagExecutorImpl(Logger *zap.SugaredLogger, pipelineRepository pipelineConfig.PipelineRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	ciArtifactRepository repository.CiArtifactRepository,
	enforcerUtil rbac.EnforcerUtil,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	pipelineStageService pipeline.PipelineStageService,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	pipelineStageRepository repository4.PipelineStageRepository,
	globalPluginRepository repository2.GlobalPluginRepository,
	eventClient client.EventClient,
	eventFactory client.EventFactory,
	customTagService pipeline.CustomTagService,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	cdWorkflowRunnerService cd.CdWorkflowRunnerService,
	ciService pipeline.CiService,
	helmAppService client2.HelmAppService,
	cdWorkflowCommonService cd.CdWorkflowCommonService,
	cdHandlerService devtronApps.HandlerService,
	userDeploymentRequestService service.UserDeploymentRequestService,
	manifestCreationService manifest.ManifestCreationService,
	commonArtifactService artifacts.CommonArtifactService,
	deploymentConfigService common2.DeploymentConfigService,
	asyncRunnable *async.Runnable,
	scanHistoryRepository repository3.ImageScanHistoryRepository,
	imageScanService imageScanning.ImageScanService,
	K8sUtil *k8s.K8sServiceImpl,
	envRepository repository5.EnvironmentRepository,
	k8sCommonService k8sPkg.K8sCommonService,
	workflowService executor.WorkflowService,
	ciHandlerService trigger.HandlerService,
	workflowTriggerAuditService auditService.WorkflowTriggerAuditService,
	fluxApplicationService fluxApplication.FluxApplicationService,
	workflowStatusUpdateService workflowStatusLatest.WorkflowStatusUpdateService,
) *WorkflowDagExecutorImpl {
	wde := &WorkflowDagExecutorImpl{logger: Logger,
		pipelineRepository:            pipelineRepository,
		pipelineOverrideRepository:    pipelineOverrideRepository,
		cdWorkflowRepository:          cdWorkflowRepository,
		ciArtifactRepository:          ciArtifactRepository,
		enforcerUtil:                  enforcerUtil,
		appWorkflowRepository:         appWorkflowRepository,
		pipelineStageService:          pipelineStageService,
		ciWorkflowRepository:          ciWorkflowRepository,
		ciPipelineRepository:          ciPipelineRepository,
		pipelineStageRepository:       pipelineStageRepository,
		globalPluginRepository:        globalPluginRepository,
		eventClient:                   eventClient,
		eventFactory:                  eventFactory,
		customTagService:              customTagService,
		pipelineStatusTimelineService: pipelineStatusTimelineService,
		helmAppService:                helmAppService,
		cdWorkflowCommonService:       cdWorkflowCommonService,
		cdHandlerService:              cdHandlerService,
		userDeploymentRequestService:  userDeploymentRequestService,
		manifestCreationService:       manifestCreationService,
		commonArtifactService:         commonArtifactService,
		deploymentConfigService:       deploymentConfigService,
		asyncRunnable:                 asyncRunnable,
		scanHistoryRepository:         scanHistoryRepository,
		imageScanService:              imageScanService,
		cdWorkflowRunnerService:       cdWorkflowRunnerService,
		ciService:                     ciService,
		K8sUtil:                       K8sUtil,
		envRepository:                 envRepository,
		k8sCommonService:              k8sCommonService,
		workflowService:               workflowService,
		ciHandlerService:              ciHandlerService,
		workflowTriggerAuditService:   workflowTriggerAuditService,
		fluxApplicationService:        fluxApplicationService,
		workflowStatusUpdateService:   workflowStatusUpdateService,
	}
	config, err := types.GetCdConfig()
	if err != nil {
		return nil
	}
	wde.config = config
	ciConfig, err := types.GetCiConfig()
	if err != nil {
		return nil
	}
	wde.ciConfig = ciConfig
	appServiceConfig, err := app.GetAppServiceConfig()
	if err != nil {
		return nil
	}
	wde.appServiceConfig = appServiceConfig
	return wde
}

func (impl *WorkflowDagExecutorImpl) UpdateCiWorkflowStatusFailure(timeoutForFailureCiBuild int) error {
	ciWorkflows, err := impl.ciWorkflowRepository.FindByStatusesIn([]string{constants2.Starting, constants2.Running})
	if err != nil {
		impl.logger.Errorw("error on fetching ci workflows", "err", err)
		return err
	}
	client, err := impl.K8sUtil.GetClientForInCluster()
	if err != nil {
		impl.logger.Errorw("error while fetching k8s client", "error", err)
		return err
	}

	for _, ciWorkflow := range ciWorkflows {
		var isExt bool
		var env *repository5.Environment
		var restConfig *rest.Config
		if ciWorkflow.Namespace != constants2.DefaultCiWorkflowNamespace {
			isExt = true
			env, err = impl.envRepository.FindById(ciWorkflow.EnvironmentId)
			if err != nil {
				impl.logger.Errorw("could not fetch stage env", "err", err)
				return err
			}
			restConfig, err = impl.getRestConfig(ciWorkflow)
			if err != nil {
				return err
			}
		}

		isEligibleToMarkFailed := false
		isPodDeleted := false
		if time.Since(ciWorkflow.StartedOn) > (time.Minute * time.Duration(timeoutForFailureCiBuild)) {

			//check weather pod is exists or not, if exits check its status
			wf, err := impl.workflowService.GetWorkflowStatus(ciWorkflow.ExecutorType, ciWorkflow.Name, ciWorkflow.Namespace, restConfig)
			if err != nil {
				impl.logger.Warnw("unable to fetch ci workflow", "err", err)
				statusError, ok := err.(*errors2.StatusError)
				if ok && statusError.Status().Code == http.StatusNotFound {
					impl.logger.Warnw("ci workflow not found", "err", err)
					isEligibleToMarkFailed = true
				} else {
					continue
					// skip this and process for next ci workflow
				}
			}

			//if ci workflow is exists, check its pod
			if !isEligibleToMarkFailed {
				ns := constants2.DefaultCiWorkflowNamespace
				if isExt {
					_, client, err = impl.k8sCommonService.GetCoreClientByClusterId(env.ClusterId)
					if err != nil {
						impl.logger.Warnw("error in getting core v1 client using GetCoreClientByClusterId", "err", err, "clusterId", env.Cluster.Id)
						continue
					}
					ns = env.Namespace
				}
				_, err := impl.K8sUtil.GetPodByName(ns, ciWorkflow.PodName, client)
				if err != nil {
					impl.logger.Warnw("unable to fetch ci workflow - pod", "err", err)
					statusError, ok := err.(*errors2.StatusError)
					if ok && statusError.Status().Code == http.StatusNotFound {
						impl.logger.Warnw("pod not found", "err", err)
						isEligibleToMarkFailed = true
					} else {
						continue
						// skip this and process for next ci workflow
					}
				}
				if ciWorkflow.ExecutorType == cdWorkflow2.WORKFLOW_EXECUTOR_TYPE_SYSTEM {
					if wf.Status == string(v1alpha1.WorkflowFailed) {
						isPodDeleted = true
					}
				} else {
					// check workflow status,get the status
					if wf.Status == string(v1alpha1.WorkflowFailed) && wf.Message == cdWorkflow2.POD_DELETED_MESSAGE {
						isPodDeleted = true
					}
				}
			}
		}
		if isEligibleToMarkFailed {
			ciWorkflow.Status = "Failed"
			ciWorkflow.PodStatus = "Failed"
			if isPodDeleted {
				ciWorkflow.Message = cdWorkflow2.POD_DELETED_MESSAGE
				// error logging handled inside handlePodDeleted
				impl.ciHandlerService.HandlePodDeleted(ciWorkflow)
			} else {
				ciWorkflow.Message = "marked failed by job"
			}
			err := impl.ciService.UpdateCiWorkflowWithStage(ciWorkflow)
			if err != nil {
				impl.logger.Errorw("unable to update ci workflow, its eligible to mark failed", "err", err)
				// skip this and process for next ci workflow
			}
			err = impl.customTagService.DeactivateImagePathReservation(ciWorkflow.ImagePathReservationId)
			if err != nil {
				impl.logger.Errorw("unable to update ci workflow, its eligible to mark failed", "err", err)
			}
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) getRestConfig(workflow *pipelineConfig.CiWorkflow) (*rest.Config, error) {
	env, err := impl.envRepository.FindById(workflow.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("could not fetch stage env", "err", err)
		return nil, err
	}

	clusterBean := adapter.GetClusterBean(*env.Cluster)

	clusterConfig := clusterBean.GetClusterConfig()
	restConfig, err := impl.K8sUtil.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.logger.Errorw("error in getting rest config by cluster id", "err", err)
		return nil, err
	}
	return restConfig, nil
}

func (impl *WorkflowDagExecutorImpl) HandleCdStageReTrigger(runner *pipelineConfig.CdWorkflowRunner) error {
	// do not re-trigger if retries = 0
	if !impl.config.WorkflowRetriesEnabled() {
		impl.logger.Debugw("cd stage workflow re-triggering is not enabled")
		return nil
	}

	impl.logger.Infow("re triggering cd stage ", "runnerId", runner.Id)
	var err error
	// add comment for this logic
	if runner.RefCdWorkflowRunnerId != 0 {
		runner, err = impl.cdWorkflowRepository.FindWorkflowRunnerById(runner.RefCdWorkflowRunnerId)
		if err != nil {
			impl.logger.Errorw("error in FindWorkflowRunnerById by id ", "err", err, "wfrId", runner.RefCdWorkflowRunnerId)
			return err
		}
	}
	retryCnt, err := impl.cdWorkflowRepository.FindRetriedWorkflowCountByReferenceId(runner.Id)
	if err != nil {
		impl.logger.Errorw("error in FindRetriedWorkflowCountByReferenceId ", "err", err, "cdWorkflowRunnerId", runner.Id)
		return err
	}

	if retryCnt >= impl.config.MaxCdWorkflowRunnerRetries {
		impl.logger.Infow("maximum retries for this workflow are exhausted, not re-triggering again", "retries", retryCnt, "wfrId", runner.Id)
		return nil
	}

	err = impl.reTriggerCdStageFromSnapshot(runner)
	if err != nil {
		impl.logger.Warnw("failed to retrigger CD stage from snapshot", "runnerId", runner.Id, "err", err)
		return err
	}

	impl.logger.Infow("cd stage re triggered for runner", "runnerId", runner.Id)
	return nil
}

// reTriggerCdStageFromSnapshot attempts to retrigger CD stage using stored workflow config snapshot
func (impl *WorkflowDagExecutorImpl) reTriggerCdStageFromSnapshot(runner *pipelineConfig.CdWorkflowRunner) error {
	impl.logger.Infow("attempting to retrigger CD stage from stored snapshot", "runnerId", runner.Id, "workflowType", runner.WorkflowType)

	triggerRequest := triggerBean.CdTriggerRequest{
		CdWf:                  runner.CdWorkflow,
		Pipeline:              runner.CdWorkflow.Pipeline,
		Artifact:              runner.CdWorkflow.CiArtifact,
		TriggeredBy:           bean7.SYSTEM_USER_ID,
		ApplyAuth:             false,
		RefCdWorkflowRunnerId: runner.Id,
		TriggerContext: triggerBean.TriggerContext{
			Context: context.Background(),
		},
		IsRetrigger: true,
	}

	if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
		_, err := impl.cdHandlerService.TriggerPreStage(triggerRequest)
		if err != nil {
			impl.logger.Errorw("error in TriggerPreStage ", "err", err, "cdWorkflowRunnerId", runner.Id)
			return err
		}
	} else if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
		_, err := impl.cdHandlerService.TriggerPostStage(triggerRequest)
		if err != nil {
			impl.logger.Errorw("error in TriggerPostStage ", "err", err, "cdWorkflowRunnerId", runner.Id)
			return err
		}
	}

	return nil
}

// UpdateWorkflowRunnerStatusForDeployment will update CD workflow runner based on release status and app status
func (impl *WorkflowDagExecutorImpl) UpdateWorkflowRunnerStatusForDeployment(appIdentifier *helmBean.AppIdentifier, wfr *pipelineConfig.CdWorkflowRunner, skipReleaseNotFound bool) bool {
	helmInstalledDevtronApp, err := impl.helmAppService.GetApplicationAndReleaseStatus(context.Background(), appIdentifier)
	if err != nil {
		impl.logger.Errorw("error in getting helm app release status", "appIdentifier", appIdentifier, "err", err)
		// Handle release not found errors
		if skipReleaseNotFound && util.GetClientErrorDetailedMessage(err) != bean6.ErrReleaseNotFound {
			// skip this error and continue for next workflow status
			impl.logger.Warnw("found error, skipping helm apps status update for this trigger", "appIdentifier", appIdentifier, "err", err)
			return false
		}
		// If release not found, mark the deployment as failure
		wfr.Status = cdWorkflow2.WorkflowFailed
		wfr.Message = fmt.Sprintf("helm client error: %s", util.GetClientErrorDetailedMessage(err))
		wfr.FinishedOn = time.Now()
		return true
	}

	switch helmInstalledDevtronApp.GetReleaseStatus() {
	case serverBean.HelmReleaseStatusSuperseded:
		// If release status is superseded, mark the deployment as failure
		wfr.Status = cdWorkflow2.WorkflowFailed
		wfr.Message = cdWorkflow2.ErrorDeploymentSuperseded.Error()
		wfr.FinishedOn = time.Now()
		return true
	case serverBean.HelmReleaseStatusFailed:
		// If release status is failed, mark the deployment as failure
		wfr.Status = cdWorkflow2.WorkflowFailed
		wfr.Message = helmInstalledDevtronApp.GetDescription()
		wfr.FinishedOn = time.Now()
		return true
	case serverBean.HelmReleaseStatusDeployed:
		//skip if there is no deployment after wfr.StartedOn and continue for next workflow status
		if helmInstalledDevtronApp.GetLastDeployed().AsTime().Before(wfr.StartedOn) {
			impl.logger.Warnw("release mismatched, skipping helm apps status update for this trigger", "appIdentifier", appIdentifier, "err", err)
			return false
		}

		if helmInstalledDevtronApp.GetApplicationStatus() == argoApplication.Healthy {
			// mark the deployment as succeed
			wfr.Status = cdWorkflow2.WorkflowSucceeded
			wfr.FinishedOn = time.Now()
			return true
		}
	}
	if wfr.Status == cdWorkflow2.WorkflowInProgress {
		return false
	}
	wfr.Status = cdWorkflow2.WorkflowInProgress
	return true
}

func (impl *WorkflowDagExecutorImpl) handleAsyncTriggerReleaseError(ctx context.Context, releaseErr error, cdWfr *pipelineConfig.CdWorkflowRunner, overrideRequest *bean.ValuesOverrideRequest) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "WorkflowDagExecutorImpl.handleAsyncTriggerReleaseError")
	defer span.End()
	// for context cancellation due to error.ServerShutDown, the new instance should pick the unfinished process and execute further.
	if releaseErr == nil || errors.Is(context.Cause(newCtx), error2.ServerShutDown) {
		// skipping
		return
	} else if errors.Is(releaseErr, context.DeadlineExceeded) || strings.Contains(releaseErr.Error(), context.DeadlineExceeded.Error()) {
		appIdentifier := triggerAdapter.NewAppIdentifierFromOverrideRequest(overrideRequest)
		if util.IsHelmApp(overrideRequest.DeploymentAppType) {
			// if context deadline is exceeded fetch release status and UpdateWorkflowRunnerStatusForDeployment
			if isWfrUpdated := impl.UpdateWorkflowRunnerStatusForDeployment(appIdentifier, cdWfr, false); !isWfrUpdated {
				// updating cdWfr to failed
				if err := impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(cdWfr, fmt.Errorf("%s: release %s took more than %d mins", "Deployment timeout", appIdentifier.ReleaseName, impl.appServiceConfig.DevtronChartHelmInstallRequestTimeout), overrideRequest.UserId); err != nil {
					impl.logger.Errorw("error while updating current runner status to failed, handleAsyncTriggerReleaseError", "cdWfr", cdWfr.Id, "err", err)
				}
			}
			cdWfr.UpdatedBy = 1
			cdWfr.UpdatedOn = time.Now()
			err := impl.cdWorkflowRunnerService.UpdateCdWorkflowRunnerWithStage(cdWfr)
			if err != nil {
				impl.logger.Errorw("error on update cd workflow runner", "wfr", cdWfr, "err", err)
				return
			}
			if util4.IsRunnerStatusFailed(cdWfr.Status) {
				if cdWfr.Message == cdWorkflow2.ErrorDeploymentSuperseded.Error() {
					dbErr := impl.pipelineStatusTimelineService.MarkPipelineStatusTimelineSuperseded(cdWfr.Id)
					if dbErr != nil {
						impl.logger.Errorw("error updating CdPipelineStatusTimeline", "err", dbErr, "releaseErr", releaseErr)
					}
				} else {
					dbErr := impl.pipelineStatusTimelineService.MarkPipelineStatusTimelineFailed(cdWfr.Id, cdWfr.Message)
					if dbErr != nil {
						impl.logger.Errorw("error updating CdPipelineStatusTimeline", "err", dbErr, "releaseErr", releaseErr)
					}
				}
			}
			if util4.IsTerminalRunnerStatus(cdWfr.Status) {
				appId := cdWfr.CdWorkflow.Pipeline.AppId
				envId := cdWfr.CdWorkflow.Pipeline.EnvironmentId
				envDeploymentConfig, err := impl.deploymentConfigService.GetConfigForDevtronApps(nil, appId, envId)
				if err != nil {
					impl.logger.Errorw("error in fetching environment deployment config by appId and envId", "appId", appId, "envId", envId, "err", err)
				}
				util4.TriggerCDMetrics(cdWorkflow.GetTriggerMetricsFromRunnerObj(cdWfr, envDeploymentConfig), impl.config.ExposeCDMetrics)
			}
			impl.logger.Infow("updated workflow runner status for helm app", "wfr", cdWfr)
		} else {
			// updating cdWfr to failed
			if err := impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(cdWfr, fmt.Errorf("%s: release %s took more than %d mins", "Deployment timeout", appIdentifier.ReleaseName, impl.appServiceConfig.DevtronChartArgoCdInstallRequestTimeout), overrideRequest.UserId); err != nil {
				impl.logger.Errorw("error while updating current runner status to failed, handleAsyncTriggerReleaseError", "cdWfr", cdWfr.Id, "err", err)
			}
		}
		return
	} else if errors.Is(releaseErr, context.Canceled) || strings.Contains(releaseErr.Error(), context.Canceled.Error()) {
		if err := impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(cdWfr, cdWorkflow2.ErrorDeploymentSuperseded, overrideRequest.UserId); err != nil {
			impl.logger.Errorw("error while updating current runner status to failed, handleAsyncTriggerReleaseError", "cdWfr", cdWfr.Id, "err", err)
		}
		return
	} else {
		if err := impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(cdWfr, releaseErr, overrideRequest.UserId); err != nil {
			impl.logger.Errorw("error while updating current runner status to failed, handleAsyncTriggerReleaseError", "cdWfr", cdWfr.Id, "err", err)
		}
		return
	}
}

func (impl *WorkflowDagExecutorImpl) ProcessDevtronAsyncInstallRequest(cdAsyncInstallReq *eventProcessorBean.UserDeploymentRequest, ctx context.Context) error {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "WorkflowDagExecutorImpl.ProcessDevtronAsyncInstallRequest")
	defer span.End()
	overrideRequest := cdAsyncInstallReq.ValuesOverrideRequest
	cdWfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(overrideRequest.WfrId)
	if err != nil {
		impl.logger.Errorw("err on fetching cd workflow runner, processDevtronAsyncHelmInstallRequest", "err", err)
		return err
	}
	if util.IsHelmApp(overrideRequest.DeploymentAppType) {
		impl.logger.Debugw("processing async install request for helm app", "cdWfrId", cdWfr.Id)
		// update workflow runner status, used in app workflow view
		err = impl.cdWorkflowCommonService.UpdateNonTerminalStatusInRunner(newCtx, overrideRequest.WfrId, overrideRequest.UserId, cdWorkflow2.WorkflowStarting)
		if err != nil {
			impl.logger.Errorw("error in updating the workflow runner status, processDevtronAsyncHelmInstallRequest", "cdWfrId", cdWfr.Id, "err", err)
			return err
		}
	}
	envDeploymentConfig, err := impl.deploymentConfigService.GetAndMigrateConfigIfAbsentForDevtronApps(nil, overrideRequest.AppId, overrideRequest.EnvId)
	if err != nil {
		impl.logger.Errorw("error in getting deployment config by appId and envId", "appId", overrideRequest.AppId, "envId", overrideRequest.EnvId, "err", err)
		return err
	}
	releaseId, _, releaseErr := impl.cdHandlerService.TriggerRelease(newCtx, overrideRequest, envDeploymentConfig, cdAsyncInstallReq.TriggeredAt, cdAsyncInstallReq.TriggeredBy)
	if releaseErr != nil {
		impl.logger.Errorw("error encountered in ProcessDevtronAsyncInstallRequest", "err", releaseErr, "cdWfrId", cdWfr.Id)
		impl.handleAsyncTriggerReleaseError(newCtx, releaseErr, cdWfr, overrideRequest)
		return releaseErr
	} else {
		impl.logger.Infow("pipeline triggered successfully !!", "cdPipelineId", overrideRequest.PipelineId, "artifactId", overrideRequest.CiArtifactId, "releaseId", releaseId)
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) handleCiSuccessEvent(triggerContext triggerBean.TriggerContext, artifact *repository.CiArtifact, async bool, triggeredBy int32) error {
	//1. get cd pipelines
	//2. get config
	//3. trigger wf/ deployment
	var pipelineID int
	if artifact.DataSource == repository.POST_CI {
		pipelineID = artifact.ComponentId
	} else {
		// TODO: need to migrate artifact.PipelineId for dataSource="CI_RUNNER" also to component_id
		pipelineID = artifact.PipelineId
	}
	pipelines, err := impl.pipelineRepository.FindByParentCiPipelineId(pipelineID)
	if err != nil {
		impl.logger.Errorw("error in fetching cd pipeline", "pipelineId", artifact.PipelineId, "err", err)
		return err
	}
	for _, pipeline := range pipelines {
		triggerRequest := triggerBean.CdTriggerRequest{
			CdWf:           nil,
			Pipeline:       pipeline,
			Artifact:       artifact,
			TriggeredBy:    triggeredBy,
			TriggerContext: triggerContext,
		}
		err = impl.triggerIfAutoStageCdPipeline(triggerRequest)
		if err != nil {
			impl.logger.Debugw("error on trigger cd pipeline", "err", err)
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) handleWebhookExternalCiEvent(artifact *repository.CiArtifact, triggeredBy int32, externalCiId int, auth func(token string, projectObject string, envObject string) bool, token string) (bool, error) {
	hasAnyTriggered := false
	appWorkflowMappings, err := impl.appWorkflowRepository.FindWFCDMappingByExternalCiId(externalCiId)
	if err != nil {
		impl.logger.Errorw("error in fetching cd pipeline", "pipelineId", artifact.PipelineId, "err", err)
		return hasAnyTriggered, err
	}

	var pipelines []*pipelineConfig.Pipeline
	for _, appWorkflowMapping := range appWorkflowMappings {
		pipeline, err := impl.pipelineRepository.FindById(appWorkflowMapping.ComponentId)
		if err != nil {
			impl.logger.Errorw("error in fetching cd pipeline", "pipelineId", artifact.PipelineId, "err", err)
			return hasAnyTriggered, err
		}
		projectObject := impl.enforcerUtil.GetAppRBACNameByAppId(pipeline.AppId)
		envObject := impl.enforcerUtil.GetAppRBACByAppIdAndPipelineId(pipeline.AppId, pipeline.Id)
		if !auth(token, projectObject, envObject) {
			err = &util.ApiError{Code: "401", HttpStatusCode: 401, UserMessage: "Unauthorized"}
			return hasAnyTriggered, err
		}
		isQualifiedForCdAutoTrigger := helper.IsCdQualifiedForAutoTriggerForWebhookCiEvent(pipeline)
		if !isQualifiedForCdAutoTrigger {
			impl.logger.Warnw("skipping deployment for manual trigger for webhook", "pipeline", pipeline)
			continue
		}
		pipelines = append(pipelines, pipeline)
	}

	for _, pipeline := range pipelines {
		//applyAuth=false, already auth applied for this flow
		triggerRequest := triggerBean.CdTriggerRequest{
			CdWf:        nil,
			Pipeline:    pipeline,
			Artifact:    artifact,
			ApplyAuth:   false,
			TriggeredBy: triggeredBy,
		}
		err = impl.triggerIfAutoStageCdPipeline(triggerRequest)
		if err != nil {
			impl.logger.Debugw("error on trigger cd pipeline", "err", err)
			return hasAnyTriggered, err
		}
		hasAnyTriggered = true
	}

	return hasAnyTriggered, err
}

// if stage is present with 0 stage steps, delete the stage
// handle corrupt data (https://github.com/devtron-labs/devtron/issues/3826)
func (impl *WorkflowDagExecutorImpl) deleteCorruptedPipelineStage(pipelineStage *repository4.PipelineStage, triggeredBy int32) (error, bool) {
	if pipelineStage != nil {
		stageReq := &bean3.PipelineStageDto{
			Id:   pipelineStage.Id,
			Type: pipelineStage.Type,
		}
		err, deleted := impl.pipelineStageService.DeletePipelineStageIfReq(stageReq, triggeredBy)
		if err != nil {
			impl.logger.Errorw("error in deleting the corrupted pipeline stage", "err", err, "pipelineStageReq", stageReq)
			return err, false
		}
		return nil, deleted
	}
	return nil, false
}

func (impl *WorkflowDagExecutorImpl) triggerIfAutoStageCdPipeline(request triggerBean.CdTriggerRequest) error {
	preStage, err := impl.getPipelineStage(request.Pipeline.Id, repository4.PIPELINE_STAGE_TYPE_PRE_CD)
	if err != nil {
		return err
	}

	//handle corrupt data (https://github.com/devtron-labs/devtron/issues/3826)
	err, deleted := impl.deleteCorruptedPipelineStage(preStage, request.TriggeredBy)
	if err != nil {
		impl.logger.Errorw("error in deleteCorruptedPipelineStage ", "cdPipelineId", request.Pipeline.Id, "err", err, "preStage", preStage, "triggeredBy", request.TriggeredBy)
		return err
	}

	request.TriggerContext.Context = context.Background()
	//for auto stage setting triggeredBy to system user no matter where the request came from
	request.TriggeredBy = bean7.SYSTEM_USER_ID
	if len(request.Pipeline.PreStageConfig) > 0 || (preStage != nil && !deleted) {
		// pre stage exists
		if request.Pipeline.PreTriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC {
			impl.logger.Debugw("trigger pre stage for pipeline", "artifactId", request.Artifact.Id, "pipelineId", request.Pipeline.Id)
			_, err = impl.cdHandlerService.TriggerPreStage(request) // TODO handle error here
			return err
		}
	} else if request.Pipeline.TriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC {
		// trigger deployment
		impl.logger.Debugw("trigger cd for pipeline", "artifactId", request.Artifact.Id, "pipelineId", request.Pipeline.Id)
		err = impl.cdHandlerService.TriggerAutomaticDeployment(request)
		return err
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) getPipelineStage(pipelineId int, stageType repository4.PipelineStageType) (*repository4.PipelineStage, error) {
	stage, err := impl.pipelineStageService.GetCdStageByCdPipelineIdAndStageType(pipelineId, stageType, false)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching CD pipeline stage", "cdPipelineId", pipelineId, "stage ", stage, "err", err)
		return nil, err
	}
	return stage, nil
}

func (impl *WorkflowDagExecutorImpl) HandlePreStageSuccessEvent(triggerContext triggerBean.TriggerContext, cdStageCompleteEvent eventProcessorBean.CdStageCompleteEvent) error {
	wfRunner, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(cdStageCompleteEvent.WorkflowRunnerId)
	if err != nil {
		return err
	}
	if wfRunner.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {

		pluginArtifacts := make(map[string][]string)
		if cdStageCompleteEvent.PluginArtifacts != nil {
			pluginArtifacts = cdStageCompleteEvent.PluginArtifacts.GetRegistryToUniqueContainerArtifactDataMapping()
		}
		util4.MergeMaps(pluginArtifacts, cdStageCompleteEvent.PluginRegistryArtifactDetails)

		pipeline, err := impl.pipelineRepository.FindById(cdStageCompleteEvent.CdPipelineId)
		if err != nil {
			return err
		}
		ciArtifact, err := impl.ciArtifactRepository.Get(cdStageCompleteEvent.CiArtifactDTO.Id)
		if err != nil {
			return err
		}
		scanEnabled, scanned := ciArtifact.ScanEnabled, ciArtifact.Scanned
		isScanPluginConfigured, isScanningDoneViaPlugin, err := impl.isArtifactScannedByPluginForPipeline(ciArtifact, cdStageCompleteEvent.CdPipelineId, repository4.PIPELINE_STAGE_TYPE_PRE_CD, bean2.ImageScanningPluginToCheckInPipelineStageStep)
		if err != nil {
			impl.logger.Errorw("error in checking if artifact scanned by plugin for a pipeline or not", "ciArtifact", ciArtifact, "err", err)
			return err
		}
		helper.UpdateScanStatusInCiArtifact(ciArtifact, isScanPluginConfigured, isScanningDoneViaPlugin)

		// if ciArtifact scanEnabled and scanned state changed from above func then update ciArtifact
		if scanEnabled != ciArtifact.ScanEnabled || scanned != ciArtifact.Scanned {
			err = impl.ciArtifactRepository.Update(ciArtifact)
			if err != nil {
				impl.logger.Errorw("error in updating ci artifact after handling scan event for this artifact", "ciArtifact", ciArtifact, "err", err)
				return err
			}
		}
		// Migration of deprecated DataSource Type
		if ciArtifact.IsMigrationRequired() {
			migrationErr := impl.ciArtifactRepository.MigrateToWebHookDataSourceType(ciArtifact.Id)
			if migrationErr != nil {
				impl.logger.Warnw("unable to migrate deprecated DataSource", "artifactId", ciArtifact.Id)
			}
		}
		PreCDArtifacts, err := impl.commonArtifactService.SavePluginArtifacts(ciArtifact, pluginArtifacts, pipeline.Id, repository.PRE_CD, cdStageCompleteEvent.TriggeredBy)
		if err != nil {
			impl.logger.Errorw("error in saving plugin artifacts", "err", err)
			return err
		}
		ciArtifactId := 0
		if len(PreCDArtifacts) > 0 {
			ciArtifactId = PreCDArtifacts[len(PreCDArtifacts)-1].Id // deployment will be trigger with artifact copied by plugin
		} else {
			ciArtifactId = cdStageCompleteEvent.CiArtifactDTO.Id
		}
		err = impl.cdHandlerService.TriggerAutoCDOnPreStageSuccess(triggerContext, cdStageCompleteEvent.CdPipelineId, ciArtifactId, cdStageCompleteEvent.WorkflowId)
		if err != nil {
			impl.logger.Errorw("error in triggering cd on pre cd succcess", "err", err)
			return err
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) HandleDeploymentSuccessEvent(triggerContext triggerBean.TriggerContext, pipelineOverride *chartConfig.PipelineOverride) error {
	if pipelineOverride == nil {
		return fmt.Errorf("invalid request, pipeline override not found")
	}
	cdWorkflow, err := impl.cdWorkflowRepository.FindById(pipelineOverride.CdWorkflowId)
	if err != nil {
		impl.logger.Errorw("error in fetching cd workflow by id", "pipelineOverride", pipelineOverride)
		return err
	}

	postStage, err := impl.getPipelineStage(pipelineOverride.PipelineId, repository4.PIPELINE_STAGE_TYPE_POST_CD)
	if err != nil {
		return err
	}

	//handle corrupt data (https://github.com/devtron-labs/devtron/issues/3826)
	err, deleted := impl.deleteCorruptedPipelineStage(postStage, bean7.SYSTEM_USER_ID)
	if err != nil {
		impl.logger.Errorw("error in deleteCorruptedPipelineStage ", "err", err, "preStage", postStage)
		return err
	}

	if len(pipelineOverride.Pipeline.PostStageConfig) > 0 || (postStage != nil && !deleted) {
		if pipelineOverride.Pipeline.PostTriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC &&
			pipelineOverride.DeploymentType != models.DEPLOYMENTTYPE_STOP &&
			pipelineOverride.DeploymentType != models.DEPLOYMENTTYPE_START {

			triggerRequest := triggerBean.CdTriggerRequest{
				CdWf:                  cdWorkflow,
				Pipeline:              pipelineOverride.Pipeline,
				TriggeredBy:           bean7.SYSTEM_USER_ID,
				TriggerContext:        triggerContext,
				RefCdWorkflowRunnerId: 0,
			}
			triggerRequest.TriggerContext.Context = context.Background()
			_, err = impl.cdHandlerService.TriggerPostStage(triggerRequest)
			if err != nil {
				impl.logger.Errorw("error in triggering post stage after successful deployment event", "err", err, "cdWorkflow", cdWorkflow)
				return err
			}
		}
	} else {
		// to trigger next pre/cd, if any
		// finding children cd by pipeline id
		err = impl.HandlePostStageSuccessEvent(triggerContext, nil, cdWorkflow.Id, pipelineOverride.PipelineId, 1, nil)
		if err != nil {
			impl.logger.Errorw("error in triggering children cd after successful deployment event", "parentCdPipelineId", pipelineOverride.PipelineId)
			return err
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) HandlePostStageSuccessEvent(triggerContext triggerBean.TriggerContext, wfr *bean4.CdWorkflowRunnerDto, cdWorkflowId int, cdPipelineId int, triggeredBy int32, pluginRegistryImageDetails map[string][]string) error {
	// finding children cd by pipeline id
	cdPipelinesMapping, err := impl.appWorkflowRepository.FindWFCDMappingByParentCDPipelineId(cdPipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting mapping of cd pipelines by parent cd pipeline id", "err", err, "parentCdPipelineId", cdPipelineId)
		return err
	}
	ciArtifact, err := impl.ciArtifactRepository.GetArtifactByCdWorkflowId(cdWorkflowId)
	if err != nil {
		impl.logger.Errorw("error in finding artifact by cd workflow id", "err", err, "cdWorkflowId", cdWorkflowId)
		return err
	}
	scanEnabled, scanned := ciArtifact.ScanEnabled, ciArtifact.Scanned
	isScanPluginConfigured, isScanningDoneViaPlugin, err := impl.isArtifactScannedByPluginForPipeline(ciArtifact, cdPipelineId, repository4.PIPELINE_STAGE_TYPE_POST_CD, bean2.ImageScanningPluginToCheckInPipelineStageStep)
	if err != nil {
		impl.logger.Errorw("error in checking if artifact scanned by plugin for a pipeline or not", "ciArtifact", ciArtifact, "err", err)
		return err
	}
	helper.UpdateScanStatusInCiArtifact(ciArtifact, isScanPluginConfigured, isScanningDoneViaPlugin)

	// if ciArtifact scanEnabled and scanned state changed from above func then update ciArtifact
	if scanEnabled != ciArtifact.ScanEnabled || scanned != ciArtifact.Scanned {
		err = impl.ciArtifactRepository.Update(ciArtifact)
		if err != nil {
			impl.logger.Errorw("error in updating ci artifact after handling scan event for this artifact", "ciArtifact", ciArtifact, "err", err)
			return err
		}
	}
	if len(pluginRegistryImageDetails) > 0 {
		PostCDArtifacts, err := impl.commonArtifactService.SavePluginArtifacts(ciArtifact, pluginRegistryImageDetails, cdPipelineId, repository.POST_CD, triggeredBy)
		if err != nil {
			impl.logger.Errorw("error in saving plugin artifacts", "err", err)
			return err
		}
		if len(PostCDArtifacts) > 0 {
			ciArtifact = PostCDArtifacts[0]
		}
	}
	for _, cdPipelineMapping := range cdPipelinesMapping {
		//find pipeline by cdPipeline ID
		pipeline, err := impl.pipelineRepository.FindById(cdPipelineMapping.ComponentId)
		if err != nil {
			impl.logger.Errorw("error in getting cd pipeline by id", "err", err, "pipelineId", cdPipelineMapping.ComponentId)
			return err
		}
		//finding ci artifact by ciPipelineID and pipelineId
		//TODO : confirm values for applyAuth, async & triggeredBy

		triggerRequest := triggerBean.CdTriggerRequest{
			CdWf:           nil,
			Pipeline:       pipeline,
			Artifact:       ciArtifact,
			TriggeredBy:    triggeredBy,
			TriggerContext: triggerContext,
		}

		err = impl.triggerIfAutoStageCdPipeline(triggerRequest)
		if err != nil {
			impl.logger.Errorw("error in triggering cd pipeline after successful post stage", "err", err, "pipelineId", pipeline.Id)
			return err
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) UpdateCiWorkflowForCiSuccess(request *bean2.CiArtifactWebhookRequest) (err error) {
	savedWorkflow, err := impl.ciWorkflowRepository.FindById(*request.WorkflowId)
	if err != nil {
		impl.logger.Errorw("cannot get saved wf", "err", err)
		return err
	}
	// if workflow already cancelled then return, this state arises when user force aborts a ci
	if savedWorkflow.Status == cdWorkflow2.WorkflowCancel {
		return err
	}
	savedWorkflow.Status = string(v1alpha1.NodeSucceeded)
	savedWorkflow.IsArtifactUploaded = workflow.GetArtifactUploadedType(request.IsArtifactUploaded)
	impl.logger.Debugw("updating workflow ", "savedWorkflow", savedWorkflow)
	err = impl.ciService.UpdateCiWorkflowWithStage(savedWorkflow)
	if err != nil {
		impl.logger.Errorw("update wf failed for id ", "err", err)
		return err
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) isArtifactScannedByPluginForPipeline(ciArtifact *repository.CiArtifact, pipelineId int,
	pipelineStage repository4.PipelineStageType, pluginName string) (bool, bool, error) {
	var isScanningDone bool
	isScanPluginConfigured, err := impl.pipelineStageService.IsScanPluginConfiguredAtPipelineStage(pipelineId, pipelineStage, pluginName)
	if err != nil {
		impl.logger.Errorw("error in fetching if a scan plugin is configured or not in a pipeline", "pipelineStage", pipelineStage, "ciArtifact", ciArtifact)
		return false, false, err
	}
	if isScanPluginConfigured {
		isScanningDone, err = impl.imageScanService.IsImageScanExecutionCompleted(ciArtifact.Image, ciArtifact.ImageDigest)
		if err != nil {
			impl.logger.Errorw("error in checking if image scanning is completed or not", "image", ciArtifact.Image, "imageDigest", ciArtifact.ImageDigest)
			return false, false, err
		}
	}
	return isScanPluginConfigured, isScanningDone, nil
}

func (impl *WorkflowDagExecutorImpl) HandleCiSuccessEvent(triggerContext triggerBean.TriggerContext, ciPipelineId int, request *bean2.CiArtifactWebhookRequest, imagePushedAt time.Time) (id int, err error) {
	impl.logger.Infow("webhook for artifact save", "req", request)
	pipelineModal, err := impl.ciPipelineRepository.FindByCiAndAppDetailsById(ciPipelineId)
	if err != nil {
		impl.logger.Errorw("unable to find pipelineModal", "ciPipelineId", ciPipelineId, "err", err)
		return 0, err
	}
	if request.PipelineName == "" {
		request.PipelineName = pipelineModal.Name
	}
	materialJson, err := helper.GetMaterialInfoJson(request.MaterialInfo)
	if err != nil {
		impl.logger.Errorw("unable to get materialJson", "materialInfo", request.MaterialInfo, "err", err)
		return 0, err
	}
	createdOn := time.Now()
	updatedOn := time.Now()
	if !imagePushedAt.IsZero() {
		createdOn = imagePushedAt
	}
	buildArtifact := adaptor.GetBuildArtifact(request, pipelineModal.Id, materialJson, createdOn, updatedOn)

	// image scanning plugin can only be applied in Post-ci, scanning in pre-ci doesn't make sense
	pipelineStage := repository4.PIPELINE_STAGE_TYPE_POST_CI
	if pipelineModal.PipelineType == buildCommonBean.CI_JOB.ToString() {
		pipelineStage = repository4.PIPELINE_STAGE_TYPE_PRE_CI
	}
	// this flag comes from ci-runner when scanning is enabled from ciPipeline modal
	if request.IsScanEnabled {
		buildArtifact.Scanned = true
		buildArtifact.ScanEnabled = true
	} else {
		isScanPluginConfigured, isScanningDoneViaPlugin, err := impl.isArtifactScannedByPluginForPipeline(buildArtifact, pipelineModal.Id, pipelineStage, bean2.ImageScanningPluginToCheckInPipelineStageStep)
		if err != nil {
			impl.logger.Errorw("error in checking if artifact scanned by plugin for a pipeline or not", "ciArtifact", buildArtifact, "err", err)
			return 0, err
		}
		helper.UpdateScanStatusInCiArtifact(buildArtifact, isScanPluginConfigured, isScanningDoneViaPlugin)
	}

	if err = impl.ciArtifactRepository.Save(buildArtifact); err != nil {
		impl.logger.Errorw("error in saving material", "err", err)
		return 0, err
	}

	var pluginArtifacts []*repository.CiArtifact
	for registry, artifacts := range request.PluginRegistryArtifactDetails {
		for _, image := range artifacts {
			if pipelineModal.PipelineType == string(buildCommonBean.CI_JOB) && image == "" {
				continue
			}
			pluginArtifact := &repository.CiArtifact{
				Image:                 image,
				ImageDigest:           request.ImageDigest,
				MaterialInfo:          string(materialJson),
				DataSource:            request.PluginArtifactStage,
				ComponentId:           pipelineModal.Id,
				PipelineId:            pipelineModal.Id,
				AuditLog:              sql.AuditLog{CreatedBy: request.UserId, UpdatedBy: request.UserId, CreatedOn: createdOn, UpdatedOn: updatedOn},
				CredentialsSourceType: repository.GLOBAL_CONTAINER_REGISTRY,
				CredentialSourceValue: registry,
				ParentCiArtifact:      buildArtifact.Id,
				Scanned:               buildArtifact.Scanned,
				ScanEnabled:           buildArtifact.ScanEnabled,
				IsArtifactUploaded:    request.IsArtifactUploaded, // for backward compatibility
			}
			pluginArtifacts = append(pluginArtifacts, pluginArtifact)
		}
	}
	if len(pluginArtifacts) > 0 {
		_, err = impl.ciArtifactRepository.SaveAll(pluginArtifacts)
		if err != nil {
			impl.logger.Errorw("error while saving ci artifacts", "err", err)
			return 0, err
		}
	}

	childrenCi, err := impl.ciPipelineRepository.FindByParentCiPipelineId(ciPipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error while fetching childern ci ", "err", err)
		return 0, err
	}

	var ciArtifactArr []*repository.CiArtifact
	for _, ci := range childrenCi {
		ciArtifact := &repository.CiArtifact{
			Image:              request.Image,
			ImageDigest:        request.ImageDigest,
			MaterialInfo:       string(materialJson),
			DataSource:         request.DataSource,
			PipelineId:         ci.Id,
			ParentCiArtifact:   buildArtifact.Id,
			IsArtifactUploaded: request.IsArtifactUploaded, // for backward compatibility
			ScanEnabled:        buildArtifact.ScanEnabled,
			Scanned:            false,
			TargetPlatforms:    utils.ConvertTargetPlatformListToString(request.TargetPlatforms),
			AuditLog:           sql.AuditLog{CreatedBy: request.UserId, UpdatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now()},
		}
		if buildArtifact.ScanEnabled {
			ciArtifact.Scanned = buildArtifact.Scanned
		}
		ciArtifactArr = append(ciArtifactArr, ciArtifact)
	}

	impl.logger.Debugw("saving ci artifacts", "art", ciArtifactArr)
	if len(ciArtifactArr) > 0 {
		_, err = impl.ciArtifactRepository.SaveAll(ciArtifactArr)
		if err != nil {
			impl.logger.Errorw("error while saving ci artifacts", "err", err)
			return 0, err
		}
	}
	if len(pluginArtifacts) == 0 {
		ciArtifactArr = append(ciArtifactArr, buildArtifact)
	} else {
		ciArtifactArr = append(ciArtifactArr, pluginArtifacts[0])
	}
	runnableFunc := func() {
		impl.WriteCiSuccessEvent(request, pipelineModal, buildArtifact)
	}
	impl.asyncRunnable.Execute(runnableFunc)
	async := false

	// execute auto trigger in batch on CI success event
	totalCIArtifactCount := len(ciArtifactArr)
	batchSize := impl.ciConfig.CIAutoTriggerBatchSize
	// handling to avoid infinite loop
	if batchSize <= 0 {
		batchSize = 1
	}
	start := time.Now()
	impl.logger.Infow("Started: auto trigger for children Stage/CD pipelines", "Artifact count", totalCIArtifactCount)
	for i := 0; i < totalCIArtifactCount; {
		// requests left to process
		remainingBatch := totalCIArtifactCount - i
		if remainingBatch < batchSize {
			batchSize = remainingBatch
		}
		var wg sync.WaitGroup
		for j := 0; j < batchSize; j++ {
			wg.Add(1)
			index := i + j
			runnableFunc := func(index int) {
				defer wg.Done()
				ciArtifact := ciArtifactArr[index]
				// handle individual CiArtifact success event
				err = impl.handleCiSuccessEvent(triggerContext, ciArtifact, async, request.UserId)
				if err != nil {
					impl.logger.Errorw("error on handle  ci success event", "ciArtifactId", ciArtifact.Id, "err", err)
				}
			}
			impl.asyncRunnable.Execute(func() { runnableFunc(index) })
		}
		wg.Wait()
		i += batchSize
	}
	impl.logger.Debugw("Completed: auto trigger for children Stage/CD pipelines", "Time taken", time.Since(start).Seconds())
	return buildArtifact.Id, err
}

func (impl *WorkflowDagExecutorImpl) WriteCiSuccessEvent(request *bean2.CiArtifactWebhookRequest, pipeline *pipelineConfig.CiPipeline, artifact *repository.CiArtifact) {
	event, _ := impl.eventFactory.Build(util2.Success, &pipeline.Id, pipeline.AppId, nil, util2.CI)
	event.CiArtifactId = artifact.Id
	if artifact.WorkflowId != nil {
		event.CiWorkflowRunnerId = *artifact.WorkflowId
	}
	event.UserId = int(request.UserId)
	event = impl.eventFactory.BuildExtraCIData(event, nil)
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("error in writing event", "err", evtErr)
	}
}

func (impl *WorkflowDagExecutorImpl) HandleCiStepFailedEvent(ciPipelineId int, request *bean2.CiArtifactWebhookRequest) (err error) {
	if request == nil || request.WorkflowId == nil {
		impl.logger.Errorw("invalid request received", "request", request)
		return fmt.Errorf("invalid request received! workflowId is required")
	}
	savedWorkflow, err := impl.ciWorkflowRepository.FindById(*request.WorkflowId)
	if err != nil {
		impl.logger.Errorw("cannot get saved wf", "wf ID: ", *request.WorkflowId, "err", err)
		return err
	}
	// update IsArtifactUploaded flag in workflow
	dbErr := impl.ciWorkflowRepository.UpdateArtifactUploaded(savedWorkflow.Id, workflow.GetArtifactUploadedType(request.IsArtifactUploaded))
	if dbErr != nil {
		impl.logger.Errorw("update workflow status", "ciWorkflowId", savedWorkflow.Id, "err", dbErr)
	}
	pipelineModel, err := impl.ciPipelineRepository.FindByCiAndAppDetailsById(ciPipelineId)
	if err != nil {
		impl.logger.Errorw("unable to find pipeline", "ID", ciPipelineId, "err", err)
		return err
	}
	customTagServiceRunnableFunc := func() {
		if len(savedWorkflow.ImagePathReservationIds) > 0 {
			err = impl.customTagService.DeactivateImagePathReservationByImageIds(savedWorkflow.ImagePathReservationIds)
			if err != nil {
				impl.logger.Errorw("unable to deactivate ImagePathReservation", "imagePathReservationIds", savedWorkflow.ImagePathReservationIds, "err", err)
			}
		}
	}
	impl.asyncRunnable.Execute(customTagServiceRunnableFunc)
	if request.FailureReason != workFlow.CiFailed.String() {
		notificationServiceRunnableFunc := func() {
			impl.WriteCiStepFailedEvent(pipelineModel, request, savedWorkflow)
		}
		impl.asyncRunnable.Execute(notificationServiceRunnableFunc)
	} else {
		// this case has been handled CiHandlerImpl.UpdateWorkflow function.
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) WriteCiStepFailedEvent(pipeline *pipelineConfig.CiPipeline, request *bean2.CiArtifactWebhookRequest, ciWorkflow *pipelineConfig.CiWorkflow) {
	event, _ := impl.eventFactory.Build(util2.Fail, &pipeline.Id, pipeline.AppId, nil, util2.CI)
	material := &bean5.MaterialTriggerInfo{}
	material.GitTriggers = ciWorkflow.GitTriggers
	event.CiWorkflowRunnerId = ciWorkflow.Id
	event.UserId = int(ciWorkflow.TriggeredBy)
	event = impl.eventFactory.BuildExtraCIData(event, material)
	event.CiArtifactId = 0
	event.Payload.FailureReason = request.FailureReason
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("error in writing event: ", event, "error: ", evtErr)
	}
}

func (impl *WorkflowDagExecutorImpl) HandleExternalCiWebhook(externalCiId int, request *bean2.CiArtifactWebhookRequest,
	auth func(token string, projectObject string, envObject string) bool, token string) (id int, err error) {
	externalCiPipeline, err := impl.ciPipelineRepository.FindExternalCiById(externalCiId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching external ci", "err", err)
		return 0, err
	}
	if externalCiPipeline.Id == 0 {
		impl.logger.Errorw("invalid external ci id", "externalCiId", externalCiId, "err", err)
		return 0, &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "invalid external ci id"}
	}

	impl.logger.Infow("request of webhook external ci", "req", request)
	materialJson, err := request.MaterialInfo.MarshalJSON()
	if err != nil {
		impl.logger.Errorw("unable to marshal material metadata", "err", err)
		return 0, err
	}
	dst := new(bytes.Buffer)
	err = json.Compact(dst, materialJson)
	if err != nil {
		impl.logger.Errorw("parsing error", "err", err)
		return 0, err
	}
	materialJson = dst.Bytes()
	artifact := &repository.CiArtifact{
		Image:                request.Image,
		ImageDigest:          request.ImageDigest,
		MaterialInfo:         string(materialJson),
		DataSource:           request.DataSource,
		WorkflowId:           request.WorkflowId,
		ExternalCiPipelineId: externalCiId,
		ScanEnabled:          false,
		Scanned:              false,
		AuditLog:             sql.AuditLog{CreatedBy: request.UserId, UpdatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now()},
	}
	if err = impl.ciArtifactRepository.Save(artifact); err != nil {
		impl.logger.Errorw("error in saving material", "err", err)
		return 0, err
	}

	hasAnyTriggered, err := impl.handleWebhookExternalCiEvent(artifact, request.UserId, externalCiId, auth, token)
	if err != nil {
		impl.logger.Errorw("error on handle ext ci webhook", "err", err)
		// if none of the child node has been triggered
		if !hasAnyTriggered {
			if err1 := impl.ciArtifactRepository.Delete(artifact); err1 != nil {
				impl.logger.Errorw("error in rollback artifact", "err", err1)
				return 0, err1
			}
		}
	}
	return artifact.Id, err
}

// TODO: move in adapter
func (impl *WorkflowDagExecutorImpl) BuildCiArtifactRequestForWebhook(event pipeline.ExternalCiWebhookDto) (*bean2.CiArtifactWebhookRequest, error) {
	ciMaterialInfos := make([]repository.CiMaterialInfo, 0)
	if event.MaterialType == "" {
		event.MaterialType = "git"
	}
	for _, p := range event.CiProjectDetails {
		var modifications []repository.Modification

		var branch string
		var tag string
		var webhookData repository.WebhookData
		if p.SourceType == constants.SOURCE_TYPE_BRANCH_FIXED {
			branch = p.SourceValue
		} else if p.SourceType == constants.SOURCE_TYPE_WEBHOOK {
			webhookData = repository.WebhookData{
				Id:              p.WebhookData.Id,
				EventActionType: p.WebhookData.EventActionType,
				Data:            p.WebhookData.Data,
			}
		}

		modification := repository.Modification{
			Revision:     p.CommitHash,
			ModifiedTime: p.CommitTime,
			Author:       p.Author,
			Branch:       branch,
			Tag:          tag,
			WebhookData:  webhookData,
			Message:      p.Message,
		}

		modifications = append(modifications, modification)
		ciMaterialInfo := repository.CiMaterialInfo{
			Material: repository.Material{
				GitConfiguration: repository.GitConfiguration{
					URL: p.GitRepository,
				},
				Type: event.MaterialType,
			},
			Changed:       true,
			Modifications: modifications,
		}
		ciMaterialInfos = append(ciMaterialInfos, ciMaterialInfo)
	}

	materialBytes, err := json.Marshal(ciMaterialInfos)
	if err != nil {
		impl.logger.Errorw("cannot build ci artifact req", "err", err)
		return nil, err
	}
	rawMaterialInfo := json.RawMessage(materialBytes)
	fmt.Printf("Raw Message : %s\n", rawMaterialInfo)

	if event.TriggeredBy == 0 {
		event.TriggeredBy = 1 // system triggered event
	}

	request := &bean2.CiArtifactWebhookRequest{
		Image:              event.DockerImage,
		ImageDigest:        event.Digest,
		DataSource:         event.DataSource,
		PipelineName:       event.PipelineName,
		MaterialInfo:       rawMaterialInfo,
		UserId:             event.TriggeredBy,
		WorkflowId:         event.WorkflowId,
		IsArtifactUploaded: event.IsArtifactUploaded,
	}
	// if DataSource is empty, repository.WEBHOOK is considered as default
	if request.DataSource == "" {
		request.DataSource = repository.WEBHOOK
	}
	return request, nil
}

func (impl *WorkflowDagExecutorImpl) UpdateWorkflowRunnerStatusForFluxDeployment(appIdentifier *bean8.FluxAppIdentifier, wfr *pipelineConfig.CdWorkflowRunner,
	pipelineOverride *chartConfig.PipelineOverride) bool {
	fluxAppDetail, err := impl.fluxApplicationService.GetFluxAppDetail(context.Background(), appIdentifier)
	if err != nil {
		impl.logger.Errorw("error in getting helm app release status", "appIdentifier", appIdentifier, "err", err)
		// Handle release not found errors
		// If release not found, mark the deployment as failure
		wfr.Status = cdWorkflow2.WorkflowUnableToFetchState
		wfr.Message = fmt.Sprintf("error in fetching app detail; err: %s", err.Error())
		wfr.FinishedOn = time.Now()
		return true
	}
	if !impl.checkIfFluxPipelineEventIsValid(fluxAppDetail.LastObservedVersion, pipelineOverride) {
		return false
	}
	wfr.FinishedOn = time.Now()
	wfr.Message = fluxAppDetail.FluxAppStatusDetail.Message
	switch fluxAppDetail.FluxAppStatusDetail.Reason {
	case bean8.InstallSucceededReason, bean8.UpgradeSucceededReason, bean8.TestSucceededReason, bean8.RollbackSucceededReason:
		if fluxAppDetail.AppHealthStatus == commonBean.HealthStatusHealthy {
			wfr.Status = cdWorkflow2.WorkflowSucceeded
		}
	case bean8.UpgradeFailedReason,
		bean8.TestFailedReason,
		bean8.RollbackFailedReason,
		bean8.UninstallFailedReason,
		bean8.ArtifactFailedReason,
		bean8.InstallFailedReason:
		wfr.Status = cdWorkflow2.WorkflowFailed
	}
	return true
}

func (impl *WorkflowDagExecutorImpl) checkIfFluxPipelineEventIsValid(lastObservedVersion string, pipelineOverride *chartConfig.PipelineOverride) bool {
	gitHash := getShortHash(lastObservedVersion)
	if !strings.HasPrefix(pipelineOverride.GitHash, gitHash) {
		pipelineOverrideByHash, err := impl.pipelineOverrideRepository.FindByPipelineLikeTriggerGitHash(gitHash)
		if err != nil {
			impl.logger.Errorw("error on update application status", "gitHash", gitHash, "err", err)
			return false
		}
		if pipelineOverrideByHash == nil || pipelineOverrideByHash.CommitTime.Before(pipelineOverride.CommitTime) {
			// we have received trigger hash which is committed before this apps actual gitHash stored by us
			// this means that the hash stored by us will be synced later, so we will drop this event
			return false
		}
	}
	return true
}

// getShortHash gets the short Git hash embedded in the version string
// with the beginning of the full Git commit hash.
//
// version: expected format like "4.22.1+<shortHash>.<buildNumber>"
// fullHash: expected to be a full 40-character Git commit SHA
func getShortHash(version string) string {
	// Split version string at '+' to extract metadata
	parts := strings.Split(version, "+")
	if len(parts) < 2 {
		return "" // No metadata found
	}

	// Metadata might look like "2b6c6b2.2" → short hash + build number
	metaParts := strings.Split(parts[1], ".")
	shortHash := metaParts[0] // Take only the short hash before the dot

	// Compare short hash with prefix of full hash
	return shortHash
}
