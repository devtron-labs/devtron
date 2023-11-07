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
	"context"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository/imageTagging"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/bean"
	repository4 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/notifier"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"strings"
	"time"
)

type EventFactory interface {
	Build(eventType util.EventType, sourceId *int, appId int, envId *int, pipelineType util.PipelineType) Event
	BuildExtraCDData(event Event, wfr *pipelineConfig.CdWorkflowRunner, pipelineOverrideId int, stage bean2.WorkflowType) Event
	BuildExtraApprovalData(event Event, approvalActionRequest bean.UserApprovalActionRequest, pipeline *pipelineConfig.Pipeline, userId int32) Event
	BuildExtraProtectConfigData(event Event, draftNotificationRequest ConfigDataForNotification) Event
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
	ciArtifactRepository         repository2.CiArtifactRepository
	DeploymentApprovalRepository pipelineConfig.DeploymentApprovalRepository
	sesNotificationRepository    repository2.SESNotificationRepository
	smtpNotificationRepository   repository2.SMTPNotificationRepository
	imageTaggingRepository       repository3.ImageTaggingRepository
	appRepo                      appRepository.AppRepository
	envRepository                repository4.EnvironmentRepository
}

func NewEventSimpleFactoryImpl(logger *zap.SugaredLogger, cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository, ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository, pipelineRepository pipelineConfig.PipelineRepository,
	userRepository repository.UserRepository, ciArtifactRepository repository2.CiArtifactRepository, DeploymentApprovalRepository pipelineConfig.DeploymentApprovalRepository,
	sesNotificationRepository repository2.SESNotificationRepository, smtpNotificationRepository repository2.SMTPNotificationRepository, imageTaggingRepository repository3.ImageTaggingRepository,
	appRepo appRepository.AppRepository, envRepository repository4.EnvironmentRepository) *EventSimpleFactoryImpl {
	return &EventSimpleFactoryImpl{
		logger:                       logger,
		cdWorkflowRepository:         cdWorkflowRepository,
		pipelineOverrideRepository:   pipelineOverrideRepository,
		ciWorkflowRepository:         ciWorkflowRepository,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		ciPipelineRepository:         ciPipelineRepository,
		pipelineRepository:           pipelineRepository,
		userRepository:               userRepository,
		ciArtifactRepository:         ciArtifactRepository,
		DeploymentApprovalRepository: DeploymentApprovalRepository,
		sesNotificationRepository:    sesNotificationRepository,
		smtpNotificationRepository:   smtpNotificationRepository,
		imageTaggingRepository:       imageTaggingRepository,
		appRepo:                      appRepo,
		envRepository:                envRepository,
	}
}

type ResourceType string

const (
	CM                 ResourceType = "ConfigMap"
	CS                 ResourceType = "Secret"
	DeploymentTemplate ResourceType = "DeploymentTemplate"
)
const AppLevelBaseUrl = "/dashboard/app/%d/edit/"
const EnvLevelBaseUrl = "/dashboard/app/%d/edit/env-override/%d/"

