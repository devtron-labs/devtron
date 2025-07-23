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

package pipeline

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/common-lib/utils"
	bean4 "github.com/devtron-labs/common-lib/utils/bean"
	pipelineAdapter "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/adapter/cdWorkflow"
	cdWorkflowBean "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	"github.com/devtron-labs/devtron/pkg/build/artifacts/imageTagging"
	buildBean "github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	repository3 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	common2 "github.com/devtron-labs/devtron/pkg/deployment/common"
	eventProcessorBean "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus"
	bean5 "github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus/bean"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	"github.com/devtron-labs/devtron/pkg/workflow/workflowStatusLatest"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	pipelineBean "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	resourceGroup2 "github.com/devtron-labs/devtron/pkg/resourceGroup"
	globalUtil "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

const (
	DEVTRON_APP_HELM_PIPELINE_STATUS_UPDATE_CRON = "DTAppHelmPipelineStatusUpdateCron"
	DEVTRON_APP_ARGO_PIPELINE_STATUS_UPDATE_CRON = "DTAppArgoPipelineStatusUpdateCron"
	DEVTRON_APP_FLUX_PIPELINE_STATUS_UPDATE_CRON = "DTAppFluxPipelineStatusUpdateCron"
	HELM_APP_ARGO_PIPELINE_STATUS_UPDATE_CRON    = "HelmAppArgoPipelineStatusUpdateCron"
)

type CdHandler interface {
	UpdateWorkflow(workflowStatus eventProcessorBean.CiCdStatus) (int, string, bool, string, error)
	GetCdBuildHistory(appId int, environmentId int, pipelineId int, offset int, size int) ([]pipelineBean.CdWorkflowWithArtifact, error)
	FetchCdWorkflowDetails(appId int, environmentId int, pipelineId int, buildId int) (types.WorkflowResponse, error)
	FetchCdPrePostStageStatus(pipelineId int) ([]pipelineBean.CdWorkflowWithArtifact, error)
	FetchAppWorkflowStatusForTriggerView(appId int) ([]*pipelineConfig.CdWorkflowStatus, error)
	FetchAppWorkflowStatusForTriggerViewForEnvironment(request resourceGroup2.ResourceGroupingRequest, token string) ([]*pipelineConfig.CdWorkflowStatus, error)
	FetchAppDeploymentStatusForEnvironments(request resourceGroup2.ResourceGroupingRequest, token string) ([]*pipelineConfig.AppDeploymentStatus, error)
	DeactivateImageReservationPathsOnFailure(imagePathReservationIds []int) error
}

type CdHandlerImpl struct {
	Logger                       *zap.SugaredLogger
	userService                  user.UserService
	ciArtifactRepository         repository.CiArtifactRepository
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
	cdWorkflowRepository         pipelineConfig.CdWorkflowRepository
	envRepository                repository3.EnvironmentRepository
	pipelineRepository           pipelineConfig.PipelineRepository
	ciWorkflowRepository         pipelineConfig.CiWorkflowRepository
	enforcerUtil                 rbac.EnforcerUtil
	resourceGroupService         resourceGroup2.ResourceGroupService
	imageTaggingService          imageTagging.ImageTaggingService
	k8sUtil                      *k8s.K8sServiceImpl
	config                       *types.CdConfig
	customTagService             CustomTagService
	deploymentConfigService      common2.DeploymentConfigService
	workflowStageStatusService   workflowStatus.WorkFlowStageStatusService
	cdWorkflowRunnerService      cd.CdWorkflowRunnerService
	WorkflowStatusLatestService  workflowStatusLatest.WorkflowStatusLatestService
	pipelineStageRepository      repository2.PipelineStageRepository
}

func NewCdHandlerImpl(Logger *zap.SugaredLogger, userService user.UserService,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	ciArtifactRepository repository.CiArtifactRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	pipelineRepository pipelineConfig.PipelineRepository, envRepository repository3.EnvironmentRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository, enforcerUtil rbac.EnforcerUtil,
	resourceGroupService resourceGroup2.ResourceGroupService,
	imageTaggingService imageTagging.ImageTaggingService, k8sUtil *k8s.K8sServiceImpl,
	customTagService CustomTagService,
	deploymentConfigService common2.DeploymentConfigService,
	workflowStageStatusService workflowStatus.WorkFlowStageStatusService,
	cdWorkflowRunnerService cd.CdWorkflowRunnerService,
	WorkflowStatusLatestService workflowStatusLatest.WorkflowStatusLatestService,
	pipelineStageRepository repository2.PipelineStageRepository,
) *CdHandlerImpl {
	cdh := &CdHandlerImpl{
		Logger:                       Logger,
		userService:                  userService,
		cdWorkflowRepository:         cdWorkflowRepository,
		ciArtifactRepository:         ciArtifactRepository,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		envRepository:                envRepository,
		pipelineRepository:           pipelineRepository,
		ciWorkflowRepository:         ciWorkflowRepository,
		enforcerUtil:                 enforcerUtil,
		resourceGroupService:         resourceGroupService,
		imageTaggingService:          imageTaggingService,
		k8sUtil:                      k8sUtil,
		customTagService:             customTagService,
		deploymentConfigService:      deploymentConfigService,
		workflowStageStatusService:   workflowStageStatusService,
		cdWorkflowRunnerService:      cdWorkflowRunnerService,
		WorkflowStatusLatestService:  WorkflowStatusLatestService,
		pipelineStageRepository:      pipelineStageRepository,
	}
	config, err := types.GetCdConfig()
	if err != nil {
		return nil
	}
	cdh.config = config
	return cdh
}

