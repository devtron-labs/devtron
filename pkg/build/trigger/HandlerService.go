package trigger

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/common-lib/async"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"github.com/devtron-labs/common-lib/utils"
	bean4 "github.com/devtron-labs/common-lib/utils/bean"
	"github.com/devtron-labs/common-lib/utils/k8s"
	commonBean "github.com/devtron-labs/common-lib/workflow"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/constants"
	repository5 "github.com/devtron-labs/devtron/internal/sql/repository"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/attributes"
	bean3 "github.com/devtron-labs/devtron/pkg/attributes/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	bean6 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/bean/common"
	"github.com/devtron-labs/devtron/pkg/build/pipeline"
	buildBean "github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	buildCommonBean "github.com/devtron-labs/devtron/pkg/build/pipeline/bean/common"
	"github.com/devtron-labs/devtron/pkg/build/trigger/adaptor"
	"github.com/devtron-labs/devtron/pkg/cluster"
	adapter2 "github.com/devtron-labs/devtron/pkg/cluster/adapter"
	clusterBean "github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/environment"
	repository6 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	eventProcessorBean "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/devtron-labs/devtron/pkg/executor"
	pipeline2 "github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/adapter"
	pipelineConfigBean "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	constants2 "github.com/devtron-labs/devtron/pkg/pipeline/constants"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	util2 "github.com/devtron-labs/devtron/pkg/pipeline/util"
	"github.com/devtron-labs/devtron/pkg/plugin"
	bean2 "github.com/devtron-labs/devtron/pkg/plugin/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/variables"
	repository4 "github.com/devtron-labs/devtron/pkg/variables/repository"
	auditService "github.com/devtron-labs/devtron/pkg/workflow/trigger/audit/service"
	"github.com/devtron-labs/devtron/util/sliceUtil"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

type HandlerService interface {
	HandlePodDeleted(ciWorkflow *pipelineConfig.CiWorkflow)
	CheckAndReTriggerCI(workflowStatus eventProcessorBean.CiCdStatus) error
	HandleCIManual(ciTriggerRequest bean.CiTriggerRequest) (int, error)
	HandleCIWebhook(gitCiTriggerRequest bean.GitCiTriggerRequest) (int, error)

	StartCiWorkflowAndPrepareWfRequest(trigger *types.CiTriggerRequest) (map[string]string, *pipelineConfig.CiWorkflow, *types.WorkflowRequest, error)

	CancelBuild(workflowId int, forceAbort bool) (int, error)
	GetRunningWorkflowLogs(workflowId int, followLogs bool) (*bufio.Reader, func() error, error)
	GetHistoricBuildLogs(workflowId int, ciWorkflow *pipelineConfig.CiWorkflow) (map[string]string, error)
	DownloadCiWorkflowArtifacts(pipelineId int, buildId int) (*os.File, error)
	abortPreviousRunningBuilds(pipelineId int, triggeredBy int32) error
}

// CATEGORY=CI_BUILDX
type BuildxGlobalFlags struct {
	BuildxCacheModeMin         bool `env:"BUILDX_CACHE_MODE_MIN" envDefault:"false" description:"To set build cache mode to minimum in buildx" `
	AsyncBuildxCacheExport     bool `env:"ASYNC_BUILDX_CACHE_EXPORT" envDefault:"false" description:"To enable async container image cache export"`
	BuildxInterruptionMaxRetry int  `env:"BUILDX_INTERRUPTION_MAX_RETRY" envDefault:"3" description:"Maximum number of retries for buildx builder interruption"`
}

type HandlerServiceImpl struct {
	Logger                       *zap.SugaredLogger
	workflowService              executor.WorkflowService
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
	ciArtifactRepository         repository5.CiArtifactRepository
	pipelineStageService         pipeline2.PipelineStageService
	userService                  user.UserService
	ciTemplateService            pipeline.CiTemplateReadService
	appCrudOperationService      app.AppCrudOperationService
	envRepository                repository6.EnvironmentRepository
	appRepository                appRepository.AppRepository
	customTagService             pipeline2.CustomTagService
	config                       *types.CiConfig
	scopedVariableManager        variables.ScopedVariableManager
	ciCdPipelineOrchestrator     pipeline2.CiCdPipelineOrchestrator
	buildxGlobalFlags            *BuildxGlobalFlags
	attributeService             attributes.AttributesService
	pluginInputVariableParser    pipeline2.PluginInputVariableParser
	globalPluginService          plugin.GlobalPluginService
	ciService                    pipeline2.CiService
	ciWorkflowRepository         pipelineConfig.CiWorkflowRepository
	gitSensorClient              gitSensor.Client
	ciLogService                 pipeline2.CiLogService
	blobConfigStorageService     pipeline2.BlobStorageConfigService
	clusterService               cluster.ClusterService
	envService                   environment.EnvironmentService
	K8sUtil                      *k8s.K8sServiceImpl
	asyncRunnable                *async.Runnable
	workflowTriggerAuditService  auditService.WorkflowTriggerAuditService
}

func NewHandlerServiceImpl(Logger *zap.SugaredLogger, workflowService executor.WorkflowService,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	ciArtifactRepository repository5.CiArtifactRepository,
	pipelineStageService pipeline2.PipelineStageService,
	userService user.UserService,
	ciTemplateService pipeline.CiTemplateReadService,
	appCrudOperationService app.AppCrudOperationService,
	envRepository repository6.EnvironmentRepository,
	appRepository appRepository.AppRepository,
	scopedVariableManager variables.ScopedVariableManager,
	customTagService pipeline2.CustomTagService,
	ciCdPipelineOrchestrator pipeline2.CiCdPipelineOrchestrator, attributeService attributes.AttributesService,
	pluginInputVariableParser pipeline2.PluginInputVariableParser,
	globalPluginService plugin.GlobalPluginService,
	ciService pipeline2.CiService,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	gitSensorClient gitSensor.Client,
	ciLogService pipeline2.CiLogService,
	blobConfigStorageService pipeline2.BlobStorageConfigService,
	clusterService cluster.ClusterService,
	envService environment.EnvironmentService,
	K8sUtil *k8s.K8sServiceImpl,
	asyncRunnable *async.Runnable,
	workflowTriggerAuditService auditService.WorkflowTriggerAuditService,
) *HandlerServiceImpl {
	buildxCacheFlags := &BuildxGlobalFlags{}
	err := env.Parse(buildxCacheFlags)
	if err != nil {
		Logger.Infow("error occurred while parsing BuildxGlobalFlags env,so setting BuildxCacheModeMin and AsyncBuildxCacheExport to default value", "err", err)
	}
	cis := &HandlerServiceImpl{
		Logger:                       Logger,
		workflowService:              workflowService,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		ciPipelineRepository:         ciPipelineRepository,
		ciArtifactRepository:         ciArtifactRepository,
		pipelineStageService:         pipelineStageService,
		userService:                  userService,
		ciTemplateService:            ciTemplateService,
		appCrudOperationService:      appCrudOperationService,
		envRepository:                envRepository,
		appRepository:                appRepository,
		scopedVariableManager:        scopedVariableManager,
		customTagService:             customTagService,
		ciCdPipelineOrchestrator:     ciCdPipelineOrchestrator,
		buildxGlobalFlags:            buildxCacheFlags,
		attributeService:             attributeService,
		pluginInputVariableParser:    pluginInputVariableParser,
		globalPluginService:          globalPluginService,
		ciService:                    ciService,
		ciWorkflowRepository:         ciWorkflowRepository,
		gitSensorClient:              gitSensorClient,
		ciLogService:                 ciLogService,
		blobConfigStorageService:     blobConfigStorageService,
		clusterService:               clusterService,
		envService:                   envService,
		K8sUtil:                      K8sUtil,
		asyncRunnable:                asyncRunnable,
		workflowTriggerAuditService:  workflowTriggerAuditService,
	}
	config, err := types.GetCiConfig()
	if err != nil {
		return nil
	}
	cis.config = config
	return cis
}

func (impl *HandlerServiceImpl) HandlePodDeleted(ciWorkflow *pipelineConfig.CiWorkflow) {
	if !impl.config.WorkflowRetriesEnabled() {
		impl.Logger.Debug("ci workflow retry feature disabled")
		return
	}
	retryCount, refCiWorkflow, err := impl.getRefWorkflowAndCiRetryCount(ciWorkflow)
	if err != nil {
		impl.Logger.Errorw("error in getRefWorkflowAndCiRetryCount", "ciWorkflowId", ciWorkflow.Id, "err", err)
	}
	impl.Logger.Infow("re-triggering ci by UpdateCiWorkflowStatusFailedCron", "refCiWorkflowId", refCiWorkflow.Id, "ciWorkflow.Status", ciWorkflow.Status, "ciWorkflow.Message", ciWorkflow.Message, "retryCount", retryCount)
	err = impl.reTriggerCi(retryCount, refCiWorkflow)
	if err != nil {
		impl.Logger.Errorw("error in reTriggerCi", "ciWorkflowId", refCiWorkflow.Id, "workflowStatus", ciWorkflow.Status, "ciWorkflowMessage", "ciWorkflow.Message", "retryCount", retryCount, "err", err)
	}
}

func (impl *HandlerServiceImpl) CheckAndReTriggerCI(workflowStatus eventProcessorBean.CiCdStatus) error {

	// return if re-trigger feature is disabled
	if !impl.config.WorkflowRetriesEnabled() {
		impl.Logger.Debug("CI re-trigger is disabled")
		return nil
	}

	status, message, ciWorkFlow, err := impl.extractPodStatusAndWorkflow(workflowStatus)
	if err != nil {
		impl.Logger.Errorw("error in extractPodStatusAndWorkflow", "err", err)
		return err
	}

	if !executors.CheckIfReTriggerRequired(status, message, ciWorkFlow.Status) {
		impl.Logger.Debugw("not re-triggering ci", "status", status, "message", message, "ciWorkflowStatus", ciWorkFlow.Status)
		return nil
	}

	impl.Logger.Debugw("re-triggering ci", "status", status, "message", message, "ciWorkflowStatus", ciWorkFlow.Status, "ciWorkFlowId", ciWorkFlow.Id)

	retryCount, refCiWorkflow, err := impl.getRefWorkflowAndCiRetryCount(ciWorkFlow)
	if err != nil {
		impl.Logger.Errorw("error while getting retry count value for a ciWorkflow", "ciWorkFlowId", ciWorkFlow.Id)
		return err
	}

	err = impl.reTriggerCi(retryCount, refCiWorkflow)
	if err != nil {
		impl.Logger.Errorw("error in reTriggerCi", "err", err, "status", status, "message", message, "retryCount", retryCount, "ciWorkFlowId", ciWorkFlow.Id)
	}
	return err
}

func (impl *HandlerServiceImpl) reTriggerCi(retryCount int, refCiWorkflow *pipelineConfig.CiWorkflow) error {
	if retryCount >= impl.config.MaxCiWorkflowRetries {
		impl.Logger.Infow("maximum retries exhausted for this ciWorkflow", "ciWorkflowId", refCiWorkflow.Id, "retries", retryCount, "configuredRetries", impl.config.MaxCiWorkflowRetries)
		return nil
	}
	impl.Logger.Infow("re-triggering ci for a ci workflow", "ReferenceCiWorkflowId", refCiWorkflow.Id)

	// Try to use stored workflow config snapshot for retrigger
	err := impl.reTriggerCiFromSnapshot(refCiWorkflow)
	if err != nil {
		impl.Logger.Errorw("failed to retrigger from snapshot", "ciWorkflowId", refCiWorkflow.Id, "err", err)
		return err
	}
	return nil
}

// reTriggerCiFromSnapshot attempts to retrigger CI using a stored workflow config snapshot
func (impl *HandlerServiceImpl) reTriggerCiFromSnapshot(refCiWorkflow *pipelineConfig.CiWorkflow) error {
	impl.Logger.Infow("attempting to retrigger CI from stored snapshot", "refCiWorkflowId", refCiWorkflow.Id)

	// Retrieve workflow request from snapshot
	workflowRequest, err := impl.workflowTriggerAuditService.GetWorkflowRequestFromSnapshotForRetrigger(refCiWorkflow.Id, types.CI_WORKFLOW_TYPE)
	if err != nil {
		impl.Logger.Errorw("error retrieving workflow request from snapshot", "ciWorkflowId", refCiWorkflow.Id, "err", err)
		return err
	}
	//create a new CI workflow entry for retrigger
	newCiWorkflow, err := impl.createNewCiWorkflowForRetrigger(refCiWorkflow)
	if err != nil {
		impl.Logger.Errorw("error creating new CI workflow for retrigger", "refCiWorkflowId", refCiWorkflow.Id, "err", err)
		return err
	}

	impl.updateWorkflowRequestForRetrigger(workflowRequest, newCiWorkflow)
	ciPipelineMaterialIds := make([]int, 0, len(refCiWorkflow.GitTriggers))
	for id, _ := range refCiWorkflow.GitTriggers {
		ciPipelineMaterialIds = append(ciPipelineMaterialIds, id)
	}
	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByIdsIncludeDeleted(ciPipelineMaterialIds)
	if err != nil {
		impl.Logger.Errorw("error in getting ci Pipeline Materials using ciPipeline Material Ids", "ciPipelineMaterialIds", ciPipelineMaterialIds, "err", err)
		return err
	}

	trigger := &types.CiTriggerRequest{IsRetrigger: true, RetriggerWorkflowRequest: workflowRequest, RetriggerCiWorkflow: newCiWorkflow}
	trigger.BuildTriggerObject(refCiWorkflow, ciMaterials, bean6.SYSTEM_USER_ID, true, nil, "")

	_, err = impl.triggerCiPipeline(trigger)
	if err != nil {
		impl.Logger.Errorw("error occurred in re-triggering ciWorkflow", "triggerDetails", trigger, "err", err)
		return err
	}

	impl.Logger.Infow("successfully retriggered CI from snapshot", "originalCiWorkflowId", refCiWorkflow.Id, "newCiWorkflowId", newCiWorkflow.Id)
	return nil
}

