package devtronApps

import (
	"context"
	"errors"
	"fmt"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	util5 "github.com/devtron-labs/common-lib/utils/k8s"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	bean6 "github.com/devtron-labs/devtron/api/helm-app/bean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	client2 "github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/client/argocdServer"
	bean7 "github.com/devtron-labs/devtron/client/argocdServer/bean"
	client "github.com/devtron-labs/devtron/client/events"
	gitSensorClient "github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/models"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	bean4 "github.com/devtron-labs/devtron/pkg/app/bean"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest"
	bean5 "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/helper"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/out"
	"github.com/devtron-labs/devtron/pkg/imageDigestPolicy"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	bean8 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/plugin"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	util2 "github.com/devtron-labs/devtron/util/event"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	status2 "google.golang.org/grpc/status"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"path"
	"strconv"
	"strings"
	"time"
)

type TriggerService interface {
	TriggerPostStage(request bean.TriggerRequest) error
	TriggerPreStage(request bean.TriggerRequest) error

	TriggerStageForBulk(triggerRequest bean.TriggerRequest) error

	ManualCdTrigger(triggerContext bean.TriggerContext, overrideRequest *bean3.ValuesOverrideRequest) (int, error)
	TriggerAutomaticDeployment(request bean.TriggerRequest) error

	HandleCDTriggerRelease(overrideRequest *bean3.ValuesOverrideRequest, ctx context.Context,
		triggeredAt time.Time, deployedBy int32) (releaseNo int, manifest []byte, err error)

	TriggerRelease(overrideRequest *bean3.ValuesOverrideRequest, valuesOverrideResponse *app.ValuesOverrideResponse,
		builtChartPath string, ctx context.Context, triggeredAt time.Time, triggeredBy int32) (releaseNo int, manifest []byte, err error)

	//TODO: make this method private and move all usages in this service since TriggerService should own if async mode is enabled and if yes then how to act on it
	IsDevtronAsyncInstallModeEnabled(deploymentAppType string) bool
}

type TriggerServiceImpl struct {
	logger                              *zap.SugaredLogger
	cdWorkflowCommonService             cd.CdWorkflowCommonService
	gitOpsManifestPushService           app.GitOpsPushService
	argoK8sClient                       argocdServer.ArgoK8sClient
	ACDConfig                           *argocdServer.ACDConfig
	argoClientWrapperService            argocdServer.ArgoClientWrapperService
	pipelineStatusTimelineService       status.PipelineStatusTimelineService
	chartTemplateService                util.ChartTemplateService
	eventFactory                        client.EventFactory
	eventClient                         client.EventClient
	globalEnvVariables                  *util3.GlobalEnvVariables
	workflowEventPublishService         out.WorkflowEventPublishService
	manifestCreationService             manifest.ManifestCreationService
	deployedConfigurationHistoryService history.DeployedConfigurationHistoryService
	argoUserService                     argo.ArgoUserService
	pipelineStageService                pipeline.PipelineStageService
	globalPluginService                 plugin.GlobalPluginService
	customTagService                    pipeline.CustomTagService
	pluginInputVariableParser           pipeline.PluginInputVariableParser
	prePostCdScriptHistoryService       history.PrePostCdScriptHistoryService
	scopedVariableManager               variables.ScopedVariableCMCSManager
	cdWorkflowService                   pipeline.WorkflowService
	imageDigestPolicyService            imageDigestPolicy.ImageDigestPolicyService
	userService                         user.UserService
	gitSensorGrpcClient                 gitSensorClient.Client
	config                              *types.CdConfig

	helmAppService client2.HelmAppService

	enforcerUtil  rbac.EnforcerUtil
	helmAppClient gRPC.HelmAppClient //TODO refactoring: use helm app service instead

	scanResultRepository          security.ImageScanResultRepository
	cvePolicyRepository           security.CvePolicyRepository
	ciPipelineMaterialRepository  pipelineConfig.CiPipelineMaterialRepository
	imageScanHistoryRepository    security.ImageScanHistoryRepository
	imageScanDeployInfoRepository security.ImageScanDeployInfoRepository
	pipelineRepository            pipelineConfig.PipelineRepository
	pipelineOverrideRepository    chartConfig.PipelineOverrideRepository
	manifestPushConfigRepository  repository.ManifestPushConfigRepository
	chartRepository               chartRepoRepository.ChartRepository
	envRepository                 repository2.EnvironmentRepository
	cdWorkflowRepository          pipelineConfig.CdWorkflowRepository
	ciWorkflowRepository          pipelineConfig.CiWorkflowRepository
	ciArtifactRepository          repository3.CiArtifactRepository
	ciTemplateRepository          pipelineConfig.CiTemplateRepository
	materialRepository            pipelineConfig.MaterialRepository
	appLabelRepository            pipelineConfig.AppLabelRepository
	ciPipelineRepository          pipelineConfig.CiPipelineRepository
	appWorkflowRepository         appWorkflow.AppWorkflowRepository
}

func NewTriggerServiceImpl(logger *zap.SugaredLogger, cdWorkflowCommonService cd.CdWorkflowCommonService,
	gitOpsManifestPushService app.GitOpsPushService,
	argoK8sClient argocdServer.ArgoK8sClient,
	ACDConfig *argocdServer.ACDConfig,
	argoClientWrapperService argocdServer.ArgoClientWrapperService,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	chartTemplateService util.ChartTemplateService,
	workflowEventPublishService out.WorkflowEventPublishService,
	manifestCreationService manifest.ManifestCreationService,
	deployedConfigurationHistoryService history.DeployedConfigurationHistoryService,
	argoUserService argo.ArgoUserService,
	pipelineStageService pipeline.PipelineStageService,
	globalPluginService plugin.GlobalPluginService,
	customTagService pipeline.CustomTagService,
	pluginInputVariableParser pipeline.PluginInputVariableParser,
	prePostCdScriptHistoryService history.PrePostCdScriptHistoryService,
	scopedVariableManager variables.ScopedVariableCMCSManager,
	cdWorkflowService pipeline.WorkflowService,
	imageDigestPolicyService imageDigestPolicy.ImageDigestPolicyService,
	userService user.UserService,
	gitSensorGrpcClient gitSensorClient.Client,
	config *types.CdConfig,
	helmAppService client2.HelmAppService,
	enforcerUtil rbac.EnforcerUtil,
	helmAppClient gRPC.HelmAppClient,
	eventFactory client.EventFactory,
	eventClient client.EventClient,
	globalEnvVariables *util3.GlobalEnvVariables,
	scanResultRepository security.ImageScanResultRepository,
	cvePolicyRepository security.CvePolicyRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	imageScanHistoryRepository security.ImageScanHistoryRepository,
	imageScanDeployInfoRepository security.ImageScanDeployInfoRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
	manifestPushConfigRepository repository.ManifestPushConfigRepository,
	chartRepository chartRepoRepository.ChartRepository,
	envRepository repository2.EnvironmentRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	ciArtifactRepository repository3.CiArtifactRepository,
	ciTemplateRepository pipelineConfig.CiTemplateRepository,
	materialRepository pipelineConfig.MaterialRepository,
	appLabelRepository pipelineConfig.AppLabelRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	appWorkflowRepository appWorkflow.AppWorkflowRepository) *TriggerServiceImpl {
	return &TriggerServiceImpl{
		logger:                              logger,
		cdWorkflowCommonService:             cdWorkflowCommonService,
		gitOpsManifestPushService:           gitOpsManifestPushService,
		argoK8sClient:                       argoK8sClient,
		ACDConfig:                           ACDConfig,
		argoClientWrapperService:            argoClientWrapperService,
		pipelineStatusTimelineService:       pipelineStatusTimelineService,
		chartTemplateService:                chartTemplateService,
		workflowEventPublishService:         workflowEventPublishService,
		manifestCreationService:             manifestCreationService,
		deployedConfigurationHistoryService: deployedConfigurationHistoryService,
		argoUserService:                     argoUserService,
		pipelineStageService:                pipelineStageService,
		globalPluginService:                 globalPluginService,
		customTagService:                    customTagService,
		pluginInputVariableParser:           pluginInputVariableParser,
		prePostCdScriptHistoryService:       prePostCdScriptHistoryService,
		scopedVariableManager:               scopedVariableManager,
		cdWorkflowService:                   cdWorkflowService,
		imageDigestPolicyService:            imageDigestPolicyService,
		userService:                         userService,
		gitSensorGrpcClient:                 gitSensorGrpcClient,
		config:                              config,
		helmAppService:                      helmAppService,
		enforcerUtil:                        enforcerUtil,
		eventFactory:                        eventFactory,
		eventClient:                         eventClient,
		globalEnvVariables:                  globalEnvVariables,
		helmAppClient:                       helmAppClient,
		scanResultRepository:                scanResultRepository,
		cvePolicyRepository:                 cvePolicyRepository,
		ciPipelineMaterialRepository:        ciPipelineMaterialRepository,
		imageScanHistoryRepository:          imageScanHistoryRepository,
		imageScanDeployInfoRepository:       imageScanDeployInfoRepository,
		pipelineRepository:                  pipelineRepository,
		pipelineOverrideRepository:          pipelineOverrideRepository,
		manifestPushConfigRepository:        manifestPushConfigRepository,
		chartRepository:                     chartRepository,
		envRepository:                       envRepository,
		cdWorkflowRepository:                cdWorkflowRepository,
		ciWorkflowRepository:                ciWorkflowRepository,
		ciArtifactRepository:                ciArtifactRepository,
		ciTemplateRepository:                ciTemplateRepository,
		materialRepository:                  materialRepository,
		appLabelRepository:                  appLabelRepository,
		ciPipelineRepository:                ciPipelineRepository,
		appWorkflowRepository:               appWorkflowRepository,
	}
}

