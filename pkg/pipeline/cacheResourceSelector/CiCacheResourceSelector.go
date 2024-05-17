package cacheResourceSelector

import (
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/enterprise/pkg/expressionEvaluators"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"go.uber.org/zap"
)

type CiCacheResourceSelector interface {
	GetAvailResource(scope resourceQualifiers.Scope) (string, string, error)
	UpdateResourceStatus(resourceName string, status ResourceStatus) error
}

type CiCacheResourceSelectorImpl struct {
	logger          *zap.SugaredLogger
	celEvalService  expressionEvaluators.CELEvaluatorService
	resourcesStatus map[string]ResourceStatus
	config          *Config
}

func NewCiCacheResourceSelectorImpl(logger *zap.SugaredLogger, celEvalService expressionEvaluators.CELEvaluatorService) *CiCacheResourceSelectorImpl {
	config := &Config{}
	err := env.Parse(config)
	if err != nil {
		logger.Fatalw("failed to load cache selector config", "err", err)
	}
	resourcesStatus := make(map[string]ResourceStatus)
	selectorImpl := &CiCacheResourceSelectorImpl{
		logger:          logger,
		celEvalService:  celEvalService,
		config:          config,
		resourcesStatus: resourcesStatus,
	}
	go selectorImpl.updateCacheResourceStatus()
	return selectorImpl
}

func (impl *CiCacheResourceSelectorImpl) GetAvailResource(scope resourceQualifiers.Scope) (string, string, error) {

}

func (impl *CiCacheResourceSelectorImpl) UpdateResourceStatus(resourceName string, status ResourceStatus) error {
	//TODO implement me
	panic("implement me")
}
func (impl *CiCacheResourceSelectorImpl) updateCacheResourceStatus() {

}