// createNewCiWorkflowForRetrigger creates a new CI workflow entry for retrigger
func (impl *HandlerServiceImpl) createNewCiWorkflowForRetrigger(refCiWorkflow *pipelineConfig.CiWorkflow) (*pipelineConfig.CiWorkflow, error) {
	newCiWorkflow := adaptor.GetCiWorkflowFromRefCiWorkflow(refCiWorkflow, cdWorkflow.WorkflowStarting, bean6.SYSTEM_USER_ID)
	err := impl.ciService.SaveCiWorkflowWithStage(newCiWorkflow)
	if err != nil {
		impl.Logger.Errorw("error saving new CI workflow for retrigger", "newCiWorkflow", newCiWorkflow, "err", err)
		return nil, err
	}
	return newCiWorkflow, nil
}

// updateWorkflowRequestForRetrigger updates dynamic fields in workflow request for retrigger, like workflowId,WorkflowNamePrefix and triggeredBy
func (impl *HandlerServiceImpl) updateWorkflowRequestForRetrigger(workflowRequest *types.WorkflowRequest, newCiWorkflow *pipelineConfig.CiWorkflow) {
	// Update only the dynamic fields that need to change for retrigger
	workflowRequest.WorkflowId = newCiWorkflow.Id
	workflowRequest.WorkflowNamePrefix = fmt.Sprintf("%d-%s", newCiWorkflow.Id, newCiWorkflow.Name)
	workflowRequest.TriggeredBy = bean6.SYSTEM_USER_ID
	// Keep all other fields from snapshot (CI project details, build configs, etc.) unchanged
	// This ensures we use the exact same configuration that was used during the original trigger
}

func (impl *HandlerServiceImpl) HandleCIManual(ciTriggerRequest bean.CiTriggerRequest) (int, error) {
	impl.Logger.Debugw("HandleCIManual for pipeline ", "PipelineId", ciTriggerRequest.PipelineId)
	commitHashes, runtimeParams, err := impl.buildManualTriggerCommitHashes(ciTriggerRequest)
	if err != nil {
		return 0, err
	}

	ciArtifact, err := impl.ciArtifactRepository.GetLatestArtifactTimeByCiPipelineId(ciTriggerRequest.PipelineId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("Error in GetLatestArtifactTimeByCiPipelineId", "err", err, "pipelineId", ciTriggerRequest.PipelineId)
		return 0, err
	}

	createdOn := time.Time{}
	if err != pg.ErrNoRows {
		createdOn = ciArtifact.CreatedOn
	}

	trigger := &types.CiTriggerRequest{
		PipelineId:          ciTriggerRequest.PipelineId,
		CommitHashes:        commitHashes,
		CiMaterials:         nil,
		TriggeredBy:         ciTriggerRequest.TriggeredBy,
		InvalidateCache:     ciTriggerRequest.InvalidateCache,
		RuntimeParameters:   runtimeParams,
		EnvironmentId:       ciTriggerRequest.EnvironmentId,
		PipelineType:        ciTriggerRequest.PipelineType,
		CiArtifactLastFetch: createdOn,
	}
	id, err := impl.triggerCiPipeline(trigger)

	if err != nil {
		return 0, err
	}
	return id, nil
}

func (impl *HandlerServiceImpl) HandleCIWebhook(gitCiTriggerRequest bean.GitCiTriggerRequest) (int, error) {
	impl.Logger.Debugw("HandleCIWebhook for material ", "material", gitCiTriggerRequest.CiPipelineMaterial)
	ciPipeline, err := impl.GetCiPipeline(gitCiTriggerRequest.CiPipelineMaterial.Id)
	if err != nil {
		impl.Logger.Errorw("err in getting ci_pipeline by ciPipelineMaterialId", "ciPipelineMaterialId", gitCiTriggerRequest.CiPipelineMaterial.Id, "err", err)
		return 0, err
	}
	if ciPipeline.IsManual || ciPipeline.PipelineType == buildCommonBean.LINKED_CD.ToString() {
		impl.Logger.Debugw("not handling for manual pipeline or in case of linked cd", "pipelineId", ciPipeline.Id)
		return 0, err
	}

	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(ciPipeline.Id)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return 0, err
	}
	isValidBuildSequence, err := impl.validateBuildSequence(gitCiTriggerRequest, ciPipeline.Id)
	if !isValidBuildSequence {
		return 0, errors.New("ignoring older build for ciMaterial " + strconv.Itoa(gitCiTriggerRequest.CiPipelineMaterial.Id) +
			" commit " + gitCiTriggerRequest.CiPipelineMaterial.GitCommit.Commit)
	}
	// updating runtime params
	runtimeParams := common.NewRuntimeParameters()
	for k, v := range gitCiTriggerRequest.ExtraEnvironmentVariables {
		runtimeParams = runtimeParams.AddSystemVariable(k, v)
	}
	runtimeParams, err = impl.updateRuntimeParamsForAutoCI(ciPipeline.Id, runtimeParams)
	if err != nil {
		impl.Logger.Errorw("err, updateRuntimeParamsForAutoCI", "ciPipelineId", ciPipeline.Id,
			"runtimeParameters", runtimeParams, "err", err)
		return 0, err
	}
	commitHashes, err := impl.buildAutomaticTriggerCommitHashes(ciMaterials, gitCiTriggerRequest)
	if err != nil {
		return 0, err
	}

	trigger := &types.CiTriggerRequest{
		PipelineId:        ciPipeline.Id,
		CommitHashes:      commitHashes,
		CiMaterials:       ciMaterials,
		TriggeredBy:       gitCiTriggerRequest.TriggeredBy,
		RuntimeParameters: runtimeParams,
	}
	id, err := impl.triggerCiPipeline(trigger)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (impl *HandlerServiceImpl) extractPodStatusAndWorkflow(workflowStatus eventProcessorBean.CiCdStatus) (string, string, *pipelineConfig.CiWorkflow, error) {
	workflowName, status, _, message, _, _ := pipeline2.ExtractWorkflowStatus(workflowStatus)
	if workflowName == "" {
		impl.Logger.Errorw("extract workflow status, invalid wf name", "workflowName", workflowName, "status", status, "message", message)
		return status, message, nil, errors.New("invalid wf name")
	}
	workflowId, err := strconv.Atoi(workflowName[:strings.Index(workflowName, "-")])
	if err != nil {
		impl.Logger.Errorw("extract workflowId, invalid wf name", "workflowName", workflowName, "err", err)
		return status, message, nil, err
	}

	savedWorkflow, err := impl.ciWorkflowRepository.FindById(workflowId)
	if err != nil {
		impl.Logger.Errorw("cannot get saved wf", "workflowId", workflowId, "err", err)
		return status, message, nil, err
	}

	return status, message, savedWorkflow, err

}

func (impl *HandlerServiceImpl) getRefWorkflowAndCiRetryCount(savedWorkflow *pipelineConfig.CiWorkflow) (int, *pipelineConfig.CiWorkflow, error) {
	var err error

	if savedWorkflow.ReferenceCiWorkflowId != 0 {
		savedWorkflow, err = impl.ciWorkflowRepository.FindById(savedWorkflow.ReferenceCiWorkflowId)
		if err != nil {
			impl.Logger.Errorw("cannot get saved wf", "err", err)
			return 0, savedWorkflow, err
		}
	}
	retryCount, err := impl.ciWorkflowRepository.FindRetriedWorkflowCountByReferenceId(savedWorkflow.Id)
	return retryCount, savedWorkflow, err
}

func (impl *HandlerServiceImpl) validateBuildSequence(gitCiTriggerRequest bean.GitCiTriggerRequest, pipelineId int) (bool, error) {
	isValid := true
	lastTriggeredBuild, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflow(pipelineId)
	if !(lastTriggeredBuild.Status == string(v1alpha1.NodePending) || lastTriggeredBuild.Status == string(v1alpha1.NodeRunning)) {
		return true, nil
	}
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("cannot get last build for pipeline", "pipelineId", pipelineId)
		return false, err
	}

	ciPipelineMaterial := gitCiTriggerRequest.CiPipelineMaterial

	if ciPipelineMaterial.Type == string(constants.SOURCE_TYPE_BRANCH_FIXED) {
		if ciPipelineMaterial.GitCommit.Date.Before(lastTriggeredBuild.GitTriggers[ciPipelineMaterial.Id].Date) {
			impl.Logger.Warnw("older commit cannot be built for pipeline", "pipelineId", pipelineId, "ciMaterial", gitCiTriggerRequest.CiPipelineMaterial.Id)
			isValid = false
		}
	}

	return isValid, nil
}

func (impl *HandlerServiceImpl) buildAutomaticTriggerCommitHashes(ciMaterials []*pipelineConfig.CiPipelineMaterial, request bean.GitCiTriggerRequest) (map[int]pipelineConfig.GitCommit, error) {
	commitHashes := map[int]pipelineConfig.GitCommit{}
	for _, ciMaterial := range ciMaterials {
		if ciMaterial.Id == request.CiPipelineMaterial.Id || len(ciMaterials) == 1 {
			request.CiPipelineMaterial.GitCommit = SetGitCommitValuesForBuildingCommitHash(ciMaterial, request.CiPipelineMaterial.GitCommit)
			commitHashes[ciMaterial.Id] = request.CiPipelineMaterial.GitCommit
		} else {
			// this is possible in case of non Webhook, as there would be only one pipeline material per git material in case of PR
			lastCommit, err := impl.getLastSeenCommit(ciMaterial.Id)
			if err != nil {
				return map[int]pipelineConfig.GitCommit{}, err
			}
			lastCommit = SetGitCommitValuesForBuildingCommitHash(ciMaterial, lastCommit)
			commitHashes[ciMaterial.Id] = lastCommit
		}
	}
	return commitHashes, nil
}

func (impl *HandlerServiceImpl) GetCiPipeline(ciMaterialId int) (*pipelineConfig.CiPipeline, error) {
	ciMaterial, err := impl.ciPipelineMaterialRepository.GetById(ciMaterialId)
	if err != nil {
		return nil, err
	}
	ciPipeline := ciMaterial.CiPipeline
	return ciPipeline, nil
}

func (impl *HandlerServiceImpl) buildManualTriggerCommitHashes(ciTriggerRequest bean.CiTriggerRequest) (map[int]pipelineConfig.GitCommit, *common.RuntimeParameters, error) {
	commitHashes := map[int]pipelineConfig.GitCommit{}
	runtimeParams := impl.getRuntimeParamsForBuildingManualTriggerHashes(ciTriggerRequest)
	for _, ciPipelineMaterial := range ciTriggerRequest.CiPipelineMaterial {

		pipeLineMaterialFromDb, err := impl.ciPipelineMaterialRepository.GetById(ciPipelineMaterial.Id)
		if err != nil {
			impl.Logger.Errorw("err in fetching pipeline material by id", "err", err)
			return map[int]pipelineConfig.GitCommit{}, nil, err
		}

		pipelineType := pipeLineMaterialFromDb.Type
		if pipelineType == constants.SOURCE_TYPE_BRANCH_FIXED {
			gitCommit, err := impl.BuildManualTriggerCommitHashesForSourceTypeBranchFix(ciPipelineMaterial, pipeLineMaterialFromDb)
			if err != nil {
				impl.Logger.Errorw("err", "err", err)
				return map[int]pipelineConfig.GitCommit{}, nil, err
			}
			commitHashes[ciPipelineMaterial.Id] = gitCommit

		} else if pipelineType == constants.SOURCE_TYPE_WEBHOOK {
			gitCommit, extraEnvVariables, err := impl.BuildManualTriggerCommitHashesForSourceTypeWebhook(ciPipelineMaterial, pipeLineMaterialFromDb)
			if err != nil {
				impl.Logger.Errorw("err", "err", err)
				return map[int]pipelineConfig.GitCommit{}, nil, err
			}
			commitHashes[ciPipelineMaterial.Id] = gitCommit
			for key, value := range extraEnvVariables {
				runtimeParams = runtimeParams.AddSystemVariable(key, value)
			}
		}
	}
	return commitHashes, runtimeParams, nil
}

func (impl *HandlerServiceImpl) BuildManualTriggerCommitHashesForSourceTypeBranchFix(ciPipelineMaterial bean.CiPipelineMaterial, pipeLineMaterialFromDb *pipelineConfig.CiPipelineMaterial) (pipelineConfig.GitCommit, error) {
	commitMetadataRequest := &gitSensor.CommitMetadataRequest{
		PipelineMaterialId: ciPipelineMaterial.Id,
		GitHash:            ciPipelineMaterial.GitCommit.Commit,
		GitTag:             ciPipelineMaterial.GitTag,
	}
	gitCommitResponse, err := impl.gitSensorClient.GetCommitMetadataForPipelineMaterial(context.Background(), commitMetadataRequest)
	if err != nil {
		impl.Logger.Errorw("err in fetching commit metadata", "commitMetadataRequest", commitMetadataRequest, "err", err)
		return pipelineConfig.GitCommit{}, err
	}
	if gitCommitResponse == nil {
		return pipelineConfig.GitCommit{}, errors.New("commit not found")
	}

	gitCommit := pipelineConfig.GitCommit{
		Commit:                 gitCommitResponse.Commit,
		Author:                 gitCommitResponse.Author,
		Date:                   gitCommitResponse.Date,
		Message:                gitCommitResponse.Message,
		Changes:                gitCommitResponse.Changes,
		GitRepoName:            pipeLineMaterialFromDb.GitMaterial.Name[strings.Index(pipeLineMaterialFromDb.GitMaterial.Name, "-")+1:],
		GitRepoUrl:             pipeLineMaterialFromDb.GitMaterial.Url,
		CiConfigureSourceValue: pipeLineMaterialFromDb.Value,
		CiConfigureSourceType:  pipeLineMaterialFromDb.Type,
	}

	return gitCommit, nil
}

