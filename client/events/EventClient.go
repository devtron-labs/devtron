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

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/pkg/attributes/bean"
	"github.com/devtron-labs/devtron/pkg/module"
	bean3 "github.com/devtron-labs/devtron/pkg/module/bean"
	"net/http"
	"time"

	"github.com/caarlos0/env"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util "github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
)

type EventClientConfig struct {
	DestinationURL     string             `env:"EVENT_URL" envDefault:"http://localhost:3000/notify" description:"Notifier service url"`
	NotificationMedium NotificationMedium `env:"NOTIFICATION_MEDIUM" envDefault:"rest" description:"notification medium"`
	EnableNotifierV2   bool               `env:"ENABLE_NOTIFIER_V2" envDefault:"false" description:"enable notifier v2"`
}
type NotificationMedium string

const PUB_SUB NotificationMedium = "nats"

func GetEventClientConfig() (*EventClientConfig, error) {
	cfg := &EventClientConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, errors.New("could not get event service url")
	}
	return cfg, err
}

type EventClient interface {
	WriteNotificationEvent(event Event) (bool, error)
	WriteNatsEvent(channel string, payload interface{}) error
}

type Event struct {
	EventTypeId         int               `json:"eventTypeId"`
	EventName           string            `json:"eventName"`
	PipelineId          int               `json:"pipelineId"`
	PipelineType        string            `json:"pipelineType"`
	CorrelationId       string            `json:"correlationId"`
	Payload             *Payload          `json:"payload"`
	EventTime           string            `json:"eventTime"`
	TeamId              int               `json:"teamId"`
	AppId               int               `json:"appId"`
	EnvId               int               `json:"envId"`
	IsProdEnv           bool              `json:"isProdEnv"`
	ClusterId           int               `json:"clusterId"`
	CdWorkflowType      bean.WorkflowType `json:"cdWorkflowType,omitempty"`
	CdWorkflowRunnerId  int               `json:"cdWorkflowRunnerId"`
	CiWorkflowRunnerId  int               `json:"ciWorkflowRunnerId"`
	CiArtifactId        int               `json:"ciArtifactId"`
	EnvIdsForCiPipeline []int             `json:"envIdsForCiPipeline"`
	BaseUrl             string            `json:"baseUrl"`
	UserId              int               `json:"-"`
}

type Payload struct {
	AppName               string               `json:"appName"`
	EnvName               string               `json:"envName"`
	PipelineName          string               `json:"pipelineName"`
	Source                string               `json:"source"`
	DockerImageUrl        string               `json:"dockerImageUrl"`
	TriggeredBy           string               `json:"triggeredBy"`
	Stage                 string               `json:"stage"`
	DeploymentHistoryLink string               `json:"deploymentHistoryLink"`
	AppDetailLink         string               `json:"appDetailLink"`
	DownloadLink          string               `json:"downloadLink"`
	BuildHistoryLink      string               `json:"buildHistoryLink"`
	MaterialTriggerInfo   *MaterialTriggerInfo `json:"material"`
	FailureReason         string               `json:"failureReason"`
}

type CiPipelineMaterialResponse struct {
	Id              int                    `json:"id"`
	GitMaterialId   int                    `json:"gitMaterialId"`
	GitMaterialUrl  string                 `json:"gitMaterialUrl"`
	GitMaterialName string                 `json:"gitMaterialName"`
	Type            string                 `json:"type"`
	Value           string                 `json:"value"`
	Active          bool                   `json:"active"`
	History         []*gitSensor.GitCommit `json:"history,omitempty"`
	LastFetchTime   time.Time              `json:"lastFetchTime"`
	IsRepoError     bool                   `json:"isRepoError"`
	RepoErrorMsg    string                 `json:"repoErrorMsg"`
	IsBranchError   bool                   `json:"isBranchError"`
	BranchErrorMsg  string                 `json:"branchErrorMsg"`
	Url             string                 `json:"url"`
}

type MaterialTriggerInfo struct {
	GitTriggers map[int]pipelineConfig.GitCommit `json:"gitTriggers"`
	CiMaterials []CiPipelineMaterialResponse     `json:"ciMaterials"`
}

type EventRESTClientImpl struct {
	logger                         *zap.SugaredLogger
	client                         *http.Client
	config                         *EventClientConfig
	pubsubClient                   *pubsub.PubSubClientServiceImpl
	ciPipelineRepository           pipelineConfig.CiPipelineRepository
	pipelineRepository             pipelineConfig.PipelineRepository
	attributesRepository           repository.AttributesRepository
	moduleService                  module.ModuleService
	notificationSettingsRepository repository.NotificationSettingsRepository
}

func NewEventRESTClientImpl(logger *zap.SugaredLogger, client *http.Client, config *EventClientConfig, pubsubClient *pubsub.PubSubClientServiceImpl,
	ciPipelineRepository pipelineConfig.CiPipelineRepository, pipelineRepository pipelineConfig.PipelineRepository,
	attributesRepository repository.AttributesRepository, moduleService module.ModuleService,
	notificationSettingsRepository repository.NotificationSettingsRepository) *EventRESTClientImpl {
	return &EventRESTClientImpl{logger: logger, client: client, config: config, pubsubClient: pubsubClient,
		ciPipelineRepository: ciPipelineRepository, pipelineRepository: pipelineRepository,
		attributesRepository: attributesRepository, moduleService: moduleService,
		notificationSettingsRepository: notificationSettingsRepository}
}

