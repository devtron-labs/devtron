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
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type AppLabelService interface {
	Create(request *bean.AppLabelDto, tx *pg.Tx) (*bean.AppLabelDto, error)
	UpdateLabelsInApp(request *bean.AppLabelsDto) (*bean.AppLabelsDto, error)
	FindById(id int) (*bean.AppLabelDto, error)
	FindAll() ([]*bean.AppLabelDto, error)
	GetAppMetaInfo(appId int) (*bean.AppMetaInfoDto, error)
}
type AppLabelServiceImpl struct {
	logger             *zap.SugaredLogger
	appLabelRepository pipelineConfig.AppLabelRepository
	appRepository      pipelineConfig.AppRepository
	userRepository     repository.UserRepository
}

func NewAppLabelServiceImpl(appLabelRepository pipelineConfig.AppLabelRepository,
	logger *zap.SugaredLogger, appRepository pipelineConfig.AppRepository, userRepository repository.UserRepository) *AppLabelServiceImpl {
	return &AppLabelServiceImpl{
		appLabelRepository: appLabelRepository,
		logger:             logger,
		appRepository:      appRepository,
		userRepository:     userRepository,
	}
}

func (impl AppLabelServiceImpl) Create(request *bean.AppLabelDto, tx *pg.Tx) (*bean.AppLabelDto, error) {
	_, err := impl.appLabelRepository.FindByAppIdAndKeyAndValue(request.AppId, request.Key, request.Value)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching app label", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		model := &pipelineConfig.AppLabel{
			Key:   request.Key,
			Value: request.Value,
			AppId: request.AppId,
		}
		model.CreatedBy = request.UserId
		model.UpdatedBy = request.UserId
		model.CreatedOn = time.Now()
		model.UpdatedOn = time.Now()
		_, err = impl.appLabelRepository.Create(model, tx)
		if err != nil {
			impl.logger.Errorw("error in creating new app labels", "error", err)
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("duplicate key found for app %d, %s", request.AppId, request.Key)
	}
	return request, nil
}

func (impl AppLabelServiceImpl) UpdateLabelsInApp(request *bean.AppLabelsDto) (*bean.AppLabelsDto, error) {
	dbConnection := impl.appRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	appLabels, err := impl.appLabelRepository.FindAllByAppId(request.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching app label", "error", err)
		return nil, err
	}
	appLabelMap := make(map[string]*pipelineConfig.AppLabel)
	for _, appLabel := range appLabels {
		uniqueLabelExists := fmt.Sprintf("%s:%s", appLabel.Key, appLabel.Value)
		if _, ok := appLabelMap[uniqueLabelExists]; !ok {
			appLabelMap[uniqueLabelExists] = appLabel
		}
	}

	for _, label := range request.Labels {
		uniqueLabelRequest := fmt.Sprintf("%s:%s", label.Key, label.Value)
		if _, ok := appLabelMap[uniqueLabelRequest]; !ok {
			// create new
			model := &pipelineConfig.AppLabel{
				Key:   label.Key,
				Value: label.Value,
				AppId: request.AppId,
			}
			model.CreatedBy = request.UserId
			model.UpdatedBy = request.UserId
			model.CreatedOn = time.Now()
			model.UpdatedOn = time.Now()
			_, err = impl.appLabelRepository.Create(model, tx)
			if err != nil {
				impl.logger.Errorw("error in creating new app labels", "error", err)
				return nil, err
			}
		} else {
			// delete from map so that item remain live, all other item will be delete from this app
			delete(appLabelMap, uniqueLabelRequest)
		}
	}
	for _, appLabel := range appLabelMap {
		err = impl.appLabelRepository.Delete(appLabel, tx)
		if err != nil {
			impl.logger.Errorw("error in delete app label", "error", err)
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in commit repo", "error", err)
		return nil, err
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
		return &bean.AppLabelDto{}, nil
	}
	label := &bean.AppLabelDto{
		Key:   model.Key,
		Value: model.Value,
		AppId: model.AppId,
	}
	return label, nil
}

func (impl AppLabelServiceImpl) FindAll() ([]*bean.AppLabelDto, error) {
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
			AppId: model.AppId,
			Key:   model.Key,
			Value: model.Value,
		}
		results = append(results, dto)
	}
	return results, nil
}

func (impl AppLabelServiceImpl) GetAppMetaInfo(appId int) (*bean.AppMetaInfoDto, error) {
	app, err := impl.appRepository.FindAppAndProjectByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching GetAppMetaInfo", "error", err)
		return nil, err
	}
	labels := make([]*bean.Label, 0)
	models, err := impl.appLabelRepository.FindAllByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching GetAppMetaInfo", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		impl.logger.Infow("no labels found for app", "app", app)
	} else {
		for _, model := range models {
			dto := &bean.Label{
				Key:   model.Key,
				Value: model.Value,
			}
			labels = append(labels, dto)
		}
	}

	user, err := impl.userRepository.GetById(app.CreatedBy)
	if err != nil {
		impl.logger.Errorw("error in fetching user for app meta info", "error", err)
		return nil, err
	}
	info := &bean.AppMetaInfoDto{
		AppId:       app.Id,
		AppName:     app.AppName,
		ProjectId:   app.TeamId,
		ProjectName: app.Team.Name,
		CreatedBy:   user.EmailId,
		CreatedOn:   app.CreatedOn,
		Labels:      labels,
		Active:      app.Active,
	}
	return info, nil
}
