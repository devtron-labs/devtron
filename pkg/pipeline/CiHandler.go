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
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"github.com/devtron-labs/common-lib/utils/k8s"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository3 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	k8s2 "github.com/devtron-labs/devtron/pkg/k8s"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	resourceGroup "github.com/devtron-labs/devtron/pkg/resourceGroup"
	"github.com/devtron-labs/devtron/util/rbac"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	util2 "github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CiHandler interface {
	HandleCIWebhook(gitCiTriggerRequest bean.GitCiTriggerRequest) (int, error)
	HandleCIManual(ciTriggerRequest bean.CiTriggerRequest) (int, error)
	CheckAndReTriggerCI(workflowStatus v1alpha1.WorkflowStatus) error
	FetchMaterialsByPipelineId(pipelineId int, showAll bool) ([]pipelineConfig.CiPipelineMaterialResponse, error)
	FetchMaterialsByPipelineIdAndGitMaterialId(pipelineId int, gitMaterialId int, showAll bool) ([]pipelineConfig.CiPipelineMaterialResponse, error)
	FetchWorkflowDetails(appId int, pipelineId int, buildId int) (types.WorkflowResponse, error)
	FetchArtifactsForCiJob(buildId int) (*types.ArtifactsForCiJob, error)
	//FetchBuildById(appId int, pipelineId int) (WorkflowResponse, error)
	CancelBuild(workflowId int, forceAbort bool) (int, error)

	GetRunningWorkflowLogs(pipelineId int, workflowId int) (*bufio.Reader, func() error, error)
	GetHistoricBuildLogs(pipelineId int, workflowId int, ciWorkflow *pipelineConfig.CiWorkflow) (map[string]string, error)
	//SyncWorkflows() error

	GetBuildHistory(pipelineId int, appId int, offset int, size int) ([]types.WorkflowResponse, error)
	DownloadCiWorkflowArtifacts(pipelineId int, buildId int) (*os.File, error)
	UpdateWorkflow(workflowStatus v1alpha1.WorkflowStatus) (int, error)

	FetchCiStatusForTriggerView(appId int) ([]*pipelineConfig.CiWorkflowStatus, error)
	FetchCiStatusForTriggerViewV1(appId int) ([]*pipelineConfig.CiWorkflowStatus, error)
	RefreshMaterialByCiPipelineMaterialId(gitMaterialId int) (refreshRes *gitSensor.RefreshGitMaterialResponse, err error)
	FetchMaterialInfoByArtifactId(ciArtifactId int, envId int) (*types.GitTriggerInfoResponse, error)
	UpdateCiWorkflowStatusFailure(timeoutForFailureCiBuild int) error
	FetchCiStatusForTriggerViewForEnvironment(request resourceGroup.ResourceGroupingRequest, token string) ([]*pipelineConfig.CiWorkflowStatus, error)
}

type CiHandlerImpl struct {
	Logger                       *zap.SugaredLogger
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
	ciService                    CiService
	gitSensorClient              gitSensor.Client
	ciWorkflowRepository         pipelineConfig.CiWorkflowRepository
	workflowService              WorkflowService
	ciLogService                 CiLogService
	ciArtifactRepository         repository.CiArtifactRepository
	userService                  user.UserService
	eventClient                  client.EventClient
	eventFactory                 client.EventFactory
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
	appListingRepository         repository.AppListingRepository
	K8sUtil                      *k8s.K8sServiceImpl
	cdPipelineRepository         pipelineConfig.PipelineRepository
	enforcerUtil                 rbac.EnforcerUtil
	resourceGroupService         resourceGroup.ResourceGroupService
	envRepository                repository3.EnvironmentRepository
	imageTaggingService          ImageTaggingService
	customTagService             CustomTagService
	appWorkflowRepository        appWorkflow.AppWorkflowRepository
	config                       *types.CiConfig
	k8sCommonService             k8s2.K8sCommonService
	clusterService               cluster.ClusterService
	blobConfigStorageService     BlobStorageConfigService
	envService                   cluster.EnvironmentService
}

func NewCiHandlerImpl(Logger *zap.SugaredLogger, ciService CiService, ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository, gitSensorClient gitSensor.Client, ciWorkflowRepository pipelineConfig.CiWorkflowRepository, workflowService WorkflowService,
	ciLogService CiLogService, ciArtifactRepository repository.CiArtifactRepository, userService user.UserService, eventClient client.EventClient, eventFactory client.EventFactory, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	appListingRepository repository.AppListingRepository, K8sUtil *k8s.K8sServiceImpl, cdPipelineRepository pipelineConfig.PipelineRepository, enforcerUtil rbac.EnforcerUtil, resourceGroupService resourceGroup.ResourceGroupService, envRepository repository3.EnvironmentRepository,
	imageTaggingService ImageTaggingService, k8sCommonService k8s2.K8sCommonService, clusterService cluster.ClusterService, blobConfigStorageService BlobStorageConfigService, appWorkflowRepository appWorkflow.AppWorkflowRepository, customTagService CustomTagService,
	envService cluster.EnvironmentService) *CiHandlerImpl {
	cih := &CiHandlerImpl{
		Logger:                       Logger,
		ciService:                    ciService,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		gitSensorClient:              gitSensorClient,
		ciWorkflowRepository:         ciWorkflowRepository,
		workflowService:              workflowService,
		ciLogService:                 ciLogService,
		ciArtifactRepository:         ciArtifactRepository,
		userService:                  userService,
		eventClient:                  eventClient,
		eventFactory:                 eventFactory,
		ciPipelineRepository:         ciPipelineRepository,
		appListingRepository:         appListingRepository,
		K8sUtil:                      K8sUtil,
		cdPipelineRepository:         cdPipelineRepository,
		enforcerUtil:                 enforcerUtil,
		resourceGroupService:         resourceGroupService,
		envRepository:                envRepository,
		imageTaggingService:          imageTaggingService,
		customTagService:             customTagService,
		appWorkflowRepository:        appWorkflowRepository,
		k8sCommonService:             k8sCommonService,
		clusterService:               clusterService,
		blobConfigStorageService:     blobConfigStorageService,
		envService:                   envService,
	}
	config, err := types.GetCiConfig()
	if err != nil {
		return nil
	}
	cih.config = config

	return cih
}

