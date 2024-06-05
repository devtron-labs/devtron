package cacheResourceSelector

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/caarlos0/env"
	k8s2 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/enterprise/pkg/expressionEvaluators"
	"github.com/devtron-labs/devtron/pkg/k8s"
	"github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strconv"
	"strings"
	"sync"
)

type CiCacheResourceSelector interface {
	GetAvailResource(scope resourceQualifiers.Scope, appLabels map[string]string, ciWorkflowId int) (*CiCacheResource, error)
	UpdateResourceStatus(ciWorkflowId int, podName string, namespace string, status string) bool
}

type CiCacheResourceSelectorImpl struct {
	logger                *zap.SugaredLogger
	celEvalService        expressionEvaluators.CELEvaluatorService
	k8sApplicationService application.K8sApplicationService
	k8sCommonService      k8s.K8sCommonService
	config                *Config

	// locked
	synced               bool
	resourcesStatus      map[string]ResourceStatus
	ciWorkflowPVCMapping map[int]string
	lock                 *sync.RWMutex
}

func NewCiCacheResourceSelectorImpl(logger *zap.SugaredLogger, celEvalService expressionEvaluators.CELEvaluatorService, k8sApplicationService application.K8sApplicationService,
	k8sCommonService k8s.K8sCommonService) *CiCacheResourceSelectorImpl {
	config := &Config{}
	err := env.Parse(config)
	if err != nil {
		logger.Fatalw("failed to load cache selector config", "err", err)
	}
	resourcesStatus := make(map[string]ResourceStatus)
	ciWorkflowPVCMapping := make(map[int]string)
	selectorImpl := &CiCacheResourceSelectorImpl{
		logger:                logger,
		celEvalService:        celEvalService,
		k8sApplicationService: k8sApplicationService,
		k8sCommonService:      k8sCommonService,
		config:                config,
		resourcesStatus:       resourcesStatus,
		ciWorkflowPVCMapping:  ciWorkflowPVCMapping,
		lock:                  &sync.RWMutex{},
	}
	go selectorImpl.updateCacheResourceStatus()
	return selectorImpl
}

