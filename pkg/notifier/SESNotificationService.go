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
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/team"
	util2 "github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

const SES_CONFIG_TYPE = "ses"

type SESNotificationService interface {
	SaveOrEditNotificationConfig(channelReq []*SESConfigDto, userId int32) ([]int, error)
	FetchSESNotificationConfigById(id int) (*SESConfigDto, error)
	FetchAllSESNotificationConfig() ([]*SESConfigDto, error)
	FetchAllSESNotificationConfigAutocomplete() ([]*NotificationChannelAutoResponse, error)
	DeleteNotificationConfig(channelReq *SESConfigDto, userId int32) error
}

type SESNotificationServiceImpl struct {
	logger                         *zap.SugaredLogger
	teamService                    team.TeamService
	sesRepository                  repository.SESNotificationRepository
	notificationSettingsRepository repository.NotificationSettingsRepository
}

type SESChannelConfig struct {
	Channel       util2.Channel   `json:"channel" validate:"required"`
	SESConfigDtos []*SESConfigDto `json:"configs"`
}

type SESConfigDto struct {
	OwnerId      int32  `json:"userId" validate:"number"`
	TeamId       int    `json:"teamId" validate:"number"`
	Region       string `json:"region" validate:"required"`
	AccessKey    string `json:"accessKey" validate:"required"`
	SecretKey    string `json:"secretKey" validate:"required"`
	FromEmail    string `json:"fromEmail" validate:"email,required"`
	ToEmail      string `json:"toEmail"`
	SessionToken string `json:"sessionToken"`
	ConfigName   string `json:"configName" validate:"required"`
	Description  string `json:"description"`
	Id           int    `json:"id" validate:"number"`
	Default      bool   `json:"default,notnull"`
}

type NotificationChannelAutoResponse struct {
	ConfigName string `json:"configName"`
	Id         int    `json:"id"`
	TeamId     int    `json:"-"`
}

type NotificationRecipientListingResponse struct {
	Dest      util2.Channel `json:"dest"`
	ConfigId  int           `json:"configId"`
	Recipient string        `json:"recipient"`
}

func NewSESNotificationServiceImpl(logger *zap.SugaredLogger, sesRepository repository.SESNotificationRepository,
	teamService team.TeamService, notificationSettingsRepository repository.NotificationSettingsRepository) *SESNotificationServiceImpl {
	return &SESNotificationServiceImpl{
		logger:                         logger,
		teamService:                    teamService,
		sesRepository:                  sesRepository,
		notificationSettingsRepository: notificationSettingsRepository,
	}
}

func (impl *SESNotificationServiceImpl) SaveOrEditNotificationConfig(channelReq []*SESConfigDto, userId int32) ([]int, error) {
	var responseIds []int
	sesConfigs := buildSESNewConfigs(channelReq, userId)
	for _, config := range sesConfigs {
		if config.Id != 0 {

			if config.Default {
				_, err := impl.sesRepository.UpdateSESConfigDefault()
				if err != nil && !util.IsErrNoRows(err) {
					impl.logger.Errorw("err while updating ses config", "err", err)
					return []int{}, err
				}
			} else {
				_, err := impl.sesRepository.FindDefault()
				if err != nil && !util.IsErrNoRows(err) {
					impl.logger.Errorw("err while updating ses config", "err", err)
					return []int{}, err
				} else if util.IsErrNoRows(err) {
					config.Default = true
				}
			}

			model, err := impl.sesRepository.FindOne(config.Id)
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("err while fetching ses config", "err", err)
				return []int{}, err
			}
			impl.buildConfigUpdateModel(config, model, userId)
			model, uErr := impl.sesRepository.UpdateSESConfig(model)
			if uErr != nil {
				impl.logger.Errorw("err while updating ses config", "err", err)
				return []int{}, uErr
			}
		} else {

			if config.Default {
				_, err := impl.sesRepository.UpdateSESConfigDefault()
				if err != nil && !util.IsErrNoRows(err) {
					impl.logger.Errorw("err while updating ses config", "err", err)
					return []int{}, err
				}
			} else {
				_, err := impl.sesRepository.FindDefault()
				if err != nil && !util.IsErrNoRows(err) {
					impl.logger.Errorw("err while updating ses config", "err", err)
					return []int{}, err
				} else if util.IsErrNoRows(err) {
					config.Default = true
				}
			}

			_, iErr := impl.sesRepository.SaveSESConfig(config)
			if iErr != nil {
				impl.logger.Errorw("err while inserting ses config", "err", iErr)
				return []int{}, iErr
			}
		}
		responseIds = append(responseIds, config.Id)
	}
	return responseIds, nil
}

