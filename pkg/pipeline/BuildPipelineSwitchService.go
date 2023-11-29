package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"time"
)

type BuildPipelineSwitchService interface {
	SwitchToExternalCi(tx *pg.Tx, appWorkflowMapping *appWorkflow.AppWorkflowMapping, switchFromCiPipelineId int, userId int32) error
	HandleSwitchCiPipeline(request *bean.CiPatchRequest, ciConfig *bean.CiConfigRequest) (*bean.CiConfigRequest, error)
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

	ciPipeline, err := impl.ciPipelineRepository.FindById(switchFromCiPipelineId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in finding ci-pipeline by id", "ciPipelineId", switchFromCiPipelineId, "err", err)
		return err
	}

	if ciPipeline == nil || err == pg.ErrNoRows {
		return errors.New("ci pipeline doesn't exist")
	}
	oldWorkflowMapping, err := impl.deleteOldPipelineAndWorkflowMappingBeforeSwitch(tx, ciPipeline, 0, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting old ci-pipeline and getting the appWorkflow mapping of that", "err", err, "userId", userId)
		return err
	}

	err = impl.updateLinkedAppWorkflowMappings(tx, oldWorkflowMapping, appWorkflowMapping)
	if err != nil {
		impl.logger.Errorw("error in updating linked app-workflow-mappings ", "oldAppWorkflowMappingId", oldWorkflowMapping.Id, "currentAppWorkflowMapId", appWorkflowMapping.Id, "err", err, "userId", userId)
		return err
	}

	//setting new ci_pipeline_id to 0 because we dont store ci_pipeline_id if the ci_pipeline is external/webhook type.
	err = impl.pipelineRepository.UpdateOldCiPipelineIdToNewCiPipelineId(tx, switchFromCiPipelineId, 0)
	if err != nil {
		impl.logger.Errorw("error in updating pipelines ci_pipeline_ids with new ci_pipelineId", "oldCiPipelineId", switchFromCiPipelineId)
		return err
	}

	return nil
}

func (impl *BuildPipelineSwitchServiceImpl) HandleSwitchCiPipeline(request *bean.CiPatchRequest, ciConfig *bean.CiConfigRequest) (*bean.CiConfigRequest, error) {
	if request.SwitchFromCiPipelineId != 0 && request.SwitchFromExternalCiPipelineId != 0 {
		return nil, errors.New("invalid request payload, both switchFromCiPipelineId and switchFromExternalCiPipelineId cannot be set in the payload")
	}

	//get the ciPipeline
	var switchFromType bean.PipelineType
	var ciPipeline *pipelineConfig.CiPipeline
	var err error
	if request.SwitchFromExternalCiPipelineId != 0 {
		switchFromType = bean.EXTERNAL
	} else {
		ciPipeline, err = impl.ciPipelineRepository.FindById(request.SwitchFromCiPipelineId)
		if err != nil {
			impl.logger.Errorw("error in finding ci-pipeline by id", "ciPipelineId", request.SwitchFromCiPipelineId, "err", err)
			return nil, err
		}
	}

	//validate switch request
	err = impl.validateCiPipelineSwitch(request.SwitchFromCiPipelineId, request.SwitchFromExternalCiPipelineId, request.CiPipeline.PipelineType, switchFromType)
	if err != nil {
		impl.logger.Errorw("error occurred when validating ci-pipeline switch", "err", err)
		return nil, err
	}

	//delete old pipeline and it's appworkflow mapping
	tx, err := impl.ciPipelineRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction", "err", err)
		return nil, err
	}
	defer impl.ciPipelineRepository.RollbackTx(tx)
	oldAppWorkflowMapping, err := impl.deleteOldPipelineAndWorkflowMappingBeforeSwitch(tx, ciPipeline, request.SwitchFromExternalCiPipelineId, request.UserId)
	if err != nil {
		impl.logger.Errorw("error in deleting old ci-pipeline and getting the appWorkflow mapping of that", "err", err, "userId", request.UserId)
		return nil, err
	}

	ciConfig.CiPipelines = []*bean.CiPipeline{request.CiPipeline} //request.CiPipeline
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
	//we don't store ci-pipeline-id in pipeline table for
	if request.SwitchFromCiPipelineId > 0 {
		// ciPipeline id is being set in res object in the addpipelineToTemplate method.
		//
		err = impl.pipelineRepository.UpdateOldCiPipelineIdToNewCiPipelineId(tx, request.SwitchFromCiPipelineId, res.CiPipelines[0].Id)
		if err != nil {
			impl.logger.Errorw("error in updating pipelines ci_pipeline_ids with new ci_pipelineId", "oldCiPipelineId", request.SwitchFromCiPipelineId, "newCiPipelineId", res.CiPipelines[0].Id)
			return nil, err
		}
	}

	err = impl.ciPipelineRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in commiting the transaction")
		return nil, err
	}
	return res, nil
}

