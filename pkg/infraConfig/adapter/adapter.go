package adapter

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/devtron-labs/devtron/pkg/infraConfig/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/util"
	"math"
	"strconv"
)

func ConvertToPlatformMap(infraProfileConfigurationEntities []*repository.InfraProfileConfigurationEntity, profileName string) (map[string][]*bean.ConfigurationBean, error) {
	// Validate input parameters
	if len(infraProfileConfigurationEntities) == 0 {
		return nil, fmt.Errorf("input infraProfileConfigurationEntities is empty")
	}
	if profileName == "" {
		return nil, fmt.Errorf("profileName cannot be empty")
	}
	platformMap := make(map[string][]*bean.ConfigurationBean)
	for _, infraProfileConfiguration := range infraProfileConfigurationEntities {
		if infraProfileConfiguration == nil {
			return nil, fmt.Errorf("infraProfileConfiguration for profile %s is nil", profileName)
		}

		configurationBean, err := GetConfigurationBean(infraProfileConfiguration, profileName)
		if err != nil {
			return nil, fmt.Errorf("failed to get configuration bean for profile from infraConfiguration '%s': %w", profileName, err)
		}
		platform := infraProfileConfiguration.ProfilePlatformMapping.Platform
		if len(platform) == 0 {
			platform = bean.RUNNER_PLATFORM
		}

		// Add the ConfigurationBean to the corresponding platform entry in the map
		platformMap[platform] = append(platformMap[platform], configurationBean)
	}
	return platformMap, nil
}

// ConvertFromPlatformMap converts map[platform][]*ConfigurationBean back to []InfraProfileConfigurationEntity
func ConvertFromPlatformMap(platformMap map[string][]*bean.ConfigurationBean, profileBean *bean.ProfileBeanDto, userId int32) []*repository.InfraProfileConfigurationEntity {
	var entities []*repository.InfraProfileConfigurationEntity
	for platform, beans := range platformMap {
		for _, configBean := range beans {
			entity := getInfraProfileEntity(configBean, profileBean, platform, userId)
			entities = append(entities, entity)
		}
	}
	return entities
}

// Function to convert valueString to interface{} based on key
func convertValueStringToInterface(configKey bean.ConfigKeyStr, valueString string) (interface{}, error) {
	switch configKey {
	case bean.CPU_LIMIT, bean.CPU_REQUEST, bean.MEMORY_LIMIT, bean.MEMORY_REQUEST:
		// Convert string to float64 and truncate to 2 decimal places
		valueFloat, err := strconv.ParseFloat(valueString, 64)
		truncateValue := util2.TruncateFloat(valueFloat, 2)
		return truncateValue, err // Returning float64 for resource values
	case bean.TIME_OUT:
		// Convert string to float64 and ensure it's within integer range
		valueFloat, err := strconv.ParseFloat(valueString, 64)
		modifiedValue := math.Min(math.Floor(valueFloat), math.MaxInt64)
		return modifiedValue, err // Returning float64 for timeout

	// Add more cases as needed for different config keys
	default:
		// Default case, return the string as is
		err := errors.New(fmt.Sprintf("unsupported key found %s", configKey))
		return nil, err
	}
}

func GetConfigurationBean(infraProfileConfiguration *repository.InfraProfileConfigurationEntity, profileName string) (*bean.ConfigurationBean, error) {
	valueString := infraProfileConfiguration.ValueString
	//handle old values
	if len(valueString) == 0 && infraProfileConfiguration.Unit > 0 {
		valueString = strconv.FormatFloat(infraProfileConfiguration.Value, 'f', -1, 64)
	}
	valueInterface, err := convertValueStringToInterface(utils.GetConfigKeyStr(infraProfileConfiguration.Key), valueString)
	if err != nil {
		return &bean.ConfigurationBean{}, err
	}
	return &bean.ConfigurationBean{
		ConfigurationBeanAbstract: bean.ConfigurationBeanAbstract{
			Id:                       infraProfileConfiguration.Id,
			Key:                      utils.GetConfigKeyStr(infraProfileConfiguration.Key),
			Unit:                     utils.GetUnitSuffixStr(infraProfileConfiguration.Key, infraProfileConfiguration.Unit),
			Active:                   infraProfileConfiguration.Active,
			ProfileId:                infraProfileConfiguration.ProfilePlatformMapping.ProfileId,
			ProfileName:              profileName,
			ProfilePlatformMappingId: infraProfileConfiguration.ProfilePlatformMapping.Id,
		},
		Value: valueInterface,
	}, nil
}

