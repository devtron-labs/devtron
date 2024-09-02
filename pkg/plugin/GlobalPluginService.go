/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugin

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/plugin/bean"
	helper2 "github.com/devtron-labs/devtron/pkg/plugin/helper"
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/plugin/utils"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"strconv"
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

const (
	APP                          = "app"
	JOB                          = "job"
	DOCKER_IMAGE                 = "DOCKER_IMAGE"
	DEPLOYMENT_RELEASE_ID        = "DEPLOYMENT_RELEASE_ID"
	DEPLOYMENT_UNIQUE_ID         = "DEPLOYMENT_UNIQUE_ID"
	CD_TRIGGERED_BY              = "CD_TRIGGERED_BY"
	CD_TRIGGER_TIME              = "CD_TRIGGER_TIME"
	APP_NAME                     = "APP_NAME"
	JOB_NAME                     = "JOB_NAME"
	DEVTRON_CD_TRIGGERED_BY      = "DEVTRON_CD_TRIGGERED_BY"
	DEVTRON_CD_TRIGGER_TIME      = "DEVTRON_CD_TRIGGER_TIME"
	CD_PIPELINE_ENV_NAME_KEY     = "CD_PIPELINE_ENV_NAME"
	CD_PIPELINE_CLUSTER_NAME_KEY = "CD_PIPELINE_CLUSTER_NAME"
	GIT_METADATA                 = "GIT_METADATA"
	CHILD_CD_METADATA            = "CHILD_CD_METADATA"
	APP_LABEL_METADATA           = "APP_LABEL_METADATA"
)

type GlobalPluginService interface {
	GetAllGlobalVariables(appType helper.AppType) ([]*GlobalVariable, error)
	ListAllPlugins(stageTypeReq string) ([]*bean2.PluginListComponentDto, error)
	GetPluginDetailById(pluginId int) (*bean2.PluginDetailDto, error)
	GetRefPluginIdByRefPluginName(pluginName string) (refPluginId int, err error)
	PatchPlugin(pluginDto *bean2.PluginMetadataDto, userId int32) (*bean2.PluginMetadataDto, error)
	GetDetailedPluginInfoByPluginId(pluginId int) (*bean2.PluginMetadataDto, error)
	GetAllDetailedPluginInfo() ([]*bean2.PluginMetadataDto, error)

	ListAllPluginsV2(filter *bean2.PluginsListFilter) (*bean2.PluginsDto, error)
	GetPluginDetailV2(pluginVersionIds, parentPluginIds []int, fetchAllVersionDetails bool) (*bean2.PluginsDto, error)
	GetAllUniqueTags() (*bean2.PluginTagsDto, error)
	MigratePluginData() error
}

func NewGlobalPluginService(logger *zap.SugaredLogger, globalPluginRepository repository.GlobalPluginRepository,
	pipelineStageRepository repository2.PipelineStageRepository, userService user.UserService) *GlobalPluginServiceImpl {
	return &GlobalPluginServiceImpl{
		logger:                  logger,
		globalPluginRepository:  globalPluginRepository,
		pipelineStageRepository: pipelineStageRepository,
		userService:             userService,
	}
}

type GlobalPluginServiceImpl struct {
	logger                  *zap.SugaredLogger
	globalPluginRepository  repository.GlobalPluginRepository
	pipelineStageRepository repository2.PipelineStageRepository
	userService             user.UserService
}

func (impl *GlobalPluginServiceImpl) GetAllGlobalVariables(appType helper.AppType) ([]*GlobalVariable, error) {
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
	appName := APP_NAME
	entityType := APP
	if appType == helper.Job {
		appName = JOB_NAME
		entityType = JOB
	}
	globalVariable := &GlobalVariable{
		Name:        appName,
		Format:      string(repository.PLUGIN_VARIABLE_FORMAT_TYPE_STRING),
		Description: fmt.Sprintf("Name of the %s this pipeline resides in.", entityType),
		Type:        "ci",
	}
	globalVariables = append(globalVariables, globalVariable)
	return globalVariables, nil
}

