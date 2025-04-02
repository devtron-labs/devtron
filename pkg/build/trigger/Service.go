package trigger

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"github.com/devtron-labs/common-lib/utils"
	bean4 "github.com/devtron-labs/common-lib/utils/bean"
	commonBean "github.com/devtron-labs/common-lib/workflow"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/constants"
	repository5 "github.com/devtron-labs/devtron/internal/sql/repository"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/attributes"
	bean3 "github.com/devtron-labs/devtron/pkg/attributes/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/bean/common"
	"github.com/devtron-labs/devtron/pkg/build/pipeline"
	buildBean "github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	buildCommonBean "github.com/devtron-labs/devtron/pkg/build/pipeline/bean/common"
	repository6 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	pipeline2 "github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/adapter"
	pipelineConfigBean "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/plugin"
	bean2 "github.com/devtron-labs/devtron/pkg/plugin/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/variables"
	repository4 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/devtron-labs/devtron/util/sliceUtil"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Service interface {
	TriggerCiPipeline(trigger types.Trigger) (int, error)
}

type BuildxCacheFlags struct {
	BuildxCacheModeMin     bool `env:"BUILDX_CACHE_MODE_MIN" envDefault:"false"`
	AsyncBuildxCacheExport bool `env:"ASYNC_BUILDX_CACHE_EXPORT" envDefault:"false"`
}

type ServiceImpl struct {
	Logger                       *zap.SugaredLogger
	workflowService              pipeline2.WorkflowService
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
	ciArtifactRepository         repository5.CiArtifactRepository
	pipelineStageService         pipeline2.PipelineStageService
	userService                  user.UserService
	ciTemplateService            pipeline.CiTemplateReadService
	appCrudOperationService      app.AppCrudOperationService
	envRepository                repository6.EnvironmentRepository
	appRepository                appRepository.AppRepository
	customTagService             pipeline2.CustomTagService
	config                       *types.CiConfig
	scopedVariableManager        variables.ScopedVariableManager
	ciCdPipelineOrchestrator     pipeline2.CiCdPipelineOrchestrator
	buildxCacheFlags             *BuildxCacheFlags
	attributeService             attributes.AttributesService
	pluginInputVariableParser    pipeline2.PluginInputVariableParser
	globalPluginService          plugin.GlobalPluginService
	ciService                    pipeline2.CiService
}

func NewServiceImpl(Logger *zap.SugaredLogger, workflowService pipeline2.WorkflowService,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	ciArtifactRepository repository5.CiArtifactRepository,
	pipelineStageService pipeline2.PipelineStageService,
	userService user.UserService,
	ciTemplateService pipeline.CiTemplateReadService,
	appCrudOperationService app.AppCrudOperationService,
	envRepository repository6.EnvironmentRepository,
	appRepository appRepository.AppRepository,
	scopedVariableManager variables.ScopedVariableManager,
	customTagService pipeline2.CustomTagService,
	ciCdPipelineOrchestrator pipeline2.CiCdPipelineOrchestrator, attributeService attributes.AttributesService,
	pluginInputVariableParser pipeline2.PluginInputVariableParser,
	globalPluginService plugin.GlobalPluginService,
	ciService pipeline2.CiService,
) *ServiceImpl {
	buildxCacheFlags := &BuildxCacheFlags{}
	err := env.Parse(buildxCacheFlags)
	if err != nil {
		Logger.Infow("error occurred while parsing BuildxCacheFlags env,so setting BuildxCacheModeMin and AsyncBuildxCacheExport to default value", "err", err)
	}
	cis := &ServiceImpl{
		Logger:                       Logger,
		workflowService:              workflowService,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		ciPipelineRepository:         ciPipelineRepository,
		ciArtifactRepository:         ciArtifactRepository,
		pipelineStageService:         pipelineStageService,
		userService:                  userService,
		ciTemplateService:            ciTemplateService,
		appCrudOperationService:      appCrudOperationService,
		envRepository:                envRepository,
		appRepository:                appRepository,
		scopedVariableManager:        scopedVariableManager,
		customTagService:             customTagService,
		ciCdPipelineOrchestrator:     ciCdPipelineOrchestrator,
		buildxCacheFlags:             buildxCacheFlags,
		attributeService:             attributeService,
		pluginInputVariableParser:    pluginInputVariableParser,
		globalPluginService:          globalPluginService,
		ciService:                    ciService,
	}
	config, err := types.GetCiConfig()
	if err != nil {
		return nil
	}
	cis.config = config
	return cis
}

