package pipeline

import (
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PipelineStageService interface {
	GetCiPipelineStageData(ciPipelineId int) (preCiStage *bean.PipelineStageDto, postCiStage *bean.PipelineStageDto, err error)
	CreateCiPreStage(createRequest *bean.PipelineStageDto, ciPipelineId int, userId int32) error
	CreateCiPostStage(createRequest *bean.PipelineStageDto, ciPipelineId int, userId int32) error
}

func NewPipelineStageService(logger *zap.SugaredLogger,
	pipelineStageRepository repository.PipelineStageRepository) *PipelineStageServiceImpl {
	return &PipelineStageServiceImpl{
		logger:                  logger,
		pipelineStageRepository: pipelineStageRepository,
	}
}

type PipelineStageServiceImpl struct {
	logger                  *zap.SugaredLogger
	pipelineStageRepository repository.PipelineStageRepository
}

func (impl *PipelineStageServiceImpl) GetCiPipelineStageData(ciPipelineId int) (*bean.PipelineStageDto, *bean.PipelineStageDto, error) {

	//getting all stages by ci pipeline id
	ciStages, err := impl.pipelineStageRepository.GetAllCiStagesByCiPipelineId(ciPipelineId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting all ciStages by ciPipelineId", "err", err, "ciPipelineId", ciStages)
		return nil, nil, err
	}
	var preCiStage *bean.PipelineStageDto
	var postCiStage *bean.PipelineStageDto
	for _, ciStage := range ciStages {
		if ciStage.Type == repository.PIPELINE_STAGE_TYPE_PRE_CI {
			preCiStage, err = impl.BuildCiStageData(ciStage)
			if err != nil {
				impl.logger.Errorw("error in getting ci stage data", "err", err, "ciStage", ciStage)
				return nil, nil, err
			}
		} else if ciStage.Type == repository.PIPELINE_STAGE_TYPE_POST_CI {
			postCiStage, err = impl.BuildCiStageData(ciStage)
			if err != nil {
				impl.logger.Errorw("error in getting ci stage data", "err", err, "ciStage", ciStage)
				return nil, nil, err
			}
		} else {
			impl.logger.Errorw("found improper stage mapped with ciPipeline", "ciPipelineId", ciPipelineId, "stage", ciStage)
		}
	}
	return preCiStage, postCiStage, nil
}

func (impl *PipelineStageServiceImpl) CreateCiPreStage(preStageReq *bean.PipelineStageDto, ciPipelineId int, userId int32) error {
	preStage := &repository.PipelineStage{
		Name:         preStageReq.Name,
		Description:  preStageReq.Description,
		Type:         repository.PIPELINE_STAGE_TYPE_PRE_CI,
		Deleted:      false,
		CiPipelineId: ciPipelineId,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
	preStage, err := impl.pipelineStageRepository.CreateCiStage(preStage)
	if err != nil {
		impl.logger.Errorw("error in creating entry for preCiStage", "err", err, "preCiStage", preStage)
		return err
	}

	//creating stage steps and all related data
	err = impl.CreateStageSteps(preStageReq.Steps, preStage.Id, userId)
	if err != nil {
		impl.logger.Errorw("error in creating stage steps for pre ci stage", "err", err, "preStageId", preStage.Id)
		return err
	}
	return nil
}

func (impl *PipelineStageServiceImpl) CreateCiPostStage(postStageReq *bean.PipelineStageDto, ciPipelineId int, userId int32) error {
	postStage := &repository.PipelineStage{
		Name:         postStageReq.Name,
		Description:  postStageReq.Description,
		Type:         repository.PIPELINE_STAGE_TYPE_POST_CI,
		Deleted:      false,
		CiPipelineId: ciPipelineId,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
	postStage, err := impl.pipelineStageRepository.CreateCiStage(postStage)
	if err != nil {
		impl.logger.Errorw("error in creating entry for postCiStage", "err", err, "postCiStage", postStage)
		return err
	}
	//creating stage steps and all related data
	err = impl.CreateStageSteps(postStageReq.Steps, postStage.Id, userId)
	if err != nil {
		impl.logger.Errorw("error in creating stage steps for post ci stage", "err", err, "postStageId", postStage.Id)
		return err
	}
	return nil
}

func (impl *PipelineStageServiceImpl) CreateStageSteps(steps []*bean.PipelineStageStepDto, stageId int, userId int32) (err error) {
	for _, step := range steps {
		var stepId int
		var inputVariables []*bean.StepVariableDto
		var outputVariables []*bean.StepVariableDto
		var conditionDetails []*bean.ConditionDetailDto
		if step.StepType == repository.PIPELINE_STEP_TYPE_INLINE {
			inlineStepDetail := step.InlineStepDetail
			//creating script entry first, because step entry needs scriptId
			scriptEntryId, err := impl.CreateScriptAndMappingForInlineStep(inlineStepDetail, userId)
			if err != nil {
				impl.logger.Errorw("error in creating script and mapping for inline step", "err", err, "inlineStepDetail", inlineStepDetail)
				return err
			}
			inlineStep := &repository.PipelineStageStep{
				PipelineStageId:     stageId,
				Name:                step.Name,
				Description:         step.Description,
				Index:               step.Index,
				StepType:            step.StepType,
				ScriptId:            scriptEntryId,
				ReportDirectoryPath: step.ReportDirectoryPath,
				Deleted:             false,
				AuditLog: sql.AuditLog{
					CreatedOn: time.Now(),
					CreatedBy: userId,
					UpdatedOn: time.Now(),
					UpdatedBy: userId,
				},
			}
			inlineStep, err = impl.pipelineStageRepository.CreatePipelineStageStep(inlineStep)
			if err != nil {
				impl.logger.Errorw("error in creating inline step", "err", err, "step", inlineStep)
				return err
			}
			stepId = inlineStep.Id
			inputVariables = inlineStepDetail.InputVariables
			outputVariables = inlineStepDetail.OutputVariables
			conditionDetails = inlineStepDetail.ConditionDetails
		} else if step.StepType == repository.PIPELINE_STEP_TYPE_REF_PLUGIN {
			refPluginStepDetail := step.RefPluginStepDetail
			refPluginStep := &repository.PipelineStageStep{
				PipelineStageId:     stageId,
				Name:                step.Name,
				Description:         step.Description,
				Index:               step.Index,
				StepType:            step.StepType,
				RefPluginId:         refPluginStepDetail.PluginId,
				ReportDirectoryPath: step.ReportDirectoryPath,
				Deleted:             false,
				AuditLog: sql.AuditLog{
					CreatedOn: time.Now(),
					CreatedBy: userId,
					UpdatedOn: time.Now(),
					UpdatedBy: userId,
				},
			}
			refPluginStep, err = impl.pipelineStageRepository.CreatePipelineStageStep(refPluginStep)
			if err != nil {
				impl.logger.Errorw("error in creating ref plugin step", "err", err, "step", refPluginStep)
				return err
			}
			stepId = refPluginStep.Id
			inputVariables = refPluginStepDetail.InputVariables
			outputVariables = refPluginStepDetail.OutputVariables
			conditionDetails = refPluginStepDetail.ConditionDetails
		}
		//creating input variables
		inputVariablesRepo, err := impl.CreateVariablesEntryInDb(stepId, inputVariables, repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_INPUT, userId)
		if err != nil {
			impl.logger.Errorw("error in creating input variables for step", "err", err, "stepId", stepId, "stepType", step.StepType)
			return err
		}
		//creating output variables
		outputVariablesRepo, err := impl.CreateVariablesEntryInDb(stepId, outputVariables, repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_OUTPUT, userId)
		if err != nil {
			impl.logger.Errorw("error in creating output variables for step", "err", err, "stepId", stepId, "stepType", step.StepType)
			return err
		}
		variableNameIdMap := make(map[string]int)
		for _, inVar := range inputVariablesRepo {
			variableNameIdMap[inVar.Name] = inVar.Id
		}
		for _, outVar := range outputVariablesRepo {
			variableNameIdMap[outVar.Name] = outVar.Id
		}
		//creating conditions
		_, err = impl.CreateConditionsEntryInDb(stepId, conditionDetails, variableNameIdMap, userId)
		if err != nil {
			impl.logger.Errorw("error in creating conditions for step", "err", err, "stepId", stepId, "stepType", step.StepType)
			return err
		}
	}
	return nil
}

func (impl *PipelineStageServiceImpl) CreateScriptAndMappingForInlineStep(inlineStepDetail *bean.InlineStepDetailDto, userId int32) (scriptId int, err error) {
	scriptEntry := &repository.PluginPipelineScript{
		Script:               inlineStepDetail.Script,
		Type:                 inlineStepDetail.ScriptType,
		DockerfileExists:     inlineStepDetail.DockerfileExists,
		MountPath:            inlineStepDetail.MountPath,
		MountCodeToContainer: inlineStepDetail.MountCodeToContainer,
		ConfigureMountPath:   inlineStepDetail.ConfigureMountPath,
		ContainerImagePath:   inlineStepDetail.ContainerImagePath,
		ImagePullSecretType:  inlineStepDetail.ImagePullSecretType,
		ImagePullSecret:      inlineStepDetail.ImagePullSecret,
		Deleted:              false,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
	scriptEntry, err = impl.pipelineStageRepository.CreatePipelineScript(scriptEntry)
	if err != nil {
		impl.logger.Errorw("error in creating script entry for inline step", "err", err, "scriptEntry", scriptEntry)
		return 0, err
	}
	var mountPathMap []repository.ScriptPathArgPortMapping
	for _, mountPath := range inlineStepDetail.MountPathMap {
		repositoryEntry := repository.ScriptPathArgPortMapping{
			TypeOfMapping:       repository2.SCRIPT_MAPPING_TYPE_FILE_PATH,
			FilePathOnDisk:      mountPath.FilePathOnDisk,
			FilePathOnContainer: mountPath.FilePathOnContainer,
			ScriptId:            scriptEntry.Id,
			Deleted:             false,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: userId,
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		mountPathMap = append(mountPathMap, repositoryEntry)
	}
	err = impl.pipelineStageRepository.CreateScriptMapping(mountPathMap)
	if err != nil {
		impl.logger.Errorw("error in creating script mappings", "err", err, "scriptMappings", mountPathMap)
		return 0, err
	}
	return scriptEntry.Id, nil
}

func (impl *PipelineStageServiceImpl) CreateVariablesEntryInDb(stepId int, variables []*bean.StepVariableDto, variableType repository.PipelineStageStepVariableType, userId int32) ([]repository.PipelineStageStepVariable, error) {
	var variablesRepo []repository.PipelineStageStepVariable
	var err error
	for _, v := range variables {
		inVarRepo := repository.PipelineStageStepVariable{
			PipelineStageStepId:   stepId,
			Name:                  v.Name,
			Format:                v.Format,
			Description:           v.Description,
			IsExposed:             v.IsExposed,
			AllowEmptyValue:       v.AllowEmptyValue,
			DefaultValue:          v.DefaultValue,
			Value:                 v.Value,
			ValueType:             v.ValueType,
			VariableType:          variableType,
			PreviousStepIndex:     v.PreviousStepIndex,
			ReferenceVariableName: v.ReferenceVariableName,
			Deleted:               false,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: userId,
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		variablesRepo = append(variablesRepo, inVarRepo)
	}
	//saving input variables
	variablesRepo, err = impl.pipelineStageRepository.CreatePipelineStageStepVariables(variablesRepo)
	if err != nil {
		impl.logger.Errorw("error in creating variables for pipeline stage steps", "err", err, "variables", variablesRepo)
		return nil, err
	}
	return variablesRepo, nil
}

func (impl *PipelineStageServiceImpl) CreateConditionsEntryInDb(stepId int, conditions []*bean.ConditionDetailDto, variableNameIdMap map[string]int, userId int32) ([]repository.PipelineStageStepCondition, error) {
	var conditionsRepo []repository.PipelineStageStepCondition
	var err error
	for _, condition := range conditions {
		varId := variableNameIdMap[condition.ConditionOnVariable]
		conditionRepo := repository.PipelineStageStepCondition{
			PipelineStageStepId: stepId,
			ConditionVariableId: varId,
			ConditionType:       condition.ConditionType,
			ConditionalOperator: condition.ConditionalOperator,
			ConditionalValue:    condition.ConditionalValue,
			Deleted:             false,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: userId,
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		conditionsRepo = append(conditionsRepo, conditionRepo)
	}
	//saving conditions
	conditionsRepo, err = impl.pipelineStageRepository.CreatePipelineStageStepConditions(conditionsRepo)
	if err != nil {
		impl.logger.Errorw("error in creating pipeline stage step conditions", "err", err, "conditionsRepo", conditionsRepo)
		return nil, err
	}
	return conditionsRepo, nil
}

func (impl *PipelineStageServiceImpl) BuildCiStageData(ciStage *repository.PipelineStage) (*bean.PipelineStageDto, error) {
	stageData := &bean.PipelineStageDto{
		Id:          ciStage.Id,
		Name:        ciStage.Name,
		Description: ciStage.Description,
		Type:        ciStage.Type,
	}
	//getting all steps in this stage
	steps, err := impl.pipelineStageRepository.GetAllStepsByStageId(ciStage.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting ci steps by stage id", "err", err, "ciStageId", ciStage.Id)
		return nil, err
	}
	var stepsDto []*bean.PipelineStageStepDto
	for _, step := range steps {
		stepDto := &bean.PipelineStageStepDto{
			Id:                  step.Id,
			Name:                step.Name,
			Index:               step.Index,
			Description:         step.Description,
			ReportDirectoryPath: step.ReportDirectoryPath,
			StepType:            step.StepType,
		}
		if step.StepType == repository.PIPELINE_STEP_TYPE_INLINE {
			inlineStepDetail, err := impl.BuildInlineStepData(step)
			if err != nil {
				impl.logger.Errorw("error in getting inline step data", "err", err, "step", step)
				return nil, err
			}
			stepDto.InlineStepDetail = inlineStepDetail
		} else if step.StepType == repository.PIPELINE_STEP_TYPE_REF_PLUGIN {
			refPluginStepDetail, err := impl.BuildRefPluginStepData(step)
			if err != nil {
				impl.logger.Errorw("error in getting ref plugin step data", "err", err, "step", step)
				return nil, err
			}
			stepDto.RefPluginStepDetail = refPluginStepDetail
		}
	}
	stageData.Steps = stepsDto
	return stageData, nil
}

func (impl *PipelineStageServiceImpl) BuildInlineStepData(step *repository.PipelineStageStep) (*bean.InlineStepDetailDto, error) {
	//getting script details for step
	scriptDetail, err := impl.pipelineStageRepository.GetScriptDetailById(step.ScriptId)
	if err != nil {
		impl.logger.Errorw("error in getting script details by id", "err", err, "scriptId", step.ScriptId)
		return nil, err
	}
	inlineStepDetail := &bean.InlineStepDetailDto{
		ScriptType:           scriptDetail.Type,
		Script:               scriptDetail.Script,
		DockerfileExists:     scriptDetail.DockerfileExists,
		MountPath:            scriptDetail.MountPath,
		MountCodeToContainer: scriptDetail.MountCodeToContainer,
		ConfigureMountPath:   scriptDetail.ConfigureMountPath,
		ContainerImagePath:   scriptDetail.ContainerImagePath,
		ImagePullSecretType:  scriptDetail.ImagePullSecretType,
		ImagePullSecret:      scriptDetail.ImagePullSecret,
	}
	//getting script mapping details
	scriptMappings, err := impl.pipelineStageRepository.GetScriptMappingDetailByScriptId(step.ScriptId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting script mapping by scriptId", "err", err, "scriptId", step.ScriptId)
		return nil, err
	}
	//only getting mount path map, because port mapping and docker args are not supported yet(CONTAINER_IMAGE type step)
	var mountPathMap []*bean.MountPathMap
	for _, scriptMapping := range scriptMappings {
		if scriptMapping.TypeOfMapping == repository2.SCRIPT_MAPPING_TYPE_FILE_PATH {
			mapEntry := &bean.MountPathMap{
				FilePathOnDisk:      scriptMapping.FilePathOnDisk,
				FilePathOnContainer: scriptMapping.FilePathOnContainer,
			}
			mountPathMap = append(mountPathMap, mapEntry)
		}
	}
	inlineStepDetail.MountPathMap = mountPathMap
	inputVariablesDto, outputVariablesDto, conditionsDto, err := impl.BuildVariableAndConditionDataByStepId(step.Id)
	if err != nil {
		impl.logger.Errorw("error in getting variables and conditions data by stepId", "err", err, "stepId", step.Id)
		return nil, err
	}
	inlineStepDetail.InputVariables = inputVariablesDto
	inlineStepDetail.OutputVariables = outputVariablesDto
	inlineStepDetail.ConditionDetails = conditionsDto
	return inlineStepDetail, nil
}

func (impl *PipelineStageServiceImpl) BuildRefPluginStepData(step *repository.PipelineStageStep) (*bean.RefPluginStepDetailDto, error) {
	refPluginStepDetail := &bean.RefPluginStepDetailDto{
		PluginId: step.RefPluginId,
	}
	inputVariablesDto, outputVariablesDto, conditionsDto, err := impl.BuildVariableAndConditionDataByStepId(step.Id)
	if err != nil {
		impl.logger.Errorw("error in getting variables and conditions data by stepId", "err", err, "stepId", step.Id)
		return nil, err
	}
	refPluginStepDetail.InputVariables = inputVariablesDto
	refPluginStepDetail.OutputVariables = outputVariablesDto
	refPluginStepDetail.ConditionDetails = conditionsDto
	return refPluginStepDetail, nil
}

func (impl *PipelineStageServiceImpl) BuildVariableAndConditionDataByStepId(stepId int) ([]*bean.StepVariableDto, []*bean.StepVariableDto, []*bean.ConditionDetailDto, error) {
	var inputVariablesDto []*bean.StepVariableDto
	var outputVariablesDto []*bean.StepVariableDto
	var conditionsDto []*bean.ConditionDetailDto
	//getting all variables in the step
	variables, err := impl.pipelineStageRepository.GetVariablesByStepId(stepId)
	if err != nil {
		impl.logger.Errorw("error in getting variables by stepId", "err", err, "stepId", stepId)
		return nil, nil, nil, err
	}
	for _, variable := range variables {
		variableDto := &bean.StepVariableDto{
			Id:                    variable.Id,
			Name:                  variable.Name,
			Format:                variable.Format,
			Description:           variable.Description,
			IsExposed:             variable.IsExposed,
			AllowEmptyValue:       variable.AllowEmptyValue,
			DefaultValue:          variable.DefaultValue,
			Value:                 variable.Value,
			ValueType:             variable.ValueType,
			PreviousStepIndex:     variable.PreviousStepIndex,
			ReferenceVariableName: variable.ReferenceVariableName,
		}
		if variable.VariableType == repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_INPUT {
			inputVariablesDto = append(inputVariablesDto, variableDto)
		} else if variable.VariableType == repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_OUTPUT {
			outputVariablesDto = append(outputVariablesDto, variableDto)
		}
		conditions, err := impl.pipelineStageRepository.GetConditionsByVariableId(variable.Id)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting conditions by variableId", "err", err, "variableId", variable.Id)
			return nil, nil, nil, err
		}
		for _, condition := range conditions {
			conditionDto := &bean.ConditionDetailDto{
				Id:                  condition.Id,
				ConditionOnVariable: variable.Name,
				ConditionalOperator: condition.ConditionalOperator,
				ConditionalValue:    condition.ConditionalValue,
				ConditionType:       condition.ConditionType,
			}
			conditionsDto = append(conditionsDto, conditionDto)
		}
	}
	return inputVariablesDto, outputVariablesDto, conditionsDto, nil
}
