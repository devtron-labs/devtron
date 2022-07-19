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

package chartRepoRepository

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"strings"
)

type Chart struct {
	tableName               struct{}           `sql:"charts" pg:",discard_unknown_columns"`
	Id                      int                `sql:"id,pk"`
	AppId                   int                `sql:"app_id"`
	ChartRepoId             int                `sql:"chart_repo_id"`
	ChartName               string             `sql:"chart_name"` //use composite key as unique id
	ChartVersion            string             `sql:"chart_version"`
	ChartRepo               string             `sql:"chart_repo"`
	ChartRepoUrl            string             `sql:"chart_repo_url"`
	Values                  string             `sql:"values_yaml"`       //json format // used at for release. this should be always updated
	GlobalOverride          string             `sql:"global_override"`   //json format    // global overrides visible to user only
	ReleaseOverride         string             `sql:"release_override"`  //json format   //image descriptor template used for injecting tigger metadata injection
	PipelineOverride        string             `sql:"pipeline_override"` //json format  // pipeline values -> strategy values
	Status                  models.ChartStatus `sql:"status"`            //(new , deployment-in-progress, deployed-To-production, error )
	Active                  bool               `sql:"active"`
	GitRepoUrl              string             `sql:"git_repo_url"`   //git repository where chart is stored
	ChartLocation           string             `sql:"chart_location"` //location within git repo where current chart is pointing
	ReferenceTemplate       string             `sql:"reference_template"`
	ImageDescriptorTemplate string             `sql:"image_descriptor_template"`
	ChartRefId              int                `sql:"chart_ref_id"`
	Latest                  bool               `sql:"latest,notnull"`
	Previous                bool               `sql:"previous,notnull"`
	ReferenceChart          []byte             `sql:"reference_chart"`
	sql.AuditLog
}

type ChartRepository interface {
	//ChartReleasedToProduction(chartRepo, appName, chartVersion string) (bool, error)
	FindOne(chartRepo, appName, chartVersion string) (*Chart, error)
	Save(*Chart) error
	FindCurrentChartVersion(chartRepo, chartName, chartVersionPattern string) (string, error)
	FindActiveChart(appId int) (chart *Chart, err error)
	FindLatestByAppId(appId int) (chart *Chart, err error)
	FindById(id int) (chart *Chart, err error)
	Update(chart *Chart) error

	FindActiveChartsByAppId(appId int) (charts []*Chart, err error)
	FindLatestChartForAppByAppId(appId int) (chart *Chart, err error)
	FindChartByAppIdAndRefId(appId int, chartRefId int) (chart *Chart, err error)
	FindNoLatestChartForAppByAppId(appId int) ([]*Chart, error)
	FindPreviousChartByAppId(appId int) (chart *Chart, err error)
	FindNumberOfAppsWithDeploymentTemplate(appIds []int) (int, error)
	FindChartByGitRepoUrl(gitRepoUrl string) (*Chart, error)
}

func NewChartRepository(dbConnection *pg.DB) *ChartRepositoryImpl {
	return &ChartRepositoryImpl{dbConnection: dbConnection}
}

type ChartRepositoryImpl struct {
	dbConnection *pg.DB
}

func (repositoryImpl ChartRepositoryImpl) FindOne(chartRepo, chartName, chartVersion string) (*Chart, error) {
	chart := &Chart{}
	err := repositoryImpl.dbConnection.
		Model(chart).
		Where("chart_name= ?", chartName).
		Where("chart_version = ?", chartVersion).
		Where("chart_repo = ? ", chartRepo).
		Select()
	return chart, err
}
func (repositoryImpl ChartRepositoryImpl) FindCurrentChartVersion(chartRepo, chartName, chartVersionPattern string) (string, error) {
	chart := &Chart{}
	err := repositoryImpl.dbConnection.
		Model(chart).
		Where("chart_name= ?", chartName).
		Where("chart_version like ?", chartVersionPattern+"%").
		Where("chart_repo = ? ", chartRepo).
		Order("id Desc").
		Limit(1).
		Select()
	return chart.ChartVersion, err
}

//Deprecated
func (repositoryImpl ChartRepositoryImpl) FindActiveChart(appId int) (chart *Chart, err error) {
	chart = &Chart{}
	err = repositoryImpl.dbConnection.
		Model(chart).
		Where("app_id= ?", appId).
		Where("active =?", true).
		Select()
	return chart, err
}

//Deprecated
func (repositoryImpl ChartRepositoryImpl) FindLatestByAppId(appId int) (chart *Chart, err error) {
	chart = &Chart{}
	err = repositoryImpl.dbConnection.
		Model(chart).
		Where("app_id= ?", appId).
		Select()
	return chart, err
}

