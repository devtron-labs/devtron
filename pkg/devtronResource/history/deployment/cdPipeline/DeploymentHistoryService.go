package cdPipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/history/deployment/cdPipeline/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/read"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/devtron-labs/devtron/pkg/devtronResource/util"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"go.uber.org/zap"
)

type DeploymentHistoryService interface {
	GetCdPipelineDeploymentHistory(offset, limit, appId, environmentId,
		pipelineId, filterByReleaseId int) (resp bean.DeploymentHistoryResp, err error)
	GetCdPipelineDeploymentHistoryConfigList(baseConfigurationId, pipelineId int,
		historyComponent, historyComponentName string) (resp []*history.DeployedHistoryComponentMetadataDto, err error)
}

type DeploymentHistoryServiceImpl struct {
	logger                              *zap.SugaredLogger
	cdHandler                           pipeline.CdHandler
	imageTaggingService                 pipeline.ImageTaggingService
	pipelineRepository                  pipelineConfig.PipelineRepository
	dtResourceTaskRunRepository         repository.DevtronResourceTaskRunRepository
	dtResourceSchemaRepository          repository.DevtronResourceSchemaRepository
	dtResourceReadService               read.ReadService
	deployedConfigurationHistoryService history.DeployedConfigurationHistoryService
}

func NewDeploymentHistoryServiceImpl(logger *zap.SugaredLogger,
	cdHandler pipeline.CdHandler,
	imageTaggingService pipeline.ImageTaggingService,
	pipelineRepository pipelineConfig.PipelineRepository,
	dtResourceTaskRunRepository repository.DevtronResourceTaskRunRepository,
	dtResourceSchemaRepository repository.DevtronResourceSchemaRepository,
	dtResourceReadService read.ReadService,
	deployedConfigurationHistoryService history.DeployedConfigurationHistoryService) *DeploymentHistoryServiceImpl {
	return &DeploymentHistoryServiceImpl{
		logger:                              logger,
		cdHandler:                           cdHandler,
		imageTaggingService:                 imageTaggingService,
		pipelineRepository:                  pipelineRepository,
		dtResourceTaskRunRepository:         dtResourceTaskRunRepository,
		dtResourceSchemaRepository:          dtResourceSchemaRepository,
		dtResourceReadService:               dtResourceReadService,
		deployedConfigurationHistoryService: deployedConfigurationHistoryService,
	}
}

func (impl *DeploymentHistoryServiceImpl) GetCdPipelineDeploymentHistory(offset, limit, appId, environmentId,
	pipelineId, filterByReleaseId int) (resp bean.DeploymentHistoryResp, err error) {
	toGetOnlyWfrIds := make([]int, 0)
	wfrIdReleaseIdMap := make(map[int]int)
	releaseIdsForRunSourceData := make([]int, 0)
	toFetchRunnerData := true
	// finding task target identifier resource and schema id assuming it tobe cd -pipeline here
	cdPipelineSchema, err := impl.dtResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(bean2.DevtronResourceCdPipeline.ToString(), "",
		bean2.DevtronResourceVersion1.ToString())
	if err != nil {
		impl.logger.Errorw("error, FindSchemaByKindSubKindAndVersion for cd pipeline", "err", err)
		return resp, err
	}
	runTargetIdentifier := helper.GetTaskRunIdentifier(pipelineId, bean2.OldObjectId, cdPipelineSchema.DevtronResourceId, cdPipelineSchema.Id)
	var deploymentTaskRuns []*repository.DevtronResourceTaskRun
	if filterByReleaseId > 0 { //filtering by release, offset and limit is valid for release runSourced workflow runners only
		//get these runners and get only their data
		releaseSchema, err := impl.dtResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(bean2.DevtronResourceRelease.ToString(), "",
			bean2.DevtronResourceVersionAlpha1.ToString())
		if err != nil {
			impl.logger.Errorw("error, FindSchemaByKindSubKindAndVersion for release", "err", err)
			return resp, err
		}
		runSourceIdentifier := helper.GetTaskRunSourceIdentifier(filterByReleaseId, bean2.ResourceObjectIdType, releaseSchema.DevtronResourceId, releaseSchema.Id)
		deploymentTaskRuns, err = impl.dtResourceTaskRunRepository.GetByRunSourceTargetAndTaskTypes(runSourceIdentifier,
			runTargetIdentifier, bean2.CdPipelineAllDeploymentTaskRuns, offset, limit)
		if err != nil && util2.IsErrNoRows(err) {
			impl.logger.Errorw("error, GetByRunSourceTargetAndTaskTypes", "err", err, "runSourceIdentifier", runSourceIdentifier, "runTargetIdentifier", runTargetIdentifier)
			return resp, err
		}
		for _, deploymentTaskRun := range deploymentTaskRuns {
			toGetOnlyWfrIds = append(toGetOnlyWfrIds, deploymentTaskRun.TaskTypeIdentifier)
		}
		toFetchRunnerData = len(toGetOnlyWfrIds) != 0
	}
	var wfs []bean3.CdWorkflowWithArtifact

	if toFetchRunnerData {
		wfs, err = impl.cdHandler.GetCdBuildHistory(appId, environmentId, pipelineId, toGetOnlyWfrIds, offset, limit)
		if err != nil {
			impl.logger.Errorw("service err, List", "err", err, "appId", appId, "environmentId", environmentId, "pipelineId", pipelineId, "offset", offset)
			return resp, err
		}
		if filterByReleaseId == 0 && len(wfs) != 0 { //not filtering by release, to get run source data for result workflows only for optimised processing
			//not getting runners for a specific release, have to get runSource of the above result wfRunners
			filteredWfrIds := make([]int, 0, len(wfs))
			for _, wf := range wfs {
				filteredWfrIds = append(filteredWfrIds, wf.Id)
			}
			deploymentTaskRuns, err = impl.dtResourceTaskRunRepository.GetByTaskTypeAndIdentifiers(filteredWfrIds, bean2.CdPipelineAllDeploymentTaskRuns)
			if err != nil && util2.IsErrNoRows(err) {
				impl.logger.Errorw("error, GetByRunSourceTargetAndTaskTypes", "err", err, "filteredWfrIds", filteredWfrIds)
				return resp, err
			}

		}
		for _, deploymentTaskRun := range deploymentTaskRuns {
			//decode runSourceIdentifier
			identifier, err := util.DecodeTaskRunSourceIdentifier(deploymentTaskRun.RunSourceIdentifier)
			if err != nil {
				return resp, err
			}
			wfrIdReleaseIdMap[deploymentTaskRun.TaskTypeIdentifier] = identifier.Id
			releaseIdsForRunSourceData = append(releaseIdsForRunSourceData, identifier.Id)
		}

		if len(releaseIdsForRunSourceData) > 0 {
			runSourceMap, err := impl.dtResourceReadService.GetTaskRunSourceInfoForReleases(releaseIdsForRunSourceData)
			if err != nil {
				impl.logger.Errorw("error, GetTaskRunSourceInfoForReleases", "err", err, "releaseIds", releaseIdsForRunSourceData)
				return resp, err
			}
			for i := range wfs {
				wfrId := wfs[i].Id
				releaseId := wfrIdReleaseIdMap[wfrId]
				runSource := runSourceMap[releaseId]
				wfs[i].RunSource = runSource
			}
		}
	}

	resp.CdWorkflows = wfs
	appTags, err := impl.imageTaggingService.GetUniqueTagsByAppId(appId)
	if err != nil {
		impl.logger.Errorw("service err, GetTagsByAppId", "err", err, "appId", appId)
		return resp, err
	}
	resp.AppReleaseTagNames = appTags

	prodEnvExists, err := impl.imageTaggingService.GetProdEnvByCdPipelineId(pipelineId)
	resp.TagsEditable = prodEnvExists
	resp.HideImageTaggingHardDelete = impl.imageTaggingService.GetImageTaggingServiceConfig().HideImageTaggingHardDelete
	if err != nil {
		impl.logger.Errorw("service err, GetProdEnvFromParentAndLinkedWorkflow", "err", err, "cdPipelineId", pipelineId)
		return resp, err
	}
	return resp, nil
}

