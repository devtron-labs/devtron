package pipeline

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type PipelineStageService interface {
	GetCiPipelineStageData(ciPipelineId int) (preCiStage *bean.PipelineStageDto, postCiStage *bean.PipelineStageDto, err error)
	CreateCiStage(stageReq *bean.PipelineStageDto, stageType repository.PipelineStageType, ciPipelineId int, userId int32) error
	UpdateCiStage(stageReq *bean.PipelineStageDto, stageType repository.PipelineStageType, ciPipelineId int, userId int32) error
	DeleteCiStage(stageReq *bean.PipelineStageDto, userId int32, tx *pg.Tx) error
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

func (impl *PipelineStageServiceImpl) CreateCiStage(stageReq *bean.PipelineStageDto, stageType repository.PipelineStageType, ciPipelineId int, userId int32) error {
	stage := &repository.PipelineStage{
		Name:         stageReq.Name,
		Description:  stageReq.Description,
		Type:         stageType,
		Deleted:      false,
		CiPipelineId: ciPipelineId,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
	stage, err := impl.pipelineStageRepository.CreateCiStage(stage)
	if err != nil {
		impl.logger.Errorw("error in creating entry for ciStage", "err", err, "ciStage", stage)
		return err
	}
	//creating stage steps and all related data
	err = impl.CreateStageSteps(stageReq.Steps, stage.Id, userId)
	if err != nil {
		impl.logger.Errorw("error in creating stage steps for ci stage", "err", err, "stageId", stage.Id)
		return err
	}
	return nil
}

func (impl *PipelineStageServiceImpl) UpdateCiStage(stageReq *bean.PipelineStageDto, stageType repository.PipelineStageType, ciPipelineId int, userId int32) error {
	//getting stage by stageType and ciPipelineId
	stageOld, err := impl.pipelineStageRepository.GetCiStageByCiPipelineIdAndStageType(ciPipelineId, stageType)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting stageId by ciPipelineId and stageType", "err", err, "ciPipelineId", ciPipelineId, "stageType", stageType)
		return err
	} else if err == pg.ErrNoRows {
		//no stage found, creating new stage
		stageReq.Id = 0
		err = impl.CreateCiStage(stageReq, stageType, ciPipelineId, userId)
		if err != nil {
			impl.logger.Errorw("error in creating new ci stage", "err", err, "ciStageReq", stageReq)
			return err
		}
	} else {
		//stageId found, to handle as an update request
		stageReq.Id = stageOld.Id
		stageUpdateReq := stageOld
		stageUpdateReq.Name = stageReq.Name
		stageUpdateReq.Description = stageReq.Description
		stageUpdateReq.UpdatedBy = userId
		stageUpdateReq.UpdatedOn = time.Now()
		_, err = impl.pipelineStageRepository.UpdateCiStage(stageUpdateReq)
		if err != nil {
			impl.logger.Errorw("error in updating entry for ciStage", "err", err, "ciStage", stageUpdateReq)
			return err
		}
		// filtering(if steps/variables/conditions are updated or newly added) and performing relevant actions on update request
		err = impl.FilterAndActOnStepsInCiStageUpdateRequest(stageReq, userId)
		if err != nil {
			impl.logger.Errorw("error in filtering and performing actions on steps in ci stage update request", "err", err, "stageReq", stageReq)
			return err
		}
	}
	return nil
}

func (impl *PipelineStageServiceImpl) DeleteCiStage(stageReq *bean.PipelineStageDto, userId int32, tx *pg.Tx) error {
	//marking stage deleted
	err := impl.pipelineStageRepository.MarkCiStageDeletedById(stageReq.Id, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in marking ci stage deleted", "err", err, "ciStageId", stageReq.Id)
		return err
	}
	//marking all steps deleted
	err = impl.pipelineStageRepository.MarkCiStageStepsDeletedByStageId(stageReq.Id, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in marking ci stage steps deleted by stageId", "err", err, "ciStageId", stageReq.Id)
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
			impl.logger.Errorw("error in marking ci stage step variables deleted by variableIds", "err", err, "variableIds", variableIds)
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
			impl.logger.Errorw("error in marking ci stage step conditions deleted by conditionIds", "err", err, "conditionIds", conditionIds)
			return err
		}
	}
	return nil
}

