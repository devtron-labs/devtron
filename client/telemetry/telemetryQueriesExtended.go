package telemetry

import (
	"github.com/devtron-labs/devtron/pkg/build/pipeline/bean/common"
	pluginRepository "github.com/devtron-labs/devtron/pkg/plugin/repository"
)

// FULL-mode specific telemetry methods for TelemetryEventClientImplExtended

func (impl *TelemetryEventClientImplExtended) getDevtronAppCount() int {
	devtronAppCount, err := impl.appRepository.FindDevtronAppCount()
	if err != nil {
		impl.logger.Errorw("error getting all apps for devtron app count", "err", err)
		return -1
	}
	return devtronAppCount
}

func (impl *TelemetryEventClientImplExtended) getJobCount() int {
	jobCount, err := impl.appRepository.FindJobCount()
	if err != nil {
		impl.logger.Errorw("error getting all apps for job count", "err", err)
		return -1
	}
	return jobCount
}

func (impl *TelemetryEventClientImplExtended) getJobPipelineTriggeredLast24h() int {
	// Get build type and status data for the last 24 hours
	buildTypeStatusData, err := impl.ciWorkflowRepository.FindBuildTypeAndStatusDataOfLast1Day()
	if err != nil {
		impl.logger.Warnw("no build type status data available for last 24 hours")
		return -1
	}

	// Count job pipeline triggers
	// Job pipelines have build type "CI_JOB"
	jobTriggeredCount := 0
	for _, data := range buildTypeStatusData {
		if data.Type == string(common.CI_JOB) {
			jobTriggeredCount += data.Count
		}
	}

	return jobTriggeredCount
}

func (impl *TelemetryEventClientImplExtended) getJobPipelineSucceededLast24h() int {
	// Get build type and status data for the last 24 hours
	buildTypeStatusData, err := impl.ciWorkflowRepository.FindBuildTypeAndStatusDataOfLast1Day()
	if err != nil {
		impl.logger.Warnw("no build type status data available for last 24 hours")
		return -1
	}

	// Count successful job pipeline runs
	// Job pipelines have build type "CI_JOB"
	successfulJobCount := 0
	for _, data := range buildTypeStatusData {
		if data.Type == "CI_JOB" && data.Status == "Succeeded" {
			successfulJobCount += data.Count
		}
	}

	impl.logger.Debugw("counted successful job pipeline runs in last 24h", "count", successfulJobCount)
	return successfulJobCount
}

func (impl *TelemetryEventClientImplExtended) getUserCreatedPluginCount() int {
	// Get all user-created plugins (SHARED type)
	plugins, err := impl.pluginRepository.GetAllPluginMinDataByType(string(pluginRepository.PLUGIN_TYPE_SHARED))
	if err != nil {
		impl.logger.Errorw("error getting user created plugin count", "err", err)
		return -1
	}

	return len(plugins)
}

func (impl *TelemetryEventClientImplExtended) getPolicyCount() int {
	// Get global policies
	globalPolicies, err := impl.cvePolicyRepository.GetGlobalPolicies()
	if err != nil {
		impl.logger.Errorw("error getting global CVE policies", "err", err)
		return -1
	}
	return len(globalPolicies)
}

func (impl *TelemetryEventClientImplExtended) getGitOpsPipelineCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT p.id)
		FROM pipeline p
		WHERE p.deleted = false AND p.deployment_app_type = 'argo_cd'
	`

	dbConnection := impl.cdWorkflowRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting GitOps pipeline count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted GitOps pipelines", "count", count)
	return count
}

func (impl *TelemetryEventClientImplExtended) helmPipelineCount() int {
	// Get the pipeline repository from cdWorkflowRepository connection
	var count int
	query := `
		SELECT COUNT(DISTINCT p.id)
		FROM pipeline p
		WHERE p.deleted = false AND p.deployment_app_type = 'helm'
	`

	dbConnection := impl.cdWorkflowRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting No-GitOps pipeline count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted No-GitOps pipelines", "count", count)
	return count
}

// getJobPipelineCount returns 0 for now as implementation is not yet available
func (impl *TelemetryEventClientImplExtended) getJobPipelineCount() int {
	return -1
}

// getAppliedPolicyRowCount returns 0 for now as implementation is not yet available
func (impl *TelemetryEventClientImplExtended) getAppliedPolicyRowCount() int {
	return -1
}
