package plugin

import (
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strings"
	"time"
)

type GlobalVariable struct {
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	Format      string `json:"format"`
	Description string `json:"description"`
	Type        string `json:"stageType"`
}

type GlobalPluginService interface {
	GetAllGlobalVariables() ([]*GlobalVariable, error)
	ListAllPlugins(stageType int) ([]*PluginListComponentDto, error)
	GetPluginDetailById(pluginId int) (*PluginDetailDto, error)
	PatchPlugin(pluginDto *PluginMetadataDto, userId int32) (*PluginMetadataDto, error)
	GetDetailedPluginInfoByPluginId(pluginId int) (*PluginMetadataDto, error)
	GetAllDetailedPluginInfo() ([]*PluginMetadataDto, error)
}

func NewGlobalPluginService(logger *zap.SugaredLogger, globalPluginRepository repository.GlobalPluginRepository) *GlobalPluginServiceImpl {
	return &GlobalPluginServiceImpl{
		logger:                 logger,
		globalPluginRepository: globalPluginRepository,
	}
}

type GlobalPluginServiceImpl struct {
	logger                 *zap.SugaredLogger
	globalPluginRepository repository.GlobalPluginRepository
}

func (impl *GlobalPluginServiceImpl) GetAllGlobalVariables() ([]*GlobalVariable, error) {
	globalVariables := []*GlobalVariable{
		{
			Name:        "WORKING_DIRECTORY",
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Directory in which git material is cloned.The home path of repo = WORKING_DIRECTORY + CHECKOUT_PATH",
			Type:        "ci",
		},
		{
			Name:        "DOCKER_IMAGE_TAG",
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Tag going to be used to push image.",
			Type:        "ci",
		},
		{
			Name:        "DOCKER_REPOSITORY",
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Name of the repository to be used for pushing images.",
			Type:        "ci",
		},
		{
			Name:        "DOCKER_REGISTRY_URL",
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Url of the container registry used for this pipeline.",
			Type:        "ci",
		},
		{
			Name:        "DOCKER_IMAGE",
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Complete image name(repository+registry+tag).",
			Type:        "ci",
		},
		{
			Name:        "APP_NAME",
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Name of the app this pipeline resides in.",
			Type:        "ci",
		},
		{
			Name:        "TRIGGER_BY_AUTHOR",
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Email-Id/Name of the user who triggers pipeline.",
			Type:        "ci",
		},
		{
			Name:        pipeline.CD_PIPELINE_ENV_NAME_KEY,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "The name of the environment for which this deployment pipeline is configured.",
			Type:        "cd",
		},
		{
			Name:        pipeline.CD_PIPELINE_CLUSTER_NAME_KEY,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "The name of the cluster to which the environment belongs for which this deployment pipeline is configured.",
			Type:        "cd",
		},
		{
			Name:        pipeline.DOCKER_IMAGE,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Complete image name(repository+registry+tag).",
			Type:        "cd",
		},
		{
			Name:        pipeline.APP_NAME,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "The name of the app this pipeline resides in.",
			Type:        "cd",
		},
		{
			Name:        pipeline.DEPLOYMENT_RELEASE_ID,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Auto-incremented counter for deployment triggers.",
			Type:        "post-cd",
		},
		{
			Name:        pipeline.DEPLOYMENT_UNIQUE_ID,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Auto-incremented counter for deployment triggers. Counter is shared between Pre/Post/Deployment stages.",
			Type:        "cd",
		},
		{
			Name:        pipeline.CD_TRIGGERED_BY,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Email-Id/Name of the user who triggered the deployment pipeline.",
			Type:        "post-cd",
		},
		{
			Name:        pipeline.CD_TRIGGER_TIME,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Time when the deployment pipeline was triggered.",
			Type:        "post-cd",
		},
		{
			Name:        pipeline.GIT_METADATA,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "GIT_METADATA consists of GIT_COMMIT_HASH, GIT_SOURCE_TYPE, GIT_SOURCE_VALUE.",
			Type:        "cd",
		},
		{
			Name:        pipeline.APP_LABEL_METADATA,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "APP_LABEL_METADATA consists of APP_LABEL_KEY, APP_LABEL_VALUE. APP_LABEL_METADATA will only be available if workflow has External CI.",
			Type:        "cd",
		},
		{
			Name:        pipeline.CHILD_CD_METADATA,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "CHILD_CD_METADATA consists of CHILD_CD_ENV_NAME, CHILD_CD_CLUSTER_NAME. CHILD_CD_METADATA will only be available if this CD pipeline has a Child CD pipeline.",
			Type:        "cd",
		},
	}
	return globalVariables, nil
}

