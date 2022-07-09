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
)

type AppContainer struct {
	AppId                   int                        `json:"appId"`
	AppName                 string                     `json:"appName"`
	ProjectId               int                        `json:"projectId"`
	AppEnvironmentContainer []*AppEnvironmentContainer `json:"environments"`
	DefaultEnv              AppEnvironmentContainer    `json:"-"`
}

type AppContainerResponse struct {
	AppContainers      []*AppContainer    `json:"appContainers"`
	AppCount           int                `json:"appCount"`
	DeploymentGroupDTO DeploymentGroupDTO `json:"deploymentGroup,omitempty"`
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
	CdStageStatus               *string                   `json:"cdStageStatus"`
	PreStageStatus              *string                   `json:"preStageStatus"`
	PostStageStatus             *string                   `json:"postStageStatus"`
	LastDeployedTime            string                    `json:"lastDeployedTime,omitempty"`
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
	ClusterId                     int             `json:"clusterId"`
	DeploymentAppType             string          `json:"deploymentAppType"`
}

type AppDetailContainer struct {
	DeploymentDetailContainer `json:",inline"`
	InstanceDetail            []InstanceDetail       `json:"instanceDetail"` //pod list with cpu, memory usage percent
	Environments              []Environment          `json:"otherEnvironment,omitempty"`
	LinkOuts                  []LinkOuts             `json:"linkOuts,omitempty"`
	ResourceTree              map[string]interface{} `json:"resourceTree,omitempty"`
}

type Environment struct {
	EnvironmentId   int    `json:"environmentId"`
	EnvironmentName string `json:"environmentName"`
	AppMetrics      *bool  `json:"appMetrics"`
	InfraMetrics    *bool  `json:"infraMetrics"`
	Prod            bool   `json:"prod"`
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
