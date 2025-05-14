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

package bean

import (
	"fmt"
	"github.com/deckarep/golang-set"
	"github.com/devtron-labs/devtron/pkg/bean"
)

const (
	CD_PIPELINE_TYPE = "CD_PIPELINE"
	CI_PIPELINE_TYPE = "CI_PIPELINE"
)

type WorkflowsFilterQuery struct {
	EnvIdsString string `schema:"envIds"`
	EnvIds       []int  `schema:"-"`
}

func NewWorkflowsFilterQuery() *WorkflowsFilterQuery {
	return &WorkflowsFilterQuery{}
}

func (q *WorkflowsFilterQuery) WithEnvIds(envIds []int) *WorkflowsFilterQuery {
	q.EnvIds = envIds
	return q
}

type WorkflowMappingsNotFoundError struct {
	WorkflowIds []int
}

func (w WorkflowMappingsNotFoundError) Error() string {
	return fmt.Sprintf("workflow not found %v not found", w.WorkflowIds)
}

type WorkflowComponents struct {
	CiPipelineId         int
	ExternalCiPipelineId int
	CdPipelineIds        []int
}

type AppWorkflowListRespDto struct {
	Workflows                 []AppWorkflowDto `json:"workflows"`
	AppId                     int              `json:"appId"`
	AppName                   string           `json:"appName"`
	IsGitOpsRepoNotConfigured bool             `json:"isGitOpsRepoNotConfigured"`
}

type AppWorkflowDto struct {
	Id                    int                     `json:"id,omitempty"`
	Name                  string                  `json:"name"`
	AppId                 int                     `json:"appId"`
	AppWorkflowMappingDto []AppWorkflowMappingDto `json:"tree,omitempty"`
	UserId                int32                   `json:"-"`
}

type TriggerViewWorkflowConfig struct {
	Workflows        []AppWorkflowDto          `json:"workflows"`
	CiConfig         *bean.TriggerViewCiConfig `json:"ciConfig"`
	CdPipelines      *bean.CdPipelines         `json:"cdConfig"`
	ExternalCiConfig []*bean.ExternalCiConfig  `json:"externalCiConfig"`
}

type AppWorkflowMappingDto struct {
	Id                         int        `json:"id,omitempty"`
	AppWorkflowId              int        `json:"appWorkflowId"`
	Type                       string     `json:"type"`
	ComponentId                int        `json:"componentId"`
	ParentId                   int        `json:"parentId"`
	ParentType                 string     `json:"parentType"`
	DeploymentAppDeleteRequest bool       `json:"deploymentAppDeleteRequest"`
	UserId                     int32      `json:"-"`
	IsLast                     bool       `json:"isLast"`
	ChildPipelinesIds          mapset.Set `json:"-"`
}

func (dto AppWorkflowMappingDto) GetPipelineIdentifier() PipelineIdentifier {
	return PipelineIdentifier{
		PipelineType: dto.Type,
		PipelineId:   dto.ComponentId,
	}
}

func (dto AppWorkflowMappingDto) GetParentPipelineIdentifier() PipelineIdentifier {
	return PipelineIdentifier{
		PipelineType: dto.ParentType,
		PipelineId:   dto.ParentId,
	}
}

type AllAppWorkflowComponentDetails struct {
	Workflows []*WorkflowComponentNamesDto `json:"workflows"`
}

type WorkflowComponentNamesDto struct {
	Id             int      `json:"id"`
	Name           string   `json:"name"`
	CiPipelineId   int      `json:"ciPipelineId"`
	CiPipelineName string   `json:"ciPipelineName"`
	CdPipelines    []string `json:"cdPipelines"`
}

type WorkflowNamesResponse struct {
	AppIdWorkflowNamesMapping map[string][]string `json:"appIdWorkflowNamesMapping"`
}

type WorkflowNamesRequest struct {
	AppNames []string `json:"appNames"`
}

type WorkflowCloneRequest struct {
	WorkflowName  string `json:"workflowName,omitempty"`
	AppId         int    `json:"appId,omitempty"`
	EnvironmentId int    `json:"environmentId,omitempty"`
	WorkflowId    int    `json:"workflowId,omitempty"`
	UserId        int32  `json:"-"`
}

type PipelineIdentifier struct {
	PipelineType string
	PipelineId   int
}
