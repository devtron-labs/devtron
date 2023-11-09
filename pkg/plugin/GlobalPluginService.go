package plugin

import (
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type GlobalVariable struct {
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	Format      string `json:"format"`
	Description string `json:"description"`
	Type        string `json:"stageType"`
}

const (
	DOCKER_IMAGE                 = "DOCKER_IMAGE"
	DEPLOYMENT_RELEASE_ID        = "DEPLOYMENT_RELEASE_ID"
	DEPLOYMENT_UNIQUE_ID         = "DEPLOYMENT_UNIQUE_ID"
	CD_TRIGGERED_BY              = "CD_TRIGGERED_BY"
	CD_TRIGGER_TIME              = "CD_TRIGGER_TIME"
	APP_NAME                     = "APP_NAME"
	DEVTRON_CD_TRIGGERED_BY      = "DEVTRON_CD_TRIGGERED_BY"
	DEVTRON_CD_TRIGGER_TIME      = "DEVTRON_CD_TRIGGER_TIME"
	CD_PIPELINE_ENV_NAME_KEY     = "CD_PIPELINE_ENV_NAME"
	CD_PIPELINE_CLUSTER_NAME_KEY = "CD_PIPELINE_CLUSTER_NAME"
	GIT_METADATA                 = "GIT_METADATA"
	CHILD_CD_METADATA            = "CHILD_CD_METADATA"
	APP_LABEL_METADATA           = "APP_LABEL_METADATA"
)

type GlobalPluginService interface {
	GetAllGlobalVariables() ([]*GlobalVariable, error)
	ListAllPlugins(stageType int) ([]*PluginListComponentDto, error)
	GetPluginDetailById(pluginId int) (*PluginDetailDto, error)
	GetRefPluginIdByRefPluginName(pluginName string) (refPluginId int, err error)
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
			Description: "RedirectionUrl of the container registry used for this pipeline.",
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
			Name:        CD_PIPELINE_ENV_NAME_KEY,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "The name of the environment for which this deployment pipeline is configured.",
			Type:        "cd",
		},
		{
			Name:        CD_PIPELINE_CLUSTER_NAME_KEY,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "The name of the cluster to which the environment belongs for which this deployment pipeline is configured.",
			Type:        "cd",
		},
		{
			Name:        DOCKER_IMAGE,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Complete image name(repository+registry+tag).",
			Type:        "cd",
		},
		{
			Name:        APP_NAME,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "The name of the app this pipeline resides in.",
			Type:        "cd",
		},
		{
			Name:        DEPLOYMENT_RELEASE_ID,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Auto-incremented counter for deployment triggers.",
			Type:        "post-cd",
		},
		{
			Name:        DEPLOYMENT_UNIQUE_ID,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Auto-incremented counter for deployment triggers. Counter is shared between Pre/Post/Deployment stages.",
			Type:        "cd",
		},
		{
			Name:        CD_TRIGGERED_BY,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Email-Id/Name of the user who triggered the deployment pipeline.",
			Type:        "post-cd",
		},
		{
			Name:        CD_TRIGGER_TIME,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Time when the deployment pipeline was triggered.",
			Type:        "post-cd",
		},
		{
			Name:        GIT_METADATA,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "GIT_METADATA consists of GIT_COMMIT_HASH, GIT_SOURCE_TYPE, GIT_SOURCE_VALUE.",
			Type:        "cd",
		},
		{
			Name:        APP_LABEL_METADATA,
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "APP_LABEL_METADATA consists of APP_LABEL_KEY, APP_LABEL_VALUE. APP_LABEL_METADATA will only be available if workflow has External CI.",
			Type:        "cd",
		},
		{
			Name:        CHILD_CD_METADATA,
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

func (impl *GlobalPluginServiceImpl) GetRefPluginIdByRefPluginName(pluginName string) (refPluginId int, err error) {
	pluginMetadata, err := impl.globalPluginRepository.GetPluginByName(pluginName)
	if err != nil {
		impl.logger.Errorw("error in fetching plugin metadata by name", "err", err)
		return 0, err
	}
	if pluginMetadata == nil {
		return 0, nil
	}
	return pluginMetadata[0].Id, nil
}
