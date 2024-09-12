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

package notifier

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/notifier/beans"
	"github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
	"time"
)

type NotificationConfigBuilder interface {
	BuildNotificationSettingsConfig(notificationSettingsRequest *beans.NotificationConfigRequest, existingNotificationSettingsConfig *repository.NotificationSettingsView, userId int32) (*repository.NotificationSettingsView, error)
	BuildNewNotificationSettings(notificationSettingsRequest *beans.NotificationConfigRequest, notificationSettingsView *repository.NotificationSettingsView) ([]repository.NotificationSettings, error)
	BuildNotificationSettingWithPipeline(teamId *int, envId *int, appId *int, pipelineId *int, clusterId *int, pipelineType util.PipelineType, eventTypeId int, viewId int, providers []*beans.Provider) (repository.NotificationSettings, error)
}

type NotificationConfigBuilderImpl struct {
	logger *zap.SugaredLogger
}

func NewNotificationConfigBuilderImpl(logger *zap.SugaredLogger) *NotificationConfigBuilderImpl {
	return &NotificationConfigBuilderImpl{
		logger: logger,
	}
}

func (impl NotificationConfigBuilderImpl) BuildNotificationSettingsConfig(notificationSettingsRequest *beans.NotificationConfigRequest, existingNotificationSettingsConfig *repository.NotificationSettingsView, userId int32) (*repository.NotificationSettingsView, error) {
	nsConfig := &beans.NSConfig{}
	nsConfig.TeamId = notificationSettingsRequest.TeamId
	nsConfig.AppId = notificationSettingsRequest.AppId
	nsConfig.EnvId = notificationSettingsRequest.EnvId
	nsConfig.ClusterId = notificationSettingsRequest.ClusterId
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

func (impl NotificationConfigBuilderImpl) BuildNewNotificationSettings(notificationSettingsRequest *beans.NotificationConfigRequest, notificationSettingsView *repository.NotificationSettingsView) ([]repository.NotificationSettings, error) {
	// tempRequest := generateSettingCombinationsV1(notificationSettingsRequest)
	tempRequest := notificationSettingsRequest.GenerateSettingCombinations()
	var notificationSettings []repository.NotificationSettings
	for _, item := range tempRequest {
		for _, e := range notificationSettingsRequest.EventTypeIds {
			notificationSetting, err := impl.BuildNotificationSettingWithPipeline(item.TeamId, item.EnvId, item.AppId, item.PipelineId, item.ClusterId, notificationSettingsRequest.PipelineType, e, notificationSettingsView.Id, notificationSettingsRequest.Providers)
			if err != nil {
				impl.logger.Error(err)
				return nil, err
			}
			notificationSettings = append(notificationSettings, notificationSetting)
		}
	}
	return notificationSettings, nil
}

func (impl NotificationConfigBuilderImpl) buildNotificationSetting(notificationSettingsRequest *beans.NotificationSettingRequest, notificationSettingsView *repository.NotificationSettingsView, eventTypeId int) (repository.NotificationSettings, error) {
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

func (impl NotificationConfigBuilderImpl) BuildNotificationSettingWithPipeline(teamId *int, envId *int, appId *int, pipelineId *int, clusterId *int, pipelineType util.PipelineType, eventTypeId int, viewId int, providers []*beans.Provider) (repository.NotificationSettings, error) {

	if teamId == nil && appId == nil && envId == nil && pipelineId == nil && clusterId == nil {
		return repository.NotificationSettings{}, errors.New("no filter criteria is selected")
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
		ClusterId:    clusterId,
	}
	return notificationSetting, nil
}
