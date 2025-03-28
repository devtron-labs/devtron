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
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PipelineStageType string
type PipelineStepType string

func (p PipelineStepType) String() string {
	return string(p)
}

type PipelineStageStepVariableType string

func (p PipelineStageStepVariableType) IsOutput() bool {
	return p == PIPELINE_STAGE_STEP_VARIABLE_TYPE_OUTPUT
}

func (p PipelineStageStepVariableType) IsInput() bool {
	return p == PIPELINE_STAGE_STEP_VARIABLE_TYPE_INPUT
}

type PipelineStageStepVariableValueType string

func (p PipelineStageStepVariableValueType) String() string {
	return string(p)
}

func (p PipelineStageStepVariableValueType) IsGlobalDefinedValue() bool {
	return p == PIPELINE_STAGE_STEP_VARIABLE_VALUE_TYPE_GLOBAL
}

func (p PipelineStageStepVariableValueType) IsUserDefinedValue() bool {
	return p == PIPELINE_STAGE_STEP_VARIABLE_VALUE_TYPE_NEW
}

func (p PipelineStageStepVariableValueType) IsPreviousOutputDefinedValue() bool {
	return p == PIPELINE_STAGE_STEP_VARIABLE_VALUE_TYPE_PREVIOUS
}

type PipelineStageStepConditionType string

// PipelineStageStepVariableFormatType - Duplicate of repository.PluginStepVariableFormatType
// TODO: analyse if we can remove the duplicates.
// Should ideally be a subset of repository.PluginStepVariableFormatType
type PipelineStageStepVariableFormatType string

func (p PipelineStageStepVariableFormatType) String() string {
	return string(p)
}

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
	PIPELINE_STAGE_STEP_CONDITION_TYPE_SUCCESS       PipelineStageStepConditionType      = "PASS"
	PIPELINE_STAGE_STEP_CONDITION_TYPE_FAIL          PipelineStageStepConditionType      = "FAIL"
	PIPELINE_STAGE_STEP_VARIABLE_FORMAT_TYPE_STRING  PipelineStageStepVariableFormatType = "STRING"
	PIPELINE_STAGE_STEP_VARIABLE_FORMAT_TYPE_NUMBER  PipelineStageStepVariableFormatType = "NUMBER"
	PIPELINE_STAGE_STEP_VARIABLE_FORMAT_TYPE_BOOL    PipelineStageStepVariableFormatType = "BOOL"
	PIPELINE_STAGE_STEP_VARIABLE_FORMAT_TYPE_DATE    PipelineStageStepVariableFormatType = "DATE"
)

func (r PipelineStageType) ToString() string {
	return string(r)
}
func (r PipelineStageType) IsStageTypePreCi() bool {
	return r == PIPELINE_STAGE_TYPE_PRE_CI
}
func (r PipelineStageType) IsStageTypePreCd() bool {
	return r == PIPELINE_STAGE_TYPE_PRE_CD
}
func (r PipelineStageType) IsStageTypePostCi() bool {
	return r == PIPELINE_STAGE_TYPE_POST_CI
}
func (r PipelineStageType) IsStageTypePostCd() bool {
	return r == PIPELINE_STAGE_TYPE_POST_CD
}

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

func (ps *PipelineStage) IsPipelineStageExists() bool {
	if ps == nil {
		return false
	}
	// if id is not 0 then it exists; special case id is -1 for dummy stage.
	// Dummy stage is used if no stage is found for a pipeline but mandatory stage plugins is required.
	return ps.Id != 0
}

type PipelineStageStep struct {
	tableName                struct{}         `sql:"pipeline_stage_step" pg:",discard_unknown_columns"`
	Id                       int              `sql:"id,pk"`
	PipelineStageId          int              `sql:"pipeline_stage_id"`
	Name                     string           `sql:"name"`
	Description              string           `sql:"description"`
	Index                    int              `sql:"index"`
	StepType                 PipelineStepType `sql:"step_type"`
	ScriptId                 int              `sql:"script_id"`
	RefPluginId              int              `sql:"ref_plugin_id"` //id of plugin used as reference
	OutputDirectoryPath      []string         `sql:"output_directory_path" pg:",array"`
	DependentOnStep          string           `sql:"dependent_on_step"`
	Deleted                  bool             `sql:"deleted,notnull"`
	TriggerIfParentStageFail bool             `sql:"trigger_if_parent_stage_fail"`
	sql.AuditLog
}

