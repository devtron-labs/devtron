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

package repository

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strings"
	"time"
)

type PluginType string

func (p PluginType) ToString() string {
	return string(p)
}

type ScriptType string
type ScriptImagePullSecretType string
type ScriptMappingType string
type PluginStepType string
type PluginStepVariableType string

func (p PluginStepVariableType) IsOutput() bool {
	return p == PLUGIN_VARIABLE_TYPE_OUTPUT
}

type PluginStepVariableValueType string

func (p PluginStepVariableValueType) String() string {
	return string(p)
}

func (p PluginStepVariableValueType) IsGlobalDefinedValue() bool {
	return p == PLUGIN_VARIABLE_VALUE_TYPE_GLOBAL
}

func (p PluginStepVariableValueType) IsPreviousOutputDefinedValue() bool {
	return p == PLUGIN_VARIABLE_VALUE_TYPE_PREVIOUS
}

type PluginStepConditionType string

type PluginStepVariableFormatType string

func (p PluginStepVariableFormatType) String() string {
	return string(p)
}

const (
	PLUGIN_TYPE_SHARED                  PluginType                   = "SHARED"
	PLUGIN_TYPE_PRESET                  PluginType                   = "PRESET"
	SCRIPT_TYPE_SHELL                   ScriptType                   = "SHELL"
	SCRIPT_TYPE_DOCKERFILE              ScriptType                   = "DOCKERFILE"
	SCRIPT_TYPE_CONTAINER_IMAGE         ScriptType                   = "CONTAINER_IMAGE"
	IMAGE_PULL_TYPE_CONTAINER_REGISTRY  ScriptImagePullSecretType    = "CONTAINER_REGISTRY"
	IMAGE_PULL_TYPE_SECRET_PATH         ScriptImagePullSecretType    = "SECRET_PATH"
	SCRIPT_MAPPING_TYPE_FILE_PATH       ScriptMappingType            = "FILE_PATH"
	SCRIPT_MAPPING_TYPE_DOCKER_ARG      ScriptMappingType            = "DOCKER_ARG"
	SCRIPT_MAPPING_TYPE_PORT            ScriptMappingType            = "PORT"
	PLUGIN_STEP_TYPE_INLINE             PluginStepType               = "INLINE"
	PLUGIN_STEP_TYPE_REF_PLUGIN         PluginStepType               = "REF_PLUGIN"
	PLUGIN_VARIABLE_TYPE_INPUT          PluginStepVariableType       = "INPUT"
	PLUGIN_VARIABLE_TYPE_OUTPUT         PluginStepVariableType       = "OUTPUT"
	PLUGIN_VARIABLE_VALUE_TYPE_NEW      PluginStepVariableValueType  = "NEW"
	PLUGIN_VARIABLE_VALUE_TYPE_PREVIOUS PluginStepVariableValueType  = "FROM_PREVIOUS_STEP"
	PLUGIN_VARIABLE_VALUE_TYPE_GLOBAL   PluginStepVariableValueType  = "GLOBAL"
	PLUGIN_CONDITION_TYPE_SKIP          PluginStepConditionType      = "SKIP"
	PLUGIN_CONDITION_TYPE_TRIGGER       PluginStepConditionType      = "TRIGGER"
	PLUGIN_CONDITION_TYPE_SUCCESS       PluginStepConditionType      = "SUCCESS"
	PLUGIN_CONDITION_TYPE_FAIL          PluginStepConditionType      = "FAIL"
	PLUGIN_VARIABLE_FORMAT_TYPE_STRING  PluginStepVariableFormatType = "STRING"
	PLUGIN_VARIABLE_FORMAT_TYPE_NUMBER  PluginStepVariableFormatType = "NUMBER"
	PLUGIN_VARIABLE_FORMAT_TYPE_BOOL    PluginStepVariableFormatType = "BOOL"
	PLUGIN_VARIABLE_FORMAT_TYPE_DATE    PluginStepVariableFormatType = "DATE"
)

const (
	CI                 = 1
	CD                 = 2
	CI_CD              = 3
	SCANNER            = 4
	CD_STAGE_TYPE      = "cd"
	CI_STAGE_TYPE      = "ci"
	CI_CD_STAGE_TYPE   = "ci_cd"
	SCANNER_STAGE_TYPE = "scanner"
	EXISTING_TAG_TYPE  = "existing_tags"
	NEW_TAG_TYPE       = "new_tags"
)

type PluginParentMetadata struct {
	tableName   struct{}   `sql:"plugin_parent_metadata" pg:",discard_unknown_columns"`
	Id          int        `sql:"id,pk"`
	Name        string     `sql:"name, notnull"`
	Identifier  string     `sql:"identifier, notnull"`
	Description string     `sql:"description"`
	Type        PluginType `sql:"type"`
	Icon        string     `sql:"icon"`
	Deleted     bool       `sql:"deleted, notnull"`
	IsExposed   bool       `sql:"is_exposed, notnull"` // it's not user driven, used internally to make decision weather to show plugin or not in plugin list
	sql.AuditLog
}

func NewPluginParentMetadata() *PluginParentMetadata {
	return &PluginParentMetadata{}
}

func (r *PluginParentMetadata) CreateAuditLog(userId int32) *PluginParentMetadata {
	r.CreatedBy = userId
	r.CreatedOn = time.Now()
	r.UpdatedBy = userId
	r.UpdatedOn = time.Now()
	return r
}
func (r *PluginParentMetadata) WithIsExposed(isExposed bool) *PluginParentMetadata {
	r.IsExposed = isExposed
	return r
}

func (r *PluginParentMetadata) WithBasicMetadata(name, identifier, description, icon string, pluginType PluginType, isExposed *bool) *PluginParentMetadata {
	r.Name = name
	r.Identifier = identifier
	r.Description = description
	r.Icon = icon
	r.Type = pluginType
	r.Deleted = false
	if isExposed != nil {
		r.IsExposed = *isExposed
	} else {
		//default set true
		r.IsExposed = true
	}
	return r
}