func (impl *TriggerServiceImpl) TriggerStageForBulk(triggerRequest bean.TriggerRequest) error {

	preStage, err := impl.pipelineStageService.GetCdStageByCdPipelineIdAndStageType(triggerRequest.Pipeline.Id, repository.PIPELINE_STAGE_TYPE_PRE_CD)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching CD pipeline stage", "cdPipelineId", triggerRequest.Pipeline.Id, "stage ", repository.PIPELINE_STAGE_TYPE_PRE_CD, "err", err)
		return err
	}

	//handle corrupt data (https://github.com/devtron-labs/devtron/issues/3826)
	err, deleted := impl.deleteCorruptedPipelineStage(preStage, triggerRequest.TriggeredBy)
	if err != nil {
		impl.logger.Errorw("error in deleteCorruptedPipelineStage ", "cdPipelineId", triggerRequest.Pipeline.Id, "err", err, "preStage", preStage, "triggeredBy", triggerRequest.TriggeredBy)
		return err
	}

	triggerRequest.TriggerContext.Context = context.Background()
	if len(triggerRequest.Pipeline.PreStageConfig) > 0 || (preStage != nil && !deleted) {
		//pre stage exists
		impl.logger.Debugw("trigger pre stage for pipeline", "artifactId", triggerRequest.Artifact.Id, "pipelineId", triggerRequest.Pipeline.Id)
		triggerRequest.RefCdWorkflowRunnerId = 0
		err = impl.TriggerPreStage(triggerRequest) // TODO handle error here
		return err
	} else {
		// trigger deployment
		impl.logger.Debugw("trigger cd for pipeline", "artifactId", triggerRequest.Artifact.Id, "pipelineId", triggerRequest.Pipeline.Id)
		err = impl.TriggerAutomaticDeployment(triggerRequest)
		return err
	}
}

