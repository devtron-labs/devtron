package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type PluginMetadata struct {
	tableName   struct{} `sql:"plugin_metadata" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	Name        string   `sql:"name"`
	Description string   `sql:"description"`
	Type        string   `sql:"type"` // SHARED, PRESET etc
	Icon        string   `sql:"icon"`
	Deleted     bool     `sql:"deleted, notnull"`
	sql.AuditLog
}

type PluginTags struct {
	tableName struct{} `sql:"plugin_tags" pg:",discard_unknown_columns"`
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
	tableName            struct{} `sql:"plugin_pipeline_script" pg:",discard_unknown_columns"`
	Id                   int      `sql:"id,pk"`
	Script               string   `sql:"name"`
	Type                 string   `sql:"type"` // SHELL, DOCKERFILE, CONTAINER_IMAGE etc
	DockerfileExists     bool     `sql:"dockerfile_exists, notnull"`
	StoreScriptAt        string   `sql:"store_script_at"`
	MountPath            string   `sql:"mount_path"`
	MountCodeToContainer bool     `sql:"mount_code_to_container,notnull"`
	ConfigureMountPath   bool     `sql:"configure_mount_path,notnull"`
	ContainerImagePath   string   `sql:"container_image_path"`
	ImagePullSecretType  string   `sql:"image_pull_secret_type"` //CONTAINER_REGISTRY or SECRET_PATH
	ImagePullSecret      string   `sql:"image_pull_secret"`
	Deleted              bool     `sql:"deleted, notnull"`
	sql.AuditLog
}

type ScriptPathArgPortMappings struct {
	tableName           struct{} `sql:"script_path_arg_port_mappings" pg:",discard_unknown_columns"`
	Id                  int      `sql:"id,pk"`
	TypeOfMapping       string   `sql:"type_of_mapping"` // FILE_PATH, DOCKER_ARG, PORT
	FilePathOnDisk      string   `sql:"file_path_on_disk"`
	FilePathOnContainer string   `sql:"file_path_on_container"`
	Command             string   `sql:"command"`
	Arg                 string   `sql:"arg"`
	PortOnLocal         int      `sql:"port_on_local"`
	PortOnContainer     int      `sql:"port_on_container"`
	ScriptId            int      `sql:"script_id"`
	sql.AuditLog
}

type PluginSteps struct {
	tableName   struct{} `sql:"plugin_steps" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	PluginId    int      `sql:"plugin_id"` //id of plugin - parent of this step
	Name        string   `sql:"name"`
	Description string   `sql:"description"`
	Index       int      `sql:"index"`
	StepType    string   `sql:"step_type"` //INLINE or REF_PLUGIN
	ScriptId    int      `sql:"script_id"`
	RefPluginId int      `sql:"ref_plugin_id"` //id of plugin used as reference
	Deleted     bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

type PluginStepVariables struct {
	tableName             struct{} `sql:"plugin_step_variables" pg:",discard_unknown_columns"`
	Id                    int      `sql:"id,pk"`
	PluginStepId          int      `sql:"plugin_step_id"`
	Name                  string   `sql:"name"`
	Format                string   `sql:"format"`
	Description           string   `sql:"description"`
	IsExposed             bool     `sql:"is_exposed,notnull"`
	AllowEmptyValue       bool     `sql:"allow_empty_value,notnull"`
	DefaultValue          string   `sql:"default_value"`
	Value                 string   `sql:"value"`
	VariableType          string   `sql:"variable_type"` //INPUT or OUTPUT
	Index                 int      `sql:"index"`
	ValueType             string   `sql:"value_type"` //NEW, FROM_PREVIOUS_STEP or GLOBAL
	PreviousStepIndex     int      `sql:"previous_step_index"`
	ReferenceVariableName string   `sql:"reference_variable_name"`
	Deleted               bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

type PluginStepConditions struct {
	tableName           struct{} `sql:"plugin_step_conditions" pg:",discard_unknown_columns"`
	Id                  int      `sql:"id,pk"`
	PluginStepId        int      `sql:"plugin_step_id"`
	ConditionVariableId int      `sql:"condition_variable_id"` //id of variable on which condition is written
	ConditionType       string   `sql:"condition_type"`        //SKIP, TRIGGER, SUCCESS or FAILURE
	ConditionalOperator string   `sql:"conditional_operator"`
	ConditionalValue    string   `sql:"conditional_value"`
	Deleted             bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

type GlobalPluginRepository interface {
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

func (impl *GlobalPluginRepositoryImpl) GetAllPlugins() ([]*PluginMetadata, error) {
	var plugins []*PluginMetadata
	err := impl.dbConnection.Model(&plugins).
		Where("deleted = false").Select()
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