func (impl *ServiceImpl) TriggerCiPipeline(trigger types.Trigger) (int, error) {
	impl.Logger.Debug("ci pipeline manual trigger")
	ciMaterials, err := impl.GetCiMaterials(trigger.PipelineId, trigger.CiMaterials)
	if err != nil {
		return 0, err
	}

	ciPipelineScripts, err := impl.ciPipelineRepository.FindCiScriptsByCiPipelineId(trigger.PipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		return 0, err
	}

	var pipeline *pipelineConfig.CiPipeline
	for _, m := range ciMaterials {
		pipeline = m.CiPipeline
		break
	}

	scope := resourceQualifiers.Scope{
		AppId: pipeline.App.Id,
	}
	ciWorkflowConfigNamespace := impl.config.GetDefaultNamespace()
	envModal, isJob, err := impl.getEnvironmentForJob(pipeline, trigger)
	if err != nil {
		return 0, err
	}
	if isJob && envModal != nil {
		ciWorkflowConfigNamespace = envModal.Namespace

		// This will be populated for jobs running in selected environment
		scope.EnvId = envModal.Id
		scope.ClusterId = envModal.ClusterId

		scope.SystemMetadata = &resourceQualifiers.SystemMetadata{
			EnvironmentName: envModal.Name,
			ClusterName:     envModal.Cluster.ClusterName,
			Namespace:       envModal.Namespace,
		}
	}
	if scope.SystemMetadata == nil {
		scope.SystemMetadata = &resourceQualifiers.SystemMetadata{
			Namespace: ciWorkflowConfigNamespace,
			AppName:   pipeline.App.AppName,
		}
	}

	savedCiWf, err := impl.saveNewWorkflow(pipeline, ciWorkflowConfigNamespace, trigger.CommitHashes, trigger.TriggeredBy, trigger.EnvironmentId, isJob, trigger.ReferenceCiWorkflowId)
	if err != nil {
		impl.Logger.Errorw("could not save new workflow", "err", err)
		return 0, err
	}

	// preCiSteps, postCiSteps, refPluginsData, err := impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(pipeline.Id, ciEvent)
	request := pipelineConfigBean.NewBuildPrePostStepDataReq(pipeline.Id, pipelineConfigBean.CiStage, scope)
	prePostAndRefPluginResponse, err := impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(request)
	if err != nil {
		impl.Logger.Errorw("error in getting pre steps data for wf request", "err", err, "ciPipelineId", pipeline.Id)
		dbErr := impl.markCurrentCiWorkflowFailed(savedCiWf, err)
		if dbErr != nil {
			impl.Logger.Errorw("saving workflow error", "err", dbErr)
		}
		return 0, err
	}
	preCiSteps := prePostAndRefPluginResponse.PreStageSteps
	postCiSteps := prePostAndRefPluginResponse.PostStageSteps
	refPluginsData := prePostAndRefPluginResponse.RefPluginData
	variableSnapshot := prePostAndRefPluginResponse.VariableSnapshot

	if len(preCiSteps) == 0 && isJob {
		errMsg := fmt.Sprintf("No tasks are configured in this job pipeline")
		validationErr := util.NewApiError(http.StatusNotFound, errMsg, errMsg)

		return 0, validationErr
	}

	// get env variables of git trigger data and add it in the extraEnvVariables
	gitTriggerEnvVariables, _, err := impl.ciCdPipelineOrchestrator.GetGitCommitEnvVarDataForCICDStage(savedCiWf.GitTriggers)
	if err != nil {
		impl.Logger.Errorw("error in getting gitTrigger env data for stage", "gitTriggers", savedCiWf.GitTriggers, "err", err)
		return 0, err
	}

	for k, v := range gitTriggerEnvVariables {
		trigger.RuntimeParameters = trigger.RuntimeParameters.AddSystemVariable(k, v)
	}

	workflowRequest, err := impl.buildWfRequestForCiPipeline(pipeline, trigger, ciMaterials, savedCiWf, ciWorkflowConfigNamespace, ciPipelineScripts, preCiSteps, postCiSteps, refPluginsData, isJob)
	if err != nil {
		impl.Logger.Errorw("make workflow req", "err", err)
		return 0, err
	}
	err = impl.handleRuntimeParamsValidations(trigger, ciMaterials, workflowRequest)
	if err != nil {
		savedCiWf.Status = cdWorkflow.WorkflowAborted
		savedCiWf.Message = err.Error()
		err1 := impl.ciService.UpdateCiWorkflowWithStage(savedCiWf)
		if err1 != nil {
			impl.Logger.Errorw("could not save workflow, after failing due to conflicting image tag")
		}
		return 0, err
	}

	workflowRequest.Scope = scope
	workflowRequest.AppId = pipeline.AppId
	workflowRequest.BuildxCacheModeMin = impl.buildxCacheFlags.BuildxCacheModeMin
	workflowRequest.AsyncBuildxCacheExport = impl.buildxCacheFlags.AsyncBuildxCacheExport
	if impl.config != nil && impl.config.BuildxK8sDriverOptions != "" {
		err = impl.setBuildxK8sDriverData(workflowRequest)
		if err != nil {
			impl.Logger.Errorw("error in setBuildxK8sDriverData", "BUILDX_K8S_DRIVER_OPTIONS", impl.config.BuildxK8sDriverOptions, "err", err)
			return 0, err
		}
	}

	// savedCiWf.LogLocation = impl.ciCdConfig.CiDefaultBuildLogsKeyPrefix + "/" + workflowRequest.WorkflowNamePrefix + "/main.log"
	savedCiWf.LogLocation = fmt.Sprintf("%s/%s/main.log", impl.config.GetDefaultBuildLogsKeyPrefix(), workflowRequest.WorkflowNamePrefix)
	err = impl.updateCiWorkflow(workflowRequest, savedCiWf)

	appLabels, err := impl.appCrudOperationService.GetLabelsByAppId(pipeline.AppId)
	if err != nil {
		return 0, err
	}
	workflowRequest.AppLabels = appLabels
	workflowRequest.Env = envModal
	if isJob {
		workflowRequest.Type = pipelineConfigBean.JOB_WORKFLOW_PIPELINE_TYPE
	} else {
		workflowRequest.Type = pipelineConfigBean.CI_WORKFLOW_PIPELINE_TYPE
	}

	workflowRequest.CiPipelineType = trigger.PipelineType
	err = impl.executeCiPipeline(workflowRequest)
	if err != nil {
		impl.Logger.Errorw("error in executing ci pipeline", "err", err)
		dbErr := impl.markCurrentCiWorkflowFailed(savedCiWf, err)
		if dbErr != nil {
			impl.Logger.Errorw("update ci workflow error", "err", dbErr)
		}
		return 0, err
	}
	impl.Logger.Debugw("ci triggered", " pipeline ", trigger.PipelineId)

	var variableSnapshotHistories = sliceUtil.GetBeansPtr(
		repository4.GetSnapshotBean(savedCiWf.Id, repository4.HistoryReferenceTypeCIWORKFLOW, variableSnapshot))
	if len(variableSnapshotHistories) > 0 {
		err = impl.scopedVariableManager.SaveVariableHistoriesForTrigger(variableSnapshotHistories, trigger.TriggeredBy)
		if err != nil {
			impl.Logger.Errorf("Not able to save variable snapshot for CI trigger %s", err)
		}
	}

	middleware.CiTriggerCounter.WithLabelValues(pipeline.App.AppName, pipeline.Name).Inc()
	go impl.ciService.WriteCITriggerEvent(trigger, pipeline, workflowRequest)
	return savedCiWf.Id, err
}

func (impl *ServiceImpl) GetCiMaterials(pipelineId int, ciMaterials []*pipelineConfig.CiPipelineMaterial) ([]*pipelineConfig.CiPipelineMaterial, error) {
	if !(len(ciMaterials) == 0) {
		return ciMaterials, nil
	} else {
		ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(pipelineId)
		if err != nil {
			impl.Logger.Errorw("err", "err", err)
			return nil, err
		}
		impl.Logger.Debug("ciMaterials for pipeline trigger ", ciMaterials)
		return ciMaterials, nil
	}
}

func (impl *ServiceImpl) setBuildxK8sDriverData(workflowRequest *types.WorkflowRequest) error {
	ciBuildConfig := workflowRequest.CiBuildConfig
	if ciBuildConfig != nil {
		if dockerBuildConfig := ciBuildConfig.DockerBuildConfig; dockerBuildConfig != nil {
			k8sDriverOptions, err := impl.getK8sDriverOptions()
			if err != nil {
				errMsg := "error in parsing BUILDX_K8S_DRIVER_OPTIONS from the devtron-cm, "
				err = errors.New(errMsg + "error : " + err.Error())
				impl.Logger.Errorw(errMsg, "err", err)
			}
			dockerBuildConfig.BuildxK8sDriverOptions = k8sDriverOptions

		}
	}
	return nil
}

