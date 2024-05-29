/*
 * Copyright (c) 2024. Devtron Inc.
 */

package bean

import (
	"fmt"
	"github.com/deckarep/golang-set"
	"github.com/devtron-labs/devtron/pkg/bean"
)

const (
	WEBHOOK_TYPE     = "WEBHOOK"
	CD_PIPELINE_TYPE = "CD_PIPELINE"
	CI_PIPELINE_TYPE = "CI_PIPELINE"
)

type AppWorkflowDto struct {
	Id                        int                        `json:"id,omitempty"`
	Name                      string                     `json:"name"`
	AppId                     int                        `json:"appId"`
	AppWorkflowMappingDto     []AppWorkflowMappingDto    `json:"tree,omitempty"`
	ArtifactPromotionMetadata *ArtifactPromotionMetadata `json:"artifactPromotionMetadata"`
	UserId                    int32                      `json:"-"`
}

type ArtifactPromotionMetadata struct {
	IsApprovalPendingForPromotion bool `json:"isApprovalPendingForPromotion"`
	IsConfigured                  bool `json:"isConfigured"`
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
	EnvironmentName            string     `json:"environmentName"`
	HelmPackageName            string     `json:"helmPackageName"`
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

type AllAppWorkflowComponentNames struct {
	Workflows []*WorkflowComponentNamesDto `json:"workflows"`
}

type AppWorkflowComponentDetails struct {
	Id           int                        `json:"id"`
	WorkflowName string                     `json:"workflowName"`
	AppId        int                        `json:"-"`
	CiPipeline   *bean.CiComponentDetails   `json:"ciPipeline"`
	CdPipelines  []*bean.CdComponentDetails `json:"cdPipelines"`
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
