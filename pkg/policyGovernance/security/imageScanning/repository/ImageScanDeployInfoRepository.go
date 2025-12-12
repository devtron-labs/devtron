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
	"fmt"
	"time"

	repoBean "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util"

	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

/*
*
this table contains scanned images registry for deployed object and apps,
images which are deployed on cluster by anyway and has scanned result
*/
type ImageScanDeployInfo struct {
	tableName                   struct{} `sql:"image_scan_deploy_info" pg:",discard_unknown_columns"`
	Id                          int      `sql:"id,pk"`
	ImageScanExecutionHistoryId []int    `sql:"image_scan_execution_history_id,notnull" pg:",array"`
	ScanObjectMetaId            int      `sql:"scan_object_meta_id,notnull"`
	ObjectType                  string   `sql:"object_type,notnull"`
	EnvId                       int      `sql:"env_id,notnull"`
	ClusterId                   int      `sql:"cluster_id,notnull"`
	sql.AuditLog
}

func (r *ImageScanDeployInfo) IsObjectTypeApp() bool {
	return r.ObjectType == ScanObjectType_APP
}

const (
	ScanObjectType_APP   string = "app"
	ScanObjectType_CHART string = "chart"
	ScanObjectType_POD   string = "pod"
)

type ImageScanListingResponse struct {
	Id               int       `json:"id"`
	ScanObjectMetaId int       `json:"scanObjectMetaId"`
	ObjectName       string    `json:"objectName"`
	ObjectType       string    `json:"objectType"`
	SecurityScan     string    `json:"securityScan"`
	EnvironmentName  string    `json:"environmentName"`
	LastChecked      time.Time `json:"lastChecked"`
	TotalCount       int       `json:"totalCount"`
}

type DeploymentScannedCount struct {
	ScannedCount   int
	UnscannedCount int
}

type ImageScanDeployInfoRepository interface {
	Save(model *ImageScanDeployInfo) error
	FindAll() ([]*ImageScanDeployInfo, error)
	FindOne(id int) (*ImageScanDeployInfo, error)
	FindByIds(ids []int) ([]*ImageScanDeployInfo, error)
	Update(model *ImageScanDeployInfo) error
	FetchListingGroupByObject(size int, offset int) ([]*ImageScanDeployInfo, error)
	FetchByAppIdAndEnvId(appId int, envId int, objectType []string) (*ImageScanDeployInfo, error)
	FindByTypeMetaAndTypeId(scanObjectMetaId int, objectType string) (*ImageScanDeployInfo, error)
	ScanListingWithFilter(request *repoBean.ImageScanFilter, size int, offset int, deployInfoIds []int) ([]*ImageScanListingResponse, error)

	// Security Overview methods
	GetActiveDeploymentCountByFilters(envIds, clusterIds, appIds []int) (int, error)
	GetActiveDeploymentCountWithVulnerabilitiesByFilters(envIds, clusterIds, appIds []int) (int, error)
	GetActiveDeploymentScannedUnscannedCountByFilters(envIds, clusterIds, appIds []int) (*DeploymentScannedCount, error)
	GetNonScannedAppEnvCombinations(request *repoBean.ImageScanFilter, size int, offset int, deployInfoIds []int) ([]*ImageScanDeployInfo, error)
	GetNonScannedAppEnvCombinationsCount(request *repoBean.ImageScanFilter, deployInfoIds []int) (int, error)
}

type ImageScanDeployInfoRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewImageScanDeployInfoRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *ImageScanDeployInfoRepositoryImpl {
	return &ImageScanDeployInfoRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl ImageScanDeployInfoRepositoryImpl) Save(model *ImageScanDeployInfo) error {
	err := impl.dbConnection.Insert(model)
	return err
}

func (impl ImageScanDeployInfoRepositoryImpl) FindAll() ([]*ImageScanDeployInfo, error) {
	var models []*ImageScanDeployInfo
	err := impl.dbConnection.Model(&models).Select()
	return models, err
}

