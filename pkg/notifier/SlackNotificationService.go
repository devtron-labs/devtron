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
	"github.com/devtron-labs/devtron/pkg/notifier/adapter"
	"github.com/devtron-labs/devtron/pkg/notifier/beans"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	repository2 "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/team"
	util2 "github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type SlackNotificationService interface {
	SaveOrEditNotificationConfig(channelReq []beans.SlackConfigDto, userId int32) ([]int, error)
	FetchSlackNotificationConfigById(id int) (*beans.SlackConfigDto, error)
	FetchAllSlackNotificationConfig() ([]*beans.SlackConfigDto, error)
	FetchAllSlackNotificationConfigAutocomplete() ([]*beans.NotificationChannelAutoResponse, error)
	RecipientListingSuggestion(value string) ([]*beans.NotificationRecipientListingResponse, error)
	DeleteNotificationConfig(deleteReq *beans.SlackConfigDto, userId int32) error
}

type SlackNotificationServiceImpl struct {
	logger                         *zap.SugaredLogger
	teamService                    team.TeamService
	slackRepository                repository.SlackNotificationRepository
	webhookRepository              repository.WebhookNotificationRepository
	userRepository                 repository2.UserRepository
	notificationSettingsRepository repository.NotificationSettingsRepository
}

func NewSlackNotificationServiceImpl(logger *zap.SugaredLogger, slackRepository repository.SlackNotificationRepository, webhookRepository repository.WebhookNotificationRepository, teamService team.TeamService,
	userRepository repository2.UserRepository, notificationSettingsRepository repository.NotificationSettingsRepository) *SlackNotificationServiceImpl {
	return &SlackNotificationServiceImpl{
		logger:                         logger,
		teamService:                    teamService,
		slackRepository:                slackRepository,
		webhookRepository:              webhookRepository,
		userRepository:                 userRepository,
		notificationSettingsRepository: notificationSettingsRepository,
	}
}

func (impl *SlackNotificationServiceImpl) SaveOrEditNotificationConfig(channelReq []beans.SlackConfigDto, userId int32) ([]int, error) {
	var responseIds []int
	slackConfigs := adapter.BuildSlackNewConfigs(channelReq, userId)
	for _, config := range slackConfigs {
		if config.Id != 0 {
			model, err := impl.slackRepository.FindOne(config.Id)
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("err while fetching slack config", "err", err)
				return []int{}, err
			}
			adapter.BuildConfigUpdateModelForSlack(config, model, userId)
			model, uErr := impl.slackRepository.UpdateSlackConfig(model)
			if uErr != nil {
				impl.logger.Errorw("err while updating slack config", "err", err)
				return []int{}, uErr
			}
		} else {
			_, iErr := impl.slackRepository.SaveSlackConfig(config)
			if iErr != nil {
				impl.logger.Errorw("err while inserting slack config", "err", iErr)
				return []int{}, iErr
			}
		}
		responseIds = append(responseIds, config.Id)
	}
	return responseIds, nil
}

func (impl *SlackNotificationServiceImpl) FetchSlackNotificationConfigById(id int) (*beans.SlackConfigDto, error) {
	slackConfig, err := impl.slackRepository.FindOne(id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return nil, err
	}
	slackConfigDto := adapter.AdaptSlackConfig(*slackConfig)
	return &slackConfigDto, nil
}

func (impl *SlackNotificationServiceImpl) FetchAllSlackNotificationConfig() ([]*beans.SlackConfigDto, error) {
	var responseDto []*beans.SlackConfigDto
	slackConfigs, err := impl.slackRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return []*beans.SlackConfigDto{}, err
	}
	for _, slackConfig := range slackConfigs {
		slackConfigDto := adapter.AdaptSlackConfig(slackConfig)
		responseDto = append(responseDto, &slackConfigDto)
	}
	if responseDto == nil {
		responseDto = make([]*beans.SlackConfigDto, 0)
	}
	return responseDto, nil
}