func (impl *ServiceImpl) getK8sDriverOptions() ([]map[string]string, error) {
	buildxK8sDriverOptions := make([]map[string]string, 0)
	err := json.Unmarshal([]byte(impl.config.BuildxK8sDriverOptions), &buildxK8sDriverOptions)
	if err != nil {
		return nil, err
	} else {
		return buildxK8sDriverOptions, nil
	}
}

func (impl *ServiceImpl) getEnvironmentForJob(pipeline *pipelineConfig.CiPipeline, trigger types.Trigger) (*repository6.Environment, bool, error) {
	app, err := impl.appRepository.FindById(pipeline.AppId)
	if err != nil {
		impl.Logger.Errorw("could not find app", "err", err)
		return nil, false, err
	}

	var env *repository6.Environment
	isJob := false
	if app.AppType == helper.Job {
		isJob = true
		if trigger.EnvironmentId != 0 {
			env, err = impl.envRepository.FindById(trigger.EnvironmentId)
			if err != nil {
				impl.Logger.Errorw("could not find environment", "err", err)
				return nil, isJob, err
			}
			return env, isJob, nil
		}
	}
	return nil, isJob, nil
}

// TODO: Send all trigger data
func (impl *ServiceImpl) BuildPayload(trigger types.Trigger, pipeline *pipelineConfig.CiPipeline) *client.Payload {
	payload := &client.Payload{}
	payload.AppName = pipeline.App.AppName
	payload.PipelineName = pipeline.Name
	return payload
}

func (impl *ServiceImpl) saveNewWorkflow(pipeline *pipelineConfig.CiPipeline, ciWorkflowConfigNamespace string,
	commitHashes map[int]pipelineConfig.GitCommit, userId int32, EnvironmentId int, isJob bool, refCiWorkflowId int) (wf *pipelineConfig.CiWorkflow, error error) {

	ciWorkflow := &pipelineConfig.CiWorkflow{
		Name:                  pipeline.Name + "-" + strconv.Itoa(pipeline.Id),
		Status:                cdWorkflow.WorkflowStarting, // starting CIStage
		Message:               "",
		StartedOn:             time.Now(),
		CiPipelineId:          pipeline.Id,
		Namespace:             impl.config.GetDefaultNamespace(),
		BlobStorageEnabled:    impl.config.BlobStorageEnabled,
		GitTriggers:           commitHashes,
		LogLocation:           "",
		TriggeredBy:           userId,
		ReferenceCiWorkflowId: refCiWorkflowId,
		ExecutorType:          impl.config.GetWorkflowExecutorType(),
	}
	if isJob {
		ciWorkflow.Namespace = ciWorkflowConfigNamespace
		ciWorkflow.EnvironmentId = EnvironmentId
	}
	err := impl.ciService.SaveCiWorkflowWithStage(ciWorkflow)
	if err != nil {
		impl.Logger.Errorw("saving workflow error", "err", err)
		return &pipelineConfig.CiWorkflow{}, err
	}
	impl.Logger.Debugw("workflow saved ", "id", ciWorkflow.Id)
	return ciWorkflow, nil
}

func (impl *ServiceImpl) executeCiPipeline(workflowRequest *types.WorkflowRequest) error {
	_, _, err := impl.workflowService.SubmitWorkflow(workflowRequest)
	if err != nil {
		impl.Logger.Errorw("workflow error", "err", err)
		return err
	}
	return nil
}

func (impl *ServiceImpl) buildS3ArtifactLocation(ciWorkflowConfigLogsBucket string, savedWf *pipelineConfig.CiWorkflow) (string, string, string) {
	ciArtifactLocationFormat := impl.config.GetArtifactLocationFormat()
	ArtifactLocation := fmt.Sprintf("s3://"+path.Join(ciWorkflowConfigLogsBucket, ciArtifactLocationFormat), savedWf.Id, savedWf.Id)
	artifactFileName := fmt.Sprintf(ciArtifactLocationFormat, savedWf.Id, savedWf.Id)
	return ArtifactLocation, ciWorkflowConfigLogsBucket, artifactFileName
}

func (impl *ServiceImpl) buildDefaultArtifactLocation(savedWf *pipelineConfig.CiWorkflow) string {
	ciArtifactLocationFormat := impl.config.GetArtifactLocationFormat()
	ArtifactLocation := fmt.Sprintf(ciArtifactLocationFormat, savedWf.Id, savedWf.Id)
	return ArtifactLocation
}