func getInfraProfileEntity(configurationBean *bean.ConfigurationBean, profileBean *bean.ProfileBeanDto, platform string, userId int32) *repository.InfraProfileConfigurationEntity {

	infraProfile := &repository.InfraProfileConfigurationEntity{
		Id:          configurationBean.Id,
		Key:         utils.GetConfigKey(configurationBean.Key),
		ValueString: FormatTypedValueAsString(configurationBean.Key, configurationBean.Value),
		Unit:        utils.GetUnitSuffix(configurationBean.Key, configurationBean.Unit),
		Active:      configurationBean.Active,
		UniqueId:    repository.GetUniqueId(profileBean.Id, platform),
		ProfileId:   profileBean.Id, // maintained for backward compatibility
		ProfilePlatformMapping: &repository.ProfilePlatformMapping{
			ProfileId: profileBean.Id,
			Platform:  platform,
		},
		AuditLog: sql.NewDefaultAuditLog(userId),
	}
	if profileBean.Name == bean.GLOBAL_PROFILE_NAME {
		infraProfile.Active = true
	}
	return infraProfile
}

func FormatTypedValueAsString(configKey bean.ConfigKeyStr, configValue interface{}) string {
	if configKey == bean.CPU_LIMIT ||
		configKey == bean.CPU_REQUEST ||
		configKey == bean.MEMORY_LIMIT ||
		configKey == bean.MEMORY_REQUEST {
		var valueFloat float64
		// Handle string input or directly as float64
		switch v := configValue.(type) {
		case string:
			valueFloat, _ = strconv.ParseFloat(v, 64)
		case float64:
			valueFloat = v
		}
		// Truncate and format the float value
		truncateValue := util2.TruncateFloat(valueFloat, 2)
		return strconv.FormatFloat(truncateValue, 'f', -1, 64)
		//valueFloat, _ := strconv.ParseFloat(configValue.(float64), 64)
	}

	if configKey == bean.TIME_OUT {
		var valueFloat float64
		switch v := configValue.(type) {
		case string:
			valueFloat, _ = strconv.ParseFloat(v, 64)
		case float64:
			valueFloat = v
		}
		//valueFloat, _ := strconv.ParseFloat(configValue, 64)
		modifiedValue := math.Min(math.Floor(valueFloat), math.MaxInt64)
		return strconv.FormatFloat(modifiedValue, 'f', -1, 64)
	}

	return configValue.(string)
}

func GetV0ProfileBean(profileBean *bean.ProfileBeanDto) *bean.ProfileBeanV0 {
	if profileBean == nil {
		return &bean.ProfileBeanV0{}
	}
	profileName := profileBean.Name
	if profileName == bean.GLOBAL_PROFILE_NAME {
		profileName = bean.DEFAULT_PROFILE_NAME
	}

	profileType := profileBean.Type
	if profileType == bean.GLOBAL {
		profileType = bean.DEFAULT
	}

	ciRunnerConfig := profileBean.Configurations[bean.RUNNER_PLATFORM]
	return &bean.ProfileBeanV0{
		ProfileBeanAbstract: bean.ProfileBeanAbstract{
			Id:               profileBean.Id,
			Name:             profileName,
			Description:      profileBean.Description,
			BuildxDriverType: profileBean.BuildxDriverType,
			Active:           profileBean.Active,
			Type:             profileType,
			AppCount:         profileBean.AppCount,
			CreatedBy:        profileBean.CreatedBy,
			CreatedOn:        profileBean.CreatedOn,
			UpdatedBy:        profileBean.UpdatedBy,
			UpdatedOn:        profileBean.UpdatedOn,
		},
		Configurations: GetV0ConfigurationBeans(ciRunnerConfig),
	}

}

