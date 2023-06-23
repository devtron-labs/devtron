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
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	repository3 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/team"
	repository4 "github.com/devtron-labs/devtron/pkg/user/repository"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	util "github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type NotificationConfigService interface {
	CreateOrUpdateNotificationSettings(notificationSettingsRequest *NotificationRequest, userId int32) (int, error)
	FindAll(offset int, size int) ([]*repository.NotificationSettingsViewWithAppEnv, int, error)
	BuildNotificationSettingsResponse(notificationSettings []*repository.NotificationSettingsViewWithAppEnv) ([]*NotificationSettingsResponse, int, error)
	DeleteNotificationSettings(request NSDeleteRequest) error
	FindNotificationSettingOptions(request *repository.SearchRequest) ([]*SearchFilterResponse, error)

	UpdateNotificationSettings(notificationSettingsRequest *NotificationUpdateRequest, userId int32) (int, error)
	FetchNSViewByIds(ids []*int) ([]*NSConfig, error)
}

type NotificationConfigServiceImpl struct {
	logger                         *zap.SugaredLogger
	notificationConfigBuilder      NotificationConfigBuilder
	notificationSettingsRepository repository.NotificationSettingsRepository
	ciPipelineRepository           pipelineConfig.CiPipelineRepository
	pipelineRepository             pipelineConfig.PipelineRepository
	slackRepository                repository.SlackNotificationRepository
	sesRepository                  repository.SESNotificationRepository
	smtpRepository                 repository.SMTPNotificationRepository
	teamRepository                 repository2.TeamRepository
	environmentRepository          repository3.EnvironmentRepository
	appRepository                  app.AppRepository
	userRepository                 repository4.UserRepository
	ciPipelineMaterialRepository   pipelineConfig.CiPipelineMaterialRepository
}

type NotificationSettingRequest struct {
	Id     int  `json:"id"`
	TeamId int  `json:"teamId"`
	AppId  *int `json:"appId"`
	EnvId  *int `json:"envId"`
	//Pipelines    []int             `json:"pipelineIds"`
	PipelineType util.PipelineType `json:"pipelineType" validate:"required"`
	EventTypeIds []int             `json:"eventTypeIds" validate:"required"`
	Providers    []Provider        `json:"providers" validate:"required"`
}

type Provider struct {
	Destination util.Channel `json:"dest"`
	Rule        string       `json:"rule"`
	ConfigId    int          `json:"configId"`
	Recipient   string       `json:"recipient"`
}

type Providers struct {
	Providers []Provider `json:"providers"`
}

type NSDeleteRequest struct {
	Id []*int `json:"id"`
}

type NotificationRequest struct {
	UpdateType                util.UpdateType              `json:"updateType,omitempty"`
	SesConfigId               int                          `json:"sesConfigId,omitempty"`
	Providers                 []*Provider                  `json:"providers"`
	NotificationConfigRequest []*NotificationConfigRequest `json:"notificationConfigRequest" validate:"required"`
}
type NotificationUpdateRequest struct {
	UpdateType                util.UpdateType              `json:"updateType,omitempty"`
	NotificationConfigRequest []*NotificationConfigRequest `json:"notificationConfigRequest" validate:"required"`
}
type NotificationConfigRequest struct {
	Id           int               `json:"id"`
	TeamId       []*int            `json:"teamId"`
	AppId        []*int            `json:"appId"`
	EnvId        []*int            `json:"envId"`
	PipelineId   *int              `json:"pipelineId"`
	PipelineType util.PipelineType `json:"pipelineType" validate:"required"`
	EventTypeIds []int             `json:"eventTypeIds" validate:"required"`
	Providers    []*Provider       `json:"providers"`
}

type NSViewResponse struct {
	Total                        int                             `json:"total"`
	NotificationSettingsResponse []*NotificationSettingsResponse `json:"settings"`
}