func (impl *DeploymentHistoryServiceImpl) GetCdPipelineDeploymentHistoryConfigList(baseConfigurationId, pipelineId int,
	historyComponent, historyComponentName string) (resp []*history.DeployedHistoryComponentMetadataDto, err error) {
	res, err := impl.deployedConfigurationHistoryService.GetDeployedHistoryComponentList(pipelineId, baseConfigurationId, historyComponent, historyComponentName)
	if err != nil {
		impl.logger.Errorw("service err, GetDeployedHistoryComponentList", "err", err, "pipelineId", pipelineId)
		return
	}
	respWfrIds := make([]int, 0, len(res))
	for _, r := range res {
		respWfrIds = append(respWfrIds, r.WfrId)
	}
	//getting run source for the response wfrIds
	deploymentTaskRuns, err := impl.dtResourceTaskRunRepository.GetByTaskTypeAndIdentifiers(respWfrIds, bean2.CdPipelineAllDeploymentTaskRuns)
	if err != nil && util2.IsErrNoRows(err) {
		impl.logger.Errorw("error, GetByRunSourceTargetAndTaskTypes", "err", err, "respWfrIds", respWfrIds)
		return resp, err
	}
	wfrIdReleaseIdMap := make(map[int]int)
	releaseIdsForRunSourceData := make([]int, 0)
	for _, deploymentTaskRun := range deploymentTaskRuns {
		//decode runSourceIdentifier
		identifier, err := util.DecodeTaskRunSourceIdentifier(deploymentTaskRun.RunSourceIdentifier)
		if err != nil {
			return resp, err
		}
		wfrIdReleaseIdMap[deploymentTaskRun.TaskTypeIdentifier] = identifier.Id
		releaseIdsForRunSourceData = append(releaseIdsForRunSourceData, identifier.Id)
	}

	if len(releaseIdsForRunSourceData) > 0 {
		runSourceMap, err := impl.dtResourceReadService.GetTaskRunSourceInfoForReleases(releaseIdsForRunSourceData)
		if err != nil {
			impl.logger.Errorw("error, GetTaskRunSourceInfoForReleases", "err", err, "releaseIds", releaseIdsForRunSourceData)
			return resp, err
		}
		for i := range res {
			wfrId := res[i].WfrId
			releaseId := wfrIdReleaseIdMap[wfrId]
			runSource := runSourceMap[releaseId]
			res[i].RunSource = runSource
		}
	}
	return res, nil
}