func (impl *CdHandlerImpl) UpdateWorkflow(workflowStatus eventProcessorBean.CiCdStatus) (int, string, bool, string, error) {
	wfStatusRs := impl.extractWorkflowStatus(workflowStatus)
	workflowName, status, podStatus, message, podName := wfStatusRs.WorkflowName, wfStatusRs.Status, wfStatusRs.PodStatus, wfStatusRs.Message, wfStatusRs.PodName
	impl.Logger.Debugw("cd workflow status update event for", "wf ", workflowName, "status", status)
	if workflowName == "" {
		return 0, "", false, "", errors.New("invalid wf name")
	}
	workflowId, err := strconv.Atoi(workflowName[:strings.Index(workflowName, "-")])
	if err != nil {
		impl.Logger.Errorw("invalid wf status update req", "workflowName", workflowName, "err", err)
		return 0, "", false, "", err
	}

	savedWorkflow, err := impl.cdWorkflowRepository.FindPreOrPostCdWorkflowRunnerById(workflowId)
	if err != nil {
		impl.Logger.Error("cannot get saved wf", "workflowId", workflowId, "err", err)
		return 0, "", false, "", err
	}

	cdArtifactLocationFormat := impl.config.GetArtifactLocationFormat()
	cdArtifactLocation := fmt.Sprintf(cdArtifactLocationFormat, savedWorkflow.CdWorkflowId, savedWorkflow.Id)
	if impl.stateChanged(status, podStatus, message, workflowStatus.FinishedAt.Time, savedWorkflow) {
		if !slices.Contains(cdWorkflowBean.WfrTerminalStatusList, savedWorkflow.PodStatus) {
			savedWorkflow.Message = message
			if !slices.Contains(cdWorkflowBean.WfrTerminalStatusList, savedWorkflow.Status) {
				savedWorkflow.FinishedOn = workflowStatus.FinishedAt.Time
			}
		} else {
			impl.Logger.Warnw("cd stage already in terminal state. skipped message and finishedOn from being updated",
				"wfId", savedWorkflow.Id, "podStatus", savedWorkflow.PodStatus, "status", savedWorkflow.Status, "message", message, "finishedOn", workflowStatus.FinishedAt.Time)
		}
		if savedWorkflow.Status != cdWorkflowBean.WorkflowCancel {
			savedWorkflow.Status = status
		}
		savedWorkflow.CdArtifactLocation = cdArtifactLocation
		savedWorkflow.PodStatus = podStatus
		savedWorkflow.Name = workflowName
		// removed log location from here since we are saving it at trigger
		savedWorkflow.PodName = podName
		savedWorkflow.UpdateAuditLog(1)
		impl.Logger.Debugw("updating cd workflow runner", "workflow", savedWorkflow)
		err = impl.cdWorkflowRunnerService.UpdateCdWorkflowRunnerWithStage(savedWorkflow)
		if err != nil {
			impl.Logger.Errorw("update wf failed for id", "wfId", savedWorkflow.Id, "err", err)
			return savedWorkflow.Id, "", true, "", err
		}
		appId := savedWorkflow.CdWorkflow.Pipeline.AppId
		envId := savedWorkflow.CdWorkflow.Pipeline.EnvironmentId
		envDeploymentConfig, err := impl.deploymentConfigService.GetConfigForDevtronApps(nil, appId, envId)
		if err != nil {
			impl.Logger.Errorw("error in fetching environment deployment config by appId and envId", "appId", appId, "envId", envId, "err", err)
			return savedWorkflow.Id, savedWorkflow.Status, true, "", err
		}
		globalUtil.TriggerCDMetrics(pipelineAdapter.GetTriggerMetricsFromRunnerObj(savedWorkflow, envDeploymentConfig), impl.config.ExposeCDMetrics)
		if string(v1alpha1.NodeError) == savedWorkflow.Status || string(v1alpha1.NodeFailed) == savedWorkflow.Status {
			impl.Logger.Warnw("cd stage failed for workflow", "wfId", savedWorkflow.Id)
		}
		return savedWorkflow.Id, savedWorkflow.Status, true, message, nil
	}
	return savedWorkflow.Id, status, false, message, nil
}

func (impl *CdHandlerImpl) extractWorkflowStatus(workflowStatus eventProcessorBean.CiCdStatus) *types.WorkflowStatus {
	workflowName := ""
	status := string(workflowStatus.Phase)
	podStatus := "Pending"
	message := ""
	logLocation := ""
	podName := ""
	for k, v := range workflowStatus.Nodes {
		impl.Logger.Debugw("ExtractWorkflowStatus", "workflowName", k, "v", v)
		if v.TemplateName == pipelineBean.CD_WORKFLOW_NAME {
			if v.BoundaryID == "" {
				workflowName = k
			} else {
				workflowName = v.BoundaryID
			}
			podName = k
			podStatus = string(v.Phase)
			message = v.Message
			if v.Outputs != nil && len(v.Outputs.Artifacts) > 0 {
				if v.Outputs.Artifacts[0].S3 != nil {
					logLocation = v.Outputs.Artifacts[0].S3.Key
				} else if v.Outputs.Artifacts[0].GCS != nil {
					logLocation = v.Outputs.Artifacts[0].GCS.Key
				}
			}
			break
		}
	}
	workflowStatusRes := &types.WorkflowStatus{
		WorkflowName: workflowName,
		Status:       status,
		PodStatus:    podStatus,
		Message:      message,
		LogLocation:  logLocation,
		PodName:      podName,
	}
	return workflowStatusRes
}

func (impl *CdHandlerImpl) stateChanged(status string, podStatus string, msg string,
	finishedAt time.Time, savedWorkflow *pipelineConfig.CdWorkflowRunner) bool {
	return savedWorkflow.Status != status || savedWorkflow.PodStatus != podStatus || savedWorkflow.Message != msg || savedWorkflow.FinishedOn != finishedAt
}

