package adapter

import (
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean"
	"github.com/devtron-labs/devtron/pkg/infraConfig/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/util"
	"math"
	"strconv"
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
	valueString := infraProfileConfiguration.ValueString
	//handle old values
	if len(valueString) == 0 && infraProfileConfiguration.Unit > 0 {
		valueString = strconv.FormatFloat(infraProfileConfiguration.Value, 'f', -1, 64)
	}
	return &bean.ConfigurationBean{
		ConfigurationBeanAbstract: bean.ConfigurationBeanAbstract{
			Id:  infraProfileConfiguration.Id,
			Key: util.GetConfigKeyStr(infraProfileConfiguration.Key),

			Unit:        util.GetUnitSuffixStr(infraProfileConfiguration.Key, infraProfileConfiguration.Unit),
			ProfileId:   infraProfileConfiguration.ProfileId,
			Active:      infraProfileConfiguration.Active,
			ProfileName: profileName,
		},
		Value: valueString,
	}
}

func getInfraProfileEntity(configurationBean *bean.ConfigurationBean, profileBean *bean.ProfileBean, platform string, userId int32) *bean.InfraProfileConfigurationEntity {

	infraProfile := &bean.InfraProfileConfigurationEntity{
		Id:          configurationBean.Id,
		Key:         util.GetConfigKey(configurationBean.Key),
		ValueString: formatFloatIfNeeded(configurationBean.Key, configurationBean.Value),
		Unit:        util.GetUnitSuffix(configurationBean.Key, configurationBean.Unit),
		ProfileId:   profileBean.Id,
		Platform:    platform,
		Active:      configurationBean.Active,
		AuditLog:    sql.NewDefaultAuditLog(userId),
	}
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

func GetV0ProfileBean(profileBean *bean.ProfileBean) *bean.ProfileBeanV0 {
	if profileBean == nil {
		return &bean.ProfileBeanV0{}
	}
	ciRunnerConfig := profileBean.Configurations[util.CI_RUNNER_PLATFORM]
	return &bean.ProfileBeanV0{
		ProfileBeanAbstract: bean.ProfileBeanAbstract{
			Id:          profileBean.Id,
			Name:        profileBean.Name,
			Description: profileBean.Description,
			Active:      profileBean.Active,
			Type:        profileBean.Type,
			AppCount:    profileBean.AppCount,
			CreatedBy:   profileBean.CreatedBy,
			CreatedOn:   profileBean.CreatedOn,
			UpdatedBy:   profileBean.UpdatedBy,
			UpdatedOn:   profileBean.UpdatedOn,
		},
		Configurations: GetV0ConfigurationBeans(ciRunnerConfig),
	}

}

func GetV1ProfileBean(profileBean *bean.ProfileBeanV0) *bean.ProfileBean {
	if profileBean == nil {
		return nil
	}
	return &bean.ProfileBean{
		ProfileBeanAbstract: bean.ProfileBeanAbstract{
			Id:          profileBean.Id,
			Name:        profileBean.Name,
			Description: profileBean.Description,
			Active:      profileBean.Active,
			Type:        profileBean.Type,
			AppCount:    profileBean.AppCount,
			CreatedBy:   profileBean.CreatedBy,
			CreatedOn:   profileBean.CreatedOn,
			UpdatedBy:   profileBean.UpdatedBy,
			UpdatedOn:   profileBean.UpdatedOn,
		},
		Configurations: map[string][]*bean.ConfigurationBean{util.CI_RUNNER_PLATFORM: GetV1ConfigurationBeans(profileBean.Configurations)},
	}

}

func GetV1ConfigurationBeans(configBeans []bean.ConfigurationBeanV0) []*bean.ConfigurationBean {
	if len(configBeans) == 0 {
		return nil
	}
	resp := make([]*bean.ConfigurationBean, 0)
	for _, configBean := range configBeans {
		valueString := strconv.FormatFloat(configBean.Value, 'f', -1, 64)

		configBeanV1 := &bean.ConfigurationBean{
			ConfigurationBeanAbstract: bean.ConfigurationBeanAbstract{
				Id:          configBean.Id,
				Key:         configBean.Key,
				Unit:        configBean.Unit,
				ProfileName: configBean.ProfileName,
				ProfileId:   configBean.ProfileId,
				Active:      configBean.Active,
			},
			Value: valueString,
		}
		resp = append(resp, configBeanV1)
	}
	return resp
}

func GetV0ConfigurationBeans(configBeans []*bean.ConfigurationBean) []bean.ConfigurationBeanV0 {
	if len(configBeans) == 0 {
		return []bean.ConfigurationBeanV0{}
	}

	resp := make([]bean.ConfigurationBeanV0, 0)
	for _, configBean := range configBeans {
		valueFloat, _ := strconv.ParseFloat(configBean.Value, 64)

		beanv0 := bean.ConfigurationBeanV0{
			ConfigurationBeanAbstract: bean.ConfigurationBeanAbstract{
				Id:          configBean.Id,
				Key:         configBean.Key,
				Unit:        configBean.Unit,
				ProfileName: configBean.ProfileName,
				ProfileId:   configBean.ProfileId,
				Active:      configBean.Active,
			},
			Value: valueFloat,
		}
		resp = append(resp, beanv0)
	}

	return resp

}
