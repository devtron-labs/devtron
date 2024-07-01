/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package pubsub_lib

import (
	"encoding/json"
	"github.com/caarlos0/env"
	"github.com/nats-io/nats.go"
	"log"
	"time"
)

const (
	CI_RUNNER_STREAM                              string = "CI-RUNNER"
	ORCHESTRATOR_STREAM                           string = "ORCHESTRATOR"
	KUBEWATCH_STREAM                              string = "KUBEWATCH"
	GIT_SENSOR_STREAM                             string = "GIT-SENSOR"
	IMAGE_SCANNER_STREAM                          string = "IMAGE-SCANNER"
	BULK_APPSTORE_DEPLOY_TOPIC                    string = "APP-STORE.BULK-DEPLOY"
	BULK_APPSTORE_DEPLOY_GROUP                    string = "APP-STORE-BULK-DEPLOY-GROUP-1"
	BULK_APPSTORE_DEPLOY_DURABLE                  string = "APP-STORE-BULK-DEPLOY-DURABLE-1"
	CD_STAGE_COMPLETE_TOPIC                       string = "CD-STAGE-COMPLETE"
	CD_COMPLETE_GROUP                             string = "CD-COMPLETE_GROUP-1"
	CD_COMPLETE_DURABLE                           string = "CD-COMPLETE_DURABLE-1"
	BULK_DEPLOY_TOPIC                             string = "CD.BULK"
	BULK_HIBERNATE_TOPIC                          string = "CD.BULK-HIBERNATE"
	BULK_DEPLOY_GROUP                             string = "CD.BULK.GROUP-1"
	BULK_HIBERNATE_GROUP                          string = "CD.BULK-HIBERNATE.GROUP-1"
	BULK_DEPLOY_DURABLE                           string = "CD-BULK-DURABLE-1"
	BULK_HIBERNATE_DURABLE                        string = "CD-BULK-HIBERNATE-DURABLE-1"
	CI_COMPLETE_TOPIC                             string = "CI-COMPLETE"
	CI_COMPLETE_GROUP                             string = "CI-COMPLETE_GROUP-1"
	CI_COMPLETE_DURABLE                           string = "CI-COMPLETE_DURABLE-1"
	IMAGE_SCANNING_SUCCESS_TOPIC                  string = "IMAGE-SCANNING-SUCCESS"
	IMAGE_SCANNING_SUCCESS_GROUP                  string = "IMAGE-SCANNING-SUCCESS-GROUP"
	IMAGE_SCANNING_SUCCESS_DURABLE                string = "IMAGE-SCANNING-SUCCESS-DURABLE"
	APPLICATION_STATUS_UPDATE_TOPIC               string = "APPLICATION_STATUS_UPDATE"
	APPLICATION_STATUS_UPDATE_GROUP               string = "APPLICATION_STATUS_UPDATE_GROUP-1"
	APPLICATION_STATUS_UPDATE_DURABLE             string = "APPLICATION_STATUS_UPDATE_DURABLE-1"
	APPLICATION_STATUS_DELETE_TOPIC               string = "APPLICATION_STATUS_DELETE"
	APPLICATION_STATUS_DELETE_GROUP               string = "APPLICATION_STATUS_DELETE_GROUP-1"
	APPLICATION_STATUS_DELETE_DURABLE             string = "APPLICATION_STATUS_DELETE_DURABLE-1"
	CRON_EVENTS                                   string = "CRON_EVENTS"
	CRON_EVENTS_GROUP                             string = "CRON_EVENTS_GROUP-2"
	CRON_EVENTS_DURABLE                           string = "CRON_EVENTS_DURABLE-2"
	WORKFLOW_STATUS_UPDATE_TOPIC                  string = "WORKFLOW_STATUS_UPDATE"
	WORKFLOW_STATUS_UPDATE_GROUP                  string = "WORKFLOW_STATUS_UPDATE_GROUP-1"
	WORKFLOW_STATUS_UPDATE_DURABLE                string = "WORKFLOW_STATUS_UPDATE_DURABLE-1"
	CD_WORKFLOW_STATUS_UPDATE                     string = "CD_WORKFLOW_STATUS_UPDATE"
	CD_WORKFLOW_STATUS_UPDATE_GROUP               string = "CD_WORKFLOW_STATUS_UPDATE_GROUP-1"
	CD_WORKFLOW_STATUS_UPDATE_DURABLE             string = "CD_WORKFLOW_STATUS_UPDATE_DURABLE-1"
	NEW_CI_MATERIAL_TOPIC                         string = "NEW-CI-MATERIAL"
	NEW_CI_MATERIAL_TOPIC_GROUP                   string = "NEW-CI-MATERIAL_GROUP-1"
	NEW_CI_MATERIAL_TOPIC_DURABLE                 string = "NEW-CI-MATERIAL_DURABLE-1"
	CD_SUCCESS                                    string = "CD.TRIGGER"
	CD_TRIGGER_GROUP                              string = "CD.TRIGGER_GRP1"
	CD_TRIGGER_DURABLE                            string = "CD-TRIGGER-DURABLE1"
	WEBHOOK_EVENT_TOPIC                           string = "WEBHOOK_EVENT"
	WEBHOOK_EVENT_GROUP                           string = "WEBHOOK_EVENT_GRP"
	WEBHOOK_EVENT_DURABLE                         string = "WEBHOOK_EVENT_DURABLE"
	DEVTRON_TEST_TOPIC                            string = "Test_Topic"
	DEVTRON_TEST_STREAM                           string = "Devtron_Test_Stream"
	DEVTRON_TEST_QUEUE                            string = "Test_Topic_Queue"
	DEVTRON_TEST_CONSUMER                         string = "Test_Topic_Consumer"
	TOPIC_CI_SCAN                                 string = "CI-SCAN"
	TOPIC_CI_SCAN_GRP                             string = "CI-SCAN-GRP-1"
	TOPIC_CI_SCAN_DURABLE                         string = "CI-SCAN-DURABLE-1"
	ARGO_PIPELINE_STATUS_UPDATE_TOPIC             string = "ARGO_PIPELINE_STATUS_UPDATE"
	ARGO_PIPELINE_STATUS_UPDATE_GROUP             string = "ARGO_PIPELINE_STATUS_UPDATE_GROUP-1"
	ARGO_PIPELINE_STATUS_UPDATE_DURABLE           string = "ARGO_PIPELINE_STATUS_UPDATE_DURABLE-1"
	CD_BULK_DEPLOY_TRIGGER_TOPIC                  string = "CD-BULK-DEPLOY-TRIGGER"
	CD_BULK_DEPLOY_TRIGGER_GROUP                  string = "CD-BULK-DEPLOY-TRIGGER-GROUP-1"
	CD_BULK_DEPLOY_TRIGGER_DURABLE                string = "CD-BULK-DEPLOY-TRIGGER-DURABLE-1"
	HELM_CHART_INSTALL_STATUS_TOPIC               string = "HELM-CHART-INSTALL-STATUS-TOPIC"
	HELM_CHART_INSTALL_STATUS_GROUP               string = "HELM-CHART-INSTALL-STATUS-GROUP"
	HELM_CHART_INSTALL_STATUS_DURABLE             string = "HELM-CHART-INSTALL-STATUS-DURABLE"
	DEVTRON_CHART_INSTALL_TOPIC                   string = "DEVTRON-CHART-INSTALL-TOPIC"
	DEVTRON_CHART_INSTALL_GROUP                   string = "DEVTRON-CHART-INSTALL-GROUP"
	DEVTRON_CHART_INSTALL_DURABLE                 string = "DEVTRON-CHART-INSTALL-DURABLE"
	DEVTRON_CHART_PRIORITY_INSTALL_TOPIC          string = "DEVTRON-CHART-PRIORITY-INSTALL-TOPIC"
	DEVTRON_CHART_PRIORITY_INSTALL_GROUP          string = "DEVTRON-CHART-PRIORITY-INSTALL-GROUP"
	DEVTRON_CHART_PRIORITY_INSTALL_DURABLE        string = "DEVTRON-CHART-PRIORITY-INSTALL-DURABLE"
	DEVTRON_CHART_GITOPS_INSTALL_TOPIC            string = "DEVTRON-CHART-GITOPS-INSTALL-TOPIC"
	DEVTRON_CHART_GITOPS_INSTALL_GROUP            string = "DEVTRON-CHART-GITOPS-INSTALL-GROUP"
	DEVTRON_CHART_GITOPS_INSTALL_DURABLE          string = "DEVTRON-CHART-GITOPS-INSTALL-DURABLE"
	DEVTRON_CHART_GITOPS_PRIORITY_INSTALL_TOPIC   string = "DEVTRON-CHART-PRIORITY-GITOPS-INSTALL-TOPIC"
	DEVTRON_CHART_GITOPS_PRIORITY_INSTALL_GROUP   string = "DEVTRON-CHART-PRIORITY-GITOPS-INSTALL-GROUP"
	DEVTRON_CHART_GITOPS_PRIORITY_INSTALL_DURABLE string = "DEVTRON-CHART-PRIORITY-GITOPS-INSTALL-DURABLE"
	PANIC_ON_PROCESSING_TOPIC                     string = "PANIC-ON-PROCESSING-TOPIC"
	PANIC_ON_PROCESSING_GROUP                     string = "PANIC-ON-PROCESSING-GROUP"
	PANIC_ON_PROCESSING_DURABLE                   string = "PANIC-ON-PROCESSING-DURABLE"
	CD_STAGE_SUCCESS_EVENT_TOPIC                  string = "CD-STAGE-SUCCESS-EVENT"
	CD_STAGE_SUCCESS_EVENT_GROUP                  string = "CD-STAGE-SUCCESS-EVENT-GROUP"
	CD_STAGE_SUCCESS_EVENT_DURABLE                string = "CD-STAGE-SUCCESS-EVENT-DURABLE"
	CD_PIPELINE_DELETE_EVENT_TOPIC                string = "CD-PIPELINE-DELETE-EVENT"
	CD_PIPELINE_DELETE_EVENT_GROUP                string = "CD-PIPELINE-DELETE-EVENT-GROUP"
	CD_PIPELINE_DELETE_EVENT_DURABLE              string = "CD-PIPELINE-DELETE-EVENT-DURABLE"
	CHART_SCAN_TOPIC                              string = "CHART-SCAN-TOPIC"
	CHART_SCAN_GROUP                              string = "CHART-SCAN-GROUP"
	CHART_SCAN_DURABLE                            string = "CHART-SCAN-DURABLE"
	NOTIFICATION_EVENT_TOPIC            		  string = "NOTIFICATION_EVENT_TOPIC"
	NOTIFICATION_EVENT_GROUP            		  string = "NOTIFICATION_EVENT_GROUP"
	NOTIFICATION_EVENT_DURABLE          		  string = "NOTIFICATION_EVENT_DURABLE"
)

