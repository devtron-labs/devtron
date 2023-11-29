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
	"github.com/devtron-labs/devtron/internal/sql/repository"
	dockerArtifactStoreRegistry "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"sort"
	"strings"
)

type AppArtifactManager interface {
	//RetrieveArtifactsByCDPipeline : RetrieveArtifactsByCDPipeline returns all the artifacts for the cd pipeline (pre / deploy / post)
	RetrieveArtifactsByCDPipeline(pipeline *pipelineConfig.Pipeline, stage bean.WorkflowType) (*bean2.CiArtifactResponse, error)

	RetrieveArtifactsByCDPipelineV2(pipeline *pipelineConfig.Pipeline, stage bean.WorkflowType, artifactListingFilterOpts *bean.ArtifactsListFilterOptions) (*bean2.CiArtifactResponse, error)

	//FetchArtifactForRollback :
	FetchArtifactForRollback(cdPipelineId, appId, offset, limit int, searchString string) (bean2.CiArtifactResponse, error)

	FetchArtifactForRollbackV2(cdPipelineId, appId, offset, limit int, searchString string, app *bean2.CreateAppDTO, deploymentPipeline *pipelineConfig.Pipeline) (bean2.CiArtifactResponse, error)

	BuildArtifactsForCdStage(pipelineId int, stageType bean.WorkflowType, ciArtifacts []bean2.CiArtifactBean, artifactMap map[int]int, parent bool, limit int, parentCdId int) ([]bean2.CiArtifactBean, map[int]int, int, string, error)

	BuildArtifactsForParentStage(cdPipelineId int, parentId int, parentType bean.WorkflowType, ciArtifacts []bean2.CiArtifactBean, artifactMap map[int]int, limit int, parentCdId int) ([]bean2.CiArtifactBean, error)
}

type AppArtifactManagerImpl struct {
	logger                  *zap.SugaredLogger
	cdWorkflowRepository    pipelineConfig.CdWorkflowRepository
	userService             user.UserService
	imageTaggingService     ImageTaggingService
	ciArtifactRepository    repository.CiArtifactRepository
	ciWorkflowRepository    pipelineConfig.CiWorkflowRepository
	pipelineStageService    PipelineStageService
	cdPipelineConfigService CdPipelineConfigService
	dockerArtifactRegistry  dockerArtifactStoreRegistry.DockerArtifactStoreRepository
	CiPipelineRepository    pipelineConfig.CiPipelineRepository
	ciTemplateService       CiTemplateService
}

func NewAppArtifactManagerImpl(
	logger *zap.SugaredLogger,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	userService user.UserService,
	imageTaggingService ImageTaggingService,
	ciArtifactRepository repository.CiArtifactRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	pipelineStageService PipelineStageService,
	cdPipelineConfigService CdPipelineConfigService,
	dockerArtifactRegistry dockerArtifactStoreRegistry.DockerArtifactStoreRepository,
	CiPipelineRepository pipelineConfig.CiPipelineRepository,
	ciTemplateService CiTemplateService) *AppArtifactManagerImpl {

	return &AppArtifactManagerImpl{
		logger:                  logger,
		cdWorkflowRepository:    cdWorkflowRepository,
		userService:             userService,
		imageTaggingService:     imageTaggingService,
		ciArtifactRepository:    ciArtifactRepository,
		ciWorkflowRepository:    ciWorkflowRepository,
		cdPipelineConfigService: cdPipelineConfigService,
		pipelineStageService:    pipelineStageService,
		dockerArtifactRegistry:  dockerArtifactRegistry,
		CiPipelineRepository:    CiPipelineRepository,
		ciTemplateService:       ciTemplateService,
	}
}

func (impl *AppArtifactManagerImpl) BuildArtifactsForParentStage(cdPipelineId int, parentId int, parentType bean.WorkflowType, ciArtifacts []bean2.CiArtifactBean, artifactMap map[int]int, limit int, parentCdId int) ([]bean2.CiArtifactBean, error) {
	var ciArtifactsFinal []bean2.CiArtifactBean
	var err error
	if parentType == bean.CI_WORKFLOW_TYPE {
		ciArtifactsFinal, err = impl.BuildArtifactsForCIParent(cdPipelineId, parentId, parentType, ciArtifacts, artifactMap, limit)
	} else if parentType == bean.WEBHOOK_WORKFLOW_TYPE {
		ciArtifactsFinal, err = impl.BuildArtifactsForCIParent(cdPipelineId, parentId, parentType, ciArtifacts, artifactMap, limit)
	} else {
		//parent type is PRE, POST or DEPLOY type
		ciArtifactsFinal, _, _, _, err = impl.BuildArtifactsForCdStage(parentId, parentType, ciArtifacts, artifactMap, true, limit, parentCdId)
	}
	return ciArtifactsFinal, err
}

