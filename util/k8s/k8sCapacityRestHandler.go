package k8s

import (
	"go.uber.org/zap"
)

type K8sCapacityRestHandler interface {
}
type K8sCapacityRestHandlerImpl struct {
	logger             *zap.SugaredLogger
	k8sCapacityService K8sCapacityService
}

func NewK8sCapacityRestHandlerImpl(logger *zap.SugaredLogger,
	k8sCapacityService K8sCapacityService) *K8sCapacityRestHandlerImpl {
	return &K8sCapacityRestHandlerImpl{
		logger:             logger,
		k8sCapacityService: k8sCapacityService,
	}
}