type NatsTopic struct {
	topicName    string
	streamName   string
	queueName    string // needed for load balancing
	consumerName string
}
type ConfigJson struct {
	// StreamConfigJson is a json string of map[string]NatsStreamConfig
	StreamConfigJson string `env:"STREAM_CONFIG_JSON"`
	// ConsumerConfigJson is a json string of map[string]NatsConsumerConfig
	// eg: "{\"ARGO_PIPELINE_STATUS_UPDATE_DURABLE-1\" : \"{\"natsMsgProcessingBatchSize\" : 3, \"natsMsgBufferSize\" : 3, \"ackWaitInSecs\": 300}\"}"
	ConsumerConfigJson string `env:"CONSUMER_CONFIG_JSON"`
}

var natsTopicMapping = map[string]NatsTopic{

	BULK_APPSTORE_DEPLOY_TOPIC:   {topicName: BULK_APPSTORE_DEPLOY_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: BULK_APPSTORE_DEPLOY_GROUP, consumerName: BULK_APPSTORE_DEPLOY_DURABLE},
	BULK_DEPLOY_TOPIC:            {topicName: BULK_DEPLOY_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: BULK_DEPLOY_GROUP, consumerName: BULK_DEPLOY_DURABLE},
	BULK_HIBERNATE_TOPIC:         {topicName: BULK_HIBERNATE_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: BULK_HIBERNATE_GROUP, consumerName: BULK_HIBERNATE_DURABLE},
	CD_SUCCESS:                   {topicName: CD_SUCCESS, streamName: ORCHESTRATOR_STREAM, queueName: CD_TRIGGER_GROUP, consumerName: CD_TRIGGER_DURABLE},
	WEBHOOK_EVENT_TOPIC:          {topicName: WEBHOOK_EVENT_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: WEBHOOK_EVENT_GROUP, consumerName: WEBHOOK_EVENT_DURABLE},
	CD_BULK_DEPLOY_TRIGGER_TOPIC: {topicName: CD_BULK_DEPLOY_TRIGGER_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: CD_BULK_DEPLOY_TRIGGER_GROUP, consumerName: CD_BULK_DEPLOY_TRIGGER_DURABLE},

	CI_COMPLETE_TOPIC:            {topicName: CI_COMPLETE_TOPIC, streamName: CI_RUNNER_STREAM, queueName: CI_COMPLETE_GROUP, consumerName: CI_COMPLETE_DURABLE},
	CD_STAGE_COMPLETE_TOPIC:      {topicName: CD_STAGE_COMPLETE_TOPIC, streamName: CI_RUNNER_STREAM, queueName: CD_COMPLETE_GROUP, consumerName: CD_COMPLETE_DURABLE},
	IMAGE_SCANNING_SUCCESS_TOPIC: {topicName: IMAGE_SCANNING_SUCCESS_TOPIC, streamName: CI_RUNNER_STREAM, queueName: IMAGE_SCANNING_SUCCESS_GROUP, consumerName: IMAGE_SCANNING_SUCCESS_DURABLE},

	APPLICATION_STATUS_UPDATE_TOPIC: {topicName: APPLICATION_STATUS_UPDATE_TOPIC, streamName: KUBEWATCH_STREAM, queueName: APPLICATION_STATUS_UPDATE_GROUP, consumerName: APPLICATION_STATUS_UPDATE_DURABLE},
	APPLICATION_STATUS_DELETE_TOPIC: {topicName: APPLICATION_STATUS_DELETE_TOPIC, streamName: KUBEWATCH_STREAM, queueName: APPLICATION_STATUS_DELETE_GROUP, consumerName: APPLICATION_STATUS_DELETE_DURABLE},
	CRON_EVENTS:                     {topicName: CRON_EVENTS, streamName: KUBEWATCH_STREAM, queueName: CRON_EVENTS_GROUP, consumerName: CRON_EVENTS_DURABLE},
	WORKFLOW_STATUS_UPDATE_TOPIC:    {topicName: WORKFLOW_STATUS_UPDATE_TOPIC, streamName: KUBEWATCH_STREAM, queueName: WORKFLOW_STATUS_UPDATE_GROUP, consumerName: WORKFLOW_STATUS_UPDATE_DURABLE},
	CD_WORKFLOW_STATUS_UPDATE:       {topicName: CD_WORKFLOW_STATUS_UPDATE, streamName: KUBEWATCH_STREAM, queueName: CD_WORKFLOW_STATUS_UPDATE_GROUP, consumerName: CD_WORKFLOW_STATUS_UPDATE_DURABLE},

	NEW_CI_MATERIAL_TOPIC: {topicName: NEW_CI_MATERIAL_TOPIC, streamName: GIT_SENSOR_STREAM, queueName: NEW_CI_MATERIAL_TOPIC_GROUP, consumerName: NEW_CI_MATERIAL_TOPIC_DURABLE},

	DEVTRON_TEST_TOPIC:                {topicName: DEVTRON_TEST_TOPIC, streamName: DEVTRON_TEST_STREAM, queueName: DEVTRON_TEST_QUEUE, consumerName: DEVTRON_TEST_CONSUMER},
	TOPIC_CI_SCAN:                     {topicName: TOPIC_CI_SCAN, streamName: IMAGE_SCANNER_STREAM, queueName: TOPIC_CI_SCAN_GRP, consumerName: TOPIC_CI_SCAN_DURABLE},
	ARGO_PIPELINE_STATUS_UPDATE_TOPIC: {topicName: ARGO_PIPELINE_STATUS_UPDATE_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: ARGO_PIPELINE_STATUS_UPDATE_GROUP, consumerName: ARGO_PIPELINE_STATUS_UPDATE_DURABLE},

	HELM_CHART_INSTALL_STATUS_TOPIC:             {topicName: HELM_CHART_INSTALL_STATUS_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: HELM_CHART_INSTALL_STATUS_GROUP, consumerName: HELM_CHART_INSTALL_STATUS_DURABLE},
	DEVTRON_CHART_INSTALL_TOPIC:                 {topicName: DEVTRON_CHART_INSTALL_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: DEVTRON_CHART_INSTALL_GROUP, consumerName: DEVTRON_CHART_INSTALL_DURABLE},
	DEVTRON_CHART_PRIORITY_INSTALL_TOPIC:        {topicName: DEVTRON_CHART_PRIORITY_INSTALL_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: DEVTRON_CHART_PRIORITY_INSTALL_GROUP, consumerName: DEVTRON_CHART_PRIORITY_INSTALL_DURABLE},
	DEVTRON_CHART_GITOPS_INSTALL_TOPIC:          {topicName: DEVTRON_CHART_GITOPS_INSTALL_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: DEVTRON_CHART_GITOPS_INSTALL_GROUP, consumerName: DEVTRON_CHART_GITOPS_INSTALL_DURABLE},
	DEVTRON_CHART_GITOPS_PRIORITY_INSTALL_TOPIC: {topicName: DEVTRON_CHART_GITOPS_PRIORITY_INSTALL_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: DEVTRON_CHART_GITOPS_PRIORITY_INSTALL_GROUP, consumerName: DEVTRON_CHART_GITOPS_PRIORITY_INSTALL_DURABLE},

	PANIC_ON_PROCESSING_TOPIC:      {topicName: PANIC_ON_PROCESSING_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: PANIC_ON_PROCESSING_GROUP, consumerName: PANIC_ON_PROCESSING_DURABLE},
	CD_STAGE_SUCCESS_EVENT_TOPIC:   {topicName: CD_STAGE_SUCCESS_EVENT_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: CD_STAGE_SUCCESS_EVENT_GROUP, consumerName: CD_STAGE_SUCCESS_EVENT_DURABLE},
	CD_PIPELINE_DELETE_EVENT_TOPIC: {topicName: CD_PIPELINE_DELETE_EVENT_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: CD_PIPELINE_DELETE_EVENT_GROUP, consumerName: CD_PIPELINE_DELETE_EVENT_DURABLE},
	NOTIFICATION_EVENT_TOPIC:       {topicName: NOTIFICATION_EVENT_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: NOTIFICATION_EVENT_GROUP, consumerName: NOTIFICATION_EVENT_DURABLE},
	CHART_SCAN_TOPIC:               {topicName: CHART_SCAN_TOPIC, streamName: ORCHESTRATOR_STREAM, queueName: CHART_SCAN_GROUP, consumerName: CHART_SCAN_DURABLE},
}

