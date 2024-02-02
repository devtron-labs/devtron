package service

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/middleware"
	repository5 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/plugin"
	repository4 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/devtron-labs/devtron/pkg/workflow/pipeline/ci/materials/service"
	materialTypes "github.com/devtron-labs/devtron/pkg/workflow/pipeline/ci/materials/types"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/ci/types"
	wfTypes "github.com/devtron-labs/devtron/pkg/workflow/trigger/core/types"
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

type CiTriggerService interface {
	TriggerPipeline(ctx bean.RequestContext, trigger types.CiTriggerRequest) (int, error)
}

type CiTriggerServiceImpl struct {
	logger                     *zap.SugaredLogger
	materialService            service.MaterialService
	ciPipelineConfigService    pipeline.CiPipelineConfigService
	pipelineStageService       pipeline.PipelineStageService
	ciCdConfig                 wfTypes.CiCdConfig
	customTagService           pipeline.CustomTagService
	ciTemplateService          pipeline.CiTemplateService
	globalPluginService        plugin.GlobalPluginService
	pluginInputVariableParser  pipeline.PluginInputVariableParser
	ciWorkflowRunnerRepository pipelineConfig.CiWorkflowRunnerRepository
	appConfigService           app.AppConfigService
	environmentService         cluster.EnvironmentService
	appCrudOperationService    app.AppCrudOperationService
}

func NewCiTriggerServiceImpl(logger *zap.SugaredLogger, service service.MaterialService, ciPipelineConfigService pipeline.CiPipelineConfigService,
	pipelineStageService pipeline.PipelineStageService, customTagService pipeline.CustomTagService, ciTemplateService pipeline.CiTemplateService,
	globalPluginService plugin.GlobalPluginService, pluginInputVariableParser pipeline.PluginInputVariableParser,
	ciWorkflowRunnerRepository pipelineConfig.CiWorkflowRunnerRepository, appConfigService app.AppConfigService,
	environmentService cluster.EnvironmentService, appCrudOperationService app.AppCrudOperationService) *CiTriggerServiceImpl {
	return &CiTriggerServiceImpl{
		logger:                     logger,
		materialService:            service,
		ciPipelineConfigService:    ciPipelineConfigService,
		pipelineStageService:       pipelineStageService,
		customTagService:           customTagService,
		ciTemplateService:          ciTemplateService,
		globalPluginService:        globalPluginService,
		pluginInputVariableParser:  pluginInputVariableParser,
		ciWorkflowRunnerRepository: ciWorkflowRunnerRepository,
		appConfigService:           appConfigService,
		environmentService:         environmentService,
		appCrudOperationService:    appCrudOperationService,
	}
}

// LoadContext(USER, PIPELINE, APP, ENV)
func (impl *CiTriggerServiceImpl) TriggerPipeline(ctx bean.RequestContext, triggerRequest types.CiTriggerRequest) (int, error) {
	ciPipelineId := triggerRequest.PipelineId
	impl.logger.Infow("ci trigger request came", "pipelineId", ciPipelineId)

	workflowRequest := &wfTypes.WorkflowTriggerRequest{}
	workflowRequest.LoadFromUserContext(ctx.GetUser())

	//TODO KB: load Job metadata
	err := impl.loadTriggerMetadata(triggerRequest, workflowRequest)
	if err != nil {
		return 0, err
	}
	// fetch ci materials
	err = impl.loadCiMaterials(triggerRequest, workflowRequest)
	if err != nil {
		return 0, err
	}
	savedCiWf, err := impl.saveNewWorkflow(triggerRequest, workflowRequest)
	if err != nil {
		return 0, err
	}
	savedCiWfId := savedCiWf.Id
	err = impl.loadFromEnvAndWorkflowConfig(savedCiWfId, ciPipelineId, workflowRequest)
	if err != nil {
		return 0, err
	}

	err = impl.loadPipelineScripts(ciPipelineId, workflowRequest)
	if err != nil {
		return 0, err
	}

	err = impl.loadDockerData(savedCiWfId, triggerRequest.CommitHashes, workflowRequest)
	if errors.Is(err, bean2.ErrImagePathInUse) {
		savedCiWf.Status = pipelineConfig.WorkflowFailed
		savedCiWf.Message = bean2.ImageTagUnavailableMessage
		err1 := impl.updateCiWorkflow(workflowRequest, savedCiWf)
		//err1 := impl.ciWorkflowRunnerRepository.UpdateWorkFlow(savedCiWf) // directly create wf
		if err1 != nil {
			impl.logger.Errorw("could not save workflow, after failing to load docker data", "err", err)
		}
		return 0, err
	}
	savedCiWf.ImagePathReservationIds = workflowRequest.ImagePathReservationIds

	//savedCiWf.LogLocation = impl.ciCdConfig.CiDefaultBuildLogsKeyPrefix + "/" + workflowRequest.WorkflowNamePrefix + "/main.log"
	savedCiWf.LogLocation = fmt.Sprintf("%s/%s/main.log", impl.ciCdConfig.GetDefaultBuildLogsKeyPrefix(), workflowRequest.WorkflowNamePrefix)
	err = impl.updateCiWorkflow(workflowRequest, savedCiWf)

	//workflowRequest.Env = env
	err = impl.executeCiPipeline(workflowRequest)
	if err != nil {
		impl.logger.Errorw("workflow error", "err", err)
		return 0, err
	}
	impl.logger.Debugw("ci triggered", " pipeline ", triggerRequest.PipelineId)

	impl.performPostTriggerOperations(savedCiWf, err, workflowRequest)

	return savedCiWfId, nil

}

