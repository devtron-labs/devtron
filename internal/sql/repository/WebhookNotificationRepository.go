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

package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type WebhookNotificationRepository interface {
	FindOne(id int) (*WebhookConfig, error)
	UpdateWebhookConfig(webhookConfig *WebhookConfig) (*WebhookConfig, error)
	SaveWebhookConfig(webhookConfig *WebhookConfig) (*WebhookConfig, error)
	FindAll() ([]WebhookConfig, error)
	FindByName(value string) ([]WebhookConfig, error)
	FindByIds(ids []*int) ([]*WebhookConfig, error)
	MarkWebhookConfigDeleted(webhookConfig *WebhookConfig) error
}

type WebhookNotificationRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewWebhookNotificationRepositoryImpl(dbConnection *pg.DB) *WebhookNotificationRepositoryImpl {
	return &WebhookNotificationRepositoryImpl{dbConnection: dbConnection}
}

type WebhookConfig struct {
	tableName   struct{}               `sql:"webhook_config" pg:",discard_unknown_columns"`
	Id          int                    `sql:"id,pk"`
	WebHookUrl  string                 `sql:"web_hook_url"`
	ConfigName  string                 `sql:"config_name"`
	Header      map[string]interface{} `sql:"header"`
	Payload     string                 `sql:"payload"`
	Description string                 `sql:"description"`
	OwnerId     int32                  `sql:"owner_id"`
	Active      bool                   `sql:"active"`
	Deleted     bool                   `sql:"deleted,notnull"`
	sql.AuditLog
}

func (impl *WebhookNotificationRepositoryImpl) FindOne(id int) (*WebhookConfig, error) {
	details := &WebhookConfig{}
	err := impl.dbConnection.Model(details).Where("id = ?", id).
		Where("deleted = ?", false).Select()

	return details, err
}

func (impl *WebhookNotificationRepositoryImpl) FindAll() ([]WebhookConfig, error) {
	var webhookConfigs []WebhookConfig
	err := impl.dbConnection.Model(&webhookConfigs).
		Where("deleted = ?", false).Select()
	return webhookConfigs, err
}

func (impl *WebhookNotificationRepositoryImpl) UpdateWebhookConfig(webhookConfig *WebhookConfig) (*WebhookConfig, error) {
	return webhookConfig, impl.dbConnection.Update(webhookConfig)
}

func (impl *WebhookNotificationRepositoryImpl) SaveWebhookConfig(webhookConfig *WebhookConfig) (*WebhookConfig, error) {
	return webhookConfig, impl.dbConnection.Insert(webhookConfig)
}

func (impl *WebhookNotificationRepositoryImpl) FindByName(value string) ([]WebhookConfig, error) {
	var webhookConfigs []WebhookConfig
	err := impl.dbConnection.Model(&webhookConfigs).Where(`config_name like ?`, "%"+value+"%").
		Where("deleted = ?", false).Select()
	return webhookConfigs, err

}

func (repo *WebhookNotificationRepositoryImpl) FindByIds(ids []*int) ([]*WebhookConfig, error) {
	var objects []*WebhookConfig
	err := repo.dbConnection.Model(&objects).Where("id in (?)", pg.In(ids)).
		Where("deleted = ?", false).Select()
	return objects, err
}

func (impl *WebhookNotificationRepositoryImpl) MarkWebhookConfigDeleted(webhookConfig *WebhookConfig) error {
	webhookConfig.Deleted = true
	return impl.dbConnection.Update(webhookConfig)
}