var NatsStreamWiseConfigMapping = map[string]NatsStreamConfig{
	ORCHESTRATOR_STREAM:  {},
	CI_RUNNER_STREAM:     {},
	KUBEWATCH_STREAM:     {},
	GIT_SENSOR_STREAM:    {},
	IMAGE_SCANNER_STREAM: {},
	DEVTRON_TEST_STREAM:  {},
}

var NatsConsumerWiseConfigMapping = map[string]NatsConsumerConfig{
	ARGO_PIPELINE_STATUS_UPDATE_DURABLE:           {},
	TOPIC_CI_SCAN_DURABLE:                         {},
	NEW_CI_MATERIAL_TOPIC_DURABLE:                 {},
	CD_WORKFLOW_STATUS_UPDATE_DURABLE:             {},
	WORKFLOW_STATUS_UPDATE_DURABLE:                {},
	CRON_EVENTS_DURABLE:                           {},
	APPLICATION_STATUS_UPDATE_DURABLE:             {},
	APPLICATION_STATUS_DELETE_DURABLE:             {},
	CD_COMPLETE_DURABLE:                           {},
	CI_COMPLETE_DURABLE:                           {},
	IMAGE_SCANNING_SUCCESS_DURABLE:                {},
	WEBHOOK_EVENT_DURABLE:                         {},
	CD_TRIGGER_DURABLE:                            {},
	BULK_HIBERNATE_DURABLE:                        {},
	BULK_DEPLOY_DURABLE:                           {},
	BULK_APPSTORE_DEPLOY_DURABLE:                  {},
	CD_BULK_DEPLOY_TRIGGER_DURABLE:                {},
	HELM_CHART_INSTALL_STATUS_DURABLE:             {},
	DEVTRON_CHART_INSTALL_DURABLE:                 {},
	DEVTRON_CHART_PRIORITY_INSTALL_DURABLE:        {},
	DEVTRON_CHART_GITOPS_INSTALL_DURABLE:          {},
	DEVTRON_CHART_GITOPS_PRIORITY_INSTALL_DURABLE: {},
	PANIC_ON_PROCESSING_DURABLE:                   {},
	DEVTRON_TEST_CONSUMER:                         {},
	CD_STAGE_SUCCESS_EVENT_DURABLE:                {},
	CD_PIPELINE_DELETE_EVENT_DURABLE:              {},
	NOTIFICATION_EVENT_DURABLE:          		   {},
}