func (impl *CiTriggerServiceImpl) updateCiWorkflow(request *wfTypes.WorkflowTriggerRequest, savedWf *pipelineConfig.CiWorkflow) error {
	ciBuildConfig := request.CiBuildConfig
	ciBuildType := string(ciBuildConfig.CiBuildType)
	savedWf.CiBuildType = ciBuildType
	return impl.ciWorkflowRunnerRepository.UpdateWorkFlow(savedWf)
}

func (impl *CiTriggerServiceImpl) performPostTriggerOperations(savedCiWf *pipelineConfig.CiWorkflow, err error, workflowRequest *wfTypes.WorkflowTriggerRequest) {
	var variableSnapshotHistories = util3.GetBeansPtr(
		repository4.GetSnapshotBean(savedCiWf.Id, repository4.HistoryReferenceTypeCIWORKFLOW, variableSnapshot))
	if len(variableSnapshotHistories) > 0 {
		err = impl.scopedVariableManager.SaveVariableHistoriesForTrigger(variableSnapshotHistories, trigger.TriggeredBy)
		if err != nil {
			impl.logger.Errorf("Not able to save variable snapshot for CI trigger %s", err)
		}
	}

	middleware.CiTriggerCounter.WithLabelValues(pipeline.App.AppName, pipeline.Name).Inc()
	go impl.WriteCITriggerEvent(trigger, pipeline, workflowRequest)
}

func (impl *CiTriggerServiceImpl) loadFromEnvAndWorkflowConfig(uniqueId int, ciPipelineId int, workflowRequest *wfTypes.WorkflowTriggerRequest) error {
	// TODO KB: load ci workflow config
	ciWorkflowConfig, err := impl.ciWorkflowRunnerRepository.FindConfigByPipelineId(ciPipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("could not fetch ci config", "ciPipeline", ciPipelineId, "err", err)
		return err
	}

	err = workflowRequest.UpdateFromEnvAndWorkflowConfig(uniqueId, ciWorkflowConfig, impl.ciCdConfig)
	return err
}

func (impl *CiTriggerServiceImpl) loadCiMaterials(triggerRequest types.CiTriggerRequest, workflowRequest *wfTypes.WorkflowTriggerRequest) error {
	materialModels, err := impl.getCiMaterials(triggerRequest)
	if err != nil {
		return err
	}
	workflowRequest.UpdateProjectMaterials(triggerRequest.CommitHashes, materialModels)
	return nil
}

func (impl *CiTriggerServiceImpl) loadPipelineScripts(ciPipelineId int, workflowRequest *wfTypes.WorkflowTriggerRequest) error {
	prePostAndRefPluginResponse, err := impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(ciPipelineId, bean2.CiStage, workflowRequest.Scope)
	if err != nil {
		impl.logger.Errorw("error in getting steps data for wf request", "ciPipelineId", ciPipelineId, "err", err)
		return err
	}

	//fetch pipeline scripts
	beforeScripts, afterScripts, err := impl.ciPipelineConfigService.FindCiScriptsByCiPipelineId(ciPipelineId)
	if err != nil {
		return nil
	}
	workflowRequest.UpdatePipelineScripts(prePostAndRefPluginResponse, beforeScripts, afterScripts)
	return nil
}

