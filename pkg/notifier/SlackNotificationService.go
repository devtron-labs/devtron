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
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/team"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	util2 "github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
	"strings"
	"time"
)

type SlackNotificationService interface {
	SaveOrEditNotificationConfig(channelReq []SlackConfigDto, userId int32) ([]int, error)
	FetchSlackNotificationConfigById(id int) (*SlackConfigDto, error)
	FetchAllSlackNotificationConfig() ([]*SlackConfigDto, error)
	FetchAllSlackNotificationConfigAutocomplete() ([]*NotificationChannelAutoResponse, error)
	RecipientListingSuggestion(value string) ([]*NotificationRecipientListingResponse, error)
}

type SlackNotificationServiceImpl struct {
	logger                         *zap.SugaredLogger
	teamService                    team.TeamService
	slackRepository                repository.SlackNotificationRepository
	userRepository                 repository2.UserRepository
	notificationSettingsRepository repository.NotificationSettingsRepository
}

type SlackChannelConfig struct {
	Channel         util2.Channel    `json:"channel" validate:"required"`
	SlackConfigDtos []SlackConfigDto `json:"configs"`
}

type SlackConfigDto struct {
	OwnerId     int32  `json:"userId" validate:"number"`
	TeamId      int    `json:"teamId" validate:"required"`
	WebhookUrl  string `json:"webhookUrl" validate:"required"`
	ConfigName  string `json:"configName" validate:"required"`
	Description string `json:"description"`
	Id          int    `json:"id" validate:"number"`
}

func NewSlackNotificationServiceImpl(logger *zap.SugaredLogger, slackRepository repository.SlackNotificationRepository, teamService team.TeamService,
	userRepository repository2.UserRepository, notificationSettingsRepository repository.NotificationSettingsRepository) *SlackNotificationServiceImpl {
	return &SlackNotificationServiceImpl{
		logger:                         logger,
		teamService:                    teamService,
		slackRepository:                slackRepository,
		userRepository:                 userRepository,
		notificationSettingsRepository: notificationSettingsRepository,
	}
}

func (impl *SlackNotificationServiceImpl) SaveOrEditNotificationConfig(channelReq []SlackConfigDto, userId int32) ([]int, error) {
	var responseIds []int
	slackConfigs := buildSlackNewConfigs(channelReq, userId)
	for _, config := range slackConfigs {
		if config.Id != 0 {
			model, err := impl.slackRepository.FindOne(config.Id)
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("err while fetching slack config", "err", err)
				return []int{}, err
			}
			impl.buildConfigUpdateModel(config, model, userId)
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

func (impl *SlackNotificationServiceImpl) FetchSlackNotificationConfigById(id int) (*SlackConfigDto, error) {
	slackConfig, err := impl.slackRepository.FindOne(id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return nil, err
	}
	slackConfigDto := impl.adaptSlackConfig(*slackConfig)
	return &slackConfigDto, nil
}

func (impl *SlackNotificationServiceImpl) FetchAllSlackNotificationConfig() ([]*SlackConfigDto, error) {
	var responseDto []*SlackConfigDto
	slackConfigs, err := impl.slackRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return []*SlackConfigDto{}, err
	}
	for _, slackConfig := range slackConfigs {
		slackConfigDto := impl.adaptSlackConfig(slackConfig)
		responseDto = append(responseDto, &slackConfigDto)
	}
	if responseDto == nil {
		responseDto = make([]*SlackConfigDto, 0)
	}
	return responseDto, nil
}

func (impl *SlackNotificationServiceImpl) FetchAllSlackNotificationConfigAutocomplete() ([]*NotificationChannelAutoResponse, error) {
	var responseDto []*NotificationChannelAutoResponse
	slackConfigs, err := impl.slackRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return []*NotificationChannelAutoResponse{}, err
	}
	for _, slackConfig := range slackConfigs {
		slackConfigDto := &NotificationChannelAutoResponse{
			Id:         slackConfig.Id,
			ConfigName: slackConfig.ConfigName,
			TeamId:     slackConfig.TeamId,
		}
		responseDto = append(responseDto, slackConfigDto)
	}
	return responseDto, nil
}

