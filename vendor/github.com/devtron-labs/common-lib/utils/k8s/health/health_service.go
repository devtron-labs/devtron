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
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func getServiceHealth(obj *unstructured.Unstructured) (*HealthStatus, error) {
	gvk := obj.GroupVersionKind()
	switch gvk {
	case corev1.SchemeGroupVersion.WithKind(commonBean.ServiceKind):
		var service corev1.Service
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &service)
		if err != nil {
			return nil, fmt.Errorf("failed to convert unstructured Service to typed: %v", err)
		}
		return getCorev1ServiceHealth(&service)
	default:
		return nil, fmt.Errorf("unsupported Service GVK: %s", gvk)
	}
}

func getCorev1ServiceHealth(service *corev1.Service) (*HealthStatus, error) {
	health := HealthStatus{Status: HealthStatusHealthy}
	if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
		if len(service.Status.LoadBalancer.Ingress) > 0 {
			health.Status = HealthStatusHealthy
		} else {
			health.Status = HealthStatusProgressing
		}
	}
	return &health, nil
}