// TODO: write a wrapper to handle auto and manual trigger
func (impl *TriggerServiceImpl) ManualCdTrigger(triggerContext bean.TriggerContext, overrideRequest *bean3.ValuesOverrideRequest) (int, error) {
	//setting triggeredAt variable to have consistent data for various audit log places in db for deployment time
	triggeredAt := time.Now()
	releaseId := 0
	ctx := triggerContext.Context
	var err error
	_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineRepository.FindById")
	cdPipeline, err := impl.pipelineRepository.FindById(overrideRequest.PipelineId)
	span.End()
	if err != nil {
		impl.logger.Errorw("manual trigger request with invalid pipelineId, ManualCdTrigger", "pipelineId", overrideRequest.PipelineId, "err", err)
		return 0, err
	}
	setPipelineFieldsInOverrideRequest(overrideRequest, cdPipeline)

	switch overrideRequest.CdWorkflowType {
	case bean3.CD_WORKFLOW_TYPE_PRE:
		_, span = otel.Tracer("orchestrator").Start(ctx, "ciArtifactRepository.Get")
		artifact, err := impl.ciArtifactRepository.Get(overrideRequest.CiArtifactId)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in getting CiArtifact", "CiArtifactId", overrideRequest.CiArtifactId, "err", err)
			return 0, err
		}
		// Migration of deprecated DataSource Type
		if artifact.IsMigrationRequired() {
			migrationErr := impl.ciArtifactRepository.MigrateToWebHookDataSourceType(artifact.Id)
			if migrationErr != nil {
				impl.logger.Warnw("unable to migrate deprecated DataSource", "artifactId", artifact.Id)
			}
		}
		_, span = otel.Tracer("orchestrator").Start(ctx, "TriggerPreStage")
		triggerRequest := bean.TriggerRequest{
			CdWf:                  nil,
			Artifact:              artifact,
			Pipeline:              cdPipeline,
			TriggeredBy:           overrideRequest.UserId,
			ApplyAuth:             false,
			TriggerContext:        triggerContext,
			RefCdWorkflowRunnerId: 0,
		}
		err = impl.TriggerPreStage(triggerRequest)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in TriggerPreStage, ManualCdTrigger", "err", err)
			return 0, err
		}
	case bean3.CD_WORKFLOW_TYPE_DEPLOY:
		if overrideRequest.DeploymentType == models.DEPLOYMENTTYPE_UNKNOWN {
			overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_DEPLOY
		}

		cdWf, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(ctx, overrideRequest.CdWorkflowId, bean3.CD_WORKFLOW_TYPE_PRE)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("error in getting cdWorkflow, ManualCdTrigger", "CdWorkflowId", overrideRequest.CdWorkflowId, "err", err)
			return 0, err
		}

		cdWorkflowId := cdWf.CdWorkflowId
		if cdWf.CdWorkflowId == 0 {
			cdWf := &pipelineConfig.CdWorkflow{
				CiArtifactId: overrideRequest.CiArtifactId,
				PipelineId:   overrideRequest.PipelineId,
				AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
			}
			err := impl.cdWorkflowRepository.SaveWorkFlow(ctx, cdWf)
			if err != nil {
				impl.logger.Errorw("error in creating cdWorkflow, ManualCdTrigger", "PipelineId", overrideRequest.PipelineId, "err", err)
				return 0, err
			}
			cdWorkflowId = cdWf.Id
		}

		runner := &pipelineConfig.CdWorkflowRunner{
			Name:         cdPipeline.Name,
			WorkflowType: bean3.CD_WORKFLOW_TYPE_DEPLOY,
			ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
			Status:       pipelineConfig.WorkflowInitiated, //deployment Initiated for manual trigger
			TriggeredBy:  overrideRequest.UserId,
			StartedOn:    triggeredAt,
			Namespace:    impl.config.GetDefaultNamespace(),
			CdWorkflowId: cdWorkflowId,
			AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
			ReferenceId:  triggerContext.ReferenceId,
		}
		savedWfr, err := impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
		overrideRequest.WfrId = savedWfr.Id
		if err != nil {
			impl.logger.Errorw("err in creating cdWorkflowRunner, ManualCdTrigger", "cdWorkflowId", cdWorkflowId, "err", err)
			return 0, err
		}
		runner.CdWorkflow = &pipelineConfig.CdWorkflow{
			Pipeline: cdPipeline,
		}
		overrideRequest.CdWorkflowId = cdWorkflowId
		// creating cd pipeline status timeline for deployment initialisation
		timeline := impl.pipelineStatusTimelineService.GetTimelineDbObjectByTimelineStatusAndTimelineDescription(savedWfr.Id, 0, pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_INITIATED, pipelineConfig.TIMELINE_DESCRIPTION_DEPLOYMENT_INITIATED, overrideRequest.UserId, time.Now())
		_, span = otel.Tracer("orchestrator").Start(ctx, "cdPipelineStatusTimelineRepo.SaveTimelineForACDHelmApps")
		err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)

		span.End()
		if err != nil {
			impl.logger.Errorw("error in creating timeline status for deployment initiation, ManualCdTrigger", "err", err, "timeline", timeline)
		}

		//checking vulnerability for deploying image
		_, span = otel.Tracer("orchestrator").Start(ctx, "ciArtifactRepository.Get")
		artifact, err := impl.ciArtifactRepository.Get(overrideRequest.CiArtifactId)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in getting ciArtifact, ManualCdTrigger", "CiArtifactId", overrideRequest.CiArtifactId, "err", err)
			return 0, err
		}
		// Migration of deprecated DataSource Type
		if artifact.IsMigrationRequired() {
			migrationErr := impl.ciArtifactRepository.MigrateToWebHookDataSourceType(artifact.Id)
			if migrationErr != nil {
				impl.logger.Warnw("unable to migrate deprecated DataSource", "artifactId", artifact.Id)
			}
		}
		isVulnerable, err := impl.GetArtifactVulnerabilityStatus(artifact, cdPipeline, ctx)
		if err != nil {
			impl.logger.Errorw("error in getting Artifact vulnerability status, ManualCdTrigger", "err", err)
			return 0, err
		}

		if isVulnerable == true {
			// if image vulnerable, update timeline status and return
			if err = impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(runner, errors.New(pipelineConfig.FOUND_VULNERABILITY), overrideRequest.UserId); err != nil {
				impl.logger.Errorw("error while updating current runner status to failed, TriggerDeployment", "wfrId", runner.Id, "err", err)
			}
			return 0, fmt.Errorf("found vulnerability for image digest %s", artifact.ImageDigest)
		}

		// Deploy the release
		_, span = otel.Tracer("orchestrator").Start(ctx, "appService.TriggerRelease")
		var releaseErr error
		releaseId, _, releaseErr = impl.HandleCDTriggerRelease(overrideRequest, ctx, triggeredAt, overrideRequest.UserId)
		span.End()
		// if releaseErr found, then the mark current deployment Failed and return
		if releaseErr != nil {
			err := impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(runner, releaseErr, overrideRequest.UserId)
			if err != nil {
				impl.logger.Errorw("error while updating current runner status to failed, updatePreviousDeploymentStatus", "cdWfr", runner.Id, "err", err)
			}
			return 0, releaseErr
		}

		// skip updatePreviousDeploymentStatus if Async Install is enabled; handled inside SubscribeDevtronAsyncHelmInstallRequest
		if !impl.IsDevtronAsyncInstallModeEnabled(cdPipeline.DeploymentAppType) {
			// Update previous deployment runner status (in transaction): Failed
			_, span = otel.Tracer("orchestrator").Start(ctx, "updatePreviousDeploymentStatus")
			err1 := impl.cdWorkflowCommonService.UpdatePreviousDeploymentStatus(runner, cdPipeline.Id, triggeredAt, overrideRequest.UserId)
			span.End()
			if err1 != nil {
				impl.logger.Errorw("error while update previous cd workflow runners, ManualCdTrigger", "err", err, "runner", runner, "pipelineId", cdPipeline.Id)
				return 0, err1
			}
		}

		if overrideRequest.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_MANIFEST_DOWNLOAD {
			runner := &pipelineConfig.CdWorkflowRunner{
				Id:           runner.Id,
				Name:         cdPipeline.Name,
				WorkflowType: bean3.CD_WORKFLOW_TYPE_DEPLOY,
				ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
				TriggeredBy:  overrideRequest.UserId,
				StartedOn:    triggeredAt,
				Status:       pipelineConfig.WorkflowSucceeded,
				Namespace:    impl.config.GetDefaultNamespace(),
				CdWorkflowId: overrideRequest.CdWorkflowId,
				AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
			}
			updateErr := impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
			if updateErr != nil {
				impl.logger.Errorw("error in updating runner for manifest_download type, ManualCdTrigger", "CdWorkflowId", overrideRequest.CdWorkflowId, "err", err)
				return 0, updateErr
			}
		}

	case bean3.CD_WORKFLOW_TYPE_POST:
		cdWfRunner, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(ctx, overrideRequest.CdWorkflowId, bean3.CD_WORKFLOW_TYPE_DEPLOY)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("err in getting cdWorkflowRunner, ManualCdTrigger", "cdWorkflowId", overrideRequest.CdWorkflowId, "err", err)
			return 0, err
		}

		var cdWf *pipelineConfig.CdWorkflow
		if cdWfRunner.CdWorkflowId == 0 {
			cdWf = &pipelineConfig.CdWorkflow{
				CiArtifactId: overrideRequest.CiArtifactId,
				PipelineId:   overrideRequest.PipelineId,
				AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
			}
			err := impl.cdWorkflowRepository.SaveWorkFlow(ctx, cdWf)
			if err != nil {
				impl.logger.Errorw("error in creating cdWorkflow, ManualCdTrigger", "CdWorkflowId", overrideRequest.CdWorkflowId, "err", err)
				return 0, err
			}
		} else {
			_, span = otel.Tracer("orchestrator").Start(ctx, "cdWorkflowRepository.FindById")
			cdWf, err = impl.cdWorkflowRepository.FindById(overrideRequest.CdWorkflowId)
			span.End()
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("error in getting cdWorkflow, ManualCdTrigger", "CdWorkflowId", overrideRequest.CdWorkflowId, "err", err)
				return 0, err
			}
		}
		_, span = otel.Tracer("orchestrator").Start(ctx, "TriggerPostStage")
		triggerRequest := bean.TriggerRequest{
			CdWf:                  cdWf,
			Pipeline:              cdPipeline,
			TriggeredBy:           overrideRequest.UserId,
			RefCdWorkflowRunnerId: 0,
			TriggerContext:        triggerContext,
		}
		err = impl.TriggerPostStage(triggerRequest)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in TriggerPostStage, ManualCdTrigger", "CdWorkflowId", cdWf.Id, "err", err)
			return 0, err
		}
	default:
		impl.logger.Errorw("invalid CdWorkflowType, ManualCdTrigger", "CdWorkflowType", overrideRequest.CdWorkflowType, "err", err)
		return 0, fmt.Errorf("invalid CdWorkflowType %s for the trigger request", string(overrideRequest.CdWorkflowType))
	}

	return releaseId, err
}