func (impl *GlobalPluginServiceImpl) ListAllPlugins(stageTypeReq string) ([]*bean2.PluginListComponentDto, error) {
	impl.logger.Infow("request received, ListAllPlugins")
	var pluginDetails []*bean2.PluginListComponentDto
	pluginsMetadata := make([]*repository.PluginMetadata, 0)
	var err error

	//getting all plugins metadata(without tags)
	if len(stageTypeReq) == 0 {
		pluginsMetadata, err = impl.globalPluginRepository.GetMetaDataForAllPlugins()
		if err != nil {
			impl.logger.Errorw("error in getting plugins", "err", err)
			return nil, err
		}
	} else {
		stageType, err := utils.GetStageType(stageTypeReq)
		if err != nil {
			return nil, err
		}
		pluginsMetadata, err = impl.globalPluginRepository.GetMetaDataForPluginWithStageType(stageType)
		if err != nil {
			impl.logger.Errorw("error in getting plugins", "err", err)
			return nil, err
		}
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
		pluginMetadataDto := &bean2.PluginMetadataDto{
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
		pluginDetail := &bean2.PluginListComponentDto{
			PluginMetadataDto: pluginMetadataDto,
			InputVariables:    pluginIdInputVariablesMap[pluginMetadata.Id],
			OutputVariables:   pluginIdOutputVariablesMap[pluginMetadata.Id],
		}
		pluginDetails = append(pluginDetails, pluginDetail)
	}
	return pluginDetails, nil
}

func (impl *GlobalPluginServiceImpl) GetPluginDetailById(pluginId int) (*bean2.PluginDetailDto, error) {
	impl.logger.Infow("request received, GetPluginDetail", "pluginId", pluginId)

	//getting metadata
	pluginMetadata, err := impl.globalPluginRepository.GetMetaDataByPluginId(pluginId)
	if err != nil {
		impl.logger.Errorw("error in getting plugins", "err", err, "pluginId", pluginId)
		return nil, err
	}
	metadataDto := &bean2.PluginMetadataDto{
		Id:          pluginMetadata.Id,
		Name:        pluginMetadata.Name,
		Type:        string(pluginMetadata.Type),
		Description: pluginMetadata.Description,
		Icon:        pluginMetadata.Icon,
	}
	pluginDetail := &bean2.PluginDetailDto{
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

func (impl *GlobalPluginServiceImpl) getPluginIdVariablesMap() (map[int][]*bean2.PluginVariableDto, map[int][]*bean2.PluginVariableDto, error) {
	variables, err := impl.globalPluginRepository.GetExposedVariablesForAllPlugins()
	if err != nil {
		impl.logger.Errorw("error in getting exposed vars for all plugins", "err", err)
		return nil, nil, err
	}
	pluginIdInputVarsMap, pluginIdOutputVarsMap := make(map[int][]*bean2.PluginVariableDto), make(map[int][]*bean2.PluginVariableDto)
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

func (impl *GlobalPluginServiceImpl) getIOVariablesOfAPlugin(pluginId int) (inputVariablesDto, outputVariablesDto []*bean2.PluginVariableDto, err error) {
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

func getVariableDto(pluginVariable *repository.PluginStepVariable) *bean2.PluginVariableDto {
	return &bean2.PluginVariableDto{
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

func (impl *GlobalPluginServiceImpl) PatchPlugin(pluginDto *bean2.PluginMetadataDto, userId int32) (*bean2.PluginMetadataDto, error) {

	switch pluginDto.Action {
	case bean2.CREATEPLUGIN:
		pluginData, err := impl.createPlugin(pluginDto, userId)
		if err != nil {
			impl.logger.Errorw("error in creating plugin", "err", err, "pluginDto", pluginDto)
			return nil, err
		}
		return pluginData, nil
	case bean2.UPDATEPLUGIN:
		pluginData, err := impl.updatePlugin(pluginDto, userId)
		if err != nil {
			impl.logger.Errorw("error in updating plugin", "err", err, "pluginDto", pluginDto)
			return nil, err
		}
		return pluginData, nil
	case bean2.DELETEPLUGIN:
		pluginData, err := impl.deletePlugin(pluginDto, userId)
		if err != nil {
			impl.logger.Errorw("error in deleting plugin", "err", err, "pluginDto", pluginDto)
			return nil, err
		}
		return pluginData, nil
	default:
		impl.logger.Errorw("unsupported operation ", "op", pluginDto.Action)
		return nil, fmt.Errorf("unsupported operation %d", pluginDto.Action)
	}

	return nil, nil
}

func (impl *GlobalPluginServiceImpl) validatePluginRequest(pluginReq *bean2.PluginMetadataDto) error {
	if len(pluginReq.Type) == 0 {
		return errors.New("invalid plugin type, should be of the type PRESET or SHARED")
	}

	plugins, err := impl.globalPluginRepository.GetMetaDataForAllPlugins()
	if err != nil {
		impl.logger.Errorw("error in getting all plugins", "err", err)
		return err
	}
	for _, plugin := range plugins {
		if plugin.Name == pluginReq.Name {
			return errors.New("plugin with the same name exists, please choose another name")
		}
	}
	return nil
}

func (impl *GlobalPluginServiceImpl) createPlugin(pluginReq *bean2.PluginMetadataDto, userId int32) (*bean2.PluginMetadataDto, error) {
	err := impl.validatePluginRequest(pluginReq)
	if err != nil {
		return nil, err
	}

	dbConnection := impl.globalPluginRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	//create entry in plugin_metadata
	pluginMetadata := &repository.PluginMetadata{}
	if pluginReq != nil {
		pluginMetadata = pluginReq.GetPluginMetadataSqlObj(userId)
	}
	pluginMetadata, err = impl.globalPluginRepository.SavePluginMetadata(pluginMetadata, tx)
	if err != nil {
		impl.logger.Errorw("createPlugin, error in saving plugin", "pluginDto", pluginReq, "err", err)
		return nil, err
	}
	pluginReq.Id = pluginMetadata.Id
	pluginStage := repository.CI_CD
	if pluginReq.PluginStage == bean2.CI_TYPE_PLUGIN {
		pluginStage = repository.CI
	} else if pluginReq.PluginStage == bean2.CD_TYPE_PLUGIN {
		pluginStage = repository.CD
	}
	pluginStageMapping := &repository.PluginStageMapping{
		PluginId:  pluginMetadata.Id,
		StageType: pluginStage,
		AuditLog:  sql.NewDefaultAuditLog(userId),
	}
	_, err = impl.globalPluginRepository.SavePluginStageMapping(pluginStageMapping, tx)
	if err != nil {
		impl.logger.Errorw("createPlugin, error in saving plugin stage mapping", "pluginDto", pluginReq, "err", err)
		return nil, err
	}

	err = impl.saveDeepPluginStepData(pluginMetadata.Id, pluginReq.PluginSteps, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in saving plugin step data", "err", err)
		return nil, err
	}
	isUpdateReq := false
	err = impl.CreateNewPluginTagsAndRelationsIfRequired(pluginReq, isUpdateReq, userId, tx)
	if err != nil {
		impl.logger.Errorw("createPlugin, error in CreateNewPluginTagsAndRelationsIfRequired", "err", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("createPlugin, error in committing db transaction", "err", err)
		return nil, err
	}
	return pluginReq, nil
}

func (impl *GlobalPluginServiceImpl) CreateNewPluginTagsAndRelationsIfRequired(pluginReq *bean2.PluginMetadataDto, isUpdateReq bool, userId int32, tx *pg.Tx) error {
	allPluginTags, err := impl.globalPluginRepository.GetAllPluginTags()
	if err != nil {
		impl.logger.Errorw("error in getting all plugin tags", "err", err)
		return err
	}
	//check for new tags, then create new plugin_tag and plugin_tag_relation entry in db when new tags are present in request
	newPluginTagsToCreate := make([]*repository.PluginTag, 0)
	newPluginTagRelationsToCreate := make([]*repository.PluginTagRelation, 0)

	pluginTagNameToPluginTagFromDb := make(map[string]map[string]repository.PluginTag)

	for _, pluginTagReq := range pluginReq.Tags {
		tagAlreadyExists := false
		for _, presentPluginTags := range allPluginTags {
			if strings.ToLower(pluginTagReq) == strings.ToLower(presentPluginTags.Name) {
				tagAlreadyExists = true
				if _, ok := pluginTagNameToPluginTagFromDb[repository.EXISTING_TAG_TYPE]; !ok {
					pluginTagNameToPluginTagFromDb[repository.EXISTING_TAG_TYPE] = make(map[string]repository.PluginTag)
				}
				pluginTagNameToPluginTagFromDb[repository.EXISTING_TAG_TYPE][pluginTagReq] = *presentPluginTags
			}
		}
		if !tagAlreadyExists {
			newPluginTag := &repository.PluginTag{
				Name:     pluginTagReq,
				AuditLog: sql.NewDefaultAuditLog(userId),
			}
			newPluginTagsToCreate = append(newPluginTagsToCreate, newPluginTag)
		}
	}
	if len(newPluginTagsToCreate) > 0 {
		err = impl.globalPluginRepository.SavePluginTagInBulk(newPluginTagsToCreate, tx)
		if err != nil {
			impl.logger.Errorw("error in saving plugin tag", "newPluginTags", newPluginTagsToCreate, "err", err)
			return err
		}
	}
	for _, newPluginTag := range newPluginTagsToCreate {
		if _, ok := pluginTagNameToPluginTagFromDb[repository.NEW_TAG_TYPE]; !ok {
			pluginTagNameToPluginTagFromDb[repository.NEW_TAG_TYPE] = make(map[string]repository.PluginTag)
		}
		pluginTagNameToPluginTagFromDb[repository.NEW_TAG_TYPE][newPluginTag.Name] = *newPluginTag
	}

	for _, tagReq := range pluginReq.Tags {
		for tagType, tagMapping := range pluginTagNameToPluginTagFromDb {
			if tagType == repository.EXISTING_TAG_TYPE && isUpdateReq {
				continue
			}
			if _, ok := tagMapping[tagReq]; ok {
				newPluginTagRelation := &repository.PluginTagRelation{
					TagId:    tagMapping[tagReq].Id,
					PluginId: pluginReq.Id,
					AuditLog: sql.NewDefaultAuditLog(userId),
				}
				newPluginTagRelationsToCreate = append(newPluginTagRelationsToCreate, newPluginTagRelation)
			}

		}
	}

	if len(newPluginTagRelationsToCreate) > 0 {
		err = impl.globalPluginRepository.SavePluginTagRelationInBulk(newPluginTagRelationsToCreate, tx)
		if err != nil {
			impl.logger.Errorw("error in saving plugin tag relation in bulk", "newPluginTagRelationsToCreate", newPluginTagRelationsToCreate, "err", err)
			return err
		}
	}
	return nil
}

func (impl *GlobalPluginServiceImpl) CreateScriptPathArgPortMappingForPluginInlineStep(scriptPathArgPortMappings []*bean2.ScriptPathArgPortMapping, pluginPipelineScriptId int, userId int32, tx *pg.Tx) error {
	//fetch previous ScriptPathArgPortMapping by pluginPipelineScriptId and mark previous as deleted before creating new mappings
	dbScriptPathArgPortMappings, err := impl.pipelineStageRepository.GetScriptMappingDetailByScriptId(pluginPipelineScriptId)
	if err != nil {
		impl.logger.Errorw("error in getting plugin step script", "err", err)
		return err
	}
	for _, scriptPathArgPortMapping := range dbScriptPathArgPortMappings {
		scriptPathArgPortMapping.Deleted = true
		scriptPathArgPortMapping.UpdatedBy = userId
		scriptPathArgPortMapping.UpdatedOn = time.Now()
	}
	if len(dbScriptPathArgPortMappings) > 0 {
		err = impl.pipelineStageRepository.UpdateScriptMapping(dbScriptPathArgPortMappings, tx)
		if err != nil {
			impl.logger.Errorw("error in updating previous plugin script path arg port mapping by script id", "scriptId", pluginPipelineScriptId, "err", err)
			return err
		}
	}
	var scriptMap []repository2.ScriptPathArgPortMapping
	for _, scriptPathArgPortMapping := range scriptPathArgPortMappings {
		scriptPathArgPortMapping.ScriptId = pluginPipelineScriptId

		if len(scriptPathArgPortMapping.FilePathOnDisk) > 0 && len(scriptPathArgPortMapping.FilePathOnContainer) > 0 {
			repositoryEntry := repository2.ScriptPathArgPortMapping{
				TypeOfMapping:       repository.SCRIPT_MAPPING_TYPE_FILE_PATH,
				FilePathOnDisk:      scriptPathArgPortMapping.FilePathOnDisk,
				FilePathOnContainer: scriptPathArgPortMapping.FilePathOnContainer,
				ScriptId:            pluginPipelineScriptId,
				Deleted:             false,
				AuditLog:            sql.NewDefaultAuditLog(userId),
			}
			scriptMap = append(scriptMap, repositoryEntry)
		}
		if len(scriptPathArgPortMapping.Command) > 0 || len(scriptPathArgPortMapping.Args) > 0 {
			repositoryEntry := repository2.ScriptPathArgPortMapping{
				TypeOfMapping: repository.SCRIPT_MAPPING_TYPE_DOCKER_ARG,
				Command:       scriptPathArgPortMapping.Command,
				Args:          scriptPathArgPortMapping.Args,
				ScriptId:      pluginPipelineScriptId,
				Deleted:       false,
				AuditLog:      sql.NewDefaultAuditLog(userId),
			}
			scriptMap = append(scriptMap, repositoryEntry)
		}
		if scriptPathArgPortMapping.PortOnContainer > 0 && scriptPathArgPortMapping.PortOnLocal > 0 {
			repositoryEntry := repository2.ScriptPathArgPortMapping{
				TypeOfMapping:   repository.SCRIPT_MAPPING_TYPE_PORT,
				PortOnLocal:     scriptPathArgPortMapping.PortOnLocal,
				PortOnContainer: scriptPathArgPortMapping.PortOnContainer,
				ScriptId:        pluginPipelineScriptId,
				Deleted:         false,
				AuditLog:        sql.NewDefaultAuditLog(userId),
			}
			scriptMap = append(scriptMap, repositoryEntry)
		}
	}

	if len(scriptMap) > 0 {
		err := impl.pipelineStageRepository.CreateScriptMapping(scriptMap, tx)
		if err != nil {
			impl.logger.Errorw("error in creating script mappings", "err", err, "scriptMappings", scriptMap)
			return err
		}
	}
	return nil
}

func (impl *GlobalPluginServiceImpl) UpdatePluginPipelineScript(dbPluginPipelineScript *repository.PluginPipelineScript, pluginPipelineScriptReq *bean2.PluginPipelineScript, userId int32, tx *pg.Tx) error {
	dbPluginPipelineScript.Script = pluginPipelineScriptReq.Script
	dbPluginPipelineScript.StoreScriptAt = pluginPipelineScriptReq.StoreScriptAt
	dbPluginPipelineScript.Type = pluginPipelineScriptReq.Type
	dbPluginPipelineScript.DockerfileExists = pluginPipelineScriptReq.DockerfileExists
	dbPluginPipelineScript.MountPath = pluginPipelineScriptReq.MountPath
	dbPluginPipelineScript.MountCodeToContainer = pluginPipelineScriptReq.MountCodeToContainer
	dbPluginPipelineScript.MountCodeToContainerPath = pluginPipelineScriptReq.MountCodeToContainerPath
	dbPluginPipelineScript.MountDirectoryFromHost = pluginPipelineScriptReq.MountDirectoryFromHost
	dbPluginPipelineScript.ContainerImagePath = pluginPipelineScriptReq.ContainerImagePath
	dbPluginPipelineScript.ImagePullSecretType = pluginPipelineScriptReq.ImagePullSecretType
	dbPluginPipelineScript.ImagePullSecret = pluginPipelineScriptReq.ImagePullSecret
	dbPluginPipelineScript.UpdatedBy = userId
	dbPluginPipelineScript.UpdatedOn = time.Now()

	err := impl.globalPluginRepository.UpdatePluginPipelineScript(dbPluginPipelineScript, tx)
	if err != nil {
		impl.logger.Errorw("error in updating plugin step script", "err", err)
		return err
	}

	return nil
}

func (impl *GlobalPluginServiceImpl) saveDeepPluginStepData(pluginMetadataId int, pluginStepsReq []*bean2.PluginStepsDto, userId int32, tx *pg.Tx) error {
	for _, pluginStep := range pluginStepsReq {
		pluginStepData := &repository.PluginStep{
			PluginId:            pluginMetadataId,
			Name:                pluginStep.Name,
			Description:         pluginStep.Description,
			Index:               pluginStep.Index,
			StepType:            pluginStep.StepType,
			RefPluginId:         pluginStep.RefPluginId,
			OutputDirectoryPath: pluginStep.OutputDirectoryPath,
			DependentOnStep:     pluginStep.DependentOnStep,
			AuditLog:            sql.NewDefaultAuditLog(userId),
		}
		//get the script saved for this plugin step
		if pluginStep.PluginPipelineScript != nil {
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
				AuditLog:                 sql.NewDefaultAuditLog(userId),
			}
			pluginPipelineScript, err := impl.globalPluginRepository.SavePluginPipelineScript(pluginPipelineScript, tx)
			if err != nil {
				impl.logger.Errorw("error in saving plugin pipeline script", "pluginPipelineScript", pluginPipelineScript, "err", err)
				return err
			}
			err = impl.CreateScriptPathArgPortMappingForPluginInlineStep(pluginStep.PluginPipelineScript.PathArgPortMapping, pluginPipelineScript.Id, userId, tx)
			if err != nil {
				impl.logger.Errorw("error in CreateScriptPathArgPortMappingForPluginInlineStep", "pluginMetadataId", pluginMetadataId, "err", err)
				return err
			}
			pluginStep.PluginPipelineScript.Id = pluginPipelineScript.Id
			pluginStepData.ScriptId = pluginPipelineScript.Id
		}

		pluginStepData, err := impl.globalPluginRepository.SavePluginSteps(pluginStepData, tx)
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
				VariableType:              pluginStepVariable.VariableType,
				ValueType:                 pluginStepVariable.ValueType,
				PreviousStepIndex:         pluginStepVariable.PreviousStepIndex,
				VariableStepIndex:         pluginStepVariable.VariableStepIndex,
				VariableStepIndexInPlugin: pluginStepVariable.VariableStepIndexInPlugin,
				ReferenceVariableName:     pluginStepVariable.ReferenceVariableName,
				AuditLog:                  sql.NewDefaultAuditLog(userId),
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
					AuditLog:            sql.NewDefaultAuditLog(userId),
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

func (impl *GlobalPluginServiceImpl) updatePlugin(pluginUpdateReq *bean2.PluginMetadataDto, userId int32) (*bean2.PluginMetadataDto, error) {
	if len(pluginUpdateReq.Type) == 0 {
		return nil, errors.New("invalid plugin type, should be of the type PRESET or SHARED")
	}

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
	//update entry in plugin_ metadata
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
	pluginStageMapping, err := impl.globalPluginRepository.GetPluginStageMappingByPluginId(pluginUpdateReq.Id)
	if err != nil {
		impl.logger.Errorw("updatePlugin, error in getting pluginStageMapping", "pluginId", pluginUpdateReq.Id, "err", err)
		return nil, err
	}
	pluginStage := repository.CI_CD
	if pluginUpdateReq.PluginStage == bean2.CI_TYPE_PLUGIN {
		pluginStage = repository.CI
	} else if pluginUpdateReq.PluginStage == bean2.CD_TYPE_PLUGIN {
		pluginStage = repository.CD
	}
	pluginStageMapping.StageType = pluginStage
	pluginStageMapping.UpdatedBy = userId
	pluginStageMapping.UpdatedOn = time.Now()

	err = impl.globalPluginRepository.UpdatePluginStageMapping(pluginStageMapping, tx)
	if err != nil {
		impl.logger.Errorw("error in updating plugin stage mapping", "pluginId", pluginUpdateReq.Id, "err", err)
		return nil, err
	}

	pluginSteps, err := impl.globalPluginRepository.GetPluginStepsByPluginId(pluginUpdateReq.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("updatePlugin, error in getting pluginSteps", "pluginId", pluginUpdateReq.Id, "err", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		impl.logger.Infow("updatePlugin,no plugin steps found for this plugin", "pluginId", pluginUpdateReq.Id, "err", err)
	}
	pluginStepVariables, err := impl.globalPluginRepository.GetExposedVariablesByPluginId(pluginUpdateReq.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("updatePlugin, error in getting pluginStepVariables", "pluginId", pluginUpdateReq.Id, "err", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		impl.logger.Infow("updatePlugin,no plugin step variables found for this plugin step", "pluginId", pluginUpdateReq.Id, "err", err)
	}
	pluginStepConditions, err := impl.globalPluginRepository.GetConditionsByPluginId(pluginUpdateReq.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("updatePlugin, error in getting pluginStepConditions", "pluginId", pluginUpdateReq.Id, "err", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		impl.logger.Infow("updatePlugin,no plugin step variable conditions found for this plugin", "pluginId", pluginUpdateReq.Id, "err", err)
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
		//update here with deleted as true
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
	isUpdateReq := true
	err = impl.CreateNewPluginTagsAndRelationsIfRequired(pluginUpdateReq, isUpdateReq, userId, tx)
	if err != nil {
		impl.logger.Errorw("updatePlugin, error in CreateNewPluginTagsAndRelationsIfRequired", "err", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("updatePlugin, error in committing db transaction", "err", err)
		return nil, err
	}
	return pluginUpdateReq, nil
}

func (impl *GlobalPluginServiceImpl) updateDeepPluginStepData(pluginStepsToUpdate []*bean2.PluginStepsDto, pluginStepVariables []*repository.PluginStepVariable,
	pluginStepConditions []*repository.PluginStepCondition, pluginSteps []*repository.PluginStep, userId int32, tx *pg.Tx) error {

	pluginStepIdsToStepDtoMapping := make(map[int]*bean2.PluginStepsDto)
	for _, pluginStepUpdateReq := range pluginStepsToUpdate {
		pluginStepIdsToStepDtoMapping[pluginStepUpdateReq.Id] = pluginStepUpdateReq
	}
	for _, pluginStep := range pluginSteps {
		if _, ok := pluginStepIdsToStepDtoMapping[pluginStep.Id]; ok {
			pluginStep.Name = pluginStepIdsToStepDtoMapping[pluginStep.Id].Name
			pluginStep.Description = pluginStepIdsToStepDtoMapping[pluginStep.Id].Description
			pluginStep.Index = pluginStepIdsToStepDtoMapping[pluginStep.Id].Index
			pluginStep.StepType = pluginStepIdsToStepDtoMapping[pluginStep.Id].StepType
			pluginStep.RefPluginId = pluginStepIdsToStepDtoMapping[pluginStep.Id].RefPluginId
			pluginStep.OutputDirectoryPath = pluginStepIdsToStepDtoMapping[pluginStep.Id].OutputDirectoryPath
			pluginStep.DependentOnStep = pluginStepIdsToStepDtoMapping[pluginStep.Id].DependentOnStep
			pluginStep.UpdatedBy = userId
			pluginStep.UpdatedOn = time.Now()

			err := impl.globalPluginRepository.UpdatePluginSteps(pluginStep, tx)
			if err != nil {
				impl.logger.Errorw("error in updating plugin steps", "pluginMetadataId", pluginStep.PluginId, "pluginStepId", pluginStep.Id, "err", err)
				return err
			}

			pluginStepScript, err := impl.globalPluginRepository.GetScriptDetailById(pluginStep.ScriptId)
			if err != nil {
				impl.logger.Errorw("error in getting plugin step script", "scriptId", pluginStep.ScriptId, "pluginStepId", pluginStep.Id, "err", err)
				return err
			}
			err = impl.UpdatePluginPipelineScript(pluginStepScript, pluginStepIdsToStepDtoMapping[pluginStep.Id].PluginPipelineScript, userId, tx)
			if err != nil {
				impl.logger.Errorw("error in updating plugin step script and script args and port mappings", "scriptId", pluginStep.ScriptId, "pluginStepId", pluginStep.Id, "err", err)
				return err
			}
		}
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
		//update ScriptPathArgPortMapping in db
		err := impl.CreateScriptPathArgPortMappingForPluginInlineStep(pluginStepUpdateReq.PluginPipelineScript.PathArgPortMapping, pluginStepUpdateReq.PluginPipelineScript.Id, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in CreateScriptPathArgPortMappingForPluginInlineStep", "pluginMetadataId", pluginStepUpdateReq.PluginPipelineScript.Id, "err", err)
			return err
		}
	}

	return nil
}

func (impl *GlobalPluginServiceImpl) updateDeepPluginStepVariableData(pluginStepId int, pluginStepVariablesToUpdate []*bean2.PluginVariableDto,
	pluginStepVariables []*repository.PluginStepVariable, pluginStepConditions []*repository.PluginStepCondition, userId int32, tx *pg.Tx) error {

	stepVariableIdsToStepVariableMapping := make(map[int]*bean2.PluginVariableDto)
	for _, stepVariable := range pluginStepVariablesToUpdate {
		stepVariableIdsToStepVariableMapping[stepVariable.Id] = stepVariable
	}
	for _, dbStepVariable := range pluginStepVariables {
		if _, ok := stepVariableIdsToStepVariableMapping[dbStepVariable.Id]; ok {

			dbStepVariable.Name = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].Name
			dbStepVariable.Format = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].Format
			dbStepVariable.Description = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].Description
			dbStepVariable.IsExposed = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].IsExposed
			dbStepVariable.AllowEmptyValue = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].AllowEmptyValue
			dbStepVariable.DefaultValue = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].DefaultValue
			dbStepVariable.Value = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].Value
			dbStepVariable.VariableType = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].VariableType
			dbStepVariable.ValueType = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].ValueType
			dbStepVariable.PreviousStepIndex = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].PreviousStepIndex
			dbStepVariable.VariableStepIndex = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].VariableStepIndex
			dbStepVariable.VariableStepIndexInPlugin = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].VariableStepIndexInPlugin
			dbStepVariable.ReferenceVariableName = stepVariableIdsToStepVariableMapping[dbStepVariable.Id].ReferenceVariableName
			dbStepVariable.UpdatedBy = userId
			dbStepVariable.UpdatedOn = time.Now()

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
			AuditLog:            sql.NewDefaultAuditLog(userId),
		}
		pluginStepConditionData, err := impl.globalPluginRepository.SavePluginStepConditions(pluginStepConditionData, tx)
		if err != nil {
			impl.logger.Errorw("saveDeepStepVariableConditionsData, error in saving plugin step condition", "pluginStepId", pluginStepVariableId, "err", err)
			return err
		}
		pluginStepCondition.Id = pluginStepConditionData.Id
	}
	return nil
}

func (impl *GlobalPluginServiceImpl) deleteDeepStepVariableConditionsData(stepVariableConditionsToDelete []*repository.PluginStepCondition, pluginStepConditions []*repository.PluginStepCondition, userId int32, tx *pg.Tx) error {

	stepVariableConditionsToDeleteIdsMapping := make(map[int]bool)
	for _, stepVariableConditionRemoveReq := range stepVariableConditionsToDelete {
		stepVariableConditionsToDeleteIdsMapping[stepVariableConditionRemoveReq.Id] = true
	}

	for _, stepVariableCondition := range pluginStepConditions {
		if _, ok := stepVariableConditionsToDeleteIdsMapping[stepVariableCondition.Id]; ok {
			stepVariableCondition.Deleted = true
			stepVariableCondition.UpdatedOn = time.Now()
			stepVariableCondition.UpdatedBy = userId
		}
	}
	err := impl.globalPluginRepository.UpdateInBulkPluginStepConditions(pluginStepConditions, tx)
	if err != nil {
		impl.logger.Errorw("deleteDeepStepVariableConditionsData, error in updating plugin step conditions in bulk", "err", err)
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
	for _, stepVariableCondition := range pluginStepConditions {
		if _, ok := stepVariableConditionsToDeleteIdsMapping[stepVariableCondition.Id]; ok {
			stepVariableCondition.ConditionType = stepVariableConditionsToDeleteIdsMapping[stepVariableCondition.Id].ConditionType
			stepVariableCondition.ConditionalOperator = stepVariableConditionsToDeleteIdsMapping[stepVariableCondition.Id].ConditionalOperator
			stepVariableCondition.ConditionalValue = stepVariableConditionsToDeleteIdsMapping[stepVariableCondition.Id].ConditionalValue
			stepVariableCondition.UpdatedOn = time.Now()
			stepVariableCondition.UpdatedBy = userId
		}
	}
	err := impl.globalPluginRepository.UpdateInBulkPluginStepConditions(pluginStepConditions, tx)
	if err != nil {
		impl.logger.Errorw("updateDeepStepVariableConditionsData, error in updating plugin step conditions in bulk", "err", err)
		return err
	}
	return nil
}

func filterPluginStepVariableConditions(stepVariableId int, pluginStepConditionsInDb []*repository.PluginStepCondition, pluginStepConditionReq []*bean2.PluginStepCondition, userId int32) ([]*repository.PluginStepCondition, []*repository.PluginStepCondition, []*repository.PluginStepCondition) {
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
			stepVariableCondition := getStepVariableConditionDbObject(stepVariableConditionReq)

			if _, ok := stepVariableConditionMapping[stepVariableConditionReq.Id]; !ok {
				newStepVariableConditionsToCreate = append(newStepVariableConditionsToCreate, stepVariableCondition)
			} else {
				stepVariableConditionsToUpdate = append(stepVariableConditionsToUpdate, stepVariableCondition)
			}
		}
	} else if len(pluginStepConditionReq) < len(stepIdToDbStepVariableConditionsMapping[stepVariableId]) {
		//it means there are deleted variable conditions in update request for a particular variable, filter out plugin variable conditions to delete
		stepVariableConditionMapping := make(map[int]*repository.PluginStepCondition)
		for _, variableConditionReq := range pluginStepConditionReq {
			stepVariableCondition := getStepVariableConditionDbObject(variableConditionReq)
			stepVariableConditionMapping[variableConditionReq.Id] = stepVariableCondition
		}

		for _, existingStepVariableCondition := range stepIdToDbStepVariableConditionsMapping[stepVariableId] {
			if _, ok := stepVariableConditionMapping[existingStepVariableCondition.Id]; !ok {
				stepVariableConditionsToRemove = append(stepVariableConditionsToRemove, existingStepVariableCondition)
			} else {
				stepVariableConditionsToUpdate = append(stepVariableConditionsToUpdate, stepVariableConditionMapping[existingStepVariableCondition.Id])
			}
		}
	} else {
		pluginStepConditionDbObject := make([]*repository.PluginStepCondition, 0)
		for _, variableCondition := range pluginStepConditionReq {
			stepVariableCondition := getStepVariableConditionDbObject(variableCondition)
			pluginStepConditionDbObject = append(pluginStepConditionDbObject, stepVariableCondition)
		}
		return nil, nil, pluginStepConditionDbObject
	}

	return newStepVariableConditionsToCreate, stepVariableConditionsToRemove, stepVariableConditionsToUpdate
}
func getStepVariableConditionDbObject(stepVariableConditionReq *bean2.PluginStepCondition) *repository.PluginStepCondition {
	stepVariableCondition := &repository.PluginStepCondition{
		Id:                  stepVariableConditionReq.Id,
		PluginStepId:        stepVariableConditionReq.PluginStepId,
		ConditionVariableId: stepVariableConditionReq.ConditionVariableId,
		ConditionType:       stepVariableConditionReq.ConditionType,
		ConditionalOperator: stepVariableConditionReq.ConditionalOperator,
		ConditionalValue:    stepVariableConditionReq.ConditionalValue,
		Deleted:             stepVariableConditionReq.Deleted,
	}
	return stepVariableCondition
}
func filterPluginStepVariable(pluginStepId int, existingPluginStepVariables []*repository.PluginStepVariable,
	pluginStepVariableUpdateReq []*bean2.PluginVariableDto, userId int32) ([]*bean2.PluginVariableDto, []*bean2.PluginVariableDto, []*bean2.PluginVariableDto) {

	newPluginStepVariablesToCreate := make([]*bean2.PluginVariableDto, 0)
	pluginStepVariablesToRemove := make([]*bean2.PluginVariableDto, 0)
	pluginStepVariablesToUpdate := make([]*bean2.PluginVariableDto, 0)

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
		pluginStepVariableIdToStepVariableMapping := make(map[int]*bean2.PluginVariableDto)
		for _, stepVariableUpdateReq := range pluginStepVariableUpdateReq {
			pluginStepVariableIdToStepVariableMapping[stepVariableUpdateReq.Id] = stepVariableUpdateReq
		}

		for _, existingStepVariable := range stepIdToDbStepVariablesMapping[pluginStepId] {
			if _, ok := pluginStepVariableIdToStepVariableMapping[existingStepVariable.Id]; !ok {
				pluginStepVariablesToRemove = append(pluginStepVariablesToRemove, &bean2.PluginVariableDto{Id: existingStepVariable.Id})
			} else {
				pluginStepVariablesToUpdate = append(pluginStepVariablesToUpdate, pluginStepVariableIdToStepVariableMapping[existingStepVariable.Id])
			}
		}
	} else {
		return nil, nil, pluginStepVariableUpdateReq
	}

	return newPluginStepVariablesToCreate, pluginStepVariablesToRemove, pluginStepVariablesToUpdate
}

