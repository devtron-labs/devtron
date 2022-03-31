package plugin

import (
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type GlobalPluginService interface {
	ListAllPlugins() ([]*PluginMetadataDto, error)
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

func (impl *GlobalPluginServiceImpl) ListAllPlugins() ([]*PluginMetadataDto, error) {
	impl.logger.Infow("request received, ListAllPlugins")

	var plugins []*PluginMetadataDto

	//getting all plugins metadata(without tags)
	pluginsMetadata, err := impl.globalPluginRepository.GetMetaDataForAllPlugins()
	if err != nil {
		impl.logger.Errorw("error in getting plugins", "err", err)
		return nil, err
	}
	for _, pluginMetadata := range pluginsMetadata {
		//getting tags for this pluginId
		tags, err := impl.globalPluginRepository.GetTagsByPluginId(pluginMetadata.Id)
		if err != nil && err != pg.ErrNoRows {
			//only logging err, not returning err for tags
			impl.logger.Errorw("error in getting tags by pluginId", "err", err, "pluginId", pluginMetadata.Id)
		}
		plugin := &PluginMetadataDto{
			Id:          pluginMetadata.Id,
			Name:        pluginMetadata.Name,
			Type:        string(pluginMetadata.Type),
			Description: pluginMetadata.Description,
			Icon:        pluginMetadata.Icon,
			Tags:        tags,
		}
		plugins = append(plugins, plugin)
	}
	return plugins, nil
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

	//getting exposed variables
	pluginVariables, err := impl.globalPluginRepository.GetExposedVariablesByPluginId(pluginId)
	if err != nil {
		impl.logger.Errorw("error in getting pluginVariables by pluginId", "err", err, "pluginId", pluginId)
		return nil, err
	}

	var inputVariablesDto []*PluginVariableDto
	var outputVariablesDto []*PluginVariableDto

	for _, pluginVariable := range pluginVariables {
		variableDto := &PluginVariableDto{
			Id:                    pluginVariable.Id,
			Name:                  pluginVariable.Name,
			Format:                pluginVariable.Format,
			Description:           pluginVariable.Description,
			IsExposed:             pluginVariable.IsExposed,
			AllowEmptyValue:       pluginVariable.AllowEmptyValue,
			DefaultValue:          pluginVariable.DefaultValue,
			Value:                 pluginVariable.Value,
			ValueType:             string(pluginVariable.ValueType),
			VariableType:          string(pluginVariable.VariableType),
			PreviousStepIndex:     pluginVariable.PreviousStepIndex,
			ReferenceVariableName: pluginVariable.ReferenceVariableName,
			Deleted:               pluginVariable.Deleted,
		}
		if pluginVariable.VariableType == repository.PLUGIN_VARIABLE_TYPE_INPUT {
			inputVariablesDto = append(inputVariablesDto, variableDto)
		} else if pluginVariable.VariableType == repository.PLUGIN_VARIABLE_TYPE_OUTPUT {
			outputVariablesDto = append(outputVariablesDto, variableDto)
		}
	}
	pluginDetail.InputVariables = inputVariablesDto
	pluginDetail.OutputVariables = outputVariablesDto
	return pluginDetail, nil
}
