package telemetry

import (
	"github.com/devtron-labs/devtron/pkg/build/pipeline/bean/common"
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
)

func (impl *TelemetryEventClientImpl) getHelmAppCount() int {
	count, err := impl.installedAppReadService.GetActiveInstalledAppCount()
	if err != nil {
		impl.logger.Errorw("error getting helm app count", "err", err)
		return -1
	}
	return count
}

func (impl *TelemetryEventClientImpl) getDevtronAppCount() int {
	devtronAppCount, err := impl.appRepository.FindDevtronAppCount()
	if err != nil {
		impl.logger.Errorw("error getting all apps for devtron app count", "err", err)
		return -1
	}
	return devtronAppCount
}

func (impl *TelemetryEventClientImpl) getJobCount() int {
	jobCount, err := impl.appRepository.FindJobCount()
	if err != nil {
		impl.logger.Errorw("error getting all apps for job count", "err", err)
		return -1
	}

	return jobCount
}

func (impl *TelemetryEventClientImpl) getUserCreatedPluginCount() int {
	// Get all user-created plugins (SHARED type)
	plugins, err := impl.pluginRepository.GetAllPluginMinDataByType(string(repository.PLUGIN_TYPE_SHARED))
	if err != nil {
		impl.logger.Errorw("error getting user created plugin count", "err", err)
		return 0
	}

	return len(plugins)
}

func (impl *TelemetryEventClientImpl) getPolicyCount() int {
	// Get global policies
	globalPolicies, err := impl.cvePolicyRepository.GetGlobalPolicies()
	if err != nil {
		impl.logger.Errorw("error getting global CVE policies", "err", err)
	}
	return len(globalPolicies)
}

func (impl *TelemetryEventClientImpl) getClusterCounts() (physicalCount int, isolatedCount int) {
	clusters, err := impl.clusterService.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error getting cluster counts", "err", err)
		return -1, -1
	}

	physicalCount = 0
	isolatedCount = 0

	for _, cluster := range clusters {
		if cluster.IsVirtualCluster {
			isolatedCount++
		} else {
			physicalCount++
		}
	}

	return physicalCount, isolatedCount
}

func (impl *TelemetryEventClientImpl) getJobPipelineCount() int {
	return 0
}

func (impl *TelemetryEventClientImpl) getJobPipelineTriggeredLast24h() int {
	// Get build type and status data for the last 24 hours
	buildTypeStatusData := impl.ciWorkflowRepository.FindBuildTypeAndStatusDataOfLast1Day()
	if buildTypeStatusData == nil {
		impl.logger.Warnw("no build type status data available for last 24 hours")
		return 0
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

func (impl *TelemetryEventClientImpl) getJobPipelineSucceededLast24h() int {
	// Get build type and status data for the last 24 hours
	buildTypeStatusData := impl.ciWorkflowRepository.FindBuildTypeAndStatusDataOfLast1Day()
	if buildTypeStatusData == nil {
		impl.logger.Warnw("no build type status data available for last 24 hours")
		return 0
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

func (impl *TelemetryEventClientImpl) getAppliedPolicyRowCount() int {
	return 0
}
