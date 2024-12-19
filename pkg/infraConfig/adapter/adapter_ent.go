package adapter

import (
	"errors"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
)

func ConvertToProfilePlatformMap(infraProfileConfigurationEntities []*repository.InfraProfileConfigurationEntity, profilesMap map[int]bean.ProfileBeanDto, profilePlatforms []*repository.ProfilePlatformMapping) (map[int]map[string][]*bean.ConfigurationBean, error) {
	return nil, errors.New("method ConvertToProfilePlatformMap not implemented")
}

func SetProfilePlatformMappings(platforms []string, infraProfileId int, userId int32) []*repository.ProfilePlatformMapping {
	return nil
}

func GetV0ProfileBeans(profileBeans []bean.ProfileBeanDto) []bean.ProfileBeanV0 {
	return nil
}

func GetV0Identifiers(identifiersV1 []*bean.Identifier) []*bean.IdentifierV0 {
	return nil
}

func GetV0ProfileName(profileNames []string) []string {
	return nil
}

func FillMissingConfigurationsForThePayloadV0(profileToUpdate *bean.ProfileBeanDto, platformMapConfigs map[string][]*bean.ConfigurationBean) {
	return
}

func LoadInfraConfigFromCM(driverOpts map[string]string, platform string, defaultConfigKeyMap map[bean.ConfigKeyStr]bool) ([]*repository.InfraProfileConfigurationEntity, error) {
	return nil, errors.New("method LoadInfraConfigFromCM not implemented")
}

func UpdateProfileMissingConfigurationsWithDefaultV0(profile bean.ProfileBeanDto, defaultConfigurationsMap map[string][]*bean.ConfigurationBean) bean.ProfileBeanDto {
	return bean.ProfileBeanDto{}
}

func UpdateProfileMissingConfigurationsWithDefault(profile bean.ProfileBeanDto, defaultConfigurationsMap map[string][]*bean.ConfigurationBean) bean.ProfileBeanDto {
	return bean.ProfileBeanDto{}
}

func LoadDefaultValueFromEnv(configKeyStr bean.ConfigKeyStr, platform string) (*repository.InfraProfileConfigurationEntity, error) {
	return nil, errors.New("method LoadDefaultValueFromEnv not implemented")
}