const DefaultCiWorkflowNamespace = "devtron-ci"
const Running = "Running"
const Starting = "Starting"
const POD_DELETED_MESSAGE = "pod deleted"
const TERMINATE_MESSAGE = "workflow shutdown with strategy: Terminate"
const ABORT_MESSAGE_AFTER_STARTING_STAGE = "workflow shutdown with strategy: Force Abort"

func (impl *CiHandlerImpl) CheckAndReTriggerCI(workflowStatus v1alpha1.WorkflowStatus) error {

	//return if re-trigger feature is disabled
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

func (impl *CiHandlerImpl) reTriggerCi(retryCount int, refCiWorkflow *pipelineConfig.CiWorkflow) error {
	if retryCount >= impl.config.MaxCiWorkflowRetries {
		impl.Logger.Infow("maximum retries exhausted for this ciWorkflow", "ciWorkflowId", refCiWorkflow.Id, "retries", retryCount, "configuredRetries", impl.config.MaxCiWorkflowRetries)
		return nil
	}
	impl.Logger.Infow("re-triggering ci for a ci workflow", "ReferenceCiWorkflowId", refCiWorkflow.Id)
	ciPipelineMaterialIds := make([]int, 0, len(refCiWorkflow.GitTriggers))
	for id, _ := range refCiWorkflow.GitTriggers {
		ciPipelineMaterialIds = append(ciPipelineMaterialIds, id)
	}
	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByIdsIncludeDeleted(ciPipelineMaterialIds)
	if err != nil {
		impl.Logger.Errorw("error in getting ci Pipeline Materials using ciPipeline Material Ids", "ciPipelineMaterialIds", ciPipelineMaterialIds, "err", err)
		return err
	}

	trigger := types.Trigger{}
	trigger.BuildTriggerObject(refCiWorkflow, ciMaterials, 1, true, nil, "")
	_, err = impl.ciService.TriggerCiPipeline(trigger)

	if err != nil {
		impl.Logger.Errorw("error occurred in re-triggering ciWorkflow", "triggerDetails", trigger, "err", err)
		return err
	}
	return nil
}

func (impl *CiHandlerImpl) HandleCIManual(ciTriggerRequest bean.CiTriggerRequest) (int, error) {
	impl.Logger.Debugw("HandleCIManual for pipeline ", "PipelineId", ciTriggerRequest.PipelineId)
	commitHashes, extraEnvironmentVariables, err := impl.buildManualTriggerCommitHashes(ciTriggerRequest)
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

	trigger := types.Trigger{
		PipelineId:                ciTriggerRequest.PipelineId,
		CommitHashes:              commitHashes,
		CiMaterials:               nil,
		TriggeredBy:               ciTriggerRequest.TriggeredBy,
		InvalidateCache:           ciTriggerRequest.InvalidateCache,
		ExtraEnvironmentVariables: extraEnvironmentVariables,
		EnvironmentId:             ciTriggerRequest.EnvironmentId,
		PipelineType:              ciTriggerRequest.PipelineType,
		CiArtifactLastFetch:       createdOn,
	}
	id, err := impl.ciService.TriggerCiPipeline(trigger)

	if err != nil {
		return 0, err
	}
	return id, nil
}

func (impl *CiHandlerImpl) HandleCIWebhook(gitCiTriggerRequest bean.GitCiTriggerRequest) (int, error) {
	impl.Logger.Debugw("HandleCIWebhook for material ", "material", gitCiTriggerRequest.CiPipelineMaterial)
	ciPipeline, err := impl.GetCiPipeline(gitCiTriggerRequest.CiPipelineMaterial.Id)
	if err != nil {
		return 0, err
	}
	if ciPipeline.IsManual {
		impl.Logger.Debugw("not handling manual pipeline", "pipelineId", ciPipeline.Id)
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

	commitHashes, err := impl.buildAutomaticTriggerCommitHashes(ciMaterials, gitCiTriggerRequest)
	if err != nil {
		return 0, err
	}

	trigger := types.Trigger{
		PipelineId:                ciPipeline.Id,
		CommitHashes:              commitHashes,
		CiMaterials:               ciMaterials,
		TriggeredBy:               gitCiTriggerRequest.TriggeredBy,
		ExtraEnvironmentVariables: gitCiTriggerRequest.ExtraEnvironmentVariables,
	}
	id, err := impl.ciService.TriggerCiPipeline(trigger)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (impl *CiHandlerImpl) validateBuildSequence(gitCiTriggerRequest bean.GitCiTriggerRequest, pipelineId int) (bool, error) {
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

	if ciPipelineMaterial.Type == string(pipelineConfig.SOURCE_TYPE_BRANCH_FIXED) {
		if ciPipelineMaterial.GitCommit.Date.Before(lastTriggeredBuild.GitTriggers[ciPipelineMaterial.Id].Date) {
			impl.Logger.Warnw("older commit cannot be built for pipeline", "pipelineId", pipelineId, "ciMaterial", gitCiTriggerRequest.CiPipelineMaterial.Id)
			isValid = false
		}
	}

	return isValid, nil
}

func (impl *CiHandlerImpl) RefreshMaterialByCiPipelineMaterialId(gitMaterialId int) (refreshRes *gitSensor.RefreshGitMaterialResponse, err error) {
	impl.Logger.Debugw("refreshing git material", "id", gitMaterialId)
	refreshRes, err = impl.gitSensorClient.RefreshGitMaterial(context.Background(),
		&gitSensor.RefreshGitMaterialRequest{GitMaterialId: gitMaterialId},
	)
	return refreshRes, err
}

func (impl *CiHandlerImpl) FetchMaterialsByPipelineIdAndGitMaterialId(pipelineId int, gitMaterialId int, showAll bool) ([]pipelineConfig.CiPipelineMaterialResponse, error) {
	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineIdAndGitMaterialId(pipelineId, gitMaterialId)
	if err != nil {
		impl.Logger.Errorw("ciMaterials fetch failed", "err", err)
	}
	var ciPipelineMaterialResponses []pipelineConfig.CiPipelineMaterialResponse
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
			return []pipelineConfig.CiPipelineMaterialResponse{}, apiErr
		}
		ciMaterialHistoryMap[m] = changesResp
	}

	for k, v := range ciMaterialHistoryMap {
		r := pipelineConfig.CiPipelineMaterialResponse{
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
		return []pipelineConfig.CiPipelineMaterialResponse{}, err
	}
	for _, k := range regexMaterials {
		r := pipelineConfig.CiPipelineMaterialResponse{
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

func (impl *CiHandlerImpl) FetchMaterialsByPipelineId(pipelineId int, showAll bool) ([]pipelineConfig.CiPipelineMaterialResponse, error) {
	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(pipelineId)
	if err != nil {
		impl.Logger.Errorw("ciMaterials fetch failed", "err", err)
	}
	var ciPipelineMaterialResponses []pipelineConfig.CiPipelineMaterialResponse
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
			return []pipelineConfig.CiPipelineMaterialResponse{}, apiErr
		}
		ciMaterialHistoryMap[m] = changesResp
	}

	for k, v := range ciMaterialHistoryMap {
		r := pipelineConfig.CiPipelineMaterialResponse{
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
		return []pipelineConfig.CiPipelineMaterialResponse{}, err
	}
	for _, k := range regexMaterials {
		r := pipelineConfig.CiPipelineMaterialResponse{
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
	var ciPipelineMaterialResponses []pipelineConfig.CiPipelineMaterialResponse
	for _, m := range ciMaterials {
		r := pipelineConfig.CiPipelineMaterialResponse{
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
	//this map contains artifactId -> array of tags of that artifact
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
	var artifactIds []int
	for _, w := range workFlows {
		artifactIds = append(artifactIds, w.CiArtifactId)
	}
	//this map contains artifactId -> imageComment of that artifact
	imageCommetnsDataMap, err := impl.imageTaggingService.GetImageCommentsDataMapByArtifactIds(artifactIds)
	if err != nil {
		impl.Logger.Errorw("error in fetching imageCommetnsDataMap", "err", err, "appId", appId, "artifactIds", artifactIds)
		return nil, err
	}

	var ciWorkLowResponses []types.WorkflowResponse
	for _, w := range workFlows {
		wfResponse := types.WorkflowResponse{
			Id:                  w.Id,
			Name:                w.Name,
			Status:              w.Status,
			PodStatus:           w.PodStatus,
			Message:             w.Message,
			StartedOn:           w.StartedOn,
			FinishedOn:          w.FinishedOn,
			CiPipelineId:        w.CiPipelineId,
			Namespace:           w.Namespace,
			LogLocation:         w.LogFilePath,
			GitTriggers:         w.GitTriggers,
			CiMaterials:         ciPipelineMaterialResponses,
			Artifact:            w.Image,
			TriggeredBy:         w.TriggeredBy,
			TriggeredByEmail:    w.EmailId,
			ArtifactId:          w.CiArtifactId,
			BlobStorageEnabled:  w.BlobStorageEnabled,
			IsArtifactUploaded:  w.IsArtifactUploaded,
			EnvironmentId:       w.EnvironmentId,
			EnvironmentName:     w.EnvironmentName,
			ReferenceWorkflowId: w.RefCiWorkflowId,
			PodName:             w.PodName,
		}
		if w.Message == bean3.ImageTagUnavailableMessage {
			customTag, err := impl.customTagService.GetCustomTagByEntityKeyAndValue(bean3.EntityTypeCiPipelineId, strconv.Itoa(w.CiPipelineId))
			if err != nil && err != pg.ErrNoRows {
				//err == pg.ErrNoRows should never happen
				return nil, err
			}
			appWorkflows, err := impl.appWorkflowRepository.FindWFCIMappingByCIPipelineId(w.CiPipelineId)
			if err != nil && err != pg.ErrNoRows {
				return nil, err
			}
			wfResponse.AppWorkflowId = appWorkflows[0].AppWorkflowId //it is guaranteed there will always be 1 entry (in case of ci_pipeline_id)
			wfResponse.CustomTag = &bean2.CustomTagErrorResponse{
				TagPattern:           customTag.TagPattern,
				AutoIncreasingNumber: customTag.AutoIncreasingNumber,
				Message:              bean3.ImageTagUnavailableMessage,
			}
		}
		if imageTagsDataMap[w.CiArtifactId] != nil {
			wfResponse.ImageReleaseTags = imageTagsDataMap[w.CiArtifactId] //if artifact is not yet created,empty list will be sent
		}
		if imageCommetnsDataMap[w.CiArtifactId] != nil {
			wfResponse.ImageComment = imageCommetnsDataMap[w.CiArtifactId]
		}
		ciWorkLowResponses = append(ciWorkLowResponses, wfResponse)
	}
	return ciWorkLowResponses, nil
}

func (impl *CiHandlerImpl) CancelBuild(workflowId int, forceAbort bool) (int, error) {
	workflow, err := impl.ciWorkflowRepository.FindById(workflowId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return 0, err
	}
	if !(string(v1alpha1.NodePending) == workflow.Status || string(v1alpha1.NodeRunning) == workflow.Status) {
		if forceAbort {
			return impl.cancelBuildAfterStartWorkflowStage(workflow)
		} else {
			return 0, &util.ApiError{Code: "200", HttpStatusCode: 400, UserMessage: "cannot cancel build, build not in progress"}
		}
	}
	//this arises when someone deletes the workflow in resource browser and wants to force abort a ci
	if workflow.Status == string(v1alpha1.NodeRunning) && forceAbort {
		return impl.cancelBuildAfterStartWorkflowStage(workflow)
	}
	isExt := workflow.Namespace != DefaultCiWorkflowNamespace
	var env *repository3.Environment
	var restConfig *rest.Config
	if isExt {
		restConfig, err = impl.getRestConfig(workflow)
		if err != nil {
			return 0, err
		}
	}

	// Terminate workflow
	err = impl.workflowService.TerminateWorkflow(workflow.ExecutorType, workflow.Name, workflow.Namespace, restConfig, isExt, env)
	if err != nil && strings.Contains(err.Error(), "cannot find workflow") {
		return 0, &util.ApiError{Code: "200", HttpStatusCode: http.StatusBadRequest, UserMessage: err.Error()}
	} else if err != nil {
		impl.Logger.Errorw("cannot terminate wf", "err", err)
		return 0, err
	}

	workflow.Status = executors.WorkflowCancel
	if workflow.ExecutorType == pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM {
		workflow.PodStatus = "Failed"
		workflow.Message = TERMINATE_MESSAGE
	}
	err = impl.ciWorkflowRepository.UpdateWorkFlow(workflow)
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

func (impl *CiHandlerImpl) cancelBuildAfterStartWorkflowStage(workflow *pipelineConfig.CiWorkflow) (int, error) {
	workflow.Status = executors.WorkflowCancel
	workflow.PodStatus = string(bean.Failed)
	workflow.Message = ABORT_MESSAGE_AFTER_STARTING_STAGE
	err := impl.ciWorkflowRepository.UpdateWorkFlow(workflow)
	if err != nil {
		impl.Logger.Errorw("error in updating workflow status", "err", err)
		return 0, err
	}
	return workflow.Id, nil
}

func (impl *CiHandlerImpl) getRestConfig(workflow *pipelineConfig.CiWorkflow) (*rest.Config, error) {
	env, err := impl.envRepository.FindById(workflow.EnvironmentId)
	if err != nil {
		impl.Logger.Errorw("could not fetch stage env", "err", err)
		return nil, err
	}

	clusterBean := cluster.GetClusterBean(*env.Cluster)

	clusterConfig, err := clusterBean.GetClusterConfig()
	if err != nil {
		impl.Logger.Errorw("error in getting cluster config", "err", err, "clusterId", clusterBean.Id)
		return nil, err
	}
	restConfig, err := impl.K8sUtil.GetRestConfigByCluster(clusterConfig)
	if err != nil {
		impl.Logger.Errorw("error in getting rest config by cluster id", "err", err)
		return nil, err
	}
	return restConfig, nil
}

func (impl *CiHandlerImpl) FetchWorkflowDetails(appId int, pipelineId int, buildId int) (types.WorkflowResponse, error) {
	workflow, err := impl.ciWorkflowRepository.FindById(buildId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return types.WorkflowResponse{}, err
	}
	triggeredByUserEmailId, err := impl.userService.GetEmailById(workflow.TriggeredBy)
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
	environmentName := ""
	if workflow.EnvironmentId != 0 {
		env, err := impl.envRepository.FindById(workflow.EnvironmentId)
		if err != nil && err != pg.ErrNoRows {
			impl.Logger.Errorw("error in fetching environment details ", "err", err)
			return types.WorkflowResponse{}, err
		}
		environmentName = env.Name
	}
	workflowResponse := types.WorkflowResponse{
		Id:                 workflow.Id,
		Name:               workflow.Name,
		Status:             workflow.Status,
		PodStatus:          workflow.PodStatus,
		Message:            workflow.Message,
		StartedOn:          workflow.StartedOn,
		FinishedOn:         workflow.FinishedOn,
		CiPipelineId:       workflow.CiPipelineId,
		Namespace:          workflow.Namespace,
		LogLocation:        workflow.LogLocation,
		BlobStorageEnabled: workflow.BlobStorageEnabled, //TODO default value if value not found in db
		GitTriggers:        workflow.GitTriggers,
		CiMaterials:        ciMaterialsArr,
		TriggeredBy:        workflow.TriggeredBy,
		TriggeredByEmail:   triggeredByUserEmailId,
		Artifact:           ciArtifact.Image,
		ArtifactId:         ciArtifact.Id,
		IsArtifactUploaded: ciArtifact.IsArtifactUploaded,
		EnvironmentId:      workflow.EnvironmentId,
		EnvironmentName:    environmentName,
		PipelineType:       workflow.CiPipeline.PipelineType,
		PodName:            workflow.PodName,
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
func (impl *CiHandlerImpl) GetRunningWorkflowLogs(pipelineId int, workflowId int) (*bufio.Reader, func() error, error) {
	ciWorkflow, err := impl.ciWorkflowRepository.FindById(workflowId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return nil, nil, err
	}
	return impl.getWorkflowLogs(pipelineId, ciWorkflow)
}

func (impl *CiHandlerImpl) getWorkflowLogs(pipelineId int, ciWorkflow *pipelineConfig.CiWorkflow) (*bufio.Reader, func() error, error) {
	if string(v1alpha1.NodePending) == ciWorkflow.PodStatus {
		return bufio.NewReader(strings.NewReader("")), nil, nil
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
		var clusterBean cluster.ClusterBean
		if env != nil && env.Cluster != nil {
			clusterBean = cluster.GetClusterBean(*env.Cluster)
		}
		clusterConfig, err = clusterBean.GetClusterConfig()
		if err != nil {
			impl.Logger.Errorw("error in getting cluster config", "err", err, "clusterId", clusterBean.Id)
			return nil, nil, err
		}
		isExt = true
	}

	logStream, cleanUp, err := impl.ciLogService.FetchRunningWorkflowLogs(ciLogRequest, clusterConfig, isExt)
	if logStream == nil || err != nil {
		if !ciWorkflow.BlobStorageEnabled {
			return nil, nil, &util.ApiError{Code: "200", HttpStatusCode: 400, UserMessage: "logs-not-stored-in-repository"}
		} else if string(v1alpha1.NodeSucceeded) == ciWorkflow.Status || string(v1alpha1.NodeError) == ciWorkflow.Status || string(v1alpha1.NodeFailed) == ciWorkflow.Status || ciWorkflow.Status == executors.WorkflowCancel {
			impl.Logger.Errorw("err", "err", err)
			return impl.getLogsFromRepository(pipelineId, ciWorkflow, clusterConfig, isExt)
		}
		impl.Logger.Errorw("err", "err", err)
		return nil, nil, &util.ApiError{Code: "200", HttpStatusCode: 400, UserMessage: err.Error()}
	}
	logReader := bufio.NewReader(logStream)
	return logReader, cleanUp, err
}

func (impl *CiHandlerImpl) getLogsFromRepository(pipelineId int, ciWorkflow *pipelineConfig.CiWorkflow, clusterConfig *k8s.ClusterConfig, isExt bool) (*bufio.Reader, func() error, error) {
	impl.Logger.Debug("getting historic logs")

	ciConfig, err := impl.ciWorkflowRepository.FindConfigByPipelineId(pipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", "err", err)
		return nil, nil, err
	}

	if ciConfig.LogsBucket == "" {
		ciConfig.LogsBucket = impl.config.GetDefaultBuildLogsBucket()
	}
	if ciConfig.CiCacheRegion == "" {
		ciConfig.CiCacheRegion = impl.config.DefaultCacheBucketRegion
	}
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
			BucketName:        ciConfig.LogsBucket,
			Region:            ciConfig.CiCacheRegion,
			VersioningEnabled: impl.config.BlobStorageS3BucketVersioned,
		},
		GcpBlobBaseConfig: &blob_storage.GcpBlobBaseConfig{
			BucketName:             ciConfig.LogsBucket,
			CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
		},
	}
	useExternalBlobStorage := isExternalBlobStorageEnabled(isExt, impl.config.UseBlobStorageConfigInCiWorkflow)
	if useExternalBlobStorage {
		//fetch extClusterBlob cm and cs from k8s client, if they are present then read creds
		//from them else return.
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

func (impl *CiHandlerImpl) DownloadCiWorkflowArtifacts(pipelineId int, buildId int) (*os.File, error) {
	ciWorkflow, err := impl.ciWorkflowRepository.FindById(buildId)
	if err != nil {
		impl.Logger.Errorw("unable to fetch ciWorkflow", "err", err)
		return nil, err
	}
	useExternalBlobStorage := isExternalBlobStorageEnabled(ciWorkflow.IsExternalRunInJobType(), impl.config.UseBlobStorageConfigInCiWorkflow)
	if !ciWorkflow.BlobStorageEnabled {
		return nil, errors.New("logs-not-stored-in-repository")
	}

	if ciWorkflow.CiPipelineId != pipelineId {
		impl.Logger.Error("invalid request, wf not in pipeline")
		return nil, errors.New("invalid request, wf not in pipeline")
	}

	ciConfig, err := impl.ciWorkflowRepository.FindConfigByPipelineId(pipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("unable to fetch ciCdConfig", "err", err)
		return nil, err
	}

	if ciConfig.LogsBucket == "" {
		ciConfig.LogsBucket = impl.config.GetDefaultBuildLogsBucket()
	}

	item := strconv.Itoa(ciWorkflow.Id)
	if ciConfig.CiCacheRegion == "" {
		ciConfig.CiCacheRegion = impl.config.DefaultCacheBucketRegion
	}
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
		BucketName:        ciConfig.LogsBucket,
		Region:            ciConfig.CiCacheRegion,
		VersioningEnabled: impl.config.BlobStorageS3BucketVersioned,
	}
	gcpBlobBaseConfig := &blob_storage.GcpBlobBaseConfig{
		BucketName:             ciConfig.LogsBucket,
		CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
	}

	key := fmt.Sprintf("%s/"+impl.config.GetArtifactLocationFormat(), impl.config.GetDefaultArtifactKeyPrefix(), ciWorkflow.Id, ciWorkflow.Id)

	baseLogLocationPathConfig := impl.config.BaseLogLocationPath
	blobStorageService := blob_storage.NewBlobStorageServiceImpl(nil)
	destinationKey := filepath.Clean(filepath.Join(baseLogLocationPathConfig, item))
	request := &blob_storage.BlobStorageRequest{
		StorageType:         impl.config.CloudProvider,
		SourceKey:           key,
		DestinationKey:      baseLogLocationPathConfig + item,
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
		//fetch extClusterBlob cm and cs from k8s client, if they are present then read creds
		//from them else return.
		cmConfig, secretConfig, err := impl.blobConfigStorageService.FetchCmAndSecretBlobConfigFromExternalCluster(clusterConfig, ciWorkflow.Namespace)
		if err != nil {
			impl.Logger.Errorw("error in fetching config map and secret from external cluster", "err", err, "clusterConfig", clusterConfig)
			return nil, err
		}
		request = updateRequestWithExtClusterCmAndSecret(request, cmConfig, secretConfig)
	}
	_, numBytes, err := blobStorageService.Get(request)
	if err != nil {
		impl.Logger.Errorw("error occurred while downloading file", "request", request)
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

func (impl *CiHandlerImpl) GetHistoricBuildLogs(pipelineId int, workflowId int, ciWorkflow *pipelineConfig.CiWorkflow) (map[string]string, error) {
	ciConfig, err := impl.ciWorkflowRepository.FindConfigByPipelineId(pipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", "err", err)
		return nil, err
	}
	if ciWorkflow == nil {
		ciWorkflow, err = impl.ciWorkflowRepository.FindById(workflowId)
		if err != nil {
			impl.Logger.Errorw("err", "err", err)
			return nil, err
		}
	}
	if ciConfig.LogsBucket == "" {
		ciConfig.LogsBucket = impl.config.GetDefaultBuildLogsBucket()
	}
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
			BucketName:        ciConfig.LogsBucket,
			Region:            ciConfig.CiCacheRegion,
			VersioningEnabled: impl.config.BlobStorageS3BucketVersioned,
		},
		GcpBlobBaseConfig: &blob_storage.GcpBlobBaseConfig{
			BucketName:             ciConfig.LogsBucket,
			CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
		},
	}
	useExternalBlobStorage := isExternalBlobStorageEnabled(ciWorkflow.IsExternalRunInJobType(), impl.config.UseBlobStorageConfigInCiWorkflow)
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
		//fetch extClusterBlob cm and cs from k8s client, if they are present then read creds
		//from them else return.
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

func ExtractWorkflowStatus(workflowStatus v1alpha1.WorkflowStatus) (string, string, string, string, string, string) {
	workflowName := ""
	status := string(workflowStatus.Phase)
	podStatus := ""
	message := ""
	podName := ""
	logLocation := ""
	for k, v := range workflowStatus.Nodes {
		if v.TemplateName == bean3.CI_WORKFLOW_NAME {
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

const CiStageFailErrorCode = 2

func (impl *CiHandlerImpl) extractPodStatusAndWorkflow(workflowStatus v1alpha1.WorkflowStatus) (string, string, *pipelineConfig.CiWorkflow, error) {
	workflowName, status, _, message, _, _ := ExtractWorkflowStatus(workflowStatus)
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

func (impl *CiHandlerImpl) getRefWorkflowAndCiRetryCount(savedWorkflow *pipelineConfig.CiWorkflow) (int, *pipelineConfig.CiWorkflow, error) {
	var err error

	if savedWorkflow.ReferenceCiWorkflowId != 0 {
		savedWorkflow, err = impl.ciWorkflowRepository.FindById(savedWorkflow.ReferenceCiWorkflowId)
	}
	if err != nil {
		impl.Logger.Errorw("cannot get saved wf", "err", err)
		return 0, savedWorkflow, err
	}
	retryCount, err := impl.ciWorkflowRepository.FindRetriedWorkflowCountByReferenceId(savedWorkflow.Id)
	return retryCount, savedWorkflow, err
}

func (impl *CiHandlerImpl) UpdateWorkflow(workflowStatus v1alpha1.WorkflowStatus) (int, error) {
	workflowName, status, podStatus, message, _, podName := ExtractWorkflowStatus(workflowStatus)
	if workflowName == "" {
		impl.Logger.Errorw("extract workflow status, invalid wf name", "workflowName", workflowName, "status", status, "podStatus", podStatus, "message", message)
		return 0, errors.New("invalid wf name")
	}
	workflowId, err := strconv.Atoi(workflowName[:strings.Index(workflowName, "-")])
	if err != nil {
		impl.Logger.Errorw("invalid wf status update req", "err", err)
		return 0, err
	}

	savedWorkflow, err := impl.ciWorkflowRepository.FindById(workflowId)
	if err != nil {
		impl.Logger.Errorw("cannot get saved wf", "err", err)
		return 0, err
	}

	ciWorkflowConfig, err := impl.ciWorkflowRepository.FindConfigByPipelineId(savedWorkflow.CiPipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("unable to fetch ciWorkflowConfig", "err", err)
		return 0, err
	}

	ciArtifactLocationFormat := ciWorkflowConfig.CiArtifactLocationFormat
	if ciArtifactLocationFormat == "" {
		ciArtifactLocationFormat = impl.config.GetArtifactLocationFormat()
	}
	ciArtifactLocation := fmt.Sprintf(ciArtifactLocationFormat, ciWorkflowConfig.LogsBucket, savedWorkflow.Id, savedWorkflow.Id)

	if impl.stateChanged(status, podStatus, message, workflowStatus.FinishedAt.Time, savedWorkflow) {
		if savedWorkflow.Status != executors.WorkflowCancel {
			savedWorkflow.Status = status
		}
		savedWorkflow.PodStatus = podStatus
		savedWorkflow.Message = message
		// NOTE: we are doing this for a quick fix where ci pending message become larger than 250 and in db we had set the charter limit to 250
		if len(message) > 250 {
			savedWorkflow.Message = message[:250]
		}
		if savedWorkflow.ExecutorType == pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM && savedWorkflow.Status == executors.WorkflowCancel {
			savedWorkflow.PodStatus = "Failed"
			savedWorkflow.Message = TERMINATE_MESSAGE
		}
		savedWorkflow.FinishedOn = workflowStatus.FinishedAt.Time
		savedWorkflow.Name = workflowName
		//savedWorkflow.LogLocation = "/ci-pipeline/" + strconv.Itoa(savedWorkflow.CiPipelineId) + "/workflow/" + strconv.Itoa(savedWorkflow.Id) + "/logs" //TODO need to fetch from workflow object
		//savedWorkflow.LogLocation = logLocation // removed because we are saving log location at trigger
		savedWorkflow.CiArtifactLocation = ciArtifactLocation
		savedWorkflow.PodName = podName
		impl.Logger.Debugw("updating workflow ", "workflow", savedWorkflow)
		err = impl.ciWorkflowRepository.UpdateWorkFlow(savedWorkflow)
		if err != nil {
			impl.Logger.Error("update wf failed for id " + strconv.Itoa(savedWorkflow.Id))
			return 0, err
		}
		if string(v1alpha1.NodeError) == savedWorkflow.Status || string(v1alpha1.NodeFailed) == savedWorkflow.Status {
			impl.Logger.Warnw("ci failed for workflow: ", "wfId", savedWorkflow.Id)

			if extractErrorCode(savedWorkflow.Message) != CiStageFailErrorCode {
				go impl.WriteCIFailEvent(savedWorkflow, ciWorkflowConfig.CiImage)
			} else {
				impl.Logger.Infof("Step failed notification received for wfID %d with message %s", savedWorkflow.Id, savedWorkflow.Message)
			}
		}
	}
	return savedWorkflow.Id, nil
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

func (impl *CiHandlerImpl) WriteCIFailEvent(ciWorkflow *pipelineConfig.CiWorkflow, ciImage string) {
	event := impl.eventFactory.Build(util2.Fail, &ciWorkflow.CiPipelineId, ciWorkflow.CiPipeline.AppId, nil, util2.CI)
	material := &client.MaterialTriggerInfo{}
	material.GitTriggers = ciWorkflow.GitTriggers
	event.CiWorkflowRunnerId = ciWorkflow.Id
	event.UserId = int(ciWorkflow.TriggeredBy)
	event = impl.eventFactory.BuildExtraCIData(event, material, ciImage)
	event.CiArtifactId = 0
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.Logger.Errorw("error in writing event", "err", evtErr)
	}
}

func (impl *CiHandlerImpl) BuildPayload(ciWorkflow *pipelineConfig.CiWorkflow) *client.Payload {
	payload := &client.Payload{}
	payload.AppName = ciWorkflow.CiPipeline.App.AppName
	payload.PipelineName = ciWorkflow.CiPipeline.Name
	//payload["buildName"] = ciWorkflow.Name
	//payload["podStatus"] = ciWorkflow.PodStatus
	//payload["message"] = ciWorkflow.Message
	return payload
}

func (impl *CiHandlerImpl) stateChanged(status string, podStatus string, msg string,
	finishedAt time.Time, savedWorkflow *pipelineConfig.CiWorkflow) bool {
	return savedWorkflow.Status != status || savedWorkflow.PodStatus != podStatus || savedWorkflow.Message != msg || savedWorkflow.FinishedOn != finishedAt
}

func (impl *CiHandlerImpl) GetCiPipeline(ciMaterialId int) (*pipelineConfig.CiPipeline, error) {
	ciMaterial, err := impl.ciPipelineMaterialRepository.GetById(ciMaterialId)
	if err != nil {
		return nil, err
	}
	ciPipeline := ciMaterial.CiPipeline
	return ciPipeline, nil
}

func (impl *CiHandlerImpl) buildAutomaticTriggerCommitHashes(ciMaterials []*pipelineConfig.CiPipelineMaterial, request bean.GitCiTriggerRequest) (map[int]pipelineConfig.GitCommit, error) {
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

func SetGitCommitValuesForBuildingCommitHash(ciMaterial *pipelineConfig.CiPipelineMaterial, oldGitCommit pipelineConfig.GitCommit) pipelineConfig.GitCommit {
	newGitCommit := oldGitCommit
	newGitCommit.CiConfigureSourceType = ciMaterial.Type
	newGitCommit.CiConfigureSourceValue = ciMaterial.Value
	newGitCommit.GitRepoUrl = ciMaterial.GitMaterial.Url
	newGitCommit.GitRepoName = ciMaterial.GitMaterial.Name[strings.Index(ciMaterial.GitMaterial.Name, "-")+1:]
	return newGitCommit
}

func (impl *CiHandlerImpl) buildManualTriggerCommitHashes(ciTriggerRequest bean.CiTriggerRequest) (map[int]pipelineConfig.GitCommit, map[string]string, error) {
	commitHashes := map[int]pipelineConfig.GitCommit{}
	extraEnvironmentVariables := make(map[string]string)
	for _, ciPipelineMaterial := range ciTriggerRequest.CiPipelineMaterial {

		pipeLineMaterialFromDb, err := impl.ciPipelineMaterialRepository.GetById(ciPipelineMaterial.Id)
		if err != nil {
			impl.Logger.Errorw("err in fetching pipeline material by id", "err", err)
			return map[int]pipelineConfig.GitCommit{}, nil, err
		}

		pipelineType := pipeLineMaterialFromDb.Type
		if pipelineType == pipelineConfig.SOURCE_TYPE_BRANCH_FIXED {
			gitCommit, err := impl.BuildManualTriggerCommitHashesForSourceTypeBranchFix(ciPipelineMaterial, pipeLineMaterialFromDb)
			if err != nil {
				impl.Logger.Errorw("err", "err", err)
				return map[int]pipelineConfig.GitCommit{}, nil, err
			}
			commitHashes[ciPipelineMaterial.Id] = gitCommit

		} else if pipelineType == pipelineConfig.SOURCE_TYPE_WEBHOOK {
			gitCommit, extraEnvVariables, err := impl.BuildManualTriggerCommitHashesForSourceTypeWebhook(ciPipelineMaterial, pipeLineMaterialFromDb)
			if err != nil {
				impl.Logger.Errorw("err", "err", err)
				return map[int]pipelineConfig.GitCommit{}, nil, err
			}
			commitHashes[ciPipelineMaterial.Id] = gitCommit
			extraEnvironmentVariables = extraEnvVariables
		}

	}
	return commitHashes, extraEnvironmentVariables, nil
}

func (impl *CiHandlerImpl) BuildManualTriggerCommitHashesForSourceTypeBranchFix(ciPipelineMaterial bean.CiPipelineMaterial, pipeLineMaterialFromDb *pipelineConfig.CiPipelineMaterial) (pipelineConfig.GitCommit, error) {
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

func (impl *CiHandlerImpl) BuildManualTriggerCommitHashesForSourceTypeWebhook(ciPipelineMaterial bean.CiPipelineMaterial, pipeLineMaterialFromDb *pipelineConfig.CiPipelineMaterial) (pipelineConfig.GitCommit, map[string]string, error) {
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

func (impl *CiHandlerImpl) getLastSeenCommit(ciMaterialId int) (pipelineConfig.GitCommit, error) {
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

func (impl *CiHandlerImpl) FetchCiStatusForTriggerViewV1(appId int) ([]*pipelineConfig.CiWorkflowStatus, error) {
	ciWorkflowStatuses, err := impl.ciWorkflowRepository.FIndCiWorkflowStatusesByAppId(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err in fetching ciWorkflowStatuses from ciWorkflowRepository", "appId", appId, "err", err)
		return ciWorkflowStatuses, err
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
		pipelineId := 0
		if pipeline.ParentCiPipeline == 0 {
			pipelineId = pipeline.Id
		} else {
			pipelineId = pipeline.ParentCiPipeline
		}
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

	ciMaterialsArr := make([]pipelineConfig.CiPipelineMaterialResponse, 0)
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

		triggeredByUserEmailId, err = impl.userService.GetEmailById(workflow.TriggeredBy)
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

			res := pipelineConfig.CiPipelineMaterialResponse{
				Id:              m.Id,
				GitMaterialId:   m.GitMaterialId,
				GitMaterialName: m.GitMaterial.Name[strings.Index(m.GitMaterial.Name, "-")+1:],
				Type:            string(m.Type),
				Value:           m.Value,
				Active:          m.Active,
				Url:             m.GitMaterial.Url,
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
		AppId:            ciPipeline.AppId,
		AppName:          deployDetail.AppName,
		EnvironmentId:    deployDetail.EnvironmentId,
		EnvironmentName:  deployDetail.EnvironmentName,
		LastDeployedTime: deployDetail.LastDeployedTime,
		Default:          deployDetail.Default,
		ImageTaggingData: *imageTaggingData,
		Image:            ciArtifact.Image,
	}
	return gitTriggerInfoResponse, nil
}

func (impl *CiHandlerImpl) UpdateCiWorkflowStatusFailure(timeoutForFailureCiBuild int) error {
	ciWorkflows, err := impl.ciWorkflowRepository.FindByStatusesIn([]string{Starting, Running})
	if err != nil {
		impl.Logger.Errorw("error on fetching ci workflows", "err", err)
		return err
	}
	client, err := impl.K8sUtil.GetClientForInCluster()
	if err != nil {
		impl.Logger.Errorw("error while fetching k8s client", "error", err)
		return err
	}

	for _, ciWorkflow := range ciWorkflows {
		var isExt bool
		var env *repository3.Environment
		var restConfig *rest.Config
		if ciWorkflow.Namespace != DefaultCiWorkflowNamespace {
			isExt = true
			env, err = impl.envRepository.FindById(ciWorkflow.EnvironmentId)
			if err != nil {
				impl.Logger.Errorw("could not fetch stage env", "err", err)
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
				impl.Logger.Warnw("unable to fetch ci workflow", "err", err)
				statusError, ok := err.(*errors2.StatusError)
				if ok && statusError.Status().Code == http.StatusNotFound {
					impl.Logger.Warnw("ci workflow not found", "err", err)
					isEligibleToMarkFailed = true
				} else {
					continue
					// skip this and process for next ci workflow
				}
			}

			//if ci workflow is exists, check its pod
			if !isEligibleToMarkFailed {
				ns := DefaultCiWorkflowNamespace
				if isExt {
					_, client, err = impl.k8sCommonService.GetCoreClientByClusterId(env.ClusterId)
					if err != nil {
						impl.Logger.Warnw("error in getting core v1 client using GetCoreClientByClusterId", "err", err, "clusterId", env.Cluster.Id)
						continue
					}
					ns = env.Namespace
				}
				_, err := impl.K8sUtil.GetPodByName(ns, ciWorkflow.PodName, client)
				if err != nil {
					impl.Logger.Warnw("unable to fetch ci workflow - pod", "err", err)
					statusError, ok := err.(*errors2.StatusError)
					if ok && statusError.Status().Code == http.StatusNotFound {
						impl.Logger.Warnw("pod not found", "err", err)
						isEligibleToMarkFailed = true
					} else {
						continue
						// skip this and process for next ci workflow
					}
				}
				if ciWorkflow.ExecutorType == pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM {
					if wf.Status == string(v1alpha1.WorkflowFailed) {
						isPodDeleted = true
					}
				} else {
					//check workflow status,get the status
					if wf.Status == string(v1alpha1.WorkflowFailed) && wf.Message == POD_DELETED_MESSAGE {
						isPodDeleted = true
					}
				}
			}
		}
		if isEligibleToMarkFailed {
			ciWorkflow.Status = "Failed"
			ciWorkflow.PodStatus = "Failed"
			if isPodDeleted {
				ciWorkflow.Message = executors.POD_DELETED_MESSAGE
				//error logging handled inside handlePodDeleted
				impl.handlePodDeleted(ciWorkflow)
			} else {
				ciWorkflow.Message = "marked failed by job"
			}
			err := impl.ciWorkflowRepository.UpdateWorkFlow(ciWorkflow)
			if err != nil {
				impl.Logger.Errorw("unable to update ci workflow, its eligible to mark failed", "err", err)
				// skip this and process for next ci workflow
			}
			err = impl.customTagService.DeactivateImagePathReservation(ciWorkflow.ImagePathReservationId)
			if err != nil {
				impl.Logger.Errorw("unable to update ci workflow, its eligible to mark failed", "err", err)
			}
		}
	}
	return nil
}

func (impl *CiHandlerImpl) handlePodDeleted(ciWorkflow *pipelineConfig.CiWorkflow) {
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
func (impl *CiHandlerImpl) FetchCiStatusForTriggerViewForEnvironment(request resourceGroup.ResourceGroupingRequest, token string) ([]*pipelineConfig.CiWorkflowStatus, error) {
	ciWorkflowStatuses := make([]*pipelineConfig.CiWorkflowStatus, 0)
	var cdPipelines []*pipelineConfig.Pipeline
	var err error
	if request.ResourceGroupId > 0 {
		appIds, err := impl.resourceGroupService.GetResourceIdsByResourceGroupId(request.ResourceGroupId)
		if err != nil {
			return nil, err
		}
		//override appIds if already provided app group id in request.
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
	//authorization block starts here
	var appObjectArr []string
	objects := impl.enforcerUtil.GetAppObjectByCiPipelineIds(ciPipelineIds)
	ciPipelineIds = []int{}
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object)
	}
	appResults, _ := request.CheckAuthBatch(token, appObjectArr, []string{})
	for _, ciPipeline := range ciPipelines {
		appObject := objects[ciPipeline.Id] //here only app permission have to check
		if !appResults[appObject] {
			//if user unauthorized, skip items
			continue
		}
		ciPipelineId := 0
		if ciPipeline.ParentCiPipeline == 0 {
			ciPipelineId = ciPipeline.Id
		} else {
			ciPipelineId = ciPipeline.ParentCiPipeline
		}
		ciPipelineIds = append(ciPipelineIds, ciPipelineId)
	}
	if len(ciPipelineIds) == 0 {
		return ciWorkflowStatuses, nil
	}
	ciWorkflows, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflowByCiIds(ciPipelineIds)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", "ciPipelineIds", ciPipelineIds, "err", err)
		return ciWorkflowStatuses, err
	}

	notTriggeredWorkflows := make(map[int]bool)
	latestCiWorkflows := make(map[int]*pipelineConfig.CiWorkflow)
	for _, ciWorkflow := range ciWorkflows {
		//adding only latest status in the list
		if _, ok := latestCiWorkflows[ciWorkflow.CiPipelineId]; !ok {
			latestCiWorkflows[ciWorkflow.CiPipelineId] = ciWorkflow
		}
	}
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
