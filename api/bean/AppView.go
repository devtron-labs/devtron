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

package bean

import (
	"encoding/json"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"time"
)

type AppContainer struct {
	AppId                   int                        `json:"appId"`
	AppName                 string                     `json:"appName"`
	ProjectId               int                        `json:"projectId"`
	AppEnvironmentContainer []*AppEnvironmentContainer `json:"environments"`
	DefaultEnv              AppEnvironmentContainer    `json:"-"`
	Description             GenericNoteResponseBean    `json:"description"`
}

type AppContainerResponse struct {
	AppContainers      []*AppContainer    `json:"appContainers"`
	AppCount           int                `json:"appCount"`
	DeploymentGroupDTO DeploymentGroupDTO `json:"deploymentGroup,omitempty"`
}

type JobContainerResponse struct {
	JobContainers []*JobContainer `json:"jobContainers"`
	JobCount      int             `json:"jobCount"`
}

type DeploymentGroupDTO struct {
	Id             int             `json:"id"`
	Name           string          `json:"name"`
	AppCount       int             `json:"appCount"`
	NoOfApps       string          `json:"noOfApps"`
	EnvironmentId  int             `json:"environmentId"`
	CiPipelineId   int             `json:"ciPipelineId"`
	CiMaterialDTOs []CiMaterialDTO `json:"ciMaterialDTOs"`
}

type CiMaterialDTO struct {
	Name        string `json:"name"`
	SourceType  string `json:"type"`
	SourceValue string `json:"value"`
}

type GenericNoteResponseBean struct {
	Id          int       `json:"id" validate:"number"`
	Description string    `json:"description"`
	UpdatedBy   string    `json:"updatedBy"`
	UpdatedOn   time.Time `json:"updatedOn"`
}

type JobContainer struct {
	JobId          int                     `json:"jobId"`
	JobName        string                  `json:"jobName""`
	JobActualName  string                  `json:"appName""`
	Description    GenericNoteResponseBean `json:"description"`
	JobCiPipelines []JobCIPipeline         `json:"ciPipelines"'`
	ProjectId      int                     `json:"projectId"`
}

type JobCIPipeline struct {
	CiPipelineId                 int       `json:"ciPipelineId"`
	CiPipelineName               string    `json:"ciPipelineName"`
	Status                       string    `json:"status"`
	LastRunAt                    time.Time `json:"lastRunAt"`
	LastSuccessAt                time.Time `json:"lastSuccessAt"`
	EnvironmentId                int       `json:"environmentId"`
	EnvironmentName              string    `json:"environmentName"`
	LastTriggeredEnvironmentName string    `json:"lastTriggeredEnvironmentName"`
}

type JobListingContainer struct {
	JobId                        int       `sql:"job_id" json:"jobId"`
	JobName                      string    `sql:"job_name" json:"jobName"`
	JobActualName                string    `sql:"app_name" json:"appName"`
	Description                  string    `sql:"description" json:"description"`
	CiPipelineID                 int       `sql:"ci_pipeline_id" json:"ciPipelineID"`
	CiPipelineName               string    `sql:"ci_pipeline_name" json:"ciPipelineName"`
	Status                       string    `sql:"status" json:"status"`
	StartedOn                    time.Time `sql:"started_on" json:"startedOn"`
	EnvironmentId                int       `sql:"environment_id" json:"environmentId"`
	EnvironmentName              string    `sql:"environment_name" json:"environmentName"`
	LastTriggeredEnvironmentName string    `sql:"last_triggered_environment_name" json:"lastTriggeredEnvironmentName"`
	LastTriggeredEnvironmentId   int       `sql:"last_triggered_environment_id" json:"lastEnvironmentId"`
	ProjectId                    int       `sql:"team_id" json:"projectId"`
}

type CiPipelineLastSucceededTime struct {
	CiPipelineID    int       `json:"ci_pipeline_id"`
	LastSucceededOn time.Time `json:"last_succeeded_on"`
}

