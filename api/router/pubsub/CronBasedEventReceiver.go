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

package pubsub

import (
	"encoding/json"
	pubsub_lib "github.com/devtron-labs/common-lib/pubsub-lib"

	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/pkg/event"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
)

type CronBasedEventReceiver interface {
	Subscribe() error
}

type CronBasedEventReceiverImpl struct {
	logger       *zap.SugaredLogger
	pubSubClient *pubsub_lib.PubSubClientServiceImpl
	eventService event.EventService
}

func NewCronBasedEventReceiverImpl(logger *zap.SugaredLogger, pubSubClient *pubsub_lib.PubSubClientServiceImpl, eventService event.EventService) *CronBasedEventReceiverImpl {
	cronBasedEventReceiverImpl := &CronBasedEventReceiverImpl{
		logger:       logger,
		pubSubClient: pubSubClient,
		eventService: eventService,
	}
	err := cronBasedEventReceiverImpl.Subscribe()
	if err != nil {
		logger.Errorw("err while subscribe", "err", err)
		return nil
	}
	return cronBasedEventReceiverImpl
}

func (impl *CronBasedEventReceiverImpl) Subscribe() error {
	err := impl.pubSubClient.Subscribe(util.CRON_EVENTS, func(msg *pubsub_lib.PubSubMsg) {
		impl.logger.Debug("received cron event")
		event := client.Event{}
		err := json.Unmarshal([]byte(msg.Data), &event)
		if err != nil {
			impl.logger.Errorw("Error while unmarshalling json data", "error", err)
			return
		}
		err = impl.eventService.HandleEvent(event)
		if err != nil {
			impl.logger.Errorw("err while handle event on subscribe", "err", err)
			return
		}
	})

	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return err
	}
	return nil
}