func (impl *CdHandlerImpl) GetCdBuildHistory(appId int, environmentId int, pipelineId int, offset int, size int) ([]pipelineBean.CdWorkflowWithArtifact, error) {

	var cdWorkflowArtifact []pipelineBean.CdWorkflowWithArtifact
	// this map contains artifactId -> array of tags of that artifact
	imageTagsDataMap, err := impl.imageTaggingService.GetTagsDataMapByAppId(appId)
	if err != nil {
		impl.Logger.Errorw("error in fetching image tags with appId", "err", err, "appId", appId)
		return cdWorkflowArtifact, err
	}
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
		if err == pg.ErrNoRows || wfrList == nil {
			return cdWorkflowArtifact, nil
		}
		var ciArtifactIds []int
		for _, cdWfA := range cdWorkflowArtifact {
			ciArtifactIds = append(ciArtifactIds, cdWfA.CiArtifactId)
		}
		parentCiArtifact := make(map[int]int)
		isLinked := false
		ciArtifacts, err := impl.ciArtifactRepository.GetArtifactParentCiAndWorkflowDetailsByIds(ciArtifactIds)
		if err != nil || len(ciArtifacts) == 0 {
			impl.Logger.Errorw("error fetching artifact data", "err", err)
			return cdWorkflowArtifact, err
		}
		var newCiArtifactIds []int
		for _, ciArtifact := range ciArtifacts {
			if ciArtifact.ParentCiArtifact > 0 && ciArtifact.WorkflowId == nil {
				// parent ci artifact ID can be greater than zero when pipeline is linked or when image is copied at plugin level from some other image
				isLinked = true
				newCiArtifactIds = append(newCiArtifactIds, ciArtifact.ParentCiArtifact)
				parentCiArtifact[ciArtifact.Id] = ciArtifact.ParentCiArtifact
			} else {
				newCiArtifactIds = append(newCiArtifactIds, ciArtifact.Id)
			}
		}
		// handling linked ci pipeline
		if isLinked {
			ciArtifactIds = newCiArtifactIds
		}

		ciWfs, err := impl.ciWorkflowRepository.FindAllLastTriggeredWorkflowByArtifactId(ciArtifactIds)
		if err != nil && err != pg.ErrNoRows {
			impl.Logger.Errorw("error in fetching ci wfs", "artifactIds", ciArtifactIds, "err", err)
			return cdWorkflowArtifact, err
		} else if len(ciWfs) == 0 {
			return cdWorkflowArtifact, nil
		}

		wfGitTriggers := make(map[int]map[int]pipelineConfig.GitCommit)
		var ciPipelineId int
		for _, ciWf := range ciWfs {
			ciPipelineId = ciWf.CiPipelineId
			wfGitTriggers[ciWf.Id] = ciWf.GitTriggers
		}
		ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineIdForRegexAndFixed(ciPipelineId)
		if err != nil && err != pg.ErrNoRows {
			impl.Logger.Errorw("err in fetching ci materials", "ciMaterials", ciMaterials, "err", err)
			return cdWorkflowArtifact, err
		}

		var ciMaterialsArr []buildBean.CiPipelineMaterialResponse
		for _, ciMaterial := range ciMaterials {
			res := buildBean.CiPipelineMaterialResponse{
				Id:              ciMaterial.Id,
				GitMaterialId:   ciMaterial.GitMaterialId,
				GitMaterialName: ciMaterial.GitMaterial.Name[strings.Index(ciMaterial.GitMaterial.Name, "-")+1:],
				Type:            string(ciMaterial.Type),
				Value:           ciMaterial.Value,
				Active:          ciMaterial.Active,
				Url:             ciMaterial.GitMaterial.Url,
			}
			ciMaterialsArr = append(ciMaterialsArr, res)
		}
		var newCdWorkflowArtifact []pipelineBean.CdWorkflowWithArtifact
		for _, cdWfA := range cdWorkflowArtifact {

			gitTriggers := make(map[int]pipelineConfig.GitCommit)
			if isLinked {
				if gitTriggerVal, ok := wfGitTriggers[parentCiArtifact[cdWfA.CiArtifactId]]; ok {
					gitTriggers = gitTriggerVal
				}
			} else {
				if gitTriggerVal, ok := wfGitTriggers[cdWfA.CiArtifactId]; ok {
					gitTriggers = gitTriggerVal
				}
			}

			cdWfA.GitTriggers = gitTriggers
			cdWfA.CiMaterials = ciMaterialsArr
			newCdWorkflowArtifact = append(newCdWorkflowArtifact, cdWfA)

		}
		cdWorkflowArtifact = newCdWorkflowArtifact
	}

	var artifactIds []int
	for _, item := range cdWorkflowArtifact {
		artifactIds = append(artifactIds, item.CiArtifactId)
	}
	imageCommentsDataMap, err := impl.imageTaggingService.GetImageCommentsDataMapByArtifactIds(artifactIds)
	if err != nil {
		impl.Logger.Errorw("error in fetching imageCommentsDataMap", "err", err, "artifactIds", artifactIds, "appId", appId)
		return cdWorkflowArtifact, err
	}
	for i, item := range cdWorkflowArtifact {

		if imageTagsDataMap[item.CiArtifactId] != nil {
			item.ImageReleaseTags = imageTagsDataMap[item.CiArtifactId]
		}
		if imageCommentsDataMap[item.CiArtifactId] != nil {
			item.ImageComment = imageCommentsDataMap[item.CiArtifactId]
		}
		cdWorkflowArtifact[i] = item
	}

	//process pre/post cd stage data
	//prepare a map of wfId and wf type to pass to next function
	wfIdToWfTypeMap := make(map[int]pipelineBean.CdWorkflowWithArtifact)
	for _, item := range cdWorkflowArtifact {
		if item.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE.String() || item.WorkflowType == bean.CD_WORKFLOW_TYPE_POST.String() {
			wfIdToWfTypeMap[item.Id] = item
		}
	}
	wfRunnerIdToStageDetailMap, err := impl.cdWorkflowRunnerService.GetPrePostWorkflowStagesByWorkflowRunnerIdsList(wfIdToWfTypeMap)
	if err != nil {
		impl.Logger.Errorw("error in fetching pre/post stage data", "err", err)
		return cdWorkflowArtifact, err
	}

	//now for each cdWorkflowArtifact, set the workflowStage Data from wfRunnerIdToStageDetailMap using workflowId as key wfRunnerIdToStageDetailMap
	if len(wfRunnerIdToStageDetailMap) > 0 {
		for i, item := range cdWorkflowArtifact {
			if val, ok := wfRunnerIdToStageDetailMap[item.Id]; ok {
				cdWorkflowArtifact[i].WorkflowExecutionStage = val
			}
		}
	}
	return cdWorkflowArtifact, nil
}