func (impl *HandlerServiceImpl) BuildManualTriggerCommitHashesForSourceTypeWebhook(ciPipelineMaterial bean.CiPipelineMaterial, pipeLineMaterialFromDb *pipelineConfig.CiPipelineMaterial) (pipelineConfig.GitCommit, map[string]string, error) {
	webhookDataInput := ciPipelineMaterial.GitCommit.WebhookData

	// fetch webhook data on the basis of Id
	webhookDataRequest := &gitSensor.WebhookDataRequest{
		Id:                   webhookDataInput.Id,
		CiPipelineMaterialId: ciPipelineMaterial.Id,
	}

	webhookAndCiData, err := impl.gitSensorClient.GetWebhookData(context.Background(), webhookDataRequest)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return pipelineConfig.GitCommit{}, nil, err
	}
	webhookData := webhookAndCiData.WebhookData

	// if webhook event is of merged type, then fetch latest commit for target branch
	if webhookData.EventActionType == bean.WEBHOOK_EVENT_MERGED_ACTION_TYPE {

		// get target branch name from webhook
		targetBranchName := webhookData.Data[bean.WEBHOOK_SELECTOR_TARGET_BRANCH_NAME_NAME]
		if targetBranchName == "" {
			impl.Logger.Error("target branch not found from webhook data")
			return pipelineConfig.GitCommit{}, nil, err
		}

		// get latest commit hash for target branch
		latestCommitMetadataRequest := &gitSensor.CommitMetadataRequest{
			PipelineMaterialId: ciPipelineMaterial.Id,
			BranchName:         targetBranchName,
		}

		latestCommit, err := impl.gitSensorClient.GetCommitMetadata(context.Background(), latestCommitMetadataRequest)

		if err != nil {
			impl.Logger.Errorw("err", "err", err)
			return pipelineConfig.GitCommit{}, nil, err
		}

		// update webhookData (local) with target latest hash
		webhookData.Data[bean.WEBHOOK_SELECTOR_TARGET_CHECKOUT_NAME] = latestCommit.Commit

	}

	// build git commit
	gitCommit := pipelineConfig.GitCommit{
		GitRepoName:            pipeLineMaterialFromDb.GitMaterial.Name[strings.Index(pipeLineMaterialFromDb.GitMaterial.Name, "-")+1:],
		GitRepoUrl:             pipeLineMaterialFromDb.GitMaterial.Url,
		CiConfigureSourceValue: pipeLineMaterialFromDb.Value,
		CiConfigureSourceType:  pipeLineMaterialFromDb.Type,
		WebhookData: pipelineConfig.WebhookData{
			Id:              int(webhookData.Id),
			EventActionType: webhookData.EventActionType,
			Data:            webhookData.Data,
		},
	}

	return gitCommit, webhookAndCiData.ExtraEnvironmentVariables, nil
}

func (impl *HandlerServiceImpl) getLastSeenCommit(ciMaterialId int) (pipelineConfig.GitCommit, error) {
	var materialIds []int
	materialIds = append(materialIds, ciMaterialId)
	headReq := &gitSensor.HeadRequest{
		MaterialIds: materialIds,
	}
	res, err := impl.gitSensorClient.GetHeadForPipelineMaterials(context.Background(), headReq)
	if err != nil {
		return pipelineConfig.GitCommit{}, err
	}
	if len(res) == 0 {
		return pipelineConfig.GitCommit{}, errors.New("received empty response")
	}
	gitCommit := pipelineConfig.GitCommit{
		Commit:  res[0].GitCommit.Commit,
		Author:  res[0].GitCommit.Author,
		Date:    res[0].GitCommit.Date,
		Message: res[0].GitCommit.Message,
		Changes: res[0].GitCommit.Changes,
	}
	return gitCommit, nil
}

func SetGitCommitValuesForBuildingCommitHash(ciMaterial *pipelineConfig.CiPipelineMaterial, oldGitCommit pipelineConfig.GitCommit) pipelineConfig.GitCommit {
	newGitCommit := oldGitCommit
	newGitCommit.CiConfigureSourceType = ciMaterial.Type
	newGitCommit.CiConfigureSourceValue = ciMaterial.Value
	newGitCommit.GitRepoUrl = ciMaterial.GitMaterial.Url
	newGitCommit.GitRepoName = ciMaterial.GitMaterial.Name[strings.Index(ciMaterial.GitMaterial.Name, "-")+1:]
	return newGitCommit
}

func (impl *HandlerServiceImpl) fetchVariableSnapshotForCiRetrigger(trigger *types.CiTriggerRequest) (map[string]string, error) {
	scope := resourceQualifiers.Scope{
		AppId: trigger.RetriggerWorkflowRequest.AppId,
	}
	request := pipelineConfigBean.NewBuildPrePostStepDataReq(trigger.RetriggerWorkflowRequest.PipelineId, pipelineConfigBean.CiStage, scope)
	request = updateBuildPrePostStepDataReq(request, trigger)
	prePostAndRefPluginResponse, err := impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(request)
	if err != nil {
		impl.Logger.Errorw("error in getting pre steps data for wf request", "ciPipelineId", trigger.RetriggerWorkflowRequest.PipelineId, "err", err)
		dbErr := impl.markCurrentCiWorkflowFailed(trigger.RetriggerCiWorkflow, err)
		if dbErr != nil {
			impl.Logger.Errorw("saving workflow error", "err", dbErr)
		}
		return nil, err
	}
	return prePostAndRefPluginResponse.VariableSnapshot, nil
}

func (impl *HandlerServiceImpl) prepareCiWfRequest(trigger *types.CiTriggerRequest) (map[string]string, *pipelineConfig.CiWorkflow, *types.WorkflowRequest, error) {
	var variableSnapshot map[string]string
	var savedCiWf *pipelineConfig.CiWorkflow
	var workflowRequest *types.WorkflowRequest
	var err error
	if trigger.IsRetrigger {
		variableSnapshot, err = impl.fetchVariableSnapshotForCiRetrigger(trigger)
		if err != nil {
			impl.Logger.Errorw("error in fetchVariableSnapshotForCiRetrigger", "triggerRequest", trigger, "err", err)
			return nil, nil, nil, err
		}
		savedCiWf, workflowRequest = trigger.RetriggerCiWorkflow, trigger.RetriggerWorkflowRequest
		if trigger.RetriggerCiWorkflow != nil {
			workflowRequest.ReferenceCiWorkflowId = trigger.RetriggerCiWorkflow.ReferenceCiWorkflowId
		}
		workflowRequest.IsReTrigger = true
	} else {
		variableSnapshot, savedCiWf, workflowRequest, err = impl.StartCiWorkflowAndPrepareWfRequest(trigger)
		if err != nil {
			impl.Logger.Errorw("error in starting ci workflow and preparing wf request", "triggerRequest", trigger, "err", err)
			return nil, nil, nil, err
		}
		workflowRequest.CiPipelineType = trigger.PipelineType
	}
	return variableSnapshot, savedCiWf, workflowRequest, nil
}

func (impl *HandlerServiceImpl) triggerCiPipeline(trigger *types.CiTriggerRequest) (int, error) {
	variableSnapshot, savedCiWf, workflowRequest, err := impl.prepareCiWfRequest(trigger)
	if err != nil {
		impl.Logger.Errorw("error in preparing wf request", "triggerRequest", trigger, "err", err)
		return 0, err
	}

	// Check if auto-abort is enabled for this pipeline and abort previous builds if needed
	err = impl.abortPreviousRunningBuilds(trigger.PipelineId, trigger.TriggeredBy)
	if err != nil {
		impl.Logger.Errorw("error in aborting previous running builds", "pipelineId", trigger.PipelineId, "err", err)
		// Log error but don't fail the trigger - previous builds aborting is a best-effort operation
	}

	err = impl.executeCiPipeline(workflowRequest)
	if err != nil {
		impl.Logger.Errorw("error in executing ci pipeline", "err", err)
		dbErr := impl.markCurrentCiWorkflowFailed(savedCiWf, err)
		if dbErr != nil {
			impl.Logger.Errorw("update ci workflow error", "err", dbErr)
		}
		return 0, err
	}
	impl.Logger.Debugw("ci triggered", " pipeline ", trigger.PipelineId)

	var variableSnapshotHistories = sliceUtil.GetBeansPtr(
		repository4.GetSnapshotBean(savedCiWf.Id, repository4.HistoryReferenceTypeCIWORKFLOW, variableSnapshot))
	if len(variableSnapshotHistories) > 0 {
		err = impl.scopedVariableManager.SaveVariableHistoriesForTrigger(variableSnapshotHistories, trigger.TriggeredBy)
		if err != nil {
			impl.Logger.Errorf("Not able to save variable snapshot for CI trigger %s", err)
		}
	}

	middleware.CiTriggerCounter.WithLabelValues(workflowRequest.AppName, workflowRequest.PipelineName).Inc()

	runnableFunc := func() {
		impl.ciService.WriteCITriggerEvent(trigger, workflowRequest)
	}
	impl.asyncRunnable.Execute(runnableFunc)

	return savedCiWf.Id, err
}

func (impl *HandlerServiceImpl) GetCiMaterials(pipelineId int, ciMaterials []*pipelineConfig.CiPipelineMaterial) ([]*pipelineConfig.CiPipelineMaterial, error) {
	if !(len(ciMaterials) == 0) {
		return ciMaterials, nil
	} else {
		ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(pipelineId)
		if err != nil {
			impl.Logger.Errorw("err", "err", err)
			return nil, err
		}
		impl.Logger.Debug("ciMaterials for pipeline trigger ", ciMaterials)
		return ciMaterials, nil
	}
}

func (impl *HandlerServiceImpl) StartCiWorkflowAndPrepareWfRequest(trigger *types.CiTriggerRequest) (map[string]string, *pipelineConfig.CiWorkflow, *types.WorkflowRequest, error) {
	impl.Logger.Debugw("ci pipeline manual trigger", "request", trigger)
	ciMaterials, err := impl.GetCiMaterials(trigger.PipelineId, trigger.CiMaterials)
	if err != nil {
		impl.Logger.Errorw("error in getting ci materials", "pipelineId", trigger.PipelineId, "ciMaterials", trigger.CiMaterials, "err", err)
		return nil, nil, nil, err
	}

	ciPipelineScripts, err := impl.ciPipelineRepository.FindCiScriptsByCiPipelineId(trigger.PipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("error in getting ci script by pipeline id", "pipelineId", trigger.PipelineId, "err", err)
		return nil, nil, nil, err
	}

	var pipeline *pipelineConfig.CiPipeline
	for _, m := range ciMaterials {
		pipeline = m.CiPipeline
		break
	}

	scope := resourceQualifiers.Scope{
		AppId: pipeline.App.Id,
	}
	ciWorkflowConfigNamespace := impl.config.GetDefaultNamespace()
	envModal, isJob, err := impl.getEnvironmentForJob(pipeline, trigger)
	if err != nil {
		impl.Logger.Errorw("error in getting environment for job", "pipelineId", trigger.PipelineId, "err", err)
		return nil, nil, nil, err
	}
	if isJob && envModal != nil {

		err = impl.checkArgoSetupRequirement(envModal)
		if err != nil {
			impl.Logger.Errorw("error in checking argo setup requirement", "envModal", envModal, "err", err)
			return nil, nil, nil, err
		}

		ciWorkflowConfigNamespace = envModal.Namespace

		// This will be populated for jobs running in selected environment
		scope.EnvId = envModal.Id
		scope.ClusterId = envModal.ClusterId

		scope.SystemMetadata = &resourceQualifiers.SystemMetadata{
			EnvironmentName: envModal.Name,
			ClusterName:     envModal.Cluster.ClusterName,
			Namespace:       envModal.Namespace,
		}
	}
	if scope.SystemMetadata == nil {
		scope.SystemMetadata = &resourceQualifiers.SystemMetadata{
			Namespace: ciWorkflowConfigNamespace,
			AppName:   pipeline.App.AppName,
		}
	}
	savedCiWf, err := impl.saveNewWorkflowForCITrigger(pipeline, ciWorkflowConfigNamespace, trigger.CommitHashes, trigger.TriggeredBy, ciMaterials, trigger.EnvironmentId, isJob, trigger.ReferenceCiWorkflowId)
	if err != nil {
		impl.Logger.Errorw("could not save new workflow", "err", err)
		return nil, nil, nil, err
	}
	// preCiSteps, postCiSteps, refPluginsData, err := impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(pipeline.Id, ciEvent)
	request := pipelineConfigBean.NewBuildPrePostStepDataReq(pipeline.Id, pipelineConfigBean.CiStage, scope)
	request = updateBuildPrePostStepDataReq(request, trigger)
	prePostAndRefPluginResponse, err := impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(request)
	if err != nil {
		impl.Logger.Errorw("error in getting pre steps data for wf request", "err", err, "ciPipelineId", pipeline.Id)
		dbErr := impl.markCurrentCiWorkflowFailed(savedCiWf, err)
		if dbErr != nil {
			impl.Logger.Errorw("saving workflow error", "err", dbErr)
		}
		return nil, nil, nil, err
	}
	preCiSteps := prePostAndRefPluginResponse.PreStageSteps
	postCiSteps := prePostAndRefPluginResponse.PostStageSteps
	refPluginsData := prePostAndRefPluginResponse.RefPluginData
	variableSnapshot := prePostAndRefPluginResponse.VariableSnapshot

	if len(preCiSteps) == 0 && isJob {
		errMsg := fmt.Sprintf("No tasks are configured in this job pipeline")
		validationErr := util.NewApiError(http.StatusNotFound, errMsg, errMsg)
		return nil, nil, nil, validationErr
	}

	// get env variables of git trigger data and add it in the extraEnvVariables
	gitTriggerEnvVariables, _, err := impl.ciCdPipelineOrchestrator.GetGitCommitEnvVarDataForCICDStage(savedCiWf.GitTriggers)
	if err != nil {
		impl.Logger.Errorw("error in getting gitTrigger env data for stage", "gitTriggers", savedCiWf.GitTriggers, "err", err)
		return nil, nil, nil, err
	}

	for k, v := range gitTriggerEnvVariables {
		trigger.RuntimeParameters = trigger.RuntimeParameters.AddSystemVariable(k, v)
	}

	workflowRequest, err := impl.buildWfRequestForCiPipeline(pipeline, trigger, ciMaterials, savedCiWf, ciWorkflowConfigNamespace, ciPipelineScripts, preCiSteps, postCiSteps, refPluginsData, isJob)
	if err != nil {
		impl.Logger.Errorw("make workflow req", "err", err)
		return nil, nil, nil, err
	}
	err = impl.handleRuntimeParamsValidations(trigger, ciMaterials, workflowRequest)
	if err != nil {
		savedCiWf.Status = cdWorkflow.WorkflowAborted
		savedCiWf.Message = err.Error()
		err1 := impl.ciService.UpdateCiWorkflowWithStage(savedCiWf)
		if err1 != nil {
			impl.Logger.Errorw("could not save workflow, after failing due to conflicting image tag")
		}
		return nil, nil, nil, err
	}

	workflowRequest.Scope = scope
	workflowRequest.AppId = pipeline.AppId
	workflowRequest.Env = envModal
	if isJob {
		workflowRequest.Type = pipelineConfigBean.JOB_WORKFLOW_PIPELINE_TYPE
	} else {
		workflowRequest.Type = pipelineConfigBean.CI_WORKFLOW_PIPELINE_TYPE
	}
	workflowRequest, err = impl.updateWorkflowRequestWithBuildxFlags(workflowRequest, scope)
	if err != nil {
		impl.Logger.Errorw("error, updateWorkflowRequestWithBuildxFlags", "workflowRequest", workflowRequest, "err", err)
		return nil, nil, nil, err
	}
	if impl.canSetK8sDriverData(workflowRequest) {
		err = impl.setBuildxK8sDriverData(workflowRequest)
		if err != nil {
			impl.Logger.Errorw("error in setBuildxK8sDriverData", "BUILDX_K8S_DRIVER_OPTIONS", impl.config.BuildxK8sDriverOptions, "err", err)
			return nil, nil, nil, err
		}
	}
	savedCiWf.LogLocation = fmt.Sprintf("%s/%s/main.log", impl.config.GetDefaultBuildLogsKeyPrefix(), workflowRequest.WorkflowNamePrefix)
	err = impl.updateCiWorkflow(workflowRequest, savedCiWf)
	appLabels, err := impl.appCrudOperationService.GetLabelsByAppId(pipeline.AppId)
	if err != nil {
		impl.Logger.Errorw("error in getting labels by appId", "appId", pipeline.AppId, "err", err)
		return nil, nil, nil, err
	}
	workflowRequest.AppLabels = appLabels
	workflowRequest = impl.updateWorkflowRequestWithEntSupportData(workflowRequest)
	return variableSnapshot, savedCiWf, workflowRequest, nil
}

