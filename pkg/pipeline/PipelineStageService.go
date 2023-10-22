package pipeline

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables"
	repository3 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PipelineStageService interface {
	GetCiPipelineStageData(ciPipelineId int) (preCiStage *bean.PipelineStageDto, postCiStage *bean.PipelineStageDto, err error)
	CreatePipelineStage(stageReq *bean.PipelineStageDto, stageType repository.PipelineStageType, pipelineId int, userId int32) error
	UpdatePipelineStage(stageReq *bean.PipelineStageDto, stageType repository.PipelineStageType, pipelineId int, userId int32) error
	DeletePipelineStage(stageReq *bean.PipelineStageDto, userId int32, tx *pg.Tx) error
	BuildPrePostAndRefPluginStepsDataForWfRequest(pipelineId int, stageType string, scope resourceQualifiers.Scope) (*bean.PrePostAndRefPluginStepsResponse, error)
	GetCiPipelineStageDataDeepCopy(ciPipelineId int) (preCiStage *bean.PipelineStageDto, postCiStage *bean.PipelineStageDto, err error)
	GetCdPipelineStageDataDeepCopy(cdPipelineId int) (*bean.PipelineStageDto, *bean.PipelineStageDto, error)
	GetCdPipelineStageDataDeepCopyForPipelineIds(cdPipelineIds []int) (map[int][]*bean.PipelineStageDto, error)
	GetCdStageByCdPipelineIdAndStageType(cdPipelineId int, stageType repository.PipelineStageType) (*repository.PipelineStage, error)
	// DeletePipelineStageIfReq function is used to delete corrupted pipelineStage data
	// , there was a bug(https://github.com/devtron-labs/devtron/issues/3826) where we were not deleting pipeline stage entry even after deleting all the pipelineStageSteps
	// , this will delete those pipelineStage entry
	DeletePipelineStageIfReq(stageReq *bean.PipelineStageDto, userId int32) (error, bool)
}

func NewPipelineStageService(logger *zap.SugaredLogger,
	pipelineStageRepository repository.PipelineStageRepository,
	globalPluginRepository repository2.GlobalPluginRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	scopedVariableManager variables.ScopedVariableManager,
	// scopedVariableService variables.ScopedVariableService,
	// variableEntityMappingService variables.VariableEntityMappingService,
	// variableTemplateParser parsers.VariableTemplateParser,
) *PipelineStageServiceImpl {
	return &PipelineStageServiceImpl{
		logger:                  logger,
		pipelineStageRepository: pipelineStageRepository,
		globalPluginRepository:  globalPluginRepository,
		pipelineRepository:      pipelineRepository,
		scopedVariableManager:   scopedVariableManager,
		//scopedVariableService:        scopedVariableService,
		//variableEntityMappingService: variableEntityMappingService,
		//variableTemplateParser:       variableTemplateParser,
	}
}

type PipelineStageServiceImpl struct {
	logger                  *zap.SugaredLogger
	pipelineStageRepository repository.PipelineStageRepository
	globalPluginRepository  repository2.GlobalPluginRepository
	pipelineRepository      pipelineConfig.PipelineRepository
	scopedVariableManager   variables.ScopedVariableManager
	//scopedVariableService        variables.ScopedVariableService
	//variableEntityMappingService variables.VariableEntityMappingService
	//variableTemplateParser       parsers.VariableTemplateParser
}

func (impl *PipelineStageServiceImpl) GetCiPipelineStageDataDeepCopy(ciPipelineId int) (*bean.PipelineStageDto, *bean.PipelineStageDto, error) {

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
			preCiStage, err = impl.BuildPipelineStageDataDeepCopy(ciStage)
			if err != nil {
				impl.logger.Errorw("error in getting ci stage data", "err", err, "ciStage", ciStage)
				return nil, nil, err
			}
		} else if ciStage.Type == repository.PIPELINE_STAGE_TYPE_POST_CI {
			postCiStage, err = impl.BuildPipelineStageDataDeepCopy(ciStage)
			if err != nil {
				impl.logger.Errorw("error in getting ci stage data", "err", err, "ciStage", ciStage)
				return nil, nil, err
			}
		} else {
			impl.logger.Errorw("found improper stage mapped with ciPipelineRequest", "ciPipelineId", ciPipelineId, "stage", ciStage)
		}
	}
	return preCiStage, postCiStage, nil
}

func (impl *PipelineStageServiceImpl) GetCdStageByCdPipelineIdAndStageType(cdPipelineId int, stageType repository.PipelineStageType) (*repository.PipelineStage, error) {
	return impl.pipelineStageRepository.GetCdStageByCdPipelineIdAndStageType(cdPipelineId, stageType)
}

func (impl *PipelineStageServiceImpl) GetCdPipelineStageDataDeepCopy(cdPipelineId int) (*bean.PipelineStageDto, *bean.PipelineStageDto, error) {
	//getting all stages by cd pipeline id
	cdStages, err := impl.pipelineStageRepository.GetAllCdStagesByCdPipelineId(cdPipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting all cdStages by cdPipelineId", "err", err, "cdPipelineStages", cdStages)
		return nil, nil, err
	}
	if len(cdStages) == 0 {
		//no entry for cdStages in db
		return nil, nil, nil
	}
	var preDeployStage *bean.PipelineStageDto
	var postDeployStage *bean.PipelineStageDto
	for _, cdStage := range cdStages {
		if cdStage.Type == repository.PIPELINE_STAGE_TYPE_PRE_CD {
			preDeployStage, err = impl.BuildPipelineStageDataDeepCopy(cdStage)
			if err != nil {
				impl.logger.Errorw("error in getting cd stage data", "err", err, "cdStage", cdStage)
				return nil, nil, err
			}
		} else if cdStage.Type == repository.PIPELINE_STAGE_TYPE_POST_CD {
			postDeployStage, err = impl.BuildPipelineStageDataDeepCopy(cdStage)
			if err != nil {
				impl.logger.Errorw("error in getting cd stage data", "err", err, "cdStage", cdStage)
				return nil, nil, err
			}
		} else {
			impl.logger.Errorw("found improper stage mapped with cdPipeline", "cdPipelineId", cdPipelineId, "stage", cdStage)
		}
	}
	if preDeployStage != nil {
		preDeployStage.Name = "Pre-Deployment"
	}
	if postDeployStage != nil {
		postDeployStage.Name = "Post-Deployment"
	}
	return preDeployStage, postDeployStage, nil
}

func (impl *PipelineStageServiceImpl) GetCdPipelineStageDataDeepCopyForPipelineIds(cdPipelineIds []int) (map[int][]*bean.PipelineStageDto, error) {
	pipelinePrePostStageMappingResp := make(map[int][]*bean.PipelineStageDto)
	pipelineStageMapping := make(map[int][]*repository.PipelineStage)
	pipelineStages, err := impl.pipelineStageRepository.GetAllCdStagesByCdPipelineIds(cdPipelineIds)
	if err != nil {
		impl.logger.Errorw("error in getting pipelineStages from cdPipelineIds", "err", err, "cdPipelineIds", cdPipelineIds)
		return pipelinePrePostStageMappingResp, err
	}
	if len(pipelineStages) == 0 {
		return pipelinePrePostStageMappingResp, nil
	}
	for _, pipelineStage := range pipelineStages {
		pipelineStageMapping[pipelineStage.CdPipelineId] = append(pipelineStageMapping[pipelineStage.CdPipelineId], pipelineStage)
	}

	for pipelineId, pipelineStages := range pipelineStageMapping {
		var preDeployStage *bean.PipelineStageDto
		var postDeployStage *bean.PipelineStageDto
		for _, pipelineStage := range pipelineStages {
			if pipelineStage.Type == repository.PIPELINE_STAGE_TYPE_PRE_CD {
				preDeployStage, err = impl.BuildPipelineStageDataDeepCopy(pipelineStage)
				if err != nil {
					impl.logger.Errorw("error in getting cd stage data", "err", err, "cdStage", bean.CdStage)
					return pipelinePrePostStageMappingResp, err
				}
				preDeployStage.Name = "Pre-Deployment"
			} else if pipelineStage.Type == repository.PIPELINE_STAGE_TYPE_POST_CD {
				postDeployStage, err = impl.BuildPipelineStageDataDeepCopy(pipelineStage)
				if err != nil {
					impl.logger.Errorw("error in getting cd stage data", "err", err, "cdStage", bean.CdStage)
					return pipelinePrePostStageMappingResp, err
				}
				postDeployStage.Name = "Post-Deployment"
			} else {
				impl.logger.Errorw("found improper stage mapped with cdPipeline", "cdPipelineId", pipelineId, "stage", bean.CdStage)
			}
		}
		pipelinePrePostStageMappingResp[pipelineId] = append(pipelinePrePostStageMappingResp[pipelineId], preDeployStage)
		pipelinePrePostStageMappingResp[pipelineId] = append(pipelinePrePostStageMappingResp[pipelineId], postDeployStage)

	}
	return pipelinePrePostStageMappingResp, nil
}

func (impl *PipelineStageServiceImpl) BuildPipelineStageDataDeepCopy(pipelineStage *repository.PipelineStage) (*bean.PipelineStageDto, error) {
	stageData := &bean.PipelineStageDto{
		Id:          pipelineStage.Id,
		Name:        pipelineStage.Name,
		Description: pipelineStage.Description,
		Type:        pipelineStage.Type,
	}
	if pipelineStage.Type == repository.PIPELINE_STAGE_TYPE_PRE_CD || pipelineStage.Type == repository.PIPELINE_STAGE_TYPE_POST_CD {
		pipeline, err := impl.pipelineRepository.FindById(pipelineStage.CdPipelineId)
		if err != nil {
			impl.logger.Errorw("error in getting pipeline from cdPipelineId", "err", err, "cdPipelineId", pipelineStage.CdPipelineId)
			return nil, err
		}
		if pipelineStage.Type == repository.PIPELINE_STAGE_TYPE_PRE_CD {
			stageData.TriggerType = pipeline.PreTriggerType
		}
		if pipelineStage.Type == repository.PIPELINE_STAGE_TYPE_POST_CD {
			stageData.TriggerType = pipeline.PostTriggerType
		}
	}

	//getting all steps in this stage
	steps, err := impl.pipelineStageRepository.GetAllStepsByStageId(pipelineStage.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting pipeline steps by stage id", "err", err, "pipelineStage", pipelineStage)
		return nil, err
	}
	var stepsDto []*bean.PipelineStageStepDto
	for _, step := range steps {
		stepDto := &bean.PipelineStageStepDto{
			Id:                       step.Id,
			Name:                     step.Name,
			Index:                    step.Index,
			Description:              step.Description,
			OutputDirectoryPath:      step.OutputDirectoryPath,
			StepType:                 step.StepType,
			TriggerIfParentStageFail: step.TriggerIfParentStageFail,
		}
		if step.StepType == repository.PIPELINE_STEP_TYPE_INLINE {
			inlineStepDetail, err := impl.BuildInlineStepDataDeepCopy(step)
			if err != nil {
				impl.logger.Errorw("error in getting inline step data", "err", err, "step", step)
				return nil, err
			}
			stepDto.InlineStepDetail = inlineStepDetail
		} else if step.StepType == repository.PIPELINE_STEP_TYPE_REF_PLUGIN {
			refPluginStepDetail, err := impl.BuildRefPluginStepDataDeepCopy(step)
			if err != nil {
				impl.logger.Errorw("error in getting ref plugin step data", "err", err, "step", step)
				return nil, err
			}
			stepDto.RefPluginStepDetail = refPluginStepDetail
		}
		stepsDto = append(stepsDto, stepDto)
	}
	stageData.Steps = stepsDto
	return stageData, nil
}

