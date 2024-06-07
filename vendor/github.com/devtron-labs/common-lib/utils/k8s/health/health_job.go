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

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func getJobHealth(obj *unstructured.Unstructured) (*HealthStatus, error) {
	gvk := obj.GroupVersionKind()
	switch gvk {
	case batchv1.SchemeGroupVersion.WithKind(commonBean.JobKind):
		var job batchv1.Job
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &job)
		if err != nil {
			return nil, fmt.Errorf("failed to convert unstructured Job to typed: %v", err)
		}
		return getBatchv1JobHealth(&job)
	default:
		return nil, fmt.Errorf("unsupported Job GVK: %s", gvk)
	}
}

func getBatchv1JobHealth(job *batchv1.Job) (*HealthStatus, error) {
	failed := false
	var failMsg string
	complete := false
	var message string
	for _, condition := range job.Status.Conditions {
		switch condition.Type {
		case batchv1.JobFailed:
			failed = true
			complete = true
			failMsg = condition.Message
		case batchv1.JobComplete:
			complete = true
			message = condition.Message
		}
	}
	if !complete {
		return &HealthStatus{
			Status:  HealthStatusProgressing,
			Message: message,
		}, nil
	} else if failed {
		return &HealthStatus{
			Status:  HealthStatusDegraded,
			Message: failMsg,
		}, nil
	} else {
		return &HealthStatus{
			Status:  HealthStatusHealthy,
			Message: message,
		}, nil
	}
}
