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
type PluginStepVariableFormatType string

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
	CI            = 0
	CD            = 1
	CI_CD         = 2
	CD_STAGE_TYPE = "cd"
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
	sql.AuditLog
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
	GetMetaDataForAllPlugins(stageType int) ([]*PluginMetadata, error)
	GetMetaDataByPluginId(pluginId int) (*PluginMetadata, error)
	GetAllPluginTags() ([]*PluginTag, error)
	GetAllPluginTagRelations() ([]*PluginTagRelation, error)
	GetTagsByPluginId(pluginId int) ([]string, error)
	GetScriptDetailById(id int) (*PluginPipelineScript, error)
	GetScriptMappingDetailByScriptId(scriptId int) ([]*ScriptPathArgPortMapping, error)
	GetVariablesByStepId(stepId int) ([]*PluginStepVariable, error)
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

	UpdatePluginMetadata(pluginMetadata *PluginMetadata, tx *pg.Tx) error
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

func (impl *GlobalPluginRepositoryImpl) GetMetaDataForAllPlugins(stageType int) ([]*PluginMetadata, error) {
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

func (impl *GlobalPluginRepositoryImpl) GetTagsByPluginId(pluginId int) ([]string, error) {
	var tags []string
	query := "SELECT pt.name from plugin_tag pt INNER JOIN plugin_tag_relation ptr on pt.id = ptr.tag_id where ptr.plugin_id = ? and pt.deleted = false"
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
