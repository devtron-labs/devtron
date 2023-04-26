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
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type ChartRepoFields struct {
	Id                      int                 `sql:"id,pk"`
	Name                    string              `sql:"name"`
	Url                     string              `sql:"url"`
	Active                  bool                `sql:"active,notnull"`
	Default                 bool                `sql:"is_default,notnull"`
	UserName                string              `sql:"user_name"`
	Password                string              `sql:"password"`
	SshKey                  string              `sql:"ssh_key"`
	AccessToken             string              `sql:"access_token"`
	AuthMode                repository.AuthMode `sql:"auth_mode,notnull"`
	External                bool                `sql:"external,notnull"`
	Deleted                 bool                `sql:"deleted,notnull"`
	AllowInsecureConnection bool                `sql:"allow_insecure_connection"`
}
type ChartRepo struct {
	tableName struct{} `sql:"chart_repo"`
	ChartRepoFields
	sql.AuditLog
}
type ChartRepoWithDeploymentCount struct {
	ChartRepoFields
	sql.AuditLog
	ActiveDeploymentCount int `sql:"deployment_count,notnull"`
}

type ChartRepoRepository interface {
	Save(chartRepo *ChartRepo, tx *pg.Tx) error
	Update(chartRepo *ChartRepo, tx *pg.Tx) error
	GetDefault() (*ChartRepo, error)
	FindById(id int) (*ChartRepo, error)
	FindAll() ([]*ChartRepo, error)
	FindAllWithDeploymentCount() ([]*ChartRepoWithDeploymentCount, error)
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
func (impl ChartRepoRepositoryImpl) FindAllWithDeploymentCount() ([]*ChartRepoWithDeploymentCount, error) {
	var repo []*ChartRepoWithDeploymentCount
	query := "select chart_repo.*,count(jq.ia_id ) as deployment_count" +
		" from chart_repo left join" +
		" (select aps.chart_repo_id as cr_id ,ia.id as ia_id from installed_app_versions iav" +
		" inner join installed_apps ia on iav.installed_app_id = ia.id" +
		" inner join app_store_application_version asav on iav.app_store_application_version_id = asav.id" +
		" inner join app_store aps on asav.app_store_id = aps.id" +
		" where ia.active=true and iav.active=true) jq" +
		" on jq.cr_id = chart_repo.id" +
		" where chart_repo.deleted = false Group by chart_repo.id;"
	_, err := impl.dbConnection.Query(&repo, query)
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