func (impl *CdHandlerImpl) FetchCdWorkflowDetails(appId int, environmentId int, pipelineId int, buildId int) (types.WorkflowResponse, error) {
	workflowR, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(buildId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("err", "err", err)
		return types.WorkflowResponse{}, err
	} else if err == pg.ErrNoRows {
		return types.WorkflowResponse{}, nil
	}
	workflow := impl.converterWFR(*workflowR)

	if workflowR.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE || workflowR.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
		//get execution stage data
		impl.Logger.Infow("fetching pre/post workflow stages", "workflowId", workflowR.Id, "workflowType", workflowR.WorkflowType)
		workflowStageData, err := impl.workflowStageStatusService.GetWorkflowStagesByWorkflowIdAndType(workflowR.Id, workflowR.WorkflowType.String())
		if err != nil {
			impl.Logger.Errorw("error in fetching pre/post workflow stages", "err", err)
			return types.WorkflowResponse{}, err
		}
		workflow.WorkflowExecutionStage = impl.workflowStageStatusService.ConvertDBWorkflowStageToMap(workflowStageData, workflow.Id, workflow.Status, workflow.PodStatus, workflow.Message, workflow.WorkflowType, workflow.StartedOn, workflow.FinishedOn)
	} else {
		workflow.WorkflowExecutionStage = map[string][]*bean5.WorkflowStageDto{}
	}

	triggeredByUserEmailId, err := impl.userService.GetActiveEmailById(workflow.TriggeredBy)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", "err", err)
		return types.WorkflowResponse{}, err
	}
	if len(triggeredByUserEmailId) == 0 {
		triggeredByUserEmailId = "anonymous"
	}
	ciArtifactId := workflow.CiArtifactId
	targetPlatforms := []*bean4.TargetPlatform{}
	if ciArtifactId > 0 {
		ciArtifact, err := impl.ciArtifactRepository.Get(ciArtifactId)
		if err != nil {
			impl.Logger.Errorw("error fetching artifact data", "err", err)
			return types.WorkflowResponse{}, err
		}

		targetPlatforms = utils.ConvertTargetPlatformStringToObject(ciArtifact.TargetPlatforms)

		// handling linked ci pipeline
		if ciArtifact.ParentCiArtifact > 0 && ciArtifact.WorkflowId == nil {
			ciArtifactId = ciArtifact.ParentCiArtifact
		}
	}
	ciWf, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflowByArtifactId(ciArtifactId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in fetching ci wf", "artifactId", workflow.CiArtifactId, "err", err)
		return types.WorkflowResponse{}, err
	}
	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineIdForRegexAndFixed(ciWf.CiPipelineId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return types.WorkflowResponse{}, err
	}

	var ciMaterialsArr []buildBean.CiPipelineMaterialResponse
	for _, m := range ciMaterials {
		res := buildBean.CiPipelineMaterialResponse{
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
	gitTriggers := make(map[int]pipelineConfig.GitCommit)
	if ciWf.GitTriggers != nil {
		gitTriggers = ciWf.GitTriggers
	}

	workflowResponse := types.WorkflowResponse{
		Id:                     workflow.Id,
		Name:                   workflow.Name,
		Status:                 workflow.Status,
		PodStatus:              workflow.PodStatus,
		Message:                workflow.Message,
		StartedOn:              workflow.StartedOn,
		FinishedOn:             workflow.FinishedOn,
		Namespace:              workflow.Namespace,
		CiMaterials:            ciMaterialsArr,
		TriggeredBy:            workflow.TriggeredBy,
		TriggeredByEmail:       triggeredByUserEmailId,
		Artifact:               workflow.Image,
		Stage:                  workflow.WorkflowType,
		GitTriggers:            gitTriggers,
		BlobStorageEnabled:     workflow.BlobStorageEnabled,
		IsVirtualEnvironment:   workflowR.CdWorkflow.Pipeline.Environment.IsVirtualEnvironment,
		PodName:                workflowR.PodName,
		ArtifactId:             workflow.CiArtifactId,
		IsArtifactUploaded:     workflow.IsArtifactUploaded,
		CiPipelineId:           ciWf.CiPipelineId,
		TargetPlatforms:        targetPlatforms,
		WorkflowExecutionStage: workflow.WorkflowExecutionStage,
	}
	return workflowResponse, nil

}

func (impl *CdHandlerImpl) converterWFR(wfr pipelineConfig.CdWorkflowRunner) pipelineBean.CdWorkflowWithArtifact {
	workflow := pipelineBean.CdWorkflowWithArtifact{}
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
		workflow.TargetPlatforms = utils.ConvertTargetPlatformStringToObject(wfr.CdWorkflow.CiArtifact.TargetPlatforms)
		workflow.PipelineId = wfr.CdWorkflow.PipelineId
		workflow.CiArtifactId = wfr.CdWorkflow.CiArtifactId
		// TODO: FIXME :- if wfr status is terminal then only migrate isArtifactUploaded flag.
		isArtifactUploaded, isMigrationRequired := wfr.GetIsArtifactUploaded()
		if isMigrationRequired {
			// Migrate isArtifactUploaded. For old records, set isArtifactUploaded -> Uploaded
			impl.cdWorkflowRepository.MigrateIsArtifactUploaded(wfr.Id, true)
			isArtifactUploaded = true
		}
		workflow.IsArtifactUploaded = isArtifactUploaded
		workflow.BlobStorageEnabled = wfr.BlobStorageEnabled
		workflow.RefCdWorkflowRunnerId = wfr.RefCdWorkflowRunnerId
	}
	return workflow
}

