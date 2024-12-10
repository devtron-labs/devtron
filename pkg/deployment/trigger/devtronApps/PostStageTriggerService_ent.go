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

package devtronApps

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean4 "github.com/devtron-labs/devtron/pkg/app/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	bean5 "github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"time"
)

func (impl *TriggerServiceImpl) checkFeasibilityForPostStage(pipeline *pipelineConfig.Pipeline, request *bean.TriggerRequest,
	env *repository.Environment, cdWf *pipelineConfig.CdWorkflow, triggeredBy int32) (interface{}, error) {
	//here return type is interface as ResourceFilterEvaluationAudit is not present in this version
	return nil, nil
}

func (impl *TriggerServiceImpl) getManifestPushTemplateForPostStage(request bean.TriggerRequest, envDevploymentConfig *bean5.DeploymentConfig,
	jobHelmPackagePath string, cdStageWorkflowRequest *types.WorkflowRequest, cdWf *pipelineConfig.CdWorkflow, runner *pipelineConfig.CdWorkflowRunner,
	pipeline *pipelineConfig.Pipeline, triggeredBy int32, triggeredAt time.Time) (*bean4.ManifestPushTemplate, error) {
	return nil, nil
}
