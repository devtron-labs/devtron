package adaptor

import (
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/bean/configMapBean"
	bean2 "github.com/devtron-labs/devtron/pkg/configDiff/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/adapter"
)

func GetConfigProperty(id int, name string, configType configMapBean.ResourceType, State bean2.ConfigState) *bean2.ConfigProperty {
	return &bean2.ConfigProperty{
		Id:          id,
		Name:        name,
		Type:        configType,
		ConfigState: State,
	}
}

func GetCmCsAppAndEnvLevelMap(cMCSNamesAppLevel, cMCSNamesEnvLevel []configMapBean.ConfigNameAndType) (map[string]*bean2.ConfigProperty, map[string]*bean2.ConfigProperty) {
	cMCSNamesAppLevelMap, cMCSNamesEnvLevelMap := make(map[string]*bean2.ConfigProperty, len(cMCSNamesAppLevel)), make(map[string]*bean2.ConfigProperty, len(cMCSNamesEnvLevel))

	for _, cmcs := range cMCSNamesAppLevel {
		property := GetConfigProperty(cmcs.Id, cmcs.Name, cmcs.Type, bean2.PublishedConfigState)
		cMCSNamesAppLevelMap[property.GetKey()] = property
	}
	for _, cmcs := range cMCSNamesEnvLevel {
		property := GetConfigProperty(cmcs.Id, cmcs.Name, cmcs.Type, bean2.PublishedConfigState)
		cMCSNamesEnvLevelMap[property.GetKey()] = property
	}
	return cMCSNamesAppLevelMap, cMCSNamesEnvLevelMap
}

func ConfigListConvertor(r bean3.ConfigList) configMapBean.ConfigsList {
	pipelineConfigData := make([]*configMapBean.ConfigData, 0, len(r.ConfigData))
	for _, item := range r.ConfigData {
		pipelineConfigData = append(pipelineConfigData, adapter.ConvertConfigDataToPipelineConfigData(item))
	}
	return configMapBean.ConfigsList{ConfigData: pipelineConfigData}
}

func SecretListConvertor(r bean3.SecretList) configMapBean.SecretsList {
	pipelineConfigData := make([]*configMapBean.ConfigData, 0, len(r.ConfigData))
	for _, item := range r.ConfigData {
		pipelineConfigData = append(pipelineConfigData, adapter.ConvertConfigDataToPipelineConfigData(item))
	}
	return configMapBean.SecretsList{ConfigData: pipelineConfigData}
}

func ReverseConfigListConvertor(r configMapBean.ConfigsList) bean3.ConfigList {
	configData := make([]*bean3.ConfigData, 0, len(r.ConfigData))
	for _, item := range r.ConfigData {
		configData = append(configData, adapter.ConvertPipelineConfigDataToConfigData(item))
	}
	return bean3.ConfigList{ConfigData: configData}
}

func ReverseSecretListConvertor(r configMapBean.SecretsList) bean3.SecretList {
	configData := make([]*bean3.ConfigData, 0, len(r.ConfigData))
	for _, item := range r.ConfigData {
		configData = append(configData, adapter.ConvertPipelineConfigDataToConfigData(item))
	}
	return bean3.SecretList{ConfigData: configData}
}
