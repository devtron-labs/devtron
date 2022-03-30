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

	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/pkg/event"
	"github.com/devtron-labs/devtron/util"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type CronBasedEventReceiver interface {
	Subscribe() error
}

type CronBasedEventReceiverImpl struct {
	logger       *zap.SugaredLogger
	pubsubClient *pubsub.PubSubClient
	eventService event.EventService
}

func NewCronBasedEventReceiverImpl(logger *zap.SugaredLogger, pubsubClient *pubsub.PubSubClient, eventService event.EventService) *CronBasedEventReceiverImpl {
	cronBasedEventReceiverImpl := &CronBasedEventReceiverImpl{
		logger:       logger,
		pubsubClient: pubsubClient,
		eventService: eventService,
	}
	err := util.AddStream(cronBasedEventReceiverImpl.pubsubClient.JetStrCtxt, util.KUBEWATCH_STREAM)
	if err != nil {
		logger.Errorw("err while adding stream", "err", err)
		return nil
	}
	err = cronBasedEventReceiverImpl.Subscribe()
	if err != nil {
		logger.Errorw("err while subscribe", "err", err)
		return nil
	}
	return cronBasedEventReceiverImpl
}

func (impl *CronBasedEventReceiverImpl) Subscribe() error {
	_, err := impl.pubsubClient.JetStrCtxt.QueueSubscribe(util.CRON_EVENTS, util.CRON_EVENTS_GROUP, func(msg *nats.Msg) {
		impl.logger.Debug("received cron event")
		defer msg.Ack()
		event := client.Event{}
		err := json.Unmarshal([]byte(string(msg.Data)), &event)
		if err != nil {
			impl.logger.Errorw("Error while unmarshalling json data", "error", err)
			return
		}
		err = impl.eventService.HandleEvent(event)
		if err != nil {
			impl.logger.Errorw("err while handle event on subscribe", "err", err)
			return
		}
	}, nats.Durable(util.CRON_EVENTS_DURABLE), nats.DeliverLast(), nats.ManualAck(), nats.BindStream(util.KUBEWATCH_STREAM))

	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return err
	}
	return nil
}