func (impl *CiCacheResourceSelectorImpl) GetAvailResource(scope resourceQualifiers.Scope, appLabels map[string]string, ciWorkflowId int) (ciCacheResource *CiCacheResource, err error) {
	if !impl.isPVCStatusSynced() {
		// no error but not synced yet. callers should handle accordingly
		return nil, nil
	}
	if len(appLabels) == 0 {
		return nil, nil
	}

	pvcResolverExpression := impl.config.PVCNameExpression
	mountPathResolverExpression := impl.config.MountPathExpression
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

	return &CiCacheResource{
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
func (impl *CiCacheResourceSelectorImpl) UpdateResourceStatus(ciWorkflowId int, podName string, namespace string, status string) bool {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	if pvc, ok := impl.ciWorkflowPVCMapping[ciWorkflowId]; ok {
		if status == string(v1alpha1.NodePending) || status == string(v1alpha1.NodeRunning) {
			// 	pvc is busy, no need to update
		} else {
			// do not have to clean-up pod in success/completed state as k8s itself frees the resources that are in hold of this pod.
			if status == string(v1alpha1.NodeError) || status == string(v1alpha1.NodeFailed) {
				impl.cleanupPod(podName, namespace)
			}

			// pvc got free
			if status == string(v1alpha1.NodeError) || status == string(v1alpha1.NodeFailed) {
				impl.cleanupPod(podName, namespace)
			}
			impl.resourcesStatus[pvc] = AvailableResourceStatus
			// TODO KB: run a particular command to make PVC unavailable
			impl.cleanupResources(pvc)
			return true
		}
	}
	return false
}

func (impl *CiCacheResourceSelectorImpl) cleanupPod(podName, namespace string) {
	resourceRequestBean := &k8s.ResourceRequestBean{
		ClusterId: 1,
		K8sRequest: &k8s2.K8sRequestBean{
			ResourceIdentifier: k8s2.ResourceIdentifier{
				GroupVersionKind: schema.GroupVersionKind{
					Version: "v1",
					Kind:    "Pod",
				},
				Name:      podName,
				Namespace: namespace,
			},
			ForceDelete: true,
		},
	}
	impl.k8sCommonService.DeleteResource(context.Background(), resourceRequestBean)
}

func (impl *CiCacheResourceSelectorImpl) cleanupResources(pvcName string) {
	pvName, done := impl.getPVName(pvcName)
	if done {
		return
	}

	resourceRequestBean := &k8s.ResourceRequestBean{
		ClusterId: 1,
		K8sRequest: &k8s2.K8sRequestBean{
			ResourceIdentifier: k8s2.ResourceIdentifier{
				GroupVersionKind: schema.GroupVersionKind{
					Group:   "storage.k8s.io",
					Version: "v1",
					Kind:    "VolumeAttachment",
				},
			},
		},
		Filter: fmt.Sprintf("self.spec.source.persistentVolumeName == '%s'", pvName),
	}
	resourceList, err := impl.k8sApplicationService.GetResourceList(context.Background(), "", resourceRequestBean, func(token string, clusterName string, request k8s.ResourceRequestBean, casbinAction string) bool {
		return true
	})
	if err != nil {
		return
	}
	volAttachData := resourceList.Data
	if len(volAttachData) > 0 {
		volAttachName := volAttachData[0]["name"].(string)
		volAttachDelRequest := &k8s.ResourceRequestBean{
			ClusterId: 1,
			K8sRequest: &k8s2.K8sRequestBean{
				ResourceIdentifier: k8s2.ResourceIdentifier{
					Name: volAttachName,
					GroupVersionKind: schema.GroupVersionKind{
						Group:   "storage.k8s.io",
						Version: "v1",
						Kind:    "VolumeAttachment",
					},
				},
				ForceDelete: true,
			},
		}
		impl.k8sCommonService.DeleteResource(context.Background(), volAttachDelRequest)
	}
}

func (impl *CiCacheResourceSelectorImpl) getPVName(pvcName string) (string, bool) {
	var pvName string
	var ok bool
	pvcResourceRequest := &k8s.ResourceRequestBean{
		ClusterId: 1,
		K8sRequest: &k8s2.K8sRequestBean{
			ResourceIdentifier: k8s2.ResourceIdentifier{
				Name:      pvcName,
				Namespace: "devtron-ci",
				GroupVersionKind: schema.GroupVersionKind{
					Version: "v1",
					Kind:    "PersistentVolumeClaim",
				},
			},
			ForceDelete: true,
		},
	}
	pvcResource, err := impl.k8sCommonService.GetResource(context.Background(), pvcResourceRequest)
	if err != nil {
		return pvName, true
	}
	manifestResponse := pvcResource.ManifestResponse
	if manifestResponse == nil {
		return pvName, true
	}

	pvcManifestObject := manifestResponse.Manifest.Object
	pvcSpec := pvcManifestObject["spec"].(map[string]interface{})
	if pvName, ok = pvcSpec["volumeName"].(string); !ok {
		return pvName, true
	}

	return pvName, false
}

// updateCacheResourceStatus triggers at service startup
// fetch all the active ci's that are running and get the list of pvc's
// fetch the above list from k8s using label selector.
func (impl *CiCacheResourceSelectorImpl) updateCacheResourceStatus() {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	// getting the pod list for default cluster and devtron-ci namespace
	// TODO: might need external env support as well.
	buildPVCLabelSelectorExpr := fmt.Sprintf("%s=%s", BuildPVCLabelKey1, BuildPVCLabelValue1)
	pods, err := impl.k8sApplicationService.GetPodListByLabel(1, "devtron-ci", buildPVCLabelSelectorExpr)
	if err != nil {
		impl.logger.Errorw("error in getting pods with label selector", "clusterId", 1, "namespace", "devtron-ci", "labelSelector", buildPVCLabelSelectorExpr, "err", err)
		return
	}
	// build status map
	statusMap, ciWorkflowMapping := computePVCStatusMap(pods)
	cachePVCs := impl.config.CachePVCs
	for _, cachePVC := range cachePVCs {
		if _, ok := statusMap[cachePVC]; !ok {
			statusMap[cachePVC] = AvailableResourceStatus
		}
	}
	impl.resourcesStatus = statusMap
	impl.ciWorkflowPVCMapping = ciWorkflowMapping
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
func computePVCStatusMap(pods []v1.Pod) (map[string]ResourceStatus, map[int]string) {
	statusMap := make(map[string]ResourceStatus)
	ciWorkflowMapping := make(map[int]string)
	for _, pod := range pods {
		labels := pod.Labels
		for key, val := range labels {
			if key == BuildPVCLabelKey2 {
				ciWorkflowId := labels[BuildWorkflowId]
				if ciWorkflowIdInt, err := strconv.Atoi(ciWorkflowId); err == nil {
					ciWorkflowMapping[ciWorkflowIdInt] = val
				}
				if pod.Status.Phase == v1.PodRunning || pod.Status.Phase == v1.PodPending {
					statusMap[val] = UnAvailableResourceStatus
				} else {
					statusMap[val] = AvailableResourceStatus
				}
			}
		}
	}
	return statusMap, ciWorkflowMapping
}
