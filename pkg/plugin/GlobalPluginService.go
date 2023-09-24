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
	case 2:
		pluginData, err := impl.deletePlugin(pluginDto, userId)
		if err != nil {
			impl.logger.Errorw("error in deleting plugin", "err", err, "pluginDto", pluginDto)
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

	err = impl.saveDeepPluginStepData(pluginMetadata.Id, pluginReq.PluginSteps, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in saving plugin step data", "err", err)
		return nil, err
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

func (impl *GlobalPluginServiceImpl) saveDeepPluginStepData(pluginMetadataId int, pluginStepsReq []*PluginStepsDto, userId int32, tx *pg.Tx) error {
	for _, pluginStep := range pluginStepsReq {
		//get the script saved for this plugin step
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
		pluginPipelineScript, err := impl.globalPluginRepository.SavePluginPipelineScript(pluginPipelineScript, tx)
		if err != nil {
			impl.logger.Errorw("error in saving plugin pipeline script", "pluginPipelineScript", pluginPipelineScript, "err", err)
			return err
		}
		pluginStep.PluginPipelineScript.Id = pluginPipelineScript.Id

		pluginStepData := &repository.PluginStep{
			PluginId:            pluginMetadataId,
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
			return err
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
				return err
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
					return err
				}
				pluginStepCondition.Id = pluginStepConditionData.Id
			}
		}
	}
	return nil
}

func (impl *GlobalPluginServiceImpl) updatePlugin(pluginUpdateReq *PluginMetadataDto, userId int32) (*PluginMetadataDto, error) {
	dbConnection := impl.globalPluginRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	pluginMetaData, err := impl.globalPluginRepository.GetMetaDataByPluginId(pluginUpdateReq.Id)
	if err != nil {
		impl.logger.Errorw("updatePlugin, error in getting pluginMetadata, pluginId does not exist", "pluginId", pluginUpdateReq.Id, "err", err)
		return nil, err
	}
	pluginStageMapping, err := impl.globalPluginRepository.GetPluginStageMappingByPluginId(pluginUpdateReq.Id)
	if err != nil {
		impl.logger.Errorw("updatePlugin, error in getting pluginStageMapping", "pluginId", pluginUpdateReq.Id, "err", err)
		return nil, err
	}
	pluginSteps, err := impl.globalPluginRepository.GetPluginStepsByPluginId(pluginUpdateReq.Id)
	if err != nil {
		impl.logger.Errorw("updatePlugin, error in getting pluginSteps", "pluginId", pluginUpdateReq.Id, "err", err)
		return nil, err
	}
	pluginStepVariables, err := impl.globalPluginRepository.GetExposedVariablesByPluginId(pluginUpdateReq.Id)
	if err != nil {
		impl.logger.Errorw("updatePlugin, error in getting pluginStepVariables", "pluginId", pluginUpdateReq.Id, "err", err)
		return nil, err
	}
	pluginStepConditions, err := impl.globalPluginRepository.GetConditionsByPluginId(pluginUpdateReq.Id)
	if err != nil {
		impl.logger.Errorw("updatePlugin, error in getting pluginStepConditions", "pluginId", pluginUpdateReq.Id, "err", err)
		return nil, err
	}

	//update entry in plugin_metadata
	pluginMetaData.Name = pluginUpdateReq.Name
	pluginMetaData.Description = pluginUpdateReq.Description
	pluginMetaData.Type = repository.PluginType(pluginUpdateReq.Type)
	pluginMetaData.Icon = pluginUpdateReq.Icon
	pluginMetaData.UpdatedOn = time.Now()
	pluginMetaData.UpdatedBy = userId

	err = impl.globalPluginRepository.UpdatePluginMetadata(pluginMetaData, tx)
	if err != nil {
		impl.logger.Errorw("error in updating plugin metadata", "pluginId", pluginUpdateReq.Id, "err", err)
		return nil, err
	}
	pluginStage := 2
	if pluginUpdateReq.PluginStage == "CI" {
		pluginStage = 0
	} else if pluginUpdateReq.PluginStage == "CD" {
		pluginStage = 1
	}
	pluginStageMapping.StageType = pluginStage
	pluginStageMapping.UpdatedBy = userId
	pluginStageMapping.UpdatedOn = time.Now()

	err = impl.globalPluginRepository.UpdatePluginStageMapping(pluginStageMapping, tx)
	if err != nil {
		impl.logger.Errorw("error in updating plugin stage mapping", "pluginId", pluginUpdateReq.Id, "err", err)
		return nil, err
	}
	newPluginStepsToCreate, pluginStepsToRemove, pluginStepsToUpdate := filterPluginStepData(pluginSteps, pluginUpdateReq.PluginSteps)

	if len(newPluginStepsToCreate) > 0 {
		err = impl.saveDeepPluginStepData(pluginMetaData.Id, newPluginStepsToCreate, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in saveDeepPluginStepData", "pluginMetadataId", pluginMetaData.Id, "err", err)
			return nil, err
		}
	}
	if len(pluginStepsToRemove) > 0 {
		//basically we need to update here with deleted as true
		err = impl.deleteDeepPluginStepData(pluginStepsToRemove, pluginStepVariables, pluginStepConditions, pluginSteps, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in deleteDeepPluginStepData", "pluginMetadataId", pluginMetaData.Id, "err", err)
			return nil, err
		}
	}
	if len(pluginStepsToUpdate) > 0 {
		err = impl.updateDeepPluginStepData(pluginStepsToUpdate, pluginStepVariables, pluginStepConditions, pluginSteps, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in updateDeepPluginStepData", "pluginMetadataId", pluginMetaData.Id, "err", err)
			return nil, err
		}
	}

	allPluginTags, err := impl.globalPluginRepository.GetAllPluginTags()
	if err != nil {
		impl.logger.Errorw("error in getting all plugin tags", "err", err)
		return nil, err
	}
	//TODO update tags, if any new tags are added
	for _, pluginTagReq := range pluginUpdateReq.Tags {
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
				PluginId: pluginUpdateReq.Id,
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
		impl.logger.Errorw("updatePlugin, error in committing db transaction", "err", err)
		return nil, err
	}
	return pluginUpdateReq, nil
}

