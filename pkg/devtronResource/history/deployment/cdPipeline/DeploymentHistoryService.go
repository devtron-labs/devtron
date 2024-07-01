package cdPipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	historyBean "github.com/devtron-labs/devtron/pkg/devtronResource/bean/history"
	"github.com/devtron-labs/devtron/pkg/pipeline"
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
	deployedConfigurationHistoryService history.DeployedConfigurationHistoryService
}

func NewDeploymentHistoryServiceImpl(logger *zap.SugaredLogger,
	cdHandler pipeline.CdHandler,
	imageTaggingService pipeline.ImageTaggingService,
	pipelineRepository pipelineConfig.PipelineRepository,
	deployedConfigurationHistoryService history.DeployedConfigurationHistoryService) *DeploymentHistoryServiceImpl {
	return &DeploymentHistoryServiceImpl{
		logger:                              logger,
		cdHandler:                           cdHandler,
		imageTaggingService:                 imageTaggingService,
		pipelineRepository:                  pipelineRepository,
		deployedConfigurationHistoryService: deployedConfigurationHistoryService,
	}
}

func (impl *DeploymentHistoryServiceImpl) GetCdPipelineDeploymentHistory(req *historyBean.CdPipelineDeploymentHistoryListReq) (resp historyBean.DeploymentHistoryResp, err error) {
	var wfs []pipelineConfig.CdWorkflowWithArtifact
	wfs, err = impl.cdHandler.GetCdBuildHistory(req.AppId, req.EnvId, req.PipelineId, req.Offset, req.Limit)
	if err != nil {
		impl.logger.Errorw("service err, List", "err", err, "req", req)
		return resp, err
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
	return res, nil
}