type ConfigDataForNotification struct {
	AppId        int
	EnvId        int
	Resource     ResourceType
	ResourceName string
	UserComment  string
	UserId       int32
	EmailIds     []string
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

func (impl *EventSimpleFactoryImpl) BuildExtraCDData(event Event, wfr *pipelineConfig.CdWorkflowRunner, pipelineOverrideId int, stage bean2.WorkflowType) Event {
	//event.CdWorkflowRunnerId =
	event.CdWorkflowType = stage
	payload := event.Payload
	if payload == nil {
		payload = &Payload{}
		payload.Stage = string(stage)
		event.Payload = payload
	}
	var emailIDs []string

	if wfr != nil && wfr.DeploymentApprovalRequestId >= 0 {
		deploymentUserData, err := impl.DeploymentApprovalRepository.FetchApprovedDataByApprovalId(wfr.DeploymentApprovalRequestId)
		if err != nil {
			impl.logger.Errorw("error in getting deploymentUserData", "err", err, "deploymentApprovalRequestId", wfr.DeploymentApprovalRequestId)
		}
		if deploymentUserData != nil {
			userIDs := []int32{}
			for _, userData := range deploymentUserData {
				userIDs = append(userIDs, userData.UserId)
			}
			users, err := impl.userRepository.GetByIds(userIDs)
			if err != nil {
				impl.logger.Errorw("UserModel not found for users", err)
			}
			emailIDs = []string{}
			for _, user := range users {
				emailIDs = append(emailIDs, user.EmailId)
			}

		}
	}

	payload.ApprovedByEmail = emailIDs
	if wfr != nil && wfr.WorkflowType != bean2.CD_WORKFLOW_TYPE_DEPLOY {
		material, err := impl.getCiMaterialInfo(wfr.CdWorkflow.Pipeline.CiPipelineId, wfr.CdWorkflow.CiArtifactId)
		if err != nil {
			impl.logger.Errorw("found error on payload build for cd stages, skipping this error ", "event", event, "stage", stage, "workflow runner", wfr, "pipelineOverrideId", pipelineOverrideId)
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
			impl.logger.Errorw("found error on payload build for cd stages, skipping this error ", "event", event, "stage", stage, "workflow runner", wfr, "pipelineOverrideId", pipelineOverrideId)
		}
		if pipelineOverride != nil && pipelineOverride.Id > 0 {
			cdWorkflow, err := impl.cdWorkflowRepository.FindById(pipelineOverride.CdWorkflowId)
			if err != nil {
				impl.logger.Errorw("found error on payload build for cd stages, skipping this error ", "cdWorkflow", cdWorkflow, "event", event, "stage", stage, "workflow runner", wfr, "pipelineOverrideId", pipelineOverrideId)
			}
			wfr, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(context.Background(), cdWorkflow.Id, stage)
			if err != nil {
				impl.logger.Errorw("found error on payload build for cd stages, skipping this error ", "wfr", wfr, "event", event, "stage", stage, "workflow runner", wfr, "pipelineOverrideId", pipelineOverrideId)
			}
			if wfr.Id > 0 {
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
			impl.logger.Errorw("error on fetching materials for", "ciPipelineId", ciPipelineId, "err", err)
			return nil, err
		}

		var ciMaterialsArr []CiPipelineMaterialResponse
		for _, m := range ciMaterials {
			if m.GitMaterial == nil {
				impl.logger.Warnw("git material are empty", "material", m)
				continue
			}
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
		ciArtifact, err := impl.ciArtifactRepository.Get(ciArtifactId)
		if err != nil {
			impl.logger.Errorw("error fetching artifact data", "err", err)
			return nil, err
		}

		// handling linked ci pipeline
		if ciArtifact.ParentCiArtifact > 0 && ciArtifact.WorkflowId == nil {
			ciArtifactId = ciArtifact.ParentCiArtifact
		}
		ciWf, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflowByArtifactId(ciArtifactId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error fetching ci workflow data by artifact", "err", err)
			return nil, err
		}
		if ciWf != nil {
			materialTriggerInfo.GitTriggers = ciWf.GitTriggers
		}
	}
	return materialTriggerInfo, nil
}
func (impl *EventSimpleFactoryImpl) BuildExtraApprovalData(event Event, approvalActionRequest bean.UserApprovalActionRequest, cdPipeline *pipelineConfig.Pipeline, userId int32) Event {
	defaultSesConfig, defaultSmtpConfig, err := impl.getDefaultSESOrSMTPConfig()
	if err != nil {
		impl.logger.Errorw("found error in getting defaultSesConfig or  defaultSmtpConfig data", "err", err)
	}
	payload := event.Payload
	if payload == nil {
		payload = &Payload{}
		event.Payload = payload
	}
	imageComment, err := impl.imageTaggingRepository.GetImageComment(approvalActionRequest.ArtifactId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error fetching imageComment", "imageComment", imageComment, "err", err)
		return event
	}
	payload.ImageComment = imageComment.Comment
	imageTags, err := impl.imageTaggingRepository.GetTagsByArtifactId(approvalActionRequest.ArtifactId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error fetching imageTags", "imageTags", imageTags, "err", err)
		return event
	}
	var imageTagNames []string
	if imageTags != nil && len(imageTags) != 0 {
		for _, tag := range imageTags {
			imageTagNames = append(imageTagNames, tag.TagName)
		}
	}
	payload.ImageTagNames = imageTagNames
	ciArtifact, err := impl.ciArtifactRepository.Get(approvalActionRequest.ArtifactId)
	if err != nil {
		impl.logger.Errorw("error fetching defaultSesConfig", "ciArtifact", ciArtifact, "err", err)
		return event
	}
	payload.AppName = cdPipeline.App.AppName
	payload.EnvName = cdPipeline.Environment.Name
	payload.PipelineName = cdPipeline.Name
	payload.DockerImageUrl = ciArtifact.Image
	lastColonIndex := strings.LastIndex(payload.DockerImageUrl, ":")
	dockerImageTag := ""
	if lastColonIndex != -1 && lastColonIndex < len(payload.DockerImageUrl)-1 {
		dockerImageTag = payload.DockerImageUrl[lastColonIndex+1:]
	}
	payload.ImageApprovalLink = fmt.Sprintf("/dashboard/app/%d/trigger?approval-node=%d&imageTag=%s", event.AppId, cdPipeline.Id, dockerImageTag)
	for _, emailId := range approvalActionRequest.ApprovalNotificationConfig.EmailIds {
		provider := &notifier.Provider{
			//Rule:      "",
			ConfigId:  0,
			Recipient: emailId,
		}
		if defaultSesConfig.Id != 0 {
			provider.Destination = "ses"
		} else if defaultSmtpConfig.Id != 0 {
			provider.Destination = "smtp"
		}
		event.Payload.Providers = append(event.Payload.Providers, provider)
	}
	if userId > 0 {
		user, err := impl.userRepository.GetById(userId)
		if err != nil {
			impl.logger.Errorw("found error on getting user data ", "user", user)
		}
		event.Payload.TriggeredBy = user.EmailId
	}

	return event
}
func (impl *EventSimpleFactoryImpl) BuildExtraProtectConfigData(event Event, request ConfigDataForNotification) Event {
	defaultSesConfig, defaultSmtpConfig, err := impl.getDefaultSESOrSMTPConfig()
	if err != nil {
		impl.logger.Errorw("found error in getting defaultSesConfig or  defaultSmtpConfig data", "err", err)
	}
	payload := &Payload{}
	setProviderForNotification(request.EmailIds, defaultSesConfig, defaultSmtpConfig, payload)
	err = impl.setEventPayload(request, payload)
	if err != nil {
		impl.logger.Errorw("error in setting payload", "error", err)
		return event
	}
	if request.UserId == 0 {
		return event
	}
	user, err := impl.userRepository.GetById(request.UserId)
	if err != nil {
		impl.logger.Errorw("found error on getting user data ", "user", user)
	}
	payload.TriggeredBy = user.EmailId
	event.Payload = payload
	return event
}
func setProviderForNotification(emailIds []string, defaultSesConfig *repository2.SESConfig, defaultSmtpConfig *repository2.SMTPConfig, payload *Payload) {
	for _, emailId := range emailIds {
		provider := &notifier.Provider{
			ConfigId:  0,
			Recipient: emailId,
		}
		if defaultSesConfig != nil && defaultSesConfig.Id != 0 {
			provider.Destination = notifier.SES_CONFIG_TYPE
		} else if defaultSmtpConfig != nil && defaultSmtpConfig.Id != 0 {
			provider.Destination = notifier.SMTP_CONFIG_TYPE
		}
		payload.Providers = append(payload.Providers, provider)
	}
}

func (impl *EventSimpleFactoryImpl) setEventPayload(request ConfigDataForNotification, payload *Payload) error {

	protectConfigLink := setProtectConfigLink(request)
	payload.ProtectConfigLink = protectConfigLink
	application, err := impl.appRepo.FindById(request.AppId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching application", "err", err)
		return err
	}
	environment := &repository4.Environment{}
	if request.EnvId != -1 {
		environment, err = impl.envRepository.FindById(request.EnvId)
		if err != nil {
			impl.logger.Errorw("error occurred while fetching environment", "err", err)
			return err
		}
	}
	payload.AppName = application.AppName
	payload.EnvName = environment.Name
	payload.ProtectConfigFileName = request.ResourceName
	payload.ProtectConfigComment = request.UserComment
	payload.ProtectConfigFileType = string(request.Resource)
	return nil
}
func setProtectConfigLink(request ConfigDataForNotification) string {
	var ProtectConfigLink string
	var isAppLevel bool
	if request.EnvId == -1 {
		isAppLevel = true
	}
	if isAppLevel {
		ProtectConfigLink = getAppLevelUrl(request)
	} else {
		ProtectConfigLink = getEnvLevelUrl(request)
	}
	return ProtectConfigLink
}

func getEnvLevelUrl(request ConfigDataForNotification) (ProtectConfigLink string) {
	switch request.Resource {
	case CM:
		ProtectConfigLink = fmt.Sprintf(EnvLevelBaseUrl+"configmap/%s", request.AppId, request.EnvId, request.ResourceName)
	case CS:
		ProtectConfigLink = fmt.Sprintf(EnvLevelBaseUrl+"secrets/%s", request.AppId, request.EnvId, request.ResourceName)
	case DeploymentTemplate:
		ProtectConfigLink = fmt.Sprintf(EnvLevelBaseUrl+"deployment-template", request.AppId, request.EnvId)

	}
	return ProtectConfigLink
}

func getAppLevelUrl(request ConfigDataForNotification) (ProtectConfigLink string) {
	switch request.Resource {
	case CM:
		ProtectConfigLink = fmt.Sprintf(AppLevelBaseUrl+"configmap/%s", request.AppId, request.ResourceName)
	case CS:
		ProtectConfigLink = fmt.Sprintf(AppLevelBaseUrl+"secrets/%s", request.AppId, request.ResourceName)
	case DeploymentTemplate:
		ProtectConfigLink = fmt.Sprintf(AppLevelBaseUrl+"deployment-template", request.AppId)
	}
	return ProtectConfigLink
}

func (impl *EventSimpleFactoryImpl) getDefaultSESOrSMTPConfig() (*repository2.SESConfig, *repository2.SMTPConfig, error) {
	defaultSesConfig, err := impl.sesNotificationRepository.FindDefault()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error fetching defaultSesConfig", "defaultSesConfig", defaultSesConfig, "err", err)
		return defaultSesConfig, nil, nil

	}
	defaultSmtpConfig := &repository2.SMTPConfig{}
	if err == pg.ErrNoRows {
		defaultSmtpConfig, err = impl.smtpNotificationRepository.FindDefault()
		if err != nil {
			impl.logger.Errorw("error fetching defaultSmtpConfig", "defaultSmtpConfig", defaultSmtpConfig, "err", err)
			return defaultSesConfig, defaultSmtpConfig, nil
		}
	}
	return defaultSesConfig, defaultSmtpConfig, nil
}