func (impl *ServiceImpl) buildWfRequestForCiPipeline(pipeline *pipelineConfig.CiPipeline, trigger types.Trigger, ciMaterials []*pipelineConfig.CiPipelineMaterial, savedWf *pipelineConfig.CiWorkflow, ciWorkflowConfigNamespace string, ciPipelineScripts []*pipelineConfig.CiPipelineScript, preCiSteps []*pipelineConfigBean.StepObject, postCiSteps []*pipelineConfigBean.StepObject, refPluginsData []*pipelineConfigBean.RefPluginObject, isJob bool) (*types.WorkflowRequest, error) {
	var ciProjectDetails []pipelineConfigBean.CiProjectDetails
	commitHashes := trigger.CommitHashes
	for _, ciMaterial := range ciMaterials {
		// ignore those materials which have inactive git material
		if ciMaterial == nil || ciMaterial.GitMaterial == nil || !ciMaterial.GitMaterial.Active {
			continue
		}
		commitHashForPipelineId := commitHashes[ciMaterial.Id]
		ciProjectDetail := pipelineConfigBean.CiProjectDetails{
			GitRepository:   ciMaterial.GitMaterial.Url,
			MaterialName:    ciMaterial.GitMaterial.Name,
			CheckoutPath:    ciMaterial.GitMaterial.CheckoutPath,
			FetchSubmodules: ciMaterial.GitMaterial.FetchSubmodules,
			CommitHash:      commitHashForPipelineId.Commit,
			Author:          commitHashForPipelineId.Author,
			SourceType:      ciMaterial.Type,
			SourceValue:     ciMaterial.Value,
			GitTag:          ciMaterial.GitTag,
			Message:         commitHashForPipelineId.Message,
			Type:            string(ciMaterial.Type),
			CommitTime:      commitHashForPipelineId.Date.Format(bean.LayoutRFC3339),
			GitOptions: pipelineConfigBean.GitOptions{
				UserName:              ciMaterial.GitMaterial.GitProvider.UserName,
				Password:              ciMaterial.GitMaterial.GitProvider.Password,
				SshPrivateKey:         ciMaterial.GitMaterial.GitProvider.SshPrivateKey,
				AccessToken:           ciMaterial.GitMaterial.GitProvider.AccessToken,
				AuthMode:              ciMaterial.GitMaterial.GitProvider.AuthMode,
				EnableTLSVerification: ciMaterial.GitMaterial.GitProvider.EnableTLSVerification,
				TlsKey:                ciMaterial.GitMaterial.GitProvider.TlsKey,
				TlsCert:               ciMaterial.GitMaterial.GitProvider.TlsCert,
				CaCert:                ciMaterial.GitMaterial.GitProvider.CaCert,
			},
		}

		if ciMaterial.Type == constants.SOURCE_TYPE_WEBHOOK {
			webhookData := commitHashForPipelineId.WebhookData
			ciProjectDetail.WebhookData = pipelineConfig.WebhookData{
				Id:              webhookData.Id,
				EventActionType: webhookData.EventActionType,
				Data:            webhookData.Data,
			}
		}

		ciProjectDetails = append(ciProjectDetails, ciProjectDetail)
	}

	var beforeDockerBuildScripts []*bean.CiScript
	var afterDockerBuildScripts []*bean.CiScript
	for _, ciPipelineScript := range ciPipelineScripts {
		ciTask := &bean.CiScript{
			Id:             ciPipelineScript.Id,
			Index:          ciPipelineScript.Index,
			Name:           ciPipelineScript.Name,
			Script:         ciPipelineScript.Script,
			OutputLocation: ciPipelineScript.OutputLocation,
		}

		if ciPipelineScript.Stage == buildCommonBean.BEFORE_DOCKER_BUILD {
			beforeDockerBuildScripts = append(beforeDockerBuildScripts, ciTask)
		} else if ciPipelineScript.Stage == buildCommonBean.AFTER_DOCKER_BUILD {
			afterDockerBuildScripts = append(afterDockerBuildScripts, ciTask)
		}
	}

	if !(len(beforeDockerBuildScripts) == 0 && len(afterDockerBuildScripts) == 0) {
		// found beforeDockerBuildScripts/afterDockerBuildScripts
		// building preCiSteps & postCiSteps from them, refPluginsData not needed
		preCiSteps = buildCiStepsDataFromDockerBuildScripts(beforeDockerBuildScripts)
		postCiSteps = buildCiStepsDataFromDockerBuildScripts(afterDockerBuildScripts)
		refPluginsData = []*pipelineConfigBean.RefPluginObject{}
	}

	host, err := impl.attributeService.GetByKey(bean3.HostUrlKey)
	if err != nil {
		impl.Logger.Errorw("error in getting host url", "err", err, "hostUrl", host.Value)
		return nil, err
	}
	ciWorkflowConfigCiCacheBucket := impl.config.DefaultCacheBucket

	ciWorkflowConfigCiCacheRegion := impl.config.DefaultCacheBucketRegion

	ciWorkflowConfigCiImage := impl.config.GetDefaultImage()

	ciTemplate := pipeline.CiTemplate
	ciLevelArgs := pipeline.DockerArgs

	if ciLevelArgs == "" {
		ciLevelArgs = "{}"
	}

	if pipeline.CiTemplate.DockerBuildOptions == "" {
		pipeline.CiTemplate.DockerBuildOptions = "{}"
	}
	userEmailId, err := impl.userService.GetActiveEmailById(trigger.TriggeredBy)
	if err != nil {
		impl.Logger.Errorw("unable to find user email by id", "err", err, "id", trigger.TriggeredBy)
		return nil, err
	}
	var dockerfilePath string
	var dockerRepository string
	var checkoutPath string
	var ciBuildConfigBean *buildBean.CiBuildConfigBean
	dockerRegistry := &repository3.DockerArtifactStore{}
	ciBaseBuildConfigEntity := ciTemplate.CiBuildConfig
	ciBaseBuildConfigBean, err := adapter.ConvertDbBuildConfigToBean(ciBaseBuildConfigEntity)
	if err != nil {
		impl.Logger.Errorw("error occurred while converting buildconfig dbEntity to configBean", "ciBuildConfigEntity", ciBaseBuildConfigEntity, "err", err)
		return nil, errors.New("error while parsing ci build config")
	}
	if !pipeline.IsExternal && pipeline.IsDockerConfigOverridden {
		templateOverrideBean, err := impl.ciTemplateService.FindTemplateOverrideByCiPipelineId(pipeline.Id)
		if err != nil {
			return nil, err
		}
		ciBuildConfigBean = templateOverrideBean.CiBuildConfig
		// updating args coming from ciBaseBuildConfigEntity because it is not part of Ci override
		if ciBuildConfigBean != nil && ciBuildConfigBean.DockerBuildConfig != nil && ciBaseBuildConfigBean != nil && ciBaseBuildConfigBean.DockerBuildConfig != nil {
			ciBuildConfigBean.DockerBuildConfig.Args = ciBaseBuildConfigBean.DockerBuildConfig.Args
		}
		templateOverride := templateOverrideBean.CiTemplateOverride
		checkoutPath = templateOverride.GitMaterial.CheckoutPath
		dockerfilePath = templateOverride.DockerfilePath
		dockerRepository = templateOverride.DockerRepository
		dockerRegistry = templateOverride.DockerRegistry
	} else {
		checkoutPath = ciTemplate.GitMaterial.CheckoutPath
		dockerfilePath = ciTemplate.DockerfilePath
		dockerRegistry = ciTemplate.DockerRegistry
		dockerRepository = ciTemplate.DockerRepository
		ciBuildConfigBean = ciBaseBuildConfigBean
		if ciBuildConfigBean != nil {
			ciBuildConfigBean.BuildContextGitMaterialId = ciTemplate.BuildContextGitMaterialId
		}

	}
	if checkoutPath == "" {
		checkoutPath = "./"
	}
	var dockerImageTag string
	customTag, err := impl.customTagService.GetActiveCustomTagByEntityKeyAndValue(pipelineConfigBean.EntityTypeCiPipelineId, strconv.Itoa(pipeline.Id))
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	if customTag.Id != 0 && customTag.Enabled == true {
		imagePathReservation, err := impl.customTagService.GenerateImagePath(pipelineConfigBean.EntityTypeCiPipelineId, strconv.Itoa(pipeline.Id), dockerRegistry.RegistryURL, dockerRepository)
		if err != nil {
			if errors.Is(err, pipelineConfigBean.ErrImagePathInUse) {
				errMsg := pipelineConfigBean.ImageTagUnavailableMessage
				validationErr := util.NewApiError(http.StatusConflict, errMsg, errMsg)
				dbErr := impl.markCurrentCiWorkflowFailed(savedWf, validationErr)
				if dbErr != nil {
					impl.Logger.Errorw("could not save workflow, after failing due to conflicting image tag", "err", dbErr, "savedWf", savedWf.Id)
				}
				return nil, err
			}
			return nil, err
		}
		savedWf.ImagePathReservationIds = []int{imagePathReservation.Id}
		// imagePath = docker.io/avd0/dashboard:fd23414b
		imagePathSplit := strings.Split(imagePathReservation.ImagePath, ":")
		if len(imagePathSplit) >= 1 {
			dockerImageTag = imagePathSplit[len(imagePathSplit)-1]
		}
	} else {
		dockerImageTag = impl.buildImageTag(commitHashes, pipeline.Id, savedWf.Id)
	}

	// copyContainerImage plugin specific logic
	var registryCredentialMap map[string]bean2.RegistryCredentials
	var pluginArtifactStage string
	var imageReservationIds []int
	var registryDestinationImageMap map[string][]string
	if !isJob {
		registryDestinationImageMap, registryCredentialMap, pluginArtifactStage, imageReservationIds, err = impl.GetWorkflowRequestVariablesForCopyContainerImagePlugin(preCiSteps, postCiSteps, dockerImageTag, customTag.Id,
			fmt.Sprintf(pipelineConfigBean.ImagePathPattern,
				dockerRegistry.RegistryURL,
				dockerRepository,
				dockerImageTag),
			dockerRegistry.Id)
		if err != nil {
			impl.Logger.Errorw("error in getting env variables for copyContainerImage plugin")
			dbErr := impl.markCurrentCiWorkflowFailed(savedWf, err)
			if dbErr != nil {
				impl.Logger.Errorw("could not save workflow, after failing due to conflicting image tag", "err", dbErr, "savedWf", savedWf.Id)
			}
			return nil, err
		}

		savedWf.ImagePathReservationIds = append(savedWf.ImagePathReservationIds, imageReservationIds...)
	}
	// mergedArgs := string(merged)
	oldArgs := ciTemplate.Args
	ciBuildConfigBean, err = adapter.OverrideCiBuildConfig(dockerfilePath, oldArgs, ciLevelArgs, ciTemplate.DockerBuildOptions, ciTemplate.TargetPlatform, ciBuildConfigBean)
	if err != nil {
		impl.Logger.Errorw("error occurred while overriding ci build config", "oldArgs", oldArgs, "ciLevelArgs", ciLevelArgs, "error", err)
		return nil, errors.New("error while parsing ci build config")
	}
	buildContextCheckoutPath, err := impl.ciPipelineMaterialRepository.GetCheckoutPath(ciBuildConfigBean.BuildContextGitMaterialId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error occurred while getting checkout path from git material", "gitMaterialId", ciBuildConfigBean.BuildContextGitMaterialId, "error", err)
		return nil, err
	}
	if buildContextCheckoutPath == "" {
		buildContextCheckoutPath = checkoutPath
	}
	if ciBuildConfigBean.UseRootBuildContext {
		// use root build context i.e '.'
		buildContextCheckoutPath = "."
	}

	ciBuildConfigBean.PipelineType = trigger.PipelineType

	if ciBuildConfigBean.CiBuildType == buildBean.SELF_DOCKERFILE_BUILD_TYPE || ciBuildConfigBean.CiBuildType == buildBean.MANAGED_DOCKERFILE_BUILD_TYPE {
		ciBuildConfigBean.DockerBuildConfig.BuildContext = filepath.Join(buildContextCheckoutPath, ciBuildConfigBean.DockerBuildConfig.BuildContext)
		dockerBuildConfig := ciBuildConfigBean.DockerBuildConfig
		dockerfilePath = filepath.Join(checkoutPath, dockerBuildConfig.DockerfilePath)
		dockerBuildConfig.DockerfilePath = dockerfilePath
		checkoutPath = dockerfilePath[:strings.LastIndex(dockerfilePath, "/")+1]
	} else if ciBuildConfigBean.CiBuildType == buildBean.BUILDPACK_BUILD_TYPE {
		buildPackConfig := ciBuildConfigBean.BuildPackConfig
		checkoutPath = filepath.Join(checkoutPath, buildPackConfig.ProjectPath)
	}

	defaultTargetPlatform := impl.config.DefaultTargetPlatform
	useBuildx := impl.config.UseBuildx

	if ciBuildConfigBean.DockerBuildConfig != nil {
		if ciBuildConfigBean.DockerBuildConfig.TargetPlatform == "" && useBuildx {
			ciBuildConfigBean.DockerBuildConfig.TargetPlatform = defaultTargetPlatform
			ciBuildConfigBean.DockerBuildConfig.UseBuildx = useBuildx
		}
		ciBuildConfigBean.DockerBuildConfig.BuildxProvenanceMode = impl.config.BuildxProvenanceMode
	}

	workflowRequest := &types.WorkflowRequest{
		WorkflowNamePrefix:          strconv.Itoa(savedWf.Id) + "-" + savedWf.Name,
		PipelineName:                pipeline.Name,
		PipelineId:                  pipeline.Id,
		CiCacheFileName:             pipeline.Name + "-" + strconv.Itoa(pipeline.Id) + ".tar.gz",
		CiProjectDetails:            ciProjectDetails,
		Namespace:                   ciWorkflowConfigNamespace,
		BlobStorageConfigured:       savedWf.BlobStorageEnabled,
		CiImage:                     ciWorkflowConfigCiImage,
		WorkflowId:                  savedWf.Id,
		TriggeredBy:                 savedWf.TriggeredBy,
		CacheLimit:                  impl.config.CacheLimit,
		ScanEnabled:                 pipeline.ScanEnabled,
		CloudProvider:               impl.config.CloudProvider,
		DefaultAddressPoolBaseCidr:  impl.config.GetDefaultAddressPoolBaseCidr(),
		DefaultAddressPoolSize:      impl.config.GetDefaultAddressPoolSize(),
		PreCiSteps:                  preCiSteps,
		PostCiSteps:                 postCiSteps,
		RefPlugins:                  refPluginsData,
		AppName:                     pipeline.App.AppName,
		TriggerByAuthor:             userEmailId,
		CiBuildConfig:               ciBuildConfigBean,
		CiBuildDockerMtuValue:       impl.config.CiRunnerDockerMTUValue,
		CacheInvalidate:             trigger.InvalidateCache,
		SystemEnvironmentVariables:  trigger.RuntimeParameters.GetSystemVariables(),
		EnableBuildContext:          impl.config.EnableBuildContext,
		OrchestratorHost:            impl.config.OrchestratorHost,
		HostUrl:                     host.Value,
		OrchestratorToken:           impl.config.OrchestratorToken,
		ImageRetryCount:             impl.config.ImageRetryCount,
		ImageRetryInterval:          impl.config.ImageRetryInterval,
		WorkflowExecutor:            impl.config.GetWorkflowExecutorType(),
		Type:                        pipelineConfigBean.CI_WORKFLOW_PIPELINE_TYPE,
		CiArtifactLastFetch:         trigger.CiArtifactLastFetch,
		RegistryDestinationImageMap: registryDestinationImageMap,
		RegistryCredentialMap:       registryCredentialMap,
		PluginArtifactStage:         pluginArtifactStage,
		ImageScanMaxRetries:         impl.config.ImageScanMaxRetries,
		ImageScanRetryDelay:         impl.config.ImageScanRetryDelay,
		UseDockerApiToGetDigest:     impl.config.UseDockerApiToGetDigest,
	}
	workflowRequest.SetAwsInspectorConfig("")
	//in oss, there is no pipeline level workflow cache config, so we pass inherit to get the app level config
	workflowCacheConfig := impl.ciCdPipelineOrchestrator.GetWorkflowCacheConfig(pipeline.App.AppType, trigger.PipelineType, common.WorkflowCacheConfigInherit)
	workflowRequest.IgnoreDockerCachePush = !workflowCacheConfig.Value
	workflowRequest.IgnoreDockerCachePull = !workflowCacheConfig.Value
	impl.Logger.Debugw("Ignore Cache values", "IgnoreDockerCachePush", workflowRequest.IgnoreDockerCachePush, "IgnoreDockerCachePull", workflowRequest.IgnoreDockerCachePull)
	if pipeline.App.AppType == helper.Job {
		workflowRequest.AppName = pipeline.App.DisplayName
	}
	if pipeline.ScanEnabled {
		scanToolMetadata, scanVia, err := impl.fetchImageScanExecutionMedium()
		if err != nil {
			impl.Logger.Errorw("error occurred getting scanned via", "err", err)
			return nil, err
		}
		workflowRequest.SetExecuteImageScanningVia(scanVia)
		if scanVia.IsScanMediumExternal() {
			imageScanExecutionSteps, refPlugins, err := impl.fetchImageScanExecutionStepsForWfRequest(scanToolMetadata)
			if err != nil {
				impl.Logger.Errorw("error occurred, fetchImageScanExecutionStepsForWfRequest", "scanToolMetadata", scanToolMetadata, "err", err)
				return nil, err
			}
			workflowRequest.SetImageScanningSteps(imageScanExecutionSteps)
			workflowRequest.RefPlugins = append(workflowRequest.RefPlugins, refPlugins...)
		}
	}
	if dockerRegistry != nil {
		workflowRequest.DockerRegistryId = dockerRegistry.Id
		workflowRequest.DockerRegistryType = string(dockerRegistry.RegistryType)
		workflowRequest.DockerImageTag = dockerImageTag
		workflowRequest.DockerRegistryURL = dockerRegistry.RegistryURL
		workflowRequest.DockerRepository = dockerRepository
		workflowRequest.CheckoutPath = checkoutPath
		workflowRequest.DockerUsername = dockerRegistry.Username
		workflowRequest.DockerPassword = dockerRegistry.Password
		workflowRequest.AwsRegion = dockerRegistry.AWSRegion
		workflowRequest.AccessKey = dockerRegistry.AWSAccessKeyId
		workflowRequest.SecretKey = dockerRegistry.AWSSecretAccessKey
		workflowRequest.DockerConnection = dockerRegistry.Connection
		workflowRequest.DockerCert = dockerRegistry.Cert

	}
	ciWorkflowConfigLogsBucket := impl.config.GetDefaultBuildLogsBucket()

	switch workflowRequest.CloudProvider {
	case types.BLOB_STORAGE_S3:
		// No AccessKey is used for uploading artifacts, instead IAM based auth is used
		workflowRequest.CiCacheRegion = ciWorkflowConfigCiCacheRegion
		workflowRequest.CiCacheLocation = ciWorkflowConfigCiCacheBucket
		workflowRequest.CiArtifactLocation, workflowRequest.CiArtifactBucket, workflowRequest.CiArtifactFileName = impl.buildS3ArtifactLocation(ciWorkflowConfigLogsBucket, savedWf)
		workflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			AccessKey:                  impl.config.BlobStorageS3AccessKey,
			Passkey:                    impl.config.BlobStorageS3SecretKey,
			EndpointUrl:                impl.config.BlobStorageS3Endpoint,
			IsInSecure:                 impl.config.BlobStorageS3EndpointInsecure,
			CiCacheBucketName:          ciWorkflowConfigCiCacheBucket,
			CiCacheRegion:              ciWorkflowConfigCiCacheRegion,
			CiCacheBucketVersioning:    impl.config.BlobStorageS3BucketVersioned,
			CiArtifactBucketName:       workflowRequest.CiArtifactBucket,
			CiArtifactRegion:           impl.config.GetDefaultCdLogsBucketRegion(),
			CiArtifactBucketVersioning: impl.config.BlobStorageS3BucketVersioned,
			CiLogBucketName:            impl.config.GetDefaultBuildLogsBucket(),
			CiLogRegion:                impl.config.GetDefaultCdLogsBucketRegion(),
			CiLogBucketVersioning:      impl.config.BlobStorageS3BucketVersioned,
		}
	case types.BLOB_STORAGE_GCP:
		workflowRequest.GcpBlobConfig = &blob_storage.GcpBlobConfig{
			CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
			CacheBucketName:        ciWorkflowConfigCiCacheBucket,
			LogBucketName:          ciWorkflowConfigLogsBucket,
			ArtifactBucketName:     ciWorkflowConfigLogsBucket,
		}
		workflowRequest.CiArtifactLocation = impl.buildDefaultArtifactLocation(savedWf)
		workflowRequest.CiArtifactFileName = workflowRequest.CiArtifactLocation
	case types.BLOB_STORAGE_AZURE:
		workflowRequest.AzureBlobConfig = &blob_storage.AzureBlobConfig{
			Enabled:               impl.config.CloudProvider == types.BLOB_STORAGE_AZURE,
			AccountName:           impl.config.AzureAccountName,
			BlobContainerCiCache:  impl.config.AzureBlobContainerCiCache,
			AccountKey:            impl.config.AzureAccountKey,
			BlobContainerCiLog:    impl.config.AzureBlobContainerCiLog,
			BlobContainerArtifact: impl.config.AzureBlobContainerCiLog,
		}
		workflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			EndpointUrl:           impl.config.AzureGatewayUrl,
			IsInSecure:            impl.config.AzureGatewayConnectionInsecure,
			CiLogBucketName:       impl.config.AzureBlobContainerCiLog,
			CiLogRegion:           impl.config.DefaultCacheBucketRegion,
			CiLogBucketVersioning: impl.config.BlobStorageS3BucketVersioned,
			AccessKey:             impl.config.AzureAccountName,
		}
		workflowRequest.CiArtifactLocation = impl.buildDefaultArtifactLocation(savedWf)
		workflowRequest.CiArtifactFileName = workflowRequest.CiArtifactLocation
	default:
		if impl.config.BlobStorageEnabled {
			return nil, fmt.Errorf("blob storage %s not supported", workflowRequest.CloudProvider)
		}
	}
	return workflowRequest, nil
}