func (impl *HandlerServiceImpl) setBuildxK8sDriverData(workflowRequest *types.WorkflowRequest) error {
	dockerBuildConfig := workflowRequest.CiBuildConfig.DockerBuildConfig
	k8sDriverOptions, err := impl.getK8sDriverOptions(workflowRequest, dockerBuildConfig.TargetPlatform)
	if err != nil {
		impl.Logger.Errorw("error in parsing BUILDX_K8S_DRIVER_OPTIONS from the devtron-cm", "err", err)
	}
	dockerBuildConfig.BuildxK8sDriverOptions = k8sDriverOptions
	return nil
}

func (impl *HandlerServiceImpl) getEnvironmentForJob(pipeline *pipelineConfig.CiPipeline, trigger *types.CiTriggerRequest) (*repository6.Environment, bool, error) {
	app, err := impl.appRepository.FindById(pipeline.AppId)
	if err != nil {
		impl.Logger.Errorw("could not find app", "err", err)
		return nil, false, err
	}

	var env *repository6.Environment
	isJob := false
	if app.AppType == helper.Job {
		isJob = true
		if trigger.EnvironmentId != 0 {
			env, err = impl.envRepository.FindById(trigger.EnvironmentId)
			if err != nil {
				impl.Logger.Errorw("could not find environment", "err", err)
				return nil, isJob, err
			}
			return env, isJob, nil
		}
	}
	return nil, isJob, nil
}

// TODO: Send all trigger data
func (impl *HandlerServiceImpl) BuildPayload(trigger types.CiTriggerRequest, pipeline *pipelineConfig.CiPipeline) *client.Payload {
	payload := &client.Payload{}
	payload.AppName = pipeline.App.AppName
	payload.PipelineName = pipeline.Name
	return payload
}

func (impl *HandlerServiceImpl) saveNewWorkflowForCITrigger(pipeline *pipelineConfig.CiPipeline, ciWorkflowConfigNamespace string,
	commitHashes map[int]pipelineConfig.GitCommit, userId int32, ciMaterials []*pipelineConfig.CiPipelineMaterial, EnvironmentId int, isJob bool, refCiWorkflowId int) (*pipelineConfig.CiWorkflow, error) {

	isCiTriggerBlocked, err := impl.checkIfCITriggerIsBlocked(pipeline, ciMaterials, isJob)
	if err != nil {
		impl.Logger.Errorw("error, checkIfCITriggerIsBlocked", "pipelineId", pipeline.Id, "err", err)
		return &pipelineConfig.CiWorkflow{}, err
	}
	ciWorkflow := &pipelineConfig.CiWorkflow{
		Name:                  pipeline.Name + "-" + strconv.Itoa(pipeline.Id),
		Status:                cdWorkflow.WorkflowStarting, // starting CIStage
		Message:               "",
		StartedOn:             time.Now(),
		CiPipelineId:          pipeline.Id,
		Namespace:             impl.config.GetDefaultNamespace(),
		BlobStorageEnabled:    impl.config.BlobStorageEnabled,
		GitTriggers:           commitHashes,
		LogLocation:           "",
		TriggeredBy:           userId,
		ReferenceCiWorkflowId: refCiWorkflowId,
		ExecutorType:          impl.config.GetWorkflowExecutorType(),
	}
	if isJob {
		ciWorkflow.Namespace = ciWorkflowConfigNamespace
		ciWorkflow.EnvironmentId = EnvironmentId
	}
	if isCiTriggerBlocked {
		return impl.handleWFIfCITriggerIsBlocked(ciWorkflow)
	}
	err = impl.ciService.SaveCiWorkflowWithStage(ciWorkflow)
	if err != nil {
		impl.Logger.Errorw("saving workflow error", "err", err)
		return &pipelineConfig.CiWorkflow{}, err
	}
	impl.Logger.Debugw("workflow saved ", "id", ciWorkflow.Id)
	return ciWorkflow, nil
}

func (impl *HandlerServiceImpl) executeCiPipeline(workflowRequest *types.WorkflowRequest) error {
	_, _, err := impl.workflowService.SubmitWorkflow(workflowRequest)
	if err != nil {
		impl.Logger.Errorw("workflow error", "err", err)
		return err
	}
	return nil
}

func (impl *HandlerServiceImpl) buildS3ArtifactLocation(ciWorkflowConfigLogsBucket string, savedWf *pipelineConfig.CiWorkflow) (string, string, string) {
	ciArtifactLocationFormat := impl.config.GetArtifactLocationFormat()
	ArtifactLocation := fmt.Sprintf("s3://"+path.Join(ciWorkflowConfigLogsBucket, ciArtifactLocationFormat), savedWf.Id, savedWf.Id)
	artifactFileName := fmt.Sprintf(ciArtifactLocationFormat, savedWf.Id, savedWf.Id)
	return ArtifactLocation, ciWorkflowConfigLogsBucket, artifactFileName
}

func (impl *HandlerServiceImpl) buildDefaultArtifactLocation(savedWf *pipelineConfig.CiWorkflow) string {
	ciArtifactLocationFormat := impl.config.GetArtifactLocationFormat()
	ArtifactLocation := fmt.Sprintf(ciArtifactLocationFormat, savedWf.Id, savedWf.Id)
	return ArtifactLocation
}

