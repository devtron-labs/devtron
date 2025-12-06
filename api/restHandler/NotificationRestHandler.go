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

package restHandler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider"
	"github.com/devtron-labs/devtron/pkg/cluster/environment"
	"github.com/devtron-labs/devtron/pkg/notifier"
	"github.com/devtron-labs/devtron/pkg/notifier/beans"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team/read"
	util "github.com/devtron-labs/devtron/util/event"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/devtron-labs/devtron/util/response"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	SLACK_CONFIG_DELETE_SUCCESS_RESP   = "Slack config deleted successfully."
	WEBHOOK_CONFIG_DELETE_SUCCESS_RESP = "Webhook config deleted successfully."
	SES_CONFIG_DELETE_SUCCESS_RESP     = "SES config deleted successfully."
	SMTP_CONFIG_DELETE_SUCCESS_RESP    = "SMTP config deleted successfully."
)

type NotificationRestHandler interface {
	SaveNotificationSettingsV2(w http.ResponseWriter, r *http.Request)
	UpdateNotificationSettings(w http.ResponseWriter, r *http.Request)
	SaveNotificationChannelConfig(w http.ResponseWriter, r *http.Request)
	FindSESConfig(w http.ResponseWriter, r *http.Request)
	FindSlackConfig(w http.ResponseWriter, r *http.Request)
	FindSMTPConfig(w http.ResponseWriter, r *http.Request)
	FindWebhookConfig(w http.ResponseWriter, r *http.Request)
	GetWebhookVariables(w http.ResponseWriter, r *http.Request)
	FindAllNotificationConfig(w http.ResponseWriter, r *http.Request)
	GetAllNotificationSettings(w http.ResponseWriter, r *http.Request)
	DeleteNotificationSettings(w http.ResponseWriter, r *http.Request)
	DeleteNotificationChannelConfig(w http.ResponseWriter, r *http.Request)

	RecipientListingSuggestion(w http.ResponseWriter, r *http.Request)
	FindAllNotificationConfigAutocomplete(w http.ResponseWriter, r *http.Request)
	GetOptionsForNotificationSettings(w http.ResponseWriter, r *http.Request)
}
type NotificationRestHandlerImpl struct {
	dockerRegistryConfig pipeline.DockerRegistryConfig
	logger               *zap.SugaredLogger
	gitRegistryConfig    gitProvider.GitRegistryConfig
	userAuthService      user.UserService
	validator            *validator.Validate
	notificationService  notifier.NotificationConfigService
	slackService         notifier.SlackNotificationService
	webhookService       notifier.WebhookNotificationService
	sesService           notifier.SESNotificationService
	smtpService          notifier.SMTPNotificationService
	enforcer             casbin.Enforcer
	environmentService   environment.EnvironmentService
	pipelineBuilder      pipeline.PipelineBuilder
	enforcerUtil         rbac.EnforcerUtil
	teamReadService      read.TeamReadService
}

type ChannelDto struct {
	Channel util.Channel `json:"channel" validate:"required"`
}

func NewNotificationRestHandlerImpl(dockerRegistryConfig pipeline.DockerRegistryConfig,
	logger *zap.SugaredLogger, gitRegistryConfig gitProvider.GitRegistryConfig,
	userAuthService user.UserService,
	validator *validator.Validate, notificationService notifier.NotificationConfigService,
	slackService notifier.SlackNotificationService, webhookService notifier.WebhookNotificationService, sesService notifier.SESNotificationService, smtpService notifier.SMTPNotificationService,
	enforcer casbin.Enforcer, environmentService environment.EnvironmentService, pipelineBuilder pipeline.PipelineBuilder,
	enforcerUtil rbac.EnforcerUtil,
	teamReadService read.TeamReadService) *NotificationRestHandlerImpl {
	return &NotificationRestHandlerImpl{
		dockerRegistryConfig: dockerRegistryConfig,
		logger:               logger,
		gitRegistryConfig:    gitRegistryConfig,
		userAuthService:      userAuthService,
		validator:            validator,
		notificationService:  notificationService,
		slackService:         slackService,
		webhookService:       webhookService,
		sesService:           sesService,
		smtpService:          smtpService,
		enforcer:             enforcer,
		environmentService:   environmentService,
		pipelineBuilder:      pipelineBuilder,
		enforcerUtil:         enforcerUtil,
		teamReadService:      teamReadService,
	}
}

