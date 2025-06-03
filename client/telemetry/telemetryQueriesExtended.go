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

// getJobPipelineCount returns the number of CI pipelines configured as job pipelines
func (impl *TelemetryEventClientImplExtended) getJobPipelineCount() int {
	var count int
	query := `
		SELECT COUNT(*)
		FROM ci_pipeline cp
		INNER JOIN app a ON cp.app_id = a.id
		WHERE cp.active = true
		AND cp.deleted = false
		AND a.active = true
		AND cp.ci_pipeline_type = 'CI_JOB'
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting job pipeline count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted job pipelines", "count", count)
	return count
}

// getAppliedPolicyRowCount returns the total number of applied policies in the system
func (impl *TelemetryEventClientImplExtended) getAppliedPolicyRowCount() int {
	var totalCount int

	// Count CVE/Security policies
	var cvePolicyCount int
	cvePolicyQuery := `
		SELECT COUNT(*)
		FROM cve_policy_control cpc
		WHERE cpc.deleted = false
	`

	// Count RBAC policies
	var rbacPolicyCount int
	rbacPolicyQuery := `
		SELECT COUNT(*)
		FROM rbac_policy_data rpd
		WHERE rpd.deleted = false
	`

	dbConnection := impl.appRepository.GetConnection()

	// Get CVE policy count
	_, err := dbConnection.Query(&cvePolicyCount, cvePolicyQuery)
	if err != nil {
		impl.logger.Errorw("error getting CVE policy count", "err", err)
		return -1
	}

	// Get RBAC policy count
	_, err = dbConnection.Query(&rbacPolicyCount, rbacPolicyQuery)
	if err != nil {
		impl.logger.Errorw("error getting RBAC policy count", "err", err)
		return -1
	}

	totalCount = cvePolicyCount + rbacPolicyCount

	impl.logger.Debugw("counted applied policies", "cvePolicyCount", cvePolicyCount, "rbacPolicyCount", rbacPolicyCount, "totalCount", totalCount)
	return totalCount
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

// getHighestEnvironmentCountInApp returns the highest number of environments configured in a single application
func (impl *TelemetryEventClientImplExtended) getHighestEnvironmentCountInApp() int {
	var maxCount int
	query := `
		SELECT COALESCE(MAX(env_count), 0) as max_env_count
		FROM (
			SELECT COUNT(DISTINCT p.environment_id) as env_count
			FROM pipeline p
			INNER JOIN app a ON p.app_id = a.id
			WHERE p.deleted = false AND a.active = true
			GROUP BY p.app_id
		) app_env_counts
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&maxCount, query)
	if err != nil {
		impl.logger.Errorw("error getting highest environment count in app", "err", err)
		return -1
	}

	impl.logger.Debugw("counted highest environment count in app", "count", maxCount)
	return maxCount
}

// getHighestAppCountInEnvironment returns the highest number of applications deployed in a single environment
func (impl *TelemetryEventClientImplExtended) getHighestAppCountInEnvironment() int {
	var maxCount int
	query := `
		SELECT COALESCE(MAX(app_count), 0) as max_app_count
		FROM (
			SELECT COUNT(DISTINCT p.app_id) as app_count
			FROM pipeline p
			INNER JOIN app a ON p.app_id = a.id
			WHERE p.deleted = false AND a.active = true
			GROUP BY p.environment_id
		) env_app_counts
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&maxCount, query)
	if err != nil {
		impl.logger.Errorw("error getting highest app count in environment", "err", err)
		return -1
	}

	impl.logger.Debugw("counted highest app count in environment", "count", maxCount)
	return maxCount
}

// getHighestWorkflowCountInApp returns the highest number of workflows in a single application
func (impl *TelemetryEventClientImplExtended) getHighestWorkflowCountInApp() int {
	var maxCount int
	query := `
		SELECT COALESCE(MAX(workflow_count), 0) as max_workflow_count
		FROM (
			SELECT COUNT(*) as workflow_count
			FROM app_workflow aw
			INNER JOIN app a ON aw.app_id = a.id
			WHERE aw.active = true AND a.active = true
			GROUP BY aw.app_id
		) app_workflow_counts
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&maxCount, query)
	if err != nil {
		impl.logger.Errorw("error getting highest workflow count in app", "err", err)
		return -1
	}

	impl.logger.Debugw("counted highest workflow count in app", "count", maxCount)
	return maxCount
}