func (impl *GlobalPluginServiceImpl) updateDeepPluginStepData(pluginStepsToUpdate []*PluginStepsDto, pluginStepVariables []*repository.PluginStepVariable,
	pluginStepConditions []*repository.PluginStepCondition, pluginSteps []*repository.PluginStep, userId int32, tx *pg.Tx) error {

	pluginStepIdsToStepDtoMapping := make(map[int]*PluginStepsDto)
	for _, pluginStepUpdateReq := range pluginStepsToUpdate {
		pluginStepIdsToStepDtoMapping[pluginStepUpdateReq.Id] = pluginStepUpdateReq
	}
	for index, pluginStep := range pluginSteps {
		if _, ok := pluginStepIdsToStepDtoMapping[pluginStep.Id]; ok {
			pluginSteps[index].Name = pluginStepIdsToStepDtoMapping[pluginStep.Id].Name
			pluginSteps[index].Description = pluginStepIdsToStepDtoMapping[pluginStep.Id].Description
			pluginSteps[index].Index = pluginStepIdsToStepDtoMapping[pluginStep.Id].Index
			pluginSteps[index].StepType = pluginStepIdsToStepDtoMapping[pluginStep.Id].StepType
			pluginSteps[index].RefPluginId = pluginStepIdsToStepDtoMapping[pluginStep.Id].RefPluginId
			pluginSteps[index].OutputDirectoryPath = pluginStepIdsToStepDtoMapping[pluginStep.Id].OutputDirectoryPath
			pluginSteps[index].DependentOnStep = pluginStepIdsToStepDtoMapping[pluginStep.Id].DependentOnStep
			pluginSteps[index].UpdatedBy = userId
			pluginSteps[index].UpdatedOn = time.Now()

			pluginStepScript, err := impl.globalPluginRepository.GetScriptDetailById(pluginStep.ScriptId)
			if err != nil {
				impl.logger.Errorw("error in getting plugin step script", "scriptId", pluginStep.ScriptId, "pluginStepId", pluginStep.Id, "err", err)
				return err
			}
			pluginStepScript.Script = pluginStepIdsToStepDtoMapping[pluginStep.Id].PluginPipelineScript.Script
			pluginStepScript.StoreScriptAt = pluginStepIdsToStepDtoMapping[pluginStep.Id].PluginPipelineScript.StoreScriptAt
			pluginStepScript.Type = pluginStepIdsToStepDtoMapping[pluginStep.Id].PluginPipelineScript.Type
			pluginStepScript.DockerfileExists = pluginStepIdsToStepDtoMapping[pluginStep.Id].PluginPipelineScript.DockerfileExists
			pluginStepScript.MountPath = pluginStepIdsToStepDtoMapping[pluginStep.Id].PluginPipelineScript.MountPath
			pluginStepScript.MountCodeToContainer = pluginStepIdsToStepDtoMapping[pluginStep.Id].PluginPipelineScript.MountCodeToContainer
			pluginStepScript.MountCodeToContainerPath = pluginStepIdsToStepDtoMapping[pluginStep.Id].PluginPipelineScript.MountCodeToContainerPath
			pluginStepScript.MountDirectoryFromHost = pluginStepIdsToStepDtoMapping[pluginStep.Id].PluginPipelineScript.MountDirectoryFromHost
			pluginStepScript.ContainerImagePath = pluginStepIdsToStepDtoMapping[pluginStep.Id].PluginPipelineScript.ContainerImagePath
			pluginStepScript.ImagePullSecretType = pluginStepIdsToStepDtoMapping[pluginStep.Id].PluginPipelineScript.ImagePullSecretType
			pluginStepScript.ImagePullSecret = pluginStepIdsToStepDtoMapping[pluginStep.Id].PluginPipelineScript.ImagePullSecret
			pluginStepScript.UpdatedBy = userId
			pluginStepScript.UpdatedOn = time.Now()

			err = impl.globalPluginRepository.UpdatePluginPipelineScript(pluginStepScript, tx)
			if err != nil {
				impl.logger.Errorw("error in updating plugin step script", "err", err)
				return err
			}
		}
	}
	err := impl.globalPluginRepository.UpdateInBulkPluginSteps(pluginSteps, tx)
	if err != nil {
		impl.logger.Errorw("error in updating plugin steps in bulk", "err", err)
		return err
	}

	for _, pluginStepUpdateReq := range pluginStepsToUpdate {
		pluginStepVariablesToCreate, pluginStepVariablesToDelete, pluginStepVariablesToUpdate := filterPluginStepVariable(pluginStepUpdateReq.Id, pluginStepVariables, pluginStepUpdateReq.PluginStepVariable, userId)

		if len(pluginStepVariablesToCreate) > 0 {
			err := impl.saveDeepPluginStepVariableData(pluginStepUpdateReq.Id, pluginStepVariablesToCreate, userId, tx)
			if err != nil {
				impl.logger.Errorw("error in saveDeepPluginStepVariableData", "err", err)
				return err
			}
		}
		if len(pluginStepVariablesToDelete) > 0 {
			err := impl.deleteDeepPluginStepVariableData(pluginStepVariablesToDelete, pluginStepVariables, pluginStepConditions, userId, tx)
			if err != nil {
				impl.logger.Errorw("error in deleteDeepPluginStepVariableData", "err", err)
				return err
			}
		}
		if len(pluginStepVariablesToUpdate) > 0 {
			err := impl.updateDeepPluginStepVariableData(pluginStepUpdateReq.Id, pluginStepVariablesToUpdate, pluginStepVariables, pluginStepConditions, userId, tx)
			if err != nil {
				impl.logger.Errorw("error in updateDeepPluginStepVariableData", "err", err)
				return err
			}
		}
	}

	return nil
}

