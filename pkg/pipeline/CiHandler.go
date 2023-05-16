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
	"archive/zip"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/gitSensor"
	appGroup2 "github.com/devtron-labs/devtron/pkg/appGroup"
	"github.com/devtron-labs/devtron/util/rbac"
	"io/ioutil"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/user"
	util2 "github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type CiHandler interface {
	HandleCIWebhook(gitCiTriggerRequest bean.GitCiTriggerRequest) (int, error)
	HandleCIManual(ciTriggerRequest bean.CiTriggerRequest) (int, error)

	FetchMaterialsByPipelineId(pipelineId int, showAll bool) ([]CiPipelineMaterialResponse, error)
	FetchMaterialsByPipelineIdAndGitMaterialId(pipelineId int, gitMaterialId int, showAll bool) ([]CiPipelineMaterialResponse, error)
	FetchWorkflowDetails(appId int, pipelineId int, buildId int) (WorkflowResponse, error)

	//FetchBuildById(appId int, pipelineId int) (WorkflowResponse, error)
	CancelBuild(workflowId int) (int, error)

	GetRunningWorkflowLogs(pipelineId int, workflowId int) (*bufio.Reader, func() error, error)
	GetHistoricBuildLogs(pipelineId int, workflowId int, ciWorkflow *pipelineConfig.CiWorkflow) (map[string]string, error)
	//SyncWorkflows() error

	GetBuildHistory(pipelineId int, offset int, size int) ([]WorkflowResponse, error)
	DownloadCiWorkflowArtifacts(pipelineId int, buildId int) (*os.File, error)
	UpdateWorkflow(workflowStatus v1alpha1.WorkflowStatus) (int, error)

	FetchCiStatusForTriggerView(appId int) ([]*pipelineConfig.CiWorkflowStatus, error)
	FetchCiStatusForTriggerViewV1(appId int) ([]*pipelineConfig.CiWorkflowStatus, error)
	RefreshMaterialByCiPipelineMaterialId(gitMaterialId int) (refreshRes *gitSensor.RefreshGitMaterialResponse, err error)
	FetchMaterialInfoByArtifactId(ciArtifactId int) (*GitTriggerInfoResponse, error)
	WriteToCreateTestSuites(pipelineId int, buildId int, triggeredBy int)
	UpdateCiWorkflowStatusFailure(timeoutForFailureCiBuild int) error
	FetchCiStatusForTriggerViewForEnvironment(request appGroup2.AppGroupingRequest) ([]*pipelineConfig.CiWorkflowStatus, error)
}

type CiHandlerImpl struct {
	Logger                       *zap.SugaredLogger
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
	ciService                    CiService
	gitSensorClient              gitSensor.Client
	ciWorkflowRepository         pipelineConfig.CiWorkflowRepository
	workflowService              WorkflowService
	ciLogService                 CiLogService
	ciConfig                     *CiConfig
	ciArtifactRepository         repository.CiArtifactRepository
	userService                  user.UserService
	eventClient                  client.EventClient
	eventFactory                 client.EventFactory
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
	appListingRepository         repository.AppListingRepository
	K8sUtil                      *util.K8sUtil
	cdPipelineRepository         pipelineConfig.PipelineRepository
	enforcerUtil                 rbac.EnforcerUtil
	appGroupService              appGroup2.AppGroupService
}

func NewCiHandlerImpl(Logger *zap.SugaredLogger, ciService CiService, ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	gitSensorClient gitSensor.Client, ciWorkflowRepository pipelineConfig.CiWorkflowRepository, workflowService WorkflowService,
	ciLogService CiLogService, ciConfig *CiConfig, ciArtifactRepository repository.CiArtifactRepository, userService user.UserService, eventClient client.EventClient,
	eventFactory client.EventFactory, ciPipelineRepository pipelineConfig.CiPipelineRepository, appListingRepository repository.AppListingRepository,
	K8sUtil *util.K8sUtil, cdPipelineRepository pipelineConfig.PipelineRepository, enforcerUtil rbac.EnforcerUtil,
	appGroupService appGroup2.AppGroupService) *CiHandlerImpl {
	return &CiHandlerImpl{
		Logger:                       Logger,
		ciService:                    ciService,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		gitSensorClient:              gitSensorClient,
		ciWorkflowRepository:         ciWorkflowRepository,
		workflowService:              workflowService,
		ciLogService:                 ciLogService,
		ciConfig:                     ciConfig,
		ciArtifactRepository:         ciArtifactRepository,
		userService:                  userService,
		eventClient:                  eventClient,
		eventFactory:                 eventFactory,
		ciPipelineRepository:         ciPipelineRepository,
		appListingRepository:         appListingRepository,
		K8sUtil:                      K8sUtil,
		cdPipelineRepository:         cdPipelineRepository,
		enforcerUtil:                 enforcerUtil,
		appGroupService:              appGroupService,
	}
}

