package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type PluginType string
type ScriptType string
type ScriptImagePullSecretType string
type ScriptMappingType string
type PluginStepType string
type PluginStepVariableType string
type PluginStepVariableValueType string
type PluginStepConditionType string

const (
	PLUGIN_TYPE_SHARED                  PluginType                  = "SHARED"
	PLUGIN_TYPE_PRESET                  PluginType                  = "PRESET"
	SCRIPT_TYPE_SHELL                   ScriptType                  = "SHELL"
	SCRIPT_TYPE_DOCKERFILE              ScriptType                  = "DOCKERFILE"
	SCRIPT_TYPE_CONTAINER_IMAGE         ScriptType                  = "CONTAINER_IMAGE"
	IMAGE_PULL_TYPE_CONTAINER_REGISTRY  ScriptImagePullSecretType   = "CONTAINER_REGISTRY"
	IMAGE_PULL_TYPE_SECRET_PATH         ScriptImagePullSecretType   = "SECRET_PATH"
	SCRIPT_MAPPING_TYPE_FILE_PATH       ScriptMappingType           = "FILE_PATH"
	SCRIPT_MAPPING_TYPE_DOCKER_ARG      ScriptMappingType           = "DOCKER_ARG"
	SCRIPT_MAPPING_TYPE_PORT            ScriptMappingType           = "PORT"
	PLUGIN_STEP_TYPE_INLINE             PluginStepType              = "INLINE"
	PLUGIN_STEP_TYPE_REF_PLUGIN         PluginStepType              = "REF_PLUGIN"
	PLUGIN_VARIABLE_TYPE_INPUT          PluginStepVariableType      = "INPUT"
	PLUGIN_VARIABLE_TYPE_OUTPUT         PluginStepVariableType      = "OUTPUT"
	PLUGIN_VARIABLE_VALUE_TYPE_NEW      PluginStepVariableValueType = "NEW"
	PLUGIN_VARIABLE_VALUE_TYPE_PREVIOUS PluginStepVariableValueType = "FROM_PREVIOUS_STEP"
	PLUGIN_VARIABLE_VALUE_TYPE_GLOBAL   PluginStepVariableValueType = "GLOBAL"
	PLUGIN_CONDITION_TYPE_SKIP          PluginStepConditionType     = "SKIP"
	PLUGIN_CONDITION_TYPE_TRIGGER       PluginStepConditionType     = "TRIGGER"
	PLUGIN_CONDITION_TYPE_SUCCESS       PluginStepConditionType     = "SUCCESS"
	PLUGIN_CONDITION_TYPE_FAIL          PluginStepConditionType     = "FAIL"
)

type PluginMetadata struct {
	tableName   struct{}   `sql:"plugin_metadata" pg:",discard_unknown_columns"`
	Id          int        `sql:"id,pk"`
	Name        string     `sql:"name"`
	Description string     `sql:"description"`
	Type        PluginType `sql:"type"`
	Icon        string     `sql:"icon"`
	Deleted     bool       `sql:"deleted, notnull"`
	sql.AuditLog
}

type PluginTag struct {
	tableName struct{} `sql:"plugin_tag" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	Name      string   `sql:"name"`
	Deleted   bool     `sql:"deleted, notnull"`
	sql.AuditLog
}

type PluginTagRelation struct {
	tableName struct{} `sql:"plugin_tag_relation" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	TagId     int      `sql:"tag_id"`
	PluginId  int      `sql:"plugin_id"`
}

// Below two tables are used at pipeline-steps level too

type PluginPipelineScript struct {
	tableName            struct{}                  `sql:"plugin_pipeline_script" pg:",discard_unknown_columns"`
	Id                   int                       `sql:"id,pk"`
	Script               string                    `sql:"name"`
	Type                 ScriptType                `sql:"type"`
	DockerfileExists     bool                      `sql:"dockerfile_exists, notnull"`
	StoreScriptAt        string                    `sql:"store_script_at"`
	MountPath            string                    `sql:"mount_path"`
	MountCodeToContainer bool                      `sql:"mount_code_to_container,notnull"`
	ConfigureMountPath   bool                      `sql:"configure_mount_path,notnull"`
	ContainerImagePath   string                    `sql:"container_image_path"`
	ImagePullSecretType  ScriptImagePullSecretType `sql:"image_pull_secret_type"`
	ImagePullSecret      string                    `sql:"image_pull_secret"`
	Deleted              bool                      `sql:"deleted, notnull"`
	sql.AuditLog
}

type ScriptPathArgPortMapping struct {
	tableName           struct{}          `sql:"script_path_arg_port_mapping" pg:",discard_unknown_columns"`
	Id                  int               `sql:"id,pk"`
	TypeOfMapping       ScriptMappingType `sql:"type_of_mapping"`
	FilePathOnDisk      string            `sql:"file_path_on_disk"`
	FilePathOnContainer string            `sql:"file_path_on_container"`
	Command             string            `sql:"command"`
	Arg                 string            `sql:"arg"`
	PortOnLocal         int               `sql:"port_on_local"`
	PortOnContainer     int               `sql:"port_on_container"`
	ScriptId            int               `sql:"script_id"`
	sql.AuditLog
}

