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