func (impl *AppArtifactManagerImpl) BuildArtifactsForCdStage(pipelineId int, stageType bean.WorkflowType, ciArtifacts []bean2.CiArtifactBean, artifactMap map[int]int, parent bool, limit int, parentCdId int) ([]bean2.CiArtifactBean, map[int]int, int, string, error) {
	//getting running artifact id for parent cd
	parentCdRunningArtifactId := 0
	if parentCdId > 0 && parent {
		parentCdWfrList, err := impl.cdWorkflowRepository.FindArtifactByPipelineIdAndRunnerType(parentCdId, bean.CD_WORKFLOW_TYPE_DEPLOY, 1, nil)
		if err != nil || len(parentCdWfrList) == 0 {
			impl.logger.Errorw("error in getting artifact for parent cd", "parentCdPipelineId", parentCdId)
			return ciArtifacts, artifactMap, 0, "", err
		}
		parentCdRunningArtifactId = parentCdWfrList[0].CdWorkflow.CiArtifact.Id
	}
	//getting wfr for parent and updating artifacts
	parentWfrList, err := impl.cdWorkflowRepository.FindArtifactByPipelineIdAndRunnerType(pipelineId, stageType, limit, nil)
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
					CiPipelineId:                  wfr.CdWorkflow.CiArtifact.PipelineId,
					CredentialsSourceType:         wfr.CdWorkflow.CiArtifact.CredentialsSourceType,
					CredentialsSourceValue:        wfr.CdWorkflow.CiArtifact.CredentialSourceValue,
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

func (impl *AppArtifactManagerImpl) BuildArtifactsForCIParent(cdPipelineId int, parentId int, parentType bean.WorkflowType, ciArtifacts []bean2.CiArtifactBean, artifactMap map[int]int, limit int) ([]bean2.CiArtifactBean, error) {
	artifacts, err := impl.ciArtifactRepository.GetArtifactsByCDPipeline(cdPipelineId, limit, parentId, parentType)
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
				Id:                     artifact.Id,
				Image:                  artifact.Image,
				ImageDigest:            artifact.ImageDigest,
				MaterialInfo:           mInfo,
				ScanEnabled:            artifact.ScanEnabled,
				Scanned:                artifact.Scanned,
				CiPipelineId:           artifact.PipelineId,
				CredentialsSourceType:  artifact.CredentialsSourceType,
				CredentialsSourceValue: artifact.CredentialSourceValue,
			})
		}
	}
	return ciArtifacts, nil
}

