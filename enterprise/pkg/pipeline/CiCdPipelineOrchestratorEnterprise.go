/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