func (impl *ServiceImpl) GetWorkflowRequestVariablesForCopyContainerImagePlugin(preCiSteps []*pipelineConfigBean.StepObject, postCiSteps []*pipelineConfigBean.StepObject, customTag string, customTagId int, buildImagePath string, buildImagedockerRegistryId string) (map[string][]string, map[string]bean2.RegistryCredentials, string, []int, error) {

	copyContainerImagePluginDetail, err := impl.globalPluginService.GetRefPluginIdByRefPluginName(buildCommonBean.COPY_CONTAINER_IMAGE)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in getting copyContainerImage plugin id", "err", err)
		return nil, nil, "", nil, err
	}

	pluginIdToVersionMap := make(map[int]string)
	for _, p := range copyContainerImagePluginDetail {
		pluginIdToVersionMap[p.Id] = p.Version
	}

	for _, step := range preCiSteps {
		if _, ok := pluginIdToVersionMap[step.RefPluginId]; ok {
			// for copyContainerImage plugin parse destination images and save its data in image path reservation table
			return nil, nil, "", nil, errors.New("copyContainerImage plugin not allowed in pre-ci step, please remove it and try again")
		}
	}

	registryCredentialMap := make(map[string]bean2.RegistryCredentials)
	registryDestinationImageMap := make(map[string][]string)
	var allDestinationImages []string //saving all images to be reserved in this array

	for _, step := range postCiSteps {
		if version, ok := pluginIdToVersionMap[step.RefPluginId]; ok {
			destinationImageMap, credentialMap, err := impl.pluginInputVariableParser.HandleCopyContainerImagePluginInputVariables(step.InputVars, customTag, buildImagePath, buildImagedockerRegistryId)
			if err != nil {
				impl.Logger.Errorw("error in parsing copyContainerImage input variable", "err", err)
				return nil, nil, "", nil, err
			}
			if version == buildCommonBean.COPY_CONTAINER_IMAGE_VERSION_V1 {
				// this is needed in ci runner only for v1
				registryDestinationImageMap = destinationImageMap
			}
			for _, images := range destinationImageMap {
				allDestinationImages = append(allDestinationImages, images...)
			}
			for k, v := range credentialMap {
				registryCredentialMap[k] = v
			}
		}
	}

	pluginArtifactStage := repository5.POST_CI
	for _, image := range allDestinationImages {
		if image == buildImagePath {
			return nil, registryCredentialMap, pluginArtifactStage, nil,
				pipelineConfigBean.ErrImagePathInUse
		}
	}

	var imagePathReservationIds []int
	if len(allDestinationImages) > 0 {
		savedCIArtifacts, err := impl.ciArtifactRepository.FindCiArtifactByImagePaths(allDestinationImages)
		if err != nil {
			impl.Logger.Errorw("error in fetching artifacts by image path", "err", err)
			return nil, nil, pluginArtifactStage, nil, err
		}
		if len(savedCIArtifacts) > 0 {
			// if already present in ci artifact, return "image path already in use error"
			return nil, nil, pluginArtifactStage, nil, pipelineConfigBean.ErrImagePathInUse
		}
		imagePathReservationIds, err = impl.ReserveImagesGeneratedAtPlugin(customTagId, allDestinationImages)
		if err != nil {
			return nil, nil, pluginArtifactStage, imagePathReservationIds, err
		}
	}
	return registryDestinationImageMap, registryCredentialMap, pluginArtifactStage, imagePathReservationIds, nil
}