func (impl *AppArtifactManagerImpl) FetchArtifactForRollback(cdPipelineId, appId, offset, limit int, searchString string) (bean2.CiArtifactResponse, error) {
	var deployedCiArtifacts []bean2.CiArtifactBean
	var deployedCiArtifactsResponse bean2.CiArtifactResponse

	cdWfrs, err := impl.cdWorkflowRepository.FetchArtifactsByCdPipelineId(cdPipelineId, bean.CD_WORKFLOW_TYPE_DEPLOY, offset, limit, searchString)
	if err != nil {
		impl.logger.Errorw("error in getting artifacts for rollback by cdPipelineId", "err", err, "cdPipelineId", cdPipelineId)
		return deployedCiArtifactsResponse, err
	}
	var ids []int32
	for _, item := range cdWfrs {
		ids = append(ids, item.TriggeredBy)
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
	deployedCiArtifactsResponse.CiArtifacts = deployedCiArtifacts

	return deployedCiArtifactsResponse, nil
}

func (impl *AppArtifactManagerImpl) FetchArtifactForRollbackV2(cdPipelineId, appId, offset, limit int, searchString string, app *bean2.CreateAppDTO, deploymentPipeline *pipelineConfig.Pipeline) (bean2.CiArtifactResponse, error) {
	var deployedCiArtifactsResponse bean2.CiArtifactResponse
	imageTagsDataMap, err := impl.imageTaggingService.GetTagsDataMapByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in getting image tagging data with appId", "err", err, "appId", appId)
		return deployedCiArtifactsResponse, err
	}

	artifactListingFilterOpts := bean.ArtifactsListFilterOptions{}
	artifactListingFilterOpts.PipelineId = cdPipelineId
	artifactListingFilterOpts.StageType = bean.CD_WORKFLOW_TYPE_DEPLOY
	artifactListingFilterOpts.SearchString = "%" + searchString + "%"
	artifactListingFilterOpts.Limit = limit
	artifactListingFilterOpts.Offset = offset
	deployedCiArtifacts, artifactIds, totalCount, err := impl.BuildRollbackArtifactsList(artifactListingFilterOpts)
	if err != nil {
		impl.logger.Errorw("error in building ci artifacts for rollback", "err", err, "cdPipelineId", cdPipelineId)
		return deployedCiArtifactsResponse, err
	}

	imageCommentsDataMap, err := impl.imageTaggingService.GetImageCommentsDataMapByArtifactIds(artifactIds)
	if err != nil {
		impl.logger.Errorw("error in getting GetImageCommentsDataMapByArtifactIds", "err", err, "appId", appId, "artifactIds", artifactIds)
		return deployedCiArtifactsResponse, err
	}

	for i, _ := range deployedCiArtifacts {
		imageTaggingResp := imageTagsDataMap[deployedCiArtifacts[i].Id]
		if imageTaggingResp != nil {
			deployedCiArtifacts[i].ImageReleaseTags = imageTaggingResp
		}
		if imageCommentResp := imageCommentsDataMap[deployedCiArtifacts[i].Id]; imageCommentResp != nil {
			deployedCiArtifacts[i].ImageComment = imageCommentResp
		}
		var dockerRegistryId string
		if deployedCiArtifacts[i].DataSource == repository.POST_CI || deployedCiArtifacts[i].DataSource == repository.PRE_CD || deployedCiArtifacts[i].DataSource == repository.POST_CD {
			if deployedCiArtifacts[i].CredentialsSourceType == repository.GLOBAL_CONTAINER_REGISTRY {
				dockerRegistryId = deployedCiArtifacts[i].CredentialsSourceValue
			}
		} else {
			ciPipeline, err := impl.CiPipelineRepository.FindById(deployedCiArtifacts[i].CiPipelineId)
			if err != nil {
				impl.logger.Errorw("error in fetching ciPipeline", "ciPipelineId", ciPipeline.Id, "error", err)
				return deployedCiArtifactsResponse, err
			}
			dockerRegistryId = *ciPipeline.CiTemplate.DockerRegistryId
		}
		if len(dockerRegistryId) > 0 {
			dockerArtifact, err := impl.dockerArtifactRegistry.FindOne(dockerRegistryId)
			if err != nil {
				impl.logger.Errorw("error in getting docker registry details", "err", err, "dockerArtifactStoreId", dockerRegistryId)
			}
			deployedCiArtifacts[i].RegistryType = string(dockerArtifact.RegistryType)
			deployedCiArtifacts[i].RegistryName = dockerRegistryId
		}
	}

	deployedCiArtifactsResponse.CdPipelineId = cdPipelineId
	if deployedCiArtifacts == nil {
		deployedCiArtifacts = []bean2.CiArtifactBean{}
	}
	deployedCiArtifactsResponse.CiArtifacts = deployedCiArtifacts
	deployedCiArtifactsResponse.TotalCount = totalCount
	return deployedCiArtifactsResponse, nil
}