func (repositoryImpl ChartRepositoryImpl) FindActiveChartsByAppId(appId int) (charts []*Chart, err error) {
	var activeCharts []*Chart
	err = repositoryImpl.dbConnection.
		Model(&activeCharts).
		Where("app_id= ?", appId).
		Where("active= ?", true).
		Select()
	return activeCharts, err
}

func (repositoryImpl ChartRepositoryImpl) FindLatestChartForAppByAppId(appId int) (chart *Chart, err error) {
	chart = &Chart{}
	err = repositoryImpl.dbConnection.
		Model(chart).
		Where("app_id= ?", appId).
		Where("latest= ?", true).
		Select()
	return chart, err
}

func (repositoryImpl ChartRepositoryImpl) FindChartByAppIdAndRefId(appId int, chartRefId int) (chart *Chart, err error) {
	chart = &Chart{}
	err = repositoryImpl.dbConnection.
		Model(chart).
		Where("app_id= ?", appId).
		Where("chart_ref_id= ?", chartRefId).
		Select()
	return chart, err
}

func (repositoryImpl ChartRepositoryImpl) FindNoLatestChartForAppByAppId(appId int) ([]*Chart, error) {
	var charts []*Chart
	err := repositoryImpl.dbConnection.
		Model(&charts).
		Where("app_id= ?", appId).
		Where("latest= ?", false).
		Select()
	return charts, err
}

func (repositoryImpl ChartRepositoryImpl) FindLatestChartForAppByAppIdAndEnvId(appId int, envId int) (chart *Chart, err error) {
	chart = &Chart{}
	err = repositoryImpl.dbConnection.
		Model(chart).
		Where("app_id= ?", appId).
		Where("latest= ?", true).
		Select()
	return chart, err
}

func (repositoryImpl ChartRepositoryImpl) FindPreviousChartByAppId(appId int) (chart *Chart, err error) {
	chart = &Chart{}
	err = repositoryImpl.dbConnection.
		Model(chart).
		Where("app_id= ?", appId).
		Where("previous= ?", true).
		Select()
	return chart, err
}

func (repositoryImpl ChartRepositoryImpl) Save(chart *Chart) error {
	return repositoryImpl.dbConnection.Insert(chart)
}

func (repositoryImpl ChartRepositoryImpl) Update(chart *Chart) error {
	_, err := repositoryImpl.dbConnection.Model(chart).WherePK().UpdateNotNull()
	return err
}

func (repositoryImpl ChartRepositoryImpl) FindById(id int) (chart *Chart, err error) {
	chart = &Chart{}
	err = repositoryImpl.dbConnection.Model(chart).
		Where("id = ?", id).Select()
	return chart, err
}

func (repositoryImpl ChartRepositoryImpl) FindChartByGitRepoUrl(gitRepoUrl string) (*Chart, error) {
	var chart Chart
	err := repositoryImpl.dbConnection.Model(&chart).
		Join("INNER JOIN app ON app.id=app_id").
		Where("app.active = ?", true).
		Where("chart.git_repo_url = ?", gitRepoUrl).
		Where("chart.active = ?", true).
		Limit(1).
		Select()
	return &chart, err
}

func (repositoryImpl ChartRepositoryImpl) FindNumberOfAppsWithDeploymentTemplate(appIds []int) (int, error) {
	var charts []*Chart
	count, err := repositoryImpl.dbConnection.
		Model(&charts).
		ColumnExpr("DISTINCT app_id").
		Where("app_id in (?)", pg.In(appIds)).
		Count()
	if err != nil {
		return 0, err
	}

	return count, nil
}

//---------------------------chart repository------------------

type ChartRepo struct {
	tableName   struct{}            `sql:"chart_repo"`
	Id          int                 `sql:"id,pk"`
	Name        string              `sql:"name"`
	Url         string              `sql:"url"`
	Active      bool                `sql:"active,notnull"`
	Default     bool                `sql:"is_default,notnull"`
	UserName    string              `sql:"user_name"`
	Password    string              `sql:"password"`
	SshKey      string              `sql:"ssh_key"`
	AccessToken string              `sql:"access_token"`
	AuthMode    repository.AuthMode `sql:"auth_mode,notnull"`
	External    bool                `sql:"external,notnull"`
	Deleted     bool                `sql:"deleted,notnull"`
	sql.AuditLog
}

