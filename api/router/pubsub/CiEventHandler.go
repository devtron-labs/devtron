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
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"encoding/json"
	"fmt"
	"github.com/nats-io/stan"
	"go.uber.org/zap"
	"time"
)

type CiEventHandler interface {
	Subscribe() error
	BuildCiArtifactRequest(event CiCompleteEvent) (*pipeline.CiArtifactWebhookRequest, error)
}

type CiEventHandlerImpl struct {
	logger         *zap.SugaredLogger
	pubsubClient   *pubsub.PubSubClient
	webhookService pipeline.WebhookService
}

type CiCompleteEvent struct {
	CiProjectDetails []pipeline.CiProjectDetails `json:"ciProjectDetails"`
	DockerImage      string                      `json:"dockerImage" validate:"required"`
	Digest           string                      `json:"digest" validate:"required"`
	PipelineId       int                         `json:"pipelineId"`
	WorkflowId       *int                        `json:"workflowId"`
	TriggeredBy      int32                       `json:"triggeredBy"`
	PipelineName     string                      `json:"pipelineName"`
	DataSource       string                      `json:"dataSource"`
	MaterialType     string                      `json:"materialType" validate:"required"`
}

const CI_COMPLETE_TOPIC = "CI-RUNNER.CI-COMPLETE"
const CI_COMPLETE_GROUP = "CI-RUNNER.CI-COMPLETE_GROUP-1"
const CI_COMPLETE_DURABLE = "CI-RUNNER.CI-COMPLETE_DURABLE-1"

func NewCiEventHandlerImpl(logger *zap.SugaredLogger, pubsubClient *pubsub.PubSubClient, webhookService pipeline.WebhookService) *CiEventHandlerImpl {
	ciEventHandlerImpl := &CiEventHandlerImpl{
		logger:         logger,
		pubsubClient:   pubsubClient,
		webhookService: webhookService,
	}
	err := ciEventHandlerImpl.Subscribe()
	if err != nil {
		logger.Error(err)
		return nil
	}
	return ciEventHandlerImpl
}

func (impl *CiEventHandlerImpl) Subscribe() error {
	_, err := impl.pubsubClient.Conn.QueueSubscribe(CI_COMPLETE_TOPIC, CI_COMPLETE_GROUP, func(msg *stan.Msg) {
		impl.logger.Debug("ci complete event received")
		defer msg.Ack()
		ciCompleteEvent := CiCompleteEvent{}
		err := json.Unmarshal([]byte(string(msg.Data)), &ciCompleteEvent)
		if err != nil {
			impl.logger.Error("err", err)
			return
		}
		impl.logger.Debugw("ci complete event for ci", "ciPipelineId", ciCompleteEvent.PipelineId)
		req, err := impl.BuildCiArtifactRequest(ciCompleteEvent)
		if err != nil {
			return
		}
		resp, err := impl.webhookService.SaveCiArtifactWebhook(ciCompleteEvent.PipelineId, req)
		if err != nil {
			impl.logger.Error(err)
			return
		}
		impl.logger.Debug(resp)
	}, stan.DurableName(CI_COMPLETE_DURABLE), stan.StartWithLastReceived(), stan.AckWait(time.Duration(impl.pubsubClient.AckDuration)*time.Second), stan.SetManualAckMode(), stan.MaxInflight(1))
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}

func (impl *CiEventHandlerImpl) BuildCiArtifactRequest(event CiCompleteEvent) (*pipeline.CiArtifactWebhookRequest, error) {
	var ciMaterialInfos []repository.CiMaterialInfo
	for _, p := range event.CiProjectDetails {
		var modifications []repository.Modification

		var branch string
		var tag string
		if p.SourceType == pipelineConfig.SOURCE_TYPE_BRANCH_FIXED {
			branch = p.SourceValue
		}
		if p.SourceType == pipelineConfig.SOURCE_TYPE_TAG_REGEX {
			tag = p.SourceValue
		}

		modification := repository.Modification{
			Revision:     p.CommitHash,
			ModifiedTime: p.CommitTime.Format(bean.LayoutRFC3339),
			Author:       p.Author,
			Branch:       branch,
			Tag:          tag,
			Message:      p.Message,
		}
		modifications = append(modifications, modification)
		ciMaterialInfo := repository.CiMaterialInfo{
			Material: repository.Material{
				GitConfiguration: repository.GitConfiguration{
					URL: p.GitRepository,
				},
				Type: event.MaterialType,
			},
			Changed:       true,
			Modifications: modifications,
		}
		ciMaterialInfos = append(ciMaterialInfos, ciMaterialInfo)
	}

	materialBytes, err := json.Marshal(ciMaterialInfos)
	if err != nil {
		impl.logger.Errorw("cannot build ci artifact req", "err", err)
		return nil, err
	}
	rawMaterialInfo := json.RawMessage(materialBytes)
	fmt.Printf("Raw Message : %s\n", rawMaterialInfo)

	if event.TriggeredBy == 0 {
		event.TriggeredBy = 1 // system triggered event
	}

	request := &pipeline.CiArtifactWebhookRequest{
		Image:        event.DockerImage,
		ImageDigest:  event.Digest,
		DataSource:   event.DataSource,
		PipelineName: event.PipelineName,
		MaterialInfo: rawMaterialInfo,
		UserId:       event.TriggeredBy,
		WorkflowId:   event.WorkflowId,
	}
	return request, nil
}
