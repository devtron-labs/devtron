package adaptor

import (
	pluginBean "github.com/devtron-labs/devtron/pkg/plugin/bean"
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
)

func GetPluginParentMetadataDbObject(pluginDto *pluginBean.PluginParentMetadataDto, userId int32) *repository.PluginParentMetadata {
	return repository.NewPluginParentMetadata().CreateAuditLog(userId).
		WithBasicMetadata(pluginDto.Name, pluginDto.PluginIdentifier, pluginDto.Description, pluginDto.Icon, repository.PLUGIN_TYPE_SHARED)
}

func GetPluginVersionMetadataDbObject(pluginDto *pluginBean.PluginParentMetadataDto, userId int32) *repository.PluginMetadata {
	versionDto := pluginDto.Versions.DetailedPluginVersionData[0]
	return repository.NewPluginVersionMetadata().CreateAuditLog(userId).WithBasicMetadata(pluginDto.Name, versionDto.Description, versionDto.Version, versionDto.DocLink)
}

func GetPluginStepDbObject(pluginStepDto *pluginBean.PluginStepsDto, pluginVersionMetadataId int, userId int32) *repository.PluginStep {
	return &repository.PluginStep{
		PluginId:            pluginVersionMetadataId,
		Name:                pluginStepDto.Name,
		Description:         pluginStepDto.Description,
		Index:               1,
		StepType:            repository.PLUGIN_STEP_TYPE_INLINE,
		RefPluginId:         pluginStepDto.RefPluginId,
		OutputDirectoryPath: pluginStepDto.OutputDirectoryPath,
		DependentOnStep:     pluginStepDto.DependentOnStep,
		AuditLog:            sql.NewDefaultAuditLog(userId),
	}
}
func GetPluginPipelineScriptDbObject(pluginPipelineScript *pluginBean.PluginPipelineScript, userId int32) *repository.PluginPipelineScript {
	return &repository.PluginPipelineScript{
		Script:                   pluginPipelineScript.Script,
		StoreScriptAt:            pluginPipelineScript.StoreScriptAt,
		Type:                     pluginPipelineScript.Type,
		DockerfileExists:         pluginPipelineScript.DockerfileExists,
		MountPath:                pluginPipelineScript.MountPath,
		MountCodeToContainer:     pluginPipelineScript.MountCodeToContainer,
		MountCodeToContainerPath: pluginPipelineScript.MountCodeToContainerPath,
		MountDirectoryFromHost:   pluginPipelineScript.MountDirectoryFromHost,
		ContainerImagePath:       pluginPipelineScript.ContainerImagePath,
		ImagePullSecretType:      pluginPipelineScript.ImagePullSecretType,
		ImagePullSecret:          pluginPipelineScript.ImagePullSecret,
		AuditLog:                 sql.NewDefaultAuditLog(userId),
	}

}

func GetPluginStepVariableDbObject(pluginStepId int, pluginVariableDto *pluginBean.PluginVariableDto, userId int32) *repository.PluginStepVariable {
	model := &repository.PluginStepVariable{
		PluginStepId:              pluginStepId,
		Name:                      pluginVariableDto.Name,
		Format:                    pluginVariableDto.Format,
		Description:               pluginVariableDto.Description,
		IsExposed:                 true, //currently hard coding this, later after plugin creation gets more mature will let user decide
		AllowEmptyValue:           pluginVariableDto.AllowEmptyValue,
		DefaultValue:              pluginVariableDto.DefaultValue,
		Value:                     pluginVariableDto.Value,
		VariableType:              pluginVariableDto.VariableType,
		ValueType:                 pluginVariableDto.ValueType,
		PreviousStepIndex:         pluginVariableDto.PreviousStepIndex,
		VariableStepIndex:         1, //currently hard coding this, later after plugin creation gets more mature will let user decide
		VariableStepIndexInPlugin: pluginVariableDto.VariableStepIndexInPlugin,
		ReferenceVariableName:     pluginVariableDto.ReferenceVariableName,
		AuditLog:                  sql.NewDefaultAuditLog(userId),
	}
	return model
}

func GetPluginStepConditionDbObject(stepDataId, pluginStepVariableId int, pluginStepCondition *pluginBean.PluginStepCondition,
	userId int32) *repository.PluginStepCondition {
	return &repository.PluginStepCondition{
		PluginStepId:        stepDataId,
		ConditionVariableId: pluginStepVariableId,
		ConditionType:       pluginStepCondition.ConditionType,
		ConditionalOperator: pluginStepCondition.ConditionalOperator,
		ConditionalValue:    pluginStepCondition.ConditionalValue,
		AuditLog:            sql.NewDefaultAuditLog(userId),
	}
}
