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
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/adapter"
	pipelineConfigBean "github.com/devtron-labs/devtron/pkg/pipeline/bean/CiPipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type SwitchBuildPipelineValidationError string

const (
	cannotConvertToSameType                           SwitchBuildPipelineValidationError = "cannot convert this pipeline to same type"
	cannotConvertToExternalCi                         SwitchBuildPipelineValidationError = "current ci-pipeline cannot be converted to external-webhook type"
	cannotConvertIfLinkedCiFound                      SwitchBuildPipelineValidationError = "cannot convert this ci-pipeline as it contains some linked ci-pipeline's"
	cannotConvertIfLatestWorkflowIsInNonTerminalState SwitchBuildPipelineValidationError = "cannot convert this ci-pipeline as recent build of this ci-pipeline is in progressing state"
)

type BuildPipelineSwitchService interface {
	SwitchToExternalCi(tx *pg.Tx, appWorkflowMapping *appWorkflow.AppWorkflowMapping, switchFromCiPipelineId int, userId int32) error
	SwitchToCiPipelineExceptExternal(request *bean.CiPatchRequest, ciConfig *bean.CiConfigRequest) (*bean.CiConfigRequest, error)
}

type BuildPipelineSwitchServiceImpl struct {
	logger                       *zap.SugaredLogger
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
	ciCdPipelineOrchestrator     CiCdPipelineOrchestrator
	pipelineRepository           pipelineConfig.PipelineRepository
	ciWorkflowRepository         pipelineConfig.CiWorkflowRepository
	appWorkflowRepository        appWorkflow.AppWorkflowRepository
	ciPipelineHistoryService     history.CiPipelineHistoryService
	ciTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
}

func NewBuildPipelineSwitchServiceImpl(logger *zap.SugaredLogger,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	ciCdPipelineOrchestrator CiCdPipelineOrchestrator,
	pipelineRepository pipelineConfig.PipelineRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	ciPipelineHistoryService history.CiPipelineHistoryService,
	ciTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository) *BuildPipelineSwitchServiceImpl {
	return &BuildPipelineSwitchServiceImpl{
		logger:                       logger,
		ciPipelineRepository:         ciPipelineRepository,
		ciCdPipelineOrchestrator:     ciCdPipelineOrchestrator,
		pipelineRepository:           pipelineRepository,
		ciWorkflowRepository:         ciWorkflowRepository,
		appWorkflowRepository:        appWorkflowRepository,
		ciPipelineHistoryService:     ciPipelineHistoryService,
		ciTemplateOverrideRepository: ciTemplateOverrideRepository,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
	}
}

func (impl *BuildPipelineSwitchServiceImpl) SwitchToExternalCi(tx *pg.Tx, appWorkflowMapping *appWorkflow.AppWorkflowMapping, switchFromCiPipelineId int, userId int32) error {

	err := impl.validateSwitchPreConditions(switchFromCiPipelineId)
	if err != nil {
		return err
	}

	err = impl.deleteCiAndItsWorkflowMappings(tx, switchFromCiPipelineId, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting old ci-pipeline and getting the appWorkflow mapping of that", "err", err, "userId", userId)
		return err
	}
	oldWorkflowMapping, err := impl.deleteAndGetAppWorkflowMappings(tx, appWorkflow.CIPIPELINE, switchFromCiPipelineId, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting workflow mapping", "switchFromCiPipelineId", switchFromCiPipelineId, "err", err)
		return err
	}
	err = impl.updateLinkedAppWorkflowMappings(tx, oldWorkflowMapping, appWorkflowMapping)
	if err != nil {
		impl.logger.Errorw("error in updating linked app-workflow-mappings ", "oldAppWorkflowMappingId", oldWorkflowMapping.Id, "currentAppWorkflowMapId", appWorkflowMapping.Id, "err", err, "userId", userId)
		return err
	}

	// setting new ci_pipeline_id to 0 because we dont store ci_pipeline_id if the ci_pipeline is external/webhook type.
	err = impl.pipelineRepository.UpdateOldCiPipelineIdToNewCiPipelineId(tx, switchFromCiPipelineId, 0)
	if err != nil {
		impl.logger.Errorw("error in updating pipelines ci_pipeline_ids with new ci_pipelineId", "oldCiPipelineId", switchFromCiPipelineId)
		return err
	}

	return nil
}

