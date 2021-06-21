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

package chartConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/go-pg/pg"
)
type App struct{
	tableName struct{}  		`sql:"app" pg:",discard_unknown_columns"`
	Id int `sql:"id"`
	AppName string `sql:"app_name"`
	Active bool `sql:"active"`
	CreatedOn string `sql:"created_on"`
	CreatedBy string `sql:"created_by"`
	UpdatedOn string `sql:"updated_on"`
	UpdatedBy int `sql:"updated_by"`
	TeamId int `sql:"team_id"`
	AppStore bool `sql:"app_store"`
}
type ChartEnvConfigOverride struct{
	tableName struct{}  		`sql:"chart_env_config_override" pg:",discard_unknown_columns"`
	Id int     					`sql:"id"`
	ChartId int   				`sql:"chart_id"`
	TargetEnvironment int    	`sql:"target_environment"`
	EnvOverrideYaml string 		`sql:"env_override_yaml"`
	Status string				`sql:"status"`
	Reviewed bool				`sql:"reviewed"`
	Active bool					`sql:"active"`
	CreatedBy int				`sql:"created_by"`
	UpdatedBy int				`sql:"updated_by"`
	Namespace string			`sql:"namespace"`
	Latest bool					`sql:"latest"`
	Previous bool				`sql:"previous"`
	IsOverride bool				`sql:"is_override"`
	models.AuditLog
}
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
	models.AuditLog
}

type ChartRepository interface {
	//ChartReleasedToProduction(chartRepo, appName, chartVersion string) (bool, error)
	FindBulkAppByAppNameSubstring(appNameIncludes string,appNameExcludes string) ([]*App, error)
	BulkUpdateChartByAppId(final map[int]string) error
	FindBulkChartByAppId(appId []int)([]*Chart,error)
	FindBulkEnvByChartIdAndEnvId(chartId []int,envId int)([]*ChartEnvConfigOverride,error)
	FindBulkChartByChartId(chartId []int)([]*Chart,error)
	FindBulkAppByAppId(appId []int)([]*App,error)
	BulkUpdateChartByChartIdAndEnvID(final map[int]string) error

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
}

func NewChartRepository(dbConnection *pg.DB) *ChartRepositoryImpl {
	return &ChartRepositoryImpl{dbConnection: dbConnection}
}

type ChartRepositoryImpl struct {
	dbConnection *pg.DB
}


func (repositoryImpl ChartRepositoryImpl) FindBulkAppByAppNameSubstring(appNameIncludes string,appNameExcludes string) ([]*App, error){
	var apps []*App
	err := repositoryImpl.dbConnection.
		Model(&apps).
		Where("app_name like ? and app_name not like ?", "%"+appNameIncludes+"%","%"+appNameExcludes+"%").
		Select()
	return apps,err
}
func (repositoryImpl ChartRepositoryImpl)BulkUpdateChartByAppId(final map[int]string)error{
	chart := &Chart{}
	for appid,patch:=range final {
		_,err := repositoryImpl.dbConnection.
			Model(chart).
			Set("values_yaml = ?",patch).
			Where("app_id = ?",appid).
			Where("id is not null").
			Update()
		if err!=nil{
			return err
		}
	}
	return nil
}
func (repositoryImpl ChartRepositoryImpl) FindBulkChartByAppId(appId []int)([]*Chart,error){
	var charts []*Chart
	var err error
	for _,appid := range appId {
		temp :=  &Chart{}
		exists,_:= repositoryImpl.dbConnection.
			Model(temp).
			Where("app_id = ?", appid).
			Exists()
		if exists {
			err = repositoryImpl.dbConnection.
				Model(temp).
				Where("app_id = ?", appid).
				Select()
			charts = append(charts, temp)
		}
	}
	return charts,err
}
func (repositoryImpl ChartRepositoryImpl) FindBulkEnvByChartIdAndEnvId(chartId []int,envId int)([]*ChartEnvConfigOverride,error){
	var charts []*ChartEnvConfigOverride
	var err error
	for _,chartid := range chartId {
		temp :=  &ChartEnvConfigOverride{}
		exists,_:= repositoryImpl.dbConnection.
			Model(temp).
			Where("chart_id = ?", chartid).
			Where("target_environment = ?",envId).
			Exists()
		if exists {
			err = repositoryImpl.dbConnection.
				Model(temp).
				Where("chart_id = ?", chartid).
				Where("target_environment = ?", envId).
				Select()
			charts = append(charts, temp)
		}
	}
	return charts,err
}
func (repositoryImpl ChartRepositoryImpl) FindBulkChartByChartId(chartId []int)([]*Chart,error){
	var charts []*Chart
	for _,chartid := range chartId {
		temp :=  &Chart{}
		exists,_:= repositoryImpl.dbConnection.
			Model(temp).
			Where("id = ?", chartid).
			Exists()
			if exists {
				err := repositoryImpl.dbConnection.
					Model(temp).
					Where("id = ?", chartid).
					Select()
				if err != nil {
					panic(err)
				}
				charts = append(charts, temp)
			}
	}
	return charts,nil
}
func (repositoryImpl ChartRepositoryImpl) FindBulkAppByAppId(appId []int)([]*App,error){
	var apps []*App
	for _,appid:=range appId {
		app:=&App{}
		exists,_:= repositoryImpl.dbConnection.
			Model(app).
			Where("id = ?", appid).
			Exists()
		if exists {
			err := repositoryImpl.dbConnection.
				Model(app).
				Where("id = ?", appid).
				Select()
			if err != nil {
				panic(err)
			}
			apps = append(apps, app)
		}
	}
	return apps,nil
}
func (repositoryImpl ChartRepositoryImpl)BulkUpdateChartByChartIdAndEnvID(final map[int]string)error{
	chart := &ChartEnvConfigOverride{}
	for chartid,patch:=range final {
		_,err := repositoryImpl.dbConnection.
			Model(chart).
			Set("env_override_yaml = ?",patch).
			Where("chart_id = ?",chartid).
			Update()
		if err!=nil{
			return err
		}
	}
	return nil
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
	models.AuditLog
}

type ChartRepoRepository interface {
	Save(chartRepo *ChartRepo, tx *pg.Tx) error
	Update(chartRepo *ChartRepo, tx *pg.Tx) error
	GetDefault() (*ChartRepo, error)
	FindById(id int) (*ChartRepo, error)
	FindAll() ([]*ChartRepo, error)
	GetConnection() *pg.DB
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
		Where("active = ?", true).Select()
	return repo, err
}

func (impl ChartRepoRepositoryImpl) FindById(id int) (*ChartRepo, error) {
	repo := &ChartRepo{}
	err := impl.dbConnection.Model(repo).
		Where("id = ?", id).Select()
	return repo, err
}

func (impl ChartRepoRepositoryImpl) FindAll() ([]*ChartRepo, error) {
	var repo []*ChartRepo
	err := impl.dbConnection.Model(&repo).Select()
	return repo, err
}

// ------------------------ CHART REF REPOSITORY ---------------

type ChartRef struct {
	tableName struct{} `sql:"chart_ref" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	Location  string   `sql:"location"`
	Version   string   `sql:"version"`
	Active    bool     `sql:"active"`
	Default   bool     `sql:"is_default"`
	models.AuditLog
}

type ChartRefRepository interface {
	Save(chartRepo *ChartRef) error
	GetDefault() (*ChartRef, error)
	FindById(id int) (*ChartRef, error)
	GetAll() ([]*ChartRef, error)
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