func (impl *AppArtifactManagerImpl) BuildRollbackArtifactsList(artifactListingFilterOpts bean.ArtifactsListFilterOptions) ([]bean2.CiArtifactBean, []int, int, error) {
	var deployedCiArtifacts []bean2.CiArtifactBean
	totalCount := 0

	//1)get current deployed artifact on this pipeline
	latestWf, err := impl.cdWorkflowRepository.FindArtifactByPipelineIdAndRunnerType(artifactListingFilterOpts.PipelineId, artifactListingFilterOpts.StageType, 1, []string{application.Healthy, application.SUCCEEDED, application.Progressing})
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting latest workflow by pipelineId", "pipelineId", artifactListingFilterOpts.PipelineId, "currentStageType", artifactListingFilterOpts.StageType)
		return deployedCiArtifacts, nil, totalCount, err
	}
	if len(latestWf) > 0 {
		//we should never show current deployed artifact in rollback API
		artifactListingFilterOpts.ExcludeWfrIds = []int{latestWf[0].Id}
	}

	ciArtifacts, totalCount, err := impl.ciArtifactRepository.FetchArtifactsByCdPipelineIdV2(artifactListingFilterOpts)

	if err != nil {
		impl.logger.Errorw("error in getting artifacts for rollback by cdPipelineId", "err", err, "cdPipelineId", artifactListingFilterOpts.PipelineId)
		return deployedCiArtifacts, nil, totalCount, err
	}

	var ids []int32
	for _, item := range ciArtifacts {
		ids = append(ids, item.TriggeredBy)
	}

	userEmails := make(map[int32]string)
	users, err := impl.userService.GetByIds(ids)
	if err != nil {
		impl.logger.Errorw("unable to fetch users by ids", "err", err, "ids", ids)
	}
	for _, item := range users {
		userEmails[item.Id] = item.EmailId
	}

	artifactIds := make([]int, 0)

	for _, ciArtifact := range ciArtifacts {
		mInfo, err := parseMaterialInfo([]byte(ciArtifact.MaterialInfo), ciArtifact.DataSource)
		if err != nil {
			mInfo = []byte("[]")
			impl.logger.Errorw("error in parsing ciArtifact material info", "err", err, "ciArtifact", ciArtifact)
		}
		userEmail := userEmails[ciArtifact.TriggeredBy]
		deployedCiArtifacts = append(deployedCiArtifacts, bean2.CiArtifactBean{
			Id:                     ciArtifact.Id,
			Image:                  ciArtifact.Image,
			MaterialInfo:           mInfo,
			DeployedTime:           formatDate(ciArtifact.StartedOn, bean2.LayoutRFC3339),
			WfrId:                  ciArtifact.CdWorkflowRunnerId,
			DeployedBy:             userEmail,
			Scanned:                ciArtifact.Scanned,
			ScanEnabled:            ciArtifact.ScanEnabled,
			CiPipelineId:           ciArtifact.PipelineId,
			CredentialsSourceType:  ciArtifact.CredentialsSourceType,
			CredentialsSourceValue: ciArtifact.CredentialSourceValue,
			DataSource:             ciArtifact.DataSource,
		})
		artifactIds = append(artifactIds, ciArtifact.Id)
	}
	return deployedCiArtifacts, artifactIds, totalCount, nil

}

