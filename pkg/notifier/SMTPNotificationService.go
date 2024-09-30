/*
 * Copyright (c) 2024. Devtron Inc.
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
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/notifier/beans"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/team"
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
	smtpConfigs := buildSMTPNewConfigs(channelReq, userId)
	for _, config := range smtpConfigs {
		if config.Id != 0 {

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

			model, err := impl.smtpRepository.FindOne(config.Id)
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("err while fetching smtp config", "err", err)
				return []int{}, err
			}
			impl.buildConfigUpdateModel(config, model, userId)
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
	smtpConfigDto := impl.adaptSMTPConfig(smtpConfig)
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
		smtpConfigDto := impl.adaptSMTPConfig(smtpConfig)
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

func (impl *SMTPNotificationServiceImpl) adaptSMTPConfig(smtpConfig *repository.SMTPConfig) *beans.SMTPConfigDto {
	smtpConfigDto := &beans.SMTPConfigDto{
		OwnerId:      smtpConfig.OwnerId,
		Port:         smtpConfig.Port,
		Host:         smtpConfig.Host,
		AuthType:     smtpConfig.AuthType,
		AuthUser:     smtpConfig.AuthUser,
		AuthPassword: smtpConfig.AuthPassword,
		FromEmail:    smtpConfig.FromEmail,
		ConfigName:   smtpConfig.ConfigName,
		Description:  smtpConfig.Description,
		Id:           smtpConfig.Id,
		Default:      smtpConfig.Default,
		Deleted:      false,
	}
	return smtpConfigDto
}

func buildSMTPNewConfigs(smtpReq []*beans.SMTPConfigDto, userId int32) []*repository.SMTPConfig {
	var smtpConfigs []*repository.SMTPConfig
	for _, c := range smtpReq {
		smtpConfig := &repository.SMTPConfig{
			Id:           c.Id,
			Port:         c.Port,
			Host:         c.Host,
			AuthType:     c.AuthType,
			AuthUser:     c.AuthUser,
			AuthPassword: c.AuthPassword,
			ConfigName:   c.ConfigName,
			FromEmail:    c.FromEmail,
			Deleted:      false,
			Description:  c.Description,
			Default:      c.Default,
			AuditLog: sql.AuditLog{
				CreatedBy: userId,
				CreatedOn: time.Now(),
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		smtpConfig.OwnerId = userId
		smtpConfigs = append(smtpConfigs, smtpConfig)
	}
	return smtpConfigs
}

func (impl *SMTPNotificationServiceImpl) buildConfigUpdateModel(smtpConfig *repository.SMTPConfig, model *repository.SMTPConfig, userId int32) {
	model.Id = smtpConfig.Id
	model.OwnerId = smtpConfig.OwnerId
	model.Port = smtpConfig.Port
	model.Host = smtpConfig.Host
	model.AuthUser = smtpConfig.AuthUser
	model.AuthType = smtpConfig.AuthType
	model.AuthPassword = smtpConfig.AuthPassword
	model.FromEmail = smtpConfig.FromEmail
	model.ConfigName = smtpConfig.ConfigName
	model.Description = smtpConfig.Description
	model.Default = smtpConfig.Default
	model.UpdatedOn = time.Now()
	model.UpdatedBy = userId
	model.Deleted = false
}

func (impl *SMTPNotificationServiceImpl) DeleteNotificationConfig(deleteReq *beans.SMTPConfigDto, userId int32) error {
	existingConfig, err := impl.smtpRepository.FindOne(deleteReq.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete", "err", err, "id", deleteReq.Id)
		return err
	}
	notifications, err := impl.notificationSettingsRepository.FindNotificationSettingsByConfigIdAndConfigType(deleteReq.Id, beans.SMTP_CONFIG_TYPE)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in deleting smtp config", "config", deleteReq)
		return err
	}
	if len(notifications) > 0 {
		impl.logger.Errorw("found notifications using this config, cannot delete", "config", deleteReq)
		return fmt.Errorf(" Please delete all notifications using this config before deleting")
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
