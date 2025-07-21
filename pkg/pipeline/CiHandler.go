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
	"context"
	"errors"
	"fmt"
	"github.com/devtron-labs/common-lib/utils"
	"github.com/devtron-labs/common-lib/utils/workFlow"
	cdWorkflowBean "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	"github.com/devtron-labs/devtron/pkg/build/artifacts/imageTagging"
	buildBean "github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	eventProcessorBean "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/constants"
	"github.com/devtron-labs/devtron/pkg/pipeline/workflowStatus"
	"github.com/devtron-labs/devtron/pkg/workflow/status/workflowStatusLatest"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	k8sPkg "github.com/devtron-labs/devtron/pkg/k8s"
	pipelineConfigBean "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/resourceGroup"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CiHandler interface {
	//	HandleCIWebhook(gitCiTriggerRequest bean.GitCiTriggerRequest) (int, error)
	//HandleCIManual(ciTriggerRequest bean.CiTriggerRequest) (int, error)
	//CheckAndReTriggerCI(workflowStatus eventProcessorBean.CiCdStatus) error
	FetchMaterialsByPipelineId(pipelineId int, showAll bool) ([]buildBean.CiPipelineMaterialResponse, error)
	FetchMaterialsByPipelineIdAndGitMaterialId(pipelineId int, gitMaterialId int, showAll bool) ([]buildBean.CiPipelineMaterialResponse, error)
	FetchWorkflowDetails(appId int, pipelineId int, buildId int) (types.WorkflowResponse, error)
	FetchArtifactsForCiJob(buildId int) (*types.ArtifactsForCiJob, error)

	GetBuildHistory(pipelineId int, appId int, offset int, size int) ([]types.WorkflowResponse, error)
	UpdateWorkflow(workflowStatus eventProcessorBean.CiCdStatus) (int, bool, error)

	FetchCiStatusForTriggerView(appId int) ([]*pipelineConfig.CiWorkflowStatus, error)
	FetchCiStatusForTriggerViewV1(appId int) ([]*pipelineConfig.CiWorkflowStatus, error)
	RefreshMaterialByCiPipelineMaterialId(gitMaterialId int) (refreshRes *gitSensor.RefreshGitMaterialResponse, err error)
	FetchMaterialInfoByArtifactId(ciArtifactId int, envId int) (*types.GitTriggerInfoResponse, error)
	//UpdateCiWorkflowStatusFailure(timeoutForFailureCiBuild int) error
	FetchCiStatusForTriggerViewForEnvironment(request resourceGroup.ResourceGroupingRequest, token string) ([]*pipelineConfig.CiWorkflowStatus, error)
	CiHandlerEnt
}

type CiHandlerImpl struct {
	Logger                       *zap.SugaredLogger
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
	ciService                    CiService
	gitSensorClient              gitSensor.Client
	ciWorkflowRepository         pipelineConfig.CiWorkflowRepository
	ciArtifactRepository         repository.CiArtifactRepository
	userService                  user.UserService
	eventClient                  client.EventClient
	eventFactory                 client.EventFactory
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
	appListingRepository         repository.AppListingRepository
	cdPipelineRepository         pipelineConfig.PipelineRepository
	enforcerUtil                 rbac.EnforcerUtil
	resourceGroupService         resourceGroup.ResourceGroupService
	envRepository                repository2.EnvironmentRepository
	imageTaggingService          imageTagging.ImageTaggingService
	customTagService             CustomTagService
	appWorkflowRepository        appWorkflow.AppWorkflowRepository
	config                       *types.CiConfig
	k8sCommonService             k8sPkg.K8sCommonService
	workFlowStageStatusService   workflowStatus.WorkFlowStageStatusService
	workflowStatusUpdateService  workflowStatusLatest.WorkflowStatusUpdateService
}

func NewCiHandlerImpl(Logger *zap.SugaredLogger, ciService CiService, ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository, gitSensorClient gitSensor.Client, ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	ciArtifactRepository repository.CiArtifactRepository, userService user.UserService, eventClient client.EventClient, eventFactory client.EventFactory, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	appListingRepository repository.AppListingRepository, cdPipelineRepository pipelineConfig.PipelineRepository, enforcerUtil rbac.EnforcerUtil, resourceGroupService resourceGroup.ResourceGroupService, envRepository repository2.EnvironmentRepository,
	imageTaggingService imageTagging.ImageTaggingService, k8sCommonService k8sPkg.K8sCommonService, appWorkflowRepository appWorkflow.AppWorkflowRepository, customTagService CustomTagService,
	workFlowStageStatusService workflowStatus.WorkFlowStageStatusService,
	workflowStatusUpdateService workflowStatusLatest.WorkflowStatusUpdateService,
) *CiHandlerImpl {
	cih := &CiHandlerImpl{
		Logger:                       Logger,
		ciService:                    ciService,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		gitSensorClient:              gitSensorClient,
		ciWorkflowRepository:         ciWorkflowRepository,
		ciArtifactRepository:         ciArtifactRepository,
		userService:                  userService,
		eventClient:                  eventClient,
		eventFactory:                 eventFactory,
		ciPipelineRepository:         ciPipelineRepository,
		appListingRepository:         appListingRepository,
		cdPipelineRepository:         cdPipelineRepository,
		enforcerUtil:                 enforcerUtil,
		resourceGroupService:         resourceGroupService,
		envRepository:                envRepository,
		imageTaggingService:          imageTaggingService,
		customTagService:             customTagService,
		appWorkflowRepository:        appWorkflowRepository,
		k8sCommonService:             k8sCommonService,
		workFlowStageStatusService:   workFlowStageStatusService,
		workflowStatusUpdateService:  workflowStatusUpdateService,
	}
	config, err := types.GetCiConfig()
	if err != nil {
		return nil
	}
	cih.config = config

	return cih
}

