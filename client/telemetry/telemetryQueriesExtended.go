package telemetry

import (
	"github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	buildBean "github.com/devtron-labs/devtron/pkg/build/pipeline/bean/common"
	chartRefBean "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
	pluginRepository "github.com/devtron-labs/devtron/pkg/plugin/repository"
)

// Policy type constants
const (
	POLICY_TYPE_DEPLOYMENT_WINDOW  = "DEPLOYMENT_WINDOW"
	POLICY_TYPE_APPROVAL           = "APPROVAL"
	POLICY_TYPE_PLUGIN             = "PLUGIN"
	POLICY_TYPE_LOCK_CONFIGURATION = "LOCK_CONFIGURATION"
)

// Deployment type constants
const (
	DEPLOYMENT_TYPE_ARGOCD = "argo_cd"
	DEPLOYMENT_TYPE_HELM   = "helm"
)

// Build type constants
const (
	BUILD_TYPE_MANAGED_DOCKERFILE = "MANAGED_DOCKERFILE_BUILD_TYPE"
	BUILD_TYPE_BUILDPACK          = "BUILDPACK_BUILD_TYPE"
)

// Merge strategy constants
const (
	MERGE_STRATEGY_PATCH   = "patch"
	MERGE_STRATEGY_REPLACE = "replace"
)

// Config map type constants
const (
	CONFIG_MAP_TYPE_EXTERNAL = "true"
	CONFIG_MAP_TYPE_INTERNAL = "false"
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
	jobTriggeredCount := 0
	for _, data := range buildTypeStatusData {
		if data.Type == string(bean.SKIP_BUILD_TYPE) {
			jobTriggeredCount += data.Count
		}
	}

	return jobTriggeredCount
}

func (impl *TelemetryEventClientImplExtended) getJobPipelineSucceededLast24h() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT cp.id)
		FROM ci_pipeline cp
		INNER JOIN app a ON cp.app_id = a.id
		INNER JOIN ci_workflow cw ON cp.id = cw.ci_pipeline_id
		WHERE cp.active = true
		AND cp.deleted = false
		AND a.active = true
		AND cp.ci_pipeline_type = ?
		AND cw.status = 'Succeeded'
		AND cw.created_on >= NOW() - INTERVAL '24 hours'
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query, buildBean.CI_JOB)
	if err != nil {
		impl.logger.Errorw("error getting job pipeline succeeded count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted job pipelines succeeded in last 24h", "count", count)
	return count
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

// getDeploymentWindowPolicyCount returns the count of deployment window policies
func (impl *TelemetryEventClientImplExtended) getDeploymentWindowPolicyCount() int {
	var count int
	query := `
		SELECT COUNT(*)
		FROM global_policy gp
		WHERE gp.policy_of = ? AND gp.deleted = false
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query, POLICY_TYPE_DEPLOYMENT_WINDOW)
	if err != nil {
		impl.logger.Debugw("global_policy table not found or query failed for deployment window policies", "err", err)
		return 0
	}

	impl.logger.Debugw("counted deployment window policies", "count", count)
	return count
}

// getApprovalPolicyCount returns the count of approval policies
func (impl *TelemetryEventClientImplExtended) getApprovalPolicyCount() int {
	var count int
	query := `
		SELECT COUNT(*)
		FROM global_policy gp
		WHERE gp.policy_of = ? AND gp.deleted = false
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query, POLICY_TYPE_APPROVAL)
	if err != nil {
		impl.logger.Debugw("global_policy table not found or query failed", "err", err)
		return 0
	}

	impl.logger.Debugw("counted approval policies", "count", count)
	return count
}

// getPluginPolicyCount returns the count of plugin policies
func (impl *TelemetryEventClientImplExtended) getPluginPolicyCount() int {
	var count int
	query := `
		SELECT COUNT(*)
		FROM global_policy gp
		WHERE gp.policy_of = ? AND gp.deleted = false
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query, POLICY_TYPE_PLUGIN)
	if err != nil {
		impl.logger.Debugw("global_policy table not found or query failed", "err", err)
		return 0
	}
	impl.logger.Debugw("counted plugin policies", "count", count)
	return count
}

// getTagsPolicyCount returns the count of tags policies
func (impl *TelemetryEventClientImplExtended) getTagsPolicyCount() int {
	// Count tag-related policies using available plugin tag repository
	tags, err := impl.pluginRepository.GetAllPluginTags()
	if err != nil {
		impl.logger.Errorw("error getting tags policies", "err", err)
		return -1
	}

	impl.logger.Debugw("counted tags policies", "count", len(tags))
	return len(tags)
}

// getFilterConditionPolicyCount returns the count of filter condition policies
func (impl *TelemetryEventClientImplExtended) getFilterConditionPolicyCount() int {
	// TODO: Implement when filter condition policy repository is available
	// For now, return 0 as placeholder
	var count int
	query := `
		SELECT COUNT(*)
		FROM resource_filter rf
		WHERE rf.deleted = false
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Debugw("filter condition policy table not found or query failed", "err", err)
		return 0
	}

	impl.logger.Debugw("counted filter condition policies", "count", count)
	return count
}

