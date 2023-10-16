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
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"sort"
)

type AppArtifactManager interface {
	//RetrieveArtifactsByCDPipeline : RetrieveArtifactsByCDPipeline returns all the artifacts for the cd pipeline (pre / deploy / post)
	RetrieveArtifactsByCDPipeline(pipeline *pipelineConfig.Pipeline, stage bean.WorkflowType, searchString string, isApprovalNode bool) (*bean2.CiArtifactResponse, error)

	//FetchArtifactForRollback :
	FetchArtifactForRollback(cdPipelineId, appId, offset, limit int) (bean2.CiArtifactResponse, error)

	BuildArtifactsForCdStage(pipelineId int, stageType bean.WorkflowType, ciArtifacts []bean2.CiArtifactBean, artifactMap map[int]int, parent bool, searchString string, limit int, parentCdId int) ([]bean2.CiArtifactBean, map[int]int, int, string, error)

	BuildArtifactsForParentStage(cdPipelineId int, parentId int, parentType bean.WorkflowType, ciArtifacts []bean2.CiArtifactBean, artifactMap map[int]int, searchString string, limit int, parentCdId int) ([]bean2.CiArtifactBean, error)
}

type AppArtifactManagerImpl struct {
	logger                *zap.SugaredLogger
	cdWorkflowRepository  pipelineConfig.CdWorkflowRepository
	userService           user.UserService
	imageTaggingService   ImageTaggingService
	ciArtifactRepository  repository.CiArtifactRepository
	ciWorkflowRepository  pipelineConfig.CiWorkflowRepository
	pipelineStageService    PipelineStageService
	workflowDagExecutor   WorkflowDagExecutor
	celService            resourceFilter.CELEvaluatorService
	resourceFilterService resourceFilter.ResourceFilterService

	cdPipelineConfigService CdPipelineConfigService
}

func NewAppArtifactManagerImpl(
	logger *zap.SugaredLogger,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	userService user.UserService,
	imageTaggingService ImageTaggingService,
	ciArtifactRepository repository.CiArtifactRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	workflowDagExecutor WorkflowDagExecutor,
	celService resourceFilter.CELEvaluatorService,
	resourceFilterService resourceFilter.ResourceFilterService,
	pipelineStageService PipelineStageService,
	cdPipelineConfigService CdPipelineConfigService) *AppArtifactManagerImpl {

	return &AppArtifactManagerImpl{
		logger:                  logger,
		cdWorkflowRepository:    cdWorkflowRepository,
		userService:             userService,
		imageTaggingService:     imageTaggingService,
		ciArtifactRepository:    ciArtifactRepository,
		ciWorkflowRepository:    ciWorkflowRepository,
		workflowDagExecutor:     workflowDagExecutor,
		celService:              celService,
		resourceFilterService:   resourceFilterService,
		cdPipelineConfigService: cdPipelineConfigService,
		pipelineStageService:    pipelineStageService,
	}
}

func (impl *AppArtifactManagerImpl) BuildArtifactsForParentStage(cdPipelineId int, parentId int, parentType bean.WorkflowType, ciArtifacts []bean2.CiArtifactBean, artifactMap map[int]int, searchString string, limit int, parentCdId int) ([]bean2.CiArtifactBean, error) {
	var ciArtifactsFinal []bean2.CiArtifactBean
	var err error
	if parentType == bean.CI_WORKFLOW_TYPE {
		ciArtifactsFinal, err = impl.BuildArtifactsForCIParent(cdPipelineId, parentId, parentType, ciArtifacts, artifactMap, searchString, limit)
	} else if parentType == bean.WEBHOOK_WORKFLOW_TYPE {
		ciArtifactsFinal, err = impl.BuildArtifactsForCIParent(cdPipelineId, parentId, parentType, ciArtifacts, artifactMap, searchString, limit)
	} else {
		//parent type is PRE, POST or DEPLOY type
		ciArtifactsFinal, _, _, _, err = impl.BuildArtifactsForCdStage(parentId, parentType, ciArtifacts, artifactMap, true, searchString, limit, parentCdId)
	}
	return ciArtifactsFinal, err
}

