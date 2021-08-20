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

package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"go.uber.org/zap"
	"time"
)

type WebhookEventDataConfig interface {
	Save(webhookEventDataRequest *WebhookEventDataRequest) error
}

type WebhookEventDataConfigImpl struct {
	logger                     *zap.SugaredLogger
	webhookEventDataRepository repository.WebhookEventDataRepository
}

func NewWebhookEventDataConfigImpl(logger *zap.SugaredLogger, webhookEventDataRepository repository.WebhookEventDataRepository) *WebhookEventDataConfigImpl {
	return &WebhookEventDataConfigImpl{
		logger:                     logger,
		webhookEventDataRepository: webhookEventDataRepository,
	}
}

type WebhookEventDataRequest struct {
	GitHostId          int       `json:"gitHostId"`
	EventType          string    `json:"eventType"`
	RequestPayloadJson string    `json:"requestPayloadJson"`
	CreatedOn          time.Time `json:"createdOn"`
}

func (impl WebhookEventDataConfigImpl) Save(webhookEventDataRequest *WebhookEventDataRequest) error {
	impl.logger.Debug("save event data request")

	webhookEventDataRequestSql := &repository.WebhookEventData{
		GitHostId:   webhookEventDataRequest.GitHostId,
		EventType:   webhookEventDataRequest.EventType,
		PayloadJson: webhookEventDataRequest.RequestPayloadJson,
		CreatedOn:   time.Now(),
	}

	err := impl.webhookEventDataRepository.Save(webhookEventDataRequestSql)
	if err != nil {
		impl.logger.Errorw("error in saving webhook event data in db", "err", err)
		return err
	}
	return nil
}
