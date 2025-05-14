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

package cdWorkflow

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/util"
	"time"
)

type UpdateOptions = func(cdWfr *pipelineConfig.CdWorkflowRunner)

func GetTriggerMetricsFromRunnerObj(runner *pipelineConfig.CdWorkflowRunner, deploymentConfig *bean.DeploymentConfig) util.CDMetrics {
	return util.CDMetrics{
		AppName:         runner.CdWorkflow.Pipeline.DeploymentAppName,
		Status:          runner.Status,
		DeploymentType:  deploymentConfig.DeploymentAppType,
		EnvironmentName: runner.CdWorkflow.Pipeline.Environment.Name,
		Time:            time.Since(runner.StartedOn).Seconds() - time.Since(runner.FinishedOn).Seconds(),
	}
}