func (impl *CiTriggerServiceImpl) getCiMaterials(ciTriggerRequest types.CiTriggerRequest) ([]*materialTypes.CiPipelineMaterialModel, error) {
	var pipelineId int
	var ciMaterials []*materialTypes.CiPipelineMaterialModel
	var err error
	if !(len(ciMaterials) == 0) {
		return ciMaterials, nil
	} else {
		ciMaterials, err = impl.materialService.GetByPipelineId(pipelineId)
		if err != nil {
			impl.logger.Errorw("error occurred while fetching pipeline material info", "pipelineId", pipelineId, "err", err)
			return nil, err
		}
		if ciTriggerRequest.PipelineType == bean2.CI_JOB && len(ciMaterials) != 0 {
			ciMaterials = []*materialTypes.CiPipelineMaterialModel{ciMaterials[0]}
			ciMaterials[0].GitMaterial = nil
			ciMaterials[0].GitMaterialId = 0
		}
		return ciMaterials, nil
	}
}

func (impl *CiTriggerServiceImpl) loadDockerData(uniqueId int, commitHashes map[int]pipelineConfig.GitCommit, request *wfTypes.WorkflowTriggerRequest) error {

	ciPipelineId := request.PipelineId
	ciTemplateMetadata, err := impl.ciPipelineConfigService.GetCiTemplateMetadata(request.AppId, ciPipelineId)
	if err != nil {
		return err
	}
	// TODO KB: load template metadata to WorkflowTriggerRequest
	err = impl.getDockerImageTag(uniqueId, commitHashes, request, ciTemplateMetadata)
	if err != nil {
		return err
	}
	request.CheckoutPath = ciTemplateMetadata.CheckoutPath
	return nil
}

func (impl *CiTriggerServiceImpl) getDockerImageTag(uniqueId int, commitHashes map[int]pipelineConfig.GitCommit, request *wfTypes.WorkflowTriggerRequest, ciTemplateMetadata *bean2.CiTemplateMetadata) error {
	var imageReservationIds []int
	var dockerImageTag string
	dockerRegistry := ciTemplateMetadata.DockerRegistry
	dockerRepository := ciTemplateMetadata.DockerRepository
	ciPipelineId := request.PipelineId
	ciPipelineIdString := strconv.Itoa(ciPipelineId)
	customTag, err := impl.customTagService.GetActiveCustomTagByEntityKeyAndValue(bean2.EntityTypeCiPipelineId, ciPipelineIdString)
	if err != nil && err != pg.ErrNoRows {
		return err
	}
	if customTag.Id != 0 && customTag.Enabled == true {
		imagePathReservation, err := impl.customTagService.GenerateImagePath(bean2.EntityTypeCiPipelineId, ciPipelineIdString, dockerRegistry.RegistryURL, dockerRepository)
		if err != nil {
			return err
		}
		imageReservationIds = []int{imagePathReservation.Id}
		imagePathSplit := strings.Split(imagePathReservation.ImagePath, ":")
		if len(imagePathSplit) >= 1 {
			dockerImageTag = imagePathSplit[len(imagePathSplit)-1]
		}
	} else {
		dockerImageTag = impl.buildImageTag(commitHashes, ciPipelineId, uniqueId)
	}
	// image copy plugin
	if !request.IsJob() {
		buildImagePath := fmt.Sprintf(bean2.ImagePathPattern, dockerRegistry.RegistryURL, dockerRepository, dockerImageTag)
		err = impl.loadCopyContainerImagePluginMetadata(dockerImageTag, customTag.Id, buildImagePath, dockerRegistry.Id, request)
		if err != nil {
			//TODO KB: save it in db
			impl.logger.Errorw("error in getting env variables for copyContainerImage plugin", "ciPipelineId", ciPipelineId, "err", err)
			//savedWf.Status = pipelineConfig.WorkflowFailed
			//savedWf.Message = err.Error()
			//err1 := impl.ciWorkflowRepository.UpdateWorkFlow(savedWf)
			//if err1 != nil {
			//	impl.logger.Errorw("could not save workflow, after failing due to conflicting image tag")
			//}
			return err
		}

		imageReservationIds = append(imageReservationIds, request.ImagePathReservationIds...)
	}
	request.ImagePathReservationIds = imageReservationIds
	request.DockerImageTag = dockerImageTag
	return nil
}