func (impl *ServiceImpl) ReserveImagesGeneratedAtPlugin(customTagId int, destinationImages []string) ([]int, error) {
	var imagePathReservationIds []int
	for _, image := range destinationImages {
		imagePathReservationData, err := impl.customTagService.ReserveImagePath(image, customTagId)
		if err != nil {
			impl.Logger.Errorw("Error in marking custom tag reserved", "err", err)
			return imagePathReservationIds, err
		}
		imagePathReservationIds = append(imagePathReservationIds, imagePathReservationData.Id)
	}
	return imagePathReservationIds, nil
}

func buildCiStepsDataFromDockerBuildScripts(dockerBuildScripts []*bean.CiScript) []*pipelineConfigBean.StepObject {
	// before plugin support, few variables were set as env vars in ci-runner
	// these variables are now moved to global vars in plugin steps, but to avoid error in old scripts adding those variables in payload
	inputVars := []*commonBean.VariableObject{
		{
			Name:                  "DOCKER_IMAGE_TAG",
			Format:                "STRING",
			VariableType:          commonBean.VariableTypeRefGlobal,
			ReferenceVariableName: "DOCKER_IMAGE_TAG",
		},
		{
			Name:                  "DOCKER_REPOSITORY",
			Format:                "STRING",
			VariableType:          commonBean.VariableTypeRefGlobal,
			ReferenceVariableName: "DOCKER_REPOSITORY",
		},
		{
			Name:                  "DOCKER_REGISTRY_URL",
			Format:                "STRING",
			VariableType:          commonBean.VariableTypeRefGlobal,
			ReferenceVariableName: "DOCKER_REGISTRY_URL",
		},
		{
			Name:                  "DOCKER_IMAGE",
			Format:                "STRING",
			VariableType:          commonBean.VariableTypeRefGlobal,
			ReferenceVariableName: "DOCKER_IMAGE",
		},
	}
	var ciSteps []*pipelineConfigBean.StepObject
	for _, dockerBuildScript := range dockerBuildScripts {
		ciStep := &pipelineConfigBean.StepObject{
			Name:          dockerBuildScript.Name,
			Index:         dockerBuildScript.Index,
			Script:        dockerBuildScript.Script,
			ArtifactPaths: []string{dockerBuildScript.OutputLocation},
			StepType:      string(repository.PIPELINE_STEP_TYPE_INLINE),
			ExecutorType:  string(repository2.SCRIPT_TYPE_SHELL),
			InputVars:     inputVars,
		}
		ciSteps = append(ciSteps, ciStep)
	}
	return ciSteps
}