// getConsumerConfigMap will fetch the consumer wise config from the json string
// this will only fetch consumerConfigs for given consumers in the jsonString
func getConsumerConfigMap(jsonString string) map[string]NatsConsumerConfig {
	resMap := map[string]NatsConsumerConfig{}
	if jsonString == "" {
		return resMap
	}
	object := map[string]NatsConsumerConfig{}
	err := json.Unmarshal([]byte(jsonString), &object)
	if err != nil {
		log.Println("error while unmarshalling in getConsumerConfigMap")
		return resMap
	}

	for key, val := range object {
		resMap[key] = val
	}
	return resMap
}

// getStreamConfigMap will fetch the stream wise config from the json string
// this will only fetch streamConfigs for given streams in the jsonString
func getStreamConfigMap(jsonString string) map[string]NatsStreamConfig {
	resMap := map[string]NatsStreamConfig{}
	if jsonString == "" {
		return resMap
	}
	object := map[string]NatsStreamConfig{}
	err := json.Unmarshal([]byte(jsonString), &object)
	if err != nil {
		log.Println("error while unmarshalling in getStreamConfigMap")
		return resMap
	}

	for key, val := range object {
		resMap[key] = val
	}
	return resMap
}

func ParseAndFillStreamWiseAndConsumerWiseConfigMaps() error {
	configJson := ConfigJson{}
	err := env.Parse(&configJson)
	if err != nil {
		log.Println("error while parsing config from environment params", " err", err)
		return err
	}

	// fetch the consumer configs that were given explicitly in the configJson.ConsumerConfigJson
	consumerConfigMap := getConsumerConfigMap(configJson.ConsumerConfigJson)
	// fetch the stream configs that were given explicitly in the configJson.StreamConfigJson
	streamConfigMap := getStreamConfigMap(configJson.StreamConfigJson)

	// default nats configuration values
	defaultConfig := NatsClientConfig{}
	err = env.Parse(&defaultConfig)
	if err != nil {
		log.Print("error while parsing config from environment params", "err", err)
		return err
	}

	// default stream and consumer config values
	defaultStreamConfigVal := defaultConfig.GetDefaultNatsStreamConfig()
	defaultConsumerConfigVal := defaultConfig.GetDefaultNatsConsumerConfig()

	// initialise all the consumer wise config with default values or user defined values
	updateNatsConsumerConfigMapping(defaultConsumerConfigVal, consumerConfigMap)

	// initialise all the stream wise config with default values or user defined values
	updateNatsStreamConfigMapping(defaultStreamConfigVal, streamConfigMap)
	return nil
}