// TODO: write a wrapper to handle auto and manual trigger
func (impl *TriggerServiceImpl) TriggerAutomaticDeployment(request bean.TriggerRequest) error {
	//in case of manual trigger auth is already applied and for auto triggers there is no need for auth check here
	triggeredBy := request.TriggeredBy
	pipeline := request.Pipeline
	artifact := request.Artifact

	//setting triggeredAt variable to have consistent data for various audit log places in db for deployment time
	triggeredAt := time.Now()
	cdWf := request.CdWf

	if cdWf == nil || (cdWf != nil && cdWf.CiArtifactId != artifact.Id) {
		// cdWf != nil && cdWf.CiArtifactId != artifact.Id for auto trigger case when deployment is triggered with image generated by plugin
		cdWf = &pipelineConfig.CdWorkflow{
			CiArtifactId: artifact.Id,
			PipelineId:   pipeline.Id,
			AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: 1, UpdatedOn: triggeredAt, UpdatedBy: 1},
		}
		err := impl.cdWorkflowRepository.SaveWorkFlow(context.Background(), cdWf)
		if err != nil {
			return err
		}
	}

	runner := &pipelineConfig.CdWorkflowRunner{
		Name:         pipeline.Name,
		WorkflowType: bean3.CD_WORKFLOW_TYPE_DEPLOY,
		ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM,
		Status:       pipelineConfig.WorkflowInitiated, // deployment Initiated for auto trigger
		TriggeredBy:  1,
		StartedOn:    triggeredAt,
		Namespace:    impl.config.GetDefaultNamespace(),
		CdWorkflowId: cdWf.Id,
		AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: triggeredBy, UpdatedOn: triggeredAt, UpdatedBy: triggeredBy},
		ReferenceId:  request.TriggerContext.ReferenceId,
	}
	savedWfr, err := impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
	if err != nil {
		return err
	}
	runner.CdWorkflow = &pipelineConfig.CdWorkflow{
		Pipeline: pipeline,
	}
	// creating cd pipeline status timeline for deployment initialisation
	timeline := &pipelineConfig.PipelineStatusTimeline{
		CdWorkflowRunnerId: runner.Id,
		Status:             pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_INITIATED,
		StatusDetail:       "Deployment initiated successfully.",
		StatusTime:         time.Now(),
		AuditLog: sql.AuditLog{
			CreatedBy: 1,
			CreatedOn: time.Now(),
			UpdatedBy: 1,
			UpdatedOn: time.Now(),
		},
	}
	isAppStore := false
	err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, isAppStore)
	if err != nil {
		impl.logger.Errorw("error in creating timeline status for deployment initiation", "err", err, "timeline", timeline)
	}
	//checking vulnerability for deploying image
	isVulnerable := false
	if len(artifact.ImageDigest) > 0 {
		var cveStores []*security.CveStore
		imageScanResult, err := impl.scanResultRepository.FindByImageDigest(artifact.ImageDigest)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error fetching image digest", "digest", artifact.ImageDigest, "err", err)
			return err
		}
		for _, item := range imageScanResult {
			cveStores = append(cveStores, &item.CveStore)
		}
		env, err := impl.envRepository.FindById(pipeline.EnvironmentId)
		if err != nil {
			impl.logger.Errorw("error while fetching env", "err", err)
			return err
		}
		blockCveList, err := impl.cvePolicyRepository.GetBlockedCVEList(cveStores, env.ClusterId, pipeline.EnvironmentId, pipeline.AppId, false)
		if err != nil {
			impl.logger.Errorw("error while fetching blocked cve list", "err", err)
			return err
		}
		if len(blockCveList) > 0 {
			isVulnerable = true
		}
	}
	if isVulnerable == true {
		if err = impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(runner, errors.New(pipelineConfig.FOUND_VULNERABILITY), triggeredBy); err != nil {
			impl.logger.Errorw("error while updating current runner status to failed, TriggerDeployment", "wfrId", runner.Id, "err", err)
		}
		return nil
	}

	releaseErr := impl.TriggerCD(artifact, cdWf.Id, savedWfr.Id, pipeline, triggeredAt)
	// if releaseErr found, then the mark current deployment Failed and return
	if releaseErr != nil {
		err := impl.cdWorkflowCommonService.MarkCurrentDeploymentFailed(runner, releaseErr, triggeredBy)
		if err != nil {
			impl.logger.Errorw("error while updating current runner status to failed, updatePreviousDeploymentStatus", "cdWfr", runner.Id, "err", err)
		}
		return releaseErr
	}
	//skip updatePreviousDeploymentStatus if Async Install is enabled; handled inside SubscribeDevtronAsyncHelmInstallRequest
	if !impl.IsDevtronAsyncInstallModeEnabled(pipeline.DeploymentAppType) {
		err1 := impl.cdWorkflowCommonService.UpdatePreviousDeploymentStatus(runner, pipeline.Id, triggeredAt, triggeredBy)
		if err1 != nil {
			impl.logger.Errorw("error while update previous cd workflow runners", "err", err, "runner", runner, "pipelineId", pipeline.Id)
			return err1
		}
	}
	return nil
}

func (impl *TriggerServiceImpl) TriggerCD(artifact *repository3.CiArtifact, cdWorkflowId, wfrId int, pipeline *pipelineConfig.Pipeline, triggeredAt time.Time) error {
	impl.logger.Debugw("automatic pipeline trigger attempt async", "artifactId", artifact.Id)

	return impl.triggerReleaseAsync(artifact, cdWorkflowId, wfrId, pipeline, triggeredAt)
}

func (impl *TriggerServiceImpl) triggerReleaseAsync(artifact *repository3.CiArtifact, cdWorkflowId, wfrId int, pipeline *pipelineConfig.Pipeline, triggeredAt time.Time) error {
	err := impl.validateAndTrigger(pipeline, artifact, cdWorkflowId, wfrId, triggeredAt)
	if err != nil {
		impl.logger.Errorw("error in trigger for pipeline", "pipelineId", strconv.Itoa(pipeline.Id))
	}
	impl.logger.Debugw("trigger attempted for all pipeline ", "artifactId", artifact.Id)
	return err
}

func (impl *TriggerServiceImpl) validateAndTrigger(p *pipelineConfig.Pipeline, artifact *repository3.CiArtifact, cdWorkflowId, wfrId int, triggeredAt time.Time) error {
	//TODO: verify this logicc
	object := impl.enforcerUtil.GetAppRBACNameByAppId(p.AppId)
	envApp := strings.Split(object, "/")
	if len(envApp) != 2 {
		impl.logger.Error("invalid req, app and env not found from rbac")
		return errors.New("invalid req, app and env not found from rbac")
	}
	err := impl.releasePipeline(p, artifact, cdWorkflowId, wfrId, triggeredAt)
	return err
}

func (impl *TriggerServiceImpl) releasePipeline(pipeline *pipelineConfig.Pipeline, artifact *repository3.CiArtifact, cdWorkflowId, wfrId int, triggeredAt time.Time) error {
	impl.logger.Debugw("triggering release for ", "cdPipelineId", pipeline.Id, "artifactId", artifact.Id)

	pipeline, err := impl.pipelineRepository.FindById(pipeline.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline by pipelineId", "err", err)
		return err
	}

	request := &bean3.ValuesOverrideRequest{
		PipelineId:           pipeline.Id,
		UserId:               artifact.CreatedBy,
		CiArtifactId:         artifact.Id,
		AppId:                pipeline.AppId,
		CdWorkflowId:         cdWorkflowId,
		ForceTrigger:         true,
		DeploymentWithConfig: bean3.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED,
		WfrId:                wfrId,
	}
	setPipelineFieldsInOverrideRequest(request, pipeline)

	ctx, err := impl.argoUserService.BuildACDContext()
	if err != nil {
		impl.logger.Errorw("error in creating acd sync context", "pipelineId", pipeline.Id, "artifactId", artifact.Id, "err", err)
		return err
	}
	//setting deployedBy as 1(system user) since case of auto trigger
	id, _, err := impl.HandleCDTriggerRelease(request, ctx, triggeredAt, 1)
	if err != nil {
		impl.logger.Errorw("error in auto  cd pipeline trigger", "pipelineId", pipeline.Id, "artifactId", artifact.Id, "err", err)
	} else {
		impl.logger.Infow("pipeline successfully triggered ", "cdPipelineId", pipeline.Id, "artifactId", artifact.Id, "releaseId", id)
	}
	return err
}