type PluginStep struct {
	tableName   struct{}       `sql:"plugin_step" pg:",discard_unknown_columns"`
	Id          int            `sql:"id,pk"`
	PluginId    int            `sql:"plugin_id"` //id of plugin - parent of this step
	Name        string         `sql:"name"`
	Description string         `sql:"description"`
	Index       int            `sql:"index"`
	StepType    PluginStepType `sql:"step_type"`
	ScriptId    int            `sql:"script_id"`
	RefPluginId int            `sql:"ref_plugin_id"` //id of plugin used as reference
	Deleted     bool           `sql:"deleted,notnull"`
	sql.AuditLog
}

type PluginStepVariable struct {
	tableName             struct{}                    `sql:"plugin_step_variable" pg:",discard_unknown_columns"`
	Id                    int                         `sql:"id,pk"`
	PluginStepId          int                         `sql:"plugin_step_id"`
	Name                  string                      `sql:"name"`
	Format                string                      `sql:"format"`
	Description           string                      `sql:"description"`
	IsExposed             bool                        `sql:"is_exposed,notnull"`
	AllowEmptyValue       bool                        `sql:"allow_empty_value,notnull"`
	DefaultValue          string                      `sql:"default_value"`
	Value                 string                      `sql:"value"`
	VariableType          PluginStepVariableType      `sql:"variable_type"`
	ValueType             PluginStepVariableValueType `sql:"value_type"`
	PreviousStepIndex     int                         `sql:"previous_step_index"`
	ReferenceVariableName string                      `sql:"reference_variable_name"`
	Deleted               bool                        `sql:"deleted,notnull"`
	sql.AuditLog
}

type PluginStepCondition struct {
	tableName           struct{}                `sql:"plugin_step_condition" pg:",discard_unknown_columns"`
	Id                  int                     `sql:"id,pk"`
	PluginStepId        int                     `sql:"plugin_step_id"`
	ConditionVariableId int                     `sql:"condition_variable_id"` //id of variable on which condition is written
	ConditionType       PluginStepConditionType `sql:"condition_type"`
	ConditionalOperator string                  `sql:"conditional_operator"`
	ConditionalValue    string                  `sql:"conditional_value"`
	Deleted             bool                    `sql:"deleted,notnull"`
	sql.AuditLog
}

type GlobalPluginRepository interface {
	GetMetaDataForAllPlugins() ([]*PluginMetadata, error)
	GetMetaDataByPluginId(pluginId int) (*PluginMetadata, error)
	GetTagsByPluginId(pluginId int) ([]string, error)
	GetExposedVariablesByPluginIdAndVariableType(pluginId int, variableType PluginStepVariableType) ([]*PluginStepVariable, error)
	GetExposedVariablesByPluginId(pluginId int) ([]*PluginStepVariable, error)
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

func (impl *GlobalPluginRepositoryImpl) GetMetaDataForAllPlugins() ([]*PluginMetadata, error) {
	var plugins []*PluginMetadata
	err := impl.dbConnection.Model(&plugins).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting all plugins", "err", err)
		return nil, err
	}
	return plugins, nil
}

func (impl *GlobalPluginRepositoryImpl) GetTagsByPluginId(pluginId int) ([]string, error) {
	var tags []string
	query := "SELECT pt.name from plugin_tags pt INNER JOIN plugin_tag_relation ptr on pt.id = ptr.tag_id where ptr.plugin_id = ? and pt.deleted = false"
	_, err := impl.dbConnection.Query(&tags, query, pluginId)
	if err != nil {
		impl.logger.Errorw("err in getting tags by pluginId", "err", err, "pluginId", pluginId)
		return nil, err
	}
	return tags, nil
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

func (impl *GlobalPluginRepositoryImpl) GetExposedVariablesByPluginIdAndVariableType(pluginId int, variableType PluginStepVariableType) ([]*PluginStepVariable, error) {
	var pluginVariables []*PluginStepVariable
	err := impl.dbConnection.Model(&pluginVariables).
		Column("plugin_step_variables.*").
		Join("INNER JOIN plugin_steps ps on ps.id = plugin_step_variables.plugin_step_id").
		Join("INNER JOIN plugin_metadata pm on pm.id = ps.plugin_id").
		Where("plugin_step_variables.deleted = ?", false).
		Where("plugin_step_variables.is_exposed = ?", true).
		Where("plugin_step_variables.variable_type = ?", variableType).
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
		Column("plugin_step_variables.*").
		Join("INNER JOIN plugin_steps ps on ps.id = plugin_step_variables.plugin_step_id").
		Join("INNER JOIN plugin_metadata pm on pm.id = ps.plugin_id").
		Where("plugin_step_variables.deleted = ?", false).
		Where("plugin_step_variables.is_exposed = ?", true).
		Where("ps.deleted = ?", false).
		Where("pm.deleted = ?", false).
		Where("pm.id = ?", pluginId).Select()
	if err != nil {
		impl.logger.Errorw("err in getting exposed variables by pluginId", "err", err, "pluginId", pluginId)
		return nil, err
	}
	return pluginVariables, nil
}
