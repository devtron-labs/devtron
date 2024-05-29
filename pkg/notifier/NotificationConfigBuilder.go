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
	"fmt"
	"github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	notifierBean "github.com/devtron-labs/devtron/pkg/notifier/bean"
	"github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"strings"
	"time"
)

type NotificationConfigBuilder interface {
	BuildNotificationSettingsConfig(notificationSettingsRequest *NotificationConfigRequest, existingNotificationSettingsConfig *repository.NotificationSettingsView, userId int32) (*repository.NotificationSettingsView, error)
	BuildNewNotificationSettings(notificationSettingsRequest *NotificationConfigRequest, notificationSettingsView *repository.NotificationSettingsView, eventIdRuleIdMapping map[int]int) ([]repository.NotificationSettings, error)
	BuildNotificationSettingWithPipeline(teamId *int, envId *int, appId *int, pipelineId *int, pipelineType util.PipelineType, eventTypeId int, viewId int, providers []*client.Provider) (repository.NotificationSettings, error)
	BuildNotificationRules(filterCondition map[int]string, userId int32) (map[int]*repository.NotificationRule, error)
	GenerateAdditionalConfigByEventIds(additionalConfigJson map[int]string) (map[int]string, error)
	UpdateExpressionInNotificationRules(eventIdToNotificationRule map[int]*repository.NotificationRule, filterCondition map[int]string, userId int32) ([]*repository.NotificationRule, error)
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
	TeamId       []*int             `json:"teamId"`
	AppId        []*int             `json:"appId"`
	EnvId        []*int             `json:"envId"`
	PipelineId   *int               `json:"pipelineId"`
	PipelineType util.PipelineType  `json:"pipelineType" validate:"required"`
	EventTypeIds []int              `json:"eventTypeIds" validate:"required"`
	Providers    []*client.Provider `json:"providers" validate:"required"`
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
	if slices.Contains(notificationSettingsRequest.EventTypeIds, int(util.ImageScanning)) {
		notificationSettingsRequest.IsInternal = true
	}
	notificationSettingsView := &repository.NotificationSettingsView{
		Config:   string(config),
		Internal: notificationSettingsRequest.IsInternal,
		// ConfigName:    notificationSettingsRequest.ConfigName,
		// AppId:         notificationSettingsRequest.AppId,
		// EnvironmentId: notificationSettingsRequest.EnvId,
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

func (impl NotificationConfigBuilderImpl) BuildNewNotificationSettings(notificationSettingsRequest *NotificationConfigRequest, notificationSettingsView *repository.NotificationSettingsView, eventIdRuleIdMapping map[int]int) ([]repository.NotificationSettings, error) {
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
	eventIdToAdditionalConfig, err := impl.GenerateAdditionalConfigByEventIds(notificationSettingsRequest.AdditionalConfigJson)
	if err != nil {
		impl.logger.Error("error in unmarshalling additional config", "additionalConfig", notificationSettingsRequest.AdditionalConfigJson, "err", err)
		return nil, err
	}
	for _, item := range tempRequest {
		for _, e := range notificationSettingsRequest.EventTypeIds {
			notificationSetting, err := impl.BuildNotificationSettingWithPipeline(item.TeamId, item.EnvId, item.AppId, item.PipelineId, notificationSettingsRequest.PipelineType, e, notificationSettingsView.Id, notificationSettingsRequest.Providers)
			if err != nil {
				impl.logger.Error("error in json marshalling", "providers", notificationSettingsRequest.Providers, "err", err)
				return nil, err
			}
			if eventIdRuleIdMapping != nil && eventIdRuleIdMapping[notificationSetting.EventTypeId] != 0 {
				notificationSetting.NotificationRuleId = eventIdRuleIdMapping[notificationSetting.EventTypeId]
			}
			if eventIdToAdditionalConfig != nil && len(eventIdToAdditionalConfig[notificationSetting.EventTypeId]) != 0 {
				notificationSetting.AdditionalConfigJson = eventIdToAdditionalConfig[notificationSetting.EventTypeId]
			}
			notificationSettings = append(notificationSettings, notificationSetting)
		}
	}
	return notificationSettings, nil
}

func (impl NotificationConfigBuilderImpl) UpdateExpressionInNotificationRules(eventIdToNotificationRule map[int]*repository.NotificationRule, filterCondition map[int]string, userId int32) ([]*repository.NotificationRule, error) {
	expression, err := impl.GenerateFilterExpression(filterCondition)
	if err != nil {
		impl.logger.Errorw("error in filter expression", "filterCondition", filterCondition)
		return nil, err
	}
	if expression == nil {
		return nil, nil
	}
	var updatedNotificationRules []*repository.NotificationRule
	for eventId, notificationRule := range eventIdToNotificationRule {
		if len(expression[eventId]) == 0 || notificationRule == nil {
			continue
		}
		updatedNotificationRule := *notificationRule
		updatedNotificationRule.Expression = expression[eventId]
		updatedNotificationRule.UpdateAuditLog(userId)
		updatedNotificationRules = append(updatedNotificationRules, &updatedNotificationRule)
	}
	return updatedNotificationRules, nil
}

func (impl NotificationConfigBuilderImpl) BuildNotificationRules(filterCondition map[int]string, userId int32) (map[int]*repository.NotificationRule, error) {
	expression, err := impl.GenerateFilterExpression(filterCondition)
	if err != nil {
		impl.logger.Errorw("error in generating filter expression", "filterCondition", filterCondition, "err", err)
		return nil, err
	}
	if expression == nil {
		return nil, nil
	}
	notificationRulesMap := make(map[int]*repository.NotificationRule)
	for eventId, _ := range filterCondition {
		if !slices.Contains(util.RulesSupportedEvents, eventId) ||
			expression[eventId] == "" {
			continue
		}
		notificationRule := repository.NotificationRule{
			ConditionType: notifierBean.PASS,
			Expression:    expression[eventId],
		}
		notificationRule.CreateAuditLog(userId)
		notificationRulesMap[eventId] = &notificationRule
	}
	return notificationRulesMap, nil
}

func (impl NotificationConfigBuilderImpl) GenerateAdditionalConfigByEventIds(additionalConfigJson map[int]string) (map[int]string, error) {
	if additionalConfigJson == nil {
		return nil, nil
	}
	if value, ok := additionalConfigJson[int(util.ImageScanning)]; ok {
		var filterFlagsJson map[int]string
		if value == "" {
			return filterFlagsJson, nil
		}
		var flags repository.ImageScanFilterFlags
		err := json.Unmarshal([]byte(value), &flags)
		if err != nil {
			return filterFlagsJson, err
		}
		flagsJson, err := json.Marshal(flags)
		if err != nil {
			return filterFlagsJson, err
		}
		filterFlagsJson = make(map[int]string)
		filterFlagsJson[int(util.ImageScanning)] = string(flagsJson)
		return filterFlagsJson, nil
	}
	return nil, nil
}

func (impl NotificationConfigBuilderImpl) GenerateFilterExpression(filterCondition map[int]string) (map[int]string, error) {
	if filterCondition == nil {
		return nil, nil
	}
	if value, ok := filterCondition[int(util.ImageScanning)]; ok {
		var filterExpression map[int]string
		if value == "" {
			return filterExpression, nil
		}
		var filters repository.ImageScanFilterConditions
		err := json.Unmarshal([]byte(value), &filters)
		if err != nil {
			return filterExpression, err
		}
		var severityExpression, policyExpression, expression string
		if filters.Severity != nil {
			severityExpression = fmt.Sprintf("%s in ['%s']", resourceFilter.Severity, strings.Join(filters.Severity, "', '"))
		}
		if filters.PolicyPermission != nil {
			policyExpression = fmt.Sprintf("%s in ['%s']", resourceFilter.PolicyPermission, strings.Join(filters.PolicyPermission, "', '"))
		}

		if len(severityExpression) != 0 && len(policyExpression) != 0 {
			expression = fmt.Sprintf("%s && %s", severityExpression, policyExpression)
		} else if len(severityExpression) != 0 {
			expression = severityExpression
		} else if len(policyExpression) != 0 {
			expression = policyExpression
		}
		filterExpression = make(map[int]string)
		filterExpression[int(util.ImageScanning)] = expression
		return filterExpression, nil
	}
	return nil, nil
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

func (impl NotificationConfigBuilderImpl) BuildNotificationSettingWithPipeline(teamId *int, envId *int, appId *int, pipelineId *int, pipelineType util.PipelineType, eventTypeId int, viewId int, providers []*client.Provider) (repository.NotificationSettings, error) {

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
