package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/attributes"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"log"
	"strconv"
)

type PresetContainerRegistryHandler interface {
	SyncAndUpdatePresetContainerRegistry()
	GetPresetDockerRegistryConfigBean() *PresetDockerRegistryConfigBean
}

type PresetContainerRegistryHandlerImpl struct {
	logger                         *zap.SugaredLogger
	dockerRegistryConfig           DockerRegistryConfig
	presetDockerRegistryConfigBean *PresetDockerRegistryConfigBean
	cron                           *cron.Cron
	attributeService               attributes.AttributesService
}

const PresetRegistryRepoNameKey = "PresetRegistryRepoName"
const PresetRegistryExpiryTimeKey = "PresetRegistryExpiryTime"

func NewPresetContainerRegistryHandlerImpl(logger *zap.SugaredLogger,
	dockerRegistryConfig DockerRegistryConfig, attributeService attributes.AttributesService) *PresetContainerRegistryHandlerImpl {
	cron := cron.New(
		cron.WithChain())
	cron.Start()
	presetDockerRegistryConfigBean := LoadConfig()
	impl := &PresetContainerRegistryHandlerImpl{
		logger:                         logger,
		cron:                           cron,
		dockerRegistryConfig:           dockerRegistryConfig,
		presetDockerRegistryConfigBean: presetDockerRegistryConfigBean,
		attributeService:               attributeService,
	}
	_, err := cron.AddFunc(presetDockerRegistryConfigBean.PresetRegistryUpdateCronExpr, impl.SyncAndUpdatePresetContainerRegistry)
	if err != nil {
		logger.Errorw("error in starting preset container registry update cron job", "err", err)
		return nil
	}
	go impl.loadExpiryAndRepoNameFromDb()
	return impl
}

type PresetDockerRegistryConfigBean struct {
	PresetRegistrySyncUrl               string `env:"PRESET_REGISTRY_SYNC_URL" envDefault:"https://api.devtron.ai/presetCR"`
	PresetRegistryUpdateCronExpr        string `env:"PRESET_REGISTRY_UPDATE_CRON_EXPR" envDefault:"0 */1 * * *"`
	PresetPublicRegistryImgTagValue     string `env:"PRESET_PUBLIC_REGISTRY_IMG_TAG" envDefault:"24h"`
	PresetPublicRegistry                string `env:"PRESET_PUBLIC_REGISTRY" envDefault:"ttl.sh"`
	PresetRegistryRepoName              string
	PresetRegistryImageExpiryTimeInSecs int
}

func LoadConfig() *PresetDockerRegistryConfigBean {
	cfg := &PresetDockerRegistryConfigBean{}
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal("error occurred while loading docker registry config ")
	}
	return cfg
}

func (impl *PresetContainerRegistryHandlerImpl) GetPresetDockerRegistryConfigBean() *PresetDockerRegistryConfigBean {

	cfg := impl.presetDockerRegistryConfigBean
	if cfg.PresetRegistryImageExpiryTimeInSecs == 0 || cfg.PresetRegistryRepoName == "" {
		impl.loadExpiryAndRepoNameFromDb()
	}
	return cfg
}

func (impl *PresetContainerRegistryHandlerImpl) loadExpiryAndRepoNameFromDb() {
	imageExpiryTimeInSecs := 0
	presetRepoName := ""
	keys := []string{PresetRegistryExpiryTimeKey, PresetRegistryRepoNameKey}
	keyAttributes, err := impl.attributeService.GetByKeys(keys)
	if err != nil {
		return
	}
	for _, attribute := range keyAttributes {
		attrValue := attribute.Value
		attrKey := attribute.Key
		if attrKey == PresetRegistryRepoNameKey {
			presetRepoName = attrValue
		} else if attrKey == PresetRegistryExpiryTimeKey {
			imageExpiryTimeInSecs, _ = strconv.Atoi(attrValue)
		}
	}
	registryConfigBean := impl.presetDockerRegistryConfigBean
	if registryConfigBean.PresetRegistryImageExpiryTimeInSecs == 0 {
		registryConfigBean.PresetRegistryImageExpiryTimeInSecs = imageExpiryTimeInSecs
	}
	if registryConfigBean.PresetRegistryRepoName == "" {
		registryConfigBean.PresetRegistryRepoName = presetRepoName
	}
}

