/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package repository

import (
	"time"

	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ImageScanExecutionResult struct {
	tableName                   struct{} `sql:"image_scan_execution_result" pg:",discard_unknown_columns"`
	Id                          int      `sql:"id,pk"`
	CveStoreName                string   `sql:"cve_store_name,notnull"`
	ImageScanExecutionHistoryId int      `sql:"image_scan_execution_history_id"`
	ScanToolId                  int      `sql:"scan_tool_id"`
	Package                     string   `sql:"package"`
	Version                     string   `sql:"version"`
	FixedVersion                string   `sql:"fixed_version"`
	Target                      string   `sql:"target"`
	Type                        string   `sql:"type"`
	Class                       string   `sql:"class"`
	CveStore                    CveStore
	ImageScanExecutionHistory   ImageScanExecutionHistory
}

type VulnerabilityData struct {
	CveStoreName string
	FixedVersion string
}

type SeverityInsightData struct {
	CveStoreName  string
	Severity      int       // Severity enum value from cve_store
	ExecutionTime time.Time // From image_scan_execution_history
}

type VulnerabilityTrendData struct {
	CveStoreName  string
	Severity      int       // Severity enum value from cve_store
	ExecutionTime time.Time // From image_scan_execution_history
}

type VulnerabilityListingData struct {
	CveStoreName   string    `sql:"cve_store_name"`
	Severity       int       `sql:"severity"`
	AppId          int       `sql:"app_id"`
	AppName        string    `sql:"app_name"`
	EnvId          int       `sql:"env_id"`
	EnvName        string    `sql:"env_name"`
	DiscoveredAt   time.Time `sql:"discovered_at"`
	Package        string    `sql:"package"`
	CurrentVersion string    `sql:"current_version"`
	FixedVersion   string    `sql:"fixed_version"`
	TotalCount     int       `sql:"total_count"`
}

// VulnerabilityRawData represents raw CVE data before aggregation (for code-level optimization)
type VulnerabilityRawData struct {
	CveStoreName   string    `sql:"cve_store_name"`
	Severity       int       `sql:"severity"`
	AppId          int       `sql:"app_id"`
	AppName        string    `sql:"app_name"`
	EnvId          int       `sql:"env_id"`
	EnvName        string    `sql:"env_name"`
	ExecutionTime  time.Time `sql:"execution_time"`
	Package        string    `sql:"package"`
	CurrentVersion string    `sql:"current_version"`
	FixedVersion   string    `sql:"fixed_version"`
}

type ImageScanResultRepository interface {
	Save(model *ImageScanExecutionResult) error
	FindAll() ([]*ImageScanExecutionResult, error)
	FindOne(id int) (*ImageScanExecutionResult, error)
	FindByCveName(name string) ([]*ImageScanExecutionResult, error)
	Update(model *ImageScanExecutionResult) error
	FetchByScanExecutionId(id int) ([]*ImageScanExecutionResult, error)
	FetchByScanExecutionIds(ids []int) ([]*ImageScanExecutionResult, error)
	FindByImageDigest(imageDigest string) ([]*ImageScanExecutionResult, error)
	FindByImageDigests(digest []string) ([]*ImageScanExecutionResult, error)
	FindByImage(image string) ([]*ImageScanExecutionResult, error)

	// Security Overview methods
	GetVulnerabilitiesWithFixedVersionByFilters(envIds, clusterIds, appIds []int) ([]*VulnerabilityData, error)
	GetSeverityInsightDataByFilters(envIds, clusterIds, appIds []int, isProd *bool) ([]*SeverityInsightData, error)
	GetVulnerabilityTrendDataByFilters(from, to *time.Time, isProd *bool) ([]*VulnerabilityTrendData, error)

	// Vulnerability Listing
	GetVulnerabilityRawData(cveName string, severities, envIds, clusterIds, appIds, deployInfoIds []int) ([]*VulnerabilityRawData, error)
}

type ImageScanResultRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewImageScanResultRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *ImageScanResultRepositoryImpl {
	return &ImageScanResultRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl ImageScanResultRepositoryImpl) Save(model *ImageScanExecutionResult) error {
	err := impl.dbConnection.Insert(model)
	return err
}

func (impl ImageScanResultRepositoryImpl) FindAll() ([]*ImageScanExecutionResult, error) {
	var models []*ImageScanExecutionResult
	err := impl.dbConnection.Model(&models).Select()
	return models, err
}

func (impl ImageScanResultRepositoryImpl) FindOne(id int) (*ImageScanExecutionResult, error) {
	var model *ImageScanExecutionResult
	err := impl.dbConnection.Model(&model).
		Where("id = ?", id).Select()
	return model, err
}

func (impl ImageScanResultRepositoryImpl) FindByCveName(name string) ([]*ImageScanExecutionResult, error) {
	var model []*ImageScanExecutionResult
	err := impl.dbConnection.Model(&model).
		Where("cve_store_name = ?", name).Select()
	return model, err
}

func (impl ImageScanResultRepositoryImpl) Update(team *ImageScanExecutionResult) error {
	err := impl.dbConnection.Update(team)
	return err
}

func (impl ImageScanResultRepositoryImpl) FetchByScanExecutionId(scanExecutionId int) ([]*ImageScanExecutionResult, error) {
	var models []*ImageScanExecutionResult
	/*err := impl.dbConnection.Model(&models).Column("image_scan_execution_result.*", "cs.*").
	Join("inner join cve_store cs on cs.name=image_scan_execution_result.cve_name").
	Where("image_scan_execution_result.scan_execution_id = ?", id).Select()
	*/

	err := impl.dbConnection.Model(&models).Column("image_scan_execution_result.*", "CveStore").
		Where("image_scan_execution_result.image_scan_execution_history_id = ?", scanExecutionId).
		Select()
	return models, err
}

func (impl ImageScanResultRepositoryImpl) FetchByScanExecutionIds(ids []int) ([]*ImageScanExecutionResult, error) {
	var models []*ImageScanExecutionResult
	err := impl.dbConnection.Model(&models).Column("image_scan_execution_result.*", "ImageScanExecutionHistory", "CveStore").
		Where("image_scan_execution_result.image_scan_execution_history_id in(?)", pg.In(ids)).
		Select()
	return models, err
}

func (impl ImageScanResultRepositoryImpl) FindByImageDigest(imageDigest string) ([]*ImageScanExecutionResult, error) {
	var model []*ImageScanExecutionResult
	err := impl.dbConnection.Model(&model).Column("image_scan_execution_result.*", "ImageScanExecutionHistory", "CveStore").
		Where("image_scan_execution_history.image_hash = ?", imageDigest).Order("image_scan_execution_history.execution_time desc").Select()
	return model, err
}

func (impl ImageScanResultRepositoryImpl) FindByImageDigests(digest []string) ([]*ImageScanExecutionResult, error) {
	var models []*ImageScanExecutionResult
	err := impl.dbConnection.Model(&models).Column("image_scan_execution_result.*", "ImageScanExecutionHistory", "CveStore").
		Where("image_hash in (?)", pg.In(digest)).Order("execution_time desc").Select()
	return models, err
}

func (impl ImageScanResultRepositoryImpl) FindByImage(image string) ([]*ImageScanExecutionResult, error) {
	var model []*ImageScanExecutionResult
	err := impl.dbConnection.Model(&model).Column("image_scan_execution_result.*", "ImageScanExecutionHistory", "CveStore").
		Where("image_scan_execution_history.image = ?", image).Order("image_scan_execution_history.execution_time desc").Select()
	return model, err
}

// ============================================================================
// Security Overview Methods
// ============================================================================

func (impl ImageScanResultRepositoryImpl) GetVulnerabilitiesWithFixedVersionByFilters(envIds, clusterIds, appIds []int) ([]*VulnerabilityData, error) {
	var results []*VulnerabilityData

	query := `
		WITH LatestDeployments AS (
			SELECT
				p.app_id,
				p.environment_id,
				env.cluster_id,
				cia.image,
				ROW_NUMBER() OVER (PARTITION BY p.app_id, p.environment_id ORDER BY cwr.id DESC) AS rn
			FROM cd_workflow_runner cwr
			INNER JOIN cd_workflow cw ON cw.id = cwr.cd_workflow_id
			INNER JOIN pipeline p ON p.id = cw.pipeline_id
			INNER JOIN environment env ON env.id = p.environment_id
			INNER JOIN ci_artifact cia ON cia.id = cw.ci_artifact_id
			WHERE cwr.workflow_type = 'DEPLOY'
			AND p.deleted = false
			AND env.active = true
	`

	var queryParams []interface{}

	// Add filters to CTE
	if len(envIds) > 0 {
		query += " AND p.environment_id = ANY(?)"
		queryParams = append(queryParams, pg.Array(envIds))
	}

	if len(clusterIds) > 0 {
		query += " AND env.cluster_id = ANY(?)"
		queryParams = append(queryParams, pg.Array(clusterIds))
	}

	if len(appIds) > 0 {
		query += " AND p.app_id = ANY(?)"
		queryParams = append(queryParams, pg.Array(appIds))
	}

	// Complete the CTE and join with image_scan_deploy_info to get only scanned deployments
	// Then fetch vulnerabilities from image_scan_execution_result
	query += `
		)
		SELECT
			iser.cve_store_name,
			iser.fixed_version
		FROM LatestDeployments ld
		INNER JOIN image_scan_deploy_info isdi
			ON isdi.scan_object_meta_id = ld.app_id
			AND isdi.env_id = ld.environment_id
			AND isdi.object_type = 'app'
		INNER JOIN image_scan_execution_result iser
			ON iser.image_scan_execution_history_id = isdi.image_scan_execution_history_id[1]
		INNER JOIN image_scan_execution_history iseh
			ON iseh.id = isdi.image_scan_execution_history_id[1]
		WHERE ld.image = iseh.image
		AND ld.rn = 1
		AND isdi.image_scan_execution_history_id[1] != -1
	`

	_, err := impl.dbConnection.Query(&results, query, queryParams...)
	if err != nil {
		impl.logger.Errorw("error in getting vulnerabilities with fixed version", "err", err)
		return nil, err
	}

	return results, nil
}

// GetSeverityInsightDataByFilters returns vulnerability data with severity and execution time
// for calculating severity distribution and age distribution in a single query
// Only returns vulnerabilities from the LATEST deployed artifact for each app+env combination
// isProd: nil = all environments, true = prod only, false = non-prod only
func (impl ImageScanResultRepositoryImpl) GetSeverityInsightDataByFilters(envIds, clusterIds, appIds []int, isProd *bool) ([]*SeverityInsightData, error) {
	var results []*SeverityInsightData

	// Query to get vulnerabilities from latest deployed images per app+env
	// Step 1: Get latest deployment per app+env from cd_workflow_runner (source of truth for all deployments)
	// Step 2: Join with image_scan_deploy_info to verify if this app+env has scanned image deployed
	//         image_scan_deploy_info contains env_id mapping and scan_execution_history_id for scanned images
	//         For object_type='app', the array image_scan_execution_history_id has length 1 (current deployed image's scan)
	// Step 3: Get execution_time from image_scan_execution_history for age distribution
	// Step 4: Fetch vulnerabilities with severity from image_scan_execution_result
	// Images without scan data (not in image_scan_deploy_info) will not appear in results (zero vulnerabilities)
	query := `
		WITH LatestDeployments AS (
			SELECT
				p.app_id,
				p.environment_id,
				env.cluster_id,
				cia.image,
				ROW_NUMBER() OVER (PARTITION BY p.app_id, p.environment_id ORDER BY cwr.id DESC) AS rn
			FROM cd_workflow_runner cwr
			INNER JOIN cd_workflow cw ON cw.id = cwr.cd_workflow_id
			INNER JOIN pipeline p ON p.id = cw.pipeline_id
			INNER JOIN environment env ON env.id = p.environment_id
			INNER JOIN ci_artifact cia ON cia.id = cw.ci_artifact_id
			WHERE cwr.workflow_type = 'DEPLOY'
			AND p.deleted = false
			AND env.active = true
	`

	var queryParams []interface{}

	// Add prod/non-prod filter only if isProd is not nil
	if isProd != nil {
		query += " AND env.default = ?"
		queryParams = append(queryParams, *isProd)
	}

	// Add filters to CTE
	if len(envIds) > 0 {
		query += " AND p.environment_id = ANY(?)"
		queryParams = append(queryParams, pg.Array(envIds))
	}

	if len(clusterIds) > 0 {
		query += " AND env.cluster_id = ANY(?)"
		queryParams = append(queryParams, pg.Array(clusterIds))
	}

	if len(appIds) > 0 {
		query += " AND p.app_id = ANY(?)"
		queryParams = append(queryParams, pg.Array(appIds))
	}

	// Complete the CTE and join with image_scan_deploy_info to get only scanned deployments
	// Then fetch vulnerabilities with severity and execution_time
	query += `
		)
		SELECT
			iser.cve_store_name,
			COALESCE(cs.standard_severity, cs.severity) as severity,
			iseh.execution_time
		FROM LatestDeployments ld
		INNER JOIN image_scan_deploy_info isdi
			ON isdi.scan_object_meta_id = ld.app_id
			AND isdi.env_id = ld.environment_id
			AND isdi.object_type = 'app'
		INNER JOIN image_scan_execution_history iseh
			ON iseh.id = isdi.image_scan_execution_history_id[1]
			AND iseh.image = ld.image
		INNER JOIN image_scan_execution_result iser
			ON iser.image_scan_execution_history_id = iseh.id
		INNER JOIN cve_store cs ON cs.name = iser.cve_store_name
		WHERE ld.rn = 1
		AND isdi.image_scan_execution_history_id[1] != -1
	`

	_, err := impl.dbConnection.Query(&results, query, queryParams...)
	if err != nil {
		impl.logger.Errorw("error in getting severity insight data", "err", err, "isProd", isProd)
		return nil, err
	}

	return results, nil
}

// GetVulnerabilityTrendDataByFilters returns vulnerability data with severity and execution time
// for calculating time-series vulnerability trend grouped by severity
// Only returns vulnerabilities from the LATEST deployed artifact for each app+env combination
// isProd: nil = all environments, true = prod only, false = non-prod only
func (impl ImageScanResultRepositoryImpl) GetVulnerabilityTrendDataByFilters(from, to *time.Time, isProd *bool) ([]*VulnerabilityTrendData, error) {
	var results []*VulnerabilityTrendData

	// Query to get vulnerabilities from latest deployed images per app+env
	// Step 1: Get latest deployment per app+env from cd_workflow_runner (source of truth for all deployments)
	// Step 2: Join with image_scan_deploy_info to verify if this app+env has scanned image deployed
	//         image_scan_deploy_info contains env_id mapping and scan_execution_history_id for scanned images
	//         For object_type='app', the array image_scan_execution_history_id has length 1 (current deployed image's scan)
	// Step 3: Get execution_time from image_scan_execution_history for trend analysis
	// Step 4: Fetch vulnerabilities with severity from image_scan_execution_result
	// Images without scan data (not in image_scan_deploy_info) will not appear in results (zero vulnerabilities)
	// Filters by execution_time range for trend analysis
	query := `
		WITH LatestDeployments AS (
			SELECT
				p.app_id,
				p.environment_id,
				env.cluster_id,
				cia.image,
				ROW_NUMBER() OVER (PARTITION BY p.app_id, p.environment_id ORDER BY cwr.id DESC) AS rn
			FROM cd_workflow_runner cwr
			INNER JOIN cd_workflow cw ON cw.id = cwr.cd_workflow_id
			INNER JOIN pipeline p ON p.id = cw.pipeline_id
			INNER JOIN environment env ON env.id = p.environment_id
			INNER JOIN ci_artifact cia ON cia.id = cw.ci_artifact_id
			WHERE cwr.workflow_type = 'DEPLOY'
			AND p.deleted = false
			AND env.active = true
	`

	var queryParams []interface{}

	// Add prod/non-prod filter only if isProd is not nil
	if isProd != nil {
		query += " AND env.default = ?"
		queryParams = append(queryParams, *isProd)
	}

	// Complete the CTE and join with image_scan_deploy_info to get only scanned deployments
	// Then fetch vulnerabilities with severity and execution_time, filtered by time range
	query += `
		)
		SELECT
			iser.cve_store_name,
			COALESCE(cs.standard_severity, cs.severity) as severity,
			iseh.execution_time
		FROM LatestDeployments ld
		INNER JOIN image_scan_deploy_info isdi
			ON isdi.scan_object_meta_id = ld.app_id
			AND isdi.env_id = ld.environment_id
			AND isdi.object_type = 'app'
		INNER JOIN image_scan_execution_history iseh
			ON iseh.id = isdi.image_scan_execution_history_id[1]
			AND iseh.image = ld.image
		INNER JOIN image_scan_execution_result iser
			ON iser.image_scan_execution_history_id = iseh.id
		INNER JOIN cve_store cs ON cs.name = iser.cve_store_name
		WHERE ld.rn = 1
		AND isdi.image_scan_execution_history_id[1] != -1
		AND iseh.execution_time >= ? AND iseh.execution_time <= ?
	`

	queryParams = append(queryParams, from, to)

	_, err := impl.dbConnection.Query(&results, query, queryParams...)
	if err != nil {
		impl.logger.Errorw("error in getting vulnerability trend data", "err", err, "from", from, "to", to, "isProd", isProd)
		return nil, err
	}

	return results, nil
}

func (impl ImageScanResultRepositoryImpl) GetVulnerabilityRawData(cveName string, severities, envIds, clusterIds, appIds, deployInfoIds []int) ([]*VulnerabilityRawData, error) {
	var results []*VulnerabilityRawData

	query := `
		WITH LatestDeployments AS (
			SELECT DISTINCT ON (p.app_id, p.environment_id)
				p.app_id,
				a.app_name,
				p.environment_id,
				env.environment_name as env_name,
				env.cluster_id,
				cia.image
			FROM cd_workflow_runner cwr
			INNER JOIN cd_workflow cw ON cw.id = cwr.cd_workflow_id
			INNER JOIN pipeline p ON p.id = cw.pipeline_id
			INNER JOIN app a ON a.id = p.app_id
			INNER JOIN environment env ON env.id = p.environment_id
			INNER JOIN ci_artifact cia ON cia.id = cw.ci_artifact_id
			WHERE cwr.workflow_type = 'DEPLOY'
			AND p.deleted = false
			AND a.active = true
			AND env.active = true
	`

	var queryParams []interface{}

	// Add filters to CTE
	if len(envIds) > 0 {
		query += " AND p.environment_id = ANY(?)"
		queryParams = append(queryParams, pg.Array(envIds))
	}

	if len(clusterIds) > 0 {
		query += " AND env.cluster_id = ANY(?)"
		queryParams = append(queryParams, pg.Array(clusterIds))
	}

	if len(appIds) > 0 {
		query += " AND p.app_id = ANY(?)"
		queryParams = append(queryParams, pg.Array(appIds))
	}

	query += `
			ORDER BY p.app_id, p.environment_id, cwr.id DESC
		)
		SELECT
			iser.cve_store_name,
			COALESCE(cs.standard_severity, cs.severity) as severity,
			ld.app_id,
			ld.app_name,
			ld.environment_id as env_id,
			ld.env_name,
			iseh.execution_time,
			iser.package,
			iser.version as current_version,
			iser.fixed_version
		FROM LatestDeployments ld
		INNER JOIN image_scan_deploy_info isdi
			ON isdi.scan_object_meta_id = ld.app_id
			AND isdi.env_id = ld.environment_id
			AND isdi.object_type = 'app'
			AND isdi.image_scan_execution_history_id[1] != -1
		INNER JOIN image_scan_execution_history iseh
			ON iseh.id = isdi.image_scan_execution_history_id[1]
			AND iseh.image = ld.image
		INNER JOIN image_scan_execution_result iser
			ON iser.image_scan_execution_history_id = iseh.id
	`

	// Add CVE name filter
	if cveName != "" {
		query += " AND iser.cve_store_name ILIKE ?"
		queryParams = append(queryParams, "%"+cveName+"%")
	}

	query += `
		INNER JOIN cve_store cs ON cs.name = iser.cve_store_name
	`

	// Add RBAC filter for deploy info IDs
	if len(deployInfoIds) > 0 {
		query += " WHERE isdi.id = ANY(?)"
		queryParams = append(queryParams, pg.Array(deployInfoIds))

		// Add severity filter with AND since WHERE already exists
		if len(severities) > 0 {
			query += " AND COALESCE(cs.standard_severity, cs.severity) = ANY(?)"
			queryParams = append(queryParams, pg.Array(severities))
		}
	} else {
		// Add severity filter with WHERE since no deploy info filter
		if len(severities) > 0 {
			query += " WHERE COALESCE(cs.standard_severity, cs.severity) = ANY(?)"
			queryParams = append(queryParams, pg.Array(severities))
		}
	}

	_, err := impl.dbConnection.Query(&results, query, queryParams...)
	if err != nil {
		impl.logger.Errorw("error in getting vulnerability raw data", "err", err, "cveName", cveName, "severities", severities)
		return nil, err
	}

	return results, nil
}