func (impl *SESNotificationServiceImpl) FetchSESNotificationConfigById(id int) (*SESConfigDto, error) {
	sesConfig, err := impl.sesRepository.FindOne(id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return nil, err
	}
	sesConfigDto := impl.adaptSESConfig(sesConfig)
	return sesConfigDto, nil
}

func (impl *SESNotificationServiceImpl) FetchAllSESNotificationConfig() ([]*SESConfigDto, error) {
	var responseDto []*SESConfigDto
	sesConfigs, err := impl.sesRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return []*SESConfigDto{}, err
	}
	for _, sesConfig := range sesConfigs {
		sesConfigDto := impl.adaptSESConfig(sesConfig)
		sesConfigDto.SecretKey = "**********"
		responseDto = append(responseDto, sesConfigDto)
	}
	if responseDto == nil {
		responseDto = make([]*SESConfigDto, 0)
	}
	return responseDto, nil
}

func (impl *SESNotificationServiceImpl) FetchAllSESNotificationConfigAutocomplete() ([]*NotificationChannelAutoResponse, error) {
	var responseDto []*NotificationChannelAutoResponse
	sesConfigs, err := impl.sesRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return []*NotificationChannelAutoResponse{}, err
	}
	for _, sesConfig := range sesConfigs {
		sesConfigDto := &NotificationChannelAutoResponse{
			Id:         sesConfig.Id,
			ConfigName: sesConfig.ConfigName}
		responseDto = append(responseDto, sesConfigDto)
	}
	return responseDto, nil
}

func (impl *SESNotificationServiceImpl) adaptSESConfig(sesConfig *repository.SESConfig) *SESConfigDto {
	sesConfigDto := &SESConfigDto{
		OwnerId:      sesConfig.OwnerId,
		Region:       sesConfig.Region,
		AccessKey:    sesConfig.AccessKey,
		SecretKey:    sesConfig.SecretKey,
		FromEmail:    sesConfig.FromEmail,
		SessionToken: sesConfig.SessionToken,
		ConfigName:   sesConfig.ConfigName,
		Description:  sesConfig.Description,
		Id:           sesConfig.Id,
		Default:      sesConfig.Default,
	}
	return sesConfigDto
}

func buildSESNewConfigs(sesReq []*SESConfigDto, userId int32) []*repository.SESConfig {
	var sesConfigs []*repository.SESConfig
	for _, c := range sesReq {
		sesConfig := &repository.SESConfig{
			Id:           c.Id,
			Region:       c.Region,
			AccessKey:    c.AccessKey,
			SecretKey:    c.SecretKey,
			ConfigName:   c.ConfigName,
			FromEmail:    c.FromEmail,
			SessionToken: c.SessionToken,
			Description:  c.Description,
			Default:      c.Default,
			AuditLog: sql.AuditLog{
				CreatedBy: userId,
				CreatedOn: time.Now(),
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}

		sesConfig.OwnerId = userId
		sesConfigs = append(sesConfigs, sesConfig)
	}
	return sesConfigs
}

func (impl *SESNotificationServiceImpl) buildConfigUpdateModel(sesConfig *repository.SESConfig, model *repository.SESConfig, userId int32) {
	model.Id = sesConfig.Id
	model.OwnerId = sesConfig.OwnerId
	model.Region = sesConfig.Region
	model.AccessKey = sesConfig.AccessKey
	model.SecretKey = sesConfig.SecretKey
	model.FromEmail = sesConfig.FromEmail
	model.SessionToken = sesConfig.SessionToken
	model.ConfigName = sesConfig.ConfigName
	model.Description = sesConfig.Description
	model.Default = sesConfig.Default
	model.UpdatedOn = time.Now()
	model.UpdatedBy = userId
}

func (impl *SESNotificationServiceImpl) DeleteNotificationConfig(deleteReq *SESConfigDto, userId int32) error {
	existingConfig, err := impl.sesRepository.FindOne(deleteReq.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete", "err", err, "id", deleteReq.Id)
		return err
	}
	notifications, err := impl.notificationSettingsRepository.FindNotificationSettingsByConfigIdAndConfigType(deleteReq.Id, SES_CONFIG_TYPE)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in deleting ses config", "config", deleteReq)
		return err
	}
	if len(notifications) > 0 {
		impl.logger.Errorw("found notifications using this config, cannot delete", "config", deleteReq)
		return fmt.Errorf(" Please delete all notifications using this config before deleting")
	}
	existingConfig.UpdatedOn = time.Now()
	existingConfig.UpdatedBy = userId
	//deleting slack config
	err = impl.sesRepository.MarkSESConfigDeleted(existingConfig)
	if err != nil {
		impl.logger.Errorw("error in deleting ses config", "err", err, "id", existingConfig.Id)
		return err
	}
	return nil
}