func (impl ImageScanDeployInfoRepositoryImpl) FindOne(id int) (*ImageScanDeployInfo, error) {
	var model ImageScanDeployInfo
	err := impl.dbConnection.Model(&model).
		Where("id = ?", id).Select()
	return &model, err
}

func (impl ImageScanDeployInfoRepositoryImpl) FindByIds(ids []int) ([]*ImageScanDeployInfo, error) {
	var models []*ImageScanDeployInfo
	err := impl.dbConnection.Model(&models).Where("id in (?)", pg.In(ids)).Where("image_scan_execution_history_id is not null").Select()
	return models, err
}

func (impl ImageScanDeployInfoRepositoryImpl) Update(model *ImageScanDeployInfo) error {
	err := impl.dbConnection.Update(model)
	return err
}

func (impl ImageScanDeployInfoRepositoryImpl) FetchListingGroupByObject(size int, offset int) ([]*ImageScanDeployInfo, error) {
	var models []*ImageScanDeployInfo
	query := "select scan_object_meta_id,object_type, max(id) as id from image_scan_deploy_info" +
		" group by scan_object_meta_id,object_type order by id desc limit ? offset ?"
	_, err := impl.dbConnection.Query(&models, query, size, offset)
	if err != nil {
		impl.logger.Error("err", err)
		return []*ImageScanDeployInfo{}, err
	}
	return models, err
}

func (impl ImageScanDeployInfoRepositoryImpl) FetchByAppIdAndEnvId(appId int, envId int, objectType []string) (*ImageScanDeployInfo, error) {
	var model ImageScanDeployInfo
	err := impl.dbConnection.Model(&model).
		Where("scan_object_meta_id = ?", appId).
		Where("env_id = ?", envId).Where("object_type in (?)", pg.In(objectType)).
		Order("created_on desc").Limit(1).
		Select()
	return &model, err
}

func (impl ImageScanDeployInfoRepositoryImpl) FindByTypeMetaAndTypeId(scanObjectMetaId int, objectType string) (*ImageScanDeployInfo, error) {
	var model ImageScanDeployInfo
	err := impl.dbConnection.Model(&model).
		Where("scan_object_meta_id = ?", scanObjectMetaId).
		Where("object_type = ?", objectType).
		Select()
	return &model, err
}

func (impl ImageScanDeployInfoRepositoryImpl) ScanListingWithFilter(request *repoBean.ImageScanFilter, size int, offset int, deployInfoIds []int) ([]*ImageScanListingResponse, error) {
	var models []*ImageScanListingResponse
	query, queryParams := impl.scanListingQueryBuilder(request, size, offset, deployInfoIds)
	_, err := impl.dbConnection.Query(&models, query, queryParams...)
	if err != nil {
		impl.logger.Error("err", err)
		return []*ImageScanListingResponse{}, err
	}
	return models, err
}

