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

package pipeline

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

type CiArtifactWebhookRequest struct {
	Image        string          `json:"image"`
	ImageDigest  string          `json:"imageDigest"`
	MaterialInfo json.RawMessage `json:"materialInfo"`
	DataSource   string          `json:"dataSource"`
	PipelineName string          `json:"pipelineName"`
	WorkflowId   *int            `json:"workflowId"`
	UserId       int32           `json:"userId"`
}

type WebhookService interface {
	AuthenticateExternalCiWebhook(apiKey string) (int, error)
	SaveCiArtifactWebhook(ciPipelineId int, request *CiArtifactWebhookRequest) (id int, err error)
}

type WebhookServiceImpl struct {
	ciArtifactRepository repository.CiArtifactRepository
	logger               *zap.SugaredLogger
	ciPipelineRepository pipelineConfig.CiPipelineRepository
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository
	appService           app.AppService
	eventClient          client.EventClient
	eventFactory         client.EventFactory
	workflowDagExecutor  WorkflowDagExecutor
	ciHandler            CiHandler
}

func NewWebhookServiceImpl(
	ciArtifactRepository repository.CiArtifactRepository,
	logger *zap.SugaredLogger,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	appService app.AppService, eventClient client.EventClient,
	eventFactory client.EventFactory,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	workflowDagExecutor WorkflowDagExecutor, ciHandler CiHandler) *WebhookServiceImpl {
	return &WebhookServiceImpl{
		ciArtifactRepository: ciArtifactRepository,
		logger:               logger,
		ciPipelineRepository: ciPipelineRepository,
		appService:           appService,
		eventClient:          eventClient,
		eventFactory:         eventFactory,
		ciWorkflowRepository: ciWorkflowRepository,
		workflowDagExecutor:  workflowDagExecutor,
		ciHandler:            ciHandler,
	}
}

func (impl WebhookServiceImpl) AuthenticateExternalCiWebhook(apiKey string) (int, error) {
	impl.logger.Debug("external ci webhook auth")
	splitKey := strings.Split(apiKey, ".")

	if len(splitKey) != 2 {
		return 0, fmt.Errorf("invalid key")
	}

	encodedCiPipelineId := splitKey[0]
	sha := splitKey[1]

	ciPipelineId, err := base64.StdEncoding.DecodeString(encodedCiPipelineId)
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return 0, fmt.Errorf("invalid ci pipeline")
	}
	id, err := strconv.Atoi(string(ciPipelineId))
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return 0, fmt.Errorf("invalid ci pipeline")
	}
	externalCiPipeline, err := impl.ciPipelineRepository.FindExternalCiByCiPipelineId(id)
	if externalCiPipeline.AccessToken != sha {
		return 0, fmt.Errorf("invalid key, auth failed")
	}
	return id, nil
}