type CiPipelineMaterialResponse struct {
	Id              int                    `json:"id"`
	GitMaterialId   int                    `json:"gitMaterialId"`
	GitMaterialUrl  string                 `json:"gitMaterialUrl"`
	GitMaterialName string                 `json:"gitMaterialName"`
	Type            string                 `json:"type"`
	Value           string                 `json:"value"`
	Active          bool                   `json:"active"`
	History         []*gitSensor.GitCommit `json:"history,omitempty"`
	LastFetchTime   time.Time              `json:"lastFetchTime"`
	IsRepoError     bool                   `json:"isRepoError"`
	RepoErrorMsg    string                 `json:"repoErrorMsg"`
	IsBranchError   bool                   `json:"isBranchError"`
	BranchErrorMsg  string                 `json:"branchErrorMsg"`
	Url             string                 `json:"url"`
	Regex           string                 `json:"regex"`
}

type WorkflowResponse struct {
	Id                 int                              `json:"id"`
	Name               string                           `json:"name"`
	Status             string                           `json:"status"`
	PodStatus          string                           `json:"podStatus"`
	Message            string                           `json:"message"`
	StartedOn          time.Time                        `json:"startedOn"`
	FinishedOn         time.Time                        `json:"finishedOn"`
	CiPipelineId       int                              `json:"ciPipelineId"`
	Namespace          string                           `json:"namespace"`
	LogLocation        string                           `json:"logLocation"`
	BlobStorageEnabled bool                             `json:"blobStorageEnabled"`
	GitTriggers        map[int]pipelineConfig.GitCommit `json:"gitTriggers"`
	CiMaterials        []CiPipelineMaterialResponse     `json:"ciMaterials"`
	TriggeredBy        int32                            `json:"triggeredBy"`
	Artifact           string                           `json:"artifact"`
	TriggeredByEmail   string                           `json:"triggeredByEmail"`
	Stage              string                           `json:"stage"`
	ArtifactId         int                              `json:"artifactId"`
	IsArtifactUploaded bool                             `json:"isArtifactUploaded"`
}

type GitTriggerInfoResponse struct {
	CiMaterials      []CiPipelineMaterialResponse `json:"ciMaterials"`
	TriggeredByEmail string                       `json:"triggeredByEmail"`
	LastDeployedTime string                       `json:"lastDeployedTime,omitempty"`
	AppId            int                          `json:"appId"`
	AppName          string                       `json:"appName"`
	EnvironmentId    int                          `json:"environmentId"`
	EnvironmentName  string                       `json:"environmentName"`
	Default          bool                         `json:"default,omitempty"`
}

type Trigger struct {
	PipelineId                int
	CommitHashes              map[int]bean.GitCommit
	CiMaterials               []*pipelineConfig.CiPipelineMaterial
	TriggeredBy               int32
	InvalidateCache           bool
	ExtraEnvironmentVariables map[string]string // extra env variables which will be used for CI
}

const WorkflowCancel = "CANCELLED"
const DefaultCiWorkflowNamespace = "devtron-ci"
const Running = "Running"
const Starting = "Starting"