func (impl ImageScanDeployInfoRepositoryImpl) scanListQueryWithoutObject(request *repoBean.ImageScanFilter, size int, offset int, deployInfoIds []int) (string, []interface{}) {
	var queryParams []interface{}
	query := `select info.scan_object_meta_id,a.app_name as object_name, info.object_type, env.environment_name, max(info.id) as id, COUNT(*) OVER() AS total_count 
				 from image_scan_deploy_info info `
	if len(request.CVEName) > 0 || len(request.Severity) > 0 {
		query = query + ` INNER JOIN image_scan_execution_history his on his.id = any (info.image_scan_execution_history_id) 
						  INNER JOIN image_scan_execution_result res on res.image_scan_execution_history_id=his.id 
		                  INNER JOIN cve_store cs on cs.name= res.cve_store_name`
	}
	query = query + ` INNER JOIN environment env on env.id=info.env_id 
	 				  INNER JOIN cluster clus on clus.id=env.cluster_id 
	                  LEFT JOIN app a on a.id = info.scan_object_meta_id and info.object_type='app' WHERE a.active=true 
	                  AND info.scan_object_meta_id > 0 and env.active=true and info.image_scan_execution_history_id[1] != -1`
	if len(deployInfoIds) > 0 {
		query += " AND info.id IN (?) "
		queryParams = append(queryParams, pg.In(deployInfoIds))
	}
	if len(request.CVEName) > 0 {
		query += " AND res.cve_store_name ILIKE ?"
		queryParams = append(queryParams, util.GetLIKEClauseQueryParam(request.CVEName))
	}
	if len(request.Severity) > 0 {
		// use pg.In to inject values here wherever calling this func in case severity exists, to avoid sql injections
		query = query + " AND (cs.standard_severity IN (?) OR (cs.severity IN (?) AND cs.standard_severity IS NULL))"
		queryParams = append(queryParams, pg.In(request.Severity), pg.In(request.Severity))
	}
	if len(request.EnvironmentIds) > 0 {
		query += " AND env.id IN (?)"
		queryParams = append(queryParams, pg.In(request.EnvironmentIds))
	}
	if len(request.ClusterIds) > 0 {
		query += " AND clus.id IN (?)"
		queryParams = append(queryParams, pg.In(request.ClusterIds))
	}
	query = query + " GROUP BY info.scan_object_meta_id, a.app_name, info.object_type, env.environment_name"
	queryTemp, queryParamsTemp := getOrderByQueryPart(request.SortBy, request.SortOrder)
	query += queryTemp
	queryParams = append(queryParams, queryParamsTemp...)
	if size > 0 {
		query = query + " LIMIT ? OFFSET ? "
		queryParams = append(queryParams, size, offset)
	}
	query = query + " ;"
	return query, queryParams
}

func getOrderByQueryPart(sortBy repoBean.SortBy, sortOrder repoBean.SortOrder) (string, []interface{}) {
	var queryParams []interface{}
	var sort string
	if sortBy == "appName" {
		sort = "a.app_name"
	} else if sortBy == "envName" {
		sort = "environment_name"
	} else {
		// id is to sort by time.
		// id with desc fetches latest scans
		sort = "id"
	}

	query := fmt.Sprintf(" ORDER BY %s ", sort)
	if sortOrder == repoBean.Desc {
		query += " DESC "
	}
	return query, queryParams
}

func (impl ImageScanDeployInfoRepositoryImpl) scanListQueryWithObject(request *repoBean.ImageScanFilter, size int, offset int, deployInfoIds []int) (string, []interface{}) {
	var queryParams []interface{}

	query := ` select info.scan_object_meta_id, a.app_name as object_name, info.object_type, env.environment_name, max(info.id) as id, COUNT(*) OVER() AS total_count 
               from image_scan_deploy_info info
	 		   INNER JOIN app a on a.id=info.scan_object_meta_id `

	if len(request.Severity) > 0 {
		query = query + ` INNER JOIN image_scan_execution_history his on his.id = any (info.image_scan_execution_history_id) 
		 				INNER JOIN image_scan_execution_result res on res.image_scan_execution_history_id=his.id 
		 				INNER JOIN cve_store cs on cs.name= res.cve_store_name `
	}

	query = query + ` INNER JOIN environment env on env.id=info.env_id 
	 				  INNER JOIN cluster c on c.id=env.cluster_id 
					  WHERE info.scan_object_meta_id > 0 and env.active=true and info.image_scan_execution_history_id[1] != -1 
 					  AND a.app_name like ? `

	queryParams = append(queryParams, util.GetLIKEClauseQueryParam(request.AppName))

	if len(deployInfoIds) > 0 {
		query += " AND info.id IN (?) "
		queryParams = append(queryParams, pg.In(deployInfoIds))
	}

	if len(request.Severity) > 0 {
		query += " AND (cs.standard_severity IN (?) OR (cs.severity IN (?) AND cs.standard_severity IS NULL)) "
		queryParams = append(queryParams, pg.In(request.Severity), pg.In(request.Severity))
	}
	if len(request.EnvironmentIds) > 0 {
		query += " AND env.id IN (?) "
		queryParams = append(queryParams, pg.In(request.EnvironmentIds))
	}
	if len(request.ClusterIds) > 0 {
		query += " AND c.id IN (?) "
		queryParams = append(queryParams, pg.In(request.ClusterIds))
	}

	query = query + " GROUP BY info.scan_object_meta_id, a.app_name, info.object_type, env.environment_name "

	queryTemp, queryParamsTemp := getOrderByQueryPart(request.SortBy, request.SortOrder)
	query += queryTemp
	queryParams = append(queryParams, queryParamsTemp...)
	if size > 0 {
		query += " LIMIT ? OFFSET ? "
		queryParams = append(queryParams, size, offset)
	}
	query = query + " ;"
	return query, queryParams
}

