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
	"strings"
	"time"

	bean2 "github.com/devtron-labs/devtron/api/bean"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/apiToken"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/bean"
	repository4 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
)

type EventFactory interface {
	Build(eventType util.EventType, sourceId *int, appId int, envId *int, pipelineType util.PipelineType) Event
	BuildExtraCDData(event Event, wfr *pipelineConfig.CdWorkflowRunner, pipelineOverrideId int, stage bean2.WorkflowType) Event
	BuildExtraApprovalData(event Event, approvalActionRequest bean.UserApprovalActionRequest, pipeline *pipelineConfig.Pipeline, userId int32, imageTagNames []string, imageComment string) []Event
	BuildExtraProtectConfigData(event Event, draftNotificationRequest ConfigDataForNotification, draftId int, DraftVersionId int) []Event
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
	appRepo                      appRepository.AppRepository
	envRepository                repository4.EnvironmentRepository
	apiTokenServiceImpl          *apiToken.ApiTokenServiceImpl
}

func NewEventSimpleFactoryImpl(logger *zap.SugaredLogger, cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository, ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository, pipelineRepository pipelineConfig.PipelineRepository,
	userRepository repository.UserRepository, ciArtifactRepository repository2.CiArtifactRepository, DeploymentApprovalRepository pipelineConfig.DeploymentApprovalRepository,
	sesNotificationRepository repository2.SESNotificationRepository, smtpNotificationRepository repository2.SMTPNotificationRepository,
	appRepo appRepository.AppRepository, envRepository repository4.EnvironmentRepository, apiTokenServiceImpl *apiToken.ApiTokenServiceImpl,
) *EventSimpleFactoryImpl {
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
		appRepo:                      appRepo,
		envRepository:                envRepository,
		apiTokenServiceImpl:          apiTokenServiceImpl,
	}

}

type ResourceType string

