package repository

import (
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
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
	tableName                struct{}                             `sql:"plugin_pipeline_script" pg:",discard_unknown_columns"`
	Id                       int                                  `sql:"id,pk"`
	Script                   string                               `sql:"name"`
	StoreScriptAt            string                               `sql:"store_script_at"`
	Type                     repository.ScriptType                `sql:"type"`
	DockerfileExists         bool                                 `sql:"dockerfile_exists, notnull"`
	MountPath                string                               `sql:"mount_path"`
	MountCodeToContainer     bool                                 `sql:"mount_code_to_container,notnull"`
	MountCodeToContainerPath string                               `sql:"mount_code_to_container_path"`
	ConfigureMountPath       bool                                 `sql:"configure_mount_path,notnull"`
	ContainerImagePath       string                               `sql:"container_image_path"`
	ImagePullSecretType      repository.ScriptImagePullSecretType `sql:"image_pull_secret_type"`
	ImagePullSecret          string                               `sql:"image_pull_secret"`
	Deleted                  bool                                 `sql:"deleted, notnull"`
	sql.AuditLog
}

type ScriptPathArgPortMapping struct {
	tableName           struct{}                     `sql:"script_path_arg_port_mapping" pg:",discard_unknown_columns"`
	Id                  int                          `sql:"id,pk"`
	TypeOfMapping       repository.ScriptMappingType `sql:"type_of_mapping"`
	FilePathOnDisk      string                       `sql:"file_path_on_disk"`
	FilePathOnContainer string                       `sql:"file_path_on_container"`
	Command             string                       `sql:"command"`
	Args                string                       `sql:"args"`
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
	CreateCiStage(ciStage *PipelineStage) (*PipelineStage, error)
	UpdateCiStage(ciStage *PipelineStage) (*PipelineStage, error)
	MarkCiStageDeletedById(ciStageId int, updatedBy int32, tx *pg.Tx) error
	GetAllCiStagesByCiPipelineId(ciPipelineId int) ([]*PipelineStage, error)
	GetCiStageByCiPipelineIdAndStageType(ciPipelineId int, stageType PipelineStageType) (*PipelineStage, error)

	GetStepIdsByStageId(stageId int) ([]int, error)
	CreatePipelineStageStep(step *PipelineStageStep) (*PipelineStageStep, error)
	UpdatePipelineStageStep(step *PipelineStageStep) (*PipelineStageStep, error)
	MarkCiStageStepsDeletedByStageId(ciStageId int, updatedBy int32, tx *pg.Tx) error
	GetAllStepsByStageId(stageId int) ([]*PipelineStageStep, error)
	GetStepById(stepId int) (*PipelineStageStep, error)
	MarkStepsDeletedExcludingActiveStepsInUpdateReq(activeStepIdsPresentInReq []int) error

	CreatePipelineScript(pipelineScript *PluginPipelineScript) (*PluginPipelineScript, error)
	UpdatePipelineScript(pipelineScript *PluginPipelineScript) (*PluginPipelineScript, error)
	MarkPipelineScriptsDeletedByStageId(ciStageId int, updatedBy int32, tx *pg.Tx) error
	GetScriptDetailById(id int) (*PluginPipelineScript, error)
	MarkScriptDeletedById(scriptId int) error

	MarkScriptMappingDeletedByScriptId(scriptId int) error
	CreateScriptMapping(mappings []ScriptPathArgPortMapping) error
	MarkPipelineScriptMappingsDeletedByStageId(ciStageId int, updatedBy int32, tx *pg.Tx) error
	GetScriptMappingDetailByScriptId(scriptId int) ([]*ScriptPathArgPortMapping, error)
	CheckIfFilePathMappingExists(filePathOnDisk string, filePathOnContainer string, scriptId int) (bool, error)
	CheckIfCommandArgMappingExists(command string, arg string, scriptId int) (bool, error)
	CheckIfPortMappingExists(portOnLocal int, portOnContainer int, scriptId int) (bool, error)

	CreatePipelineStageStepVariables([]PipelineStageStepVariable) ([]PipelineStageStepVariable, error)
	UpdatePipelineStageStepVariables(variables []PipelineStageStepVariable) ([]PipelineStageStepVariable, error)
	MarkPipelineStageStepVariablesDeletedByStageId(ciStageId int, updatedBy int32, tx *pg.Tx) error
	GetVariablesByStepId(stepId int) ([]*PipelineStageStepVariable, error)
	GetVariableIdsByStepId(stepId int) ([]int, error)
	MarkVariablesDeletedExcludingActiveVariablesInUpdateReq(activeVariableIdsPresentInReq []int) error

	CreatePipelineStageStepConditions([]PipelineStageStepCondition) ([]PipelineStageStepCondition, error)
	UpdatePipelineStageStepConditions(conditions []PipelineStageStepCondition) ([]PipelineStageStepCondition, error)
	MarkPipelineStageStepConditionDeletedByStageId(ciStageId int, updatedBy int32, tx *pg.Tx) error
	GetConditionsByVariableId(variableId int) ([]*PipelineStageStepCondition, error)
	GetConditionIdsByStepId(stepId int) ([]int, error)
	MarkConditionsDeletedExcludingActiveVariablesInUpdateReq(activeConditionIdsPresentInReq []int) error
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

func (impl *PipelineStageRepositoryImpl) GetCiStageByCiPipelineIdAndStageType(ciPipelineId int, stageType PipelineStageType) (*PipelineStage, error) {
	var pipelineStage PipelineStage
	err := impl.dbConnection.Model(&pipelineStage).
		Where("ci_pipeline_id = ?", ciPipelineId).
		Where("type = ?", stageType).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting ci stage by ciPipelineId", "err", err, "ciPipelineId", ciPipelineId)
		return nil, err
	}
	return &pipelineStage, nil
}