type ChartRepoRepository interface {
	Save(chartRepo *ChartRepo, tx *pg.Tx) error
	Update(chartRepo *ChartRepo, tx *pg.Tx) error
	GetDefault() (*ChartRepo, error)
	FindById(id int) (*ChartRepo, error)
	FindAll() ([]*ChartRepo, error)
	GetConnection() *pg.DB
	MarkChartRepoDeleted(chartRepo *ChartRepo, tx *pg.Tx) error
	FindByName(name string) (*ChartRepo, error)
}
type ChartRepoRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewChartRepoRepositoryImpl(dbConnection *pg.DB) *ChartRepoRepositoryImpl {
	return &ChartRepoRepositoryImpl{
		dbConnection: dbConnection,
	}
}

func (impl ChartRepoRepositoryImpl) GetConnection() *pg.DB {
	return impl.dbConnection
}

func (impl ChartRepoRepositoryImpl) Save(chartRepo *ChartRepo, tx *pg.Tx) error {
	return tx.Insert(chartRepo)
}

func (impl ChartRepoRepositoryImpl) Update(chartRepo *ChartRepo, tx *pg.Tx) error {
	return tx.Update(chartRepo)
}

func (impl ChartRepoRepositoryImpl) GetDefault() (*ChartRepo, error) {
	repo := &ChartRepo{}
	err := impl.dbConnection.Model(repo).
		Where("is_default = ?", true).
		Where("active = ?", true).
		Where("deleted = ?", false).
		Select()
	return repo, err
}

func (impl ChartRepoRepositoryImpl) FindById(id int) (*ChartRepo, error) {
	repo := &ChartRepo{}
	err := impl.dbConnection.Model(repo).
		Where("id = ?", id).
		Where("deleted = ?", false).
		Select()
	return repo, err
}

func (impl ChartRepoRepositoryImpl) FindAll() ([]*ChartRepo, error) {
	var repo []*ChartRepo
	err := impl.dbConnection.Model(&repo).
		Where("deleted = ?", false).
		Select()
	return repo, err
}

func (impl ChartRepoRepositoryImpl) MarkChartRepoDeleted(chartRepo *ChartRepo, tx *pg.Tx) error {
	chartRepo.Deleted = true
	return tx.Update(chartRepo)
}

func (impl ChartRepoRepositoryImpl) FindByName(name string) (*ChartRepo, error) {
	repo := &ChartRepo{}
	err := impl.dbConnection.Model(repo).
		Where("name = ?", name).
		Where("deleted = ?", false).
		Select()
	return repo, err
}

// ------------------------ CHART REF REPOSITORY ---------------
type RefChartDir string
type ChartRef struct {
	tableName        struct{} `sql:"chart_ref" pg:",discard_unknown_columns"`
	Id               int      `sql:"id,pk"`
	Location         string   `sql:"location"`
	Version          string   `sql:"version"`
	Active           bool     `sql:"active,notnull"`
	Default          bool     `sql:"is_default,notnull"`
	Name             string   `sql:"name"`
	ChartData        []byte   `sql:"chart_data"`
	ChartDescription string   `sql:"chart_description"`
	UserUploaded     bool     `sql:"user_uploaded,notnull"`
	sql.AuditLog
}

type ChartRefRepository interface {
	Save(chartRepo *ChartRef) error
	GetDefault() (*ChartRef, error)
	FindById(id int) (*ChartRef, error)
	GetAll() ([]*ChartRef, error)
	CheckIfDataExists(name string, version string) (bool, error)
	FetchChart(name string) ([]*ChartRef, error)
	FetchChartInfoByUploadFlag(userUploaded bool) ([]*ChartRef, error)
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

func (impl ChartRefRepositoryImpl) GetAll() ([]*ChartRef, error) {
	var chartRefs []*ChartRef
	err := impl.dbConnection.Model(&chartRefs).
		Where("active = ?", true).Select()
	return chartRefs, err
}

func (impl ChartRefRepositoryImpl) CheckIfDataExists(name string, version string) (bool, error) {
	repo := &ChartRef{}
	return impl.dbConnection.Model(repo).
		Where("lower(name) = ?", strings.ToLower(name)).
		Where("version = ? ", version).Exists()
}

func (impl ChartRefRepositoryImpl) FetchChart(name string) ([]*ChartRef, error) {
	var chartRefs []*ChartRef
	err := impl.dbConnection.Model(&chartRefs).Where("lower(name) = ?", strings.ToLower(name)).Select()
	if err != nil {
		return nil, err
	}
	return chartRefs, err
}

func (impl ChartRefRepositoryImpl) FetchChartInfoByUploadFlag(userUploaded bool) ([]*ChartRef, error) {
	var repo []*ChartRef
	err := impl.dbConnection.Model(&repo).
		Where("user_uploaded = ?", userUploaded).
		Where("active = ?", true).Select()
	if err != nil {
		return repo, err
	}
	return repo, err
}