func (impl *CiTriggerServiceImpl) loadCopyContainerImagePluginMetadata(customTag string, customTagId int, buildImagePath string, buildImageDockerRegistryId string, workflowRequest *wfTypes.WorkflowTriggerRequest) error {
	preCiSteps := workflowRequest.PreCiSteps
	postCiSteps := workflowRequest.PostCiSteps
	var registryDestinationImageMap map[string][]string
	var registryCredentialMap map[string]plugin.RegistryCredentials
	var pluginArtifactStage string
	var imagePathReservationIds []int
	copyContainerImagePluginId, err := impl.globalPluginService.GetRefPluginIdByRefPluginName(types.COPY_CONTAINER_IMAGE)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting copyContainerImage plugin id", "err", err)
		return err
	}
	for _, step := range preCiSteps {
		if copyContainerImagePluginId != 0 && step.RefPluginId == copyContainerImagePluginId {
			// for copyContainerImage plugin parse destination images and save its data in image path reservation table
			return errors.New("copyContainerImage plugin not allowed in pre-ci step, please remove it and try again")
		}
	}
	for _, step := range postCiSteps {
		if copyContainerImagePluginId != 0 && step.RefPluginId == copyContainerImagePluginId {
			// for copyContainerImage plugin parse destination images and save its data in image path reservation table
			registryDestinationImageMap, registryCredentialMap, err = impl.pluginInputVariableParser.HandleCopyContainerImagePluginInputVariables(step.InputVars, customTag, buildImagePath, buildImageDockerRegistryId)
			if err != nil {
				impl.logger.Errorw("error in parsing copyContainerImage input variable", "err", err)
				return err
			}
			pluginArtifactStage = repository5.POST_CI
		}
	}
	for _, images := range registryDestinationImageMap {
		for _, image := range images {
			if image == buildImagePath {
				return bean2.ErrImagePathInUse
			}
		}
	}
	workflowRequest.RegistryDestinationImageMap = registryDestinationImageMap
	workflowRequest.RegistryCredentialMap = registryCredentialMap
	workflowRequest.PluginArtifactStage = pluginArtifactStage
	if len(registryDestinationImageMap) > 0 {
		workflowRequest.PushImageBeforePostCI = true
	}
	imagePathReservationIds, err = impl.ReserveImagesGeneratedAtPlugin(customTagId, registryDestinationImageMap)
	workflowRequest.ImagePathReservationIds = imagePathReservationIds
	return err
}

func (impl *CiTriggerServiceImpl) ReserveImagesGeneratedAtPlugin(customTagId int, registryImageMap map[string][]string) ([]int, error) {
	var imagePathReservationIds []int
	for _, images := range registryImageMap {
		for _, image := range images {
			imagePathReservationData, err := impl.customTagService.ReserveImagePath(image, customTagId)
			if err != nil {
				impl.logger.Errorw("Error in marking custom tag reserved", "err", err)
				return imagePathReservationIds, err
			}
			imagePathReservationIds = append(imagePathReservationIds, imagePathReservationData.Id)
		}
	}
	return imagePathReservationIds, nil
}

func (impl *CiTriggerServiceImpl) buildImageTag(commitHashes map[int]pipelineConfig.GitCommit, id int, wfId int) string {
	dockerImageTag := ""
	toAppendDevtronParamInTag := true
	for _, v := range commitHashes {
		if v.WebhookData.Id == 0 {
			if v.Commit == "" {
				continue
			}
			dockerImageTag = getUpdatedDockerImageTagWithCommitOrCheckOutData(dockerImageTag, getTruncatedImageTag(v.Commit))
		} else {
			_targetCheckout := v.WebhookData.Data[bean.WEBHOOK_SELECTOR_TARGET_CHECKOUT_NAME]
			if _targetCheckout == "" {
				continue
			}
			// if not PR based then meaning tag based
			isPRBasedEvent := v.WebhookData.EventActionType == bean.WEBHOOK_EVENT_MERGED_ACTION_TYPE
			if !isPRBasedEvent && impl.ciCdConfig.UseImageTagFromGitProviderForTagBasedBuild {
				dockerImageTag = getUpdatedDockerImageTagWithCommitOrCheckOutData(dockerImageTag, _targetCheckout)
			} else {
				dockerImageTag = getUpdatedDockerImageTagWithCommitOrCheckOutData(dockerImageTag, getTruncatedImageTag(_targetCheckout))
			}
			if isPRBasedEvent {
				_sourceCheckout := v.WebhookData.Data[bean.WEBHOOK_SELECTOR_SOURCE_CHECKOUT_NAME]
				dockerImageTag = getUpdatedDockerImageTagWithCommitOrCheckOutData(dockerImageTag, getTruncatedImageTag(_sourceCheckout))
			} else {
				toAppendDevtronParamInTag = !impl.ciCdConfig.UseImageTagFromGitProviderForTagBasedBuild
			}
		}
	}
	toAppendDevtronParamInTag = toAppendDevtronParamInTag && dockerImageTag != ""
	if toAppendDevtronParamInTag {
		dockerImageTag = fmt.Sprintf("%s-%d-%d", dockerImageTag, id, wfId)
	}
	// replace / with underscore, as docker image tag doesn't support slash. it gives error
	dockerImageTag = strings.ReplaceAll(dockerImageTag, "/", "_")
	return dockerImageTag
}