func (impl *GlobalPluginServiceImpl) saveDeepPluginStepVariableData(pluginStepId int, pluginStepVariablesToCreate []*bean2.PluginVariableDto, userId int32, tx *pg.Tx) error {
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
			AuditLog:                  sql.NewDefaultAuditLog(userId),
		}
		pluginStepVariableData, err := impl.globalPluginRepository.SavePluginStepVariables(pluginStepVariableData, tx)
		if err != nil {
			impl.logger.Errorw("saveDeepPluginStepVariableData, error in saving plugin step variable", "pluginStepVariableData", pluginStepVariableData, "err", err)
			return err
		}
		pluginStepVariable.Id = pluginStepVariableData.Id
		for _, pluginStepCondition := range pluginStepVariable.PluginStepCondition {
			pluginStepConditionData := &repository.PluginStepCondition{
				PluginStepId:        pluginStepId,
				ConditionVariableId: pluginStepVariableData.Id,
				ConditionType:       pluginStepCondition.ConditionType,
				ConditionalOperator: pluginStepCondition.ConditionalOperator,
				ConditionalValue:    pluginStepCondition.ConditionalValue,
				AuditLog:            sql.NewDefaultAuditLog(userId),
			}
			pluginStepConditionData, err = impl.globalPluginRepository.SavePluginStepConditions(pluginStepConditionData, tx)
			if err != nil {
				impl.logger.Errorw("saveDeepPluginStepVariableData, error in saving plugin step condition", "pluginStepId", pluginStepId, "err", err)
				return err
			}
			pluginStepCondition.Id = pluginStepConditionData.Id
		}
	}
	return nil
}

