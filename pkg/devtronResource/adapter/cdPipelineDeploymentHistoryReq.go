package adapter

import (
	apiBean "github.com/devtron-labs/devtron/api/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean/history"
)

func GetCDDeploymentHistoryListReq(queryParams *apiBean.GetHistoryQueryParams,
	decodedReq *bean.DeploymentHistoryGetReqDecoderBean) *history.CdPipelineDeploymentHistoryListReq {
	resp := &history.CdPipelineDeploymentHistoryListReq{}
	if queryParams != nil {
		resp.Offset = queryParams.OffSet
		resp.Limit = queryParams.Limit
	}
	if decodedReq != nil {
		resp.AppId = decodedReq.AppId
		resp.EnvId = decodedReq.EnvId
		resp.PipelineId = decodedReq.PipelineId
	}
	return resp
}

func GetCDDeploymentHistoryConfigListReq(queryParams *apiBean.GetHistoryConfigQueryParams,
	decodedReq *bean.DeploymentHistoryGetReqDecoderBean) *history.CdPipelineDeploymentHistoryConfigListReq {
	resp := &history.CdPipelineDeploymentHistoryConfigListReq{}
	if queryParams != nil {
		resp.BaseConfigurationId = queryParams.BaseConfigurationId
		resp.HistoryComponent = queryParams.HistoryComponent
		resp.HistoryComponentName = queryParams.HistoryComponentName
	}
	if decodedReq != nil {
		resp.PipelineId = decodedReq.PipelineId
	}
	return resp
}
