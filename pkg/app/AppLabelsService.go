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
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type AppLabelsCreateRequest struct {
	Labels []string `json:"labels,notnull"`
	AppId  int      `json:"appId"`
	UserId int32    `json:"-"`
}

type AppLabelsDto struct {
	Id     int    `json:"id,pk"`
	Label  string `json:"label,notnull"`
	AppId  int    `json:"appId"`
	Active bool   `json:"active,notnull"`
	UserId int32  `json:"-"`
}

type AppMetaInfoDto struct {
	AppId       int             `json:"appId"`
	AppName     string          `json:"appName"`
	ProjectId   int             `json:"projectId"`
	ProjectName string          `json:"projectName"`
	CreatedBy   string          `json:"createdBy"`
	CreatedOn   time.Time       `json:"createdOn"`
	Active      bool            `json:"active,notnull"`
	Labels      []*AppLabelsDto `json:"labels"`
	UserId      int32           `json:"-"`
}

type AppLabelsService interface {
	Create(request *AppLabelsDto) (*AppLabelsDto, error)
	EditAppLabels(request *AppLabelsCreateRequest) (*AppLabelsCreateRequest, error)
	FindById(id int) (*AppLabelsDto, error)
	FindAllActive() ([]*AppLabelsDto, error)
	GetAppMetaInfo(appId int) (*AppMetaInfoDto, error)
}
type AppLabelsServiceImpl struct {
	logger              *zap.SugaredLogger
	appLabelsRepository pipelineConfig.AppLabelsRepository
	appRepository       pipelineConfig.AppRepository
}

func NewAppLabelsServiceImpl(appLabelsRepository pipelineConfig.AppLabelsRepository,
	logger *zap.SugaredLogger, appRepository pipelineConfig.AppRepository) *AppLabelsServiceImpl {
	return &AppLabelsServiceImpl{
		appLabelsRepository: appLabelsRepository,
		logger:              logger,
		appRepository:       appRepository,
	}
}

func (impl AppLabelsServiceImpl) Create(request *AppLabelsDto) (*AppLabelsDto, error) {
	model := &pipelineConfig.AppLabels{
	}
	model.Active = true
	model.CreatedBy = request.UserId
	model.UpdatedBy = request.UserId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	_, err := impl.appLabelsRepository.Create(model)
	if err != nil {
		impl.logger.Errorw("error in creating new app labels", "error", err)
		return nil, err
	}
	request.Id = model.Id
	return request, nil
}

func (impl AppLabelsServiceImpl) EditAppLabels(request *AppLabelsCreateRequest) (*AppLabelsCreateRequest, error) {
	for _, label := range request.Labels {
		model, err := impl.appLabelsRepository.FindByLabel(label)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in fetching app label", "error", err)
			return nil, err
		}
		if err == pg.ErrNoRows {
			model = &pipelineConfig.AppLabels{}
			model.Active = true
			model.UpdatedBy = request.UserId
			model.CreatedOn = time.Now()
			model.UpdatedOn = time.Now()
			_, err = impl.appLabelsRepository.Create(model)
			if err != nil {
				impl.logger.Errorw("error in creating new app labels", "error", err)
				return nil, err
			}
		}
	}
	return request, nil
}

func (impl AppLabelsServiceImpl) FindById(id int) (*AppLabelsDto, error) {
	model, err := impl.appLabelsRepository.FindById(id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching app labels", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return nil, nil
	}
	ssoLoginDto := &AppLabelsDto{
		Id:     model.Id,
		Active: model.Active,
	}
	return ssoLoginDto, nil
}

func (impl AppLabelsServiceImpl) FindAllActive() ([]*AppLabelsDto, error) {
	results := make([]*AppLabelsDto, 0)
	models, err := impl.appLabelsRepository.FindAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching FindAllActive app labels", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return results, nil
	}
	for _, model := range models {
		dto := &AppLabelsDto{
			Id:    model.Id,
			Label: model.Label,
		}
		results = append(results, dto)
	}
	return results, nil
}

func (impl AppLabelsServiceImpl) GetAppMetaInfo(appId int) (*AppMetaInfoDto, error) {
	app, err := impl.appRepository.FindAppAndProjectByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching GetAppMetaInfo", "error", err)
		return nil, err
	}

	models, err := impl.appLabelsRepository.FindAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching GetAppMetaInfo", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return nil, nil
	}
	var labels []*AppLabelsDto
	for _, model := range models {
		dto := &AppLabelsDto{
			Id:    model.Id,
			Label: model.Label,
		}
		labels = append(labels, dto)
	}
	info := &AppMetaInfoDto{
		AppId:       app.Id,
		AppName:     app.AppName,
		ProjectId:   app.TeamId,
		ProjectName: app.Team.Name,
		Labels:      labels,
		Active:      app.Active,
	}
	return info, nil
}
