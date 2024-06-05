/*
 * Copyright (c) 2024. Devtron Inc.
 */

package bean

import (
	"context"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/enterprise/pkg/deploymentWindow"
	"github.com/devtron-labs/devtron/enterprise/pkg/expressionEvaluators"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"time"
)

const (
	ARGOCD_SYNC_ERROR = "error in syncing argoCD app"
)

type TriggerEvent struct {
	PerformChartPush           bool
	PerformDeploymentOnCluster bool
	GetManifestInResponse      bool
	DeploymentAppType          string
	ManifestStorageType        string
	TriggeredBy                int32
	TriggerdAt                 time.Time
}

type TriggerRequest struct {
	CdWf                   *pipelineConfig.CdWorkflow
	Pipeline               *pipelineConfig.Pipeline
	Artifact               *repository.CiArtifact
	ApplyAuth              bool
	TriggeredBy            int32
	RefCdWorkflowRunnerId  int
	RunStageInEnvNamespace string
	WorkflowType           bean.WorkflowType
	TriggerMessage         string
	DeploymentWindowState  *deploymentWindow.EnvironmentState
	CdWorkflowRunnerId     int // current used for release if runner id comes we dont create runner
	TriggerContext
}

type TriggerContext struct {
	// Context is a context object to be passed to the pipeline trigger
	// +optional
	Context context.Context
	// ReferenceId is a unique identifier for the workflow runner
	// refer pipelineConfig.CdWorkflowRunner
	ReferenceId *string

	// manual or automatic
	TriggerType TriggerType
}

type TriggerType int

const (
	Automatic TriggerType = 1
	Manual    TriggerType = 2
)

func (context TriggerContext) IsAutoTrigger() bool {
	return context.TriggerType == Automatic
}

func (context TriggerContext) ToTriggerTypeString() string {
	if context.IsAutoTrigger() {
		return "AUTO"
	}
	return "MANUAL"
}

type DeploymentType = string

const (
	Helm                    DeploymentType = "helm"
	ArgoCd                  DeploymentType = "argo_cd"
	ManifestDownload        DeploymentType = "manifest_download"
	GitOpsWithoutDeployment DeploymentType = "git_ops_without_deployment"
	ManifestPush            DeploymentType = "manifest_push"
)

const ImagePromotionPolicyValidationErr = "error in cd trigger, user who has approved the image for promotion cannot deploy"

type TriggerRequirementRequestDto struct {
	Scope          resourceQualifiers.Scope
	TriggerRequest TriggerRequest
	Stage          resourceFilter.ReferenceType
	DeploymentType models.DeploymentType
}

type TriggerFeasibilityResponse struct {
	ApprovalRequestId int
	TriggerRequest    TriggerRequest
	FilterIdVsState   map[int]expressionEvaluators.FilterState
	Filters           []*resourceFilter.FilterMetaDataBean
}

type VulnerabilityCheckRequest struct {
	ImageDigest string
	CdPipeline  *pipelineConfig.Pipeline
}

type TriggerOperationDto struct {
	TriggerRequest  TriggerRequest
	ExecutorType    pipelineConfig.WorkflowExecutorType
	PipelineId      int
	Scope           resourceQualifiers.Scope
	TriggeredAt     time.Time
	OverrideCdWrfId int
}

const (
	CronJobChartRegexExpression = "cronjob-chart_1-(2|3|4|5)-0"
)
