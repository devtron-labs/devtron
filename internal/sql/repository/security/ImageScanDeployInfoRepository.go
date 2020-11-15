/*
 * Copyright (c) 2020 Devtron Labs
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
 *
 */

package security

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

/**
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
	models.AuditLog
}

const (
	ScanObjectType_APP   string = "app"
	ScanObjectType_CHART string = "chart"
	ScanObjectType_POD   string = "pod"
)

type SortBy string
type SortOrder string

const (
	Asc  SortOrder = "ASC"
	Desc SortOrder = "DESC"
)

type ImageScanFilter struct {
	Offset         int    `json:"offset"`
	Size           int    `json:"size"`
	CVEName        string `json:"cveName"`
	AppName        string `json:"appName"`
	ObjectName     string `json:"objectName"`
	EnvironmentIds []int  `json:"envIds"`
	ClusterIds     []int  `json:"clusterIds"`
	Severity       []int  `json:"severity"`
}

type ImageScanListingResponse struct {
	Id               int       `json:"id"`
	ScanObjectMetaId int       `json:"scanObjectMetaId"`
	ObjectName       string    `json:"objectName"`
	ObjectType       string    `json:"objectType"`
	SecurityScan     string    `json:"securityScan"`
	EnvironmentName  string    `json:"environmentName"`
	LastChecked      time.Time `json:"lastChecked"`
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
	ScanListingWithFilter(request *ImageScanFilter, size int, offset int, deployInfoIds []int) ([]*ImageScanListingResponse, error)
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

func (impl ImageScanDeployInfoRepositoryImpl) Update(team *ImageScanDeployInfo) error {
	err := impl.dbConnection.Update(team)
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

func (impl ImageScanDeployInfoRepositoryImpl) ScanListingWithFilter(request *ImageScanFilter, size int, offset int, deployInfoIds []int) ([]*ImageScanListingResponse, error) {
	var models []*ImageScanListingResponse
	query := impl.scanListingQueryBuilder(request, size, offset, deployInfoIds)
	_, err := impl.dbConnection.Query(&models, query, size, offset)
	if err != nil {
		impl.logger.Error("err", err)
		return []*ImageScanListingResponse{}, err
	}
	return models, err
}

func (impl ImageScanDeployInfoRepositoryImpl) scanListQueryWithoutObject(request *ImageScanFilter, size int, offset int, deployInfoIds []int) string {
	query := ""
	query = query + "select info.scan_object_meta_id, info.object_type, env.environment_name, max(info.id) as id"
	query = query + " from image_scan_deploy_info info"
	if len(request.CVEName) > 0 || len(request.Severity) > 0 {
		query = query + " INNER JOIN image_scan_execution_history his on his.id = any (info.image_scan_execution_history_id)"
		query = query + " INNER JOIN image_scan_execution_result res on res.image_scan_execution_history_id=his.id"
		query = query + " INNER JOIN cve_store cs on cs.name= res.cve_store_name"
	}
	query = query + " INNER JOIN environment env on env.id=info.env_id"
	query = query + " INNER JOIN cluster clus on clus.id=env.cluster_id"
	query = query + " WHERE info.scan_object_meta_id > 0"
	if len(deployInfoIds) > 0 {
		ids := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(deployInfoIds)), ","), "[]")
		query = query + " AND info.id IN (" + ids + ")"
	}
	if len(request.CVEName) > 0 {
		query = query + " AND res.cve_store_name like '" + request.CVEName + "'"
	}
	if len(request.Severity) > 0 {
		severities := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(request.Severity)), ","), "[]")
		query = query + " AND cs.severity IN (" + severities + ")"
	}
	if len(request.EnvironmentIds) > 0 {
		envIds := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(request.EnvironmentIds)), ","), "[]")
		query = query + " AND env.id IN (" + envIds + ")"
	}
	if len(request.ClusterIds) > 0 {
		clusterIds := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(request.ClusterIds)), ","), "[]")
		query = query + " AND clus.id IN (" + clusterIds + ")"
	}
	query = query + " group by info.scan_object_meta_id, info.object_type, env.environment_name"
	query = query + " order by id desc"
	if size > 0 {
		query = query + " limit " + strconv.Itoa(size) + " offset " + strconv.Itoa(offset) + ""
	}
	query = query + " ;"
	return query
}

func (impl ImageScanDeployInfoRepositoryImpl) scanListQueryWithObject(request *ImageScanFilter, size int, offset int, deployInfoIds []int) string {
	query := ""
	if len(request.AppName) > 0 {
		query = query + " select info.scan_object_meta_id, a.app_name as object_name, info.object_type, env.environment_name, max(info.id) as id"
		query = query + " from image_scan_deploy_info info"
		query = query + " INNER JOIN app a on a.id=info.scan_object_meta_id"
	} else if len(request.ObjectName) > 0 {
		query = query + " select info.scan_object_meta_id, om.name as object_name,info.object_type, env.environment_name, max(info.id) as id"
		query = query + " from image_scan_deploy_info info"
		query = query + " INNER JOIN image_scan_object_meta om on om.id=info.scan_object_meta_id"
	}
	if len(request.Severity) > 0 {
		query = query + " INNER JOIN image_scan_execution_history his on his.id = any (info.image_scan_execution_history_id)"
		query = query + " INNER JOIN image_scan_execution_result res on res.image_scan_execution_history_id=his.id"
		query = query + " INNER JOIN cve_store cs on cs.name= res.cve_store_name"
	}
	query = query + " INNER JOIN environment env on env.id=info.env_id"
	query = query + " INNER JOIN cluster c on c.id=env.cluster_id"
	query = query + " WHERE info.scan_object_meta_id > 0"
	if len(deployInfoIds) > 0 {
		ids := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(deployInfoIds)), ","), "[]")
		query = query + " AND info.id IN (" + ids + ")"
	}
	if len(request.AppName) > 0 {
		query = query + " AND a.app_name like '%" + request.AppName + "%'"
	} else if len(request.ObjectName) > 0 {
		query = query + " AND om.name like '%" + request.ObjectName + "%'"
	}
	if len(request.Severity) > 0 {
		severities := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(request.Severity)), ","), "[]")
		query = query + " AND cs.severity IN (" + severities + ")"
	}
	if len(request.EnvironmentIds) > 0 {
		envIds := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(request.EnvironmentIds)), ","), "[]")
		query = query + " AND env.id IN (" + envIds + ")"
	}
	if len(request.ClusterIds) > 0 {
		clusterIds := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(request.ClusterIds)), ","), "[]")
		query = query + " AND c.id IN (" + clusterIds + ")"
	}

	if len(request.AppName) > 0 {
		query = query + " group by info.scan_object_meta_id, a.app_name, info.object_type, env.environment_name"
	} else if len(request.ObjectName) > 0 {
		query = query + " group by info.scan_object_meta_id, om.name, info.object_type, env.environment_name"
	}
	query = query + " order by id desc"
	if size > 0 {
		query = query + " limit " + strconv.Itoa(size) + " offset " + strconv.Itoa(offset) + ""
	}
	query = query + " ;"
	return query
}

func (impl ImageScanDeployInfoRepositoryImpl) scanListingQueryBuilder(request *ImageScanFilter, size int, offset int, deployInfoIds []int) string {
	query := ""
	if len(request.AppName) == 0 && len(request.CVEName) == 0 && len(request.ObjectName) == 0 {
		query = impl.scanListQueryWithoutObject(request, size, offset, deployInfoIds)
	} else if len(request.CVEName) > 0 {
		query = impl.scanListQueryWithoutObject(request, size, offset, deployInfoIds)
	} else if len(request.AppName) > 0 || len(request.ObjectName) > 0 {
		query = impl.scanListQueryWithObject(request, size, offset, deployInfoIds)
	}

	return query
}
