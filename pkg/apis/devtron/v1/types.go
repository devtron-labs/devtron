/*
 * Copyright (c) 2020 Devtron Labs
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
 *
 */

package v1

// App defines model for app .
type App struct {

	// API version of this configuration
	ApiVersion string       `json:"apiVersion"`
	ConfigMaps []DataHolder `json:"configMaps"`

	// Unique identification of resource
	Destination    *ResourcePath `json:"destination,omitempty"`
	DockerConfig   DockerConfig  `json:"dockerConfig"`
	DockerRegistry string        `json:"dockerRegistry"`
	DockerRepo     string        `json:"dockerRepo"`

	// Action to be taken on the component
	Operation Operation    `json:"operation"`
	Repo      []Repo       `json:"repo"`
	Secrets   []DataHolder `json:"secrets"`

	// Unique identification of resource
	Source *ResourcePath `json:"source,omitempty"`
	Team   string        `json:"team"`

	// Workflow for app
	Workflow *[]Workflow `json:"workflow,omitempty"`
}

// BlueGreenStrategy defines model for blueGreenStrategy.
type BlueGreenStrategy struct {
	AutoPromotionEnabled  bool  `json:"autoPromotionEnabled"`
	AutoPromotionSeconds  int32 `json:"autoPromotionSeconds"`
	PreviewReplicaCount   int32 `json:"previewReplicaCount"`
	ScaleDownDelaySeconds int32 `json:"scaleDownDelaySeconds"`
}

// Build defines model for build.
type Build struct {
	AccessKey *string `json:"accessKey,omitempty"`

	// API version of this configuration
	ApiVersion     string          `json:"apiVersion"`
	BuildMaterials []BuildMaterial `json:"buildMaterials"`

	// Unique identification of resource
	Destination     *ResourcePath          `json:"destination,omitempty"`
	DockerArguments map[string]interface{} `json:"dockerArguments"`
	NextPipeline    *Pipeline              `json:"nextPipeline,omitempty"`

	// Action to be taken on the component
	Operation Operation `json:"operation"`
	Payload   *string   `json:"payload,omitempty"`
	PostBuild *Task     `json:"postBuild,omitempty"`
	PreBuild  *Task     `json:"preBuild,omitempty"`
	Repo      []Repo    `json:"repo"`

	// Unique identification of resource
	Source *ResourcePath `json:"source,omitempty"`

	// How will this action be initiated
	Trigger    Trigger `json:"trigger"`
	WebHookUrl *string `json:"webHookUrl,omitempty"`
}

// BuildMaterial defines model for buildMaterial.
type BuildMaterial struct {
	GitMaterialUrl string `json:"gitMaterialUrl"`
	Source         struct {
		Type  string `json:"type"`
		Value string `json:"value"`
		Regex string `json:"regex"`
	} `json:"source"`
}

// CanaryStrategy defines model for canaryStrategy.
type CanaryStrategy struct {
	MaxSurge       string        `json:"maxSurge"`
	MaxUnavailable int32         `json:"maxUnavailable"`
	Steps          []interface{} `json:"steps"`
}

// ConfigMaps defines model for configMaps.
type ConfigMaps DataHolder

// DataHolder defines model for dataHolder.
type DataHolder struct {

	// API version of this configuration
	ApiVersion string `json:"apiVersion"`

	// If operation is clone, leaving value empty results in deletion of key in destination.
	Data map[string]interface{} `json:"data"`

	// Unique identification of resource
	Destination  *ResourcePath `json:"destination,omitempty"`
	External     bool          `json:"external"`
	ExternalType string        `json:"externalType"`
	Global       bool          `json:"global"`
	MountPath    string        `json:"mountPath"`

	// Action to be taken on the component
	Operation Operation `json:"operation"`

	// Unique identification of resource
	Source *ResourcePath `json:"source,omitempty"`
	Type   string        `json:"type"`
}

// Deployment defines model for deployment.
type Deployment struct {

	// API version of this configuration
	ApiVersion string `json:"apiVersion"`

	// These are applied for environment
	ConfigMaps []DataHolder `json:"configMaps"`

	// Unique identification of resource
	Destination  *ResourcePath `json:"destination,omitempty"`
	NextPipeline *Pipeline     `json:"nextPipeline,omitempty"`

	// Action to be taken on the component
	Operation         Operation `json:"operation"`
	PostDeployment    *Task     `json:"postDeployment,omitempty"`
	PreDeployment     *Task     `json:"preDeployment,omitempty"`
	PreviousPipeline  *Pipeline `json:"previousPipeline,omitempty"`
	RunPostStageInEnv bool      `json:"runPostStageInEnv"`
	RunPreStageInEnv  bool      `json:"runPreStageInEnv"`

	// These are applied for environment
	Secrets []DataHolder `json:"secrets"`

	// Unique identification of resource
	Source *ResourcePath `json:"source,omitempty"`

	// Strategy as defined by devtron template, this overrides at environment level
	Strategy DeploymentStrategy  `json:"strategy"`
	Template *DeploymentTemplate `json:"template,omitempty"`

	// How will this action be initiated
	Trigger *Trigger `json:"trigger,omitempty"`
}