func setPipelineFieldsInOverrideRequest(overrideRequest *bean3.ValuesOverrideRequest, pipeline *pipelineConfig.Pipeline) {
	overrideRequest.PipelineId = pipeline.Id
	overrideRequest.PipelineName = pipeline.Name
	overrideRequest.EnvId = pipeline.EnvironmentId
	overrideRequest.EnvName = pipeline.Environment.Name
	overrideRequest.ClusterId = pipeline.Environment.ClusterId
	overrideRequest.AppId = pipeline.AppId
	overrideRequest.AppName = pipeline.App.AppName
	overrideRequest.DeploymentAppType = pipeline.DeploymentAppType
}

func (impl *TriggerServiceImpl) HandleCDTriggerRelease(overrideRequest *bean3.ValuesOverrideRequest, ctx context.Context,
	triggeredAt time.Time, deployedBy int32) (releaseNo int, manifest []byte, err error) {
	if impl.IsDevtronAsyncInstallModeEnabled(overrideRequest.DeploymentAppType) {
		// asynchronous mode of installation starts
		return impl.workflowEventPublishService.TriggerHelmAsyncRelease(overrideRequest, ctx, triggeredAt, deployedBy)
	}
	// synchronous mode of installation starts

	valuesOverrideResponse, builtChartPath, err := impl.manifestCreationService.BuildManifestForTrigger(overrideRequest, triggeredAt, ctx)
	_, span := otel.Tracer("orchestrator").Start(ctx, "CreateHistoriesForDeploymentTrigger")
	err1 := impl.deployedConfigurationHistoryService.CreateHistoriesForDeploymentTrigger(valuesOverrideResponse.Pipeline, valuesOverrideResponse.PipelineStrategy, valuesOverrideResponse.EnvOverride, triggeredAt, deployedBy)
	if err1 != nil {
		impl.logger.Errorw("error in saving histories for trigger", "err", err1, "pipelineId", valuesOverrideResponse.Pipeline.Id, "wfrId", overrideRequest.WfrId)
	}
	span.End()
	if err != nil {
		impl.logger.Errorw("error in building merged manifest for trigger", "err", err)
		return releaseNo, manifest, err
	}
	return impl.TriggerRelease(overrideRequest, valuesOverrideResponse, builtChartPath, ctx, triggeredAt, deployedBy)
}

// TriggerRelease will trigger Install/Upgrade request for Devtron App releases synchronously
func (impl *TriggerServiceImpl) TriggerRelease(overrideRequest *bean3.ValuesOverrideRequest, valuesOverrideResponse *app.ValuesOverrideResponse,
	builtChartPath string, ctx context.Context, triggeredAt time.Time, triggeredBy int32) (releaseNo int, manifest []byte, err error) {
	// Handling for auto trigger
	if overrideRequest.UserId == 0 {
		overrideRequest.UserId = triggeredBy
	}
	triggerEvent := helper.GetTriggerEvent(overrideRequest.DeploymentAppType, triggeredAt, triggeredBy)
	releaseNo, manifest, err = impl.triggerPipeline(overrideRequest, valuesOverrideResponse, builtChartPath, triggerEvent, ctx)
	if err != nil {
		return 0, manifest, err
	}
	return releaseNo, manifest, nil
}

func (impl *TriggerServiceImpl) triggerPipeline(overrideRequest *bean3.ValuesOverrideRequest, valuesOverrideResponse *app.ValuesOverrideResponse,
	builtChartPath string, triggerEvent bean.TriggerEvent, ctx context.Context) (releaseNo int, manifest []byte, err error) {
	isRequestValid, err := helper.ValidateTriggerEvent(triggerEvent)
	if !isRequestValid {
		return releaseNo, manifest, err
	}

	if triggerEvent.PerformChartPush {
		//update workflow runner status, used in app workflow view
		err = impl.cdWorkflowCommonService.UpdateCDWorkflowRunnerStatus(ctx, overrideRequest, triggerEvent.TriggerdAt, pipelineConfig.WorkflowInProgress, "")
		if err != nil {
			impl.logger.Errorw("error in updating the workflow runner status, createHelmAppForCdPipeline", "err", err)
			return releaseNo, manifest, err
		}
		manifestPushTemplate, err := impl.buildManifestPushTemplate(overrideRequest, valuesOverrideResponse, builtChartPath, &manifest)
		if err != nil {
			impl.logger.Errorw("error in building manifest push template", "err", err)
			return releaseNo, manifest, err
		}
		manifestPushService := impl.getManifestPushService(triggerEvent)
		manifestPushResponse := manifestPushService.PushChart(manifestPushTemplate, ctx)
		if manifestPushResponse.Error != nil {
			impl.logger.Errorw("Error in pushing manifest to git", "err", err, "git_repo_url", manifestPushTemplate.RepoUrl)
			return releaseNo, manifest, manifestPushResponse.Error
		}
		pipelineOverrideUpdateRequest := &chartConfig.PipelineOverride{
			Id:                     valuesOverrideResponse.PipelineOverride.Id,
			GitHash:                manifestPushResponse.CommitHash,
			CommitTime:             manifestPushResponse.CommitTime,
			EnvConfigOverrideId:    valuesOverrideResponse.EnvOverride.Id,
			PipelineOverrideValues: valuesOverrideResponse.ReleaseOverrideJSON,
			PipelineId:             overrideRequest.PipelineId,
			CiArtifactId:           overrideRequest.CiArtifactId,
			PipelineMergedValues:   valuesOverrideResponse.MergedValues,
			AuditLog:               sql.AuditLog{UpdatedOn: triggerEvent.TriggerdAt, UpdatedBy: overrideRequest.UserId},
		}
		_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineOverrideRepository.Update")
		err = impl.pipelineOverrideRepository.Update(pipelineOverrideUpdateRequest)
		span.End()
	}

	if triggerEvent.PerformDeploymentOnCluster {
		err = impl.deployApp(overrideRequest, valuesOverrideResponse, triggerEvent.TriggerdAt, ctx)
		if err != nil {
			impl.logger.Errorw("error in deploying app", "err", err)
			return releaseNo, manifest, err
		}
	}

	go impl.writeCDTriggerEvent(overrideRequest, valuesOverrideResponse.Artifact, valuesOverrideResponse.PipelineOverride.PipelineReleaseCounter, valuesOverrideResponse.PipelineOverride.Id)

	_, span := otel.Tracer("orchestrator").Start(ctx, "markImageScanDeployed")
	_ = impl.markImageScanDeployed(overrideRequest.AppId, valuesOverrideResponse.EnvOverride.TargetEnvironment, valuesOverrideResponse.Artifact.ImageDigest, overrideRequest.ClusterId, valuesOverrideResponse.Artifact.ScanEnabled)
	span.End()

	middleware.CdTriggerCounter.WithLabelValues(overrideRequest.AppName, overrideRequest.EnvName).Inc()

	return valuesOverrideResponse.PipelineOverride.PipelineReleaseCounter, manifest, nil

}

