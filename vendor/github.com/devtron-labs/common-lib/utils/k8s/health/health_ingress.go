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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func getIngressHealth(obj *unstructured.Unstructured) (*HealthStatus, error) {
	ingresses, _, _ := unstructured.NestedSlice(obj.Object, "status", "loadBalancer", "ingress")
	health := HealthStatus{}
	if len(ingresses) > 0 {
		health.Status = HealthStatusHealthy
	} else {
		health.Status = HealthStatusProgressing
	}
	return &health, nil
}
