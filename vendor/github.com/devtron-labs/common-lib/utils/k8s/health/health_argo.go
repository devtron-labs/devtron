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
	"k8s.io/apimachinery/pkg/runtime"
)

type nodePhase string

// Workflow and node statuses
// See: https://github.com/argoproj/argo-workflows/blob/master/pkg/apis/workflow/v1alpha1/workflow_phase.go
const (
	nodePending   nodePhase = "Pending"
	nodeRunning   nodePhase = "Running"
	nodeSucceeded nodePhase = "Succeeded"
	nodeFailed    nodePhase = "Failed"
	nodeError     nodePhase = "Error"
)

// An agnostic workflow object only considers Status.Phase and Status.Message. It is agnostic to the API version or any
// other fields.
type argoWorkflow struct {
	Status struct {
		Phase   nodePhase
		Message string
	}
}

func getArgoWorkflowHealth(obj *unstructured.Unstructured) (*HealthStatus, error) {
	var wf argoWorkflow
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &wf)
	if err != nil {
		return nil, err
	}
	switch wf.Status.Phase {
	case "", nodePending, nodeRunning:
		return &HealthStatus{Status: HealthStatusProgressing, Message: wf.Status.Message}, nil
	case nodeSucceeded:
		return &HealthStatus{Status: HealthStatusHealthy, Message: wf.Status.Message}, nil
	case nodeFailed, nodeError:
		return &HealthStatus{Status: HealthStatusDegraded, Message: wf.Status.Message}, nil
	}
	return &HealthStatus{Status: HealthStatusUnknown, Message: wf.Status.Message}, nil
}