func (impl *CiHandlerImpl) RefreshMaterialByCiPipelineMaterialId(gitMaterialId int) (refreshRes *gitSensor.RefreshGitMaterialResponse, err error) {
	impl.Logger.Debugw("refreshing git material", "id", gitMaterialId)
	refreshRes, err = impl.gitSensorClient.RefreshGitMaterial(context.Background(),
		&gitSensor.RefreshGitMaterialRequest{GitMaterialId: gitMaterialId},
	)
	return refreshRes, err
}

func (impl *CiHandlerImpl) FetchMaterialsByPipelineIdAndGitMaterialId(pipelineId int, gitMaterialId int, showAll bool) ([]buildBean.CiPipelineMaterialResponse, error) {
	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineIdAndGitMaterialId(pipelineId, gitMaterialId)
	if err != nil {
		impl.Logger.Errorw("ciMaterials fetch failed", "err", err)
	}
	var ciPipelineMaterialResponses []buildBean.CiPipelineMaterialResponse
	var responseMap = make(map[int]bool)

	ciMaterialHistoryMap := make(map[*pipelineConfig.CiPipelineMaterial]*gitSensor.MaterialChangeResp)
	for _, m := range ciMaterials {
		// git material should be active in this case
		if m == nil || m.GitMaterial == nil || !m.GitMaterial.Active {
			continue
		}
		changesRequest := &gitSensor.FetchScmChangesRequest{
			PipelineMaterialId: m.Id,
			ShowAll:            showAll,
		}
		changesResp, apiErr := impl.gitSensorClient.FetchChanges(context.Background(), changesRequest)
		impl.Logger.Debugw("commits for material ", "m", m, "commits: ", changesResp)
		if apiErr != nil {
			impl.Logger.Warnw("git sensor FetchChanges failed for material", "id", m.Id)
			return []buildBean.CiPipelineMaterialResponse{}, apiErr
		}
		ciMaterialHistoryMap[m] = changesResp
	}

	for k, v := range ciMaterialHistoryMap {
		r := buildBean.CiPipelineMaterialResponse{
			Id:              k.Id,
			GitMaterialId:   k.GitMaterialId,
			GitMaterialName: k.GitMaterial.Name[strings.Index(k.GitMaterial.Name, "-")+1:],
			Type:            string(k.Type),
			Value:           k.Value,
			Active:          k.Active,
			GitMaterialUrl:  k.GitMaterial.Url,
			History:         v.Commits,
			LastFetchTime:   v.LastFetchTime,
			IsRepoError:     v.IsRepoError,
			RepoErrorMsg:    v.RepoErrorMsg,
			IsBranchError:   v.IsBranchError,
			BranchErrorMsg:  v.BranchErrorMsg,
			Regex:           k.Regex,
		}
		responseMap[k.GitMaterialId] = true
		ciPipelineMaterialResponses = append(ciPipelineMaterialResponses, r)
	}

	regexMaterials, err := impl.ciPipelineMaterialRepository.GetRegexByPipelineId(pipelineId)
	if err != nil {
		impl.Logger.Errorw("regex ciMaterials fetch failed", "err", err)
		return []buildBean.CiPipelineMaterialResponse{}, err
	}
	for _, k := range regexMaterials {
		r := buildBean.CiPipelineMaterialResponse{
			Id:              k.Id,
			GitMaterialId:   k.GitMaterialId,
			GitMaterialName: k.GitMaterial.Name[strings.Index(k.GitMaterial.Name, "-")+1:],
			Type:            string(k.Type),
			Value:           k.Value,
			Active:          k.Active,
			GitMaterialUrl:  k.GitMaterial.Url,
			History:         nil,
			IsRepoError:     false,
			RepoErrorMsg:    "",
			IsBranchError:   false,
			BranchErrorMsg:  "",
			Regex:           k.Regex,
		}
		_, exists := responseMap[k.GitMaterialId]
		if !exists {
			ciPipelineMaterialResponses = append(ciPipelineMaterialResponses, r)
		}
	}
	return ciPipelineMaterialResponses, nil
}