func (impl *PipelineStageRepositoryImpl) CreateCiStage(ciStage *PipelineStage) (*PipelineStage, error) {
	err := impl.dbConnection.Insert(ciStage)
	if err != nil {
		impl.logger.Errorw("error in creating pre stage entry", "err", err, "ciStage", ciStage)
		return nil, err
	}
	return ciStage, nil
}

func (impl *PipelineStageRepositoryImpl) UpdateCiStage(ciStage *PipelineStage) (*PipelineStage, error) {
	err := impl.dbConnection.Update(ciStage)
	if err != nil {
		impl.logger.Errorw("error in updating pre stage entry", "err", err, "ciStage", ciStage)
		return nil, err
	}
	return ciStage, nil
}

func (impl *PipelineStageRepositoryImpl) MarkCiStageDeletedById(ciStageId int, updatedBy int32, tx *pg.Tx) error {
	var stage PipelineStage
	_, err := tx.Model(&stage).Set("deleted = ?", true).Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", updatedBy).Where("id = ?", ciStageId).Update()
	if err != nil {
		impl.logger.Errorw("error in marking ci stage deleted", "err", err, "ciStageId", ciStageId)
		return err
	}
	return nil
}

func (impl *PipelineStageRepositoryImpl) GetStepIdsByStageId(stageId int) ([]int, error) {
	var ids []int
	query := "SELECT pss.id from plugin_stage_step pss where pss.pipeline_stage_id = ? and pss.deleted = false"
	_, err := impl.dbConnection.Query(&ids, query, stageId)
	if err != nil {
		impl.logger.Errorw("err in getting stepIds by stageId", "err", err, "stageId", stageId)
		return nil, err
	}
	return ids, nil
}

func (impl *PipelineStageRepositoryImpl) CreatePipelineStageStep(step *PipelineStageStep) (*PipelineStageStep, error) {
	err := impl.dbConnection.Insert(step)
	if err != nil {
		impl.logger.Errorw("error in creating pipeline stage step", "err", err, "step", step)
		return nil, err
	}
	return step, nil
}

func (impl *PipelineStageRepositoryImpl) UpdatePipelineStageStep(step *PipelineStageStep) (*PipelineStageStep, error) {
	err := impl.dbConnection.Update(step)
	if err != nil {
		impl.logger.Errorw("error in updating pipeline stage step", "err", err, "step", step)
		return nil, err
	}
	return step, nil
}

