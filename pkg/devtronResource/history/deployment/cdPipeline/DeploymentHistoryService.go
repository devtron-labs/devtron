package cdPipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	historyBean "github.com/devtron-labs/devtron/pkg/devtronResource/bean/history"
	"github.com/devtron-labs/devtron/pkg/devtronResource/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/read"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/devtron-labs/devtron/pkg/devtronResource/util"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"go.uber.org/zap"
)

type DeploymentHistoryService interface {
	GetCdPipelineDeploymentHistory(req *historyBean.CdPipelineDeploymentHistoryListReq) (resp historyBean.DeploymentHistoryResp, err error)
	GetCdPipelineDeploymentHistoryConfigList(req *historyBean.CdPipelineDeploymentHistoryConfigListReq) (resp []*history.DeployedHistoryComponentMetadataDto, err error)
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

func (impl *DeploymentHistoryServiceImpl) GetCdPipelineDeploymentHistory(req *historyBean.CdPipelineDeploymentHistoryListReq) (resp historyBean.DeploymentHistoryResp, err error) {
	toGetOnlyWfrIds := make([]int, 0, req.Limit)
	wfrIdReleaseIdMap := make(map[int]int, req.Limit)
	releaseIdsForRunSourceData := make([]int, 0, req.Limit)
	toFetchRunnerData := true
	// finding task target identifier resource and schema id assuming it tobe cd -pipeline here
	cdPipelineSchema, err := impl.dtResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(bean2.DevtronResourceCdPipeline.ToString(), "",
		bean2.DevtronResourceVersion1.ToString())
	if err != nil {
		impl.logger.Errorw("error, FindSchemaByKindSubKindAndVersion for cd pipeline", "err", err)
		return resp, err
	}
	runTargetIdentifier := helper.GetTaskRunIdentifier(req.PipelineId, bean2.OldObjectId, cdPipelineSchema.DevtronResourceId, cdPipelineSchema.Id)
	var deploymentTaskRuns []*repository.DevtronResourceTaskRun
	if req.FilterByReleaseId > 0 { //filtering by release, offset and limit is valid for release runSourced workflow runners only
		//get these runners and get only their data
		releaseSchema, err := impl.dtResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(bean2.DevtronResourceRelease.ToString(), "",
			bean2.DevtronResourceVersionAlpha1.ToString())
		if err != nil {
			impl.logger.Errorw("error, FindSchemaByKindSubKindAndVersion for release", "err", err)
			return resp, err
		}
		runSourceIdentifier := helper.GetTaskRunSourceIdentifier(req.FilterByReleaseId, bean2.ResourceObjectIdType, releaseSchema.DevtronResourceId, releaseSchema.Id)
		deploymentTaskRuns, err = impl.dtResourceTaskRunRepository.GetByRunSourceTargetAndTaskTypes(runSourceIdentifier,
			runTargetIdentifier, bean2.CdPipelineAllDeploymentTaskRuns, req.Offset, req.Limit)
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
		wfs, err = impl.cdHandler.GetCdBuildHistory(req.AppId, req.EnvId, req.PipelineId, toGetOnlyWfrIds, req.Offset, req.Limit)
		if err != nil {
			impl.logger.Errorw("service err, List", "err", err, "req", req)
			return resp, err
		}
		if req.FilterByReleaseId == 0 && len(wfs) != 0 { //not filtering by release, to get run source data for result workflows only for optimised processing
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
	appTags, err := impl.imageTaggingService.GetUniqueTagsByAppId(req.AppId)
	if err != nil {
		impl.logger.Errorw("service err, GetTagsByAppId", "err", err, "appId", req.AppId)
		return resp, err
	}
	resp.AppReleaseTagNames = appTags

	prodEnvExists, err := impl.imageTaggingService.GetProdEnvByCdPipelineId(req.PipelineId)
	resp.TagsEditable = prodEnvExists
	resp.HideImageTaggingHardDelete = impl.imageTaggingService.GetImageTaggingServiceConfig().HideImageTaggingHardDelete
	if err != nil {
		impl.logger.Errorw("service err, GetProdEnvFromParentAndLinkedWorkflow", "err", err, "cdPipelineId", req.PipelineId)
		return resp, err
	}
	return resp, nil
}

func (impl *DeploymentHistoryServiceImpl) GetCdPipelineDeploymentHistoryConfigList(req *historyBean.CdPipelineDeploymentHistoryConfigListReq) (resp []*history.DeployedHistoryComponentMetadataDto, err error) {
	res, err := impl.deployedConfigurationHistoryService.GetDeployedHistoryComponentList(req.PipelineId, req.BaseConfigurationId, req.HistoryComponent, req.HistoryComponentName)
	if err != nil {
		impl.logger.Errorw("service err, GetDeployedHistoryComponentList", "err", err, "pipelineId", req.PipelineId)
		return nil, err
	}
	respWfrIds := make([]int, 0, len(res))
	for _, r := range res {
		respWfrIds = append(respWfrIds, r.WfrId)
	}
	//getting run source for the response wfrIds
	deploymentTaskRuns, err := impl.dtResourceTaskRunRepository.GetByTaskTypeAndIdentifiers(respWfrIds, bean2.CdPipelineAllDeploymentTaskRuns)
	if err != nil && util2.IsErrNoRows(err) {
		impl.logger.Errorw("error, GetByRunSourceTargetAndTaskTypes", "err", err, "respWfrIds", respWfrIds)
		return nil, err
	}
	wfrIdReleaseIdMap := make(map[int]int, len(deploymentTaskRuns))
	releaseIdsForRunSourceData := make([]int, 0, len(deploymentTaskRuns))
	for _, deploymentTaskRun := range deploymentTaskRuns {
		//decode runSourceIdentifier
		identifier, err := util.DecodeTaskRunSourceIdentifier(deploymentTaskRun.RunSourceIdentifier)
		if err != nil {
			return nil, err
		}
		wfrIdReleaseIdMap[deploymentTaskRun.TaskTypeIdentifier] = identifier.Id
		releaseIdsForRunSourceData = append(releaseIdsForRunSourceData, identifier.Id)
	}

	if len(releaseIdsForRunSourceData) > 0 {
		runSourceMap, err := impl.dtResourceReadService.GetTaskRunSourceInfoForReleases(releaseIdsForRunSourceData)
		if err != nil {
			impl.logger.Errorw("error, GetTaskRunSourceInfoForReleases", "err", err, "releaseIds", releaseIdsForRunSourceData)
			return nil, err
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