func (impl *CiHandlerImpl) FetchMaterialsByPipelineId(pipelineId int, showAll bool) ([]buildBean.CiPipelineMaterialResponse, error) {
	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(pipelineId)
	if err != nil {
		impl.Logger.Errorw("ciMaterials fetch failed", "err", err)
	}
	var ciPipelineMaterialResponses []buildBean.CiPipelineMaterialResponse
	var responseMap = make(map[int]bool)

	ciMaterialHistoryMap := make(map[*pipelineConfig.CiPipelineMaterial]*gitSensor.MaterialChangeResp)
	for _, m := range ciMaterials {
		// git material should be active in this case
		if m == nil || m.GitMaterial == nil || !m.GitMaterial.Active {
			continue
		}
		changesRequest := &gitSensor.FetchScmChangesRequest{
			PipelineMaterialId: m.Id,
			ShowAll:            showAll,
		}
		changesResp, apiErr := impl.gitSensorClient.FetchChanges(context.Background(), changesRequest)
		impl.Logger.Debugw("commits for material ", "m", m, "commits: ", changesResp)
		if apiErr != nil {
			impl.Logger.Warnw("git sensor FetchChanges failed for material", "id", m.Id)
			return nil, apiErr
		}
		ciMaterialHistoryMap[m] = changesResp
	}

	for k, v := range ciMaterialHistoryMap {
		r := buildBean.CiPipelineMaterialResponse{
			Id:              k.Id,
			GitMaterialId:   k.GitMaterialId,
			GitMaterialName: k.GitMaterial.Name[strings.Index(k.GitMaterial.Name, "-")+1:],
			Type:            string(k.Type),
			Value:           k.Value,
			Active:          k.Active,
			GitMaterialUrl:  k.GitMaterial.Url,
			History:         v.Commits,
			LastFetchTime:   v.LastFetchTime,
			IsRepoError:     v.IsRepoError,
			RepoErrorMsg:    v.RepoErrorMsg,
			IsBranchError:   v.IsBranchError,
			BranchErrorMsg:  v.BranchErrorMsg,
			Regex:           k.Regex,
		}
		responseMap[k.GitMaterialId] = true
		ciPipelineMaterialResponses = append(ciPipelineMaterialResponses, r)
	}

	regexMaterials, err := impl.ciPipelineMaterialRepository.GetRegexByPipelineId(pipelineId)
	if err != nil {
		impl.Logger.Errorw("regex ciMaterials fetch failed", "err", err)
		return nil, err
	}
	for _, k := range regexMaterials {
		r := buildBean.CiPipelineMaterialResponse{
			Id:              k.Id,
			GitMaterialId:   k.GitMaterialId,
			GitMaterialName: k.GitMaterial.Name[strings.Index(k.GitMaterial.Name, "-")+1:],
			Type:            string(k.Type),
			Value:           k.Value,
			Active:          k.Active,
			GitMaterialUrl:  k.GitMaterial.Url,
			History:         nil,
			IsRepoError:     false,
			RepoErrorMsg:    "",
			IsBranchError:   false,
			BranchErrorMsg:  "",
			Regex:           k.Regex,
		}
		_, exists := responseMap[k.GitMaterialId]
		if !exists {
			ciPipelineMaterialResponses = append(ciPipelineMaterialResponses, r)
		}
	}

	return ciPipelineMaterialResponses, nil
}