// getLockDeploymentConfigurationPolicyCount returns the count of lock deployment configuration policies
func (impl *TelemetryEventClientImplExtended) getLockDeploymentConfigurationPolicyCount() int {
	var count int
	query := `
		SELECT COUNT(*)
		FROM global_policy gp
		WHERE gp.policy_of = ? AND gp.deleted = false
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query, POLICY_TYPE_LOCK_CONFIGURATION)
	if err != nil {
		impl.logger.Debugw("lock deployment configuration policy table not found or query failed", "err", err)
		return 0
	}

	impl.logger.Debugw("counted lock deployment configuration policies", "count", count)
	return count
}

func (impl *TelemetryEventClientImplExtended) getGitOpsPipelineCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT p.id)
		FROM pipeline p
		WHERE p.deleted = false AND p.deployment_app_type = ?
	`

	dbConnection := impl.cdWorkflowRepository.GetConnection()
	_, err := dbConnection.Query(&count, query, DEPLOYMENT_TYPE_ARGOCD)
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
		WHERE p.deleted = false AND p.deployment_app_type = ?
	`

	dbConnection := impl.cdWorkflowRepository.GetConnection()
	_, err := dbConnection.Query(&count, query, DEPLOYMENT_TYPE_HELM)
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
		AND cp.ci_pipeline_type = ?
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query, buildBean.CI_JOB)
	if err != nil {
		impl.logger.Errorw("error getting job pipeline count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted job pipelines", "count", count)
	return count
}

// getAppliedPolicyRowCount returns the total number of applied policies in the system
func (impl *TelemetryEventClientImplExtended) getAppliedPolicyRowCount() int {
	var count int
	query := `
		WITH policy_counts AS (
			-- Count global policies
			SELECT COUNT(*) as count
			FROM global_policy gp
			WHERE gp.deleted = false
			UNION ALL
			-- Count resource mapped policies
			SELECT COUNT(DISTINCT rqm.resource_id)
			FROM resource_qualifier_mapping rqm
			INNER JOIN resource_qualifier_mapping_criteria rqmc ON rqm.id = rqmc.id
			INNER JOIN global_policy gp ON rqmc.id = gp.id
			WHERE rqm.active = true
			AND gp.deleted = false
		)
		SELECT COALESCE(SUM(count), 0) as total_count
		FROM policy_counts
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query)
	if err != nil {
		impl.logger.Errorw("error getting applied policy count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted applied policies", "count", count)
	return count
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
		FROM (
			SELECT a.id
			FROM app a
			WHERE a.active = true
			AND a.description IS NOT NULL
			AND TRIM(COALESCE(a.description, '')) != ''
			UNION
			SELECT ia.id
			FROM installed_apps ia
			WHERE ia.active = true
			AND ia.notes IS NOT NULL
			AND TRIM(COALESCE(ia.notes, '')) != ''
		) apps_with_description
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
		SELECT COUNT(DISTINCT gn.identifier)
		FROM generic_note gn
		INNER JOIN app a ON gn.identifier = a.id
		WHERE gn.identifier_type = 1 
		AND a.active = true
		AND gn.description IS NOT NULL
		AND TRIM(gn.description) != ''
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
		AND jsonb_array_length(gm.filter_pattern::jsonb) > 0
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
		WHERE cbc.type = ?
		AND ct.active = true
		AND a.active = true
		AND cbc.build_metadata IS NOT NULL
		ORDER BY language
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&languages, query, BUILD_TYPE_MANAGED_DOCKERFILE)
	if err != nil {
		impl.logger.Errorw("error getting dockerfile languages list", "err", err)
		return []string{}
	}

	impl.logger.Debugw("retrieved dockerfile languages list", "languages", languages)
	return languages
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
		WHERE cbc.type = ?
		AND ct.active = true
		AND a.active = true
		AND cbc.build_metadata IS NOT NULL
		ORDER BY language
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&languages, query, BUILD_TYPE_BUILDPACK)
	if err != nil {
		impl.logger.Errorw("error getting buildpack languages list", "err", err)
		return []string{}
	}

	impl.logger.Debugw("retrieved buildpack languages list", "languages", languages)
	return languages
}

