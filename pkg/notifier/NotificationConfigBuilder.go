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

package notifier

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
	"strings"
	"time"
)

type NotificationConfigBuilder interface {
	BuildNotificationSettingsConfig(notificationSettingsRequest *NotificationConfigRequest, existingNotificationSettingsConfig *repository.NotificationSettingsView, userId int32) (*repository.NotificationSettingsView, error)
	BuildNewNotificationSettings(notificationSettingsRequest *NotificationConfigRequest, notificationSettingsView *repository.NotificationSettingsView) ([]repository.NotificationSettings, error)
	BuildNotificationSettingWithPipeline(teamId *int, envId *int, appId *int, pipelineId *int, pipelineType util.PipelineType, eventTypeId int, viewId int, providers []*Provider) (repository.NotificationSettings, error)
}

type NotificationConfigBuilderImpl struct {
	logger *zap.SugaredLogger
}

func NewNotificationConfigBuilderImpl(logger *zap.SugaredLogger) *NotificationConfigBuilderImpl {
	return &NotificationConfigBuilderImpl{
		logger: logger,
	}
}

type NSConfig struct {
	TeamId       []*int            `json:"teamId"`
	AppId        []*int            `json:"appId"`
	EnvId        []*int            `json:"envId"`
	PipelineId   *int              `json:"pipelineId"`
	PipelineType util.PipelineType `json:"pipelineType" validate:"required"`
	EventTypeIds []int             `json:"eventTypeIds" validate:"required"`
	Providers    []*Provider       `json:"providers" validate:"required"`
}

func (impl NotificationConfigBuilderImpl) BuildNotificationSettingsConfig(notificationSettingsRequest *NotificationConfigRequest, existingNotificationSettingsConfig *repository.NotificationSettingsView, userId int32) (*repository.NotificationSettingsView, error) {
	nsConfig := &NSConfig{}
	nsConfig.TeamId = notificationSettingsRequest.TeamId
	nsConfig.AppId = notificationSettingsRequest.AppId
	nsConfig.EnvId = notificationSettingsRequest.EnvId
	nsConfig.PipelineId = notificationSettingsRequest.PipelineId
	nsConfig.PipelineType = notificationSettingsRequest.PipelineType
	nsConfig.EventTypeIds = notificationSettingsRequest.EventTypeIds
	nsConfig.Providers = notificationSettingsRequest.Providers

	config, err := json.Marshal(nsConfig)
	if err != nil {
		impl.logger.Error(err)
		return nil, err
	}
	currentTime := time.Now()
	notificationSettingsView := &repository.NotificationSettingsView{
		Config: string(config),
		//ConfigName:    notificationSettingsRequest.ConfigName,
		//AppId:         notificationSettingsRequest.AppId,
		//EnvironmentId: notificationSettingsRequest.EnvId,
	}
	if notificationSettingsRequest.Id != 0 {
		notificationSettingsView.Id = notificationSettingsRequest.Id
		notificationSettingsView.UpdatedOn = currentTime
		notificationSettingsView.UpdatedBy = userId
		notificationSettingsView.CreatedBy = existingNotificationSettingsConfig.CreatedBy
		notificationSettingsView.CreatedOn = existingNotificationSettingsConfig.CreatedOn
	} else {
		notificationSettingsView.AuditLog.CreatedOn = currentTime
		notificationSettingsView.AuditLog.CreatedBy = userId
	}
	return notificationSettingsView, nil
}