func (impl *CiHandlerImpl) GetBuildHistory(pipelineId int, appId int, offset int, size int) ([]types.WorkflowResponse, error) {
	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineIdForRegexAndFixed(pipelineId)
	if err != nil {
		impl.Logger.Errorw("ciMaterials fetch failed", "err", err)
	}
	var ciPipelineMaterialResponses []buildBean.CiPipelineMaterialResponse
	for _, m := range ciMaterials {
		r := buildBean.CiPipelineMaterialResponse{
			Id:              m.Id,
			GitMaterialId:   m.GitMaterialId,
			Type:            string(m.Type),
			Value:           m.Value,
			Active:          m.Active,
			GitMaterialName: m.GitMaterial.Name[strings.Index(m.GitMaterial.Name, "-")+1:],
			Url:             m.GitMaterial.Url,
		}
		ciPipelineMaterialResponses = append(ciPipelineMaterialResponses, r)
	}
	// this map contains artifactId -> array of tags of that artifact
	imageTagsDataMap, err := impl.imageTaggingService.GetTagsDataMapByAppId(appId)
	if err != nil {
		impl.Logger.Errorw("error in fetching image tags with appId", "err", err, "appId", appId)
		return nil, err
	}
	workFlows, err := impl.ciWorkflowRepository.FindByPipelineId(pipelineId, offset, size)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", "err", err)
		return nil, err
	}

	var workflowIds []int
	var artifactIds []int
	for _, w := range workFlows {
		artifactIds = append(artifactIds, w.CiArtifactId)
		workflowIds = append(workflowIds, w.Id)
	}

	allWfStagesDetail, err := impl.workFlowStageStatusService.GetWorkflowStagesByWorkflowIdsAndWfType(workflowIds, bean2.CI_WORKFLOW_TYPE.String())
	if err != nil {
		impl.Logger.Errorw("error in fetching allWfStagesDetail", "err", err, "workflowIds", workflowIds)
		return nil, err
	}

	// this map contains artifactId -> imageComment of that artifact
	imageCommetnsDataMap, err := impl.imageTaggingService.GetImageCommentsDataMapByArtifactIds(artifactIds)
	if err != nil {
		impl.Logger.Errorw("error in fetching imageCommetnsDataMap", "err", err, "appId", appId, "artifactIds", artifactIds)
		return nil, err
	}

	var ciWorkLowResponses []types.WorkflowResponse
	for _, w := range workFlows {
		isArtifactUploaded, isMigrationRequired := w.GetIsArtifactUploaded()
		if isMigrationRequired {
			// Migrate isArtifactUploaded. For old records, set isArtifactUploaded -> w.IsArtifactUploaded
			impl.ciWorkflowRepository.MigrateIsArtifactUploaded(w.Id, w.IsArtifactUploaded)
			isArtifactUploaded = w.IsArtifactUploaded
		}
		wfResponse := types.WorkflowResponse{
			Id:                     w.Id,
			Name:                   w.Name,
			Status:                 w.Status,
			PodStatus:              w.PodStatus,
			Message:                w.Message,
			StartedOn:              w.StartedOn,
			FinishedOn:             w.FinishedOn,
			CiPipelineId:           w.CiPipelineId,
			Namespace:              w.Namespace,
			LogLocation:            w.LogFilePath,
			GitTriggers:            w.GitTriggers,
			CiMaterials:            ciPipelineMaterialResponses,
			Artifact:               w.Image,
			TriggeredBy:            w.TriggeredBy,
			TriggeredByEmail:       w.EmailId,
			ArtifactId:             w.CiArtifactId,
			BlobStorageEnabled:     w.BlobStorageEnabled,
			IsArtifactUploaded:     isArtifactUploaded,
			EnvironmentId:          w.EnvironmentId,
			EnvironmentName:        w.EnvironmentName,
			ReferenceWorkflowId:    w.RefCiWorkflowId,
			PodName:                w.PodName,
			TargetPlatforms:        utils.ConvertTargetPlatformStringToObject(w.TargetPlatforms),
			WorkflowExecutionStage: impl.workFlowStageStatusService.ConvertDBWorkflowStageToMap(allWfStagesDetail, w.Id, w.Status, w.PodStatus, w.Message, bean2.CI_WORKFLOW_TYPE.String(), w.StartedOn, w.FinishedOn),
		}

		if w.Message == pipelineConfigBean.ImageTagUnavailableMessage {
			customTag, err := impl.customTagService.GetCustomTagByEntityKeyAndValue(pipelineConfigBean.EntityTypeCiPipelineId, strconv.Itoa(w.CiPipelineId))
			if err != nil && err != pg.ErrNoRows {
				// err == pg.ErrNoRows should never happen
				return nil, err
			}
			appWorkflows, err := impl.appWorkflowRepository.FindWFCIMappingByCIPipelineId(w.CiPipelineId)
			if err != nil && err != pg.ErrNoRows {
				return nil, err
			}
			wfResponse.AppWorkflowId = appWorkflows[0].AppWorkflowId // it is guaranteed there will always be 1 entry (in case of ci_pipeline_id)
			wfResponse.CustomTag = &bean2.CustomTagErrorResponse{
				TagPattern:           customTag.TagPattern,
				AutoIncreasingNumber: customTag.AutoIncreasingNumber,
				Message:              pipelineConfigBean.ImageTagUnavailableMessage,
			}
		}
		if imageTagsDataMap[w.CiArtifactId] != nil {
			wfResponse.ImageReleaseTags = imageTagsDataMap[w.CiArtifactId] // if artifact is not yet created,empty list will be sent
		}
		if imageCommetnsDataMap[w.CiArtifactId] != nil {
			wfResponse.ImageComment = imageCommetnsDataMap[w.CiArtifactId]
		}
		ciWorkLowResponses = append(ciWorkLowResponses, wfResponse)
	}
	return ciWorkLowResponses, nil
}