func updateNatsConsumerConfigMapping(defaultConsumerConfigVal NatsConsumerConfig, consumerConfigMap map[string]NatsConsumerConfig) {
	//iterating through all nats topic mappings (assuming source of truth) to update any consumers if not present in consumer mapping
	for _, natsTopic := range natsTopicMapping {
		if _, ok := NatsConsumerWiseConfigMapping[natsTopic.consumerName]; !ok {
			NatsConsumerWiseConfigMapping[natsTopic.consumerName] = NatsConsumerConfig{}
		}
	}
	//initialise all the consumer wise config with default values or user defined values
	for key, _ := range NatsConsumerWiseConfigMapping {
		consumerConfig := defaultConsumerConfigVal
		if _, ok := consumerConfigMap[key]; ok {
			consumerConfig = consumerConfigMap[key]
		}
		NatsConsumerWiseConfigMapping[key] = consumerConfig
	}
}

func updateNatsStreamConfigMapping(defaultStreamConfigVal NatsStreamConfig, streamConfigMap map[string]NatsStreamConfig) {
	for key, _ := range NatsStreamWiseConfigMapping {
		streamConfig := defaultStreamConfigVal
		if _, ok := streamConfigMap[key]; ok {
			streamConfig = streamConfigMap[key]
		}
		NatsStreamWiseConfigMapping[key] = streamConfig
	}
}

