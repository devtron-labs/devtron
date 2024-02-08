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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/common-lib-private/utils/k8s"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	k8s2 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	resourceGroup2 "github.com/devtron-labs/devtron/pkg/resourceGroup"
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

const (
	DEVTRON_APP_HELM_PIPELINE_STATUS_UPDATE_CRON = "DTAppHelmPipelineStatusUpdateCron"
	DEVTRON_APP_ARGO_PIPELINE_STATUS_UPDATE_CRON = "DTAppArgoPipelineStatusUpdateCron"
	HELM_APP_ARGO_PIPELINE_STATUS_UPDATE_CRON    = "HelmAppArgoPipelineStatusUpdateCron"
)

type CdHandler interface {
	UpdateWorkflow(workflowStatus v1alpha1.WorkflowStatus) (int, string, error)
	GetCdBuildHistory(appId int, environmentId int, pipelineId int, offset int, size int) ([]bean2.CdWorkflowWithArtifact, error)
	GetRunningWorkflowLogs(environmentId int, pipelineId int, workflowId int) (*bufio.Reader, func() error, error)
	FetchCdWorkflowDetails(appId int, environmentId int, pipelineId int, buildId int, showAppliedFilters bool) (types.WorkflowResponse, error)
	DownloadCdWorkflowArtifacts(pipelineId int, buildId int) (*os.File, error)
	FetchCdPrePostStageStatus(pipelineId int) ([]bean2.CdWorkflowWithArtifact, error)
	CancelStage(workflowRunnerId int, userId int32) (int, error)
	FetchAppWorkflowStatusForTriggerView(appId int) ([]*pipelineConfig.CdWorkflowStatus, error)
	FetchAppWorkflowStatusForTriggerViewForEnvironment(request resourceGroup2.ResourceGroupingRequest, token string) ([]*pipelineConfig.CdWorkflowStatus, error)
	FetchAppDeploymentStatusForEnvironments(request resourceGroup2.ResourceGroupingRequest, token string) ([]*pipelineConfig.AppDeploymentStatus, error)
	DeactivateImageReservationPathsOnFailure(imagePathReservationIds []int) error
}

type CdHandlerImpl struct {
	Logger                       *zap.SugaredLogger
	userService                  user.UserService
	ciLogService                 CiLogService
	ciArtifactRepository         repository.CiArtifactRepository
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
	cdWorkflowRepository         pipelineConfig.CdWorkflowRepository
	envRepository                repository2.EnvironmentRepository
	pipelineRepository           pipelineConfig.PipelineRepository
	ciWorkflowRepository         pipelineConfig.CiWorkflowRepository
	enforcerUtil                 rbac.EnforcerUtil
	resourceGroupService         resourceGroup2.ResourceGroupService
	imageTaggingService          ImageTaggingService
	k8sUtil                      *k8s.K8sUtilExtended
	workflowService              WorkflowService
	config                       *types.CdConfig
	clusterService               cluster.ClusterService
	blobConfigStorageService     BlobStorageConfigService
	customTagService             CustomTagService
	deploymentApprovalRepository pipelineConfig.DeploymentApprovalRepository
	resourceFilterService        resourceFilter.ResourceFilterService
}

func NewCdHandlerImpl(Logger *zap.SugaredLogger, userService user.UserService,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository, ciLogService CiLogService,
	ciArtifactRepository repository.CiArtifactRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	pipelineRepository pipelineConfig.PipelineRepository, envRepository repository2.EnvironmentRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository, enforcerUtil rbac.EnforcerUtil,
	resourceGroupService resourceGroup2.ResourceGroupService,
	imageTaggingService ImageTaggingService, k8sUtil *k8s.K8sUtilExtended,
	workflowService WorkflowService, clusterService cluster.ClusterService,
	blobConfigStorageService BlobStorageConfigService, customTagService CustomTagService,
	deploymentApprovalRepository pipelineConfig.DeploymentApprovalRepository,
	resourceFilterService resourceFilter.ResourceFilterService) *CdHandlerImpl {
	cdh := &CdHandlerImpl{
		Logger:                       Logger,
		userService:                  userService,
		ciLogService:                 ciLogService,
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
		workflowService:              workflowService,
		clusterService:               clusterService,
		blobConfigStorageService:     blobConfigStorageService,
		customTagService:             customTagService,
		deploymentApprovalRepository: deploymentApprovalRepository,
		resourceFilterService:        resourceFilterService,
	}
	config, err := types.GetCdConfig()
	if err != nil {
		return nil
	}
	cdh.config = config
	return cdh
}

