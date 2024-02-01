package deploymentTemplate

import "go.uber.org/zap"

type DeploymentTemplateService interface {
}

type DeploymentTemplateServiceImpl struct {
	logger *zap.SugaredLogger
}

func NewDeploymentTemplateServiceImpl(logger *zap.SugaredLogger) *DeploymentTemplateServiceImpl {
	return &DeploymentTemplateServiceImpl{
		logger: logger,
	}
}