func GetNatsTopic(topicName string) NatsTopic {
	return natsTopicMapping[topicName]
}

func GetStreamSubjects(streamName string) []string {
	var subjArr []string

	for _, natsTopic := range natsTopicMapping {
		if natsTopic.streamName == streamName {
			subjArr = append(subjArr, natsTopic.topicName)
		}
	}
	return subjArr
}

func AddStream(isClustered bool, js nats.JetStreamContext, streamConfig *nats.StreamConfig, streamNames ...string) error {
	var err error
	for _, streamName := range streamNames {
		streamInfo, err := js.StreamInfo(streamName)
		if err == nats.ErrStreamNotFound || streamInfo == nil {
			log.Print("No stream was created already. Need to create one. ", "Stream name: ", streamName)
			// Stream doesn't already exist. Create a new stream from jetStreamContext
			cfgToSet := getNewConfig(streamName, streamConfig)
			if cfgToSet.Replicas > 1 && !isClustered {
				cfgToSet.Replicas = 0
				log.Println("replicas > 1 not possible in non clustered mode")
			}
			_, err = js.AddStream(cfgToSet)
			if err != nil {
				log.Println("Error while creating stream. ", "stream name: ", streamName, "error: ", err)
				return err
			}
		} else if err != nil {
			log.Println("Error while getting stream info. ", "stream name: ", streamName, "error: ", err)
			return err
		} else {
			config := streamInfo.Config
			streamConfig.Name = streamName
			if checkConfigChangeReqd(&config, streamConfig, isClustered) {
				_, err1 := js.UpdateStream(&config)
				if err1 != nil {
					log.Println("error occurred while updating stream config. ", "streamName: ", streamName, "streamConfig: ", config, "error: ", err1)
				}
			}
		}
	}
	return err
}