func (impl *CiHandlerImpl) HandleCIManual(ciTriggerRequest bean.CiTriggerRequest) (int, error) {
	impl.Logger.Debugw("HandleCIManual for pipeline ", "PipelineId", ciTriggerRequest.PipelineId)
	commitHashes, extraEnvironmentVariables, err := impl.buildManualTriggerCommitHashes(ciTriggerRequest)
	if err != nil {
		return 0, err
	}
	trigger := Trigger{
		PipelineId:                ciTriggerRequest.PipelineId,
		CommitHashes:              commitHashes,
		CiMaterials:               nil,
		TriggeredBy:               ciTriggerRequest.TriggeredBy,
		InvalidateCache:           ciTriggerRequest.InvalidateCache,
		ExtraEnvironmentVariables: extraEnvironmentVariables,
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

	trigger := Trigger{
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

func (impl *CiHandlerImpl) FetchMaterialsByPipelineIdAndGitMaterialId(pipelineId int, gitMaterialId int, showAll bool) ([]CiPipelineMaterialResponse, error) {
	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineIdAndGitMaterialId(pipelineId, gitMaterialId)
	if err != nil {
		impl.Logger.Errorw("ciMaterials fetch failed", "err", err)
	}
	var ciPipelineMaterialResponses []CiPipelineMaterialResponse
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
			return []CiPipelineMaterialResponse{}, apiErr
		}
		ciMaterialHistoryMap[m] = changesResp
	}

	for k, v := range ciMaterialHistoryMap {
		r := CiPipelineMaterialResponse{
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
		return []CiPipelineMaterialResponse{}, err
	}
	for _, k := range regexMaterials {
		r := CiPipelineMaterialResponse{
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

func (impl *CiHandlerImpl) FetchMaterialsByPipelineId(pipelineId int, showAll bool) ([]CiPipelineMaterialResponse, error) {
	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(pipelineId)
	if err != nil {
		impl.Logger.Errorw("ciMaterials fetch failed", "err", err)
	}
	var ciPipelineMaterialResponses []CiPipelineMaterialResponse
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
			return []CiPipelineMaterialResponse{}, apiErr
		}
		ciMaterialHistoryMap[m] = changesResp
	}

	for k, v := range ciMaterialHistoryMap {
		r := CiPipelineMaterialResponse{
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
		return []CiPipelineMaterialResponse{}, err
	}
	for _, k := range regexMaterials {
		r := CiPipelineMaterialResponse{
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

func (impl *CiHandlerImpl) GetBuildHistory(pipelineId int, offset int, size int) ([]WorkflowResponse, error) {
	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineIdForRegexAndFixed(pipelineId)
	if err != nil {
		impl.Logger.Errorw("ciMaterials fetch failed", "err", err)
	}
	var ciPipelineMaterialResponses []CiPipelineMaterialResponse
	for _, m := range ciMaterials {
		r := CiPipelineMaterialResponse{
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

	workFlows, err := impl.ciWorkflowRepository.FindByPipelineId(pipelineId, offset, size)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", "err", err)
		return nil, err
	}
	var ciWorkLowResponses []WorkflowResponse
	for _, w := range workFlows {
		wfResponse := WorkflowResponse{
			Id:                 w.Id,
			Name:               w.Name,
			Status:             w.Status,
			PodStatus:          w.PodStatus,
			Message:            w.Message,
			StartedOn:          w.StartedOn,
			FinishedOn:         w.FinishedOn,
			CiPipelineId:       w.CiPipelineId,
			Namespace:          w.Namespace,
			LogLocation:        w.LogFilePath,
			GitTriggers:        w.GitTriggers,
			CiMaterials:        ciPipelineMaterialResponses,
			Artifact:           w.Image,
			TriggeredBy:        w.TriggeredBy,
			TriggeredByEmail:   w.EmailId,
			ArtifactId:         w.CiArtifactId,
			BlobStorageEnabled: w.BlobStorageEnabled,
			IsArtifactUploaded: w.IsArtifactUploaded,
		}
		ciWorkLowResponses = append(ciWorkLowResponses, wfResponse)
	}
	return ciWorkLowResponses, nil
}

func (impl *CiHandlerImpl) CancelBuild(workflowId int) (int, error) {
	workflow, err := impl.ciWorkflowRepository.FindById(workflowId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return 0, err
	}
	if !(string(v1alpha1.NodePending) == workflow.Status || string(v1alpha1.NodeRunning) == workflow.Status) {
		impl.Logger.Warn("cannot cancel build, build not in progress")
		return 0, errors.New("cannot cancel build, build not in progress")
	}
	runningWf, err := impl.workflowService.GetWorkflow(workflow.Name, workflow.Namespace)
	if err != nil {
		impl.Logger.Errorw("cannot find workflow ", "err", err)
		return 0, errors.New("cannot find workflow " + workflow.Name)
	}

	// Terminate workflow
	err = impl.workflowService.TerminateWorkflow(runningWf.Name, runningWf.Namespace)
	if err != nil {
		impl.Logger.Errorw("cannot terminate wf", "err", err)
		return 0, err
	}

	workflow.Status = WorkflowCancel
	err = impl.ciWorkflowRepository.UpdateWorkFlow(workflow)
	if err != nil {
		impl.Logger.Errorw("cannot update deleted workflow status, but wf deleted", "err", err)
		return 0, err
	}
	return workflow.Id, nil
}

func (impl *CiHandlerImpl) FetchWorkflowDetails(appId int, pipelineId int, buildId int) (WorkflowResponse, error) {
	workflow, err := impl.ciWorkflowRepository.FindById(buildId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return WorkflowResponse{}, err
	}
	triggeredByUser, err := impl.userService.GetById(workflow.TriggeredBy)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", "err", err)
		return WorkflowResponse{}, err
	}

	if workflow.CiPipeline.AppId != appId {
		impl.Logger.Error("pipeline does not exist for this app")
		return WorkflowResponse{}, errors.New("invalid app and pipeline combination")
	}

	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(pipelineId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return WorkflowResponse{}, err
	}

	ciArtifact, err := impl.ciArtifactRepository.GetByWfId(workflow.Id)
	if err != nil && !util.IsErrNoRows(err) {
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
	workflowResponse := WorkflowResponse{
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
		TriggeredByEmail:   triggeredByUser.EmailId,
		Artifact:           ciArtifact.Image,
		ArtifactId:         ciArtifact.Id,
		IsArtifactUploaded: ciArtifact.IsArtifactUploaded,
	}
	return workflowResponse, nil
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
	ciLogRequest := BuildLogRequest{
		PodName:   ciWorkflow.PodName,
		Namespace: ciWorkflow.Namespace,
	}
	logStream, cleanUp, err := impl.ciLogService.FetchRunningWorkflowLogs(ciLogRequest, "", "", false)
	if logStream == nil || err != nil {
		if !ciWorkflow.BlobStorageEnabled {
			return nil, nil, errors.New("logs-not-stored-in-repository")
		} else if string(v1alpha1.NodeSucceeded) == ciWorkflow.Status || string(v1alpha1.NodeError) == ciWorkflow.Status || string(v1alpha1.NodeFailed) == ciWorkflow.Status || ciWorkflow.Status == WorkflowCancel {
			impl.Logger.Errorw("err", "err", err)
			return impl.getLogsFromRepository(pipelineId, ciWorkflow)
		}
		impl.Logger.Errorw("err", "err", err)
		return nil, nil, err
	}
	logReader := bufio.NewReader(logStream)
	return logReader, cleanUp, err
}

func (impl *CiHandlerImpl) getLogsFromRepository(pipelineId int, ciWorkflow *pipelineConfig.CiWorkflow) (*bufio.Reader, func() error, error) {
	impl.Logger.Debug("getting historic logs")

	ciConfig, err := impl.ciWorkflowRepository.FindConfigByPipelineId(pipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("err", "err", err)
		return nil, nil, err
	}

	if ciConfig.LogsBucket == "" {
		ciConfig.LogsBucket = impl.ciConfig.DefaultBuildLogsBucket
	}
	if ciConfig.CiCacheRegion == "" {
		ciConfig.CiCacheRegion = impl.ciConfig.DefaultCacheBucketRegion
	}
	logsFilePath := impl.ciConfig.DefaultBuildLogsKeyPrefix + "/" + ciWorkflow.Name + "/main.log" // this is for backward compatibilty
	if strings.Contains(ciWorkflow.LogLocation, "main.log") {
		logsFilePath = ciWorkflow.LogLocation
	}
	ciLogRequest := BuildLogRequest{
		PipelineId:    ciWorkflow.CiPipelineId,
		WorkflowId:    ciWorkflow.Id,
		PodName:       ciWorkflow.PodName,
		LogsFilePath:  logsFilePath,
		CloudProvider: impl.ciConfig.CloudProvider,
		AzureBlobConfig: &blob_storage.AzureBlobBaseConfig{
			Enabled:           impl.ciConfig.CloudProvider == BLOB_STORAGE_AZURE,
			AccountName:       impl.ciConfig.AzureAccountName,
			BlobContainerName: impl.ciConfig.AzureBlobContainerCiLog,
			AccountKey:        impl.ciConfig.AzureAccountKey,
		},
		AwsS3BaseConfig: &blob_storage.AwsS3BaseConfig{
			AccessKey:         impl.ciConfig.BlobStorageS3AccessKey,
			Passkey:           impl.ciConfig.BlobStorageS3SecretKey,
			EndpointUrl:       impl.ciConfig.BlobStorageS3Endpoint,
			IsInSecure:        impl.ciConfig.BlobStorageS3EndpointInsecure,
			BucketName:        ciConfig.LogsBucket,
			Region:            ciConfig.CiCacheRegion,
			VersioningEnabled: impl.ciConfig.BlobStorageS3BucketVersioned,
		},
		GcpBlobBaseConfig: &blob_storage.GcpBlobBaseConfig{
			BucketName:             ciConfig.LogsBucket,
			CredentialFileJsonData: impl.ciConfig.BlobStorageGcpCredentialJson,
		},
	}
	oldLogsStream, cleanUp, err := impl.ciLogService.FetchLogs(impl.ciConfig.BaseLogLocationPath, ciLogRequest)
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

	if !ciWorkflow.BlobStorageEnabled {
		return nil, errors.New("logs-not-stored-in-repository")
	}

	if ciWorkflow.CiPipelineId != pipelineId {
		impl.Logger.Error("invalid request, wf not in pipeline")
		return nil, errors.New("invalid request, wf not in pipeline")
	}

	ciConfig, err := impl.ciWorkflowRepository.FindConfigByPipelineId(pipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("unable to fetch ciConfig", "err", err)
		return nil, err
	}

	if ciConfig.LogsBucket == "" {
		ciConfig.LogsBucket = impl.ciConfig.DefaultBuildLogsBucket
	}

	item := strconv.Itoa(ciWorkflow.Id)
	if ciConfig.CiCacheRegion == "" {
		ciConfig.CiCacheRegion = impl.ciConfig.DefaultCacheBucketRegion
	}
	azureBlobConfig := &blob_storage.AzureBlobBaseConfig{
		Enabled:           impl.ciConfig.CloudProvider == BLOB_STORAGE_AZURE,
		AccountName:       impl.ciConfig.AzureAccountName,
		BlobContainerName: impl.ciConfig.AzureBlobContainerCiLog,
		AccountKey:        impl.ciConfig.AzureAccountKey,
	}
	awsS3BaseConfig := &blob_storage.AwsS3BaseConfig{
		AccessKey:         impl.ciConfig.BlobStorageS3AccessKey,
		Passkey:           impl.ciConfig.BlobStorageS3SecretKey,
		EndpointUrl:       impl.ciConfig.BlobStorageS3Endpoint,
		IsInSecure:        impl.ciConfig.BlobStorageS3EndpointInsecure,
		BucketName:        ciConfig.LogsBucket,
		Region:            ciConfig.CiCacheRegion,
		VersioningEnabled: impl.ciConfig.BlobStorageS3BucketVersioned,
	}
	gcpBlobBaseConfig := &blob_storage.GcpBlobBaseConfig{
		BucketName:             ciConfig.LogsBucket,
		CredentialFileJsonData: impl.ciConfig.BlobStorageGcpCredentialJson,
	}

	key := fmt.Sprintf("%s/"+impl.ciConfig.CiArtifactLocationFormat, impl.ciConfig.DefaultArtifactKeyPrefix, ciWorkflow.Id, ciWorkflow.Id)

	baseLogLocationPathConfig := impl.ciConfig.BaseLogLocationPath
	blobStorageService := blob_storage.NewBlobStorageServiceImpl(nil)
	destinationKey := filepath.Clean(filepath.Join(baseLogLocationPathConfig, item))
	request := &blob_storage.BlobStorageRequest{
		StorageType:         impl.ciConfig.CloudProvider,
		SourceKey:           key,
		DestinationKey:      baseLogLocationPathConfig + item,
		AzureBlobBaseConfig: azureBlobConfig,
		AwsS3BaseConfig:     awsS3BaseConfig,
		GcpBlobBaseConfig:   gcpBlobBaseConfig,
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
		ciConfig.LogsBucket = impl.ciConfig.DefaultBuildLogsBucket
	}
	ciLogRequest := BuildLogRequest{
		PipelineId:    ciWorkflow.CiPipelineId,
		WorkflowId:    ciWorkflow.Id,
		PodName:       ciWorkflow.PodName,
		LogsFilePath:  ciWorkflow.LogLocation,
		CloudProvider: impl.ciConfig.CloudProvider,
		AzureBlobConfig: &blob_storage.AzureBlobBaseConfig{
			Enabled:           impl.ciConfig.CloudProvider == BLOB_STORAGE_AZURE,
			AccountName:       impl.ciConfig.AzureAccountName,
			BlobContainerName: impl.ciConfig.AzureBlobContainerCiLog,
			AccountKey:        impl.ciConfig.AzureAccountKey,
		},
		AwsS3BaseConfig: &blob_storage.AwsS3BaseConfig{
			AccessKey:         impl.ciConfig.BlobStorageS3AccessKey,
			Passkey:           impl.ciConfig.BlobStorageS3SecretKey,
			EndpointUrl:       impl.ciConfig.BlobStorageS3Endpoint,
			IsInSecure:        impl.ciConfig.BlobStorageS3EndpointInsecure,
			BucketName:        ciConfig.LogsBucket,
			Region:            ciConfig.CiCacheRegion,
			VersioningEnabled: impl.ciConfig.BlobStorageS3BucketVersioned,
		},
		GcpBlobBaseConfig: &blob_storage.GcpBlobBaseConfig{
			BucketName:             ciConfig.LogsBucket,
			CredentialFileJsonData: impl.ciConfig.BlobStorageGcpCredentialJson,
		},
	}
	logsFile, cleanUp, err := impl.ciLogService.FetchLogs(impl.ciConfig.BaseLogLocationPath, ciLogRequest)
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

func (impl *CiHandlerImpl) extractWorkfowStatus(workflowStatus v1alpha1.WorkflowStatus) (string, string, string, string, string, string) {
	workflowName := ""
	status := string(workflowStatus.Phase)
	podStatus := ""
	message := ""
	podName := ""
	logLocation := ""
	for k, v := range workflowStatus.Nodes {
		if v.TemplateName == CI_WORKFLOW_NAME {
			impl.Logger.Infow("extractWorkflowStatus", "workflowName", k, "v", v)
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

func (impl *CiHandlerImpl) UpdateWorkflow(workflowStatus v1alpha1.WorkflowStatus) (int, error) {
	workflowName, status, podStatus, message, logLocation, podName := impl.extractWorkfowStatus(workflowStatus)
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
		ciArtifactLocationFormat = impl.ciConfig.CiArtifactLocationFormat
	}
	ciArtifactLocation := fmt.Sprintf(ciArtifactLocationFormat, ciWorkflowConfig.LogsBucket, savedWorkflow.Id, savedWorkflow.Id)

	if impl.stateChanged(status, podStatus, message, workflowStatus.FinishedAt.Time, savedWorkflow) {
		if savedWorkflow.Status != WorkflowCancel {
			savedWorkflow.Status = status
		}
		savedWorkflow.PodStatus = podStatus
		savedWorkflow.Message = message
		savedWorkflow.FinishedOn = workflowStatus.FinishedAt.Time
		savedWorkflow.Name = workflowName
		//savedWorkflow.LogLocation = "/ci-pipeline/" + strconv.Itoa(savedWorkflow.CiPipelineId) + "/workflow/" + strconv.Itoa(savedWorkflow.Id) + "/logs" //TODO need to fetch from workflow object
		savedWorkflow.LogLocation = logLocation
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

			impl.WriteToCreateTestSuites(savedWorkflow.CiPipelineId, workflowId, int(savedWorkflow.TriggeredBy))
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

func (impl *CiHandlerImpl) buildAutomaticTriggerCommitHashes(ciMaterials []*pipelineConfig.CiPipelineMaterial, request bean.GitCiTriggerRequest) (map[int]bean.GitCommit, error) {
	commitHashes := map[int]bean.GitCommit{}
	for _, ciMaterial := range ciMaterials {
		if ciMaterial.Id == request.CiPipelineMaterial.Id || len(ciMaterials) == 1 {
			request.CiPipelineMaterial.GitCommit = SetGitCommitValuesForBuildingCommitHash(ciMaterial, request.CiPipelineMaterial.GitCommit)
			commitHashes[ciMaterial.Id] = request.CiPipelineMaterial.GitCommit
		} else {
			// this is possible in case of non Webhook, as there would be only one pipeline material per git material in case of PR
			lastCommit, err := impl.getLastSeenCommit(ciMaterial.Id)
			if err != nil {
				return map[int]bean.GitCommit{}, err
			}
			lastCommit = SetGitCommitValuesForBuildingCommitHash(ciMaterial, lastCommit)
			commitHashes[ciMaterial.Id] = lastCommit
		}
	}
	return commitHashes, nil
}

func SetGitCommitValuesForBuildingCommitHash(ciMaterial *pipelineConfig.CiPipelineMaterial, oldGitCommit bean.GitCommit) bean.GitCommit {
	newGitCommit := oldGitCommit
	newGitCommit.CiConfigureSourceType = ciMaterial.Type
	newGitCommit.CiConfigureSourceValue = ciMaterial.Value
	newGitCommit.GitRepoUrl = ciMaterial.GitMaterial.Url
	newGitCommit.GitRepoName = ciMaterial.GitMaterial.Name[strings.Index(ciMaterial.GitMaterial.Name, "-")+1:]
	return newGitCommit
}

func (impl *CiHandlerImpl) buildManualTriggerCommitHashes(ciTriggerRequest bean.CiTriggerRequest) (map[int]bean.GitCommit, map[string]string, error) {
	commitHashes := map[int]bean.GitCommit{}
	extraEnvironmentVariables := make(map[string]string)
	for _, ciPipelineMaterial := range ciTriggerRequest.CiPipelineMaterial {

		pipeLineMaterialFromDb, err := impl.ciPipelineMaterialRepository.GetById(ciPipelineMaterial.Id)
		if err != nil {
			impl.Logger.Errorw("err in fetching pipeline material by id", "err", err)
			return map[int]bean.GitCommit{}, nil, err
		}

		pipelineType := pipeLineMaterialFromDb.Type
		if pipelineType == pipelineConfig.SOURCE_TYPE_BRANCH_FIXED {
			gitCommit, err := impl.BuildManualTriggerCommitHashesForSourceTypeBranchFix(ciPipelineMaterial, pipeLineMaterialFromDb)
			if err != nil {
				impl.Logger.Errorw("err", "err", err)
				return map[int]bean.GitCommit{}, nil, err
			}
			commitHashes[ciPipelineMaterial.Id] = gitCommit

		} else if pipelineType == pipelineConfig.SOURCE_TYPE_WEBHOOK {
			gitCommit, extraEnvVariables, err := impl.BuildManualTriggerCommitHashesForSourceTypeWebhook(ciPipelineMaterial, pipeLineMaterialFromDb)
			if err != nil {
				impl.Logger.Errorw("err", "err", err)
				return map[int]bean.GitCommit{}, nil, err
			}
			commitHashes[ciPipelineMaterial.Id] = gitCommit
			extraEnvironmentVariables = extraEnvVariables
		}

	}
	return commitHashes, extraEnvironmentVariables, nil
}

func (impl *CiHandlerImpl) BuildManualTriggerCommitHashesForSourceTypeBranchFix(ciPipelineMaterial bean.CiPipelineMaterial, pipeLineMaterialFromDb *pipelineConfig.CiPipelineMaterial) (bean.GitCommit, error) {
	commitMetadataRequest := &gitSensor.CommitMetadataRequest{
		PipelineMaterialId: ciPipelineMaterial.Id,
		GitHash:            ciPipelineMaterial.GitCommit.Commit,
		GitTag:             ciPipelineMaterial.GitTag,
	}
	gitCommitResponse, err := impl.gitSensorClient.GetCommitMetadataForPipelineMaterial(context.Background(), commitMetadataRequest)
	if err != nil {
		impl.Logger.Errorw("err in fetching commit metadata", "commitMetadataRequest", commitMetadataRequest, "err", err)
		return bean.GitCommit{}, err
	}
	if gitCommitResponse == nil {
		return bean.GitCommit{}, errors.New("commit not found")
	}

	gitCommit := bean.GitCommit{
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

func (impl *CiHandlerImpl) BuildManualTriggerCommitHashesForSourceTypeWebhook(ciPipelineMaterial bean.CiPipelineMaterial, pipeLineMaterialFromDb *pipelineConfig.CiPipelineMaterial) (bean.GitCommit, map[string]string, error) {
	webhookDataInput := ciPipelineMaterial.GitCommit.WebhookData

	// fetch webhook data on the basis of Id
	webhookDataRequest := &gitSensor.WebhookDataRequest{
		Id:                   webhookDataInput.Id,
		CiPipelineMaterialId: ciPipelineMaterial.Id,
	}

	webhookAndCiData, err := impl.gitSensorClient.GetWebhookData(context.Background(), webhookDataRequest)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return bean.GitCommit{}, nil, err
	}
	webhookData := webhookAndCiData.WebhookData

	// if webhook event is of merged type, then fetch latest commit for target branch
	if webhookData.EventActionType == bean.WEBHOOK_EVENT_MERGED_ACTION_TYPE {

		// get target branch name from webhook
		targetBranchName := webhookData.Data[bean.WEBHOOK_SELECTOR_TARGET_BRANCH_NAME_NAME]
		if targetBranchName == "" {
			impl.Logger.Error("target branch not found from webhook data")
			return bean.GitCommit{}, nil, err
		}

		// get latest commit hash for target branch
		latestCommitMetadataRequest := &gitSensor.CommitMetadataRequest{
			PipelineMaterialId: ciPipelineMaterial.Id,
			BranchName:         targetBranchName,
		}

		latestCommit, err := impl.gitSensorClient.GetCommitMetadata(context.Background(), latestCommitMetadataRequest)

		if err != nil {
			impl.Logger.Errorw("err", "err", err)
			return bean.GitCommit{}, nil, err
		}

		// update webhookData (local) with target latest hash
		webhookData.Data[bean.WEBHOOK_SELECTOR_TARGET_CHECKOUT_NAME] = latestCommit.Commit

	}

	// build git commit
	gitCommit := bean.GitCommit{
		GitRepoName:            pipeLineMaterialFromDb.GitMaterial.Name[strings.Index(pipeLineMaterialFromDb.GitMaterial.Name, "-")+1:],
		GitRepoUrl:             pipeLineMaterialFromDb.GitMaterial.Url,
		CiConfigureSourceValue: pipeLineMaterialFromDb.Value,
		CiConfigureSourceType:  pipeLineMaterialFromDb.Type,
		WebhookData: &bean.WebhookData{
			Id:              int(webhookData.Id),
			EventActionType: webhookData.EventActionType,
			Data:            webhookData.Data,
		},
	}

	return gitCommit, webhookAndCiData.ExtraEnvironmentVariables, nil
}

func (impl *CiHandlerImpl) getLastSeenCommit(ciMaterialId int) (bean.GitCommit, error) {
	var materialIds []int
	materialIds = append(materialIds, ciMaterialId)
	headReq := &gitSensor.HeadRequest{
		MaterialIds: materialIds,
	}
	res, err := impl.gitSensorClient.GetHeadForPipelineMaterials(context.Background(), headReq)
	if err != nil {
		return bean.GitCommit{}, err
	}
	if len(res) == 0 {
		return bean.GitCommit{}, errors.New("received empty response")
	}
	gitCommit := bean.GitCommit{
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

func (impl *CiHandlerImpl) FetchMaterialInfoByArtifactId(ciArtifactId int) (*GitTriggerInfoResponse, error) {

	ciArtifact, err := impl.ciArtifactRepository.Get(ciArtifactId)
	if err != nil {
		impl.Logger.Errorw("err", "ciArtifactId", ciArtifactId, "err", err)
		return &GitTriggerInfoResponse{}, err
	}

	ciPipeline, err := impl.ciPipelineRepository.FindById(ciArtifact.PipelineId)
	if err != nil {
		impl.Logger.Errorw("err", "ciArtifactId", ciArtifactId, "err", err)
		return &GitTriggerInfoResponse{}, err
	}

	ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(ciPipeline.Id)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return &GitTriggerInfoResponse{}, err
	}

	deployDetail, err := impl.appListingRepository.DeploymentDetailByArtifactId(ciArtifactId)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return &GitTriggerInfoResponse{}, err
	}

	ciMaterialsArr := make([]CiPipelineMaterialResponse, 0)
	triggeredByUser := &bean2.UserInfo{}
	//check workflow data only for non external builds
	if !ciPipeline.IsExternal {
		var workflow *pipelineConfig.CiWorkflow
		if ciArtifact.ParentCiArtifact > 0 {
			workflow, err = impl.ciWorkflowRepository.FindLastTriggeredWorkflowByArtifactId(ciArtifact.ParentCiArtifact)
			if err != nil {
				impl.Logger.Errorw("err", "ciArtifactId", ciArtifact.ParentCiArtifact, "err", err)
				return &GitTriggerInfoResponse{}, err
			}
		} else {
			workflow, err = impl.ciWorkflowRepository.FindLastTriggeredWorkflowByArtifactId(ciArtifactId)
			if err != nil {
				impl.Logger.Errorw("err", "ciArtifactId", ciArtifactId, "err", err)
				return &GitTriggerInfoResponse{}, err
			}
		}

		triggeredByUser, err = impl.userService.GetById(workflow.TriggeredBy)
		if err != nil && !util.IsErrNoRows(err) {
			impl.Logger.Errorw("err", "err", err)
			return &GitTriggerInfoResponse{}, err
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

			res := CiPipelineMaterialResponse{
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
	gitTriggerInfoResponse := &GitTriggerInfoResponse{
		//GitTriggers:      workflow.GitTriggers,
		CiMaterials:      ciMaterialsArr,
		TriggeredByEmail: triggeredByUser.EmailId,
		AppId:            ciPipeline.AppId,
		AppName:          deployDetail.AppName,
		EnvironmentId:    deployDetail.EnvironmentId,
		EnvironmentName:  deployDetail.EnvironmentName,
		LastDeployedTime: deployDetail.LastDeployedTime,
		Default:          deployDetail.Default,
	}
	return gitTriggerInfoResponse, nil
}

func (impl *CiHandlerImpl) WriteToCreateTestSuites(pipelineId int, buildId int, triggeredBy int) {
	testReportFile, err := impl.DownloadCiWorkflowArtifacts(pipelineId, buildId)
	if err != nil {
		impl.Logger.Errorw("WriteTestSuite, error in fetching report file from s3", "err", err, "pipelineId", pipelineId, "buildId", buildId)
		return
	}
	if testReportFile == nil {
		return
	}
	read, err := zip.OpenReader(testReportFile.Name())
	if err != nil {
		impl.Logger.Errorw("WriteTestSuite, error while open reader", "name", testReportFile.Name())
		return
	}
	defer read.Close()
	const CreatedBy = "created_by"
	const TriggerId = "trigger_id"
	const CiPipelineId = "ci_pipeline_id"
	const XML = "xml"
	payload := make(map[string]interface{})
	var reports []string
	payload[CreatedBy] = triggeredBy
	payload[TriggerId] = buildId
	payload[CiPipelineId] = pipelineId
	payload[XML] = reports
	for _, file := range read.File {
		if payload, err = impl.listFiles(file, payload); err != nil {
			impl.Logger.Errorw("WriteTestSuite, failed to read from zip", "file", file.Name, "error", err)
			return
		}
	}
	b, err := json.Marshal(payload)
	if err != nil {
		impl.Logger.Errorw("WriteTestSuite, payload marshal error", "error", err)
		return
	}
	impl.Logger.Debugw("WriteTestSuite, sending to create", "TriggerId", buildId)
	_, err = impl.eventClient.SendTestSuite(b)
	if err != nil {
		impl.Logger.Errorw("WriteTestSuite, error while making test suit post request", "err", err)
		return
	}
}

func (impl *CiHandlerImpl) listFiles(file *zip.File, payload map[string]interface{}) (map[string]interface{}, error) {
	fileRead, err := file.Open()
	if err != nil {
		return payload, err
	}
	defer fileRead.Close()

	if strings.Contains(file.Name, ".xml") {
		content, err := ioutil.ReadAll(fileRead)
		if err != nil {
			impl.Logger.Errorw("panic error", "err", err)
			return payload, err
		}
		var reports []string
		if _, ok := payload["xml"]; !ok {
			reports = append(reports, string([]byte(content)))
			payload["xml"] = reports
		} else {
			reports = payload["xml"].([]string)
			reports = append(reports, string([]byte(content)))
			payload["xml"] = reports
		}
	}
	return payload, nil
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
		isEligibleToMarkFailed := false
		if time.Since(ciWorkflow.StartedOn) > (time.Minute * time.Duration(timeoutForFailureCiBuild)) {
			//check weather pod is exists or not, if exits check its status
			_, err := impl.workflowService.GetWorkflow(ciWorkflow.Name, DefaultCiWorkflowNamespace)
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
				_, err = impl.K8sUtil.GetPodByName(DefaultCiWorkflowNamespace, ciWorkflow.PodName, client)
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
			}
		}
		if isEligibleToMarkFailed {
			ciWorkflow.Status = "Failed"
			ciWorkflow.PodStatus = "Failed"
			ciWorkflow.Message = "marked failed by job"
			err := impl.ciWorkflowRepository.UpdateWorkFlow(ciWorkflow)
			if err != nil {
				impl.Logger.Errorw("unable to update ci workflow, its eligible to mark failed", "err", err)
				continue
				// skip this and process for next ci workflow
			}
		}
	}
	return nil
}

func (impl *CiHandlerImpl) FetchCiStatusForTriggerViewForEnvironment(request appGroup2.AppGroupingRequest) ([]*pipelineConfig.CiWorkflowStatus, error) {
	ciWorkflowStatuses := make([]*pipelineConfig.CiWorkflowStatus, 0)
	var cdPipelines []*pipelineConfig.Pipeline
	var err error
	if request.AppGroupId > 0 {
		appIds, err := impl.appGroupService.GetAppIdsByAppGroupId(request.AppGroupId)
		if err != nil {
			return nil, err
		}
		//override appIds if already provided app group id in request.
		request.AppIds = appIds
	}
	if len(request.AppIds) > 0 {
		cdPipelines, err = impl.cdPipelineRepository.FindActiveByInFilter(request.EnvId, request.AppIds)
	} else {
		cdPipelines, err = impl.cdPipelineRepository.FindActiveByEnvId(request.EnvId)
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
	appResults, _ := request.CheckAuthBatch(request.EmailId, appObjectArr, []string{})
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
