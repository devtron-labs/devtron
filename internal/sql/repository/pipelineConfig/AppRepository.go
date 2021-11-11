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

package pipelineConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/team"
	"github.com/go-pg/pg"
)

type App struct {
	tableName struct{} `sql:"app" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	AppName   string   `sql:"app_name,notnull"` //same as app name
	Active    bool     `sql:"active, notnull"`
	TeamId    int      `sql:"team_id"`
	AppStore  bool     `sql:"app_store, notnull"`
	Team      team.Team
	models.AuditLog
}

type AppRepository interface {
	Save(pipelineGroup *App) error
	SaveWithTxn(pipelineGroup *App, tx *pg.Tx) error
	Update(app *App) error
	UpdateWithTxn(app *App, tx *pg.Tx) error
	FindActiveByName(appName string) (pipelineGroup *App, err error)
	FindActiveListByName(appName string) ([]*App, error)
	FindById(id int) (pipelineGroup *App, err error)
	FindAppsByTeamId(teamId int) ([]App, error)
	FindAppsByTeamIds(teamId []int) ([]App, error)
	FindAppsByTeamName(teamName string) ([]App, error)
	FindAll() ([]App, error)
	FindAppsByEnvironmentId(environmentId int) ([]App, error)
	FindAllActiveAppsWithTeam() ([]*App, error)
	CheckAppExists(appNames []string) ([]*App, error)

	FindByIds(ids []*int) ([]*App, error)
	FetchAppsByFilterV2(appNameIncludes string, appNameExcludes string, environmentId int) ([]*App, error)
	FindAppAndProjectByAppId(appId int) (*App, error)
	GetConnection() *pg.DB
}

type AppRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewAppRepositoryImpl(dbConnection *pg.DB) *AppRepositoryImpl {
	return &AppRepositoryImpl{dbConnection: dbConnection}
}

func (repo AppRepositoryImpl) GetConnection() *pg.DB {
	return repo.dbConnection
}

func (repo AppRepositoryImpl) Save(pipelineGroup *App) error {
	err := repo.dbConnection.Insert(pipelineGroup)
	return err
}

func (repo AppRepositoryImpl) SaveWithTxn(pipelineGroup *App, tx *pg.Tx) error {
	err := tx.Insert(pipelineGroup)
	return err
}

func (repo AppRepositoryImpl) Update(app *App) error {
	_, err := repo.dbConnection.Model(app).WherePK().UpdateNotNull()
	return err
}

func (repo AppRepositoryImpl) UpdateWithTxn(app *App, tx *pg.Tx) error {
	err := tx.Update(app)
	return err
}

func (repo AppRepositoryImpl) FindActiveByName(appName string) (*App, error) {
	pipelineGroup := &App{}
	err := repo.dbConnection.
		Model(pipelineGroup).
		Where("app_name = ?", appName).
		Where("active = ?", true).
		Order("id DESC").Limit(1).
		Select()
	// there is only single active app will be present in db with a same name.
	return pipelineGroup, err
}

func (repo AppRepositoryImpl) FindActiveListByName(appName string) ([]*App, error) {
	var apps []*App
	err := repo.dbConnection.
		Model(&apps).
		Where("app_name = ?", appName).
		Where("active = ?", true).
		Order("id ASC").
		Select()
	// there is only single active app will be present in db with a same name. but check for concurrency
	return apps, err
}

func (repo AppRepositoryImpl) CheckAppExists(appNames []string) ([]*App, error) {
	var apps []*App
	err := repo.dbConnection.
		Model(&apps).
		Where("app_name in (?)", pg.In(appNames)).
		Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindById(id int) (*App, error) {
	pipelineGroup := &App{}
	err := repo.dbConnection.Model(pipelineGroup).Where("id = ?", id).Select()
	return pipelineGroup, err
}

func (repo AppRepositoryImpl) FindAppsByTeamId(teamId int) ([]App, error) {
	var apps []App
	err := repo.dbConnection.Model(&apps).Where("team_id = ?", teamId).Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindAppsByTeamIds(teamId []int) ([]App, error) {
	var apps []App
	err := repo.dbConnection.Model(&apps).Column("app.*", "Team").Where("team_id in (?)", pg.In(teamId)).
		Where("app.active=?", true).Where("app.app_store=?", false).Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindAppsByTeamName(teamName string) ([]App, error) {
	var apps []App
	err := repo.dbConnection.Model(&apps).Column("app.*").
		Join("inner join team t on t.id = app.team_id").Where("t.name = ?", teamName).
		Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindAll() ([]App, error) {
	var apps []App
	err := repo.dbConnection.Model(&apps).Where("active = ?", true).Where("app_store = ?", false).Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindAppsByEnvironmentId(environmentId int) ([]App, error) {
	var apps []App
	err := repo.dbConnection.Model(&apps).ColumnExpr("DISTINCT app.*").
		Join("inner join pipeline p on p.app_id=app.id").Where("p.environment_id = ?", environmentId).Where("p.deleted = ?", false).
		Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindAllActiveAppsWithTeam() ([]*App, error) {
	var apps []*App
	err := repo.dbConnection.Model(&apps).Column("Team").
		Where("app.active = ?", true).Where("app.app_store = ?", false).
		Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindByIds(ids []*int) ([]*App, error) {
	var apps []*App
	err := repo.dbConnection.Model(&apps).Where("active = ?", true).Where("id in (?)", pg.In(ids)).Select()
	return apps, err
}

func (repo AppRepositoryImpl) FetchAppsByFilterV2(appNameIncludes string, appNameExcludes string, environmentId int) ([]*App, error) {
	var apps []*App
	var err error
	if environmentId > 0 && len(appNameExcludes) > 0 {
		err = repo.dbConnection.Model(&apps).ColumnExpr("DISTINCT app.*").
			Join("inner join pipeline p on p.app_id=app.id").
			Where("app.app_name like ?", ""+appNameIncludes+"%").Where("app.app_name not like ?", ""+appNameExcludes+"%").
			Where("app.active=?", true).Where("app_store=?", false).
			Where("p.environment_id = ?", environmentId).Where("p.deleted = ?", false).
			Select()
	} else if environmentId > 0 && appNameExcludes == "" {
		err = repo.dbConnection.Model(&apps).ColumnExpr("DISTINCT app.*").
			Join("inner join pipeline p on p.app_id=app.id").
			Where("app.app_name like ?", ""+appNameIncludes+"%").
			Where("app.active=?", true).Where("app_store=?", false).
			Where("p.environment_id = ?", environmentId).Where("p.deleted = ?", false).
			Select()
	} else if environmentId == 0 && len(appNameExcludes) > 0 {
		err = repo.dbConnection.Model(&apps).ColumnExpr("DISTINCT app.*").
			Where("app.app_name like ?", ""+appNameIncludes+"%").Where("app.app_name not like ?", ""+appNameExcludes+"%").
			Where("app.active=?", true).Where("app_store=?", false).
			Select()
	} else if environmentId == 0 && appNameExcludes == "" {
		err = repo.dbConnection.Model(&apps).ColumnExpr("DISTINCT app.*").
			Where("app.app_name like ?", ""+appNameIncludes+"%").
			Where("app.active=?", true).Where("app_store=?", false).
			Select()
	}
	return apps, err
}

func (repo AppRepositoryImpl) FindAppAndProjectByAppId(appId int) (*App, error) {
	app := &App{}
	err := repo.dbConnection.Model(app).Column("Team").Where("app.id = ?", appId).Select()
	return app, err
}
