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

package util

import (
	"log"

	"github.com/nats-io/nats.go"
)

const (
	CI_RUNNER_STREAM                  string = "CI-RUNNER"
	ORCHESTRATOR_STREAM               string = "ORCHESTRATOR"
	KUBEWATCH_STREAM                  string = "KUBEWATCH"
	GIT_SENSOR_STREAM                 string = "GIT-SENSOR"
	BULK_APPSTORE_DEPLOY_TOPIC        string = "APP-STORE.BULK-DEPLOY"
	BULK_APPSTORE_DEPLOY_GROUP        string = "APP-STORE-BULK-DEPLOY-GROUP-1"
	BULK_APPSTORE_DEPLOY_DURABLE      string = "APP-STORE-BULK-DEPLOY-DURABLE-1"
	CD_STAGE_COMPLETE_TOPIC           string = "CD-STAGE-COMPLETE"
	CD_COMPLETE_GROUP                 string = "CD-COMPLETE_GROUP-1"
	CD_COMPLETE_DURABLE               string = "CD-COMPLETE_DURABLE-1"
	BULK_DEPLOY_TOPIC                 string = "CD.BULK"
	BULK_HIBERNATE_TOPIC              string = "CD.BULK-HIBERNATE"
	BULK_DEPLOY_GROUP                 string = "CD.BULK.GROUP-1"
	BULK_HIBERNATE_GROUP              string = "CD.BULK-HIBERNATE.GROUP-1"
	BULK_DEPLOY_DURABLE               string = "CD-BULK-DURABLE-1"
	BULK_HIBERNATE_DURABLE            string = "CD-BULK-HIBERNATE-DURABLE-1"
	CI_COMPLETE_TOPIC                 string = "CI-COMPLETE"
	CI_COMPLETE_GROUP                 string = "CI-COMPLETE_GROUP-1"
	CI_COMPLETE_DURABLE               string = "CI-COMPLETE_DURABLE-1"
	APPLICATION_STATUS_UPDATE_TOPIC   string = "APPLICATION_STATUS_UPDATE"
	APPLICATION_STATUS_UPDATE_GROUP   string = "APPLICATION_STATUS_UPDATE_GROUP-1"
	APPLICATION_STATUS_UPDATE_DURABLE string = "APPLICATION_STATUS_UPDATE_DURABLE-1"
	CRON_EVENTS                       string = "CRON_EVENTS"
	CRON_EVENTS_GROUP                 string = "CRON_EVENTS_GROUP-2"
	CRON_EVENTS_DURABLE               string = "CRON_EVENTS_DURABLE-2"
	WORKFLOW_STATUS_UPDATE_TOPIC      string = "WORKFLOW_STATUS_UPDATE"
	WORKFLOW_STATUS_UPDATE_GROUP      string = "WORKFLOW_STATUS_UPDATE_GROUP-1"
	WORKFLOW_STATUS_UPDATE_DURABLE    string = "WORKFLOW_STATUS_UPDATE_DURABLE-1"
	CD_WORKFLOW_STATUS_UPDATE         string = "CD_WORKFLOW_STATUS_UPDATE"
	CD_WORKFLOW_STATUS_UPDATE_GROUP   string = "CD_WORKFLOW_STATUS_UPDATE_GROUP-1"
	CD_WORKFLOW_STATUS_UPDATE_DURABLE string = "CD_WORKFLOW_STATUS_UPDATE_DURABLE-1"
	NEW_CI_MATERIAL_TOPIC             string = "NEW-CI-MATERIAL"
	NEW_CI_MATERIAL_TOPIC_GROUP       string = "NEW-CI-MATERIAL_GROUP-1"
	NEW_CI_MATERIAL_TOPIC_DURABLE     string = "NEW-CI-MATERIAL_DURABLE-1"
	CD_SUCCESS                        string = "CD.TRIGGER"
	WEBHOOK_EVENT_TOPIC               string = "WEBHOOK_EVENT"
)

var ORCHESTRATOR_SUBJECTS = []string{BULK_APPSTORE_DEPLOY_TOPIC, BULK_DEPLOY_TOPIC, BULK_HIBERNATE_TOPIC, CD_SUCCESS, WEBHOOK_EVENT_TOPIC}
var CI_RUNNER_SUBJECTS = []string{CI_COMPLETE_TOPIC, CD_STAGE_COMPLETE_TOPIC}
var KUBEWATCH_SUBJECTS = []string{APPLICATION_STATUS_UPDATE_TOPIC, CRON_EVENTS, WORKFLOW_STATUS_UPDATE_TOPIC, CD_WORKFLOW_STATUS_UPDATE}
var GIT_SENSOR_SUBJECTS = []string{NEW_CI_MATERIAL_TOPIC}

func GetStreamSubjects(streamName string) []string {
	var subjArr []string
	switch streamName {
	case ORCHESTRATOR_STREAM:
		subjArr = ORCHESTRATOR_SUBJECTS
	case CI_RUNNER_STREAM:
		subjArr = CI_RUNNER_SUBJECTS
	case KUBEWATCH_STREAM:
		subjArr = KUBEWATCH_SUBJECTS
	case GIT_SENSOR_STREAM:
		subjArr = GIT_SENSOR_SUBJECTS
	}
	return subjArr
}

func AddStream(js nats.JetStreamContext, streamNames ...string) error {
	var err error
	for _, streamName := range streamNames {
		streamInfo, err := js.StreamInfo(streamName)
		if err == nats.ErrStreamNotFound || streamInfo == nil {
			log.Print("No stream was created already. Need to create one.", "Stream name", streamName)
			//Stream doesn't already exist. Create a new stream from jetStreamContext
			_, err = js.AddStream(&nats.StreamConfig{
				Name:     streamName,
				Subjects: GetStreamSubjects(streamName),
			})
			if err != nil {
				log.Fatal("Error while creating stream", "stream name", streamName, "error", err)
				return err
			}
		} else if err != nil {
			log.Fatal("Error while getting stream info", "stream name", streamName, "error", err)
		}
	}
	return err
}
