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

import (
	"context"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"time"
)

const (
	ARGOCD_SYNC_ERROR = "error in syncing argoCD app"
)

type TriggerEvent struct {
	PerformChartPush           bool
	PerformDeploymentOnCluster bool
	DeploymentAppType          string
	ManifestStorageType        string
	TriggeredBy                int32
	TriggeredAt                time.Time
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
	TriggerContext
}

type TriggerContext struct {
	// Context is a context object to be passed to the pipeline trigger
	// +optional
	Context context.Context
	// ReferenceId is a unique identifier for the workflow runner
	// refer pipelineConfig.CdWorkflowRunner
	ReferenceId *string
}

type DeploymentType = string

const (
	Helm                    DeploymentType = "helm"
	ArgoCd                  DeploymentType = "argo_cd"
	ManifestDownload        DeploymentType = "manifest_download"
	GitOpsWithoutDeployment DeploymentType = "git_ops_without_deployment"
)

type TriggerRequirementRequestDto struct {
	TriggerRequest TriggerRequest
}

type VulnerabilityCheckRequest struct {
	ImageDigest string
	CdPipeline  *pipelineConfig.Pipeline
}

const (
	CronJobChartRegexExpression = "cronjob-chart_1-(2|3|4|5)-0"
)
