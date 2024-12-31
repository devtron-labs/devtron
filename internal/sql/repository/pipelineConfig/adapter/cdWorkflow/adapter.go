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
