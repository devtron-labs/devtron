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

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func getDaemonSetHealth(obj *unstructured.Unstructured) (*HealthStatus, error) {
	gvk := obj.GroupVersionKind()
	switch gvk {
	case appsv1.SchemeGroupVersion.WithKind(commonBean.DaemonSetKind):
		var daemon appsv1.DaemonSet
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &daemon)
		if err != nil {
			return nil, fmt.Errorf("failed to convert unstructured DaemonSet to typed: %v", err)
		}
		return getAppsv1DaemonSetHealth(&daemon)
	default:
		return nil, fmt.Errorf("unsupported DaemonSet GVK: %s", gvk)
	}
}

func getAppsv1DaemonSetHealth(daemon *appsv1.DaemonSet) (*HealthStatus, error) {
	// Borrowed at kubernetes/kubectl/rollout_status.go https://github.com/kubernetes/kubernetes/blob/5232ad4a00ec93942d0b2c6359ee6cd1201b46bc/pkg/kubectl/rollout_status.go#L110
	if daemon.Generation <= daemon.Status.ObservedGeneration {
		if daemon.Spec.UpdateStrategy.Type == appsv1.OnDeleteDaemonSetStrategyType {
			return &HealthStatus{
				Status:  HealthStatusHealthy,
				Message: fmt.Sprintf("daemon set %d out of %d new pods have been updated", daemon.Status.UpdatedNumberScheduled, daemon.Status.DesiredNumberScheduled),
			}, nil
		}
		if daemon.Status.UpdatedNumberScheduled < daemon.Status.DesiredNumberScheduled {
			return &HealthStatus{
				Status:  HealthStatusProgressing,
				Message: fmt.Sprintf("Waiting for daemon set %q rollout to finish: %d out of %d new pods have been updated...", daemon.Name, daemon.Status.UpdatedNumberScheduled, daemon.Status.DesiredNumberScheduled),
			}, nil
		}
		if daemon.Status.NumberAvailable < daemon.Status.DesiredNumberScheduled {
			return &HealthStatus{
				Status:  HealthStatusProgressing,
				Message: fmt.Sprintf("Waiting for daemon set %q rollout to finish: %d of %d updated pods are available...", daemon.Name, daemon.Status.NumberAvailable, daemon.Status.DesiredNumberScheduled),
			}, nil
		}
	} else {
		return &HealthStatus{
			Status:  HealthStatusProgressing,
			Message: "Waiting for rollout to finish: observed daemon set generation less than desired generation",
		}, nil
	}
	return &HealthStatus{
		Status: HealthStatusHealthy,
	}, nil
}