func (impl *BuildPipelineSwitchServiceImpl) SwitchToCiPipelineExceptExternal(request *bean.CiPatchRequest, ciConfig *bean.CiConfigRequest) (*bean.CiConfigRequest, error) {
	if request.SwitchFromCiPipelineId != 0 && request.SwitchFromExternalCiPipelineId != 0 {
		return nil, errors.New("invalid request payload, both switchFromCiPipelineId and switchFromExternalCiPipelineId cannot be set in the payload")
	}

	//get the ciPipeline
	var switchFromType pipelineConfigBean.PipelineType
	var switchFromPipelineId int
	if request.SwitchFromExternalCiPipelineId != 0 {
		switchFromType = pipelineConfigBean.EXTERNAL
		switchFromPipelineId = request.SwitchFromExternalCiPipelineId
	} else {
		switchFromPipelineId = request.SwitchFromCiPipelineId
		switchFromType = request.SwitchFromCiPipelineType
	}

	//validate switch request
	err := impl.validateCiPipelineSwitch(switchFromPipelineId, request.CiPipeline.PipelineType, switchFromType)
	if err != nil {
		impl.logger.Errorw("validating failed for ci-pipeline switch request", "switchFromPipelineId", "switchFromType", switchFromType, "pipelineType", request.CiPipeline.PipelineType, "err", err)
		return nil, err
	}

	// delete old pipeline and it's appworkflow mapping
	return impl.createNewPipelineAndReplaceOldPipelineLinks(request.CiPipeline, ciConfig, switchFromPipelineId, switchFromType, request.UserId)
}

func (impl *BuildPipelineSwitchServiceImpl) createNewPipelineAndReplaceOldPipelineLinks(ciPipelineReq *bean.CiPipeline, ciConfig *bean.CiConfigRequest, switchFromPipelineId int, switchFromType pipelineConfigBean.PipelineType, userId int32) (*bean.CiConfigRequest, error) {

	isSelfLinkedCiPipeline := switchFromType != pipelineConfigBean.EXTERNAL && ciPipelineReq.IsLinkedCi() && ciPipelineReq.ParentCiPipeline == switchFromPipelineId
	if isSelfLinkedCiPipeline {
		errMsg := "cannot create linked ci pipeline from the same source"
		return nil, util.NewApiError().WithInternalMessage(errMsg).WithUserMessage(errMsg).WithHttpStatusCode(http.StatusBadRequest)
	}

	tx, err := impl.ciPipelineRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction", "switchFromPipelineId", switchFromPipelineId, "switchFromType", switchFromType, "userId", userId, "err", err)
		return nil, err
	}
	defer impl.ciPipelineRepository.RollbackTx(tx)
	oldAppWorkflowMapping, err := impl.deleteOldCiPipelineAndWorkflowMappingBeforeSwitch(tx, switchFromPipelineId, switchFromType, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting old ci-pipeline and getting the appWorkflow mapping of that", "err", err, "userId", userId)
		return nil, err
	}

	ciConfig.CiPipelines = []*bean.CiPipeline{ciPipelineReq} //request.CiPipeline
	res, err := impl.ciCdPipelineOrchestrator.AddPipelineToTemplate(ciConfig, true)
	if err != nil {
		impl.logger.Errorw("error in adding pipeline to template", "ciConf", ciConfig, "err", err)
		return nil, err
	}

	//go and update all the app workflow mappings of old ci_pipeline with new ci_pipeline_id.
	err = impl.updateLinkedAppWorkflowMappings(tx, oldAppWorkflowMapping, res.AppWorkflowMapping)
	if err != nil {
		impl.logger.Errorw("error in updating app workflow mappings", "err", err)
		return nil, err
	}

	if switchFromPipelineId > 0 {
		// get all the cd workflow mappings whose parent component is our old pipeline
		cdwfmappings, err := impl.appWorkflowRepository.FindWFCDMappingsByWorkflowId(oldAppWorkflowMapping.AppWorkflowId)
		if err != nil {
			impl.logger.Errorw("error in finding parent cd workflowMappings using parent component details", "parentComponentType", oldAppWorkflowMapping.Type, "parentComponentId", oldAppWorkflowMapping.ComponentId, "err", err)
			return nil, err
		}
		pipelineIds := make([]int, 0, len(cdwfmappings))
		for _, cdwfMapping := range cdwfmappings {
			pipelineIds = append(pipelineIds, cdwfMapping.ComponentId)
		}

		err = impl.pipelineRepository.UpdateCiPipelineId(tx, pipelineIds, res.CiPipelines[0].Id)
		if err != nil {
			impl.logger.Errorw("error in updating pipelines ci_pipeline_ids with new ci_pipelineId", "oldCiPipelineId", switchFromPipelineId, "newCiPipelineId", res.CiPipelines[0].Id, "err", err)
			return nil, err
		}
	}

	err = impl.ciPipelineRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing the transaction", "switchFromPipelineId", switchFromPipelineId, "switchFromType", switchFromType, "userId", userId, "err", err)
		return nil, err
	}
	return res, nil
}

