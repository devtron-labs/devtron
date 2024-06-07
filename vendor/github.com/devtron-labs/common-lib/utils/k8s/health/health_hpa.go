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
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	progressingStatus = &HealthStatus{
		Status:  HealthStatusProgressing,
		Message: "Waiting to Autoscale",
	}
)

type hpaCondition struct {
	Type    string
	Reason  string
	Message string
	Status  string
}

func getHPAHealth(obj *unstructured.Unstructured) (*HealthStatus, error) {
	gvk := obj.GroupVersionKind()
	failedConversionMsg := "failed to convert unstructured HPA to typed: %v"

	switch gvk {
	case autoscalingv1.SchemeGroupVersion.WithKind(commonBean.HorizontalPodAutoscalerKind):
		var hpa autoscalingv1.HorizontalPodAutoscaler
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &hpa)
		if err != nil {
			return nil, fmt.Errorf(failedConversionMsg, err)
		}
		return getAutoScalingV1HPAHealth(&hpa)
	case autoscalingv2beta1.SchemeGroupVersion.WithKind(commonBean.HorizontalPodAutoscalerKind):
		var hpa autoscalingv2beta1.HorizontalPodAutoscaler
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &hpa)
		if err != nil {
			return nil, fmt.Errorf(failedConversionMsg, err)
		}
		return getAutoScalingV2beta1HPAHealth(&hpa)
	case autoscalingv2beta2.SchemeGroupVersion.WithKind(commonBean.HorizontalPodAutoscalerKind):
		var hpa autoscalingv2beta2.HorizontalPodAutoscaler
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &hpa)
		if err != nil {
			return nil, fmt.Errorf(failedConversionMsg, err)
		}
		return getAutoScalingV2beta2HPAHealth(&hpa)
	case autoscalingv2.SchemeGroupVersion.WithKind(commonBean.HorizontalPodAutoscalerKind):
		var hpa autoscalingv2.HorizontalPodAutoscaler
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &hpa)
		if err != nil {
			return nil, fmt.Errorf(failedConversionMsg, err)
		}
		return getAutoScalingV2HPAHealth(&hpa)
	default:
		return nil, fmt.Errorf("unsupported HPA GVK: %s", gvk)
	}
}

func getAutoScalingV2HPAHealth(hpa *autoscalingv2.HorizontalPodAutoscaler) (*HealthStatus, error) {
	statusConditions := hpa.Status.Conditions
	conditions := make([]hpaCondition, 0, len(statusConditions))
	for _, statusCondition := range statusConditions {
		conditions = append(conditions, hpaCondition{
			Type:    string(statusCondition.Type),
			Reason:  statusCondition.Reason,
			Message: statusCondition.Message,
			Status:  string(statusCondition.Status),
		})
	}

	return checkConditions(conditions, progressingStatus)
}

func getAutoScalingV2beta2HPAHealth(hpa *autoscalingv2beta2.HorizontalPodAutoscaler) (*HealthStatus, error) {
	statusConditions := hpa.Status.Conditions
	conditions := make([]hpaCondition, 0, len(statusConditions))
	for _, statusCondition := range statusConditions {
		conditions = append(conditions, hpaCondition{
			Type:    string(statusCondition.Type),
			Reason:  statusCondition.Reason,
			Message: statusCondition.Message,
			Status:  string(statusCondition.Status),
		})
	}

	return checkConditions(conditions, progressingStatus)
}

func getAutoScalingV2beta1HPAHealth(hpa *autoscalingv2beta1.HorizontalPodAutoscaler) (*HealthStatus, error) {
	statusConditions := hpa.Status.Conditions
	conditions := make([]hpaCondition, 0, len(statusConditions))
	for _, statusCondition := range statusConditions {
		conditions = append(conditions, hpaCondition{
			Type:    string(statusCondition.Type),
			Reason:  statusCondition.Reason,
			Message: statusCondition.Message,
			Status:  string(statusCondition.Status),
		})
	}

	return checkConditions(conditions, progressingStatus)
}

func getAutoScalingV1HPAHealth(hpa *autoscalingv1.HorizontalPodAutoscaler) (*HealthStatus, error) {
	annotation, ok := hpa.GetAnnotations()["autoscaling.alpha.kubernetes.io/conditions"]
	if !ok {
		return progressingStatus, nil
	}

	var conditions []hpaCondition
	err := json.Unmarshal([]byte(annotation), &conditions)
	if err != nil {
		failedMessage := "failed to convert conditions annotation to typed: %v"
		return nil, fmt.Errorf(failedMessage, err)
	}

	if len(conditions) == 0 {
		return progressingStatus, nil
	}

	return checkConditions(conditions, progressingStatus)
}

func checkConditions(conditions []hpaCondition, progressingStatus *HealthStatus) (*HealthStatus, error) {
	for _, condition := range conditions {
		if isDegraded(&condition) {
			return &HealthStatus{
				Status:  HealthStatusDegraded,
				Message: condition.Message,
			}, nil
		}

		if isHealthy(&condition) {
			return &HealthStatus{
				Status:  HealthStatusHealthy,
				Message: condition.Message,
			}, nil
		}
	}

	return progressingStatus, nil
}

func isDegraded(condition *hpaCondition) bool {
	degraded_states := []hpaCondition{
		{Type: "AbleToScale", Reason: "FailedGetScale"},
		{Type: "AbleToScale", Reason: "FailedUpdateScale"},
		{Type: "ScalingActive", Reason: "FailedGetResourceMetric"},
		{Type: "ScalingActive", Reason: "InvalidSelector"},
	}
	for _, degraded_state := range degraded_states {
		if condition.Type == degraded_state.Type && condition.Reason == degraded_state.Reason {
			return true
		}
	}
	return false
}

func isHealthy(condition *hpaCondition) bool {
	healthyConditionTypes := []string{"AbleToScale", "ScalingLimited"}
	for _, conditionType := range healthyConditionTypes {
		if condition.Type == conditionType && condition.Status == "True" {
			return true
		}
	}
	return false
}