func (impl *AppArtifactManagerImpl) BuildArtifactsForCdStage(pipelineId int, stageType bean.WorkflowType, ciArtifacts []bean2.CiArtifactBean, artifactMap map[int]int, parent bool, searchString string, limit int, parentCdId int) ([]bean2.CiArtifactBean, map[int]int, int, string, error) {
	//getting running artifact id for parent cd
	parentCdRunningArtifactId := 0
	if parentCdId > 0 && parent {
		parentCdWfrList, err := impl.cdWorkflowRepository.FindArtifactByPipelineIdAndRunnerType(parentCdId, bean.CD_WORKFLOW_TYPE_DEPLOY, searchString, 1)
		if err != nil || len(parentCdWfrList) == 0 {
			impl.logger.Errorw("error in getting artifact for parent cd", "parentCdPipelineId", parentCdId)
			return ciArtifacts, artifactMap, 0, "", err
		}
		parentCdRunningArtifactId = parentCdWfrList[0].CdWorkflow.CiArtifact.Id
	}
	//getting wfr for parent and updating artifacts
	parentWfrList, err := impl.cdWorkflowRepository.FindArtifactByPipelineIdAndRunnerType(pipelineId, stageType, searchString, limit)
	if err != nil {
		impl.logger.Errorw("error in getting artifact for deployed items", "cdPipelineId", pipelineId)
		return ciArtifacts, artifactMap, 0, "", err
	}
	deploymentArtifactId := 0
	deploymentArtifactStatus := ""
	for index, wfr := range parentWfrList {
		if !parent && index == 0 {
			deploymentArtifactId = wfr.CdWorkflow.CiArtifact.Id
			deploymentArtifactStatus = wfr.Status
		}
		if wfr.Status == application.Healthy || wfr.Status == application.SUCCEEDED {
			lastSuccessfulTriggerOnParent := parent && index == 0
			latest := !parent && index == 0
			runningOnParentCd := parentCdRunningArtifactId == wfr.CdWorkflow.CiArtifact.Id
			if ciArtifactIndex, ok := artifactMap[wfr.CdWorkflow.CiArtifact.Id]; !ok {
				//entry not present, creating new entry
				mInfo, err := parseMaterialInfo([]byte(wfr.CdWorkflow.CiArtifact.MaterialInfo), wfr.CdWorkflow.CiArtifact.DataSource)
				if err != nil {
					mInfo = []byte("[]")
					impl.logger.Errorw("Error in parsing artifact material info", "err", err)
				}
				ciArtifact := bean2.CiArtifactBean{
					Id:                            wfr.CdWorkflow.CiArtifact.Id,
					Image:                         wfr.CdWorkflow.CiArtifact.Image,
					ImageDigest:                   wfr.CdWorkflow.CiArtifact.ImageDigest,
					MaterialInfo:                  mInfo,
					LastSuccessfulTriggerOnParent: lastSuccessfulTriggerOnParent,
					Latest:                        latest,
					Scanned:                       wfr.CdWorkflow.CiArtifact.Scanned,
					ScanEnabled:                   wfr.CdWorkflow.CiArtifact.ScanEnabled,
				}
				if wfr.DeploymentApprovalRequest != nil {
					ciArtifact.UserApprovalMetadata = wfr.DeploymentApprovalRequest.ConvertToApprovalMetadata()
				}
				if !parent {
					ciArtifact.Deployed = true
					ciArtifact.DeployedTime = formatDate(wfr.StartedOn, bean2.LayoutRFC3339)
				}
				if runningOnParentCd {
					ciArtifact.RunningOnParentCd = runningOnParentCd
				}
				ciArtifacts = append(ciArtifacts, ciArtifact)
				//storing index of ci artifact for using when updating old entry
				artifactMap[wfr.CdWorkflow.CiArtifact.Id] = len(ciArtifacts) - 1
			} else {
				//entry already present, updating running on parent
				if parent {
					ciArtifacts[ciArtifactIndex].LastSuccessfulTriggerOnParent = lastSuccessfulTriggerOnParent
				}
				if runningOnParentCd {
					ciArtifacts[ciArtifactIndex].RunningOnParentCd = runningOnParentCd
				}
			}
		}
	}
	return ciArtifacts, artifactMap, deploymentArtifactId, deploymentArtifactStatus, nil
}

