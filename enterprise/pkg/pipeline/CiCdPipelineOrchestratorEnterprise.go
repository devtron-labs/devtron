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
	"github.com/devtron-labs/devtron/enterprise/pkg/globalTag"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
)

type CiCdPipelineOrchestratorEnterpriseImpl struct {
	globalTagService globalTag.GlobalTagService
	*pipeline.CiCdPipelineOrchestratorImpl
}

func NewCiCdPipelineOrchestratorEnterpriseImpl(
	ciCdPipelineOrchestratorImpl *pipeline.CiCdPipelineOrchestratorImpl,
	globalTagService globalTag.GlobalTagService) *CiCdPipelineOrchestratorEnterpriseImpl {
	return &CiCdPipelineOrchestratorEnterpriseImpl{
		CiCdPipelineOrchestratorImpl: ciCdPipelineOrchestratorImpl,
		globalTagService:             globalTagService,
	}
}

func (impl *CiCdPipelineOrchestratorEnterpriseImpl) CreateApp(createRequest *bean.CreateAppDTO) (*bean.CreateAppDTO, error) {
	// validate mandatory labels against project
	labelsMap := make(map[string]string)
	for _, label := range createRequest.AppLabels {
		labelsMap[label.Key] = label.Value
	}
	err := impl.globalTagService.ValidateMandatoryLabelsForProject(createRequest.TeamId, labelsMap)
	if err != nil {
		return nil, err
	}

	// call forward
	return impl.CiCdPipelineOrchestratorImpl.CreateApp(createRequest)
}