func (impl *HandlerServiceImpl) buildWfRequestForCiPipeline(pipeline *pipelineConfig.CiPipeline, trigger *types.CiTriggerRequest, ciMaterials []*pipelineConfig.CiPipelineMaterial, savedWf *pipelineConfig.CiWorkflow, ciWorkflowConfigNamespace string, ciPipelineScripts []*pipelineConfig.CiPipelineScript, preCiSteps []*pipelineConfigBean.StepObject, postCiSteps []*pipelineConfigBean.StepObject, refPluginsData []*pipelineConfigBean.RefPluginObject, isJob bool) (*types.WorkflowRequest, error) {
	var ciProjectDetails []pipelineConfigBean.CiProjectDetails
	commitHashes := trigger.CommitHashes
	for _, ciMaterial := range ciMaterials {
		// ignore those materials which have inactive git material
		if ciMaterial == nil || ciMaterial.GitMaterial == nil || !ciMaterial.GitMaterial.Active {
			continue
		}
		commitHashForPipelineId := commitHashes[ciMaterial.Id]
		ciProjectDetail := pipelineConfigBean.CiProjectDetails{
			GitRepository:   ciMaterial.GitMaterial.Url,
			MaterialName:    ciMaterial.GitMaterial.Name,
			CheckoutPath:    ciMaterial.GitMaterial.CheckoutPath,
			FetchSubmodules: ciMaterial.GitMaterial.FetchSubmodules,
			CommitHash:      commitHashForPipelineId.Commit,
			Author:          commitHashForPipelineId.Author,
			SourceType:      ciMaterial.Type,
			SourceValue:     ciMaterial.Value,
			GitTag:          ciMaterial.GitTag,
			Message:         commitHashForPipelineId.Message,
			Type:            string(ciMaterial.Type),
			CommitTime:      commitHashForPipelineId.Date.Format(bean.LayoutRFC3339),
			GitOptions: pipelineConfigBean.GitOptions{
				UserName:              ciMaterial.GitMaterial.GitProvider.UserName,
				Password:              ciMaterial.GitMaterial.GitProvider.Password.String(),
				SshPrivateKey:         ciMaterial.GitMaterial.GitProvider.SshPrivateKey.String(),
				AccessToken:           ciMaterial.GitMaterial.GitProvider.AccessToken.String(),
				AuthMode:              ciMaterial.GitMaterial.GitProvider.AuthMode,
				EnableTLSVerification: ciMaterial.GitMaterial.GitProvider.EnableTLSVerification,
				TlsKey:                ciMaterial.GitMaterial.GitProvider.TlsKey,
				TlsCert:               ciMaterial.GitMaterial.GitProvider.TlsCert,
				CaCert:                ciMaterial.GitMaterial.GitProvider.CaCert,
			},
		}
		var err error
		ciProjectDetail, err = impl.updateCIProjectDetailWithCloningMode(pipeline.AppId, ciMaterial, ciProjectDetail)
		if err != nil {
			impl.Logger.Errorw("error, updateCIProjectDetailWithCloningMode", "pipelineId", pipeline.Id, "err", err)
			return nil, err
		}
		if ciMaterial.Type == constants.SOURCE_TYPE_WEBHOOK {
			webhookData := commitHashForPipelineId.WebhookData
			ciProjectDetail.WebhookData = pipelineConfig.WebhookData{
				Id:              webhookData.Id,
				EventActionType: webhookData.EventActionType,
				Data:            webhookData.Data,
			}
		}

		ciProjectDetails = append(ciProjectDetails, ciProjectDetail)
	}

	var beforeDockerBuildScripts []*bean.CiScript
	var afterDockerBuildScripts []*bean.CiScript
	for _, ciPipelineScript := range ciPipelineScripts {
		ciTask := &bean.CiScript{
			Id:             ciPipelineScript.Id,
			Index:          ciPipelineScript.Index,
			Name:           ciPipelineScript.Name,
			Script:         ciPipelineScript.Script,
			OutputLocation: ciPipelineScript.OutputLocation,
		}

		if ciPipelineScript.Stage == buildCommonBean.BEFORE_DOCKER_BUILD {
			beforeDockerBuildScripts = append(beforeDockerBuildScripts, ciTask)
		} else if ciPipelineScript.Stage == buildCommonBean.AFTER_DOCKER_BUILD {
			afterDockerBuildScripts = append(afterDockerBuildScripts, ciTask)
		}
	}

	if !(len(beforeDockerBuildScripts) == 0 && len(afterDockerBuildScripts) == 0) {
		// found beforeDockerBuildScripts/afterDockerBuildScripts
		// building preCiSteps & postCiSteps from them, refPluginsData not needed
		preCiSteps = buildCiStepsDataFromDockerBuildScripts(beforeDockerBuildScripts)
		postCiSteps = buildCiStepsDataFromDockerBuildScripts(afterDockerBuildScripts)
		refPluginsData = []*pipelineConfigBean.RefPluginObject{}
	}

	host, err := impl.attributeService.GetByKey(bean3.HostUrlKey)
	if err != nil {
		impl.Logger.Errorw("error in getting host url", "err", err, "hostUrl", host.Value)
		return nil, err
	}
	ciWorkflowConfigCiCacheBucket := impl.config.DefaultCacheBucket

	ciWorkflowConfigCiCacheRegion := impl.config.DefaultCacheBucketRegion

	ciWorkflowConfigCiImage := impl.config.GetDefaultImage()

	ciTemplate := pipeline.CiTemplate
	ciLevelArgs := pipeline.DockerArgs

	if ciLevelArgs == "" {
		ciLevelArgs = "{}"
	}

	if pipeline.CiTemplate.DockerBuildOptions == "" {
		pipeline.CiTemplate.DockerBuildOptions = "{}"
	}
	userEmailId, err := impl.userService.GetActiveEmailById(trigger.TriggeredBy)
	if err != nil {
		impl.Logger.Errorw("unable to find user email by id", "err", err, "id", trigger.TriggeredBy)
		return nil, err
	}
	var dockerfilePath string
	var dockerRepository string
	var checkoutPath string
	var ciBuildConfigBean *buildBean.CiBuildConfigBean
	dockerRegistry := &repository3.DockerArtifactStore{}
	ciBaseBuildConfigEntity := ciTemplate.CiBuildConfig
	ciBaseBuildConfigBean, err := adapter.ConvertDbBuildConfigToBean(ciBaseBuildConfigEntity)
	if err != nil {
		impl.Logger.Errorw("error occurred while converting buildconfig dbEntity to configBean", "ciBuildConfigEntity", ciBaseBuildConfigEntity, "err", err)
		return nil, errors.New("error while parsing ci build config")
	}
	if !pipeline.IsExternal && pipeline.IsDockerConfigOverridden {
		templateOverrideBean, err := impl.ciTemplateService.FindTemplateOverrideByCiPipelineId(pipeline.Id)
		if err != nil {
			return nil, err
		}
		ciBuildConfigBean = templateOverrideBean.CiBuildConfig
		// updating args coming from ciBaseBuildConfigEntity because it is not part of Ci override
		if ciBuildConfigBean != nil && ciBuildConfigBean.DockerBuildConfig != nil && ciBaseBuildConfigBean != nil && ciBaseBuildConfigBean.DockerBuildConfig != nil {
			ciBuildConfigBean.DockerBuildConfig.Args = ciBaseBuildConfigBean.DockerBuildConfig.Args
		}
		templateOverride := templateOverrideBean.CiTemplateOverride
		checkoutPath = templateOverride.GitMaterial.CheckoutPath
		dockerfilePath = templateOverride.DockerfilePath
		dockerRepository = templateOverride.DockerRepository
		dockerRegistry = templateOverride.DockerRegistry
	} else {
		checkoutPath = ciTemplate.GitMaterial.CheckoutPath
		dockerfilePath = ciTemplate.DockerfilePath
		dockerRegistry = ciTemplate.DockerRegistry
		dockerRepository = ciTemplate.DockerRepository
		ciBuildConfigBean = ciBaseBuildConfigBean
		if ciBuildConfigBean != nil {
			ciBuildConfigBean.BuildContextGitMaterialId = ciTemplate.BuildContextGitMaterialId
		}

	}
	if checkoutPath == "" {
		checkoutPath = "./"
	}
	var dockerImageTag string
	customTag, err := impl.customTagService.GetActiveCustomTagByEntityKeyAndValue(pipelineConfigBean.EntityTypeCiPipelineId, strconv.Itoa(pipeline.Id))
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	if customTag.Id != 0 && customTag.Enabled == true {
		imagePathReservation, err := impl.customTagService.GenerateImagePath(pipelineConfigBean.EntityTypeCiPipelineId, strconv.Itoa(pipeline.Id), dockerRegistry.RegistryURL, dockerRepository)
		if err != nil {
			if errors.Is(err, pipelineConfigBean.ErrImagePathInUse) {
				errMsg := pipelineConfigBean.ImageTagUnavailableMessage
				validationErr := util.NewApiError(http.StatusConflict, errMsg, errMsg)
				dbErr := impl.markCurrentCiWorkflowFailed(savedWf, validationErr)
				if dbErr != nil {
					impl.Logger.Errorw("could not save workflow, after failing due to conflicting image tag", "err", dbErr, "savedWf", savedWf.Id)
				}
				return nil, err
			}
			return nil, err
		}
		savedWf.ImagePathReservationIds = []int{imagePathReservation.Id}
		// imagePath = docker.io/avd0/dashboard:fd23414b
		imagePathSplit := strings.Split(imagePathReservation.ImagePath, ":")
		if len(imagePathSplit) >= 1 {
			dockerImageTag = imagePathSplit[len(imagePathSplit)-1]
		}
	} else {
		dockerImageTag = impl.buildImageTag(commitHashes, pipeline.Id, savedWf.Id)
	}

	// copyContainerImage plugin specific logic
	var registryCredentialMap map[string]bean2.RegistryCredentials
	var pluginArtifactStage string
	var imageReservationIds []int
	var registryDestinationImageMap map[string][]string
	if !isJob {
		registryDestinationImageMap, registryCredentialMap, pluginArtifactStage, imageReservationIds, err = impl.GetWorkflowRequestVariablesForCopyContainerImagePlugin(preCiSteps, postCiSteps, dockerImageTag, customTag.Id,
			fmt.Sprintf(pipelineConfigBean.ImagePathPattern,
				dockerRegistry.RegistryURL,
				dockerRepository,
				dockerImageTag),
			dockerRegistry.Id)
		if err != nil {
			impl.Logger.Errorw("error in getting env variables for copyContainerImage plugin")
			dbErr := impl.markCurrentCiWorkflowFailed(savedWf, err)
			if dbErr != nil {
				impl.Logger.Errorw("could not save workflow, after failing due to conflicting image tag", "err", dbErr, "savedWf", savedWf.Id)
			}
			return nil, err
		}

		savedWf.ImagePathReservationIds = append(savedWf.ImagePathReservationIds, imageReservationIds...)
	}
	// mergedArgs := string(merged)
	oldArgs := ciTemplate.Args
	ciBuildConfigBean, err = adapter.OverrideCiBuildConfig(dockerfilePath, oldArgs, ciLevelArgs, ciTemplate.DockerBuildOptions, ciTemplate.TargetPlatform, ciBuildConfigBean)
	if err != nil {
		impl.Logger.Errorw("error occurred while overriding ci build config", "oldArgs", oldArgs, "ciLevelArgs", ciLevelArgs, "error", err)
		return nil, errors.New("error while parsing ci build config")
	}
	buildContextCheckoutPath, err := impl.ciPipelineMaterialRepository.GetCheckoutPath(ciBuildConfigBean.BuildContextGitMaterialId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error occurred while getting checkout path from git material", "gitMaterialId", ciBuildConfigBean.BuildContextGitMaterialId, "error", err)
		return nil, err
	}
	if buildContextCheckoutPath == "" {
		buildContextCheckoutPath = checkoutPath
	}
	if ciBuildConfigBean.UseRootBuildContext {
		// use root build context i.e '.'
		buildContextCheckoutPath = "."
	}

	ciBuildConfigBean.PipelineType = trigger.PipelineType

	if ciBuildConfigBean.CiBuildType == buildBean.SELF_DOCKERFILE_BUILD_TYPE || ciBuildConfigBean.CiBuildType == buildBean.MANAGED_DOCKERFILE_BUILD_TYPE {
		ciBuildConfigBean.DockerBuildConfig.BuildContext = filepath.Join(buildContextCheckoutPath, ciBuildConfigBean.DockerBuildConfig.BuildContext)
		dockerBuildConfig := ciBuildConfigBean.DockerBuildConfig
		dockerfilePath = filepath.Join(checkoutPath, dockerBuildConfig.DockerfilePath)
		dockerBuildConfig.DockerfilePath = dockerfilePath
		checkoutPath = dockerfilePath[:strings.LastIndex(dockerfilePath, "/")+1]
	} else if ciBuildConfigBean.CiBuildType == buildBean.BUILDPACK_BUILD_TYPE {
		buildPackConfig := ciBuildConfigBean.BuildPackConfig
		checkoutPath = filepath.Join(checkoutPath, buildPackConfig.ProjectPath)
	}

	if ciBuildConfigBean.DockerBuildConfig != nil {
		ciBuildConfigBean = impl.updateCIBuildConfig(ciBuildConfigBean)
	}

	workflowRequest := &types.WorkflowRequest{
		WorkflowNamePrefix:          strconv.Itoa(savedWf.Id) + "-" + savedWf.Name,
		PipelineName:                pipeline.Name,
		PipelineId:                  pipeline.Id,
		CiCacheFileName:             pipeline.Name + "-" + strconv.Itoa(pipeline.Id) + ".tar.gz",
		CiProjectDetails:            ciProjectDetails,
		Namespace:                   ciWorkflowConfigNamespace,
		BlobStorageConfigured:       savedWf.BlobStorageEnabled,
		CiImage:                     ciWorkflowConfigCiImage,
		WorkflowId:                  savedWf.Id,
		TriggeredBy:                 savedWf.TriggeredBy,
		CacheLimit:                  impl.config.CacheLimit,
		ScanEnabled:                 pipeline.ScanEnabled,
		CloudProvider:               impl.config.CloudProvider,
		DefaultAddressPoolBaseCidr:  impl.config.GetDefaultAddressPoolBaseCidr(),
		DefaultAddressPoolSize:      impl.config.GetDefaultAddressPoolSize(),
		PreCiSteps:                  preCiSteps,
		PostCiSteps:                 postCiSteps,
		RefPlugins:                  refPluginsData,
		AppName:                     pipeline.App.AppName,
		TriggerByAuthor:             userEmailId,
		CiBuildConfig:               ciBuildConfigBean,
		CiBuildDockerMtuValue:       impl.config.CiRunnerDockerMTUValue,
		CacheInvalidate:             trigger.InvalidateCache,
		SystemEnvironmentVariables:  trigger.RuntimeParameters.GetSystemVariables(),
		EnableBuildContext:          impl.config.EnableBuildContext,
		OrchestratorHost:            impl.config.OrchestratorHost,
		HostUrl:                     host.Value,
		OrchestratorToken:           impl.config.OrchestratorToken,
		ImageRetryCount:             impl.config.ImageRetryCount,
		ImageRetryInterval:          impl.config.ImageRetryInterval,
		WorkflowExecutor:            impl.config.GetWorkflowExecutorType(),
		Type:                        pipelineConfigBean.CI_WORKFLOW_PIPELINE_TYPE,
		CiArtifactLastFetch:         trigger.CiArtifactLastFetch,
		RegistryCredentialMap:       registryCredentialMap,
		PluginArtifactStage:         pluginArtifactStage,
		ImageScanMaxRetries:         impl.config.ImageScanMaxRetries,
		ImageScanRetryDelay:         impl.config.ImageScanRetryDelay,
		UseDockerApiToGetDigest:     impl.config.UseDockerApiToGetDigest,
		RegistryDestinationImageMap: registryDestinationImageMap,
	}
	workflowRequest.SetEntOnlyFields(trigger, impl.config)
	workflowCacheConfig := impl.ciCdPipelineOrchestrator.GetWorkflowCacheConfig(pipeline.App.AppType, trigger.PipelineType, pipeline.GetWorkflowCacheConfig())
	workflowRequest.IgnoreDockerCachePush = !workflowCacheConfig.Value
	workflowRequest.IgnoreDockerCachePull = !workflowCacheConfig.Value
	impl.Logger.Debugw("Ignore Cache values", "IgnoreDockerCachePush", workflowRequest.IgnoreDockerCachePush, "IgnoreDockerCachePull", workflowRequest.IgnoreDockerCachePull)
	if pipeline.App.AppType == helper.Job {
		workflowRequest.AppName = pipeline.App.DisplayName
	}
	if pipeline.ScanEnabled {
		scanToolMetadata, scanVia, err := impl.fetchImageScanExecutionMedium()
		if err != nil {
			impl.Logger.Errorw("error occurred getting scanned via", "err", err)
			return nil, err
		}
		workflowRequest.SetExecuteImageScanningVia(scanVia)
		if scanVia.IsScanMediumExternal() {
			imageScanExecutionSteps, refPlugins, err := impl.fetchImageScanExecutionStepsForWfRequest(scanToolMetadata)
			if err != nil {
				impl.Logger.Errorw("error occurred, fetchImageScanExecutionStepsForWfRequest", "scanToolMetadata", scanToolMetadata, "err", err)
				return nil, err
			}
			workflowRequest.SetImageScanningSteps(imageScanExecutionSteps)
			workflowRequest.RefPlugins = append(workflowRequest.RefPlugins, refPlugins...)
		}
	}
	if dockerRegistry != nil {
		workflowRequest, err = impl.updateWorkflowRequestWithRemoteConnConf(dockerRegistry, workflowRequest)
		if err != nil {
			impl.Logger.Errorw("error occurred updating workflow request", "dockerRegistryId", dockerRegistry.Id, "err", err)
			return nil, err
		}
		workflowRequest.DockerRegistryId = dockerRegistry.Id
		workflowRequest.DockerRegistryType = string(dockerRegistry.RegistryType)
		workflowRequest.DockerImageTag = dockerImageTag
		workflowRequest.DockerRegistryURL = dockerRegistry.RegistryURL
		workflowRequest.DockerRepository = dockerRepository
		workflowRequest.CheckoutPath = checkoutPath
		workflowRequest.DockerUsername = dockerRegistry.Username
		workflowRequest.DockerPassword = dockerRegistry.Password.String()
		workflowRequest.AwsRegion = dockerRegistry.AWSRegion
		workflowRequest.AccessKey = dockerRegistry.AWSAccessKeyId
		workflowRequest.SecretKey = dockerRegistry.AWSSecretAccessKey.String()
		workflowRequest.DockerConnection = dockerRegistry.Connection
		workflowRequest.DockerCert = dockerRegistry.Cert

	}
	ciWorkflowConfigLogsBucket := impl.config.GetDefaultBuildLogsBucket()

	switch workflowRequest.CloudProvider {
	case types.BLOB_STORAGE_S3:
		// No AccessKey is used for uploading artifacts, instead IAM based auth is used
		workflowRequest.CiCacheRegion = ciWorkflowConfigCiCacheRegion
		workflowRequest.CiCacheLocation = ciWorkflowConfigCiCacheBucket
		workflowRequest.CiArtifactLocation, workflowRequest.CiArtifactBucket, workflowRequest.CiArtifactFileName = impl.buildS3ArtifactLocation(ciWorkflowConfigLogsBucket, savedWf)
		workflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			AccessKey:                  impl.config.BlobStorageS3AccessKey,
			Passkey:                    impl.config.BlobStorageS3SecretKey,
			EndpointUrl:                impl.config.BlobStorageS3Endpoint,
			IsInSecure:                 impl.config.BlobStorageS3EndpointInsecure,
			CiCacheBucketName:          ciWorkflowConfigCiCacheBucket,
			CiCacheRegion:              ciWorkflowConfigCiCacheRegion,
			CiCacheBucketVersioning:    impl.config.BlobStorageS3BucketVersioned,
			CiArtifactBucketName:       workflowRequest.CiArtifactBucket,
			CiArtifactRegion:           impl.config.GetDefaultCdLogsBucketRegion(),
			CiArtifactBucketVersioning: impl.config.BlobStorageS3BucketVersioned,
			CiLogBucketName:            impl.config.GetDefaultBuildLogsBucket(),
			CiLogRegion:                impl.config.GetDefaultCdLogsBucketRegion(),
			CiLogBucketVersioning:      impl.config.BlobStorageS3BucketVersioned,
		}
	case types.BLOB_STORAGE_GCP:
		workflowRequest.GcpBlobConfig = &blob_storage.GcpBlobConfig{
			CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
			CacheBucketName:        ciWorkflowConfigCiCacheBucket,
			LogBucketName:          ciWorkflowConfigLogsBucket,
			ArtifactBucketName:     ciWorkflowConfigLogsBucket,
		}
		workflowRequest.CiArtifactLocation = impl.buildDefaultArtifactLocation(savedWf)
		workflowRequest.CiArtifactFileName = workflowRequest.CiArtifactLocation
	case types.BLOB_STORAGE_AZURE:
		workflowRequest.AzureBlobConfig = &blob_storage.AzureBlobConfig{
			Enabled:               impl.config.CloudProvider == types.BLOB_STORAGE_AZURE,
			AccountName:           impl.config.AzureAccountName,
			BlobContainerCiCache:  impl.config.AzureBlobContainerCiCache,
			AccountKey:            impl.config.AzureAccountKey,
			BlobContainerCiLog:    impl.config.AzureBlobContainerCiLog,
			BlobContainerArtifact: impl.config.AzureBlobContainerCiLog,
		}
		workflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			EndpointUrl:           impl.config.AzureGatewayUrl,
			IsInSecure:            impl.config.AzureGatewayConnectionInsecure,
			CiLogBucketName:       impl.config.AzureBlobContainerCiLog,
			CiLogRegion:           impl.config.DefaultCacheBucketRegion,
			CiLogBucketVersioning: impl.config.BlobStorageS3BucketVersioned,
			AccessKey:             impl.config.AzureAccountName,
		}
		workflowRequest.CiArtifactLocation = impl.buildDefaultArtifactLocation(savedWf)
		workflowRequest.CiArtifactFileName = workflowRequest.CiArtifactLocation
	default:
		if impl.config.BlobStorageEnabled {
			return nil, fmt.Errorf("blob storage %s not supported", workflowRequest.CloudProvider)
		}
	}
	return workflowRequest, nil
}

