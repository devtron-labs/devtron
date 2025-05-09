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
	repoBean "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util"
	"time"

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
 					  AND a.app_name like '%` + request.AppName + `%' `

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
