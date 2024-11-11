package cdPipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/build/artifacts/imageTagging/read"
	historyBean "github.com/devtron-labs/devtron/pkg/devtronResource/bean/history"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/history/bean"
	"go.uber.org/zap"
)

type DeploymentHistoryService interface {
	GetCdPipelineDeploymentHistory(req *historyBean.CdPipelineDeploymentHistoryListReq) (resp historyBean.DeploymentHistoryResp, err error)
	GetCdPipelineDeploymentHistoryConfigList(req *historyBean.CdPipelineDeploymentHistoryConfigListReq) (resp []*bean2.DeployedHistoryComponentMetadataDto, err error)
}

type DeploymentHistoryServiceImpl struct {
	logger                              *zap.SugaredLogger
	cdHandler                           pipeline.CdHandler
	imageTaggingReadService             read.ImageTaggingReadService
	imageTaggingService                 pipeline.ImageTaggingService
	pipelineRepository                  pipelineConfig.PipelineRepository
	deployedConfigurationHistoryService history.DeployedConfigurationHistoryService
}

func NewDeploymentHistoryServiceImpl(logger *zap.SugaredLogger,
	cdHandler pipeline.CdHandler,
	imageTaggingReadService read.ImageTaggingReadService,
	imageTaggingService pipeline.ImageTaggingService,
	pipelineRepository pipelineConfig.PipelineRepository,
	deployedConfigurationHistoryService history.DeployedConfigurationHistoryService) *DeploymentHistoryServiceImpl {
	return &DeploymentHistoryServiceImpl{
		logger:                              logger,
		cdHandler:                           cdHandler,
		imageTaggingReadService:             imageTaggingReadService,
		imageTaggingService:                 imageTaggingService,
		pipelineRepository:                  pipelineRepository,
		deployedConfigurationHistoryService: deployedConfigurationHistoryService,
	}
}

func (impl *DeploymentHistoryServiceImpl) GetCdPipelineDeploymentHistory(req *historyBean.CdPipelineDeploymentHistoryListReq) (resp historyBean.DeploymentHistoryResp, err error) {
	var wfs []bean.CdWorkflowWithArtifact
	wfs, err = impl.cdHandler.GetCdBuildHistory(req.AppId, req.EnvId, req.PipelineId, req.Offset, req.Limit)
	if err != nil {
		impl.logger.Errorw("service err, List", "err", err, "req", req)
		return resp, err
	}

	resp.CdWorkflows = wfs
	appTags, err := impl.imageTaggingReadService.GetUniqueTagsByAppId(req.AppId)
	if err != nil {
		impl.logger.Errorw("service err, GetTagsByAppId", "err", err, "appId", req.AppId)
		return resp, err
	}
	resp.AppReleaseTagNames = appTags

	prodEnvExists, err := impl.imageTaggingService.GetProdEnvByCdPipelineId(req.PipelineId)
	if err != nil {
		impl.logger.Errorw("service err, GetProdEnvByCdPipelineId", "err", err, "cdPipelineId", req.PipelineId)
		return resp, err
	}
	resp.TagsEditable = prodEnvExists
	resp.HideImageTaggingHardDelete = impl.imageTaggingReadService.GetImageTaggingServiceConfig().IsHardDeleteHidden()
	return resp, nil
}

func (impl *DeploymentHistoryServiceImpl) GetCdPipelineDeploymentHistoryConfigList(req *historyBean.CdPipelineDeploymentHistoryConfigListReq) (resp []*bean2.DeployedHistoryComponentMetadataDto, err error) {
	res, err := impl.deployedConfigurationHistoryService.GetDeployedHistoryComponentList(req.PipelineId, req.BaseConfigurationId, req.HistoryComponent, req.HistoryComponentName)
	if err != nil {
		impl.logger.Errorw("service err, GetDeployedHistoryComponentList", "err", err, "pipelineId", req.PipelineId)
		return nil, err
	}
	return res, nil
}