// SetParentPluginMetadata method signature used only for migration purposes, sets pluginVersionsMetadata into plugin_parent_metadata
func (r *PluginParentMetadata) SetParentPluginMetadata(pluginMetadata *PluginMetadata) *PluginParentMetadata {
	r.Name = pluginMetadata.Name
	r.Deleted = false
	r.Type = pluginMetadata.Type
	r.Icon = pluginMetadata.Icon
	r.Description = pluginMetadata.Description
	return r
}

func (r *PluginParentMetadata) CreateAndSetPluginIdentifier(pluginName string, pluginId int, isIdentifierDuplicated bool) *PluginParentMetadata {
	pluginName = strings.ToLower(pluginName)
	pluginName = strings.ReplaceAll(pluginName, " ", "_")
	r.Identifier = pluginName
	if isIdentifierDuplicated {
		r.Identifier = fmt.Sprintf("%s_%d", r.Identifier, pluginId)
	}
	return r
}

type PluginMetadata struct {
	tableName              struct{}   `sql:"plugin_metadata" pg:",discard_unknown_columns"`
	Id                     int        `sql:"id,pk"`
	Name                   string     `sql:"name"`
	Description            string     `sql:"description"`
	Type                   PluginType `sql:"type"` //deprecated
	Icon                   string     `sql:"icon"` //deprecated
	Deleted                bool       `sql:"deleted, notnull"`
	PluginParentMetadataId int        `sql:"plugin_parent_metadata_id"`
	PluginVersion          string     `sql:"plugin_version, notnull"`
	IsDeprecated           bool       `sql:"is_deprecated, notnull"`
	DocLink                string     `sql:"doc_link"`
	IsLatest               bool       `sql:"is_latest, notnull"`
	IsExposed              bool       `sql:"is_exposed, notnull"` // it's not user driven, used internally to make decision weather to show plugin or not in plugin list
	sql.AuditLog
}

func NewPluginVersionMetadata() *PluginMetadata {
	return &PluginMetadata{}
}

func (r *PluginMetadata) CreateAuditLog(userId int32) *PluginMetadata {
	r.CreatedBy = userId
	r.CreatedOn = time.Now()
	r.UpdatedBy = userId
	r.UpdatedOn = time.Now()
	return r
}

func (r *PluginMetadata) WithBasicMetadata(name, description, pluginVersion, docLink string, isExposed *bool) *PluginMetadata {
	r.Name = name
	r.PluginVersion = pluginVersion
	r.Description = description
	r.DocLink = docLink
	r.Deleted = false
	r.IsDeprecated = false
	if isExposed != nil {
		r.IsExposed = *isExposed
	} else {
		//default value is true
		r.IsExposed = true
	}
	return r
}

func (r *PluginMetadata) WithPluginParentMetadataId(parentId int) *PluginMetadata {
	r.PluginParentMetadataId = parentId
	return r
}

func (r *PluginMetadata) WithIsLatestFlag(isLatest bool) *PluginMetadata {
	r.IsLatest = isLatest
	return r
}

type PluginTag struct {
	tableName struct{} `sql:"plugin_tag" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	Name      string   `sql:"name"`
	Deleted   bool     `sql:"deleted, notnull"`
	sql.AuditLog
}

func NewPluginTag() *PluginTag {
	return &PluginTag{}
}

func (r *PluginTag) WithName(name string) *PluginTag {
	r.Name = name
	return r
}

func (r *PluginTag) CreateAuditLog(userId int32) *PluginTag {
	r.CreatedBy = userId
	r.CreatedOn = time.Now()
	r.UpdatedBy = userId
	r.UpdatedOn = time.Now()
	return r
}

type PluginTagRelation struct {
	tableName struct{} `sql:"plugin_tag_relation" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	TagId     int      `sql:"tag_id"`
	PluginId  int      `sql:"plugin_id"`
	sql.AuditLog
}

func NewPluginTagRelation() *PluginTagRelation {
	return &PluginTagRelation{}
}

func (r *PluginTagRelation) WithTagAndPluginId(tagId, pluginId int) *PluginTagRelation {
	r.TagId = tagId
	r.PluginId = pluginId
	return r
}

func (r *PluginTagRelation) CreateAuditLog(userId int32) *PluginTagRelation {
	r.CreatedBy = userId
	r.CreatedOn = time.Now()
	r.UpdatedBy = userId
	r.UpdatedOn = time.Now()
	return r
}

// Below two tables are used at pipeline-steps level too

type PluginPipelineScript struct {
	tableName                struct{}                  `sql:"plugin_pipeline_script" pg:",discard_unknown_columns"`
	Id                       int                       `sql:"id,pk"`
	Script                   string                    `sql:"script"`
	StoreScriptAt            string                    `sql:"store_script_at"`
	Type                     ScriptType                `sql:"type"`
	DockerfileExists         bool                      `sql:"dockerfile_exists, notnull"`
	MountPath                string                    `sql:"mount_path"`
	MountCodeToContainer     bool                      `sql:"mount_code_to_container,notnull"`
	MountCodeToContainerPath string                    `sql:"mount_code_to_container_path"`
	MountDirectoryFromHost   bool                      `sql:"mount_directory_from_host,notnull"`
	ContainerImagePath       string                    `sql:"container_image_path"`
	ImagePullSecretType      ScriptImagePullSecretType `sql:"image_pull_secret_type"`
	ImagePullSecret          string                    `sql:"image_pull_secret"`
	Deleted                  bool                      `sql:"deleted, notnull"`
	sql.AuditLog
}