func (impl *TriggerServiceImpl) buildManifestPushTemplate(overrideRequest *bean3.ValuesOverrideRequest, valuesOverrideResponse *app.ValuesOverrideResponse, builtChartPath string, manifest *[]byte) (*bean4.ManifestPushTemplate, error) {

	manifestPushTemplate := &bean4.ManifestPushTemplate{
		WorkflowRunnerId:      overrideRequest.WfrId,
		AppId:                 overrideRequest.AppId,
		ChartRefId:            valuesOverrideResponse.EnvOverride.Chart.ChartRefId,
		EnvironmentId:         valuesOverrideResponse.EnvOverride.Environment.Id,
		UserId:                overrideRequest.UserId,
		PipelineOverrideId:    valuesOverrideResponse.PipelineOverride.Id,
		AppName:               overrideRequest.AppName,
		TargetEnvironmentName: valuesOverrideResponse.EnvOverride.TargetEnvironment,
		BuiltChartPath:        builtChartPath,
		BuiltChartBytes:       manifest,
		MergedValues:          valuesOverrideResponse.MergedValues,
	}

	manifestPushConfig, err := impl.manifestPushConfigRepository.GetManifestPushConfigByAppIdAndEnvId(overrideRequest.AppId, overrideRequest.EnvId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching manifest push config from db", "err", err)
		return manifestPushTemplate, err
	}

	if manifestPushConfig != nil {
		if manifestPushConfig.StorageType == bean2.ManifestStorageGit {
			// need to implement for git repo push
			// currently manifest push config doesn't have git push config. Gitops config is derived from charts, chart_env_config_override and chart_ref table
		}
	} else {
		manifestPushTemplate.ChartReferenceTemplate = valuesOverrideResponse.EnvOverride.Chart.ReferenceTemplate
		manifestPushTemplate.ChartName = valuesOverrideResponse.EnvOverride.Chart.ChartName
		manifestPushTemplate.ChartVersion = valuesOverrideResponse.EnvOverride.Chart.ChartVersion
		manifestPushTemplate.ChartLocation = valuesOverrideResponse.EnvOverride.Chart.ChartLocation
		manifestPushTemplate.RepoUrl = valuesOverrideResponse.EnvOverride.Chart.GitRepoUrl
	}
	return manifestPushTemplate, err
}

func (impl *TriggerServiceImpl) getManifestPushService(triggerEvent bean.TriggerEvent) app.ManifestPushService {
	var manifestPushService app.ManifestPushService
	if triggerEvent.ManifestStorageType == bean2.ManifestStorageGit {
		manifestPushService = impl.gitOpsManifestPushService
	}
	return manifestPushService
}

func (impl *TriggerServiceImpl) deployApp(overrideRequest *bean3.ValuesOverrideRequest, valuesOverrideResponse *app.ValuesOverrideResponse,
	triggeredAt time.Time, ctx context.Context) error {

	if util.IsAcdApp(overrideRequest.DeploymentAppType) {
		_, span := otel.Tracer("orchestrator").Start(ctx, "deployArgocdApp")
		err := impl.deployArgocdApp(overrideRequest, valuesOverrideResponse, triggeredAt, ctx)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in deploying app on argocd", "err", err)
			return err
		}
	} else if util.IsHelmApp(overrideRequest.DeploymentAppType) {
		_, span := otel.Tracer("orchestrator").Start(ctx, "createHelmAppForCdPipeline")
		_, err := impl.createHelmAppForCdPipeline(overrideRequest, valuesOverrideResponse, triggeredAt, ctx)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in creating or updating helm application for cd pipeline", "err", err)
			return err
		}
	}
	return nil
}

func (impl *TriggerServiceImpl) createHelmAppForCdPipeline(overrideRequest *bean3.ValuesOverrideRequest, valuesOverrideResponse *app.ValuesOverrideResponse,
	triggeredAt time.Time, ctx context.Context) (bool, error) {

	pipeline := valuesOverrideResponse.Pipeline
	envOverride := valuesOverrideResponse.EnvOverride
	mergeAndSave := valuesOverrideResponse.MergedValues

	chartMetaData := &chart.Metadata{
		Name:    pipeline.App.AppName,
		Version: envOverride.Chart.ChartVersion,
	}
	referenceTemplatePath := path.Join(bean5.RefChartDirPath, envOverride.Chart.ReferenceTemplate)

	if util.IsHelmApp(pipeline.DeploymentAppType) {
		referenceChartByte := envOverride.Chart.ReferenceChart
		// here updating reference chart into database.
		if len(envOverride.Chart.ReferenceChart) == 0 {
			refChartByte, err := impl.chartTemplateService.GetByteArrayRefChart(chartMetaData, referenceTemplatePath)
			if err != nil {
				impl.logger.Errorw("ref chart commit error on cd trigger", "err", err, "req", overrideRequest)
				return false, err
			}
			ch := envOverride.Chart
			ch.ReferenceChart = refChartByte
			ch.UpdatedOn = time.Now()
			ch.UpdatedBy = overrideRequest.UserId
			err = impl.chartRepository.Update(ch)
			if err != nil {
				impl.logger.Errorw("chart update error", "err", err, "req", overrideRequest)
				return false, err
			}
			referenceChartByte = refChartByte
		}

		releaseName := pipeline.DeploymentAppName
		cluster := envOverride.Environment.Cluster
		bearerToken := cluster.Config[util5.BearerToken]
		clusterConfig := &gRPC.ClusterConfig{
			ClusterName:           cluster.ClusterName,
			Token:                 bearerToken,
			ApiServerUrl:          cluster.ServerUrl,
			InsecureSkipTLSVerify: cluster.InsecureSkipTlsVerify,
		}
		if cluster.InsecureSkipTlsVerify == false {
			clusterConfig.KeyData = cluster.Config[util5.TlsKey]
			clusterConfig.CertData = cluster.Config[util5.CertData]
			clusterConfig.CaData = cluster.Config[util5.CertificateAuthorityData]
		}
		releaseIdentifier := &gRPC.ReleaseIdentifier{
			ReleaseName:      releaseName,
			ReleaseNamespace: envOverride.Namespace,
			ClusterConfig:    clusterConfig,
		}

		if pipeline.DeploymentAppCreated {
			req := &gRPC.UpgradeReleaseRequest{
				ReleaseIdentifier: releaseIdentifier,
				ValuesYaml:        mergeAndSave,
				HistoryMax:        impl.helmAppService.GetRevisionHistoryMaxValue(bean6.SOURCE_DEVTRON_APP),
				ChartContent:      &gRPC.ChartContent{Content: referenceChartByte},
			}
			if impl.IsDevtronAsyncInstallModeEnabled(bean.Helm) {
				req.RunInCtx = true
			}
			// For cases where helm release was not found, kubelink will install the same configuration
			updateApplicationResponse, err := impl.helmAppClient.UpdateApplication(ctx, req)
			if err != nil {
				impl.logger.Errorw("error in updating helm application for cd pipeline", "err", err)
				if util.GetGRPCErrorDetailedMessage(err) == context.Canceled.Error() {
					err = errors.New(pipelineConfig.NEW_DEPLOYMENT_INITIATED)
				}
				return false, err
			} else {
				impl.logger.Debugw("updated helm application", "response", updateApplicationResponse, "isSuccess", updateApplicationResponse.Success)
			}

		} else {

			helmResponse, err := impl.helmInstallReleaseWithCustomChart(ctx, releaseIdentifier, referenceChartByte, mergeAndSave)

			// For connection related errors, no need to update the db
			if err != nil && strings.Contains(err.Error(), "connection error") {
				impl.logger.Errorw("error in helm install custom chart", "err", err)
				return false, err
			}
			if util.GetGRPCErrorDetailedMessage(err) == context.Canceled.Error() {
				err = errors.New(pipelineConfig.NEW_DEPLOYMENT_INITIATED)
			}

			// IMP: update cd pipeline to mark deployment app created, even if helm install fails
			// If the helm install fails, it still creates the app in failed state, so trying to
			// re-create the app results in error from helm that cannot re-use name which is still in use
			_, pgErr := impl.updatePipeline(pipeline, overrideRequest.UserId)

			if err != nil {
				impl.logger.Errorw("error in helm install custom chart", "err", err)

				if pgErr != nil {
					impl.logger.Errorw("failed to update deployment app created flag in pipeline table", "err", err)
				}
				return false, err
			}

			if pgErr != nil {
				impl.logger.Errorw("failed to update deployment app created flag in pipeline table", "err", err)
				return false, err
			}

			impl.logger.Debugw("received helm release response", "helmResponse", helmResponse, "isSuccess", helmResponse.Success)
		}

		//update workflow runner status, used in app workflow view
		err := impl.cdWorkflowCommonService.UpdateCDWorkflowRunnerStatus(ctx, overrideRequest, triggeredAt, pipelineConfig.WorkflowInProgress, "")
		if err != nil {
			impl.logger.Errorw("error in updating the workflow runner status, createHelmAppForCdPipeline", "err", err)
			return false, err
		}
	}
	return true, nil
}