func (impl *AppArtifactManagerImpl) BuildArtifactsForCIParent(cdPipelineId int, parentId int, parentType bean.WorkflowType, ciArtifacts []bean2.CiArtifactBean, artifactMap map[int]int, searchString string, limit int) ([]bean2.CiArtifactBean, error) {
	artifacts, err := impl.ciArtifactRepository.GetArtifactsByCDPipeline(cdPipelineId, limit, parentId, searchString, parentType)
	if err != nil {
		impl.logger.Errorw("error in getting artifacts for ci", "err", err)
		return ciArtifacts, err
	}
	for _, artifact := range artifacts {
		if _, ok := artifactMap[artifact.Id]; !ok {
			mInfo, err := parseMaterialInfo([]byte(artifact.MaterialInfo), artifact.DataSource)
			if err != nil {
				mInfo = []byte("[]")
				impl.logger.Errorw("Error in parsing artifact material info", "err", err, "artifact", artifact)
			}
			ciArtifacts = append(ciArtifacts, bean2.CiArtifactBean{
				Id:           artifact.Id,
				Image:        artifact.Image,
				ImageDigest:  artifact.ImageDigest,
				MaterialInfo: mInfo,
				ScanEnabled:  artifact.ScanEnabled,
				Scanned:      artifact.Scanned,
			})
		}
	}
	return ciArtifacts, nil
}

func (impl *AppArtifactManagerImpl) FetchArtifactForRollback(cdPipelineId, appId, offset, limit int) (bean2.CiArtifactResponse, error) {
	var deployedCiArtifacts []bean2.CiArtifactBean
	var deployedCiArtifactsResponse bean2.CiArtifactResponse
	var pipeline *pipelineConfig.Pipeline

	cdWfrs, err := impl.cdWorkflowRepository.FetchArtifactsByCdPipelineId(cdPipelineId, bean.CD_WORKFLOW_TYPE_DEPLOY, offset, limit)
	if err != nil {
		impl.logger.Errorw("error in getting artifacts for rollback by cdPipelineId", "err", err, "cdPipelineId", cdPipelineId)
		return deployedCiArtifactsResponse, err
	}
	var ids []int32
	for _, item := range cdWfrs {
		ids = append(ids, item.TriggeredBy)
		if pipeline == nil && item.CdWorkflow != nil {
			pipeline = item.CdWorkflow.Pipeline
		}
	}
	userEmails := make(map[int32]string)
	users, err := impl.userService.GetByIds(ids)
	if err != nil {
		impl.logger.Errorw("unable to fetch users by ids", "err", err, "ids", ids)
	}
	for _, item := range users {
		userEmails[item.Id] = item.EmailId
	}

	imageTagsDataMap, err := impl.imageTaggingService.GetTagsDataMapByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in getting image tagging data with appId", "err", err, "appId", appId)
		return deployedCiArtifactsResponse, err
	}
	artifactIds := make([]int, 0)

	for _, cdWfr := range cdWfrs {
		ciArtifact := &repository.CiArtifact{}
		if cdWfr.CdWorkflow != nil && cdWfr.CdWorkflow.CiArtifact != nil {
			ciArtifact = cdWfr.CdWorkflow.CiArtifact
		}
		if ciArtifact == nil {
			continue
		}
		mInfo, err := parseMaterialInfo([]byte(ciArtifact.MaterialInfo), ciArtifact.DataSource)
		if err != nil {
			mInfo = []byte("[]")
			impl.logger.Errorw("error in parsing ciArtifact material info", "err", err, "ciArtifact", ciArtifact)
		}
		userEmail := userEmails[cdWfr.TriggeredBy]
		deployedCiArtifacts = append(deployedCiArtifacts, bean2.CiArtifactBean{
			Id:           ciArtifact.Id,
			Image:        ciArtifact.Image,
			MaterialInfo: mInfo,
			DeployedTime: formatDate(cdWfr.StartedOn, bean2.LayoutRFC3339),
			WfrId:        cdWfr.Id,
			DeployedBy:   userEmail,
		})
		artifactIds = append(artifactIds, ciArtifact.Id)
	}
	imageCommentsDataMap, err := impl.imageTaggingService.GetImageCommentsDataMapByArtifactIds(artifactIds)
	if err != nil {
		impl.logger.Errorw("error in getting GetImageCommentsDataMapByArtifactIds", "err", err, "appId", appId, "artifactIds", artifactIds)
		return deployedCiArtifactsResponse, err
	}

	for i, _ := range deployedCiArtifacts {
		if imageTaggingResp := imageTagsDataMap[deployedCiArtifacts[i].Id]; imageTaggingResp != nil {
			deployedCiArtifacts[i].ImageReleaseTags = imageTaggingResp
		}
		if imageCommentResp := imageCommentsDataMap[deployedCiArtifacts[i].Id]; imageCommentResp != nil {
			deployedCiArtifacts[i].ImageComment = imageCommentResp
		}
	}

	deployedCiArtifactsResponse.CdPipelineId = cdPipelineId
	if deployedCiArtifacts == nil {
		deployedCiArtifacts = []bean2.CiArtifactBean{}
	}
	if pipeline != nil && pipeline.ApprovalNodeConfigured() {
		deployedCiArtifacts, _, err = impl.overrideArtifactsWithUserApprovalData(pipeline, deployedCiArtifacts, false, 0)
		if err != nil {
			return deployedCiArtifactsResponse, err
		}
	}
	deployedCiArtifactsResponse.CiArtifacts = deployedCiArtifacts

	return deployedCiArtifactsResponse, nil
}