type NotificationSettingsResponse struct {
	Id               int                `json:"id"`
	ConfigName       string             `json:"configName"`
	TeamResponse     []*TeamResponse    `json:"team"`
	AppResponse      []*AppResponse     `json:"app"`
	EnvResponse      []*EnvResponse     `json:"environment"`
	PipelineResponse *PipelineResponse  `json:"pipeline"`
	PipelineType     string             `json:"pipelineType"`
	ProvidersConfig  []*ProvidersConfig `json:"providerConfigs"`
	EventTypes       []int              `json:"eventTypes"`
}

type SearchFilterResponse struct {
	TeamResponse     []*TeamResponse   `json:"team"`
	AppResponse      []*AppResponse    `json:"app"`
	EnvResponse      []*EnvResponse    `json:"environment"`
	PipelineResponse *PipelineResponse `json:"pipeline"`
	PipelineType     string            `json:"pipelineType"`
}

type TeamResponse struct {
	Id   *int   `json:"id"`
	Name string `json:"name"`
}

type AppResponse struct {
	Id   *int   `json:"id"`
	Name string `json:"name"`
}

type EnvResponse struct {
	Id   *int   `json:"id"`
	Name string `json:"name"`
}

type PipelineResponse struct {
	Id              *int     `json:"id"`
	Name            string   `json:"name"`
	EnvironmentName string   `json:"environmentName,omitempty"`
	AppName         string   `json:"appName,omitempty"`
	Branches        []string `json:"branches,omitempty"`
}

type ProvidersConfig struct {
	Id         int    `json:"id"`
	Dest       string `json:"dest"`
	ConfigName string `json:"name"`
	Recipient  string `json:"recipient"`
}

