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

type SESNotificationService interface {
	SaveOrEditNotificationConfig(channelReq []*beans.SESConfigDto, userId int32) ([]int, error)
	FetchSESNotificationConfigById(id int) (*beans.SESConfigDto, error)
	FetchAllSESNotificationConfig() ([]*beans.SESConfigDto, error)
	FetchAllSESNotificationConfigAutocomplete() ([]*beans.NotificationChannelAutoResponse, error)
	DeleteNotificationConfig(channelReq *beans.SESConfigDto, userId int32) error
}

type SESNotificationServiceImpl struct {
	logger                         *zap.SugaredLogger
	teamService                    team.TeamService
	sesRepository                  repository.SESNotificationRepository
	notificationSettingsRepository repository.NotificationSettingsRepository
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

func (impl *SESNotificationServiceImpl) SaveOrEditNotificationConfig(channelReq []*beans.SESConfigDto, userId int32) ([]int, error) {
	var responseIds []int
	sesConfigs := adapter.BuildSESNewConfigs(channelReq, userId)
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
			adapter.BuildConfigUpdateModelForSES(config, model, userId)
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

func (impl *SESNotificationServiceImpl) FetchSESNotificationConfigById(id int) (*beans.SESConfigDto, error) {
	sesConfig, err := impl.sesRepository.FindOne(id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return nil, err
	}
	sesConfigDto := adapter.AdaptSESConfig(sesConfig)
	return sesConfigDto, nil
}

func (impl *SESNotificationServiceImpl) FetchAllSESNotificationConfig() ([]*beans.SESConfigDto, error) {
	var responseDto []*beans.SESConfigDto
	sesConfigs, err := impl.sesRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return []*beans.SESConfigDto{}, err
	}
	for _, sesConfig := range sesConfigs {
		sesConfigDto := adapter.AdaptSESConfig(sesConfig)
		sesConfigDto.SecretKey = "**********"
		responseDto = append(responseDto, sesConfigDto)
	}
	if responseDto == nil {
		responseDto = make([]*beans.SESConfigDto, 0)
	}
	return responseDto, nil
}

func (impl *SESNotificationServiceImpl) FetchAllSESNotificationConfigAutocomplete() ([]*beans.NotificationChannelAutoResponse, error) {
	var responseDto []*beans.NotificationChannelAutoResponse
	sesConfigs, err := impl.sesRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find all slack config", "err", err)
		return []*beans.NotificationChannelAutoResponse{}, err
	}
	for _, sesConfig := range sesConfigs {
		sesConfigDto := &beans.NotificationChannelAutoResponse{
			Id:         sesConfig.Id,
			ConfigName: sesConfig.ConfigName}
		responseDto = append(responseDto, sesConfigDto)
	}
	return responseDto, nil
}

func (impl *SESNotificationServiceImpl) DeleteNotificationConfig(deleteReq *beans.SESConfigDto, userId int32) error {
	existingConfig, err := impl.sesRepository.FindOne(deleteReq.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete", "err", err, "id", deleteReq.Id)
		return err
	}
	notifications, err := impl.notificationSettingsRepository.FindNotificationSettingsByConfigIdAndConfigType(deleteReq.Id, eventUtil.SES.String())
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in deleting ses config", "config", deleteReq)
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
	//deleting slack config
	err = impl.sesRepository.MarkSESConfigDeleted(existingConfig)
	if err != nil {
		impl.logger.Errorw("error in deleting ses config", "err", err, "id", existingConfig.Id)
		return err
	}
	return nil
}