func (impl *ServiceImpl) buildImageTag(commitHashes map[int]pipelineConfig.GitCommit, id int, wfId int) string {
	dockerImageTag := ""
	toAppendDevtronParamInTag := true
	for _, v := range commitHashes {
		if v.WebhookData.Id == 0 {
			if v.Commit == "" {
				continue
			}
			dockerImageTag = getUpdatedDockerImageTagWithCommitOrCheckOutData(dockerImageTag, _getTruncatedImageTag(v.Commit))
		} else {
			_targetCheckout := v.WebhookData.Data[bean.WEBHOOK_SELECTOR_TARGET_CHECKOUT_NAME]
			if _targetCheckout == "" {
				continue
			}
			// if not PR based then meaning tag based
			isPRBasedEvent := v.WebhookData.EventActionType == bean.WEBHOOK_EVENT_MERGED_ACTION_TYPE
			if !isPRBasedEvent && impl.config.CiCdConfig.UseImageTagFromGitProviderForTagBasedBuild {
				dockerImageTag = getUpdatedDockerImageTagWithCommitOrCheckOutData(dockerImageTag, _targetCheckout)
			} else {
				dockerImageTag = getUpdatedDockerImageTagWithCommitOrCheckOutData(dockerImageTag, _getTruncatedImageTag(_targetCheckout))
			}
			if isPRBasedEvent {
				_sourceCheckout := v.WebhookData.Data[bean.WEBHOOK_SELECTOR_SOURCE_CHECKOUT_NAME]
				dockerImageTag = getUpdatedDockerImageTagWithCommitOrCheckOutData(dockerImageTag, _getTruncatedImageTag(_sourceCheckout))
			} else {
				toAppendDevtronParamInTag = !impl.config.CiCdConfig.UseImageTagFromGitProviderForTagBasedBuild
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

func (impl *ServiceImpl) updateCiWorkflow(request *types.WorkflowRequest, savedWf *pipelineConfig.CiWorkflow) error {
	ciBuildConfig := request.CiBuildConfig
	ciBuildType := string(ciBuildConfig.CiBuildType)
	savedWf.CiBuildType = ciBuildType
	return impl.ciService.UpdateCiWorkflowWithStage(savedWf)
}

func (impl *ServiceImpl) handleRuntimeParamsValidations(trigger types.Trigger, ciMaterials []*pipelineConfig.CiPipelineMaterial, workflowRequest *types.WorkflowRequest) error {
	// externalCi artifact is meant only for CI_JOB
	if trigger.PipelineType != string(buildCommonBean.CI_JOB) {
		return nil
	}

	// checking if user has given run time parameters for externalCiArtifact, if given then sending git material to Ci-Runner
	externalCiArtifact, exists := trigger.RuntimeParameters.GetGlobalRuntimeVariables()[buildBean.ExtraEnvVarExternalCiArtifactKey]
	// validate externalCiArtifact as docker image
	if exists {
		externalCiArtifact = strings.TrimSpace(externalCiArtifact)
		if !strings.Contains(externalCiArtifact, ":") {
			if utils.IsValidDockerTagName(externalCiArtifact) {
				fullImageUrl, err := utils.BuildDockerImagePath(bean4.DockerRegistryInfo{
					DockerImageTag:     externalCiArtifact,
					DockerRegistryId:   workflowRequest.DockerRegistryId,
					DockerRegistryType: workflowRequest.DockerRegistryType,
					DockerRegistryURL:  workflowRequest.DockerRegistryURL,
					DockerRepository:   workflowRequest.DockerRepository,
				})
				if err != nil {
					impl.Logger.Errorw("Error in building docker image", "err", err)
					return err
				}
				externalCiArtifact = fullImageUrl
			} else {
				impl.Logger.Errorw("validation error", "externalCiArtifact", externalCiArtifact)
				return fmt.Errorf("invalid image name or url given in externalCiArtifact")
			}

		}
		// This will overwrite the existing runtime parameters value for constants.externalCiArtifact
		trigger.RuntimeParameters = trigger.RuntimeParameters.AddRuntimeGlobalVariable(buildBean.ExtraEnvVarExternalCiArtifactKey, externalCiArtifact)
		var artifactExists bool
		var err error

		imageDigest, ok := trigger.RuntimeParameters.GetGlobalRuntimeVariables()[buildBean.ExtraEnvVarImageDigestKey]
		if !ok || len(imageDigest) == 0 {
			artifactExists, err = impl.ciArtifactRepository.IfArtifactExistByImage(externalCiArtifact, trigger.PipelineId)
			if err != nil {
				impl.Logger.Errorw("error in fetching ci artifact", "err", err)
				return err
			}
			if artifactExists {
				impl.Logger.Errorw("ci artifact already exists with same image name", "artifact", externalCiArtifact)
				return fmt.Errorf("ci artifact already exists with same image name")
			}
		} else {
			artifactExists, err = impl.ciArtifactRepository.IfArtifactExistByImageDigest(imageDigest, externalCiArtifact, trigger.PipelineId)
			if err != nil {
				impl.Logger.Errorw("error in fetching ci artifact", "err", err, "imageDigest", imageDigest)
				return err
			}
			if artifactExists {
				impl.Logger.Errorw("ci artifact already exists  with same digest", "artifact", externalCiArtifact)
				return fmt.Errorf("ci artifact already exists  with same digest")
			}
		}
	}
	if trigger.PipelineType == string(buildCommonBean.CI_JOB) && len(ciMaterials) != 0 && !exists && externalCiArtifact == "" {
		ciMaterials[0].GitMaterial = nil
		ciMaterials[0].GitMaterialId = 0
	}
	return nil
}

func _getTruncatedImageTag(imageTag string) string {
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

func (impl *ServiceImpl) markCurrentCiWorkflowFailed(savedCiWf *pipelineConfig.CiWorkflow, validationErr error) error {
	// currently such requirement is not there
	if savedCiWf == nil {
		return nil
	}
	if savedCiWf.Id != 0 && slices.Contains(cdWorkflow.WfrTerminalStatusList, savedCiWf.Status) {
		impl.Logger.Debug("workflow is already in terminal state", "status", savedCiWf.Status, "workflowId", savedCiWf.Id, "message", savedCiWf.Message)
		return nil
	}

	savedCiWf.Status = cdWorkflow.WorkflowFailed
	savedCiWf.Message = validationErr.Error()
	savedCiWf.FinishedOn = time.Now()

	var dbErr error
	if savedCiWf.Id == 0 {
		dbErr = impl.ciService.SaveCiWorkflowWithStage(savedCiWf)
	} else {
		dbErr = impl.ciService.UpdateCiWorkflowWithStage(savedCiWf)
	}

	if dbErr != nil {
		impl.Logger.Errorw("save/update workflow error", "err", dbErr)
		return dbErr
	}
	return nil
}
