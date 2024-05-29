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

package resourceQualifiers

type Scope struct {
	AppId          int             `json:"appId"`
	EnvId          int             `json:"envId"`
	ClusterId      int             `json:"clusterId"`
	ProjectId      int             `json:"projectId"`
	IsProdEnv      bool            `json:"isProdEnv"`
	PipelineId     int             `json:"pipelineId"`
	SystemMetadata *SystemMetadata `json:"-"`
}

type SystemMetadata struct {
	EnvironmentName string
	ClusterName     string
	Namespace       string
	ImageTag        string
	Image           string
	AppName         string
}

func (metadata *SystemMetadata) GetDataFromSystemVariable(variable SystemVariableName) string {
	switch variable {
	case DevtronNamespace:
		return metadata.Namespace
	case DevtronClusterName:
		return metadata.ClusterName
	case DevtronEnvName:
		return metadata.EnvironmentName
	case DevtronImageTag:
		return metadata.ImageTag
	case DevtronAppName:
		return metadata.AppName
	case DevtronImage:
		return metadata.Image
	}
	return ""
}

type Qualifier int

const (
	APP_AND_ENV_QUALIFIER Qualifier = 1
	APP_QUALIFIER         Qualifier = 2
	ENV_QUALIFIER         Qualifier = 3
	CLUSTER_QUALIFIER     Qualifier = 4
	GLOBAL_QUALIFIER      Qualifier = 5
	PIPELINE_QUALIFIER    Qualifier = 6
)

var CompoundQualifiers = []Qualifier{APP_AND_ENV_QUALIFIER}

func GetNumOfChildQualifiers(qualifier Qualifier) int {
	switch qualifier {
	case APP_AND_ENV_QUALIFIER:
		return 1
	}
	return 0
}

type ResourceIdentifierCount struct {
	ResourceId      int
	IdentifierCount int
}
