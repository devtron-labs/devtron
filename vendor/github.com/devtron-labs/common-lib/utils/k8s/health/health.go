/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package health

import (
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Implements custom health assessment that overrides built-in assessment
type HealthOverride interface {
	GetResourceHealth(obj *unstructured.Unstructured) (*HealthStatus, error)
}

// Holds health assessment results
type HealthStatus struct {
	Status  HealthStatusCode `json:"status,omitempty"`
	Message string           `json:"message,omitempty"`
}

// healthOrder is a list of health codes in order of most healthy to least healthy
var healthOrder = []HealthStatusCode{
	HealthStatusHealthy,
	HealthStatusSuspended,
	HealthStatusProgressing,
	HealthStatusMissing,
	HealthStatusDegraded,
	HealthStatusUnknown,
}
var healthOrderMap = map[HealthStatusCode]int{
	HealthStatusHealthy:     0,
	HealthStatusSuspended:   1,
	HealthStatusProgressing: 2,
	HealthStatusMissing:     3,
	HealthStatusDegraded:    4,
	HealthStatusUnknown:     5,
}

// IsWorse returns whether or not the new health status code is a worse condition than the current
func IsWorse(current, new HealthStatusCode) bool {
	if new == HealthStatusHealthy && current == HealthStatusHealthy {
		return false
	} else if current == HealthStatusHealthy {
		return true
	} else {
		currentIndex := 0
		newIndex := 0
		for i, code := range healthOrder {
			if current == code {
				currentIndex = i
			}
			if new == code {
				newIndex = i
			}
		}
		return newIndex > currentIndex
	}
}

func IsWorseStatus(current, new HealthStatusCode) bool {
	return healthOrderMap[new] > healthOrderMap[current]
}

// GetResourceHealth returns the health of a k8s resource
func GetResourceHealth(obj *unstructured.Unstructured, healthOverride HealthOverride) (health *HealthStatus, err error) {
	if obj.GetDeletionTimestamp() != nil {
		return &HealthStatus{
			Status:  HealthStatusProgressing,
			Message: "Pending deletion",
		}, nil
	}

	if healthOverride != nil {
		health, err := healthOverride.GetResourceHealth(obj)
		if err != nil {
			health = &HealthStatus{
				Status:  HealthStatusUnknown,
				Message: err.Error(),
			}
			return health, err
		}
		if health != nil {
			return health, nil
		}
	}

	if healthCheck := GetHealthCheckFunc(obj.GroupVersionKind()); healthCheck != nil {
		if health, err = healthCheck(obj); err != nil {
			health = &HealthStatus{
				Status:  HealthStatusUnknown,
				Message: err.Error(),
			}
		}
	}
	return health, err

}

// GetHealthCheckFunc returns built-in health check function or nil if health check is not supported
func GetHealthCheckFunc(gvk schema.GroupVersionKind) func(obj *unstructured.Unstructured) (*HealthStatus, error) {
	switch gvk.Group {
	case "apps":
		switch gvk.Kind {
		case commonBean.DeploymentKind:
			return getDeploymentHealth
		case commonBean.StatefulSetKind:
			return getStatefulSetHealth
		case commonBean.ReplicaSetKind:
			return getReplicaSetHealth
		case commonBean.DaemonSetKind:
			return getDaemonSetHealth
		}
	case "extensions":
		switch gvk.Kind {
		case commonBean.IngressKind:
			return getIngressHealth
		}
	case "argoproj.io":
		switch gvk.Kind {
		case "Workflow":
			return getArgoWorkflowHealth
		}
	case "apiregistration.k8s.io":
		switch gvk.Kind {
		case commonBean.APIServiceKind:
			return getAPIServiceHealth
		}
	case "networking.k8s.io":
		switch gvk.Kind {
		case commonBean.IngressKind:
			return getIngressHealth
		}
	case "":
		switch gvk.Kind {
		case commonBean.ServiceKind:
			return getServiceHealth
		case commonBean.PersistentVolumeClaimKind:
			return getPVCHealth
		case commonBean.PodKind:
			return getPodHealth
		}
	case "batch":
		switch gvk.Kind {
		case commonBean.JobKind:
			return getJobHealth
		}
	case "autoscaling":
		switch gvk.Kind {
		case commonBean.HorizontalPodAutoscalerKind:
			return getHPAHealth
		}
	}
	return nil
}