func checkConfigChangeReqd(existingConfig *nats.StreamConfig, toUpdateConfig *nats.StreamConfig, isClustered bool) bool {
	configChanged := false
	newStreamSubjects := GetStreamSubjects(toUpdateConfig.Name)
	if ((toUpdateConfig.MaxAge != time.Duration(0)) && (toUpdateConfig.MaxAge != existingConfig.MaxAge)) || (len(newStreamSubjects) != len(existingConfig.Subjects)) {
		existingConfig.MaxAge = toUpdateConfig.MaxAge
		existingConfig.Subjects = newStreamSubjects
		configChanged = true
	}
	if replicas := toUpdateConfig.Replicas; replicas > 0 && existingConfig.Replicas != replicas && replicas < 5 {
		if replicas > 1 && !isClustered {
			log.Println("replicas >1 is not possible in non-clustered mode ")
		} else {
			existingConfig.Replicas = replicas
			configChanged = true
		}
	}
	return configChanged
}

func getNewConfig(streamName string, toUpdateConfig *nats.StreamConfig) *nats.StreamConfig {
	cfg := &nats.StreamConfig{
		Name:     streamName,
		Subjects: GetStreamSubjects(streamName),
	}

	if toUpdateConfig.MaxAge != time.Duration(0) {
		cfg.MaxAge = toUpdateConfig.MaxAge
	}
	if toUpdateConfig.Replicas > 0 {
		cfg.Replicas = toUpdateConfig.Replicas
	}
	if toUpdateConfig.Retention != nats.RetentionPolicy(0) {
		cfg.Retention = toUpdateConfig.Retention
	}
	return cfg
}
