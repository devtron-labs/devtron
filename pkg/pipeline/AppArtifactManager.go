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
	repository2 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"sort"
)

type AppArtifactManager interface {
	//RetrieveArtifactsByCDPipeline : RetrieveArtifactsByCDPipeline returns all the artifacts for the cd pipeline (pre / deploy / post)
	RetrieveArtifactsByCDPipeline(pipeline *pipelineConfig.Pipeline, stage bean.WorkflowType, searchString string, count int, isApprovalNode bool) (*bean2.CiArtifactResponse, error)

	RetrieveArtifactsByCDPipelineV2(pipeline *pipelineConfig.Pipeline, stage bean.WorkflowType, artifactListingFilterOpts *bean.ArtifactsListFilterOptions, isApprovalNode bool) (*bean2.CiArtifactResponse, error)

	FetchApprovalPendingArtifacts(pipeline *pipelineConfig.Pipeline, artifactListingFilterOpts *bean.ArtifactsListFilterOptions) (*bean2.CiArtifactResponse, error)

	//FetchArtifactForRollback :
	FetchArtifactForRollback(cdPipelineId, appId, offset, limit int, searchString string, app *bean2.CreateAppDTO, pipeline *pipelineConfig.Pipeline) (bean2.CiArtifactResponse, error)

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
	pipelineStageService  PipelineStageService
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

func (impl *AppArtifactManagerImpl) FetchArtifactForRollback(cdPipelineId, appId, offset, limit int, searchString string, app *bean2.CreateAppDTO, deploymentPipeline *pipelineConfig.Pipeline) (bean2.CiArtifactResponse, error) {
	var deployedCiArtifacts []bean2.CiArtifactBean
	var deployedCiArtifactsResponse bean2.CiArtifactResponse
	var pipeline *pipelineConfig.Pipeline

	cdWfrs, err := impl.cdWorkflowRepository.FetchArtifactsByCdPipelineId(cdPipelineId, bean.CD_WORKFLOW_TYPE_DEPLOY, offset, limit, searchString)
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

	scope := resourceQualifiers.Scope{AppId: app.Id, EnvId: deploymentPipeline.EnvironmentId, ClusterId: deploymentPipeline.Environment.ClusterId, ProjectId: app.TeamId, IsProdEnv: deploymentPipeline.Environment.Default}
	impl.logger.Infow("scope for rollback deployment ", "scope", scope)
	filters, err := impl.resourceFilterService.GetFiltersByScope(scope)
	if err != nil {
		impl.logger.Errorw("error in getting resource filters for the pipeline", "pipelineId", pipeline.Id, "err", err)
		return deployedCiArtifactsResponse, err
	}

	for i, _ := range deployedCiArtifacts {
		imageTaggingResp := imageTagsDataMap[deployedCiArtifacts[i].Id]
		if imageTaggingResp != nil {
			deployedCiArtifacts[i].ImageReleaseTags = imageTaggingResp
		}
		if imageCommentResp := imageCommentsDataMap[deployedCiArtifacts[i].Id]; imageCommentResp != nil {
			deployedCiArtifacts[i].ImageComment = imageCommentResp
			releaseTags := make([]string, 0, len(imageTaggingResp))
			for _, imageTag := range imageTaggingResp {
				if !imageTag.Deleted {
					releaseTags = append(releaseTags, imageTag.TagName)
				}
			}

			filterState, _, err := impl.resourceFilterService.CheckForResource(filters, deployedCiArtifacts[i].Image, releaseTags)
			if err != nil {
				return deployedCiArtifactsResponse, err
			}
			deployedCiArtifacts[i].FilterState = filterState
		}
		deployedCiArtifactsResponse.ResourceFilters = filters
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

func (impl *AppArtifactManagerImpl) RetrieveArtifactsByCDPipeline(pipeline *pipelineConfig.Pipeline, stage bean.WorkflowType, searchString string, count int, isApprovalNode bool) (*bean2.CiArtifactResponse, error) {

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
	limit := count

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

	environment := pipeline.Environment
	scope := resourceQualifiers.Scope{AppId: pipeline.AppId, ProjectId: pipeline.App.TeamId, EnvId: pipeline.EnvironmentId, ClusterId: environment.ClusterId, IsProdEnv: environment.Default}
	filters, err := impl.resourceFilterService.GetFiltersByScope(scope)
	if err != nil {
		impl.logger.Errorw("error in getting resource filters for the pipeline", "pipelineId", pipeline.Id, "err", err)
		return ciArtifactsResponse, err
	}

	for i, artifact := range artifacts {
		imageTaggingResp := imageTagsDataMap[ciArtifacts[i].Id]
		if imageTaggingResp != nil {
			ciArtifacts[i].ImageReleaseTags = imageTaggingResp
		}
		if imageCommentResp := imageCommentsDataMap[ciArtifacts[i].Id]; imageCommentResp != nil {
			ciArtifacts[i].ImageComment = imageCommentResp
		}

		releaseTags := make([]string, 0, len(imageTaggingResp))
		for _, imageTag := range imageTaggingResp {
			if !imageTag.Deleted {
				releaseTags = append(releaseTags, imageTag.TagName)
			}
		}
		filterState, _, err := impl.resourceFilterService.CheckForResource(filters, ciArtifacts[i].Image, releaseTags)
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
	ciArtifactsResponse.ResourceFilters = filters
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

func (impl *AppArtifactManagerImpl) FetchApprovalPendingArtifacts(pipeline *pipelineConfig.Pipeline, artifactListingFilterOpts *bean.ArtifactsListFilterOptions) (*bean2.CiArtifactResponse, error) {
	ciArtifactsResponse := &bean2.CiArtifactResponse{}

	if pipeline.ApprovalNodeConfigured() { // for now, we are checking artifacts for deploy stage only
		ciArtifacts, err := impl.workflowDagExecutor.FetchApprovalRequestArtifacts(pipeline.Id, artifactListingFilterOpts.Limit, artifactListingFilterOpts.Offset, artifactListingFilterOpts.SearchString)
		if err != nil {
			impl.logger.Errorw("failed to fetch approval request artifacts", "err", err, "cdPipelineId", pipeline.Id)
			return ciArtifactsResponse, err
		}

		approvalConfig, err := pipeline.GetApprovalConfig()
		if err != nil {
			impl.logger.Errorw("failed to unmarshal userApprovalConfig", "err", err, "cdPipelineId", pipeline.Id, "approvalConfig", approvalConfig)
			return ciArtifactsResponse, err
		}

		var userApprovalMetadata map[int]*pipelineConfig.UserApprovalMetadata
		requiredApprovals := approvalConfig.RequiredCount
		var artifactIds []int
		for _, item := range ciArtifacts {
			artifactIds = append(artifactIds, item.Id)
		}

		userApprovalMetadata, err = impl.workflowDagExecutor.FetchApprovalDataForArtifacts(artifactIds, pipeline.Id, requiredApprovals) // it will fetch all the request data with nil cd_wfr_rnr_id
		if err != nil {
			impl.logger.Errorw("error occurred while fetching approval data for artifacts", "cdPipelineId", pipeline.Id, "artifactIds", artifactIds, "err", err)
			return ciArtifactsResponse, err
		}

		for _, artifact := range ciArtifacts {
			if approvalMetadataForArtifact, ok := userApprovalMetadata[artifact.Id]; ok { // either approved or requested
				artifact.UserApprovalMetadata = approvalMetadataForArtifact
			}
		}
		ciArtifactsResponse.CdPipelineId = pipeline.Id
		ciArtifactsResponse.CiArtifacts = ciArtifacts
		ciArtifactsResponse.UserApprovalConfig = &approvalConfig
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

func (impl *AppArtifactManagerImpl) extractParentMetaDataByPipeline(pipeline *pipelineConfig.Pipeline, stage bean.WorkflowType) (parentId int, parentType bean.WorkflowType, parentCdId int, err error) {
	// retrieve parent details
	parentId, parentType, err = impl.cdPipelineConfigService.RetrieveParentDetails(pipeline.Id)
	if err != nil {
		impl.logger.Errorw("failed to retrieve parent details",
			"cdPipelineId", pipeline.Id,
			"err", err)
		return parentId, parentType, parentCdId, err
	}

	if parentType == bean.CD_WORKFLOW_TYPE_POST || (parentType == bean.CD_WORKFLOW_TYPE_DEPLOY && stage != bean.CD_WORKFLOW_TYPE_POST) {
		// parentCdId is being set to store the artifact currently deployed on parent cd (if applicable).
		// Parent component is CD only if parent type is POST/DEPLOY
		parentCdId = parentId
	}

	if stage == bean.CD_WORKFLOW_TYPE_DEPLOY {
		pipelinePreStage, err := impl.pipelineStageService.GetCdStageByCdPipelineIdAndStageType(pipeline.Id, repository2.PIPELINE_STAGE_TYPE_PRE_CD)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in fetching PRE-CD stage by cd pipeline id", "pipelineId", pipeline.Id, "err", err)
			return parentId, parentType, parentCdId, err
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
	return parentId, parentType, parentCdId, err
}

func (impl *AppArtifactManagerImpl) RetrieveArtifactsByCDPipelineV2(pipeline *pipelineConfig.Pipeline, stage bean.WorkflowType, artifactListingFilterOpts *bean.ArtifactsListFilterOptions, isApprovalNode bool) (*bean2.CiArtifactResponse, error) {

	// retrieve parent details
	parentId, parentType, parentCdId, err := impl.extractParentMetaDataByPipeline(pipeline, stage)
	if err != nil {
		impl.logger.Errorw("error in finding parent meta data for pipeline", "pipelineId", pipeline.Id, "pipelineStage", stage, "err", err)
		return nil, err
	}
	// Build artifacts for cd stages
	var ciArtifacts []bean2.CiArtifactBean
	ciArtifactsResponse := &bean2.CiArtifactResponse{}

	artifactListingFilterOpts.PipelineId = pipeline.Id
	artifactListingFilterOpts.ParentId = parentId
	artifactListingFilterOpts.ParentCdId = parentCdId
	artifactListingFilterOpts.ParentStageType = parentType
	artifactListingFilterOpts.StageType = stage
	artifactListingFilterOpts.ApprovalNodeConfigured = pipeline.ApprovalNodeConfigured()
	artifactListingFilterOpts.SearchString = "%" + artifactListingFilterOpts.SearchString + "%"

	if artifactListingFilterOpts.ApprovalNodeConfigured {
		approvalConfig, err := pipeline.GetApprovalConfig()
		if err != nil {
			impl.logger.Errorw("failed to unmarshal userApprovalConfig", "err", err, "cdPipelineId", pipeline.Id, "approvalConfig", approvalConfig)
			return ciArtifactsResponse, err
		}
		artifactListingFilterOpts.ApproversCount = approvalConfig.RequiredCount
		ciArtifactsResponse.UserApprovalConfig = &approvalConfig
	}

	ciArtifactsRefs, latestWfArtifactId, latestWfArtifactStatus, err := impl.BuildArtifactsList(artifactListingFilterOpts, isApprovalNode)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting artifacts for child cd stage", "err", err, "stage", stage)
		return nil, err
	}

	for _, ciArtifactsRef := range ciArtifactsRefs {
		ciArtifacts = append(ciArtifacts, *ciArtifactsRef)
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

	environment := pipeline.Environment
	scope := resourceQualifiers.Scope{AppId: pipeline.AppId, ProjectId: pipeline.App.TeamId, EnvId: pipeline.EnvironmentId, ClusterId: environment.ClusterId, IsProdEnv: environment.Default}
	filters, err := impl.resourceFilterService.GetFiltersByScope(scope)
	if err != nil {
		impl.logger.Errorw("error in getting resource filters for the pipeline", "pipelineId", pipeline.Id, "err", err)
		return ciArtifactsResponse, err
	}

	for i, artifact := range ciArtifacts {
		imageTaggingResp := imageTagsDataMap[ciArtifacts[i].Id]
		if imageTaggingResp != nil {
			ciArtifacts[i].ImageReleaseTags = imageTaggingResp
		}

		if imageCommentResp := imageCommentsDataMap[ciArtifacts[i].Id]; imageCommentResp != nil {
			ciArtifacts[i].ImageComment = imageCommentResp
		}

		releaseTags := make([]string, 0, len(imageTaggingResp))
		for _, imageTag := range imageTaggingResp {
			if !imageTag.Deleted {
				releaseTags = append(releaseTags, imageTag.TagName)
			}
		}
		filterState, _, err := impl.resourceFilterService.CheckForResource(filters, ciArtifacts[i].Image, releaseTags)
		ciArtifacts[i].FilterState = filterState

		if artifact.ExternalCiPipelineId != 0 {
			// if external webhook continue
			continue
		}

		//TODO: can be optimised
		var ciWorkflow *pipelineConfig.CiWorkflow
		if artifact.ParentCiArtifact != 0 {
			ciWorkflow, err = impl.ciWorkflowRepository.FindLastTriggeredWorkflowGitTriggersByArtifactId(artifact.ParentCiArtifact)
			if err != nil {
				impl.logger.Errorw("error in getting ci_workflow for artifacts", "err", err, "artifact", artifact, "parentStage", parentType, "stage", stage)
				return ciArtifactsResponse, err
			}

		} else {
			ciWorkflow, err = impl.ciWorkflowRepository.FindCiWorkflowGitTriggersById(artifact.CiWorkflowId)
			if err != nil {
				impl.logger.Errorw("error in getting ci_workflow for artifacts", "err", err, "artifact", artifact, "parentStage", parentType, "stage", stage)
				return ciArtifactsResponse, err
			}
		}
		ciArtifacts[i].CiConfigureSourceType = ciWorkflow.GitTriggers[ciWorkflow.CiPipelineId].CiConfigureSourceType
		ciArtifacts[i].CiConfigureSourceValue = ciWorkflow.GitTriggers[ciWorkflow.CiPipelineId].CiConfigureSourceValue
	}
	ciArtifactsResponse.ResourceFilters = filters
	ciArtifactsResponse.CdPipelineId = pipeline.Id
	ciArtifactsResponse.LatestWfArtifactId = latestWfArtifactId
	ciArtifactsResponse.LatestWfArtifactStatus = latestWfArtifactStatus
	if ciArtifacts == nil {
		ciArtifacts = []bean2.CiArtifactBean{}
	}
	ciArtifactsResponse.CiArtifacts = ciArtifacts
	return ciArtifactsResponse, nil
}

func (impl *AppArtifactManagerImpl) BuildArtifactsList(listingFilterOpts *bean.ArtifactsListFilterOptions, isApprovalNode bool) ([]*bean2.CiArtifactBean, int, string, error) {

	var ciArtifacts []*bean2.CiArtifactBean

	//1)get current deployed artifact on this pipeline
	latestWf, err := impl.cdWorkflowRepository.FindArtifactByPipelineIdAndRunnerType(listingFilterOpts.PipelineId, listingFilterOpts.StageType, "", 1)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting latest workflow by pipelineId", "pipelineId", listingFilterOpts.PipelineId, "currentStageType", listingFilterOpts.StageType)
		return ciArtifacts, 0, "", err
	}

	var currentRunningArtifactBean *bean2.CiArtifactBean
	currentRunningArtifactId := 0
	currentRunningWorkflowStatus := ""

	//no artifacts deployed on this pipeline yet
	if len(latestWf) > 0 {

		currentRunningArtifact := latestWf[0].CdWorkflow.CiArtifact
		if !isApprovalNode {
			listingFilterOpts.ExcludeArtifactIds = []int{currentRunningArtifact.Id}
		}
		currentRunningArtifactId = currentRunningArtifact.Id
		currentRunningWorkflowStatus = latestWf[0].Status
		//current deployed artifact should always be computed, as we have to show it every time
		mInfo, err := parseMaterialInfo([]byte(currentRunningArtifact.MaterialInfo), currentRunningArtifact.DataSource)
		if err != nil {
			mInfo = []byte("[]")
			impl.logger.Errorw("Error in parsing artifact material info", "err", err, "artifact", currentRunningArtifact)
		}
		currentRunningArtifactBean = &bean2.CiArtifactBean{
			Id:           currentRunningArtifact.Id,
			Image:        currentRunningArtifact.Image,
			ImageDigest:  currentRunningArtifact.ImageDigest,
			MaterialInfo: mInfo,
			ScanEnabled:  currentRunningArtifact.ScanEnabled,
			Scanned:      currentRunningArtifact.Scanned,
			Deployed:     true,
			DeployedTime: formatDate(latestWf[0].CdWorkflow.CreatedOn, bean2.LayoutRFC3339),
			Latest:       true,
		}
		if currentRunningArtifact.WorkflowId != nil {
			currentRunningArtifactBean.CiWorkflowId = *currentRunningArtifact.WorkflowId
		}
	}
	//2) get artifact list limited by filterOptions

	//if approval configured and request is for deploy stage, fetch approved images only
	if listingFilterOpts.ApprovalNodeConfigured && listingFilterOpts.StageType == bean.CD_WORKFLOW_TYPE_DEPLOY && !isApprovalNode { //currently approval node is configured for this deploy stage
		ciArtifacts, err = impl.fetchApprovedArtifacts(listingFilterOpts, currentRunningArtifactBean)
		if err != nil {
			impl.logger.Errorw("error in fetching approved artifacts for cd pipeline", "pipelineId", listingFilterOpts.PipelineId, "err", err)
			return ciArtifacts, currentRunningArtifactId, currentRunningWorkflowStatus, err
		}
		return ciArtifacts, currentRunningArtifactId, currentRunningWorkflowStatus, nil

	} else {

		//if parent pipeline is CI/WEBHOOK, get all the ciArtifacts limited by listingFilterOpts
		if listingFilterOpts.ParentStageType == bean.CI_WORKFLOW_TYPE || listingFilterOpts.ParentStageType == bean.WEBHOOK_WORKFLOW_TYPE {
			ciArtifacts, err = impl.BuildArtifactsForCIParentV2(listingFilterOpts, isApprovalNode)
			if err != nil {
				impl.logger.Errorw("error in getting ci artifacts for ci/webhook type parent", "pipelineId", listingFilterOpts.PipelineId, "parentPipelineId", listingFilterOpts.ParentId, "parentStageType", listingFilterOpts.ParentStageType, "currentStageType", listingFilterOpts.StageType)
				return ciArtifacts, 0, "", err
			}
		} else {
			//if parent pipeline is PRE_CD/POST_CD/CD, then compute ciArtifacts using listingFilterOpts
			ciArtifacts, err = impl.BuildArtifactsForCdStageV2(listingFilterOpts, isApprovalNode)
			if err != nil {
				impl.logger.Errorw("error in getting ci artifacts for ci/webhook type parent", "pipelineId", listingFilterOpts.PipelineId, "parentPipelineId", listingFilterOpts.ParentId, "parentStageType", listingFilterOpts.ParentStageType, "currentStageType", listingFilterOpts.StageType)
				return ciArtifacts, 0, "", err
			}
		}

		var userApprovalMetadata map[int]*pipelineConfig.UserApprovalMetadata
		if isApprovalNode {
			artifactIds := make([]int, len(ciArtifacts))
			for i, artifact := range ciArtifacts {
				artifactIds[i] = artifact.Id
			}
			userApprovalMetadata, err = impl.workflowDagExecutor.FetchApprovalDataForArtifacts(artifactIds, listingFilterOpts.PipelineId, listingFilterOpts.ApproversCount) // it will fetch all the request data with nil cd_wfr_rnr_id
			if err != nil {
				impl.logger.Errorw("error occurred while fetching approval data for artifacts", "cdPipelineId", listingFilterOpts.PipelineId, "artifactIds", artifactIds, "err", err)
				return ciArtifacts, 0, "", err
			}
			for i, artifact := range ciArtifacts {
				if approvalMetadataForArtifact, ok := userApprovalMetadata[artifact.Id]; ok {
					ciArtifacts[i].UserApprovalMetadata = approvalMetadataForArtifact
				}
			}
		}

	}

	//we don't need currently deployed artifact for approvalNode explicitly
	//if no artifact deployed skip adding currentRunningArtifactBean in ciArtifacts arr
	if !isApprovalNode && currentRunningArtifactBean != nil {
		ciArtifacts = append(ciArtifacts, currentRunningArtifactBean)
	}
	return ciArtifacts, currentRunningArtifactId, currentRunningWorkflowStatus, nil
}

func (impl *AppArtifactManagerImpl) BuildArtifactsForCdStageV2(listingFilterOpts *bean.ArtifactsListFilterOptions, isApprovalNode bool) ([]*bean2.CiArtifactBean, error) {
	cdWfrList, err := impl.cdWorkflowRepository.FindArtifactByListFilter(listingFilterOpts, isApprovalNode)
	if err != nil {
		impl.logger.Errorw("error in fetching cd workflow runners using filter", "filterOptions", listingFilterOpts, "err", err)
		return nil, err
	}
	ciArtifacts := make([]*bean2.CiArtifactBean, 0, len(cdWfrList))

	//get artifact running on parent cd
	artifactRunningOnParentCd := 0
	if listingFilterOpts.ParentCdId > 0 {
		//TODO: check if we can fetch LastSuccessfulTriggerOnParent wfr along with last running wf
		parentCdWfrList, err := impl.cdWorkflowRepository.FindArtifactByPipelineIdAndRunnerType(listingFilterOpts.ParentCdId, bean.CD_WORKFLOW_TYPE_DEPLOY, "", 1)
		if err != nil || len(parentCdWfrList) == 0 {
			impl.logger.Errorw("error in getting artifact for parent cd", "parentCdPipelineId", listingFilterOpts.ParentCdId)
			return ciArtifacts, err
		}
		artifactRunningOnParentCd = parentCdWfrList[0].CdWorkflow.CiArtifact.Id
	}

	for _, wfr := range cdWfrList {
		mInfo, err := parseMaterialInfo([]byte(wfr.CdWorkflow.CiArtifact.MaterialInfo), wfr.CdWorkflow.CiArtifact.DataSource)
		if err != nil {
			mInfo = []byte("[]")
			impl.logger.Errorw("Error in parsing artifact material info", "err", err)
		}
		ciArtifact := &bean2.CiArtifactBean{
			Id:           wfr.CdWorkflow.CiArtifact.Id,
			Image:        wfr.CdWorkflow.CiArtifact.Image,
			ImageDigest:  wfr.CdWorkflow.CiArtifact.ImageDigest,
			MaterialInfo: mInfo,
			//TODO:LastSuccessfulTriggerOnParent
			Scanned:              wfr.CdWorkflow.CiArtifact.Scanned,
			ScanEnabled:          wfr.CdWorkflow.CiArtifact.ScanEnabled,
			RunningOnParentCd:    wfr.CdWorkflow.CiArtifact.Id == artifactRunningOnParentCd,
			ExternalCiPipelineId: wfr.CdWorkflow.CiArtifact.ExternalCiPipelineId,
			ParentCiArtifact:     wfr.CdWorkflow.CiArtifact.ParentCiArtifact,
		}
		if wfr.CdWorkflow.CiArtifact.WorkflowId != nil {
			ciArtifact.CiWorkflowId = *wfr.CdWorkflow.CiArtifact.WorkflowId
		}
		ciArtifacts = append(ciArtifacts, ciArtifact)
	}

	return ciArtifacts, nil
}

func (impl *AppArtifactManagerImpl) BuildArtifactsForCIParentV2(listingFilterOpts *bean.ArtifactsListFilterOptions, isApprovalNode bool) ([]*bean2.CiArtifactBean, error) {

	artifacts, err := impl.ciArtifactRepository.GetArtifactsByCDPipelineV3(listingFilterOpts, isApprovalNode)
	if err != nil {
		impl.logger.Errorw("error in getting artifacts for ci", "err", err)
		return nil, err
	}

	ciArtifacts := make([]*bean2.CiArtifactBean, 0, len(artifacts))
	for _, artifact := range artifacts {
		mInfo, err := parseMaterialInfo([]byte(artifact.MaterialInfo), artifact.DataSource)
		if err != nil {
			mInfo = []byte("[]")
			impl.logger.Errorw("Error in parsing artifact material info", "err", err, "artifact", artifact)
		}
		ciArtifact := &bean2.CiArtifactBean{
			Id:                   artifact.Id,
			Image:                artifact.Image,
			ImageDigest:          artifact.ImageDigest,
			MaterialInfo:         mInfo,
			ScanEnabled:          artifact.ScanEnabled,
			Scanned:              artifact.Scanned,
			Deployed:             artifact.Deployed,
			DeployedTime:         formatDate(artifact.DeployedTime, bean2.LayoutRFC3339),
			ExternalCiPipelineId: artifact.ExternalCiPipelineId,
			ParentCiArtifact:     artifact.ParentCiArtifact,
		}
		if artifact.WorkflowId != nil {
			ciArtifact.CiWorkflowId = *artifact.WorkflowId
		}
		ciArtifacts = append(ciArtifacts, ciArtifact)
	}

	return ciArtifacts, nil
}

func (impl *AppArtifactManagerImpl) fetchApprovedArtifacts(listingFilterOpts *bean.ArtifactsListFilterOptions, currentRunningArtifactBean *bean2.CiArtifactBean) ([]*bean2.CiArtifactBean, error) {
	artifacts, err := impl.ciArtifactRepository.FindApprovedArtifactsWithFilter(listingFilterOpts)
	if err != nil {
		impl.logger.Errorw("error in fetching approved image list", "pipelineId", listingFilterOpts.PipelineId, "err", err)
		return nil, err
	}
	ciArtifacts := make([]*bean2.CiArtifactBean, 0, len(artifacts))

	//get approval metadata for above ciArtifacts and current running artifact
	//TODO Gireesh: init array with default size and using append is not optimized
	artifactIds := make([]int, 0, len(artifacts)+1)
	for _, item := range artifacts {
		artifactIds = append(artifactIds, item.Id)
	}
	if currentRunningArtifactBean != nil {
		artifactIds = append(artifactIds, currentRunningArtifactBean.Id)
	}

	var userApprovalMetadata map[int]*pipelineConfig.UserApprovalMetadata
	userApprovalMetadata, err = impl.workflowDagExecutor.FetchApprovalDataForArtifacts(artifactIds, listingFilterOpts.PipelineId, listingFilterOpts.ApproversCount) // it will fetch all the request data with nil cd_wfr_rnr_id
	if err != nil {
		impl.logger.Errorw("error occurred while fetching approval data for artifacts", "cdPipelineId", listingFilterOpts.PipelineId, "artifactIds", artifactIds, "err", err)
		return ciArtifacts, err
	}

	//TODO Gireesh: this needs refactoring
	for _, artifact := range artifacts {
		mInfo, err := parseMaterialInfo([]byte(artifact.MaterialInfo), artifact.DataSource)
		if err != nil {
			mInfo = []byte("[]")
			impl.logger.Errorw("Error in parsing artifact material info", "err", err, "artifact", artifact)
		}
		ciArtifact := &bean2.CiArtifactBean{
			Id:                   artifact.Id,
			Image:                artifact.Image,
			ImageDigest:          artifact.ImageDigest,
			MaterialInfo:         mInfo,
			ScanEnabled:          artifact.ScanEnabled,
			Scanned:              artifact.Scanned,
			Deployed:             artifact.Deployed,
			DeployedTime:         formatDate(artifact.DeployedTime, bean2.LayoutRFC3339),
			ExternalCiPipelineId: artifact.ExternalCiPipelineId,
			ParentCiArtifact:     artifact.ParentCiArtifact,
		}
		if artifact.WorkflowId != nil {
			ciArtifact.CiWorkflowId = *artifact.WorkflowId
		}

		if approvalMetadataForArtifact, ok := userApprovalMetadata[artifact.Id]; ok {
			ciArtifact.UserApprovalMetadata = approvalMetadataForArtifact
		}

		ciArtifacts = append(ciArtifacts, ciArtifact)
	}

	if currentRunningArtifactBean != nil {
		if approvalMetadataForArtifact, ok := userApprovalMetadata[currentRunningArtifactBean.Id]; ok {
			currentRunningArtifactBean.UserApprovalMetadata = approvalMetadataForArtifact
		}
		ciArtifacts = append(ciArtifacts, currentRunningArtifactBean)
	}

	return ciArtifacts, nil
}