func (impl *GlobalPluginServiceImpl) ListAllPlugins(stageType int) ([]*PluginListComponentDto, error) {
	impl.logger.Infow("request received, ListAllPlugins")
	var pluginDetails []*PluginListComponentDto
	//getting all plugins metadata(without tags)
	pluginsMetadata, err := impl.globalPluginRepository.GetMetaDataForAllPlugins(stageType)
	if err != nil {
		impl.logger.Errorw("error in getting plugins", "err", err)
		return nil, err
	}
	pluginIdTagsMap, err := impl.getPluginIdTagsMap()
	if err != nil {
		impl.logger.Errorw("error, getPluginIdTagsMap", "err", err)
		return nil, err
	}
	pluginIdInputVariablesMap, pluginIdOutputVariablesMap, err := impl.getPluginIdVariablesMap()
	if err != nil {
		impl.logger.Errorw("error, getPluginIdVariablesMap", "err", err)
		return nil, err
	}
	for _, pluginMetadata := range pluginsMetadata {
		pluginMetadataDto := &PluginMetadataDto{
			Id:          pluginMetadata.Id,
			Name:        pluginMetadata.Name,
			Type:        string(pluginMetadata.Type),
			Description: pluginMetadata.Description,
			Icon:        pluginMetadata.Icon,
		}
		tags, ok := pluginIdTagsMap[pluginMetadata.Id]
		if ok {
			pluginMetadataDto.Tags = tags
		}
		pluginDetail := &PluginListComponentDto{
			PluginMetadataDto: pluginMetadataDto,
			InputVariables:    pluginIdInputVariablesMap[pluginMetadata.Id],
			OutputVariables:   pluginIdOutputVariablesMap[pluginMetadata.Id],
		}
		pluginDetails = append(pluginDetails, pluginDetail)
	}
	return pluginDetails, nil
}

func (impl *GlobalPluginServiceImpl) GetPluginDetailById(pluginId int) (*PluginDetailDto, error) {
	impl.logger.Infow("request received, GetPluginDetail", "pluginId", pluginId)

	//getting metadata
	pluginMetadata, err := impl.globalPluginRepository.GetMetaDataByPluginId(pluginId)
	if err != nil {
		impl.logger.Errorw("error in getting plugins", "err", err, "pluginId", pluginId)
		return nil, err
	}
	metadataDto := &PluginMetadataDto{
		Id:          pluginMetadata.Id,
		Name:        pluginMetadata.Name,
		Type:        string(pluginMetadata.Type),
		Description: pluginMetadata.Description,
		Icon:        pluginMetadata.Icon,
	}
	pluginDetail := &PluginDetailDto{
		Metadata: metadataDto,
	}
	pluginDetail.InputVariables, pluginDetail.OutputVariables, err = impl.getIOVariablesOfAPlugin(pluginMetadata.Id)
	if err != nil {
		impl.logger.Errorw("error, getIOVariablesOfAPlugin", "err", err)
		return nil, err
	}
	return pluginDetail, nil
}

func (impl *GlobalPluginServiceImpl) getPluginIdTagsMap() (map[int][]string, error) {
	//getting all plugin tags
	pluginTags, err := impl.globalPluginRepository.GetAllPluginTags()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting all plugin tags", "err", err)
		return nil, err
	}
	tagIdNameMap := make(map[int]string)
	for _, tag := range pluginTags {
		tagIdNameMap[tag.Id] = tag.Name
	}
	//getting plugin-tag relations
	relations, err := impl.globalPluginRepository.GetAllPluginTagRelations()
	if err != nil {
		impl.logger.Errorw("error in getting all plugin-tag relations", "err", err)
		return nil, err
	}
	pluginIdTagsMap := make(map[int][]string)
	for _, relation := range relations {
		tag, ok := tagIdNameMap[relation.TagId]
		if ok {
			pluginIdTagsMap[relation.PluginId] = append(pluginIdTagsMap[relation.PluginId], tag)
		}
	}
	return pluginIdTagsMap, nil
}

