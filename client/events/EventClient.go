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

package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util "github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type EventClientConfig struct {
	DestinationURL string `env:"EVENT_URL" envDefault:"http://localhost:3000/notify"`
	TestSuitURL    string `env:"TEST_SUIT_URL" envDefault:"http://localhost:3000"`
}

func GetEventClientConfig() (*EventClientConfig, error) {
	cfg := &EventClientConfig{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, errors.New("could not get event service url")
	}
	return cfg, err
}

type EventClient interface {
	WriteEvent(event Event) (bool, error)
	WriteNatsEvent(channel string, payload interface{}) error
	SendTestSuite(reqBody []byte) (bool, error)
}

type Event struct {
	EventTypeId        int                 `json:"eventTypeId"`
	EventName          string              `json:"eventName"`
	PipelineId         int                 `json:"pipelineId"`
	PipelineType       string              `json:"pipelineType"`
	CorrelationId      string              `json:"correlationId"`
	Payload            *Payload            `json:"payload"`
	EventTime          string              `json:"eventTime"`
	TeamId             int                 `json:"teamId"`
	AppId              int                 `json:"appId"`
	EnvId              int                 `json:"envId"`
	CdWorkflowType     bean.CdWorkflowType `json:"cdWorkflowType,omitempty"`
	CdWorkflowRunnerId int                 `json:"cdWorkflowRunnerId"`
	CiWorkflowRunnerId int                 `json:"ciWorkflowRunnerId"`
	CiArtifactId       int                 `json:"ciArtifactId"`
	UserId             int                 `json:"-"`
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
	logger               *zap.SugaredLogger
	client               *http.Client
	config               *EventClientConfig
	pubsubClient         *pubsub.PubSubClient
	ciPipelineRepository pipelineConfig.CiPipelineRepository
	pipelineRepository   pipelineConfig.PipelineRepository
}

func NewEventRESTClientImpl(logger *zap.SugaredLogger, client *http.Client, config *EventClientConfig, pubsubClient *pubsub.PubSubClient,
	ciPipelineRepository pipelineConfig.CiPipelineRepository, pipelineRepository pipelineConfig.PipelineRepository) *EventRESTClientImpl {
	return &EventRESTClientImpl{logger: logger, client: client, config: config, pubsubClient: pubsubClient,
		ciPipelineRepository: ciPipelineRepository, pipelineRepository: pipelineRepository}
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

func (impl *EventRESTClientImpl) WriteEvent(event Event) (bool, error) {
	var cdPipeline *pipelineConfig.Pipeline
	var ciPipeline *pipelineConfig.CiPipeline
	var err error
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
	if event.CdWorkflowType == "" {
		_, err = impl.SendEvent(event)
	} else if event.CdWorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
		if event.EventTypeId == int(util.Success) {
			impl.logger.Debug("skip - will send from deployment or post stage")
		} else {
			_, err = impl.SendEvent(event)
		}
	} else if event.CdWorkflowType == bean.CD_WORKFLOW_TYPE_DEPLOY {
		if isPreStageExist && event.EventTypeId == int(util.Trigger) {
			impl.logger.Debug("skip - already sent from pre stage")
		} else if isPostStageExist && event.EventTypeId == int(util.Success) {
			impl.logger.Debug("skip - will send from post stage")
		} else {
			_, err = impl.SendEvent(event)
		}
	} else if event.CdWorkflowType == bean.CD_WORKFLOW_TYPE_POST {
		if event.EventTypeId == int(util.Trigger) {
			impl.logger.Debug("skip - already sent from pre or deployment stage")
		} else {
			_, err = impl.SendEvent(event)
		}
	}
	return true, err
}

func (impl *EventRESTClientImpl) SendEvent(event Event) (bool, error) {
	impl.logger.Debugw("event before send", "event", event)
	body, err := json.Marshal(event)
	if err != nil {
		impl.logger.Errorw("error while marshaling event request ", "err", err)
		return false, err
	}
	var reqBody = []byte(body)
	req, err := http.NewRequest(http.MethodPost, impl.config.DestinationURL, bytes.NewBuffer(reqBody))
	if err != nil {
		impl.logger.Errorw("error while writing event", "err", err)
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := impl.client.Do(req)
	if err != nil {
		impl.logger.Errorw("error while UpdateJiraTransition request ", "err", err)
		return false, err
	}
	impl.logger.Debugw("event completed", "event resp", resp)
	return true, err
}

func (impl *EventRESTClientImpl) WriteNatsEvent(channel string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	err = impl.pubsubClient.Conn.Publish(channel, body)
	return err
}

func (impl *EventRESTClientImpl) SendTestSuite(reqBody []byte) (bool, error) {
	impl.logger.Debugw("request", "body", string(reqBody))
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/triggers", impl.config.TestSuitURL), bytes.NewBuffer(reqBody))
	if err != nil {
		impl.logger.Errorw("error while writing test suites", "err", err)
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := impl.client.Do(req)
	if err != nil {
		impl.logger.Errorw("error while UpdateJiraTransition request ", "err", err)
		return false, err
	}
	impl.logger.Debugw("response from test suit create api", "status code", resp.StatusCode)
	return true, err
}
