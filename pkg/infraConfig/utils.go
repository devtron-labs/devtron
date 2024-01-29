package infraConfig

import "github.com/devtron-labs/devtron/util"

func UpdateProfileMissingConfigurationsWithDefault(profile ProfileBean, defaultConfigurations []ConfigurationBean) ProfileBean {
	extraConfigurations := make([]ConfigurationBean, 0)
	for _, defaultConfiguration := range defaultConfigurations {
		// if profile doesn't have the default configuration, add it to the profile
		if !util.Contains(profile.Configurations, func(config ConfigurationBean) bool {
			return config.Key == defaultConfiguration.Key
		}) {
			extraConfigurations = append(extraConfigurations, defaultConfiguration)
		}
	}
	profile.Configurations = append(profile.Configurations, extraConfigurations...)
	return profile
}