func (impl ImageScanDeployInfoRepositoryImpl) scanListingQueryBuilder(request *repoBean.ImageScanFilter, size int, offset int, deployInfoIds []int) (string, []interface{}) {
	query := ""
	var queryParams []interface{}
	if request.AppName == "" && request.CVEName == "" && request.ObjectName == "" {
		query, queryParams = impl.scanListQueryWithoutObject(request, size, offset, deployInfoIds)
	} else if len(request.CVEName) > 0 {
		query, queryParams = impl.scanListQueryWithoutObject(request, size, offset, deployInfoIds)
	} else if len(request.AppName) > 0 {
		query, queryParams = impl.scanListQueryWithObject(request, size, offset, deployInfoIds)
	}
	return query, queryParams
}

// ============================================================================
// Security Overview Methods
// ============================================================================

// GetActiveDeploymentCountByFilters returns the count of unique active deployments (app+env combinations)
// filtered by envIds, clusterIds, and appIds
// Uses cd_workflow_runner as the source of truth for ALL deployments (scanned and unscanned)
func (impl ImageScanDeployInfoRepositoryImpl) GetActiveDeploymentCountByFilters(envIds, clusterIds, appIds []int) (int, error) {
	// Query to find latest deployment per app+environment combination from cd_workflow_runner
	// This is the source of truth for ALL active deployments, not just scanned ones
	// Partitions by (app_id, environment_id) to get the most recent deployment for each app+env
	query := `
		WITH LatestDeployments AS (
			SELECT
				p.app_id,
				p.environment_id,
				ROW_NUMBER() OVER (PARTITION BY p.app_id, p.environment_id ORDER BY cwr.id DESC) AS rn
			FROM cd_workflow_runner cwr
			INNER JOIN cd_workflow cw ON cw.id = cwr.cd_workflow_id
			INNER JOIN pipeline p ON p.id = cw.pipeline_id
			INNER JOIN environment env ON env.id = p.environment_id
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

	// Complete the CTE and count unique app+env combinations
	query += `
		)
		SELECT COUNT(DISTINCT (app_id, environment_id))
		FROM LatestDeployments
		WHERE rn = 1
	`

	var count int
	_, err := impl.dbConnection.Query(&count, query, queryParams...)
	if err != nil {
		impl.logger.Errorw("error in getting active deployment count", "err", err)
		return 0, err
	}

	return count, nil
}

// GetActiveDeploymentCountWithVulnerabilitiesByFilters returns the count of unique active deployments
// that have vulnerabilities in their LATEST deployed artifact
func (impl ImageScanDeployInfoRepositoryImpl) GetActiveDeploymentCountWithVulnerabilitiesByFilters(envIds, clusterIds, appIds []int) (int, error) {
	// Query to find latest deployment per app+environment combination and check if it has vulnerabilities
	// Partitions by (app_id, environment_id) to get the most recent deployment for each app+env
	// This handles cases where pipelines are deleted and recreated for the same app+env
	// Shows vulnerability data from all deployments (successful or failed) since vulnerability is about the image, not deployment status
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

	// Complete the CTE and count deployments with vulnerabilities
	// Join with image_scan_deploy_info to verify scanned deployments
	// Then join with image_scan_execution_history using both id and image for verification
	query += `
		)
		SELECT COUNT(DISTINCT (ld.app_id, ld.environment_id))
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
		WHERE ld.rn = 1
		AND isdi.image_scan_execution_history_id[1] != -1
	`

	var count int
	_, err := impl.dbConnection.Query(&count, query, queryParams...)
	if err != nil {
		impl.logger.Errorw("error in getting deployment count with vulnerabilities", "err", err)
		return 0, err
	}

	return count, nil
}

// GetActiveDeploymentScannedUnscannedCountByFilters returns the count of scanned and unscanned deployments
// in a single query for optimal performance. It finds the latest deployed artifact per app+env combination
// and counts how many have scanned=true vs scanned=false
func (impl ImageScanDeployInfoRepositoryImpl) GetActiveDeploymentScannedUnscannedCountByFilters(envIds, clusterIds, appIds []int) (*DeploymentScannedCount, error) {
	// Query to find latest deployment per app+environment combination and count scanned vs unscanned
	// Uses ROW_NUMBER() to get the latest deployment per app+env
	// Partitions by (app_id, environment_id) to get the most recent deployment for each app+env
	// This handles cases where pipelines are deleted and recreated for the same app+env
	// Shows scan data from all deployments (successful or failed) since scan status is about the image, not deployment status
	// Then uses conditional aggregation to count scanned and unscanned in one query
	query := `
		WITH LatestDeployments AS (
			SELECT
				p.app_id,
				p.environment_id,
				cia.scanned,
				ROW_NUMBER() OVER (PARTITION BY p.app_id, p.environment_id ORDER BY cwr.id DESC) AS rn
			FROM cd_workflow_runner cwr
			INNER JOIN cd_workflow cw ON cw.id = cwr.cd_workflow_id
			INNER JOIN pipeline p ON p.id = cw.pipeline_id
			INNER JOIN ci_artifact cia ON cia.id = cw.ci_artifact_id
			INNER JOIN environment env ON env.id = p.environment_id
			WHERE cwr.workflow_type = 'DEPLOY'
			AND p.deleted = false
			AND env.active = true
	`

	var queryParams []interface{}

	// Add filters
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
		)
		SELECT
			COUNT(*) FILTER (WHERE scanned = true) as scanned_count,
			COUNT(*) FILTER (WHERE scanned = false) as unscanned_count
		FROM LatestDeployments
		WHERE rn = 1
	`

	type queryResult struct {
		ScannedCount   int `pg:"scanned_count"`
		UnscannedCount int `pg:"unscanned_count"`
	}

	var result queryResult
	_, err := impl.dbConnection.Query(&result, query, queryParams...)
	if err != nil {
		impl.logger.Errorw("error in getting deployment scanned/unscanned counts", "err", err)
		return nil, err
	}

	return &DeploymentScannedCount{
		ScannedCount:   result.ScannedCount,
		UnscannedCount: result.UnscannedCount,
	}, nil
}