const NotTriggered string = "Not Triggered"
const NotDeployed = "Not Deployed"
const WorklowTypeDeploy = "DEPLOY"
const WorklowTypePre = "PRE"
const WorklowTypePost = "POST"
const WorkflowApprovalNode = "APPROVAL"

func (impl *CdHandlerImpl) CancelStage(workflowRunnerId int, userId int32) (int, error) {
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

	var clusterBean cluster.ClusterBean
	if env != nil && env.Cluster != nil {
		clusterBean = cluster.GetClusterBean(*env.Cluster)
	}
	clusterConfig := clusterBean.GetClusterConfig()
	var isExtCluster bool
	if workflowRunner.WorkflowType == types.PRE {
		isExtCluster = pipeline.RunPreStageInEnv
	} else if workflowRunner.WorkflowType == types.POST {
		isExtCluster = pipeline.RunPostStageInEnv
	}
	var restConfig *rest.Config
	if isExtCluster {
		restConfig, err = impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
		if err != nil {
			impl.Logger.Errorw("error in getting rest config by cluster id", "err", err)
			return 0, err
		}
	}
	// Terminate workflow
	err = impl.workflowService.TerminateWorkflow(workflowRunner.ExecutorType, workflowRunner.Name, workflowRunner.Namespace, restConfig, isExtCluster, nil)
	if err != nil {
		impl.Logger.Error("cannot terminate wf runner", "err", err)
		return 0, err
	}
	if len(workflowRunner.ImagePathReservationIds) > 0 {
		err := impl.customTagService.DeactivateImagePathReservationByImageIds(workflowRunner.ImagePathReservationIds)
		if err != nil {
			impl.Logger.Errorw("error in deactivating image path reservation ids", "err", err)
			return 0, err
		}
	}
	workflowRunner.Status = executors.WorkflowCancel
	workflowRunner.UpdatedOn = time.Now()
	workflowRunner.UpdatedBy = userId
	err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(workflowRunner)
	if err != nil {
		impl.Logger.Error("cannot update deleted workflow runner status, but wf deleted", "err", err)
		return 0, err
	}
	return workflowRunner.Id, nil
}

func (impl *CdHandlerImpl) UpdateWorkflow(workflowStatus v1alpha1.WorkflowStatus) (int, string, error) {
	wfStatusRs := impl.extractWorkfowStatus(workflowStatus)
	workflowName, status, podStatus, message, podName := wfStatusRs.WorkflowName, wfStatusRs.Status, wfStatusRs.PodStatus, wfStatusRs.Message, wfStatusRs.PodName
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
		ciArtifactLocationFormat = impl.config.GetArtifactLocationFormat()
	}

	if impl.stateChanged(status, podStatus, message, workflowStatus.FinishedAt.Time, savedWorkflow) {
		if savedWorkflow.Status != executors.WorkflowCancel {
			savedWorkflow.Status = status
		}
		savedWorkflow.PodStatus = podStatus
		savedWorkflow.Message = message
		savedWorkflow.FinishedOn = workflowStatus.FinishedAt.Time
		savedWorkflow.Name = workflowName
		// removed log location from here since we are saving it at trigger
		savedWorkflow.PodName = podName
		savedWorkflow.UpdatedOn = time.Now()
		savedWorkflow.UpdatedBy = 1
		impl.Logger.Debugw("updating workflow ", "workflow", savedWorkflow)
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(savedWorkflow)
		if err != nil {
			impl.Logger.Error("update wf failed for id " + strconv.Itoa(savedWorkflow.Id))
			return 0, "", err
		}
		cdMetrics := util3.CDMetrics{
			AppName:         savedWorkflow.CdWorkflow.Pipeline.DeploymentAppName,
			Status:          savedWorkflow.Status,
			DeploymentType:  savedWorkflow.CdWorkflow.Pipeline.DeploymentAppType,
			EnvironmentName: savedWorkflow.CdWorkflow.Pipeline.Environment.Name,
			Time:            time.Since(savedWorkflow.StartedOn).Seconds() - time.Since(savedWorkflow.FinishedOn).Seconds(),
		}
		util3.TriggerCDMetrics(cdMetrics, impl.config.ExposeCDMetrics)
		if string(v1alpha1.NodeError) == savedWorkflow.Status || string(v1alpha1.NodeFailed) == savedWorkflow.Status {
			impl.Logger.Warnw("cd stage failed for workflow: ", "wfId", savedWorkflow.Id)
		}
	}
	return savedWorkflow.Id, savedWorkflow.Status, nil
}

