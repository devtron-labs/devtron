package deploymentTemplate

import "go.uber.org/zap"

type DeploymentTemplateValidationService interface {
}

type DeploymentTemplateValidationServiceImpl struct {
	logger *zap.SugaredLogger
}

func NewDeploymentTemplateValidationServiceImpl(logger *zap.SugaredLogger) *DeploymentTemplateValidationServiceImpl {
	return &DeploymentTemplateValidationServiceImpl{
		logger: logger,
	}
}