type AppEnvironmentContainer struct {
	AppId                       int                       `json:"appId"`
	AppName                     string                    `json:"appName"`
	EnvironmentId               int                       `json:"environmentId"`
	EnvironmentName             string                    `json:"environmentName"`
	Namespace                   string                    `json:"namespace"`
	ClusterName                 string                    `json:"clusterName"`
	DeploymentCounter           int                       `json:"deploymentCounter,omitempty"`
	InstanceCounter             int                       `json:"instanceCounter,omitempty"`
	Status                      string                    `json:"status"`
	AppStatus                   string                    `json:"appStatus"`
	CdStageStatus               *string                   `json:"cdStageStatus"`
	PreStageStatus              *string                   `json:"preStageStatus"`
	PostStageStatus             *string                   `json:"postStageStatus"`
	LastDeployedTime            string                    `json:"lastDeployedTime,omitempty"`
	LastDeployedImage           string                    `json:"lastDeployedImage,omitempty"`
	LastDeployedBy              string                    `json:"lastDeployedBy,omitempty"`
	LastSuccessDeploymentDetail DeploymentDetailContainer `json:"-"`
	Default                     bool                      `json:"default"`
	Deleted                     bool                      `json:"deleted"`
	MaterialInfo                json.RawMessage           `json:"materialInfo,omitempty"`
	DataSource                  string                    `json:"dataSource,omitempty"`
	MaterialInfoJson            string                    `json:"-"`
	PipelineId                  int                       `json:"-"`
	PipelineReleaseCounter      int                       `json:"-"`
	CiArtifactId                int                       `json:"ciArtifactId"`
	Active                      bool                      `json:"-"`
	TeamId                      int                       `json:"teamId"`
	TeamName                    string                    `json:"teamName"`
	Description                 string                    `json:"description" validate:"max=40"`
	TotalCount                  int                       `json:"-"`
	Commits                     []string                  `json:"commits"`
}

type DeploymentDetailContainer struct {
	InstalledAppId                int             `json:"installedAppId,omitempty"`
	AppId                         int             `json:"appId,omitempty"`
	AppStoreInstalledAppVersionId int             `json:"appStoreInstalledAppVersionId,omitempty"`
	AppStoreChartName             string          `json:"appStoreChartName,omitempty"`
	AppStoreChartId               int             `json:"appStoreChartId,omitempty"`
	AppStoreAppName               string          `json:"appStoreAppName,omitempty"`
	AppStoreAppVersion            string          `json:"appStoreAppVersion,omitempty"`
	AppName                       string          `json:"appName"`
	EnvironmentId                 int             `json:"environmentId"`
	EnvironmentName               string          `json:"environmentName"`
	Namespace                     string          `json:"namespace,omitempty"`
	Status                        string          `json:"status,omitempty"`
	StatusMessage                 string          `json:"statusMessage,omitempty"`
	LastDeployedTime              string          `json:"lastDeployedTime,omitempty"`
	LastDeployedBy                string          `json:"lastDeployedBy,omitempty"`
	MaterialInfo                  json.RawMessage `json:"materialInfo,omitempty"`
	MaterialInfoJsonString        string          `json:"-"`
	ReleaseVersion                string          `json:"releaseVersion,omitempty"`
	Default                       bool            `json:"default,omitempty"`
	DataSource                    string          `json:"dataSource,omitempty"`
	LastDeployedPipeline          string          `json:"lastDeployedPipeline,omitempty"`
	Deprecated                    bool            `json:"deprecated"`
	K8sVersion                    string          `json:"k8sVersion"`
	CiArtifactId                  int             `json:"ciArtifactId"`
	ParentArtifactId              int             `json:"parentArtifactId"`
	ClusterId                     int             `json:"clusterId"`
	DeploymentAppType             string          `json:"deploymentAppType"`
	CiPipelineId                  int             `json:"-"`
	IsExternalCi                  bool            `json:"externalCi"`
	ClusterName                   string          `json:"clusterName,omitempty"`
	DockerRegistryId              string          `json:"dockerRegistryId,omitempty"`
	IpsAccessProvided             bool            `json:"ipsAccessProvided"`
	DeploymentAppDeleteRequest    bool            `json:"deploymentAppDeleteRequest"`
	Description                   string          `json:"description" validate:"max=40"`
	IsVirtualEnvironment          bool            `json:"isVirtualEnvironment"`
	HelmReleaseInstallStatus      string          `json:"-"`
}

type AppDetailContainer struct {
	DeploymentDetailContainer `json:",inline"`
	InstanceDetail            []InstanceDetail       `json:"instanceDetail"` //pod list with cpu, memory usage percent
	Environments              []Environment          `json:"otherEnvironment,omitempty"`
	LinkOuts                  []LinkOuts             `json:"linkOuts,omitempty"`
	ResourceTree              map[string]interface{} `json:"resourceTree,omitempty"`
	Notes                     string                 `json:"notes,omitempty"`
}
type AppDetailsContainer struct {
	ResourceTree  map[string]interface{} `json:"resourceTree,omitempty"`
	Notes         string                 `json:"notes,omitempty"`
	ReleaseStatus map[string]interface{} `json:"releaseStatus"`
}
type Notes struct {
	Notes string `json:"gitOpsNotes,omitempty"`
}