func (impl *AppArtifactManagerImpl) RetrieveArtifactsByCDPipeline(pipeline *pipelineConfig.Pipeline, stage bean.WorkflowType, searchString string, isApprovalNode bool) (*bean2.CiArtifactResponse, error) {

	// retrieve parent details
	parentId, parentType, err := impl.cdPipelineConfigService.RetrieveParentDetails(pipeline.Id)
	if err != nil {
		impl.logger.Errorw("failed to retrieve parent details",
			"cdPipelineId", pipeline.Id,
			"err", err)
		return nil, err
	}

	parentCdId := 0
	if parentType == bean.CD_WORKFLOW_TYPE_POST || (parentType == bean.CD_WORKFLOW_TYPE_DEPLOY && stage != bean.CD_WORKFLOW_TYPE_POST) {
		// parentCdId is being set to store the artifact currently deployed on parent cd (if applicable).
		// Parent component is CD only if parent type is POST/DEPLOY
		parentCdId = parentId
	}

	if stage == bean.CD_WORKFLOW_TYPE_DEPLOY {
		pipelinePreStage, err := impl.pipelineStageService.GetCdStageByCdPipelineIdAndStageType(pipeline.Id, repository2.PIPELINE_STAGE_TYPE_PRE_CD)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in fetching PRE-CD stage by cd pipeline id", "pipelineId", pipeline.Id, "err", err)
			return nil, err
		}
		if (pipelinePreStage != nil && pipelinePreStage.Id != 0) || len(pipeline.PreStageConfig) > 0 {
			// Parent type will be PRE for DEPLOY stage
			parentId = pipeline.Id
			parentType = bean.CD_WORKFLOW_TYPE_PRE
		}
	}
	if stage == bean.CD_WORKFLOW_TYPE_POST {
		// Parent type will be DEPLOY for POST stage
		parentId = pipeline.Id
		parentType = bean.CD_WORKFLOW_TYPE_DEPLOY
	}

	// Build artifacts for cd stages
	var ciArtifacts []bean2.CiArtifactBean
	ciArtifactsResponse := &bean2.CiArtifactResponse{}

	artifactMap := make(map[int]int)
	limit := 10

	ciArtifacts, artifactMap, latestWfArtifactId, latestWfArtifactStatus, err := impl.
		BuildArtifactsForCdStage(pipeline.Id, stage, ciArtifacts, artifactMap, false, searchString, limit, parentCdId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting artifacts for child cd stage", "err", err, "stage", stage)
		return nil, err
	}

	ciArtifacts, err = impl.BuildArtifactsForParentStage(pipeline.Id, parentId, parentType, ciArtifacts, artifactMap, searchString, limit, parentCdId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting artifacts for cd", "err", err, "parentStage", parentType, "stage", stage)
		return nil, err
	}

	//sorting ci artifacts on the basis of creation time
	if ciArtifacts != nil {
		sort.SliceStable(ciArtifacts, func(i, j int) bool {
			return ciArtifacts[i].Id > ciArtifacts[j].Id
		})
	}

	artifactIds := make([]int, 0, len(ciArtifacts))
	for _, artifact := range ciArtifacts {
		artifactIds = append(artifactIds, artifact.Id)
	}

	artifacts, err := impl.ciArtifactRepository.GetArtifactParentCiAndWorkflowDetailsByIdsInDesc(artifactIds)
	if err != nil {
		return ciArtifactsResponse, err
	}
	imageTagsDataMap, err := impl.imageTaggingService.GetTagsDataMapByAppId(pipeline.AppId)
	if err != nil {
		impl.logger.Errorw("error in getting image tagging data with appId", "err", err, "appId", pipeline.AppId)
		return ciArtifactsResponse, err
	}

	imageCommentsDataMap, err := impl.imageTaggingService.GetImageCommentsDataMapByArtifactIds(artifactIds)
	if err != nil {
		impl.logger.Errorw("error in getting GetImageCommentsDataMapByArtifactIds", "err", err, "appId", pipeline.AppId, "artifactIds", artifactIds)
		return ciArtifactsResponse, err
	}

	for i, artifact := range artifacts {
		if imageTaggingResp := imageTagsDataMap[ciArtifacts[i].Id]; imageTaggingResp != nil {
			ciArtifacts[i].ImageReleaseTags = imageTaggingResp
		}
		if imageCommentResp := imageCommentsDataMap[ciArtifacts[i].Id]; imageCommentResp != nil {
			ciArtifacts[i].ImageComment = imageCommentResp
		}

		environment := pipeline.Environment
		scope := resourceQualifiers.Scope{AppId: pipeline.AppId, ProjectId: pipeline.App.TeamId, EnvId: pipeline.EnvironmentId, ClusterId: environment.ClusterId, IsProdEnv: environment.Default}
		params := impl.celService.GetParamsFromArtifact(ciArtifacts[i].Image)
		metadata := resourceFilter.ExpressionMetadata{
			Params: params,
		}
		filterState, err := impl.resourceFilterService.CheckForResource(scope, metadata)
		if err != nil {
			return ciArtifactsResponse, err
		}
		ciArtifacts[i].FilterState = filterState

		if artifact.ExternalCiPipelineId != 0 {
			// if external webhook continue
			continue
		}

		var ciWorkflow *pipelineConfig.CiWorkflow
		if artifact.ParentCiArtifact != 0 {
			ciWorkflow, err = impl.ciWorkflowRepository.FindLastTriggeredWorkflowGitTriggersByArtifactId(artifact.ParentCiArtifact)
			if err != nil {
				impl.logger.Errorw("error in getting ci_workflow for artifacts", "err", err, "artifact", artifact, "parentStage", parentType, "stage", stage)
				return ciArtifactsResponse, err
			}

		} else {
			ciWorkflow, err = impl.ciWorkflowRepository.FindCiWorkflowGitTriggersById(*artifact.WorkflowId)
			if err != nil {
				impl.logger.Errorw("error in getting ci_workflow for artifacts", "err", err, "artifact", artifact, "parentStage", parentType, "stage", stage)
				return ciArtifactsResponse, err
			}
		}
		ciArtifacts[i].TriggeredBy = ciWorkflow.TriggeredBy
		ciArtifacts[i].CiConfigureSourceType = ciWorkflow.GitTriggers[ciWorkflow.CiPipelineId].CiConfigureSourceType
		ciArtifacts[i].CiConfigureSourceValue = ciWorkflow.GitTriggers[ciWorkflow.CiPipelineId].CiConfigureSourceValue
	}

	ciArtifactsResponse.CdPipelineId = pipeline.Id
	ciArtifactsResponse.LatestWfArtifactId = latestWfArtifactId
	ciArtifactsResponse.LatestWfArtifactStatus = latestWfArtifactStatus
	if ciArtifacts == nil {
		ciArtifacts = []bean2.CiArtifactBean{}
	}
	ciArtifactsResponse.CiArtifacts = ciArtifacts

	if pipeline.ApprovalNodeConfigured() && stage == bean.CD_WORKFLOW_TYPE_DEPLOY { // for now, we are checking artifacts for deploy stage only
		ciArtifactsFinal, approvalConfig, err := impl.overrideArtifactsWithUserApprovalData(pipeline, ciArtifactsResponse.CiArtifacts, isApprovalNode, latestWfArtifactId)
		if err != nil {
			return ciArtifactsResponse, err
		}
		ciArtifactsResponse.UserApprovalConfig = &approvalConfig
		ciArtifactsResponse.CiArtifacts = ciArtifactsFinal
	}
	return ciArtifactsResponse, nil
}