func (impl *GlobalPluginServiceImpl) updateDeepPluginStepVariableData(pluginStepId int, pluginStepVariablesToUpdate []*PluginVariableDto,
	pluginStepVariables []*repository.PluginStepVariable, pluginStepConditions []*repository.PluginStepCondition, userId int32, tx *pg.Tx) error {

	stepVariableIdsToStepVariableMapping := make(map[int]*PluginVariableDto)
	for _, stepVariable := range pluginStepVariablesToUpdate {
		stepVariableIdsToStepVariableMapping[stepVariable.Id] = stepVariable
	}
	for index, dbStepVariable := range pluginStepVariables {
		if _, ok := stepVariableIdsToStepVariableMapping[dbStepVariable.Id]; ok {

			pluginStepVariables[index].Name = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].Name
			pluginStepVariables[index].Format = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].Format
			pluginStepVariables[index].Description = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].Description
			pluginStepVariables[index].IsExposed = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].IsExposed
			pluginStepVariables[index].AllowEmptyValue = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].AllowEmptyValue
			pluginStepVariables[index].DefaultValue = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].DefaultValue
			pluginStepVariables[index].Value = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].Value
			pluginStepVariables[index].VariableType = repository.PluginStepVariableType(stepVariableIdsToStepVariableMapping[dbStepVariable.Id].VariableType)
			pluginStepVariables[index].ValueType = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].ValueType
			pluginStepVariables[index].PreviousStepIndex = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].PreviousStepIndex
			pluginStepVariables[index].VariableStepIndex = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].VariableStepIndex
			pluginStepVariables[index].VariableStepIndexInPlugin = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].VariableStepIndexInPlugin
			pluginStepVariables[index].ReferenceVariableName = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].ReferenceVariableName
			pluginStepVariables[index].UpdatedBy = userId
			pluginStepVariables[index].UpdatedOn = time.Now()
		}
	}
	err := impl.globalPluginRepository.UpdateInBulkPluginStepVariables(pluginStepVariables, tx)
	if err != nil {
		impl.logger.Errorw("error in updating plugin step variables in bulk", "err", err)
		return err
	}
	for _, pluginStepVariableReq := range pluginStepVariablesToUpdate {
		stepVariableConditionsToCreate, stepVariableConditionsToDelete, stepVariableConditionsToUpdate := filterPluginStepVariableConditions(pluginStepVariableReq.Id, pluginStepConditions, pluginStepVariableReq.PluginStepCondition, userId)

		if len(stepVariableConditionsToCreate) > 0 {
			err := impl.saveDeepStepVariableConditionsData(pluginStepId, pluginStepVariableReq.Id, stepVariableConditionsToCreate, userId, tx)
			if err != nil {
				impl.logger.Errorw("error in saveDeepStepVariableConditionsData", "err", err)
				return err
			}
		}
		if len(stepVariableConditionsToDelete) > 0 {
			err := impl.deleteDeepStepVariableConditionsData(stepVariableConditionsToDelete, pluginStepConditions, userId, tx)
			if err != nil {
				impl.logger.Errorw("error in deleteDeepStepVariableConditionsData", "err", err)
				return err
			}
		}
		if len(stepVariableConditionsToUpdate) > 0 {
			err := impl.updateDeepStepVariableConditionsData(stepVariableConditionsToUpdate, pluginStepConditions, userId, tx)
			if err != nil {
				impl.logger.Errorw("error in updateDeepStepVariableConditionsData", "err", err)
				return err
			}
		}
	}

	return nil
}