func NewNotificationConfigServiceImpl(logger *zap.SugaredLogger, notificationSettingsRepository repository.NotificationSettingsRepository, notificationConfigBuilder NotificationConfigBuilder, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	pipelineRepository pipelineConfig.PipelineRepository, slackRepository repository.SlackNotificationRepository,
	sesRepository repository.SESNotificationRepository, smtpRepository repository.SMTPNotificationRepository,
	teamRepository repository2.TeamRepository,
	environmentRepository repository3.EnvironmentRepository, appRepository app.AppRepository,
	userRepository repository4.UserRepository, ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository) *NotificationConfigServiceImpl {
	return &NotificationConfigServiceImpl{
		logger:                         logger,
		notificationSettingsRepository: notificationSettingsRepository,
		notificationConfigBuilder:      notificationConfigBuilder,
		pipelineRepository:             pipelineRepository,
		ciPipelineRepository:           ciPipelineRepository,
		sesRepository:                  sesRepository,
		slackRepository:                slackRepository,
		smtpRepository:                 smtpRepository,
		teamRepository:                 teamRepository,
		environmentRepository:          environmentRepository,
		appRepository:                  appRepository,
		userRepository:                 userRepository,
		ciPipelineMaterialRepository:   ciPipelineMaterialRepository,
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

func (impl *NotificationConfigServiceImpl) DeleteNotificationSettings(request NSDeleteRequest) error {
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
	Providers    []Provider        `json:"providers" validate:"required"`
}

func (impl *NotificationConfigServiceImpl) CreateOrUpdateNotificationSettings(notificationSettingsRequest *NotificationRequest, userId int32) (int, error) {
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

func (impl *NotificationConfigServiceImpl) UpdateNotificationSettings(notificationSettingsRequest *NotificationUpdateRequest, userId int32) (int, error) {
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

func (impl *NotificationConfigServiceImpl) BuildNotificationSettingsResponse(notificationSettingViews []*repository.NotificationSettingsViewWithAppEnv) ([]*NotificationSettingsResponse, int, error) {
	var notificationSettingsResponses []*NotificationSettingsResponse
	deletedItemCount := 0
	for _, ns := range notificationSettingViews {
		notificationSettingsResponse := &NotificationSettingsResponse{
			Id:         ns.Id,
			ConfigName: ns.ConfigName,
		}

		var config NSConfig
		configJson := []byte(ns.Config)
		err := json.Unmarshal(configJson, &config)
		if err != nil {
			impl.logger.Errorw("unmarshal error", "err", err)
			return notificationSettingsResponses, deletedItemCount, err
		}

		if config.TeamId != nil && len(config.TeamId) > 0 {
			teams, err := impl.teamRepository.FindByIds(config.TeamId)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error in fetching team", "err", err)
				return notificationSettingsResponses, deletedItemCount, err
			}
			var teamResponse []*TeamResponse
			for _, item := range teams {
				teamResponse = append(teamResponse, &TeamResponse{Id: &item.Id, Name: item.Name})
			}
			notificationSettingsResponse.TeamResponse = teamResponse
		}

		if config.EnvId != nil && len(config.EnvId) > 0 {
			environments, err := impl.environmentRepository.FindByIds(config.EnvId)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error in fetching env", "err", err)
				return notificationSettingsResponses, deletedItemCount, err
			}
			var envResponse []*EnvResponse
			for _, item := range environments {
				envResponse = append(envResponse, &EnvResponse{Id: &item.Id, Name: item.Name})
			}
			notificationSettingsResponse.EnvResponse = envResponse
		}

		if config.AppId != nil && len(config.AppId) > 0 {
			applications, err := impl.appRepository.FindByIds(config.AppId)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error in fetching app", "err", err)
				return notificationSettingsResponses, deletedItemCount, err
			}
			var appResponse []*AppResponse
			for _, item := range applications {
				appResponse = append(appResponse, &AppResponse{Id: &item.Id, Name: item.AppName})
			}
			notificationSettingsResponse.AppResponse = appResponse
		}

		if config.Providers != nil && len(config.Providers) > 0 {
			var slackIds []*int
			var sesUserIds []int32
			var smtpUserIds []int32
			var providerConfigs []*ProvidersConfig
			for _, item := range config.Providers {
				// if item.ConfigId > 0 that means, user is of user repository, else user email is custom
				if item.ConfigId > 0 {
					if item.Destination == util.Slack {
						slackIds = append(slackIds, &item.ConfigId)
					} else if item.Destination == util.SES {
						sesUserIds = append(sesUserIds, int32(item.ConfigId))
					} else if item.Destination == util.SMTP {
						smtpUserIds = append(smtpUserIds, int32(item.ConfigId))
					}
				} else {
					providerConfigs = append(providerConfigs, &ProvidersConfig{Dest: string(item.Destination), Recipient: item.Recipient})
				}
			}
			if len(slackIds) > 0 {
				slackConfigs, err := impl.slackRepository.FindByIds(slackIds)
				if err != nil && err != pg.ErrNoRows {
					impl.logger.Errorw("error in fetching slack config", "err", err)
					return notificationSettingsResponses, deletedItemCount, err
				}
				for _, item := range slackConfigs {
					providerConfigs = append(providerConfigs, &ProvidersConfig{Id: item.Id, ConfigName: item.ConfigName, Dest: string(util.Slack)})
				}
			}

			if len(sesUserIds) > 0 {
				sesConfigs, err := impl.userRepository.GetByIds(sesUserIds)
				if err != nil && err != pg.ErrNoRows {
					impl.logger.Errorw("error in fetching user", "error", err)
					return notificationSettingsResponses, deletedItemCount, err
				}
				for _, item := range sesConfigs {
					providerConfigs = append(providerConfigs, &ProvidersConfig{Id: int(item.Id), ConfigName: item.EmailId, Dest: string(util.SES)})
				}
			}
			if len(smtpUserIds) > 0 {
				smtpConfigs, err := impl.userRepository.GetByIds(smtpUserIds)
				if err != nil && err != pg.ErrNoRows {
					impl.logger.Errorw("error in fetching user", "error", err)
					return notificationSettingsResponses, deletedItemCount, err
				}
				for _, item := range smtpConfigs {
					providerConfigs = append(providerConfigs, &ProvidersConfig{Id: int(item.Id), ConfigName: item.EmailId, Dest: string(util.SMTP)})
				}
			}
			notificationSettingsResponse.ProvidersConfig = providerConfigs
		}

		if config.PipelineId != nil && *config.PipelineId > 0 {
			pipelineResponse := &PipelineResponse{}
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

func (impl *NotificationConfigServiceImpl) buildProvidersConfig(config config) ([]ProvidersConfig, error) {
	var providerConfigs []ProvidersConfig
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
				return []ProvidersConfig{}, err
			}
			for _, s := range slackConfigs {
				slackConfigNameMap[s.Id] = s.ConfigName
			}
		}
		if len(sesIds) > 0 {
			sesConfigs, err := impl.sesRepository.FindByIdsIn(sesIds)
			if err != nil {
				impl.logger.Errorw("error on fetch ses configs", "err", err)
				return []ProvidersConfig{}, err
			}
			for _, s := range sesConfigs {
				sesConfigNamesMap[s.Id] = s.ConfigName
			}
		}
		if len(smtpIds) > 0 {
			smtpConfigs, err := impl.smtpRepository.FindByIdsIn(sesIds)
			if err != nil {
				impl.logger.Errorw("error on fetch smtp configs", "err", err)
				return []ProvidersConfig{}, err
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
			providerConfig := ProvidersConfig{
				Id:         c.ConfigId,
				Dest:       string(c.Destination),
				ConfigName: configName,
			}
			providerConfigs = append(providerConfigs, providerConfig)
		}
	}
	return providerConfigs, nil
}

func (impl *NotificationConfigServiceImpl) buildPipelineResponses(config config, ciPipelines []*pipelineConfig.CiPipeline, cdPipelines []*pipelineConfig.Pipeline) (util.PipelineType, []PipelineResponse, error) {
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
			return "", []PipelineResponse{}, err
		}
	}
	var pipelineResponses []PipelineResponse
	if util.CI == pipelineType {
		for _, ci := range ciPipelines {
			pipelineResponse := PipelineResponse{
				Id:   &ci.Id,
				Name: ci.Name,
			}
			pipelineResponses = append(pipelineResponses, pipelineResponse)
		}
	} else if util.CD == pipelineType {
		for _, cd := range cdPipelines {
			pipelineResponse := PipelineResponse{
				Id:   &cd.Id,
				Name: cd.Name,
			}
			pipelineResponses = append(pipelineResponses, pipelineResponse)
		}
	}
	return pipelineType, pipelineResponses, nil
}

