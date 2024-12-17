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

package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	commonBean "github.com/devtron-labs/common-lib/workflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/adapter"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/plugin"
	repository2 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables"
	repository3 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/devtron-labs/devtron/util/sliceUtil"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type PipelineStageService interface {
	GetCiPipelineStageData(ciPipelineId int) (preCiStage *bean.PipelineStageDto, postCiStage *bean.PipelineStageDto, err error)
	CreatePipelineStage(stageReq *bean.PipelineStageDto, stageType repository.PipelineStageType, pipelineId int, userId int32) error
	UpdatePipelineStage(stageReq *bean.PipelineStageDto, stageType repository.PipelineStageType, pipelineId int, userId int32) error
	DeletePipelineStage(stageReq *bean.PipelineStageDto, userId int32, tx *pg.Tx) error
	BuildPrePostAndRefPluginStepsDataForWfRequest(request *bean.BuildPrePostStepDataRequest) (*bean.PrePostAndRefPluginStepsResponse, error)
	GetCiPipelineStageDataDeepCopy(ciPipelineId int) (preCiStage *bean.PipelineStageDto, postCiStage *bean.PipelineStageDto, err error)
	GetCdPipelineStageDataDeepCopy(cdPipeline *pipelineConfig.Pipeline) (*bean.PipelineStageDto, *bean.PipelineStageDto, error)
	GetCdPipelineStageDataDeepCopyForPipelineIds(cdPipelineIds []int, pipelineMap map[int]*pipelineConfig.Pipeline) (map[int][]*bean.PipelineStageDto, error)
	GetCdStageByCdPipelineIdAndStageType(cdPipelineId int, stageType repository.PipelineStageType) (*repository.PipelineStage, error)
	// DeletePipelineStageIfReq function is used to delete corrupted pipelineStage data
	// , there was a bug(https://github.com/devtron-labs/devtron/issues/3826) where we were not deleting pipeline stage entry even after deleting all the pipelineStageSteps
	// , this will delete those pipelineStage entry
	DeletePipelineStageIfReq(stageReq *bean.PipelineStageDto, userId int32) (error, bool)
	IsScanPluginConfiguredAtPipelineStage(pipelineId int, pipelineStage repository.PipelineStageType, pluginName string) (bool, error)
}

func NewPipelineStageService(logger *zap.SugaredLogger,
	pipelineStageRepository repository.PipelineStageRepository,
	globalPluginRepository repository2.GlobalPluginRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	scopedVariableManager variables.ScopedVariableManager,
	globalPluginService plugin.GlobalPluginService,

) *PipelineStageServiceImpl {
	return &PipelineStageServiceImpl{
		logger:                  logger,
		pipelineStageRepository: pipelineStageRepository,
		globalPluginRepository:  globalPluginRepository,
		pipelineRepository:      pipelineRepository,
		scopedVariableManager:   scopedVariableManager,
		globalPluginService:     globalPluginService,
	}
}