// getHighestEnvironmentCountInWorkflow returns the highest number of environments in a single workflow
func (impl *TelemetryEventClientImplExtended) getHighestEnvironmentCountInWorkflow() int {
	var maxCount int
	query := `
		SELECT COALESCE(MAX(env_count), 0) as max_env_count
		FROM (
			SELECT COUNT(DISTINCT p.environment_id) as env_count
			FROM app_workflow_mapping awm
			INNER JOIN pipeline p ON awm.component_id = p.id
			INNER JOIN app_workflow aw ON awm.app_workflow_id = aw.id
			WHERE awm.type = 'CD_PIPELINE'
			AND awm.active = true
			AND aw.active = true
			AND p.deleted = false
			GROUP BY awm.app_workflow_id
		) workflow_env_counts
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&maxCount, query)
	if err != nil {
		impl.logger.Errorw("error getting highest environment count in workflow", "err", err)
		return -1
	}

	impl.logger.Debugw("counted highest environment count in workflow", "count", maxCount)
	return maxCount
}

// getHighestGitRepoCountInApp returns the highest number of git repositories in a single application
func (impl *TelemetryEventClientImplExtended) getHighestGitRepoCountInApp() int {
	var maxCount int
	query := `
		SELECT COALESCE(MAX(git_repo_count), 0) as max_git_repo_count
		FROM (
			SELECT COUNT(*) as git_repo_count
			FROM git_material gm
			INNER JOIN app a ON gm.app_id = a.id
			WHERE gm.active = true AND a.active = true
			GROUP BY gm.app_id
		) app_git_counts
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&maxCount, query)
	if err != nil {
		impl.logger.Errorw("error getting highest git repo count in app", "err", err)
		return -1
	}

	impl.logger.Debugw("counted highest git repo count in app", "count", maxCount)
	return maxCount
}

// getAppsWithIncludeExcludeFilesCount returns the number of applications that have include/exclude file patterns defined
func (impl *TelemetryEventClientImplExtended) getAppsWithIncludeExcludeFilesCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT gm.app_id)
		FROM git_material gm
		INNER JOIN app a ON gm.app_id = a.id
		WHERE gm.active = true
		AND a.active = true
		AND gm.filter_pattern IS NOT NULL
		AND array_length(gm.filter_pattern, 1) > 0
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting apps with include/exclude files count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted apps with include/exclude files", "count", count)
	return count
}

// getAppsWithCreateDockerfileCount returns the number of applications that have dockerfile creation configured
func (impl *TelemetryEventClientImplExtended) getAppsWithCreateDockerfileCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT ct.app_id)
		FROM ci_template ct
		INNER JOIN app a ON ct.app_id = a.id
		INNER JOIN ci_build_config cbc ON ct.ci_build_config_id = cbc.id
		WHERE ct.active = true
		AND a.active = true
		AND cbc.type = 'MANAGED_DOCKERFILE_BUILD_TYPE'
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting apps with create dockerfile count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted apps with create dockerfile", "count", count)
	return count
}