func (impl *CiHandlerImpl) FetchWorkflowDetails(appId int, pipelineId int, buildId int) (types.WorkflowResponse, error) {
	workflow, err := impl.ciWorkflowRepository.FindById(buildId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return types.WorkflowResponse{}, err
	}
	triggeredByUserEmailId, err := impl.userService.GetActiveEmailById(workflow.TriggeredBy)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", "err", err)
		return types.WorkflowResponse{}, err
	}

	if workflow.CiPipeline.AppId != appId {
		impl.Logger.Error("pipeline does not exist for this app")
		return types.WorkflowResponse{}, errors.New("invalid app and pipeline combination")
	}

	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(pipelineId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return types.WorkflowResponse{}, err
	}

	ciArtifact, err := impl.ciArtifactRepository.GetByWfId(workflow.Id)
	if err != nil && !util.IsErrNoRows(err) {
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
	environmentName := ""
	if workflow.EnvironmentId != 0 {
		envModel, err := impl.envRepository.FindById(workflow.EnvironmentId)
		if err != nil && err != pg.ErrNoRows {
			impl.Logger.Errorw("error in fetching environment details ", "err", err)
			return types.WorkflowResponse{}, err
		}
		environmentName = envModel.Name
	}
	isArtifactUploaded, isMigrationRequired := workflow.GetIsArtifactUploaded()
	if isMigrationRequired {
		// Migrate isArtifactUploaded. For old records, set isArtifactUploaded -> ciArtifact.IsArtifactUploaded
		impl.ciWorkflowRepository.MigrateIsArtifactUploaded(workflow.Id, ciArtifact.IsArtifactUploaded)
		isArtifactUploaded = ciArtifact.IsArtifactUploaded
	}

	wfStagesDetail, err := impl.workFlowStageStatusService.GetWorkflowStagesByWorkflowIdsAndWfType([]int{workflow.Id}, bean2.CI_WORKFLOW_TYPE.String())
	if err != nil {
		impl.Logger.Errorw("error in fetching allWfStagesDetail", "err", err, "workflowId", workflow.Id)
		return types.WorkflowResponse{}, err
	}

	workflowResponse := types.WorkflowResponse{
		Id:                     workflow.Id,
		Name:                   workflow.Name,
		Status:                 workflow.Status,
		PodStatus:              workflow.PodStatus,
		Message:                workflow.Message,
		StartedOn:              workflow.StartedOn,
		FinishedOn:             workflow.FinishedOn,
		CiPipelineId:           workflow.CiPipelineId,
		Namespace:              workflow.Namespace,
		LogLocation:            workflow.LogLocation,
		BlobStorageEnabled:     workflow.BlobStorageEnabled, // TODO default value if value not found in db
		GitTriggers:            workflow.GitTriggers,
		CiMaterials:            ciMaterialsArr,
		TriggeredBy:            workflow.TriggeredBy,
		TriggeredByEmail:       triggeredByUserEmailId,
		Artifact:               ciArtifact.Image,
		ArtifactId:             ciArtifact.Id,
		IsArtifactUploaded:     isArtifactUploaded,
		EnvironmentId:          workflow.EnvironmentId,
		EnvironmentName:        environmentName,
		PipelineType:           workflow.CiPipeline.PipelineType,
		PodName:                workflow.PodName,
		TargetPlatforms:        utils.ConvertTargetPlatformStringToObject(ciArtifact.TargetPlatforms),
		WorkflowExecutionStage: impl.workFlowStageStatusService.ConvertDBWorkflowStageToMap(wfStagesDetail, workflow.Id, workflow.Status, workflow.PodStatus, workflow.Message, bean2.CI_WORKFLOW_TYPE.String(), workflow.StartedOn, workflow.FinishedOn),
	}
	return workflowResponse, nil
}

func (impl *CiHandlerImpl) FetchArtifactsForCiJob(buildId int) (*types.ArtifactsForCiJob, error) {
	artifacts, err := impl.ciArtifactRepository.GetArtifactsByParentCiWorkflowId(buildId)
	if err != nil {
		impl.Logger.Errorw("error in fetching artifacts by parent ci workflow id", "err", err, "buildId", buildId)
		return nil, err
	}
	artifactsResponse := &types.ArtifactsForCiJob{
		Artifacts: artifacts,
	}
	return artifactsResponse, nil
}

