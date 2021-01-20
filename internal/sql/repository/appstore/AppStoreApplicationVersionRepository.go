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

package appstore

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type AppStoreApplicationVersionRepository interface {
	FindAll() ([]AppStoreWithVersion, error)
	FindWithFilter(filter *AppStoreFilter) ([]AppStoreWithVersion, error)
	FindById(id int) (*AppStoreApplicationVersion, error)
	FindVersionsByAppStoreId(id int) ([]*AppStoreApplicationVersion, error)
	FindChartVersionByAppStoreId(id int) ([]*AppStoreApplicationVersion, error)
	FindByIds(ids []int) ([]*AppStoreApplicationVersion, error)
	GetReadMeById(id int) (*AppStoreApplicationVersion, error)
	FindByAppStoreName(name string) (*AppStoreWithVersion, error)
	SearchAppStoreChartByName(chartName string) ([]*ChartRepoSearch, error)
}

type AppStoreApplicationVersionRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewAppStoreApplicationVersionRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *AppStoreApplicationVersionRepositoryImpl {
	return &AppStoreApplicationVersionRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type AppStoreApplicationVersion struct {
	TableName   struct{}  `sql:"app_store_application_version" pg:",discard_unknown_columns"`
	Id          int       `sql:"id,pk"`
	Version     string    `sql:"version"`
	AppVersion  string    `sql:"app_version"`
	Created     time.Time `sql:"created"`
	Deprecated  bool      `sql:"deprecated"`
	Description string    `sql:"description"`
	Digest      string    `sql:"digest"`
	Icon        string    `sql:"icon"`
	Name        string    `sql:"name"`
	Source      string    `sql:"source"`
	Home        string    `sql:"home"`
	ValuesYaml  string    `sql:"values_yaml"`
	ChartYaml   string    `sql:"chart_yaml"`
	Latest      bool      `sql:"latest"`
	AppStoreId  int       `sql:"app_store_id"`
	models.AuditLog
	RawValues string `sql:"raw_values"`
	Readme    string `sql:"readme"`
	AppStore  *AppStore
}

type AppStoreWithVersion struct {
	Id                           int       `json:"id"`
	AppStoreApplicationVersionId int       `json:"appStoreApplicationVersionId"`
	Name                         string    `json:"name"`
	ChartRepoId                  int       `json:"chart_repo_id"`
	ChartName                    string    `json:"chart_name"`
	Icon                         string    `json:"icon"`
	Active                       bool      `json:"active"`
	ChartGitLocation             string    `json:"chart_git_location"`
	CreatedOn                    time.Time `json:"created_on"`
	UpdatedOn                    time.Time `json:"updated_on"`
	Version                      string    `json:"version"`
	Deprecated                   bool      `json:"deprecated"`
}

type AppStoreFilter struct {
	ChartRepoId  int    `json:"chartRepoId"`
	AppStoreName string `json:"appStoreName"`
	Deprecated   bool   `json:"deprecated"`
	Offset       int    `json:"offset"`
	Size         int    `json:"size"`
}

type ChartRepoSearch struct {
	AppStoreApplicationVersionId int    `json:"appStoreApplicationVersionId"`
	ChartId                      int    `json:"chartId"`
	ChartName                    string `json:"chartName"`
	ChartRepoId                  int    `json:"chartRepoId"`
	ChartRepoName                string `json:"chartRepoName"`
	Version                      string `json:"version"`
	Deprecated                   bool   `json:"deprecated"`
}

func (impl AppStoreApplicationVersionRepositoryImpl) GetReadMeById(id int) (*AppStoreApplicationVersion, error) {
	var appStoreWithVersion AppStoreApplicationVersion
	err := impl.dbConnection.Model(&appStoreWithVersion).Column("readme", "id").
		Where("id= ?", id).Select()
	return &appStoreWithVersion, err
}

func (impl *AppStoreApplicationVersionRepositoryImpl) FindAll() ([]AppStoreWithVersion, error) {
	var appStoreWithVersion []AppStoreWithVersion
	queryTemp := "select asv.version, asv.icon,asv.deprecated ,asv.id as app_store_application_version_id, aps.*, ch.name as chart_name" +
		" from app_store_application_version asv inner join app_store aps on asv.app_store_id = aps.id" +
		" inner join chart_repo ch on aps.chart_repo_id = ch.id" +
		" where asv.latest is TRUE and ch.active = ? order by aps.name asc;"
	_, err := impl.dbConnection.Query(&appStoreWithVersion, queryTemp, true)
	if err != nil {
		return nil, err
	}
	return appStoreWithVersion, err
}

func (impl *AppStoreApplicationVersionRepositoryImpl) FindWithFilter(filter *AppStoreFilter) ([]AppStoreWithVersion, error) {
	var appStoreWithVersion []AppStoreWithVersion
	query := "SELECT asv.version, asv.icon,asv.deprecated ,asv.id as app_store_application_version_id, aps.*, ch.name as chart_name" +
		" FROM app_store_application_version asv" +
		" INNER JOIN app_store aps ON asv.app_store_id = aps.id" +
		" INNER JOIN chart_repo ch ON aps.chart_repo_id = ch.id" +
		" WHERE asv.latest IS TRUE AND ch.active = TRUE"
	if filter.Deprecated {
		query = query + " AND asv.deprecated = TRUE"
	}
	if len(filter.AppStoreName) > 0 {
		query = query + " AND aps.name LIKE '%" + filter.AppStoreName + "%'"
	}
	if filter.ChartRepoId > 0 {
		query = query + " AND ch.id IN (?)"
	}
	query = query + " ORDER BY aps.name ASC"
	if filter.Size > 0 {
		query = query + " OFFSET " + strconv.Itoa(filter.Offset) + " LIMIT " + strconv.Itoa(filter.Size) + ""
	}
	query = query + ";"

	var err error
	if filter.ChartRepoId > 0 {
		_, err = impl.dbConnection.Query(&appStoreWithVersion, query, filter.ChartRepoId)
	} else {
		_, err = impl.dbConnection.Query(&appStoreWithVersion, query)
	}
	if err != nil {
		return nil, err
	}
	return appStoreWithVersion, err
}

func (impl AppStoreApplicationVersionRepositoryImpl) FindById(id int) (*AppStoreApplicationVersion, error) {
	appStoreWithVersion := &AppStoreApplicationVersion{}
	err := impl.dbConnection.
		Model(appStoreWithVersion).
		Column("app_store_application_version.*", "AppStore", "AppStore.ChartRepo").
		Where("app_store_application_version.id = ?", id).
		Limit(1).
		Select()
	return appStoreWithVersion, err
}

func (impl AppStoreApplicationVersionRepositoryImpl) FindByIds(ids []int) ([]*AppStoreApplicationVersion, error) {
	var appStoreApplicationVersions []*AppStoreApplicationVersion
	if len(ids) == 0 {
		return appStoreApplicationVersions, nil
	}
	err := impl.dbConnection.
		Model(&appStoreApplicationVersions).
		Column("app_store_application_version.*", "AppStore", "AppStore.ChartRepo").
		Where("app_store_application_version.id in (?)", pg.In(ids)).
		Select()
	return appStoreApplicationVersions, err
}

func (impl AppStoreApplicationVersionRepositoryImpl) FindChartVersionByAppStoreId(appStoreId int) ([]*AppStoreApplicationVersion, error) {
	var appStoreWithVersion []*AppStoreApplicationVersion
	err := impl.dbConnection.
		Model(&appStoreWithVersion).
		Column("app_store_application_version.version", "app_store_application_version.id").
		Where("app_store_application_version.app_store_id = ?", appStoreId).
		Select()
	return appStoreWithVersion, err
}

func (impl AppStoreApplicationVersionRepositoryImpl) FindVersionsByAppStoreId(id int) ([]*AppStoreApplicationVersion, error) {
	var appStoreApplicationVersions []*AppStoreApplicationVersion
	err := impl.dbConnection.
		Model(&appStoreApplicationVersions).
		Column("app_store_application_version.*", "AppStore", "AppStore.ChartRepo").
		Join("inner join app_store aps on aps.id = app_store_application_version.app_store_id").
		Join("inner join chart_repo as cr on cr.id = aps.chart_repo_id").
		Where("aps.id = ?", id).
		Order("app_store_application_version.created DESC").
		Select()
	return appStoreApplicationVersions, err
}

func (impl *AppStoreApplicationVersionRepositoryImpl) FindByAppStoreName(name string) (*AppStoreWithVersion, error) {
	var appStoreWithVersion AppStoreWithVersion
	queryTemp := "SELECT asv.version, asv.icon,asv.id as app_store_application_version_id, aps.*, ch.name as chart_name FROM app_store_application_version asv INNER JOIN app_store aps ON asv.app_store_id = aps.id INNER JOIN chart_repo ch ON aps.chart_repo_id = ch.id WHERE asv.latest IS TRUE AND aps.name LIKE ?;"
	_, err := impl.dbConnection.Query(&appStoreWithVersion, queryTemp, name)
	if err != nil {
		return nil, err
	}
	return &appStoreWithVersion, err
}

func (impl *AppStoreApplicationVersionRepositoryImpl) SearchAppStoreChartByName(chartName string) ([]*ChartRepoSearch, error) {
	var chartRepos []*ChartRepoSearch
	//eryTemp := "select asv.version, asv.icon,asv.deprecated ,asv.id as app_store_application_version_id, aps.*, ch.name as chart_name from app_store_application_version asv inner join app_store aps on asv.app_store_id = aps.id inner join chart_repo ch on aps.chart_repo_id = ch.id where asv.latest is TRUE order by aps.name asc;"
	queryTemp := "select asv.id as app_store_application_version_id, asv.version, asv.deprecated, aps.id as chart_id," +
		" aps.name as chart_name, chr.id as chart_repo_id, chr.name as chart_repo_name" +
		" from app_store_application_version asv" +
		" inner join app_store aps on asv.app_store_id = aps.id" +
		" inner join chart_repo chr on aps.chart_repo_id = chr.id" +
		" where aps.name like '%" + chartName + "%' and asv.latest is TRUE order by aps.name asc;"
	_, err := impl.dbConnection.Query(&chartRepos, queryTemp)
	if err != nil {
		return nil, err
	}
	return chartRepos, err
}