// Below two tables are used at plugin-steps level too
// TODO: remove these duplicate tables definitions

// PluginPipelineScript is the duplicate declaration of repository.PluginPipelineScript
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

// ScriptPathArgPortMapping is the duplicate declaration of repository.ScriptPathArgPortMapping
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
	tableName           struct{}                            `sql:"pipeline_stage_step_variable" pg:",discard_unknown_columns"`
	Id                  int                                 `sql:"id,pk"`
	PipelineStageStepId int                                 `sql:"pipeline_stage_step_id"`
	Name                string                              `sql:"name"`
	Format              PipelineStageStepVariableFormatType `sql:"format"` // oneof: STRING, NUMBER, BOOL, DATE
	Description         string                              `sql:"description"`
	// IsExposed: has conflicting data in DB.
	// Ideally, for user given input variables, it should always be TRUE.
	// Also, if IsExposed is false, then it's an internal variable (should not be exposed in UI).
	// TODO: investigate the conflicting data in DB.
	IsExposed                 bool                               `sql:"is_exposed,notnull"`
	AllowEmptyValue           bool                               `sql:"allow_empty_value,notnull"`
	DefaultValue              string                             `sql:"default_value"`
	Value                     string                             `sql:"value"`
	VariableType              PipelineStageStepVariableType      `sql:"variable_type"`
	ValueType                 PipelineStageStepVariableValueType `sql:"value_type"`
	PreviousStepIndex         int                                `sql:"previous_step_index,type:integer"`
	VariableStepIndexInPlugin int                                `sql:"variable_step_index_in_plugin,type:integer"`
	ReferenceVariableName     string                             `sql:"reference_variable_name,type:text"`
	ReferenceVariableStage    PipelineStageType                  `sql:"reference_variable_stage,type:text"`
	Deleted                   bool                               `sql:"deleted,notnull"`
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

	CreatePipelineStage(pipelineStage *PipelineStage, tx *pg.Tx) (*PipelineStage, error)
	UpdatePipelineStage(pipelineStage *PipelineStage) (*PipelineStage, error)
	MarkPipelineStageDeletedById(stageId int, updatedBy int32, tx *pg.Tx) error

	GetAllCiStagesByCiPipelineId(ciPipelineId int) ([]*PipelineStage, error)
	GetAllCdStagesByCdPipelineId(cdPipelineId int) ([]*PipelineStage, error)
	GetAllCdStagesByCdPipelineIds(cdPipelineIds []int) ([]*PipelineStage, error)

	GetCiStageByCiPipelineIdAndStageType(ciPipelineId int, stageType PipelineStageType) (*PipelineStage, error)
	GetCdStageByCdPipelineIdAndStageType(cdPipelineId int, stageType PipelineStageType) (*PipelineStage, error)

	GetStepIdsByStageId(stageId int) ([]int, error)
	CreatePipelineStageStep(step *PipelineStageStep, tx *pg.Tx) (*PipelineStageStep, error)
	UpdatePipelineStageStep(step *PipelineStageStep, tx *pg.Tx) (*PipelineStageStep, error)
	MarkPipelineStageStepsDeletedByStageId(stageId int, updatedBy int32, tx *pg.Tx) error
	GetAllStepsByStageId(stageId int) ([]*PipelineStageStep, error)
	GetAllCiPipelineIdsByPluginIdAndStageType(pluginId int, stageType string) ([]int, error)
	CheckPluginExistsInCiPipeline(pipelineId int, stageType string, pluginId int) (bool, error)
	GetStepById(stepId int) (*PipelineStageStep, error)
	MarkStepsDeletedByStageId(stageId int) error
	MarkStepsDeletedExcludingActiveStepsInUpdateReq(activeStepIdsPresentInReq []int, stageId int) error
	GetActiveStepsByRefPluginId(refPluginId int) ([]*PipelineStageStep, error)
	CheckIfPluginExistsInPipelineStage(pipelineId int, stageType PipelineStageType, pluginId int) (bool, error)

	CreatePipelineScript(pipelineScript *PluginPipelineScript, tx *pg.Tx) (*PluginPipelineScript, error)
	UpdatePipelineScript(pipelineScript *PluginPipelineScript) (*PluginPipelineScript, error)
	GetScriptIdsByStageId(stageId int) ([]int, error)
	MarkPipelineScriptsDeletedByIds(ids []int, updatedBy int32, tx *pg.Tx) error
	GetScriptDetailById(id int) (*PluginPipelineScript, error)
	MarkScriptDeletedById(scriptId int, tx *pg.Tx) error

	MarkScriptMappingDeletedByScriptId(scriptId int, tx *pg.Tx) error
	CreateScriptMapping(mappings []ScriptPathArgPortMapping, tx *pg.Tx) error
	UpdateScriptMapping(mappings []*ScriptPathArgPortMapping, tx *pg.Tx) error
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
	GetVariablesByStepIdAndVariableType(stepId int, variableType PipelineStageStepVariableType) (variables []*PipelineStageStepVariable, err error)
	MarkVariablesDeletedByStepIdAndVariableType(stepId int, variableType PipelineStageStepVariableType, userId int32, tx *pg.Tx) error
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