func (impl *AppArtifactManagerImpl) RetrieveArtifactsByCDPipeline(pipeline *pipelineConfig.Pipeline, stage bean.WorkflowType) (*bean2.CiArtifactResponse, error) {

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
		BuildArtifactsForCdStage(pipeline.Id, stage, ciArtifacts, artifactMap, false, limit, parentCdId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting artifacts for child cd stage", "err", err, "stage", stage)
		return nil, err
	}

	ciArtifacts, err = impl.BuildArtifactsForParentStage(pipeline.Id, parentId, parentType, ciArtifacts, artifactMap, limit, parentCdId)
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

	artifacts, err := impl.ciArtifactRepository.GetArtifactParentCiAndWorkflowDetailsByIds(artifactIds)
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

		if artifact.ExternalCiPipelineId != 0 {
			// if external webhook continue
			continue
		}
		var dockerRegistryId string
		if artifact.PipelineId != 0 {
			ciPipeline, err := impl.CiPipelineRepository.FindById(artifact.PipelineId)
			if err != nil {
				impl.logger.Errorw("error in fetching ciPipeline", "ciPipelineId", ciPipeline.Id, "error", err)
				return nil, err
			}
			dockerRegistryId = *ciPipeline.CiTemplate.DockerRegistryId
		} else {
			if artifact.CredentialsSourceType == repository.GLOBAL_CONTAINER_REGISTRY {
				dockerRegistryId = artifact.CredentialSourceValue
			}
		}
		if len(dockerRegistryId) > 0 {
			dockerArtifact, err := impl.dockerArtifactRegistry.FindOne(dockerRegistryId)
			if err != nil {
				impl.logger.Errorw("error in getting docker registry details", "err", err, "dockerArtifactStoreId", dockerRegistryId)
			}
			ciArtifacts[i].RegistryType = string(dockerArtifact.RegistryType)
			ciArtifacts[i].RegistryName = dockerRegistryId
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
	return ciArtifactsResponse, nil
}

func (impl *AppArtifactManagerImpl) extractParentMetaDataByPipeline(pipeline *pipelineConfig.Pipeline, stage bean.WorkflowType) (parentId int, parentType bean.WorkflowType, parentCdId int, err error) {
	// retrieve parent details
	parentId, parentType, err = impl.cdPipelineConfigService.RetrieveParentDetails(pipeline.Id)
	if err != nil {
		return parentId, parentType, parentCdId, err
	}

	//TODO Gireesh: why this(stage != bean.CD_WORKFLOW_TYPE_POST) check is added, explain that in comment ??
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

func (impl *AppArtifactManagerImpl) RetrieveArtifactsByCDPipelineV2(pipeline *pipelineConfig.Pipeline, stage bean.WorkflowType, artifactListingFilterOpts *bean.ArtifactsListFilterOptions) (*bean2.CiArtifactResponse, error) {

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
	artifactListingFilterOpts.SearchString = "%" + artifactListingFilterOpts.SearchString + "%"
	ciArtifactsRefs, latestWfArtifactId, latestWfArtifactStatus, totalCount, err := impl.BuildArtifactsList(artifactListingFilterOpts)
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

		ciArtifacts, err = impl.setAdditionalDataInArtifacts(ciArtifacts, pipeline)
		if err != nil {
			impl.logger.Errorw("error in setting additional data in fetched artifacts", "pipelineId", pipeline.Id, "err", err)
			return ciArtifactsResponse, err
		}
	}

	ciArtifactsResponse.CdPipelineId = pipeline.Id
	ciArtifactsResponse.LatestWfArtifactId = latestWfArtifactId
	ciArtifactsResponse.LatestWfArtifactStatus = latestWfArtifactStatus
	if ciArtifacts == nil {
		ciArtifacts = []bean2.CiArtifactBean{}
	}
	ciArtifactsResponse.CiArtifacts = ciArtifacts
	ciArtifactsResponse.TotalCount = totalCount
	return ciArtifactsResponse, nil
}

func (impl *AppArtifactManagerImpl) setAdditionalDataInArtifacts(ciArtifacts []bean2.CiArtifactBean, pipeline *pipelineConfig.Pipeline) ([]bean2.CiArtifactBean, error) {
	artifactIds := make([]int, 0, len(ciArtifacts))
	for _, artifact := range ciArtifacts {
		artifactIds = append(artifactIds, artifact.Id)
	}

	imageTagsDataMap, err := impl.imageTaggingService.GetTagsDataMapByAppId(pipeline.AppId)
	if err != nil {
		impl.logger.Errorw("error in getting image tagging data with appId", "err", err, "appId", pipeline.AppId)
		return ciArtifacts, err
	}

	imageCommentsDataMap, err := impl.imageTaggingService.GetImageCommentsDataMapByArtifactIds(artifactIds)
	if err != nil {
		impl.logger.Errorw("error in getting GetImageCommentsDataMapByArtifactIds", "err", err, "appId", pipeline.AppId, "artifactIds", artifactIds)
		return ciArtifacts, err
	}

	for i, _ := range ciArtifacts {
		imageTaggingResp := imageTagsDataMap[ciArtifacts[i].Id]
		if imageTaggingResp != nil {
			ciArtifacts[i].ImageReleaseTags = imageTaggingResp
		}
		if imageCommentResp := imageCommentsDataMap[ciArtifacts[i].Id]; imageCommentResp != nil {
			ciArtifacts[i].ImageComment = imageCommentResp
		}
		var dockerRegistryId string
		if ciArtifacts[i].DataSource == repository.POST_CI || ciArtifacts[i].DataSource == repository.PRE_CD || ciArtifacts[i].DataSource == repository.POST_CD {
			if ciArtifacts[i].CredentialsSourceType == repository.GLOBAL_CONTAINER_REGISTRY {
				dockerRegistryId = ciArtifacts[i].CredentialsSourceValue
			}
		} else if ciArtifacts[i].DataSource == repository.CI_RUNNER {
			ciPipeline, err := impl.CiPipelineRepository.FindById(ciArtifacts[i].CiPipelineId)
			if err != nil {
				impl.logger.Errorw("error in fetching ciPipeline", "ciPipelineId", ciPipeline.Id, "error", err)
				return nil, err
			}
			if !ciPipeline.IsExternal && ciPipeline.IsDockerConfigOverridden {
				ciTemplateBean, err := impl.ciTemplateService.FindTemplateOverrideByCiPipelineId(ciPipeline.Id)
				if err != nil {
					impl.logger.Errorw("error in fetching template override", "pipelineId", ciPipeline.Id, "err", err)
					return nil, err
				}
				dockerRegistryId = ciTemplateBean.CiTemplateOverride.DockerRegistryId
			} else {
				dockerRegistryId = *ciPipeline.CiTemplate.DockerRegistryId
			}
		}
		if len(dockerRegistryId) > 0 {
			dockerArtifact, err := impl.dockerArtifactRegistry.FindOne(dockerRegistryId)
			if err != nil {
				impl.logger.Errorw("error in getting docker registry details", "err", err, "dockerArtifactStoreId", dockerRegistryId)
			}
			ciArtifacts[i].RegistryType = string(dockerArtifact.RegistryType)
			ciArtifacts[i].RegistryName = dockerRegistryId
		}
	}
	return impl.setGitTriggerData(ciArtifacts)

}

func (impl *AppArtifactManagerImpl) setGitTriggerData(ciArtifacts []bean2.CiArtifactBean) ([]bean2.CiArtifactBean, error) {
	directArtifactIndexes, directWorkflowIds, artifactsWithParentIndexes, parentArtifactIds := make([]int, 0), make([]int, 0), make([]int, 0), make([]int, 0)
	for i, artifact := range ciArtifacts {
		if artifact.ExternalCiPipelineId != 0 {
			// if external webhook continue
			continue
		}
		//linked ci case
		if artifact.ParentCiArtifact != 0 {
			artifactsWithParentIndexes = append(artifactsWithParentIndexes, i)
			parentArtifactIds = append(parentArtifactIds, artifact.ParentCiArtifact)
		} else {
			directArtifactIndexes = append(directArtifactIndexes, i)
			directWorkflowIds = append(directWorkflowIds, artifact.CiWorkflowId)
		}
	}
	ciWorkflowWithArtifacts, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflowGitTriggersByArtifactIds(parentArtifactIds)
	if err != nil {
		impl.logger.Errorw("error in getting ci_workflow for artifacts", "err", err, "parentArtifactIds", parentArtifactIds)
		return ciArtifacts, err
	}

	parentArtifactIdVsCiWorkflowMap := make(map[int]*pipelineConfig.WorkflowWithArtifact)
	for _, ciWorkflow := range ciWorkflowWithArtifacts {
		parentArtifactIdVsCiWorkflowMap[ciWorkflow.CiArtifactId] = ciWorkflow
	}

	for _, index := range directArtifactIndexes {
		ciWorkflow := parentArtifactIdVsCiWorkflowMap[ciArtifacts[index].CiWorkflowId]
		if ciWorkflow != nil {
			ciArtifacts[index].CiConfigureSourceType = ciWorkflow.GitTriggers[ciWorkflow.CiPipelineId].CiConfigureSourceType
			ciArtifacts[index].CiConfigureSourceValue = ciWorkflow.GitTriggers[ciWorkflow.CiPipelineId].CiConfigureSourceValue
		}
	}

	ciWorkflows, err := impl.ciWorkflowRepository.FindCiWorkflowGitTriggersByIds(directWorkflowIds)
	if err != nil {
		impl.logger.Errorw("error in getting ci_workflow for artifacts", "err", err, "ciWorkflowIds", directWorkflowIds)
		return ciArtifacts, err
	}
	ciWorkflowMap := make(map[int]*pipelineConfig.CiWorkflow)
	for _, ciWorkflow := range ciWorkflows {
		ciWorkflowMap[ciWorkflow.Id] = ciWorkflow
	}
	for _, index := range directArtifactIndexes {
		ciWorkflow := ciWorkflowMap[ciArtifacts[index].CiWorkflowId]
		if ciWorkflow != nil {
			ciArtifacts[index].CiConfigureSourceType = ciWorkflow.GitTriggers[ciWorkflow.CiPipelineId].CiConfigureSourceType
			ciArtifacts[index].CiConfigureSourceValue = ciWorkflow.GitTriggers[ciWorkflow.CiPipelineId].CiConfigureSourceValue
		}
	}
	return ciArtifacts, nil
}

func (impl *AppArtifactManagerImpl) BuildArtifactsList(listingFilterOpts *bean.ArtifactsListFilterOptions) ([]*bean2.CiArtifactBean, int, string, int, error) {

	var ciArtifacts []*bean2.CiArtifactBean
	totalCount := 0
	//1)get current deployed artifact on this pipeline
	latestWf, err := impl.cdWorkflowRepository.FindArtifactByPipelineIdAndRunnerType(listingFilterOpts.PipelineId, listingFilterOpts.StageType, 1, []string{application.Healthy, application.SUCCEEDED, application.Progressing})
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting latest workflow by pipelineId", "pipelineId", listingFilterOpts.PipelineId, "currentStageType", listingFilterOpts.StageType)
		return ciArtifacts, 0, "", totalCount, err
	}

	var currentRunningArtifactBean *bean2.CiArtifactBean
	currentRunningArtifactId := 0
	currentRunningWorkflowStatus := ""

	//no artifacts deployed on this pipeline yet if latestWf is empty
	if len(latestWf) > 0 {

		currentRunningArtifact := latestWf[0].CdWorkflow.CiArtifact
		listingFilterOpts.ExcludeArtifactIds = []int{currentRunningArtifact.Id}
		currentRunningArtifactId = currentRunningArtifact.Id
		currentRunningWorkflowStatus = latestWf[0].Status
		// TODO Gireesh: move below logic to proper func belong to CiArtifactBean
		//current deployed artifact should always be computed, as we have to show it every time
		mInfo, err := parseMaterialInfo([]byte(currentRunningArtifact.MaterialInfo), currentRunningArtifact.DataSource)
		if err != nil {
			mInfo = []byte("[]")
			impl.logger.Errorw("Error in parsing artifact material info", "err", err, "artifact", currentRunningArtifact)
		}
		currentRunningArtifactBean = &bean2.CiArtifactBean{
			Id:                     currentRunningArtifact.Id,
			Image:                  currentRunningArtifact.Image,
			ImageDigest:            currentRunningArtifact.ImageDigest,
			MaterialInfo:           mInfo,
			ScanEnabled:            currentRunningArtifact.ScanEnabled,
			Scanned:                currentRunningArtifact.Scanned,
			Deployed:               true,
			DeployedTime:           formatDate(latestWf[0].CdWorkflow.CreatedOn, bean2.LayoutRFC3339),
			Latest:                 true,
			CreatedTime:            formatDate(currentRunningArtifact.CreatedOn, bean2.LayoutRFC3339),
			DataSource:             currentRunningArtifact.DataSource,
			CiPipelineId:           currentRunningArtifact.PipelineId,
			CredentialsSourceType:  currentRunningArtifact.CredentialsSourceType,
			CredentialsSourceValue: currentRunningArtifact.CredentialSourceValue,
		}
		if currentRunningArtifact.WorkflowId != nil {
			currentRunningArtifactBean.CiWorkflowId = *currentRunningArtifact.WorkflowId
		}
	}
	//2) get artifact list limited by filterOptions
	if listingFilterOpts.ParentStageType == bean.CI_WORKFLOW_TYPE || listingFilterOpts.ParentStageType == bean.WEBHOOK_WORKFLOW_TYPE {
		ciArtifacts, totalCount, err = impl.BuildArtifactsForCIParentV2(listingFilterOpts)
		if err != nil {
			impl.logger.Errorw("error in getting ci artifacts for ci/webhook type parent", "pipelineId", listingFilterOpts.PipelineId, "parentPipelineId", listingFilterOpts.ParentId, "parentStageType", listingFilterOpts.ParentStageType, "currentStageType", listingFilterOpts.StageType)
			return ciArtifacts, 0, "", totalCount, err
		}
	} else {
		if listingFilterOpts.ParentStageType == WorklowTypePre {
			listingFilterOpts.PluginStage = repository.PRE_CD
		} else if listingFilterOpts.ParentStageType == WorklowTypePost {
			listingFilterOpts.PluginStage = repository.POST_CD
		}
		ciArtifacts, totalCount, err = impl.BuildArtifactsForCdStageV2(listingFilterOpts)
		if err != nil {
			impl.logger.Errorw("error in getting ci artifacts for ci/webhook type parent", "pipelineId", listingFilterOpts.PipelineId, "parentPipelineId", listingFilterOpts.ParentId, "parentStageType", listingFilterOpts.ParentStageType, "currentStageType", listingFilterOpts.StageType)
			return ciArtifacts, 0, "", totalCount, err
		}
	}

	//if no artifact deployed skip adding currentRunningArtifactBean in ciArtifacts arr
	if currentRunningArtifactBean != nil {
		searchString := listingFilterOpts.SearchString[1 : len(listingFilterOpts.SearchString)-1]
		if strings.Contains(currentRunningArtifactBean.Image, searchString) {
			ciArtifacts = append(ciArtifacts, currentRunningArtifactBean)
			totalCount += 1
		}
	}

	return ciArtifacts, currentRunningArtifactId, currentRunningWorkflowStatus, totalCount, nil
}

