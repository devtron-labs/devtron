package repository

import (
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type PipelineStageType string
type PipelineStepType string
type PipelineStageStepVariableType string
type PipelineStageStepVariableValueType string
type PipelineStageStepConditionType string

const (
	PIPELINE_STAGE_TYPE_PRE_CI                       PipelineStageType                  = "PRE_CI"
	PIPELINE_STAGE_TYPE_POST_CI                      PipelineStageType                  = "POST_CI"
	PIPELINE_STAGE_TYPE_PRE_CD                       PipelineStageType                  = "PRE_CD"
	PIPELINE_STAGE_TYPE_POST_CD                      PipelineStageType                  = "POST_CD"
	PIPELINE_STEP_TYPE_INLINE                        PipelineStepType                   = "INLINE"
	PIPELINE_STEP_TYPE_REF_PLUGIN                    PipelineStepType                   = "REF_PLUGIN"
	PIPELINE_STAGE_STEP_VARIABLE_TYPE_INPUT          PipelineStageStepVariableType      = "INPUT"
	PIPELINE_STAGE_STEP_VARIABLE_TYPE_OUTPUT         PipelineStageStepVariableType      = "OUTPUT"
	PIPELINE_STAGE_STEP_VARIABLE_VALUE_TYPE_NEW      PipelineStageStepVariableValueType = "NEW"
	PIPELINE_STAGE_STEP_VARIABLE_VALUE_TYPE_PREVIOUS PipelineStageStepVariableValueType = "FROM_PREVIOUS_STEP"
	PIPELINE_STAGE_STEP_VARIABLE_VALUE_TYPE_GLOBAL   PipelineStageStepVariableValueType = "GLOBAL"
	PIPELINE_STAGE_STEP_CONDITION_TYPE_SKIP          PipelineStageStepConditionType     = "SKIP"
	PIPELINE_STAGE_STEP_CONDITION_TYPE_TRIGGER       PipelineStageStepConditionType     = "TRIGGER"
	PIPELINE_STAGE_STEP_CONDITION_TYPE_SUCCESS       PipelineStageStepConditionType     = "SUCCESS"
	PIPELINE_STAGE_STEP_CONDITION_TYPE_FAIL          PipelineStageStepConditionType     = "FAIL"
)

type PipelineStage struct {
	tableName    struct{}          `sql:"pipeline_stage" pg:",discard_unknown_columns"`
	Id           int               `sql:"id,pk"`
	Name         string            `sql:"name"`
	Description  string            `sql:"description"`
	Type         PipelineStageType `sql:"type"`
	Deleted      bool              `sql:"deleted, notnull"`
	CiPipelineId int               `sql:"ci_pipeline_id"`
	CdPipelineId int               `sql:"cd_pipeline_id"`
	sql.AuditLog
}

type PipelineStageStep struct {
	tableName           struct{}         `sql:"pipeline_stage_step" pg:",discard_unknown_columns"`
	Id                  int              `sql:"id,pk"`
	PipelineStageId     int              `sql:"pipeline_stage_id"`
	Name                string           `sql:"name"`
	Description         string           `sql:"description"`
	Index               int              `sql:"index"`
	StepType            PipelineStepType `sql:"step_type"`
	ScriptId            int              `sql:"script_id"`
	RefPluginId         int              `sql:"ref_plugin_id"` //id of plugin used as reference
	ReportDirectoryPath string           `sql:"report_directory_path"`
	Deleted             bool             `sql:"deleted,notnull"`
	sql.AuditLog
}

// Below two tables are used at plugin-steps level too

type PluginPipelineScript struct {
	tableName            struct{}                             `sql:"plugin_pipeline_script" pg:",discard_unknown_columns"`
	Id                   int                                  `sql:"id,pk"`
	Script               string                               `sql:"name"`
	Type                 repository.ScriptType                `sql:"type"`
	DockerfileExists     bool                                 `sql:"dockerfile_exists, notnull"`
	StoreScriptAt        string                               `sql:"store_script_at"`
	MountPath            string                               `sql:"mount_path"`
	MountCodeToContainer bool                                 `sql:"mount_code_to_container,notnull"`
	ConfigureMountPath   bool                                 `sql:"configure_mount_path,notnull"`
	ContainerImagePath   string                               `sql:"container_image_path"`
	ImagePullSecretType  repository.ScriptImagePullSecretType `sql:"image_pull_secret_type"`
	ImagePullSecret      string                               `sql:"image_pull_secret"`
	Deleted              bool                                 `sql:"deleted, notnull"`
	sql.AuditLog
}

type ScriptPathArgPortMapping struct {
	tableName           struct{}                     `sql:"script_path_arg_port_mapping" pg:",discard_unknown_columns"`
	Id                  int                          `sql:"id,pk"`
	TypeOfMapping       repository.ScriptMappingType `sql:"type_of_mapping"`
	FilePathOnDisk      string                       `sql:"file_path_on_disk"`
	FilePathOnContainer string                       `sql:"file_path_on_container"`
	Command             string                       `sql:"command"`
	Arg                 string                       `sql:"arg"`
	PortOnLocal         int                          `sql:"port_on_local"`
	PortOnContainer     int                          `sql:"port_on_container"`
	ScriptId            int                          `sql:"script_id"`
	Deleted             bool                         `sql:"deleted, notnull"`
	sql.AuditLog
}

type PipelineStageStepVariable struct {
	tableName             struct{}                           `sql:"pipeline_stage_step_variable" pg:",discard_unknown_columns"`
	Id                    int                                `sql:"id,pk"`
	PipelineStageStepId   int                                `sql:"pipeline_stage_step_id"`
	Name                  string                             `sql:"name"`
	Format                string                             `sql:"format"`
	Description           string                             `sql:"description"`
	IsExposed             bool                               `sql:"is_exposed,notnull"`
	AllowEmptyValue       bool                               `sql:"allow_empty_value,notnull"`
	DefaultValue          string                             `sql:"default_value"`
	Value                 string                             `sql:"value"`
	VariableType          PipelineStageStepVariableType      `sql:"variable_type"`
	ValueType             PipelineStageStepVariableValueType `sql:"value_type"`
	PreviousStepIndex     int                                `sql:"previous_step_index"`
	ReferenceVariableName string                             `sql:"reference_variable_name"`
	Deleted               bool                               `sql:"deleted,notnull"`
	sql.AuditLog
}

type PipelineStageStepCondition struct {
	tableName           struct{}                       `sql:"pipeline_stage_step_condition" pg:",discard_unknown_columns"`
	Id                  int                            `sql:"id,pk"`
	PipelineStageStepId int                            `sql:"pipeline_stage_step_id"`
	ConditionVariableId int                            `sql:"condition_variable_id"` //id of variable on which condition is written
	ConditionType       PipelineStageStepConditionType `sql:"condition_type"`
	ConditionalOperator string                         `sql:"conditional_operator"`
	ConditionalValue    string                         `sql:"conditional_value"`
	Deleted             bool                           `sql:"deleted,notnull"`
	sql.AuditLog
}

type PipelineStageRepository interface {
	GetAllCiStagesByCiPipelineId(ciPipelineId int) ([]*PipelineStage, error)
	GetAllStepsByStageId(stageId int) ([]*PipelineStageStep, error)
	GetScriptDetailById(id int) (*PluginPipelineScript, error)
	GetScriptMappingDetailByScriptId(scriptId int) ([]*ScriptPathArgPortMapping, error)
	GetVariablesByStepId(stepId int) ([]*PipelineStageStepVariable, error)
	GetConditionsByVariableId(variableId int) ([]*PipelineStageStepCondition, error)
}

func NewPipelineStageRepository(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *PipelineStageRepositoryImpl {
	return &PipelineStageRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type PipelineStageRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func (impl *PipelineStageRepositoryImpl) GetAllCiStagesByCiPipelineId(ciPipelineId int) ([]*PipelineStage, error) {
	var pipelineStages []*PipelineStage
	err := impl.dbConnection.Model(&pipelineStages).
		Where("ci_pipeline_id = ?", ciPipelineId).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting all ci stages by ciPipelineId", "err", err, "ciPipelineId", ciPipelineId)
		return nil, err
	}
	return pipelineStages, nil
}

func (impl *PipelineStageRepositoryImpl) GetAllStepsByStageId(stageId int) ([]*PipelineStageStep, error) {
	var steps []*PipelineStageStep
	err := impl.dbConnection.Model(&steps).
		Where("pipeline_stage_id = ?", stageId).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting all steps by stageId", "err", err, "stageId", stageId)
		return nil, err
	}
	return steps, nil
}

func (impl *PipelineStageRepositoryImpl) GetScriptDetailById(id int) (*PluginPipelineScript, error) {
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

func (impl *PipelineStageRepositoryImpl) GetScriptMappingDetailByScriptId(scriptId int) ([]*ScriptPathArgPortMapping, error) {
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

func (impl *PipelineStageRepositoryImpl) GetVariablesByStepId(stepId int) ([]*PipelineStageStepVariable, error) {
	var variables []*PipelineStageStepVariable
	err := impl.dbConnection.Model(&variables).
		Where("pipeline_stage_step_id = ?", stepId).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting variables by stepId", "err", err, "stepId", stepId)
		return nil, err
	}
	return variables, nil
}

func (impl *PipelineStageRepositoryImpl) GetConditionsByVariableId(variableId int) ([]*PipelineStageStepCondition, error) {
	var conditions []*PipelineStageStepCondition
	err := impl.dbConnection.Model(&conditions).
		Where("condition_variable_id = ?", variableId).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting variables by stepId", "err", err, "variableId", variableId)
		return nil, err
	}
	return conditions, nil
}
