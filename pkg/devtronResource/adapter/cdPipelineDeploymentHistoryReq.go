/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

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
