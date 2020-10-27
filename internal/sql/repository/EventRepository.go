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
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type EventRepository interface {
	Save(event *JobEvent) error
	Update(event *JobEvent) error
	FindLastCompletedEvent(eventTypeId int) (*JobEvent, error)
}

type EventRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

type JobEvent struct {
	tableName        struct{}  `sql:"job_event" pg:",discard_unknown_columns"`
	Id               int       `sql:"id,pk"`
	EventTriggerTime time.Time `sql:"event_trigger_time"`
	EventName        string    `sql:"name"`
	Status           string    `sql:"status"`
	Message          string    `sql:"message"`
	CreatedOn        time.Time `sql:"created_on"`
	UpdatedOn        time.Time `sql:"updated_on"`
}

const Success = "SUCCESS"
const Failure = "FAILURE"

func NewEventRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *EventRepositoryImpl {
	return &EventRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl *EventRepositoryImpl) FindLastCompletedEvent(eventTypeId int) (*JobEvent, error) {
	jobEvent := &JobEvent{}
	err := impl.dbConnection.
		Model(jobEvent).
		Order("id DESC").
		Limit(1).
		Select()
	return jobEvent, err
}

func (impl *EventRepositoryImpl) Save(event *JobEvent) error {
	return impl.dbConnection.Insert(event)
}

func (impl *EventRepositoryImpl) Update(event *JobEvent) error {
	return impl.dbConnection.Update(event)
}