func (impl *PipelineStageRepositoryImpl) GetAllCdStagesByCdPipelineId(cdPipelineId int) ([]*PipelineStage, error) {
	var pipelineStages []*PipelineStage
	err := impl.dbConnection.Model(&pipelineStages).
		Where("cd_pipeline_id = ?", cdPipelineId).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting all cd stages by cdPipelineId", "err", err, "cdPipelineId", cdPipelineId)
		return nil, err
	}
	return pipelineStages, nil
}

func (impl *PipelineStageRepositoryImpl) GetAllCdStagesByCdPipelineIds(cdPipelineIds []int) ([]*PipelineStage, error) {
	var pipelineStages []*PipelineStage
	err := impl.dbConnection.Model(&pipelineStages).
		Where("cd_pipeline_id in (?)", pg.In(cdPipelineIds)).
		Where("deleted = ?", false).
		Select()

	if err != nil {
		impl.logger.Errorw("err in getting all cd stages by cdPipelineIds", "err", err, "cdPipelineIds", cdPipelineIds)
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

func (impl *PipelineStageRepositoryImpl) GetCdStageByCdPipelineIdAndStageType(cdPipelineId int, stageType PipelineStageType) (*PipelineStage, error) {
	var pipelineStage PipelineStage
	err := impl.dbConnection.Model(&pipelineStage).
		Where("cd_pipeline_id = ?", cdPipelineId).
		Where("type = ?", stageType).
		Where("deleted = ?", false).
		Select()
	if err != nil {
		impl.logger.Errorw("err in getting cd stage by cdPipelineId", "err", err, "cdPipelineId", cdPipelineId)
		return nil, err
	}
	return &pipelineStage, nil
}

func (impl *PipelineStageRepositoryImpl) CreatePipelineStage(pipelineStage *PipelineStage, tx *pg.Tx) (*PipelineStage, error) {
	var err error
	if tx != nil {
		err = tx.Insert(pipelineStage)
	} else {
		err = impl.dbConnection.Insert(pipelineStage)
	}
	if err != nil {
		impl.logger.Errorw("err at CreatePipelineStage in inserting pipelineStage", err, "pipelineStage", pipelineStage)
		return nil, err
	}

	return pipelineStage, nil
}

func (impl *PipelineStageRepositoryImpl) UpdatePipelineStage(pipelineStage *PipelineStage) (*PipelineStage, error) {
	err := impl.dbConnection.Update(pipelineStage)
	if err != nil {
		impl.logger.Errorw("error in updating pre stage entry", "err", err, "pipelineStage", pipelineStage)
		return nil, err
	}
	return pipelineStage, nil
}

func (impl *PipelineStageRepositoryImpl) MarkPipelineStageDeletedById(stageId int, updatedBy int32, tx *pg.Tx) error {
	var stage PipelineStage
	_, err := tx.Model(&stage).Set("deleted = ?", true).Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", updatedBy).Where("id = ?", stageId).Update()
	if err != nil {
		impl.logger.Errorw("error in marking pipeline stage deleted", "err", err, "StageId", stageId)
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

func (impl *PipelineStageRepositoryImpl) CreatePipelineStageStep(step *PipelineStageStep, tx *pg.Tx) (*PipelineStageStep, error) {
	var err error
	if tx != nil {
		err = tx.Insert(step)
	} else {
		err = impl.dbConnection.Insert(step)
	}
	if err != nil {
		impl.logger.Errorw("error in creating pipeline stage step", "err", err, "step", step)
		return nil, err
	}

	return step, nil
}

func (impl *PipelineStageRepositoryImpl) UpdatePipelineStageStep(step *PipelineStageStep, tx *pg.Tx) (*PipelineStageStep, error) {
	err := tx.Update(step)
	if err != nil {
		impl.logger.Errorw("error in updating pipeline stage step", "err", err, "step", step)
		return nil, err
	}
	return step, nil
}

func (impl *PipelineStageRepositoryImpl) MarkPipelineStageStepsDeletedByStageId(stageId int, updatedBy int32, tx *pg.Tx) error {
	var step PipelineStageStep
	_, err := tx.Model(&step).Set("deleted = ?", true).Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", updatedBy).Where("pipeline_stage_id = ?", stageId).Update()
	if err != nil {
		impl.logger.Errorw("error in marking steps deleted by stageId", "err", err, "ciStageId", stageId)
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

func (impl *PipelineStageRepositoryImpl) GetAllCiPipelineIdsByPluginIdAndStageType(pluginId int, stageType string) ([]int, error) {
	var ciPipelineIds []int
	query := "Select ps.ci_pipeline_id from pipeline_stage ps " +
		"INNER JOIN pipeline_stage_step pss ON pss.pipeline_stage_id = ps.id " +
		"where pss.ref_plugin_id = ? and ps.type = ? and pss.deleted = false and ps.deleted = false"
	_, err := impl.dbConnection.Query(&ciPipelineIds, query, pluginId, stageType)
	if err != nil {
		impl.logger.Errorw("err in getting ciPipelineIds by PluginId and StepType", "err", err, "pluginId", pluginId, "stageType", stageType)
		return nil, err
	}
	return ciPipelineIds, nil
}

func (impl *PipelineStageRepositoryImpl) CheckPluginExistsInCiPipeline(pipelineId int, stageType string, pluginId int) (bool, error) {
	var step PipelineStageStep
	query := `Select * from pipeline_stage_step pss  
		INNER JOIN pipeline_stage ps ON ps.id = pss.pipeline_stage_id  
		where pss.ref_plugin_id = ? and ps.type = ? and pss.deleted = false and ps.deleted = false and ps.ci_pipeline_id= ?;`
	_, err := impl.dbConnection.Query(&step, query, pluginId, stageType, pipelineId)
	if err != nil {
		impl.logger.Errorw("err in getting pipelineStageStep", "err", err, "pluginId", pluginId, "pipelineId", pipelineId, "stageType", stageType)
		return false, err
	}
	return step.Id != 0, nil
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

func (impl *PipelineStageRepositoryImpl) GetActiveStepsByRefPluginId(refPluginId int) ([]*PipelineStageStep, error) {
	var steps []*PipelineStageStep
	err := impl.dbConnection.Model(&steps).
		Where("ref_plugin_id = ?", refPluginId).
		Where("deleted = ?", false).Select()
	if err != nil {
		impl.logger.Errorw("err in getting all steps by refPluginId", "err", err, "refPluginId", refPluginId)
		return nil, err
	}
	return steps, nil
}

func (impl *PipelineStageRepositoryImpl) CreatePipelineScript(pipelineScript *PluginPipelineScript, tx *pg.Tx) (*PluginPipelineScript, error) {
	var err error
	if tx != nil {
		err = tx.Insert(pipelineScript)
	} else {
		err = impl.dbConnection.Insert(pipelineScript)
	}
	if err != nil {
		impl.logger.Errorw("err at CreatePipelineScript in inserting pipelineScript", err, "pipelineScript", pipelineScript)
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

func (impl *PipelineStageRepositoryImpl) MarkScriptDeletedById(scriptId int, tx *pg.Tx) error {
	var script PluginPipelineScript
	_, err := tx.Model(&script).Set("deleted = ?", true).
		Where("id = ?", scriptId).Update()
	if err != nil {
		impl.logger.Errorw("error in marking script deleted", "err", err, "scriptId", scriptId)
		return err
	}
	return nil
}

func (impl *PipelineStageRepositoryImpl) MarkScriptMappingDeletedByScriptId(scriptId int, tx *pg.Tx) error {
	var scriptMapping ScriptPathArgPortMapping
	_, err := tx.Model(&scriptMapping).Set("deleted = ?", true).
		Where("script_id = ?", scriptId).Update()
	if err != nil {
		impl.logger.Errorw("error in marking script mappings deleted", "err", err, "scriptId", scriptId)
		return err
	}
	return nil
}

func (impl *PipelineStageRepositoryImpl) CreateScriptMapping(mappings []ScriptPathArgPortMapping, tx *pg.Tx) error {
	var err error
	if tx != nil {
		err = tx.Insert(&mappings)
	} else {
		err = impl.dbConnection.Insert(&mappings)
	}
	if err != nil {
		impl.logger.Errorw("error in creating pipeline script mappings", "err", err, "mappings", mappings)
		return err
	}
	return nil
}

func (impl *PipelineStageRepositoryImpl) UpdateScriptMapping(mappings []*ScriptPathArgPortMapping, tx *pg.Tx) error {
	var err error
	if tx != nil {
		for _, entry := range mappings {
			err = tx.Update(entry)
			if err != nil {
				impl.logger.Errorw("error in updating ScriptPathArgPortMapping", "entry", entry, "err", err)
				return err
			}
		}

	} else {
		err = impl.dbConnection.Update(&mappings)
	}
	if err != nil {
		impl.logger.Errorw("error in updating pipeline script mappings", "err", err, "mappings", mappings)
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
	var err error
	if tx != nil {
		err = tx.Insert(&variables)
	} else {
		err = impl.dbConnection.Insert(&variables)
	}
	if err != nil {
		impl.logger.Errorw("error in creating pipeline stage step variables", "err", err, "variables", variables)
		return variables, err
	}
	return variables, nil
}

func (impl *PipelineStageRepositoryImpl) UpdatePipelineStageStepVariables(variables []PipelineStageStepVariable, tx *pg.Tx) ([]PipelineStageStepVariable, error) {
	_, err := tx.Model(&variables).UpdateNotNull()
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
		Where("deleted = ?", false).
		Select()
	if err != nil {
		impl.logger.Errorw("err in getting variables by stepId", "err", err, "stepId", stepId)
		return nil, err
	}
	return variables, nil
}

func (impl *PipelineStageRepositoryImpl) GetVariablesByStepIdAndVariableType(stepId int, variableType PipelineStageStepVariableType) (variables []*PipelineStageStepVariable, err error) {
	err = impl.dbConnection.Model().
		Table("pipeline_stage_step_variable").
		Column("pipeline_stage_step_variable.*").
		Where("deleted = ?", false).
		Where("pipeline_stage_step_id = ?", stepId).
		Where("variable_type = ?", variableType).
		Select(&variables)
	return variables, nil
}

func (impl *PipelineStageRepositoryImpl) MarkVariablesDeletedByStepIdAndVariableType(stepId int, variableType PipelineStageStepVariableType, userId int32, tx *pg.Tx) error {
	var variable PipelineStageStepVariable
	_, err := tx.Model(&variable).
		Set("deleted = ?", true).
		Set("updated_by = ?", userId).
		Set("updated_on = ?", time.Now()).
		Where("pipeline_stage_step_id = ?", stepId).
		Where("variable_type = ?", variableType).
		Update()
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
	var err error
	if tx != nil {
		err = tx.Insert(&conditions)
	} else {
		err = impl.dbConnection.Insert(&conditions)
	}
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

func (impl *PipelineStageRepositoryImpl) CheckIfPluginExistsInPipelineStage(pipelineId int, stageType PipelineStageType, pluginId int) (bool, error) {
	var step PipelineStageStep
	query := impl.dbConnection.Model(&step).
		Column("pipeline_stage_step.*").
		Join("INNER JOIN pipeline_stage ps on ps.id = pipeline_stage_step.pipeline_stage_id").
		Where("pipeline_stage_step.ref_plugin_id = ?", pluginId).
		Where("ps.type = ?", stageType).
		Where("pipeline_stage_step.deleted=?", false).
		Where("ps.deleted= ?", false)

	if stageType.IsStageTypePostCi() || stageType.IsStageTypePreCi() {
		query.Where("ps.ci_pipeline_id= ?", pipelineId)
	} else if stageType.IsStageTypePostCd() || stageType.IsStageTypePreCd() {
		query.Where("ps.cd_pipeline_id= ?", pipelineId)
	}
	exists, err := query.Exists()
	if err != nil {
		impl.logger.Errorw("error in getting plugin stage step by pipelineId, stageType nad plugin id", "pipelineId", pipelineId, "stageType", stageType.ToString(), "pluginId", pluginId, "err", err)
		return false, err
	}
	return exists, nil
}
