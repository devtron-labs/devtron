package configDiff

import (
	"github.com/devtron-labs/devtron/pkg/configDiff/adaptor"
	bean2 "github.com/devtron-labs/devtron/pkg/configDiff/bean"
	"github.com/devtron-labs/devtron/pkg/configDiff/helper"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"go.uber.org/zap"
)

type DeploymentConfigurationService interface {
	ConfigAutoComplete(appId int, envId int) (*bean2.ConfigDataResponse, error)
}

type DeploymentConfigurationServiceImpl struct {
	logger           *zap.SugaredLogger
	configMapService pipeline.ConfigMapService
}

func NewDeploymentConfigurationServiceImpl(logger *zap.SugaredLogger,
	configMapService pipeline.ConfigMapService,
) (*DeploymentConfigurationServiceImpl, error) {
	deploymentConfigurationService := &DeploymentConfigurationServiceImpl{
		logger:           logger,
		configMapService: configMapService,
	}

	return deploymentConfigurationService, nil
}
func (impl *DeploymentConfigurationServiceImpl) ConfigAutoComplete(appId int, envId int) (*bean2.ConfigDataResponse, error) {
	cMCSNamesAppLevel, cMCSNamesEnvLevel, err := impl.configMapService.FetchCmCsNamesAppAndEnvLevel(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in fetching CM and CS names at app or env level", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	cmcsKeyPropertyAppLevelMap, cmcsKeyPropertyEnvLevelMap := adaptor.GetCmCsAppAndEnvLevelMap(cMCSNamesAppLevel, cMCSNamesEnvLevel)
	for key, configProperty := range cmcsKeyPropertyAppLevelMap {
		if _, ok := cmcsKeyPropertyEnvLevelMap[key]; !ok {
			if envId > 0 {
				configProperty.ConfigStage = bean2.Inheriting
			}

		}
	}
	for key, configProperty := range cmcsKeyPropertyEnvLevelMap {
		if _, ok := cmcsKeyPropertyAppLevelMap[key]; ok {
			configProperty.ConfigStage = bean2.Overridden
		} else {
			configProperty.ConfigStage = bean2.Env
		}
	}
	combinedProperties := helper.GetCombinedPropertiesMap(cmcsKeyPropertyAppLevelMap, cmcsKeyPropertyEnvLevelMap)
	combinedProperties = append(combinedProperties, adaptor.GetConfigProperty(0, "", bean.DeploymentTemplate, bean2.PublishedConfigState))

	configDataResp := bean2.NewConfigDataResponse().WithResourceConfig(combinedProperties)
	return configDataResp, nil
}