type PipelineStageServiceImpl struct {
	logger                  *zap.SugaredLogger
	pipelineStageRepository repository.PipelineStageRepository
	globalPluginRepository  repository2.GlobalPluginRepository
	pipelineRepository      pipelineConfig.PipelineRepository
	scopedVariableManager   variables.ScopedVariableManager
	globalPluginService     plugin.GlobalPluginService
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
			preCiStage, err = impl.BuildPipelineStageDataDeepCopy(ciStage, nil)
			if err != nil {
				impl.logger.Errorw("error in getting ci stage data", "err", err, "ciStage", ciStage)
				return nil, nil, err
			}
		} else if ciStage.Type == repository.PIPELINE_STAGE_TYPE_POST_CI {
			postCiStage, err = impl.BuildPipelineStageDataDeepCopy(ciStage, nil)
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

func (impl *PipelineStageServiceImpl) GetCdPipelineStageDataDeepCopy(cdPipeline *pipelineConfig.Pipeline) (*bean.PipelineStageDto, *bean.PipelineStageDto, error) {
	//migrate plugin_metadata to plugin_parent_metadata, also pluginVersionsMetadata will have updated entries after migrating, this is a one time operation
	err := impl.globalPluginService.MigratePluginData()
	if err != nil {
		impl.logger.Errorw("GetCdPipelineStageDataDeepCopy, error in migrating plugin data into parent metadata table", "err", err)
		return nil, nil, err
	}

	//getting all stages by cd pipeline id
	cdStages, err := impl.pipelineStageRepository.GetAllCdStagesByCdPipelineId(cdPipeline.Id)
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
			preDeployStage, err = impl.BuildPipelineStageDataDeepCopy(cdStage, cdPipeline)
			if err != nil {
				impl.logger.Errorw("error in getting cd stage data", "err", err, "cdStage", cdStage)
				return nil, nil, err
			}
		} else if cdStage.Type == repository.PIPELINE_STAGE_TYPE_POST_CD {
			postDeployStage, err = impl.BuildPipelineStageDataDeepCopy(cdStage, cdPipeline)
			if err != nil {
				impl.logger.Errorw("error in getting cd stage data", "err", err, "cdStage", cdStage)
				return nil, nil, err
			}
		} else {
			impl.logger.Errorw("found improper stage mapped with cdPipeline", "cdPipelineId", cdPipeline.Id, "stage", cdStage)
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

func (impl *PipelineStageServiceImpl) GetCdPipelineStageDataDeepCopyForPipelineIds(cdPipelineIds []int, pipelineMap map[int]*pipelineConfig.Pipeline) (map[int][]*bean.PipelineStageDto, error) {
	pipelinePrePostStageMappingResp := make(map[int][]*bean.PipelineStageDto, len(cdPipelineIds))
	pipelineStages, err := impl.pipelineStageRepository.GetAllCdStagesByCdPipelineIds(cdPipelineIds)
	if err != nil {
		impl.logger.Errorw("error in getting pipelineStages from cdPipelineIds", "err", err, "cdPipelineIds", cdPipelineIds)
		return nil, err
	}
	if len(pipelineStages) == 0 {
		return pipelinePrePostStageMappingResp, nil
	}
	for _, pipelineStage := range pipelineStages {
		pipelineId := pipelineStage.CdPipelineId
		var preDeployStage, postDeployStage *bean.PipelineStageDto
		pipeline := pipelineMap[pipelineId]
		if pipeline == nil {
			impl.logger.Errorw("error in finding pipeline by id", "pipelineId", pipelineId)
			return nil, &util.ApiError{Code: "404", HttpStatusCode: http.StatusNotFound, InternalMessage: fmt.Sprintf("pipeline not found, id : %d", pipelineId), UserMessage: "pipeline not found"}
		}
		if _, ok := pipelinePrePostStageMappingResp[pipelineId]; !ok {
			pipelinePrePostStageMappingResp[pipelineId] = make([]*bean.PipelineStageDto, 2) //for old logic compatibility (assumptions that pre is 0 element and post is 1 element)
		}
		if pipelineStage.Type == repository.PIPELINE_STAGE_TYPE_PRE_CD {
			preDeployStage, err = impl.BuildPipelineStageDataDeepCopy(pipelineStage, pipeline)
			if err != nil {
				impl.logger.Errorw("error in getting cd stage data", "err", err, "cdStage", bean.CdStage)
				return nil, err
			}
			preDeployStage.Name = "Pre-Deployment"
			pipelinePrePostStageMappingResp[pipelineId][0] = preDeployStage
		} else if pipelineStage.Type == repository.PIPELINE_STAGE_TYPE_POST_CD {
			postDeployStage, err = impl.BuildPipelineStageDataDeepCopy(pipelineStage, pipeline)
			if err != nil {
				impl.logger.Errorw("error in getting cd stage data", "err", err, "cdStage", bean.CdStage)
				return nil, err
			}
			postDeployStage.Name = "Post-Deployment"
			pipelinePrePostStageMappingResp[pipelineId][1] = postDeployStage
		} else {
			impl.logger.Errorw("found improper stage mapped with cdPipeline", "cdPipelineId", pipelineId, "stage", bean.CdStage)
		}
	}
	return pipelinePrePostStageMappingResp, nil
}

func (impl *PipelineStageServiceImpl) BuildPipelineStageDataDeepCopy(pipelineStage *repository.PipelineStage, pipeline *pipelineConfig.Pipeline) (*bean.PipelineStageDto, error) {
	stageData := &bean.PipelineStageDto{
		Id:          pipelineStage.Id,
		Name:        pipelineStage.Name,
		Description: pipelineStage.Description,
		Type:        pipelineStage.Type,
	}
	if pipelineStage.Type == repository.PIPELINE_STAGE_TYPE_PRE_CD || pipelineStage.Type == repository.PIPELINE_STAGE_TYPE_POST_CD {
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
	// getting all variables in the step
	variables, err := impl.pipelineStageRepository.GetVariablesByStepId(stepId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in getting variables by stepId", "err", err, "stepId", stepId)
		return nil, nil, nil, err
	}
	variableNameIdMap := make(map[int]string)
	for _, variable := range variables {
		variableNameIdMap[variable.Id] = variable.Name
		variableDto, convErr := adapter.GetStepVariableDto(variable)
		if convErr != nil {
			impl.logger.Errorw("error in converting variable to dto", "err", convErr, "variable", variable)
			return nil, nil, nil, convErr
		}
		if variable.VariableType.IsInput() {
			inputVariablesDto = append(inputVariablesDto, variableDto)
		} else if variable.VariableType.IsOutput() {
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
	//migrate plugin_metadata to plugin_parent_metadata, also pluginVersionsMetadata will have updated entries after migrating, this is a one time operation
	err := impl.globalPluginService.MigratePluginData()
	if err != nil {
		impl.logger.Errorw("GetCdPipelineStageDataDeepCopy, error in migrating plugin data into parent metadata table", "err", err)
		return nil, nil, err
	}

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
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in getting variables by stepId", "err", err, "stepId", stepId)
		return nil, nil, nil, err
	}
	variableNameIdMap := make(map[int]string)
	for _, variable := range variables {
		variableNameIdMap[variable.Id] = variable.Name
		variableDto, convErr := adapter.GetStepVariableDto(variable)
		if convErr != nil {
			impl.logger.Errorw("error in converting variable to dto", "err", convErr, "variable", variable)
			return nil, nil, nil, convErr
		}
		if variable.VariableType.IsInput() {
			inputVariablesDto = append(inputVariablesDto, variableDto)
		} else if variable.VariableType.IsOutput() {
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

// validateStepInputVariableDtoForConfigure validates the []*bean.StepVariableDto
// Note: This function should be used for configure request stage (Create/ Update)
func validateStepInputVariableDtoForConfigure(variableDtos []*bean.StepVariableDto) error {
	return validateStepVariables(variableDtos, false)
}

func (impl *PipelineStageServiceImpl) CreateVariablesEntryInDb(stepId int, variables []*bean.StepVariableDto, variableType repository.PipelineStageStepVariableType, userId int32, tx *pg.Tx) ([]repository.PipelineStageStepVariable, error) {
	// validate variables
	// for output variables, validation is not required
	if variableType.IsInput() {
		// validation is only performed for input variables
		validationErr := validateStepInputVariableDtoForConfigure(variables)
		if validationErr != nil {
			impl.logger.Errorw("validation failed for StepVariableDto", "err", validationErr, "stepId", stepId, "variableType", variableType)
			return nil, validationErr
		}
	}
	var variablesRepo []repository.PipelineStageStepVariable
	var err error
	for _, v := range variables {
		inVarRepo := repository.PipelineStageStepVariable{
			PipelineStageStepId: stepId,
			Name:                v.Name,
			Format:              v.Format,
			Description:         v.Description,
			// Hard coding to TRUE;
			// PipelineStageStepVariable are always configured by user,
			// And only exposed variables are configured by user
			IsExposed:                 true,
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
		// saving variables
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
		err = impl.FilterAndActOnStepsInPipelineStageUpdateRequest(stageReq, stageType, pipelineId, userId)
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

func (impl *PipelineStageServiceImpl) FilterAndActOnStepsInPipelineStageUpdateRequest(stageReq *bean.PipelineStageDto, stageType repository.PipelineStageType, pipelineId int, userId int32) error {
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
	stageDeleted := false
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
		stageDeleted = true
	}
	if len(stepsToBeCreated) > 0 {
		if stageDeleted {
			//create new stage and set stageReq.Id to newly created stage
			stageReq.Id = 0
			stageReq.Steps = nil
			err = impl.CreatePipelineStage(stageReq, stageType, pipelineId, userId)
			if err != nil {
				impl.logger.Errorw("error in creating new pipeline stage", "err", err, "pipelineStageReq", stageReq)
				return err
			}
		}
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
	//using tx for conditions db operation

	dbConnection := impl.pipelineStageRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in starting tx", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	err = impl.UpdateStageStepsWithTx(steps, userId, stageId, indexNameString, tx)
	if err != nil {
		impl.logger.Errorw("Error in updating stage step", "steps", steps, "err", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in tx commit", "err", err)
		return err
	}
	return nil
}

func (impl *PipelineStageServiceImpl) UpdateStageStepsWithTx(steps []*bean.PipelineStageStepDto, userId int32, stageId int, indexNameString map[int]string, tx *pg.Tx) error {
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
				err = impl.pipelineStageRepository.MarkScriptDeletedById(idOfScriptToBeDeleted, tx)
				if err != nil {
					impl.logger.Errorw("error in marking script deleted by id", "err", err, "scriptId", idOfScriptToBeDeleted)
					return err
				}
				//deleting mappings
				err = impl.pipelineStageRepository.MarkScriptMappingDeletedByScriptId(idOfScriptToBeDeleted, tx)
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
				scriptEntryId, err := impl.CreateScriptAndMappingForInlineStep(step.InlineStepDetail, userId, tx)
				if err != nil {
					impl.logger.Errorw("error in creating script and mapping for inline step", "err", err, "inlineStepDetail", step.InlineStepDetail)
					return err
				}
				//updating scriptId in step update req
				stepUpdateReq.ScriptId = scriptEntryId
			} else {
				//update script and its mappings
				err = impl.UpdateScriptAndMappingForInlineStep(step.InlineStepDetail, savedStep.ScriptId, userId, tx)
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
		_, err = impl.pipelineStageRepository.UpdatePipelineStageStep(stepUpdateReq, tx)
		if err != nil {
			impl.logger.Errorw("error in updating pipeline stage step", "err", err, "stepUpdateReq", stepUpdateReq)
			return err
		}
		inputVarNameIdMap, outputVarNameIdMap, err := impl.UpdateInputAndOutputVariables(step.Id, inputVariables, outputVariables, userId, tx)
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
		_, err = impl.UpdatePipelineStageStepConditions(step.Id, conditionDetails, varNameIdMap, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in updating step conditions", "err", err)
			return err
		}

	}
	return nil
}

func (impl *PipelineStageServiceImpl) UpdateScriptAndMappingForInlineStep(inlineStepDetail *bean.InlineStepDetailDto, scriptId int, userId int32, tx *pg.Tx) (err error) {
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
	err = impl.pipelineStageRepository.MarkScriptMappingDeletedByScriptId(scriptId, tx)
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
		err = impl.pipelineStageRepository.CreateScriptMapping(scriptMap, tx)
		if err != nil {
			impl.logger.Errorw("error in creating script mappings", "err", err, "scriptMappings", scriptMap)
			return err
		}
	}
	return nil
}

func (impl *PipelineStageServiceImpl) UpdateInputAndOutputVariables(stepId int, inputVariables []*bean.StepVariableDto, outputVariables []*bean.StepVariableDto, userId int32, tx *pg.Tx) (inputVarNameIdMap map[string]int, outputVarNameIdMap map[string]int, err error) {
	// updating input variable
	inputVarNameIdMap, err = impl.UpdatePipelineStageStepVariables(stepId, inputVariables, repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_INPUT, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in updating input variables", "err", err)
		return nil, nil, err
	}
	// updating output variable
	outputVarNameIdMap, err = impl.UpdatePipelineStageStepVariables(stepId, outputVariables, repository.PIPELINE_STAGE_STEP_VARIABLE_TYPE_OUTPUT, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in updating output variables", "err", err)
		return nil, nil, err
	}
	return inputVarNameIdMap, outputVarNameIdMap, nil
}

// updateVariablesRequest function is used to update the variables []*bean.StepVariableDto request
// - If variableId is 0, it will be handled as new variable
// - If variableId is repeated in request, from second occurrence it will be handled as new variable.
func updateVariablesRequest(variables []*bean.StepVariableDto) []*bean.StepVariableDto {
	uniqueIdsOfVariable := make(map[int]bool)
	for _, variable := range variables {
		if variable.Id == 0 {
			// variableId is 0, will be handled as new variable
			continue
		}
		_, idRepeated := uniqueIdsOfVariable[variable.Id]
		if idRepeated {
			// NOTE: repeated variableId in request will be handled as new variable.
			// This was initially placed to handle the alternative UI (not shipped).
			// TODO: Analyse the impact of removing this idRepeated logic.
			// Until then keep it as is.

			// variableId is repeated in request, will be handled as new variable
			variable.Id = 0 // setting id to 0, for repeated variableId in request
		} else {
			// variableId present in current active variables and not repeated in request
			// will be used for update
			uniqueIdsOfVariable[variable.Id] = true
		}
	}
	return variables
}

// getNewVariablesPresentInReq function is used to get the new variables []*bean.StepVariableDto present in request
// - If variableId is not present in current active variables, it will be handled as new variable
func getNewVariablesPresentInReq(variables []*bean.StepVariableDto, activeVariableIdsMap map[int]*repository.PipelineStageStepVariable) (variablesToBeCreated []*bean.StepVariableDto) {
	variablesToBeCreated = make([]*bean.StepVariableDto, 0, len(variables))
	for _, variable := range variables {
		_, idFound := activeVariableIdsMap[variable.Id]
		if !idFound {
			// variableId, not present in current active variables, will be handled as new variable
			variable.Id = 0 // setting id to 0, for repeated variableId in request
			variablesToBeCreated = append(variablesToBeCreated, variable)
		}
	}
	return variablesToBeCreated
}

// getExistingVariablesPresentInReq function is used to get the existing variables []*bean.StepVariableDto present in request
// - If variableId is present in current active variables, it will be updated
func getExistingVariablesPresentInReq(variables []*bean.StepVariableDto, activeVariableIdsMap map[int]*repository.PipelineStageStepVariable) (variablesToBeUpdated []*bean.StepVariableDto) {
	variablesToBeUpdated = make([]*bean.StepVariableDto, 0, len(variables))
	for _, variable := range variables {
		_, idFound := activeVariableIdsMap[variable.Id]
		if idFound {
			// variableId present in current active variables, will be updated
			variablesToBeUpdated = append(variablesToBeUpdated, variable)
		}
	}
	return variablesToBeUpdated
}

func (impl *PipelineStageServiceImpl) UpdatePipelineStageStepVariables(stepId int, variables []*bean.StepVariableDto, variableType repository.PipelineStageStepVariableType, userId int32, tx *pg.Tx) (map[string]int, error) {
	// validate variables
	// for output variables, validation is not required
	if variableType.IsInput() {
		// validation is only performed for input variables
		validationErr := validateStepInputVariableDtoForConfigure(variables)
		if validationErr != nil {
			impl.logger.Errorw("validation failed for input StepVariableDto", "err", validationErr, "stepId", stepId, "variableType", variableType)
			return nil, validationErr
		}
	}
	// getting ids of all currently active variables
	activeVariables, err := impl.pipelineStageRepository.GetVariablesByStepIdAndVariableType(stepId, variableType)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in getting variablesIds by stepId", "err", err, "stepId", stepId)
		return nil, err
	}
	activeVariablesMap := sliceUtil.NewMapFromFuncExec(activeVariables, func(variable *repository.PipelineStageStepVariable) int {
		return variable.Id
	})
	variables = updateVariablesRequest(variables)
	variablesToBeCreated := getNewVariablesPresentInReq(variables, activeVariablesMap)
	variablesToBeUpdated := getExistingVariablesPresentInReq(variables, activeVariablesMap)

	// activeVariableIdsPresentInReq: all the variable ids that are currently active and present in update request
	activeVariableIdsPresentInReq := sliceUtil.NewSliceFromFuncExec(variablesToBeUpdated, func(variable *bean.StepVariableDto) int {
		return variable.Id
	})

	if len(activeVariableIdsPresentInReq) > 0 {
		// deleting all variables that are currently active but not present in update request
		err = impl.pipelineStageRepository.MarkVariablesDeletedExcludingActiveVariablesInUpdateReq(activeVariableIdsPresentInReq, stepId, variableType, tx)
		if err != nil {
			impl.logger.Errorw("error in marking all variables deleted excluding active variables in update req", "err", err, "activeVariableIdsPresentInReq", activeVariableIdsPresentInReq)
			return nil, err
		}
	} else {
		// deleting all variables by stepId, since no variable is present in update request
		err = impl.pipelineStageRepository.MarkVariablesDeletedByStepIdAndVariableType(stepId, variableType, userId, tx)
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
			Id:                  v.Id,
			PipelineStageStepId: stepId,
			Name:                v.Name,
			Format:              v.Format,
			Description:         v.Description,
			// Hard coding to TRUE;
			// PipelineStageStepVariable are always configured by user,
			// And only exposed variables are configured by user
			IsExposed:                 true,
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
		}
		if _, ok := activeVariablesMap[v.Id]; ok {
			inVarRepo.CreatedOn = activeVariablesMap[v.Id].CreatedOn
			inVarRepo.CreatedBy = activeVariablesMap[v.Id].CreatedBy
		}
		inVarRepo.UpdateAuditLog(userId)
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

func (impl *PipelineStageServiceImpl) UpdatePipelineStageStepConditions(stepId int, conditions []*bean.ConditionDetailDto, variableNameIdMap map[string]int, userId int32, tx *pg.Tx) ([]repository.PipelineStageStepCondition, error) {
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
		// marking all variables deleted
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

	err = impl.scopedVariableManager.RemoveMappedVariables(stageReq.Id, repository3.EntityTypePipelineStage, userId, tx)
	if err != nil {
		return err
	}
	return nil
}

//DeleteCiStage and related methods starts

// BuildPrePostAndRefPluginStepsDataForWfRequest and related methods starts
func (impl *PipelineStageServiceImpl) BuildPrePostAndRefPluginStepsDataForWfRequest(request *bean.BuildPrePostStepDataRequest) (*bean.PrePostAndRefPluginStepsResponse, error) {
	pipelineId := request.PipelineId
	stageType := request.StageType
	scope := request.Scope
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
		steps, refIds, err := impl.buildPipelineStageDataForWfRequest(pipelineStage)
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

func (impl *PipelineStageServiceImpl) buildPipelineStageDataForWfRequest(pipelineStage *repository.PipelineStage) ([]*bean.StepObject, []int, error) {
	//getting all steps for this stage
	steps, err := impl.pipelineStageRepository.GetAllStepsByStageId(pipelineStage.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting all steps by stageId", "err", err, "stageId", pipelineStage.Id)
		return nil, nil, err
	}
	var stepsData []*bean.StepObject
	var refPluginIds []int
	for _, step := range steps {
		stepData, err := impl.buildPipelineStepDataForWfRequest(step)
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

func (impl *PipelineStageServiceImpl) buildPipelineStepDataForWfRequest(step *repository.PipelineStageStep) (*bean.StepObject, error) {
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
	variableAndConditionData, err := impl.buildVariableAndConditionDataForWfRequest(step.Id)
	if err != nil {
		impl.logger.Errorw("error in getting variable and conditions data for wf request", "err", err, "stepId", step.Id)
		return nil, err
	}
	stepData.InputVars = variableAndConditionData.GetInputVariables()
	stepData.OutputVars = variableAndConditionData.GetOutputVariables()
	stepData.TriggerSkipConditions = variableAndConditionData.GetTriggerSkipConditions()
	stepData.SuccessFailureConditions = variableAndConditionData.GetSuccessFailureConditions()
	return stepData, nil
}

func (impl *PipelineStageServiceImpl) buildVariableAndConditionDataForWfRequest(stepId int) (*bean.VariableAndConditionDataForStep, error) {
	variableAndConditionData := bean.NewVariableAndConditionDataForStep()
	//getting all variables in the step
	variables, err := impl.pipelineStageRepository.GetVariablesByStepId(stepId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in getting variables by stepId", "err", err, "stepId", stepId)
		return variableAndConditionData, err
	}
	variableNameIdMap := make(map[int]string)
	for _, variable := range variables {
		variableNameIdMap[variable.Id] = variable.Name
		// getting format
		// ignoring error as it is already validated in func validatePipelineStageStepVariableForTrigger
		format, _ := commonBean.NewFormat(variable.Format.String())
		variableData := &commonBean.VariableObject{
			Name:                       variable.Name,
			Format:                     format,
			ReferenceVariableStepIndex: variable.PreviousStepIndex,
			ReferenceVariableName:      variable.ReferenceVariableName,
			VariableStepIndexInPlugin:  variable.VariableStepIndexInPlugin,
			//  default VariableType is commonBean.VariableTypeValue
			VariableType: commonBean.VariableTypeValue,
		}
		if variable.ValueType.IsUserDefinedValue() {
			variableData.VariableType = commonBean.VariableTypeValue
		} else if variable.ValueType.IsGlobalDefinedValue() {
			variableData.VariableType = commonBean.VariableTypeRefGlobal
		} else if variable.ValueType.IsPreviousOutputDefinedValue() {
			if variable.ReferenceVariableStage == repository.PIPELINE_STAGE_TYPE_POST_CI {
				variableData.VariableType = commonBean.VariableTypeRefPostCi
			} else if variable.ReferenceVariableStage == repository.PIPELINE_STAGE_TYPE_PRE_CI {
				variableData.VariableType = commonBean.VariableTypeRefPreCi
			}
		}
		if variable.VariableType.IsInput() {
			// below checks for setting Value field is only relevant for ref_plugin
			// for an inline step it will always end up using user's choice(if value == "" then defaultValue will also be = "", as no defaultValue option in inline )
			if variable.Value == "" {
				//no value from user; will use default value
				variableData.Value = variable.DefaultValue
			} else {
				// if the format is not file, then the value will be the value provided by user
				variableData.Value = variable.Value
			}
			variableAndConditionData = variableAndConditionData.AddInputVariable(variableData)
		} else if variable.VariableType.IsOutput() {
			variableAndConditionData = variableAndConditionData.AddOutputVariable(variableData)
		}
	}
	conditions, err := impl.pipelineStageRepository.GetConditionsByStepId(stepId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting conditions by stepId", "err", err, "stepId", stepId)
		return variableAndConditionData, err
	}
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
			variableAndConditionData = variableAndConditionData.AddTriggerSkipCondition(conditionData)
		} else if condition.ConditionType == repository.PIPELINE_STAGE_STEP_CONDITION_TYPE_SUCCESS || condition.ConditionType == repository.PIPELINE_STAGE_STEP_CONDITION_TYPE_FAIL {
			variableAndConditionData = variableAndConditionData.AddSuccessFailureCondition(conditionData)
		}
	}
	return variableAndConditionData, nil
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

func (impl *PipelineStageServiceImpl) BuildPluginVariableAndConditionDataForWfRequest(stepId int) ([]*commonBean.VariableObject, []*commonBean.VariableObject, []*bean.ConditionObject, []*bean.ConditionObject, error) {
	var inputVariables []*commonBean.VariableObject
	var outputVariables []*commonBean.VariableObject
	//getting all variables in the step
	variables, err := impl.globalPluginRepository.GetVariablesByStepId(stepId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting variables by stepId", "err", err, "stepId", stepId)
		return nil, nil, nil, nil, err
	}
	variableNameIdMap := make(map[int]string)
	for _, variable := range variables {
		variableNameIdMap[variable.Id] = variable.Name
		format, err := commonBean.NewFormat(variable.Format.String())
		if err != nil {
			impl.logger.Errorw("error in creating format object", "err", err, "format", variable.Format)
			return nil, nil, nil, nil, err
		}
		variableData := &commonBean.VariableObject{
			Name:                       variable.Name,
			Format:                     format,
			ReferenceVariableStepIndex: variable.PreviousStepIndex,
			ReferenceVariableName:      variable.ReferenceVariableName,
			VariableStepIndexInPlugin:  variable.VariableStepIndexInPlugin,
			// default value for VariableType is commonBean.VariableTypeValue
			VariableType: commonBean.VariableTypeValue,
		}
		if variable.ValueType == repository2.PLUGIN_VARIABLE_VALUE_TYPE_NEW {
			variableData.VariableType = commonBean.VariableTypeValue
		} else if variable.ValueType == repository2.PLUGIN_VARIABLE_VALUE_TYPE_GLOBAL {
			variableData.VariableType = commonBean.VariableTypeRefGlobal
		} else if variable.ValueType == repository2.PLUGIN_VARIABLE_VALUE_TYPE_PREVIOUS && !variable.IsExposed {
			variableData.VariableType = commonBean.VariableTypeRefPlugin
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

	resolvedTemplate, variableSnapshot, err := impl.scopedVariableManager.GetMappedVariablesAndResolveTemplateBatch(string(responseJson), scope, entities)
	if err != nil {
		return nil, err
	}
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

	return impl.scopedVariableManager.ExtractAndMapVariables(string(requestJson), stageReq.Id, repository3.EntityTypePipelineStage, userId, tx)

}

func (impl *PipelineStageServiceImpl) IsScanPluginConfiguredAtPipelineStage(pipelineId int, pipelineStage repository.PipelineStageType, pluginName string) (bool, error) {
	plugin, err := impl.globalPluginRepository.GetPluginByName(pluginName)
	if err != nil {
		impl.logger.Errorw("error in getting image scanning plugin, Vulnerability Scanning", "pipelineId", pipelineId, "pipelineStage", pipelineStage, "err", err)
		return false, err
	}
	if len(plugin) == 0 {
		return false, nil
	}
	isScanPluginConfigured, err := impl.pipelineStageRepository.CheckIfPluginExistsInPipelineStage(pipelineId, pipelineStage, plugin[0].Id)
	if err != nil {
		impl.logger.Errorw("error in getting ci pipeline plugin", "err", err, "pipelineId", pipelineId, "pluginId", plugin[0].Id)
		return false, err
	}
	return isScanPluginConfigured, nil
}

// validateStepVariables validates the step variable
// It validates the following:
//   - variable.InternalBool is false, then it's an internal variable (not exposed in UI) and no validation is required
//   - variable.Name is mandatory
//   - format commonBean.Format is mandatory
//   - variable.Value should be a valid value for the format
//   - variable.Value is optional on few conditions: refer to &bean.StepVariableDto{}.IsEmptyValueAllowed()
//
// Input:
//   - variable: Type *bean.StepVariableDto; variable object to be validated
//   - isTriggerStage: Type bool; set to true if validation is for trigger stage (default is false)
//
// Returns:
//   - error: validation error (util.ApiError)
func validateStepVariable(variable *bean.StepVariableDto, isTriggerStage bool) error {
	if variable == nil {
		return nil
	}
	// validate name
	// if invalid, return error
	if len(variable.Name) == 0 {
		errMsg := fmt.Sprintf("variable name is mandatory")
		return util.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	// validate format
	// if invalid, return error
	// format is mandatory
	format, err := commonBean.NewFormat(variable.Format.String())
	if err != nil {
		errMsg := fmt.Sprintf("variable '%s' has invalid format '%s'", variable.Name, variable.Format)
		return util.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}

	// runtime args can be skipped if it is a configuration stage (create/ update).
	// but runtime args cannot be skipped if it is a trigger stage

	// validate value
	// if invalid, return error
	if !variable.IsEmptyValueAllowed(isTriggerStage) && variable.IsEmptyValue() {
		// runtime args can have empty value OR default value (must be a value from the choices)
		errMsg := fmt.Sprintf("variable '%s' does not allow empty value", variable.Name)
		return util.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}

	if len(variable.GetValue()) != 0 {
		// validate value based on format
		// convert value to format
		// if invalid, return error
		_, err = format.Convert(variable.GetValue())
		if err != nil {
			errMsg := fmt.Sprintf("variable '%s' has invalid value '%s' for format '%s'", variable.Name, variable.GetValue(), variable.Format)
			return util.NewApiError(http.StatusBadRequest, errMsg, errMsg)
		}
	}
	return nil
}

func validateStepVariables(variable []*bean.StepVariableDto, isTriggerStage bool) error {
	for _, v := range variable {
		err := validateStepVariable(v, isTriggerStage)
		if err != nil {
			return err
		}
	}
	return nil
}