func (impl *GlobalPluginServiceImpl) getPluginIdVariablesMap() (map[int][]*PluginVariableDto, map[int][]*PluginVariableDto, error) {
	variables, err := impl.globalPluginRepository.GetExposedVariablesForAllPlugins()
	if err != nil {
		impl.logger.Errorw("error in getting exposed vars for all plugins", "err", err)
		return nil, nil, err
	}
	pluginIdInputVarsMap, pluginIdOutputVarsMap := make(map[int][]*PluginVariableDto), make(map[int][]*PluginVariableDto)
	for _, variable := range variables {
		variableDto := getVariableDto(variable)
		if variable.VariableType == repository.PLUGIN_VARIABLE_TYPE_INPUT {
			pluginIdInputVarsMap[variable.PluginMetadataId] = append(pluginIdInputVarsMap[variable.PluginMetadataId], variableDto)
		} else if variable.VariableType == repository.PLUGIN_VARIABLE_TYPE_OUTPUT {
			pluginIdOutputVarsMap[variable.PluginMetadataId] = append(pluginIdOutputVarsMap[variable.PluginMetadataId], variableDto)
		}
	}
	return pluginIdInputVarsMap, pluginIdOutputVarsMap, nil
}

func (impl *GlobalPluginServiceImpl) getIOVariablesOfAPlugin(pluginId int) (inputVariablesDto, outputVariablesDto []*PluginVariableDto, err error) {
	//getting exposed variables
	pluginVariables, err := impl.globalPluginRepository.GetExposedVariablesByPluginId(pluginId)
	if err != nil {
		impl.logger.Errorw("error in getting pluginVariables by pluginId", "err", err, "pluginId", pluginId)
		return nil, nil, err
	}
	for _, pluginVariable := range pluginVariables {
		variableDto := getVariableDto(pluginVariable)
		if pluginVariable.VariableType == repository.PLUGIN_VARIABLE_TYPE_INPUT {
			inputVariablesDto = append(inputVariablesDto, variableDto)
		} else if pluginVariable.VariableType == repository.PLUGIN_VARIABLE_TYPE_OUTPUT {
			outputVariablesDto = append(outputVariablesDto, variableDto)
		}
	}
	return inputVariablesDto, outputVariablesDto, nil
}

func getVariableDto(pluginVariable *repository.PluginStepVariable) *PluginVariableDto {
	return &PluginVariableDto{
		Id:                    pluginVariable.Id,
		Name:                  pluginVariable.Name,
		Format:                pluginVariable.Format,
		Description:           pluginVariable.Description,
		IsExposed:             pluginVariable.IsExposed,
		AllowEmptyValue:       pluginVariable.AllowEmptyValue,
		DefaultValue:          pluginVariable.DefaultValue,
		Value:                 pluginVariable.Value,
		ValueType:             pluginVariable.ValueType,
		PreviousStepIndex:     pluginVariable.PreviousStepIndex,
		VariableStepIndex:     pluginVariable.VariableStepIndex,
		ReferenceVariableName: pluginVariable.ReferenceVariableName,
	}
}

func (impl *GlobalPluginServiceImpl) PatchPlugin(pluginDto *PluginMetadataDto, userId int32) (*PluginMetadataDto, error) {

	switch pluginDto.Action {
	case 0:
		//create action
		pluginData, err := impl.createPlugin(pluginDto, userId)
		if err != nil {
			impl.logger.Errorw("error in creating plugin", "err", err, "pluginDto", pluginDto)
			return nil, err
		}
		return pluginData, nil
	case 1:
		pluginData, err := impl.updatePlugin(pluginDto, userId)
		if err != nil {
			impl.logger.Errorw("error in updating plugin", "err", err, "pluginDto", pluginDto)
			return nil, err
		}
		return pluginData, nil
	}

	return nil, nil
}