func (impl *GlobalPluginServiceImpl) deleteDeepPluginStepVariableData(pluginStepVariablesToDelete []*bean2.PluginVariableDto,
	pluginStepVariables []*repository.PluginStepVariable, pluginStepConditions []*repository.PluginStepCondition, userId int32, tx *pg.Tx) error {

	stepVariablesToDeleteIdsMapping := make(map[int]bool)
	for _, stepVariableRemoveReq := range pluginStepVariablesToDelete {
		stepVariablesToDeleteIdsMapping[stepVariableRemoveReq.Id] = true
	}
	for _, stepVariable := range pluginStepVariables {
		if _, ok := stepVariablesToDeleteIdsMapping[stepVariable.Id]; ok {
			stepVariable.Deleted = true
			stepVariable.UpdatedOn = time.Now()
			stepVariable.UpdatedBy = userId

			err := impl.globalPluginRepository.UpdatePluginStepVariables(stepVariable, tx)
			if err != nil {
				impl.logger.Errorw("deleteDeepPluginStepData, error in deleting plugin step variables", "stepVariableId", stepVariable.Id, "err", err)
				return err
			}
		}
	}
	for _, stepVariableCondition := range pluginStepConditions {
		if _, ok := stepVariablesToDeleteIdsMapping[stepVariableCondition.ConditionVariableId]; ok {
			stepVariableCondition.Deleted = true
			stepVariableCondition.UpdatedOn = time.Now()
			stepVariableCondition.UpdatedBy = userId

			err := impl.globalPluginRepository.UpdatePluginStepConditions(stepVariableCondition, tx)
			if err != nil {
				impl.logger.Errorw("deleteDeepPluginStepData, error in deleting plugin step conditions", "stepVariableConditionId", stepVariableCondition.Id, "err", err)
				return err
			}
		}
	}
	return nil
}