func (impl *AppArtifactManagerImpl) overrideArtifactsWithUserApprovalData(pipeline *pipelineConfig.Pipeline, inputArtifacts []bean2.CiArtifactBean, isApprovalNode bool, latestArtifactId int) ([]bean2.CiArtifactBean, pipelineConfig.UserApprovalConfig, error) {
	impl.logger.Infow("approval node configured", "pipelineId", pipeline.Id, "isApproval", isApprovalNode)
	ciArtifactsFinal := make([]bean2.CiArtifactBean, 0, len(inputArtifacts))
	artifactIds := make([]int, 0, len(inputArtifacts))
	cdPipelineId := pipeline.Id
	approvalConfig, err := pipeline.GetApprovalConfig()
	if err != nil {
		impl.logger.Errorw("failed to unmarshal userApprovalConfig", "err", err, "cdPipelineId", cdPipelineId, "approvalConfig", approvalConfig)
		return ciArtifactsFinal, approvalConfig, err
	}

	for _, item := range inputArtifacts {
		artifactIds = append(artifactIds, item.Id)
	}

	var userApprovalMetadata map[int]*pipelineConfig.UserApprovalMetadata
	requiredApprovals := approvalConfig.RequiredCount
	userApprovalMetadata, err = impl.workflowDagExecutor.FetchApprovalDataForArtifacts(artifactIds, cdPipelineId, requiredApprovals) // it will fetch all the request data with nil cd_wfr_rnr_id
	if err != nil {
		impl.logger.Errorw("error occurred while fetching approval data for artifacts", "cdPipelineId", cdPipelineId, "artifactIds", artifactIds, "err", err)
		return ciArtifactsFinal, approvalConfig, err
	}
	for _, artifact := range inputArtifacts {
		approvalRuntimeState := pipelineConfig.InitApprovalState
		approvalMetadataForArtifact, ok := userApprovalMetadata[artifact.Id]
		if ok { // either approved or requested
			approvalRuntimeState = approvalMetadataForArtifact.ApprovalRuntimeState
			artifact.UserApprovalMetadata = approvalMetadataForArtifact
		} else if artifact.Deployed {
			approvalRuntimeState = pipelineConfig.ConsumedApprovalState
		}

		allowed := false
		if isApprovalNode { // return all the artifacts with state in init, requested or consumed
			allowed = approvalRuntimeState == pipelineConfig.InitApprovalState || approvalRuntimeState == pipelineConfig.RequestedApprovalState || approvalRuntimeState == pipelineConfig.ConsumedApprovalState
		} else { // return only approved state artifacts
			allowed = approvalRuntimeState == pipelineConfig.ApprovedApprovalState || artifact.Latest || artifact.Id == latestArtifactId
		}
		if allowed {
			ciArtifactsFinal = append(ciArtifactsFinal, artifact)
		}
	}
	return ciArtifactsFinal, approvalConfig, nil
}
