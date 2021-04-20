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

package team

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
)

type Team struct {
	tableName struct{} `sql:"team"`
	Id        int      `sql:"id,pk"`
	Name      string   `sql:"name,notnull"`
	Active    bool     `sql:"active,notnull"`
	models.AuditLog
}

type TeamRepository interface {
	Save(team *Team) error
	FindAll() ([]Team, error)
	FindOne(id int) (Team, error)
	FindByTeamName(name string) (Team, error)
	Update(team *Team) error

	FindTeamByAppId(appId int) (*Team, error)
	FindTeamByAppName(appName string) (*Team, error)
	FindTeamByAppNameV2() ([]*TeamRbacObjects, error)
	FindByIds(ids []*int) ([]*Team, error)
}
type TeamRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewTeamRepositoryImpl(dbConnection *pg.DB) *TeamRepositoryImpl {
	return &TeamRepositoryImpl{dbConnection: dbConnection}
}

func (impl TeamRepositoryImpl) Save(team *Team) error {
	err := impl.dbConnection.Insert(team)
	return err
}

func (impl TeamRepositoryImpl) FindAll() ([]Team, error) {
	var teams []Team
	err := impl.dbConnection.Model(&teams).Select()
	return teams, err
}

func (impl TeamRepositoryImpl) FindOne(id int) (Team, error) {
	var team Team
	err := impl.dbConnection.Model(&team).
		Where("id = ?", id).Select()
	return team, err
}

func (impl TeamRepositoryImpl) FindByTeamName(name string) (Team, error) {
	var team Team
	err := impl.dbConnection.Model(&team).
		Where("name = ?", name).Select()
	return team, err
}

func (impl TeamRepositoryImpl) Update(team *Team) error {
	err := impl.dbConnection.Update(team)
	return err
}

func (repo TeamRepositoryImpl) FindTeamByAppId(appId int) (*Team, error) {
	team := &Team{}
	err := repo.dbConnection.Model(team).Column("team.*").
		Join("inner join app a on a.team_id = team.id").Where("a.id = ?", appId).
		Select()
	return team, err
}

func (repo TeamRepositoryImpl) FindTeamByAppName(appName string) (*Team, error) {
	team := &Team{}
	err := repo.dbConnection.Model(team).Column("team.*").
		Join("inner join app a on a.team_id = team.id").Where("a.app_name = ?", appName).
		Where("a.active = ?", true).Select()
	return team, err
}

func (repo TeamRepositoryImpl) FindTeamByAppNameV2() ([]*TeamRbacObjects, error) {
	var rbacObjects []*TeamRbacObjects
	err := repo.dbConnection.Model(&rbacObjects).Column("a.app_name as appName, a.id as appId, team.name as teamName").
		Join("inner join app a on a.team_id = team.id").Where("a.active = ?", true).
		Select()
	return rbacObjects, err
}

func (repo TeamRepositoryImpl) FindByIds(ids []*int) ([]*Team, error) {
	var objects []*Team
	err := repo.dbConnection.Model(&objects).Where("active = ?", true).Where("id in (?)", pg.In(ids)).Select()
	return objects, err
}

type TeamRbacObjects struct {
	AppName  string `json:"appName"`
	TeamName string `json:"teamName"`
	AppId    int    `json:"appId"`
}