func (impl *PipelineStageServiceImpl) BuildInlineStepDataDeepCopy(step *repository.PipelineStageStep) (*bean.InlineStepDetailDto, error) {
	//getting script details for step
	scriptDetail, err := impl.pipelineStageRepository.GetScriptDetailById(step.ScriptId)
	if err != nil {
		impl.logger.Errorw("error in getting script details by id", "err", err, "scriptId", step.ScriptId)
		return nil, err
	}
	inlineStepDetail := &bean.InlineStepDetailDto{
		ScriptType:               scriptDetail.Type,
		Script:                   scriptDetail.Script,
		StoreScriptAt:            scriptDetail.StoreScriptAt,
		DockerfileExists:         scriptDetail.DockerfileExists,
		MountPath:                scriptDetail.MountPath,
		MountCodeToContainer:     scriptDetail.MountCodeToContainer,
		MountCodeToContainerPath: scriptDetail.MountCodeToContainerPath,
		MountDirectoryFromHost:   scriptDetail.MountDirectoryFromHost,
		ContainerImagePath:       scriptDetail.ContainerImagePath,
		ImagePullSecretType:      scriptDetail.ImagePullSecretType,
		ImagePullSecret:          scriptDetail.ImagePullSecret,
	}
	//getting script mapping details
	scriptMappings, err := impl.pipelineStageRepository.GetScriptMappingDetailByScriptId(step.ScriptId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting script mapping by scriptId", "err", err, "scriptId", step.ScriptId)
		return nil, err
	}
	var mountPathMap []*bean.MountPathMap
	var commandArgsMap []*bean.CommandArgsMap
	var portMap []*bean.PortMap
	for _, scriptMapping := range scriptMappings {
		if scriptMapping.TypeOfMapping == repository2.SCRIPT_MAPPING_TYPE_FILE_PATH {
			mapEntry := &bean.MountPathMap{
				FilePathOnDisk:      scriptMapping.FilePathOnDisk,
				FilePathOnContainer: scriptMapping.FilePathOnContainer,
			}
			mountPathMap = append(mountPathMap, mapEntry)
		} else if scriptMapping.TypeOfMapping == repository2.SCRIPT_MAPPING_TYPE_DOCKER_ARG {
			mapEntry := &bean.CommandArgsMap{
				Command: scriptMapping.Command,
				Args:    scriptMapping.Args,
			}
			commandArgsMap = append(commandArgsMap, mapEntry)
		} else if scriptMapping.TypeOfMapping == repository2.SCRIPT_MAPPING_TYPE_PORT {
			mapEntry := &bean.PortMap{
				PortOnLocal:     scriptMapping.PortOnLocal,
				PortOnContainer: scriptMapping.PortOnContainer,
			}
			portMap = append(portMap, mapEntry)
		}
	}
	inlineStepDetail.MountPathMap = mountPathMap
	inlineStepDetail.CommandArgsMap = commandArgsMap
	inlineStepDetail.PortMap = portMap
	inputVariablesDto, outputVariablesDto, conditionsDto, err := impl.BuildVariableAndConditionDataByStepIdDeepCopy(step.Id)
	if err != nil {
		impl.logger.Errorw("error in getting variables and conditions data by stepId", "err", err, "stepId", step.Id)
		return nil, err
	}
	inlineStepDetail.InputVariables = inputVariablesDto
	inlineStepDetail.OutputVariables = outputVariablesDto
	inlineStepDetail.ConditionDetails = conditionsDto
	return inlineStepDetail, nil
}

func (impl *PipelineStageServiceImpl) BuildRefPluginStepDataDeepCopy(step *repository.PipelineStageStep) (*bean.RefPluginStepDetailDto, error) {
	refPluginStepDetail := &bean.RefPluginStepDetailDto{
		PluginId: step.RefPluginId,
	}
	inputVariablesDto, outputVariablesDto, conditionsDto, err := impl.BuildVariableAndConditionDataByStepIdDeepCopy(step.Id)
	if err != nil {
		impl.logger.Errorw("error in getting variables and conditions data by stepId", "err", err, "stepId", step.Id)
		return nil, err
	}
	refPluginStepDetail.InputVariables = inputVariablesDto
	refPluginStepDetail.OutputVariables = outputVariablesDto
	refPluginStepDetail.ConditionDetails = conditionsDto
	return refPluginStepDetail, nil
}
func (impl *PipelineStageServiceImpl) BuildVariableAndConditionDataByStepIdDeepCopy(stepId int) ([]*bean.StepVariableDto, []*bean.StepVariableDto, []*bean.ConditionDetailDto, error) {
	var inputVariablesDto []*bean.StepVariableDto
	var outputVariablesDto []*bean.StepVariableDto
	var conditionsDto []*bean.ConditionDetailDto
	//getting all variables in the step
	variables, err := impl.pipelineStageRepository.GetVariablesByStepId(stepId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting variables by stepId", "err", err, "stepId", stepId)
		return nil, nil, nil, err
	}
	variableNameIdMap := make(map[int]string)
	for _, variable := range variables {
		variableNameIdMap[variable.Id] = variable.Name
		variableDto := &bean.StepVariableDto{
			Name:                      variable.Name,
			Format:                    variable.Format,
			Description:               variable.Description,
			IsExposed:                 variable.IsExposed,
			AllowEmptyValue:           variable.AllowEmptyValue,
			DefaultValue:              variable.DefaultValue,
			Value:                     variable.Value,
			ValueType:                 variable.ValueType,
			PreviousStepIndex:         variable.PreviousStepIndex,
			ReferenceVariableName:     variable.ReferenceVariableName,
			ReferenceVariableStage:    variable.ReferenceVariableStage,
			VariableStepIndexInPlugin: variable.VariableStepIndexInPlugin,
		}
		if variable.VariableType == repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_INPUT {
			inputVariablesDto = append(inputVariablesDto, variableDto)
		} else if variable.VariableType == repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_OUTPUT {
			outputVariablesDto = append(outputVariablesDto, variableDto)
		}
	}
	conditions, err := impl.pipelineStageRepository.GetConditionsByStepId(stepId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting conditions by stepId", "err", err, "stepId", stepId)
		return nil, nil, nil, err
	}
	for _, condition := range conditions {
		conditionDto := &bean.ConditionDetailDto{
			Id:                  condition.Id,
			ConditionalOperator: condition.ConditionalOperator,
			ConditionalValue:    condition.ConditionalValue,
			ConditionType:       condition.ConditionType,
		}
		varName, ok := variableNameIdMap[condition.ConditionVariableId]
		if ok {
			conditionDto.ConditionOnVariable = varName
		}
		conditionsDto = append(conditionsDto, conditionDto)
	}
	return inputVariablesDto, outputVariablesDto, conditionsDto, nil
}

// GetCiPipelineStageData and related methods starts
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
			impl.logger.Errorw("found improper stage mapped with ciPipelineRequest", "ciPipelineId", ciPipelineId, "stage", ciStage)
		}
	}
	return preCiStage, postCiStage, nil
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
			Id:                       step.Id,
			Name:                     step.Name,
			Index:                    step.Index,
			Description:              step.Description,
			OutputDirectoryPath:      step.OutputDirectoryPath,
			StepType:                 step.StepType,
			TriggerIfParentStageFail: step.TriggerIfParentStageFail,
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
		stepsDto = append(stepsDto, stepDto)
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
		ScriptType:               scriptDetail.Type,
		Script:                   scriptDetail.Script,
		StoreScriptAt:            scriptDetail.StoreScriptAt,
		DockerfileExists:         scriptDetail.DockerfileExists,
		MountPath:                scriptDetail.MountPath,
		MountCodeToContainer:     scriptDetail.MountCodeToContainer,
		MountCodeToContainerPath: scriptDetail.MountCodeToContainerPath,
		MountDirectoryFromHost:   scriptDetail.MountDirectoryFromHost,
		ContainerImagePath:       scriptDetail.ContainerImagePath,
		ImagePullSecretType:      scriptDetail.ImagePullSecretType,
		ImagePullSecret:          scriptDetail.ImagePullSecret,
	}
	//getting script mapping details
	scriptMappings, err := impl.pipelineStageRepository.GetScriptMappingDetailByScriptId(step.ScriptId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting script mapping by scriptId", "err", err, "scriptId", step.ScriptId)
		return nil, err
	}
	var mountPathMap []*bean.MountPathMap
	var commandArgsMap []*bean.CommandArgsMap
	var portMap []*bean.PortMap
	for _, scriptMapping := range scriptMappings {
		if scriptMapping.TypeOfMapping == repository2.SCRIPT_MAPPING_TYPE_FILE_PATH {
			mapEntry := &bean.MountPathMap{
				FilePathOnDisk:      scriptMapping.FilePathOnDisk,
				FilePathOnContainer: scriptMapping.FilePathOnContainer,
			}
			mountPathMap = append(mountPathMap, mapEntry)
		} else if scriptMapping.TypeOfMapping == repository2.SCRIPT_MAPPING_TYPE_DOCKER_ARG {
			mapEntry := &bean.CommandArgsMap{
				Command: scriptMapping.Command,
				Args:    scriptMapping.Args,
			}
			commandArgsMap = append(commandArgsMap, mapEntry)
		} else if scriptMapping.TypeOfMapping == repository2.SCRIPT_MAPPING_TYPE_PORT {
			mapEntry := &bean.PortMap{
				PortOnLocal:     scriptMapping.PortOnLocal,
				PortOnContainer: scriptMapping.PortOnContainer,
			}
			portMap = append(portMap, mapEntry)
		}
	}
	inlineStepDetail.MountPathMap = mountPathMap
	inlineStepDetail.CommandArgsMap = commandArgsMap
	inlineStepDetail.PortMap = portMap
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
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting variables by stepId", "err", err, "stepId", stepId)
		return nil, nil, nil, err
	}
	variableNameIdMap := make(map[int]string)
	for _, variable := range variables {
		variableNameIdMap[variable.Id] = variable.Name
		variableDto := &bean.StepVariableDto{
			Id:                        variable.Id,
			Name:                      variable.Name,
			Format:                    variable.Format,
			Description:               variable.Description,
			IsExposed:                 variable.IsExposed,
			AllowEmptyValue:           variable.AllowEmptyValue,
			DefaultValue:              variable.DefaultValue,
			Value:                     variable.Value,
			ValueType:                 variable.ValueType,
			PreviousStepIndex:         variable.PreviousStepIndex,
			ReferenceVariableName:     variable.ReferenceVariableName,
			ReferenceVariableStage:    variable.ReferenceVariableStage,
			VariableStepIndexInPlugin: variable.VariableStepIndexInPlugin,
		}
		if variable.VariableType == repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_INPUT {
			inputVariablesDto = append(inputVariablesDto, variableDto)
		} else if variable.VariableType == repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_OUTPUT {
			outputVariablesDto = append(outputVariablesDto, variableDto)
		}
	}
	conditions, err := impl.pipelineStageRepository.GetConditionsByStepId(stepId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting conditions by stepId", "err", err, "stepId", stepId)
		return nil, nil, nil, err
	}
	for _, condition := range conditions {
		conditionDto := &bean.ConditionDetailDto{
			Id:                  condition.Id,
			ConditionalOperator: condition.ConditionalOperator,
			ConditionalValue:    condition.ConditionalValue,
			ConditionType:       condition.ConditionType,
		}
		varName, ok := variableNameIdMap[condition.ConditionVariableId]
		if ok {
			conditionDto.ConditionOnVariable = varName
		}
		conditionsDto = append(conditionsDto, conditionDto)
	}
	return inputVariablesDto, outputVariablesDto, conditionsDto, nil
}

