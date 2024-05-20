package cacheResourceSelector

import (
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/enterprise/pkg/expressionEvaluators"
	"github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"sync"
)

type CiCacheResourceSelector interface {
	GetAvailResource(scope resourceQualifiers.Scope) (string, string, error)
	UpdateResourceStatus(resourceName string, status ResourceStatus) error
}

type CiCacheResourceSelectorImpl struct {
	logger                *zap.SugaredLogger
	celEvalService        expressionEvaluators.CELEvaluatorService
	k8sApplicationService application.K8sApplicationService
	config                *Config

	// locked
	synced          bool
	resourcesStatus map[string]ResourceStatus
	lock            *sync.RWMutex
}

func NewCiCacheResourceSelectorImpl(logger *zap.SugaredLogger, celEvalService expressionEvaluators.CELEvaluatorService, k8sApplicationService application.K8sApplicationService) *CiCacheResourceSelectorImpl {
	config := &Config{}
	err := env.Parse(config)
	if err != nil {
		logger.Fatalw("failed to load cache selector config", "err", err)
	}
	resourcesStatus := make(map[string]ResourceStatus)
	selectorImpl := &CiCacheResourceSelectorImpl{
		logger:                logger,
		celEvalService:        celEvalService,
		k8sApplicationService: k8sApplicationService,
		config:                config,
		resourcesStatus:       resourcesStatus,
	}
	go selectorImpl.updateCacheResourceStatus()
	return selectorImpl
}

func (impl *CiCacheResourceSelectorImpl) GetAvailResource(scope resourceQualifiers.Scope) (name string, path string, err error) {
	if !impl.isPVCStatusSynced() {
		// no error but not synced yet. callers should handle accordingly
		return "", "", nil
	}

	autoSelectedPvc := impl.autoSelectAvailablePVC()
	if autoSelectedPvc == nil {
		// no error. all the pvs are busy callers should handle accordingly
		return "", "", nil
	}

	name = *autoSelectedPvc
	// TODO: compute path from cel expression

	return

}

func (impl *CiCacheResourceSelectorImpl) UpdateResourceStatus(resourceName string, status ResourceStatus) error {
	// TODO implement me
	panic("implement me")
}

// updateCacheResourceStatus triggers at service startup
// fetch all the active ci's that are running and get the list of pvc's
// fetch the above list from k8s using label selector.
func (impl *CiCacheResourceSelectorImpl) updateCacheResourceStatus() {
	// getting the pod list for default cluster and devtron-ci namespace
	// TODO: might need external env support as well.
	pods, err := impl.k8sApplicationService.GetPodListByLabel(1, "devtron-ci", BuildPVCLabel)
	if err != nil {
		// 	log error
		impl.logger.Errorw("error in getting pods with label selector", "clusterId", 1, "namespace", "devtron-ci", "labelSelector", BuildPVCLabel)
		return
	}

	impl.lock.Lock()
	defer impl.lock.Unlock()
	// build status map
	statusMap := computePVCStatusMap(pods)
	impl.resourcesStatus = statusMap
	impl.synced = true
}

func (impl *CiCacheResourceSelectorImpl) isPVCStatusSynced() bool {
	impl.lock.RLock()
	defer impl.lock.RUnlock()
	synced := impl.synced
	return synced
}

func (impl *CiCacheResourceSelectorImpl) autoSelectAvailablePVC() *string {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	for pvc, status := range impl.resourcesStatus {
		if status == AvailableResourceStatus {
			pvcCopy := pvc
			return &pvcCopy
		}
	}
	return nil
}

// computePVCStatusMap computes pvc's statuses.
// if pod is in running state , then the pvc held by the pod is considered unavailable.
// else available
func computePVCStatusMap(pods []v1.Pod) map[string]ResourceStatus {
	statusMap := make(map[string]ResourceStatus)
	for _, pod := range pods {
		for key, val := range pod.Labels {
			if key == BuildPVCLabel {
				if pod.Status.Phase == v1.PodRunning {
					statusMap[val] = UnAvailableResourceStatus
				} else {
					statusMap[val] = AvailableResourceStatus
				}
			}
		}
	}
	return statusMap
}