type ScriptPathArgPortMapping struct {
	tableName           struct{}          `sql:"script_path_arg_port_mapping" pg:",discard_unknown_columns"`
	Id                  int               `sql:"id,pk"`
	TypeOfMapping       ScriptMappingType `sql:"type_of_mapping"`
	FilePathOnDisk      string            `sql:"file_path_on_disk"`
	FilePathOnContainer string            `sql:"file_path_on_container"`
	Command             string            `sql:"command"`
	Args                []string          `sql:"args" pg:",array"`
	PortOnLocal         int               `sql:"port_on_local"`
	PortOnContainer     int               `sql:"port_on_container"`
	ScriptId            int               `sql:"script_id"`
	sql.AuditLog
}

type PluginStep struct {
	tableName           struct{}       `sql:"plugin_step" pg:",discard_unknown_columns"`
	Id                  int            `sql:"id,pk"`
	PluginId            int            `sql:"plugin_id"` //id of plugin - parent of this step
	Name                string         `sql:"name"`
	Description         string         `sql:"description"`
	Index               int            `sql:"index"`
	StepType            PluginStepType `sql:"step_type"`
	ScriptId            int            `sql:"script_id"`
	RefPluginId         int            `sql:"ref_plugin_id"` //id of plugin used as reference
	OutputDirectoryPath []string       `sql:"output_directory_path" pg:",array"`
	DependentOnStep     string         `sql:"dependent_on_step"`
	Deleted             bool           `sql:"deleted,notnull"`
	sql.AuditLog
}

type PluginStepVariable struct {
	tableName                 struct{}                     `sql:"plugin_step_variable" pg:",discard_unknown_columns"`
	Id                        int                          `sql:"id,pk"`
	PluginStepId              int                          `sql:"plugin_step_id"`
	Name                      string                       `sql:"name"`
	Format                    PluginStepVariableFormatType `sql:"format"`
	Description               string                       `sql:"description"`
	IsExposed                 bool                         `sql:"is_exposed,notnull"`
	AllowEmptyValue           bool                         `sql:"allow_empty_value,notnull"`
	DefaultValue              string                       `sql:"default_value"`
	Value                     string                       `sql:"value"`
	VariableType              PluginStepVariableType       `sql:"variable_type"`
	ValueType                 PluginStepVariableValueType  `sql:"value_type"`
	PreviousStepIndex         int                          `sql:"previous_step_index,notnull"`
	VariableStepIndex         int                          `sql:"variable_step_index,notnull"`
	VariableStepIndexInPlugin int                          `sql:"variable_step_index_in_plugin,notnull"` // will contain stepIndex of variable in case of refPlugin
	ReferenceVariableName     string                       `sql:"reference_variable_name"`
	Deleted                   bool                         `sql:"deleted,notnull"`
	sql.AuditLog
	PluginMetadataId int `sql:"-"`
}

type PluginStepCondition struct {
	tableName           struct{}                `sql:"plugin_step_condition" pg:",discard_unknown_columns"`
	Id                  int                     `sql:"id,pk"`
	PluginStepId        int                     `sql:"plugin_step_id,notnull"`
	ConditionVariableId int                     `sql:"condition_variable_id,notnull"` //id of variable on which condition is written
	ConditionType       PluginStepConditionType `sql:"condition_type"`
	ConditionalOperator string                  `sql:"conditional_operator"`
	ConditionalValue    string                  `sql:"conditional_value"`
	Deleted             bool                    `sql:"deleted,notnull"`
	sql.AuditLog
}

