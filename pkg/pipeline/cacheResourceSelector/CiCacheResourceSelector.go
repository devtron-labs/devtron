package cacheResourceSelector

import (
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/enterprise/pkg/expressionEvaluators"
	"github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"strings"
	"sync"
)

type CiCacheResourceSelector interface {
	GetAvailResource(scope resourceQualifiers.Scope, appLabels map[string]string, pvcResolverExpression, mountPathResolverExpression string, ciWorkflowId int) (*CiCacheSelector, error)
	UpdateResourceStatus(ciWorkflowId int, status string) bool
}

type CiCacheResourceSelectorImpl struct {
	logger                *zap.SugaredLogger
	celEvalService        expressionEvaluators.CELEvaluatorService
	k8sApplicationService application.K8sApplicationService
	config                *Config

	// locked
	synced               bool
	resourcesStatus      map[string]ResourceStatus
	ciWorkflowPVCMapping map[int]string
	lock                 *sync.RWMutex
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

func (impl *CiCacheResourceSelectorImpl) GetAvailResource(scope resourceQualifiers.Scope, appLabels map[string]string, pvcResolverExpression, mountPathResolverExpression string, ciWorkflowId int) (ciCacheSelector *CiCacheSelector, err error) {
	if !impl.isPVCStatusSynced() {
		// no error but not synced yet. callers should handle accordingly
		return nil, nil
	}

	pvcPrefix, mountPath, err := impl.computePVCAndMountPath(appLabels, pvcResolverExpression, mountPathResolverExpression)
	if err != nil {
		impl.logger.Errorw("error in evaluating cel expression for resolving pvc prefix name and mount path", "appLabels", appLabels, "pvcResolverExpression", pvcResolverExpression, "mountPathResolverExpression", mountPathResolverExpression, "err", err)
		return nil, err
	}

	autoSelectedPvc := impl.autoSelectAvailablePrefixMatchingPVCForWorkflow(ciWorkflowId, pvcPrefix)
	if autoSelectedPvc == nil {
		// no error. all the pvc's are busy. callers should handle accordingly
		return nil, nil
	}

	return &CiCacheSelector{
		PVCName:   *autoSelectedPvc,
		MountPath: mountPath,
	}, nil

}

// UpdateResourceStatus
// status: make pvc available
// errored
// failed
// completed
// status: make pvc unavailable
// running
// pending
func (impl *CiCacheResourceSelectorImpl) UpdateResourceStatus(ciWorkflowId int, status string) bool {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	if pvc, ok := impl.ciWorkflowPVCMapping[ciWorkflowId]; ok {
		if status == string(v1alpha1.NodePending) || status == string(v1alpha1.NodeRunning) {
			// 	pvc is busy, no need to update
		} else {
			// pvc got free
			impl.resourcesStatus[pvc] = AvailableResourceStatus
			return true
		}
	}
	return false
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

func (impl *CiCacheResourceSelectorImpl) autoSelectAvailablePrefixMatchingPVCForWorkflow(ciWorkflowId int, pvcPrefix string) *string {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	for pvc, status := range impl.resourcesStatus {
		if status == AvailableResourceStatus && strings.HasPrefix(pvc, pvcPrefix) {
			pvcCopy := pvc
			impl.resourcesStatus[pvc] = UnAvailableResourceStatus
			impl.ciWorkflowPVCMapping[ciWorkflowId] = pvc
			return &pvcCopy
		}
	}

	return nil
}

func (impl *CiCacheResourceSelectorImpl) computePVCAndMountPath(appLabels map[string]string, pvcResolverExpression, mountPathResolverExpression string) (pvcPrefix string, mountPath string, err error) {
	params := []expressionEvaluators.ExpressionParam{{
		ParamName: expressionEvaluators.AppLabels,
		Value:     appLabels,
		Type:      expressionEvaluators.ParamTypeStringMap,
	},
	}

	expressionMetadata := expressionEvaluators.CELRequest{
		Expression: pvcResolverExpression,
		ExpressionMetadata: expressionEvaluators.ExpressionMetadata{
			Params: params,
		},
	}

	ok := false
	pvcPrefixIf, err := impl.celEvalService.EvaluateCELForObject(expressionMetadata)
	if pvcPrefix, ok = pvcPrefixIf.(string); !ok {
		err = errors.New("invalid object resolved for pvc prefix name, expected: string")
	}

	if err != nil {
		impl.logger.Errorw("error in evaluating cel expression for resolving pvc name", "err", err)
		return
	}

	expressionMetadata.Expression = mountPathResolverExpression
	mountPathIf, err := impl.celEvalService.EvaluateCELForObject(expressionMetadata)
	if err != nil {
		impl.logger.Errorw("error in evaluating cel expression for resolving pvc name", "err", err)
		return
	}
	if mountPath, ok = mountPathIf.(string); !ok {
		err = errors.New("invalid object resolved for pvc mount path, expected: string")
		return
	}
	return
}

// computePVCStatusMap computes pvc's statuses.
// if pod is in running state , then the pvc held by the pod is considered unavailable.
// else available
func computePVCStatusMap(pods []v1.Pod) map[string]ResourceStatus {
	statusMap := make(map[string]ResourceStatus)
	for _, pod := range pods {
		for key, val := range pod.Labels {
			if key == BuildPVCLabel {
				if pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodPending {
					statusMap[val] = UnAvailableResourceStatus
				} else {
					statusMap[val] = AvailableResourceStatus
				}
			}
		}
	}
	return statusMap
}