func (impl *AppArtifactManagerImpl) BuildArtifactsForCdStageV2(listingFilterOpts *bean.ArtifactsListFilterOptions) ([]*bean2.CiArtifactBean, int, error) {
	cdArtifacts, totalCount, err := impl.ciArtifactRepository.FindArtifactByListFilter(listingFilterOpts)
	if err != nil {
		impl.logger.Errorw("error in fetching cd workflow runners using filter", "filterOptions", listingFilterOpts, "err", err)
		return nil, totalCount, err
	}
	ciArtifacts := make([]*bean2.CiArtifactBean, 0, len(cdArtifacts))

	//get artifact running on parent cd
	artifactRunningOnParentCd := 0
	if listingFilterOpts.ParentCdId > 0 {
		//TODO: check if we can fetch LastSuccessfulTriggerOnParent wfr along with last running wf
		parentCdWfrList, err := impl.cdWorkflowRepository.FindArtifactByPipelineIdAndRunnerType(listingFilterOpts.ParentCdId, bean.CD_WORKFLOW_TYPE_DEPLOY, 1, []string{application.Healthy, application.SUCCEEDED, application.Progressing})
		if err != nil || len(parentCdWfrList) == 0 {
			impl.logger.Errorw("error in getting artifact for parent cd", "parentCdPipelineId", listingFilterOpts.ParentCdId)
			return ciArtifacts, totalCount, err
		}
		artifactRunningOnParentCd = parentCdWfrList[0].CdWorkflow.CiArtifact.Id
	}

	for _, artifact := range cdArtifacts {
		mInfo, err := parseMaterialInfo([]byte(artifact.MaterialInfo), artifact.DataSource)
		if err != nil {
			mInfo = []byte("[]")
			impl.logger.Errorw("Error in parsing artifact material info", "err", err)
		}
		ciArtifact := &bean2.CiArtifactBean{
			Id:           artifact.Id,
			Image:        artifact.Image,
			ImageDigest:  artifact.ImageDigest,
			MaterialInfo: mInfo,
			//TODO:LastSuccessfulTriggerOnParent
			Scanned:                artifact.Scanned,
			ScanEnabled:            artifact.ScanEnabled,
			RunningOnParentCd:      artifact.Id == artifactRunningOnParentCd,
			ExternalCiPipelineId:   artifact.ExternalCiPipelineId,
			ParentCiArtifact:       artifact.ParentCiArtifact,
			CreatedTime:            formatDate(artifact.CreatedOn, bean2.LayoutRFC3339),
			DataSource:             artifact.DataSource,
			CiPipelineId:           artifact.PipelineId,
			CredentialsSourceType:  artifact.CredentialsSourceType,
			CredentialsSourceValue: artifact.CredentialSourceValue,
		}
		if artifact.WorkflowId != nil {
			ciArtifact.CiWorkflowId = *artifact.WorkflowId
		}
		ciArtifacts = append(ciArtifacts, ciArtifact)
	}

	return ciArtifacts, totalCount, nil
}

