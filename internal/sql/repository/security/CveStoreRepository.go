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
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

type CveStore struct {
	tableName    struct{} `sql:"cve_store" pg:",discard_unknown_columns"`
	Name         string   `sql:"name,pk"`
	Severity     Severity `sql:"severity,notnull"`
	Package      string   `sql:"package,notnull"`
	Version      string   `sql:"version,notnull"`
	FixedVersion string   `sql:"fixed_version,notnull"`
	sql.AuditLog
}

type VulnerabilityRequest struct {
	AppName    string `json:"appName"`
	CveName    string `json:"cveName"`
	EnvIds     []int  `json:"envIds"`
	ClusterIds []int  `json:"clusterIds"`
	Offset     int    `json:"offset"`
	Size       int    `json:"size"`
}

type VulnerabilityExposure struct {
	AppName string `json:"appName"`
	EnvName string `json:"envName"`
	AppId   int    `json:"appId"`
	EnvId   int    `json:"envId"`
	//ClusterId     int    `json:"clusterId"`
	AppStore      bool `json:"appStore"`
	Blocked       bool `json:"blocked"`
	PipelineEnvId int  `json:"-"`
	ChartEnvId    int  `json:"-"`
}

type VulnerabilityExposureListingResponse struct {
	Offset                int                      `json:"offset"`
	Size                  int                      `json:"size"`
	Total                 int                      `json:"total"`
	VulnerabilityExposure []*VulnerabilityExposure `json:"list"`
}

type CveStoreRepository interface {
	Save(model *CveStore) error
	FindAll() ([]*CveStore, error)
	FindByCveNames(names []string) ([]*CveStore, error)
	FindByName(name string) (*CveStore, error)
	Update(model *CveStore) error
	VulnerabilityExposure(request *VulnerabilityRequest) ([]*VulnerabilityExposure, error)
}

type CveStoreRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewCveStoreRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *CveStoreRepositoryImpl {
	return &CveStoreRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl CveStoreRepositoryImpl) Save(model *CveStore) error {
	err := impl.dbConnection.Insert(model)
	return err
}

func (impl CveStoreRepositoryImpl) FindAll() ([]*CveStore, error) {
	var models []*CveStore
	err := impl.dbConnection.Model(&models).Select()
	return models, err
}

func (impl CveStoreRepositoryImpl) FindByCveNames(names []string) ([]*CveStore, error) {
	var models []*CveStore
	err := impl.dbConnection.Model(&models).Where("name in (?)", pg.In(names)).Select()
	return models, err
}

func (impl CveStoreRepositoryImpl) FindByName(name string) (*CveStore, error) {
	var model CveStore
	err := impl.dbConnection.Model(&model).
		Where("name = ?", name).Select()
	return &model, err
}

func (impl CveStoreRepositoryImpl) Update(team *CveStore) error {
	err := impl.dbConnection.Update(team)
	return err
}

func (impl CveStoreRepositoryImpl) VulnerabilityExposure(request *VulnerabilityRequest) ([]*VulnerabilityExposure, error) {
	var items []*VulnerabilityExposure

	query := "SELECT a.id as app_id, a.app_name, a.app_store, p.environment_id as pipeline_env_id, ia.environment_id  as chart_env_id " +
		" FROM app a" +
		" LEFT JOIN pipeline p ON p.app_id=a.id" +
		" LEFT JOIN installed_apps ia ON ia.app_id=a.id" +
		" INNER JOIN environment env ON (env.id=p.environment_id OR env.id=ia.environment_id)"
	query = query + " WHERE (p.deleted=? OR ia.active = ?) and env.active=true"
	if len(request.AppName) > 0 {
		query = query + " AND (a.app_name like '" + request.AppName + "')"
	}
	if len(request.EnvIds) > 0 {
		envIds := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(request.EnvIds)), ","), "[]")
		query = query + " AND (env.id IN (" + envIds + "))"
	}
	if len(request.ClusterIds) > 0 {
		clusterIds := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(request.ClusterIds)), ","), "[]")
		query = query + " AND (env.cluster_id IN (" + clusterIds + "))"
	}
	query = query + " ORDER BY a.id DESC"
	if request.Size > 0 {
		query = query + " LIMIT " + strconv.Itoa(request.Size) + " OFFSET " + strconv.Itoa(request.Offset) + ""
	}
	query = query + " ;"
	impl.logger.Debugw("query", "query:", query)
	_, err := impl.dbConnection.Query(&items, query, false, true)
	if err != nil {
		impl.logger.Error("err", err)
		return []*VulnerabilityExposure{}, err
	}
	return items, err
}