func ExtractWorkflowStatus(workflowStatus eventProcessorBean.CiCdStatus) (string, string, string, string, string, string) {
	workflowName := ""
	status := string(workflowStatus.Phase)
	podStatus := ""
	message := ""
	podName := ""
	logLocation := ""
	for k, v := range workflowStatus.Nodes {
		if v.TemplateName == pipelineConfigBean.CI_WORKFLOW_NAME {
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
	return workflowName, status, podStatus, message, logLocation, podName
}

func (impl *CiHandlerImpl) UpdateWorkflow(workflowStatus eventProcessorBean.CiCdStatus) (int, bool, error) {
	workflowName, status, podStatus, message, _, podName := ExtractWorkflowStatus(workflowStatus)
	if workflowName == "" {
		impl.Logger.Errorw("extract workflow status, invalid wf name", "workflowName", workflowName, "status", status, "podStatus", podStatus, "message", message)
		return 0, false, errors.New("invalid wf name")
	}
	workflowId, err := strconv.Atoi(workflowName[:strings.Index(workflowName, "-")])
	if err != nil {
		impl.Logger.Errorw("invalid wf status update req", "err", err)
		return 0, false, err
	}

	savedWorkflow, err := impl.ciWorkflowRepository.FindById(workflowId)
	if err != nil {
		impl.Logger.Errorw("cannot get saved wf", "err", err)
		return 0, false, err
	}
	impl.updateResourceStatusInCache(workflowId, podName, savedWorkflow.Namespace, status)
	ciArtifactLocationFormat := impl.config.GetArtifactLocationFormat()
	ciArtifactLocation := fmt.Sprintf(ciArtifactLocationFormat, savedWorkflow.Id, savedWorkflow.Id)

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
		savedWorkflow.PodStatus = podStatus
		if savedWorkflow.ExecutorType == cdWorkflowBean.WORKFLOW_EXECUTOR_TYPE_SYSTEM && savedWorkflow.Status == cdWorkflowBean.WorkflowCancel {
			savedWorkflow.PodStatus = "Failed"
			savedWorkflow.Message = constants.TERMINATE_MESSAGE
		}
		savedWorkflow.Name = workflowName
		// savedWorkflow.LogLocation = "/ci-pipeline/" + strconv.Itoa(savedWorkflow.CiPipelineId) + "/workflow/" + strconv.Itoa(savedWorkflow.Id) + "/logs" //TODO need to fetch from workflow object
		// savedWorkflow.LogLocation = logLocation // removed because we are saving log location at trigger
		savedWorkflow.CiArtifactLocation = ciArtifactLocation
		savedWorkflow.PodName = podName
		impl.Logger.Debugw("updating workflow ", "workflow", savedWorkflow)
		err = impl.ciService.UpdateCiWorkflowWithStage(savedWorkflow)
		if err != nil {
			impl.Logger.Error("update wf failed for id " + strconv.Itoa(savedWorkflow.Id))
			return savedWorkflow.Id, true, err
		}
		impl.sendCIFailEvent(savedWorkflow, status, message)
		return savedWorkflow.Id, true, nil
	}
	return savedWorkflow.Id, false, nil
}

func (impl *CiHandlerImpl) sendCIFailEvent(savedWorkflow *pipelineConfig.CiWorkflow, status, message string) {
	if string(v1alpha1.NodeError) == savedWorkflow.Status || string(v1alpha1.NodeFailed) == savedWorkflow.Status {
		if executors.CheckIfReTriggerRequired(status, message, savedWorkflow.Status) {
			impl.Logger.Infow("not sending failure notification for re-trigger workflow", "workflowId", savedWorkflow.Id)
			return
		}
		impl.Logger.Warnw("ci failed for workflow: ", "wfId", savedWorkflow.Id)

		if extractErrorCode(savedWorkflow.Message) != workFlow.CiStageFailErrorCode {
			impl.ciService.WriteCIFailEvent(savedWorkflow)
		} else {
			impl.Logger.Infof("Step failed notification received for wfID %d with message %s", savedWorkflow.Id, savedWorkflow.Message)
		}
	}
}

func extractErrorCode(msg string) int {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(msg, -1)
	if len(matches) > 0 {
		code, err := strconv.Atoi(matches[0])
		if err == nil {
			return code
		}
	}
	return -1
}

func (impl *CiHandlerImpl) BuildPayload(ciWorkflow *pipelineConfig.CiWorkflow) *client.Payload {
	payload := &client.Payload{}
	payload.AppName = ciWorkflow.CiPipeline.App.AppName
	payload.PipelineName = ciWorkflow.CiPipeline.Name
	return payload
}

func (impl *CiHandlerImpl) stateChanged(status string, podStatus string, msg string,
	finishedAt time.Time, savedWorkflow *pipelineConfig.CiWorkflow) bool {
	return savedWorkflow.Status != status || savedWorkflow.PodStatus != podStatus || savedWorkflow.Message != msg || savedWorkflow.FinishedOn != finishedAt
}

func (impl *CiHandlerImpl) FetchCiStatusForTriggerViewV1(appId int) ([]*pipelineConfig.CiWorkflowStatus, error) {
	ciWorkflowStatuses, err := impl.workflowStatusUpdateService.FetchCiStatusForTriggerViewOptimized(appId)
	if err != nil {
		impl.Logger.Errorw("error in fetching ci status from optimized service, falling back to old method", "appId", appId, "err", err)
		// Fallback to old method if optimized service fails
		ciWorkflowStatuses, err = impl.ciWorkflowRepository.FIndCiWorkflowStatusesByAppId(appId)
		if err != nil && !util.IsErrNoRows(err) {
			impl.Logger.Errorw("err in fetching ciWorkflowStatuses from ciWorkflowRepository", "appId", appId, "err", err)
			return ciWorkflowStatuses, err
		}
	}

	return ciWorkflowStatuses, err
}

func (impl *CiHandlerImpl) FetchCiStatusForTriggerView(appId int) ([]*pipelineConfig.CiWorkflowStatus, error) {
	var ciWorkflowStatuses []*pipelineConfig.CiWorkflowStatus

	pipelines, err := impl.ciPipelineRepository.FindByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in fetching ci pipeline", "appId", appId, "err", err)
		return ciWorkflowStatuses, err
	}
	for _, pipeline := range pipelines {
		pipelineId := impl.getPipelineIdForTriggerView(pipeline)
		workflow, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflow(pipelineId)
		if err != nil && !util.IsErrNoRows(err) {
			impl.Logger.Errorw("err", "pipelineId", pipelineId, "err", err)
			return ciWorkflowStatuses, err
		}
		ciWorkflowStatus := &pipelineConfig.CiWorkflowStatus{}
		ciWorkflowStatus.CiPipelineId = pipeline.Id
		if workflow.Id > 0 {
			ciWorkflowStatus.CiPipelineName = workflow.CiPipeline.Name
			ciWorkflowStatus.CiStatus = workflow.Status
		} else {
			ciWorkflowStatus.CiStatus = "Not Triggered"
		}
		ciWorkflowStatuses = append(ciWorkflowStatuses, ciWorkflowStatus)
	}
	return ciWorkflowStatuses, nil
}

