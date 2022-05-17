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
type PipelineStageStepVariableFormatType string

const (
	PIPELINE_STAGE_TYPE_PRE_CI                       PipelineStageType                   = "PRE_CI"
	PIPELINE_STAGE_TYPE_POST_CI                      PipelineStageType                   = "POST_CI"
	PIPELINE_STAGE_TYPE_PRE_CD                       PipelineStageType                   = "PRE_CD"
	PIPELINE_STAGE_TYPE_POST_CD                      PipelineStageType                   = "POST_CD"
	PIPELINE_STEP_TYPE_INLINE                        PipelineStepType                    = "INLINE"
	PIPELINE_STEP_TYPE_REF_PLUGIN                    PipelineStepType                    = "REF_PLUGIN"
	PIPELINE_STAGE_STEP_VARIABLE_TYPE_INPUT          PipelineStageStepVariableType       = "INPUT"
	PIPELINE_STAGE_STEP_VARIABLE_TYPE_OUTPUT         PipelineStageStepVariableType       = "OUTPUT"
	PIPELINE_STAGE_STEP_VARIABLE_VALUE_TYPE_NEW      PipelineStageStepVariableValueType  = "NEW"
	PIPELINE_STAGE_STEP_VARIABLE_VALUE_TYPE_PREVIOUS PipelineStageStepVariableValueType  = "FROM_PREVIOUS_STEP"
	PIPELINE_STAGE_STEP_VARIABLE_VALUE_TYPE_GLOBAL   PipelineStageStepVariableValueType  = "GLOBAL"
	PIPELINE_STAGE_STEP_CONDITION_TYPE_SKIP          PipelineStageStepConditionType      = "SKIP"
	PIPELINE_STAGE_STEP_CONDITION_TYPE_TRIGGER       PipelineStageStepConditionType      = "TRIGGER"
	PIPELINE_STAGE_STEP_CONDITION_TYPE_SUCCESS       PipelineStageStepConditionType      = "SUCCESS"
	PIPELINE_STAGE_STEP_CONDITION_TYPE_FAIL          PipelineStageStepConditionType      = "FAIL"
	PIPELINE_STAGE_STEP_VARIABLE_FORMAT_TYPE_STRING  PipelineStageStepVariableFormatType = "STRING"
	PIPELINE_STAGE_STEP_VARIABLE_FORMAT_TYPE_NUMBER  PipelineStageStepVariableFormatType = "NUMBER"
	PIPELINE_STAGE_STEP_VARIABLE_FORMAT_TYPE_BOOL    PipelineStageStepVariableFormatType = "BOOL"
	PIPELINE_STAGE_STEP_VARIABLE_FORMAT_TYPE_DATE    PipelineStageStepVariableFormatType = "DATE"
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
	OutputDirectoryPath []string         `sql:"output_directory_path" pg:",array"`
	DependentOnStep     string           `sql:"dependent_on_step"`
	Deleted             bool             `sql:"deleted,notnull"`
	sql.AuditLog
}

// Below two tables are used at plugin-steps level too

type PluginPipelineScript struct {
	tableName                struct{}                             `sql:"plugin_pipeline_script" pg:",discard_unknown_columns"`
	Id                       int                                  `sql:"id,pk"`
	Script                   string                               `sql:"script"`
	StoreScriptAt            string                               `sql:"store_script_at"`
	Type                     repository.ScriptType                `sql:"type"`
	DockerfileExists         bool                                 `sql:"dockerfile_exists, notnull"`
	MountPath                string                               `sql:"mount_path"`
	MountCodeToContainer     bool                                 `sql:"mount_code_to_container,notnull"`
	MountCodeToContainerPath string                               `sql:"mount_code_to_container_path"`
	MountDirectoryFromHost   bool                                 `sql:"mount_directory_from_host,notnull"`
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
	Args                []string                     `sql:"args" pg:",array"`
	PortOnLocal         int                          `sql:"port_on_local"`
	PortOnContainer     int                          `sql:"port_on_container"`
	ScriptId            int                          `sql:"script_id"`
	Deleted             bool                         `sql:"deleted, notnull"`
	sql.AuditLog
}

