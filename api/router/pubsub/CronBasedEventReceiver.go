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
	"github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/pkg/event"
	"encoding/json"
	"github.com/nats-io/stan"
	"go.uber.org/zap"
	"time"
)

type CronBasedEventReceiver interface {
	Subscribe() error
}

type CronBasedEventReceiverImpl struct {
	logger       *zap.SugaredLogger
	pubsubClient *pubsub.PubSubClient
	eventService event.EventService
}

const cronEvents = "CRON_EVENTS"
const cronEventsGroup = "CRON_EVENTS_GROUP-2"
const cronEventsDurable = "CRON_EVENTS_DURABLE-2"

func NewCronBasedEventReceiverImpl(logger *zap.SugaredLogger, pubsubClient *pubsub.PubSubClient, eventService event.EventService) *CronBasedEventReceiverImpl {
	cronBasedEventReceiverImpl := &CronBasedEventReceiverImpl{
		logger:       logger,
		pubsubClient: pubsubClient,
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
	_, err := impl.pubsubClient.Conn.QueueSubscribe(cronEvents, cronEventsGroup, func(msg *stan.Msg) {
		impl.logger.Debug("received cron event")
		defer msg.Ack()
		event := client.Event{}
		err := json.Unmarshal([]byte(string(msg.Data)), &event)
		if err != nil {
			impl.logger.Errorw("err", "err", err)
			return
		}
		err = impl.eventService.HandleEvent(event)
		if err != nil {
			impl.logger.Errorw("err while handle event on subscribe", "err", err)
			return
		}
	}, stan.DurableName(cronEventsDurable), stan.StartWithLastReceived(), stan.AckWait(time.Duration(impl.pubsubClient.AckDuration)*time.Second), stan.SetManualAckMode(), stan.MaxInflight(1))

	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return err
	}
	return nil
}
