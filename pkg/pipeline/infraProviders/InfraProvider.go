/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package infraProviders

import (
	"github.com/devtron-labs/devtron/pkg/infraConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/infraProviders/infraGetters"
	"github.com/devtron-labs/devtron/pkg/pipeline/infraProviders/infraGetters/ciPipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/infraProviders/infraGetters/job"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type InfraProvider interface {
	GetInfraProvider(providerType bean.WorkflowPipelineType) (infraGetters.InfraGetter, error)
}

type InfraProviderImpl struct {
	logger         *zap.SugaredLogger
	ciInfraGetter  infraGetters.InfraGetter
	jobInfraGetter infraGetters.InfraGetter
}

func NewInfraProviderImpl(logger *zap.SugaredLogger, service infraConfig.InfraConfigService) *InfraProviderImpl {
	return &InfraProviderImpl{
		logger:         logger,
		ciInfraGetter:  ciPipeline.NewCiInfraGetter(service),
		jobInfraGetter: job.NewJobInfraGetter(),
	}
}

func (infraProvider *InfraProviderImpl) GetInfraProvider(providerType bean.WorkflowPipelineType) (infraGetters.InfraGetter, error) {
	switch providerType {
	case bean.CI_WORKFLOW_PIPELINE_TYPE:
		return infraProvider.ciInfraGetter, nil
	case bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return infraProvider.jobInfraGetter, nil
	default:
		return nil, errors.New("Invalid workflow pipeline type")
	}
}