// GetNonScannedAppEnvCombinations returns app-env combinations that are NOT scanned
// It finds all active deployments and excludes those that exist in image_scan_deploy_info
func (impl ImageScanDeployInfoRepositoryImpl) GetNonScannedAppEnvCombinations(request *repoBean.ImageScanFilter, size int, offset int, deployInfoIds []int) ([]*ImageScanDeployInfo, error) {
	query, queryParams := impl.buildNonScannedAppEnvQuery(request, size, offset, deployInfoIds, false)

	var results []*ImageScanDeployInfo
	_, err := impl.dbConnection.Query(&results, query, queryParams...)
	if err != nil {
		impl.logger.Errorw("error in getting non-scanned app-env combinations", "err", err)
		return nil, err
	}

	return results, nil
}

// GetNonScannedAppEnvCombinationsCount returns count of app-env combinations that are NOT scanned
func (impl ImageScanDeployInfoRepositoryImpl) GetNonScannedAppEnvCombinationsCount(request *repoBean.ImageScanFilter, deployInfoIds []int) (int, error) {
	query, queryParams := impl.buildNonScannedAppEnvQuery(request, 0, 0, deployInfoIds, true)

	var count int
	_, err := impl.dbConnection.Query(&count, query, queryParams...)
	if err != nil {
		impl.logger.Errorw("error in getting non-scanned app-env combinations count", "err", err)
		return 0, err
	}

	return count, nil
}

