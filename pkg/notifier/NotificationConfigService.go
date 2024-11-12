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
	clusterService "github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/notifier/beans"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/team/read"
	util3 "github.com/devtron-labs/devtron/util"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	repository4 "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	repository3 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	util "github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type NotificationConfigService interface {
	CreateOrUpdateNotificationSettings(notificationSettingsRequest *beans.NotificationRequest, userId int32) (int, error)
	FindAll(offset int, size int) ([]*repository.NotificationSettingsViewWithAppEnv, int, error)
	BuildNotificationSettingsResponse(notificationSettings []*repository.NotificationSettingsViewWithAppEnv) ([]*beans.NotificationSettingsResponse, int, error)
	DeleteNotificationSettings(request beans.NSDeleteRequest) error
	FindNotificationSettingOptions(request *repository.SearchRequest) ([]*beans.SearchFilterResponse, error)

	UpdateNotificationSettings(notificationSettingsRequest *beans.NotificationUpdateRequest, userId int32) (int, error)
	FetchNSViewByIds(ids []*int) ([]*beans.NSConfig, error)
}

type NotificationConfigServiceImpl struct {
	logger                         *zap.SugaredLogger
	notificationConfigBuilder      NotificationConfigBuilder
	notificationSettingsRepository repository.NotificationSettingsRepository
	ciPipelineRepository           pipelineConfig.CiPipelineRepository
	pipelineRepository             pipelineConfig.PipelineRepository
	slackRepository                repository.SlackNotificationRepository
	webhookRepository              repository.WebhookNotificationRepository
	sesRepository                  repository.SESNotificationRepository
	smtpRepository                 repository.SMTPNotificationRepository
	environmentRepository          repository3.EnvironmentRepository
	clusterService                 clusterService.ClusterService
	appRepository                  app.AppRepository
	userRepository                 repository4.UserRepository
	ciPipelineMaterialRepository   pipelineConfig.CiPipelineMaterialRepository
	teamReadService                read.TeamReadService
}

const allNonProdEnvsName = "All non-prod environments"
const allProdEnvsName = "All prod environments"

func NewNotificationConfigServiceImpl(logger *zap.SugaredLogger, notificationSettingsRepository repository.NotificationSettingsRepository, notificationConfigBuilder NotificationConfigBuilder, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	pipelineRepository pipelineConfig.PipelineRepository, slackRepository repository.SlackNotificationRepository, webhookRepository repository.WebhookNotificationRepository,
	sesRepository repository.SESNotificationRepository, smtpRepository repository.SMTPNotificationRepository,
	environmentRepository repository3.EnvironmentRepository, appRepository app.AppRepository, clusterService clusterService.ClusterService,
	userRepository repository4.UserRepository, ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	teamReadService read.TeamReadService) *NotificationConfigServiceImpl {
	return &NotificationConfigServiceImpl{
		logger:                         logger,
		notificationSettingsRepository: notificationSettingsRepository,
		notificationConfigBuilder:      notificationConfigBuilder,
		pipelineRepository:             pipelineRepository,
		ciPipelineRepository:           ciPipelineRepository,
		sesRepository:                  sesRepository,
		slackRepository:                slackRepository,
		webhookRepository:              webhookRepository,
		smtpRepository:                 smtpRepository,
		environmentRepository:          environmentRepository,
		appRepository:                  appRepository,
		userRepository:                 userRepository,
		ciPipelineMaterialRepository:   ciPipelineMaterialRepository,
		clusterService:                 clusterService,
		teamReadService:                teamReadService,
	}
}

func (impl *NotificationConfigServiceImpl) FindAll(offset int, size int) ([]*repository.NotificationSettingsViewWithAppEnv, int, error) {
	var notificationSettingsViews []*repository.NotificationSettingsViewWithAppEnv
	models, err := impl.notificationSettingsRepository.FindAll(offset, size)
	if err != nil && !util2.IsErrNoRows(err) {
		impl.logger.Errorw("error while fetch notifications", "err", err)
		return notificationSettingsViews, 0, err
	}
	for _, model := range models {
		notificationSettingsView := &repository.NotificationSettingsViewWithAppEnv{
			Id:     model.Id,
			Config: model.Config,
		}
		notificationSettingsViews = append(notificationSettingsViews, notificationSettingsView)
	}

	count, err := impl.notificationSettingsRepository.FindNSViewCount()
	if err != nil && !util2.IsErrNoRows(err) {
		impl.logger.Errorw("error while fetch notifications", "err", err)
		return notificationSettingsViews, 0, err
	}
	return notificationSettingsViews, count, nil
}