func (impl *AppArtifactManagerImpl) BuildArtifactsForCIParentV2(listingFilterOpts *bean.ArtifactsListFilterOptions) ([]*bean2.CiArtifactBean, int, error) {

	artifacts, totalCount, err := impl.ciArtifactRepository.GetArtifactsByCDPipelineV3(listingFilterOpts)
	if err != nil {
		impl.logger.Errorw("error in getting artifacts for ci", "err", err)
		return nil, totalCount, err
	}

	//TODO Gireesh: if initialized then no need of using append, put value directly to index
	ciArtifacts := make([]*bean2.CiArtifactBean, 0, len(artifacts))
	for _, artifact := range artifacts {
		mInfo, err := parseMaterialInfo([]byte(artifact.MaterialInfo), artifact.DataSource)
		if err != nil {
			mInfo = []byte("[]")
			impl.logger.Errorw("Error in parsing artifact material info", "err", err, "artifact", artifact)
		}
		ciArtifact := &bean2.CiArtifactBean{
			Id:                     artifact.Id,
			Image:                  artifact.Image,
			ImageDigest:            artifact.ImageDigest,
			MaterialInfo:           mInfo,
			ScanEnabled:            artifact.ScanEnabled,
			Scanned:                artifact.Scanned,
			Deployed:               artifact.Deployed,
			DeployedTime:           formatDate(artifact.DeployedTime, bean2.LayoutRFC3339),
			ExternalCiPipelineId:   artifact.ExternalCiPipelineId,
			ParentCiArtifact:       artifact.ParentCiArtifact,
			CreatedTime:            formatDate(artifact.CreatedOn, bean2.LayoutRFC3339),
			CiPipelineId:           artifact.PipelineId,
			DataSource:             artifact.DataSource,
			CredentialsSourceType:  artifact.CredentialsSourceType,
			CredentialsSourceValue: artifact.CredentialSourceValue,
		}
		if artifact.WorkflowId != nil {
			ciArtifact.CiWorkflowId = *artifact.WorkflowId
		}
		ciArtifacts = append(ciArtifacts, ciArtifact)
	}

	return ciArtifacts, totalCount, nil
}