func (impl *HandlerServiceImpl) GetWorkflowRequestVariablesForCopyContainerImagePlugin(preCiSteps []*pipelineConfigBean.StepObject, postCiSteps []*pipelineConfigBean.StepObject, customTag string, customTagId int, buildImagePath string, buildImagedockerRegistryId string) (map[string][]string, map[string]bean2.RegistryCredentials, string, []int, error) {

	copyContainerImagePluginDetail, err := impl.globalPluginService.GetRefPluginIdByRefPluginName(buildCommonBean.COPY_CONTAINER_IMAGE)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in getting copyContainerImage plugin id", "err", err)
		return nil, nil, "", nil, err
	}

	pluginIdToVersionMap := make(map[int]string)
	for _, p := range copyContainerImagePluginDetail {
		pluginIdToVersionMap[p.Id] = p.Version
	}

	for _, step := range preCiSteps {
		if _, ok := pluginIdToVersionMap[step.RefPluginId]; ok {
			// for copyContainerImage plugin parse destination images and save its data in image path reservation table
			return nil, nil, "", nil, errors.New("copyContainerImage plugin not allowed in pre-ci step, please remove it and try again")
		}
	}

	registryCredentialMap := make(map[string]bean2.RegistryCredentials)
	registryDestinationImageMap := make(map[string][]string)
	var allDestinationImages []string //saving all images to be reserved in this array

	for _, step := range postCiSteps {
		if version, ok := pluginIdToVersionMap[step.RefPluginId]; ok {
			destinationImageMap, credentialMap, err := impl.pluginInputVariableParser.HandleCopyContainerImagePluginInputVariables(step.InputVars, customTag, buildImagePath, buildImagedockerRegistryId)
			if err != nil {
				impl.Logger.Errorw("error in parsing copyContainerImage input variable", "err", err)
				return nil, nil, "", nil, err
			}
			if version == buildCommonBean.COPY_CONTAINER_IMAGE_VERSION_V1 {
				// this is needed in ci runner only for v1
				registryDestinationImageMap = destinationImageMap
			}
			for _, images := range destinationImageMap {
				allDestinationImages = append(allDestinationImages, images...)
			}
			for k, v := range credentialMap {
				registryCredentialMap[k] = v
			}
		}
	}

	pluginArtifactStage := repository5.POST_CI
	for _, image := range allDestinationImages {
		if image == buildImagePath {
			return nil, registryCredentialMap, pluginArtifactStage, nil,
				pipelineConfigBean.ErrImagePathInUse
		}
	}

	var imagePathReservationIds []int
	if len(allDestinationImages) > 0 {
		savedCIArtifacts, err := impl.ciArtifactRepository.FindCiArtifactByImagePaths(allDestinationImages)
		if err != nil {
			impl.Logger.Errorw("error in fetching artifacts by image path", "err", err)
			return nil, nil, pluginArtifactStage, nil, err
		}
		if len(savedCIArtifacts) > 0 {
			// if already present in ci artifact, return "image path already in use error"
			return nil, nil, pluginArtifactStage, nil, pipelineConfigBean.ErrImagePathInUse
		}
		imagePathReservationIds, err = impl.ReserveImagesGeneratedAtPlugin(customTagId, allDestinationImages)
		if err != nil {
			return nil, nil, pluginArtifactStage, imagePathReservationIds, err
		}
	}
	return registryDestinationImageMap, registryCredentialMap, pluginArtifactStage, imagePathReservationIds, nil
}

func (impl *HandlerServiceImpl) ReserveImagesGeneratedAtPlugin(customTagId int, destinationImages []string) ([]int, error) {
	var imagePathReservationIds []int
	for _, image := range destinationImages {
		imagePathReservationData, err := impl.customTagService.ReserveImagePath(image, customTagId)
		if err != nil {
			impl.Logger.Errorw("Error in marking custom tag reserved", "err", err)
			return imagePathReservationIds, err
		}
		imagePathReservationIds = append(imagePathReservationIds, imagePathReservationData.Id)
	}
	return imagePathReservationIds, nil
}

func buildCiStepsDataFromDockerBuildScripts(dockerBuildScripts []*bean.CiScript) []*pipelineConfigBean.StepObject {
	// before plugin support, few variables were set as env vars in ci-runner
	// these variables are now moved to global vars in plugin steps, but to avoid error in old scripts adding those variables in payload
	inputVars := []*commonBean.VariableObject{
		{
			Name:                  "DOCKER_IMAGE_TAG",
			Format:                "STRING",
			VariableType:          commonBean.VariableTypeRefGlobal,
			ReferenceVariableName: "DOCKER_IMAGE_TAG",
		},
		{
			Name:                  "DOCKER_REPOSITORY",
			Format:                "STRING",
			VariableType:          commonBean.VariableTypeRefGlobal,
			ReferenceVariableName: "DOCKER_REPOSITORY",
		},
		{
			Name:                  "DOCKER_REGISTRY_URL",
			Format:                "STRING",
			VariableType:          commonBean.VariableTypeRefGlobal,
			ReferenceVariableName: "DOCKER_REGISTRY_URL",
		},
		{
			Name:                  "DOCKER_IMAGE",
			Format:                "STRING",
			VariableType:          commonBean.VariableTypeRefGlobal,
			ReferenceVariableName: "DOCKER_IMAGE",
		},
	}
	var ciSteps []*pipelineConfigBean.StepObject
	for _, dockerBuildScript := range dockerBuildScripts {
		ciStep := &pipelineConfigBean.StepObject{
			Name:          dockerBuildScript.Name,
			Index:         dockerBuildScript.Index,
			Script:        dockerBuildScript.Script,
			ArtifactPaths: []string{dockerBuildScript.OutputLocation},
			StepType:      string(repository.PIPELINE_STEP_TYPE_INLINE),
			ExecutorType:  string(repository2.SCRIPT_TYPE_SHELL),
			InputVars:     inputVars,
		}
		ciSteps = append(ciSteps, ciStep)
	}
	return ciSteps
}

func (impl *HandlerServiceImpl) buildImageTag(commitHashes map[int]pipelineConfig.GitCommit, id int, wfId int) string {
	dockerImageTag := ""
	toAppendDevtronParamInTag := true
	for _, v := range commitHashes {
		if v.WebhookData.Id == 0 {
			if v.Commit == "" {
				continue
			}
			dockerImageTag = getUpdatedDockerImageTagWithCommitOrCheckOutData(dockerImageTag, _getTruncatedImageTag(v.Commit))
		} else {
			_targetCheckout := v.WebhookData.Data[bean.WEBHOOK_SELECTOR_TARGET_CHECKOUT_NAME]
			if _targetCheckout == "" {
				continue
			}
			// if not PR based then meaning tag based
			isPRBasedEvent := v.WebhookData.EventActionType == bean.WEBHOOK_EVENT_MERGED_ACTION_TYPE
			if !isPRBasedEvent && impl.config.CiCdConfig.UseImageTagFromGitProviderForTagBasedBuild {
				dockerImageTag = getUpdatedDockerImageTagWithCommitOrCheckOutData(dockerImageTag, _targetCheckout)
			} else {
				dockerImageTag = getUpdatedDockerImageTagWithCommitOrCheckOutData(dockerImageTag, _getTruncatedImageTag(_targetCheckout))
			}
			if isPRBasedEvent {
				_sourceCheckout := v.WebhookData.Data[bean.WEBHOOK_SELECTOR_SOURCE_CHECKOUT_NAME]
				dockerImageTag = getUpdatedDockerImageTagWithCommitOrCheckOutData(dockerImageTag, _getTruncatedImageTag(_sourceCheckout))
			} else {
				toAppendDevtronParamInTag = !impl.config.CiCdConfig.UseImageTagFromGitProviderForTagBasedBuild
			}
		}
	}
	toAppendDevtronParamInTag = toAppendDevtronParamInTag && dockerImageTag != ""
	if toAppendDevtronParamInTag {
		dockerImageTag = fmt.Sprintf("%s-%d-%d", dockerImageTag, id, wfId)
	}
	// replace / with underscore, as docker image tag doesn't support slash. it gives error
	dockerImageTag = strings.ReplaceAll(dockerImageTag, "/", "_")
	return dockerImageTag
}

func getUpdatedDockerImageTagWithCommitOrCheckOutData(dockerImageTag, commitOrCheckOutData string) string {
	if dockerImageTag == "" {
		dockerImageTag = commitOrCheckOutData
	} else {
		if commitOrCheckOutData != "" {
			dockerImageTag = fmt.Sprintf("%s-%s", dockerImageTag, commitOrCheckOutData)
		}
	}
	return dockerImageTag
}

func (impl *HandlerServiceImpl) updateCiWorkflow(request *types.WorkflowRequest, savedWf *pipelineConfig.CiWorkflow) error {
	ciBuildConfig := request.CiBuildConfig
	ciBuildType := string(ciBuildConfig.CiBuildType)
	savedWf.CiBuildType = ciBuildType
	return impl.ciService.UpdateCiWorkflowWithStage(savedWf)
}