func (impl *CiTriggerServiceImpl) saveNewWorkflow(triggerRequest types.CiTriggerRequest, workflowRequest *wfTypes.WorkflowTriggerRequest) (*pipelineConfig.CiWorkflow, error) {
	pipelineId := workflowRequest.CdPipelineId
	ciWorkflow := &pipelineConfig.CiWorkflow{
		Name:                  workflowRequest.PipelineName + "-" + strconv.Itoa(pipelineId),
		Status:                pipelineConfig.WorkflowStarting, //starting CIStage
		Message:               "",
		StartedOn:             time.Now(),
		CiPipelineId:          pipelineId,
		Namespace:             impl.ciCdConfig.GetDefaultNamespace(),
		BlobStorageEnabled:    impl.ciCdConfig.BlobStorageEnabled,
		GitTriggers:           triggerRequest.CommitHashes,
		LogLocation:           "",
		TriggeredBy:           workflowRequest.TriggeredBy,
		ReferenceCiWorkflowId: triggerRequest.ReferenceCiWorkflowId,
		ExecutorType:          impl.ciCdConfig.GetWorkflowExecutorType(),
	}
	if workflowRequest.IsJob() {
		ciWorkflow.Namespace = workflowRequest.Namespace
		ciWorkflow.EnvironmentId = workflowRequest.EnvironmentId
	}

	err := impl.ciWorkflowRunnerRepository.SaveWorkFlow(ciWorkflow)
	if err != nil {
		impl.logger.Errorw("error occurred while saving ci workflow", "wfName", ciWorkflow.Name, "err", err)
		return nil, err
	}
	impl.logger.Debug("ci workflow saved", "wfName", ciWorkflow.Name)
	return ciWorkflow, nil
}

func (impl *CiTriggerServiceImpl) loadTriggerMetadata(triggerRequest types.CiTriggerRequest, request *wfTypes.WorkflowTriggerRequest) error {
	triggerMetadata := &types.PipelineTriggerMetadata{}
	pipelineId := triggerRequest.PipelineId
	ciPipelineMetadata, err := impl.ciPipelineConfigService.GetCiPipelineMetadata(pipelineId)
	if err != nil {
		return err
	}
	appId := ciPipelineMetadata.AppId
	triggerMetadata.AppId = appId
	appBean, err := impl.appConfigService.FindById(appId)
	if err != nil {
		return err
	}
	triggerMetadata.AppName = appBean.Name
	appType := appBean.AppType
	triggerMetadata.AppType = appType

	appLabels, err := impl.appCrudOperationService.GetLabelsByAppId(appId)
	if err != nil {
		return err
	}
	triggerMetadata.AppLabels = appLabels
	if appType == helper.Job {
		environmentBean, err := impl.environmentService.FindById(triggerRequest.EnvironmentId)
		if err != nil {
			return err
		}
		triggerMetadata.EnvId = environmentBean.Id
		triggerMetadata.EnvName = environmentBean.Environment
		triggerMetadata.ClusterId = environmentBean.ClusterId
		triggerMetadata.ClusterName = environmentBean.ClusterName
	}
	//request.Env = env
	request.LoadTriggerMetadata(triggerMetadata)
	return nil
}

func getUpdatedDockerImageTagWithCommitOrCheckOutData(dockerImageTag, commitOrCheckOutData string) string {
	if dockerImageTag == "" {
		dockerImageTag = commitOrCheckOutData
	} else {
		if commitOrCheckOutData != "" {
			dockerImageTag = fmt.Sprintf("%s-%s", dockerImageTag, commitOrCheckOutData)
		}
	}
	return dockerImageTag
}

func getTruncatedImageTag(imageTag string) string {
	_length := len(imageTag)
	if _length == 0 {
		return imageTag
	}

	_truncatedLength := 8

	if _length < _truncatedLength {
		return imageTag
	} else {
		return imageTag[:_truncatedLength]
	}
}