func (impl *GlobalPluginServiceImpl) deleteDeepPluginStepData(pluginStepsToRemove []*bean2.PluginStepsDto, pluginStepVariables []*repository.PluginStepVariable,
	pluginStepConditions []*repository.PluginStepCondition, pluginSteps []*repository.PluginStep, userId int32, tx *pg.Tx) error {
	pluginStepsToRemoveIdsMapping := make(map[int]bool)
	for _, pluginStepRemoveReq := range pluginStepsToRemove {
		pluginStepsToRemoveIdsMapping[pluginStepRemoveReq.Id] = true
	}
	for _, pluginStep := range pluginSteps {
		if _, ok := pluginStepsToRemoveIdsMapping[pluginStep.Id]; ok {
			pluginStep.Deleted = true
			pluginStep.UpdatedBy = userId
			pluginStep.UpdatedOn = time.Now()

			err := impl.globalPluginRepository.UpdatePluginSteps(pluginStep, tx)
			if err != nil {
				impl.logger.Errorw("deleteDeepPluginStepData, error in deleting plugin steps", "pluginStepId", pluginStep.Id, "err", err)
				return err
			}

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
			scriptPathArgPortMappings, err := impl.pipelineStageRepository.GetScriptMappingDetailByScriptId(pluginStep.ScriptId)
			if err != nil {
				impl.logger.Errorw("error in getting plugin step script", "err", err)
				return err
			}
			for _, scriptPathArgPortMapping := range scriptPathArgPortMappings {
				scriptPathArgPortMapping.Deleted = true
				scriptPathArgPortMapping.UpdatedBy = userId
				scriptPathArgPortMapping.UpdatedOn = time.Now()
			}
			err = impl.pipelineStageRepository.UpdateScriptMapping(scriptPathArgPortMappings, tx)
			if err != nil {
				impl.logger.Errorw("error in updating plugin script path arg port mapping", "err", err)
				return err
			}
		}
	}
	for _, pluginStepVariable := range pluginStepVariables {
		if _, ok := pluginStepsToRemoveIdsMapping[pluginStepVariable.PluginStepId]; ok {
			pluginStepVariable.Deleted = true
			pluginStepVariable.UpdatedOn = time.Now()
			pluginStepVariable.UpdatedBy = userId
		}
	}
	if len(pluginStepVariables) > 0 {
		err := impl.globalPluginRepository.UpdateInBulkPluginStepVariables(pluginStepVariables, tx)
		if err != nil {
			impl.logger.Errorw("deleteDeepPluginStepData,error in updating plugin step variables in bulk", "err", err)
			return err
		}
	}

	for _, pluginStepCondition := range pluginStepConditions {
		if _, ok := pluginStepsToRemoveIdsMapping[pluginStepCondition.PluginStepId]; ok {
			pluginStepCondition.Deleted = true
			pluginStepCondition.UpdatedOn = time.Now()
			pluginStepCondition.UpdatedBy = userId
		}
	}
	if len(pluginStepConditions) > 0 {
		err := impl.globalPluginRepository.UpdateInBulkPluginStepConditions(pluginStepConditions, tx)
		if err != nil {
			impl.logger.Errorw("deleteDeepPluginStepData, error in updating plugin step conditions in bulk", "err", err)
			return err
		}
	}
	return nil
}