// add switchType and remove other id
// make constants for error msgs
func (impl *BuildPipelineSwitchServiceImpl) validateCiPipelineSwitch(switchFromCiPipelineId int, switchToType, switchFromType pipelineConfigBean.PipelineType) error {
	// this will only allow below conversions
	// ext -> {ci_job,direct,linked}
	// direct -> {ci_job,linked}
	// linked -> {ci_job,direct}
	// ci_job -> {direct,linked}

	if switchToType == switchFromType {
		return errors.New(string(cannotConvertToSameType))
	}

	// refer SwitchToExternalCi
	if switchToType == pipelineConfigBean.EXTERNAL {
		return errors.New(string(cannotConvertToExternalCi))
	}

	// we should not check the below logic for external_ci type as builds are not built in devtron and
	// linked pipelines won't be there as per current external-ci-pipeline architecture
	if switchFromCiPipelineId > 0 && switchFromType != pipelineConfigBean.EXTERNAL {
		err := impl.validateSwitchPreConditions(switchFromCiPipelineId)
		if err != nil {
			return err
		}
	}

	return nil
}
func (impl *BuildPipelineSwitchServiceImpl) deleteCiAndItsWorkflowMappings(tx *pg.Tx, switchFromPipelineId int, userId int32) error {
	ciPipeline, err := impl.ciPipelineRepository.FindById(switchFromPipelineId)
	if err == pg.ErrNoRows {
		impl.logger.Errorw("no ci-pipeline found for given switchFromCiPipelineId", "switchFromCiPipelineId", switchFromPipelineId)
		return errors.New("requested ci pipeline doesn't exist")
	}
	if err != nil {
		impl.logger.Errorw("error in finding ci-pipeline by id", "ciPipelineId", switchFromPipelineId, "err", err)
		return err
	}
	err = impl.deleteBuildPipeline(tx, ciPipeline, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting ciPipeline", "ciPipelineId", ciPipeline, "err", err)
	}
	return err
}

func (impl *BuildPipelineSwitchServiceImpl) deleteOldCiPipelineAndWorkflowMappingBeforeSwitch(tx *pg.Tx, switchFromPipelineId int, switchFromType pipelineConfigBean.PipelineType, userId int32) (*appWorkflow.AppWorkflowMapping, error) {
	// 1) delete build pipelines
	// 2) delete app workflowMappings
	var err error
	pipelineId := switchFromPipelineId
	pipelineType := ""
	if switchFromType == pipelineConfigBean.EXTERNAL {
		err = impl.deleteExternalCi(tx, switchFromPipelineId, userId)
		pipelineType = appWorkflow.WEBHOOK
	} else {
		err = impl.deleteCiAndItsWorkflowMappings(tx, switchFromPipelineId, userId)
		pipelineType = appWorkflow.CIPIPELINE
	}
	if err != nil {
		impl.logger.Errorw("error in deleting ci pipeline before switching ", "switchFromPipelineId", switchFromPipelineId, "switchFromType", switchFromType, "err", err)
		return nil, err
	}
	return impl.deleteAndGetAppWorkflowMappings(tx, pipelineType, pipelineId, userId)
}

func (impl *BuildPipelineSwitchServiceImpl) deleteAndGetAppWorkflowMappings(tx *pg.Tx, pipelineType string, pipelineId int, userId int32) (*appWorkflow.AppWorkflowMapping, error) {
	appWorkflowMappings, err := impl.appWorkflowRepository.FindWFMappingByComponent(pipelineType, pipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting  appWorkflowMappings", "err", err, "pipelineType", pipelineType, "pipelineId", pipelineId)
		return appWorkflowMappings, err
	}
	//deleting  app workflow mapping in tx
	appWorkflowMappings.UpdatedBy = userId
	appWorkflowMappings.UpdatedOn = time.Now()
	err = impl.appWorkflowRepository.DeleteAppWorkflowMapping(appWorkflowMappings, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting workflow mapping", "CiPipelineType", pipelineType, "pipelineId", pipelineId, "err", err)
		return appWorkflowMappings, err
	}
	return appWorkflowMappings, nil
}

func (impl *BuildPipelineSwitchServiceImpl) deleteBuildPipeline(tx *pg.Tx, ciPipeline *pipelineConfig.CiPipeline, userId int32) error {
	err := impl.ciCdPipelineOrchestrator.DeleteCiPipelineAndCiEnvMappings(tx, ciPipeline, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting ci pipeline and its env mappings", "pipelineId", ciPipeline.Id, "err", err)
		return err
	}
	//not deleting ciPipeline material or template override as these can be useful for artifact built from the old ciPipeline
	return err
}

