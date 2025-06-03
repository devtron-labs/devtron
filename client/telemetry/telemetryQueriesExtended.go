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

// getProjectsWithZeroAppsCount returns the number of projects (teams) that have no applications
func (impl *TelemetryEventClientImplExtended) getProjectsWithZeroAppsCount() int {
	var count int
	query := `
		SELECT COUNT(*)
		FROM team t
		WHERE t.active = true
		AND NOT EXISTS (
			SELECT 1 FROM app a
			WHERE a.team_id = t.id AND a.active = true
		)
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting projects with zero apps count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted projects with zero apps", "count", count)
	return count
}

// getAppsWithPropagationTagsCount returns the number of apps that have at least one propagation tag
func (impl *TelemetryEventClientImplExtended) getAppsWithPropagationTagsCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT al.app_id)
		FROM app_label al
		INNER JOIN app a ON al.app_id = a.id
		WHERE al.propagate = true AND a.active = true
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting apps with propagation tags count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted apps with propagation tags", "count", count)
	return count
}

// getAppsWithNonPropagationTagsCount returns the number of apps that have at least one non-propagation tag
func (impl *TelemetryEventClientImplExtended) getAppsWithNonPropagationTagsCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT al.app_id)
		FROM app_label al
		INNER JOIN app a ON al.app_id = a.id
		WHERE al.propagate = false AND a.active = true
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting apps with non-propagation tags count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted apps with non-propagation tags", "count", count)
	return count
}

// getAppsWithDescriptionCount returns the number of apps that have a description defined
func (impl *TelemetryEventClientImplExtended) getAppsWithDescriptionCount() int {
	var count int
	query := `
		SELECT COUNT(*)
		FROM app a
		WHERE a.active = true
		AND a.description IS NOT NULL
		AND TRIM(a.description) != ''
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting apps with description count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted apps with description", "count", count)
	return count
}

// getAppsWithCatalogDataCount returns the number of apps that have catalog data (app store applications)
func (impl *TelemetryEventClientImplExtended) getAppsWithCatalogDataCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT ia.id)
		FROM installed_apps ia
		INNER JOIN installed_app_versions iav ON ia.id = iav.installed_app_id
		INNER JOIN app_store_application_version asav ON iav.app_store_application_version_id = asav.id
		WHERE ia.active = true AND iav.active = true
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting apps with catalog data count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted apps with catalog data", "count", count)
	return count
}

// getAppsWithReadmeDataCount returns the number of apps that have readme data
func (impl *TelemetryEventClientImplExtended) getAppsWithReadmeDataCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT ia.id)
		FROM installed_apps ia
		INNER JOIN installed_app_versions iav ON ia.id = iav.installed_app_id
		INNER JOIN app_store_application_version asav ON iav.app_store_application_version_id = asav.id
		WHERE ia.active = true
		AND iav.active = true
		AND asav.readme IS NOT NULL
		AND TRIM(asav.readme) != ''
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting apps with readme data count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted apps with readme data", "count", count)
	return count
}
