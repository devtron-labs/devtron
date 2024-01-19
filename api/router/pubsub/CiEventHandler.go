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
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/caarlos0/env/v6"
	pubsub "github.com/devtron-labs/common-lib-private/pubsub-lib"
	"github.com/devtron-labs/common-lib-private/pubsub-lib/model"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"k8s.io/utils/pointer"
	"time"
)

type CiEventConfig struct {
	ExposeCiMetrics bool `env:"EXPOSE_CI_METRICS" envDefault:"false"`
}

func GetCiEventConfig() (*CiEventConfig, error) {
	cfg := &CiEventConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

type CiEventHandler interface {
	Subscribe() error
	BuildCiArtifactRequest(event CiCompleteEvent) (*pipeline.CiArtifactWebhookRequest, error)
	BuildCiArtifactRequestForWebhook(event CiCompleteEvent) (*pipeline.CiArtifactWebhookRequest, error)
}

type CiEventHandlerImpl struct {
	logger         *zap.SugaredLogger
	pubsubClient   *pubsub.PubSubClientServiceImpl
	webhookService pipeline.WebhookService
	validator      *validator.Validate
	ciEventConfig  *CiEventConfig
}

type ImageDetailsFromCR struct {
	ImageDetails []types.ImageDetail `json:"imageDetails"`
	Region       string              `json:"region"`
}

type CiCompleteEvent struct {
	CiProjectDetails              []bean2.CiProjectDetails `json:"ciProjectDetails"`
	DockerImage                   string                   `json:"dockerImage" validate:"required,image-validator"`
	Digest                        string                   `json:"digest"`
	PipelineId                    int                      `json:"pipelineId"`
	WorkflowId                    *int                     `json:"workflowId"`
	TriggeredBy                   int32                    `json:"triggeredBy"`
	PipelineName                  string                   `json:"pipelineName"`
	DataSource                    string                   `json:"dataSource"`
	MaterialType                  string                   `json:"materialType"`
	Metrics                       util.CIMetrics           `json:"metrics"`
	AppName                       string                   `json:"appName"`
	IsArtifactUploaded            bool                     `json:"isArtifactUploaded"`
	FailureReason                 string                   `json:"failureReason"`
	ImageDetailsFromCR            *ImageDetailsFromCR      `json:"imageDetailsFromCR"`
	PluginRegistryArtifactDetails map[string][]string      `json:"PluginRegistryArtifactDetails"`
	PluginArtifactStage           string                   `json:"pluginArtifactStage"`
}

func NewCiEventHandlerImpl(logger *zap.SugaredLogger, pubsubClient *pubsub.PubSubClientServiceImpl, webhookService pipeline.WebhookService, validator *validator.Validate, ciEventConfig *CiEventConfig) *CiEventHandlerImpl {
	ciEventHandlerImpl := &CiEventHandlerImpl{
		logger:         logger,
		pubsubClient:   pubsubClient,
		webhookService: webhookService,
		validator:      validator,
		ciEventConfig:  ciEventConfig,
	}
	err := ciEventHandlerImpl.Subscribe()
	if err != nil {
		logger.Error(err)
		return nil
	}
	return ciEventHandlerImpl
}

func (impl *CiEventHandlerImpl) Subscribe() error {
	callback := func(msg *model.PubSubMsg) {
		ciCompleteEvent := CiCompleteEvent{}
		err := json.Unmarshal([]byte(string(msg.Data)), &ciCompleteEvent)
		if err != nil {
			impl.logger.Error("error while unmarshalling json data", "error", err)
			return
		}
		impl.logger.Debugw("ci complete event for ci", "ciPipelineId", ciCompleteEvent.PipelineId)
		req, err := impl.BuildCiArtifactRequest(ciCompleteEvent)
		if err != nil {
			return
		}

		triggerContext := pipeline.TriggerContext{
			ReferenceId: pointer.String(msg.MsgId),
		}

		if ciCompleteEvent.FailureReason != "" {
			req.FailureReason = ciCompleteEvent.FailureReason
			err := impl.webhookService.HandleCiStepFailedEvent(ciCompleteEvent.PipelineId, req)
			if err != nil {
				impl.logger.Error("Error while sending event for CI failure for pipelineID: ",
					ciCompleteEvent.PipelineId, "request: ", req, "error: ", err)
				return
			}
		} else if ciCompleteEvent.ImageDetailsFromCR != nil {
			if len(ciCompleteEvent.ImageDetailsFromCR.ImageDetails) > 0 {
				imageDetails := util.GetReverseSortedImageDetails(ciCompleteEvent.ImageDetailsFromCR.ImageDetails)
				digestWorkflowMap, err := impl.webhookService.HandleMultipleImagesFromEvent(imageDetails, *ciCompleteEvent.WorkflowId)
				if err != nil {
					impl.logger.Errorw("error in getting digest workflow map", "err", err, "workflowId", ciCompleteEvent.WorkflowId)
					return
				}
				for _, detail := range imageDetails {
					request, err := impl.BuildCIArtifactRequestForImageFromCR(detail, ciCompleteEvent.ImageDetailsFromCR.Region, ciCompleteEvent, digestWorkflowMap[*detail.ImageDigest].Id)
					if err != nil {
						impl.logger.Error("Error while creating request for pipelineID", "pipelineId", ciCompleteEvent.PipelineId, "err", err)
						return
					}
					resp, err := impl.ValidateAndHandleCiSuccessEvent(triggerContext, ciCompleteEvent.PipelineId, request, detail.ImagePushedAt)
					if err != nil {
						return
					}
					impl.logger.Debug("response of handle ci success event for multiple images from plugin", "resp", resp)
				}
			}

		} else {
			util.TriggerCIMetrics(ciCompleteEvent.Metrics, impl.ciEventConfig.ExposeCiMetrics, ciCompleteEvent.PipelineName, ciCompleteEvent.AppName)
			resp, err := impl.ValidateAndHandleCiSuccessEvent(triggerContext, ciCompleteEvent.PipelineId, req, &time.Time{})
			if err != nil {
				return
			}
			impl.logger.Debug(resp)
		}
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		ciCompleteEvent := CiCompleteEvent{}
		err := json.Unmarshal([]byte(string(msg.Data)), &ciCompleteEvent)
		if err != nil {
			return "error while unmarshalling json data", []interface{}{"error", err}
		}
		return "got message for ci-completion", []interface{}{"ciPipelineId", ciCompleteEvent.PipelineId, "workflowId", ciCompleteEvent.WorkflowId}
	}

	validations := impl.webhookService.GetTriggerValidateFuncs()
	err := impl.pubsubClient.Subscribe(pubsub.CI_COMPLETE_TOPIC, callback, loggerFunc, validations...)
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}

func (impl *CiEventHandlerImpl) ValidateAndHandleCiSuccessEvent(triggerContext pipeline.TriggerContext, ciPipelineId int, request *pipeline.CiArtifactWebhookRequest, imagePushedAt *time.Time) (int, error) {
	validationErr := impl.validator.Struct(request)
	if validationErr != nil {
		impl.logger.Errorw("validation err, HandleCiSuccessEvent", "err", validationErr, "payload", request)
		return 0, validationErr
	}
	buildArtifactId, err := impl.webhookService.HandleCiSuccessEvent(triggerContext, ciPipelineId, request, imagePushedAt)
	if err != nil {
		impl.logger.Error("Error while sending event for CI success for pipelineID",
			ciPipelineId, "request", request, "error", err)
		return 0, err
	}
	return buildArtifactId, nil
}

func (impl *CiEventHandlerImpl) BuildCiArtifactRequest(event CiCompleteEvent) (*pipeline.CiArtifactWebhookRequest, error) {
	var ciMaterialInfos []repository.CiMaterialInfo
	for _, p := range event.CiProjectDetails {
		var modifications []repository.Modification

		var branch string
		var tag string
		var webhookData repository.WebhookData
		if p.SourceType == pipelineConfig.SOURCE_TYPE_BRANCH_FIXED {
			branch = p.SourceValue
		} else if p.SourceType == pipelineConfig.SOURCE_TYPE_WEBHOOK {
			webhookData = repository.WebhookData{
				Id:              p.WebhookData.Id,
				EventActionType: p.WebhookData.EventActionType,
				Data:            p.WebhookData.Data,
			}
		}

		modification := repository.Modification{
			Revision:     p.CommitHash,
			ModifiedTime: p.CommitTime,
			Author:       p.Author,
			Branch:       branch,
			Tag:          tag,
			WebhookData:  webhookData,
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
		Image:                         event.DockerImage,
		ImageDigest:                   event.Digest,
		DataSource:                    event.DataSource,
		PipelineName:                  event.PipelineName,
		MaterialInfo:                  rawMaterialInfo,
		UserId:                        event.TriggeredBy,
		WorkflowId:                    event.WorkflowId,
		IsArtifactUploaded:            event.IsArtifactUploaded,
		PluginRegistryArtifactDetails: event.PluginRegistryArtifactDetails,
		PluginArtifactStage:           event.PluginArtifactStage,
	}
	// if DataSource is empty, repository.WEBHOOK is considered as default
	if request.DataSource == "" {
		request.DataSource = repository.WEBHOOK
	}
	return request, nil
}

func (impl *CiEventHandlerImpl) BuildCIArtifactRequestForImageFromCR(imageDetails types.ImageDetail, region string, event CiCompleteEvent, workflowId int) (*pipeline.CiArtifactWebhookRequest, error) {
	if event.TriggeredBy == 0 {
		event.TriggeredBy = 1 // system triggered event
	}
	request := &pipeline.CiArtifactWebhookRequest{
		Image:              util.ExtractEcrImage(*imageDetails.RegistryId, region, *imageDetails.RepositoryName, imageDetails.ImageTags[0]),
		ImageDigest:        *imageDetails.ImageDigest,
		DataSource:         event.DataSource,
		PipelineName:       event.PipelineName,
		UserId:             event.TriggeredBy,
		WorkflowId:         &workflowId,
		IsArtifactUploaded: event.IsArtifactUploaded,
	}
	if request.DataSource == "" {
		request.DataSource = repository.WEBHOOK
	}
	return request, nil
}

func (impl *CiEventHandlerImpl) BuildCiArtifactRequestForWebhook(event CiCompleteEvent) (*pipeline.CiArtifactWebhookRequest, error) {
	ciMaterialInfos := make([]repository.CiMaterialInfo, 0)
	if event.MaterialType == "" {
		event.MaterialType = "git"
	}
	for _, p := range event.CiProjectDetails {
		var modifications []repository.Modification

		var branch string
		var tag string
		var webhookData repository.WebhookData
		if p.SourceType == pipelineConfig.SOURCE_TYPE_BRANCH_FIXED {
			branch = p.SourceValue
		} else if p.SourceType == pipelineConfig.SOURCE_TYPE_WEBHOOK {
			webhookData = repository.WebhookData{
				Id:              p.WebhookData.Id,
				EventActionType: p.WebhookData.EventActionType,
				Data:            p.WebhookData.Data,
			}
		}

		modification := repository.Modification{
			Revision:     p.CommitHash,
			ModifiedTime: p.CommitTime,
			Author:       p.Author,
			Branch:       branch,
			Tag:          tag,
			WebhookData:  webhookData,
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
		Image:              event.DockerImage,
		ImageDigest:        event.Digest,
		DataSource:         event.DataSource,
		PipelineName:       event.PipelineName,
		MaterialInfo:       rawMaterialInfo,
		UserId:             event.TriggeredBy,
		WorkflowId:         event.WorkflowId,
		IsArtifactUploaded: event.IsArtifactUploaded,
	}
	// if DataSource is empty, repository.WEBHOOK is considered as default
	if request.DataSource == "" {
		request.DataSource = repository.WEBHOOK
	}
	return request, nil
}