func (impl *TriggerServiceImpl) deployArgocdApp(overrideRequest *bean3.ValuesOverrideRequest, valuesOverrideResponse *app.ValuesOverrideResponse, triggeredAt time.Time, ctx context.Context) error {

	impl.logger.Debugw("new pipeline found", "pipeline", valuesOverrideResponse.Pipeline)
	_, span := otel.Tracer("orchestrator").Start(ctx, "createArgoApplicationIfRequired")
	name, err := impl.createArgoApplicationIfRequired(overrideRequest.AppId, valuesOverrideResponse.EnvOverride, valuesOverrideResponse.Pipeline, overrideRequest.UserId)
	span.End()
	if err != nil {
		impl.logger.Errorw("acd application create error on cd trigger", "err", err, "req", overrideRequest)
		return err
	}
	impl.logger.Debugw("argocd application created", "name", name)

	_, span = otel.Tracer("orchestrator").Start(ctx, "updateArgoPipeline")
	updateAppInArgocd, err := impl.updateArgoPipeline(valuesOverrideResponse.Pipeline, valuesOverrideResponse.EnvOverride, ctx)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in updating argocd app ", "err", err)
		return err
	}
	syncTime := time.Now()
	err = impl.argoClientWrapperService.SyncArgoCDApplicationIfNeededAndRefresh(ctx, valuesOverrideResponse.Pipeline.DeploymentAppName)
	if err != nil {
		impl.logger.Errorw("error in getting argo application with normal refresh", "argoAppName", valuesOverrideResponse.Pipeline.DeploymentAppName)
		return fmt.Errorf("%s. err: %s", bean.ARGOCD_SYNC_ERROR, err.Error())
	}
	if !impl.ACDConfig.ArgoCDAutoSyncEnabled {
		timeline := &pipelineConfig.PipelineStatusTimeline{
			CdWorkflowRunnerId: overrideRequest.WfrId,
			StatusTime:         syncTime,
			AuditLog: sql.AuditLog{
				CreatedBy: 1,
				CreatedOn: time.Now(),
				UpdatedBy: 1,
				UpdatedOn: time.Now(),
			},
			Status:       pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED,
			StatusDetail: "argocd sync completed",
		}
		_, err, _ = impl.pipelineStatusTimelineService.SavePipelineStatusTimelineIfNotAlreadyPresent(overrideRequest.WfrId, timeline.Status, timeline, false)
		if err != nil {
			impl.logger.Errorw("error in saving pipeline status timeline", "err", err)
		}
	}
	if updateAppInArgocd {
		impl.logger.Debug("argo-cd successfully updated")
	} else {
		impl.logger.Debug("argo-cd failed to update, ignoring it")
	}
	return nil
}

// update repoUrl, revision and argo app sync mode (auto/manual) if needed
func (impl *TriggerServiceImpl) updateArgoPipeline(pipeline *pipelineConfig.Pipeline, envOverride *chartConfig.EnvConfigOverride, ctx context.Context) (bool, error) {
	if ctx == nil {
		impl.logger.Errorw("err in syncing ACD, ctx is NULL", "pipelineName", pipeline.Name)
		return false, nil
	}
	argoAppName := pipeline.DeploymentAppName
	impl.logger.Infow("received payload, updateArgoPipeline", "appId", pipeline.AppId, "pipelineName", pipeline.Name, "envId", envOverride.TargetEnvironment, "argoAppName", argoAppName, "context", ctx)
	argoApplication, err := impl.argoClientWrapperService.GetArgoAppByName(ctx, argoAppName)
	if err != nil {
		impl.logger.Errorw("no argo app exists", "app", argoAppName, "pipeline", pipeline.Name)
		return false, err
	}
	//if status, ok:=status.FromError(err);ok{
	appStatus, _ := status2.FromError(err)
	if appStatus.Code() == codes.OK {
		impl.logger.Debugw("argo app exists", "app", argoAppName, "pipeline", pipeline.Name)
		if argoApplication.Spec.Source.Path != envOverride.Chart.ChartLocation || argoApplication.Spec.Source.TargetRevision != "master" {
			patchRequestDto := &bean7.ArgoCdAppPatchReqDto{
				ArgoAppName:    argoAppName,
				ChartLocation:  envOverride.Chart.ChartLocation,
				GitRepoUrl:     envOverride.Chart.GitRepoUrl,
				TargetRevision: bean7.TargetRevisionMaster,
				PatchType:      bean7.PatchTypeMerge,
			}
			err = impl.argoClientWrapperService.PatchArgoCdApp(ctx, patchRequestDto)
			if err != nil {
				impl.logger.Errorw("error in patching argo pipeline", "err", err, "req", patchRequestDto)
				return false, err
			}
			impl.logger.Debugw("pipeline update req", "res", patchRequestDto)
		} else {
			impl.logger.Debug("pipeline no need to update ")
		}
		err := impl.argoClientWrapperService.UpdateArgoCDSyncModeIfNeeded(ctx, argoApplication)
		if err != nil {
			impl.logger.Errorw("error in updating argocd sync mode", "err", err)
			return false, err
		}
		return true, nil
	} else if appStatus.Code() == codes.NotFound {
		impl.logger.Errorw("argo app not found", "app", argoAppName, "pipeline", pipeline.Name)
		return false, nil
	} else {
		impl.logger.Errorw("err in checking application on argoCD", "err", err, "pipeline", pipeline.Name)
		return false, err
	}
}

func (impl *TriggerServiceImpl) createArgoApplicationIfRequired(appId int, envConfigOverride *chartConfig.EnvConfigOverride, pipeline *pipelineConfig.Pipeline, userId int32) (string, error) {
	//repo has been registered while helm create
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("no chart found ", "app", appId)
		return "", err
	}
	envModel, err := impl.envRepository.FindById(envConfigOverride.TargetEnvironment)
	if err != nil {
		return "", err
	}
	argoAppName := pipeline.DeploymentAppName
	if pipeline.DeploymentAppCreated {
		return argoAppName, nil
	} else {
		//create
		appNamespace := envConfigOverride.Namespace
		if appNamespace == "" {
			appNamespace = "default"
		}
		namespace := argocdServer.DevtronInstalationNs

		appRequest := &argocdServer.AppTemplate{
			ApplicationName: argoAppName,
			Namespace:       namespace,
			TargetNamespace: appNamespace,
			TargetServer:    envModel.Cluster.ServerUrl,
			Project:         "default",
			ValuesFile:      getValuesFileForEnv(envModel.Id),
			RepoPath:        chart.ChartLocation,
			RepoUrl:         chart.GitRepoUrl,
			AutoSyncEnabled: impl.ACDConfig.ArgoCDAutoSyncEnabled,
		}
		argoAppName, err := impl.argoK8sClient.CreateAcdApp(appRequest, envModel.Cluster, argocdServer.ARGOCD_APPLICATION_TEMPLATE)
		if err != nil {
			return "", err
		}
		//update cd pipeline to mark deployment app created
		_, err = impl.updatePipeline(pipeline, userId)
		if err != nil {
			impl.logger.Errorw("error in update cd pipeline for deployment app created or not", "err", err)
			return "", err
		}
		return argoAppName, nil
	}
}

