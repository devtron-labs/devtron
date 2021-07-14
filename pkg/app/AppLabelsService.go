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

package app

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type AppLabelService interface {
	Create(request *bean.AppLabelDto) (*bean.AppLabelDto, error)
	UpdateLabelsInApp(request *bean.AppLabelsDto) (*bean.AppLabelsDto, error)
	FindById(id int) (*bean.AppLabelDto, error)
	FindAllActive() ([]*bean.AppLabelDto, error)
	GetAppMetaInfo(appId int) (*bean.AppMetaInfoDto, error)
}
type AppLabelServiceImpl struct {
	logger             *zap.SugaredLogger
	appLabelRepository pipelineConfig.AppLabelRepository
	appRepository      pipelineConfig.AppRepository
}

func NewAppLabelServiceImpl(appLabelRepository pipelineConfig.AppLabelRepository,
	logger *zap.SugaredLogger, appRepository pipelineConfig.AppRepository) *AppLabelServiceImpl {
	return &AppLabelServiceImpl{
		appLabelRepository: appLabelRepository,
		logger:             logger,
		appRepository:      appRepository,
	}
}

func (impl AppLabelServiceImpl) Create(request *bean.AppLabelDto) (*bean.AppLabelDto, error) {
	model := &pipelineConfig.AppLabel{
		Key:   request.Label,
		Value: request.Label,
		AppId: request.AppId,
	}
	model.CreatedBy = request.UserId
	model.UpdatedBy = request.UserId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	_, err := impl.appLabelRepository.Create(model)
	if err != nil {
		impl.logger.Errorw("error in creating new app labels", "error", err)
		return nil, err
	}
	request.Id = model.Id
	return request, nil
}

func (impl AppLabelServiceImpl) UpdateLabelsInApp(request *bean.AppLabelsDto) (*bean.AppLabelsDto, error) {
	for _, label := range request.Labels {
		model, err := impl.appLabelRepository.FindByAppIdAndKeyAndValue(request.AppId, label, label)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in fetching app label", "error", err)
			return nil, err
		}
		if err == pg.ErrNoRows {
			model = &pipelineConfig.AppLabel{
				Key:   label,
				Value: label,
				AppId: request.AppId,
			}
			model.CreatedBy = request.UserId
			model.UpdatedBy = request.UserId
			model.CreatedOn = time.Now()
			model.UpdatedOn = time.Now()
			_, err = impl.appLabelRepository.Create(model)
			if err != nil {
				impl.logger.Errorw("error in creating new app labels", "error", err)
				return nil, err
			}
		}
	}
	return request, nil
}

func (impl AppLabelServiceImpl) FindById(id int) (*bean.AppLabelDto, error) {
	model, err := impl.appLabelRepository.FindById(id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching app labels", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return nil, nil
	}
	ssoLoginDto := &bean.AppLabelDto{
		Id: model.Id,
	}
	return ssoLoginDto, nil
}

func (impl AppLabelServiceImpl) FindAllActive() ([]*bean.AppLabelDto, error) {
	results := make([]*bean.AppLabelDto, 0)
	models, err := impl.appLabelRepository.FindAll()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching FindAll app labels", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return results, nil
	}
	for _, model := range models {
		dto := &bean.AppLabelDto{
			Id:    model.Id,
			Label: fmt.Sprintf("%s:%s", model.Key, model.Value),
		}
		results = append(results, dto)
	}
	return results, nil
}

func (impl AppLabelServiceImpl) GetAppMetaInfo(appId int) (*bean.AppMetaInfoDto, error) {
	app, err := impl.appRepository.FindAppAndProjectByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching GetAppMetaInfo", "error", err)
		return nil, err
	}

	models, err := impl.appLabelRepository.FindAll()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching GetAppMetaInfo", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return nil, nil
	}
	var labels []*bean.AppLabelDto
	for _, model := range models {
		dto := &bean.AppLabelDto{
			Id:    model.Id,
			Label: fmt.Sprintf("%s:%s", model.Key, model.Value),
		}
		labels = append(labels, dto)
	}
	info := &bean.AppMetaInfoDto{
		AppId:       app.Id,
		AppName:     app.AppName,
		ProjectId:   app.TeamId,
		ProjectName: app.Team.Name,
		Labels:      labels,
		Active:      app.Active,
	}
	return info, nil
}