func GetV1ProfileBean(profileBean *bean.ProfileBeanV0) *bean.ProfileBeanDto {
	if profileBean == nil {
		return nil
	}
	profileName := profileBean.Name
	if profileName == bean.DEFAULT_PROFILE_NAME {
		profileName = bean.GLOBAL_PROFILE_NAME
	}
	profileType := profileBean.Type
	if profileType == bean.DEFAULT {
		profileType = bean.GLOBAL
	}
	return &bean.ProfileBeanDto{
		ProfileBeanAbstract: bean.ProfileBeanAbstract{
			Id:               profileBean.Id,
			Name:             profileName,
			Description:      profileBean.Description,
			Active:           profileBean.Active,
			Type:             profileType,
			AppCount:         profileBean.AppCount,
			CreatedBy:        profileBean.CreatedBy,
			CreatedOn:        profileBean.CreatedOn,
			UpdatedBy:        profileBean.UpdatedBy,
			UpdatedOn:        profileBean.UpdatedOn,
			BuildxDriverType: profileBean.BuildxDriverType,
		},
		Configurations: map[string][]*bean.ConfigurationBean{bean.RUNNER_PLATFORM: GetV1ConfigurationBeans(profileBean.Configurations, profileName)},
	}

}

func GetV1ConfigurationBeans(configBeans []bean.ConfigurationBeanV0, profileName string) []*bean.ConfigurationBean {
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
				Active:      configBean.Active,
				ProfileId:   configBean.ProfileId,
				ProfileName: profileName,
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
		// Use the GetTypedValue function to decode the value
		typedValue, _ := utils.GetTypedValue(configBean.Key, configBean.Value)
		// Cast the returned value to float64 for supported keys
		valueFloat, ok := typedValue.(float64)
		if !ok {
			//here skipping the value for the NodeSelectors and TolerationsKey
			continue
		}
		profileName := configBean.ProfileName
		if profileName == bean.GLOBAL_PROFILE_NAME {
			profileName = bean.DEFAULT_PROFILE_NAME
		}

		beanv0 := bean.ConfigurationBeanV0{
			ConfigurationBeanAbstract: bean.ConfigurationBeanAbstract{
				Id:          configBean.Id,
				Key:         configBean.Key,
				Unit:        configBean.Unit,
				Active:      configBean.Active,
				ProfileId:   configBean.ProfileId,
				ProfileName: profileName,
			},
			Value: valueFloat,
		}
		resp = append(resp, beanv0)
	}
	return resp
}

func ConvertToProfileBean(infraProfile *repository.InfraProfileEntity) bean.ProfileBeanDto {
	profileType := bean.GLOBAL
	if infraProfile.Name != bean.GLOBAL_PROFILE_NAME {
		profileType = bean.NORMAL
	}
	return bean.ProfileBeanDto{
		ProfileBeanAbstract: bean.ProfileBeanAbstract{
			Id:               infraProfile.Id,
			Name:             infraProfile.Name,
			Type:             profileType,
			Description:      infraProfile.Description,
			BuildxDriverType: infraProfile.BuildxDriverType,
			Active:           infraProfile.Active,
			CreatedBy:        infraProfile.CreatedBy,
			CreatedOn:        infraProfile.CreatedOn,
			UpdatedBy:        infraProfile.UpdatedBy,
			UpdatedOn:        infraProfile.UpdatedOn,
		},
	}
}

func ConvertToInfraProfileEntity(profileBean *bean.ProfileBeanDto) *repository.InfraProfileEntity {
	return &repository.InfraProfileEntity{
		Id:               profileBean.Id,
		Name:             profileBean.Name,
		Description:      profileBean.Description,
		BuildxDriverType: profileBean.BuildxDriverType,
	}
}