func (impl *NotificationConfigServiceImpl) saveNotificationSetting(notificationSettingsRequest *NotificationConfigRequest, userId int32, tx *pg.Tx) (int, error) {
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

func (impl *NotificationConfigServiceImpl) updateNotificationSetting(notificationSettingsRequest *NotificationConfigRequest, updateType util.UpdateType, userId int32, tx *pg.Tx) (int, error) {
	var err error
	existingNotificationSettingsConfig, err := impl.notificationSettingsRepository.FindNotificationSettingsViewById(notificationSettingsRequest.Id)
	if err != nil {
		impl.logger.Errorw("failed to fetch existing notification settings view", "err", err)
		return 0, err
	}

	nsConfig := &NSConfig{}
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
					notificationSetting, err := impl.notificationConfigBuilder.BuildNotificationSettingWithPipeline(item.TeamId, item.EnvId, item.AppId, item.PipelineId, util.PipelineType(item.PipelineType), e, notificationSettingsRequest.Id, nsConfig.Providers)
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

func (impl *NotificationConfigServiceImpl) FindNotificationSettingOptions(settingRequest *repository.SearchRequest) ([]*SearchFilterResponse, error) {
	var searchFilterResponse []*SearchFilterResponse

	settingOptionResponseDeployment, err := impl.notificationSettingsRepository.FindNotificationSettingDeploymentOptions(settingRequest)
	if err != nil && !util2.IsErrNoRows(err) {
		impl.logger.Errorw("error while fetching notification deployment option", "err", err)
		return searchFilterResponse, err
	}
	for _, item := range settingOptionResponseDeployment {
		item.PipelineType = string(util.CD)
		result := &SearchFilterResponse{
			PipelineType:     item.PipelineType,
			PipelineResponse: &PipelineResponse{Id: &item.PipelineId, Name: item.PipelineName, EnvironmentName: item.EnvironmentName, AppName: item.AppName},
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
		result := &SearchFilterResponse{
			PipelineType:     item.PipelineType,
			PipelineResponse: &PipelineResponse{Id: &item.PipelineId, Name: item.PipelineName, AppName: item.AppName, Branches: branches},
		}
		searchFilterResponse = append(searchFilterResponse, result)
	}

	var teamResponse []*TeamResponse
	var appResponse []*AppResponse
	var envResponse []*EnvResponse
	if settingRequest.TeamId != nil && len(settingRequest.TeamId) > 0 {
		teams, err := impl.teamRepository.FindByIds(settingRequest.TeamId)
		if err != nil {
			impl.logger.Errorw("error on fetch teams", "err", err)
			return searchFilterResponse, err
		}
		for _, item := range teams {
			teamResponse = append(teamResponse, &TeamResponse{Id: &item.Id, Name: item.Name})
		}
	}
	if settingRequest.EnvId != nil && len(settingRequest.EnvId) > 0 {
		environments, err := impl.environmentRepository.FindByIds(settingRequest.EnvId)
		if err != nil {
			impl.logger.Errorw("error on fetching environments", "err", err)
			return searchFilterResponse, err
		}
		for _, item := range environments {
			envResponse = append(envResponse, &EnvResponse{Id: &item.Id, Name: item.Name})
		}
	}
	if settingRequest.AppId != nil && len(settingRequest.AppId) > 0 {
		applications, err := impl.appRepository.FindByIds(settingRequest.AppId)
		if err != nil {
			impl.logger.Errorw("error on fetching apps", "err", err)
			return searchFilterResponse, err
		}
		for _, item := range applications {
			appResponse = append(appResponse, &AppResponse{Id: &item.Id, Name: item.AppName})
		}
	}
	ciMatching := &SearchFilterResponse{
		PipelineType: string(util.CI),
		TeamResponse: teamResponse,
		AppResponse:  appResponse,
	}
	searchFilterResponse = append(searchFilterResponse, ciMatching)

	cdMatching := &SearchFilterResponse{
		PipelineType: string(util.CD),
		TeamResponse: teamResponse,
		AppResponse:  appResponse,
		EnvResponse:  envResponse,
	}
	searchFilterResponse = append(searchFilterResponse, cdMatching)

	if searchFilterResponse == nil {
		searchFilterResponse = make([]*SearchFilterResponse, 0)
	}

	return searchFilterResponse, nil
}

func (impl *NotificationConfigServiceImpl) FetchNSViewByIds(ids []*int) ([]*NSConfig, error) {
	var configs []*NSConfig
	notificationSettingViewList, err := impl.notificationSettingsRepository.FindNotificationSettingsViewByIds(ids)
	if err != nil {
		impl.logger.Errorw("failed to fetch existing notification settings view", "err", err)
		return configs, err
	}
	for _, item := range notificationSettingViewList {
		nsConfig := &NSConfig{}
		err = json.Unmarshal([]byte(item.Config), nsConfig)
		if err != nil {
			return configs, err
		}
		configs = append(configs, nsConfig)
	}

	return configs, nil
}
