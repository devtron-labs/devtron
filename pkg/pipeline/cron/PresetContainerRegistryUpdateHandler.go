package cron

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type PresetContainerRegistryUpdateHandler interface {
	SyncAndUpdatePresetContainerRegistry()
}

type PresetContainerRegistryUpdateHandlerImpl struct {
	logger                 *zap.SugaredLogger
	cron                   *cron.Cron
	dockerRegConfigService pipeline.DockerRegistryConfig
	config                 *DockerRegistryConfigBean
}

type DockerRegistryConfigBean struct {
	PresetRegistrySyncUrl        string `env:"PRESET_REGISTRY_SYNC_URL" envDefault:"https://api-stage.devtron.ai/presetCR"`
	PresetRegistryUpdateCronExpr string `env:"PRESET_REGISTRY_UPDATE_CRON_EXPR" envDefault:"0 */1 * * *"`
}

type PresetRegistrySyncResponseBean struct {
	code   float64
	status string
	result *pipeline.DockerArtifactStoreBean
}

func GetConfig() (*DockerRegistryConfigBean, error) {
	cfg := &DockerRegistryConfigBean{}
	err := env.Parse(cfg)
	return cfg, err
}

func NewPresetContainerRegistryHandlerImpl(logger *zap.SugaredLogger, dockerRegConfig pipeline.DockerRegistryConfig) *PresetContainerRegistryUpdateHandlerImpl {
	cron := cron.New(
		cron.WithChain())
	cron.Start()

	config, err := GetConfig()
	if err != nil {
		logger.Fatalw("error occurred while loading docker registry config ")
	}
	impl := &PresetContainerRegistryUpdateHandlerImpl{
		logger:                 logger,
		cron:                   cron,
		dockerRegConfigService: dockerRegConfig,
		config:                 config,
	}
	_, err = cron.AddFunc(config.PresetRegistryUpdateCronExpr, impl.SyncAndUpdatePresetContainerRegistry)
	if err != nil {
		logger.Errorw("error in starting preset container registry update cron job", "err", err)
		return nil
	}
	return impl
}

func (impl *PresetContainerRegistryUpdateHandlerImpl) SyncAndUpdatePresetContainerRegistry() {
	presetSyncUrl := impl.config.PresetRegistrySyncUrl
	presetContainerRegistryByteArr, err := util2.ReadFromUrlWithRetry(presetSyncUrl)
	centralDockerRegistryConfig, err := extractRegistryConfig(presetContainerRegistryByteArr)
	if err != nil {
		impl.logger.Errorw("err during unmarshal for preset container registry response from central-api", "err", err)
		return
	}

	registryId := util2.DockerPresetContainerRegistry
	dockerArtifactStore, err := impl.dockerRegConfigService.FetchOneDockerAccount(registryId)
	if err != nil {
		impl.logger.Errorw("err in extracting docker registry from DB", "id", registryId, "err", err)
		return
	}
	if changed := impl.compareCentralRegistryAndConfigured(centralDockerRegistryConfig, dockerArtifactStore); changed {
		centralDockerRegistryConfig.Id = registryId
		centralDockerRegistryConfig.User = 1 // system
		_, err := impl.dockerRegConfigService.Update(centralDockerRegistryConfig)
		if err != nil {
			impl.logger.Errorw("err in updating central-api docker registry into DB", "id", registryId, "err", err)
			return
		}
		impl.logger.Info("docker preset container registry updated from central api")
	} else {
		impl.logger.Debug("docker preset container registry not updated as there is not diff, will check after sometime again!!")
	}

}

func extractRegistryConfig(arr []byte) (*pipeline.DockerArtifactStoreBean, error) {

	var result map[string]interface{}
	err := json.Unmarshal(arr, &result)
	if err != nil {
		return nil, err
	}
	statusCode := result["code"]
	sCodeStr := statusCode.(float64)
	if sCodeStr != 200 {
		return nil, errors.New("api failed with code" + fmt.Sprint(sCodeStr))
	}
	responseBean := result["result"]
	response1, _ := json.Marshal(responseBean.(map[string]interface{}))
	centralDockerRegistryConfig := &pipeline.DockerArtifactStoreBean{}
	err = json.Unmarshal(response1, centralDockerRegistryConfig)
	return centralDockerRegistryConfig, err
}

func (impl *PresetContainerRegistryUpdateHandlerImpl) compareCentralRegistryAndConfigured(centralDockerRegistry *pipeline.DockerArtifactStoreBean,
	dbDockerRegistry *pipeline.DockerArtifactStoreBean) bool {
	if centralDockerRegistry.PluginId != dbDockerRegistry.PluginId {
		return true
	}
	if centralDockerRegistry.RegistryURL != dbDockerRegistry.RegistryURL {
		return true
	}
	if centralDockerRegistry.RegistryType != dbDockerRegistry.RegistryType {
		return true
	}
	if centralDockerRegistry.Username != dbDockerRegistry.Username {
		return true
	}
	if centralDockerRegistry.Password != dbDockerRegistry.Password {
		return true
	}
	if centralDockerRegistry.AWSRegion != dbDockerRegistry.AWSRegion {
		return true
	}
	if centralDockerRegistry.AWSAccessKeyId != dbDockerRegistry.AWSAccessKeyId {
		return true
	}
	if centralDockerRegistry.AWSSecretAccessKey != dbDockerRegistry.AWSSecretAccessKey {
		return true
	}
	if centralDockerRegistry.Active != dbDockerRegistry.Active {
		return true
	}
	if centralDockerRegistry.Cert != dbDockerRegistry.Cert {
		return true
	}
	if centralDockerRegistry.Connection != dbDockerRegistry.Connection {
		return true
	}

	return false
}