type Environment struct {
	AppStatus                  string   `json:"appStatus"` //this is not the status of environment , this make sense with a specific app only
	EnvironmentId              int      `json:"environmentId"`
	EnvironmentName            string   `json:"environmentName"`
	AppMetrics                 *bool    `json:"appMetrics"`
	InfraMetrics               *bool    `json:"infraMetrics"`
	Prod                       bool     `json:"prod"`
	ChartRefId                 int      `json:"chartRefId"`
	LastDeployed               string   `json:"lastDeployed"`
	LastDeployedBy             string   `json:"lastDeployedBy"`
	LastDeployedImage          string   `json:"lastDeployedImage"`
	DeploymentAppDeleteRequest bool     `json:"deploymentAppDeleteRequest"`
	Description                string   `json:"description" validate:"max=40"`
	IsVirtualEnvironment       bool     `json:"isVirtualEnvironment"`
	ClusterId                  int      `json:"clusterId"`
	PipelineId                 int      `json:"pipelineId"`
	LatestCdWorkflowRunnerId   int      `json:"latestCdWorkflowRunnerId,omitempty"`
	CiArtifactId               int      `json:"ciArtifactId"`
	ParentCiArtifactId         int      `json:"parentCiArtifactId"`
	Commits                    []string `json:"commits"`
}

type InstanceDetail struct {
	PodName            string  `json:"podName,omitempty"`
	CpuUsage           float64 `json:"cpuUsage,omitempty"`
	MemoryUsage        int64   `json:"memoryUsage,omitempty"`
	CpuRequest         float64 `json:"cpuRequest,omitempty"`
	MemoryRequest      int64   `json:"memoryRequest,omitempty"`
	CpuUsagePercent    float64 `json:"cpuUsagePercent,omitempty"`
	MemoryUsagePercent float64 `json:"memoryUsagePercent,omitempty"`
}

type ResourceUsage struct {
	AppId         int     `json:"appId"` //app_id
	EnvironmentId int     `json:"environmentId"`
	CpuUsage      float64 `json:"cpuUsage,omitempty"`
	MemoryUsage   float64 `json:"memoryUsage,omitempty"`
}

type TriggerView struct {
	CiPipelineId           int                             `json:"ciPipelineId"`
	CiPipelineName         string                          `json:"ciPipelineName"`
	CdPipelineId           int                             `json:"cdPipelineId"`
	CdPipelineName         string                          `json:"cdPipelineName"`
	Status                 string                          `json:"status"`
	StatusMessage          string                          `json:"statusMessage,omitempty"`
	LastDeployedTime       string                          `json:"lastDeployedTime,omitempty"`
	LastDeployedBy         string                          `json:"lastDeployedBy,omitempty"`
	MaterialInfo           json.RawMessage                 `json:"materialInfo,omitempty"`
	MaterialInfoJsonString string                          `json:"-"`
	ReleaseVersion         string                          `json:"releaseVersion,omitempty"`
	DataSource             string                          `json:"dataSource,omitempty"`
	Conditions             []v1alpha1.ApplicationCondition `json:"conditions"`
	AppName                string                          `json:"appName"`
	EnvironmentName        string                          `json:"environmentName"`
}

type DeploymentDetailStat struct {
	AppId         int           `json:"appId"`
	EnvironmentId int           `json:"environmentId"`
	NewPodStats   PodDetailStat `json:"newPodStats,omitempty"`
	OldPodStats   PodDetailStat `json:"oldPodStats,omitempty"`
}

type PodDetailStat struct {
	PodCount           string  `json:"podCount,omitempty"`
	CpuUsagePercent    float64 `json:"cpuUsagePercent,omitempty"`
	MemoryUsagePercent float64 `json:"memoryUsagePercent,omitempty"`
}

type AppStageStatus struct {
	Stage     int    `json:"stage"`
	StageName string `json:"stageName,omitempty"`
	Status    bool   `json:"status"`
	Required  bool   `json:"required"`
}

type LinkOuts struct {
	Id            int    `json:"id"`
	AppId         int    `json:"appId,omitempty"`
	EnvironmentId int    `json:"environmentId,omitempty"`
	Name          string `json:"name"`
	AppName       string `json:"appName,omitempty"`
	EnvName       string `json:"envName,omitempty"`
	PodName       string `json:"podName,omitempty"`
	ContainerName string `json:"containerName,omitempty"`
	Link          string `json:"link,omitempty"`
	Description   string `json:"description,omitempty"`
}