func (impl *BuildPipelineSwitchServiceImpl) deleteExternalCi(tx *pg.Tx, externalCiPipelineId int, userId int32) error {
	externalCiPipeline, err := impl.ciPipelineRepository.FindExternalCiById(externalCiPipelineId)
	externalCiPipeline.Active = false
	externalCiPipeline.AuditLog = sql.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()}
	_, err = impl.ciPipelineRepository.UpdateExternalCi(externalCiPipeline, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting workflow mapping", "externalCiPipelineId", externalCiPipelineId, "err", err)
		return err
	}
	return nil
}

func (impl *BuildPipelineSwitchServiceImpl) DeleteCiMaterial(tx *pg.Tx, ciPipeline *pipelineConfig.CiPipeline) ([]*pipelineConfig.CiPipelineMaterial, error) {
	materialDbObject, err := impl.ciPipelineMaterialRepository.GetByPipelineId(ciPipeline.Id)
	var materials []*pipelineConfig.CiPipelineMaterial
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting ci pipeline materials", "externalCiPipelineId", "ciPipelineId", ciPipeline.Id, "err", err)
		return materials, err
	}
	if len(materialDbObject) == 0 {
		return materials, nil
	}
	for _, material := range materialDbObject {
		material.Active = false
		materials = append(materials, material)
	}
	err = impl.ciPipelineMaterialRepository.Update(tx, materials...)
	if err != nil {
		impl.logger.Errorw("error in updating ci pipeline materials, DeleteCiPipeline", "err", err, "pipelineId", ciPipeline.Id)
		return materials, err
	}
	return materials, nil
}

func (impl *BuildPipelineSwitchServiceImpl) saveHistoryOfOverriddenTemplate(ciPipeline *pipelineConfig.CiPipeline, userId int32, materials []*pipelineConfig.CiPipelineMaterial) error {
	ciTemplate, err := impl.ciTemplateOverrideRepository.FindByCiPipelineId(ciPipeline.Id)
	if err != nil {
		impl.logger.Errorw("error in getting ciTemplate ", "err", err, "ciTemplate", ciTemplate.Id)
		return err
	}
	buildConfig, err := adapter.ConvertDbBuildConfigToBean(ciTemplate.CiBuildConfig)
	if err != nil {
		impl.logger.Errorw("error in ConvertDbBuildConfigToBean ", "err", err, "buildConfigId", buildConfig.Id)
		return err
	}
	CiTemplateBean := impl.ciCdPipelineOrchestrator.CreateCiTemplateBean(ciPipeline.Id, ciTemplate.DockerRegistry.Id, ciTemplate.DockerRepository, ciTemplate.GitMaterialId, buildConfig, userId)
	err = impl.ciPipelineHistoryService.SaveHistory(ciPipeline, materials, &CiTemplateBean, repository4.TRIGGER_DELETE)
	if err != nil {
		impl.logger.Errorw("error in saving delete history for ci pipeline material and ci template overridden", "err", err)
	}
	return nil
}

func (impl *BuildPipelineSwitchServiceImpl) updateLinkedAppWorkflowMappings(tx *pg.Tx, oldAppWorkflowMapping *appWorkflow.AppWorkflowMapping, newAppWorkflowMapping *appWorkflow.AppWorkflowMapping) error {
	return impl.appWorkflowRepository.UpdateParentComponentDetails(tx, oldAppWorkflowMapping.ComponentId, oldAppWorkflowMapping.Type, newAppWorkflowMapping.ComponentId, newAppWorkflowMapping.Type, nil)
}

func (impl *BuildPipelineSwitchServiceImpl) validateSwitchPreConditions(switchFromCiPipelineId int) error {

	// old ci_pipeline should not contain any linked ci_pipelines.
	linkedCiPipelines, err := impl.ciPipelineRepository.FindLinkedCiCount(switchFromCiPipelineId)
	if err != nil {
		impl.logger.Errorw("error in finding the linkedCi count for the pipeline", "ciPipelineId", switchFromCiPipelineId, "err", err)
		return err
	}
	if linkedCiPipelines > 0 {
		return errors.New(string(cannotConvertIfLinkedCiFound))
	}

	// note: ideally we should have found any builds running on old ci_pipeline, if yes block this conversion with proper message.
	// but checking only latest wf for now.
	ciWorkflow, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflow(switchFromCiPipelineId)
	// no build is triggered case
	if err == pg.ErrNoRows {
		return nil
	}
	if err != nil {
		impl.logger.Errorw("error in finding latest ciwokflow by ciPipelineId", "ciPipelineId", switchFromCiPipelineId)
		return err
	}

	if ciWorkflow.InProgress() {
		return errors.New(string(cannotConvertIfLatestWorkflowIsInNonTerminalState))
	}

	return nil
}
