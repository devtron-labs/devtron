/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package repository

import (
	"github.com/go-pg/pg"
	"time"
)

type WebhookEventData struct {
	tableName   struct{}  `sql:"webhook_event_data" pg:",discard_unknown_columns"`
	Id          int       `sql:"id,pk"`
	GitHostId   int       `sql:"git_host_id,notnull"`
	EventType   string    `sql:"event_type,notnull"`
	PayloadJson string    `sql:"payload_json,notnull"`
	CreatedOn   time.Time `sql:"created_on,notnull"`
}

type WebhookEventDataRepository interface {
	Save(webhookEventData *WebhookEventData) error
	GetById(id int) (*WebhookEventData, error)
}

type WebhookEventDataRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewWebhookEventDataRepositoryImpl(dbConnection *pg.DB) *WebhookEventDataRepositoryImpl {
	return &WebhookEventDataRepositoryImpl{dbConnection: dbConnection}
}

func (impl WebhookEventDataRepositoryImpl) Save(webhookEventData *WebhookEventData) error {
	_, err := impl.dbConnection.Model(webhookEventData).Insert()
	return err
}

func (impl WebhookEventDataRepositoryImpl) GetById(id int) (*WebhookEventData, error) {
	var webhookEventData WebhookEventData
	err := impl.dbConnection.Model(&webhookEventData).
		Where("id = ?", id).Select()
	return &webhookEventData, err
}