func (impl *GlobalPluginServiceImpl) saveDeepStepVariableConditionsData(pluginStepId int, pluginStepVariableId int, stepVariableConditionsToCreate []*repository.PluginStepCondition, userId int32, tx *pg.Tx) error {
	for _, pluginStepCondition := range stepVariableConditionsToCreate {
		pluginStepConditionData := &repository.PluginStepCondition{
			PluginStepId:        pluginStepId,
			ConditionVariableId: pluginStepVariableId,
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
		pluginStepConditionData, err := impl.globalPluginRepository.SavePluginStepConditions(pluginStepConditionData, tx)
		if err != nil {
			impl.logger.Errorw("saveDeepStepVariableConditionsData, error in saving plugin step condition", "pluginStepId", pluginStepVariableId, "err", err)
			return err
		}
	}
	return nil
}

func (impl *GlobalPluginServiceImpl) deleteDeepStepVariableConditionsData(stepVariableConditionsToDelete []*repository.PluginStepCondition, pluginStepConditions []*repository.PluginStepCondition, userId int32, tx *pg.Tx) error {

	stepVariableConditionsToDeleteIdsMapping := make(map[int]bool)
	for _, stepVariableConditionRemoveReq := range stepVariableConditionsToDelete {
		stepVariableConditionsToDeleteIdsMapping[stepVariableConditionRemoveReq.Id] = true
	}

	for index, stepVariableCondition := range pluginStepConditions {
		if _, ok := stepVariableConditionsToDeleteIdsMapping[stepVariableCondition.Id]; ok {
			pluginStepConditions[index].Deleted = true
			pluginStepConditions[index].UpdatedOn = time.Now()
			pluginStepConditions[index].UpdatedBy = userId
		}
	}
	err := impl.globalPluginRepository.UpdateInBulkPluginStepConditions(pluginStepConditions, tx)
	if err != nil {
		impl.logger.Errorw("deleteDeepStepVariableConditionsData, error in deleting plugin step conditions in bulk", "err", err)
		return err
	}

	return nil
}

func (impl *GlobalPluginServiceImpl) updateDeepStepVariableConditionsData(stepVariableConditionsToUpdate []*repository.PluginStepCondition,
	pluginStepConditions []*repository.PluginStepCondition, userId int32, tx *pg.Tx) error {

	stepVariableConditionsToDeleteIdsMapping := make(map[int]*repository.PluginStepCondition)
	for _, stepVariableConditionRemoveReq := range stepVariableConditionsToUpdate {
		stepVariableConditionsToDeleteIdsMapping[stepVariableConditionRemoveReq.Id] = stepVariableConditionRemoveReq
	}
	for index, stepVariableCondition := range pluginStepConditions {
		if _, ok := stepVariableConditionsToDeleteIdsMapping[stepVariableCondition.Id]; ok {
			pluginStepConditions[index].ConditionType = stepVariableConditionsToDeleteIdsMapping[stepVariableCondition.Id].ConditionType
			pluginStepConditions[index].ConditionalOperator = stepVariableConditionsToDeleteIdsMapping[stepVariableCondition.Id].ConditionalOperator
			pluginStepConditions[index].ConditionalValue = stepVariableConditionsToDeleteIdsMapping[stepVariableCondition.Id].ConditionalValue
			pluginStepConditions[index].UpdatedOn = time.Now()
			pluginStepConditions[index].UpdatedBy = userId
		}
	}
	err := impl.globalPluginRepository.UpdateInBulkPluginStepConditions(pluginStepConditions, tx)
	if err != nil {
		impl.logger.Errorw("updateDeepStepVariableConditionsData, error in updating plugin step conditions in bulk", "err", err)
		return err
	}
	return nil
}

func filterPluginStepVariableConditions(stepVariableId int, pluginStepConditionsInDb []*repository.PluginStepCondition, pluginStepConditionReq []*repository.PluginStepCondition, userId int32) ([]*repository.PluginStepCondition, []*repository.PluginStepCondition, []*repository.PluginStepCondition) {
	newStepVariableConditionsToCreate := make([]*repository.PluginStepCondition, 0)
	stepVariableConditionsToRemove := make([]*repository.PluginStepCondition, 0)
	stepVariableConditionsToUpdate := make([]*repository.PluginStepCondition, 0)

	stepIdToDbStepVariableConditionsMapping := make(map[int][]*repository.PluginStepCondition)
	for _, stepVariableConditionInDb := range pluginStepConditionsInDb {
		stepIdToDbStepVariableConditionsMapping[stepVariableConditionInDb.ConditionVariableId] = append(stepIdToDbStepVariableConditionsMapping[stepVariableConditionInDb.ConditionVariableId], stepVariableConditionInDb)
	}

	if len(pluginStepConditionReq) > len(stepIdToDbStepVariableConditionsMapping[stepVariableId]) {
		//it means there are new conditions for a variable in update request for a particular variable
		//filter out plugin variable conditions to create
		stepVariableConditionMapping := make(map[int]bool)
		for _, existingStepVariableCondition := range pluginStepConditionsInDb {
			stepVariableConditionMapping[existingStepVariableCondition.Id] = true
		}

		for _, stepVariableConditionReq := range pluginStepConditionReq {
			if _, ok := stepVariableConditionMapping[stepVariableConditionReq.Id]; !ok {
				newStepVariableConditionsToCreate = append(newStepVariableConditionsToCreate, stepVariableConditionReq)
			} else {
				stepVariableConditionsToUpdate = append(stepVariableConditionsToUpdate, stepVariableConditionReq)
			}
		}
	} else if len(pluginStepConditionReq) < len(stepIdToDbStepVariableConditionsMapping[stepVariableId]) {
		//it means there are deleted variable conditions in update request for a particular variable, filter out plugin variable conditions to delete
		stepVariableConditionMapping := make(map[int]*repository.PluginStepCondition)
		for _, variableConditionReq := range pluginStepConditionReq {
			stepVariableConditionMapping[variableConditionReq.Id] = variableConditionReq
		}

		for _, existingStepVariableCondition := range pluginStepConditionsInDb {
			if _, ok := stepVariableConditionMapping[existingStepVariableCondition.Id]; !ok {
				stepVariableConditionsToRemove = append(stepVariableConditionsToRemove, existingStepVariableCondition)
			} else {
				stepVariableConditionsToUpdate = append(stepVariableConditionsToUpdate, stepVariableConditionMapping[existingStepVariableCondition.Id])
			}
		}
	} else {
		return nil, nil, pluginStepConditionReq
	}

	return newStepVariableConditionsToCreate, stepVariableConditionsToRemove, stepVariableConditionsToUpdate
}

func filterPluginStepVariable(pluginStepId int, existingPluginStepVariables []*repository.PluginStepVariable,
	pluginStepVariableUpdateReq []*PluginVariableDto, userId int32) ([]*PluginVariableDto, []*PluginVariableDto, []*PluginVariableDto) {

	newPluginStepVariablesToCreate := make([]*PluginVariableDto, 0)
	pluginStepVariablesToRemove := make([]*PluginVariableDto, 0)
	pluginStepVariablesToUpdate := make([]*PluginVariableDto, 0)

	stepIdToDbStepVariablesMapping := make(map[int][]*repository.PluginStepVariable)
	for _, pluginStepVariableInDb := range existingPluginStepVariables {
		stepIdToDbStepVariablesMapping[pluginStepVariableInDb.PluginStepId] = append(stepIdToDbStepVariablesMapping[pluginStepVariableInDb.PluginStepId], pluginStepVariableInDb)
	}

	if len(pluginStepVariableUpdateReq) > len(stepIdToDbStepVariablesMapping[pluginStepId]) {
		//it means there are new variables in update request for a particular step, filter out plugin variables to create
		pluginStepVariableIdToStepVariableMapping := make(map[int]bool)
		for _, existingPluginStepVariable := range existingPluginStepVariables {
			pluginStepVariableIdToStepVariableMapping[existingPluginStepVariable.Id] = true
		}
		for _, stepVariableUpdateReq := range pluginStepVariableUpdateReq {
			if _, ok := pluginStepVariableIdToStepVariableMapping[stepVariableUpdateReq.Id]; !ok {
				newPluginStepVariablesToCreate = append(newPluginStepVariablesToCreate, stepVariableUpdateReq)
			} else {
				pluginStepVariablesToUpdate = append(pluginStepVariablesToUpdate, stepVariableUpdateReq)
			}
		}
	} else if len(pluginStepVariableUpdateReq) < len(stepIdToDbStepVariablesMapping[pluginStepId]) {
		//it means there are deleted variables in update request for a particular step, filter out plugin variables to delete
		pluginStepVariableIdToStepVariableMapping := make(map[int]*PluginVariableDto)
		for _, stepVariableUpdateReq := range pluginStepVariableUpdateReq {
			pluginStepVariableIdToStepVariableMapping[stepVariableUpdateReq.Id] = stepVariableUpdateReq
		}

		for _, existingStepVariable := range existingPluginStepVariables {
			if _, ok := pluginStepVariableIdToStepVariableMapping[existingStepVariable.Id]; !ok {
				pluginStepVariablesToRemove = append(pluginStepVariablesToRemove, &PluginVariableDto{Id: existingStepVariable.Id})
			} else {
				pluginStepVariablesToUpdate = append(pluginStepVariablesToUpdate, pluginStepVariableIdToStepVariableMapping[existingStepVariable.Id])
			}
		}
	} else {
		return nil, nil, pluginStepVariableUpdateReq
	}

	return newPluginStepVariablesToCreate, pluginStepVariablesToRemove, pluginStepVariablesToUpdate
}

func (impl *GlobalPluginServiceImpl) saveDeepPluginStepVariableData(pluginStepId int, pluginStepVariablesToCreate []*PluginVariableDto, userId int32, tx *pg.Tx) error {
	for _, pluginStepVariable := range pluginStepVariablesToCreate {
		pluginStepVariableData := &repository.PluginStepVariable{
			PluginStepId:              pluginStepId,
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
		pluginStepVariableData, err := impl.globalPluginRepository.SavePluginStepVariables(pluginStepVariableData, tx)
		if err != nil {
			impl.logger.Errorw("saveDeepPluginStepVariableData, error in saving plugin step variable", "pluginStepVariableData", pluginStepVariableData, "err", err)
			return err
		}

		for _, pluginStepCondition := range pluginStepVariable.PluginStepCondition {
			pluginStepConditionData := &repository.PluginStepCondition{
				PluginStepId:        pluginStepId,
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
				impl.logger.Errorw("saveDeepPluginStepVariableData, error in saving plugin step condition", "pluginStepId", pluginStepId, "err", err)
				return err
			}
		}
	}
	return nil
}

func (impl *GlobalPluginServiceImpl) deleteDeepPluginStepVariableData(pluginStepVariablesToDelete []*PluginVariableDto,
	pluginStepVariables []*repository.PluginStepVariable, pluginStepConditions []*repository.PluginStepCondition, userId int32, tx *pg.Tx) error {

	stepVariablesToDeleteIdsMapping := make(map[int]bool)
	for _, stepVariableRemoveReq := range pluginStepVariablesToDelete {
		stepVariablesToDeleteIdsMapping[stepVariableRemoveReq.Id] = true
	}
	for index, stepVariable := range pluginStepVariables {
		if _, ok := stepVariablesToDeleteIdsMapping[stepVariable.Id]; ok {
			pluginStepVariables[index].Deleted = true
			pluginStepVariables[index].UpdatedOn = time.Now()
			pluginStepVariables[index].UpdatedBy = userId
		}
	}
	for index, stepVariableCondition := range pluginStepConditions {
		if _, ok := stepVariablesToDeleteIdsMapping[stepVariableCondition.ConditionVariableId]; ok {
			pluginStepConditions[index].Deleted = true
			pluginStepConditions[index].UpdatedOn = time.Now()
			pluginStepConditions[index].UpdatedBy = userId
		}
	}
	err := impl.globalPluginRepository.UpdateInBulkPluginStepVariables(pluginStepVariables, tx)
	if err != nil {
		impl.logger.Errorw("deleteDeepPluginStepData, error in deleting plugin step variables", "err", err)
		return err
	}
	err = impl.globalPluginRepository.UpdateInBulkPluginStepConditions(pluginStepConditions, tx)
	if err != nil {
		impl.logger.Errorw("deleteDeepPluginStepData, error in deleting plugin step conditions", "err", err)
		return err
	}

	return nil
}

func (impl *GlobalPluginServiceImpl) deleteDeepPluginStepData(pluginStepsToRemove []*PluginStepsDto, pluginStepVariables []*repository.PluginStepVariable,
	pluginStepConditions []*repository.PluginStepCondition, pluginSteps []*repository.PluginStep, userId int32, tx *pg.Tx) error {
	pluginStepsToRemoveIdsMapping := make(map[int]bool)
	for _, pluginStepRemoveReq := range pluginStepsToRemove {
		pluginStepsToRemoveIdsMapping[pluginStepRemoveReq.Id] = true
	}
	for index, pluginStep := range pluginSteps {
		if _, ok := pluginStepsToRemoveIdsMapping[pluginStep.Id]; ok {
			pluginSteps[index].Deleted = true
			pluginSteps[index].UpdatedBy = userId
			pluginSteps[index].UpdatedOn = time.Now()

			pluginStepScript, err := impl.globalPluginRepository.GetScriptDetailById(pluginStep.ScriptId)
			if err != nil {
				impl.logger.Errorw("error in getting plugin step script", "scriptId", pluginStep.ScriptId, "pluginStepId", pluginStep.Id, "err", err)
				return err
			}
			pluginStepScript.Deleted = true
			pluginStepScript.UpdatedBy = userId
			pluginStepScript.UpdatedOn = time.Now()
			err = impl.globalPluginRepository.UpdatePluginPipelineScript(pluginStepScript, tx)
			if err != nil {
				impl.logger.Errorw("error in updating plugin step script", "err", err)
				return err
			}
		}
	}
	for index, _ := range pluginStepVariables {
		if _, ok := pluginStepsToRemoveIdsMapping[pluginStepVariables[index].PluginStepId]; ok {
			pluginStepVariables[index].Deleted = true
			pluginStepVariables[index].UpdatedOn = time.Now()
			pluginStepVariables[index].UpdatedBy = userId
		}
	}
	for index, _ := range pluginStepConditions {
		if _, ok := pluginStepsToRemoveIdsMapping[pluginStepConditions[index].PluginStepId]; ok {
			pluginStepConditions[index].Deleted = true
			pluginStepConditions[index].UpdatedOn = time.Now()
			pluginStepConditions[index].UpdatedBy = userId
		}
	}
	err := impl.globalPluginRepository.UpdateInBulkPluginSteps(pluginSteps, tx)
	if err != nil {
		impl.logger.Errorw("deleteDeepPluginStepData, error in deleting plugin steps", "err", err)
		return err
	}
	err = impl.globalPluginRepository.UpdateInBulkPluginStepVariables(pluginStepVariables, tx)
	if err != nil {
		impl.logger.Errorw("deleteDeepPluginStepData, error in deleting plugin step variables", "err", err)
		return err
	}
	err = impl.globalPluginRepository.UpdateInBulkPluginStepConditions(pluginStepConditions, tx)
	if err != nil {
		impl.logger.Errorw("deleteDeepPluginStepData, error in deleting plugin step conditions", "err", err)
		return err
	}
	return nil
}

func filterPluginStepData(existingPluginStepsInDb []*repository.PluginStep, pluginStepUpdateReq []*PluginStepsDto) ([]*PluginStepsDto, []*PluginStepsDto, []*PluginStepsDto) {
	newPluginStepsToCreate := make([]*PluginStepsDto, 0)
	pluginStepsToRemove := make([]*PluginStepsDto, 0)
	pluginStepsToUpdate := make([]*PluginStepsDto, 0)

	if len(pluginStepUpdateReq) > len(existingPluginStepsInDb) {
		//new plugin step found
		pluginIdMapping := make(map[int]bool)
		for _, existingPluginStep := range existingPluginStepsInDb {
			pluginIdMapping[existingPluginStep.Id] = true
		}
		for _, pluginStepReq := range pluginStepUpdateReq {
			if _, ok := pluginIdMapping[pluginStepReq.Id]; !ok {
				newPluginStepsToCreate = append(newPluginStepsToCreate, pluginStepReq)
			} else {
				pluginStepsToUpdate = append(pluginStepsToUpdate, pluginStepReq)
			}
		}
	} else if len(pluginStepUpdateReq) < len(existingPluginStepsInDb) {
		pluginIdMapping := make(map[int]*PluginStepsDto)
		for _, pluginStepReq := range pluginStepUpdateReq {
			pluginIdMapping[pluginStepReq.Id] = pluginStepReq
		}
		for _, existingPluginStep := range existingPluginStepsInDb {
			if _, ok := pluginIdMapping[existingPluginStep.Id]; !ok {
				pluginStepsToRemove = append(pluginStepsToRemove, &PluginStepsDto{Id: existingPluginStep.Id})
			} else {
				pluginStepsToUpdate = append(pluginStepsToUpdate, pluginIdMapping[existingPluginStep.Id])
			}
		}
	} else {
		return nil, nil, pluginStepUpdateReq
	}

	return newPluginStepsToCreate, pluginStepsToRemove, pluginStepsToUpdate
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

func (impl *GlobalPluginServiceImpl) deletePlugin(pluginDeleteReq *PluginMetadataDto, userId int32) (*PluginMetadataDto, error) {
	dbConnection := impl.globalPluginRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error
	defer tx.Rollback()
	pluginMetaData, err := impl.globalPluginRepository.GetMetaDataByPluginId(pluginDeleteReq.Id)
	if err != nil {
		impl.logger.Errorw("deletePlugin, error in getting pluginMetadata, pluginId does not exist", "pluginId", pluginDeleteReq.Id, "err", err)
		return nil, err
	}
	pluginMetaData.Deleted = true
	pluginMetaData.UpdatedBy = userId
	pluginMetaData.UpdatedOn = time.Now()
	err = impl.globalPluginRepository.UpdatePluginMetadata(pluginMetaData, tx)
	if err != nil {
		impl.logger.Errorw("deletePlugin, error in deleting pluginMetadata", "pluginId", pluginDeleteReq.Id, "err", err)
		return nil, err
	}
	pluginSteps, err := impl.globalPluginRepository.GetPluginStepsByPluginId(pluginDeleteReq.Id)
	if err != nil {
		impl.logger.Errorw("deletePlugin, error in getting pluginSteps", "pluginId", pluginDeleteReq.Id, "err", err)
		return nil, err
	}

	for _, pluginStep := range pluginSteps {
		pluginStep.Deleted = true
		pluginStep.UpdatedBy = userId
		pluginStep.UpdatedOn = time.Now()
	}
	err = impl.globalPluginRepository.UpdateInBulkPluginSteps(pluginSteps, tx)
	if err != nil {
		impl.logger.Errorw("deletePlugin, error in deleting plugin steps in bulk", "pluginId", pluginMetaData.Id, "err", err)
		return nil, err
	}

	pluginStepVariables, err := impl.globalPluginRepository.GetExposedVariablesByPluginId(pluginDeleteReq.Id)
	if err != nil {
		impl.logger.Errorw("deletePlugin, error in getting pluginStepVariables", "pluginId", pluginDeleteReq.Id, "err", err)
		return nil, err
	}
	for index, _ := range pluginStepVariables {
		pluginStepVariables[index].Deleted = true
		pluginStepVariables[index].UpdatedBy = userId
		pluginStepVariables[index].UpdatedOn = time.Now()
	}
	err = impl.globalPluginRepository.UpdateInBulkPluginStepVariables(pluginStepVariables, tx)
	if err != nil {
		impl.logger.Errorw("deletePlugin, error in deleting plugin step variables in bulk", "pluginId", pluginMetaData.Id, "err", err)
		return nil, err
	}

	pluginStepConditions, err := impl.globalPluginRepository.GetConditionsByPluginId(pluginDeleteReq.Id)
	if err != nil {
		impl.logger.Errorw("deletePlugin, error in getting pluginStepConditions", "pluginId", pluginDeleteReq.Id, "err", err)
		return nil, err
	}

	for index, _ := range pluginStepConditions {
		pluginStepConditions[index].Deleted = true
		pluginStepConditions[index].UpdatedBy = userId
		pluginStepConditions[index].UpdatedOn = time.Now()
	}

	err = impl.globalPluginRepository.UpdateInBulkPluginStepConditions(pluginStepConditions, tx)
	if err != nil {
		impl.logger.Errorw("deletePlugin, error in deleting plugin step variable conditions in bulk", "pluginId", pluginMetaData.Id, "err", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("deletePlugin, error in committing db transaction", "err", err)
		return nil, err
	}
	return pluginDeleteReq, nil
}
