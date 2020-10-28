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

package event

import (
	"github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
	"time"
)

type EventServiceImpl struct {
	logger                   *zap.SugaredLogger
	eventRepository          repository.EventRepository
	deploymentFailureHandler app.DeploymentFailureHandler
}

type EventService interface {
	HandleEvent(event client.Event) error
}

func NewEventServiceImpl(logger *zap.SugaredLogger, eventRepository repository.EventRepository, deploymentFailureHandler app.DeploymentFailureHandler) *EventServiceImpl {
	eventServiceImpl := &EventServiceImpl{
		logger:                   logger,
		eventRepository:          eventRepository,
		deploymentFailureHandler: deploymentFailureHandler,
	}
	return eventServiceImpl
}

const cronMinuteWiseEventName string = "minute-event"
const timeLayout = "2006-01-02 15:04:05"

func (impl *EventServiceImpl) HandleEvent(event client.Event) error {
	jobEvent, err := impl.eventRepository.FindLastCompletedEvent(int(util.Success))
	if err != nil && !util2.IsErrNoRows(err) {
		impl.logger.Errorw("err", err)
		return err
	}
	if cronMinuteWiseEventName == event.EventName {
		err := impl.deploymentFailureHandler.Handle(jobEvent, event)
		if err != nil {
			impl.logger.Error("err", err)
			return err
		}
		sourceEventTime, err := time.Parse(timeLayout, event.EventTime)
		if err != nil {
			impl.logger.Error("err", err)
			return err
		}
		completedJob := &repository.JobEvent{
			EventTriggerTime: sourceEventTime,
			EventName:        event.EventName,
			CreatedOn:        time.Now(),
			UpdatedOn:        time.Now(),
		}
		if err == nil {
			completedJob.Status = repository.Success
		} else {
			completedJob.Status = repository.Failure
			completedJob.Message = err.Error()
		}
		err = impl.eventRepository.Save(completedJob)
		if err != nil {
			impl.logger.Errorw("error while save job event", "err", err)
		}
		return err
	}
	return nil
}
