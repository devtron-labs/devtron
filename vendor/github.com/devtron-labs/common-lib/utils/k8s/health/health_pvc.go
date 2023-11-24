package health

import (
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func getPVCHealth(obj *unstructured.Unstructured) (*HealthStatus, error) {
	gvk := obj.GroupVersionKind()
	switch gvk {
	case corev1.SchemeGroupVersion.WithKind(commonBean.PersistentVolumeClaimKind):
		var pvc corev1.PersistentVolumeClaim
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &pvc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert unstructured PersistentVolumeClaim to typed: %v", err)
		}
		return getCorev1PVCHealth(&pvc)
	default:
		return nil, fmt.Errorf("unsupported PersistentVolumeClaim GVK: %s", gvk)
	}
}

func getCorev1PVCHealth(pvc *corev1.PersistentVolumeClaim) (*HealthStatus, error) {
	var status HealthStatusCode
	switch pvc.Status.Phase {
	case corev1.ClaimLost:
		status = HealthStatusDegraded
	case corev1.ClaimPending:
		status = HealthStatusProgressing
	case corev1.ClaimBound:
		status = HealthStatusHealthy
	default:
		status = HealthStatusUnknown
	}
	return &HealthStatus{Status: status}, nil
}