func (impl *CdHandlerImpl) extractWorkfowStatus(workflowStatus v1alpha1.WorkflowStatus) *types.WorkflowStatus {
	workflowName := ""
	status := string(workflowStatus.Phase)
	podStatus := "Pending"
	message := ""
	logLocation := ""
	podName := ""
	for k, v := range workflowStatus.Nodes {
		impl.Logger.Debugw("ExtractWorkflowStatus", "workflowName", k, "v", v)
		if v.TemplateName == bean2.CD_WORKFLOW_NAME {
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

func (impl *CdHandlerImpl) fillAppliedFiltersData(cdWorkflowArtifacts []bean2.CdWorkflowWithArtifact) []bean2.CdWorkflowWithArtifact {
	artifactIds := make([]int, len(cdWorkflowArtifacts))
	workflowRunnerIds := make([]int, len(cdWorkflowArtifacts))
	for i, cdWorkflowArtifact := range cdWorkflowArtifacts {
		artifactIds[i] = cdWorkflowArtifact.CiArtifactId
		workflowRunnerIds[i] = cdWorkflowArtifact.Id
	}
	appliedFiltersMap, appliedFiltersTimeStampMap, err := impl.resourceFilterService.GetEvaluatedFiltersForSubjectsAndReferenceIds(resourceFilter.Artifact, artifactIds, workflowRunnerIds, resourceFilter.CdWorkflowRunner)
	if err != nil {
		// not returning error by choice
		impl.Logger.Errorw("error in fetching applied filters when this image was born", "cdWorkflowRunnerIds", workflowRunnerIds, "artifactIds", artifactIds, "err", err)
		return cdWorkflowArtifacts
	}
	for i, cdWorkflowArtifact := range cdWorkflowArtifacts {
		artifactWfrKey := fmt.Sprintf("%v-%v", cdWorkflowArtifact.CiArtifactId, cdWorkflowArtifact.Id)
		cdWorkflowArtifacts[i].AppliedFilters = appliedFiltersMap[artifactWfrKey]
		cdWorkflowArtifacts[i].AppliedFiltersTimestamp = appliedFiltersTimeStampMap[artifactWfrKey]
		// we are setting this data in workflow runner list, which means these got triggered because filters are allowed or no filters configured at all
		cdWorkflowArtifact.AppliedFiltersState = resourceFilter.ALLOW
	}
	return cdWorkflowArtifacts
}

func (impl *CdHandlerImpl) GetCdBuildHistory(appId int, environmentId int, pipelineId int, offset int, size int) ([]bean2.CdWorkflowWithArtifact, error) {

	var cdWorkflowArtifact []bean2.CdWorkflowWithArtifact
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
		ciArtifacts, err := impl.ciArtifactRepository.GetArtifactParentCiAndWorkflowDetailsByIdsInDesc(ciArtifactIds)
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

		var ciMaterialsArr []pipelineConfig.CiPipelineMaterialResponse
		for _, ciMaterial := range ciMaterials {
			res := pipelineConfig.CiPipelineMaterialResponse{
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
		var newCdWorkflowArtifact []bean2.CdWorkflowWithArtifact
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
	cdWorkflowArtifact = impl.fillAppliedFiltersData(cdWorkflowArtifact)
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
	var clusterBean cluster.ClusterBean
	if env != nil && env.Cluster != nil {
		clusterBean = cluster.GetClusterBean(*env.Cluster)
	}
	clusterConfig := clusterBean.GetClusterConfig()
	var isExtCluster bool
	if cdWorkflow.WorkflowType == types.PRE {
		isExtCluster = pipeline.RunPreStageInEnv
	} else if cdWorkflow.WorkflowType == types.POST {
		isExtCluster = pipeline.RunPostStageInEnv
	}
	return impl.getWorkflowLogs(pipelineId, cdWorkflow, clusterConfig, isExtCluster)
}

func (impl *CdHandlerImpl) getWorkflowLogs(pipelineId int, cdWorkflow *pipelineConfig.CdWorkflowRunner, clusterConfig *k8s2.ClusterConfig, runStageInEnv bool) (*bufio.Reader, func() error, error) {
	cdLogRequest := types.BuildLogRequest{
		PodName:   cdWorkflow.PodName,
		Namespace: cdWorkflow.Namespace,
	}

	logStream, cleanUp, err := impl.ciLogService.FetchRunningWorkflowLogs(cdLogRequest, clusterConfig, runStageInEnv)
	if logStream == nil || err != nil {
		if !cdWorkflow.BlobStorageEnabled {
			return nil, nil, errors.New("logs-not-stored-in-repository")
		} else if string(v1alpha1.NodeSucceeded) == cdWorkflow.Status || string(v1alpha1.NodeError) == cdWorkflow.Status || string(v1alpha1.NodeFailed) == cdWorkflow.Status || cdWorkflow.Status == executors.WorkflowCancel {
			impl.Logger.Debugw("pod is not live ", "err", err)
			return impl.getLogsFromRepository(pipelineId, cdWorkflow, clusterConfig, runStageInEnv)
		}
		impl.Logger.Errorw("err on fetch workflow logs", "err", err)
		return nil, nil, err
	}
	logReader := bufio.NewReader(logStream)
	return logReader, cleanUp, err
}

func (impl *CdHandlerImpl) getLogsFromRepository(pipelineId int, cdWorkflow *pipelineConfig.CdWorkflowRunner, clusterConfig *k8s2.ClusterConfig, isExt bool) (*bufio.Reader, func() error, error) {
	impl.Logger.Debug("getting historic logs", "pipelineId", pipelineId)

	cdConfig, err := impl.cdWorkflowRepository.FindConfigByPipelineId(pipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", err)
		return nil, nil, err
	}

	if cdConfig.LogsBucket == "" {
		cdConfig.LogsBucket = impl.config.GetDefaultBuildLogsBucket() // TODO -fixme
	}
	if cdConfig.CdCacheRegion == "" {
		cdConfig.CdCacheRegion = impl.config.GetDefaultCdLogsBucketRegion()
	}

	cdLogRequest := types.BuildLogRequest{
		PipelineId:    cdWorkflow.CdWorkflow.PipelineId,
		WorkflowId:    cdWorkflow.Id,
		PodName:       cdWorkflow.PodName,
		LogsFilePath:  cdWorkflow.LogLocation, // impl.ciCdConfig.CiDefaultBuildLogsKeyPrefix + "/" + cdWorkflow.Name + "/main.log", //TODO - fixme
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
			BucketName:        cdConfig.LogsBucket,
			Region:            cdConfig.CdCacheRegion,
			VersioningEnabled: impl.config.BlobStorageS3BucketVersioned,
		},
		GcpBlobBaseConfig: &blob_storage.GcpBlobBaseConfig{
			BucketName:             cdConfig.LogsBucket,
			CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
		},
	}
	useExternalBlobStorage := isExternalBlobStorageEnabled(isExt, impl.config.UseBlobStorageConfigInCdWorkflow)
	if useExternalBlobStorage {
		// fetch extClusterBlob cm and cs from k8s client, if they are present then read creds
		// from them else return.
		cmConfig, secretConfig, err := impl.blobConfigStorageService.FetchCmAndSecretBlobConfigFromExternalCluster(clusterConfig, cdWorkflow.Namespace)
		if err != nil {
			impl.Logger.Errorw("error in fetching config map and secret from external cluster", "err", err, "clusterConfig", clusterConfig)
			return nil, nil, err
		}
		rq := &cdLogRequest
		rq.SetBuildLogRequest(cmConfig, secretConfig)
	}

	impl.Logger.Infow("s3 log req ", "req", cdLogRequest)
	oldLogsStream, cleanUp, err := impl.ciLogService.FetchLogs(impl.config.BaseLogLocationPath, cdLogRequest)
	if err != nil {
		impl.Logger.Errorw("err", err)
		return nil, nil, err
	}
	logReader := bufio.NewReader(oldLogsStream)
	return logReader, cleanUp, err
}
func isExternalBlobStorageEnabled(isExternalRun bool, useBlobStorageConfigInCdWorkflow bool) bool {
	// TODO impl.config.UseBlobStorageConfigInCdWorkflow fetches the live status, we need to check from db as well, we should put useExternalBlobStorage in db
	return isExternalRun && !useBlobStorageConfigInCdWorkflow
}

func (impl *CdHandlerImpl) FetchCdWorkflowDetails(appId int, environmentId int, pipelineId int, buildId int, showAppliedFilters bool) (types.WorkflowResponse, error) {
	workflowR, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(buildId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("err", "err", err)
		return types.WorkflowResponse{}, err
	} else if err == pg.ErrNoRows {
		return types.WorkflowResponse{}, nil
	}

	var userIds []int32
	var approvalRequestedUserId int32
	approvalRequest := workflowR.DeploymentApprovalRequest
	if approvalRequest != nil {
		approvalReqId := workflowR.DeploymentApprovalRequestId
		approvalUserData, err := impl.deploymentApprovalRepository.FetchApprovalDataForRequests([]int{approvalReqId})
		if err != nil {
			return types.WorkflowResponse{}, err
		}
		approvalRequest.DeploymentApprovalUserData = approvalUserData
		approvalRequestedUserId = approvalRequest.CreatedBy
		userIds = append(userIds, approvalRequestedUserId)
	}

	triggeredBy := workflowR.TriggeredBy

	triggeredByUserEmailId := "anonymous"

	userIds = append(userIds, triggeredBy)
	userInfos, err := impl.userService.GetByIds(userIds)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", "err", err)
		return types.WorkflowResponse{}, err
	}
	for _, userInfo := range userInfos {
		if userInfo.Id == triggeredBy {
			triggeredByUserEmailId = userInfo.EmailId
		}
		if userInfo.Id == approvalRequestedUserId {
			approvalRequest.UserEmail = userInfo.EmailId
		}
	}

	workflow := impl.converterWFR(*workflowR)

	ciArtifactId := workflow.CiArtifactId
	if ciArtifactId > 0 {
		ciArtifact, err := impl.ciArtifactRepository.Get(ciArtifactId)
		if err != nil {
			impl.Logger.Errorw("error fetching artifact data", "err", err)
			return types.WorkflowResponse{}, err
		}

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

	var ciMaterialsArr []pipelineConfig.CiPipelineMaterialResponse
	for _, m := range ciMaterials {
		res := pipelineConfig.CiPipelineMaterialResponse{
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

	var imageTag string
	if len(workflow.Image) > 0 {
		imageTag = strings.Split(workflow.Image, ":")[1]
	}
	appName := workflowR.CdWorkflow.Pipeline.App.AppName
	if workflowR.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
		appName = fmt.Sprintf("%s-%s", bean.CD_WORKFLOW_TYPE_PRE, appName)
	} else if workflowR.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
		appName = fmt.Sprintf("%s-%s", bean.CD_WORKFLOW_TYPE_POST, appName)
	}
	helmPackageName := fmt.Sprintf("%s-%s-%s",
		appName,
		workflowR.CdWorkflow.Pipeline.Environment.Name,
		imageTag)

	workflowResponse := types.WorkflowResponse{
		Id:                   workflow.Id,
		Name:                 workflow.Name,
		Status:               workflow.Status,
		PodStatus:            workflow.PodStatus,
		Message:              workflow.Message,
		StartedOn:            workflow.StartedOn,
		FinishedOn:           workflow.FinishedOn,
		Namespace:            workflow.Namespace,
		CiMaterials:          ciMaterialsArr,
		TriggeredBy:          workflow.TriggeredBy,
		TriggeredByEmail:     triggeredByUserEmailId,
		Artifact:             workflow.Image,
		Stage:                workflow.WorkflowType,
		GitTriggers:          gitTriggers,
		BlobStorageEnabled:   workflow.BlobStorageEnabled,
		UserApprovalMetadata: workflow.UserApprovalMetadata,
		IsVirtualEnvironment: workflowR.CdWorkflow.Pipeline.Environment.IsVirtualEnvironment,
		PodName:              workflowR.PodName,
		CdWorkflowId:         workflowR.CdWorkflowId,
		HelmPackageName:      helmPackageName,
		ArtifactId:           workflow.CiArtifactId,
		CiPipelineId:         ciWf.CiPipelineId,
	}

	if showAppliedFilters {

		appliedFiltersMap, appliedFiltersTimeStampMap, err := impl.resourceFilterService.GetEvaluatedFiltersForSubjects(resourceFilter.Artifact, []int{workflow.CiArtifactId}, workflow.Id, resourceFilter.CdWorkflowRunner)
		if err != nil {
			// not returning error by choice
			impl.Logger.Errorw("error in fetching applied filters when this image was born", "cdWorkflowRunnerId", workflow.Id, "err", err)
		}
		workflowResponse.AppliedFiltersState = resourceFilter.ALLOW
		workflowResponse.AppliedFilters = appliedFiltersMap[workflow.CiArtifactId]
		workflowResponse.AppliedFiltersTimestamp = appliedFiltersTimeStampMap[workflow.CiArtifactId]
	}
	return workflowResponse, nil

}

func (impl *CdHandlerImpl) DownloadCdWorkflowArtifacts(pipelineId int, buildId int) (*os.File, error) {
	wfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(buildId)
	if err != nil {
		impl.Logger.Errorw("unable to fetch ciWorkflow", "err", err)
		return nil, err
	}
	useExternalBlobStorage := isExternalBlobStorageEnabled(wfr.IsExternalRun(), impl.config.UseBlobStorageConfigInCdWorkflow)
	if !wfr.BlobStorageEnabled {
		return nil, errors.New("logs-not-stored-in-repository")
	}

	cdConfig, err := impl.cdWorkflowRepository.FindConfigByPipelineId(pipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("unable to fetch ciCdConfig", "err", err)
		return nil, err
	}

	if cdConfig.LogsBucket == "" {
		cdConfig.LogsBucket = impl.config.GetDefaultBuildLogsBucket()
	}
	if cdConfig.CdCacheRegion == "" {
		cdConfig.CdCacheRegion = impl.config.GetDefaultCdLogsBucketRegion()
	}

	item := strconv.Itoa(wfr.Id)
	awsS3BaseConfig := &blob_storage.AwsS3BaseConfig{
		AccessKey:         impl.config.BlobStorageS3AccessKey,
		Passkey:           impl.config.BlobStorageS3SecretKey,
		EndpointUrl:       impl.config.BlobStorageS3Endpoint,
		IsInSecure:        impl.config.BlobStorageS3EndpointInsecure,
		BucketName:        cdConfig.LogsBucket,
		Region:            cdConfig.CdCacheRegion,
		VersioningEnabled: impl.config.BlobStorageS3BucketVersioned,
	}
	azureBlobBaseConfig := &blob_storage.AzureBlobBaseConfig{
		Enabled:           impl.config.CloudProvider == types.BLOB_STORAGE_AZURE,
		AccountKey:        impl.config.AzureAccountKey,
		AccountName:       impl.config.AzureAccountName,
		BlobContainerName: impl.config.AzureBlobContainerCiLog,
	}
	gcpBlobBaseConfig := &blob_storage.GcpBlobBaseConfig{
		BucketName:             cdConfig.LogsBucket,
		CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
	}
	key := fmt.Sprintf("%s/"+impl.config.GetArtifactLocationFormat(), impl.config.GetDefaultArtifactKeyPrefix(), wfr.CdWorkflow.Id, wfr.Id)
	baseLogLocationPathConfig := impl.config.BaseLogLocationPath
	blobStorageService := blob_storage.NewBlobStorageServiceImpl(nil)
	destinationKey := filepath.Clean(filepath.Join(baseLogLocationPathConfig, item))
	request := &blob_storage.BlobStorageRequest{
		StorageType:         impl.config.CloudProvider,
		SourceKey:           key,
		DestinationKey:      destinationKey,
		AzureBlobBaseConfig: azureBlobBaseConfig,
		AwsS3BaseConfig:     awsS3BaseConfig,
		GcpBlobBaseConfig:   gcpBlobBaseConfig,
	}
	if useExternalBlobStorage {
		clusterConfig, err := impl.clusterService.GetClusterConfigByClusterId(wfr.CdWorkflow.Pipeline.Environment.ClusterId)
		if err != nil {
			impl.Logger.Errorw("GetClusterConfigByClusterId, error in fetching clusterConfig", "err", err, "clusterId", wfr.CdWorkflow.Pipeline.Environment.ClusterId)
			return nil, err
		}
		// fetch extClusterBlob cm and cs from k8s client, if they are present then read creds
		// from them else return.
		cmConfig, secretConfig, err := impl.blobConfigStorageService.FetchCmAndSecretBlobConfigFromExternalCluster(clusterConfig, wfr.Namespace)
		if err != nil {
			impl.Logger.Errorw("error in fetching config map and secret from external cluster", "err", err, "clusterConfig", clusterConfig)
			return nil, err
		}
		request = updateRequestWithExtClusterCmAndSecret(request, cmConfig, secretConfig)
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

	impl.Logger.Infow("Downloaded ", "name", file.Name(), "bytes", numBytes)
	return file, nil
}

func (impl *CdHandlerImpl) converterWFR(wfr pipelineConfig.CdWorkflowRunner) bean2.CdWorkflowWithArtifact {
	workflow := bean2.CdWorkflowWithArtifact{}
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
		workflow.BlobStorageEnabled = wfr.BlobStorageEnabled
		if wfr.DeploymentApprovalRequest != nil {
			workflow.UserApprovalMetadata = wfr.DeploymentApprovalRequest.ConvertToApprovalMetadata()
		}
		workflow.RefCdWorkflowRunnerId = wfr.RefCdWorkflowRunnerId
	}
	return workflow
}

func (impl *CdHandlerImpl) converterWFRList(wfrList []pipelineConfig.CdWorkflowRunner) []bean2.CdWorkflowWithArtifact {
	var workflowList []bean2.CdWorkflowWithArtifact
	var results []bean2.CdWorkflowWithArtifact
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

func (impl *CdHandlerImpl) FetchCdPrePostStageStatus(pipelineId int) ([]bean2.CdWorkflowWithArtifact, error) {
	var results []bean2.CdWorkflowWithArtifact
	wfrPre, err := impl.cdWorkflowRepository.FindLastStatusByPipelineIdAndRunnerType(pipelineId, bean.CD_WORKFLOW_TYPE_PRE)
	if err != nil && err != pg.ErrNoRows {
		return results, err
	}
	if wfrPre.Id > 0 {
		workflowPre := impl.converterWFR(wfrPre)
		results = append(results, workflowPre)
	} else {
		workflowPre := bean2.CdWorkflowWithArtifact{Status: "Notbuilt", WorkflowType: string(bean.CD_WORKFLOW_TYPE_PRE), PipelineId: pipelineId}
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
		workflowPost := bean2.CdWorkflowWithArtifact{Status: "Notbuilt", WorkflowType: string(bean.CD_WORKFLOW_TYPE_POST), PipelineId: pipelineId}
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
			if item.WorkflowType == WorklowTypePre {
				cdWorkflowStatus.PreStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypeDeploy {
				cdWorkflowStatus.DeployStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypePost {
				cdWorkflowStatus.PostStatus = statusMap[item.WfrId]
			}
			cdMap[item.PipelineId] = cdWorkflowStatus
		} else {
			cdWorkflowStatus := cdMap[item.PipelineId]
			cdWorkflowStatus.PipelineId = item.PipelineId
			cdWorkflowStatus.CiPipelineId = item.CiPipelineId
			if item.WorkflowType == WorklowTypePre {
				cdWorkflowStatus.PreStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypeDeploy {
				cdWorkflowStatus.DeployStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypePost {
				cdWorkflowStatus.PostStatus = statusMap[item.WfrId]
			}
			cdMap[item.PipelineId] = cdWorkflowStatus
		}
		cdMap[item.PipelineId].DeploymentAppDeleteRequest = partialDeletedPipelines[item.PipelineId]
	}

	for _, item := range cdMap {
		if item.PreStatus == "" {
			item.PreStatus = NotTriggered
		}
		if item.DeployStatus == "" {
			item.DeployStatus = NotDeployed
		}
		if item.PostStatus == "" {
			item.PostStatus = NotTriggered
		}
		cdWorkflowStatus = append(cdWorkflowStatus, item)
	}

	if len(cdWorkflowStatus) == 0 {
		for _, item := range pipelineIds {
			cdWs := &pipelineConfig.CdWorkflowStatus{}
			cdWs.PipelineId = item
			cdWs.PreStatus = NotTriggered
			cdWs.DeployStatus = NotDeployed
			cdWs.PostStatus = NotTriggered
			cdWorkflowStatus = append(cdWorkflowStatus, cdWs)
		}
	} else {
		for _, item := range pipelineIds {
			if _, ok := cdMap[item]; !ok {
				cdWs := &pipelineConfig.CdWorkflowStatus{}
				cdWs.PipelineId = item
				cdWs.PreStatus = NotTriggered
				cdWs.DeployStatus = NotDeployed
				cdWs.PostStatus = NotTriggered
				cdWorkflowStatus = append(cdWorkflowStatus, cdWs)
			}
		}
	}

	return cdWorkflowStatus, err
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
	appResults, envResults := request.CheckAuthBatch(token, appObjectArr, envObjectArr)
	for _, pipeline := range pipelines {
		appObject := objects[pipeline.Id][0]
		envObject := objects[pipeline.Id][1]
		if !(appResults[appObject] && envResults[envObject]) {
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
	wfrStatus, err := impl.cdWorkflowRepository.FetchAllCdStagesLatestEntity(pipelineIds)
	if err != nil {
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
			if item.WorkflowType == WorklowTypePre {
				cdWorkflowStatus.PreStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypeDeploy {
				cdWorkflowStatus.DeployStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypePost {
				cdWorkflowStatus.PostStatus = statusMap[item.WfrId]
			}
			cdMap[item.PipelineId] = cdWorkflowStatus
		} else {
			cdWorkflowStatus := cdMap[item.PipelineId]
			cdWorkflowStatus.PipelineId = item.PipelineId
			cdWorkflowStatus.CiPipelineId = item.CiPipelineId
			if item.WorkflowType == WorklowTypePre {
				cdWorkflowStatus.PreStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypeDeploy {
				cdWorkflowStatus.DeployStatus = statusMap[item.WfrId]
			} else if item.WorkflowType == WorklowTypePost {
				cdWorkflowStatus.PostStatus = statusMap[item.WfrId]
			}
			cdMap[item.PipelineId] = cdWorkflowStatus
		}
	}

	for _, item := range cdMap {
		if item.PreStatus == "" {
			item.PreStatus = NotTriggered
		}
		if item.DeployStatus == "" {
			item.DeployStatus = NotDeployed
		}
		if item.PostStatus == "" {
			item.PostStatus = NotTriggered
		}
		cdWorkflowStatus = append(cdWorkflowStatus, item)
	}

	if len(cdWorkflowStatus) == 0 {
		for _, item := range pipelineIds {
			cdWs := &pipelineConfig.CdWorkflowStatus{}
			cdWs.PipelineId = item
			cdWs.PreStatus = NotTriggered
			cdWs.DeployStatus = NotDeployed
			cdWs.PostStatus = NotTriggered
			cdWorkflowStatus = append(cdWorkflowStatus, cdWs)
		}
	} else {
		for _, item := range pipelineIds {
			if _, ok := cdMap[item]; !ok {
				cdWs := &pipelineConfig.CdWorkflowStatus{}
				cdWs.PipelineId = item
				cdWs.PreStatus = NotTriggered
				cdWs.DeployStatus = NotDeployed
				cdWs.PostStatus = NotTriggered
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
	result, err := impl.cdWorkflowRepository.FetchAllCdStagesLatestEntity(pipelineIds)
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
		wfrList, err := impl.cdWorkflowRepository.FetchAllCdStagesLatestEntityStatus(wfrIds)
		span.End()
		if err != nil && !util.IsErrNoRows(err) {
			return deploymentStatuses, err
		}
		for _, item := range wfrList {
			if item.Status == "" {
				statusMap[item.Id] = NotDeployed
			} else {
				statusMap[item.Id] = item.Status
			}
		}
	}

	for _, item := range result {
		if _, ok := deploymentStatusesMap[item.PipelineId]; !ok {
			deploymentStatus := &pipelineConfig.AppDeploymentStatus{}
			deploymentStatus.PipelineId = item.PipelineId
			if item.WorkflowType == WorklowTypeDeploy {
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
			deploymentStatus.DeployStatus = NotDeployed
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