func (impl *SlackNotificationServiceImpl) adaptSlackConfig(slackConfig repository.SlackConfig) SlackConfigDto {
	slackConfigDto := SlackConfigDto{
		OwnerId:     slackConfig.OwnerId,
		TeamId:      slackConfig.TeamId,
		WebhookUrl:  slackConfig.WebHookUrl,
		ConfigName:  slackConfig.ConfigName,
		Description: slackConfig.Description,
		Id:          slackConfig.Id,
	}
	return slackConfigDto
}

func buildSlackNewConfigs(slackReq []SlackConfigDto, userId int32) []*repository.SlackConfig {
	var slackConfigs []*repository.SlackConfig
	for _, c := range slackReq {
		slackConfig := &repository.SlackConfig{
			Id:          c.Id,
			ConfigName:  c.ConfigName,
			WebHookUrl:  c.WebhookUrl,
			Description: c.Description,
			AuditLog: sql.AuditLog{
				CreatedBy: userId,
				CreatedOn: time.Now(),
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		if c.TeamId != 0 {
			slackConfig.TeamId = c.TeamId
		} else {
			slackConfig.OwnerId = userId
		}
		slackConfigs = append(slackConfigs, slackConfig)
	}
	return slackConfigs
}

func (impl *SlackNotificationServiceImpl) buildConfigUpdateModel(slackConfig *repository.SlackConfig, model *repository.SlackConfig, userId int32) {
	model.WebHookUrl = slackConfig.WebHookUrl
	model.ConfigName = slackConfig.ConfigName
	model.Description = slackConfig.Description
	if slackConfig.TeamId != 0 {
		model.TeamId = slackConfig.TeamId
	} else {
		model.OwnerId = slackConfig.OwnerId
	}
	model.UpdatedOn = time.Now()
	model.UpdatedBy = userId
}

func (impl *SlackNotificationServiceImpl) RecipientListingSuggestion(value string) ([]*NotificationRecipientListingResponse, error) {
	var results []*NotificationRecipientListingResponse

	sesConfigs, err := impl.slackRepository.FindByName(value)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return []*NotificationRecipientListingResponse{}, err
	}
	for _, sesConfig := range sesConfigs {
		result := &NotificationRecipientListingResponse{
			ConfigId:  sesConfig.Id,
			Recipient: sesConfig.ConfigName,
			Dest:      util2.Slack}
		results = append(results, result)
	}

	userList, err := impl.userRepository.FetchUserMatchesByEmailId(value)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return []*NotificationRecipientListingResponse{}, err
	}
	for _, item := range userList {
		result := &NotificationRecipientListingResponse{
			ConfigId:  int(item.Id),
			Recipient: item.EmailId,
			Dest:      util2.SES}
		results = append(results, result)
	}

	nsv, err := impl.notificationSettingsRepository.FindAll(0, 20)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "error", err)
		return []*NotificationRecipientListingResponse{}, err
	}
	for _, item := range nsv {
		var dat map[string]interface{}
		if err := json.Unmarshal([]byte(item.Config), &dat); err != nil {
			impl.logger.Errorw("Unmarshal error", "error", err)
			return []*NotificationRecipientListingResponse{}, err
		}
		providers := dat["providers"]
		if providers != nil {
			data := providers.([]interface{})
			for _, item := range data {
				if item != nil {
					for k, v := range item.(map[string]interface{}) {
						if v != nil && len(value) > 0 {
							if k == "recipient" && strings.Contains(v.(string), value) {
								result := &NotificationRecipientListingResponse{
									Recipient: v.(string),
								}
								if strings.Contains(v.(string), "https://") {
									result.Dest = util2.Slack
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
