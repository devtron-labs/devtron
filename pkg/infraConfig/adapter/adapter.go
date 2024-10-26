package adapter

import (
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean"
	"github.com/devtron-labs/devtron/pkg/infraConfig/util"
	util2 "github.com/devtron-labs/devtron/util"
	"math"
	"strconv"
	"time"
)

func ConvertToPlatformMap(infraProfileConfigurationEntities []*bean.InfraProfileConfigurationEntity, profileName string) map[string][]*bean.ConfigurationBean {
	platformMap := make(map[string][]*bean.ConfigurationBean)

	for _, infraProfileConfiguration := range infraProfileConfigurationEntities {
		ConfigurationBean := getConfigurationBean(infraProfileConfiguration, profileName)

		// Add the ConfigurationBean to the corresponding platform entry in the map
		platformMap[infraProfileConfiguration.Platform] = append(platformMap[infraProfileConfiguration.Platform], ConfigurationBean)
	}

	return platformMap
}

// ConvertFromPlatformMap converts map[platform][]*ConfigurationBean back to []InfraProfileConfigurationEntity
func ConvertFromPlatformMap(platformMap map[string][]*bean.ConfigurationBean, profileBean *bean.ProfileBean, userId int32) []*bean.InfraProfileConfigurationEntity {
	var entities []*bean.InfraProfileConfigurationEntity
	for platform, beans := range platformMap {
		for _, configBean := range beans {
			entity := getInfraProfileEntity(configBean, profileBean, platform, userId)
			entities = append(entities, entity)
		}
	}
	return entities
}

func getConfigurationBean(infraProfileConfiguration *bean.InfraProfileConfigurationEntity, profileName string) *bean.ConfigurationBean {
	return &bean.ConfigurationBean{
		Id:          infraProfileConfiguration.Id,
		Key:         util.GetConfigKeyStr(infraProfileConfiguration.Key),
		Value:       infraProfileConfiguration.Value,
		Unit:        util.GetUnitSuffixStr(infraProfileConfiguration.Key, infraProfileConfiguration.Unit),
		ProfileId:   infraProfileConfiguration.ProfileId,
		Active:      infraProfileConfiguration.Active,
		ProfileName: profileName,
	}
}

func getInfraProfileEntity(configurationBean *bean.ConfigurationBean, profileBean *bean.ProfileBean, platform string, userId int32) *bean.InfraProfileConfigurationEntity {

	infraProfile := &bean.InfraProfileConfigurationEntity{
		Id:        configurationBean.Id,
		Key:       util.GetConfigKey(configurationBean.Key),
		Value:     formatFloatIfNeeded(configurationBean.Key, configurationBean.Value),
		Unit:      util.GetUnitSuffix(configurationBean.Key, configurationBean.Unit),
		ProfileId: profileBean.Id,
		Platform:  platform,
		Active:    configurationBean.Active,
	}
	infraProfile.UpdatedOn = time.Now()
	infraProfile.UpdatedBy = userId
	if profileBean.Name == util.DEFAULT_PROFILE_NAME {
		infraProfile.Active = true
	}
	return infraProfile
}

func formatFloatIfNeeded(configKey util.ConfigKeyStr, configValue string) string {
	if configKey == util.CPU_LIMIT ||
		configKey == util.CPU_REQUEST ||
		configKey == util.MEMORY_LIMIT ||
		configKey == util.MEMORY_REQUEST {
		valueFloat, _ := strconv.ParseFloat(configValue, 64)
		truncateValue := util2.TruncateFloat(valueFloat, 2)
		return strconv.FormatFloat(truncateValue, 'f', -1, 64)
	}

	if configKey == util.TIME_OUT {
		valueFloat, _ := strconv.ParseFloat(configValue, 64)
		modifiedValue := math.Min(math.Floor(valueFloat), math.MaxInt64)
		return strconv.FormatFloat(modifiedValue, 'f', -1, 64)
	}

	return configValue
}