func (impl *EventRESTClientImpl) buildFinalPayload(event Event, cdPipeline *pipelineConfig.Pipeline, ciPipeline *pipelineConfig.CiPipeline) *Payload {
	payload := event.Payload
	if payload == nil {
		payload = &Payload{}
	}
	if event.PipelineType == string(util.CD) {
		if cdPipeline != nil {
			payload.AppName = cdPipeline.App.AppName
			payload.EnvName = cdPipeline.Environment.Name
			payload.PipelineName = cdPipeline.Name
		}
		payload.Stage = string(event.CdWorkflowType)
		payload.DeploymentHistoryLink = fmt.Sprintf("/dashboard/app/%d/cd-details/%d/%d/%d/source-code", event.AppId, event.EnvId, event.PipelineId, event.CdWorkflowRunnerId)
		payload.AppDetailLink = fmt.Sprintf("/dashboard/app/%d/details/%d/pod", event.AppId, event.EnvId)
		if event.CdWorkflowType != bean.CD_WORKFLOW_TYPE_DEPLOY {
			payload.DownloadLink = fmt.Sprintf("/orchestrator/app/cd-pipeline/workflow/download/%d/%d/%d/%d", event.AppId, event.EnvId, event.PipelineId, event.CdWorkflowRunnerId)
		}
	} else if event.PipelineType == string(util.CI) {
		if ciPipeline != nil {
			payload.AppName = ciPipeline.App.AppName
			payload.PipelineName = ciPipeline.Name
		}
		payload.BuildHistoryLink = fmt.Sprintf("/dashboard/app/%d/ci-details/%d/%d/artifacts", event.AppId, event.PipelineId, event.CiWorkflowRunnerId)
	}
	return payload
}

func (impl *EventRESTClientImpl) WriteNotificationEvent(event Event) (bool, error) {
	// if notification integration is not installed then do not send the notification
	moduleInfo, err := impl.moduleService.GetModuleInfo(bean3.ModuleNameNotification)
	if err != nil {
		impl.logger.Errorw("error while getting notification module status", "err", err)
		return false, err
	}
	if moduleInfo.Status != bean3.ModuleStatusInstalled {
		impl.logger.Warnw("Notification module is not installed, hence skipping sending notification", "currentModuleStatus", moduleInfo.Status)
		return false, nil
	}

	var cdPipeline *pipelineConfig.Pipeline
	var ciPipeline *pipelineConfig.CiPipeline
	if event.PipelineId > 0 {
		if event.PipelineType == string(util.CD) {
			cdPipeline, err = impl.pipelineRepository.FindById(event.PipelineId)
			if err != nil {
				impl.logger.Errorw("error while fetching pipeline", "err", err)
				return false, err
			}
			if cdPipeline != nil {
				event.TeamId = cdPipeline.App.TeamId
			}
		} else if event.PipelineType == string(util.CI) {
			ciPipeline, err = impl.ciPipelineRepository.FindById(event.PipelineId)
			if err != nil {
				impl.logger.Errorw("error while fetching pipeline", "err", err)
				return false, err
			}
			if ciPipeline != nil {
				event.TeamId = ciPipeline.App.TeamId
			}
		}
	}

	payload := impl.buildFinalPayload(event, cdPipeline, ciPipeline)
	event.Payload = payload

	isPreStageExist := false
	isPostStageExist := false
	if cdPipeline != nil && len(cdPipeline.PreStageConfig) > 0 {
		isPreStageExist = true
	}
	if cdPipeline != nil && len(cdPipeline.PostStageConfig) > 0 {
		isPostStageExist = true
	}

	attribute, err := impl.attributesRepository.FindByKey(bean2.HostUrlKey)
	if err != nil {
		impl.logger.Errorw("there is host url configured", "ci pipeline", ciPipeline)
		return false, err
	}
	if attribute != nil {
		event.BaseUrl = attribute.Value
	}
	if event.CdWorkflowType == "" {
		_, err = impl.sendEvent(event)
	} else if event.CdWorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
		if event.EventTypeId == int(util.Success) {
			impl.logger.Debug("skip - will send from deployment or post stage")
		} else {
			_, err = impl.sendEvent(event)
		}
	} else if event.CdWorkflowType == bean.CD_WORKFLOW_TYPE_DEPLOY {
		if isPreStageExist && event.EventTypeId == int(util.Trigger) {
			impl.logger.Debug("skip - already sent from pre stage")
		} else if isPostStageExist && event.EventTypeId == int(util.Success) {
			impl.logger.Debug("skip - will send from post stage")
		} else {
			_, err = impl.sendEvent(event)
		}
	} else if event.CdWorkflowType == bean.CD_WORKFLOW_TYPE_POST {
		if event.EventTypeId == int(util.Trigger) {
			impl.logger.Debug("skip - already sent from pre or deployment stage")
		} else {
			_, err = impl.sendEvent(event)
		}
	}
	return true, err
}
func (impl *EventRESTClientImpl) sendEventsOnNats(body []byte) error {

	err := impl.pubsubClient.Publish(pubsub.NOTIFICATION_EVENT_TOPIC, string(body))
	if err != nil {
		impl.logger.Errorw("err while publishing msg for testing topic", "msg", body, "err", err)
		return err
	}
	return nil

}