func (impl *PipelineStageRepositoryImpl) MarkCiStageStepsDeletedByStageId(ciStageId int, updatedBy int32, tx *pg.Tx) error {
	var step PipelineStageStep
	_, err := tx.Model(&step).Set("deleted = ?", true).Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", updatedBy).Where("pipeline_stage_id = ?", ciStageId).Update()
	if err != nil {
		impl.logger.Errorw("error in marking steps deleted by stageId", "err", err, "ciStageId", ciStageId)
		return err
	}
	return nil
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

func (impl *PipelineStageRepositoryImpl) GetStepById(stepId int) (*PipelineStageStep, error) {
	var step PipelineStageStep
	err := impl.dbConnection.Model(&step).
		Where("id = ?", stepId).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting step by Id", "err", err, "stepId", stepId)
		return nil, err
	}
	return &step, nil
}

func (impl *PipelineStageRepositoryImpl) MarkStepsDeletedExcludingActiveStepsInUpdateReq(activeStepIdsPresentInReq []int) error {
	var step PipelineStageStep
	_, err := impl.dbConnection.Model(&step).Set("deleted = ?", true).
		Where("id in (?)", pg.In(activeStepIdsPresentInReq)).Update()
	if err != nil {
		impl.logger.Errorw("error in deleting steps by excluding active steps in update req", "err", err, "activeStepIdsPresentInReq", activeStepIdsPresentInReq)
		return err
	}
	return nil
}

func (impl *PipelineStageRepositoryImpl) CreatePipelineScript(pipelineScript *PluginPipelineScript) (*PluginPipelineScript, error) {
	err := impl.dbConnection.Insert(pipelineScript)
	if err != nil {
		impl.logger.Errorw("error in creating pipeline script", "err", err, "scriptEntry", pipelineScript)
		return nil, err
	}
	return pipelineScript, nil
}

func (impl *PipelineStageRepositoryImpl) UpdatePipelineScript(pipelineScript *PluginPipelineScript) (*PluginPipelineScript, error) {
	err := impl.dbConnection.Update(pipelineScript)
	if err != nil {
		impl.logger.Errorw("error in updating pipeline script", "err", err, "scriptEntry", pipelineScript)
		return nil, err
	}
	return pipelineScript, nil
}