func (impl NotificationRestHandlerImpl) SaveNotificationSettingsV2(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}
	var notificationSetting beans.NotificationRequest
	err = json.NewDecoder(r.Body).Decode(&notificationSetting)
	if err != nil {
		impl.logger.Errorw("request err, SaveNotificationSettings", "err", err, "payload", notificationSetting)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, SaveNotificationSettings", "err", err, "payload", notificationSetting)
	err = impl.validator.Struct(notificationSetting)
	if err != nil {
		impl.logger.Errorw("validation err, SaveNotificationSettings", "err", err, "payload", notificationSetting)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//RBAC
	token := r.Header.Get("token")
	if isSuperAdmin := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !isSuperAdmin {
		common.WriteJsonResp(w, err, nil, http.StatusForbidden)
		return
	}
	//RBAC

	res, err := impl.notificationService.CreateOrUpdateNotificationSettings(&notificationSetting, userId)
	if err != nil {
		impl.logger.Errorw("service err, SaveNotificationSettings", "err", err, "payload", notificationSetting)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (impl NotificationRestHandlerImpl) UpdateNotificationSettings(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}
	var notificationSetting beans.NotificationUpdateRequest
	err = json.NewDecoder(r.Body).Decode(&notificationSetting)
	if err != nil {
		impl.logger.Errorw("request err, UpdateNotificationSettings", "err", err, "payload", notificationSetting)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, UpdateNotificationSettings", "err", err, "payload", notificationSetting)
	err = impl.validator.Struct(notificationSetting)
	if err != nil {
		impl.logger.Errorw("validation err, UpdateNotificationSettings", "err", err, "payload", notificationSetting)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//RBAC
	token := r.Header.Get("token")
	if isSuperAdmin := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !isSuperAdmin {
		common.WriteJsonResp(w, err, nil, http.StatusForbidden)
		return
	}
	//RBAC

	res, err := impl.notificationService.UpdateNotificationSettings(&notificationSetting, userId)
	if err != nil {
		impl.logger.Errorw("service err, UpdateNotificationSettings", "err", err, "payload", notificationSetting)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (impl NotificationRestHandlerImpl) DeleteNotificationSettings(w http.ResponseWriter, r *http.Request) {
	var request beans.NSDeleteRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		impl.logger.Errorw("request err, DeleteNotificationSettings", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, DeleteNotificationSettings", "err", err, "payload", request)
	//RBAC
	token := r.Header.Get("token")
	if isSuperAdmin := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !isSuperAdmin {
		common.WriteJsonResp(w, err, nil, http.StatusForbidden)
		return
	}
	//RBAC
	err = impl.notificationService.DeleteNotificationSettings(request)
	if err != nil {
		impl.logger.Errorw("service err, DeleteNotificationSettings", "err", err, "payload", request)
	}
	common.WriteJsonResp(w, err, nil, http.StatusOK)
}

func (impl NotificationRestHandlerImpl) GetAllNotificationSettings(w http.ResponseWriter, r *http.Request) {
	// Use enhanced parameter parsing with context
	size, err := common.ExtractIntPathParamWithContext(w, r, "size")
	if err != nil {
		// Error already written by ExtractIntPathParamWithContext
		return
	}

	offset, err := common.ExtractIntPathParamWithContext(w, r, "offset")
	if err != nil {
		// Error already written by ExtractIntPathParamWithContext
		return
	}

	token := r.Header.Get("token")
	if isSuperAdmin := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !isSuperAdmin {
		common.WriteJsonResp(w, err, nil, http.StatusForbidden)
		return
	}
	notificationSettingsViews, totalCount, err := impl.notificationService.FindAll(offset, size)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("service err, GetAllNotificationSettings", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	results, deletedItemCount, err := impl.notificationService.BuildNotificationSettingsResponse(notificationSettingsViews)
	if err != nil {
		impl.logger.Errorw("service err, GetAllNotificationSettings", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	totalCount = totalCount - deletedItemCount
	if results == nil {
		results = make([]*beans.NotificationSettingsResponse, 0)
	}
	nsvResponse := beans.NSViewResponse{
		Total:                        totalCount,
		NotificationSettingsResponse: results,
	}

	common.WriteJsonResp(w, err, nsvResponse, http.StatusOK)
}

func (impl NotificationRestHandlerImpl) SaveNotificationChannelConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	var channelReq ChannelDto
	err = json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(data))).Decode(&channelReq)
	if err != nil {
		impl.logger.Errorw("request err, SaveNotificationChannelConfig", "err", err, "payload", channelReq)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, SaveNotificationChannelConfig", "err", err, "payload", channelReq)
	token := r.Header.Get("token")
	if util.Slack == channelReq.Channel {
		var slackReq *beans.SlackChannelConfig
		err = json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(data))).Decode(&slackReq)
		if err != nil {
			impl.logger.Errorw("request err, SaveNotificationChannelConfig", "err", err, "slackReq", slackReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		err = impl.validator.Struct(slackReq)
		if err != nil {
			impl.logger.Errorw("validation err, SaveNotificationChannelConfig", "err", err, "slackReq", slackReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		//RBAC
		var teamIds []*int
		for _, item := range slackReq.SlackConfigDtos {
			teamIds = append(teamIds, &item.TeamId)
		}
		teams, err := impl.teamReadService.FindByIds(teamIds)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		for _, item := range teams {
			if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, fmt.Sprintf("%s/*", item.Name)); !ok {
				common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
				return
			}
		}
		//RBAC

		res, cErr := impl.slackService.SaveOrEditNotificationConfig(slackReq.SlackConfigDtos, userId)
		if cErr != nil {
			impl.logger.Errorw("service err, SaveNotificationChannelConfig", "err", err, "slackReq", slackReq)
			common.WriteJsonResp(w, cErr, nil, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		common.WriteJsonResp(w, nil, res, http.StatusOK)
	} else if util.SES == channelReq.Channel {
		var sesReq *beans.SESChannelConfig
		err = json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(data))).Decode(&sesReq)
		if err != nil {
			impl.logger.Errorw("request err, SaveNotificationChannelConfig", "err", err, "sesReq", sesReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		err = impl.validator.Struct(sesReq)
		if err != nil {
			impl.logger.Errorw("validation err, SaveNotificationChannelConfig", "err", err, "sesReq", sesReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		// RBAC enforcer applying
		token := r.Header.Get("token")
		if ok := impl.enforcer.Enforce(token, casbin.ResourceNotification, casbin.ActionCreate, "*"); !ok {
			response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
			return
		}
		//RBAC enforcer Ends

		res, cErr := impl.sesService.SaveOrEditNotificationConfig(sesReq.SESConfigDtos, userId)
		if cErr != nil {
			impl.logger.Errorw("service err, SaveNotificationChannelConfig", "err", err, "sesReq", sesReq)
			common.WriteJsonResp(w, cErr, nil, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		common.WriteJsonResp(w, nil, res, http.StatusOK)
	} else if util.SMTP == channelReq.Channel {
		var smtpReq *beans.SMTPChannelConfig
		err = json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(data))).Decode(&smtpReq)
		if err != nil {
			impl.logger.Errorw("request err, SaveNotificationChannelConfig", "err", err, "smtpReq", smtpReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		err = impl.validator.Struct(smtpReq)
		if err != nil {
			impl.logger.Errorw("validation err, SaveNotificationChannelConfig", "err", err, "smtpReq", smtpReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		// RBAC enforcer applying
		token := r.Header.Get("token")
		if ok := impl.enforcer.Enforce(token, casbin.ResourceNotification, casbin.ActionCreate, "*"); !ok {
			response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
			return
		}
		//RBAC enforcer Ends

		res, cErr := impl.smtpService.SaveOrEditNotificationConfig(smtpReq.SMTPConfigDtos, userId)
		if cErr != nil {
			impl.logger.Errorw("service err, SaveNotificationChannelConfig", "err", err, "smtpReq", smtpReq)
			common.WriteJsonResp(w, cErr, nil, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		common.WriteJsonResp(w, nil, res, http.StatusOK)
	} else if util.Webhook == channelReq.Channel {
		var webhookReq *beans.WebhookChannelConfig
		err = json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(data))).Decode(&webhookReq)
		if err != nil {
			impl.logger.Errorw("request err, SaveNotificationChannelConfig", "err", err, "webhookReq", webhookReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		err = impl.validator.Struct(webhookReq)
		if err != nil {
			impl.logger.Errorw("validation err, SaveNotificationChannelConfig", "err", err, "webhookReq", webhookReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		// RBAC enforcer applying
		token := r.Header.Get("token")
		if ok := impl.enforcer.Enforce(token, casbin.ResourceNotification, casbin.ActionCreate, "*"); !ok {
			response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
			return
		}
		//RBAC enforcer Ends

		res, cErr := impl.webhookService.SaveOrEditNotificationConfig(*webhookReq.WebhookConfigDtos, userId)
		if cErr != nil {
			impl.logger.Errorw("service err, SaveNotificationChannelConfig", "err", err, "webhookReq", webhookReq)
			common.WriteJsonResp(w, cErr, nil, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		common.WriteJsonResp(w, nil, res, http.StatusOK)
	}
}

type ChannelResponseDTO struct {
	SlackConfigs   []*beans.SlackConfigDto   `json:"slackConfigs"`
	WebhookConfigs []*beans.WebhookConfigDto `json:"webhookConfigs"`
	SESConfigs     []*beans.SESConfigDto     `json:"sesConfigs"`
	SMTPConfigs    []*beans.SMTPConfigDto    `json:"smtpConfigs"`
}

func (impl NotificationRestHandlerImpl) FindAllNotificationConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	token := r.Header.Get("token")
	channelsResponse := &ChannelResponseDTO{}
	slackConfigs, fErr := impl.slackService.FetchAllSlackNotificationConfig()
	if fErr != nil && fErr != pg.ErrNoRows {
		impl.logger.Errorw("service err, FindAllNotificationConfig", "err", err)
		common.WriteJsonResp(w, fErr, nil, http.StatusInternalServerError)
		return
	}

	if ok := impl.enforcer.Enforce(token, casbin.ResourceNotification, casbin.ActionGet, "*"); !ok {
		// if user does not have notification level access then return unauthorized
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//RBAC
	pass := true
	if len(slackConfigs) > 0 {
		var teamIds []*int
		for _, item := range slackConfigs {
			teamIds = append(teamIds, &item.TeamId)
		}
		teams, err := impl.teamReadService.FindByIds(teamIds)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		for _, item := range teams {
			if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, fmt.Sprintf("%s/*", item.Name)); !ok {
				pass = false
				break
			}
		}
	}
	//RBAC
	if slackConfigs == nil {
		slackConfigs = make([]*beans.SlackConfigDto, 0)
	}
	if pass {
		channelsResponse.SlackConfigs = slackConfigs
	}
	webhookConfigs, fErr := impl.webhookService.FetchAllWebhookNotificationConfig()
	if fErr != nil && fErr != pg.ErrNoRows {
		impl.logger.Errorw("service err, FindAllNotificationConfig", "err", err)
		common.WriteJsonResp(w, fErr, nil, http.StatusInternalServerError)
		return
	}
	if webhookConfigs == nil {
		webhookConfigs = make([]*beans.WebhookConfigDto, 0)
	}
	if pass {
		channelsResponse.WebhookConfigs = webhookConfigs
	}
	sesConfigs, fErr := impl.sesService.FetchAllSESNotificationConfig()
	if fErr != nil && fErr != pg.ErrNoRows {
		impl.logger.Errorw("service err, FindAllNotificationConfig", "err", err)
		common.WriteJsonResp(w, fErr, nil, http.StatusInternalServerError)
		return
	}
	if sesConfigs == nil {
		sesConfigs = make([]*beans.SESConfigDto, 0)
	}
	if pass {
		channelsResponse.SESConfigs = sesConfigs
	}

	smtpConfigs, err := impl.smtpService.FetchAllSMTPNotificationConfig()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("service err, FindAllNotificationConfig", "err", err)
		common.WriteJsonResp(w, fErr, nil, http.StatusInternalServerError)
		return
	}
	if smtpConfigs == nil {
		smtpConfigs = make([]*beans.SMTPConfigDto, 0)
	}
	if pass {
		channelsResponse.SMTPConfigs = smtpConfigs
	}
	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, fErr, channelsResponse, http.StatusOK)
}

func (impl NotificationRestHandlerImpl) FindSESConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	// Use enhanced parameter parsing with context
	id, err := common.ExtractIntPathParamWithContext(w, r, "id")
	if err != nil {
		// Error already written by ExtractIntPathParamWithContext
		return
	}
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceNotification, casbin.ActionGet, "*"); !ok {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
	}

	sesConfig, fErr := impl.sesService.FetchSESNotificationConfigById(id)
	if fErr != nil && fErr != pg.ErrNoRows {
		impl.logger.Errorw("service err, FindSESConfig", "err", err)
		common.WriteJsonResp(w, fErr, nil, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, fErr, sesConfig, http.StatusOK)
}

func (impl NotificationRestHandlerImpl) FindSlackConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	// Use enhanced parameter parsing with context
	id, err := common.ExtractIntPathParamWithContext(w, r, "id")
	if err != nil {
		// Error already written by ExtractIntPathParamWithContext
		return
	}

	sesConfig, fErr := impl.slackService.FetchSlackNotificationConfigById(id)
	if fErr != nil && fErr != pg.ErrNoRows {
		impl.logger.Errorw("service err, FindSlackConfig, cannot find slack config", "err", fErr, "id", id)
		common.WriteJsonResp(w, fErr, nil, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, fErr, sesConfig, http.StatusOK)
}

func (impl NotificationRestHandlerImpl) FindSMTPConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		impl.logger.Errorw("request err, FindSMTPConfig", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceNotification, casbin.ActionGet, "*"); !ok {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
	}

	smtpConfig, fErr := impl.smtpService.FetchSMTPNotificationConfigById(id)
	if fErr != nil && fErr != pg.ErrNoRows {
		impl.logger.Errorw("service err, FindSMTPConfig", "err", err)
		common.WriteJsonResp(w, fErr, nil, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, fErr, smtpConfig, http.StatusOK)
}
func (impl NotificationRestHandlerImpl) FindWebhookConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		impl.logger.Errorw("request err, FindWebhookConfig", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	webhookConfig, fErr := impl.webhookService.FetchWebhookNotificationConfigById(id)
	if fErr != nil && fErr != pg.ErrNoRows {
		impl.logger.Errorw("service err, FindWebhookConfig, cannot find webhook config", "err", fErr, "id", id)
		common.WriteJsonResp(w, fErr, nil, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, fErr, webhookConfig, http.StatusOK)
}
func (impl NotificationRestHandlerImpl) GetWebhookVariables(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	webhookVariables, fErr := impl.webhookService.GetWebhookVariables()
	if fErr != nil && fErr != pg.ErrNoRows {
		impl.logger.Errorw("service err, GetWebhookVariables, cannot find webhook Variables", "err", fErr)
		common.WriteJsonResp(w, fErr, nil, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, fErr, webhookVariables, http.StatusOK)
}

func (impl NotificationRestHandlerImpl) RecipientListingSuggestion(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceNotification, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), "Forbidden", http.StatusForbidden)
		return
	}
	vars := mux.Vars(r)
	value := vars["value"]
	//var teams []int
	var channelsResponse []*beans.NotificationRecipientListingResponse
	channelsResponse, err = impl.slackService.RecipientListingSuggestion(value)
	if err != nil {
		impl.logger.Errorw("service err, RecipientListingSuggestion", "err", err, "value", value)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	if channelsResponse == nil {
		channelsResponse = make([]*beans.NotificationRecipientListingResponse, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, nil, channelsResponse, http.StatusOK)
}

func (impl NotificationRestHandlerImpl) FindAllNotificationConfigAutocomplete(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceNotification, casbin.ActionGet, "*"); !ok {
		response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
		return
	}
	//RBAC enforcer Ends
	vars := mux.Vars(r)
	cType := vars["type"]
	var channelsResponse []*beans.NotificationChannelAutoResponse
	if cType == string(util.Slack) {
		channelsResponseAll, err := impl.slackService.FetchAllSlackNotificationConfigAutocomplete()
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("service err, FindAllNotificationConfigAutocomplete", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		for _, item := range channelsResponseAll {
			team, err := impl.teamReadService.FindOne(item.TeamId)
			if err != nil {
				common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
				return
			}
			if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, fmt.Sprintf("%s/*", team.Name)); ok {
				channelsResponse = append(channelsResponse, item)
			}
		}

	} else if cType == string(util.Webhook) {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceNotification, casbin.ActionGet, "*"); !ok {
			response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
			return
		}
		channelsResponse, err = impl.webhookService.FetchAllWebhookNotificationConfigAutocomplete()
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("service err, FindAllNotificationConfigAutocomplete", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	} else if cType == string(util.SES) {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceNotification, casbin.ActionGet, "*"); !ok {
			response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
			return
		}
		channelsResponse, err = impl.sesService.FetchAllSESNotificationConfigAutocomplete()
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("service err, FindAllNotificationConfigAutocomplete", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	} else if cType == string(util.SMTP) {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceNotification, casbin.ActionGet, "*"); !ok {
			response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
			return
		}
		channelsResponse, err = impl.smtpService.FetchAllSMTPNotificationConfigAutocomplete()
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("service err, FindAllNotificationConfigAutocomplete", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}
	if channelsResponse == nil {
		channelsResponse = make([]*beans.NotificationChannelAutoResponse, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, nil, channelsResponse, http.StatusOK)
}

func (impl NotificationRestHandlerImpl) GetOptionsForNotificationSettings(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}
	var request repository.SearchRequest
	err = decoder.Decode(&request)
	if err != nil {
		impl.logger.Errorw("request err, GetOptionsForNotificationSettings", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId

	token := r.Header.Get("token")
	if isSuperAdmin := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !isSuperAdmin {
		common.WriteJsonResp(w, err, nil, http.StatusForbidden)
		return
	}

	notificationSettingsOptions, err := impl.notificationService.FindNotificationSettingOptions(&request)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("service err, GetOptionsForNotificationSettings", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	if notificationSettingsOptions == nil {
		notificationSettingsOptions = make([]*beans.SearchFilterResponse, 0)
	}
	common.WriteJsonResp(w, err, notificationSettingsOptions, http.StatusOK)
}

func (impl NotificationRestHandlerImpl) DeleteNotificationChannelConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	var channelReq ChannelDto
	err = json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(data))).Decode(&channelReq)
	if err != nil {
		impl.logger.Errorw("request err, DeleteNotificationChannelConfig", "err", err, "payload", channelReq)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, DeleteNotificationChannelConfig", "err", err, "payload", channelReq)
	if util.Slack == channelReq.Channel {
		var deleteReq *beans.SlackConfigDto
		err = json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(data))).Decode(&deleteReq)
		if err != nil {
			impl.logger.Errorw("request err, DeleteNotificationChannelConfig", "err", err, "deleteReq", deleteReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		err = impl.validator.Struct(deleteReq)
		if err != nil {
			impl.logger.Errorw("validation err, DeleteNotificationChannelConfig", "err", err, "deleteReq", deleteReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		// RBAC enforcer applying
		token := r.Header.Get("token")
		if ok := impl.enforcer.Enforce(token, casbin.ResourceNotification, casbin.ActionCreate, "*"); !ok {
			response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
			return
		}
		//RBAC enforcer Ends

		cErr := impl.slackService.DeleteNotificationConfig(deleteReq, userId)
		if cErr != nil {
			impl.logger.Errorw("service err, DeleteNotificationChannelConfig", "err", err, "deleteReq", deleteReq)
			common.WriteJsonResp(w, cErr, nil, http.StatusInternalServerError)
			return
		}
		common.WriteJsonResp(w, nil, SLACK_CONFIG_DELETE_SUCCESS_RESP, http.StatusOK)
	} else if util.Webhook == channelReq.Channel {
		var deleteReq *beans.WebhookConfigDto
		err = json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(data))).Decode(&deleteReq)
		if err != nil {
			impl.logger.Errorw("request err, DeleteNotificationChannelConfig", "err", err, "deleteReq", deleteReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		err = impl.validator.Struct(deleteReq)
		if err != nil {
			impl.logger.Errorw("validation err, DeleteNotificationChannelConfig", "err", err, "deleteReq", deleteReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		// RBAC enforcer applying
		token := r.Header.Get("token")
		if ok := impl.enforcer.Enforce(token, casbin.ResourceNotification, casbin.ActionCreate, "*"); !ok {
			response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
			return
		}
		//RBAC enforcer Ends

		cErr := impl.webhookService.DeleteNotificationConfig(deleteReq, userId)
		if cErr != nil {
			impl.logger.Errorw("service err, DeleteNotificationChannelConfig", "err", err, "deleteReq", deleteReq)
			common.WriteJsonResp(w, cErr, nil, http.StatusInternalServerError)
			return
		}
		common.WriteJsonResp(w, nil, WEBHOOK_CONFIG_DELETE_SUCCESS_RESP, http.StatusOK)
	} else if util.SES == channelReq.Channel {
		var deleteReq *beans.SESConfigDto
		err = json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(data))).Decode(&deleteReq)
		if err != nil {
			impl.logger.Errorw("request err, DeleteNotificationChannelConfig", "err", err, "deleteReq", deleteReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		err = impl.validator.Struct(deleteReq)
		if err != nil {
			impl.logger.Errorw("validation err, DeleteNotificationChannelConfig", "err", err, "deleteReq", deleteReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		// RBAC enforcer applying
		token := r.Header.Get("token")
		if ok := impl.enforcer.Enforce(token, casbin.ResourceNotification, casbin.ActionCreate, "*"); !ok {
			response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
			return
		}
		//RBAC enforcer Ends

		cErr := impl.sesService.DeleteNotificationConfig(deleteReq, userId)
		if cErr != nil {
			impl.logger.Errorw("service err, DeleteNotificationChannelConfig", "err", err, "deleteReq", deleteReq)
			common.WriteJsonResp(w, cErr, nil, http.StatusInternalServerError)
			return
		}
		common.WriteJsonResp(w, nil, SES_CONFIG_DELETE_SUCCESS_RESP, http.StatusOK)
	} else if util.SMTP == channelReq.Channel {
		var deleteReq *beans.SMTPConfigDto
		err = json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(data))).Decode(&deleteReq)
		if err != nil {
			impl.logger.Errorw("request err, DeleteNotificationChannelConfig", "err", err, "deleteReq", deleteReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		err = impl.validator.Struct(deleteReq)
		if err != nil {
			impl.logger.Errorw("validation err, DeleteNotificationChannelConfig", "err", err, "deleteReq", deleteReq)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		// RBAC enforcer applying
		token := r.Header.Get("token")
		if ok := impl.enforcer.Enforce(token, casbin.ResourceNotification, casbin.ActionCreate, "*"); !ok {
			response.WriteResponse(http.StatusForbidden, "FORBIDDEN", w, errors.New("unauthorized"))
			return
		}
		//RBAC enforcer Ends

		cErr := impl.smtpService.DeleteNotificationConfig(deleteReq, userId)
		if cErr != nil {
			impl.logger.Errorw("service err, DeleteNotificationChannelConfig", "err", err, "deleteReq", deleteReq)
			common.WriteJsonResp(w, cErr, nil, http.StatusInternalServerError)
			return
		}
		common.WriteJsonResp(w, nil, SMTP_CONFIG_DELETE_SUCCESS_RESP, http.StatusOK)
	} else {
		common.WriteJsonResp(w, fmt.Errorf(" The channel you requested is not supported"), nil, http.StatusBadRequest)
	}
}