type PipelineStageStepVariable struct {
	tableName                 struct{}                            `sql:"pipeline_stage_step_variable" pg:",discard_unknown_columns"`
	Id                        int                                 `sql:"id,pk"`
	PipelineStageStepId       int                                 `sql:"pipeline_stage_step_id"`
	Name                      string                              `sql:"name"`
	Format                    PipelineStageStepVariableFormatType `sql:"format"`
	Description               string                              `sql:"description"`
	IsExposed                 bool                                `sql:"is_exposed,notnull"`
	AllowEmptyValue           bool                                `sql:"allow_empty_value,notnull"`
	DefaultValue              string                              `sql:"default_value"`
	Value                     string                              `sql:"value"`
	VariableType              PipelineStageStepVariableType       `sql:"variable_type"`
	ValueType                 PipelineStageStepVariableValueType  `sql:"value_type"`
	PreviousStepIndex         int                                 `sql:"previous_step_index,type:integer"`
	VariableStepIndexInPlugin int                                 `sql:"variable_step_index_in_plugin,type:integer"`
	ReferenceVariableName     string                              `sql:"reference_variable_name,type:text"`
	ReferenceVariableStage    PipelineStageType                   `sql:"reference_variable_stage,type:text"`
	Deleted                   bool                                `sql:"deleted,notnull"`
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
	GetConnection() *pg.DB

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
	MarkStepsDeletedByStageId(stageId int) error
	MarkStepsDeletedExcludingActiveStepsInUpdateReq(activeStepIdsPresentInReq []int, stageId int) error

	CreatePipelineScript(pipelineScript *PluginPipelineScript) (*PluginPipelineScript, error)
	UpdatePipelineScript(pipelineScript *PluginPipelineScript) (*PluginPipelineScript, error)
	GetScriptIdsByStageId(stageId int) ([]int, error)
	MarkPipelineScriptsDeletedByIds(ids []int, updatedBy int32, tx *pg.Tx) error
	GetScriptDetailById(id int) (*PluginPipelineScript, error)
	MarkScriptDeletedById(scriptId int) error

	MarkScriptMappingDeletedByScriptId(scriptId int) error
	CreateScriptMapping(mappings []ScriptPathArgPortMapping) error
	GetScriptMappingIdsByStageId(stageId int) ([]int, error)
	MarkPipelineScriptMappingsDeletedByIds(ids []int, updatedBy int32, tx *pg.Tx) error
	GetScriptMappingDetailByScriptId(scriptId int) ([]*ScriptPathArgPortMapping, error)
	CheckIfFilePathMappingExists(filePathOnDisk string, filePathOnContainer string, scriptId int) (bool, error)
	CheckIfCommandArgMappingExists(command string, arg string, scriptId int) (bool, error)
	CheckIfPortMappingExists(portOnLocal int, portOnContainer int, scriptId int) (bool, error)

	CreatePipelineStageStepVariables([]PipelineStageStepVariable, *pg.Tx) ([]PipelineStageStepVariable, error)
	UpdatePipelineStageStepVariables(variables []PipelineStageStepVariable, tx *pg.Tx) ([]PipelineStageStepVariable, error)
	GetVariableIdsByStageId(stageId int) ([]int, error)
	MarkPipelineStageStepVariablesDeletedByIds(ids []int, updatedBy int32, tx *pg.Tx) error
	GetVariablesByStepId(stepId int) ([]*PipelineStageStepVariable, error)
	GetVariableIdsByStepIdAndVariableType(stepId int, variableType PipelineStageStepVariableType) ([]int, error)
	MarkVariablesDeletedByStepIdAndVariableType(stepId int, variableType PipelineStageStepVariableType, tx *pg.Tx) error
	MarkVariablesDeletedExcludingActiveVariablesInUpdateReq(activeVariableIdsPresentInReq []int, stepId int, variableType PipelineStageStepVariableType, tx *pg.Tx) error

	CreatePipelineStageStepConditions([]PipelineStageStepCondition, *pg.Tx) ([]PipelineStageStepCondition, error)
	UpdatePipelineStageStepConditions(conditions []PipelineStageStepCondition, tx *pg.Tx) ([]PipelineStageStepCondition, error)
	GetConditionIdsByStageId(stageId int) ([]int, error)
	MarkPipelineStageStepConditionDeletedByIds(ids []int, updatedBy int32, tx *pg.Tx) error
	GetConditionsByStepId(stepId int) ([]*PipelineStageStepCondition, error)
	GetConditionsByVariableId(variableId int) ([]*PipelineStageStepCondition, error)
	GetConditionIdsByStepId(stepId int) ([]int, error)
	MarkConditionsDeletedByStepId(stepId int, tx *pg.Tx) error
	MarkConditionsDeletedExcludingActiveVariablesInUpdateReq(activeConditionIdsPresentInReq []int, stepId int, tx *pg.Tx) error
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

func (impl *PipelineStageRepositoryImpl) GetConnection() *pg.DB {
	return impl.dbConnection
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
	query := "SELECT pss.id from pipeline_stage_step pss where pss.pipeline_stage_id = ? and pss.deleted = false"
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
		Where("deleted = ?", false).
		Order("index ASC").Select()
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

func (impl *PipelineStageRepositoryImpl) MarkStepsDeletedByStageId(stageId int) error {
	var step PipelineStageStep
	_, err := impl.dbConnection.Model(&step).Set("deleted = ?", true).
		Where("pipeline_stage_id = ?", stageId).Update()
	if err != nil {
		impl.logger.Errorw("error in deleting steps by stageId", "err", err, "stageId", stageId)
		return err
	}
	return nil
}

func (impl *PipelineStageRepositoryImpl) MarkStepsDeletedExcludingActiveStepsInUpdateReq(activeStepIdsPresentInReq []int, stageId int) error {
	var step PipelineStageStep
	_, err := impl.dbConnection.Model(&step).Set("deleted = ?", true).
		Where("pipeline_stage_id = ?", stageId).
		Where("id not in (?)", pg.In(activeStepIdsPresentInReq)).Update()
	if err != nil {
		impl.logger.Errorw("error in deleting steps by excluding active steps in update req", "err", err, "activeStepIdsPresentInReq", activeStepIdsPresentInReq, "stageId", stageId)
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

func (impl *PipelineStageRepositoryImpl) GetScriptIdsByStageId(stageId int) ([]int, error) {
	var ids []int
	query := "SELECT pps.id from plugin_pipeline_script pps INNER JOIN pipeline_stage_step pss ON pss.script_id = pps.id " +
		"INNER JOIN pipeline_stage ps ON ps.id = pss.pipeline_stage_id " +
		"WHERE ps.id = ? and pps.deleted=false;"
	_, err := impl.dbConnection.Query(&ids, query, stageId)
	if err != nil {
		impl.logger.Errorw("err in getting scriptIds by stageId", "err", err, "ids", ids)
		return nil, err
	}
	return ids, nil
}

func (impl *PipelineStageRepositoryImpl) MarkPipelineScriptsDeletedByIds(ids []int, updatedBy int32, tx *pg.Tx) error {
	var script PluginPipelineScript
	_, err := tx.Model(&script).
		Set("deleted = ?", true).
		Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", updatedBy).
		Where("id in (?)", pg.In(ids)).Update()
	if err != nil {
		impl.logger.Errorw("error in marking scripts deleted by ids", "err", err, "ids", ids)
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

func (impl *PipelineStageRepositoryImpl) GetScriptMappingIdsByStageId(stageId int) ([]int, error) {
	var ids []int
	query := "SELECT spapm.id from script_path_arg_port_mapping spapm INNER JOIN plugin_pipeline_script pps ON pps.id = spapm.script_id " +
		"INNER JOIN pipeline_stage_step pss ON pss.script_id = pps.id " +
		"INNER JOIN pipeline_stage ps ON ps.id = pss.pipeline_stage_id " +
		"WHERE ps.id = ? and spapm.deleted=false;"
	_, err := impl.dbConnection.Query(&ids, query, stageId)
	if err != nil {
		impl.logger.Errorw("err in getting scriptMappingIds by stageId", "err", err, "ids", ids)
		return nil, err
	}
	return ids, nil
}

func (impl *PipelineStageRepositoryImpl) MarkPipelineScriptMappingsDeletedByIds(ids []int, updatedBy int32, tx *pg.Tx) error {
	var mapping ScriptPathArgPortMapping
	_, err := tx.Model(&mapping).
		Set("deleted = ?", true).
		Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", updatedBy).
		Where("id in (?)", pg.In(ids)).Update()
	if err != nil {
		impl.logger.Errorw("error in marking script mappings deleted by ids", "err", err, "ids", ids)
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

func (impl *PipelineStageRepositoryImpl) CreatePipelineStageStepVariables(variables []PipelineStageStepVariable, tx *pg.Tx) ([]PipelineStageStepVariable, error) {
	err := tx.Insert(&variables)
	if err != nil {
		impl.logger.Errorw("error in creating pipeline stage step variables", "err", err, "variables", variables)
		return variables, err
	}
	return variables, nil
}

func (impl *PipelineStageRepositoryImpl) UpdatePipelineStageStepVariables(variables []PipelineStageStepVariable, tx *pg.Tx) ([]PipelineStageStepVariable, error) {
	_, err := tx.Model(&variables).Update()
	if err != nil {
		impl.logger.Errorw("error in updating pipeline stage step variables", "err", err, "variables", variables)
		return variables, err
	}
	return variables, nil
}

func (impl *PipelineStageRepositoryImpl) GetVariableIdsByStageId(stageId int) ([]int, error) {
	var ids []int
	query := "SELECT pssv.id from pipeline_stage_step_variable pssv INNER JOIN pipeline_stage_step pss ON pss.id = pssv.pipeline_stage_step_id " +
		"INNER JOIN pipeline_stage ps ON ps.id = pss.pipeline_stage_id " +
		"WHERE ps.id = ? and pssv.deleted=false;"
	_, err := impl.dbConnection.Query(&ids, query, stageId)
	if err != nil {
		impl.logger.Errorw("err in getting variableIds by stageId", "err", err, "stageId", stageId)
		return nil, err
	}
	return ids, nil
}

func (impl *PipelineStageRepositoryImpl) MarkPipelineStageStepVariablesDeletedByIds(ids []int, updatedBy int32, tx *pg.Tx) error {
	var variable PipelineStageStepVariable
	_, err := tx.Model(&variable).
		Set("deleted = ?", true).
		Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", updatedBy).
		Where("id in (?)", pg.In(ids)).Update()
	if err != nil {
		impl.logger.Errorw("error in marking pipeline stage step variables deleted by ids", "err", err, "ids", ids)
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

func (impl *PipelineStageRepositoryImpl) GetVariableIdsByStepIdAndVariableType(stepId int, variableType PipelineStageStepVariableType) ([]int, error) {
	var ids []int
	query := "SELECT pssv.id from pipeline_stage_step_variable pssv where pssv.pipeline_stage_step_id = ? and pssv.deleted = false and pssv.variable_type = ?;"
	_, err := impl.dbConnection.Query(&ids, query, stepId, variableType)
	if err != nil {
		impl.logger.Errorw("err in getting variableIds by stepId", "err", err, "stepId", stepId)
		return nil, err
	}
	return ids, nil
}

func (impl *PipelineStageRepositoryImpl) MarkVariablesDeletedByStepIdAndVariableType(stepId int, variableType PipelineStageStepVariableType, tx *pg.Tx) error {
	var variable PipelineStageStepVariable
	_, err := tx.Model(&variable).Set("deleted = ?", true).
		Where("pipeline_stage_step_id = ?", stepId).
		Where("variable_type = ?", variableType).Update()
	if err != nil {
		impl.logger.Errorw("error in deleting variables by stepId", "err", err, "stepId", stepId)
		return err
	}
	return nil
}

func (impl *PipelineStageRepositoryImpl) MarkVariablesDeletedExcludingActiveVariablesInUpdateReq(activeVariableIdsPresentInReq []int, stepId int, variableType PipelineStageStepVariableType, tx *pg.Tx) error {
	var variable PipelineStageStepVariable
	_, err := tx.Model(&variable).Set("deleted = ?", true).
		Where("pipeline_stage_step_id = ?", stepId).
		Where("id not in (?)", pg.In(activeVariableIdsPresentInReq)).
		Where("variable_type = ?", variableType).Update()
	if err != nil {
		impl.logger.Errorw("error in deleting variables by excluding active variables in update req", "err", err, "activeVariableIdsPresentInReq", activeVariableIdsPresentInReq)
		return err
	}
	return nil
}

func (impl *PipelineStageRepositoryImpl) CreatePipelineStageStepConditions(conditions []PipelineStageStepCondition, tx *pg.Tx) ([]PipelineStageStepCondition, error) {
	err := tx.Insert(&conditions)
	if err != nil {
		impl.logger.Errorw("error in creating pipeline stage step conditions", "err", err, "conditions", conditions)
		return conditions, err
	}
	return conditions, nil
}

func (impl *PipelineStageRepositoryImpl) UpdatePipelineStageStepConditions(conditions []PipelineStageStepCondition, tx *pg.Tx) ([]PipelineStageStepCondition, error) {
	_, err := tx.Model(&conditions).Update()
	if err != nil {
		impl.logger.Errorw("error in updating pipeline stage step conditions", "err", err, "conditions", conditions)
		return conditions, err
	}
	return conditions, nil
}

func (impl *PipelineStageRepositoryImpl) GetConditionIdsByStageId(stageId int) ([]int, error) {
	var ids []int
	query := "SELECT pssc.id from pipeline_stage_step_condition pssc INNER JOIN pipeline_stage_step pss ON pss.id = pssc.pipeline_stage_step_id " +
		"INNER JOIN pipeline_stage ps ON ps.id = pss.pipeline_stage_id " +
		"WHERE ps.id = ? and pssc.deleted=false;"
	_, err := impl.dbConnection.Query(&ids, query, stageId)
	if err != nil {
		impl.logger.Errorw("err in getting conditionIds by stageId", "err", err, "stageId", stageId)
		return nil, err
	}
	return ids, nil
}

func (impl *PipelineStageRepositoryImpl) MarkPipelineStageStepConditionDeletedByIds(ids []int, updatedBy int32, tx *pg.Tx) error {
	var condition PipelineStageStepCondition
	_, err := tx.Model(&condition).
		Set("deleted = ?", true).
		Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", updatedBy).
		Where("id in (?)", pg.In(ids)).Update()
	if err != nil {
		impl.logger.Errorw("error in marking pipeline stage step conditions deleted by ids", "err", err, "ids", ids)
		return err
	}
	return nil
}

func (impl *PipelineStageRepositoryImpl) GetConditionsByStepId(stepId int) ([]*PipelineStageStepCondition, error) {
	var conditions []*PipelineStageStepCondition
	err := impl.dbConnection.Model(&conditions).
		Where("pipeline_stage_step_id = ?", stepId).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting conditions by stepId", "err", err, "stepId", stepId)
		return nil, err
	}
	return conditions, nil
}

func (impl *PipelineStageRepositoryImpl) GetConditionsByVariableId(variableId int) ([]*PipelineStageStepCondition, error) {
	var conditions []*PipelineStageStepCondition
	err := impl.dbConnection.Model(&conditions).
		Where("condition_variable_id = ?", variableId).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting conditions by variableId", "err", err, "variableId", variableId)
		return nil, err
	}
	return conditions, nil
}

func (impl *PipelineStageRepositoryImpl) GetConditionIdsByStepId(stepId int) ([]int, error) {
	var ids []int
	query := "SELECT pssc.id from pipeline_stage_step_condition pssc where pssc.pipeline_stage_step_id = ? and pssc.deleted = false;"
	_, err := impl.dbConnection.Query(&ids, query, stepId)
	if err != nil {
		impl.logger.Errorw("err in getting conditionIds by stepId", "err", err, "stepId", stepId)
		return nil, err
	}
	return ids, nil
}

func (impl *PipelineStageRepositoryImpl) MarkConditionsDeletedByStepId(stepId int, tx *pg.Tx) error {
	var condition PipelineStageStepCondition
	_, err := tx.Model(&condition).Set("deleted = ?", true).
		Where("pipeline_stage_step_id = ?", stepId).Update()
	if err != nil {
		impl.logger.Errorw("error in deleting conditions by stepId", "err", err, "stepId", stepId)
		return err
	}
	return nil
}

func (impl *PipelineStageRepositoryImpl) MarkConditionsDeletedExcludingActiveVariablesInUpdateReq(activeConditionIdsPresentInReq []int, stepId int, tx *pg.Tx) error {
	var condition PipelineStageStepCondition
	_, err := tx.Model(&condition).Set("deleted = ?", true).
		Where("pipeline_stage_step_id = ?", stepId).
		Where("id not in (?)", pg.In(activeConditionIdsPresentInReq)).Update()
	if err != nil {
		impl.logger.Errorw("error in deleting conditions by excluding active conditions in update req", "err", err, "activeConditionIdsPresentInReq", activeConditionIdsPresentInReq)
		return err
	}
	return nil
}