func (impl *PipelineStageServiceImpl) FilterAndActOnStepsInCiStageUpdateRequest(stageReq *bean.PipelineStageDto, userId int32) error {
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
	for _, step := range stageReq.Steps {
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
		//deleting all current steps since no step is present in update request
		err = impl.pipelineStageRepository.MarkStepsDeletedByStageId(stageReq.Id)
		if err != nil {
			impl.logger.Errorw("error in marking all steps deleted by stageId", "err", err, "stageId", stageReq.Id)
			return err
		}
	}
	if len(stepsToBeCreated) > 0 {
		//creating new steps
		err = impl.CreateStageSteps(stepsToBeCreated, stageReq.Id, userId)
		if err != nil {
			impl.logger.Errorw("error in creating stage steps for ci stage", "err", err, "stageId", stageReq.Id)
			return err
		}
	}
	if len(stepsToBeUpdated) > 0 {
		//updating steps
		err = impl.UpdateStageSteps(stepsToBeUpdated, userId, stageReq.Id)
		if err != nil {
			impl.logger.Errorw("error in updating stage steps for ci stage", "err", err)
			return err
		}
	}
	return nil
}

func (impl *PipelineStageServiceImpl) UpdateStageSteps(steps []*bean.PipelineStageStepDto, userId int32, stageId int) error {
	for _, step := range steps {
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
			ReportDirectoryPath: step.ReportDirectoryPath,
			Deleted:             false,
			AuditLog: sql.AuditLog{
				CreatedOn: savedStep.CreatedOn,
				CreatedBy: savedStep.CreatedBy,
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
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
				scriptEntryId, err := impl.CreateScriptAndMappingForInlineStep(step.InlineStepDetail, userId)
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
		//updating input variable
		inputVarNameIdMap, err := impl.UpdatePipelineStageStepVariables(step.Id, inputVariables, repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_INPUT, userId)
		if err != nil {
			impl.logger.Errorw("error in updating input variables", "err", err)
			return err
		}
		//updating output variable
		outputVarNameIdMap, err := impl.UpdatePipelineStageStepVariables(step.Id, outputVariables, repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_OUTPUT, userId)
		if err != nil {
			impl.logger.Errorw("error in updating output variables", "err", err)
			return err
		}
		if len(conditionDetails) > 0 {
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
		if len(conditionDetails) > 0 {
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
	}

	return nil
}

func (impl *PipelineStageServiceImpl) CreateScriptAndMappingForInlineStep(inlineStepDetail *bean.InlineStepDetailDto, userId int32) (scriptId int, err error) {
	scriptEntry := &repository.PluginPipelineScript{
		Script:                   inlineStepDetail.Script,
		Type:                     inlineStepDetail.ScriptType,
		StoreScriptAt:            inlineStepDetail.StoreScriptAt,
		DockerfileExists:         inlineStepDetail.DockerfileExists,
		MountPath:                inlineStepDetail.MountPath,
		MountCodeToContainer:     inlineStepDetail.MountCodeToContainer,
		MountCodeToContainerPath: inlineStepDetail.MountCodeToContainerPath,
		MountDirectoryFromHost:   inlineStepDetail.MountDirectoryFromHost,
		ConfigureMountPath:       inlineStepDetail.ConfigureMountPath,
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
	scriptEntry, err = impl.pipelineStageRepository.CreatePipelineScript(scriptEntry)
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
		argsByte, err := json.Marshal(commandArgMap.Args)
		if err != nil {
			impl.logger.Errorw("error in marshaling docker args", "err", err)
			return 0, err
		}
		repositoryEntry := repository.ScriptPathArgPortMapping{
			TypeOfMapping: repository2.SCRIPT_MAPPING_TYPE_DOCKER_ARG,
			Command:       commandArgMap.Command,
			Args:          string(argsByte),
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
		err = impl.pipelineStageRepository.CreateScriptMapping(scriptMap)
		if err != nil {
			impl.logger.Errorw("error in creating script mappings", "err", err, "scriptMappings", scriptMap)
			return 0, err
		}
	}
	return scriptEntry.Id, nil
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
		ConfigureMountPath:       inlineStepDetail.ConfigureMountPath,
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
		argsByte, err := json.Marshal(commandArgMap.Args)
		if err != nil {
			impl.logger.Errorw("error in marshaling docker args", "err", err)
			return err
		}
		repositoryEntry := repository.ScriptPathArgPortMapping{
			TypeOfMapping: repository2.SCRIPT_MAPPING_TYPE_DOCKER_ARG,
			Command:       commandArgMap.Command,
			Args:          string(argsByte),
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
		err = impl.pipelineStageRepository.CreateScriptMapping(scriptMap)
		if err != nil {
			impl.logger.Errorw("error in creating script mappings", "err", err, "scriptMappings", scriptMap)
			return err
		}
	}
	return nil
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
	if len(variablesRepo) > 0 {
		//saving variables
		variablesRepo, err = impl.pipelineStageRepository.CreatePipelineStageStepVariables(variablesRepo)
		if err != nil {
			impl.logger.Errorw("error in creating variables for pipeline stage steps", "err", err, "variables", variablesRepo)
			return nil, err
		}
	}
	return variablesRepo, nil
}

func (impl *PipelineStageServiceImpl) UpdatePipelineStageStepVariables(stepId int, variables []*bean.StepVariableDto, variableType repository.PipelineStageStepVariableType, userId int32) (map[string]int, error) {
	//getting ids of all current active variables
	variableIds, err := impl.pipelineStageRepository.GetVariableIdsByStepIdAndVariableType(stepId, variableType)
	if err != nil {
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
		err = impl.pipelineStageRepository.MarkVariablesDeletedExcludingActiveVariablesInUpdateReq(activeVariableIdsPresentInReq, stepId, variableType)
		if err != nil {
			impl.logger.Errorw("error in marking all variables deleted excluding active variables in update req", "err", err, "activeVariableIdsPresentInReq", activeVariableIdsPresentInReq)
			return nil, err
		}
	} else {
		//deleting all variables by stepId, since no variable is present in update request
		err = impl.pipelineStageRepository.MarkVariablesDeletedByStepIdAndVariableType(stepId, variableType)
		if err != nil {
			impl.logger.Errorw("error in marking all variables deleted by stepId", "err", err, "stepId", stepId)
			return nil, err
		}
	}
	var newVariables []repository.PipelineStageStepVariable
	if len(variablesToBeCreated) > 0 {
		//creating new variables
		newVariables, err = impl.CreateVariablesEntryInDb(stepId, variablesToBeCreated, variableType, userId)
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
			Id:                    v.Id,
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
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}
		variablesRepo = append(variablesRepo, inVarRepo)
	}
	if len(variablesRepo) > 0 {
		variablesRepo, err = impl.pipelineStageRepository.UpdatePipelineStageStepVariables(variablesRepo)
		if err != nil {
			impl.logger.Errorw("error in updating variables for pipeline stage steps", "err", err, "variables", variablesRepo)
			return nil, err
		}
	}
	return variableNameIdMap, nil
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
	if len(conditionsRepo) > 0 {
		//saving conditions
		conditionsRepo, err = impl.pipelineStageRepository.CreatePipelineStageStepConditions(conditionsRepo)
		if err != nil {
			impl.logger.Errorw("error in creating pipeline stage step conditions", "err", err, "conditionsRepo", conditionsRepo)
			return nil, err
		}
	}
	return conditionsRepo, nil
}

func (impl *PipelineStageServiceImpl) UpdatePipelineStageStepConditions(stepId int, conditions []*bean.ConditionDetailDto, variableNameIdMap map[string]int, userId int32) ([]repository.PipelineStageStepCondition, error) {
	//getting ids of all current active variables
	conditionIds, err := impl.pipelineStageRepository.GetConditionIdsByStepId(stepId)
	if err != nil {
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
		err = impl.pipelineStageRepository.MarkConditionsDeletedExcludingActiveVariablesInUpdateReq(activeConditionIdsPresentInReq, stepId)
		if err != nil {
			impl.logger.Errorw("error in marking all conditions deleted excluding active conditions in update req", "err", err, "activeConditionIdsPresentInReq", activeConditionIdsPresentInReq)
			return nil, err
		}
	} else {
		// deleting all current conditions, since no condition is present in update request
		err = impl.pipelineStageRepository.MarkConditionsDeletedByStepId(stepId)
		if err != nil {
			impl.logger.Errorw("error in marking all conditions deleted by stepId", "err", err, "stepId", stepId)
			return nil, err
		}
	}
	//creating new conditions
	_, err = impl.CreateConditionsEntryInDb(stepId, conditionsToBeCreated, variableNameIdMap, userId)
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
	//saving conditions
	conditionsRepo, err = impl.pipelineStageRepository.UpdatePipelineStageStepConditions(conditionsRepo)
	if err != nil {
		impl.logger.Errorw("error in updating pipeline stage step conditions", "err", err, "conditionsRepo", conditionsRepo)
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
		DockerfileExists:         scriptDetail.DockerfileExists,
		MountPath:                scriptDetail.MountPath,
		MountCodeToContainer:     scriptDetail.MountCodeToContainer,
		MountCodeToContainerPath: scriptDetail.MountCodeToContainerPath,
		MountDirectoryFromHost:   scriptDetail.MountDirectoryFromHost,
		ConfigureMountPath:       scriptDetail.ConfigureMountPath,
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
			var args []string
			err = json.Unmarshal([]byte(scriptMapping.Args), &args)
			if err != nil {
				impl.logger.Errorw("error in un-marshaling docker args", "err", err)
				return nil, err
			}
			mapEntry := &bean.CommandArgsMap{
				Command: scriptMapping.Command,
				Args:    args,
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
	if err != nil {
		impl.logger.Errorw("error in getting variables by stepId", "err", err, "stepId", stepId)
		return nil, nil, nil, err
	}
	variableNameIdMap := make(map[int]string)
	for _, variable := range variables {
		variableNameIdMap[variable.Id] = variable.Name
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