// getAppsWithChartTypeCount returns the number of applications using a specific chart type
func (impl *TelemetryEventClientImplExtended) getAppsWithChartTypeCount(chartType string) int {
	var count int
	query := `
		SELECT COUNT(DISTINCT c.app_id)
		FROM charts c
		INNER JOIN chart_ref cr ON c.chart_ref_id = cr.id
		INNER JOIN app a ON c.app_id = a.id
		WHERE c.active = true
		AND a.active = true
		AND cr.active = true
		AND cr.name = ?
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query, chartType)
	if err != nil {
		impl.logger.Errorw("error getting apps with chart type count", "chartType", chartType, "err", err)
		return -1
	}

	impl.logger.Debugw("counted apps with chart type", "chartType", chartType, "count", count)
	return count
}

// getAppsWithDeploymentChartCount returns the number of applications using deployment chart
func (impl *TelemetryEventClientImplExtended) getAppsWithDeploymentChartCount() int {
	return impl.getAppsWithChartTypeCount(chartRefBean.DeploymentChartType)
}

// getAppsWithRolloutChartCount returns the number of applications using rollout chart
func (impl *TelemetryEventClientImplExtended) getAppsWithRolloutChartCount() int {
	return impl.getAppsWithChartTypeCount(chartRefBean.RolloutChartType)
}

// getAppsWithStatefulsetCount returns the number of applications using statefulset chart
func (impl *TelemetryEventClientImplExtended) getAppsWithStatefulsetCount() int {
	return impl.getAppsWithChartTypeCount(chartRefBean.StatefulSetChartType)
}

// getAppsWithJobsCronjobsCount returns the number of applications using jobs & cronjobs chart
func (impl *TelemetryEventClientImplExtended) getAppsWithJobsCronjobsCount() int {
	return impl.getAppsWithChartTypeCount(chartRefBean.JobsCronjobsChartType)
}

// getEnvironmentsWithPatchStrategyCount returns the number of environments using patch strategy
func (impl *TelemetryEventClientImplExtended) getEnvironmentsWithPatchStrategyCount() int {
	var count int
	query := `
		SELECT COUNT(DISTINCT ceco.target_environment)
		FROM chart_env_config_override ceco
		INNER JOIN environment e ON ceco.target_environment = e.id
		WHERE ceco.active = true
		AND e.active = true
		AND ceco.merge_strategy = ?
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query, MERGE_STRATEGY_PATCH)
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
		SELECT COUNT(DISTINCT ceco.target_environment)
		FROM chart_env_config_override ceco
		INNER JOIN environment e ON ceco.target_environment = e.id
		WHERE ceco.active = true
		AND e.active = true
		AND ceco.merge_strategy = ?
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query, MERGE_STRATEGY_REPLACE)
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
		SELECT COALESCE(SUM(cm_count), 0)
		FROM (
			-- Count app-level configmaps
			SELECT COUNT(*) as cm_count
			FROM config_map_app_level cmal
			INNER JOIN app a ON cmal.app_id = a.id
			WHERE a.active = true
			AND EXISTS (
				SELECT 1
				FROM jsonb_array_elements(cmal.config_map_data::jsonb->'maps') as cm
				WHERE cm->>'external' = ?
			)
			GROUP BY cmal.app_id
			UNION ALL
			-- Count environment-level configmaps
			SELECT COUNT(*) as cm_count
			FROM config_map_env_level cmel
			INNER JOIN environment e ON cmel.environment_id = e.id
			WHERE cmel.deleted = false
			AND e.active = true
			AND EXISTS (
				SELECT 1
				FROM jsonb_array_elements(cmel.config_map_data::jsonb->'maps') as cm
				WHERE cm->>'external' = ?
			)
			GROUP BY cmel.environment_id, cmel.app_id
		) external_cms
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query, CONFIG_MAP_TYPE_EXTERNAL, CONFIG_MAP_TYPE_EXTERNAL)
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
		SELECT COALESCE(SUM(cm_count), 0)
		FROM (
			-- Count app-level configmaps
			SELECT COUNT(*) as cm_count
			FROM config_map_app_level cmal
			INNER JOIN app a ON cmal.app_id = a.id
			WHERE a.active = true
			AND EXISTS (
				SELECT 1
				FROM jsonb_array_elements(cmal.config_map_data::jsonb->'maps') as cm
				WHERE cm->>'external' = ?
			)
			GROUP BY cmal.app_id
			UNION ALL
			-- Count environment-level configmaps
			SELECT COUNT(*) as cm_count
			FROM config_map_env_level cmel
			INNER JOIN environment e ON cmel.environment_id = e.id
			WHERE cmel.deleted = false
			AND e.active = true
			AND EXISTS (
				SELECT 1
				FROM jsonb_array_elements(cmel.config_map_data::jsonb->'maps') as cm
				WHERE cm->>'external' = ?
			)
			GROUP BY cmel.environment_id, cmel.app_id
		) internal_cms
	`

	dbConnection := impl.appRepository.GetConnection()
	_, err := dbConnection.Query(&count, query, CONFIG_MAP_TYPE_INTERNAL, CONFIG_MAP_TYPE_INTERNAL)
	if err != nil {
		impl.logger.Errorw("error getting internal configmap count", "err", err)
		return -1
	}

	impl.logger.Debugw("counted internal configmaps", "count", count)
	return count
}