// buildNonScannedAppEnvQuery builds query to find non-scanned app-env combinations
// It gets all active deployments from cd_workflow_runner and excludes those in image_scan_deploy_info
func (impl ImageScanDeployInfoRepositoryImpl) buildNonScannedAppEnvQuery(request *repoBean.ImageScanFilter, size int, offset int, deployInfoIds []int, isCountQuery bool) (string, []interface{}) {
	var queryParams []interface{}

	// Build the CTE to get latest deployments
	query := `
		WITH LatestDeployments AS (
			SELECT
				p.app_id,
				p.environment_id,
				env.cluster_id,
				ROW_NUMBER() OVER (PARTITION BY p.app_id, p.environment_id ORDER BY cwr.id DESC) AS rn
			FROM cd_workflow_runner cwr
			INNER JOIN cd_workflow cw ON cw.id = cwr.cd_workflow_id
			INNER JOIN pipeline p ON p.id = cw.pipeline_id
			INNER JOIN environment env ON env.id = p.environment_id
			WHERE cwr.workflow_type = 'DEPLOY'
			AND p.deleted = false
			AND env.active = true
	`

	// Add filters to CTE
	if len(request.EnvironmentIds) > 0 {
		query += " AND p.environment_id = ANY(?)"
		queryParams = append(queryParams, pg.Array(request.EnvironmentIds))
	}

	if len(request.ClusterIds) > 0 {
		query += " AND env.cluster_id = ANY(?)"
		queryParams = append(queryParams, pg.Array(request.ClusterIds))
	}

	query += `
		)
	`

	// Main query - select non-scanned app-env combinations
	if isCountQuery {
		query += `
			SELECT COUNT(*)
			FROM LatestDeployments ld
			INNER JOIN app a ON a.id = ld.app_id
			INNER JOIN environment env ON env.id = ld.environment_id
			LEFT JOIN image_scan_deploy_info isdi
				ON isdi.scan_object_meta_id = ld.app_id
				AND isdi.env_id = ld.environment_id
				AND isdi.object_type = 'app'
			WHERE ld.rn = 1
			AND a.active = true
			AND env.active = true
			AND (isdi.id IS NULL OR isdi.image_scan_execution_history_id[1] = -1)
		`
	} else {
		query += `
			SELECT
				ld.app_id as scan_object_meta_id,
				ld.environment_id as env_id,
				ld.cluster_id,
				'app' as object_type,
				-1 as id
			FROM LatestDeployments ld
			INNER JOIN app a ON a.id = ld.app_id
			INNER JOIN environment env ON env.id = ld.environment_id
			LEFT JOIN image_scan_deploy_info isdi
				ON isdi.scan_object_meta_id = ld.app_id
				AND isdi.env_id = ld.environment_id
				AND isdi.object_type = 'app'
			WHERE ld.rn = 1
			AND a.active = true
			AND env.active = true
			AND (isdi.id IS NULL OR isdi.image_scan_execution_history_id[1] = -1)
		`
	}

	// Add app name filter if provided
	if len(request.AppName) > 0 {
		query += " AND a.app_name ILIKE ?"
		queryParams = append(queryParams, util.GetLIKEClauseQueryParam(request.AppName))
	}

	// Add deployInfoIds filter if provided (for RBAC)
	if len(deployInfoIds) > 0 && !isCountQuery {
		// For non-scanned items, we can't filter by deployInfoIds since they don't exist in image_scan_deploy_info
		// This filter is only applicable for scanned items
	}

	// Add pagination for non-count queries
	if !isCountQuery && size > 0 {
		query += " ORDER BY ld.app_id, ld.environment_id LIMIT ? OFFSET ?"
		queryParams = append(queryParams, size, offset)
	}

	return query, queryParams
}