// getDockerfileLanguagesList returns a list of languages being used in create dockerfile configurations
func (impl *TelemetryEventClientImplExtended) getDockerfileLanguagesList() []string {
	var languages []string
	query := `
		SELECT DISTINCT
			CASE
				WHEN cbc.build_metadata::jsonb->>'language' IS NOT NULL
				THEN cbc.build_metadata::jsonb->>'language'
				ELSE 'unknown'
			END as language
		FROM ci_build_config cbc
		INNER JOIN ci_template ct ON cbc.id = ct.ci_build_config_id
		INNER JOIN app a ON ct.app_id = a.id
		WHERE cbc.type = 'MANAGED_DOCKERFILE_BUILD_TYPE'
		AND ct.active = true
		AND a.active = true
		AND cbc.build_metadata IS NOT NULL
		ORDER BY language
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&languages, query)
	if err != nil {
		impl.logger.Errorw("error getting dockerfile languages list", "err", err)
		return []string{}
	}

	impl.logger.Debugw("retrieved dockerfile languages list", "languages", languages)
	return languages
}

// getAppsWithDockerfileCount returns the number of applications that have a dockerfile configured
func (impl *TelemetryEventClientImplExtended) getAppsWithDockerfileCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT ct.app_id)
		FROM ci_template ct
		INNER JOIN app a ON ct.app_id = a.id
		WHERE ct.active = true
		AND a.active = true
		AND ct.dockerfile_path IS NOT NULL
		AND TRIM(ct.dockerfile_path) != ''
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting apps with dockerfile count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted apps with dockerfile", "count", count)
	return count
}

// getAppsWithBuildpacksCount returns the number of applications that use buildpacks
func (impl *TelemetryEventClientImplExtended) getAppsWithBuildpacksCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT ct.app_id)
		FROM ci_template ct
		INNER JOIN app a ON ct.app_id = a.id
		INNER JOIN ci_build_config cbc ON ct.ci_build_config_id = cbc.id
		WHERE ct.active = true
		AND a.active = true
		AND cbc.type = 'BUILDPACK_BUILD_TYPE'
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting apps with buildpacks count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted apps with buildpacks", "count", count)
	return count
}

// getBuildpackLanguagesList returns a list of languages being used in buildpack configurations
func (impl *TelemetryEventClientImplExtended) getBuildpackLanguagesList() []string {
	var languages []string
	query := `
		SELECT DISTINCT
			CASE
				WHEN cbc.build_metadata::jsonb->>'language' IS NOT NULL
				THEN cbc.build_metadata::jsonb->>'language'
				ELSE 'unknown'
			END as language
		FROM ci_build_config cbc
		INNER JOIN ci_template ct ON cbc.id = ct.ci_build_config_id
		INNER JOIN app a ON ct.app_id = a.id
		WHERE cbc.type = 'BUILDPACK_BUILD_TYPE'
		AND ct.active = true
		AND a.active = true
		AND cbc.build_metadata IS NOT NULL
		ORDER BY language
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&languages, query)
	if err != nil {
		impl.logger.Errorw("error getting buildpack languages list", "err", err)
		return []string{}
	}

	impl.logger.Debugw("retrieved buildpack languages list", "languages", languages)
	return languages
}

// getAppsWithDeploymentChartCount returns the number of applications using deployment chart
func (impl *TelemetryEventClientImplExtended) getAppsWithDeploymentChartCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT c.app_id)
		FROM charts c
		INNER JOIN chart_ref cr ON c.chart_repo_id = cr.id
		INNER JOIN app a ON c.app_id = a.id
		WHERE c.active = true
		AND a.active = true
		AND cr.active = true
		AND cr.name = 'Deployment'
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting apps with deployment chart count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted apps with deployment chart", "count", count)
	return count
}

// getAppsWithRolloutChartCount returns the number of applications using rollout chart
func (impl *TelemetryEventClientImplExtended) getAppsWithRolloutChartCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT c.app_id)
		FROM charts c
		INNER JOIN chart_ref cr ON c.chart_repo_id = cr.id
		INNER JOIN app a ON c.app_id = a.id
		WHERE c.active = true
		AND a.active = true
		AND cr.active = true
		AND cr.location LIKE 'reference-chart_%'
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting apps with rollout chart count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted apps with rollout chart", "count", count)
	return count
}

