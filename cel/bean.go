/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cel

type ParamValuesType string

const (
	ParamTypeString         ParamValuesType = "string"
	ParamTypeObject         ParamValuesType = "object"
	ParamTypeInteger        ParamValuesType = "integer"
	ParamTypeList           ParamValuesType = "list"
	ParamTypeBool           ParamValuesType = "bool"
	ParamTypeMapStringToAny ParamValuesType = "mapStringToAny"
)

type ParamName string

const AppName ParamName = "appName"
const ProjectName ParamName = "projectName"
const EnvName ParamName = "envName"
const CdPipelineName ParamName = "cdPipelineName"
const IsProdEnv ParamName = "isProdEnv"
const ClusterName ParamName = "clusterName"
const ChartRefId ParamName = "chartRefId"
const CdPipelineTriggerType ParamName = "cdPipelineTriggerType"
const ContainerRepo ParamName = "containerRepository"
const ContainerImage ParamName = "containerImage"
const ContainerImageTag ParamName = "containerImageTag"
const ImageLabels ParamName = "imageLabels"

type Request struct {
	Expression         string             `json:"expression"`
	ExpressionMetadata ExpressionMetadata `json:"params"`
}

type ExpressionMetadata struct {
	Params []ExpressionParam
}

type ExpressionParam struct {
	ParamName ParamName       `json:"paramName"`
	Value     interface{}     `json:"value"`
	Type      ParamValuesType `json:"type"`
}
