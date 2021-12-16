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
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/devtron-labs/devtron/util/event"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"strings"
	"time"
)

type EventFactory interface {
	Build(eventType util.EventType, sourceId *int, appId int, envId *int, pipelineType util.PipelineType) Event
	BuildExtraCDData(event Event, wfr *pipelineConfig.CdWorkflowRunner, pipelineOverrideId int, stage bean2.CdWorkflowType) Event
	BuildExtraCIData(event Event, material *MaterialTriggerInfo, dockerImage string) Event
	//BuildFinalData(event Event) *Payload
}

type EventSimpleFactoryImpl struct {
	logger                       *zap.SugaredLogger
	cdWorkflowRepository         pipelineConfig.CdWorkflowRepository
	pipelineOverrideRepository   chartConfig.PipelineOverrideRepository
	ciWorkflowRepository         pipelineConfig.CiWorkflowRepository
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
	pipelineRepository           pipelineConfig.PipelineRepository
	userRepository               repository.UserRepository
}

func NewEventSimpleFactoryImpl(logger *zap.SugaredLogger, cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository, ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository, pipelineRepository pipelineConfig.PipelineRepository,
	userRepository repository.UserRepository) *EventSimpleFactoryImpl {
	return &EventSimpleFactoryImpl{
		logger:                       logger,
		cdWorkflowRepository:         cdWorkflowRepository,
		pipelineOverrideRepository:   pipelineOverrideRepository,
		ciWorkflowRepository:         ciWorkflowRepository,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		ciPipelineRepository:         ciPipelineRepository,
		pipelineRepository:           pipelineRepository,
		userRepository:               userRepository,
	}
}

func (impl *EventSimpleFactoryImpl) Build(eventType util.EventType, sourceId *int, appId int, envId *int, pipelineType util.PipelineType) Event {
	correlationId := uuid.NewV4()
	event := Event{}
	event.EventTypeId = int(eventType)
	if sourceId != nil {
		event.PipelineId = *sourceId
	}
	event.AppId = appId
	if envId != nil {
		event.EnvId = *envId
	}
	event.PipelineType = string(pipelineType)
	event.CorrelationId = fmt.Sprintf("%s", correlationId)
	event.EventTime = time.Now().Format(bean.LayoutRFC3339)
	return event
}

func (impl *EventSimpleFactoryImpl) BuildExtraCDData(event Event, wfr *pipelineConfig.CdWorkflowRunner, pipelineOverrideId int, stage bean2.CdWorkflowType) Event {
	//event.CdWorkflowRunnerId =
	event.CdWorkflowType = stage
	payload := event.Payload
	if payload == nil {
		payload = &Payload{}
		payload.Stage = string(stage)
		event.Payload = payload
	}
	if wfr != nil {
		material, err := impl.getCiMaterialInfo(wfr.CdWorkflow.Pipeline.CiPipelineId, wfr.CdWorkflow.CiArtifactId)
		if err != nil {
			impl.logger.Errorw("found error on payload build for cd stages, skipping this error ", "material", material)
		}
		payload.MaterialTriggerInfo = material
		payload.DockerImageUrl = wfr.CdWorkflow.CiArtifact.Image
		event.UserId = int(wfr.TriggeredBy)
		event.Payload = payload
		event.CdWorkflowRunnerId = wfr.Id
		event.CiArtifactId = wfr.CdWorkflow.CiArtifactId
	} else if pipelineOverrideId > 0 {
		pipelineOverride, err := impl.pipelineOverrideRepository.FindById(pipelineOverrideId)
		if err != nil {
			impl.logger.Errorw("found error on payload build for cd stages, skipping this error ", "pipelineOverride", pipelineOverride)
		}
		if pipelineOverride != nil {
			cdWorkflow, err := impl.cdWorkflowRepository.FindById(pipelineOverride.CdWorkflowId)
			if err != nil {
				impl.logger.Errorw("found error on payload build for cd stages, skipping this error ", "cdWorkflow", cdWorkflow)
			}
			wfr, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(cdWorkflow.Id, stage)
			if err != nil {
				impl.logger.Errorw("found error on payload build for cd stages, skipping this error ", "wfr", wfr)
			}
			event.CdWorkflowRunnerId = wfr.Id
			event.CiArtifactId = pipelineOverride.CiArtifactId

			material, err := impl.getCiMaterialInfo(pipelineOverride.CiArtifact.PipelineId, pipelineOverride.CiArtifactId)
			if err != nil {
				impl.logger.Errorw("found error on payload build for cd stages, skipping this error ", "material", material)
			}
			payload.MaterialTriggerInfo = material
			payload.DockerImageUrl = wfr.CdWorkflow.CiArtifact.Image
			event.UserId = int(wfr.TriggeredBy)
		}
		event.Payload = payload
	} else if event.PipelineId > 0 {
		pipeline, err := impl.pipelineRepository.FindById(event.PipelineId)
		if err != nil {
			impl.logger.Errorw("found error on payload build for cd stages, skipping this error ", "pipeline", pipeline)
		}
		if pipeline != nil {
			material, err := impl.getCiMaterialInfo(pipeline.CiPipelineId, 0)
			if err != nil {
				impl.logger.Errorw("found error on payload build for cd stages, skipping this error ", "material", material)
			}
			payload.MaterialTriggerInfo = material
		}
		event.Payload = payload
	}

	if event.UserId > 0 {
		user, err := impl.userRepository.GetById(int32(event.UserId))
		if err != nil {
			impl.logger.Errorw("found error on payload build for cd stages, skipping this error ", "user", user)
		}
		payload = event.Payload
		payload.TriggeredBy = user.EmailId
		event.Payload = payload
	}
	return event
}