func getValuesFileForEnv(environmentId int) string {
	return fmt.Sprintf("_%d-values.yaml", environmentId) //-{envId}-values.yaml
}

func (impl *TriggerServiceImpl) updatePipeline(pipeline *pipelineConfig.Pipeline, userId int32) (bool, error) {
	err := impl.pipelineRepository.SetDeploymentAppCreatedInPipeline(true, pipeline.Id, userId)
	if err != nil {
		impl.logger.Errorw("error on updating cd pipeline for setting deployment app created", "err", err)
		return false, err
	}
	return true, nil
}

// helmInstallReleaseWithCustomChart performs helm install with custom chart
func (impl *TriggerServiceImpl) helmInstallReleaseWithCustomChart(ctx context.Context, releaseIdentifier *gRPC.ReleaseIdentifier, referenceChartByte []byte, valuesYaml string) (*gRPC.HelmInstallCustomResponse, error) {

	helmInstallRequest := gRPC.HelmInstallCustomRequest{
		ValuesYaml:        valuesYaml,
		ChartContent:      &gRPC.ChartContent{Content: referenceChartByte},
		ReleaseIdentifier: releaseIdentifier,
	}
	if impl.IsDevtronAsyncInstallModeEnabled(bean.Helm) {
		helmInstallRequest.RunInCtx = true
	}
	// Request exec
	return impl.helmAppClient.InstallReleaseWithCustomChart(ctx, &helmInstallRequest)
}

func (impl *TriggerServiceImpl) writeCDTriggerEvent(overrideRequest *bean3.ValuesOverrideRequest, artifact *repository3.CiArtifact, releaseId, pipelineOverrideId int) {

	event := impl.eventFactory.Build(util2.Trigger, &overrideRequest.PipelineId, overrideRequest.AppId, &overrideRequest.EnvId, util2.CD)
	impl.logger.Debugw("event writeCDTriggerEvent", "event", event)
	event = impl.eventFactory.BuildExtraCDData(event, nil, pipelineOverrideId, bean3.CD_WORKFLOW_TYPE_DEPLOY)
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("CD trigger event not sent", "error", evtErr)
	}
	deploymentEvent := app.DeploymentEvent{
		ApplicationId:      overrideRequest.AppId,
		EnvironmentId:      overrideRequest.EnvId, //check for production Environment
		ReleaseId:          releaseId,
		PipelineOverrideId: pipelineOverrideId,
		TriggerTime:        time.Now(),
		CiArtifactId:       overrideRequest.CiArtifactId,
	}
	ciPipelineMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(artifact.PipelineId)
	if err != nil {
		impl.logger.Errorw("error in ")
	}
	materialInfoMap, mErr := artifact.ParseMaterialInfo()
	if mErr != nil {
		impl.logger.Errorw("material info map error", mErr)
		return
	}
	for _, ciPipelineMaterial := range ciPipelineMaterials {
		hash := materialInfoMap[ciPipelineMaterial.GitMaterial.Url]
		pipelineMaterialInfo := &app.PipelineMaterialInfo{PipelineMaterialId: ciPipelineMaterial.Id, CommitHash: hash}
		deploymentEvent.PipelineMaterials = append(deploymentEvent.PipelineMaterials, pipelineMaterialInfo)
	}
	impl.logger.Infow("triggering deployment event", "event", deploymentEvent)
	err = impl.eventClient.WriteNatsEvent(pubsub.CD_SUCCESS, deploymentEvent)
	if err != nil {
		impl.logger.Errorw("error in writing cd trigger event", "err", err)
	}
}

func (impl *TriggerServiceImpl) markImageScanDeployed(appId int, envId int, imageDigest string, clusterId int, isScanEnabled bool) error {
	impl.logger.Debugw("mark image scan deployed for normal app, from cd auto or manual trigger", "imageDigest", imageDigest)
	executionHistory, err := impl.imageScanHistoryRepository.FindByImageDigest(imageDigest)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching execution history", "err", err)
		return err
	}
	if executionHistory == nil || executionHistory.Id == 0 {
		impl.logger.Errorw("no execution history found for digest", "digest", imageDigest)
		return fmt.Errorf("no execution history found for digest - %s", imageDigest)
	}
	impl.logger.Debugw("mark image scan deployed for normal app, from cd auto or manual trigger", "executionHistory", executionHistory)
	var ids []int
	ids = append(ids, executionHistory.Id)

	ot, err := impl.imageScanDeployInfoRepository.FetchByAppIdAndEnvId(appId, envId, []string{security.ScanObjectType_APP})

	if err == pg.ErrNoRows && !isScanEnabled {
		//ignoring if no rows are found and scan is disabled
		return nil
	}

	if err != nil && err != pg.ErrNoRows {
		return err
	} else if err == pg.ErrNoRows && isScanEnabled {
		imageScanDeployInfo := &security.ImageScanDeployInfo{
			ImageScanExecutionHistoryId: ids,
			ScanObjectMetaId:            appId,
			ObjectType:                  security.ScanObjectType_APP,
			EnvId:                       envId,
			ClusterId:                   clusterId,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: 1,
				UpdatedOn: time.Now(),
				UpdatedBy: 1,
			},
		}
		impl.logger.Debugw("mark image scan deployed for normal app, from cd auto or manual trigger", "imageScanDeployInfo", imageScanDeployInfo)
		err = impl.imageScanDeployInfoRepository.Save(imageScanDeployInfo)
		if err != nil {
			impl.logger.Errorw("error in creating deploy info", "err", err)
		}
	} else {
		// Updating Execution history for Latest Deployment to fetch out security Vulnerabilities for latest deployed info
		if isScanEnabled {
			ot.ImageScanExecutionHistoryId = ids
		} else {
			arr := []int{-1}
			ot.ImageScanExecutionHistoryId = arr
		}
		err = impl.imageScanDeployInfoRepository.Update(ot)
		if err != nil {
			impl.logger.Errorw("error in updating deploy info for latest deployed image", "err", err)
		}
	}
	return err
}

func (impl *TriggerServiceImpl) IsDevtronAsyncInstallModeEnabled(deploymentAppType string) bool {
	return impl.globalEnvVariables.EnableAsyncInstallDevtronChart &&
		deploymentAppType == bean.Helm
}

func (impl *TriggerServiceImpl) deleteCorruptedPipelineStage(pipelineStage *repository.PipelineStage, triggeredBy int32) (error, bool) {
	if pipelineStage != nil {
		stageReq := &bean8.PipelineStageDto{
			Id:   pipelineStage.Id,
			Type: pipelineStage.Type,
		}
		err, deleted := impl.pipelineStageService.DeletePipelineStageIfReq(stageReq, triggeredBy)
		if err != nil {
			impl.logger.Errorw("error in deleting the corrupted pipeline stage", "err", err, "pipelineStageReq", stageReq)
			return err, false
		}
		return nil, deleted
	}
	return nil, false
}