func (impl *CiHandlerImpl) FetchMaterialInfoByArtifactId(ciArtifactId int, envId int) (*types.GitTriggerInfoResponse, error) {

	ciArtifact, err := impl.ciArtifactRepository.Get(ciArtifactId)
	if err != nil {
		impl.Logger.Errorw("err", "ciArtifactId", ciArtifactId, "err", err)
		return &types.GitTriggerInfoResponse{}, err
	}

	ciPipeline, err := impl.ciPipelineRepository.FindByIdIncludingInActive(ciArtifact.PipelineId)
	if err != nil {
		impl.Logger.Errorw("err", "ciArtifactId", ciArtifactId, "err", err)
		return &types.GitTriggerInfoResponse{}, err
	}

	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(ciPipeline.Id)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return &types.GitTriggerInfoResponse{}, err
	}

	deployDetail, err := impl.appListingRepository.DeploymentDetailByArtifactId(ciArtifactId, envId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return &types.GitTriggerInfoResponse{}, err
	}

	ciMaterialsArr := make([]buildBean.CiPipelineMaterialResponse, 0)
	var triggeredByUserEmailId string
	//check workflow data only for non external builds
	if !ciPipeline.IsExternal {
		var workflow *pipelineConfig.CiWorkflow
		if ciArtifact.ParentCiArtifact > 0 {
			workflow, err = impl.ciWorkflowRepository.FindLastTriggeredWorkflowByArtifactId(ciArtifact.ParentCiArtifact)
			if err != nil {
				impl.Logger.Errorw("err", "ciArtifactId", ciArtifact.ParentCiArtifact, "err", err)
				return &types.GitTriggerInfoResponse{}, err
			}
		} else {
			workflow, err = impl.ciWorkflowRepository.FindLastTriggeredWorkflowByArtifactId(ciArtifactId)
			if err != nil {
				impl.Logger.Errorw("err", "ciArtifactId", ciArtifactId, "err", err)
				return &types.GitTriggerInfoResponse{}, err
			}
		}

		//getting the user including both active and inactive both
		// as there arises case of having the deleted user had triggered the deployment
		triggeredByUserEmailId, err = impl.userService.GetEmailById(int32(deployDetail.LastDeployedById))
		if err != nil && !util.IsErrNoRows(err) {
			impl.Logger.Errorw("err", "err", err)
			return &types.GitTriggerInfoResponse{}, err
		}

		for _, m := range ciMaterials {
			var history []*gitSensor.GitCommit
			_gitTrigger := workflow.GitTriggers[m.Id]

			// ignore git trigger which have commit and webhook both data nil
			if len(_gitTrigger.Commit) == 0 && _gitTrigger.WebhookData.Id == 0 {
				continue
			}

			_gitCommit := &gitSensor.GitCommit{
				Message: _gitTrigger.Message,
				Author:  _gitTrigger.Author,
				Date:    _gitTrigger.Date,
				Changes: _gitTrigger.Changes,
				Commit:  _gitTrigger.Commit,
			}

			// set webhook data
			_webhookData := _gitTrigger.WebhookData
			if _webhookData.Id > 0 {
				_gitCommit.WebhookData = &gitSensor.WebhookData{
					Id:              _webhookData.Id,
					EventActionType: _webhookData.EventActionType,
					Data:            _webhookData.Data,
				}
			}

			history = append(history, _gitCommit)

			res := buildBean.CiPipelineMaterialResponse{
				Id:              m.Id,
				GitMaterialId:   m.GitMaterialId,
				GitMaterialName: _gitTrigger.GitRepoName,
				Type:            string(m.Type),
				Value:           _gitTrigger.CiConfigureSourceValue,
				Active:          m.Active,
				Url:             _gitTrigger.GitRepoUrl,
				History:         history,
			}
			ciMaterialsArr = append(ciMaterialsArr, res)
		}
	}
	imageTaggingData, err := impl.imageTaggingService.GetTagsData(ciPipeline.Id, ciPipeline.AppId, ciArtifactId, false)
	if err != nil {
		impl.Logger.Errorw("error in fetching imageTaggingData", "err", err, "ciPipelineId", ciPipeline.Id, "appId", ciPipeline.AppId, "ciArtifactId", ciArtifactId)
		return &types.GitTriggerInfoResponse{}, err
	}
	gitTriggerInfoResponse := &types.GitTriggerInfoResponse{
		//GitTriggers:      workflow.GitTriggers,
		CiMaterials:      ciMaterialsArr,
		TriggeredByEmail: triggeredByUserEmailId,
		CiPipelineId:     ciPipeline.Id,
		AppId:            ciPipeline.AppId,
		AppName:          deployDetail.AppName,
		EnvironmentId:    deployDetail.EnvironmentId,
		EnvironmentName:  deployDetail.EnvironmentName,
		LastDeployedTime: deployDetail.LastDeployedTime,
		Default:          deployDetail.Default,
		ImageTaggingData: *imageTaggingData,
		Image:            ciArtifact.Image,
		TargetPlatforms:  utils.ConvertTargetPlatformStringToObject(ciArtifact.TargetPlatforms),
	}
	return gitTriggerInfoResponse, nil
}