func (impl *HandlerServiceImpl) handleRuntimeParamsValidations(trigger *types.CiTriggerRequest, ciMaterials []*pipelineConfig.CiPipelineMaterial, workflowRequest *types.WorkflowRequest) error {
	// externalCi artifact is meant only for CI_JOB
	if trigger.PipelineType != string(buildCommonBean.CI_JOB) {
		return nil
	}

	// checking if user has given run time parameters for externalCiArtifact, if given then sending git material to Ci-Runner
	externalCiArtifact, exists := trigger.RuntimeParameters.GetGlobalRuntimeVariables()[buildBean.ExtraEnvVarExternalCiArtifactKey]
	// validate externalCiArtifact as docker image
	if exists {
		externalCiArtifact = strings.TrimSpace(externalCiArtifact)
		if !strings.Contains(externalCiArtifact, ":") {
			if utils.IsValidDockerTagName(externalCiArtifact) {
				fullImageUrl, err := utils.BuildDockerImagePath(bean4.DockerRegistryInfo{
					DockerImageTag:     externalCiArtifact,
					DockerRegistryId:   workflowRequest.DockerRegistryId,
					DockerRegistryType: workflowRequest.DockerRegistryType,
					DockerRegistryURL:  workflowRequest.DockerRegistryURL,
					DockerRepository:   workflowRequest.DockerRepository,
				})
				if err != nil {
					impl.Logger.Errorw("Error in building docker image", "err", err)
					return err
				}
				externalCiArtifact = fullImageUrl
			} else {
				impl.Logger.Errorw("validation error", "externalCiArtifact", externalCiArtifact)
				return fmt.Errorf("invalid image name or url given in externalCiArtifact")
			}

		}
		// This will overwrite the existing runtime parameters value for constants.externalCiArtifact
		trigger.RuntimeParameters = trigger.RuntimeParameters.AddRuntimeGlobalVariable(buildBean.ExtraEnvVarExternalCiArtifactKey, externalCiArtifact)
		var artifactExists bool
		var err error

		imageDigest, ok := trigger.RuntimeParameters.GetGlobalRuntimeVariables()[buildBean.ExtraEnvVarImageDigestKey]
		if !ok || len(imageDigest) == 0 {
			artifactExists, err = impl.ciArtifactRepository.IfArtifactExistByImage(externalCiArtifact, trigger.PipelineId)
			if err != nil {
				impl.Logger.Errorw("error in fetching ci artifact", "err", err)
				return err
			}
			if artifactExists {
				impl.Logger.Errorw("ci artifact already exists with same image name", "artifact", externalCiArtifact)
				return fmt.Errorf("ci artifact already exists with same image name")
			}
			workflowRequest, err = impl.updateWorkflowRequestForDigestPull(trigger.PipelineId, workflowRequest)
			if err != nil {
				impl.Logger.Errorw("error in updating workflow request", "err", err)
				return err
			}
		} else {
			artifactExists, err = impl.ciArtifactRepository.IfArtifactExistByImageDigest(imageDigest, externalCiArtifact, trigger.PipelineId)
			if err != nil {
				impl.Logger.Errorw("error in fetching ci artifact", "err", err, "imageDigest", imageDigest)
				return err
			}
			if artifactExists {
				impl.Logger.Errorw("ci artifact already exists  with same digest", "artifact", externalCiArtifact)
				return fmt.Errorf("ci artifact already exists  with same digest")
			}
		}
	}
	if trigger.PipelineType == string(buildCommonBean.CI_JOB) && len(ciMaterials) != 0 && !exists && externalCiArtifact == "" {
		ciMaterials[0].GitMaterial = nil
		ciMaterials[0].GitMaterialId = 0
	}
	return nil
}

func _getTruncatedImageTag(imageTag string) string {
	_length := len(imageTag)
	if _length == 0 {
		return imageTag
	}

	_truncatedLength := 8

	if _length < _truncatedLength {
		return imageTag
	} else {
		return imageTag[:_truncatedLength]
	}
}

func (impl *HandlerServiceImpl) markCurrentCiWorkflowFailed(savedCiWf *pipelineConfig.CiWorkflow, validationErr error) error {
	// currently such requirement is not there
	if savedCiWf == nil {
		return nil
	}
	if savedCiWf.Id != 0 && slices.Contains(cdWorkflow.WfrTerminalStatusList, savedCiWf.Status) {
		impl.Logger.Debug("workflow is already in terminal state", "status", savedCiWf.Status, "workflowId", savedCiWf.Id, "message", savedCiWf.Message)
		return nil
	}

	savedCiWf.Status = cdWorkflow.WorkflowFailed
	savedCiWf.Message = validationErr.Error()
	savedCiWf.FinishedOn = time.Now()

	var dbErr error
	if savedCiWf.Id == 0 {
		dbErr = impl.ciService.SaveCiWorkflowWithStage(savedCiWf)
	} else {
		dbErr = impl.ciService.UpdateCiWorkflowWithStage(savedCiWf)
	}

	if dbErr != nil {
		impl.Logger.Errorw("save/update workflow error", "err", dbErr)
		return dbErr
	}
	return nil
}

func (impl *HandlerServiceImpl) CancelBuild(workflowId int, forceAbort bool) (int, error) {
	workflow, err := impl.ciWorkflowRepository.FindById(workflowId)
	if err != nil {
		impl.Logger.Errorw("error in finding ci-workflow by workflow id", "ciWorkflowId", workflowId, "err", err)
		return 0, err
	}
	isExt := workflow.Namespace != constants2.DefaultCiWorkflowNamespace
	var env *repository6.Environment
	var restConfig *rest.Config
	if isExt {
		restConfig, err = impl.getRestConfig(workflow)
		if err != nil {
			return 0, err
		}
	}
	// Terminate workflow
	cancelWfDtoRequest := &types.CancelWfRequestDto{
		ExecutorType: workflow.ExecutorType,
		WorkflowName: workflow.Name,
		Namespace:    workflow.Namespace,
		RestConfig:   restConfig,
		IsExt:        isExt,
		Environment:  env,
	}
	// Terminate workflow
	err = impl.workflowService.TerminateWorkflow(cancelWfDtoRequest)
	if err != nil && forceAbort {
		impl.Logger.Errorw("error in terminating workflow, with force abort flag flag as true", "workflowName", workflow.Name, "err", err)

		cancelWfDtoRequest.WorkflowGenerateName = fmt.Sprintf("%d-%s", workflowId, workflow.Name)
		err1 := impl.workflowService.TerminateDanglingWorkflows(cancelWfDtoRequest)
		if err1 != nil {
			impl.Logger.Errorw("error in terminating dangling workflows", "cancelWfDtoRequest", cancelWfDtoRequest, "err", err)
			// ignoring error here in case of force abort, confirmed from product
		}
	} else if err != nil && strings.Contains(err.Error(), "cannot find workflow") {
		return 0, &util.ApiError{Code: "200", HttpStatusCode: http.StatusBadRequest, UserMessage: err.Error()}
	} else if err != nil {
		impl.Logger.Errorw("cannot terminate wf", "err", err)
		return 0, err
	}
	if forceAbort {
		err = impl.handleForceAbortCaseForCi(workflow, forceAbort)
		if err != nil {
			impl.Logger.Errorw("error in handleForceAbortCaseForCi", "forceAbortFlag", forceAbort, "workflow", workflow, "err", err)
			return 0, err
		}
		return workflow.Id, nil
	}

	workflow.Status = cdWorkflow.WorkflowCancel
	if workflow.ExecutorType == cdWorkflow.WORKFLOW_EXECUTOR_TYPE_SYSTEM {
		workflow.PodStatus = "Failed"
		workflow.Message = constants2.TERMINATE_MESSAGE
	}
	err = impl.ciService.UpdateCiWorkflowWithStage(workflow)
	if err != nil {
		impl.Logger.Errorw("cannot update deleted workflow status, but wf deleted", "err", err)
		return 0, err
	}
	imagePathReservationId := workflow.ImagePathReservationId
	err = impl.customTagService.DeactivateImagePathReservation(imagePathReservationId)
	if err != nil {
		impl.Logger.Errorw("error in marking image tag unreserved", "err", err)
		return 0, err
	}
	imagePathReservationIds := workflow.ImagePathReservationIds
	if len(imagePathReservationIds) > 0 {
		err = impl.customTagService.DeactivateImagePathReservationByImageIds(imagePathReservationIds)
		if err != nil {
			impl.Logger.Errorw("error in marking image tag unreserved", "err", err)
			return 0, err
		}
	}
	return workflow.Id, nil
}

func (impl *HandlerServiceImpl) handleForceAbortCaseForCi(workflow *pipelineConfig.CiWorkflow, forceAbort bool) error {
	isWorkflowInNonTerminalStage := workflow.Status == string(v1alpha1.NodePending) || workflow.Status == string(v1alpha1.NodeRunning)
	if !isWorkflowInNonTerminalStage {
		if forceAbort {
			return impl.updateWorkflowForForceAbort(workflow)
		} else {
			return &util.ApiError{Code: "200", HttpStatusCode: 400, UserMessage: "cannot cancel build, build not in progress"}
		}
	}
	//this arises when someone deletes the workflow in resource browser and wants to force abort a ci
	if workflow.Status == string(v1alpha1.NodeRunning) && forceAbort {
		return impl.updateWorkflowForForceAbort(workflow)
	}
	return nil
}

func (impl *HandlerServiceImpl) updateWorkflowForForceAbort(workflow *pipelineConfig.CiWorkflow) error {
	workflow.Status = cdWorkflow.WorkflowCancel
	workflow.PodStatus = string(bean.Failed)
	workflow.Message = constants2.FORCE_ABORT_MESSAGE_AFTER_STARTING_STAGE
	err := impl.ciService.UpdateCiWorkflowWithStage(workflow)
	if err != nil {
		impl.Logger.Errorw("error in updating workflow status", "err", err)
		return err
	}
	return nil
}

func (impl *HandlerServiceImpl) getRestConfig(workflow *pipelineConfig.CiWorkflow) (*rest.Config, error) {
	env, err := impl.envRepository.FindById(workflow.EnvironmentId)
	if err != nil {
		impl.Logger.Errorw("could not fetch stage env", "err", err)
		return nil, err
	}

	clusterBean := adapter2.GetClusterBean(*env.Cluster)

	clusterConfig := clusterBean.GetClusterConfig()
	restConfig, err := impl.K8sUtil.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.Logger.Errorw("error in getting rest config by cluster id", "err", err)
		return nil, err
	}
	return restConfig, nil
}

func (impl *HandlerServiceImpl) GetRunningWorkflowLogs(workflowId int, followLogs bool) (*bufio.Reader, func() error, error) {
	ciWorkflow, err := impl.ciWorkflowRepository.FindById(workflowId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return nil, nil, err
	}
	return impl.getWorkflowLogs(ciWorkflow, followLogs)
}

func (impl *HandlerServiceImpl) getWorkflowLogs(ciWorkflow *pipelineConfig.CiWorkflow, followLogs bool) (*bufio.Reader, func() error, error) {
	if string(v1alpha1.NodePending) == ciWorkflow.PodStatus {
		return bufio.NewReader(strings.NewReader("")), func() error { return nil }, nil
	}
	ciLogRequest := types.BuildLogRequest{
		PodName:   ciWorkflow.PodName,
		Namespace: ciWorkflow.Namespace,
	}
	isExt := false
	clusterConfig := &k8s.ClusterConfig{}
	if ciWorkflow.EnvironmentId != 0 {
		env, err := impl.envRepository.FindById(ciWorkflow.EnvironmentId)
		if err != nil {
			return nil, nil, err
		}
		var clusterBean clusterBean.ClusterBean
		if env != nil && env.Cluster != nil {
			clusterBean = adapter2.GetClusterBean(*env.Cluster)
		}
		clusterConfig = clusterBean.GetClusterConfig()
		isExt = true
	}

	logStream, cleanUp, err := impl.ciLogService.FetchRunningWorkflowLogs(ciLogRequest, clusterConfig, isExt, followLogs)
	if logStream == nil || err != nil {
		if !ciWorkflow.BlobStorageEnabled {
			return nil, nil, &util.ApiError{Code: "200", HttpStatusCode: 400, UserMessage: "logs-not-stored-in-repository"}
		} else if string(v1alpha1.NodeSucceeded) == ciWorkflow.Status || string(v1alpha1.NodeError) == ciWorkflow.Status || string(v1alpha1.NodeFailed) == ciWorkflow.Status || ciWorkflow.Status == cdWorkflow.WorkflowCancel {
			impl.Logger.Debugw("pod is not live", "podName", ciWorkflow.PodName, "err", err)
			return impl.getLogsFromRepository(ciWorkflow, clusterConfig, isExt)
		}
		if err != nil {
			impl.Logger.Errorw("err on fetch workflow logs", "err", err)
			return nil, nil, &util.ApiError{Code: "200", HttpStatusCode: 400, UserMessage: err.Error()}
		} else if logStream == nil {
			return nil, cleanUp, fmt.Errorf("no logs found for pod %s", ciWorkflow.PodName)
		}
	}
	logReader := bufio.NewReader(logStream)
	return logReader, cleanUp, err
}

func (impl *HandlerServiceImpl) getLogsFromRepository(ciWorkflow *pipelineConfig.CiWorkflow, clusterConfig *k8s.ClusterConfig, isExt bool) (*bufio.Reader, func() error, error) {
	impl.Logger.Debug("getting historic logs", "ciWorkflowId", ciWorkflow.Id)
	ciConfigLogsBucket := impl.config.GetDefaultBuildLogsBucket()
	ciConfigCiCacheRegion := impl.config.DefaultCacheBucketRegion
	logsFilePath := impl.config.GetDefaultBuildLogsKeyPrefix() + "/" + ciWorkflow.Name + "/main.log" // this is for backward compatibilty
	if strings.Contains(ciWorkflow.LogLocation, "main.log") {
		logsFilePath = ciWorkflow.LogLocation
	}
	ciLogRequest := types.BuildLogRequest{
		PipelineId:    ciWorkflow.CiPipelineId,
		WorkflowId:    ciWorkflow.Id,
		PodName:       ciWorkflow.PodName,
		LogsFilePath:  logsFilePath,
		CloudProvider: impl.config.CloudProvider,
		AzureBlobConfig: &blob_storage.AzureBlobBaseConfig{
			Enabled:           impl.config.CloudProvider == types.BLOB_STORAGE_AZURE,
			AccountName:       impl.config.AzureAccountName,
			BlobContainerName: impl.config.AzureBlobContainerCiLog,
			AccountKey:        impl.config.AzureAccountKey,
		},
		AwsS3BaseConfig: &blob_storage.AwsS3BaseConfig{
			AccessKey:         impl.config.BlobStorageS3AccessKey,
			Passkey:           impl.config.BlobStorageS3SecretKey,
			EndpointUrl:       impl.config.BlobStorageS3Endpoint,
			IsInSecure:        impl.config.BlobStorageS3EndpointInsecure,
			BucketName:        ciConfigLogsBucket,
			Region:            ciConfigCiCacheRegion,
			VersioningEnabled: impl.config.BlobStorageS3BucketVersioned,
		},
		GcpBlobBaseConfig: &blob_storage.GcpBlobBaseConfig{
			BucketName:             ciConfigLogsBucket,
			CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
		},
	}
	useExternalBlobStorage := pipeline2.IsExternalBlobStorageEnabled(isExt, impl.config.UseBlobStorageConfigInCiWorkflow)
	if useExternalBlobStorage {
		// fetch extClusterBlob cm and cs from k8s client, if they are present then read creds
		// from them else return.
		cmConfig, secretConfig, err := impl.blobConfigStorageService.FetchCmAndSecretBlobConfigFromExternalCluster(clusterConfig, ciWorkflow.Namespace)
		if err != nil {
			impl.Logger.Errorw("error in fetching config map and secret from external cluster", "err", err, "clusterConfig", clusterConfig)
			return nil, nil, err
		}
		rq := &ciLogRequest
		rq.SetBuildLogRequest(cmConfig, secretConfig)
	}
	oldLogsStream, cleanUp, err := impl.ciLogService.FetchLogs(impl.config.BaseLogLocationPath, ciLogRequest)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return nil, nil, err
	}
	logReader := bufio.NewReader(oldLogsStream)
	return logReader, cleanUp, err
}

