package config

import (
	"fmt"
	globalUtil "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/config/read"
	"github.com/devtron-labs/devtron/pkg/infraConfig/adapter"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/devtron-labs/devtron/pkg/infraConfig/util"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/util/sliceUtil"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type InfraConfigClient interface {
	GetDefaultConfigurationForPlatform(platformName string, defaultConfigurationsMap map[string][]*v1.ConfigurationBean) []*v1.ConfigurationBean
	GetConfigurationBeansForProfile(infraProfileConfigurationEntities []*repository.InfraProfileConfigurationEntity, profileName string) (map[string][]*v1.ConfigurationBean, error)
	Validate(profileBean, defaultProfile *v1.ProfileBeanDto) (map[string]v1.InfraConfigKeys, error)
	GetConfigurationUnits() (map[v1.ConfigKeyStr]map[string]v1.Unit, error)
	GetInfraProfileConfigurationEntities(profileBean *v1.ProfileBeanDto, userId int32) ([]*repository.InfraProfileConfigurationEntity, error)
	HandlePostUpdateOperations(tx *pg.Tx, updatedInfraConfigs []*repository.InfraProfileConfigurationEntity) error
	HandlePostCreateOperations(tx *pg.Tx, createdInfraConfigs []*repository.InfraProfileConfigurationEntity) error
	GetInfraConfigEntities(profileId int, infraConfig *v1.InfraConfig) ([]*repository.InfraProfileConfigurationEntity, error)
	OverrideInfraConfig(infraConfiguration *v1.InfraConfig, configurationBean *v1.ConfigurationBean) (*v1.InfraConfig, error)
	ConvertToProfilePlatformMap(infraProfileConfigurationEntities []*repository.InfraProfileConfigurationEntity, profilesMap map[int]*v1.ProfileBeanDto, profilePlatforms []*repository.ProfilePlatformMapping) (map[int]map[string][]*v1.ConfigurationBean, error)
	MergeInfraConfigurations(supportedConfigKey v1.ConfigKeyStr, profileConfiguration *v1.ConfigurationBean, defaultConfigurations []*v1.ConfigurationBean) (*v1.ConfigurationBean, error)
	HandleInfraConfigTriggerAudit(workflowId int, triggeredBy int32, infraConfigs map[string]*v1.InfraConfig) error
	InfraConfigEntClient
}

type InfraConfigClientImpl struct {
	logger          *zap.SugaredLogger
	configFactories *configFactories
	unitFactoryMap  *unitFactories
}

func NewInfraConfigClient(logger *zap.SugaredLogger,
	scopedVariableManager variables.ScopedVariableManager,
	configReadService read.ConfigReadService) *InfraConfigClientImpl {
	return &InfraConfigClientImpl{
		logger:          logger,
		configFactories: getConfigFactory(logger, scopedVariableManager, configReadService),
		unitFactoryMap:  getUnitFactoryMap(logger),
	}
}

func (impl *InfraConfigClientImpl) getCPUConfigFactory() configFactory[float64] {
	return impl.configFactories.cpuConfigFactory
}

func (impl *InfraConfigClientImpl) getMemoryConfigFactory() configFactory[float64] {
	return impl.configFactories.memConfigFactory
}

func (impl *InfraConfigClientImpl) getTimeoutConfigFactory() configFactory[float64] {
	return impl.configFactories.timeoutConfigFactory
}

func (impl *InfraConfigClientImpl) GetDefaultConfigurationForPlatform(platformName string, defaultConfigurationsMap map[string][]*v1.ConfigurationBean) []*v1.ConfigurationBean {
	if len(defaultConfigurationsMap) == 0 {
		return []*v1.ConfigurationBean{}
	}
	// Check if the platform exists in the defaultConfigurationsMap
	defaultConfigurations, exists := defaultConfigurationsMap[platformName]
	if !exists {
		// If not, fallback to the default platform configurations
		return filterSupportedConfigurations(platformName, defaultConfigurationsMap[v1.RUNNER_PLATFORM])
	}
	return filterSupportedConfigurations(platformName, defaultConfigurations)
}

func (impl *InfraConfigClientImpl) GetConfigurationBeansForProfile(infraProfileConfigurationEntities []*repository.InfraProfileConfigurationEntity, profileName string) (map[string][]*v1.ConfigurationBean, error) {
	platformMap := make(map[string][]*v1.ConfigurationBean)
	if len(infraProfileConfigurationEntities) == 0 {
		return platformMap, nil
	}
	if profileName == "" {
		errMsg := "profileName cannot be empty"
		return platformMap, globalUtil.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	for _, infraProfileConfiguration := range infraProfileConfigurationEntities {
		if infraProfileConfiguration == nil {
			continue
		}
		configurationBean, err := impl.getConfigurationBean(infraProfileConfiguration, profileName)
		if err != nil {
			impl.logger.Errorw("failed to get configurations for profile", "err", err, "profileName", profileName)
			errMsg := fmt.Sprintf("failed to get configurations for profile '%s'", profileName)
			return platformMap, globalUtil.NewApiError(http.StatusBadRequest, errMsg, errMsg)
		}
		platform := infraProfileConfiguration.ProfilePlatformMapping.Platform
		if len(platform) == 0 {
			platform = v1.RUNNER_PLATFORM
		}
		// Add the ConfigurationBean to the corresponding platform entry in the map
		platformMap[platform] = append(platformMap[platform], configurationBean)
	}
	return platformMap, nil
}

func (impl *InfraConfigClientImpl) Validate(profileBean, defaultProfile *v1.ProfileBeanDto) (map[string]v1.InfraConfigKeys, error) {
	platformWiseDefaultConfigKeyMap := make(map[string]v1.InfraConfigKeys)
	for platformName, platformConfigurations := range profileBean.GetConfigurations() {
		// Check if the same platform exists in global profile,
		// if not exist, then take go falling on the default platform value
		defaultConfigurations := impl.GetDefaultConfigurationForPlatform(platformName, defaultProfile.GetConfigurations())
		defaultConfigKeyMap, err := impl.validateConfig(platformName, platformConfigurations, defaultConfigurations, false)
		if err != nil {
			return platformWiseDefaultConfigKeyMap, err
		}
		platformWiseDefaultConfigKeyMap[platformName] = defaultConfigKeyMap
	}
	return platformWiseDefaultConfigKeyMap, nil
}

func (impl *InfraConfigClientImpl) GetConfigurationUnits() (map[v1.ConfigKeyStr]map[string]v1.Unit, error) {
	configurationUnits := make(map[v1.ConfigKeyStr]map[string]v1.Unit)
	for configKey, supportedUnits := range impl.getCPUConfigFactory().getSupportedUnits() {
		configurationUnits[configKey] = supportedUnits
	}
	for configKey, supportedUnits := range impl.getMemoryConfigFactory().getSupportedUnits() {
		configurationUnits[configKey] = supportedUnits
	}
	for configKey, supportedUnits := range impl.getTimeoutConfigFactory().getSupportedUnits() {
		configurationUnits[configKey] = supportedUnits
	}
	entConfigurationUnits, err := impl.getEntConfigurationUnits()
	if err != nil {
		return configurationUnits, err
	}
	for configKey, supportedUnits := range entConfigurationUnits {
		configurationUnits[configKey] = supportedUnits
	}
	return configurationUnits, nil
}

// GetInfraProfileConfigurationEntities converts bean.ProfileBeanDto back to []repository.InfraProfileConfigurationEntity
func (impl *InfraConfigClientImpl) GetInfraProfileConfigurationEntities(profileBean *v1.ProfileBeanDto, userId int32) ([]*repository.InfraProfileConfigurationEntity, error) {
	var entities []*repository.InfraProfileConfigurationEntity
	for platform, configBeans := range profileBean.GetConfigurations() {
		for _, configBean := range configBeans {
			configBean.ProfileId = profileBean.Id
			configBean.ProfileName = profileBean.GetName()
			valueString, err := impl.formatTypedValueAsString(configBean.Key, configBean.Value)
			if err != nil {
				return nil, err
			}
			entity := adapter.GetInfraProfileEntity(configBean, valueString, platform, userId)
			entities = append(entities, entity)
		}
	}
	return entities, nil
}

func (impl *InfraConfigClientImpl) GetInfraConfigEntities(profileId int, infraConfig *v1.InfraConfig) ([]*repository.InfraProfileConfigurationEntity, error) {
	defaultConfigurations := make([]*repository.InfraProfileConfigurationEntity, 0)
	cpuInfraEntities, err := impl.getCPUConfigFactory().getInfraConfigEntities(infraConfig, profileId, v1.RUNNER_PLATFORM)
	if err != nil {
		impl.logger.Errorw("error in getting infra cpu config entities", "error", err, "infraConfig", infraConfig)
		return defaultConfigurations, err
	}
	defaultConfigurations = sliceUtil.Filter(defaultConfigurations, cpuInfraEntities,
		func(entity *repository.InfraProfileConfigurationEntity) bool {
			return entity != nil
		})
	memInfraEntities, err := impl.getMemoryConfigFactory().getInfraConfigEntities(infraConfig, profileId, v1.RUNNER_PLATFORM)
	if err != nil {
		impl.logger.Errorw("error in getting infra memory config entities", "error", err, "infraConfig", infraConfig)
		return defaultConfigurations, err
	}
	defaultConfigurations = sliceUtil.Filter(defaultConfigurations, memInfraEntities,
		func(entity *repository.InfraProfileConfigurationEntity) bool {
			return entity != nil
		})
	timeoutInfraEntities, err := impl.getTimeoutConfigFactory().getInfraConfigEntities(infraConfig, profileId, v1.RUNNER_PLATFORM)
	if err != nil {
		impl.logger.Errorw("error in getting infra timeout config entities", "error", err, "infraConfig", infraConfig)
		return defaultConfigurations, err
	}
	defaultConfigurations = sliceUtil.Filter(defaultConfigurations, timeoutInfraEntities,
		func(entity *repository.InfraProfileConfigurationEntity) bool {
			return entity != nil
		})
	entInfraEntities, err := impl.getInfraConfigEntEntities(profileId, infraConfig)
	if err != nil {
		impl.logger.Errorw("error in getting infra ent config entities", "error", err, "infraConfig", infraConfig)
		return defaultConfigurations, err
	}
	defaultConfigurations = append(defaultConfigurations, entInfraEntities...)
	return defaultConfigurations, nil
}

func (impl *InfraConfigClientImpl) ConvertToProfilePlatformMap(infraProfileConfigurationEntities []*repository.InfraProfileConfigurationEntity,
	profilesMap map[int]*v1.ProfileBeanDto, profilePlatforms []*repository.ProfilePlatformMapping) (map[int]map[string][]*v1.ConfigurationBean, error) {
	// Create a map to track profileId and platform presence
	profilePlatformTracker := make(map[int]map[string]bool)

	// Initialize the tracker with platforms from profilePlatforms
	for _, profilePlatform := range profilePlatforms {
		if _, exists := profilePlatformTracker[profilePlatform.ProfileId]; !exists {
			profilePlatformTracker[profilePlatform.ProfileId] = make(map[string]bool)
		}
		profilePlatformTracker[profilePlatform.ProfileId][profilePlatform.Platform] = true
	}

	profilePlatformMap := make(map[int]map[string][]*v1.ConfigurationBean)

	for _, infraProfileConfiguration := range infraProfileConfigurationEntities {
		profileId := infraProfileConfiguration.ProfilePlatformMapping.ProfileId
		profile, ok := profilesMap[profileId]

		if !ok || profile == nil {
			continue
		}

		// Initialize the inner map for the current ProfileId if it doesn't exist
		if _, exists := profilePlatformMap[profileId]; !exists {
			profilePlatformMap[profileId] = make(map[string][]*v1.ConfigurationBean)
		}

		// Convert entity to ConfigurationBean
		configurationBean, err := impl.getConfigurationBean(infraProfileConfiguration, profile.GetName())
		if err != nil {
			impl.logger.Errorw("failed to get configurations for profile", "err", err, "profileName", profile.GetName())
			errMsg := fmt.Sprintf("failed to get configurations for profile '%s'", profile.GetName())
			return nil, globalUtil.NewApiError(http.StatusBadRequest, errMsg, errMsg)
		}
		platform := infraProfileConfiguration.ProfilePlatformMapping.Platform
		if len(platform) == 0 {
			platform = v1.RUNNER_PLATFORM
		}

		// Append the ConfigurationBean to the list under the appropriate platform in the inner map
		profilePlatformMap[profileId][platform] = append(
			profilePlatformMap[profileId][platform],
			configurationBean,
		)

		// Mark the platform as processed
		if platformTracker, exists := profilePlatformTracker[profileId]; exists {
			platformTracker[platform] = false
		}
	}
	// Ensure all platforms from profilePlatformTracker are included in the result map
	for profileId, platforms := range profilePlatformTracker {
		if _, exists := profilePlatformMap[profileId]; !exists {
			profilePlatformMap[profileId] = make(map[string][]*v1.ConfigurationBean)
		}
		for platform, isMissing := range platforms {
			if isMissing {
				profilePlatformMap[profileId][platform] = []*v1.ConfigurationBean{}
			}
		}
	}
	return profilePlatformMap, nil
}

func (impl *InfraConfigClientImpl) HandleInfraConfigTriggerAudit(workflowId int, triggeredBy int32, infraConfigs map[string]*v1.InfraConfig) error {
	for platform, infraConfig := range infraConfigs {
		supportedConfigKeys := util.GetConfigKeysMapForPlatform(platform)
		err := impl.handleInfraConfigTriggerAudit(supportedConfigKeys, workflowId, triggeredBy, infraConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl *InfraConfigClientImpl) validateConfig(platformName string, platformConfigurations, defaultConfigurations []*v1.ConfigurationBean, skipError bool) (v1.InfraConfigKeys, error) {
	supportedConfigKeyMap := util.GetConfigKeysMapForPlatform(platformName)
	cpuConfigKeys := impl.getCPUConfigFactory().getConfigKeys()
	if err := impl.getCPUConfigFactory().validate(platformConfigurations, defaultConfigurations); err != nil {
		for _, cpuConfigKey := range cpuConfigKeys {
			supportedConfigKeyMap = supportedConfigKeyMap.MarkUnConfigured(cpuConfigKey)
		}
		if !skipError {
			return supportedConfigKeyMap, err
		}
	} else {
		for _, cpuConfigKey := range cpuConfigKeys {
			supportedConfigKeyMap = supportedConfigKeyMap.MarkConfigured(cpuConfigKey)
		}
	}

	memConfigKeys := impl.getMemoryConfigFactory().getConfigKeys()
	if err := impl.getMemoryConfigFactory().validate(platformConfigurations, defaultConfigurations); err != nil {
		for _, memConfigKey := range memConfigKeys {
			supportedConfigKeyMap = supportedConfigKeyMap.MarkUnConfigured(memConfigKey)
		}
		if !skipError {
			return supportedConfigKeyMap, err
		}
	} else {
		for _, memConfigKey := range memConfigKeys {
			supportedConfigKeyMap = supportedConfigKeyMap.MarkConfigured(memConfigKey)
		}
	}

	timeoutConfigKeys := impl.getTimeoutConfigFactory().getConfigKeys()
	if err := impl.getTimeoutConfigFactory().validate(platformConfigurations, defaultConfigurations); err != nil {
		for _, timeoutConfigKey := range timeoutConfigKeys {
			supportedConfigKeyMap = supportedConfigKeyMap.MarkUnConfigured(timeoutConfigKey)
		}
		if !skipError {
			return supportedConfigKeyMap, err
		}
	} else {
		for _, timeoutConfigKey := range timeoutConfigKeys {
			supportedConfigKeyMap = supportedConfigKeyMap.MarkConfigured(timeoutConfigKey)
		}
	}

	supportedConfigKeyMap, err := impl.validateEntConfig(supportedConfigKeyMap, platformConfigurations, defaultConfigurations, skipError)
	if !skipError && err != nil {
		return supportedConfigKeyMap, err
	}
	return supportedConfigKeyMap, nil
}

func (impl *InfraConfigClientImpl) getConfigurationBean(infraProfileConfiguration *repository.InfraProfileConfigurationEntity, profileName string) (*v1.ConfigurationBean, error) {
	valueString := infraProfileConfiguration.ValueString
	// handling for old values
	if len(valueString) == 0 && infraProfileConfiguration.Unit > 0 {
		valueString = strconv.FormatFloat(infraProfileConfiguration.Value, 'f', -1, 64)
	}
	valueInterface, valueCount, err := impl.convertValueStringToInterface(util.GetConfigKeyStr(infraProfileConfiguration.Key), valueString)
	if err != nil {
		return &v1.ConfigurationBean{}, err
	}
	return &v1.ConfigurationBean{
		ConfigurationBeanAbstract: v1.ConfigurationBeanAbstract{
			Id:          infraProfileConfiguration.Id,
			Key:         util.GetConfigKeyStr(infraProfileConfiguration.Key),
			Unit:        util.GetUnitSuffixStr(infraProfileConfiguration.Key, infraProfileConfiguration.Unit),
			ProfileId:   infraProfileConfiguration.ProfilePlatformMapping.ProfileId,
			Active:      impl.isConfigActive(util.GetConfigKeyStr(infraProfileConfiguration.Key), valueCount, infraProfileConfiguration.Active),
			ProfileName: profileName,
		},
		Value: valueInterface,
	}, nil
}