func (impl *NotificationConfigServiceImpl) DeleteNotificationSettings(request beans.NSDeleteRequest) error {
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	for _, id := range request.Id {
		rowsDeleted, err := impl.notificationSettingsRepository.DeleteNotificationSettingsByConfigId(*id, tx)
		if err != nil && !util2.IsErrNoRows(err) {
			impl.logger.Errorw("error while delete notifications", "err", err)
			return err
		}
		impl.logger.Debugw("deleted notificationSetting", "rows", rowsDeleted)

		rowsDeleted, err = impl.notificationSettingsRepository.DeleteNotificationSettingsViewById(*id, tx)
		if err != nil && !util2.IsErrNoRows(err) {
			impl.logger.Errorw("err", err)
			return err
		}
		impl.logger.Debugw("deleted notificationSettingView", "rows", rowsDeleted)
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

type config struct {
	AppId        int               `json:"appId"`
	EnvId        int               `json:"envId"`
	Pipelines    []int             `json:"pipelineIds"`
	PipelineType util.PipelineType `json:"pipelineType" validate:"required"`
	EventTypeIds []int             `json:"eventTypeIds" validate:"required"`
	Providers    []beans.Provider  `json:"providers" validate:"required"`
}

func (impl *NotificationConfigServiceImpl) CreateOrUpdateNotificationSettings(notificationSettingsRequest *beans.NotificationRequest, userId int32) (int, error) {
	var configId int
	var err error
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return 0, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	for _, request := range notificationSettingsRequest.NotificationConfigRequest {
		if request.Id != 0 {
			_, err := impl.notificationSettingsRepository.DeleteNotificationSettingsByConfigId(request.Id, tx)
			if err != nil {
				impl.logger.Errorw("error while create notifications settings", "err", err)
				return 0, err
			}
		}
		request.Providers = notificationSettingsRequest.Providers
		configId, err = impl.saveNotificationSetting(request, userId, tx)
		if err != nil {
			impl.logger.Errorw("failed to save notification settings", "err", err)
			return 0, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return configId, nil
}

func (impl *NotificationConfigServiceImpl) UpdateNotificationSettings(notificationSettingsRequest *beans.NotificationUpdateRequest, userId int32) (int, error) {
	var configId int
	var err error
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return 0, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	for _, item := range notificationSettingsRequest.NotificationConfigRequest {
		configId, err = impl.updateNotificationSetting(item, notificationSettingsRequest.UpdateType, userId, tx)
		if err != nil {
			impl.logger.Errorw("failed to save notification settings", "err", err)
			return 0, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return configId, nil
}

func (impl *NotificationConfigServiceImpl) BuildNotificationSettingsResponse(notificationSettingViews []*repository.NotificationSettingsViewWithAppEnv) ([]*beans.NotificationSettingsResponse, int, error) {
	var notificationSettingsResponses []*beans.NotificationSettingsResponse
	deletedItemCount := 0
	for _, ns := range notificationSettingViews {
		notificationSettingsResponse := &beans.NotificationSettingsResponse{
			Id:         ns.Id,
			ConfigName: ns.ConfigName,
		}

		var config beans.NSConfig
		configJson := []byte(ns.Config)
		err := json.Unmarshal(configJson, &config)
		if err != nil {
			impl.logger.Errorw("unmarshal error", "err", err)
			return notificationSettingsResponses, deletedItemCount, err
		}

		if config.TeamId != nil && len(config.TeamId) > 0 {
			teams, err := impl.teamReadService.FindByIds(config.TeamId)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error in fetching team", "err", err)
				return notificationSettingsResponses, deletedItemCount, err
			}
			var teamResponse []*beans.TeamResponse
			for _, item := range teams {
				teamResponse = append(teamResponse, &beans.TeamResponse{Id: &item.Id, Name: item.Name})
			}
			notificationSettingsResponse.TeamResponse = teamResponse
		}

		if config.EnvId != nil && len(config.EnvId) > 0 {
			var envResponse []*beans.EnvResponse
			envIds := make([]*int, 0)
			for _, envId := range config.EnvId {
				if *envId == resourceQualifiers.AllExistingAndFutureProdEnvsInt {
					envResponse = append(envResponse, &beans.EnvResponse{Id: envId, Name: allProdEnvsName})
				} else if *envId == resourceQualifiers.AllExistingAndFutureNonProdEnvsInt {
					envResponse = append(envResponse, &beans.EnvResponse{Id: envId, Name: allNonProdEnvsName})
				} else {
					envIds = append(envIds, envId)
				}
			}
			if len(envIds) > 0 {
				environments, err := impl.environmentRepository.FindByIds(config.EnvId)
				if err != nil && err != pg.ErrNoRows {
					impl.logger.Errorw("error in fetching env", "envIds", config.EnvId, "err", err)
					return notificationSettingsResponses, deletedItemCount, err
				}
				for _, item := range environments {
					envResponse = append(envResponse, &beans.EnvResponse{Id: &item.Id, Name: item.Name})
				}
			}

			notificationSettingsResponse.EnvResponse = envResponse
		}

		if config.AppId != nil && len(config.AppId) > 0 {
			applications, err := impl.appRepository.FindByIds(config.AppId)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error in fetching app", "appIds", config.AppId, "err", err)
				return notificationSettingsResponses, deletedItemCount, err
			}
			var appResponse []*beans.AppResponse
			for _, item := range applications {
				appResponse = append(appResponse, &beans.AppResponse{Id: &item.Id, Name: item.AppName})
			}
			notificationSettingsResponse.AppResponse = appResponse
		}

		var clusterResponse []*beans.ClusterResponse
		if len(config.ClusterId) > 0 {
			clusterIds := util3.Transform(config.ClusterId, func(id *int) int {
				return *id
			})
			clusterName, err := impl.clusterService.FindByIds(clusterIds)
			if err != nil {
				impl.logger.Errorw("error on fetching cluster", "clusterIds", clusterIds, "err", err)
				return notificationSettingsResponses, deletedItemCount, err
			}
			for _, item := range clusterName {
				clusterResponse = append(clusterResponse, &beans.ClusterResponse{Id: &item.Id, Name: item.ClusterName})
			}
			notificationSettingsResponse.ClusterResponse = clusterResponse
		}

		if config.Providers != nil && len(config.Providers) > 0 {
			var slackIds []*int
			var webhookIds []*int
			var sesUserIds []int32
			var smtpUserIds []int32
			var providerConfigs []*beans.ProvidersConfig
			for _, item := range config.Providers {
				// if item.ConfigId > 0 that means, user is of user repository, else user email is custom
				if item.ConfigId > 0 {
					if item.Destination == util.Slack {
						slackIds = append(slackIds, &item.ConfigId)
					} else if item.Destination == util.SES {
						sesUserIds = append(sesUserIds, int32(item.ConfigId))
					} else if item.Destination == util.SMTP {
						smtpUserIds = append(smtpUserIds, int32(item.ConfigId))
					} else if item.Destination == util.Webhook {
						webhookIds = append(webhookIds, &item.ConfigId)
					}
				} else {
					providerConfigs = append(providerConfigs, &beans.ProvidersConfig{Dest: string(item.Destination), Recipient: item.Recipient})
				}
			}
			if len(slackIds) > 0 {
				slackConfigs, err := impl.slackRepository.FindByIds(slackIds)
				if err != nil && err != pg.ErrNoRows {
					impl.logger.Errorw("error in fetching slack config", "slackIds", slackIds, "err", err)
					return notificationSettingsResponses, deletedItemCount, err
				}
				for _, item := range slackConfigs {
					providerConfigs = append(providerConfigs, &beans.ProvidersConfig{Id: item.Id, ConfigName: item.ConfigName, Dest: string(util.Slack)})
				}
			}
			if len(webhookIds) > 0 {
				webhookConfigs, err := impl.webhookRepository.FindByIds(webhookIds)
				if err != nil && err != pg.ErrNoRows {
					impl.logger.Errorw("error in fetching webhook config", "webhookIds", webhookIds, "err", err)
					return notificationSettingsResponses, deletedItemCount, err
				}
				for _, item := range webhookConfigs {
					providerConfigs = append(providerConfigs, &beans.ProvidersConfig{Id: item.Id, ConfigName: item.ConfigName, Dest: string(util.Webhook)})
				}
			}

			if len(sesUserIds) > 0 {
				sesConfigs, err := impl.userRepository.GetByIds(sesUserIds)
				if err != nil && err != pg.ErrNoRows {
					impl.logger.Errorw("error in fetching user", "sesUserIds", sesUserIds, "error", err)
					return notificationSettingsResponses, deletedItemCount, err
				}
				for _, item := range sesConfigs {
					providerConfigs = append(providerConfigs, &beans.ProvidersConfig{Id: int(item.Id), ConfigName: item.EmailId, Dest: string(util.SES)})
				}
			}
			if len(smtpUserIds) > 0 {
				smtpConfigs, err := impl.userRepository.GetByIds(smtpUserIds)
				if err != nil && err != pg.ErrNoRows {
					impl.logger.Errorw("error in fetching user", "smtpUserIds", smtpUserIds, "error", err)
					return notificationSettingsResponses, deletedItemCount, err
				}
				for _, item := range smtpConfigs {
					providerConfigs = append(providerConfigs, &beans.ProvidersConfig{Id: int(item.Id), ConfigName: item.EmailId, Dest: string(util.SMTP)})
				}
			}
			notificationSettingsResponse.ProvidersConfig = providerConfigs
		}

		if config.PipelineId != nil && *config.PipelineId > 0 {
			pipelineResponse := &beans.PipelineResponse{}
			pipelineResponse.Id = config.PipelineId
			if config.PipelineType == util.CD {
				pipeline, err := impl.pipelineRepository.FindById(*config.PipelineId)
				if err != nil && err != pg.ErrNoRows {
					impl.logger.Errorw("error in fetching cd pipeline", "err", err)
					return notificationSettingsResponses, deletedItemCount, err
				}
				if err == pg.ErrNoRows {
					deletedItemCount = deletedItemCount + 1
					continue
				}
				pipelineResponse.EnvironmentName = pipeline.Environment.Name
				pipelineResponse.Name = pipeline.Name
				if pipeline.App.Id > 0 {
					pipelineResponse.AppName = pipeline.App.AppName
				}
			} else if config.PipelineType == util.CI {
				pipeline, err := impl.ciPipelineRepository.FindById(*config.PipelineId)
				if err != nil && err != pg.ErrNoRows {
					impl.logger.Errorw("error in fetching ci pipeline", "err", err)
					return notificationSettingsResponses, deletedItemCount, err
				}
				if err == pg.ErrNoRows {
					deletedItemCount = deletedItemCount + 1
					continue
				}
				pipelineResponse.Name = pipeline.Name
				if pipeline.App != nil {
					pipelineResponse.AppName = pipeline.App.AppName
				}
				if pipeline.CiPipelineMaterials != nil {
					for _, item := range pipeline.CiPipelineMaterials {
						pipelineResponse.Branches = append(pipelineResponse.Branches, item.Value)
					}
				}
			}
			notificationSettingsResponse.PipelineResponse = pipelineResponse
		}

		notificationSettingsResponse.PipelineType = string(config.PipelineType)
		notificationSettingsResponse.EventTypes = config.EventTypeIds

		notificationSettingsResponses = append(notificationSettingsResponses, notificationSettingsResponse)
	}
	return notificationSettingsResponses, deletedItemCount, nil
}

func (impl *NotificationConfigServiceImpl) buildProvidersConfig(config config) ([]beans.ProvidersConfig, error) {
	var providerConfigs []beans.ProvidersConfig
	if len(config.Providers) > 0 {
		sesConfigNamesMap := map[int]string{}
		slackConfigNameMap := map[int]string{}
		smtpConfigNamesMap := map[int]string{}
		for _, c := range config.Providers {
			if util.Slack == c.Destination {
				if _, ok := slackConfigNameMap[c.ConfigId]; ok {
					continue
				}
				slackConfigNameMap[c.ConfigId] = ""
			} else if util.SES == c.Destination {
				if _, ok := sesConfigNamesMap[c.ConfigId]; ok {
					continue
				}
				sesConfigNamesMap[c.ConfigId] = ""
			} else if util.SMTP == c.Destination {
				if _, ok := smtpConfigNamesMap[c.ConfigId]; ok {
					continue
				}
				smtpConfigNamesMap[c.ConfigId] = ""
			}
		}

		slackIds := make([]int, 0, len(slackConfigNameMap))
		sesIds := make([]int, 0, len(sesConfigNamesMap))
		smtpIds := make([]int, 0, len(smtpConfigNamesMap))

		for k := range slackConfigNameMap {
			slackIds = append(slackIds, k)
		}
		for k := range sesConfigNamesMap {
			sesIds = append(sesIds, k)
		}
		for k := range smtpConfigNamesMap {
			smtpIds = append(smtpIds, k)
		}

		if len(slackIds) > 0 {
			slackConfigs, err := impl.slackRepository.FindByIdsIn(slackIds)
			if err != nil {
				impl.logger.Errorw("error in fetch slack configs", "err", err)
				return []beans.ProvidersConfig{}, err
			}
			for _, s := range slackConfigs {
				slackConfigNameMap[s.Id] = s.ConfigName
			}
		}
		if len(sesIds) > 0 {
			sesConfigs, err := impl.sesRepository.FindByIdsIn(sesIds)
			if err != nil {
				impl.logger.Errorw("error on fetch ses configs", "err", err)
				return []beans.ProvidersConfig{}, err
			}
			for _, s := range sesConfigs {
				sesConfigNamesMap[s.Id] = s.ConfigName
			}
		}
		if len(smtpIds) > 0 {
			smtpConfigs, err := impl.smtpRepository.FindByIdsIn(sesIds)
			if err != nil {
				impl.logger.Errorw("error on fetch smtp configs", "err", err)
				return []beans.ProvidersConfig{}, err
			}
			for _, s := range smtpConfigs {
				smtpConfigNamesMap[s.Id] = s.ConfigName
			}
		}
		for _, c := range config.Providers {
			var configName string
			if c.Destination == util.Slack {
				configName = slackConfigNameMap[c.ConfigId]
			} else if c.Destination == util.SES {
				configName = sesConfigNamesMap[c.ConfigId]
			} else if c.Destination == util.SMTP {
				configName = smtpConfigNamesMap[c.ConfigId]
			}
			providerConfig := beans.ProvidersConfig{
				Id:         c.ConfigId,
				Dest:       string(c.Destination),
				ConfigName: configName,
			}
			providerConfigs = append(providerConfigs, providerConfig)
		}
	}
	return providerConfigs, nil
}

func (impl *NotificationConfigServiceImpl) buildPipelineResponses(config config, ciPipelines []*pipelineConfig.CiPipeline, cdPipelines []*pipelineConfig.Pipeline) (util.PipelineType, []beans.PipelineResponse, error) {
	var pipelineType util.PipelineType
	var err error

	if util.CI == config.PipelineType {
		pipelineType = util.CI
	} else if util.CD == config.PipelineType {
		pipelineType = util.CD
	}

	if len(config.Pipelines) > 0 {
		var pipelinesIds []int
		pipelinesIds = append(pipelinesIds, config.Pipelines...)
		if util.CI == config.PipelineType {
			ciPipelines, err = impl.ciPipelineRepository.FindByIdsIn(pipelinesIds)
		} else if util.CD == config.PipelineType {
			cdPipelines, err = impl.pipelineRepository.FindByIdsIn(pipelinesIds)
		}
		if err != nil {
			impl.logger.Errorw("error while response build", "err", err)
			return "", []beans.PipelineResponse{}, err
		}
	}
	var pipelineResponses []beans.PipelineResponse
	if util.CI == pipelineType {
		for _, ci := range ciPipelines {
			pipelineResponse := beans.PipelineResponse{
				Id:   &ci.Id,
				Name: ci.Name,
			}
			pipelineResponses = append(pipelineResponses, pipelineResponse)
		}
	} else if util.CD == pipelineType {
		for _, cd := range cdPipelines {
			pipelineResponse := beans.PipelineResponse{
				Id:   &cd.Id,
				Name: cd.Name,
			}
			pipelineResponses = append(pipelineResponses, pipelineResponse)
		}
	}
	return pipelineType, pipelineResponses, nil
}

func (impl *NotificationConfigServiceImpl) saveNotificationSetting(notificationSettingsRequest *beans.NotificationConfigRequest, userId int32, tx *pg.Tx) (int, error) {
	var existingNotificationSettingsConfig *repository.NotificationSettingsView
	var err error
	if notificationSettingsRequest.Id != 0 {
		existingNotificationSettingsConfig, err = impl.notificationSettingsRepository.FindNotificationSettingsViewById(notificationSettingsRequest.Id)
		if err != nil {
			impl.logger.Errorw("failed to fetch existing notification settings view", "err", err)
			return 0, err
		}
	}
	notificationSettingsConfig, err := impl.notificationConfigBuilder.BuildNotificationSettingsConfig(notificationSettingsRequest, existingNotificationSettingsConfig, userId)
	if err != nil {
		impl.logger.Errorw("failed to build notification settings view", "err", err)
		return 0, err
	}
	if notificationSettingsConfig.Id == 0 {
		notificationSettingsConfig, err = impl.notificationSettingsRepository.SaveNotificationSettingsConfig(notificationSettingsConfig, tx)
	} else {
		notificationSettingsConfig, err = impl.notificationSettingsRepository.UpdateNotificationSettingsView(notificationSettingsConfig, tx)
	}
	if err != nil {
		impl.logger.Errorw("failed to save notification settings view", "err", err)
		return 0, err
	}
	notificationSettings, nErr := impl.notificationConfigBuilder.BuildNewNotificationSettings(notificationSettingsRequest, notificationSettingsConfig)
	if nErr != nil {
		impl.logger.Errorw("failed to build notification settings", "err", err)
		return 0, nErr
	}
	_, sErr := impl.notificationSettingsRepository.SaveAllNotificationSettings(notificationSettings, tx)
	if sErr != nil {
		impl.logger.Errorw("failed to save notification settings", "err", err)
		_, err = impl.notificationSettingsRepository.DeleteNotificationSettingsViewById(notificationSettingsConfig.Id, tx)
		if err != nil {
			impl.logger.Errorw("failed to rollback notification settings view", "err", err)
			return 0, err
		}
		return 0, sErr
	}
	impl.logger.Debug("notification settings saved")
	return notificationSettingsConfig.Id, nil
}

func (impl *NotificationConfigServiceImpl) updateNotificationSetting(notificationSettingsRequest *beans.NotificationConfigRequest, updateType util.UpdateType, userId int32, tx *pg.Tx) (int, error) {
	var err error
	existingNotificationSettingsConfig, err := impl.notificationSettingsRepository.FindNotificationSettingsViewById(notificationSettingsRequest.Id)
	if err != nil {
		impl.logger.Errorw("failed to fetch existing notification settings view", "err", err)
		return 0, err
	}

	nsConfig := &beans.NSConfig{}
	err = json.Unmarshal([]byte(existingNotificationSettingsConfig.Config), nsConfig)
	if updateType == util.UpdateEvents {
		nsConfig.EventTypeIds = notificationSettingsRequest.EventTypeIds
	} else if updateType == util.UpdateRecipients {
		nsConfig.Providers = notificationSettingsRequest.Providers
	}
	config, err := json.Marshal(nsConfig)
	if err != nil {
		impl.logger.Error(err)
		return 0, err
	}
	existingNotificationSettingsConfig.Config = string(config)
	currentTime := time.Now()
	existingNotificationSettingsConfig.UpdatedOn = currentTime
	existingNotificationSettingsConfig.UpdatedBy = userId
	existingNotificationSettingsConfig, err = impl.notificationSettingsRepository.UpdateNotificationSettingsView(existingNotificationSettingsConfig, tx)
	if err != nil {
		impl.logger.Errorw("failed to save notification settings view", "err", err)
		return 0, err
	}

	if updateType == util.UpdateEvents {
		notificationSettingsRequest.TeamId = nsConfig.TeamId
		notificationSettingsRequest.AppId = nsConfig.AppId
		notificationSettingsRequest.EnvId = nsConfig.EnvId
		notificationSettingsRequest.PipelineId = nsConfig.PipelineId
		notificationSettingsRequest.PipelineType = nsConfig.PipelineType
		notificationSettingsRequest.Providers = nsConfig.Providers
		var notificationSettings []repository.NotificationSettings
		nsOptions, err := impl.notificationSettingsRepository.FetchNotificationSettingGroupBy(notificationSettingsRequest.Id)
		if err != nil {
			impl.logger.Errorw("failed to fetch existing notification settings view", "err", err)
			return 0, err
		}
		if len(nsOptions) == 0 {
			notificationSettings, err = impl.notificationConfigBuilder.BuildNewNotificationSettings(notificationSettingsRequest, existingNotificationSettingsConfig)
			if err != nil {
				impl.logger.Error(err)
				return 0, err
			}
		} else {
			for _, item := range nsOptions {
				for _, e := range notificationSettingsRequest.EventTypeIds {
					notificationSetting, err := impl.notificationConfigBuilder.BuildNotificationSettingWithPipeline(item.TeamId, item.EnvId, item.AppId, item.PipelineId, item.ClusterId, util.PipelineType(item.PipelineType), e, notificationSettingsRequest.Id, nsConfig.Providers)
					if err != nil {
						impl.logger.Error(err)
						return 0, err
					}
					notificationSettings = append(notificationSettings, notificationSetting)
				}
			}
		}
		//deleting old items
		_, err = impl.notificationSettingsRepository.DeleteNotificationSettingsByConfigId(notificationSettingsRequest.Id, tx)
		if err != nil {
			impl.logger.Errorw("error on delete ns", "err", err)
			return 0, err
		}

		if notificationSettings != nil {
			_, sErr := impl.notificationSettingsRepository.SaveAllNotificationSettings(notificationSettings, tx)
			if sErr != nil {
				impl.logger.Errorw("failed to save notification settings", "err", err)
				_, err = impl.notificationSettingsRepository.DeleteNotificationSettingsViewById(existingNotificationSettingsConfig.Id, tx)
				if err != nil {
					impl.logger.Errorw("failed to rollback notification settings view", "err", err)
					return 0, err
				}
				return 0, sErr
			}
		}
	} else if updateType == util.UpdateRecipients {
		nsOptions, err := impl.notificationSettingsRepository.FindNotificationSettingsByViewId(notificationSettingsRequest.Id)
		if err != nil {
			impl.logger.Errorw("failed to fetch existing notification settings view", "err", err)
			return 0, err
		}

		//UPDATE - config updated, MAY BE THERE IS NO NEED OF UPDATE HERE

		for _, ns := range nsOptions {
			config, err := json.Marshal(nsConfig.Providers)
			if err != nil {
				impl.logger.Error(err)
				return 0, err
			}
			ns.Config = string(config)
			_, err = impl.notificationSettingsRepository.UpdateNotificationSettings(&ns, tx)
			if err != nil {
				impl.logger.Errorw("failed to fetch existing notification settings view", "err", err)
				return 0, err
			}
		}
	}
	return existingNotificationSettingsConfig.Id, nil
}

func (impl *NotificationConfigServiceImpl) FindNotificationSettingOptions(settingRequest *repository.SearchRequest) ([]*beans.SearchFilterResponse, error) {
	var searchFilterResponse []*beans.SearchFilterResponse

	prodEnvIdentifierFound := false
	nonProdEnvIdentifierFound := false
	for _, envId := range settingRequest.EnvId {
		if *envId == resourceQualifiers.AllExistingAndFutureProdEnvsInt {
			prodEnvIdentifierFound = true
		}
	}

	for _, envId := range settingRequest.EnvId {
		if *envId == resourceQualifiers.AllExistingAndFutureNonProdEnvsInt {
			nonProdEnvIdentifierFound = true
		}
	}

	if prodEnvIdentifierFound && nonProdEnvIdentifierFound {
		return searchFilterResponse, errors.New("cannot specify both prod and non-prod environment filter")
	}

	settingOptionResponseDeployment, err := impl.notificationSettingsRepository.FindNotificationSettingDeploymentOptions(settingRequest)
	if err != nil && !util2.IsErrNoRows(err) {
		impl.logger.Errorw("error while fetching notification deployment option", "err", err)
		return searchFilterResponse, err
	}
	for _, item := range settingOptionResponseDeployment {
		item.PipelineType = string(util.CD)
		result := &beans.SearchFilterResponse{
			PipelineType:     item.PipelineType,
			PipelineResponse: &beans.PipelineResponse{Id: &item.PipelineId, Name: item.PipelineName, EnvironmentName: item.EnvironmentName, AppName: item.AppName, ClusterName: item.ClusterName},
		}
		searchFilterResponse = append(searchFilterResponse, result)
	}

	settingOptionResponseBuild, err := impl.notificationSettingsRepository.FindNotificationSettingBuildOptions(settingRequest)
	if err != nil && !util2.IsErrNoRows(err) {
		impl.logger.Errorw("error while fetching notification deployment option", "err", err)
		return searchFilterResponse, err
	}
	for _, item := range settingOptionResponseBuild {
		item.PipelineType = string(util.CI)

		pipelineMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(item.PipelineId)
		if err != nil && !util2.IsErrNoRows(err) {
			impl.logger.Errorw("error while fetching material", "err", err)
			return searchFilterResponse, err
		}
		var branches []string
		if pipelineMaterials != nil {
			for _, pipelineMaterial := range pipelineMaterials {
				branches = append(branches, pipelineMaterial.Value)
			}
		}
		result := &beans.SearchFilterResponse{
			PipelineType:     item.PipelineType,
			PipelineResponse: &beans.PipelineResponse{Id: &item.PipelineId, Name: item.PipelineName, AppName: item.AppName, Branches: branches},
		}
		searchFilterResponse = append(searchFilterResponse, result)
	}

	var teamResponse []*beans.TeamResponse
	var appResponse []*beans.AppResponse
	var envResponse []*beans.EnvResponse
	var clusterResponse []*beans.ClusterResponse
	if settingRequest.TeamId != nil && len(settingRequest.TeamId) > 0 {
		teams, err := impl.teamReadService.FindByIds(settingRequest.TeamId)
		if err != nil {
			impl.logger.Errorw("error on fetch teams", "err", err)
			return searchFilterResponse, err
		}
		for _, item := range teams {
			teamResponse = append(teamResponse, &beans.TeamResponse{Id: &item.Id, Name: item.Name})
		}
	}
	if settingRequest.EnvId != nil && len(settingRequest.EnvId) > 0 {
		envIds := make([]*int, 0)
		for _, envId := range settingRequest.EnvId {
			if *envId == resourceQualifiers.AllExistingAndFutureProdEnvsInt {
				envResponse = append(envResponse, &beans.EnvResponse{Id: envId, Name: allProdEnvsName})
			} else if *envId == resourceQualifiers.AllExistingAndFutureNonProdEnvsInt {
				envResponse = append(envResponse, &beans.EnvResponse{Id: envId, Name: allNonProdEnvsName})
			} else {
				envIds = append(envIds, envId)
			}
		}
		if len(envIds) > 0 {
			environments, err := impl.environmentRepository.FindByIds(settingRequest.EnvId)
			if err != nil {
				impl.logger.Errorw("error on fetching environments", "envIds", settingRequest.EnvId, "err", err)
				return searchFilterResponse, err
			}
			for _, item := range environments {
				envResponse = append(envResponse, &beans.EnvResponse{Id: &item.Id, Name: item.Name})
			}
		}
	}
	if settingRequest.AppId != nil && len(settingRequest.AppId) > 0 {
		applications, err := impl.appRepository.FindByIds(settingRequest.AppId)
		if err != nil {
			impl.logger.Errorw("error on fetching apps", "appIds", settingRequest.AppId, "err", err)
			return searchFilterResponse, err
		}
		for _, item := range applications {
			appResponse = append(appResponse, &beans.AppResponse{Id: &item.Id, Name: item.AppName})
		}
	}

	if len(settingRequest.ClusterId) > 0 {
		clusterIds := util3.Transform(settingRequest.ClusterId, func(id *int) int {
			return *id
		})
		clusterName, err := impl.clusterService.FindByIds(clusterIds)
		if err != nil {
			impl.logger.Errorw("error on fetching cluster", "clusterIds", clusterIds, "err", err)
			return searchFilterResponse, err
		}
		for _, item := range clusterName {
			clusterResponse = append(clusterResponse, &beans.ClusterResponse{Id: &item.Id, Name: item.ClusterName})
		}
	}

	if teamResponse != nil || appResponse != nil {
		ciMatching := &beans.SearchFilterResponse{
			PipelineType: string(util.CI),
			TeamResponse: teamResponse,
			AppResponse:  appResponse,
		}
		searchFilterResponse = append(searchFilterResponse, ciMatching)
	}

	if teamResponse != nil || appResponse != nil || envResponse != nil || clusterResponse != nil {
		cdMatching := &beans.SearchFilterResponse{
			PipelineType:    string(util.CD),
			TeamResponse:    teamResponse,
			AppResponse:     appResponse,
			EnvResponse:     envResponse,
			ClusterResponse: clusterResponse,
		}
		searchFilterResponse = append(searchFilterResponse, cdMatching)
	}

	if searchFilterResponse == nil {
		searchFilterResponse = make([]*beans.SearchFilterResponse, 0)
	}

	return searchFilterResponse, nil
}

func (impl *NotificationConfigServiceImpl) FetchNSViewByIds(ids []*int) ([]*beans.NSConfig, error) {
	var configs []*beans.NSConfig
	notificationSettingViewList, err := impl.notificationSettingsRepository.FindNotificationSettingsViewByIds(ids)
	if err != nil {
		impl.logger.Errorw("failed to fetch existing notification settings view", "err", err)
		return configs, err
	}
	for _, item := range notificationSettingViewList {
		nsConfig := &beans.NSConfig{}
		err = json.Unmarshal([]byte(item.Config), nsConfig)
		if err != nil {
			return configs, err
		}
		configs = append(configs, nsConfig)
	}

	return configs, nil
}
