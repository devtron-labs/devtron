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

package repository

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
)

type SlackNotificationRepository interface {
	FindOne(id int) (*SlackConfig, error)
	UpdateSlackConfig(slackConfig *SlackConfig) (*SlackConfig, error)
	SaveSlackConfig(slackConfig *SlackConfig) (*SlackConfig, error)
	FindAll() ([]SlackConfig, error)
	FindByIdsIn(ids []int) ([]*SlackConfig, error)
	FindByTeamIdOrOwnerId(ownerId int32, teamIds []int) ([]SlackConfig, error)
	FindByName(value string) ([]SlackConfig, error)
	FindByIds(ids []*int) ([]*SlackConfig, error)
}

type SlackNotificationRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewSlackNotificationRepositoryImpl(dbConnection *pg.DB) *SlackNotificationRepositoryImpl {
	return &SlackNotificationRepositoryImpl{dbConnection: dbConnection}
}

type SlackConfig struct {
	tableName   struct{} `sql:"slack_config" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	WebHookUrl  string   `sql:"web_hook_url"`
	ConfigName  string   `sql:"config_name"`
	Description string   `sql:"description"`
	OwnerId     int32    `sql:"owner_id"`
	TeamId      int      `sql:"team_id"`
	models.AuditLog
}

func (impl *SlackNotificationRepositoryImpl) FindByIdsIn(ids []int) ([]*SlackConfig, error) {
	var configs []*SlackConfig
	err := impl.dbConnection.Model(&configs).
		Where("id in (?)", pg.In(ids)).
		Select()
	return configs, err
}

func (impl *SlackNotificationRepositoryImpl) FindOne(id int) (*SlackConfig, error) {
	details := &SlackConfig{}
	err := impl.dbConnection.Model(details).Where("id = ?", id).Select()
	return details, err
}

func (impl *SlackNotificationRepositoryImpl) FindAll() ([]SlackConfig, error) {
	var slackConfigs []SlackConfig
	err := impl.dbConnection.Model(&slackConfigs).Select()
	return slackConfigs, err
}

func (impl *SlackNotificationRepositoryImpl) FindByTeamIdOrOwnerId(ownerId int32, teamIds []int) ([]SlackConfig, error) {
	var slackConfigs []SlackConfig
	if len(teamIds) == 0 {
		err := impl.dbConnection.Model(&slackConfigs).Where(`owner_id = ?`, ownerId).
			Select()
		return slackConfigs, err
	} else {
		err := impl.dbConnection.Model(&slackConfigs).
			Where(`team_id in (?)`, pg.In(teamIds)).
			Select()
		return slackConfigs, err
	}
}

func (impl *SlackNotificationRepositoryImpl) UpdateSlackConfig(slackConfig *SlackConfig) (*SlackConfig, error) {
	return slackConfig, impl.dbConnection.Update(slackConfig)
}

func (impl *SlackNotificationRepositoryImpl) SaveSlackConfig(slackConfig *SlackConfig) (*SlackConfig, error) {
	return slackConfig, impl.dbConnection.Insert(slackConfig)
}

func (impl *SlackNotificationRepositoryImpl) FindByName(value string) ([]SlackConfig, error) {
	var slackConfigs []SlackConfig
	err := impl.dbConnection.Model(&slackConfigs).Where(`config_name like ?`, "%"+value+"%").
		Select()
	return slackConfigs, err

}

func (repo *SlackNotificationRepositoryImpl) FindByIds(ids []*int) ([]*SlackConfig, error) {
	var objects []*SlackConfig
	err := repo.dbConnection.Model(&objects).Where("id in (?)", pg.In(ids)).Select()
	return objects, err
}