func (impl *EventSimpleFactoryImpl) BuildExtraCIData(event Event, material *MaterialTriggerInfo, dockerImage string) Event {
	if material == nil {
		materialInfo, err := impl.getCiMaterialInfo(event.PipelineId, event.CiArtifactId)
		if err != nil {
			impl.logger.Errorw("found error on payload build for ci, skipping this error ", "materialInfo", materialInfo)
		}
		material = materialInfo
	} else if material.CiMaterials == nil {
		materialInfo, err := impl.getCiMaterialInfo(event.PipelineId, 0)
		if err != nil {
			impl.logger.Errorw("found error on payload build for ci, skipping this error ", "materialInfo", materialInfo)
		}
		materialInfo.GitTriggers = material.GitTriggers
		material = materialInfo
	}
	payload := event.Payload
	if payload == nil {
		payload = &Payload{}
		event.Payload = payload
	}
	event.Payload.MaterialTriggerInfo = material

	if event.UserId > 0 {
		user, err := impl.userRepository.GetById(int32(event.UserId))
		if err != nil {
			impl.logger.Errorw("found error on payload build for cd stages, skipping this error ", "user", user)
		}
		payload = event.Payload
		payload.TriggeredBy = user.EmailId
		event.Payload = payload
	}
	return event
}

func (impl *EventSimpleFactoryImpl) getCiMaterialInfo(ciPipelineId int, ciArtifactId int) (*MaterialTriggerInfo, error) {
	materialTriggerInfo := &MaterialTriggerInfo{}
	if ciPipelineId > 0 {
		ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(ciPipelineId)
		if err != nil {
			impl.logger.Errorw("err", err)
			return nil, err
		}

		var ciMaterialsArr []CiPipelineMaterialResponse
		for _, m := range ciMaterials {
			res := CiPipelineMaterialResponse{
				Id:              m.Id,
				GitMaterialId:   m.GitMaterialId,
				GitMaterialName: m.GitMaterial.Name[strings.Index(m.GitMaterial.Name, "-")+1:],
				Type:            string(m.Type),
				Value:           m.Value,
				Active:          m.Active,
				Url:             m.GitMaterial.Url,
			}
			ciMaterialsArr = append(ciMaterialsArr, res)
		}
		materialTriggerInfo.CiMaterials = ciMaterialsArr
	}
	if ciArtifactId > 0 {
		ciWf, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflowByArtifactId(ciArtifactId)
		if err != nil {
			return nil, err
		}
		materialTriggerInfo.GitTriggers = ciWf.GitTriggers
	}
	return materialTriggerInfo, nil
}
