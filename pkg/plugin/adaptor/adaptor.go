package adaptor

import (
	bean2 "github.com/devtron-labs/devtron/pkg/plugin/bean"
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
)

func GetPluginParentMetadataDbObject(pluginDto *bean2.PluginParentMetadataDto, userId int32) *repository.PluginParentMetadata {
	return repository.NewPluginParentMetadata().CreateAuditLog(userId).
		WithBasicMetadata(pluginDto.Name, pluginDto.PluginIdentifier, pluginDto.Description, pluginDto.Icon, repository.PLUGIN_TYPE_SHARED)
}

func GetPluginVersionMetadataDbObject(pluginDto *bean2.PluginParentMetadataDto, userId int32) *repository.PluginMetadata {
	versionDto := pluginDto.Versions.DetailedPluginVersionData[0]
	return repository.NewPluginVersionMetadata().CreateAuditLog(userId).WithBasicMetadata(pluginDto.Name, versionDto.Description, versionDto.Version, versionDto.DocLink)
}

func GetPluginStepDbObject(pluginStepDto *bean2.PluginStepsDto, pluginVersionMetadata int, userId int32) *repository.PluginStep {
	return &repository.PluginStep{
		PluginId:            pluginVersionMetadata,
		Name:                pluginStepDto.Name,
		Description:         pluginStepDto.Description,
		Index:               pluginStepDto.Index,
		StepType:            pluginStepDto.StepType,
		RefPluginId:         pluginStepDto.RefPluginId,
		OutputDirectoryPath: pluginStepDto.OutputDirectoryPath,
		DependentOnStep:     pluginStepDto.DependentOnStep,
		AuditLog:            sql.NewDefaultAuditLog(userId),
	}
}
func GetPluginPipelineScriptDbObject(pluginPipelineScript *bean2.PluginPipelineScript, userId int32) *repository.PluginPipelineScript {
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

func GetPluginStepVariableDbObject(pluginStepId int, pluginVariableDto *bean2.PluginVariableDto, userId int32) *repository.PluginStepVariable {
	return &repository.PluginStepVariable{
		PluginStepId:              pluginStepId,
		Name:                      pluginVariableDto.Name,
		Format:                    pluginVariableDto.Format,
		Description:               pluginVariableDto.Description,
		IsExposed:                 pluginVariableDto.IsExposed,
		AllowEmptyValue:           pluginVariableDto.AllowEmptyValue,
		DefaultValue:              pluginVariableDto.DefaultValue,
		Value:                     pluginVariableDto.Value,
		VariableType:              pluginVariableDto.VariableType,
		ValueType:                 pluginVariableDto.ValueType,
		PreviousStepIndex:         pluginVariableDto.PreviousStepIndex,
		VariableStepIndex:         pluginVariableDto.VariableStepIndex,
		VariableStepIndexInPlugin: pluginVariableDto.VariableStepIndexInPlugin,
		ReferenceVariableName:     pluginVariableDto.ReferenceVariableName,
		AuditLog:                  sql.NewDefaultAuditLog(userId),
	}
}

func GetPluginStepConditionDbObject(stepDataId, pluginStepVariableId int, pluginStepCondition *bean2.PluginStepCondition,
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
