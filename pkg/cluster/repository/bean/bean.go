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

package bean

import "github.com/devtron-labs/devtron/pkg/cluster/repository"

type EnvironmentBean struct {
	Id                     int      `json:"id,omitempty" validate:"number"`
	Environment            string   `json:"environment_name,omitempty" validate:"required,max=50"`
	ClusterId              int      `json:"cluster_id,omitempty" validate:"number,required"`
	ClusterName            string   `json:"cluster_name,omitempty"`
	Active                 bool     `json:"active"`
	Default                bool     `json:"default"`
	PrometheusEndpoint     string   `json:"prometheus_endpoint,omitempty"`
	Namespace              string   `json:"namespace,omitempty" validate:"name-space-component,max=50"`
	CdArgoSetup            bool     `json:"isClusterCdActive"`
	EnvironmentIdentifier  string   `json:"environmentIdentifier"`
	Description            string   `json:"description" validate:"max=40"`
	AppCount               int      `json:"appCount"`
	IsVirtualEnvironment   bool     `json:"isVirtualEnvironment"`
	AllowedDeploymentTypes []string `json:"allowedDeploymentTypes"`
	IsDigestEnforcedForEnv bool     `json:"isDigestEnforcedForEnv"`
	ClusterServerUrl       string   `json:"-"`
	ErrorInConnecting      string   `json:"-"`
}

func (environmentBean *EnvironmentBean) AdaptFromEnvironment(model *repository.Environment) {

	environmentBean.Id = model.Id
	environmentBean.Environment = model.Name
	environmentBean.ClusterId = model.Cluster.Id
	environmentBean.Active = model.Active
	environmentBean.PrometheusEndpoint = model.Cluster.PrometheusEndpoint
	environmentBean.Namespace = model.Namespace
	environmentBean.Default = model.Default
	environmentBean.EnvironmentIdentifier = model.EnvironmentIdentifier
	environmentBean.Description = model.Description

}

type VirtualEnvironmentBean struct {
	Id                   int    `json:"id,omitempty" validate:"number"`
	Environment          string `json:"environment_name,omitempty" validate:"required,max=50"`
	ClusterId            int    `json:"cluster_id,omitempty" validate:"number,required"`
	ClusterName          string `json:"cluster_name,omitempty"`
	Active               bool   `json:"active"`
	Namespace            string `json:"namespace,omitempty"`
	Description          string `json:"description" validate:"max=40"`
	IsVirtualEnvironment bool   `json:"isVirtualEnvironment"`
}

type EnvDto struct {
	EnvironmentId         int    `json:"environmentId" validate:"number"`
	EnvironmentName       string `json:"environmentName,omitempty" validate:"max=50"`
	Namespace             string `json:"namespace,omitempty" validate:"name-space-component,max=50"`
	EnvironmentIdentifier string `json:"environmentIdentifier,omitempty"`
	Description           string `json:"description" validate:"max=40"`
	IsVirtualEnvironment  bool   `json:"isVirtualEnvironment"`
	Default               bool   `json:"default"`
}

type ClusterEnvDto struct {
	ClusterId        int       `json:"clusterId"`
	ClusterName      string    `json:"clusterName,omitempty"`
	Environments     []*EnvDto `json:"environments,omitempty"`
	IsVirtualCluster bool      `json:"isVirtualCluster"`
}

type ResourceGroupingResponse struct {
	EnvList  []EnvironmentBean `json:"envList"`
	EnvCount int               `json:"envCount"`
}

const (
	PIPELINE_DEPLOYMENT_TYPE_HELM = "helm"
	PIPELINE_DEPLOYMENT_TYPE_ACD  = "argo_cd"
)
