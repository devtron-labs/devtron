package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
)

type DevtronAppCMCSService interface {
	//FetchConfigmapSecretsForCdStages : Delegating the request to appService for fetching cm/cs
	FetchConfigmapSecretsForCdStages(appId, envId, cdPipelineId int) (ConfigMapSecretsResponse, error)
	//GetDeploymentConfigMap : Retrieve deployment config values from the attributes table
	GetDeploymentConfigMap(environmentId int) (map[string]bool, error)
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

func (impl *DevtronAppCMCSServiceImpl) GetDeploymentConfigMap(environmentId int) (map[string]bool, error) {
	var deploymentConfig map[string]map[string]bool
	var deploymentConfigEnv map[string]bool
	deploymentConfigValues, err := impl.attributesRepository.FindByKey(attributes.ENFORCE_DEPLOYMENT_TYPE_CONFIG)
	if err == pg.ErrNoRows {
		return deploymentConfigEnv, nil
	}
	//if empty config received(doesn't exist in table) which can't be parsed
	if deploymentConfigValues.Value != "" {
		if err := json.Unmarshal([]byte(deploymentConfigValues.Value), &deploymentConfig); err != nil {
			rerr := &util.ApiError{
				HttpStatusCode:  http.StatusInternalServerError,
				InternalMessage: err.Error(),
				UserMessage:     "Failed to fetch deployment config values from the attributes table",
			}
			return deploymentConfigEnv, rerr
		}
		deploymentConfigEnv, _ = deploymentConfig[fmt.Sprintf("%d", environmentId)]
	}
	return deploymentConfigEnv, nil
}