// DeploymentStrategy defines model for deploymentStrategy.
type DeploymentStrategy struct {
	BlueGreen *BlueGreenStrategy `json:"blueGreen,omitempty"`
	Canary    *CanaryStrategy    `json:"canary,omitempty"`
	Default   string             `json:"default"`
	Recreate  *RecreateStrategy  `json:"recreate,omitempty"`
	Rolling   *RollingStrategy   `json:"rolling,omitempty"`
}

// DeploymentTemplate defines model for deploymentTemplate.
type DeploymentTemplate struct {

	// API version of this configuration
	ApiVersion         string                 `json:"apiVersion"`
	ChartRefId         int32                  `json:"chartRefId"`
	DefaultAppOverride map[string]interface{} `json:"defaultAppOverride"`

	// Unique identification of resource
	Destination         *ResourcePath `json:"destination,omitempty"`
	IsAppMetricsEnabled bool          `json:"isAppMetricsEnabled"`

	// Action to be taken on the component
	Operation               Operation `json:"operation"`
	RefChartTemplate        string    `json:"refChartTemplate"`
	RefChartTemplateVersion string    `json:"refChartTemplateVersion"`

	// Unique identification of resource
	Source         *ResourcePath          `json:"source,omitempty"`
	ValuesOverride map[string]interface{} `json:"valuesOverride"`
}

// DockerConfig defines model for dockerConfig.
type DockerConfig struct {
	Args                   map[string]interface{} `json:"args"`
	DockerFilePath         string                 `json:"dockerFilePath"`
	DockerFileRelativePath string                 `json:"dockerFileRelativePath"`
	DockerFileRepository   string                 `json:"dockerFileRepository"`
	GitMaterial            string                 `json:"gitMaterial"`
}

// InheritedProps defines model for inheritedProps.
type InheritedProps struct {

	// Unique identification of resource
	Destination *ResourcePath `json:"destination,omitempty"`

	// Action to be taken on the component
	Operation Operation `json:"operation"`

	// Unique identification of resource
	Source *ResourcePath `json:"source,omitempty"`
}

// Operation defines model for operation.
type Operation string

// Pipeline defines model for pipeline.
type Pipeline struct {
	Build      *Build      `json:"build,omitempty"`
	Deployment *Deployment `json:"deployment,omitempty"`
}

// PipelineRequest defines model for pipelineRequest.
type PipelineRequest struct {

	// API version of this configuration
	ApiVersion string `json:"apiVersion"`

	// Entries can be of type build or deployment
	Pipelines []Pipeline `json:"pipelines"`
}

// PostBuild defines model for postBuild.
type PostBuild Task

// PostDeployment defines model for postDeployment.
type PostDeployment Task

// PreBuild defines model for preBuild.
type PreBuild Task

// PreDeployment defines model for preDeployment.
type PreDeployment Task

// RecreateStrategy defines model for recreateStrategy.
type RecreateStrategy map[string]interface{}

// Repo defines model for repo.
type Repo struct {

	// branch to build
	Branch *string `json:"branch,omitempty"`

	// path to checkout
	Path *string `json:"path,omitempty"`

	// git url
	Url *string `json:"url,omitempty"`
}

// ResourcePath defines model for resourcePath.
type ResourcePath struct {
	App         *string `json:"app,omitempty"`
	ConfigMap   *string `json:"configMap,omitempty"`
	Environment *string `json:"environment,omitempty"`
	Pipeline    *string `json:"pipeline,omitempty"`
	Secret      *string `json:"secret,omitempty"`
	Uid         *string `json:"uid,omitempty"`
	Workflow    *string `json:"workflow,omitempty"`
}

// RollingStrategy defines model for rollingStrategy.
type RollingStrategy struct {
	MaxSurge       string `json:"maxSurge"`
	MaxUnavailable int32  `json:"maxUnavailable"`
}

// Secrets defines model for secrets.
type Secrets DataHolder

// Stage defines model for stage.
type Stage struct {
	Name string `json:"name"`

	// Action to be taken on the component
	Operation      Operation `json:"operation"`
	OutputLocation *string   `json:"outputLocation,omitempty"`
	Position       *int32    `json:"position,omitempty"`
	Script         *string   `json:"script,omitempty"`
}

// Task defines model for task.
type Task struct {

	// API version of this configuration
	ApiVersion string   `json:"apiVersion"`
	ConfigMaps []string `json:"configMaps"`

	// Unique identification of resource
	Destination *ResourcePath `json:"destination,omitempty"`

	// Action to be taken on the component
	Operation Operation `json:"operation"`
	Secrets   []string  `json:"secrets"`

	// Unique identification of resource
	Source *ResourcePath `json:"source,omitempty"`

	// Different stages in this step
	Stages []Stage `json:"stages"`

	// How will this action be initiated
	Trigger *Trigger `json:"trigger,omitempty"`
}

// Trigger defines model for trigger.
type Trigger string

// Workflow defines model for workflow.
type Workflow struct {

	// API version of this configuration
	ApiVersion string `json:"apiVersion"`

	// Unique identification of resource
	Destination *ResourcePath `json:"destination,omitempty"`

	// Action to be taken on the component
	Operation Operation `json:"operation"`

	// Entries can be of type build or deployment
	Pipelines *[]Pipeline `json:"pipelines,omitempty"`

	// Unique identification of resource
	Source *ResourcePath `json:"source,omitempty"`
}