const (
	CM                 ResourceType = "ConfigMap"
	CS                 ResourceType = "Secret"
	DeploymentTemplate ResourceType = "Deployment Template"
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
type Provider struct {
	Destination util.Channel `json:"dest"`
	Rule        string       `json:"rule"`
	ConfigId    int          `json:"configId"`
	Recipient   string       `json:"recipient"`
}

const (
	SES_CONFIG_TYPE  = "ses"
	SMTP_CONFIG_TYPE = "smtp"
)

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
func (impl *EventSimpleFactoryImpl) BuildExtraApprovalData(event Event, approvalActionRequest bean.UserApprovalActionRequest, cdPipeline *pipelineConfig.Pipeline, userId int32, imageTagNames []string, imageComment string) []Event {
	defaultSesConfig, defaultSmtpConfig, err := impl.getDefaultSESOrSMTPConfig()
	if err != nil {
		impl.logger.Errorw("found error in getting defaultSesConfig or  defaultSmtpConfig data", "err", err)
	}
	var events []Event
	if userId == 0 {
		return events
	}
	user, err := impl.userRepository.GetById(userId)
	if err != nil {
		impl.logger.Errorw("found error on getting user data ", "user", user)
	}
	EmailIds := approvalActionRequest.ApprovalNotificationConfig.EmailIds
	for _, emailId := range EmailIds {

		payload, err := impl.setApprovalEventPayload(event, approvalActionRequest, cdPipeline, imageTagNames, imageComment)
		if err != nil {
			impl.logger.Errorw("error in setting payload", "error", err)
			return events
		}
		setProviderForNotification(emailId, defaultSesConfig, defaultSmtpConfig, payload)
		reqData := &ConfigDataForNotification{
			AppId: cdPipeline.AppId,
			EnvId: cdPipeline.EnvironmentId,
		}
		deploymentApprovalRequest := setDeploymentApprovalRequest(reqData, &approvalActionRequest, emailId)
		err = impl.createAndSetToken(nil, deploymentApprovalRequest, payload)
		payload.TriggeredBy = user.EmailId
		event.Payload = payload
		events = append(events, event)
	}

	return events
}
func (impl *EventSimpleFactoryImpl) setApprovalEventPayload(event Event, approvalActionRequest bean.UserApprovalActionRequest, cdPipeline *pipelineConfig.Pipeline, imageTagNames []string, imageComment string) (*Payload, error) {
	payload := &Payload{}
	payload.ImageComment = imageComment
	payload.ImageTagNames = imageTagNames
	ciArtifact, err := impl.ciArtifactRepository.Get(approvalActionRequest.ArtifactId)
	if err != nil {
		impl.logger.Errorw("error fetching ciArtifact", "ciArtifact", ciArtifact, "err", err)
		return payload, err
	}
	payload.AppName = cdPipeline.App.AppName
	payload.EnvName = cdPipeline.Environment.Name
	payload.PipelineName = cdPipeline.Name
	payload.DockerImageUrl = ciArtifact.Image
	dockerImageTag := ""
	split := strings.Split(ciArtifact.Image, ":")
	if len(split) > 1 {
		dockerImageTag = split[len(split)-1]
	}
	payload.ImageApprovalLink = fmt.Sprintf("/dashboard/app/%d/trigger?approval-node=%d&imageTag=%s", event.AppId, cdPipeline.Id, dockerImageTag)
	return payload, err
}

func (impl *EventSimpleFactoryImpl) BuildExtraProtectConfigData(event Event, request ConfigDataForNotification, draftId int, DraftVersionId int) []Event {
	defaultSesConfig, defaultSmtpConfig, err := impl.getDefaultSESOrSMTPConfig()
	if err != nil {
		impl.logger.Errorw("found error in getting defaultSesConfig or  defaultSmtpConfig data", "err", err)
	}
	var events []Event
	if request.UserId == 0 {
		return events
	}
	user, err := impl.userRepository.GetById(request.UserId)
	if err != nil {
		impl.logger.Errorw("found error on getting user data ", "user", user)
	}
	for _, email := range request.EmailIds {
		payload, err := impl.setEventPayload(request)
		if err != nil {
			impl.logger.Errorw("error in setting payload", "error", err)
			return events
		}
		setProviderForNotification(email, defaultSesConfig, defaultSmtpConfig, payload)
		draftRequest := setDraftApprovalRequest(&request, draftId, DraftVersionId, email)
		err = impl.createAndSetToken(draftRequest, nil, payload)
		if err != nil {
			return events
		}
		payload.TriggeredBy = user.EmailId
		event.Payload = payload
		events = append(events, event)

	}

	return events
}
func (impl *EventSimpleFactoryImpl) createAndSetToken(draftRequest *apiToken.DraftApprovalRequest, deploymentApprovalRequest *apiToken.DeploymentApprovalRequest, payload *Payload) error {
	var emailId string
	if deploymentApprovalRequest != nil {
		emailId = deploymentApprovalRequest.EmailId
	} else {
		emailId = draftRequest.EmailId
	}
	user, err := impl.userRepository.FetchActiveUserByEmail(emailId)
	if err != nil {
		impl.logger.Errorw("error in fetching user", "emailId", emailId)
		return err
	}
	if deploymentApprovalRequest != nil {
		deploymentApprovalRequest.UserId = user.Id
		token, err := impl.apiTokenServiceImpl.CreateApiJwtTokenForNotification(deploymentApprovalRequest.GetClaimsForDeploymentApprovalRequest(), impl.apiTokenServiceImpl.TokenVariableConfig.GetExpiryTimeInMs())
		if err != nil {
			impl.logger.Errorw("error in generating token for deployment approval request", "err", err)
			return err
		}
		payload.ApprovalLink = fmt.Sprintf("/dashboard/deployment/approve?token=%s", token)

	} else {
		draftRequest.UserId = user.Id
		token, err := impl.apiTokenServiceImpl.CreateApiJwtTokenForNotification(draftRequest.GetClaimsForDraftApprovalRequest(), impl.apiTokenServiceImpl.TokenVariableConfig.GetExpiryTimeInMs())
		if err != nil {
			impl.logger.Errorw("error in generating token for draft approval request", "err", err)
			return err
		}
		payload.ApprovalLink = fmt.Sprintf("/dashboard/config/approve?token=%s", token)

	}
	return err
}
func setDraftApprovalRequest(request *ConfigDataForNotification, draftId int, DraftVersionId int, emailId string) *apiToken.DraftApprovalRequest {
	draftRequest := &apiToken.DraftApprovalRequest{
		DraftId:        draftId,
		DraftVersionId: DraftVersionId,
		NotificationApprovalRequest: apiToken.NotificationApprovalRequest{
			AppId:   request.AppId,
			EnvId:   request.EnvId,
			EmailId: emailId,
		},
	}
	return draftRequest
}

func setDeploymentApprovalRequest(request *ConfigDataForNotification, approvalActionRequest *bean.UserApprovalActionRequest, emailId string) *apiToken.DeploymentApprovalRequest {
	deploymentApprovalRequest := &apiToken.DeploymentApprovalRequest{
		ApprovalRequestId: approvalActionRequest.ApprovalRequestId,
		ArtifactId:        approvalActionRequest.ArtifactId,
		PipelineId:        approvalActionRequest.PipelineId,
		NotificationApprovalRequest: apiToken.NotificationApprovalRequest{
			AppId:   request.AppId,
			EnvId:   request.EnvId,
			EmailId: emailId,
		},
	}
	return deploymentApprovalRequest
}
func setProviderForNotification(emailId string, defaultSesConfig *repository2.SESConfig, defaultSmtpConfig *repository2.SMTPConfig, payload *Payload) {
	provider := &Provider{
		ConfigId:  0,
		Recipient: emailId,
	}
	if defaultSesConfig != nil && defaultSesConfig.Id != 0 {
		provider.Destination = SES_CONFIG_TYPE
	} else if defaultSmtpConfig != nil && defaultSmtpConfig.Id != 0 {
		provider.Destination = SMTP_CONFIG_TYPE
	}
	payload.Providers = append(payload.Providers, provider)

}

func (impl *EventSimpleFactoryImpl) setEventPayload(request ConfigDataForNotification) (*Payload, error) {
	payload := &Payload{}
	protectConfigLink := setProtectConfigLink(request)
	payload.ProtectConfigLink = protectConfigLink
	application, err := impl.appRepo.FindById(request.AppId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching application", "err", err)
		return payload, err
	}
	environment := &repository4.Environment{}
	if request.EnvId != -1 {
		environment, err = impl.envRepository.FindById(request.EnvId)
		if err != nil {
			impl.logger.Errorw("error occurred while fetching environment", "err", err)
			return payload, err
		}
	}
	payload.AppName = application.AppName
	payload.EnvName = environment.Name
	payload.ProtectConfigComment = request.UserComment
	payload.ProtectConfigFileType = string(request.Resource)
	if request.Resource == DeploymentTemplate {
		payload.ProtectConfigFileName = string(DeploymentTemplate)
	} else {
		payload.ProtectConfigFileName = request.ResourceName
	}

	return payload, err
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