func filterPluginStepData(existingPluginStepsInDb []*repository.PluginStep, pluginStepUpdateReq []*bean2.PluginStepsDto) ([]*bean2.PluginStepsDto, []*bean2.PluginStepsDto, []*bean2.PluginStepsDto) {
	newPluginStepsToCreate := make([]*bean2.PluginStepsDto, 0)
	pluginStepsToRemove := make([]*bean2.PluginStepsDto, 0)
	pluginStepsToUpdate := make([]*bean2.PluginStepsDto, 0)

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
		pluginIdMapping := make(map[int]*bean2.PluginStepsDto)
		for _, pluginStepReq := range pluginStepUpdateReq {
			pluginIdMapping[pluginStepReq.Id] = pluginStepReq
		}
		for _, existingPluginStep := range existingPluginStepsInDb {
			if _, ok := pluginIdMapping[existingPluginStep.Id]; !ok {
				pluginStepsToRemove = append(pluginStepsToRemove, &bean2.PluginStepsDto{Id: existingPluginStep.Id})
			} else {
				pluginStepsToUpdate = append(pluginStepsToUpdate, pluginIdMapping[existingPluginStep.Id])
			}
		}
	} else {
		return nil, nil, pluginStepUpdateReq
	}

	return newPluginStepsToCreate, pluginStepsToRemove, pluginStepsToUpdate
}