func (impl *SlackNotificationServiceImpl) FetchAllSlackNotificationConfigAutocomplete() ([]*beans.NotificationChannelAutoResponse, error) {
	var responseDto []*beans.NotificationChannelAutoResponse
	slackConfigs, err := impl.slackRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return []*beans.NotificationChannelAutoResponse{}, err
	}
	for _, slackConfig := range slackConfigs {
		slackConfigDto := &beans.NotificationChannelAutoResponse{
			Id:         slackConfig.Id,
			ConfigName: slackConfig.ConfigName,
			TeamId:     slackConfig.TeamId,
		}
		responseDto = append(responseDto, slackConfigDto)
	}
	return responseDto, nil
}

func (impl *SlackNotificationServiceImpl) RecipientListingSuggestion(value string) ([]*beans.NotificationRecipientListingResponse, error) {
	var results []*beans.NotificationRecipientListingResponse

	slackConfigs, err := impl.slackRepository.FindByName(value)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return []*beans.NotificationRecipientListingResponse{}, err
	}
	for _, slackConfig := range slackConfigs {
		result := &beans.NotificationRecipientListingResponse{
			ConfigId:  slackConfig.Id,
			Recipient: slackConfig.ConfigName,
			Dest:      util2.Slack}
		results = append(results, result)
	}
	webhookConfigs, err := impl.webhookRepository.FindByName(value)

	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all webhook config", "err", err)
		return []*beans.NotificationRecipientListingResponse{}, err
	}
	for _, webhookConfig := range webhookConfigs {
		result := &beans.NotificationRecipientListingResponse{
			ConfigId:  webhookConfig.Id,
			Recipient: webhookConfig.ConfigName,
			Dest:      util2.Webhook}
		results = append(results, result)
	}
	userList, err := impl.userRepository.FetchUserMatchesByEmailIdExcludingApiTokenUser(value)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return []*beans.NotificationRecipientListingResponse{}, err
	}
	for _, item := range userList {
		result := &beans.NotificationRecipientListingResponse{
			ConfigId:  int(item.Id),
			Recipient: item.EmailId,
			Dest:      util2.SES}
		results = append(results, result)
	}

	nsv, err := impl.notificationSettingsRepository.FindAll(0, 20)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "error", err)
		return []*beans.NotificationRecipientListingResponse{}, err
	}
	for _, item := range nsv {
		var dat map[string]interface{}
		if err := json.Unmarshal([]byte(item.Config), &dat); err != nil {
			impl.logger.Errorw("Unmarshal error", "error", err)
			return []*beans.NotificationRecipientListingResponse{}, err
		}
		providers := dat["providers"]
		if providers != nil {
			data := providers.([]interface{})
			for _, item := range data {
				if item != nil {
					for k, v := range item.(map[string]interface{}) {
						if v != nil && len(value) > 0 {
							if k == "recipient" && strings.Contains(v.(string), value) {
								result := &beans.NotificationRecipientListingResponse{
									Recipient: v.(string),
								}
								if strings.Contains(v.(string), beans.SLACK_URL) {
									result.Dest = util2.Slack
								} else if strings.Contains(v.(string), beans.WEBHOOK_URL) {
									result.Dest = util2.Webhook
								} else {
									result.Dest = util2.SES
								}
								results = append(results, result)
							}
						}
					}
				}
			}
		}
	}
	return results, nil
}

func (impl *SlackNotificationServiceImpl) DeleteNotificationConfig(deleteReq *beans.SlackConfigDto, userId int32) error {
	existingConfig, err := impl.slackRepository.FindOne(deleteReq.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete", "err", err, "id", deleteReq.Id)
		return err
	}
	notifications, err := impl.notificationSettingsRepository.FindNotificationSettingsByConfigIdAndConfigType(deleteReq.Id, beans.SLACK_CONFIG_TYPE)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in deleting slack config", "config", deleteReq)
		return err
	}
	if len(notifications) > 0 {
		impl.logger.Errorw("found notifications using this config, cannot delete", "config", deleteReq)
		return fmt.Errorf(" Please delete all notifications using this config before deleting")
	}

	existingConfig.UpdatedOn = time.Now()
	existingConfig.UpdatedBy = userId
	//deleting slack config
	err = impl.slackRepository.MarkSlackConfigDeleted(existingConfig)
	if err != nil {
		impl.logger.Errorw("error in deleting slack config", "err", err, "id", existingConfig.Id)
		return err
	}
	return nil
}