func (impl *PipelineStageRepositoryImpl) MarkPipelineScriptsDeletedByStageId(ciStageId int, updatedBy int32, tx *pg.Tx) error {
	var script PluginPipelineScript
	_, err := tx.Model(&script).
		Join("INNER JOIN pipeline_stage_step pss ON pss.script_id = plugin_pipeline_script.id").
		Join("INNER JOIN pipeline_stage ps ON ps.id = pss.pipeline_stage_id").
		Set("plugin_pipeline_script.deleted = ?", true).
		Set("plugin_pipeline_script.updated_on = ?", time.Now()).
		Set("plugin_pipeline_script.updated_by = ?", updatedBy).
		Where("ps.id = ?", ciStageId).Update()
	if err != nil {
		impl.logger.Errorw("error in marking scripts deleted by stageId", "err", err, "ciStageId", ciStageId)
		return err
	}
	return nil
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

func (impl *PipelineStageRepositoryImpl) MarkScriptDeletedById(scriptId int) error {
	var script PluginPipelineScript
	_, err := impl.dbConnection.Model(&script).Set("deleted = ?", true).
		Where("id = ?", scriptId).Update()
	if err != nil {
		impl.logger.Errorw("error in marking script deleted", "err", err, "scriptId", scriptId)
		return err
	}
	return nil
}

func (impl *PipelineStageRepositoryImpl) MarkScriptMappingDeletedByScriptId(scriptId int) error {
	var scriptMapping ScriptPathArgPortMapping
	_, err := impl.dbConnection.Model(&scriptMapping).Set("deleted = ?", true).
		Where("script_id = ?", scriptId).Update()
	if err != nil {
		impl.logger.Errorw("error in marking script mappings deleted", "err", err, "scriptId", scriptId)
		return err
	}
	return nil
}

func (impl *PipelineStageRepositoryImpl) CreateScriptMapping(mappings []ScriptPathArgPortMapping) error {
	err := impl.dbConnection.Insert(&mappings)
	if err != nil {
		impl.logger.Errorw("error in creating pipeline script mappings", "err", err, "mappings", mappings)
		return err
	}
	return nil
}

func (impl *PipelineStageRepositoryImpl) MarkPipelineScriptMappingsDeletedByStageId(ciStageId int, updatedBy int32, tx *pg.Tx) error {
	var mapping ScriptPathArgPortMapping
	_, err := tx.Model(&mapping).
		Join("INNER JOIN plugin_pipeline_script pps ON pps.id = script_path_arg_port_mapping.script_id").
		Join("INNER JOIN pipeline_stage_step pss ON pss.script_id = pps.id").
		Join("INNER JOIN pipeline_stage ps ON ps.id = pss.pipeline_stage_id").
		Set("script_path_arg_port_mapping.deleted = ?", true).
		Set("script_path_arg_port_mapping.updated_on = ?", time.Now()).
		Set("script_path_arg_port_mapping.updated_by = ?", updatedBy).
		Where("ps.id = ?", ciStageId).Update()
	if err != nil {
		impl.logger.Errorw("error in marking script mappings deleted by stageId", "err", err, "ciStageId", ciStageId)
		return err
	}
	return nil
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

func (impl *PipelineStageRepositoryImpl) CheckIfFilePathMappingExists(filePathOnDisk string, filePathOnContainer string, scriptId int) (bool, error) {
	var scriptMappingDetail ScriptPathArgPortMapping
	ifExists, err := impl.dbConnection.Model(&scriptMappingDetail).
		Where("file_path_on_disk = ?", filePathOnDisk).Where("file_path_on_container = ?", filePathOnContainer).
		Where("script_id = ?", scriptId).Where("deleted = ?", false).Exists()
	if err != nil {
		impl.logger.Errorw("err in checking if file path mapping exists or not", "err", err, "scriptId", scriptId)
		return false, err
	}
	return ifExists, nil
}

func (impl *PipelineStageRepositoryImpl) CheckIfCommandArgMappingExists(command string, arg string, scriptId int) (bool, error) {
	var scriptMappingDetail ScriptPathArgPortMapping
	ifExists, err := impl.dbConnection.Model(&scriptMappingDetail).
		Where("command = ?", command).Where("arg = ?", arg).
		Where("script_id = ?", scriptId).Where("deleted = ?", false).Exists()
	if err != nil {
		impl.logger.Errorw("err in checking if docker arg mapping exists or not", "err", err, "scriptId", scriptId)
		return false, err
	}
	return ifExists, nil
}

func (impl *PipelineStageRepositoryImpl) CheckIfPortMappingExists(portOnLocal int, portOnContainer int, scriptId int) (bool, error) {
	var scriptMappingDetail ScriptPathArgPortMapping
	ifExists, err := impl.dbConnection.Model(&scriptMappingDetail).
		Where("port_on_local = ?", portOnLocal).Where("port_on_container = ?", portOnContainer).
		Where("script_id = ?", scriptId).Where("deleted = ?", false).Exists()
	if err != nil {
		impl.logger.Errorw("err in checking if port mapping exists or not", "err", err, "scriptId", scriptId)
		return false, err
	}
	return ifExists, nil
}

func (impl *PipelineStageRepositoryImpl) CreatePipelineStageStepVariables(variables []PipelineStageStepVariable) ([]PipelineStageStepVariable, error) {
	err := impl.dbConnection.Insert(&variables)
	if err != nil {
		impl.logger.Errorw("error in creating pipeline stage step variables", "err", err, "variables", variables)
		return variables, err
	}
	return variables, nil
}

func (impl *PipelineStageRepositoryImpl) UpdatePipelineStageStepVariables(variables []PipelineStageStepVariable) ([]PipelineStageStepVariable, error) {
	err := impl.dbConnection.Update(&variables)
	if err != nil {
		impl.logger.Errorw("error in updating pipeline stage step variables", "err", err, "variables", variables)
		return variables, err
	}
	return variables, nil
}

func (impl *PipelineStageRepositoryImpl) MarkPipelineStageStepVariablesDeletedByStageId(ciStageId int, updatedBy int32, tx *pg.Tx) error {
	var variable PipelineStageStepVariable
	_, err := tx.Model(&variable).
		Join("INNER JOIN pipeline_stage_step pss ON pss.id = pipeline_stage_step_variable.pipeline_stage_step_id").
		Join("INNER JOIN pipeline_stage ps ON ps.id = pss.pipeline_stage_id").
		Set("pipeline_stage_step_variable.deleted = ?", true).
		Set("pipeline_stage_step_variable.updated_on = ?", time.Now()).
		Set("pipeline_stage_step_variable.updated_by = ?", updatedBy).
		Where("ps.id = ?", ciStageId).Update()
	if err != nil {
		impl.logger.Errorw("error in marking pipeline stage step variables deleted by stageId", "err", err, "ciStageId", ciStageId)
		return err
	}
	return nil
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

func (impl *PipelineStageRepositoryImpl) GetVariableIdsByStepId(stepId int) ([]int, error) {
	var ids []int
	query := "SELECT pssv.id from pipeline_stage_step_variable pssv where pssv.pipeline_stage_step_id = ? and pssv.deleted = false"
	_, err := impl.dbConnection.Query(&ids, query, stepId)
	if err != nil {
		impl.logger.Errorw("err in getting variableIds by stepId", "err", err, "stepId", stepId)
		return nil, err
	}
	return ids, nil
}

func (impl *PipelineStageRepositoryImpl) MarkVariablesDeletedExcludingActiveVariablesInUpdateReq(activeVariableIdsPresentInReq []int) error {
	var variable PipelineStageStepVariable
	_, err := impl.dbConnection.Model(&variable).Set("deleted = ?", true).
		Where("id in (?)", pg.In(activeVariableIdsPresentInReq)).Update()
	if err != nil {
		impl.logger.Errorw("error in deleting variables by excluding active variables in update req", "err", err, "activeVariableIdsPresentInReq", activeVariableIdsPresentInReq)
		return err
	}
	return nil
}

func (impl *PipelineStageRepositoryImpl) CreatePipelineStageStepConditions(conditions []PipelineStageStepCondition) ([]PipelineStageStepCondition, error) {
	err := impl.dbConnection.Insert(&conditions)
	if err != nil {
		impl.logger.Errorw("error in creating pipeline stage step conditions", "err", err, "conditions", conditions)
		return conditions, err
	}
	return conditions, nil
}

func (impl *PipelineStageRepositoryImpl) UpdatePipelineStageStepConditions(conditions []PipelineStageStepCondition) ([]PipelineStageStepCondition, error) {
	err := impl.dbConnection.Update(&conditions)
	if err != nil {
		impl.logger.Errorw("error in updating pipeline stage step conditions", "err", err, "conditions", conditions)
		return conditions, err
	}
	return conditions, nil
}

func (impl *PipelineStageRepositoryImpl) MarkPipelineStageStepConditionDeletedByStageId(ciStageId int, updatedBy int32, tx *pg.Tx) error {
	var condition PipelineStageStepCondition
	_, err := tx.Model(&condition).
		Join("INNER JOIN pipeline_stage_step pss ON pss.id = pipeline_stage_step_condition.pipeline_stage_step_id").
		Join("INNER JOIN pipeline_stage ps ON ps.id = pss.pipeline_stage_id").
		Set("pipeline_stage_step_condition.deleted = ?", true).
		Set("pipeline_stage_step_condition.updated_on = ?", time.Now()).
		Set("pipeline_stage_step_condition.updated_by = ?", updatedBy).
		Where("ps.id = ?", ciStageId).Update()
	if err != nil {
		impl.logger.Errorw("error in marking pipeline stage step conditions deleted by stageId", "err", err, "ciStageId", ciStageId)
		return err
	}
	return nil
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

func (impl *PipelineStageRepositoryImpl) GetConditionIdsByStepId(stepId int) ([]int, error) {
	var ids []int
	query := "SELECT pssc.id from pipeline_stage_step_condition pssc where pssc.pipeline_stage_step_id = ? and pssc.deleted = false"
	_, err := impl.dbConnection.Query(&ids, query, stepId)
	if err != nil {
		impl.logger.Errorw("err in getting conditionIds by stepId", "err", err, "stepId", stepId)
		return nil, err
	}
	return ids, nil
}

func (impl *PipelineStageRepositoryImpl) MarkConditionsDeletedExcludingActiveVariablesInUpdateReq(activeConditionIdsPresentInReq []int) error {
	var condition PipelineStageStepCondition
	_, err := impl.dbConnection.Model(&condition).Set("deleted = ?", true).
		Where("id in (?)", pg.In(activeConditionIdsPresentInReq)).Update()
	if err != nil {
		impl.logger.Errorw("error in deleting variables by excluding active variables in update req", "err", err, "activeConditionIdsPresentInReq", activeConditionIdsPresentInReq)
		return err
	}
	return nil
}
