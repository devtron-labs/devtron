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

// Below adaptors contains reverse adaptors of db object adaptors

// GetPluginStepDtoFromDbObject returns PluginStepsDto object without PluginVariableDto and PluginPipelineScript objects
func GetPluginStepDtoFromDbObject(pluginStepsDbObj *repository.PluginStep) *pluginBean.PluginStepsDto {
	if pluginStepsDbObj == nil {
		return &pluginBean.PluginStepsDto{}
	}
	return &pluginBean.PluginStepsDto{
		Id:                  pluginStepsDbObj.Id,
		Name:                pluginStepsDbObj.Name,
		Description:         pluginStepsDbObj.Description,
		Index:               pluginStepsDbObj.Index,
		StepType:            pluginStepsDbObj.StepType,
		RefPluginId:         pluginStepsDbObj.RefPluginId,
		OutputDirectoryPath: pluginStepsDbObj.OutputDirectoryPath,
		DependentOnStep:     pluginStepsDbObj.DependentOnStep,
		ScriptId:            pluginStepsDbObj.ScriptId,
	}
}

// GetPluginStepsDtoFromDbObjects takes pluginStepsDbObj array and returns transformed array of pluginSteps dto object
func GetPluginStepsDtoFromDbObjects(pluginStepsDbObj []*repository.PluginStep) []*pluginBean.PluginStepsDto {
	pluginStepsDto := make([]*pluginBean.PluginStepsDto, 0, len(pluginStepsDbObj))
	for _, pluginStep := range pluginStepsDbObj {
		pluginStepsDto = append(pluginStepsDto, GetPluginStepDtoFromDbObject(pluginStep))
	}
	return pluginStepsDto
}

// GetPluginStepVarsDtoFromDbObject returns PluginVariableDto object wihtout pluginStepConditions
func GetPluginStepVarsDtoFromDbObject(pluginStepVarDbObj *repository.PluginStepVariable) *pluginBean.PluginVariableDto {
	if pluginStepVarDbObj == nil {
		return &pluginBean.PluginVariableDto{}
	}
	return &pluginBean.PluginVariableDto{
		Id:                        pluginStepVarDbObj.Id,
		Name:                      pluginStepVarDbObj.Name,
		Format:                    pluginStepVarDbObj.Format,
		Description:               pluginStepVarDbObj.Description,
		IsExposed:                 pluginStepVarDbObj.IsExposed,
		AllowEmptyValue:           pluginStepVarDbObj.AllowEmptyValue,
		DefaultValue:              pluginStepVarDbObj.DefaultValue,
		Value:                     pluginStepVarDbObj.Value,
		VariableType:              pluginStepVarDbObj.VariableType,
		ValueType:                 pluginStepVarDbObj.ValueType,
		PreviousStepIndex:         pluginStepVarDbObj.PreviousStepIndex,
		VariableStepIndex:         pluginStepVarDbObj.VariableStepIndex,
		VariableStepIndexInPlugin: pluginStepVarDbObj.VariableStepIndexInPlugin,
		ReferenceVariableName:     pluginStepVarDbObj.ReferenceVariableName,
		PluginStepId:              pluginStepVarDbObj.PluginStepId,
	}
}

// GetPluginStepVariablesDtoFromDbObjects takes pluginStepVarsDbObj array and returns transformed array of pluginVariable dto object
func GetPluginStepVariablesDtoFromDbObjects(pluginStepVarsDbObj []*repository.PluginStepVariable) []*pluginBean.PluginVariableDto {
	pluginStepVarsDto := make([]*pluginBean.PluginVariableDto, 0, len(pluginStepVarsDbObj))
	for _, pluginStepVar := range pluginStepVarsDbObj {
		pluginStepVarsDto = append(pluginStepVarsDto, GetPluginStepVarsDtoFromDbObject(pluginStepVar))
	}
	return pluginStepVarsDto
}

// GetPluginPipelineScriptDtoFromDbObject returns PluginPipelineScript dto object without ScriptPathArgPortMapping object
func GetPluginPipelineScriptDtoFromDbObject(pluginPipelineScript *repository.PluginPipelineScript) *pluginBean.PluginPipelineScript {
	if pluginPipelineScript == nil {
		return &pluginBean.PluginPipelineScript{}
	}
	return &pluginBean.PluginPipelineScript{
		Id:                       pluginPipelineScript.Id,
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
	}
}

// GetPluginPipelineScriptsDtoFromDbObjects takes pluginPipelineScriptDbObjs array and returns transformed array of pluginPipelineScripts dto object
func GetPluginPipelineScriptsDtoFromDbObjects(pluginPipelineScriptDbObjs []*repository.PluginPipelineScript) []*pluginBean.PluginPipelineScript {
	pluginPipelineScriptsDto := make([]*pluginBean.PluginPipelineScript, 0, len(pluginPipelineScriptDbObjs))
	for _, script := range pluginPipelineScriptDbObjs {
		pluginPipelineScriptsDto = append(pluginPipelineScriptsDto, GetPluginPipelineScriptDtoFromDbObject(script))
	}
	return pluginPipelineScriptsDto
}

func GetScripPathArgPortMappingDtoFromDbObject(scripPathArgPortMappingDbObj *repository.ScriptPathArgPortMapping) *pluginBean.ScriptPathArgPortMapping {
	if scripPathArgPortMappingDbObj == nil {
		return &pluginBean.ScriptPathArgPortMapping{}
	}
	return &pluginBean.ScriptPathArgPortMapping{
		Id:                  scripPathArgPortMappingDbObj.Id,
		TypeOfMapping:       scripPathArgPortMappingDbObj.TypeOfMapping,
		FilePathOnDisk:      scripPathArgPortMappingDbObj.FilePathOnDisk,
		FilePathOnContainer: scripPathArgPortMappingDbObj.FilePathOnContainer,
		Command:             scripPathArgPortMappingDbObj.Command,
		Args:                scripPathArgPortMappingDbObj.Args,
		PortOnLocal:         scripPathArgPortMappingDbObj.PortOnLocal,
		PortOnContainer:     scripPathArgPortMappingDbObj.PortOnContainer,
		ScriptId:            scripPathArgPortMappingDbObj.ScriptId,
	}
}

// GetScripPathArgPortMappingsDtoFromDbObjects takes scripPathArgPortMappingsDbObj array and returns transformed array of scriptPathArgPortMapping dto object
func GetScripPathArgPortMappingsDtoFromDbObjects(scripPathArgPortMappingsDbObj []*repository.ScriptPathArgPortMapping) []*pluginBean.ScriptPathArgPortMapping {
	scripPathArgPortMappingsDto := make([]*pluginBean.ScriptPathArgPortMapping, 0, len(scripPathArgPortMappingsDbObj))
	for _, mapping := range scripPathArgPortMappingsDbObj {
		scripPathArgPortMappingsDto = append(scripPathArgPortMappingsDto, GetScripPathArgPortMappingDtoFromDbObject(mapping))
	}
	return scripPathArgPortMappingsDto
}

func GetNewPluginStepDtoFromRefPluginMetadata(refPluginMetadata *repository.PluginMetadata) *pluginBean.PluginStepsDto {
	if refPluginMetadata == nil {
		return &pluginBean.PluginStepsDto{}
	}
	return &pluginBean.PluginStepsDto{
		Name:        refPluginMetadata.Name,
		Description: refPluginMetadata.Description,
		StepType:    repository.PLUGIN_STEP_TYPE_REF_PLUGIN,
		RefPluginId: refPluginMetadata.Id,
	}
}
