package service

import (
	"fmt"
	repository5 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/plugin"
	"github.com/devtron-labs/devtron/pkg/workflow/pipeline/ci/materials/service"
	materialTypes "github.com/devtron-labs/devtron/pkg/workflow/pipeline/ci/materials/types"
	"github.com/devtron-labs/devtron/pkg/workflow/trigger/ci/types"
	wfTypes "github.com/devtron-labs/devtron/pkg/workflow/trigger/core/types"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

type CiTriggerService interface {
	TriggerPipeline(ctx bean.RequestContext, trigger types.CiTriggerRequest) (int, error)
}

type CiTriggerServiceImpl struct {
	logger                    *zap.SugaredLogger
	materialService           service.MaterialService
	ciPipelineConfigService   pipeline.CiPipelineConfigService
	pipelineStageService      pipeline.PipelineStageService
	ciCdConfig                wfTypes.CiCdConfig
	customTagService          pipeline.CustomTagService
	ciTemplateService         pipeline.CiTemplateService
	globalPluginService       plugin.GlobalPluginService
	pluginInputVariableParser pipeline.PluginInputVariableParser
}

func NewCiTriggerServiceImpl(logger *zap.SugaredLogger, service service.MaterialService, ciPipelineConfigService pipeline.CiPipelineConfigService,
	pipelineStageService pipeline.PipelineStageService, customTagService pipeline.CustomTagService, ciTemplateService pipeline.CiTemplateService,
	globalPluginService plugin.GlobalPluginService, pluginInputVariableParser pipeline.PluginInputVariableParser) *CiTriggerServiceImpl {
	return &CiTriggerServiceImpl{
		logger:                    logger,
		materialService:           service,
		ciPipelineConfigService:   ciPipelineConfigService,
		pipelineStageService:      pipelineStageService,
		customTagService:          customTagService,
		ciTemplateService:         ciTemplateService,
		globalPluginService:       globalPluginService,
		pluginInputVariableParser: pluginInputVariableParser,
	}
}

// LoadContext(USER, PIPELINE, APP, ENV)
func (impl *CiTriggerServiceImpl) TriggerPipeline(ctx bean.RequestContext, triggerRequest types.CiTriggerRequest) (int, error) {
	ciPipelineId := triggerRequest.PipelineId
	impl.logger.Infow("ci trigger request came", "pipelineId", ciPipelineId)

	workflowRequest := &wfTypes.WorkflowTriggerRequest{}
	workflowRequest.LoadFromUserContext(ctx.GetUser())

	ciPipelineMetadata, err := impl.ciPipelineConfigService.GetCiPipelineMetadata(ciPipelineId)
	if err != nil {
		return 0, err
	}

	// fetch ci materials
	err = impl.loadCiMaterials(triggerRequest, workflowRequest)
	if err != nil {
		return 0, err
	}

	// TODO: save ci workflow here, since its unique id is needed in workflowConfig

	err = impl.loadEnvAndWorkflowConfig(savedWfId, ciPipelineId, workflowRequest)
	if err != nil {
		return 0, err
	}

	// Load App and Env metadata and initialize workflowRequest from it

	err = impl.loadPipelineScripts(ciPipelineId, workflowRequest)
	if err != nil {
		return 0, err
	}

	err = impl.loadDockerData(savedWfId, triggerRequest.CommitHashes, workflowRequest)
	if errors.Is(err, bean2.ErrImagePathInUse) {
		savedWf.Status = pipelineConfig.WorkflowFailed
		savedWf.Message = bean2.ImageTagUnavailableMessage
		err1 := impl.ciWorkflowRepository.UpdateWorkFlow(savedWf) // directly create wf
		if err1 != nil {
			impl.logger.Errorw("could not save workflow, after failing due to conflicting image tag")
		}
		return 0, err
	}

	return savedWfId, nil

}

func (impl *CiTriggerServiceImpl) loadEnvAndWorkflowConfig(uniqueId int, ciPipelineId int, workflowRequest *wfTypes.WorkflowTriggerRequest) error {
	// TODO KB: load ci workflow config
	ciWorkflowConfig, err := impl.ciWorkflowRepository.FindConfigByPipelineId(ciPipelineId)
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
	prePostAndRefPluginResponse, err := impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(ciPipelineId, bean2.CiStage, scope)
	if err != nil {
		impl.logger.Errorw("error in getting pre steps data for wf request", "err", err, "ciPipelineId", ciPipelineId)
		return err
	}

	//fetch pipeline scripts
	beforeScripts, afterScripts, err := impl.ciPipelineConfigService.FindCiScriptsByCiPipelineId(ciPipelineId)
	if err != nil {
		return nil
	}
	workflowRequest.UpdatePipelineScripts(prePostAndRefPluginResponse, beforeScripts, afterScripts)
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

	var imageReservationIds []int
	var dockerImageTag string

	ciPipelineId := request.PipelineId
	ciTemplateMetadata, err := impl.ciPipelineConfigService.GetCiTemplateMetadata(request.AppId, ciPipelineId)
	dockerRegistry := ciTemplateMetadata.DockerRegistry
	dockerRepository := ciTemplateMetadata.DockerRepository
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
	var _imageReservationIds []int
	if !isJob {
		buildImagePath := fmt.Sprintf(bean2.ImagePathPattern, dockerRegistry.RegistryURL, dockerRepository, dockerImageTag)
		err = impl.loadCopyContainerImagePluginMetadata(dockerImageTag, customTag.Id, buildImagePath, dockerRegistry.Id, request)
		if err != nil {
			//TODO KB: save it in db
			//impl.logger.Errorw("error in getting env variables for copyContainerImage plugin")
			//savedWf.Status = pipelineConfig.WorkflowFailed
			//savedWf.Message = err.Error()
			//err1 := impl.ciWorkflowRepository.UpdateWorkFlow(savedWf)
			//if err1 != nil {
			//	impl.logger.Errorw("could not save workflow, after failing due to conflicting image tag")
			//}
			return err
		}

		imageReservationIds = append(imageReservationIds, _imageReservationIds...)
	}
	request.CheckoutPath = ciTemplateMetadata.CheckoutPath
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