func LoadCiLimitCpu(infraConfig *bean.InfraConfig) (*repository.InfraProfileConfigurationEntity, error) {
	val, suffix, err := units.ParseValAndUnit(infraConfig.CiLimitCpu)
	if err != nil {
		return nil, err
	}
	return &repository.InfraProfileConfigurationEntity{
		Key:         bean.CPULimitKey,
		ValueString: strconv.FormatFloat(val, 'f', -1, 64),
		Unit:        units.CPUUnitStr(suffix).GetCPUUnit(),
		ProfilePlatformMapping: &repository.ProfilePlatformMapping{
			Platform: bean.RUNNER_PLATFORM,
		},
	}, nil

}

func LoadCiLimitMem(infraConfig *bean.InfraConfig) (*repository.InfraProfileConfigurationEntity, error) {
	val, suffix, err := units.ParseValAndUnit(infraConfig.CiLimitMem)
	if err != nil {
		return nil, err
	}
	return &repository.InfraProfileConfigurationEntity{
		Key:         bean.MemoryLimitKey,
		ValueString: strconv.FormatFloat(val, 'f', -1, 64),
		Unit:        units.MemoryUnitStr(suffix).GetMemoryUnit(),
		ProfilePlatformMapping: &repository.ProfilePlatformMapping{
			Platform: bean.RUNNER_PLATFORM,
		},
	}, nil
}

func LoadCiReqCpu(infraConfig *bean.InfraConfig) (*repository.InfraProfileConfigurationEntity, error) {
	val, suffix, err := units.ParseValAndUnit(infraConfig.CiReqCpu)
	if err != nil {
		return nil, err
	}
	return &repository.InfraProfileConfigurationEntity{
		Key:         bean.CPURequestKey,
		ValueString: strconv.FormatFloat(val, 'f', -1, 64),
		Unit:        units.CPUUnitStr(suffix).GetCPUUnit(),
		ProfilePlatformMapping: &repository.ProfilePlatformMapping{
			Platform: bean.RUNNER_PLATFORM,
		},
	}, nil
}

func LoadCiReqMem(infraConfig *bean.InfraConfig) (*repository.InfraProfileConfigurationEntity, error) {
	val, suffix, err := units.ParseValAndUnit(infraConfig.CiReqMem)
	if err != nil {
		return nil, err
	}

	return &repository.InfraProfileConfigurationEntity{
		Key:         bean.MemoryRequestKey,
		ValueString: strconv.FormatFloat(val, 'f', -1, 64),
		Unit:        units.MemoryUnitStr(suffix).GetMemoryUnit(),
		ProfilePlatformMapping: &repository.ProfilePlatformMapping{
			Platform: bean.RUNNER_PLATFORM,
		},
	}, nil
}

func LoadDefaultTimeout(infraConfig *bean.InfraConfig) (*repository.InfraProfileConfigurationEntity, error) {
	return &repository.InfraProfileConfigurationEntity{
		Key:         bean.TimeOutKey,
		ValueString: strconv.FormatInt(infraConfig.CiDefaultTimeout, 10),
		Unit:        units.SecondStr.GetTimeUnit(),
		ProfilePlatformMapping: &repository.ProfilePlatformMapping{
			Platform: bean.RUNNER_PLATFORM,
		},
	}, nil
}
func LoadInfraConfigInEntities(infraConfig *bean.InfraConfig, nodeSelectorLabel []string, taintKey, taintValue string) ([]*repository.InfraProfileConfigurationEntity, error) {
	cpuLimit, err := LoadCiLimitCpu(infraConfig)
	if err != nil {
		return nil, err
	}
	memLimit, err := LoadCiLimitMem(infraConfig)
	if err != nil {
		return nil, err
	}
	cpuReq, err := LoadCiReqCpu(infraConfig)
	if err != nil {
		return nil, err
	}
	memReq, err := LoadCiReqMem(infraConfig)
	if err != nil {
		return nil, err
	}
	timeout, err := LoadDefaultTimeout(infraConfig)
	if err != nil {
		return nil, err
	}
	defaultConfigurations := []*repository.InfraProfileConfigurationEntity{cpuLimit, memLimit, cpuReq, memReq, timeout}
	return defaultConfigurations, nil
}
