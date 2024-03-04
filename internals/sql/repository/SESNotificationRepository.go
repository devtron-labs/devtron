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
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type SESNotificationRepository interface {
	FindOne(id int) (*SESConfig, error)
	UpdateSESConfig(sesConfig *SESConfig) (*SESConfig, error)
	SaveSESConfig(sesConfig *SESConfig) (*SESConfig, error)
	FindAll() ([]*SESConfig, error)
	FindByIdsIn(ids []int) ([]*SESConfig, error)
	FindByTeamIdOrOwnerId(ownerId int32) ([]*SESConfig, error)
	UpdateSESConfigDefault() (bool, error)
	FindByIds(ids []*int) ([]*SESConfig, error)
	FindDefault() (*SESConfig, error)
	MarkSESConfigDeleted(sesConfig *SESConfig) error
}

type SESNotificationRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewSESNotificationRepositoryImpl(dbConnection *pg.DB) *SESNotificationRepositoryImpl {
	return &SESNotificationRepositoryImpl{dbConnection: dbConnection}
}

type SESConfig struct {
	tableName    struct{} `sql:"ses_config" pg:",discard_unknown_columns"`
	Id           int      `sql:"id,pk"`
	Region       string   `sql:"region"`
	AccessKey    string   `sql:"access_key"`
	SecretKey    string   `sql:"secret_access_key"`
	FromEmail    string   `sql:"from_email"`
	SessionToken string   `sql:"session_token"`
	ConfigName   string   `sql:"config_name"`
	Description  string   `sql:"description"`
	OwnerId      int32    `sql:"owner_id"`
	Default      bool     `sql:"default,notnull"`
	Deleted      bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

func (impl *SESNotificationRepositoryImpl) FindByIdsIn(ids []int) ([]*SESConfig, error) {
	var configs []*SESConfig
	err := impl.dbConnection.Model(&configs).
		Where("id in (?)", pg.In(ids)).
		Where("deleted = ?", false).
		Select()
	return configs, err
}

func (impl *SESNotificationRepositoryImpl) FindOne(id int) (*SESConfig, error) {
	details := &SESConfig{}
	err := impl.dbConnection.Model(details).Where("id = ?", id).
		Where("deleted = ?", false).Select()
	return details, err
}

func (impl *SESNotificationRepositoryImpl) FindAll() ([]*SESConfig, error) {
	var sesConfigs []*SESConfig
	err := impl.dbConnection.Model(&sesConfigs).
		Where("deleted = ?", false).Select()
	return sesConfigs, err
}

func (impl *SESNotificationRepositoryImpl) FindByTeamIdOrOwnerId(ownerId int32) ([]*SESConfig, error) {
	var sesConfigs []*SESConfig
	err := impl.dbConnection.Model(&sesConfigs).Where(`owner_id = ?`, ownerId).
		Where("deleted = ?", false).Select()
	return sesConfigs, err
}

func (impl *SESNotificationRepositoryImpl) UpdateSESConfig(sesConfig *SESConfig) (*SESConfig, error) {
	return sesConfig, impl.dbConnection.Update(sesConfig)
}

func (impl *SESNotificationRepositoryImpl) SaveSESConfig(sesConfig *SESConfig) (*SESConfig, error) {
	return sesConfig, impl.dbConnection.Insert(sesConfig)
}

func (impl *SESNotificationRepositoryImpl) UpdateSESConfigDefault() (bool, error) {
	SESConfigs, err := impl.FindAll()
	for _, SESConfig := range SESConfigs {
		SESConfig.Default = false
		err = impl.dbConnection.Update(SESConfig)
	}
	return true, err
}

func (repo *SESNotificationRepositoryImpl) FindByIds(ids []*int) ([]*SESConfig, error) {
	var objects []*SESConfig
	err := repo.dbConnection.Model(&objects).Where("id in (?)", pg.In(ids)).
		Where("deleted = ?", false).Select()
	return objects, err
}

func (impl *SESNotificationRepositoryImpl) FindDefault() (*SESConfig, error) {
	details := &SESConfig{}
	err := impl.dbConnection.Model(details).Where("ses_config.default = ?", true).
		Where("deleted = ?", false).Select()
	return details, err
}
func (impl *SESNotificationRepositoryImpl) MarkSESConfigDeleted(sesConfig *SESConfig) error {
	sesConfig.Deleted = true
	return impl.dbConnection.Update(sesConfig)
}
