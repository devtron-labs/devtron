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
}

type GlobalPluginService interface {
	GetAllGlobalVariables() ([]*GlobalVariable, error)
	ListAllPlugins() ([]*PluginListComponentDto, error)
	GetPluginDetailById(pluginId int) (*PluginDetailDto, error)
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
		},
		{
			Name:        "DOCKER_IMAGE_TAG",
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Tag going to be used to push image.",
		},
		{
			Name:        "DOCKER_REPOSITORY",
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Name of the repository to be used for pushing images.",
		},
		{
			Name:        "DOCKER_REGISTRY_URL",
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Url of the container registry used for this pipeline.",
		},
		{
			Name:        "DOCKER_IMAGE",
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Complete image name(repository+registry+tag).",
		},
		{
			Name:        "APP_NAME",
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Name of the app this pipeline resides in.",
		},
		{
			Name:        "TRIGGER_BY_AUTHOR",
			Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
			Description: "Email-Id/Name of the user who triggers pipeline.",
		},
	}
	return globalVariables, nil
}

func (impl *GlobalPluginServiceImpl) ListAllPlugins() ([]*PluginListComponentDto, error) {
	impl.logger.Infow("request received, ListAllPlugins")
	var pluginDetails []*PluginListComponentDto
	//getting all plugins metadata(without tags)
	pluginsMetadata, err := impl.globalPluginRepository.GetMetaDataForAllPlugins()
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