func (impl *CdHandlerImpl) converterWFRList(wfrList []pipelineConfig.CdWorkflowRunner) []pipelineBean.CdWorkflowWithArtifact {
	var workflowList []pipelineBean.CdWorkflowWithArtifact
	var results []pipelineBean.CdWorkflowWithArtifact
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

func (impl *CdHandlerImpl) FetchCdPrePostStageStatus(pipelineId int) ([]pipelineBean.CdWorkflowWithArtifact, error) {
	var results []pipelineBean.CdWorkflowWithArtifact
	wfrPre, err := impl.cdWorkflowRepository.FindLatestByPipelineIdAndRunnerType(pipelineId, bean.CD_WORKFLOW_TYPE_PRE)
	if err != nil && err != pg.ErrNoRows {
		return results, err
	}
	if wfrPre.Id > 0 {
		workflowPre := impl.converterWFR(wfrPre)
		results = append(results, workflowPre)
	} else {
		workflowPre := pipelineBean.CdWorkflowWithArtifact{Status: "Notbuilt", WorkflowType: string(bean.CD_WORKFLOW_TYPE_PRE), PipelineId: pipelineId}
		results = append(results, workflowPre)
	}

	wfrPost, err := impl.cdWorkflowRepository.FindLatestByPipelineIdAndRunnerType(pipelineId, bean.CD_WORKFLOW_TYPE_POST)
	if err != nil && err != pg.ErrNoRows {
		return results, err
	}
	if wfrPost.Id > 0 {
		workflowPost := impl.converterWFR(wfrPost)
		results = append(results, workflowPost)
	} else {
		workflowPost := pipelineBean.CdWorkflowWithArtifact{Status: "Notbuilt", WorkflowType: string(bean.CD_WORKFLOW_TYPE_POST), PipelineId: pipelineId}
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
	partialDeletedPipelines := make(map[int]bool)
	// pipelineIdsMap := make(map[int]int)
	for _, pipeline := range pipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
		partialDeletedPipelines[pipeline.Id] = pipeline.DeploymentAppDeleteRequest
	}

	if len(pipelineIds) == 0 {
		return cdWorkflowStatus, nil
	}

	result, err := impl.getWfrStatusForLatestRunners(pipelineIds, pipelines)
	if err != nil {
		impl.Logger.Errorw("error in fetching wfrIds", "pipelineIds", pipelineIds, "err", err)
		return cdWorkflowStatus, err
	}

	var wfrIds []int
	for _, item := range result {
		wfrIds = append(wfrIds, item.WfrId)
	}

	var cdMap = make(map[int]*pipelineConfig.CdWorkflowStatus)
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
			if item.WorkflowType == pipelineBean.WorkflowTypePre {
				cdWorkflowStatus.PreStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == pipelineBean.WorkflowTypeDeploy {
				cdWorkflowStatus.DeployStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == pipelineBean.WorkflowTypePost {
				cdWorkflowStatus.PostStatus = statusMap[item.WfrId]
			}
			cdMap[item.PipelineId] = cdWorkflowStatus
		} else {
			cdWorkflowStatus := cdMap[item.PipelineId]
			cdWorkflowStatus.PipelineId = item.PipelineId
			cdWorkflowStatus.CiPipelineId = item.CiPipelineId
			if item.WorkflowType == pipelineBean.WorkflowTypePre {
				cdWorkflowStatus.PreStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == pipelineBean.WorkflowTypeDeploy {
				cdWorkflowStatus.DeployStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == pipelineBean.WorkflowTypePost {
				cdWorkflowStatus.PostStatus = statusMap[item.WfrId]
			}
			cdMap[item.PipelineId] = cdWorkflowStatus
		}
		cdMap[item.PipelineId].DeploymentAppDeleteRequest = partialDeletedPipelines[item.PipelineId]
	}

	for _, item := range cdMap {
		if item.PreStatus == "" {
			item.PreStatus = pipelineBean.NotTriggered
		}
		if item.DeployStatus == "" {
			item.DeployStatus = pipelineBean.NotDeployed
		}
		if item.PostStatus == "" {
			item.PostStatus = pipelineBean.NotTriggered
		}
		cdWorkflowStatus = append(cdWorkflowStatus, item)
	}

	if len(cdWorkflowStatus) == 0 {
		for _, item := range pipelineIds {
			cdWs := &pipelineConfig.CdWorkflowStatus{}
			cdWs.PipelineId = item
			cdWs.PreStatus = pipelineBean.NotTriggered
			cdWs.DeployStatus = pipelineBean.NotDeployed
			cdWs.PostStatus = pipelineBean.NotTriggered
			cdWorkflowStatus = append(cdWorkflowStatus, cdWs)
		}
	} else {
		for _, item := range pipelineIds {
			if _, ok := cdMap[item]; !ok {
				cdWs := &pipelineConfig.CdWorkflowStatus{}
				cdWs.PipelineId = item
				cdWs.PreStatus = pipelineBean.NotTriggered
				cdWs.DeployStatus = pipelineBean.NotDeployed
				cdWs.PostStatus = pipelineBean.NotTriggered
				cdWorkflowStatus = append(cdWorkflowStatus, cdWs)
			}
		}
	}

	return cdWorkflowStatus, err
}

func (impl *CdHandlerImpl) getWfrStatusForLatestRunners(pipelineIds []int, pipelines []*pipelineConfig.Pipeline) ([]*pipelineConfig.CdWorkflowStatus, error) {
	// fetching the latest pipeline from the index table - cdWorkflowLatest
	var result []*pipelineConfig.CdWorkflowStatus
	cdWorkflowLatest, err := impl.WorkflowStatusLatestService.GetCdWorkflowLatestByPipelineIds(pipelineIds)
	if err != nil {
		impl.Logger.Errorw("error in getting latest by pipelineId", "pipelineId", pipelineIds, "err", err)
		return nil, err
	}

	var pipelineIdToCiPipelineIdMap map[int]int
	for _, item := range pipelines {
		pipelineIdToCiPipelineIdMap[item.Id] = item.CiPipelineId
	}

	for _, item := range cdWorkflowLatest {
		result = append(result, &pipelineConfig.CdWorkflowStatus{
			CiPipelineId: pipelineIdToCiPipelineIdMap[item.PipelineId],
			PipelineId:   item.PipelineId,
			WorkflowType: item.WorkflowType,
			WfrId:        item.WorkflowRunnerId,
		})
	}

	var cdWorfklowLatestMap map[int]map[bean.WorkflowType]bool
	for _, item := range cdWorkflowLatest {
		if _, ok := cdWorfklowLatestMap[item.PipelineId]; !ok {
			cdWorfklowLatestMap[item.PipelineId] = make(map[bean.WorkflowType]bool)
		}
		cdWorfklowLatestMap[item.PipelineId][bean.WorkflowType(item.WorkflowType)] = true
	}

	pipelineStage, err := impl.pipelineStageRepository.GetAllCdStagesByCdPipelineIds(pipelineIds)
	if err != nil {
		impl.Logger.Errorw("error in fetching pipeline stages", "pipelineId", pipelineIds, "err", err)
		return nil, err
	}
	pipelineStageMap := make(map[int]map[bean.WorkflowType]bool)
	for _, item := range pipelineStage {
		if _, ok := pipelineStageMap[item.CdPipelineId]; !ok {
			pipelineStageMap[item.CdPipelineId] = make(map[bean.WorkflowType]bool)
		}
		if item.Type == repository2.PIPELINE_STAGE_TYPE_PRE_CD {
			pipelineStageMap[item.CdPipelineId][bean.CD_WORKFLOW_TYPE_PRE] = true
		} else if item.Type == repository2.PIPELINE_STAGE_TYPE_POST_CD {
			pipelineStageMap[item.CdPipelineId][bean.CD_WORKFLOW_TYPE_POST] = true
		}
	}

	// calculating all the pipelines not present in the index table cdWorkflowLatest
	var pipelinesAbsentInCache map[int]bean.WorkflowType
	for _, item := range pipelines {
		if _, ok := cdWorfklowLatestMap[item.Id]; !ok {
			pipelinesAbsentInCache[item.Id] = bean.CD_WORKFLOW_TYPE_PRE
			pipelinesAbsentInCache[item.Id] = bean.CD_WORKFLOW_TYPE_DEPLOY
			pipelinesAbsentInCache[item.Id] = bean.CD_WORKFLOW_TYPE_POST
		} else {
			if _, ok := pipelineStageMap[item.Id][bean.CD_WORKFLOW_TYPE_PRE]; ok {
				if val, ok := cdWorfklowLatestMap[item.Id][bean.CD_WORKFLOW_TYPE_PRE]; !ok || !val {
					pipelinesAbsentInCache[item.Id] = bean.CD_WORKFLOW_TYPE_PRE
				}
			}
			if _, ok := pipelineStageMap[item.Id][bean.CD_WORKFLOW_TYPE_POST]; ok {
				if val, ok := cdWorfklowLatestMap[item.Id][bean.CD_WORKFLOW_TYPE_POST]; !ok || !val {
					pipelinesAbsentInCache[item.Id] = bean.CD_WORKFLOW_TYPE_POST
				}
			}
			if val, ok := cdWorfklowLatestMap[item.Id][bean.CD_WORKFLOW_TYPE_DEPLOY]; !ok || !val {
				pipelinesAbsentInCache[item.Id] = bean.CD_WORKFLOW_TYPE_POST
			}
		}
	}
	if len(pipelinesAbsentInCache) > 0 {
		remainingRunners, err := impl.cdWorkflowRepository.FetchAllCdStagesLatestEntity(pipelinesAbsentInCache)
		if err != nil {
			impl.Logger.Errorw("error in fetching all cd stages latest entity", "pipelinesAbsentInCache", pipelinesAbsentInCache, "err", err)
			return nil, err
		}
		result = append(result, remainingRunners...)
	}
	return result, nil
}

func (impl *CdHandlerImpl) FetchAppWorkflowStatusForTriggerViewForEnvironment(request resourceGroup2.ResourceGroupingRequest, token string) ([]*pipelineConfig.CdWorkflowStatus, error) {
	cdWorkflowStatus := make([]*pipelineConfig.CdWorkflowStatus, 0)
	var pipelines []*pipelineConfig.Pipeline
	var err error
	if request.ResourceGroupId > 0 {
		appIds, err := impl.resourceGroupService.GetResourceIdsByResourceGroupId(request.ResourceGroupId)
		if err != nil {
			return nil, err
		}
		// override appIds if already provided app group id in request.
		request.ResourceIds = appIds
	}
	if len(request.ResourceIds) > 0 {
		pipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.ParentResourceId, request.ResourceIds)
	} else {
		pipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.ParentResourceId)
	}
	if err != nil {
		impl.Logger.Errorw("error in fetching pipelines", "request", request, "err", err)
		return nil, err
	}

	var appIds []int
	for _, pipeline := range pipelines {
		appIds = append(appIds, pipeline.AppId)
	}
	if len(appIds) == 0 {
		impl.Logger.Warnw("there is no app id found for fetching cd pipelines", "request", request)
		return cdWorkflowStatus, nil
	}
	pipelines, err = impl.pipelineRepository.FindActiveByAppIds(appIds)
	if err != nil && err != pg.ErrNoRows {
		return cdWorkflowStatus, err
	}
	pipelineIds := make([]int, 0)
	for _, pipeline := range pipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
	}
	if len(pipelineIds) == 0 {
		return cdWorkflowStatus, nil
	}
	// authorization block starts here
	var appObjectArr []string
	var envObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByPipelineIds(pipelineIds)
	pipelineIds = []int{}
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
		envObjectArr = append(envObjectArr, object[1])
	}

	// filter out pipelines for unauthorized apps but not envs
	appResults, _ := request.CheckAuthBatch(token, appObjectArr, envObjectArr)
	for _, pipeline := range pipelines {
		// Safety check to prevent index-out-of-range panic
		objectArr, ok := objects[pipeline.Id]
		if !ok {
			impl.Logger.Warnw("skipping pipeline with missing object data", "pipelineId", pipeline.Id)
			continue
		}
		appObject := objectArr[0]
		if !(appResults[appObject]) {
			// if user unauthorized, skip items
			continue
		}
		pipelineIds = append(pipelineIds, pipeline.Id)
	}
	// authorization block ends here
	if len(pipelineIds) == 0 {
		return cdWorkflowStatus, nil
	}

	cdMap := make(map[int]*pipelineConfig.CdWorkflowStatus)

	wfrStatus, err := impl.getWfrStatusForLatestRunners(pipelineIds, pipelines)
	if err != nil {
		impl.Logger.Errorw("error in fetching wfrIds", "pipelineIds", pipelineIds, "err", err)
		return cdWorkflowStatus, err
	}

	var wfrIds []int
	for _, item := range wfrStatus {
		wfrIds = append(wfrIds, item.WfrId)
	}

	statusMap := make(map[int]string)
	if len(wfrIds) > 0 {
		cdWorkflowRunners, err := impl.cdWorkflowRepository.FetchAllCdStagesLatestEntityStatus(wfrIds)
		if err != nil && !util.IsErrNoRows(err) {
			return cdWorkflowStatus, err
		}
		for _, item := range cdWorkflowRunners {
			statusMap[item.Id] = item.Status
		}
	}

	for _, item := range wfrStatus {
		if _, ok := cdMap[item.PipelineId]; !ok {
			cdWorkflowStatus := &pipelineConfig.CdWorkflowStatus{}
			cdWorkflowStatus.PipelineId = item.PipelineId
			cdWorkflowStatus.CiPipelineId = item.CiPipelineId
			if item.WorkflowType == pipelineBean.WorkflowTypePre {
				cdWorkflowStatus.PreStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == pipelineBean.WorkflowTypeDeploy {
				cdWorkflowStatus.DeployStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == pipelineBean.WorkflowTypePost {
				cdWorkflowStatus.PostStatus = statusMap[item.WfrId]
			}
			cdMap[item.PipelineId] = cdWorkflowStatus
		} else {
			cdWorkflowStatus := cdMap[item.PipelineId]
			cdWorkflowStatus.PipelineId = item.PipelineId
			cdWorkflowStatus.CiPipelineId = item.CiPipelineId
			if item.WorkflowType == pipelineBean.WorkflowTypePre {
				cdWorkflowStatus.PreStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == pipelineBean.WorkflowTypeDeploy {
				cdWorkflowStatus.DeployStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == pipelineBean.WorkflowTypePost {
				cdWorkflowStatus.PostStatus = statusMap[item.WfrId]
			}
			cdMap[item.PipelineId] = cdWorkflowStatus
		}
	}

	for _, item := range cdMap {
		if item.PreStatus == "" {
			item.PreStatus = pipelineBean.NotTriggered
		}
		if item.DeployStatus == "" {
			item.DeployStatus = pipelineBean.NotDeployed
		}
		if item.PostStatus == "" {
			item.PostStatus = pipelineBean.NotTriggered
		}
		cdWorkflowStatus = append(cdWorkflowStatus, item)
	}

	if len(cdWorkflowStatus) == 0 {
		for _, item := range pipelineIds {
			cdWs := &pipelineConfig.CdWorkflowStatus{}
			cdWs.PipelineId = item
			cdWs.PreStatus = pipelineBean.NotTriggered
			cdWs.DeployStatus = pipelineBean.NotDeployed
			cdWs.PostStatus = pipelineBean.NotTriggered
			cdWorkflowStatus = append(cdWorkflowStatus, cdWs)
		}
	} else {
		for _, item := range pipelineIds {
			if _, ok := cdMap[item]; !ok {
				cdWs := &pipelineConfig.CdWorkflowStatus{}
				cdWs.PipelineId = item
				cdWs.PreStatus = pipelineBean.NotTriggered
				cdWs.DeployStatus = pipelineBean.NotDeployed
				cdWs.PostStatus = pipelineBean.NotTriggered
				cdWorkflowStatus = append(cdWorkflowStatus, cdWs)
			}
		}
	}

	return cdWorkflowStatus, err
}

func (impl *CdHandlerImpl) FetchAppDeploymentStatusForEnvironments(request resourceGroup2.ResourceGroupingRequest, token string) ([]*pipelineConfig.AppDeploymentStatus, error) {
	_, span := otel.Tracer("orchestrator").Start(request.Ctx, "pipelineBuilder.authorizationDeploymentStatusForResourceGrouping")
	deploymentStatuses := make([]*pipelineConfig.AppDeploymentStatus, 0)
	deploymentStatusesMap := make(map[int]*pipelineConfig.AppDeploymentStatus)
	pipelineAppMap := make(map[int]int)
	statusMap := make(map[int]string)
	var cdPipelines []*pipelineConfig.Pipeline
	var err error
	if request.ResourceGroupId > 0 {
		appIds, err := impl.resourceGroupService.GetResourceIdsByResourceGroupId(request.ResourceGroupId)
		if err != nil {
			return nil, err
		}
		// override appIds if already provided app group id in request.
		request.ResourceIds = appIds
	}
	if len(request.ResourceIds) > 0 {
		cdPipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.ParentResourceId, request.ResourceIds)
	} else {
		cdPipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.ParentResourceId)
	}
	if err != nil {
		impl.Logger.Errorw("error in fetching pipelines", "request", request, "err", err)
		return nil, err
	}
	pipelineIds := make([]int, 0)
	for _, pipeline := range cdPipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
	}
	if len(pipelineIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching pipeline found"}
		return nil, err
	}
	// authorization block starts here
	var appObjectArr []string
	var envObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByPipelineIds(pipelineIds)
	pipelineIds = []int{}
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
		envObjectArr = append(envObjectArr, object[1])
	}
	appResults, envResults := request.CheckAuthBatch(token, appObjectArr, envObjectArr)
	for _, pipeline := range cdPipelines {
		if _, ok := objects[pipeline.Id]; !ok {
			impl.Logger.Warnw("skipping pipeline as no object found for it", "pipelineId", pipeline.Id)
			continue
		}
		appObject := objects[pipeline.Id][0]
		envObject := objects[pipeline.Id][1]
		if !(appResults[appObject] && envResults[envObject]) {
			// if user unauthorized, skip items
			continue
		}
		pipelineIds = append(pipelineIds, pipeline.Id)
		pipelineAppMap[pipeline.Id] = pipeline.AppId
	}
	span.End()
	// authorization block ends here

	if len(pipelineIds) == 0 {
		return deploymentStatuses, nil
	}
	_, span = otel.Tracer("orchestrator").Start(request.Ctx, "pipelineBuilder.FetchAllCdStagesLatestEntity")
	result, err := impl.getWfrStatusForLatestRunners(pipelineIds, cdPipelines)
	span.End()
	if err != nil {
		return deploymentStatuses, err
	}
	var wfrIds []int
	for _, item := range result {
		wfrIds = append(wfrIds, item.WfrId)
	}
	if len(wfrIds) > 0 {
		_, span = otel.Tracer("orchestrator").Start(request.Ctx, "pipelineBuilder.FetchAllCdStagesLatestEntityStatus")
		wfrList, err := impl.cdWorkflowRepository.FetchEnvAllCdStagesLatestEntityStatus(wfrIds, request.ParentResourceId)
		span.End()
		if err != nil && !util.IsErrNoRows(err) {
			return deploymentStatuses, err
		}
		for _, item := range wfrList {
			if item.Status == "" {
				statusMap[item.Id] = pipelineBean.NotDeployed
			} else {
				statusMap[item.Id] = item.Status
			}
		}
	}

	for _, item := range result {
		if _, ok := deploymentStatusesMap[item.PipelineId]; !ok {
			deploymentStatus := &pipelineConfig.AppDeploymentStatus{}
			deploymentStatus.PipelineId = item.PipelineId
			if item.WorkflowType == pipelineBean.WorkflowTypeDeploy {
				deploymentStatus.DeployStatus = statusMap[item.WfrId]
				deploymentStatus.AppId = pipelineAppMap[deploymentStatus.PipelineId]
				deploymentStatusesMap[item.PipelineId] = deploymentStatus
			}
		}
	}
	// in case there is no workflow found for pipeline, set all the pipeline status - Not Deployed
	for _, pipelineId := range pipelineIds {
		if _, ok := deploymentStatusesMap[pipelineId]; !ok {
			deploymentStatus := &pipelineConfig.AppDeploymentStatus{}
			deploymentStatus.PipelineId = pipelineId
			deploymentStatus.DeployStatus = pipelineBean.NotDeployed
			deploymentStatus.AppId = pipelineAppMap[deploymentStatus.PipelineId]
			deploymentStatusesMap[pipelineId] = deploymentStatus
		}
	}
	for _, deploymentStatus := range deploymentStatusesMap {
		deploymentStatuses = append(deploymentStatuses, deploymentStatus)
	}

	return deploymentStatuses, err
}

func (impl *CdHandlerImpl) DeactivateImageReservationPathsOnFailure(imagePathReservationIds []int) error {
	return impl.customTagService.DeactivateImagePathReservationByImageIds(imagePathReservationIds)
}