func (impl *CiHandlerImpl) FetchCiStatusForTriggerViewForEnvironment(request resourceGroup.ResourceGroupingRequest, token string) ([]*pipelineConfig.CiWorkflowStatus, error) {
	ciWorkflowStatuses := make([]*pipelineConfig.CiWorkflowStatus, 0)
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
		cdPipelines, err = impl.cdPipelineRepository.FindActiveByInFilter(request.ParentResourceId, request.ResourceIds)
	} else {
		cdPipelines, err = impl.cdPipelineRepository.FindActiveByEnvId(request.ParentResourceId)
	}
	if err != nil {
		impl.Logger.Errorw("error in fetching pipelines", "request", request, "err", err)
		return nil, err
	}

	var appIds []int
	for _, pipeline := range cdPipelines {
		appIds = append(appIds, pipeline.AppId)
	}
	if len(appIds) == 0 {
		impl.Logger.Warnw("there is no app id found for fetching ci pipelines", "request", request)
		return ciWorkflowStatuses, nil
	}
	ciPipelines, err := impl.ciPipelineRepository.FindByAppIds(appIds)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in fetching ci pipeline", "err", err)
		return ciWorkflowStatuses, err
	}
	ciPipelineIds := make([]int, 0)
	for _, ciPipeline := range ciPipelines {
		ciPipelineIds = append(ciPipelineIds, ciPipeline.Id)
	}
	if len(ciPipelineIds) == 0 {
		return ciWorkflowStatuses, nil
	}
	// authorization block starts here
	var appObjectArr []string
	objects := impl.enforcerUtil.GetAppObjectByCiPipelineIds(ciPipelineIds)
	ciPipelineIds = []int{}
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object)
	}
	appResults, _ := request.CheckAuthBatch(token, appObjectArr, []string{})
	for _, ciPipeline := range ciPipelines {
		appObject := objects[ciPipeline.Id] // here only app permission have to check
		if !appResults[appObject] {
			// if user unauthorized, skip items
			continue
		}
		ciPipelineId := impl.getPipelineIdForTriggerView(ciPipeline)
		ciPipelineIds = append(ciPipelineIds, ciPipelineId)
	}
	if len(ciPipelineIds) == 0 {
		return ciWorkflowStatuses, nil
	}
	latestCiWorkflows, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflowByCiIds(ciPipelineIds)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", "ciPipelineIds", ciPipelineIds, "err", err)
		return ciWorkflowStatuses, err
	}

	notTriggeredWorkflows := make(map[int]bool)

	for _, ciWorkflow := range latestCiWorkflows {
		ciWorkflowStatus := &pipelineConfig.CiWorkflowStatus{}
		ciWorkflowStatus.CiPipelineId = ciWorkflow.CiPipelineId
		ciWorkflowStatus.CiPipelineName = ciWorkflow.CiPipeline.Name
		ciWorkflowStatus.CiStatus = ciWorkflow.Status
		ciWorkflowStatus.StorageConfigured = ciWorkflow.BlobStorageEnabled
		ciWorkflowStatus.CiWorkflowId = ciWorkflow.Id
		ciWorkflowStatuses = append(ciWorkflowStatuses, ciWorkflowStatus)
		notTriggeredWorkflows[ciWorkflowStatus.CiPipelineId] = true
	}

	for _, ciPipelineId := range ciPipelineIds {
		if _, ok := notTriggeredWorkflows[ciPipelineId]; !ok {
			ciWorkflowStatus := &pipelineConfig.CiWorkflowStatus{}
			ciWorkflowStatus.CiPipelineId = ciPipelineId
			ciWorkflowStatus.CiStatus = "Not Triggered"
			ciWorkflowStatuses = append(ciWorkflowStatuses, ciWorkflowStatus)
		}
	}
	return ciWorkflowStatuses, nil
}