type PluginStageMapping struct {
	tableName struct{} `sql:"plugin_stage_mapping" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	PluginId  int      `sql:"plugin_id"`
	StageType int      `sql:"stage_type"`
	sql.AuditLog
}

type GlobalPluginRepository interface {
	GetMetaDataForAllPlugins(excludeDeprecated bool) ([]*PluginMetadata, error)
	GetMetaDataForPluginWithStageType(stageType int) ([]*PluginMetadata, error)
	GetMetaDataByPluginId(pluginId int) (*PluginMetadata, error)
	GetMetaDataByPluginIds(pluginIds []int) ([]*PluginMetadata, error)
	GetAllPluginTags() ([]*PluginTag, error)
	GetPluginTagByNames(tagNames []string) ([]*PluginTag, error)
	GetAllPluginTagRelations() ([]*PluginTagRelation, error)
	GetScriptDetailById(id int) (*PluginPipelineScript, error)
	GetScriptDetailByIds(ids []int) ([]*PluginPipelineScript, error)
	GetScriptMappingDetailByScriptId(scriptId int) ([]*ScriptPathArgPortMapping, error)
	GetScriptMappingDetailByScriptIds(scriptIds []int) ([]*ScriptPathArgPortMapping, error)
	GetVariablesByStepId(stepId int) ([]*PluginStepVariable, error)
	GetVariablesByStepIds(stepIds []int) ([]*PluginStepVariable, error)
	GetStepsByPluginIds(pluginIds []int) ([]*PluginStep, error)
	GetExposedVariablesByPluginIdAndVariableType(pluginId int, variableType PluginStepVariableType) ([]*PluginStepVariable, error)
	GetExposedVariablesByPluginId(pluginId int) ([]*PluginStepVariable, error)
	GetExposedVariablesForAllPlugins() ([]*PluginStepVariable, error)
	GetConditionsByStepId(stepId int) ([]*PluginStepCondition, error)
	GetPluginByName(pluginName string) ([]*PluginMetadata, error)
	GetAllPluginMetaData() ([]*PluginMetadata, error)
	GetPluginStepsByPluginId(pluginId int) ([]*PluginStep, error)
	GetConditionsByPluginId(pluginId int) ([]*PluginStepCondition, error)
	GetPluginStageMappingByPluginId(pluginId int) (*PluginStageMapping, error)
	GetConnection() (dbConnection *pg.DB)
	GetPluginVersionsByParentId(parentPluginId int) ([]*PluginMetadata, error)

	GetPluginParentMetadataByIdentifier(pluginIdentifier string) (*PluginParentMetadata, error)
	GetPluginParentsMetadataByIdentifiers(pluginIdentifiers ...string) ([]*PluginParentMetadata, error)
	GetAllFilteredPluginParentMetadata(searchKey string, tags []string) ([]*PluginParentMetadata, error)
	GetPluginParentMetadataByIds(ids []int) ([]*PluginParentMetadata, error)
	GetAllPluginMinData() ([]*PluginParentMetadata, error)
	GetAllPluginMinDataByType(pluginType string) ([]*PluginParentMetadata, error)
	GetPluginParentMinDataById(id int) (*PluginParentMetadata, error)
	MarkPreviousPluginVersionLatestFalse(pluginParentId int) error
	GetPluginMetadataByPluginIdentifier(identifier string) (*PluginMetadata, error)

	SavePluginMetadata(pluginMetadata *PluginMetadata, tx *pg.Tx) (*PluginMetadata, error)
	SavePluginStageMapping(pluginStageMapping *PluginStageMapping, tx *pg.Tx) (*PluginStageMapping, error)
	SavePluginSteps(pluginStep *PluginStep, tx *pg.Tx) (*PluginStep, error)
	SavePluginPipelineScript(pluginPipelineScript *PluginPipelineScript, tx *pg.Tx) (*PluginPipelineScript, error)
	SavePluginStepVariables(pluginStepVariables *PluginStepVariable, tx *pg.Tx) (*PluginStepVariable, error)
	SavePluginStepConditions(pluginStepConditions *PluginStepCondition, tx *pg.Tx) (*PluginStepCondition, error)
	SavePluginTag(pluginTag *PluginTag, tx *pg.Tx) (*PluginTag, error)
	SavePluginTagRelation(pluginTagRelation *PluginTagRelation, tx *pg.Tx) (*PluginTagRelation, error)
	SavePluginTagInBulk(pluginTag []*PluginTag, tx *pg.Tx) error
	SavePluginTagRelationInBulk(pluginTagRelation []*PluginTagRelation, tx *pg.Tx) error
	//SavePluginParentMetadataInBulk(tx *pg.Tx, pluginParentMetadata []*PluginParentMetadata) error
	SavePluginParentMetadata(tx *pg.Tx, pluginParentMetadata *PluginParentMetadata) (*PluginParentMetadata, error)

	UpdatePluginMetadata(pluginMetadata *PluginMetadata, tx *pg.Tx) error
	UpdatePluginMetadataInBulk(pluginsMetadata []*PluginMetadata, tx *pg.Tx) error
	UpdatePluginStageMapping(pluginStageMapping *PluginStageMapping, tx *pg.Tx) error
	UpdatePluginSteps(pluginStep *PluginStep, tx *pg.Tx) error
	UpdatePluginPipelineScript(pluginPipelineScript *PluginPipelineScript, tx *pg.Tx) error
	UpdatePluginStepVariables(pluginStepVariables *PluginStepVariable, tx *pg.Tx) error
	UpdatePluginStepConditions(pluginStepConditions *PluginStepCondition, tx *pg.Tx) error
	UpdatePluginTag(pluginTag *PluginTag, tx *pg.Tx) error
	UpdatePluginTagRelation(pluginTagRelation *PluginTagRelation, tx *pg.Tx) error

	UpdateInBulkPluginSteps(pluginSteps []*PluginStep, tx *pg.Tx) error
	UpdateInBulkPluginStepVariables(pluginStepVariables []*PluginStepVariable, tx *pg.Tx) error
	UpdateInBulkPluginStepConditions(pluginStepConditions []*PluginStepCondition, tx *pg.Tx) error
}

func NewGlobalPluginRepository(logger *zap.SugaredLogger, dbConnection *pg.DB) *GlobalPluginRepositoryImpl {
	return &GlobalPluginRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type GlobalPluginRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func (impl *GlobalPluginRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return impl.dbConnection
}

func (impl *GlobalPluginRepositoryImpl) GetMetaDataForAllPlugins(excludeDeprecated bool) ([]*PluginMetadata, error) {
	var plugins []*PluginMetadata
	query := impl.dbConnection.Model(&plugins).
		Where("deleted = ?", false).
		Where("is_exposed = ?", true).
		Order("id")
	if excludeDeprecated {
		query = query.Where("is_deprecated = ?", false)
	}
	err := query.Select()
	if err != nil {
		impl.logger.Errorw("err in getting all plugins", "err", err)
		return nil, err
	}
	return plugins, nil
}

func (impl *GlobalPluginRepositoryImpl) GetMetaDataForPluginWithStageType(stageType int) ([]*PluginMetadata, error) {
	var plugins []*PluginMetadata
	err := impl.dbConnection.Model(&plugins).
		Join("INNER JOIN plugin_stage_mapping psm on psm.plugin_id=plugin_metadata.id").
		Where("plugin_metadata.deleted = ?", false).
		Where("psm.stage_type= 2 or psm.stage_type= ?", stageType).
		Select()
	if err != nil {
		impl.logger.Errorw("err in getting all plugins", "err", err)
		return nil, err
	}
	return plugins, nil
}

func (impl *GlobalPluginRepositoryImpl) GetAllPluginTags() ([]*PluginTag, error) {
	var tags []*PluginTag
	err := impl.dbConnection.Model(&tags).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting all tags", "err", err)
		return nil, err
	}
	return tags, nil
}

func (impl *GlobalPluginRepositoryImpl) GetPluginTagByNames(tagNames []string) ([]*PluginTag, error) {
	var tags []*PluginTag
	err := impl.dbConnection.Model(&tags).
		Where("deleted = ?", false).
		Where("name in (?)", pg.In(tagNames)).
		Select()
	if err != nil {
		impl.logger.Errorw("err in getting all tags by names", "tagNames", tagNames, "err", err)
		return nil, err
	}
	return tags, nil
}

func (impl *GlobalPluginRepositoryImpl) GetAllPluginTagRelations() ([]*PluginTagRelation, error) {
	var rel []*PluginTagRelation
	err := impl.dbConnection.Model(&rel).
		Join("INNER JOIN plugin_metadata pm ON pm.id = plugin_tag_relation.plugin_id").
		Where("pm.deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting all tags", "err", err)
		return nil, err
	}
	return rel, nil
}

func (impl *GlobalPluginRepositoryImpl) GetMetaDataByPluginId(pluginId int) (*PluginMetadata, error) {
	var plugin PluginMetadata
	err := impl.dbConnection.Model(&plugin).Where("id = ?", pluginId).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting plugin by id", "err", err, "pluginId", pluginId)
		return nil, err
	}
	return &plugin, nil
}

func (impl *GlobalPluginRepositoryImpl) GetMetaDataByPluginIds(pluginIds []int) ([]*PluginMetadata, error) {
	var plugins []*PluginMetadata
	err := impl.dbConnection.Model(&plugins).
		Where("deleted = ?", false).
		Where("id in (?)", pg.In(pluginIds)).Select()
	if err != nil {
		impl.logger.Errorw("err in getting plugins by pluginIds", "pluginIds", pluginIds, "err", err)
		return nil, err
	}
	return plugins, nil
}

func (impl *GlobalPluginRepositoryImpl) GetStepsByPluginIds(pluginIds []int) ([]*PluginStep, error) {
	var pluginSteps []*PluginStep
	err := impl.dbConnection.Model(&pluginSteps).
		Where("deleted = ?", false).
		Where("plugin_id in (?)", pg.In(pluginIds)).Select()
	if err != nil {
		impl.logger.Errorw("err in getting plugin steps by pluginIds", "err", err, "pluginIds", pluginIds)
		return nil, err
	}
	return pluginSteps, nil
}

func (impl *GlobalPluginRepositoryImpl) GetScriptDetailById(id int) (*PluginPipelineScript, error) {
	var scriptDetail PluginPipelineScript
	err := impl.dbConnection.Model(&scriptDetail).
		Where("id = ?", id).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting script detail by id", "err", err, "id", id)
		return nil, err
	}
	return &scriptDetail, nil
}

func (impl *GlobalPluginRepositoryImpl) GetScriptDetailByIds(ids []int) ([]*PluginPipelineScript, error) {
	var scriptDetail []*PluginPipelineScript
	err := impl.dbConnection.Model(&scriptDetail).
		Where("id in (?)", pg.In(ids)).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting script detail by ids", "ids", ids, "err", err)
		return nil, err
	}
	return scriptDetail, nil
}

func (impl *GlobalPluginRepositoryImpl) GetScriptMappingDetailByScriptId(scriptId int) ([]*ScriptPathArgPortMapping, error) {
	var scriptMappingDetail []*ScriptPathArgPortMapping
	err := impl.dbConnection.Model(&scriptMappingDetail).
		Where("script_id = ?", scriptId).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting script mapping detail by id", "err", err, "scriptId", scriptId)
		return nil, err
	}
	return scriptMappingDetail, nil
}

func (impl *GlobalPluginRepositoryImpl) GetScriptMappingDetailByScriptIds(scriptIds []int) ([]*ScriptPathArgPortMapping, error) {
	var scriptMappingDetail []*ScriptPathArgPortMapping
	err := impl.dbConnection.Model(&scriptMappingDetail).
		Where("script_id in (?)", pg.In(scriptIds)).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting script mapping detail by id", "scriptIds", scriptIds, "err", err)
		return nil, err
	}
	return scriptMappingDetail, nil
}

func (impl *GlobalPluginRepositoryImpl) GetVariablesByStepId(stepId int) ([]*PluginStepVariable, error) {
	var variables []*PluginStepVariable
	err := impl.dbConnection.Model(&variables).
		Where("plugin_step_id = ?", stepId).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting variables by stepId", "err", err, "stepId", stepId)
		return nil, err
	}
	return variables, nil
}

func (impl *GlobalPluginRepositoryImpl) GetVariablesByStepIds(stepIds []int) ([]*PluginStepVariable, error) {
	var variables []*PluginStepVariable
	err := impl.dbConnection.Model(&variables).
		Where("plugin_step_id in (?)", pg.In(stepIds)).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting variables by stepIds", "stepIds", stepIds, "err", err)
		return nil, err
	}
	return variables, nil
}

func (impl *GlobalPluginRepositoryImpl) GetExposedVariablesByPluginIdAndVariableType(pluginId int, variableType PluginStepVariableType) ([]*PluginStepVariable, error) {
	var pluginVariables []*PluginStepVariable
	err := impl.dbConnection.Model(&pluginVariables).
		Column("plugin_step_variable.*").
		Join("INNER JOIN plugin_step ps on ps.id = plugin_step_variable.plugin_step_id").
		Join("INNER JOIN plugin_metadata pm on pm.id = ps.plugin_id").
		Where("plugin_step_variable.deleted = ?", false).
		Where("plugin_step_variable.is_exposed = ?", true).
		Where("plugin_step_variable.variable_type = ?", variableType).
		Where("ps.deleted = ?", false).
		Where("pm.deleted = ?", false).
		Where("pm.id = ?", pluginId).Select()
	if err != nil {
		impl.logger.Errorw("err in getting exposed variables by pluginId and variableType", "err", err, "pluginId", pluginId, "variableType", variableType)
		return nil, err
	}
	return pluginVariables, nil
}

func (impl *GlobalPluginRepositoryImpl) GetExposedVariablesByPluginId(pluginId int) ([]*PluginStepVariable, error) {
	var pluginVariables []*PluginStepVariable
	err := impl.dbConnection.Model(&pluginVariables).
		Column("plugin_step_variable.*").
		Join("INNER JOIN plugin_step ps on ps.id = plugin_step_variable.plugin_step_id").
		Join("INNER JOIN plugin_metadata pm on pm.id = ps.plugin_id").
		Where("plugin_step_variable.deleted = ?", false).
		Where("plugin_step_variable.is_exposed = ?", true).
		Where("ps.deleted = ?", false).
		Where("pm.deleted = ?", false).
		Where("pm.id = ?", pluginId).Select()
	if err != nil {
		impl.logger.Errorw("err in getting exposed variables by pluginId", "err", err, "pluginId", pluginId)
		return nil, err
	}
	return pluginVariables, nil
}

func (impl *GlobalPluginRepositoryImpl) GetExposedVariablesForAllPlugins() ([]*PluginStepVariable, error) {
	var pluginVariables []*PluginStepVariable
	query := `SELECT psv.*, pm.id as plugin_metadata_id from plugin_step_variable psv 
    				INNER JOIN plugin_step ps on ps.id = psv.plugin_step_id
    				INNER JOIN plugin_metadata pm on pm.id = ps.plugin_id 
    				WHERE psv.deleted = ? and psv.is_exposed = ? and
    				ps.deleted = ? and pm.deleted = ?;`
	_, err := impl.dbConnection.Query(&pluginVariables, query, false, true, false, false)
	if err != nil {
		impl.logger.Errorw("err in getting exposed variables for all plugins", "err", err)
		return nil, err
	}
	return pluginVariables, nil
}

func (impl *GlobalPluginRepositoryImpl) GetConditionsByStepId(stepId int) ([]*PluginStepCondition, error) {
	var conditions []*PluginStepCondition
	err := impl.dbConnection.Model(&conditions).
		Where("plugin_step_id = ?", stepId).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting conditions by stepId", "err", err, "stepId", stepId)
		return nil, err
	}
	return conditions, nil
}

func (impl *GlobalPluginRepositoryImpl) GetPluginByName(pluginName string) ([]*PluginMetadata, error) {
	var plugin []*PluginMetadata
	err := impl.dbConnection.Model(&plugin).
		Where("name = ?", pluginName).
		Where("deleted = ?", false).
		Select()
	if err != nil {
		impl.logger.Errorw("err in getting pluginMetadata by pluginName", "err", err, "pluginName", pluginName)
		return nil, err
	}
	return plugin, nil

}

func (impl *GlobalPluginRepositoryImpl) GetPluginVersionsByParentId(parentPluginId int) ([]*PluginMetadata, error) {
	var plugin []*PluginMetadata
	err := impl.dbConnection.Model(&plugin).
		Where("plugin_parent_metadata_id = ?", parentPluginId).
		Where("deleted = ?", false).
		Where("is_deprecated = ?", false).
		Select()
	if err != nil {
		impl.logger.Errorw("err in getting pluginVersionMetadata by parentPluginId", "parentPluginId", parentPluginId, "err", err)
		return nil, err
	}
	return plugin, nil
}

func (impl *GlobalPluginRepositoryImpl) GetAllPluginMetaData() ([]*PluginMetadata, error) {
	var plugins []*PluginMetadata
	err := impl.dbConnection.Model(&plugins).Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting all plugin metadata", "err", err)
		return nil, err
	}
	return plugins, nil
}

func (impl *GlobalPluginRepositoryImpl) GetPluginStepsByPluginId(pluginId int) ([]*PluginStep, error) {
	var pluginSteps []*PluginStep
	err := impl.dbConnection.Model(&pluginSteps).
		Where("deleted = ?", false).
		Where("plugin_id = ?", pluginId).Select()
	if err != nil {
		impl.logger.Errorw("err in getting plugin steps by pluginId", "err", err, "pluginId", pluginId)
		return nil, err
	}
	return pluginSteps, nil
}

// GetConditionsByPluginId fetches plugin step variable conditions by plugin id
func (impl *GlobalPluginRepositoryImpl) GetConditionsByPluginId(pluginId int) ([]*PluginStepCondition, error) {
	var pluginStepConditions []*PluginStepCondition
	err := impl.dbConnection.Model(&pluginStepConditions).
		Column("plugin_step_condition.*").
		Join("INNER JOIN plugin_step_variable psv on psv.id = plugin_step_condition.condition_variable_id").
		Join("INNER JOIN plugin_step ps on ps.id = psv.plugin_step_id").
		Join("INNER JOIN plugin_metadata pm on pm.id = ps.plugin_id").
		Where("plugin_step_condition.deleted = ?", false).
		Where("psv.deleted = ?", false).
		Where("ps.deleted = ?", false).
		Where("pm.deleted = ?", false).
		Where("pm.id = ?", pluginId).Select()
	if err != nil {
		impl.logger.Errorw("err in getting plugin step conditions", "err", err, "pluginId", pluginId)
		return nil, err
	}
	return pluginStepConditions, nil
}

func (impl *GlobalPluginRepositoryImpl) GetPluginStageMappingByPluginId(pluginId int) (*PluginStageMapping, error) {
	var pluginStageMapping PluginStageMapping
	err := impl.dbConnection.Model(&pluginStageMapping).
		Where("plugin_id = ?", pluginId).Select()
	if err != nil {
		impl.logger.Errorw("err in getting pluginStageMapping", "err", err, "pluginId", pluginId)
		return nil, err
	}
	return &pluginStageMapping, nil
}

func (impl *GlobalPluginRepositoryImpl) SavePluginMetadata(pluginMetadata *PluginMetadata, tx *pg.Tx) (*PluginMetadata, error) {
	err := tx.Insert(pluginMetadata)
	if err != nil {
		impl.logger.Errorw("error in saving pluginMetadata", "err", err)
		return pluginMetadata, err
	}
	return pluginMetadata, nil
}

func (impl *GlobalPluginRepositoryImpl) SavePluginStageMapping(pluginStageMapping *PluginStageMapping, tx *pg.Tx) (*PluginStageMapping, error) {
	err := tx.Insert(pluginStageMapping)
	if err != nil {
		impl.logger.Errorw("error in saving pluginStageMapping", "err", err)
		return pluginStageMapping, err
	}
	return pluginStageMapping, nil
}
func (impl *GlobalPluginRepositoryImpl) SavePluginSteps(pluginStep *PluginStep, tx *pg.Tx) (*PluginStep, error) {
	err := tx.Insert(pluginStep)
	if err != nil {
		impl.logger.Errorw("error in saving pluginStep", "err", err)
		return pluginStep, err
	}
	return pluginStep, nil
}

func (impl *GlobalPluginRepositoryImpl) SavePluginPipelineScript(pluginPipelineScript *PluginPipelineScript, tx *pg.Tx) (*PluginPipelineScript, error) {
	err := tx.Insert(pluginPipelineScript)
	if err != nil {
		impl.logger.Errorw("error in saving pluginPipelineScript", "err", err)
		return pluginPipelineScript, err
	}
	return pluginPipelineScript, nil
}

func (impl *GlobalPluginRepositoryImpl) SavePluginStepVariables(pluginStepVariables *PluginStepVariable, tx *pg.Tx) (*PluginStepVariable, error) {
	err := tx.Insert(pluginStepVariables)
	if err != nil {
		impl.logger.Errorw("error in saving pluginStepVariables", "err", err)
		return pluginStepVariables, err
	}
	return pluginStepVariables, nil
}

func (impl *GlobalPluginRepositoryImpl) SavePluginStepConditions(pluginStepConditions *PluginStepCondition, tx *pg.Tx) (*PluginStepCondition, error) {
	err := tx.Insert(pluginStepConditions)
	if err != nil {
		impl.logger.Errorw("error in saving pluginStepConditions", "err", err)
		return pluginStepConditions, err
	}
	return pluginStepConditions, nil
}

func (impl *GlobalPluginRepositoryImpl) SavePluginTag(pluginTags *PluginTag, tx *pg.Tx) (*PluginTag, error) {
	err := tx.Insert(pluginTags)
	if err != nil {
		impl.logger.Errorw("error in saving pluginTags", "err", err)
		return pluginTags, err
	}
	return pluginTags, nil
}

func (impl *GlobalPluginRepositoryImpl) SavePluginTagRelation(pluginTagRelation *PluginTagRelation, tx *pg.Tx) (*PluginTagRelation, error) {
	err := tx.Insert(pluginTagRelation)
	if err != nil {
		impl.logger.Errorw("error in saving pluginTagRelation", "err", err)
		return pluginTagRelation, err
	}
	return pluginTagRelation, nil
}

func (impl *GlobalPluginRepositoryImpl) SavePluginTagInBulk(pluginTag []*PluginTag, tx *pg.Tx) error {
	err := tx.Insert(&pluginTag)
	return err
}

func (impl *GlobalPluginRepositoryImpl) SavePluginTagRelationInBulk(pluginTagRelation []*PluginTagRelation, tx *pg.Tx) error {
	err := tx.Insert(&pluginTagRelation)
	return err
}

func (impl *GlobalPluginRepositoryImpl) UpdatePluginMetadata(pluginMetadata *PluginMetadata, tx *pg.Tx) error {
	return tx.Update(pluginMetadata)
}

func (impl *GlobalPluginRepositoryImpl) UpdatePluginStageMapping(pluginStageMapping *PluginStageMapping, tx *pg.Tx) error {
	return tx.Update(pluginStageMapping)
}

func (impl *GlobalPluginRepositoryImpl) UpdatePluginSteps(pluginStep *PluginStep, tx *pg.Tx) error {
	return tx.Update(pluginStep)
}

func (impl *GlobalPluginRepositoryImpl) UpdatePluginPipelineScript(pluginPipelineScript *PluginPipelineScript, tx *pg.Tx) error {
	return tx.Update(pluginPipelineScript)
}

func (impl *GlobalPluginRepositoryImpl) UpdatePluginStepVariables(pluginStepVariables *PluginStepVariable, tx *pg.Tx) error {
	return tx.Update(pluginStepVariables)
}

func (impl *GlobalPluginRepositoryImpl) UpdatePluginStepConditions(pluginStepConditions *PluginStepCondition, tx *pg.Tx) error {
	return tx.Update(pluginStepConditions)
}

func (impl *GlobalPluginRepositoryImpl) UpdatePluginTag(pluginTag *PluginTag, tx *pg.Tx) error {
	return tx.Update(pluginTag)
}

func (impl *GlobalPluginRepositoryImpl) UpdatePluginTagRelation(pluginTagRelation *PluginTagRelation, tx *pg.Tx) error {
	return tx.Update(pluginTagRelation)
}

func (impl *GlobalPluginRepositoryImpl) UpdateInBulkPluginSteps(pluginSteps []*PluginStep, tx *pg.Tx) error {
	_, err := tx.Model(&pluginSteps).Update()
	return err
}
func (impl *GlobalPluginRepositoryImpl) UpdateInBulkPluginStepVariables(pluginStepVariables []*PluginStepVariable, tx *pg.Tx) error {
	_, err := tx.Model(&pluginStepVariables).Update()
	return err
}
func (impl *GlobalPluginRepositoryImpl) UpdateInBulkPluginStepConditions(pluginStepConditions []*PluginStepCondition, tx *pg.Tx) error {
	_, err := tx.Model(&pluginStepConditions).Update()
	return err
}

func (impl *GlobalPluginRepositoryImpl) GetPluginParentMetadataByIdentifier(pluginIdentifier string) (*PluginParentMetadata, error) {
	var pluginParentMetadata PluginParentMetadata
	err := impl.dbConnection.Model(&pluginParentMetadata).
		Where("identifier = ?", pluginIdentifier).
		Where("deleted = ?", false).
		Select()
	if err != nil {
		return nil, err
	}
	return &pluginParentMetadata, nil
}

func (impl *GlobalPluginRepositoryImpl) GetPluginParentsMetadataByIdentifiers(pluginIdentifiers ...string) ([]*PluginParentMetadata, error) {
	if len(pluginIdentifiers) == 0 {
		return []*PluginParentMetadata{}, nil
	}
	var pluginParentMetadata []*PluginParentMetadata
	err := impl.dbConnection.Model(&pluginParentMetadata).
		Where("identifier IN (?)", pg.In(pluginIdentifiers)).
		Where("deleted = ?", false).
		Select()
	if err != nil {
		return nil, err
	}
	return pluginParentMetadata, nil
}

func (impl *GlobalPluginRepositoryImpl) GetPluginParentMinDataById(id int) (*PluginParentMetadata, error) {
	var pluginParentMetadata PluginParentMetadata
	err := impl.dbConnection.Model(&pluginParentMetadata).
		Column("plugin_parent_metadata.id", "plugin_parent_metadata.name").
		Where("id = ?", id).
		Where("deleted = ?", false).Select()
	if err != nil {
		return nil, err
	}
	return &pluginParentMetadata, nil
}

func (impl *GlobalPluginRepositoryImpl) SavePluginParentMetadata(tx *pg.Tx, pluginParentMetadata *PluginParentMetadata) (*PluginParentMetadata, error) {
	err := tx.Insert(pluginParentMetadata)
	return pluginParentMetadata, err
}

func (impl *GlobalPluginRepositoryImpl) UpdatePluginMetadataInBulk(pluginsMetadata []*PluginMetadata, tx *pg.Tx) error {
	_, err := tx.Model(&pluginsMetadata).Update()
	return err
}

func (impl *GlobalPluginRepositoryImpl) GetAllFilteredPluginParentMetadata(searchKey string, tags []string) ([]*PluginParentMetadata, error) {
	var plugins []*PluginParentMetadata
	query := "select ppm.id, ppm.identifier,ppm.name,ppm.description,ppm.type,ppm.icon,ppm.deleted,ppm.created_by, ppm.created_on,ppm.updated_by,ppm.updated_on from plugin_parent_metadata ppm" +
		" inner join plugin_metadata pm on pm.plugin_parent_metadata_id=ppm.id"
	whereCondition := fmt.Sprintf(" where ppm.deleted=false AND pm.deleted=false AND pm.is_latest=true AND pm.is_deprecated=false AND pm.is_exposed=true AND ppm.is_exposed=true")
	if len(tags) > 0 {
		tagFilterSubQuery := fmt.Sprintf("select ptr.plugin_id from plugin_tag_relation ptr inner join plugin_tag pt on ptr.tag_id =pt.id where pt.deleted =false and  pt.name in (%s) group by ptr.plugin_id having count(ptr.plugin_id )=%d", helper.GetCommaSepratedStringWithComma(tags), len(tags))
		whereCondition += fmt.Sprintf(" AND pm.id in (%s)", tagFilterSubQuery)
	}
	if len(searchKey) > 0 {
		searchKeyLike := "%" + searchKey + "%"
		whereCondition += fmt.Sprintf(" AND (pm.description ilike '%s' or pm.name ilike '%s')", searchKeyLike, searchKeyLike)
	}
	orderCondition := " ORDER BY ppm.name asc;"

	query += whereCondition + orderCondition
	_, err := impl.dbConnection.Query(&plugins, query)
	if err != nil {
		return nil, err
	}
	return plugins, nil
}

func (impl *GlobalPluginRepositoryImpl) GetPluginParentMetadataByIds(ids []int) ([]*PluginParentMetadata, error) {
	var plugins []*PluginParentMetadata
	err := impl.dbConnection.Model(&plugins).
		Where("id in (?)", pg.In(ids)).
		Where("deleted = ?", false).
		Select()
	if err != nil {
		impl.logger.Errorw("err in getting pluginParentMetadata by ids", "ids", ids, "err", err)
		return nil, err
	}
	return plugins, nil
}

func (impl *GlobalPluginRepositoryImpl) GetAllPluginMinData() ([]*PluginParentMetadata, error) {
	return impl.GetAllPluginMinDataByType("")
}

func (impl *GlobalPluginRepositoryImpl) GetAllPluginMinDataByType(pluginType string) ([]*PluginParentMetadata, error) {
	var plugins []*PluginParentMetadata
	query := impl.dbConnection.Model(&plugins).
		Column("plugin_parent_metadata.id", "plugin_parent_metadata.name", "plugin_parent_metadata.type", "plugin_parent_metadata.icon", "plugin_parent_metadata.identifier").
		Where("deleted = ?", false).
		Where("is_exposed = ?", true)
	if len(pluginType) != 0 {
		query.Where("type = ?", pluginType)
	}
	err := query.Select()
	if err != nil {
		impl.logger.Errorw("err in getting all plugin parent metadata min data", "err", err)
		return nil, err
	}
	return plugins, nil
}

func (impl *GlobalPluginRepositoryImpl) MarkPreviousPluginVersionLatestFalse(pluginParentId int) error {
	var model PluginMetadata
	_, err := impl.dbConnection.Model(&model).
		Set("is_latest = ?", false).
		Where("id = (select id from plugin_metadata where plugin_parent_metadata_id = ? and is_latest =true order by created_on desc limit ?)", pluginParentId, 1).
		Update()
	if err != nil {
		impl.logger.Errorw("error in updating last version isLatest as false for a plugin parent id", "pluginParentId", pluginParentId, "err", err)
		return err
	}
	return nil
}

func (impl *GlobalPluginRepositoryImpl) GetPluginMetadataByPluginIdentifier(identifier string) (*PluginMetadata, error) {
	pluginMetadata := &PluginMetadata{}
	err := impl.dbConnection.Model(pluginMetadata).
		Join("INNER JOIN plugin_parent_metadata ppm").
		JoinOn("plugin_metadata.plugin_parent_metadata_id = ppm.id").
		Where("ppm.deleted = ?", false).
		Where("plugin_metadata.deleted = ?", false).
		Where("ppm.identifier = ?", identifier).
		Select()
	if err != nil {
		impl.logger.Errorw("err in getting plugin metadata by plugin identifier", "identifier", identifier, "err", err)
		return nil, err
	}
	return pluginMetadata, nil
}
