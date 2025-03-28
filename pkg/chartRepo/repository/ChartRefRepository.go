/*
 * Copyright (c) 2024. Devtron Inc.
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

package chartRepoRepository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strings"
)

type ChartRef struct {
	tableName              struct{} `sql:"chart_ref" pg:",discard_unknown_columns"`
	Id                     int      `sql:"id,pk"`
	Location               string   `sql:"location"`
	Version                string   `sql:"version"`
	Active                 bool     `sql:"active,notnull"`
	Default                bool     `sql:"is_default,notnull"`
	Name                   string   `sql:"name"`
	ChartData              []byte   `sql:"chart_data"`
	ChartDescription       string   `sql:"chart_description"`
	UserUploaded           bool     `sql:"user_uploaded,notnull"`
	IsAppMetricsSupported  bool     `sql:"is_app_metrics_supported,notnull"`
	DeploymentStrategyPath string   `sql:"deployment_strategy_path"`
	JsonPathForStrategy    string   `sql:"json_path_for_strategy"`
	sql.AuditLog
}

type ChartRefExt struct {
	ChartRef
	EmailId string `sql:"email_id"`
}
type ChartRefMetaData struct {
	tableName        struct{} `sql:"chart_ref_metadata" pg:",discard_unknown_columns"`
	ChartName        string   `sql:"chart_name,pk"`
	ChartDescription string   `sql:"chart_description"`
}

type ChartRefRepository interface {
	Save(chartRepo *ChartRef) error
	GetDefault() (*ChartRef, error)
	FindById(id int) (*ChartRef, error)
	FindByIds(ids []int) ([]*ChartRef, error)
	GetAll() ([]*ChartRef, error)
	GetAllChartMetadata() ([]*ChartRefMetaData, error)
	FindByVersionAndName(name, version string) (*ChartRef, error)
	CheckIfDataExists(location string) (bool, error)
	FetchChart(name string) ([]*ChartRef, error)
	FetchInfoOfChartConfiguredInApp(appId int) (*ChartRef, error)
	FetchAllNonUserUploadedChartInfo() ([]*ChartRef, error)
	GetAllChartsWithUserUploadedEmail() ([]*ChartRefExt, error)
}
type ChartRefRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewChartRefRepositoryImpl(dbConnection *pg.DB) *ChartRefRepositoryImpl {
	return &ChartRefRepositoryImpl{
		dbConnection: dbConnection,
	}
}

func (impl ChartRefRepositoryImpl) Save(chartRepo *ChartRef) error {
	return impl.dbConnection.Insert(chartRepo)
}

func (impl ChartRefRepositoryImpl) GetDefault() (*ChartRef, error) {
	repo := &ChartRef{}
	err := impl.dbConnection.Model(repo).
		Where("is_default = ?", true).
		Where("active = ?", true).Select()
	return repo, err
}

func (impl ChartRefRepositoryImpl) FindById(id int) (*ChartRef, error) {
	repo := &ChartRef{}
	err := impl.dbConnection.Model(repo).
		Where("id = ?", id).
		Where("active = ?", true).Select()
	return repo, err
}
func (impl ChartRefRepositoryImpl) FindByIds(ids []int) ([]*ChartRef, error) {
	var chartRefs []*ChartRef
	if len(ids) == 0 {
		return nil, nil
	}
	err := impl.dbConnection.Model(&chartRefs).
		Where("id in (?)", pg.In(ids)).
		Where("active = ?", true).Select()
	return chartRefs, err
}
func (impl ChartRefRepositoryImpl) FindByVersionAndName(name, version string) (*ChartRef, error) {
	repo := &ChartRef{}
	var err error
	if len(name) > 0 {
		err = impl.dbConnection.Model(repo).
			Where("name = ?", name).
			Where("version= ?", version).
			Where("active = ?", true).Select()
	} else {
		err = impl.dbConnection.Model(repo).
			Where("name is NULL", name).
			Where("version= ?", version).
			Where("active = ?", true).Select()
	}
	return repo, err
}

func (impl ChartRefRepositoryImpl) GetAll() ([]*ChartRef, error) {
	var chartRefs []*ChartRef
	err := impl.dbConnection.Model(&chartRefs).
		Where("active = ?", true).Select()
	return chartRefs, err
}

func (impl ChartRefRepositoryImpl) GetAllChartMetadata() ([]*ChartRefMetaData, error) {
	var chartRefMetaDatas []*ChartRefMetaData
	err := impl.dbConnection.Model(&chartRefMetaDatas).Select()
	return chartRefMetaDatas, err
}

func (impl ChartRefRepositoryImpl) CheckIfDataExists(location string) (bool, error) {
	repo := &ChartRef{}
	return impl.dbConnection.Model(repo).
		Where("location = ?", location).
		Exists()
}

func (impl ChartRefRepositoryImpl) FetchChart(name string) ([]*ChartRef, error) {
	var chartRefs []*ChartRef
	err := impl.dbConnection.
		Model(&chartRefs).
		Where("lower(name) = ?", strings.ToLower(name)).
		Select()
	if err != nil {
		return nil, err
	}
	return chartRefs, err
}

func (impl ChartRefRepositoryImpl) FetchAllNonUserUploadedChartInfo() ([]*ChartRef, error) {
	var repo []*ChartRef
	err := impl.dbConnection.Model(&repo).
		Where("user_uploaded = ?", false).
		Select()
	if err != nil {
		return repo, err
	}
	return repo, err
}

func (impl ChartRefRepositoryImpl) FetchInfoOfChartConfiguredInApp(appId int) (*ChartRef, error) {
	var repo ChartRef
	err := impl.dbConnection.Model(&repo).
		Join("inner join charts on charts.chart_ref_id=chart_ref.id").
		Where("charts.app_id= ?", appId).
		Where("charts.latest= ?", true).
		Where("chart_ref.active = ?", true).Select()
	if err != nil {
		return &repo, err
	}
	return &repo, nil
}

func (impl ChartRefRepositoryImpl) GetAllChartsWithUserUploadedEmail() ([]*ChartRefExt, error) {
	var chartRefs []*ChartRefExt
	err := impl.dbConnection.
		Model().
		Table("chart_ref").
		Column("chart_ref.id", "chart_ref.name", "chart_ref.chart_description", "chart_ref.version", "chart_ref.user_uploaded"). // Include user email in the query
		ColumnExpr("users.email_id AS email_id").
		Join("INNER JOIN users"). // Join with users table
		JoinOn("users.id = chart_ref.created_by"). // Join with users table
		Where("chart_ref.active = ?", true). // Filter by active charts
		Select(&chartRefs)
	return chartRefs, err
}

// pipeline strategy metadata repository starts here
type DeploymentStrategy string

const (
	DEPLOYMENT_STRATEGY_BLUE_GREEN DeploymentStrategy = "BLUE-GREEN"
	DEPLOYMENT_STRATEGY_ROLLING    DeploymentStrategy = "ROLLING"
	DEPLOYMENT_STRATEGY_CANARY     DeploymentStrategy = "CANARY"
	DEPLOYMENT_STRATEGY_RECREATE   DeploymentStrategy = "RECREATE"
)

type GlobalStrategyMetadata struct {
	tableName   struct{}           `sql:"global_strategy_metadata" pg:",discard_unknown_columns"`
	Id          int                `sql:"id,pk"`
	Name        DeploymentStrategy `sql:"name"`
	Key         string             `sql:"key"`
	Description string             `sql:"description"`
	Deleted     bool               `sql:"deleted,notnull"`
	sql.AuditLog
}

type GlobalStrategyMetadataRepository interface {
	GetByChartRefId(chartRefId int) ([]*GlobalStrategyMetadata, error)
}
type GlobalStrategyMetadataRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewGlobalStrategyMetadataRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *GlobalStrategyMetadataRepositoryImpl {
	return &GlobalStrategyMetadataRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl *GlobalStrategyMetadataRepositoryImpl) GetByChartRefId(chartRefId int) ([]*GlobalStrategyMetadata, error) {
	var globalStrategies []*GlobalStrategyMetadata
	err := impl.dbConnection.Model(&globalStrategies).
		Join("INNER JOIN global_strategy_metadata_chart_ref_mapping as gsmcrm on gsmcrm.global_strategy_metadata_id=global_strategy_metadata.id").
		Where("gsmcrm.chart_ref_id = ?", chartRefId).
		Where("gsmcrm.active = ?", true).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("error in getting global strategies metadata by chartRefId", "err", err, "chartRefId", chartRefId)
		return nil, err
	}
	return globalStrategies, err
}

// pipeline strategy metadata and chart_ref mapping repository starts here
type GlobalStrategyMetadataChartRefMapping struct {
	tableName                struct{} `sql:"global_strategy_metadata_chart_ref_mapping" pg:",discard_unknown_columns"`
	Id                       int      `sql:"id,pk"`
	GlobalStrategyMetadataId int      `sql:"global_strategy_metadata_id"`
	ChartRefId               int      `sql:"chart_ref_id"`
	Active                   bool     `sql:"active,notnull"`
	Default                  bool     `sql:"default,notnull"`
	GlobalStrategyMetadata   *GlobalStrategyMetadata
	sql.AuditLog
}

type GlobalStrategyMetadataChartRefMappingRepository interface {
	GetByChartRefId(chartRefId int) ([]*GlobalStrategyMetadataChartRefMapping, error)
}
type GlobalStrategyMetadataChartRefMappingRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewGlobalStrategyMetadataChartRefMappingRepositoryImpl(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *GlobalStrategyMetadataChartRefMappingRepositoryImpl {
	return &GlobalStrategyMetadataChartRefMappingRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (impl *GlobalStrategyMetadataChartRefMappingRepositoryImpl) GetByChartRefId(chartRefId int) ([]*GlobalStrategyMetadataChartRefMapping, error) {
	var globalStrategies []*GlobalStrategyMetadataChartRefMapping
	err := impl.dbConnection.Model(&globalStrategies).
		Column("global_strategy_metadata_chart_ref_mapping.*", "GlobalStrategyMetadata").
		Where("global_strategy_metadata_chart_ref_mapping.chart_ref_id = ?", chartRefId).
		Where("global_strategy_metadata_chart_ref_mapping.active = ?", true).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting global strategies metadata mapping by chartRefId", "err", err, "chartRefId", chartRefId)
		return nil, err
	}
	return globalStrategies, err
}
