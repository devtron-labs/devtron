/*
 * Copyright (c) 2024. Devtron Inc.
 */

package infraProviders

import (
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/infraProviders/infraGetters"
	"github.com/devtron-labs/devtron/pkg/pipeline/infraProviders/infraGetters/ci"
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

func NewInfraProviderImpl(logger *zap.SugaredLogger,
	jobInfraGetter *job.InfraGetter,
	ciInfraGetter *ci.InfraGetter) *InfraProviderImpl {
	return &InfraProviderImpl{
		logger:         logger,
		ciInfraGetter:  ciInfraGetter,
		jobInfraGetter: jobInfraGetter,
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