// do not call this method if notification module is not installed
func (impl *EventRESTClientImpl) sendEvent(event Event) (bool, error) {
	impl.logger.Debugw("event before send", "event", event)

	// Step 1: Create payload and destination URL based on config
	bodyBytes, destinationUrl, err := impl.createPayloadAndDestination(event)
	if err != nil {
		return false, err
	}

	// Step 2: Send via appropriate medium (NATS or REST)
	return impl.deliverEvent(bodyBytes, destinationUrl)
}

func (impl *EventRESTClientImpl) createPayloadAndDestination(event Event) ([]byte, string, error) {
	if impl.config.EnableNotifierV2 {
		return impl.createV2PayloadAndDestination(event)
	}
	return impl.createDefaultPayloadAndDestination(event)
}

func (impl *EventRESTClientImpl) createV2PayloadAndDestination(event Event) ([]byte, string, error) {
	destinationUrl := impl.config.DestinationURL + "/v2"

	// Fetch notification settings
	req := repository.GetRulesRequest{
		TeamId:              event.TeamId,
		EnvId:               event.EnvId,
		AppId:               event.AppId,
		PipelineId:          event.PipelineId,
		PipelineType:        event.PipelineType,
		IsProdEnv:           &event.IsProdEnv,
		ClusterId:           event.ClusterId,
		EnvIdsForCiPipeline: event.EnvIdsForCiPipeline,
	}
	notificationSettings, err := impl.notificationSettingsRepository.FindNotificationSettingsWithRules(
		context.Background(), event.EventTypeId, req,
	)
	if err != nil {
		impl.logger.Errorw("error while fetching notification settings", "err", err)
		return nil, "", err
	}

	// Process notification settings into beans
	notificationSettingsBean, err := impl.processNotificationSettings(notificationSettings)
	if err != nil {
		return nil, "", err
	}

	// Create combined payload
	combinedPayload := map[string]interface{}{
		"event":                event,
		"notificationSettings": notificationSettingsBean,
	}

	bodyBytes, err := json.Marshal(combinedPayload)
	if err != nil {
		impl.logger.Errorw("error while marshaling combined event request", "err", err)
		return nil, "", err
	}

	return bodyBytes, destinationUrl, nil
}

func (impl *EventRESTClientImpl) createDefaultPayloadAndDestination(event Event) ([]byte, string, error) {
	bodyBytes, err := json.Marshal(event)
	if err != nil {
		impl.logger.Errorw("error while marshaling event request", "err", err)
		return nil, "", err
	}
	return bodyBytes, impl.config.DestinationURL, nil
}

func (impl *EventRESTClientImpl) processNotificationSettings(notificationSettings []repository.NotificationSettings) ([]*repository.NotificationSettingsBean, error) {
	notificationSettingsBean := make([]*repository.NotificationSettingsBean, 0)
	for _, item := range notificationSettings {
		config := make([]repository.ConfigEntry, 0)
		if item.Config != "" {
			if err := json.Unmarshal([]byte(item.Config), &config); err != nil {
				impl.logger.Errorw("error while unmarshaling config", "err", err)
				return nil, err
			}
		}
		notificationSettingsBean = append(notificationSettingsBean, &repository.NotificationSettingsBean{
			Id:           item.Id,
			TeamId:       item.TeamId,
			AppId:        item.AppId,
			EnvId:        item.EnvId,
			PipelineId:   item.PipelineId,
			PipelineType: item.PipelineType,
			EventTypeId:  item.EventTypeId,
			Config:       config,
			ViewId:       item.ViewId,
		})
	}
	return notificationSettingsBean, nil
}

func (impl *EventRESTClientImpl) deliverEvent(bodyBytes []byte, destinationUrl string) (bool, error) {
	if impl.config.NotificationMedium == PUB_SUB {
		if err := impl.sendEventsOnNats(bodyBytes); err != nil {
			impl.logger.Errorw("error while publishing event", "err", err)
			return false, err
		}
		return true, nil
	}

	req, err := http.NewRequest(http.MethodPost, destinationUrl, bytes.NewBuffer(bodyBytes))
	if err != nil {
		impl.logger.Errorw("error while creating HTTP request", "err", err)
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := impl.client.Do(req)
	if err != nil {
		impl.logger.Errorw("error while sending HTTP request", "err", err)
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		impl.logger.Errorw("unexpected response from notifier", "status", resp.StatusCode)
		return false, fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}

	impl.logger.Debugw("event successfully delivered", "status", resp.StatusCode)
	return true, nil
}

func (impl *EventRESTClientImpl) WriteNatsEvent(topic string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	err = impl.pubsubClient.Publish(topic, string(body))
	return err
}