// getAppsWithStatefulsetCount returns the number of applications using statefulset chart
func (impl *TelemetryEventClientImplExtended) getAppsWithStatefulsetCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT c.app_id)
		FROM charts c
		INNER JOIN chart_ref cr ON c.chart_repo_id = cr.id
		INNER JOIN app a ON c.app_id = a.id
		WHERE c.active = true
		AND a.active = true
		AND cr.active = true
		AND cr.name = 'StatefulSet'
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting apps with statefulset count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted apps with statefulset", "count", count)
	return count
}

// getAppsWithJobsCronjobsCount returns the number of applications using jobs & cronjobs chart
func (impl *TelemetryEventClientImplExtended) getAppsWithJobsCronjobsCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT c.app_id)
		FROM charts c
		INNER JOIN chart_ref cr ON c.chart_repo_id = cr.id
		INNER JOIN app a ON c.app_id = a.id
		WHERE c.active = true
		AND a.active = true
		AND cr.active = true
		AND (cr.name = 'Cron Job & Job' OR cr.location LIKE 'cronjob-chart_%')
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting apps with jobs/cronjobs count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted apps with jobs/cronjobs", "count", count)
	return count
}

// getEnvironmentsWithPatchStrategyCount returns the number of environments using patch strategy
func (impl *TelemetryEventClientImplExtended) getEnvironmentsWithPatchStrategyCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT ceco.environment_id)
		FROM chart_env_config_override ceco
		INNER JOIN environment e ON ceco.environment_id = e.id
		WHERE ceco.active = true
		AND e.active = true
		AND ceco.merge_strategy = 'patch'
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting environments with patch strategy count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted environments with patch strategy", "count", count)
	return count
}

// getEnvironmentsWithReplaceStrategyCount returns the number of environments using replace strategy
func (impl *TelemetryEventClientImplExtended) getEnvironmentsWithReplaceStrategyCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT ceco.environment_id)
		FROM chart_env_config_override ceco
		INNER JOIN environment e ON ceco.environment_id = e.id
		WHERE ceco.active = true
		AND e.active = true
		AND ceco.merge_strategy = 'replace'
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting environments with replace strategy count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted environments with replace strategy", "count", count)
	return count
}

// getExternalConfigMapCount returns the number of external configmaps
func (impl *TelemetryEventClientImplExtended) getExternalConfigMapCount() int {
	var count int
	query := `
		SELECT COUNT(*)
		FROM (
			SELECT COUNT(*) as cm_count
			FROM config_map_env_level cmel
			INNER JOIN environment e ON cmel.environment_id = e.id
			WHERE cmel.deleted = false
			AND e.active = true
			AND cmel.config_map_data::jsonb->'ConfigMaps'->>'enabled' = 'true'
			AND EXISTS (
				SELECT 1
				FROM jsonb_array_elements(cmel.config_map_data::jsonb->'ConfigMaps'->'maps') as cm
				WHERE cm->>'external' = 'true'
			)
			GROUP BY cmel.environment_id, cmel.app_id
		) external_cms
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting external configmap count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted external configmaps", "count", count)
	return count
}

// getInternalConfigMapCount returns the number of internal configmaps
func (impl *TelemetryEventClientImplExtended) getInternalConfigMapCount() int {
	var count int
	query := `
		SELECT COUNT(*)
		FROM (
			SELECT COUNT(*) as cm_count
			FROM config_map_env_level cmel
			INNER JOIN environment e ON cmel.environment_id = e.id
			WHERE cmel.deleted = false
			AND e.active = true
			AND cmel.config_map_data::jsonb->'ConfigMaps'->>'enabled' = 'true'
			AND EXISTS (
				SELECT 1
				FROM jsonb_array_elements(cmel.config_map_data::jsonb->'ConfigMaps'->'maps') as cm
				WHERE cm->>'external' = 'false'
			)
			GROUP BY cmel.environment_id, cmel.app_id
		) internal_cms
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting internal configmap count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted internal configmaps", "count", count)
	return count
}
