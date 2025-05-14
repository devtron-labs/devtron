/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package notifier

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/notifier/adapter"
	"github.com/devtron-labs/devtron/pkg/notifier/beans"
	"github.com/devtron-labs/devtron/pkg/team"
	eventUtil "github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type SMTPNotificationService interface {
	SaveOrEditNotificationConfig(channelReq []*beans.SMTPConfigDto, userId int32) ([]int, error)
	FetchSMTPNotificationConfigById(id int) (*beans.SMTPConfigDto, error)
	FetchAllSMTPNotificationConfig() ([]*beans.SMTPConfigDto, error)
	FetchAllSMTPNotificationConfigAutocomplete() ([]*beans.NotificationChannelAutoResponse, error)
	DeleteNotificationConfig(channelReq *beans.SMTPConfigDto, userId int32) error
}

type SMTPNotificationServiceImpl struct {
	logger                         *zap.SugaredLogger
	teamService                    team.TeamService
	smtpRepository                 repository.SMTPNotificationRepository
	notificationSettingsRepository repository.NotificationSettingsRepository
}

func NewSMTPNotificationServiceImpl(logger *zap.SugaredLogger, smtpRepository repository.SMTPNotificationRepository,
	teamService team.TeamService, notificationSettingsRepository repository.NotificationSettingsRepository) *SMTPNotificationServiceImpl {
	return &SMTPNotificationServiceImpl{
		logger:                         logger,
		teamService:                    teamService,
		smtpRepository:                 smtpRepository,
		notificationSettingsRepository: notificationSettingsRepository,
	}
}

func (impl *SMTPNotificationServiceImpl) SaveOrEditNotificationConfig(channelReq []*beans.SMTPConfigDto, userId int32) ([]int, error) {
	var responseIds []int
	smtpConfigs := adapter.BuildSMTPNewConfigs(channelReq, userId)
	for _, config := range smtpConfigs {
		if config.Id != 0 {

			model, err := impl.smtpRepository.FindOne(config.Id)
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("err while fetching smtp config", "err", err)
				return []int{}, err
			}

			if config.Default {
				_, err := impl.smtpRepository.UpdateSMTPConfigDefault()
				if err != nil && !util.IsErrNoRows(err) {
					impl.logger.Errorw("err while updating smtp config", "err", err)
					return []int{}, err
				}
			} else {
				// check if this config is already default, we don't want to reverse it
				if model.Default {
					return []int{}, fmt.Errorf("cannot update default config to non default")
				}
				_, err := impl.smtpRepository.FindDefault()
				if err != nil && !util.IsErrNoRows(err) {
					impl.logger.Errorw("err while updating smtp config", "err", err)
					return []int{}, err
				} else if util.IsErrNoRows(err) {
					config.Default = true
				}
			}

			adapter.BuildConfigUpdateModelForSMTP(config, model, userId)
			model, uErr := impl.smtpRepository.UpdateSMTPConfig(model)
			if uErr != nil {
				impl.logger.Errorw("err while updating smtp config", "err", err)
				return []int{}, uErr
			}
		} else {

			if config.Default {
				_, err := impl.smtpRepository.UpdateSMTPConfigDefault()
				if err != nil && !util.IsErrNoRows(err) {
					impl.logger.Errorw("err while updating smtp config", "err", err)
					return []int{}, err
				}
			} else {
				_, err := impl.smtpRepository.FindDefault()
				if err != nil && !util.IsErrNoRows(err) {
					impl.logger.Errorw("err while updating smtp config", "err", err)
					return []int{}, err
				} else if util.IsErrNoRows(err) {
					config.Default = true
				}
			}

			_, iErr := impl.smtpRepository.SaveSMTPConfig(config)
			if iErr != nil {
				impl.logger.Errorw("err while inserting smtp config", "err", iErr)
				return []int{}, iErr
			}
		}
		responseIds = append(responseIds, config.Id)
	}
	return responseIds, nil
}

func (impl *SMTPNotificationServiceImpl) FetchSMTPNotificationConfigById(id int) (*beans.SMTPConfigDto, error) {
	smtpConfig, err := impl.smtpRepository.FindOne(id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all smtp config", "err", err)
		return nil, err
	}
	smtpConfigDto := adapter.AdaptSMTPConfig(smtpConfig)
	return smtpConfigDto, nil
}

func (impl *SMTPNotificationServiceImpl) FetchAllSMTPNotificationConfig() ([]*beans.SMTPConfigDto, error) {
	var responseDto []*beans.SMTPConfigDto
	smtpConfigs, err := impl.smtpRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all smtp config", "err", err)
		return []*beans.SMTPConfigDto{}, err
	}
	for _, smtpConfig := range smtpConfigs {
		smtpConfigDto := adapter.AdaptSMTPConfig(smtpConfig)
		smtpConfigDto.AuthPassword = "**********"
		responseDto = append(responseDto, smtpConfigDto)
	}
	if responseDto == nil {
		responseDto = make([]*beans.SMTPConfigDto, 0)
	}
	return responseDto, nil
}

func (impl *SMTPNotificationServiceImpl) FetchAllSMTPNotificationConfigAutocomplete() ([]*beans.NotificationChannelAutoResponse, error) {
	var responseDto []*beans.NotificationChannelAutoResponse
	smtpConfigs, err := impl.smtpRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all smtp config", "err", err)
		return []*beans.NotificationChannelAutoResponse{}, err
	}
	for _, smtpConfig := range smtpConfigs {
		smtpConfigDto := &beans.NotificationChannelAutoResponse{
			Id:         smtpConfig.Id,
			ConfigName: smtpConfig.ConfigName}
		responseDto = append(responseDto, smtpConfigDto)
	}
	return responseDto, nil
}

func (impl *SMTPNotificationServiceImpl) DeleteNotificationConfig(deleteReq *beans.SMTPConfigDto, userId int32) error {
	existingConfig, err := impl.smtpRepository.FindOne(deleteReq.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete", "err", err, "id", deleteReq.Id)
		return err
	}
	notifications, err := impl.notificationSettingsRepository.FindNotificationSettingsByConfigIdAndConfigType(deleteReq.Id, eventUtil.SMTP.String())
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in deleting smtp config", "config", deleteReq)
		return err
	}
	if len(notifications) > 0 {
		impl.logger.Errorw("found notifications using this config, cannot delete", "config", deleteReq)
		return fmt.Errorf(" Please delete all notifications using this config before deleting")
	}
	// check if default then dont delete
	if existingConfig.Default {
		return fmt.Errorf("default configuration cannot be deleted")
	}
	existingConfig.UpdatedOn = time.Now()
	existingConfig.UpdatedBy = userId
	//deleting smtp config
	err = impl.smtpRepository.MarkSMTPConfigDeleted(existingConfig)
	if err != nil {
		impl.logger.Errorw("error in deleting smtp config", "err", err, "id", existingConfig.Id)
		return err
	}
	return nil
}