func (impl NotificationConfigBuilderImpl) BuildNewNotificationSettings(notificationSettingsRequest *NotificationConfigRequest, notificationSettingsView *repository.NotificationSettingsView) ([]repository.NotificationSettings, error) {
	var notificationSettings []repository.NotificationSettings
	type LocalRequest struct {
		Id         int  `json:"id"`
		TeamId     *int `json:"teamId"`
		AppId      *int `json:"appId"`
		EnvId      *int `json:"envId"`
		PipelineId *int `json:"pipelineId"`
	}
	var tempRequest []*LocalRequest
	if len(notificationSettingsRequest.TeamId) == 0 && len(notificationSettingsRequest.EnvId) == 0 && len(notificationSettingsRequest.AppId) > 0 {
		for _, item := range notificationSettingsRequest.AppId {
			tempRequest = append(tempRequest, &LocalRequest{AppId: item})
		}
	} else if len(notificationSettingsRequest.TeamId) == 0 && len(notificationSettingsRequest.EnvId) > 0 && len(notificationSettingsRequest.AppId) == 0 {
		for _, item := range notificationSettingsRequest.EnvId {
			tempRequest = append(tempRequest, &LocalRequest{EnvId: item})
		}
	} else if len(notificationSettingsRequest.TeamId) > 0 && len(notificationSettingsRequest.EnvId) == 0 && len(notificationSettingsRequest.AppId) == 0 {
		for _, item := range notificationSettingsRequest.TeamId {
			tempRequest = append(tempRequest, &LocalRequest{TeamId: item})
		}
	} else if len(notificationSettingsRequest.TeamId) == 0 && len(notificationSettingsRequest.EnvId) > 0 && len(notificationSettingsRequest.AppId) > 0 {
		for _, itemE := range notificationSettingsRequest.EnvId {
			for _, itemA := range notificationSettingsRequest.AppId {
				tempRequest = append(tempRequest, &LocalRequest{EnvId: itemE, AppId: itemA})
			}
		}
	} else if len(notificationSettingsRequest.TeamId) > 0 && len(notificationSettingsRequest.EnvId) > 0 && len(notificationSettingsRequest.AppId) == 0 {
		for _, itemT := range notificationSettingsRequest.TeamId {
			for _, itemE := range notificationSettingsRequest.EnvId {
				tempRequest = append(tempRequest, &LocalRequest{TeamId: itemT, EnvId: itemE})
			}
		}
	} else if len(notificationSettingsRequest.TeamId) > 0 && len(notificationSettingsRequest.EnvId) == 0 && len(notificationSettingsRequest.AppId) > 0 {
		for _, itemT := range notificationSettingsRequest.TeamId {
			for _, itemA := range notificationSettingsRequest.AppId {
				tempRequest = append(tempRequest, &LocalRequest{TeamId: itemT, AppId: itemA})
			}
		}
	} else if len(notificationSettingsRequest.TeamId) > 0 && len(notificationSettingsRequest.EnvId) > 0 && len(notificationSettingsRequest.AppId) > 0 {
		for _, itemT := range notificationSettingsRequest.TeamId {
			for _, itemE := range notificationSettingsRequest.EnvId {
				for _, itemA := range notificationSettingsRequest.AppId {
					tempRequest = append(tempRequest, &LocalRequest{TeamId: itemT, EnvId: itemE, AppId: itemA})
				}
			}
		}
	} else {
		tempRequest = append(tempRequest, &LocalRequest{PipelineId: notificationSettingsRequest.PipelineId})
	}

	for _, item := range tempRequest {
		for _, e := range notificationSettingsRequest.EventTypeIds {
			notificationSetting, err := impl.BuildNotificationSettingWithPipeline(item.TeamId, item.EnvId, item.AppId, item.PipelineId, notificationSettingsRequest.PipelineType, e, notificationSettingsView.Id, notificationSettingsRequest.Providers)
			if err != nil {
				impl.logger.Error(err)
				return nil, err
			}
			notificationSettings = append(notificationSettings, notificationSetting)
		}
	}
	return notificationSettings, nil
}

func (impl NotificationConfigBuilderImpl) buildNotificationSetting(notificationSettingsRequest *NotificationSettingRequest, notificationSettingsView *repository.NotificationSettingsView, eventTypeId int) (repository.NotificationSettings, error) {
	providersJson, err := json.Marshal(notificationSettingsRequest.Providers)
	if err != nil {
		impl.logger.Error(err)
		return repository.NotificationSettings{}, err
	}
	notificationSetting := repository.NotificationSettings{
		AppId:        notificationSettingsRequest.AppId,
		EnvId:        notificationSettingsRequest.EnvId,
		EventTypeId:  eventTypeId,
		PipelineType: string(notificationSettingsRequest.PipelineType),
		Config:       string(providersJson),
		ViewId:       notificationSettingsView.Id,
	}
	return notificationSetting, nil
}

func (impl NotificationConfigBuilderImpl) BuildNotificationSettingWithPipeline(teamId *int, envId *int, appId *int, pipelineId *int, pipelineType util.PipelineType, eventTypeId int, viewId int, providers []*Provider) (repository.NotificationSettings, error) {

	for _, provider := range providers {
		if len(provider.Recipient) > 0 {
			if strings.Contains(provider.Recipient, "@") {
				provider.Destination = util.SES
			} else {
				provider.Destination = util.Slack
			}
		}
	}

	providersJson, err := json.Marshal(providers)
	if err != nil {
		impl.logger.Error(err)
		return repository.NotificationSettings{}, err
	}
	notificationSetting := repository.NotificationSettings{
		TeamId:       teamId,
		AppId:        appId,
		EnvId:        envId,
		PipelineId:   pipelineId,
		PipelineType: string(pipelineType),
		EventTypeId:  eventTypeId,
		Config:       string(providersJson),
		ViewId:       viewId,
	}
	return notificationSetting, nil
}