func (impl *BuildPipelineSwitchServiceImpl) validateCiPipelineSwitch(switchFromCiPipelineId int, switchFromExternalCiPipelineId int, switchToType, switchFromType bean.PipelineType) error {
	// this will only allow below conversions
	// ext -> {ci_job,direct,linked}
	// direct -> {ci_job,linked}
	// linked -> {ci_job,direct}
	// ci_job -> {direct,linked}

	if switchFromCiPipelineId != 0 && switchFromExternalCiPipelineId != 0 {
		return errors.New("invalid request payload, both switchFromCiPipelineId and switchFromExternalCiPipelineId cannot be set in the payload")
	}
	if switchToType == switchFromType {
		return errors.New("cannot convert this pipeline to same type")
	}

	if switchToType == bean.EXTERNAL {
		return errors.New("current ci-pipeline cannot be converted to external-webhook type")
	}

	// we should not check the below logic for external_ci type as builds are not built in devtron and
	// linked pipelines won't be there as per current external-ci-pipeline architecture
	if switchFromCiPipelineId > 0 {
		// old ci_pipeline should not contain any linked ci_pipelines.
		linkedCiPipelines, err := impl.ciPipelineRepository.FindLinkedCiCount(switchFromCiPipelineId)
		if err != nil {
			return nil
		}
		if linkedCiPipelines > 0 {
			return errors.New("cannot convert this ci-pipeline as it contains some linked ci-pipeline's")
		}

		// find any builds running on old ci_pipeline, if yes block this conversion with proper message.
		inProgressWorkflows, err := impl.ciWorkflowRepository.FindRunningWorkflowCount(switchFromCiPipelineId)
		if inProgressWorkflows > 0 {
			return errors.New("cannot convert this ci-pipeline as some builds of this ci-pipeline are in progressing state")
		}
	}

	return nil
}

func (impl *BuildPipelineSwitchServiceImpl) deleteOldPipelineAndWorkflowMappingBeforeSwitch(tx *pg.Tx, ciPipeline *pipelineConfig.CiPipeline, externalCiPipelineId int, userId int32) (*appWorkflow.AppWorkflowMapping, error) {
	//deleting ciPipeline in tx
	err := impl.deleteBuildPipeline(tx, ciPipeline, externalCiPipelineId, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting ciPipeline", "err", err)
		return nil, err
	}
	//getting appWorkflowMapping of current ciPipeline
	pipelineType, pipelineId := impl.checkForExternalPipelineAndGetID(externalCiPipelineId, ciPipeline)
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
		impl.logger.Errorw("error in deleting workflow mapping", "err", err)
		return appWorkflowMappings, err
	}
	return appWorkflowMappings, nil
}

func (impl *BuildPipelineSwitchServiceImpl) deleteBuildPipeline(tx *pg.Tx, ciPipeline *pipelineConfig.CiPipeline, externalCiPipelineId int, userId int32) error {
	if externalCiPipelineId != 0 {
		err := impl.deleteExternalCi(tx, externalCiPipelineId, userId)
		if err != nil {
			impl.logger.Errorw("error in deleting external ci pipeline", "externalCiPipelineId", externalCiPipelineId, "err", err)
			return err
		}
		return nil
	}
	err := impl.ciCdPipelineOrchestrator.DeleteCiEnvMapping(tx, ciPipeline, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting ci-env mappings", "ciPipelineId", ciPipeline.Id, "err", err)
		return err
	}
	ciPipeline.Deleted = true
	ciPipeline.Active = false
	ciPipeline.UpdatedOn = time.Now()
	ciPipeline.UpdatedBy = userId
	err = impl.ciPipelineRepository.Update(ciPipeline, tx)
	if err != nil {
		impl.logger.Errorw("error in updating ci pipeline, DeleteCiPipeline", "err", err, "pipelineId", ciPipeline.Id)
		return err
	}
	materials, err := impl.DeleteCiMaterial(tx, ciPipeline)
	if err != nil {
		return err
	}
	if !ciPipeline.IsDockerConfigOverridden {
		err = impl.ciCdPipelineOrchestrator.SaveHistoryOfBaseTemplate(userId, ciPipeline, materials)
		if err != nil {
			return err
		}
	} else if ciPipeline.ParentCiPipeline == 0 {
		err = impl.saveHistoryOfOverriddenTemplate(ciPipeline, userId, materials)
		if err != nil {
			return err
		}
	}
	return err
}

func (impl *BuildPipelineSwitchServiceImpl) deleteExternalCi(tx *pg.Tx, externalCiPipelineId int, userId int32) error {
	externalCiPipeline, err := impl.ciPipelineRepository.FindExternalCiById(externalCiPipelineId)
	externalCiPipeline.Active = false
	externalCiPipeline.AuditLog = sql.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()}
	_, err = impl.ciPipelineRepository.UpdateExternalCi(externalCiPipeline, tx)
	if err != nil {
		impl.logger.Errorw("error in deleting workflow mapping", "err", err)
		return err
	}
	return nil
}

func (impl *BuildPipelineSwitchServiceImpl) DeleteCiMaterial(tx *pg.Tx, ciPipeline *pipelineConfig.CiPipeline) ([]*pipelineConfig.CiPipelineMaterial, error) {
	materialDbObject, err := impl.ciPipelineMaterialRepository.GetByPipelineId(ciPipeline.Id)
	var materials []*pipelineConfig.CiPipelineMaterial
	if err != nil && err != pg.ErrNoRows {
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
	buildConfig, err := bean3.ConvertDbBuildConfigToBean(ciTemplate.CiBuildConfig)
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
	return impl.appWorkflowRepository.UpdateParentComponentDetails(tx, oldAppWorkflowMapping.ComponentId, oldAppWorkflowMapping.Type, newAppWorkflowMapping.ComponentId, newAppWorkflowMapping.Type)
}

func (impl *BuildPipelineSwitchServiceImpl) checkForExternalPipelineAndGetID(externalCiPipelineId int, ciPipeline *pipelineConfig.CiPipeline) (string, int) {
	if externalCiPipelineId != 0 {
		return appWorkflow.WEBHOOK, externalCiPipelineId
	}
	return appWorkflow.CIPIPELINE, ciPipeline.Id
}