func (impl WebhookServiceImpl) SaveCiArtifactWebhook(ciPipelineId int, request *CiArtifactWebhookRequest) (id int, err error) {
	impl.logger.Infow("webhook for artifact save", "req", request)
	if request.WorkflowId != nil {
		savedWorkflow, err := impl.ciWorkflowRepository.FindById(*request.WorkflowId)
		if err != nil {
			impl.logger.Errorw("cannot get saved wf", "err", err)
			return 0, err
		}
		savedWorkflow.Status = string(v1alpha1.NodeSucceeded)
		impl.logger.Debugw("updating workflow ", "savedWorkflow", savedWorkflow)
		err = impl.ciWorkflowRepository.UpdateWorkFlow(savedWorkflow)
		if err != nil {
			impl.logger.Errorw("update wf failed for id ", "err", err)
			return 0, err
		}
	}

	pipeline, err := impl.ciPipelineRepository.FindByCiAndAppDetailsById(ciPipelineId)
	if request.PipelineName == "" {
		request.PipelineName = pipeline.Name
	}
	if request.DataSource == "" {
		request.DataSource = "EXTERNAL"
	}
	if err != nil {
		impl.logger.Errorw("unable to find pipeline", "name", request.PipelineName, "err", err)
		return 0, err
	}
	materialJson, err := request.MaterialInfo.MarshalJSON()
	if err != nil {
		impl.logger.Errorw("unable to marshal material metadata", "err", err)
		return 0, err
	}
	dst := new(bytes.Buffer)
	err = json.Compact(dst, materialJson)
	if err != nil {
		return 0, err
	}
	materialJson = dst.Bytes()
	artifact := &repository.CiArtifact{
		Image:        request.Image,
		ImageDigest:  request.ImageDigest,
		MaterialInfo: string(materialJson),
		DataSource:   request.DataSource,
		PipelineId:   pipeline.Id,
		WorkflowId:   request.WorkflowId,
		ScanEnabled:  pipeline.ScanEnabled,
		Scanned:      false,
		AuditLog:     sql.AuditLog{CreatedBy: request.UserId, UpdatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now()},
	}
	if pipeline.ScanEnabled {
		artifact.Scanned = true
	}
	if err = impl.ciArtifactRepository.Save(artifact); err != nil {
		impl.logger.Errorw("error in saving material", "err", err)
		return 0, err
	}

	childrenCi, err := impl.ciPipelineRepository.FindByParentCiPipelineId(ciPipelineId)
	if err != nil && !util2.IsErrNoRows(err) {
		impl.logger.Errorw("error while fetching childern ci ", "err", err)
		return 0, err
	}

	var ciArtifactArr []*repository.CiArtifact
	for _, ci := range childrenCi {
		ciArtifact := &repository.CiArtifact{
			Image:            request.Image,
			ImageDigest:      request.ImageDigest,
			MaterialInfo:     string(materialJson),
			DataSource:       request.DataSource,
			PipelineId:       ci.Id,
			ParentCiArtifact: artifact.Id,
			ScanEnabled:      ci.ScanEnabled,
			Scanned:          false,
			AuditLog:         sql.AuditLog{CreatedBy: request.UserId, UpdatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now()},
		}
		if ci.ScanEnabled {
			ciArtifact.Scanned = true
		}
		ciArtifactArr = append(ciArtifactArr, ciArtifact)
	}

	impl.logger.Debugw("saving ci artifacts", "art", ciArtifactArr)
	if len(ciArtifactArr) > 0 {
		err = impl.ciArtifactRepository.SaveAll(ciArtifactArr)
		if err != nil {
			impl.logger.Errorw("error while saving ci artifacts", "err", err)
			return 0, err
		}
	}
	ciArtifactArr = append(ciArtifactArr, artifact)

	go impl.WriteCISuccessEvent(request, pipeline, artifact)

	impl.ciHandler.WriteToCreateTestSuites(pipeline.Id, *request.WorkflowId, int(request.UserId))

	isCiManual := true
	if request.UserId == 1 {
		impl.logger.Debugw("Trigger (auto) by system user", "userId", request.UserId)
		isCiManual = false
	} else {
		impl.logger.Debugw("Trigger (manual) by user", "userId", request.UserId)
	}
	//FIXME
	/*async := true
	if len(ciArtifactArr) > 1 {
		async = false
	}*/
	async := false
	for _, ciArtifact := range ciArtifactArr {
		err = impl.workflowDagExecutor.HandleCiSuccessEvent(ciArtifact, isCiManual, async, request.UserId)
		if err != nil {
			impl.logger.Errorw("err", "err", err)
			return 0, err
		}
	}
	return artifact.Id, err
}

func (impl *WebhookServiceImpl) WriteCISuccessEvent(request *CiArtifactWebhookRequest, pipeline *pipelineConfig.CiPipeline, artifact *repository.CiArtifact) {
	event := impl.eventFactory.Build(util.Success, &pipeline.Id, pipeline.AppId, nil, util.CI)
	event.CiArtifactId = artifact.Id
	if artifact.WorkflowId != nil {
		event.CiWorkflowRunnerId = *artifact.WorkflowId
	}
	event.UserId = int(request.UserId)
	event = impl.eventFactory.BuildExtraCIData(event, nil, artifact.Image)
	_, evtErr := impl.eventClient.WriteEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("error in writing event", "err", evtErr)
	}
}

func (impl *WebhookServiceImpl) BuildPayload(request *CiArtifactWebhookRequest, pipeline *pipelineConfig.CiPipeline) *client.Payload {
	payload := &client.Payload{}
	payload.AppName = pipeline.App.AppName
	payload.PipelineName = pipeline.Name

	var ciMaterials []*repository.CiMaterialInfo
	err := json.Unmarshal(request.MaterialInfo, &ciMaterials)
	if err != nil {
		impl.logger.Errorw("err", "err", err)
	}

	for _, material := range ciMaterials {
		if material.Modifications != nil && len(material.Modifications) > 0 {
			revision := material.Modifications[0].Revision
			if payload.Source == "" {
				payload.Source = revision
			}
			payload.Source = payload.Source + "," + revision
		}
	}
	payload.DockerImageUrl = request.Image
	return payload
}