func (impl *GlobalPluginServiceImpl) GetAllDetailedPluginInfo() ([]*bean2.PluginMetadataDto, error) {
	allPlugins, err := impl.globalPluginRepository.GetAllPluginMetaData()
	if err != nil {
		impl.logger.Errorw("GetAllDetailedPluginInfo, error in getting all pluginsMetadata", "err", err)
		return nil, err
	}
	allPluginMetadata := make([]*bean2.PluginMetadataDto, 0, len(allPlugins))
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

func (impl *GlobalPluginServiceImpl) GetDetailedPluginInfoByPluginId(pluginId int) (*bean2.PluginMetadataDto, error) {

	pluginMetaData, err := impl.globalPluginRepository.GetMetaDataByPluginId(pluginId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GetDetailedPluginInfoByPluginId, error in getting pluginMetadata", "pluginId", pluginId, "err", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return nil, errors.New("no plugin found for this id")
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
	pluginStage := bean2.CI_CD_TYPE_PLUGIN
	if pluginStageMapping.StageType == repository.CI {
		pluginStage = bean2.CI_TYPE_PLUGIN
	} else if pluginStageMapping.StageType == repository.CD {
		pluginStage = bean2.CD_TYPE_PLUGIN
	}
	pluginIdTagsMap, err := impl.getPluginIdTagsMap()
	if err != nil {
		impl.logger.Errorw("GetDetailedPluginInfoByPluginId, error in getting pluginIdTagsMap", "pluginId", pluginId, "err", err)
		return nil, err
	}
	pluginMetadataResponse := &bean2.PluginMetadataDto{
		Id:          pluginMetaData.Id,
		Name:        pluginMetaData.Name,
		Description: pluginMetaData.Description,
		Type:        string(pluginMetaData.Type),
		Icon:        pluginMetaData.Icon,
		Tags:        pluginIdTagsMap[pluginMetaData.Id],
		PluginStage: pluginStage,
	}

	pluginStepsResp := make([]*bean2.PluginStepsDto, 0)
	scriptPathArgPortMapping := make([]*bean2.ScriptPathArgPortMapping, 0)
	for _, pluginStep := range pluginSteps {
		pluginScript, err := impl.globalPluginRepository.GetScriptDetailById(pluginStep.ScriptId)
		if err != nil {
			impl.logger.Errorw("GetDetailedPluginInfoByPluginId, error in getting pluginScript", "pluginScriptId", pluginStep.ScriptId, "pluginId", pluginId, "err", err)
			return nil, err
		}
		pluginScriptDto := &bean2.PluginPipelineScript{
			Id:                       pluginScript.Id,
			Script:                   pluginScript.Script,
			StoreScriptAt:            pluginScript.StoreScriptAt,
			Type:                     pluginScript.Type,
			DockerfileExists:         pluginScript.DockerfileExists,
			MountPath:                pluginScript.MountPath,
			MountCodeToContainer:     pluginScript.MountCodeToContainer,
			MountCodeToContainerPath: pluginScript.MountCodeToContainerPath,
			MountDirectoryFromHost:   pluginScript.MountDirectoryFromHost,
			ContainerImagePath:       pluginScript.ContainerImagePath,
			ImagePullSecretType:      pluginScript.ImagePullSecretType,
			ImagePullSecret:          pluginScript.ImagePullSecret,
			Deleted:                  pluginScript.Deleted,
		}
		//fetch ScriptPathArgPortMapping for each plugin step
		scriptPathArgPortMappings, err := impl.pipelineStageRepository.GetScriptMappingDetailByScriptId(pluginStep.ScriptId)
		if err != nil {
			impl.logger.Errorw("error in getting scriptPathArgPortMappings", "err", err)
			return nil, err
		}
		for _, scriptMapping := range scriptPathArgPortMappings {
			mapping := &bean2.ScriptPathArgPortMapping{
				Id:                  scriptMapping.Id,
				TypeOfMapping:       scriptMapping.TypeOfMapping,
				FilePathOnDisk:      scriptMapping.FilePathOnDisk,
				FilePathOnContainer: scriptMapping.FilePathOnContainer,
				Command:             scriptMapping.Command,
				Args:                scriptMapping.Args,
				PortOnLocal:         scriptMapping.PortOnLocal,
				PortOnContainer:     scriptMapping.PortOnContainer,
				ScriptId:            scriptMapping.ScriptId,
			}
			scriptPathArgPortMapping = append(scriptPathArgPortMapping, mapping)
		}
		pluginScriptDto.PathArgPortMapping = scriptPathArgPortMapping

		pluginStepDto := &bean2.PluginStepsDto{
			Id:                   pluginStep.Id,
			Name:                 pluginStep.Name,
			Description:          pluginStep.Description,
			Index:                pluginStep.Index,
			StepType:             pluginStep.StepType,
			RefPluginId:          pluginStep.RefPluginId,
			OutputDirectoryPath:  pluginStep.OutputDirectoryPath,
			DependentOnStep:      pluginStep.DependentOnStep,
			PluginPipelineScript: pluginScriptDto,
		}
		pluginStepVariableResp := make([]*bean2.PluginVariableDto, 0, len(pluginStepVariables))
		for _, pluginStepVariable := range pluginStepVariables {
			if pluginStepVariable.PluginStepId == pluginStep.Id {
				pluginStepConditionDto := make([]*bean2.PluginStepCondition, 0, len(pluginStepConditions))
				for _, pluginStepCondition := range pluginStepConditions {
					if pluginStepCondition.ConditionVariableId == pluginStepVariable.Id {
						pluginStepConditionDto = append(pluginStepConditionDto, &bean2.PluginStepCondition{
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
				pluginStepVariableResp = append(pluginStepVariableResp, &bean2.PluginVariableDto{
					Id:                        pluginStepVariable.Id,
					Name:                      pluginStepVariable.Name,
					Format:                    pluginStepVariable.Format,
					Description:               pluginStepVariable.Description,
					IsExposed:                 pluginStepVariable.IsExposed,
					AllowEmptyValue:           pluginStepVariable.AllowEmptyValue,
					DefaultValue:              pluginStepVariable.DefaultValue,
					Value:                     pluginStepVariable.Value,
					VariableType:              pluginStepVariable.VariableType,
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

func (impl *GlobalPluginServiceImpl) deletePlugin(pluginDeleteReq *bean2.PluginMetadataDto, userId int32) (*bean2.PluginMetadataDto, error) {
	dbConnection := impl.globalPluginRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error
	defer tx.Rollback()
	//check if this plugin is being used in some ci or cd pipeline, if yes then  return with error
	pipelineStageStep, err := impl.pipelineStageRepository.GetActiveStepsByRefPluginId(pluginDeleteReq.Id)
	if err != nil {
		impl.logger.Errorw("deletePlugin, error in getting all pluginStageSteps where this plugin is being used", "pluginId", pluginDeleteReq.Id, "err", err)
		return nil, err
	}
	if len(pipelineStageStep) > 0 {
		return nil, errors.New("this plugin is being used in multiple pre or post ci or cd pipelines, please remove them before deleting this plugin")
	}
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

		err := impl.globalPluginRepository.UpdatePluginSteps(pluginStep, tx)
		if err != nil {
			impl.logger.Errorw("deletePlugin, error in deleting plugin steps", "pluginId", pluginMetaData.Id, "err", err)
			return nil, err
		}
	}

	pluginStepVariables, err := impl.globalPluginRepository.GetExposedVariablesByPluginId(pluginDeleteReq.Id)
	if err != nil {
		impl.logger.Errorw("deletePlugin, error in getting pluginStepVariables", "pluginId", pluginDeleteReq.Id, "err", err)
		return nil, err
	}
	for _, pluginStepVariable := range pluginStepVariables {
		pluginStepVariable.Deleted = true
		pluginStepVariable.UpdatedBy = userId
		pluginStepVariable.UpdatedOn = time.Now()

		err = impl.globalPluginRepository.UpdatePluginStepVariables(pluginStepVariable, tx)
		if err != nil {
			impl.logger.Errorw("deletePlugin, error in deleting plugin step variables", "pluginId", pluginMetaData.Id, "err", err)
			return nil, err
		}
	}

	pluginStepConditions, err := impl.globalPluginRepository.GetConditionsByPluginId(pluginDeleteReq.Id)
	if err != nil {
		impl.logger.Errorw("deletePlugin, error in getting pluginStepConditions", "pluginId", pluginDeleteReq.Id, "err", err)
		return nil, err
	}

	for _, pluginStepCondition := range pluginStepConditions {
		pluginStepCondition.Deleted = true
		pluginStepCondition.UpdatedBy = userId
		pluginStepCondition.UpdatedOn = time.Now()

		err = impl.globalPluginRepository.UpdatePluginStepConditions(pluginStepCondition, tx)
		if err != nil {
			impl.logger.Errorw("deletePlugin, error in deleting plugin step variable conditions", "pluginId", pluginMetaData.Id, "err", err)
			return nil, err
		}
	}

	//delete entry for ScriptPathArgPortMappings in db
	for _, pluginStepDeleteReq := range pluginDeleteReq.PluginSteps {
		scriptPathArgPortMappings, err := impl.pipelineStageRepository.GetScriptMappingDetailByScriptId(pluginStepDeleteReq.PluginPipelineScript.Id)
		if err != nil {
			impl.logger.Errorw("error in getting script path arg port mappings", "err", err)
			return nil, err
		}
		for _, scriptPathArgPortMapping := range scriptPathArgPortMappings {
			scriptPathArgPortMapping.Deleted = true
			scriptPathArgPortMapping.UpdatedBy = userId
			scriptPathArgPortMapping.UpdatedOn = time.Now()
		}
		err = impl.pipelineStageRepository.UpdateScriptMapping(scriptPathArgPortMappings, tx)
		if err != nil {
			impl.logger.Errorw("error in updating plugin script path arg port mapping", "err", err)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("deletePlugin, error in committing db transaction", "err", err)
		return nil, err
	}
	return pluginDeleteReq, nil
}

func (impl *GlobalPluginServiceImpl) getUserIdToEmailMap(pluginVersionsMetadata []*repository.PluginMetadata) (map[int32]string, error) {
	userIds := make([]int32, 0, len(pluginVersionsMetadata))
	for _, versionMetadata := range pluginVersionsMetadata {
		userIds = append(userIds, versionMetadata.UpdatedBy)
	}

	userIdVsEmailMap := make(map[int32]string, len(userIds))
	allUsers, err := impl.userService.GetByIds(userIds)
	if err != nil {
		impl.logger.Errorw("error in getting all user info", "err", err)
		return nil, err
	}
	for _, user := range allUsers {
		if _, ok := userIdVsEmailMap[user.Id]; !ok {
			userIdVsEmailMap[user.Id] = user.EmailId
		}
	}
	return userIdVsEmailMap, nil
}

// getPluginEntitiesIdToPluginEntitiesDtoMap returns two maps one is plugin parent ids vs plugin parent dto map and
// another map is plugin version ids vs plugin version details map, also returns error if any else nil.
func (impl *GlobalPluginServiceImpl) getPluginEntitiesIdToPluginEntitiesDtoMap(pluginVersionsMetadata []*repository.PluginMetadata, pluginsParentMetadata []*repository.PluginParentMetadata) (map[int]*bean2.PluginParentMetadataDto,
	map[int]map[int]*bean2.PluginsVersionDetail, error) {

	pluginParentIdVsPluginParentDtoMap, pluginParentIdMap := helper2.GetParentPluginDtoMappings(pluginsParentMetadata)

	filteredPluginVersionMetadata := make([]*repository.PluginMetadata, 0, len(pluginVersionsMetadata))
	for _, pluginVersion := range pluginVersionsMetadata {
		if _, ok := pluginParentIdMap[pluginVersion.PluginParentMetadataId]; ok {
			filteredPluginVersionMetadata = append(filteredPluginVersionMetadata, pluginVersion)
		}
	}

	userIdVsEmailMap, err := impl.getUserIdToEmailMap(filteredPluginVersionMetadata)
	if err != nil {
		return nil, nil, err
	}

	pluginVersionsVsPluginsVersionDetailMap := helper2.GetPluginVersionAndDetailsMapping(filteredPluginVersionMetadata, userIdVsEmailMap)

	helper2.AppendMinimalVersionDetailsInParentMetadataDtos(pluginParentIdVsPluginParentDtoMap, pluginVersionsVsPluginsVersionDetailMap)

	return pluginParentIdVsPluginParentDtoMap, pluginVersionsVsPluginsVersionDetailMap, nil
}

// GetPluginParentMetadataDtos populates PluginParentMetadataDto with all metadata(if a version is not latest then ignores populating heavy
// objects such as input and output variables in the dto) and returns the same with error if any.
func (impl *GlobalPluginServiceImpl) GetPluginParentMetadataDtos(parentIdVsPluginParentDtoMap map[int]*bean2.PluginParentMetadataDto,
	versionIdVsPluginsVersionDetailMap map[int]map[int]*bean2.PluginsVersionDetail, pluginVersionsIdToInclude map[int]bool, fetchAllVersionDetails bool) ([]*bean2.PluginParentMetadataDto, error) {
	//fetch input and output variables mappings
	pluginIdInputVariablesMap, pluginIdOutputVariablesMap, err := impl.getPluginIdVariablesMap()
	if err != nil {
		impl.logger.Errorw("error, getPluginIdVariablesMap", "err", err)
		return nil, err
	}
	pluginIdTagsMap, err := impl.getPluginIdTagsMap()
	if err != nil {
		impl.logger.Errorw("error, getPluginIdTagsMap", "err", err)
		return nil, err
	}

	pluginParentMetadataDtos := make([]*bean2.PluginParentMetadataDto, 0, len(parentIdVsPluginParentDtoMap))

	for parentPluginId, versionMap := range versionIdVsPluginsVersionDetailMap {
		detailedPluginVersionsMetadataDtos := make([]*bean2.PluginsVersionDetail, 0, len(versionMap)) //contains detailed plugin version data

		for pluginVersionId, versionDetail := range versionMap {
			pluginVersionDetail := *versionDetail
			if pluginVersionsIdToInclude != nil {
				if _, ok := pluginVersionsIdToInclude[pluginVersionId]; !ok {
					continue
				}
			}
			if fetchAllVersionDetails || pluginVersionDetail.IsLatest {
				inputVariables, ok := pluginIdInputVariablesMap[pluginVersionId]
				if ok {
					pluginVersionDetail.WithInputVariables(inputVariables)
				}
				outputVariables, ok := pluginIdOutputVariablesMap[pluginVersionId]
				if ok {
					pluginVersionDetail.WithOutputVariables(outputVariables)
				}
				tags, ok := pluginIdTagsMap[pluginVersionId]
				if ok {
					pluginVersionDetail.WithTags(tags)
				}
				detailedPluginVersionsMetadataDtos = append(detailedPluginVersionsMetadataDtos, &pluginVersionDetail)
			}
		}
		parentIdVsPluginParentDtoMap[parentPluginId].Versions.WithDetailedPluginVersionData(detailedPluginVersionsMetadataDtos)
	}

	for _, pluginParentDto := range parentIdVsPluginParentDtoMap {
		pluginParentMetadataDtos = append(pluginParentMetadataDtos, pluginParentDto)
	}

	return pluginParentMetadataDtos, nil
}

func (impl *GlobalPluginServiceImpl) ListAllPluginsV2(filter *bean2.PluginsListFilter) (*bean2.PluginsDto, error) {
	impl.logger.Infow("request received, ListAllPluginsV2", "filter", filter)
	pluginVersionsMetadata, err := impl.globalPluginRepository.GetMetaDataForAllPlugins()
	if err != nil {
		impl.logger.Errorw("ListAllPluginsV2, error in getting plugins", "err", err)
		return nil, err
	}

	allPluginParentMetadata, err := impl.globalPluginRepository.GetAllFilteredPluginParentMetadata(filter.SearchKey, filter.Tags)
	if err != nil {
		impl.logger.Errorw("ListAllPluginsV2, error in getting all plugin parent metadata", "err", err)
		return nil, err
	}
	if allPluginParentMetadata == nil {
		return bean2.NewPluginsDto(), nil
	}

	paginatedPluginParentMetadata := helper2.PaginatePluginParentMetadata(allPluginParentMetadata, filter.Limit, filter.Offset)

	parentIdVsPluginParentDtoMap, versionIdVsPluginsVersionDetailMap, err := impl.getPluginEntitiesIdToPluginEntitiesDtoMap(pluginVersionsMetadata, paginatedPluginParentMetadata)
	if err != nil {
		impl.logger.Errorw("ListAllPluginsV2, error in getPluginEntitiesIdToPluginEntitiesDtoMap", "err", err)
		return nil, err
	}

	pluginParentMetadataDtos, err := impl.GetPluginParentMetadataDtos(parentIdVsPluginParentDtoMap, versionIdVsPluginsVersionDetailMap, nil, filter.FetchAllVersionDetails)
	if err != nil {
		impl.logger.Errorw("ListAllPluginsV2, error in getting plugin parent metadata dtos for plugin list", "err", err)
		return nil, err
	}

	utils.SortParentMetadataDtoSliceByName(pluginParentMetadataDtos)
	pluginDetails := bean2.NewPluginsDto().WithParentPlugins(pluginParentMetadataDtos).WithTotalCount(len(allPluginParentMetadata))

	return pluginDetails, nil
}

// GetPluginDetailV2 returns all details of the of a plugin version according to the pluginVersionIds and parentPluginIds
// provided by user, and minimal data for all versions of that plugin.
func (impl *GlobalPluginServiceImpl) GetPluginDetailV2(pluginVersionIds, parentPluginIds []int, fetchAllVersionDetails bool) (*bean2.PluginsDto, error) {
	pluginParentMetadataDtos := make([]*bean2.PluginParentMetadataDto, 0, len(pluginVersionIds)+len(parentPluginIds))
	if len(pluginVersionIds) == 0 && len(parentPluginIds) == 0 {
		return nil, &util.ApiError{HttpStatusCode: http.StatusBadRequest, Code: strconv.Itoa(http.StatusBadRequest), InternalMessage: bean2.NoPluginOrParentIdProvidedErr, UserMessage: bean2.NoPluginOrParentIdProvidedErr}
	}
	pluginVersionIdsMap, parentPluginIdsMap := helper2.GetPluginVersionAndParentPluginIdsMap(pluginVersionIds, parentPluginIds)

	var err error
	pluginParentMetadataIds := make([]int, 0, len(pluginVersionIds)+len(parentPluginIds))
	pluginVersionsIdToInclude := make(map[int]bool, len(pluginVersionIds)+len(parentPluginIds))
	pluginVersionsMetadata, err := impl.globalPluginRepository.GetMetaDataForAllPlugins()
	if err != nil {
		impl.logger.Errorw("GetPluginDetailV2, error in getting all plugins versions metadata", "err", err)
		return nil, err
	}

	filteredPluginVersionMetadata := helper2.GetPluginVersionsMetadataByVersionAndParentPluginIds(pluginVersionsMetadata, pluginVersionIdsMap, parentPluginIdsMap)
	if len(filteredPluginVersionMetadata) == 0 {
		return nil, &util.ApiError{HttpStatusCode: http.StatusNotFound, Code: strconv.Itoa(http.StatusNotFound), InternalMessage: bean2.NoPluginFoundForThisSearchQueryErr, UserMessage: bean2.NoPluginFoundForThisSearchQueryErr}
	}
	for _, version := range filteredPluginVersionMetadata {
		_, found := pluginVersionIdsMap[version.Id]
		wantDetailedData := found || fetchAllVersionDetails || version.IsLatest
		if wantDetailedData {
			pluginVersionsIdToInclude[version.Id] = true
		}
		pluginParentMetadataIds = append(pluginParentMetadataIds, version.PluginParentMetadataId)
	}

	pluginParentDetails, err := impl.globalPluginRepository.GetPluginParentMetadataByIds(pluginParentMetadataIds)
	if err != nil {
		impl.logger.Errorw("GetPluginDetailV2, error in getting all plugin parent metadata by ids", "err", err)
		return nil, err
	}
	parentIdVsPluginParentDtoMap, versionIdVsPluginsVersionDetailMap, err := impl.getPluginEntitiesIdToPluginEntitiesDtoMap(pluginVersionsMetadata, pluginParentDetails)
	if err != nil {
		impl.logger.Errorw("GetPluginDetailV2, error in getPluginEntitiesIdToPluginEntitiesDtoMap", "err", err)
		return nil, err
	}
	pluginParentMetadataDtos, err = impl.GetPluginParentMetadataDtos(parentIdVsPluginParentDtoMap, versionIdVsPluginsVersionDetailMap, pluginVersionsIdToInclude, true)
	if err != nil {
		impl.logger.Errorw("GetPluginDetailV2, error in getting plugin parent metadata dtos by pluginParentMetadata ids", "pluginParentMetadataIds", pluginParentMetadataIds, "err", err)
		return nil, err
	}

	pluginsDto := bean2.NewPluginsDto().WithParentPlugins(pluginParentMetadataDtos)
	return pluginsDto, nil
}

func (impl *GlobalPluginServiceImpl) GetAllUniqueTags() (*bean2.PluginTagsDto, error) {
	tags, err := impl.globalPluginRepository.GetAllPluginTags()
	if err != nil {
		impl.logger.Errorw("GetAllUniqueTags, error in getting all plugin tags", "err", err)
		return nil, err
	}
	allUniqueTags := helper2.GetAllUniqueTags(tags)

	return bean2.NewPluginTagsDto().WithTagNames(allUniqueTags), nil
}

func (impl *GlobalPluginServiceImpl) MigratePluginData() error {
	pluginVersionsMetadata, err := impl.globalPluginRepository.GetMetaDataForAllPlugins()
	if err != nil {
		impl.logger.Errorw("MigratePluginData, error in getting plugins", "err", err)
		return err
	}
	err = impl.MigratePluginDataToParentPluginMetadata(pluginVersionsMetadata)
	if err != nil {
		impl.logger.Errorw("MigratePluginData, error in migrating plugin data into parent metadata table", "err", err)
		return err
	}
	return nil
}

// MigratePluginDataToParentPluginMetadata migrates pre-existing plugin metadata from plugin_metadata table into plugin_parent_metadata table,
// and also populate plugin_parent_metadata_id in plugin_metadata.
// this operation will happen only once when the get all plugin list v2 api is being called, returns error if any
func (impl *GlobalPluginServiceImpl) MigratePluginDataToParentPluginMetadata(pluginsMetadata []*repository.PluginMetadata) error {
	dbConnection := impl.globalPluginRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("MigratePluginDataToParentPluginMetadata, error in beginning transaction", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	pluginMetadataToUpdate := make([]*repository.PluginMetadata, 0, len(pluginsMetadata))
	identifierVsPluginMetadata := make(map[string]*repository.PluginMetadata, len(pluginsMetadata))
	pluginsIdentifierSlice := make([]string, 0, len(pluginsMetadata))
	for _, item := range pluginsMetadata {
		if item.PluginParentMetadataId > 0 {
			//data already migrated
			continue
		}
		identifier := utils.CreateUniqueIdentifier(item.Name, 0)
		if _, ok := identifierVsPluginMetadata[identifier]; ok {
			identifier = utils.CreateUniqueIdentifier(item.Name, item.Id)
		}
		identifierVsPluginMetadata[identifier] = item
		pluginsIdentifierSlice = append(pluginsIdentifierSlice, identifier)
	}

	for _, identifier := range pluginsIdentifierSlice {
		if pluginMetadata, ok := identifierVsPluginMetadata[identifier]; ok {
			pluginParentMetadata, err := impl.globalPluginRepository.GetPluginParentMetadataByIdentifier(identifier)
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("MigratePluginDataToParentPluginMetadata, error in GetPluginParentMetadataByIdentifier", "pluginIdentifier", identifier, "err", err)
				return err
			}
			if pluginParentMetadata != nil && pluginParentMetadata.Id > 0 {
				continue
			}
			parentMetadata := repository.NewPluginParentMetadata()
			parentMetadata.SetParentPluginMetadata(pluginMetadata).CreateAuditLog(bean.SystemUserId)
			parentMetadata.Identifier = identifier
			parentMetadata, err = impl.globalPluginRepository.SavePluginParentMetadata(tx, parentMetadata)
			if err != nil {
				impl.logger.Errorw("MigratePluginDataToParentPluginMetadata, error in saving plugin parent metadata", "err", err)
				return err
			}
			pluginMetadata.PluginParentMetadataId = parentMetadata.Id
			pluginMetadataToUpdate = append(pluginMetadataToUpdate, pluginMetadata)
		}
	}

	if len(pluginMetadataToUpdate) > 0 {
		err = impl.globalPluginRepository.UpdatePluginMetadataInBulk(pluginMetadataToUpdate, tx)
		if err != nil {
			impl.logger.Errorw("MigratePluginDataToParentPluginMetadata, error in updating plugin metadata in bulk", "err", err)
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("MigratePluginDataToParentPluginMetadata, error in committing db transaction", "err", err)
		return err
	}
	return nil
}
