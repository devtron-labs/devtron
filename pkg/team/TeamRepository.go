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
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type Team struct {
	tableName struct{} `sql:"team"`
	Id        int      `sql:"id,pk"`
	Name      string   `sql:"name,notnull"`
	Active    bool     `sql:"active,notnull"`
	sql.AuditLog
}

type TeamRepository interface {
	Save(team *Team) error
	FindAllActive() ([]Team, error)
	FindOne(id int) (Team, error)
	FindByTeamName(name string) (Team, error)
	Update(team *Team) error

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

func (impl TeamRepositoryImpl) FindAllActive() ([]Team, error) {
	var teams []Team
	err := impl.dbConnection.Model(&teams).Where("active = ?", true).Select()
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