func (impl *PresetContainerRegistryHandlerImpl) SyncAndUpdatePresetContainerRegistry() {
	presetSyncUrl := impl.presetDockerRegistryConfigBean.PresetRegistrySyncUrl
	presetContainerRegistryByteArr, err := util2.ReadFromUrlWithRetry(presetSyncUrl)
	centralDockerRegistryConfig, registryRepoName, registryExpiryTimeInSecs, err := impl.extractRegistryConfig(presetContainerRegistryByteArr)
	if err != nil {
		impl.logger.Errorw("err during unmarshal for preset container registry response from central-api", "err", err)
		return
	}

	impl.compareAndUpdateExpiryTime(registryExpiryTimeInSecs)
	impl.compareAndUpdateRepoName(registryRepoName)

	registryId := util2.DockerPresetContainerRegistryId
	dockerArtifactStore, err := impl.dockerRegistryConfig.FetchOneDockerAccount(registryId)
	if err != nil {
		impl.logger.Errorw("err in extracting docker registry from DB", "id", registryId, "err", err)
		return
	}
	if changed := impl.compareCentralRegistryAndConfigured(centralDockerRegistryConfig, dockerArtifactStore); changed {
		centralDockerRegistryConfig.Id = registryId
		centralDockerRegistryConfig.User = 1 // system
		_, err := impl.dockerRegistryConfig.Update(centralDockerRegistryConfig)
		if err != nil {
			impl.logger.Errorw("err in updating central-api docker registry into DB", "id", registryId, "err", err)
			return
		}
		impl.logger.Info("docker preset container registry updated from central api")
	} else {
		impl.logger.Debug("docker preset container registry not updated as there is not diff, will check after sometime again!!")
	}

}

func (impl *PresetContainerRegistryHandlerImpl) extractRegistryConfig(arr []byte) (*DockerArtifactStoreBean, string, int, error) {

	var result map[string]interface{}
	err := json.Unmarshal(arr, &result)
	if err != nil {
		return nil, "", 0, err
	}
	statusCode := result["code"]
	sCodeStr := statusCode.(float64)
	if sCodeStr != 200 {
		return nil, "", 0, errors.New("api failed with code" + fmt.Sprint(sCodeStr))
	}
	responseBean := result["result"]
	responseBeanMap := responseBean.(map[string]interface{})
	expiryTimeIsSecs := responseBeanMap["expiryTimeInSecs"]
	registryDefaultRepoName := responseBeanMap["presetRepoName"]
	response1, _ := json.Marshal(responseBeanMap)
	centralDockerRegistryConfig := &DockerArtifactStoreBean{}
	err = json.Unmarshal(response1, centralDockerRegistryConfig)
	return centralDockerRegistryConfig, registryDefaultRepoName.(string), int(expiryTimeIsSecs.(float64)), err
}

func (impl *PresetContainerRegistryHandlerImpl) compareCentralRegistryAndConfigured(centralDockerRegistry *DockerArtifactStoreBean,
	dbDockerRegistry *DockerArtifactStoreBean) bool {
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

func (impl *PresetContainerRegistryHandlerImpl) compareAndUpdateExpiryTime(syncedExpiryTimeInSecs int) int {
	cachedExpiryTimeInSecs := impl.presetDockerRegistryConfigBean.PresetRegistryImageExpiryTimeInSecs

	if syncedExpiryTimeInSecs != cachedExpiryTimeInSecs {
		attributesDto, err := impl.attributeService.GetByKey(PresetRegistryExpiryTimeKey)
		request := &attributes.AttributesDto{
			Id:     attributesDto.Id,
			Key:    PresetRegistryExpiryTimeKey,
			Value:  strconv.Itoa(syncedExpiryTimeInSecs),
			UserId: 1,
		}
		updateAttributeValue, err := impl.attributeService.UpdateAttributes(request)
		if err != nil {
			return syncedExpiryTimeInSecs
		}
		expiryTimeInSecStr := updateAttributeValue.Value
		updatedExpiryTimeInSec, err := strconv.Atoi(expiryTimeInSecStr)
		impl.presetDockerRegistryConfigBean.PresetRegistryImageExpiryTimeInSecs = updatedExpiryTimeInSec
	}
	return syncedExpiryTimeInSecs
}

func (impl *PresetContainerRegistryHandlerImpl) compareAndUpdateRepoName(syncedRepoName string) string {
	cachedRegistryRepoName := impl.presetDockerRegistryConfigBean.PresetRegistryRepoName

	if cachedRegistryRepoName != syncedRepoName {
		attributesDto, err := impl.attributeService.GetByKey(PresetRegistryRepoNameKey)
		request := &attributes.AttributesDto{
			Id:     attributesDto.Id,
			Key:    PresetRegistryRepoNameKey,
			Value:  syncedRepoName,
			UserId: 1,
		}
		updateAttributeValue, err := impl.attributeService.UpdateAttributes(request)
		if err != nil {
			return syncedRepoName
		}
		updatedRegistryRepoName := updateAttributeValue.Value
		impl.presetDockerRegistryConfigBean.PresetRegistryRepoName = updatedRegistryRepoName
	}
	return syncedRepoName
}