//GetCiPipelineStageData and related methods ends

// CreatePipelineStage and related methods starts
func (impl *PipelineStageServiceImpl) CreatePipelineStage(stageReq *bean.PipelineStageDto, stageType repository.PipelineStageType, pipelineId int, userId int32) error {
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	stage := &repository.PipelineStage{
		Name:        stageReq.Name,
		Description: stageReq.Description,
		Type:        stageType,
		Deleted:     false,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
	if stageType == repository.PIPELINE_STAGE_TYPE_PRE_CI || stageType == repository.PIPELINE_STAGE_TYPE_POST_CI {
		stage.CiPipelineId = pipelineId
	} else if stageType == repository.PIPELINE_STAGE_TYPE_PRE_CD || stageType == repository.PIPELINE_STAGE_TYPE_POST_CD {
		stage.CdPipelineId = pipelineId
	} else {
		return errors.New("unknown stage type")
	}
	stage, err = impl.pipelineStageRepository.CreatePipelineStage(stage, tx)
	if err != nil {
		impl.logger.Errorw("error in creating entry for pipeline Stage", "err", err, "pipelineStage", stage)
		return err
	}
	stageReq.Id = stage.Id
	indexNameString := make(map[int]string)
	for _, step := range stageReq.Steps {
		indexNameString[step.Index] = step.Name
	}
	//creating stage steps and all related data
	err = impl.CreateStageSteps(stageReq.Steps, stage.Id, userId, indexNameString, tx)
	if err != nil {
		impl.logger.Errorw("error in creating stage steps for ci stage", "err", err, "stageId", stage.Id)
		return err
	}

	err = impl.extractAndMapScopedVariables(stageReq, userId, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in tx commit", "err", err)
		return err
	}
	return nil
}

func (impl *PipelineStageServiceImpl) CreateStageSteps(steps []*bean.PipelineStageStepDto, stageId int, userId int32, indexNameString map[int]string, tx *pg.Tx) error {
	for _, step := range steps {
		//setting dependentStep detail
		var dependentOnStep string
		if step.Index > 1 {
			//since starting index is independent of any step we will be setting dependent detail for further indexes
			name, ok := indexNameString[step.Index-1]
			if ok {
				dependentOnStep = name
			}
		}
		var stepId int
		var inputVariables []*bean.StepVariableDto
		var outputVariables []*bean.StepVariableDto
		var conditionDetails []*bean.ConditionDetailDto
		if step.StepType == repository.PIPELINE_STEP_TYPE_INLINE {
			inlineStepDetail := step.InlineStepDetail
			//creating script entry first, because step entry needs scriptId
			scriptEntryId, err := impl.CreateScriptAndMappingForInlineStep(inlineStepDetail, userId, tx)
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
				OutputDirectoryPath: step.OutputDirectoryPath,
				DependentOnStep:     dependentOnStep,
				Deleted:             false,
				AuditLog: sql.AuditLog{
					CreatedOn: time.Now(),
					CreatedBy: userId,
					UpdatedOn: time.Now(),
					UpdatedBy: userId,
				},
				TriggerIfParentStageFail: step.TriggerIfParentStageFail,
			}
			inlineStep, err = impl.pipelineStageRepository.CreatePipelineStageStep(inlineStep, tx)
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
				OutputDirectoryPath: step.OutputDirectoryPath,
				DependentOnStep:     dependentOnStep,
				Deleted:             false,
				AuditLog: sql.AuditLog{
					CreatedOn: time.Now(),
					CreatedBy: userId,
					UpdatedOn: time.Now(),
					UpdatedBy: userId,
				},
				TriggerIfParentStageFail: step.TriggerIfParentStageFail,
			}
			refPluginStep, err := impl.pipelineStageRepository.CreatePipelineStageStep(refPluginStep, tx)
			if err != nil {
				impl.logger.Errorw("error in creating ref plugin step", "err", err, "step", refPluginStep)
				return err
			}
			stepId = refPluginStep.Id
			inputVariables = refPluginStepDetail.InputVariables
			outputVariables = refPluginStepDetail.OutputVariables
			conditionDetails = refPluginStepDetail.ConditionDetails
		}
		inputVariablesRepo, outputVariablesRepo, err := impl.CreateInputAndOutputVariables(stepId, inputVariables, outputVariables, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in creating variables for step", "err", err, "stepId", stepId, "inputVariables", inputVariables, "outputVariables", outputVariables)
			return err
		}
		if len(conditionDetails) > 0 {
			err = impl.CreateConditions(stepId, conditionDetails, inputVariablesRepo, outputVariablesRepo, userId, tx)
			if err != nil {
				impl.logger.Errorw("error in creating conditions", "err", err, "conditionDetails", conditionDetails)
				return err
			}
		}
	}
	return nil
}

