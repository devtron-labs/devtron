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

package models

import (
	"fmt"
	"strings"
)

type VariableUsageType string

const (
	UsageTypeDeploymentTemplate VariableUsageType = "deploymentTemplate"
	UsageTypeConfigMap          VariableUsageType = "configMap"
	UsageTypeSecret             VariableUsageType = "secret"
	UsageTypePipelineStage      VariableUsageType = "pipelineStage"
	UsageTypeInfraConfig        VariableUsageType = "infraConfig"
)

// VariableUsage is one live reference of a scoped variable in saved configuration.
type VariableUsage struct {
	VariableName string            `json:"variableName"`
	UsageType    VariableUsageType `json:"usageType"`
	AppId        int               `json:"appId,omitempty"`
	AppName      string            `json:"appName,omitempty"`
	EnvId        int               `json:"envId,omitempty"`
	EnvName      string            `json:"envName,omitempty"`
	PipelineName string            `json:"pipelineName,omitempty"`
	StageType    string            `json:"stageType,omitempty"`
}

// VariableDeletionBlockedError is returned when a variable save would delete variables
// that are still referenced by live configuration.
type VariableDeletionBlockedError struct {
	BlockedVariables []string         `json:"blockedVariables"`
	Usages           []*VariableUsage `json:"usages"`
}

func (err *VariableDeletionBlockedError) Error() string {
	return fmt.Sprintf("variable(s) [%s] cannot be deleted because they are currently in use; remove their references first", strings.Join(err.BlockedVariables, ", "))
}

// UsageSummary renders at most limit usages as a human-readable list for error messages.
func (err *VariableDeletionBlockedError) UsageSummary(limit int) string {
	parts := make([]string, 0, len(err.Usages))
	for _, usage := range err.Usages {
		if len(parts) == limit {
			parts = append(parts, fmt.Sprintf("and %d more", len(err.Usages)-limit))
			break
		}
		location := usage.AppName
		if usage.EnvName != "" {
			location = location + "/" + usage.EnvName
		}
		if usage.PipelineName != "" {
			location = location + "/" + usage.PipelineName
		}
		parts = append(parts, fmt.Sprintf("%s is used in %s (%s)", usage.VariableName, location, usage.UsageType))
	}
	return strings.Join(parts, "; ")
}