func (impl *GlobalPluginServiceImpl) createPlugin(pluginReq *PluginMetadataDto, userId int32) (*PluginMetadataDto, error) {
	dbConnection := impl.globalPluginRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	//create entry in plugin_metadata
	pluginMetadata := &repository.PluginMetadata{
		Name:        pluginReq.Name,
		Description: pluginReq.Description,
		Type:        repository.PluginType(pluginReq.Type),
		Icon:        pluginReq.Icon,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
	pluginMetadata, err = impl.globalPluginRepository.SavePluginMetadata(pluginMetadata, tx)
	if err != nil {
		impl.logger.Errorw("error in saving plugin", "pluginDto", pluginReq, "err", err)
		return nil, err
	}
	pluginReq.Id = pluginMetadata.Id
	pluginStage := 2
	if pluginReq.PluginStage == "CI" {
		pluginStage = 0
	} else if pluginReq.PluginStage == "CD" {
		pluginStage = 1
	}
	pluginStageMapping := &repository.PluginStageMapping{
		PluginId:  pluginMetadata.Id,
		StageType: pluginStage,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
	_, err = impl.globalPluginRepository.SavePluginStageMapping(pluginStageMapping, tx)
	if err != nil {
		impl.logger.Errorw("error in saving plugin stage mapping", "pluginDto", pluginReq, "err", err)
		return nil, err
	}

	for _, pluginStep := range pluginReq.PluginSteps {
		//create entry in plugin_pipeline_script
		pluginPipelineScript := &repository.PluginPipelineScript{
			Script:                   pluginStep.PluginPipelineScript.Script,
			StoreScriptAt:            pluginStep.PluginPipelineScript.StoreScriptAt,
			Type:                     pluginStep.PluginPipelineScript.Type,
			DockerfileExists:         pluginStep.PluginPipelineScript.DockerfileExists,
			MountPath:                pluginStep.PluginPipelineScript.MountPath,
			MountCodeToContainer:     pluginStep.PluginPipelineScript.MountCodeToContainer,
			MountCodeToContainerPath: pluginStep.PluginPipelineScript.MountCodeToContainerPath,
			MountDirectoryFromHost:   pluginStep.PluginPipelineScript.MountDirectoryFromHost,
			ContainerImagePath:       pluginStep.PluginPipelineScript.ContainerImagePath,
			ImagePullSecretType:      pluginStep.PluginPipelineScript.ImagePullSecretType,
			ImagePullSecret:          pluginStep.PluginPipelineScript.ImagePullSecret,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: userId,
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		//create entry in plugin_step
		pluginPipelineScript, err = impl.globalPluginRepository.SavePluginPipelineScript(pluginPipelineScript, tx)
		if err != nil {
			impl.logger.Errorw("error in saving plugin pipeline script", "pluginPipelineScript", pluginPipelineScript, "err", err)
			return nil, err
		}
		pluginStep.PluginPipelineScript.Id = pluginPipelineScript.Id

		pluginStepData := &repository.PluginStep{
			PluginId:            pluginMetadata.Id,
			Name:                pluginStep.Name,
			Description:         pluginStep.Description,
			Index:               pluginStep.Index,
			StepType:            pluginStep.StepType,
			ScriptId:            pluginPipelineScript.Id,
			RefPluginId:         pluginStep.RefPluginId,
			OutputDirectoryPath: pluginStep.OutputDirectoryPath,
			DependentOnStep:     pluginStep.DependentOnStep,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: userId,
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		pluginStepData, err = impl.globalPluginRepository.SavePluginSteps(pluginStepData, tx)
		if err != nil {
			impl.logger.Errorw("error in saving plugin step", "pluginStepData", pluginStepData, "err", err)
			return nil, err
		}
		pluginStep.Id = pluginStepData.Id
		//create entry in plugin_step_variable
		for _, pluginStepVariable := range pluginStep.PluginStepVariable {
			pluginStepVariableData := &repository.PluginStepVariable{
				PluginStepId:              pluginStepData.Id,
				Name:                      pluginStepVariable.Name,
				Format:                    pluginStepVariable.Format,
				Description:               pluginStepVariable.Description,
				IsExposed:                 pluginStepVariable.IsExposed,
				AllowEmptyValue:           pluginStepVariable.AllowEmptyValue,
				DefaultValue:              pluginStepVariable.DefaultValue,
				Value:                     pluginStepVariable.Value,
				VariableType:              repository.PluginStepVariableType(pluginStepVariable.VariableType),
				ValueType:                 pluginStepVariable.ValueType,
				PreviousStepIndex:         pluginStepVariable.PreviousStepIndex,
				VariableStepIndex:         pluginStepVariable.VariableStepIndex,
				VariableStepIndexInPlugin: pluginStepVariable.VariableStepIndexInPlugin,
				ReferenceVariableName:     pluginStepVariable.ReferenceVariableName,
				AuditLog: sql.AuditLog{
					CreatedOn: time.Now(),
					CreatedBy: userId,
					UpdatedOn: time.Now(),
					UpdatedBy: userId,
				},
			}
			pluginStepVariableData, err = impl.globalPluginRepository.SavePluginStepVariables(pluginStepVariableData, tx)
			if err != nil {
				impl.logger.Errorw("error in saving plugin step variable", "pluginStepVariableData", pluginStepVariableData, "err", err)
				return nil, err
			}
			pluginStepVariable.Id = pluginStepVariableData.Id
			//create entry in plugin_step_condition
			for _, pluginStepCondition := range pluginStepVariable.PluginStepCondition {
				pluginStepConditionData := &repository.PluginStepCondition{
					PluginStepId:        pluginStepData.Id,
					ConditionVariableId: pluginStepVariableData.Id,
					ConditionType:       pluginStepCondition.ConditionType,
					ConditionalOperator: pluginStepCondition.ConditionalOperator,
					ConditionalValue:    pluginStepCondition.ConditionalValue,
					AuditLog: sql.AuditLog{
						CreatedOn: time.Now(),
						CreatedBy: userId,
						UpdatedOn: time.Now(),
						UpdatedBy: userId,
					},
				}
				pluginStepConditionData, err = impl.globalPluginRepository.SavePluginStepConditions(pluginStepConditionData, tx)
				if err != nil {
					impl.logger.Errorw("error in saving plugin step condition", "pluginStepConditionData", pluginStepConditionData, "err", err)
					return nil, err
				}
				pluginStepCondition.Id = pluginStepConditionData.Id
			}
		}
	}

	allPluginTags, err := impl.globalPluginRepository.GetAllPluginTags()
	if err != nil {
		impl.logger.Errorw("error in getting all plugin tags", "err", err)
		return nil, err
	}
	//check for new tags, then create new plugin_tag and plugin_tag_relation entry in db when new tags are present in request
	for _, pluginTagReq := range pluginReq.Tags {
		tagAlreadyExists := false
		for _, presentPluginTags := range allPluginTags {
			if strings.ToLower(pluginTagReq) == strings.ToLower(presentPluginTags.Name) {
				tagAlreadyExists = true
			}
		}
		if !tagAlreadyExists {
			newPluginTag := &repository.PluginTag{
				Name: pluginTagReq,
				AuditLog: sql.AuditLog{
					CreatedOn: time.Now(),
					CreatedBy: userId,
					UpdatedOn: time.Now(),
					UpdatedBy: userId,
				},
			}
			newPluginTag, err = impl.globalPluginRepository.SavePluginTag(newPluginTag, tx)
			if err != nil {
				impl.logger.Errorw("error in saving plugin tag", "newPluginTag", newPluginTag, "err", err)
				return nil, err
			}
			newPluginTagRelation := &repository.PluginTagRelation{
				TagId:    newPluginTag.Id,
				PluginId: pluginReq.Id,
				AuditLog: sql.AuditLog{
					CreatedOn: time.Now(),
					CreatedBy: userId,
					UpdatedOn: time.Now(),
					UpdatedBy: userId,
				},
			}
			newPluginTagRelation, err = impl.globalPluginRepository.SavePluginTagRelation(newPluginTagRelation, tx)
			if err != nil {
				impl.logger.Errorw("error in saving plugin tag relation", "newPluginTagRelation", newPluginTagRelation, "err", err)
				return nil, err
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("createPlugin, error in committing db transaction", "err", err)
		return nil, err
	}
	return pluginReq, nil
}

func (impl *GlobalPluginServiceImpl) updatePlugin(pluginUpdateReq *PluginMetadataDto, userId int32) (*PluginMetadataDto, error) {
	dbConnection := impl.globalPluginRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	//pluginMetaData, err := impl.globalPluginRepository.GetMetaDataByPluginId(pluginUpdateReq.Id)
	//if err != nil {
	//	impl.logger.Errorw("updatePlugin, error in getting pluginMetadata", "pluginId", pluginUpdateReq.Id, "err", err)
	//	return nil, err
	//}
	//pluginStageMapping, err := impl.globalPluginRepository.GetPluginStageMappingByPluginId(pluginUpdateReq.Id)
	//if err != nil {
	//	impl.logger.Errorw("updatePlugin, error in getting pluginStageMapping", "pluginId", pluginUpdateReq.Id, "err", err)
	//	return nil, err
	//}
	//pluginSteps, err := impl.globalPluginRepository.GetPluginStepsByPluginId(pluginUpdateReq.Id)
	//if err != nil {
	//	impl.logger.Errorw("updatePlugin, error in getting pluginSteps", "pluginId", pluginUpdateReq.Id, "err", err)
	//	return nil, err
	//}
	//pluginStepVariables, err := impl.globalPluginRepository.GetExposedVariablesByPluginId(pluginUpdateReq.Id)
	//if err != nil {
	//	impl.logger.Errorw("updatePlugin, error in getting pluginStepVariables", "pluginId", pluginUpdateReq.Id, "err", err)
	//	return nil, err
	//}
	//pluginStepConditions, err := impl.globalPluginRepository.GetConditionsByPluginId(pluginUpdateReq.Id)
	//if err != nil {
	//	impl.logger.Errorw("updatePlugin, error in getting pluginStepConditions", "pluginId", pluginUpdateReq.Id, "err", err)
	//	return nil, err
	//}

	return pluginUpdateReq, nil
}

func (impl *GlobalPluginServiceImpl) GetAllDetailedPluginInfo() ([]*PluginMetadataDto, error) {
	allPlugins, err := impl.globalPluginRepository.GetAllPluginMetaData()
	if err != nil {
		impl.logger.Errorw("GetAllDetailedPluginInfo, error in getting all pluginsMetadata", "err", err)
		return nil, err
	}
	allPluginMetadata := make([]*PluginMetadataDto, 0)
	for _, plugin := range allPlugins {
		pluginDetailedInfo, err := impl.GetDetailedPluginInfoByPluginId(plugin.Id)
		if err != nil {
			impl.logger.Errorw("GetAllDetailedPluginInfo, error in getting pluginDetailedInfo", "pluginId", plugin.Id, "err", err)
			return nil, err
		}
		allPluginMetadata = append(allPluginMetadata, pluginDetailedInfo)
	}
	return allPluginMetadata, nil
}

func (impl *GlobalPluginServiceImpl) GetDetailedPluginInfoByPluginId(pluginId int) (*PluginMetadataDto, error) {

	pluginMetaData, err := impl.globalPluginRepository.GetMetaDataByPluginId(pluginId)
	if err != nil {
		impl.logger.Errorw("GetDetailedPluginInfoByPluginId, error in getting pluginMetadata", "pluginId", pluginId, "err", err)
		return nil, err
	}
	pluginStageMapping, err := impl.globalPluginRepository.GetPluginStageMappingByPluginId(pluginId)
	if err != nil {
		impl.logger.Errorw("GetDetailedPluginInfoByPluginId, error in getting pluginStageMapping", "pluginId", pluginId, "err", err)
		return nil, err
	}
	pluginSteps, err := impl.globalPluginRepository.GetPluginStepsByPluginId(pluginId)
	if err != nil {
		impl.logger.Errorw("GetDetailedPluginInfoByPluginId, error in getting pluginSteps", "pluginId", pluginId, "err", err)
		return nil, err
	}
	pluginStepVariables, err := impl.globalPluginRepository.GetExposedVariablesByPluginId(pluginId)
	if err != nil {
		impl.logger.Errorw("GetDetailedPluginInfoByPluginId, error in getting pluginStepVariables", "pluginId", pluginId, "err", err)
		return nil, err
	}
	pluginStepConditions, err := impl.globalPluginRepository.GetConditionsByPluginId(pluginId)
	if err != nil {
		impl.logger.Errorw("GetDetailedPluginInfoByPluginId, error in getting pluginStepConditions", "pluginId", pluginId, "err", err)
		return nil, err
	}
	pluginStage := "CI_CD"
	if pluginStageMapping.StageType == 0 {
		pluginStage = "CI"
	} else if pluginStageMapping.StageType == 1 {
		pluginStage = "CD"
	}
	pluginIdTagsMap, err := impl.getPluginIdTagsMap()
	if err != nil {
		impl.logger.Errorw("GetDetailedPluginInfoByPluginId, error in getting pluginIdTagsMap", "pluginId", pluginId, "err", err)
		return nil, err
	}
	pluginMetadataResponse := &PluginMetadataDto{
		Id:          pluginMetaData.Id,
		Name:        pluginMetaData.Name,
		Description: pluginMetaData.Description,
		Type:        string(pluginMetaData.Type),
		Icon:        pluginMetaData.Icon,
		Tags:        pluginIdTagsMap[pluginMetaData.Id],
		PluginStage: pluginStage,
	}

	pluginStepsResp := make([]*PluginStepsDto, 0)
	for _, pluginStep := range pluginSteps {
		pluginScript, err := impl.globalPluginRepository.GetScriptDetailById(pluginStep.ScriptId)
		if err != nil {
			impl.logger.Errorw("GetDetailedPluginInfoByPluginId, error in getting pluginScript", "pluginScriptId", pluginStep.ScriptId, "pluginId", pluginId, "err", err)
			return nil, err
		}
		pluginStepDto := &PluginStepsDto{
			Id:                   pluginStep.Id,
			Name:                 pluginStep.Name,
			Description:          pluginStep.Description,
			Index:                pluginStep.Index,
			StepType:             pluginStep.StepType,
			RefPluginId:          pluginStep.RefPluginId,
			OutputDirectoryPath:  pluginStep.OutputDirectoryPath,
			DependentOnStep:      pluginStep.DependentOnStep,
			PluginPipelineScript: pluginScript,
		}
		pluginStepVariableResp := make([]*PluginVariableDto, 0)
		for _, pluginStepVariable := range pluginStepVariables {
			if pluginStepVariable.PluginStepId == pluginStep.Id {
				pluginStepConditionDto := make([]*repository.PluginStepCondition, 0)
				for _, pluginStepCondition := range pluginStepConditions {
					if pluginStepCondition.ConditionVariableId == pluginStepVariable.Id {
						pluginStepConditionDto = append(pluginStepConditionDto, &repository.PluginStepCondition{
							Id:                  pluginStepCondition.Id,
							PluginStepId:        pluginStepCondition.PluginStepId,
							ConditionVariableId: pluginStepCondition.ConditionVariableId,
							ConditionType:       pluginStepCondition.ConditionType,
							ConditionalOperator: pluginStepCondition.ConditionalOperator,
							ConditionalValue:    pluginStepCondition.ConditionalValue,
							Deleted:             pluginStepCondition.Deleted,
						})
					}
				}
				pluginStepVariableResp = append(pluginStepVariableResp, &PluginVariableDto{
					Id:                        pluginStepVariable.Id,
					Name:                      pluginStepVariable.Name,
					Format:                    pluginStepVariable.Format,
					Description:               pluginStepVariable.Description,
					IsExposed:                 pluginStepVariable.IsExposed,
					AllowEmptyValue:           pluginStepVariable.AllowEmptyValue,
					DefaultValue:              pluginStepVariable.DefaultValue,
					Value:                     pluginStepVariable.Value,
					VariableType:              string(pluginStepVariable.VariableType),
					ValueType:                 pluginStepVariable.ValueType,
					PreviousStepIndex:         pluginStepVariable.PreviousStepIndex,
					VariableStepIndex:         pluginStepVariable.VariableStepIndex,
					VariableStepIndexInPlugin: pluginStepVariable.VariableStepIndexInPlugin,
					ReferenceVariableName:     pluginStepVariable.ReferenceVariableName,
					PluginStepCondition:       pluginStepConditionDto,
				})
			}
		}
		pluginStepDto.PluginStepVariable = pluginStepVariableResp
		pluginStepsResp = append(pluginStepsResp, pluginStepDto)
	}
	pluginMetadataResponse.PluginSteps = pluginStepsResp

	return pluginMetadataResponse, nil
}