func (impl *PipelineStageServiceImpl) CreateScriptAndMappingForInlineStep(inlineStepDetail *bean.InlineStepDetailDto, userId int32, tx *pg.Tx) (scriptId int, err error) {
	scriptEntry := &repository.PluginPipelineScript{
		Script:                   inlineStepDetail.Script,
		Type:                     inlineStepDetail.ScriptType,
		StoreScriptAt:            inlineStepDetail.StoreScriptAt,
		DockerfileExists:         inlineStepDetail.DockerfileExists,
		MountPath:                inlineStepDetail.MountPath,
		MountCodeToContainer:     inlineStepDetail.MountCodeToContainer,
		MountCodeToContainerPath: inlineStepDetail.MountCodeToContainerPath,
		MountDirectoryFromHost:   inlineStepDetail.MountDirectoryFromHost,
		ContainerImagePath:       inlineStepDetail.ContainerImagePath,
		ImagePullSecretType:      inlineStepDetail.ImagePullSecretType,
		ImagePullSecret:          inlineStepDetail.ImagePullSecret,
		Deleted:                  false,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
	scriptEntry, err = impl.pipelineStageRepository.CreatePipelineScript(scriptEntry, tx)
	if err != nil {
		impl.logger.Errorw("error in creating script entry for inline step", "err", err, "scriptEntry", scriptEntry)
		return 0, err
	}
	var scriptMap []repository.ScriptPathArgPortMapping
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
		scriptMap = append(scriptMap, repositoryEntry)
	}
	for _, commandArgMap := range inlineStepDetail.CommandArgsMap {
		repositoryEntry := repository.ScriptPathArgPortMapping{
			TypeOfMapping: repository2.SCRIPT_MAPPING_TYPE_DOCKER_ARG,
			Command:       commandArgMap.Command,
			Args:          commandArgMap.Args,
			ScriptId:      scriptEntry.Id,
			Deleted:       false,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: userId,
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		scriptMap = append(scriptMap, repositoryEntry)
	}
	for _, portMap := range inlineStepDetail.PortMap {
		repositoryEntry := repository.ScriptPathArgPortMapping{
			TypeOfMapping:   repository2.SCRIPT_MAPPING_TYPE_PORT,
			PortOnLocal:     portMap.PortOnLocal,
			PortOnContainer: portMap.PortOnContainer,
			ScriptId:        scriptEntry.Id,
			Deleted:         false,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: userId,
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		scriptMap = append(scriptMap, repositoryEntry)
	}
	if len(scriptMap) > 0 {
		err = impl.pipelineStageRepository.CreateScriptMapping(scriptMap, tx)
		if err != nil {
			impl.logger.Errorw("error in creating script mappings", "err", err, "scriptMappings", scriptMap)
			return 0, err
		}
	}
	return scriptEntry.Id, nil
}

func (impl *PipelineStageServiceImpl) CreateInputAndOutputVariables(stepId int, inputVariables []*bean.StepVariableDto, outputVariables []*bean.StepVariableDto, userId int32, tx *pg.Tx) (inputVariablesRepo []repository.PipelineStageStepVariable, outputVariablesRepo []repository.PipelineStageStepVariable, err error) {
	if len(inputVariables) > 0 {
		//creating input variables
		inputVariablesRepo, err = impl.CreateVariablesEntryInDb(stepId, inputVariables, repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_INPUT, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in creating input variables for step", "err", err, "stepId", stepId)
			return nil, nil, err
		}
	}
	if len(outputVariables) > 0 {
		//creating output variables
		outputVariablesRepo, err = impl.CreateVariablesEntryInDb(stepId, outputVariables, repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_OUTPUT, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in creating output variables for step", "err", err, "stepId", stepId)
			return nil, nil, err
		}
	}
	return inputVariablesRepo, outputVariablesRepo, nil
}

func (impl *PipelineStageServiceImpl) CreateConditions(stepId int, conditions []*bean.ConditionDetailDto, inputVariablesRepo []repository.PipelineStageStepVariable, outputVariablesRepo []repository.PipelineStageStepVariable, userId int32, tx *pg.Tx) error {
	variableNameIdMap := make(map[string]int)
	for _, inVar := range inputVariablesRepo {
		variableNameIdMap[inVar.Name] = inVar.Id
	}
	for _, outVar := range outputVariablesRepo {
		variableNameIdMap[outVar.Name] = outVar.Id
	}
	//creating conditions
	_, err := impl.CreateConditionsEntryInDb(stepId, conditions, variableNameIdMap, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in creating conditions for step", "err", err, "stepId", stepId)
		return err
	}
	return nil
}

func (impl *PipelineStageServiceImpl) CreateVariablesEntryInDb(stepId int, variables []*bean.StepVariableDto, variableType repository.PipelineStageStepVariableType, userId int32, tx *pg.Tx) ([]repository.PipelineStageStepVariable, error) {
	var variablesRepo []repository.PipelineStageStepVariable
	var err error
	for _, v := range variables {
		inVarRepo := repository.PipelineStageStepVariable{
			PipelineStageStepId:       stepId,
			Name:                      v.Name,
			Format:                    v.Format,
			Description:               v.Description,
			IsExposed:                 v.IsExposed,
			AllowEmptyValue:           v.AllowEmptyValue,
			DefaultValue:              v.DefaultValue,
			Value:                     v.Value,
			ValueType:                 v.ValueType,
			VariableType:              variableType,
			PreviousStepIndex:         v.PreviousStepIndex,
			ReferenceVariableName:     v.ReferenceVariableName,
			ReferenceVariableStage:    v.ReferenceVariableStage,
			VariableStepIndexInPlugin: v.VariableStepIndexInPlugin,
			Deleted:                   false,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: userId,
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		variablesRepo = append(variablesRepo, inVarRepo)
	}
	if len(variablesRepo) > 0 {
		//saving variables
		variablesRepo, err = impl.pipelineStageRepository.CreatePipelineStageStepVariables(variablesRepo, tx)
		if err != nil {
			impl.logger.Errorw("error in creating variables for pipeline stage steps", "err", err, "variables", variablesRepo)
			return nil, err
		}
	}
	return variablesRepo, nil
}

func (impl *PipelineStageServiceImpl) CreateConditionsEntryInDb(stepId int, conditions []*bean.ConditionDetailDto, variableNameIdMap map[string]int, userId int32, tx *pg.Tx) ([]repository.PipelineStageStepCondition, error) {
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
	if len(conditionsRepo) > 0 {
		//saving conditions
		conditionsRepo, err = impl.pipelineStageRepository.CreatePipelineStageStepConditions(conditionsRepo, tx)
		if err != nil {
			impl.logger.Errorw("error in creating pipeline stage step conditions", "err", err, "conditionsRepo", conditionsRepo)
			return nil, err
		}
	}
	return conditionsRepo, nil
}

//CreateCiStage and related methods ends

// UpdatePipelineStage and related methods starts
func (impl *PipelineStageServiceImpl) UpdatePipelineStage(stageReq *bean.PipelineStageDto, stageType repository.PipelineStageType, pipelineId int, userId int32) error {
	var stageOld *repository.PipelineStage
	var err error
	if stageType == repository.PIPELINE_STAGE_TYPE_PRE_CI || stageType == repository.PIPELINE_STAGE_TYPE_POST_CI {
		//getting stage by stageType and ciPipelineId
		stageOld, err = impl.pipelineStageRepository.GetCiStageByCiPipelineIdAndStageType(pipelineId, stageType)
	} else if stageType == repository.PIPELINE_STAGE_TYPE_PRE_CD || stageType == repository.PIPELINE_STAGE_TYPE_POST_CD {
		stageOld, err = impl.pipelineStageRepository.GetCdStageByCdPipelineIdAndStageType(pipelineId, stageType)
	}
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting stageId by pipelineId and stageType", "err", err, "pipelineId", pipelineId, "stageType", stageType)
		return err
	}

	//if stage is present with 0 stage steps, delete the stage and create new stage
	//handle corrupt data (https://github.com/devtron-labs/devtron/issues/3826)
	createNewPipStage := false
	if err == nil && stageOld != nil {
		stageReq.Id = stageOld.Id
		err, createNewPipStage = impl.DeletePipelineStageIfReq(stageReq, userId)
		if err != nil {
			impl.logger.Errorw("error in deleting the corrupted pipeline stage", "err", err, "pipelineStageReq", stageReq)
			return err
		}
	}

	if err == pg.ErrNoRows || createNewPipStage {
		//no stage found, creating new stage
		stageReq.Id = 0
		if len(stageReq.Steps) > 0 {
			err = impl.CreatePipelineStage(stageReq, stageType, pipelineId, userId)
			if err != nil {
				impl.logger.Errorw("error in creating new pipeline stage", "err", err, "pipelineStageReq", stageReq)
				return err
			}
		}
	} else {
		//stageId found, to handle as an update request
		stageReq.Id = stageOld.Id
		stageUpdateReq := stageOld
		stageUpdateReq.Name = stageReq.Name
		stageUpdateReq.Description = stageReq.Description
		stageUpdateReq.UpdatedBy = userId
		stageUpdateReq.UpdatedOn = time.Now()
		_, err = impl.pipelineStageRepository.UpdatePipelineStage(stageUpdateReq)
		if err != nil {
			impl.logger.Errorw("error in updating entry for pipelineStage", "err", err, "pipelineStage", stageUpdateReq)
			return err
		}
		// filtering(if steps/variables/conditions are updated or newly added) and performing relevant actions on update request
		err = impl.FilterAndActOnStepsInPipelineStageUpdateRequest(stageReq, userId)
		if err != nil {
			impl.logger.Errorw("error in filtering and performing actions on steps in pipelineStage update request", "err", err, "stageReq", stageReq)
			return err
		}

		err := impl.extractAndMapScopedVariables(stageReq, userId, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeletePipelineStageIfReq function is used to delete corrupted pipelineStage data
// , there was a bug (https://github.com/devtron-labs/devtron/issues/3826) where we were not deleting pipeline stage entry even after deleting all the pipelineStageSteps
// , this will delete those pipelineStage entry
func (impl *PipelineStageServiceImpl) DeletePipelineStageIfReq(stageReq *bean.PipelineStageDto, userId int32) (error, bool) {
	steps, err := impl.pipelineStageRepository.GetAllStepsByStageId(stageReq.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching pipeline stage steps by pipelineStageId", "err", err, "pipelineStageID", stageReq.Id)
		return err, false
	}
	if err == pg.ErrNoRows || len(steps) == 0 {
		impl.logger.Infow("deletePipelineStageWithTx ", "stageReq", stageReq, "userId", userId)
		err = impl.deletePipelineStageWithTx(stageReq, userId)
		if err != nil {
			impl.logger.Errorw("error in deleting the corrupted pipeline stage", "err", err, "pipelineStageId", stageReq.Id)
			return err, false
		}

	}

	return nil, len(steps) == 0
}

// deletePipelineStageWithTx transaction wrapper method around DeletePipelineStage method
func (impl *PipelineStageServiceImpl) deletePipelineStageWithTx(stageReq *bean.PipelineStageDto, userId int32) error {
	tx, err := impl.pipelineStageRepository.GetConnection().Begin()
	if err != nil {
		impl.logger.Errorw("error in starting transaction", "err", err, "stageReq", stageReq, "userId", userId)
		return err
	}
	defer tx.Rollback()
	err = impl.DeletePipelineStage(stageReq, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in DeletePipelineStage", "err", err, "stageReq", stageReq, "userId", userId)
		return err
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err, "stageReq", stageReq, "userId", userId)
	}
	return err
}

func (impl *PipelineStageServiceImpl) FilterAndActOnStepsInPipelineStageUpdateRequest(stageReq *bean.PipelineStageDto, userId int32) error {
	//getting all stepIds for current active (steps not deleted) steps
	activeStepIds, err := impl.pipelineStageRepository.GetStepIdsByStageId(stageReq.Id)
	if err != nil {
		impl.logger.Errorw("error in getting stepIds by stageId", "err", err, "stageId", stageReq.Id)
		return err
	}
	//creating map of the above ids
	activeStepIdsMap := make(map[int]bool)
	for _, activeStepId := range activeStepIds {
		activeStepIdsMap[activeStepId] = true
	}

	var stepsToBeCreated []*bean.PipelineStageStepDto
	var stepsToBeUpdated []*bean.PipelineStageStepDto
	var activeStepIdsPresentInReq []int
	idsOfStepsToBeUpdated := make(map[int]bool)
	indexNameString := make(map[int]string)
	for _, step := range stageReq.Steps {
		indexNameString[step.Index] = step.Name
		_, ok := activeStepIdsMap[step.Id]
		_, ok2 := idsOfStepsToBeUpdated[step.Id]
		if ok && !ok2 {
			// stepId present in current active steps and not repeated in request, will be updated
			stepsToBeUpdated = append(stepsToBeUpdated, step)
			idsOfStepsToBeUpdated[step.Id] = true
			activeStepIdsPresentInReq = append(activeStepIdsPresentInReq, step.Id)
		} else {
			// stepId present in current active steps but repeated in request, will be handled as new step
			// OR
			// stepId not present in current active steps, will be handled as new step
			step.Id = 0
			stepsToBeCreated = append(stepsToBeCreated, step)
		}
	}
	if len(activeStepIdsPresentInReq) > 0 {
		// deleting all steps which are currently active but not present in update request
		err = impl.pipelineStageRepository.MarkStepsDeletedExcludingActiveStepsInUpdateReq(activeStepIdsPresentInReq, stageReq.Id)
		if err != nil {
			impl.logger.Errorw("error in marking all steps deleted excluding active steps in update req", "err", err, "activeStepIdsPresentInReq", activeStepIdsPresentInReq)
			return err
		}
	} else {
		//since no step is present in update request, deleting the stage and all the related data of this stage
		err = impl.deletePipelineStageWithTx(stageReq, userId)
		if err != nil {
			impl.logger.Errorw("error in marking all steps deleted by stageId", "err", err, "stageId", stageReq.Id)
			return err
		}
	}
	if len(stepsToBeCreated) > 0 {
		//creating new steps
		err = impl.CreateStageSteps(stepsToBeCreated, stageReq.Id, userId, indexNameString, nil)
		if err != nil {
			impl.logger.Errorw("error in creating stage steps for ci stage", "err", err, "stageId", stageReq.Id)
			return err
		}
	}
	if len(stepsToBeUpdated) > 0 {
		//updating steps
		err = impl.UpdateStageSteps(stepsToBeUpdated, userId, stageReq.Id, indexNameString)
		if err != nil {
			impl.logger.Errorw("error in updating stage steps for ci stage", "err", err)
			return err
		}
	}
	return nil
}

func (impl *PipelineStageServiceImpl) UpdateStageSteps(steps []*bean.PipelineStageStepDto, userId int32, stageId int, indexNameString map[int]string) error {
	for _, step := range steps {
		//setting dependentStep detail
		var dependentOnStep string
		if step.Index > 1 {
			//since starting index is independent of any step we will be setting dependent detail for further indexes
			name, ok := indexNameString[step.Index-1]
			if ok {
				dependentOnStep = name
			}
		}
		//getting saved step from db
		savedStep, err := impl.pipelineStageRepository.GetStepById(step.Id)
		if err != nil {
			impl.logger.Errorw("error in getting saved step from db", "err", err, "stepId", step.Id)
			return err
		}
		stepUpdateReq := &repository.PipelineStageStep{
			Id:                  step.Id,
			PipelineStageId:     stageId,
			Name:                step.Name,
			Description:         step.Description,
			Index:               step.Index,
			StepType:            step.StepType,
			OutputDirectoryPath: step.OutputDirectoryPath,
			DependentOnStep:     dependentOnStep,
			Deleted:             false,
			AuditLog: sql.AuditLog{
				CreatedOn: savedStep.CreatedOn,
				CreatedBy: savedStep.CreatedBy,
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
			TriggerIfParentStageFail: step.TriggerIfParentStageFail,
		}
		var inputVariables []*bean.StepVariableDto
		var outputVariables []*bean.StepVariableDto
		var conditionDetails []*bean.ConditionDetailDto
		//handling changes in stepType
		if step.StepType == repository.PIPELINE_STEP_TYPE_REF_PLUGIN {
			if savedStep.StepType == repository.PIPELINE_STEP_TYPE_INLINE {
				//step changed from inline to ref plugin, delete script and its mappings
				idOfScriptToBeDeleted := savedStep.ScriptId

				//deleting script
				err = impl.pipelineStageRepository.MarkScriptDeletedById(idOfScriptToBeDeleted)
				if err != nil {
					impl.logger.Errorw("error in marking script deleted by id", "err", err, "scriptId", idOfScriptToBeDeleted)
					return err
				}
				//deleting mappings
				err = impl.pipelineStageRepository.MarkScriptMappingDeletedByScriptId(idOfScriptToBeDeleted)
				if err != nil {
					impl.logger.Errorw("error in marking script mappings deleted by scriptId", "err", err)
					return err
				}
			}
			//updating ref plugin id in step update req
			stepUpdateReq.RefPluginId = step.RefPluginStepDetail.PluginId
			inputVariables = step.RefPluginStepDetail.InputVariables
			outputVariables = step.RefPluginStepDetail.OutputVariables
			conditionDetails = step.RefPluginStepDetail.ConditionDetails
		} else if step.StepType == repository.PIPELINE_STEP_TYPE_INLINE {
			if savedStep.StepType == repository.PIPELINE_STEP_TYPE_REF_PLUGIN {
				//step changed from ref plugin to inline, create script and mapping
				scriptEntryId, err := impl.CreateScriptAndMappingForInlineStep(step.InlineStepDetail, userId, nil)
				if err != nil {
					impl.logger.Errorw("error in creating script and mapping for inline step", "err", err, "inlineStepDetail", step.InlineStepDetail)
					return err
				}
				//updating scriptId in step update req
				stepUpdateReq.ScriptId = scriptEntryId
			} else {
				//update script and its mappings
				err = impl.UpdateScriptAndMappingForInlineStep(step.InlineStepDetail, savedStep.ScriptId, userId)
				if err != nil {
					impl.logger.Errorw("error in updating script and its mapping", "err", err)
					return err
				}
				stepUpdateReq.ScriptId = savedStep.ScriptId
			}
			inputVariables = step.InlineStepDetail.InputVariables
			outputVariables = step.InlineStepDetail.OutputVariables
			conditionDetails = step.InlineStepDetail.ConditionDetails
		}
		//updating step
		_, err = impl.pipelineStageRepository.UpdatePipelineStageStep(stepUpdateReq)
		if err != nil {
			impl.logger.Errorw("error in updating pipeline stage step", "err", err, "stepUpdateReq", stepUpdateReq)
			return err
		}
		inputVarNameIdMap, outputVarNameIdMap, err := impl.UpdateInputAndOutputVariables(step.Id, inputVariables, outputVariables, userId)
		if err != nil {
			impl.logger.Errorw("error in updating variables for step", "err", err, "stepId", step.Id, "inputVariables", inputVariables, "outputVariables", outputVariables)
			return err
		}
		//combining both maps
		varNameIdMap := inputVarNameIdMap
		for k, v := range outputVarNameIdMap {
			varNameIdMap[k] = v
		}
		//updating conditions
		_, err = impl.UpdatePipelineStageStepConditions(step.Id, conditionDetails, varNameIdMap, userId)
		if err != nil {
			impl.logger.Errorw("error in updating step conditions", "err", err)
			return err
		}

	}
	return nil
}

func (impl *PipelineStageServiceImpl) UpdateScriptAndMappingForInlineStep(inlineStepDetail *bean.InlineStepDetailDto, scriptId int, userId int32) (err error) {
	scriptEntry := &repository.PluginPipelineScript{
		Id:                       scriptId,
		Script:                   inlineStepDetail.Script,
		StoreScriptAt:            inlineStepDetail.StoreScriptAt,
		Type:                     inlineStepDetail.ScriptType,
		DockerfileExists:         inlineStepDetail.DockerfileExists,
		MountPath:                inlineStepDetail.MountPath,
		MountCodeToContainer:     inlineStepDetail.MountCodeToContainer,
		MountCodeToContainerPath: inlineStepDetail.MountCodeToContainerPath,
		MountDirectoryFromHost:   inlineStepDetail.MountDirectoryFromHost,
		ContainerImagePath:       inlineStepDetail.ContainerImagePath,
		ImagePullSecretType:      inlineStepDetail.ImagePullSecretType,
		ImagePullSecret:          inlineStepDetail.ImagePullSecret,
		Deleted:                  false,
		AuditLog: sql.AuditLog{
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
	scriptEntry, err = impl.pipelineStageRepository.UpdatePipelineScript(scriptEntry)
	if err != nil {
		impl.logger.Errorw("error in updating script entry for inline step", "err", err, "scriptEntry", scriptEntry)
		return err
	}
	//marking all old scripts deleted
	err = impl.pipelineStageRepository.MarkScriptMappingDeletedByScriptId(scriptId)
	if err != nil {
		impl.logger.Errorw("error in marking script mappings deleted by scriptId", "err", err, "scriptId", scriptId)
		return err
	}
	//creating new mappings
	var scriptMap []repository.ScriptPathArgPortMapping
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
		scriptMap = append(scriptMap, repositoryEntry)
	}
	for _, commandArgMap := range inlineStepDetail.CommandArgsMap {
		repositoryEntry := repository.ScriptPathArgPortMapping{
			TypeOfMapping: repository2.SCRIPT_MAPPING_TYPE_DOCKER_ARG,
			Command:       commandArgMap.Command,
			Args:          commandArgMap.Args,
			ScriptId:      scriptEntry.Id,
			Deleted:       false,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: userId,
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		scriptMap = append(scriptMap, repositoryEntry)
	}
	for _, portMap := range inlineStepDetail.PortMap {
		repositoryEntry := repository.ScriptPathArgPortMapping{
			TypeOfMapping:   repository2.SCRIPT_MAPPING_TYPE_PORT,
			PortOnLocal:     portMap.PortOnLocal,
			PortOnContainer: portMap.PortOnContainer,
			ScriptId:        scriptEntry.Id,
			Deleted:         false,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: userId,
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		scriptMap = append(scriptMap, repositoryEntry)
	}
	if len(scriptMap) > 0 {
		err = impl.pipelineStageRepository.CreateScriptMapping(scriptMap, nil)
		if err != nil {
			impl.logger.Errorw("error in creating script mappings", "err", err, "scriptMappings", scriptMap)
			return err
		}
	}
	return nil
}

func (impl *PipelineStageServiceImpl) UpdateInputAndOutputVariables(stepId int, inputVariables []*bean.StepVariableDto, outputVariables []*bean.StepVariableDto, userId int32) (inputVarNameIdMap map[string]int, outputVarNameIdMap map[string]int, err error) {
	//using tx for variables db operation
	dbConnection := impl.pipelineStageRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in starting tx", "err", err)
		return nil, nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	//updating input variable
	inputVarNameIdMap, err = impl.UpdatePipelineStageStepVariables(stepId, inputVariables, repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_INPUT, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in updating input variables", "err", err)
		return nil, nil, err
	}
	//updating output variable
	outputVarNameIdMap, err = impl.UpdatePipelineStageStepVariables(stepId, outputVariables, repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_OUTPUT, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in updating output variables", "err", err)
		return nil, nil, err
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in tx commit", "err", err)
		return nil, nil, err
	}
	return inputVarNameIdMap, outputVarNameIdMap, nil
}

func (impl *PipelineStageServiceImpl) UpdatePipelineStageStepVariables(stepId int, variables []*bean.StepVariableDto, variableType repository.PipelineStageStepVariableType, userId int32, tx *pg.Tx) (map[string]int, error) {
	//getting ids of all current active variables
	variableIds, err := impl.pipelineStageRepository.GetVariableIdsByStepIdAndVariableType(stepId, variableType)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting variablesIds by stepId", "err", err, "stepId", stepId)
		return nil, err
	}
	activeVariableIdsMap := make(map[int]bool)
	for _, variableId := range variableIds {
		activeVariableIdsMap[variableId] = true
	}
	var variablesToBeCreated []*bean.StepVariableDto
	var variablesToBeUpdated []*bean.StepVariableDto
	var activeVariableIdsPresentInReq []int
	idsOfVariablesToBeUpdated := make(map[int]bool)
	for _, variable := range variables {
		_, ok := activeVariableIdsMap[variable.Id]
		_, ok2 := idsOfVariablesToBeUpdated[variable.Id]
		if ok && !ok2 {
			// variableId present in current active variables and not repeated in request, will be updated
			variablesToBeUpdated = append(variablesToBeUpdated, variable)
			idsOfVariablesToBeUpdated[variable.Id] = true
			activeVariableIdsPresentInReq = append(activeVariableIdsPresentInReq, variable.Id)
		} else {
			// variableId present in current active variables but repeated in request, will be handled as new variable
			// OR
			// variableId not present in current active variables, will be handled as new variable
			variable.Id = 0
			variablesToBeCreated = append(variablesToBeCreated, variable)
		}
	}
	if len(activeVariableIdsPresentInReq) > 0 {
		// deleting all variables which are currently active but not present in update request
		err = impl.pipelineStageRepository.MarkVariablesDeletedExcludingActiveVariablesInUpdateReq(activeVariableIdsPresentInReq, stepId, variableType, tx)
		if err != nil {
			impl.logger.Errorw("error in marking all variables deleted excluding active variables in update req", "err", err, "activeVariableIdsPresentInReq", activeVariableIdsPresentInReq)
			return nil, err
		}
	} else {
		//deleting all variables by stepId, since no variable is present in update request
		err = impl.pipelineStageRepository.MarkVariablesDeletedByStepIdAndVariableType(stepId, variableType, tx)
		if err != nil {
			impl.logger.Errorw("error in marking all variables deleted by stepId", "err", err, "stepId", stepId)
			return nil, err
		}
	}
	var newVariables []repository.PipelineStageStepVariable
	if len(variablesToBeCreated) > 0 {
		//creating new variables
		newVariables, err = impl.CreateVariablesEntryInDb(stepId, variablesToBeCreated, variableType, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in creating variables", "err", err, "stepId", stepId)
			return nil, err
		}
	}
	variableNameIdMap := make(map[string]int)
	for _, inVar := range newVariables {
		variableNameIdMap[inVar.Name] = inVar.Id
	}

	//updating variables
	var variablesRepo []repository.PipelineStageStepVariable
	for _, v := range variablesToBeUpdated {
		variableNameIdMap[v.Name] = v.Id
		inVarRepo := repository.PipelineStageStepVariable{
			Id:                        v.Id,
			PipelineStageStepId:       stepId,
			Name:                      v.Name,
			Format:                    v.Format,
			Description:               v.Description,
			IsExposed:                 v.IsExposed,
			AllowEmptyValue:           v.AllowEmptyValue,
			DefaultValue:              v.DefaultValue,
			Value:                     v.Value,
			ValueType:                 v.ValueType,
			VariableType:              variableType,
			PreviousStepIndex:         v.PreviousStepIndex,
			ReferenceVariableName:     v.ReferenceVariableName,
			ReferenceVariableStage:    v.ReferenceVariableStage,
			VariableStepIndexInPlugin: v.VariableStepIndexInPlugin,
			Deleted:                   false,
			AuditLog: sql.AuditLog{
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		variablesRepo = append(variablesRepo, inVarRepo)
	}
	if len(variablesRepo) > 0 {
		variablesRepo, err = impl.pipelineStageRepository.UpdatePipelineStageStepVariables(variablesRepo, tx)
		if err != nil {
			impl.logger.Errorw("error in updating variables for pipeline stage steps", "err", err, "variables", variablesRepo)
			return nil, err
		}
	}
	return variableNameIdMap, nil
}

func (impl *PipelineStageServiceImpl) UpdatePipelineStageStepConditions(stepId int, conditions []*bean.ConditionDetailDto, variableNameIdMap map[string]int, userId int32) ([]repository.PipelineStageStepCondition, error) {
	//using tx for conditions db operation
	dbConnection := impl.pipelineStageRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in starting tx", "err", err)
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	//getting ids of all current active variables
	conditionIds, err := impl.pipelineStageRepository.GetConditionIdsByStepId(stepId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting variablesIds by stepId", "err", err, "stepId", stepId)
		return nil, err
	}
	activeConditionIdsMap := make(map[int]bool)
	for _, conditionId := range conditionIds {
		activeConditionIdsMap[conditionId] = true
	}
	var conditionsToBeCreated []*bean.ConditionDetailDto
	var conditionsToBeUpdated []*bean.ConditionDetailDto
	var activeConditionIdsPresentInReq []int
	idsOfConditionsToBeUpdated := make(map[int]bool)
	for _, condition := range conditions {
		_, ok := activeConditionIdsMap[condition.Id]
		_, ok2 := idsOfConditionsToBeUpdated[condition.Id]
		if ok && !ok2 {
			// conditionId present in current active conditions and not repeated in request, will be updated
			conditionsToBeUpdated = append(conditionsToBeUpdated, condition)
			idsOfConditionsToBeUpdated[condition.Id] = true
			activeConditionIdsPresentInReq = append(activeConditionIdsPresentInReq, condition.Id)
		} else {
			// conditionId present in current active conditions but repeated in request, will be handled as new condition
			// OR
			// conditionId not present in current active conditions, will be handled as new condition
			condition.Id = 0
			conditionsToBeCreated = append(conditionsToBeCreated, condition)
		}
	}
	if len(activeConditionIdsPresentInReq) > 0 {
		// deleting all conditions which are currently active but not present in update request
		err = impl.pipelineStageRepository.MarkConditionsDeletedExcludingActiveVariablesInUpdateReq(activeConditionIdsPresentInReq, stepId, tx)
		if err != nil {
			impl.logger.Errorw("error in marking all conditions deleted excluding active conditions in update req", "err", err, "activeConditionIdsPresentInReq", activeConditionIdsPresentInReq)
			return nil, err
		}
	} else {
		// deleting all current conditions, since no condition is present in update request
		err = impl.pipelineStageRepository.MarkConditionsDeletedByStepId(stepId, tx)
		if err != nil {
			impl.logger.Errorw("error in marking all conditions deleted by stepId", "err", err, "stepId", stepId)
			return nil, err
		}
	}
	//creating new conditions
	_, err = impl.CreateConditionsEntryInDb(stepId, conditionsToBeCreated, variableNameIdMap, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in creating conditions", "err", err, "stepId", stepId)
		return nil, err
	}

	var conditionsRepo []repository.PipelineStageStepCondition
	for _, condition := range conditionsToBeUpdated {
		varId := variableNameIdMap[condition.ConditionOnVariable]
		conditionRepo := repository.PipelineStageStepCondition{
			Id:                  condition.Id,
			PipelineStageStepId: stepId,
			ConditionVariableId: varId,
			ConditionType:       condition.ConditionType,
			ConditionalOperator: condition.ConditionalOperator,
			ConditionalValue:    condition.ConditionalValue,
			Deleted:             false,
			AuditLog: sql.AuditLog{
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		conditionsRepo = append(conditionsRepo, conditionRepo)
	}
	if len(conditionsRepo) > 0 {
		//updating conditions
		conditionsRepo, err = impl.pipelineStageRepository.UpdatePipelineStageStepConditions(conditionsRepo, tx)
		if err != nil {
			impl.logger.Errorw("error in updating pipeline stage step conditions", "err", err, "conditionsRepo", conditionsRepo)
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in tx commit", "err", err)
		return nil, err
	}
	return conditionsRepo, nil
}

//UpdateCiStage and related methods ends

// DeletePipelineStage and related methods starts
func (impl *PipelineStageServiceImpl) DeletePipelineStage(stageReq *bean.PipelineStageDto, userId int32, tx *pg.Tx) error {
	//marking stage deleted
	err := impl.pipelineStageRepository.MarkPipelineStageDeletedById(stageReq.Id, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in marking pipeline stage deleted", "err", err, "pipelineStageId", stageReq.Id)
		return err
	}
	//marking all steps deleted
	err = impl.pipelineStageRepository.MarkPipelineStageStepsDeletedByStageId(stageReq.Id, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in marking pipeline stage steps deleted by stageId", "err", err, "pipelineStageId", stageReq.Id)
		return err
	}
	//getting scriptIds by stageId
	scriptIds, err := impl.pipelineStageRepository.GetScriptIdsByStageId(stageReq.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting scriptIds by stageId", "err", err, "stageId", stageReq.Id)
		return err
	}
	if len(scriptIds) > 0 {
		//marking all scripts deleted
		err = impl.pipelineStageRepository.MarkPipelineScriptsDeletedByIds(scriptIds, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in marking pipeline stage scripts deleted by scriptIds", "err", err, "scriptIds", scriptIds)
			return err
		}
	}
	//getting scriptMappingIds by stageId
	scriptMappingIds, err := impl.pipelineStageRepository.GetScriptMappingIdsByStageId(stageReq.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting scriptMappingIds by stageId", "err", err, "stageId", stageReq.Id)
		return err
	}
	if len(scriptMappingIds) > 0 {
		//marking all script mappings deleted
		err = impl.pipelineStageRepository.MarkPipelineScriptMappingsDeletedByIds(scriptMappingIds, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in marking pipeline script mapping deleted by scriptMappingIds", "err", err, "scriptMappingIds", scriptMappingIds)
			return err
		}
	}
	//getting variableIds by stageId
	variableIds, err := impl.pipelineStageRepository.GetVariableIdsByStageId(stageReq.Id)
	if err != nil {
		impl.logger.Errorw("error in getting variableIds by stageId", "err", err, "stageId", stageReq.Id)
		return err
	}
	if len(variableIds) > 0 {
		//marking all variables deleted
		err = impl.pipelineStageRepository.MarkPipelineStageStepVariablesDeletedByIds(variableIds, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in marking pipeline stage step variables deleted by variableIds", "err", err, "variableIds", variableIds)
			return err
		}
	}
	//getting conditionIds by stageId
	conditionIds, err := impl.pipelineStageRepository.GetConditionIdsByStageId(stageReq.Id)
	if err != nil {
		impl.logger.Errorw("error in getting conditionIds by stageId", "err", err, "stageId", stageReq.Id)
		return err
	}
	if len(conditionIds) > 0 {
		//marking all conditions deleted
		err = impl.pipelineStageRepository.MarkPipelineStageStepConditionDeletedByIds(conditionIds, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in marking pipeline stage step conditions deleted by conditionIds", "err", err, "conditionIds", conditionIds)
			return err
		}
	}

	//err = impl.variableEntityMappingService.DeleteMappingsForEntities([]repository3.Entity{{
	//	EntityType: repository3.EntityTypePipelineStage,
	//	EntityId:   stageReq.Id,
	//}}, userId, tx)
	err = impl.scopedVariableManager.RemoveMappedVariables(stageReq.Id, repository3.EntityTypePipelineStage, userId, tx)
	if err != nil {
		return err
	}
	return nil
}

//DeleteCiStage and related methods starts

// BuildPrePostAndRefPluginStepsDataForWfRequest and related methods starts
func (impl *PipelineStageServiceImpl) BuildPrePostAndRefPluginStepsDataForWfRequest(pipelineId int, stageType string, scope resourceQualifiers.Scope) (*bean.PrePostAndRefPluginStepsResponse, error) {
	//get all stages By pipelineId (it can be ciPipelineId or cdPipelineId)
	var pipelineStages []*repository.PipelineStage
	var err error
	if stageType == bean.CiStage {
		pipelineStages, err = impl.pipelineStageRepository.GetAllCiStagesByCiPipelineId(pipelineId)
	} else if stageType == preCdStage || stageType == postCdStage {
		//cdEvent
		//pipelineStages, err = impl.pipelineStageRepository.GetAllCdStagesByCdPipelineId(pipelineId)
		var pipelineStage *repository.PipelineStage
		pipelineStage, err = impl.pipelineStageRepository.GetCdStageByCdPipelineIdAndStageType(pipelineId, getPipelineStageFromStageType(stageType))
		pipelineStages = append(pipelineStages, pipelineStage)
	}
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting all ci stages by pipelineId", "err", err, "pipelineId", pipelineId, "stageType", stageType)
		return nil, err
	}
	var preCiSteps []*bean.StepObject
	var postCiSteps []*bean.StepObject
	var preCdSteps []*bean.StepObject
	var postCdSteps []*bean.StepObject
	var refPluginsData []*bean.RefPluginObject
	var refPluginIds []int
	var pipelineStageIds []int
	for _, pipelineStage := range pipelineStages {
		var refIds []int
		steps, refIds, err := impl.BuildPipelineStageDataForWfRequest(pipelineStage)
		switch pipelineStage.Type {
		case repository.PIPELINE_STAGE_TYPE_PRE_CI:
			preCiSteps = steps
			if err != nil {
				impl.logger.Errorw("error in getting pre ci steps data for wf request", "err", err, "ciStage", pipelineStage)
				return nil, err
			}
		case repository.PIPELINE_STAGE_TYPE_POST_CI:
			postCiSteps = steps
			if err != nil {
				impl.logger.Errorw("error in getting post ci steps data for wf request", "err", err, "ciStage", pipelineStage)
				return nil, err
			}
		case repository.PIPELINE_STAGE_TYPE_PRE_CD:
			preCdSteps = steps
			if err != nil {
				impl.logger.Errorw("error in getting post cd steps data for wf request", "err", err, "cdStage", pipelineStage)
				return nil, err
			}
		case repository.PIPELINE_STAGE_TYPE_POST_CD:
			postCdSteps = steps
			if err != nil {
				impl.logger.Errorw("error in getting post cd steps data for wf request", "err", err, "cdStage", pipelineStage)
				return nil, err
			}
		}
		refPluginIds = append(refPluginIds, refIds...)
		pipelineStageIds = append(pipelineStageIds, pipelineStage.Id)
	}
	if len(refPluginIds) > 0 {
		refPluginsData, err = impl.BuildRefPluginStepDataForWfRequest(refPluginIds)
		if err != nil {
			impl.logger.Errorw("error in building ref plugin step data", "err", err, "refPluginIds", refPluginIds)
			return nil, err
		}
	}
	unresolvedResponse := &bean.PrePostAndRefPluginStepsResponse{RefPluginData: refPluginsData}

	if stageType == bean.CiStage {
		unresolvedResponse.PreStageSteps = preCiSteps
		unresolvedResponse.PostStageSteps = postCiSteps
	} else {
		unresolvedResponse.PreStageSteps = preCdSteps
		unresolvedResponse.PostStageSteps = postCdSteps
	}

	resolvedResponse, err := impl.fetchScopedVariablesAndResolveTemplate(unresolvedResponse, pipelineStageIds, scope)
	if err != nil {
		impl.logger.Errorw("error in resolving stage request", "err", err, "pipelineStageIds", pipelineStageIds)
		return resolvedResponse, err
	}
	return resolvedResponse, nil
}

func getPipelineStageFromStageType(stageType string) repository.PipelineStageType {
	var pipelineStageType repository.PipelineStageType
	if stageType == preCdStage {
		pipelineStageType = repository.PIPELINE_STAGE_TYPE_PRE_CD
	} else if stageType == postCdStage {
		pipelineStageType = repository.PIPELINE_STAGE_TYPE_POST_CD
	}
	return pipelineStageType
}

func (impl *PipelineStageServiceImpl) BuildPipelineStageDataForWfRequest(pipelineStage *repository.PipelineStage) ([]*bean.StepObject, []int, error) {
	//getting all steps for this stage
	steps, err := impl.pipelineStageRepository.GetAllStepsByStageId(pipelineStage.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting all steps by stageId", "err", err, "stageId", pipelineStage.Id)
		return nil, nil, err
	}
	var stepsData []*bean.StepObject
	var refPluginIds []int
	for _, step := range steps {
		stepData, err := impl.BuildPipelineStepDataForWfRequest(step)
		if err != nil {
			impl.logger.Errorw("error in getting pipeline step data for WF request", "err", err)
			return nil, nil, err
		}
		if step.StepType == repository.PIPELINE_STEP_TYPE_REF_PLUGIN {
			refPluginIds = append(refPluginIds, stepData.RefPluginId)
		}
		stepsData = append(stepsData, stepData)
	}

	return stepsData, refPluginIds, nil
}

func (impl *PipelineStageServiceImpl) BuildRefPluginStepDataForWfRequest(refPluginIds []int) ([]*bean.RefPluginObject, error) {
	var refPluginsData []*bean.RefPluginObject
	pluginIdsEvaluated := make(map[int]bool)
	var refPluginIdsRequest []int
	for _, refPluginId := range refPluginIds {
		_, ok := pluginIdsEvaluated[refPluginId]
		if !ok {
			refPluginIdsRequest = append(refPluginIdsRequest, refPluginId)
			pluginIdsEvaluated[refPluginId] = true
		}
	}
	pluginIdStepsMap, err := impl.GetRefPluginStepsByIds(refPluginIdsRequest, pluginIdsEvaluated)
	if err != nil {
		impl.logger.Errorw("error in getting map of pluginIds and their steps", "err", err, "refPluginIdsRequest", refPluginIdsRequest)
		return nil, err
	}
	for pluginId, steps := range pluginIdStepsMap {
		refPluginData := &bean.RefPluginObject{
			Id:    pluginId,
			Steps: steps,
		}
		refPluginsData = append(refPluginsData, refPluginData)
	}
	return refPluginsData, nil
}

func (impl *PipelineStageServiceImpl) GetRefPluginStepsByIds(refPluginIds []int, pluginIdsEvaluated map[int]bool) (map[int][]*bean.StepObject, error) {
	//get refPlugin steps by ids
	pluginSteps, err := impl.globalPluginRepository.GetStepsByPluginIds(refPluginIds)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting steps by pluginIds", "err", err, "pluginIds", refPluginIds)
		return nil, err
	}

	pluginIdStepsMap := make(map[int][]*bean.StepObject)
	var nestedRefPluginIds []int
	for _, pluginStep := range pluginSteps {
		stepData, err := impl.BuildPluginStepDataForWfRequest(pluginStep)
		if err != nil {
			impl.logger.Errorw("error in getting plugin step data for WF request", "err", err)
			return nil, err
		}
		if pluginStep.StepType == repository2.PLUGIN_STEP_TYPE_REF_PLUGIN {
			_, ok := pluginIdsEvaluated[stepData.RefPluginId]
			if !ok {
				nestedRefPluginIds = append(nestedRefPluginIds, stepData.RefPluginId)
				pluginIdsEvaluated[stepData.RefPluginId] = true
			}
		}
		stepsInMap, ok := pluginIdStepsMap[pluginStep.PluginId]
		if ok {
			stepsInMap = append(stepsInMap, stepData)
			pluginIdStepsMap[pluginStep.PluginId] = stepsInMap
		} else {
			pluginIdStepsMap[pluginStep.PluginId] = []*bean.StepObject{stepData}
		}
	}
	if len(nestedRefPluginIds) > 0 {
		nestedPluginIdStepsMap, err := impl.GetRefPluginStepsByIds(nestedRefPluginIds, pluginIdsEvaluated)
		if err != nil {
			impl.logger.Errorw("error in getting nested ref plugin steps by ids", "err", err, "nestedRefPluginIds", nestedPluginIdStepsMap)
			return nil, err
		}
		for k, v := range nestedPluginIdStepsMap {
			pluginIdStepsMap[k] = v
		}
	}
	return pluginIdStepsMap, nil
}

func (impl *PipelineStageServiceImpl) BuildPipelineStepDataForWfRequest(step *repository.PipelineStageStep) (*bean.StepObject, error) {
	stepData := &bean.StepObject{
		Name:                     step.Name,
		Index:                    step.Index,
		StepType:                 string(step.StepType),
		ArtifactPaths:            step.OutputDirectoryPath,
		TriggerIfParentStageFail: step.TriggerIfParentStageFail,
	}
	if step.StepType == repository.PIPELINE_STEP_TYPE_INLINE {
		//get script and mapping data
		//getting script details for step
		scriptDetail, err := impl.pipelineStageRepository.GetScriptDetailById(step.ScriptId)
		if err != nil {
			impl.logger.Errorw("error in getting script details by id", "err", err, "scriptId", step.ScriptId)
			return nil, err
		}
		//getting script mapping details
		scriptMappings, err := impl.pipelineStageRepository.GetScriptMappingDetailByScriptId(step.ScriptId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting script mapping by scriptId", "err", err, "scriptId", step.ScriptId)
			return nil, err
		}
		var extraMappings []*bean.MountPath
		portMap := make(map[int]int)
		for _, scriptMapping := range scriptMappings {
			if scriptMapping.TypeOfMapping == repository2.SCRIPT_MAPPING_TYPE_FILE_PATH {
				extraMapping := &bean.MountPath{
					SourcePath:      scriptMapping.FilePathOnDisk,
					DestinationPath: scriptMapping.FilePathOnContainer,
				}
				extraMappings = append(extraMappings, extraMapping)
			} else if scriptMapping.TypeOfMapping == repository2.SCRIPT_MAPPING_TYPE_DOCKER_ARG {
				stepData.Command = scriptMapping.Command
				stepData.Args = scriptMapping.Args
			} else if scriptMapping.TypeOfMapping == repository2.SCRIPT_MAPPING_TYPE_PORT {
				portMap[scriptMapping.PortOnLocal] = scriptMapping.PortOnContainer
			}
		}
		stepData.ExecutorType = string(scriptDetail.Type)
		stepData.DockerImage = scriptDetail.ContainerImagePath
		stepData.Script = scriptDetail.Script
		if len(portMap) > 0 {
			stepData.ExposedPorts = portMap
		}
		if len(scriptDetail.StoreScriptAt) > 0 {
			stepData.CustomScriptMount = &bean.MountPath{
				DestinationPath: scriptDetail.StoreScriptAt,
			}
		}
		if scriptDetail.MountCodeToContainer && len(scriptDetail.MountCodeToContainerPath) > 0 {
			stepData.SourceCodeMount = &bean.MountPath{
				DestinationPath: scriptDetail.MountCodeToContainerPath,
			}
		}
		if scriptDetail.MountDirectoryFromHost && len(extraMappings) > 0 {
			stepData.ExtraVolumeMounts = extraMappings
		}
	} else if step.StepType == repository.PIPELINE_STEP_TYPE_REF_PLUGIN {
		stepData.ExecutorType = "PLUGIN" //added only to avoid un-marshaling issues at ci-runner side, will not be used
		stepData.RefPluginId = step.RefPluginId
	}
	inputVars, outputVars, triggerSkipConditions, successFailureConditions, err := impl.BuildVariableAndConditionDataForWfRequest(step.Id)
	if err != nil {
		impl.logger.Errorw("error in getting variable and conditions data for wf request", "err", err, "stepId", step.Id)
		return nil, err
	}
	stepData.InputVars = inputVars
	stepData.OutputVars = outputVars
	stepData.TriggerSkipConditions = triggerSkipConditions
	stepData.SuccessFailureConditions = successFailureConditions
	return stepData, nil
}

func (impl *PipelineStageServiceImpl) BuildVariableAndConditionDataForWfRequest(stepId int) ([]*bean.VariableObject, []*bean.VariableObject, []*bean.ConditionObject, []*bean.ConditionObject, error) {
	var inputVariables []*bean.VariableObject
	var outputVariables []*bean.VariableObject
	//getting all variables in the step
	variables, err := impl.pipelineStageRepository.GetVariablesByStepId(stepId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting variables by stepId", "err", err, "stepId", stepId)
		return nil, nil, nil, nil, err
	}
	variableNameIdMap := make(map[int]string)
	for _, variable := range variables {
		variableNameIdMap[variable.Id] = variable.Name
		variableData := &bean.VariableObject{
			Name:                       variable.Name,
			Format:                     string(variable.Format),
			ReferenceVariableStepIndex: variable.PreviousStepIndex,
			ReferenceVariableName:      variable.ReferenceVariableName,
			VariableStepIndexInPlugin:  variable.VariableStepIndexInPlugin,
		}
		if variable.ValueType == repository.PIPELINE_STAGE_STEP_VARIABLE_VALUE_TYPE_NEW {
			variableData.VariableType = bean.VARIABLE_TYPE_VALUE
		} else if variable.ValueType == repository.PIPELINE_STAGE_STEP_VARIABLE_VALUE_TYPE_GLOBAL {
			variableData.VariableType = bean.VARIABLE_TYPE_REF_GLOBAL
		} else if variable.ValueType == repository.PIPELINE_STAGE_STEP_VARIABLE_VALUE_TYPE_PREVIOUS {
			if variable.ReferenceVariableStage == repository.PIPELINE_STAGE_TYPE_POST_CI {
				variableData.VariableType = bean.VARIABLE_TYPE_REF_POST_CI
			} else if variable.ReferenceVariableStage == repository.PIPELINE_STAGE_TYPE_PRE_CI {
				variableData.VariableType = bean.VARIABLE_TYPE_REF_PRE_CI
			}
		}
		if variable.VariableType == repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_INPUT {
			//below checks for setting Value field is only relevant for ref_plugin
			//for inline step it will always end up using user's choice(if value == "" then defaultValue will also be = "", as no defaultValue option in inline )
			if variable.Value == "" {
				//no value from user; will use default value
				variableData.Value = variable.DefaultValue
			} else {
				variableData.Value = variable.Value
			}
			inputVariables = append(inputVariables, variableData)
		} else if variable.VariableType == repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_OUTPUT {
			outputVariables = append(outputVariables, variableData)
		}
	}
	conditions, err := impl.pipelineStageRepository.GetConditionsByStepId(stepId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting conditions by stepId", "err", err, "stepId", stepId)
		return nil, nil, nil, nil, err
	}
	var triggerSkipConditions []*bean.ConditionObject
	var successFailureConditions []*bean.ConditionObject
	for _, condition := range conditions {
		conditionData := &bean.ConditionObject{
			ConditionalOperator: condition.ConditionalOperator,
			ConditionalValue:    condition.ConditionalValue,
			ConditionType:       string(condition.ConditionType),
		}
		varName, ok := variableNameIdMap[condition.ConditionVariableId]
		if ok {
			conditionData.ConditionOnVariable = varName
		}
		if condition.ConditionType == repository.PIPELINE_STAGE_STEP_CONDITION_TYPE_TRIGGER || condition.ConditionType == repository.PIPELINE_STAGE_STEP_CONDITION_TYPE_SKIP {
			triggerSkipConditions = append(triggerSkipConditions, conditionData)
		} else if condition.ConditionType == repository.PIPELINE_STAGE_STEP_CONDITION_TYPE_SUCCESS || condition.ConditionType == repository.PIPELINE_STAGE_STEP_CONDITION_TYPE_FAIL {
			successFailureConditions = append(successFailureConditions, conditionData)
		}
	}
	return inputVariables, outputVariables, triggerSkipConditions, successFailureConditions, nil
}

func (impl *PipelineStageServiceImpl) BuildPluginStepDataForWfRequest(step *repository2.PluginStep) (*bean.StepObject, error) {
	stepData := &bean.StepObject{
		Name:          step.Name,
		Index:         step.Index,
		StepType:      string(step.StepType),
		ArtifactPaths: step.OutputDirectoryPath,
	}
	if step.StepType == repository2.PLUGIN_STEP_TYPE_INLINE {
		//get script and mapping data
		//getting script details for step
		scriptDetail, err := impl.globalPluginRepository.GetScriptDetailById(step.ScriptId)
		if err != nil {
			impl.logger.Errorw("error in getting script details by id", "err", err, "scriptId", step.ScriptId)
			return nil, err
		}
		//getting script mapping details
		scriptMappings, err := impl.globalPluginRepository.GetScriptMappingDetailByScriptId(step.ScriptId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting script mapping by scriptId", "err", err, "scriptId", step.ScriptId)
			return nil, err
		}
		var extraMappings []*bean.MountPath
		portMap := make(map[int]int)
		for _, scriptMapping := range scriptMappings {
			if scriptMapping.TypeOfMapping == repository2.SCRIPT_MAPPING_TYPE_FILE_PATH {
				extraMapping := &bean.MountPath{
					SourcePath:      scriptMapping.FilePathOnDisk,
					DestinationPath: scriptMapping.FilePathOnContainer,
				}
				extraMappings = append(extraMappings, extraMapping)
			} else if scriptMapping.TypeOfMapping == repository2.SCRIPT_MAPPING_TYPE_DOCKER_ARG {
				stepData.Command = scriptMapping.Command
				stepData.Args = scriptMapping.Args
			} else if scriptMapping.TypeOfMapping == repository2.SCRIPT_MAPPING_TYPE_PORT {
				portMap[scriptMapping.PortOnLocal] = scriptMapping.PortOnContainer
			}
		}
		stepData.ExecutorType = string(scriptDetail.Type)
		stepData.DockerImage = scriptDetail.ContainerImagePath
		stepData.Script = scriptDetail.Script
		if len(portMap) > 0 {
			stepData.ExposedPorts = portMap
		}
		if len(scriptDetail.StoreScriptAt) > 0 {
			stepData.CustomScriptMount = &bean.MountPath{
				DestinationPath: scriptDetail.StoreScriptAt,
			}
		}
		if scriptDetail.MountCodeToContainer && len(scriptDetail.MountCodeToContainerPath) > 0 {
			stepData.SourceCodeMount = &bean.MountPath{
				DestinationPath: scriptDetail.MountCodeToContainerPath,
			}
		}
		if scriptDetail.MountDirectoryFromHost && len(extraMappings) > 0 {
			stepData.ExtraVolumeMounts = extraMappings
		}
	} else if step.StepType == repository2.PLUGIN_STEP_TYPE_REF_PLUGIN {
		stepData.ExecutorType = "PLUGIN" //added only to avoid un-marshaling issues at ci-runner side, will not be used
		stepData.RefPluginId = step.RefPluginId
	}
	inputVars, outputVars, triggerSkipConditions, successFailureConditions, err := impl.BuildPluginVariableAndConditionDataForWfRequest(step.Id)
	if err != nil {
		impl.logger.Errorw("error in getting variable and conditions data for wf request", "err", err, "stepId", step.Id)
		return nil, err
	}
	stepData.InputVars = inputVars
	stepData.OutputVars = outputVars
	stepData.TriggerSkipConditions = triggerSkipConditions
	stepData.SuccessFailureConditions = successFailureConditions
	return stepData, nil
}

func (impl *PipelineStageServiceImpl) BuildPluginVariableAndConditionDataForWfRequest(stepId int) ([]*bean.VariableObject, []*bean.VariableObject, []*bean.ConditionObject, []*bean.ConditionObject, error) {
	var inputVariables []*bean.VariableObject
	var outputVariables []*bean.VariableObject
	//getting all variables in the step
	variables, err := impl.globalPluginRepository.GetVariablesByStepId(stepId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting variables by stepId", "err", err, "stepId", stepId)
		return nil, nil, nil, nil, err
	}
	variableNameIdMap := make(map[int]string)
	for _, variable := range variables {
		variableNameIdMap[variable.Id] = variable.Name
		variableData := &bean.VariableObject{
			Name:                       variable.Name,
			Format:                     string(variable.Format),
			ReferenceVariableStepIndex: variable.PreviousStepIndex,
			ReferenceVariableName:      variable.ReferenceVariableName,
			VariableStepIndexInPlugin:  variable.VariableStepIndexInPlugin,
		}
		if variable.ValueType == repository2.PLUGIN_VARIABLE_VALUE_TYPE_NEW {
			variableData.VariableType = bean.VARIABLE_TYPE_VALUE
		} else if variable.ValueType == repository2.PLUGIN_VARIABLE_VALUE_TYPE_GLOBAL {
			variableData.VariableType = bean.VARIABLE_TYPE_REF_GLOBAL
		} else if variable.ValueType == repository2.PLUGIN_VARIABLE_VALUE_TYPE_PREVIOUS {
			variableData.VariableType = bean.VARIABLE_TYPE_REF_PLUGIN
		}
		if variable.VariableType == repository2.PLUGIN_VARIABLE_TYPE_INPUT {
			if variable.DefaultValue == "" {
				//no default value; will use value received from user, as it must be exposed
				variableData.Value = variable.Value
			} else {
				if variable.IsExposed {
					//this value will be empty as value is set in plugin_stage_step_variable
					//& that variable is sent in pre/post steps data
					variableData.Value = variable.Value
				} else {
					variableData.Value = variable.DefaultValue
				}
			}
			inputVariables = append(inputVariables, variableData)
		} else if variable.VariableType == repository2.PLUGIN_VARIABLE_TYPE_OUTPUT {
			outputVariables = append(outputVariables, variableData)
		}
	}
	conditions, err := impl.globalPluginRepository.GetConditionsByStepId(stepId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting conditions by stepId", "err", err, "stepId", stepId)
		return nil, nil, nil, nil, err
	}
	var triggerSkipConditions []*bean.ConditionObject
	var successFailureConditions []*bean.ConditionObject
	for _, condition := range conditions {
		conditionData := &bean.ConditionObject{
			ConditionalOperator: condition.ConditionalOperator,
			ConditionalValue:    condition.ConditionalValue,
			ConditionType:       string(condition.ConditionType),
		}
		varName, ok := variableNameIdMap[condition.ConditionVariableId]
		if ok {
			conditionData.ConditionOnVariable = varName
		}
		if condition.ConditionType == repository2.PLUGIN_CONDITION_TYPE_TRIGGER || condition.ConditionType == repository2.PLUGIN_CONDITION_TYPE_SKIP {
			triggerSkipConditions = append(triggerSkipConditions, conditionData)
		} else if condition.ConditionType == repository2.PLUGIN_CONDITION_TYPE_SUCCESS || condition.ConditionType == repository2.PLUGIN_CONDITION_TYPE_FAIL {
			successFailureConditions = append(successFailureConditions, conditionData)
		}
	}
	return inputVariables, outputVariables, triggerSkipConditions, successFailureConditions, nil
}

//BuildPrePostAndRefPluginStepsDataForWfRequest and related methods ends

func (impl *PipelineStageServiceImpl) fetchScopedVariablesAndResolveTemplate(unresolvedResponse *bean.PrePostAndRefPluginStepsResponse, pipelineStageIds []int, scope resourceQualifiers.Scope) (*bean.PrePostAndRefPluginStepsResponse, error) {

	entities := make([]repository3.Entity, 0)
	for _, stageId := range pipelineStageIds {
		entities = append(entities, repository3.Entity{
			EntityType: repository3.EntityTypePipelineStage,
			EntityId:   stageId,
		})
	}

	responseJson, err := json.Marshal(unresolvedResponse)
	if err != nil {
		impl.logger.Errorw("Error in marshaling stage", "error", err, "unresolvedResponse", unresolvedResponse)
		return nil, err
	}

	resolvedTemplate, variableSnapshot, err := impl.scopedVariableManager.GetMappedVariablesAndResolveTemplateBatch(string(responseJson), entities, scope)
	if err != nil {
		return nil, err
	}
	//
	//mappingsForEntities, err := impl.variableEntityMappingService.GetAllMappingsForEntities(entities)
	//if err != nil {
	//	impl.logger.Errorw("Error in fetching mapped variables in stage request", "error", err)
	//	return nil, err
	//}
	//
	////early exit if no variables found
	//if len(mappingsForEntities) == 0 {
	//	return unresolvedResponse, nil
	//}

	// collecting all unique variable names in a stage
	//varNamesSet := mapset.NewSet()
	//for _, variableNames := range mappingsForEntities {
	//	for _, variableName := range variableNames {
	//		varNamesSet.Add(variableName)
	//	}
	//}
	//varNames := utils.ToStringArray(varNamesSet.ToSlice())

	//scopedVariables, err := impl.scopedVariableService.GetScopedVariables(scope, varNames, true)
	if err != nil {
		return nil, err
	}

	//variableSnapshot := make(map[string]string)
	//for _, variable := range scopedVariables {
	//	variableSnapshot[variable.VariableName] = variable.VariableValue.StringValue()
	//}

	//
	//parserResponse := impl.variableTemplateParser.ParseTemplate(parsers.VariableParserRequest{
	//	TemplateType:           parsers.StringVariableTemplate,
	//	Template:               string(responseJson),
	//	Variables:              scopedVariables,
	//	IgnoreUnknownVariables: true,
	//})
	//err = parserResponse.Error
	//if err != nil {
	//	impl.logger.Errorw("Error in parsing stage", "error", err, "template", responseJson, "vars", scopedVariables)
	//	return nil, err
	//}
	resolvedResponse := &bean.PrePostAndRefPluginStepsResponse{}
	err = json.Unmarshal([]byte(resolvedTemplate), resolvedResponse)
	if err != nil {
		impl.logger.Errorw("Error in unmarshalling stage", "error", err)
		return nil, err
	}
	resolvedResponse.VariableSnapshot = variableSnapshot
	return resolvedResponse, nil
}

func (impl *PipelineStageServiceImpl) extractAndMapScopedVariables(stageReq *bean.PipelineStageDto, userId int32, tx *pg.Tx) error {
	requestJson, err := json.Marshal(stageReq)
	if err != nil {
		impl.logger.Errorw("Error in marshalling stage request", "error", err)
		return err
	}
	//scopedVariables, err := impl.variableTemplateParser.ExtractVariables(string(requestJson), parsers.JsonVariableTemplate)
	//if err != nil {
	//	impl.logger.Errorw("Error in parsing variable in stage request", "error", err)
	//	return err
	//}
	//
	//if len(scopedVariables) == 0 {
	//	return nil
	//}
	//entity := repository3.Entity{
	//	EntityType: repository3.EntityTypePipelineStage,
	//	EntityId:   stageReq.Id,
	//}
	return impl.scopedVariableManager.ExtractAndMapVariables(string(requestJson), stageReq.Id, repository3.EntityTypePipelineStage, userId, tx)

	//err = impl.variableEntityMappingService.UpdateVariablesForEntity(scopedVariables, entity, userId, tx)
	//if err != nil {
	//	impl.logger.Errorw("Error in updating variables for stage request", "error", err)
	//	return err
	//}
	//return nil
}