func (impl *HandlerServiceImpl) GetHistoricBuildLogs(workflowId int, ciWorkflow *pipelineConfig.CiWorkflow) (map[string]string, error) {
	var err error
	if ciWorkflow == nil {
		ciWorkflow, err = impl.ciWorkflowRepository.FindById(workflowId)
		if err != nil {
			impl.Logger.Errorw("err", "err", err)
			return nil, err
		}
	}
	ciConfigLogsBucket := impl.config.GetDefaultBuildLogsBucket()
	ciConfigCiCacheRegion := impl.config.DefaultCacheBucketRegion
	ciLogRequest := types.BuildLogRequest{
		PipelineId:    ciWorkflow.CiPipelineId,
		WorkflowId:    ciWorkflow.Id,
		PodName:       ciWorkflow.PodName,
		LogsFilePath:  ciWorkflow.LogLocation,
		CloudProvider: impl.config.CloudProvider,
		AzureBlobConfig: &blob_storage.AzureBlobBaseConfig{
			Enabled:           impl.config.CloudProvider == types.BLOB_STORAGE_AZURE,
			AccountName:       impl.config.AzureAccountName,
			BlobContainerName: impl.config.AzureBlobContainerCiLog,
			AccountKey:        impl.config.AzureAccountKey,
		},
		AwsS3BaseConfig: &blob_storage.AwsS3BaseConfig{
			AccessKey:         impl.config.BlobStorageS3AccessKey,
			Passkey:           impl.config.BlobStorageS3SecretKey,
			EndpointUrl:       impl.config.BlobStorageS3Endpoint,
			IsInSecure:        impl.config.BlobStorageS3EndpointInsecure,
			BucketName:        ciConfigLogsBucket,
			Region:            ciConfigCiCacheRegion,
			VersioningEnabled: impl.config.BlobStorageS3BucketVersioned,
		},
		GcpBlobBaseConfig: &blob_storage.GcpBlobBaseConfig{
			BucketName:             ciConfigLogsBucket,
			CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
		},
	}
	useExternalBlobStorage := pipeline2.IsExternalBlobStorageEnabled(ciWorkflow.IsExternalRunInJobType(), impl.config.UseBlobStorageConfigInCiWorkflow)
	if useExternalBlobStorage {
		envBean, err := impl.envService.FindById(ciWorkflow.EnvironmentId)
		if err != nil {
			impl.Logger.Errorw("error in getting envBean by envId", "err", err, "envId", ciWorkflow.EnvironmentId)
			return nil, err
		}
		clusterConfig, err := impl.clusterService.GetClusterConfigByClusterId(envBean.ClusterId)
		if err != nil {
			impl.Logger.Errorw("GetClusterConfigByClusterId, error in fetching clusterConfig by clusterId", "err", err, "clusterId", envBean.ClusterId)
			return nil, err
		}
		// fetch extClusterBlob cm and cs from k8s client, if they are present then read creds
		// from them else return.
		cmConfig, secretConfig, err := impl.blobConfigStorageService.FetchCmAndSecretBlobConfigFromExternalCluster(clusterConfig, ciWorkflow.Namespace)
		if err != nil {
			impl.Logger.Errorw("error in fetching config map and secret from external cluster", "err", err, "clusterConfig", clusterConfig)
			return nil, err
		}
		rq := &ciLogRequest
		rq.SetBuildLogRequest(cmConfig, secretConfig)
	}
	logsFile, cleanUp, err := impl.ciLogService.FetchLogs(impl.config.BaseLogLocationPath, ciLogRequest)
	logs, err := ioutil.ReadFile(logsFile.Name())
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return map[string]string{}, err
	}
	logStr := string(logs)
	resp := make(map[string]string)
	resp["logs"] = logStr
	defer cleanUp()
	return resp, err
}

func (impl *HandlerServiceImpl) DownloadCiWorkflowArtifacts(pipelineId int, buildId int) (*os.File, error) {
	ciWorkflow, err := impl.ciWorkflowRepository.FindById(buildId)
	if err != nil {
		impl.Logger.Errorw("unable to fetch ciWorkflow", "err", err)
		return nil, err
	}
	useExternalBlobStorage := pipeline2.IsExternalBlobStorageEnabled(ciWorkflow.IsExternalRunInJobType(), impl.config.UseBlobStorageConfigInCiWorkflow)
	if !ciWorkflow.BlobStorageEnabled {
		return nil, errors.New("logs-not-stored-in-repository")
	}

	if ciWorkflow.CiPipelineId != pipelineId {
		impl.Logger.Error("invalid request, wf not in pipeline")
		return nil, errors.New("invalid request, wf not in pipeline")
	}

	ciConfigLogsBucket := impl.config.GetDefaultBuildLogsBucket()
	item := strconv.Itoa(ciWorkflow.Id)
	ciConfigCiCacheRegion := impl.config.DefaultCacheBucketRegion
	azureBlobConfig := &blob_storage.AzureBlobBaseConfig{
		Enabled:           impl.config.CloudProvider == types.BLOB_STORAGE_AZURE,
		AccountName:       impl.config.AzureAccountName,
		BlobContainerName: impl.config.AzureBlobContainerCiLog,
		AccountKey:        impl.config.AzureAccountKey,
	}
	awsS3BaseConfig := &blob_storage.AwsS3BaseConfig{
		AccessKey:         impl.config.BlobStorageS3AccessKey,
		Passkey:           impl.config.BlobStorageS3SecretKey,
		EndpointUrl:       impl.config.BlobStorageS3Endpoint,
		IsInSecure:        impl.config.BlobStorageS3EndpointInsecure,
		BucketName:        ciConfigLogsBucket,
		Region:            ciConfigCiCacheRegion,
		VersioningEnabled: impl.config.BlobStorageS3BucketVersioned,
	}
	gcpBlobBaseConfig := &blob_storage.GcpBlobBaseConfig{
		BucketName:             ciConfigLogsBucket,
		CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
	}

	ciArtifactLocationFormat := impl.config.GetArtifactLocationFormat()
	key := fmt.Sprintf(ciArtifactLocationFormat, ciWorkflow.Id, ciWorkflow.Id)
	if len(ciWorkflow.CiArtifactLocation) != 0 && util2.IsValidUrlSubPath(ciWorkflow.CiArtifactLocation) {
		key = ciWorkflow.CiArtifactLocation
	} else if util2.IsValidUrlSubPath(key) {
		impl.ciWorkflowRepository.MigrateCiArtifactLocation(ciWorkflow.Id, key)
	}
	baseLogLocationPathConfig := impl.config.BaseLogLocationPath
	blobStorageService := blob_storage.NewBlobStorageServiceImpl(impl.Logger)
	destinationKey := filepath.Clean(filepath.Join(baseLogLocationPathConfig, item))
	request := &blob_storage.BlobStorageRequest{
		StorageType:         impl.config.CloudProvider,
		SourceKey:           key,
		DestinationKey:      destinationKey,
		AzureBlobBaseConfig: azureBlobConfig,
		AwsS3BaseConfig:     awsS3BaseConfig,
		GcpBlobBaseConfig:   gcpBlobBaseConfig,
	}
	if useExternalBlobStorage {
		envBean, err := impl.envService.FindById(ciWorkflow.EnvironmentId)
		if err != nil {
			impl.Logger.Errorw("error in getting envBean by envId", "err", err, "envId", ciWorkflow.EnvironmentId)
			return nil, err
		}
		clusterConfig, err := impl.clusterService.GetClusterConfigByClusterId(envBean.ClusterId)
		if err != nil {
			impl.Logger.Errorw("GetClusterConfigByClusterId, error in fetching clusterConfig by clusterId", "err", err, "clusterId", envBean.ClusterId)
			return nil, err
		}
		// fetch extClusterBlob cm and cs from k8s client, if they are present then read creds
		// from them else return.
		cmConfig, secretConfig, err := impl.blobConfigStorageService.FetchCmAndSecretBlobConfigFromExternalCluster(clusterConfig, ciWorkflow.Namespace)
		if err != nil {
			impl.Logger.Errorw("error in fetching config map and secret from external cluster", "err", err, "clusterConfig", clusterConfig)
			return nil, err
		}
		request = pipeline2.UpdateRequestWithExtClusterCmAndSecret(request, cmConfig, secretConfig)
	}
	_, numBytes, err := blobStorageService.Get(request)
	if err != nil {
		impl.Logger.Errorw("error occurred while downloading file", "request", request, "error", err)
		return nil, errors.New("failed to download resource")
	}

	file, err := os.Open(destinationKey)
	if err != nil {
		impl.Logger.Errorw("unable to open file", "file", item, "err", err)
		return nil, errors.New("unable to open file")
	}

	impl.Logger.Infow("Downloaded ", "filename", file.Name(), "bytes", numBytes)
	return file, nil
}

// abortPreviousRunningBuilds checks if auto-abort is enabled for the pipeline and aborts previous running builds
func (impl *HandlerServiceImpl) abortPreviousRunningBuilds(pipelineId int, triggeredBy int32) error {
	// Get pipeline configuration to check if auto-abort is enabled
	ciPipeline, err := impl.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		impl.Logger.Errorw("error in finding ci pipeline", "pipelineId", pipelineId, "err", err)
		return err
	}

	// Check if auto-abort is enabled for this pipeline
	if !ciPipeline.AutoAbortPreviousBuilds {
		impl.Logger.Debugw("auto-abort not enabled for pipeline", "pipelineId", pipelineId)
		return nil
	}

	// Find all running/pending workflows for this pipeline
	runningWorkflows, err := impl.ciWorkflowRepository.FindRunningWorkflowsForPipeline(pipelineId)
	if err != nil {
		impl.Logger.Errorw("error in finding running workflows for pipeline", "pipelineId", pipelineId, "err", err)
		return err
	}

	if len(runningWorkflows) == 0 {
		impl.Logger.Debugw("no running workflows found to abort for pipeline", "pipelineId", pipelineId)
		return nil
	}

	impl.Logger.Infow("found running workflows to abort due to auto-abort configuration", 
		"pipelineId", pipelineId, "workflowCount", len(runningWorkflows), "triggeredBy", triggeredBy)

	// Abort each running workflow
	for _, workflow := range runningWorkflows {
		// Check if the workflow is in a critical phase that should not be aborted
		if impl.isWorkflowInCriticalPhase(workflow) {
			impl.Logger.Infow("skipping abort of workflow in critical phase", 
				"workflowId", workflow.Id, "status", workflow.Status, "pipelineId", pipelineId)
			continue
		}

		// Attempt to cancel the build
		_, err := impl.CancelBuild(workflow.Id, false)
		if err != nil {
			impl.Logger.Errorw("error aborting previous running build", 
				"workflowId", workflow.Id, "pipelineId", pipelineId, "err", err)
			// Continue with other workflows even if one fails
			continue
		}

		impl.Logger.Infow("successfully aborted previous running build due to auto-abort", 
			"workflowId", workflow.Id, "pipelineId", pipelineId, "abortedBy", triggeredBy)
	}

	return nil
}

// isWorkflowInCriticalPhase determines if a workflow is in a critical phase and should not be aborted
// This protects builds that are in the final stages like pushing cache or artifacts
func (impl *HandlerServiceImpl) isWorkflowInCriticalPhase(workflow *pipelineConfig.CiWorkflow) bool {
	// For now, we consider "Starting" as safe to abort, but "Running" needs more careful consideration
	// In the future, this could be extended to check actual workflow steps/stages
	
	// If workflow has been running for less than 2 minutes, it's likely still in setup phase
	if workflow.Status == "Running" && workflow.StartedOn.IsZero() == false {
		runningDuration := time.Since(workflow.StartedOn)
		if runningDuration < 2*time.Minute {
			impl.Logger.Debugw("workflow is in early running phase, safe to abort", 
				"workflowId", workflow.Id, "runningDuration", runningDuration.String())
			return false
		}
		
		// For workflows running longer, we should be more cautious
		// This could be extended to check actual workflow phases using workflow service APIs
		impl.Logger.Debugw("workflow has been running for a while, considering as critical phase", 
			"workflowId", workflow.Id, "runningDuration", runningDuration.String())
		return true
	}
	
	// "Starting" and "Pending" are generally safe to abort
	return false
}
