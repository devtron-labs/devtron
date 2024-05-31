/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package pipeline

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/app"
	"go.uber.org/zap"
)

type DevtronAppCMCSService interface {
	//FetchConfigmapSecretsForCdStages : Delegating the request to appService for fetching cm/cs
	FetchConfigmapSecretsForCdStages(appId, envId, cdPipelineId int) (ConfigMapSecretsResponse, error)
}

type DevtronAppCMCSServiceImpl struct {
	logger               *zap.SugaredLogger
	appService           app.AppService
	attributesRepository repository.AttributesRepository
}

func NewDevtronAppCMCSServiceImpl(
	logger *zap.SugaredLogger,
	appService app.AppService,
	attributesRepository repository.AttributesRepository) *DevtronAppCMCSServiceImpl {

	return &DevtronAppCMCSServiceImpl{
		logger:               logger,
		appService:           appService,
		attributesRepository: attributesRepository,
	}
}

func (impl *DevtronAppCMCSServiceImpl) FetchConfigmapSecretsForCdStages(appId, envId, cdPipelineId int) (ConfigMapSecretsResponse, error) {
	configMapSecrets, err := impl.appService.GetConfigMapAndSecretJson(appId, envId, cdPipelineId)
	if err != nil {
		impl.logger.Errorw("error while fetching config secrets ", "err", err)
		return ConfigMapSecretsResponse{}, err
	}
	existingConfigMapSecrets := ConfigMapSecretsResponse{}
	err = json.Unmarshal([]byte(configMapSecrets), &existingConfigMapSecrets)
	if err != nil {
		impl.logger.Error(err)
		return ConfigMapSecretsResponse{}, err
	}
	return existingConfigMapSecrets, nil
}
