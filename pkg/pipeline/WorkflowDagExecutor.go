/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package pipeline

import (
	"context"
	"encoding/json"
	errors3 "errors"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	util5 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/common-lib/utils/k8s/health"
	client2 "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/client/argocdServer"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	gitSensorClient "github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/middleware"
	app2 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	bean4 "github.com/devtron-labs/devtron/pkg/app/bean"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/dockerRegistry"
	"github.com/devtron-labs/devtron/pkg/k8s"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	repository5 "github.com/devtron-labs/devtron/pkg/variables/repository"
	util4 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	errors2 "github.com/juju/errors"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	status2 "google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	history2 "github.com/devtron-labs/devtron/pkg/pipeline/history"
	repository3 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util3 "github.com/devtron-labs/devtron/pkg/util"

	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/user"
	util2 "github.com/devtron-labs/devtron/util/event"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type WorkflowDagExecutor interface {
	HandleCiSuccessEvent(artifact *repository.CiArtifact, applyAuth bool, async bool, triggeredBy int32) error
	HandleWebhookExternalCiEvent(artifact *repository.CiArtifact, triggeredBy int32, externalCiId int, auth func(email string, projectObject string, envObject string) bool) (bool, error)
	HandlePreStageSuccessEvent(cdStageCompleteEvent CdStageCompleteEvent) error
	HandleDeploymentSuccessEvent(pipelineOverride *chartConfig.PipelineOverride) error
	HandlePostStageSuccessEvent(cdWorkflowId int, cdPipelineId int, triggeredBy int32) error
	Subscribe() error
	TriggerPostStage(cdWf *pipelineConfig.CdWorkflow, cdPipeline *pipelineConfig.Pipeline, triggeredBy int32, refCdWorkflowRunnerId int) error
	TriggerPreStage(ctx context.Context, cdWf *pipelineConfig.CdWorkflow, artifact *repository.CiArtifact, pipeline *pipelineConfig.Pipeline, triggeredBy int32, applyAuth bool, refCdWorkflowRunnerId int) error
	TriggerDeployment(cdWf *pipelineConfig.CdWorkflow, artifact *repository.CiArtifact, pipeline *pipelineConfig.Pipeline, applyAuth bool, triggeredBy int32) error
	ManualCdTrigger(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context) (int, error)
	TriggerBulkDeploymentAsync(requests []*BulkTriggerRequest, UserId int32) (interface{}, error)
	StopStartApp(stopRequest *StopAppRequest, ctx context.Context) (int, error)
	TriggerBulkHibernateAsync(request StopDeploymentGroupRequest, ctx context.Context) (interface{}, error)
	RotatePods(ctx context.Context, podRotateRequest *PodRotateRequest) (*k8s.RotatePodResponse, error)
}

type WorkflowDagExecutorImpl struct {
	logger                        *zap.SugaredLogger
	pipelineRepository            pipelineConfig.PipelineRepository
	cdWorkflowRepository          pipelineConfig.CdWorkflowRepository
	pubsubClient                  *pubsub.PubSubClientServiceImpl
	appService                    app.AppService
	cdWorkflowService             WorkflowService
	ciPipelineRepository          pipelineConfig.CiPipelineRepository
	materialRepository            pipelineConfig.MaterialRepository
	pipelineOverrideRepository    chartConfig.PipelineOverrideRepository
	ciArtifactRepository          repository.CiArtifactRepository
	user                          user.UserService
	enforcer                      casbin.Enforcer
	enforcerUtil                  rbac.EnforcerUtil
	groupRepository               repository.DeploymentGroupRepository
	tokenCache                    *util3.TokenCache
	acdAuthConfig                 *util3.ACDAuthConfig
	envRepository                 repository2.EnvironmentRepository
	eventFactory                  client.EventFactory
	eventClient                   client.EventClient
	cvePolicyRepository           security.CvePolicyRepository
	scanResultRepository          security.ImageScanResultRepository
	appWorkflowRepository         appWorkflow.AppWorkflowRepository
	prePostCdScriptHistoryService history2.PrePostCdScriptHistoryService
	argoUserService               argo.ArgoUserService
	cdPipelineStatusTimelineRepo  pipelineConfig.PipelineStatusTimelineRepository
	pipelineStatusTimelineService status.PipelineStatusTimelineService
	CiTemplateRepository          pipelineConfig.CiTemplateRepository
	ciWorkflowRepository          pipelineConfig.CiWorkflowRepository
	appLabelRepository            pipelineConfig.AppLabelRepository
	gitSensorGrpcClient           gitSensorClient.Client
	k8sCommonService              k8s.K8sCommonService
	pipelineStageRepository       repository4.PipelineStageRepository
	pipelineStageService          PipelineStageService
	config                        *CdConfig

	variableSnapshotHistoryService variables.VariableSnapshotHistoryService

	deploymentTemplateHistoryService    history2.DeploymentTemplateHistoryService
	configMapHistoryService             history2.ConfigMapHistoryService
	pipelineStrategyHistoryService      history2.PipelineStrategyHistoryService
	manifestPushConfigRepository        repository4.ManifestPushConfigRepository
	gitOpsManifestPushService           app.GitOpsPushService
	ciPipelineMaterialRepository        pipelineConfig.CiPipelineMaterialRepository
	imageScanHistoryRepository          security.ImageScanHistoryRepository
	imageScanDeployInfoRepository       security.ImageScanDeployInfoRepository
	appCrudOperationService             app.AppCrudOperationService
	pipelineConfigRepository            chartConfig.PipelineConfigRepository
	dockerRegistryIpsConfigService      dockerRegistry.DockerRegistryIpsConfigService
	chartRepository                     chartRepoRepository.ChartRepository
	chartTemplateService                util.ChartTemplateService
	strategyHistoryRepository           repository3.PipelineStrategyHistoryRepository
	appRepository                       app2.AppRepository
	deploymentTemplateHistoryRepository repository3.DeploymentTemplateHistoryRepository
	argoK8sClient                       argocdServer.ArgoK8sClient
	configMapRepository                 chartConfig.ConfigMapRepository
	configMapHistoryRepository          repository3.ConfigMapHistoryRepository
	refChartDir                         chartRepoRepository.RefChartDir
	helmAppService                      client2.HelmAppService
	helmAppClient                       client2.HelmAppClient
	chartRefRepository                  chartRepoRepository.ChartRefRepository
	environmentConfigRepository         chartConfig.EnvConfigOverrideRepository
	appLevelMetricsRepository           repository.AppLevelMetricsRepository
	envLevelMetricsRepository           repository.EnvLevelAppMetricsRepository
	dbMigrationConfigRepository         pipelineConfig.DbMigrationConfigRepository
	mergeUtil                           *util.MergeUtil
	gitOpsConfigRepository              repository.GitOpsConfigRepository
	gitFactory                          *util.GitFactory
	acdClient                           application2.ServiceClient
	variableEntityMappingService        variables.VariableEntityMappingService
	variableTemplateParser              parsers.VariableTemplateParser
	argoClientWrapperService            argocdServer.ArgoClientWrapperService
	scopedVariableService               variables.ScopedVariableService
}

const kedaAutoscaling = "kedaAutoscaling"
const horizontalPodAutoscaler = "HorizontalPodAutoscaler"
const fullnameOverride = "fullnameOverride"
const nameOverride = "nameOverride"
const enabled = "enabled"
const replicaCount = "replicaCount"

const (
	CD_PIPELINE_ENV_NAME_KEY     = "CD_PIPELINE_ENV_NAME"
	CD_PIPELINE_CLUSTER_NAME_KEY = "CD_PIPELINE_CLUSTER_NAME"
	GIT_COMMIT_HASH_PREFIX       = "GIT_COMMIT_HASH"
	GIT_SOURCE_TYPE_PREFIX       = "GIT_SOURCE_TYPE"
	GIT_SOURCE_VALUE_PREFIX      = "GIT_SOURCE_VALUE"
	GIT_METADATA                 = "GIT_METADATA"
	GIT_SOURCE_COUNT             = "GIT_SOURCE_COUNT"
	APP_LABEL_KEY_PREFIX         = "APP_LABEL_KEY"
	APP_LABEL_VALUE_PREFIX       = "APP_LABEL_VALUE"
	APP_LABEL_METADATA           = "APP_LABEL_METADATA"
	APP_LABEL_COUNT              = "APP_LABEL_COUNT"
	CHILD_CD_ENV_NAME_PREFIX     = "CHILD_CD_ENV_NAME"
	CHILD_CD_CLUSTER_NAME_PREFIX = "CHILD_CD_CLUSTER_NAME"
	CHILD_CD_METADATA            = "CHILD_CD_METADATA"
	CHILD_CD_COUNT               = "CHILD_CD_COUNT"
	DOCKER_IMAGE                 = "DOCKER_IMAGE"
	DEPLOYMENT_RELEASE_ID        = "DEPLOYMENT_RELEASE_ID"
	DEPLOYMENT_UNIQUE_ID         = "DEPLOYMENT_UNIQUE_ID"
	CD_TRIGGERED_BY              = "CD_TRIGGERED_BY"
	CD_TRIGGER_TIME              = "CD_TRIGGER_TIME"
	APP_NAME                     = "APP_NAME"
	DEVTRON_CD_TRIGGERED_BY      = "DEVTRON_CD_TRIGGERED_BY"
	DEVTRON_CD_TRIGGER_TIME      = "DEVTRON_CD_TRIGGER_TIME"
)

type CiArtifactDTO struct {
	Id                   int    `json:"id"`
	PipelineId           int    `json:"pipelineId"` //id of the ci pipeline from which this webhook was triggered
	Image                string `json:"image"`
	ImageDigest          string `json:"imageDigest"`
	MaterialInfo         string `json:"materialInfo"` //git material metadata json array string
	DataSource           string `json:"dataSource"`
	WorkflowId           *int   `json:"workflowId"`
	ciArtifactRepository repository.CiArtifactRepository
}

type CdStageCompleteEvent struct {
	CiProjectDetails []bean3.CiProjectDetails     `json:"ciProjectDetails"`
	WorkflowId       int                          `json:"workflowId"`
	WorkflowRunnerId int                          `json:"workflowRunnerId"`
	CdPipelineId     int                          `json:"cdPipelineId"`
	TriggeredBy      int32                        `json:"triggeredBy"`
	StageYaml        string                       `json:"stageYaml"`
	ArtifactLocation string                       `json:"artifactLocation"`
	PipelineName     string                       `json:"pipelineName"`
	CiArtifactDTO    pipelineConfig.CiArtifactDTO `json:"ciArtifactDTO"`
}

type GitMetadata struct {
	GitCommitHash  string `json:"GIT_COMMIT_HASH"`
	GitSourceType  string `json:"GIT_SOURCE_TYPE"`
	GitSourceValue string `json:"GIT_SOURCE_VALUE"`
}

type AppLabelMetadata struct {
	AppLabelKey   string `json:"APP_LABEL_KEY"`
	AppLabelValue string `json:"APP_LABEL_VALUE"`
}

type ChildCdMetadata struct {
	ChildCdEnvName     string `json:"CHILD_CD_ENV_NAME"`
	ChildCdClusterName string `json:"CHILD_CD_CLUSTER_NAME"`
}

func NewWorkflowDagExecutorImpl(Logger *zap.SugaredLogger, pipelineRepository pipelineConfig.PipelineRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pubsubClient *pubsub.PubSubClientServiceImpl,
	appService app.AppService,
	cdWorkflowService WorkflowService,
	ciArtifactRepository repository.CiArtifactRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	materialRepository pipelineConfig.MaterialRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
	user user.UserService,
	groupRepository repository.DeploymentGroupRepository,
	envRepository repository2.EnvironmentRepository,
	enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil, tokenCache *util3.TokenCache,
	acdAuthConfig *util3.ACDAuthConfig, eventFactory client.EventFactory,
	eventClient client.EventClient, cvePolicyRepository security.CvePolicyRepository,
	scanResultRepository security.ImageScanResultRepository,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	prePostCdScriptHistoryService history2.PrePostCdScriptHistoryService,
	argoUserService argo.ArgoUserService,
	cdPipelineStatusTimelineRepo pipelineConfig.PipelineStatusTimelineRepository,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	CiTemplateRepository pipelineConfig.CiTemplateRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	appLabelRepository pipelineConfig.AppLabelRepository, gitSensorGrpcClient gitSensorClient.Client,
	pipelineStageService PipelineStageService, k8sCommonService k8s.K8sCommonService,
	variableSnapshotHistoryService variables.VariableSnapshotHistoryService,

	deploymentTemplateHistoryService history2.DeploymentTemplateHistoryService,
	configMapHistoryService history2.ConfigMapHistoryService,
	pipelineStrategyHistoryService history2.PipelineStrategyHistoryService,
	manifestPushConfigRepository repository4.ManifestPushConfigRepository,
	gitOpsManifestPushService app.GitOpsPushService,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	imageScanHistoryRepository security.ImageScanHistoryRepository,
	imageScanDeployInfoRepository security.ImageScanDeployInfoRepository,
	appCrudOperationService app.AppCrudOperationService,
	pipelineConfigRepository chartConfig.PipelineConfigRepository,
	dockerRegistryIpsConfigService dockerRegistry.DockerRegistryIpsConfigService,
	chartRepository chartRepoRepository.ChartRepository,
	chartTemplateService util.ChartTemplateService,
	strategyHistoryRepository repository3.PipelineStrategyHistoryRepository,
	appRepository app2.AppRepository,
	deploymentTemplateHistoryRepository repository3.DeploymentTemplateHistoryRepository,
	ArgoK8sClient argocdServer.ArgoK8sClient,
	configMapRepository chartConfig.ConfigMapRepository,
	configMapHistoryRepository repository3.ConfigMapHistoryRepository,
	refChartDir chartRepoRepository.RefChartDir,
	helmAppService client2.HelmAppService,
	helmAppClient client2.HelmAppClient,
	chartRefRepository chartRepoRepository.ChartRefRepository,
	environmentConfigRepository chartConfig.EnvConfigOverrideRepository,
	appLevelMetricsRepository repository.AppLevelMetricsRepository,
	envLevelMetricsRepository repository.EnvLevelAppMetricsRepository,
	dbMigrationConfigRepository pipelineConfig.DbMigrationConfigRepository,
	mergeUtil *util.MergeUtil,
	gitOpsConfigRepository repository.GitOpsConfigRepository,
	gitFactory *util.GitFactory,
	acdClient application2.ServiceClient,
	variableEntityMappingService variables.VariableEntityMappingService,
	variableTemplateParser parsers.VariableTemplateParser,
	argoClientWrapperService argocdServer.ArgoClientWrapperService,
	scopedVariableService variables.ScopedVariableService,
) *WorkflowDagExecutorImpl {
	wde := &WorkflowDagExecutorImpl{logger: Logger,
		pipelineRepository:             pipelineRepository,
		cdWorkflowRepository:           cdWorkflowRepository,
		pubsubClient:                   pubsubClient,
		appService:                     appService,
		cdWorkflowService:              cdWorkflowService,
		ciPipelineRepository:           ciPipelineRepository,
		ciArtifactRepository:           ciArtifactRepository,
		materialRepository:             materialRepository,
		pipelineOverrideRepository:     pipelineOverrideRepository,
		user:                           user,
		enforcer:                       enforcer,
		enforcerUtil:                   enforcerUtil,
		groupRepository:                groupRepository,
		tokenCache:                     tokenCache,
		acdAuthConfig:                  acdAuthConfig,
		envRepository:                  envRepository,
		eventFactory:                   eventFactory,
		eventClient:                    eventClient,
		cvePolicyRepository:            cvePolicyRepository,
		scanResultRepository:           scanResultRepository,
		appWorkflowRepository:          appWorkflowRepository,
		prePostCdScriptHistoryService:  prePostCdScriptHistoryService,
		argoUserService:                argoUserService,
		cdPipelineStatusTimelineRepo:   cdPipelineStatusTimelineRepo,
		pipelineStatusTimelineService:  pipelineStatusTimelineService,
		CiTemplateRepository:           CiTemplateRepository,
		ciWorkflowRepository:           ciWorkflowRepository,
		appLabelRepository:             appLabelRepository,
		gitSensorGrpcClient:            gitSensorGrpcClient,
		k8sCommonService:               k8sCommonService,
		pipelineStageService:           pipelineStageService,
		variableSnapshotHistoryService: variableSnapshotHistoryService,

		deploymentTemplateHistoryService:    deploymentTemplateHistoryService,
		configMapHistoryService:             configMapHistoryService,
		pipelineStrategyHistoryService:      pipelineStrategyHistoryService,
		manifestPushConfigRepository:        manifestPushConfigRepository,
		gitOpsManifestPushService:           gitOpsManifestPushService,
		ciPipelineMaterialRepository:        ciPipelineMaterialRepository,
		imageScanHistoryRepository:          imageScanHistoryRepository,
		imageScanDeployInfoRepository:       imageScanDeployInfoRepository,
		appCrudOperationService:             appCrudOperationService,
		pipelineConfigRepository:            pipelineConfigRepository,
		dockerRegistryIpsConfigService:      dockerRegistryIpsConfigService,
		chartRepository:                     chartRepository,
		chartTemplateService:                chartTemplateService,
		strategyHistoryRepository:           strategyHistoryRepository,
		appRepository:                       appRepository,
		deploymentTemplateHistoryRepository: deploymentTemplateHistoryRepository,
		argoK8sClient:                       ArgoK8sClient,
		configMapRepository:                 configMapRepository,
		configMapHistoryRepository:          configMapHistoryRepository,
		refChartDir:                         refChartDir,
		helmAppService:                      helmAppService,
		helmAppClient:                       helmAppClient,
		chartRefRepository:                  chartRefRepository,
		environmentConfigRepository:         environmentConfigRepository,
		appLevelMetricsRepository:           appLevelMetricsRepository,
		envLevelMetricsRepository:           envLevelMetricsRepository,
		dbMigrationConfigRepository:         dbMigrationConfigRepository,
		mergeUtil:                           mergeUtil,
		gitOpsConfigRepository:              gitOpsConfigRepository,
		gitFactory:                          gitFactory,
		acdClient:                           acdClient,
		variableEntityMappingService:        variableEntityMappingService,
		variableTemplateParser:              variableTemplateParser,
		argoClientWrapperService:            argoClientWrapperService,
		scopedVariableService:               scopedVariableService,
	}
	config, err := GetCdConfig()
	if err != nil {
		return nil
	}
	wde.config = config
	err = wde.Subscribe()
	if err != nil {
		return nil
	}
	err = wde.subscribeTriggerBulkAction()
	if err != nil {
		return nil
	}
	err = wde.subscribeHibernateBulkAction()
	if err != nil {
		return nil
	}
	return wde
}

func (impl *WorkflowDagExecutorImpl) Subscribe() error {
	callback := func(msg *pubsub.PubSubMsg) {
		impl.logger.Debug("cd stage event received")
		//defer msg.Ack()
		cdStageCompleteEvent := CdStageCompleteEvent{}
		err := json.Unmarshal([]byte(string(msg.Data)), &cdStageCompleteEvent)
		if err != nil {
			impl.logger.Errorw("error while unmarshalling cdStageCompleteEvent object", "err", err, "msg", string(msg.Data))
			return
		}
		impl.logger.Debugw("cd stage event:", "workflowRunnerId", cdStageCompleteEvent.WorkflowRunnerId)
		wf, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(cdStageCompleteEvent.WorkflowRunnerId)
		if err != nil {
			impl.logger.Errorw("could not get wf runner", "err", err)
			return
		}
		if wf.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
			impl.logger.Debugw("received pre stage success event for workflow runner ", "wfId", strconv.Itoa(wf.Id))
			err = impl.HandlePreStageSuccessEvent(cdStageCompleteEvent)
			if err != nil {
				impl.logger.Errorw("deployment success event error", "err", err)
				return
			}
		} else if wf.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
			impl.logger.Debugw("received post stage success event for workflow runner ", "wfId", strconv.Itoa(wf.Id))
			err = impl.HandlePostStageSuccessEvent(wf.CdWorkflowId, cdStageCompleteEvent.CdPipelineId, cdStageCompleteEvent.TriggeredBy)
			if err != nil {
				impl.logger.Errorw("deployment success event error", "err", err)
				return
			}
		}
	}
	err := impl.pubsubClient.Subscribe(pubsub.CD_STAGE_COMPLETE_TOPIC, callback)
	if err != nil {
		impl.logger.Error("error", "err", err)
		return err
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) HandleCiSuccessEvent(artifact *repository.CiArtifact, applyAuth bool, async bool, triggeredBy int32) error {
	//1. get cd pipelines
	//2. get config
	//3. trigger wf/ deployment
	pipelines, err := impl.pipelineRepository.FindByParentCiPipelineId(artifact.PipelineId)
	if err != nil {
		impl.logger.Errorw("error in fetching cd pipeline", "pipelineId", artifact.PipelineId, "err", err)
		return err
	}
	for _, pipeline := range pipelines {
		err = impl.triggerStage(nil, pipeline, artifact, applyAuth, triggeredBy)
		if err != nil {
			impl.logger.Debugw("error on trigger cd pipeline", "err", err)
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) HandleWebhookExternalCiEvent(artifact *repository.CiArtifact, triggeredBy int32, externalCiId int, auth func(email string, projectObject string, envObject string) bool) (bool, error) {
	hasAnyTriggered := false
	appWorkflowMappings, err := impl.appWorkflowRepository.FindWFCDMappingByExternalCiId(externalCiId)
	if err != nil {
		impl.logger.Errorw("error in fetching cd pipeline", "pipelineId", artifact.PipelineId, "err", err)
		return hasAnyTriggered, err
	}
	user, err := impl.user.GetById(triggeredBy)
	if err != nil {
		return hasAnyTriggered, err
	}

	var pipelines []*pipelineConfig.Pipeline
	for _, appWorkflowMapping := range appWorkflowMappings {
		pipeline, err := impl.pipelineRepository.FindById(appWorkflowMapping.ComponentId)
		if err != nil {
			impl.logger.Errorw("error in fetching cd pipeline", "pipelineId", artifact.PipelineId, "err", err)
			return hasAnyTriggered, err
		}
		projectObject := impl.enforcerUtil.GetAppRBACNameByAppId(pipeline.AppId)
		envObject := impl.enforcerUtil.GetAppRBACByAppIdAndPipelineId(pipeline.AppId, pipeline.Id)
		if !auth(user.EmailId, projectObject, envObject) {
			err = &util.ApiError{Code: "401", HttpStatusCode: 401, UserMessage: "Unauthorized"}
			return hasAnyTriggered, err
		}
		if pipeline.TriggerType == pipelineConfig.TRIGGER_TYPE_MANUAL {
			impl.logger.Warnw("skipping deployment for manual trigger for webhook", "pipeline", pipeline)
			continue
		}
		pipelines = append(pipelines, pipeline)
	}

	for _, pipeline := range pipelines {
		//applyAuth=false, already auth applied for this flow
		err = impl.triggerStage(nil, pipeline, artifact, false, triggeredBy)
		if err != nil {
			impl.logger.Debugw("error on trigger cd pipeline", "err", err)
			return hasAnyTriggered, err
		}
		hasAnyTriggered = true
	}

	return hasAnyTriggered, err
}

// if stage is present with 0 stage steps, delete the stage
// handle corrupt data (https://github.com/devtron-labs/devtron/issues/3826)
func (impl *WorkflowDagExecutorImpl) deleteCorruptedPipelineStage(pipelineStage *repository4.PipelineStage, triggeredBy int32) (error, bool) {
	if pipelineStage != nil {
		stageReq := &bean3.PipelineStageDto{
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

func (impl *WorkflowDagExecutorImpl) triggerStage(cdWf *pipelineConfig.CdWorkflow, pipeline *pipelineConfig.Pipeline, artifact *repository.CiArtifact, applyAuth bool, triggeredBy int32) error {

	preStage, err := impl.getPipelineStage(pipeline.Id, repository4.PIPELINE_STAGE_TYPE_PRE_CD)
	if err != nil {
		return err
	}

	//handle corrupt data (https://github.com/devtron-labs/devtron/issues/3826)
	err, deleted := impl.deleteCorruptedPipelineStage(preStage, triggeredBy)
	if err != nil {
		impl.logger.Errorw("error in deleteCorruptedPipelineStage ", "cdPipelineId", pipeline.Id, "err", err, "preStage", preStage, "triggeredBy", triggeredBy)
		return err
	}

	if len(pipeline.PreStageConfig) > 0 || (preStage != nil && !deleted) {
		// pre stage exists
		if pipeline.PreTriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC {
			impl.logger.Debugw("trigger pre stage for pipeline", "artifactId", artifact.Id, "pipelineId", pipeline.Id)
			err = impl.TriggerPreStage(context.Background(), cdWf, artifact, pipeline, artifact.UpdatedBy, applyAuth, 0) //TODO handle error here
			return err
		}
	} else if pipeline.TriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC {
		// trigger deployment
		impl.logger.Debugw("trigger cd for pipeline", "artifactId", artifact.Id, "pipelineId", pipeline.Id)
		err = impl.TriggerDeployment(cdWf, artifact, pipeline, applyAuth, triggeredBy)
		return err
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) getPipelineStage(pipelineId int, stageType repository4.PipelineStageType) (*repository4.PipelineStage, error) {
	stage, err := impl.pipelineStageService.GetCdStageByCdPipelineIdAndStageType(pipelineId, stageType)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching CD pipeline stage", "cdPipelineId", pipelineId, "stage ", stage, "err", err)
		return nil, err
	}
	return stage, nil
}

func (impl *WorkflowDagExecutorImpl) triggerStageForBulk(cdWf *pipelineConfig.CdWorkflow, pipeline *pipelineConfig.Pipeline, artifact *repository.CiArtifact, applyAuth bool, async bool, triggeredBy int32) error {

	preStage, err := impl.getPipelineStage(pipeline.Id, repository4.PIPELINE_STAGE_TYPE_PRE_CD)
	if err != nil {
		return err
	}

	//handle corrupt data (https://github.com/devtron-labs/devtron/issues/3826)
	err, deleted := impl.deleteCorruptedPipelineStage(preStage, triggeredBy)
	if err != nil {
		impl.logger.Errorw("error in deleteCorruptedPipelineStage ", "cdPipelineId", pipeline.Id, "err", err, "preStage", preStage, "triggeredBy", triggeredBy)
		return err
	}

	if len(pipeline.PreStageConfig) > 0 || (preStage != nil && !deleted) {
		//pre stage exists
		impl.logger.Debugw("trigger pre stage for pipeline", "artifactId", artifact.Id, "pipelineId", pipeline.Id)
		err = impl.TriggerPreStage(context.Background(), cdWf, artifact, pipeline, artifact.UpdatedBy, applyAuth, 0) //TODO handle error here
		return err
	} else {
		// trigger deployment
		impl.logger.Debugw("trigger cd for pipeline", "artifactId", artifact.Id, "pipelineId", pipeline.Id)
		err = impl.TriggerDeployment(cdWf, artifact, pipeline, applyAuth, triggeredBy)
		return err
	}
}
func (impl *WorkflowDagExecutorImpl) HandlePreStageSuccessEvent(cdStageCompleteEvent CdStageCompleteEvent) error {
	wfRunner, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(cdStageCompleteEvent.WorkflowRunnerId)
	if err != nil {
		return err
	}
	if wfRunner.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
		pipeline, err := impl.pipelineRepository.FindById(cdStageCompleteEvent.CdPipelineId)
		if err != nil {
			return err
		}
		if pipeline.TriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC {
			ciArtifact, err := impl.ciArtifactRepository.Get(cdStageCompleteEvent.CiArtifactDTO.Id)
			if err != nil {
				return err
			}
			cdWorkflow, err := impl.cdWorkflowRepository.FindById(cdStageCompleteEvent.WorkflowId)
			if err != nil {
				return err
			}
			//TODO : confirm about this logic used for applyAuth
			applyAuth := false
			if cdStageCompleteEvent.TriggeredBy != 1 {
				applyAuth = true
			}
			err = impl.TriggerDeployment(cdWorkflow, ciArtifact, pipeline, applyAuth, cdStageCompleteEvent.TriggeredBy)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) TriggerPreStage(ctx context.Context, cdWf *pipelineConfig.CdWorkflow, artifact *repository.CiArtifact, pipeline *pipelineConfig.Pipeline, triggeredBy int32, applyAuth bool, refCdWorkflowRunnerId int) error {
	//setting triggeredAt variable to have consistent data for various audit log places in db for deployment time
	triggeredAt := time.Now()

	//in case of pre stage manual trigger auth is already applied
	if applyAuth {
		user, err := impl.user.GetById(artifact.UpdatedBy)
		if err != nil {
			impl.logger.Errorw("error in fetching user for auto pipeline", "UpdatedBy", artifact.UpdatedBy)
			return nil
		}
		token := user.EmailId
		object := impl.enforcerUtil.GetAppRBACNameByAppId(pipeline.AppId)
		impl.logger.Debugw("Triggered Request (App Permission Checking):", "object", object)
		if ok := impl.enforcer.EnforceByEmail(strings.ToLower(token), casbin.ResourceApplications, casbin.ActionTrigger, object); !ok {
			impl.logger.Warnw("unauthorized for pipeline ", "pipelineId", strconv.Itoa(pipeline.Id))
			return fmt.Errorf("unauthorized for pipeline " + strconv.Itoa(pipeline.Id))
		}
	}
	var err error
	if cdWf == nil {
		cdWf = &pipelineConfig.CdWorkflow{
			CiArtifactId: artifact.Id,
			PipelineId:   pipeline.Id,
			AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: 1, UpdatedOn: triggeredAt, UpdatedBy: 1},
		}
		err := impl.cdWorkflowRepository.SaveWorkFlow(ctx, cdWf)
		if err != nil {
			return err
		}
	}
	cdWorkflowExecutorType := impl.config.GetWorkflowExecutorType()
	runner := &pipelineConfig.CdWorkflowRunner{
		Name:                  pipeline.Name,
		WorkflowType:          bean.CD_WORKFLOW_TYPE_PRE,
		ExecutorType:          cdWorkflowExecutorType,
		Status:                pipelineConfig.WorkflowStarting, //starting
		TriggeredBy:           triggeredBy,
		StartedOn:             triggeredAt,
		Namespace:             impl.config.GetDefaultNamespace(),
		BlobStorageEnabled:    impl.config.BlobStorageEnabled,
		CdWorkflowId:          cdWf.Id,
		LogLocation:           fmt.Sprintf("%s/%s%s-%s/main.log", impl.config.GetDefaultBuildLogsKeyPrefix(), strconv.Itoa(cdWf.Id), string(bean.CD_WORKFLOW_TYPE_PRE), pipeline.Name),
		AuditLog:              sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: 1, UpdatedOn: triggeredAt, UpdatedBy: 1},
		RefCdWorkflowRunnerId: refCdWorkflowRunnerId,
	}
	var env *repository2.Environment
	if pipeline.RunPreStageInEnv {
		_, span := otel.Tracer("orchestrator").Start(ctx, "envRepository.FindById")
		env, err = impl.envRepository.FindById(pipeline.EnvironmentId)
		span.End()
		if err != nil {
			impl.logger.Errorw(" unable to find env ", "err", err)
			return err
		}
		impl.logger.Debugw("env", "env", env)
		runner.Namespace = env.Namespace
	}
	_, span := otel.Tracer("orchestrator").Start(ctx, "cdWorkflowRepository.SaveWorkFlowRunner")
	_, err = impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
	span.End()
	if err != nil {
		return err
	}

	//checking vulnerability for the selected image
	isVulnerable, err := impl.GetArtifactVulnerabilityStatus(artifact, pipeline, ctx)
	if err != nil {
		impl.logger.Errorw("error in getting Artifact vulnerability status, TriggerPreStage", "err", err)
		return err
	}
	if isVulnerable {
		// if image vulnerable, update timeline status and return
		runner.Status = pipelineConfig.WorkflowFailed
		runner.Message = "Found vulnerability on image"
		runner.FinishedOn = time.Now()
		runner.UpdatedOn = time.Now()
		runner.UpdatedBy = triggeredBy
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
		if err != nil {
			impl.logger.Errorw("error in updating wfr status due to vulnerable image", "err", err)
			return err
		}
		return fmt.Errorf("found vulnerability for image digest %s", artifact.ImageDigest)
	}

	_, span = otel.Tracer("orchestrator").Start(ctx, "buildWFRequest")
	cdStageWorkflowRequest, err := impl.buildWFRequest(runner, cdWf, pipeline, triggeredBy)
	span.End()
	if err != nil {
		return err
	}
	cdStageWorkflowRequest.StageType = PRE
	_, span = otel.Tracer("orchestrator").Start(ctx, "cdWorkflowService.SubmitWorkflow")
	cdStageWorkflowRequest.Pipeline = pipeline
	cdStageWorkflowRequest.Env = env
	cdStageWorkflowRequest.Type = bean3.CD_WORKFLOW_PIPELINE_TYPE
	_, err = impl.cdWorkflowService.SubmitWorkflow(cdStageWorkflowRequest)
	span.End()

	err = impl.sendPreStageNotification(ctx, cdWf, pipeline)
	if err != nil {
		return err
	}
	//creating cd config history entry
	_, span = otel.Tracer("orchestrator").Start(ctx, "prePostCdScriptHistoryService.CreatePrePostCdScriptHistory")
	err = impl.prePostCdScriptHistoryService.CreatePrePostCdScriptHistory(pipeline, nil, repository3.PRE_CD_TYPE, true, triggeredBy, triggeredAt)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in creating pre cd script entry", "err", err, "pipeline", pipeline)
		return err
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) sendPreStageNotification(ctx context.Context, cdWf *pipelineConfig.CdWorkflow, pipeline *pipelineConfig.Pipeline) error {
	wfr, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(ctx, cdWf.Id, bean.CD_WORKFLOW_TYPE_PRE)
	if err != nil {
		return err
	}

	event := impl.eventFactory.Build(util2.Trigger, &pipeline.Id, pipeline.AppId, &pipeline.EnvironmentId, util2.CD)
	impl.logger.Debugw("event PreStageTrigger", "event", event)
	event = impl.eventFactory.BuildExtraCDData(event, &wfr, 0, bean.CD_WORKFLOW_TYPE_PRE)
	_, span := otel.Tracer("orchestrator").Start(ctx, "eventClient.WriteNotificationEvent")
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	span.End()
	if evtErr != nil {
		impl.logger.Errorw("CD trigger event not sent", "error", evtErr)
	}
	return nil
}

func convert(ts string) (*time.Time, error) {
	//layout := "2006-01-02T15:04:05Z"
	t, err := time.Parse(bean2.LayoutRFC3339, ts)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (impl *WorkflowDagExecutorImpl) TriggerPostStage(cdWf *pipelineConfig.CdWorkflow, pipeline *pipelineConfig.Pipeline, triggeredBy int32, refCdWorkflowRunnerId int) error {
	//setting triggeredAt variable to have consistent data for various audit log places in db for deployment time
	triggeredAt := time.Now()

	runner := &pipelineConfig.CdWorkflowRunner{
		Name:                  pipeline.Name,
		WorkflowType:          bean.CD_WORKFLOW_TYPE_POST,
		ExecutorType:          impl.config.GetWorkflowExecutorType(),
		Status:                pipelineConfig.WorkflowStarting,
		TriggeredBy:           triggeredBy,
		StartedOn:             triggeredAt,
		Namespace:             impl.config.GetDefaultNamespace(),
		BlobStorageEnabled:    impl.config.BlobStorageEnabled,
		CdWorkflowId:          cdWf.Id,
		LogLocation:           fmt.Sprintf("%s/%s%s-%s/main.log", impl.config.GetDefaultBuildLogsKeyPrefix(), strconv.Itoa(cdWf.Id), string(bean.CD_WORKFLOW_TYPE_POST), pipeline.Name),
		AuditLog:              sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: triggeredBy, UpdatedOn: triggeredAt, UpdatedBy: triggeredBy},
		RefCdWorkflowRunnerId: refCdWorkflowRunnerId,
	}
	var env *repository2.Environment
	var err error
	if pipeline.RunPostStageInEnv {
		env, err = impl.envRepository.FindById(pipeline.EnvironmentId)
		if err != nil {
			impl.logger.Errorw(" unable to find env ", "err", err)
			return err
		}
		runner.Namespace = env.Namespace
	}

	_, err = impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
	if err != nil {
		return err
	}

	if cdWf.CiArtifact == nil || cdWf.CiArtifact.Id == 0 {
		cdWf.CiArtifact, err = impl.ciArtifactRepository.Get(cdWf.CiArtifactId)
		if err != nil {
			impl.logger.Errorw("error fetching artifact data", "err", err)
			return err
		}
	}
	//checking vulnerability for the selected image
	isVulnerable, err := impl.GetArtifactVulnerabilityStatus(cdWf.CiArtifact, pipeline, context.Background())
	if err != nil {
		impl.logger.Errorw("error in getting Artifact vulnerability status, TriggerPostStage", "err", err)
		return err
	}
	if isVulnerable {
		// if image vulnerable, update timeline status and return
		runner.Status = pipelineConfig.WorkflowFailed
		runner.Message = "Found vulnerability on image"
		runner.FinishedOn = time.Now()
		runner.UpdatedOn = time.Now()
		runner.UpdatedBy = triggeredBy
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
		if err != nil {
			impl.logger.Errorw("error in updating wfr status due to vulnerable image", "err", err)
			return err
		}
		return fmt.Errorf("found vulnerability for image digest %s", cdWf.CiArtifact.ImageDigest)
	}

	cdStageWorkflowRequest, err := impl.buildWFRequest(runner, cdWf, pipeline, triggeredBy)
	if err != nil {
		impl.logger.Errorw("error in building wfRequest", "err", err, "runner", runner, "cdWf", cdWf, "pipeline", pipeline)
		return err
	}
	cdStageWorkflowRequest.StageType = POST
	cdStageWorkflowRequest.Pipeline = pipeline
	cdStageWorkflowRequest.Env = env
	cdStageWorkflowRequest.Type = bean3.CD_WORKFLOW_PIPELINE_TYPE
	_, err = impl.cdWorkflowService.SubmitWorkflow(cdStageWorkflowRequest)
	if err != nil {
		impl.logger.Errorw("error in submitting workflow", "err", err, "cdStageWorkflowRequest", cdStageWorkflowRequest, "pipeline", pipeline, "env", env)
		return err
	}

	wfr, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(context.Background(), cdWf.Id, bean.CD_WORKFLOW_TYPE_POST)
	if err != nil {
		impl.logger.Errorw("error in getting wfr by workflowId and runnerType", "err", err, "wfId", cdWf.Id)
		return err
	}

	event := impl.eventFactory.Build(util2.Trigger, &pipeline.Id, pipeline.AppId, &pipeline.EnvironmentId, util2.CD)
	impl.logger.Debugw("event Cd Post Trigger", "event", event)
	event = impl.eventFactory.BuildExtraCDData(event, &wfr, 0, bean.CD_WORKFLOW_TYPE_POST)
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("CD trigger event not sent", "error", evtErr)
	}
	//creating cd config history entry
	err = impl.prePostCdScriptHistoryService.CreatePrePostCdScriptHistory(pipeline, nil, repository3.POST_CD_TYPE, true, triggeredBy, triggeredAt)
	if err != nil {
		impl.logger.Errorw("error in creating post cd script entry", "err", err, "pipeline", pipeline)
		return err
	}
	return nil
}
func (impl *WorkflowDagExecutorImpl) buildArtifactLocationForS3(cdWorkflowConfig *pipelineConfig.CdWorkflowConfig, cdWf *pipelineConfig.CdWorkflow, runner *pipelineConfig.CdWorkflowRunner) (string, string, string) {
	cdArtifactLocationFormat := cdWorkflowConfig.CdArtifactLocationFormat
	if cdArtifactLocationFormat == "" {
		cdArtifactLocationFormat = impl.config.GetArtifactLocationFormat()
	}
	if cdWorkflowConfig.LogsBucket == "" {
		cdWorkflowConfig.LogsBucket = impl.config.GetDefaultBuildLogsBucket()
	}
	ArtifactLocation := fmt.Sprintf("s3://%s/"+impl.config.GetDefaultArtifactKeyPrefix()+"/"+cdArtifactLocationFormat, cdWorkflowConfig.LogsBucket, cdWf.Id, runner.Id)
	artifactFileName := fmt.Sprintf(impl.config.GetDefaultArtifactKeyPrefix()+"/"+cdArtifactLocationFormat, cdWf.Id, runner.Id)
	return ArtifactLocation, cdWorkflowConfig.LogsBucket, artifactFileName
}

func (impl *WorkflowDagExecutorImpl) getDeployStageDetails(pipelineId int) (pipelineConfig.CdWorkflowRunner, *bean.UserInfo, int, error) {
	deployStageWfr := pipelineConfig.CdWorkflowRunner{}
	//getting deployment pipeline latest wfr by pipelineId
	deployStageWfr, err := impl.cdWorkflowRepository.FindLastStatusByPipelineIdAndRunnerType(pipelineId, bean.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		impl.logger.Errorw("error in getting latest status of deploy type wfr by pipelineId", "err", err, "pipelineId", pipelineId)
		return deployStageWfr, nil, 0, err
	}
	deployStageTriggeredByUser, err := impl.user.GetById(deployStageWfr.TriggeredBy)
	if err != nil {
		impl.logger.Errorw("error in getting userDetails by id", "err", err, "userId", deployStageWfr.TriggeredBy)
		return deployStageWfr, nil, 0, err
	}
	pipelineReleaseCounter, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(pipelineId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching latest release counter for pipeline", "pipelineId", pipelineId, "err", err)
		return deployStageWfr, nil, 0, err
	}
	return deployStageWfr, deployStageTriggeredByUser, pipelineReleaseCounter, nil
}

func isExtraVariableDynamic(variableName string, webhookAndCiData *gitSensorClient.WebhookAndCiData) bool {
	if strings.Contains(variableName, GIT_COMMIT_HASH_PREFIX) || strings.Contains(variableName, GIT_SOURCE_TYPE_PREFIX) || strings.Contains(variableName, GIT_SOURCE_VALUE_PREFIX) ||
		strings.Contains(variableName, APP_LABEL_VALUE_PREFIX) || strings.Contains(variableName, APP_LABEL_KEY_PREFIX) ||
		strings.Contains(variableName, CHILD_CD_ENV_NAME_PREFIX) || strings.Contains(variableName, CHILD_CD_CLUSTER_NAME_PREFIX) ||
		strings.Contains(variableName, CHILD_CD_COUNT) || strings.Contains(variableName, APP_LABEL_COUNT) || strings.Contains(variableName, GIT_SOURCE_COUNT) ||
		webhookAndCiData != nil {

		return true
	}
	return false
}

func setExtraEnvVariableInDeployStep(deploySteps []*bean3.StepObject, extraEnvVariables map[string]string, webhookAndCiData *gitSensorClient.WebhookAndCiData) {
	for _, deployStep := range deploySteps {
		for variableKey, variableValue := range extraEnvVariables {
			if isExtraVariableDynamic(variableKey, webhookAndCiData) && deployStep.StepType == "INLINE" {
				extraInputVar := &bean3.VariableObject{
					Name:                  variableKey,
					Format:                "STRING",
					Value:                 variableValue,
					VariableType:          bean3.VARIABLE_TYPE_REF_GLOBAL,
					ReferenceVariableName: variableKey,
				}
				deployStep.InputVars = append(deployStep.InputVars, extraInputVar)
			}
		}
	}
}
func (impl *WorkflowDagExecutorImpl) buildWFRequest(runner *pipelineConfig.CdWorkflowRunner, cdWf *pipelineConfig.CdWorkflow, cdPipeline *pipelineConfig.Pipeline, triggeredBy int32) (*WorkflowRequest, error) {
	cdWorkflowConfig, err := impl.cdWorkflowRepository.FindConfigByPipelineId(cdPipeline.Id)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}

	workflowExecutor := runner.ExecutorType

	artifact, err := impl.ciArtifactRepository.Get(cdWf.CiArtifactId)
	if err != nil {
		return nil, err
	}

	ciMaterialInfo, err := repository.GetCiMaterialInfo(artifact.MaterialInfo, artifact.DataSource)
	if err != nil {
		impl.logger.Errorw("parsing error", "err", err)
		return nil, err
	}

	var ciProjectDetails []bean3.CiProjectDetails
	var ciPipeline *pipelineConfig.CiPipeline
	if cdPipeline.CiPipelineId > 0 {
		ciPipeline, err = impl.ciPipelineRepository.FindById(cdPipeline.CiPipelineId)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("cannot find ciPipelineRequest", "err", err)
			return nil, err
		}

		for _, m := range ciPipeline.CiPipelineMaterials {
			// git material should be active in this case
			if m == nil || m.GitMaterial == nil || !m.GitMaterial.Active {
				continue
			}
			var ciMaterialCurrent repository.CiMaterialInfo
			for _, ciMaterial := range ciMaterialInfo {
				if ciMaterial.Material.GitConfiguration.URL == m.GitMaterial.Url {
					ciMaterialCurrent = ciMaterial
					break
				}
			}
			gitMaterial, err := impl.materialRepository.FindById(m.GitMaterialId)
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("could not fetch git materials", "err", err)
				return nil, err
			}

			ciProjectDetail := bean3.CiProjectDetails{
				GitRepository:   ciMaterialCurrent.Material.GitConfiguration.URL,
				MaterialName:    gitMaterial.Name,
				CheckoutPath:    gitMaterial.CheckoutPath,
				FetchSubmodules: gitMaterial.FetchSubmodules,
				SourceType:      m.Type,
				SourceValue:     m.Value,
				Type:            string(m.Type),
				GitOptions: bean3.GitOptions{
					UserName:      gitMaterial.GitProvider.UserName,
					Password:      gitMaterial.GitProvider.Password,
					SshPrivateKey: gitMaterial.GitProvider.SshPrivateKey,
					AccessToken:   gitMaterial.GitProvider.AccessToken,
					AuthMode:      gitMaterial.GitProvider.AuthMode,
				},
			}

			if len(ciMaterialCurrent.Modifications) > 0 {
				ciProjectDetail.CommitHash = ciMaterialCurrent.Modifications[0].Revision
				ciProjectDetail.Author = ciMaterialCurrent.Modifications[0].Author
				ciProjectDetail.GitTag = ciMaterialCurrent.Modifications[0].Tag
				ciProjectDetail.Message = ciMaterialCurrent.Modifications[0].Message
				commitTime, err := convert(ciMaterialCurrent.Modifications[0].ModifiedTime)
				if err != nil {
					return nil, err
				}
				ciProjectDetail.CommitTime = commitTime.Format(bean2.LayoutRFC3339)
			} else {
				impl.logger.Debugw("devtronbug#1062", ciPipeline.Id, cdPipeline.Id)
				return nil, fmt.Errorf("modifications not found for %d", ciPipeline.Id)
			}

			// set webhook data
			if m.Type == pipelineConfig.SOURCE_TYPE_WEBHOOK && len(ciMaterialCurrent.Modifications) > 0 {
				webhookData := ciMaterialCurrent.Modifications[0].WebhookData
				ciProjectDetail.WebhookData = pipelineConfig.WebhookData{
					Id:              webhookData.Id,
					EventActionType: webhookData.EventActionType,
					Data:            webhookData.Data,
				}
			}

			ciProjectDetails = append(ciProjectDetails, ciProjectDetail)
		}
	}
	var stageYaml string
	var deployStageWfr pipelineConfig.CdWorkflowRunner
	deployStageTriggeredByUser := &bean.UserInfo{}
	var pipelineReleaseCounter int
	var preDeploySteps []*bean3.StepObject
	var postDeploySteps []*bean3.StepObject
	var refPluginsData []*bean3.RefPluginObject
	//if pipeline_stage_steps present for pre-CD or post-CD then no need to add stageYaml to cdWorkflowRequest in that
	//case add PreDeploySteps and PostDeploySteps to cdWorkflowRequest, this is done for backward compatibility
	pipelineStage, err := impl.getPipelineStage(cdPipeline.Id, runner.WorkflowType.WorkflowTypeToStageType())
	if err != nil {
		return nil, err
	}
	env, err := impl.envRepository.FindById(cdPipeline.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in getting environment by id", "err", err)
		return nil, err
	}

	if pipelineStage != nil {
		//Scope will pick the environment of CD pipeline irrespective of in-cluster mode,
		//since user sees the environment of the CD pipeline
		scope := resourceQualifiers.Scope{
			AppId:     cdPipeline.App.Id,
			EnvId:     env.Id,
			ClusterId: env.ClusterId,
			SystemMetadata: &resourceQualifiers.SystemMetadata{
				EnvironmentName: env.Name,
				ClusterName:     env.Cluster.ClusterName,
				Namespace:       env.Namespace,
				Image:           artifact.Image,
				ImageTag:        util3.GetImageTagFromImage(artifact.Image),
			},
		}
		var variableSnapshot map[string]string
		if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
			//preDeploySteps, _, refPluginsData, err = impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(cdPipeline.Id, cdStage)
			prePostAndRefPluginResponse, err := impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(cdPipeline.Id, preCdStage, scope)
			if err != nil {
				impl.logger.Errorw("error in getting pre, post & refPlugin steps data for wf request", "err", err, "cdPipelineId", cdPipeline.Id)
				return nil, err
			}
			preDeploySteps = prePostAndRefPluginResponse.PreStageSteps
			refPluginsData = prePostAndRefPluginResponse.RefPluginData
			variableSnapshot = prePostAndRefPluginResponse.VariableSnapshot
		} else if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
			//_, postDeploySteps, refPluginsData, err = impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(cdPipeline.Id, cdStage)
			prePostAndRefPluginResponse, err := impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(cdPipeline.Id, postCdStage, scope)
			if err != nil {
				impl.logger.Errorw("error in getting pre, post & refPlugin steps data for wf request", "err", err, "cdPipelineId", cdPipeline.Id)
				return nil, err
			}
			postDeploySteps = prePostAndRefPluginResponse.PostStageSteps
			refPluginsData = prePostAndRefPluginResponse.RefPluginData
			variableSnapshot = prePostAndRefPluginResponse.VariableSnapshot
			deployStageWfr, deployStageTriggeredByUser, pipelineReleaseCounter, err = impl.getDeployStageDetails(cdPipeline.Id)
			if err != nil {
				impl.logger.Errorw("error in getting deployStageWfr, deployStageTriggeredByUser and pipelineReleaseCounter wf request", "err", err, "cdPipelineId", cdPipeline.Id)
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("unsupported workflow triggerd")
		}

		//Save Scoped VariableSnapshot
		if len(variableSnapshot) > 0 {
			variableMapBytes, _ := json.Marshal(variableSnapshot)
			err := impl.variableSnapshotHistoryService.SaveVariableHistoriesForTrigger([]*repository5.VariableSnapshotHistoryBean{{
				VariableSnapshot: variableMapBytes,
				HistoryReference: repository5.HistoryReference{
					HistoryReferenceId:   runner.Id,
					HistoryReferenceType: repository5.HistoryReferenceTypeCDWORKFLOWRUNNER,
				},
			}}, runner.TriggeredBy)
			if err != nil {
				impl.logger.Errorf("Not able to save variable snapshot for CD trigger %s %d %s", err, runner.Id, variableSnapshot)
			}
		}
	} else {
		//in this case no plugin script is not present for this cdPipeline hence going with attaching preStage or postStage config
		if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
			stageYaml = cdPipeline.PreStageConfig
		} else if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
			stageYaml = cdPipeline.PostStageConfig
			deployStageWfr, deployStageTriggeredByUser, pipelineReleaseCounter, err = impl.getDeployStageDetails(cdPipeline.Id)
			if err != nil {
				impl.logger.Errorw("error in getting deployStageWfr, deployStageTriggeredByUser and pipelineReleaseCounter wf request", "err", err, "cdPipelineId", cdPipeline.Id)
				return nil, err
			}

		} else {
			return nil, fmt.Errorf("unsupported workflow triggerd")
		}
	}

	cdStageWorkflowRequest := &WorkflowRequest{
		EnvironmentId:         cdPipeline.EnvironmentId,
		AppId:                 cdPipeline.AppId,
		WorkflowId:            cdWf.Id,
		WorkflowRunnerId:      runner.Id,
		WorkflowNamePrefix:    strconv.Itoa(runner.Id) + "-" + runner.Name,
		WorkflowPrefixForLog:  strconv.Itoa(cdWf.Id) + string(runner.WorkflowType) + "-" + runner.Name,
		CdImage:               impl.config.GetDefaultImage(),
		CdPipelineId:          cdWf.PipelineId,
		TriggeredBy:           triggeredBy,
		StageYaml:             stageYaml,
		CiProjectDetails:      ciProjectDetails,
		Namespace:             runner.Namespace,
		ActiveDeadlineSeconds: impl.config.GetDefaultTimeout(),
		CiArtifactDTO: CiArtifactDTO{
			Id:           artifact.Id,
			PipelineId:   artifact.PipelineId,
			Image:        artifact.Image,
			ImageDigest:  artifact.ImageDigest,
			MaterialInfo: artifact.MaterialInfo,
			DataSource:   artifact.DataSource,
			WorkflowId:   artifact.WorkflowId,
		},
		OrchestratorHost:  impl.config.OrchestratorHost,
		OrchestratorToken: impl.config.OrchestratorToken,
		CloudProvider:     impl.config.CloudProvider,
		WorkflowExecutor:  workflowExecutor,
		RefPlugins:        refPluginsData,
	}

	extraEnvVariables := make(map[string]string)
	if env != nil {
		extraEnvVariables[CD_PIPELINE_ENV_NAME_KEY] = env.Name
		if env.Cluster != nil {
			extraEnvVariables[CD_PIPELINE_CLUSTER_NAME_KEY] = env.Cluster.ClusterName
		}
	}
	ciWf, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflowByArtifactId(artifact.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting ciWf by artifactId", "err", err, "artifactId", artifact.Id)
		return nil, err
	}
	var webhookAndCiData *gitSensorClient.WebhookAndCiData
	if ciWf != nil && ciWf.GitTriggers != nil {
		i := 1
		var gitCommitEnvVariables []GitMetadata

		for ciPipelineMaterialId, gitTrigger := range ciWf.GitTriggers {
			extraEnvVariables[fmt.Sprintf("%s_%d", GIT_COMMIT_HASH_PREFIX, i)] = gitTrigger.Commit
			extraEnvVariables[fmt.Sprintf("%s_%d", GIT_SOURCE_TYPE_PREFIX, i)] = string(gitTrigger.CiConfigureSourceType)
			extraEnvVariables[fmt.Sprintf("%s_%d", GIT_SOURCE_VALUE_PREFIX, i)] = gitTrigger.CiConfigureSourceValue

			gitCommitEnvVariables = append(gitCommitEnvVariables, GitMetadata{
				GitCommitHash:  gitTrigger.Commit,
				GitSourceType:  string(gitTrigger.CiConfigureSourceType),
				GitSourceValue: gitTrigger.CiConfigureSourceValue,
			})

			// CODE-BLOCK starts - store extra environment variables if webhook
			if gitTrigger.CiConfigureSourceType == pipelineConfig.SOURCE_TYPE_WEBHOOK {
				webhookDataId := gitTrigger.WebhookData.Id
				if webhookDataId > 0 {
					webhookDataRequest := &gitSensorClient.WebhookDataRequest{
						Id:                   webhookDataId,
						CiPipelineMaterialId: ciPipelineMaterialId,
					}
					webhookAndCiData, err = impl.gitSensorGrpcClient.GetWebhookData(context.Background(), webhookDataRequest)
					if err != nil {
						impl.logger.Errorw("err while getting webhook data from git-sensor", "err", err, "webhookDataRequest", webhookDataRequest)
						return nil, err
					}
					if webhookAndCiData != nil {
						for extEnvVariableKey, extEnvVariableVal := range webhookAndCiData.ExtraEnvironmentVariables {
							extraEnvVariables[extEnvVariableKey] = extEnvVariableVal
						}
					}
				}
			}
			// CODE_BLOCK ends

			i++
		}
		gitMetadata, err := json.Marshal(&gitCommitEnvVariables)
		if err != nil {
			impl.logger.Errorw("err while marshaling git metdata", "err", err)
			return nil, err
		}
		extraEnvVariables[GIT_METADATA] = string(gitMetadata)

		extraEnvVariables[GIT_SOURCE_COUNT] = strconv.Itoa(len(ciWf.GitTriggers))
	}

	childCdIds, err := impl.appWorkflowRepository.FindChildCDIdsByParentCDPipelineId(cdPipeline.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting child cdPipelineIds by parent cdPipelineId", "err", err, "parent cdPipelineId", cdPipeline.Id)
		return nil, err
	}
	if len(childCdIds) > 0 {
		childPipelines, err := impl.pipelineRepository.FindByIdsIn(childCdIds)
		if err != nil {
			impl.logger.Errorw("error in getting pipelines by ids", "err", err, "ids", childCdIds)
			return nil, err
		}
		var childCdEnvVariables []ChildCdMetadata
		for i, childPipeline := range childPipelines {
			extraEnvVariables[fmt.Sprintf("%s_%d", CHILD_CD_ENV_NAME_PREFIX, i+1)] = childPipeline.Environment.Name
			extraEnvVariables[fmt.Sprintf("%s_%d", CHILD_CD_CLUSTER_NAME_PREFIX, i+1)] = childPipeline.Environment.Cluster.ClusterName

			childCdEnvVariables = append(childCdEnvVariables, ChildCdMetadata{
				ChildCdEnvName:     childPipeline.Environment.Name,
				ChildCdClusterName: childPipeline.Environment.Cluster.ClusterName,
			})
		}
		childCdEnvVariablesMetadata, err := json.Marshal(&childCdEnvVariables)
		if err != nil {
			impl.logger.Errorw("err while marshaling childCdEnvVariables", "err", err)
			return nil, err
		}
		extraEnvVariables[CHILD_CD_METADATA] = string(childCdEnvVariablesMetadata)

		extraEnvVariables[CHILD_CD_COUNT] = strconv.Itoa(len(childPipelines))
	}
	if ciPipeline != nil && ciPipeline.Id > 0 {
		extraEnvVariables["APP_NAME"] = ciPipeline.App.AppName
		cdStageWorkflowRequest.DockerUsername = ciPipeline.CiTemplate.DockerRegistry.Username
		cdStageWorkflowRequest.DockerPassword = ciPipeline.CiTemplate.DockerRegistry.Password
		cdStageWorkflowRequest.AwsRegion = ciPipeline.CiTemplate.DockerRegistry.AWSRegion
		cdStageWorkflowRequest.DockerConnection = ciPipeline.CiTemplate.DockerRegistry.Connection
		cdStageWorkflowRequest.DockerCert = ciPipeline.CiTemplate.DockerRegistry.Cert
		cdStageWorkflowRequest.AccessKey = ciPipeline.CiTemplate.DockerRegistry.AWSAccessKeyId
		cdStageWorkflowRequest.SecretKey = ciPipeline.CiTemplate.DockerRegistry.AWSSecretAccessKey
		cdStageWorkflowRequest.DockerRegistryType = string(ciPipeline.CiTemplate.DockerRegistry.RegistryType)
		cdStageWorkflowRequest.DockerRegistryURL = ciPipeline.CiTemplate.DockerRegistry.RegistryURL
	} else if cdPipeline.AppId > 0 {
		ciTemplate, err := impl.CiTemplateRepository.FindByAppId(cdPipeline.AppId)
		if err != nil {
			return nil, err
		}
		extraEnvVariables["APP_NAME"] = ciTemplate.App.AppName
		cdStageWorkflowRequest.DockerUsername = ciTemplate.DockerRegistry.Username
		cdStageWorkflowRequest.DockerPassword = ciTemplate.DockerRegistry.Password
		cdStageWorkflowRequest.AwsRegion = ciTemplate.DockerRegistry.AWSRegion
		cdStageWorkflowRequest.DockerConnection = ciTemplate.DockerRegistry.Connection
		cdStageWorkflowRequest.DockerCert = ciTemplate.DockerRegistry.Cert
		cdStageWorkflowRequest.AccessKey = ciTemplate.DockerRegistry.AWSAccessKeyId
		cdStageWorkflowRequest.SecretKey = ciTemplate.DockerRegistry.AWSSecretAccessKey
		cdStageWorkflowRequest.DockerRegistryType = string(ciTemplate.DockerRegistry.RegistryType)
		cdStageWorkflowRequest.DockerRegistryURL = ciTemplate.DockerRegistry.RegistryURL
		appLabels, err := impl.appLabelRepository.FindAllByAppId(cdPipeline.AppId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting labels by appId", "err", err, "appId", cdPipeline.AppId)
			return nil, err
		}
		var appLabelEnvVariables []AppLabelMetadata
		for i, appLabel := range appLabels {
			extraEnvVariables[fmt.Sprintf("%s_%d", APP_LABEL_KEY_PREFIX, i+1)] = appLabel.Key
			extraEnvVariables[fmt.Sprintf("%s_%d", APP_LABEL_VALUE_PREFIX, i+1)] = appLabel.Value
			appLabelEnvVariables = append(appLabelEnvVariables, AppLabelMetadata{
				AppLabelKey:   appLabel.Key,
				AppLabelValue: appLabel.Value,
			})
		}
		if len(appLabels) > 0 {
			extraEnvVariables[APP_LABEL_COUNT] = strconv.Itoa(len(appLabels))
			appLabelEnvVariablesMetadata, err := json.Marshal(&appLabelEnvVariables)
			if err != nil {
				impl.logger.Errorw("err while marshaling appLabelEnvVariables", "err", err)
				return nil, err
			}
			extraEnvVariables[APP_LABEL_METADATA] = string(appLabelEnvVariablesMetadata)

		}
	}
	cdStageWorkflowRequest.ExtraEnvironmentVariables = extraEnvVariables
	if deployStageTriggeredByUser != nil {
		cdStageWorkflowRequest.DeploymentTriggerTime = deployStageWfr.StartedOn
		cdStageWorkflowRequest.DeploymentTriggeredBy = deployStageTriggeredByUser.EmailId
	}
	if pipelineReleaseCounter > 0 {
		cdStageWorkflowRequest.DeploymentReleaseCounter = pipelineReleaseCounter
	}
	if cdWorkflowConfig.CdCacheRegion == "" {
		cdWorkflowConfig.CdCacheRegion = impl.config.GetDefaultCdLogsBucketRegion()
	}

	if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
		//populate input variables of steps with extra env variables
		setExtraEnvVariableInDeployStep(preDeploySteps, extraEnvVariables, webhookAndCiData)
		cdStageWorkflowRequest.PrePostDeploySteps = preDeploySteps
	} else if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
		setExtraEnvVariableInDeployStep(postDeploySteps, extraEnvVariables, webhookAndCiData)
		cdStageWorkflowRequest.PrePostDeploySteps = postDeploySteps
	}
	cdStageWorkflowRequest.BlobStorageConfigured = runner.BlobStorageEnabled
	switch cdStageWorkflowRequest.CloudProvider {
	case BLOB_STORAGE_S3:
		//No AccessKey is used for uploading artifacts, instead IAM based auth is used
		cdStageWorkflowRequest.CdCacheRegion = cdWorkflowConfig.CdCacheRegion
		cdStageWorkflowRequest.CdCacheLocation = cdWorkflowConfig.CdCacheBucket
		cdStageWorkflowRequest.ArtifactLocation, cdStageWorkflowRequest.CiArtifactBucket, cdStageWorkflowRequest.CiArtifactFileName = impl.buildArtifactLocationForS3(cdWorkflowConfig, cdWf, runner)
		cdStageWorkflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			AccessKey:                  impl.config.BlobStorageS3AccessKey,
			Passkey:                    impl.config.BlobStorageS3SecretKey,
			EndpointUrl:                impl.config.BlobStorageS3Endpoint,
			IsInSecure:                 impl.config.BlobStorageS3EndpointInsecure,
			CiCacheBucketName:          cdWorkflowConfig.CdCacheBucket,
			CiCacheRegion:              cdWorkflowConfig.CdCacheRegion,
			CiCacheBucketVersioning:    impl.config.BlobStorageS3BucketVersioned,
			CiArtifactBucketName:       cdStageWorkflowRequest.CiArtifactBucket,
			CiArtifactRegion:           cdWorkflowConfig.CdCacheRegion,
			CiArtifactBucketVersioning: impl.config.BlobStorageS3BucketVersioned,
			CiLogBucketName:            impl.config.GetDefaultBuildLogsBucket(),
			CiLogRegion:                impl.config.GetDefaultCdLogsBucketRegion(),
			CiLogBucketVersioning:      impl.config.BlobStorageS3BucketVersioned,
		}
	case BLOB_STORAGE_GCP:
		cdStageWorkflowRequest.GcpBlobConfig = &blob_storage.GcpBlobConfig{
			CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
			ArtifactBucketName:     impl.config.GetDefaultBuildLogsBucket(),
			LogBucketName:          impl.config.GetDefaultBuildLogsBucket(),
		}
		cdStageWorkflowRequest.ArtifactLocation = impl.buildDefaultArtifactLocation(cdWorkflowConfig, cdWf, runner)
		cdStageWorkflowRequest.CiArtifactFileName = cdStageWorkflowRequest.ArtifactLocation
	case BLOB_STORAGE_AZURE:
		cdStageWorkflowRequest.AzureBlobConfig = &blob_storage.AzureBlobConfig{
			Enabled:               true,
			AccountName:           impl.config.AzureAccountName,
			BlobContainerCiCache:  impl.config.AzureBlobContainerCiCache,
			AccountKey:            impl.config.AzureAccountKey,
			BlobContainerCiLog:    impl.config.AzureBlobContainerCiLog,
			BlobContainerArtifact: impl.config.AzureBlobContainerCiLog,
		}
		cdStageWorkflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			EndpointUrl:     impl.config.AzureGatewayUrl,
			IsInSecure:      impl.config.AzureGatewayConnectionInsecure,
			CiLogBucketName: impl.config.AzureBlobContainerCiLog,
			CiLogRegion:     "",
			AccessKey:       impl.config.AzureAccountName,
		}
		cdStageWorkflowRequest.ArtifactLocation = impl.buildDefaultArtifactLocation(cdWorkflowConfig, cdWf, runner)
		cdStageWorkflowRequest.CiArtifactFileName = cdStageWorkflowRequest.ArtifactLocation
	default:
		if impl.config.BlobStorageEnabled {
			return nil, fmt.Errorf("blob storage %s not supported", cdStageWorkflowRequest.CloudProvider)
		}
	}
	cdStageWorkflowRequest.DefaultAddressPoolBaseCidr = impl.config.GetDefaultAddressPoolBaseCidr()
	cdStageWorkflowRequest.DefaultAddressPoolSize = impl.config.GetDefaultAddressPoolSize()
	return cdStageWorkflowRequest, nil
}

func (impl *WorkflowDagExecutorImpl) buildDefaultArtifactLocation(cdWorkflowConfig *pipelineConfig.CdWorkflowConfig, savedWf *pipelineConfig.CdWorkflow, runner *pipelineConfig.CdWorkflowRunner) string {
	cdArtifactLocationFormat := cdWorkflowConfig.CdArtifactLocationFormat
	if cdArtifactLocationFormat == "" {
		cdArtifactLocationFormat = impl.config.GetArtifactLocationFormat()
	}
	ArtifactLocation := fmt.Sprintf("%s/"+cdArtifactLocationFormat, impl.config.GetDefaultArtifactKeyPrefix(), savedWf.Id, runner.Id)
	return ArtifactLocation
}

func (impl *WorkflowDagExecutorImpl) HandleDeploymentSuccessEvent(pipelineOverride *chartConfig.PipelineOverride) error {
	if pipelineOverride == nil {
		return fmt.Errorf("invalid request, pipeline override not found")
	}
	cdWorkflow, err := impl.cdWorkflowRepository.FindById(pipelineOverride.CdWorkflowId)
	if err != nil {
		impl.logger.Errorw("error in fetching cd workflow by id", "pipelineOverride", pipelineOverride)
		return err
	}

	postStage, err := impl.getPipelineStage(pipelineOverride.PipelineId, repository4.PIPELINE_STAGE_TYPE_POST_CD)
	if err != nil {
		return err
	}

	var triggeredByUser int32 = 1
	//handle corrupt data (https://github.com/devtron-labs/devtron/issues/3826)
	err, deleted := impl.deleteCorruptedPipelineStage(postStage, triggeredByUser)
	if err != nil {
		impl.logger.Errorw("error in deleteCorruptedPipelineStage ", "err", err, "preStage", postStage, "triggeredBy", triggeredByUser)
		return err
	}

	if len(pipelineOverride.Pipeline.PostStageConfig) > 0 || (postStage != nil && !deleted) {
		if pipelineOverride.Pipeline.PostTriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC &&
			pipelineOverride.DeploymentType != models.DEPLOYMENTTYPE_STOP &&
			pipelineOverride.DeploymentType != models.DEPLOYMENTTYPE_START {

			err = impl.TriggerPostStage(cdWorkflow, pipelineOverride.Pipeline, triggeredByUser, 0)
			if err != nil {
				impl.logger.Errorw("error in triggering post stage after successful deployment event", "err", err, "cdWorkflow", cdWorkflow)
				return err
			}
		}
	} else {
		// to trigger next pre/cd, if any
		// finding children cd by pipeline id
		err = impl.HandlePostStageSuccessEvent(cdWorkflow.Id, pipelineOverride.PipelineId, 1)
		if err != nil {
			impl.logger.Errorw("error in triggering children cd after successful deployment event", "parentCdPipelineId", pipelineOverride.PipelineId)
			return err
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) HandlePostStageSuccessEvent(cdWorkflowId int, cdPipelineId int, triggeredBy int32) error {
	// finding children cd by pipeline id
	cdPipelinesMapping, err := impl.appWorkflowRepository.FindWFCDMappingByParentCDPipelineId(cdPipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting mapping of cd pipelines by parent cd pipeline id", "err", err, "parentCdPipelineId", cdPipelineId)
		return err
	}
	ciArtifact, err := impl.ciArtifactRepository.GetArtifactByCdWorkflowId(cdWorkflowId)
	if err != nil {
		impl.logger.Errorw("error in finding artifact by cd workflow id", "err", err, "cdWorkflowId", cdWorkflowId)
		return err
	}
	//TODO : confirm about this logic used for applyAuth
	applyAuth := false
	if triggeredBy != 1 {
		applyAuth = true
	}
	for _, cdPipelineMapping := range cdPipelinesMapping {
		//find pipeline by cdPipeline ID
		pipeline, err := impl.pipelineRepository.FindById(cdPipelineMapping.ComponentId)
		if err != nil {
			impl.logger.Errorw("error in getting cd pipeline by id", "err", err, "pipelineId", cdPipelineMapping.ComponentId)
			return err
		}
		//finding ci artifact by ciPipelineID and pipelineId
		//TODO : confirm values for applyAuth, async & triggeredBy
		err = impl.triggerStage(nil, pipeline, ciArtifact, applyAuth, triggeredBy)
		if err != nil {
			impl.logger.Errorw("error in triggering cd pipeline after successful post stage", "err", err, "pipelineId", pipeline.Id)
			return err
		}
	}
	return nil
}

// Only used for auto trigger
func (impl *WorkflowDagExecutorImpl) TriggerDeployment(cdWf *pipelineConfig.CdWorkflow, artifact *repository.CiArtifact, pipeline *pipelineConfig.Pipeline, applyAuth bool, triggeredBy int32) error {
	//in case of manual ci RBAC need to apply, this method used for auto cd deployment
	if applyAuth {
		user, err := impl.user.GetById(triggeredBy)
		if err != nil {
			impl.logger.Errorw("error in fetching user for auto pipeline", "UpdatedBy", artifact.UpdatedBy)
			return nil
		}
		token := user.EmailId
		object := impl.enforcerUtil.GetAppRBACNameByAppId(pipeline.AppId)
		impl.logger.Debugw("Triggered Request (App Permission Checking):", "object", object)
		if ok := impl.enforcer.EnforceByEmail(strings.ToLower(token), casbin.ResourceApplications, casbin.ActionTrigger, object); !ok {
			err = &util.ApiError{Code: "401", HttpStatusCode: 401, UserMessage: "unauthorized for pipeline " + strconv.Itoa(pipeline.Id)}
			return err
		}
	}

	//setting triggeredAt variable to have consistent data for various audit log places in db for deployment time
	triggeredAt := time.Now()

	if cdWf == nil {
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
		WorkflowType: bean.CD_WORKFLOW_TYPE_DEPLOY,
		ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM,
		Status:       pipelineConfig.WorkflowInProgress, //starting
		TriggeredBy:  1,
		StartedOn:    triggeredAt,
		Namespace:    impl.config.GetDefaultNamespace(),
		CdWorkflowId: cdWf.Id,
		AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: triggeredBy, UpdatedOn: triggeredAt, UpdatedBy: triggeredBy},
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
		runner.Status = pipelineConfig.WorkflowFailed
		runner.Message = "Found vulnerability on image"
		runner.FinishedOn = time.Now()
		runner.UpdatedOn = time.Now()
		runner.UpdatedBy = triggeredBy
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
		if err != nil {
			impl.logger.Errorw("error in updating status", "err", err)
			return err
		}
		cdMetrics := util4.CDMetrics{
			AppName:         runner.CdWorkflow.Pipeline.DeploymentAppName,
			Status:          runner.Status,
			DeploymentType:  runner.CdWorkflow.Pipeline.DeploymentAppType,
			EnvironmentName: runner.CdWorkflow.Pipeline.Environment.Name,
			Time:            time.Since(runner.StartedOn).Seconds() - time.Since(runner.FinishedOn).Seconds(),
		}
		util4.TriggerCDMetrics(cdMetrics, impl.config.ExposeCDMetrics)
		// creating cd pipeline status timeline for deployment failed
		timeline := &pipelineConfig.PipelineStatusTimeline{
			CdWorkflowRunnerId: runner.Id,
			Status:             pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_FAILED,
			StatusDetail:       "Deployment failed: Vulnerability policy violated.",
			StatusTime:         time.Now(),
			AuditLog: sql.AuditLog{
				CreatedBy: 1,
				CreatedOn: time.Now(),
				UpdatedBy: 1,
				UpdatedOn: time.Now(),
			},
		}
		err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, isAppStore)
		if err != nil {
			impl.logger.Errorw("error in creating timeline status for deployment fail - cve policy violation", "err", err, "timeline", timeline)
		}
		return nil
	}

	err = impl.TriggerCD(artifact, cdWf.Id, savedWfr.Id, pipeline, triggeredAt)
	err1 := impl.updatePreviousDeploymentStatus(runner, pipeline.Id, err, triggeredAt, triggeredBy)
	if err1 != nil || err != nil {
		impl.logger.Errorw("error while update previous cd workflow runners", "err", err, "runner", runner, "pipelineId", pipeline.Id)
		return err
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) updatePreviousDeploymentStatus(currentRunner *pipelineConfig.CdWorkflowRunner, pipelineId int, err error, triggeredAt time.Time, triggeredBy int32) error {
	if err != nil {
		//creating cd pipeline status timeline for deployment failed
		terminalStatusExists, timelineErr := impl.cdPipelineStatusTimelineRepo.CheckIfTerminalStatusTimelinePresentByWfrId(currentRunner.Id)
		if timelineErr != nil {
			impl.logger.Errorw("error in checking if terminal status timeline exists by wfrId", "err", timelineErr, "wfrId", currentRunner.Id)
			return timelineErr
		}
		if !terminalStatusExists {
			impl.logger.Infow("marking pipeline deployment failed", "err", err)
			timeline := &pipelineConfig.PipelineStatusTimeline{
				CdWorkflowRunnerId: currentRunner.Id,
				Status:             pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_FAILED,
				StatusDetail:       fmt.Sprintf("Deployment failed: %v", err),
				StatusTime:         time.Now(),
				AuditLog: sql.AuditLog{
					CreatedBy: 1,
					CreatedOn: time.Now(),
					UpdatedBy: 1,
					UpdatedOn: time.Now(),
				},
			}
			timelineErr = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
			if timelineErr != nil {
				impl.logger.Errorw("error in creating timeline status for deployment fail", "err", timelineErr, "timeline", timeline)
			}
		}
		impl.logger.Errorw("error in triggering cd WF, setting wf status as fail ", "wfId", currentRunner.Id, "err", err)
		currentRunner.Status = pipelineConfig.WorkflowFailed
		currentRunner.Message = err.Error()
		currentRunner.FinishedOn = triggeredAt
		currentRunner.UpdatedOn = time.Now()
		currentRunner.UpdatedBy = triggeredBy
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(currentRunner)
		if err != nil {
			impl.logger.Errorw("error updating cd wf runner status", "err", err, "currentRunner", currentRunner)
			return err
		}
		cdMetrics := util4.CDMetrics{
			AppName:         currentRunner.CdWorkflow.Pipeline.DeploymentAppName,
			Status:          currentRunner.Status,
			DeploymentType:  currentRunner.CdWorkflow.Pipeline.DeploymentAppType,
			EnvironmentName: currentRunner.CdWorkflow.Pipeline.Environment.Name,
			Time:            time.Since(currentRunner.StartedOn).Seconds() - time.Since(currentRunner.FinishedOn).Seconds(),
		}
		util4.TriggerCDMetrics(cdMetrics, impl.config.ExposeCDMetrics)
		return nil
		//update current WF with error status
	} else {
		//update [n,n-1] statuses as failed if not terminal
		terminalStatus := []string{string(health.HealthStatusHealthy), pipelineConfig.WorkflowAborted, pipelineConfig.WorkflowFailed, pipelineConfig.WorkflowSucceeded}
		previousNonTerminalRunners, err := impl.cdWorkflowRepository.FindPreviousCdWfRunnerByStatus(pipelineId, currentRunner.Id, terminalStatus)
		if err != nil {
			impl.logger.Errorw("error fetching previous wf runner, updating cd wf runner status,", "err", err, "currentRunner", currentRunner)
			return err
		} else if len(previousNonTerminalRunners) == 0 {
			impl.logger.Errorw("no previous runner found in updating cd wf runner status,", "err", err, "currentRunner", currentRunner)
			return nil
		}
		dbConnection := impl.cdWorkflowRepository.GetConnection()
		tx, err := dbConnection.Begin()
		if err != nil {
			impl.logger.Errorw("error on update status, txn begin failed", "err", err)
			return err
		}
		// Rollback tx on error.
		defer tx.Rollback()
		var timelines []*pipelineConfig.PipelineStatusTimeline
		for _, previousRunner := range previousNonTerminalRunners {
			if previousRunner.Status == string(health.HealthStatusHealthy) ||
				previousRunner.Status == pipelineConfig.WorkflowSucceeded ||
				previousRunner.Status == pipelineConfig.WorkflowAborted ||
				previousRunner.Status == pipelineConfig.WorkflowFailed {
				//terminal status return
				impl.logger.Infow("skip updating cd wf runner status as previous runner status is", "status", previousRunner.Status)
				continue
			}
			impl.logger.Infow("updating cd wf runner status as previous runner status is", "status", previousRunner.Status)
			previousRunner.FinishedOn = triggeredAt
			previousRunner.Message = "A new deployment was initiated before this deployment completed"
			previousRunner.Status = pipelineConfig.WorkflowFailed
			previousRunner.UpdatedOn = time.Now()
			previousRunner.UpdatedBy = triggeredBy
			timeline := &pipelineConfig.PipelineStatusTimeline{
				CdWorkflowRunnerId: previousRunner.Id,
				Status:             pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_SUPERSEDED,
				StatusDetail:       "This deployment is superseded.",
				StatusTime:         time.Now(),
				AuditLog: sql.AuditLog{
					CreatedBy: 1,
					CreatedOn: time.Now(),
					UpdatedBy: 1,
					UpdatedOn: time.Now(),
				},
			}
			timelines = append(timelines, timeline)
		}

		err = impl.cdWorkflowRepository.UpdateWorkFlowRunnersWithTxn(previousNonTerminalRunners, tx)
		if err != nil {
			impl.logger.Errorw("error updating cd wf runner status", "err", err, "previousNonTerminalRunners", previousNonTerminalRunners)
			return err
		}
		err = impl.cdPipelineStatusTimelineRepo.SaveTimelinesWithTxn(timelines, tx)
		if err != nil {
			impl.logger.Errorw("error updating pipeline status timelines", "err", err, "timelines", timelines)
			return err
		}
		err = tx.Commit()
		if err != nil {
			impl.logger.Errorw("error in db transaction commit", "err", err)
			return err
		}
		return nil
	}
}

type RequestType string

const START RequestType = "START"
const STOP RequestType = "STOP"

type StopAppRequest struct {
	AppId         int         `json:"appId" validate:"required"`
	EnvironmentId int         `json:"environmentId" validate:"required"`
	UserId        int32       `json:"userId"`
	RequestType   RequestType `json:"requestType" validate:"oneof=START STOP"`
}

type StopDeploymentGroupRequest struct {
	DeploymentGroupId int         `json:"deploymentGroupId" validate:"required"`
	UserId            int32       `json:"userId"`
	RequestType       RequestType `json:"requestType" validate:"oneof=START STOP"`
}

type PodRotateRequest struct {
	AppId               int                        `json:"appId" validate:"required"`
	EnvironmentId       int                        `json:"environmentId" validate:"required"`
	UserId              int32                      `json:"-"`
	ResourceIdentifiers []util5.ResourceIdentifier `json:"resources" validate:"required"`
}

func (impl *WorkflowDagExecutorImpl) RotatePods(ctx context.Context, podRotateRequest *PodRotateRequest) (*k8s.RotatePodResponse, error) {
	impl.logger.Infow("rotate pod request", "payload", podRotateRequest)
	//extract cluster id and namespace from env id
	environmentId := podRotateRequest.EnvironmentId
	environment, err := impl.envRepository.FindById(environmentId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching env details", "envId", environmentId, "err", err)
		return nil, err
	}
	var resourceIdentifiers []util5.ResourceIdentifier
	for _, resourceIdentifier := range podRotateRequest.ResourceIdentifiers {
		resourceIdentifier.Namespace = environment.Namespace
		resourceIdentifiers = append(resourceIdentifiers, resourceIdentifier)
	}
	rotatePodRequest := &k8s.RotatePodRequest{
		ClusterId: environment.ClusterId,
		Resources: resourceIdentifiers,
	}
	response, err := impl.k8sCommonService.RotatePods(ctx, rotatePodRequest)
	if err != nil {
		return nil, err
	}
	//TODO KB: make entry in cd workflow runner
	return response, nil
}

func (impl *WorkflowDagExecutorImpl) StopStartApp(stopRequest *StopAppRequest, ctx context.Context) (int, error) {
	pipelines, err := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(stopRequest.AppId, stopRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline", "app", stopRequest.AppId, "env", stopRequest.EnvironmentId, "err", err)
		return 0, err
	}
	if len(pipelines) == 0 {
		return 0, fmt.Errorf("no pipeline found")
	}
	pipeline := pipelines[0]

	//find pipeline with default
	var pipelineIds []int
	for _, p := range pipelines {
		impl.logger.Debugw("adding pipelineId", "pipelineId", p.Id)
		pipelineIds = append(pipelineIds, p.Id)
		//FIXME
	}
	wf, err := impl.cdWorkflowRepository.FindLatestCdWorkflowByPipelineId(pipelineIds)
	if err != nil {
		impl.logger.Errorw("error in fetching latest release", "err", err)
		return 0, err
	}
	stopTemplate := `{"replicaCount":0,"autoscaling":{"MinReplicas":0,"MaxReplicas":0 ,"enabled": false} }`
	overrideRequest := &bean.ValuesOverrideRequest{
		PipelineId:     pipeline.Id,
		AppId:          stopRequest.AppId,
		CiArtifactId:   wf.CiArtifactId,
		UserId:         stopRequest.UserId,
		CdWorkflowType: bean.CD_WORKFLOW_TYPE_DEPLOY,
	}
	if stopRequest.RequestType == STOP {
		overrideRequest.AdditionalOverride = json.RawMessage([]byte(stopTemplate))
		overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_STOP
	} else if stopRequest.RequestType == START {
		overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_START
	} else {
		return 0, fmt.Errorf("unsupported operation %s", stopRequest.RequestType)
	}
	id, err := impl.ManualCdTrigger(overrideRequest, ctx)
	if err != nil {
		impl.logger.Errorw("error in stopping app", "err", err, "appId", stopRequest.AppId, "envId", stopRequest.EnvironmentId)
		return 0, err
	}
	return id, err
}

func (impl *WorkflowDagExecutorImpl) GetArtifactVulnerabilityStatus(artifact *repository.CiArtifact, cdPipeline *pipelineConfig.Pipeline, ctx context.Context) (bool, error) {
	isVulnerable := false
	if len(artifact.ImageDigest) > 0 {
		var cveStores []*security.CveStore
		_, span := otel.Tracer("orchestrator").Start(ctx, "scanResultRepository.FindByImageDigest")
		imageScanResult, err := impl.scanResultRepository.FindByImageDigest(artifact.ImageDigest)
		span.End()
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error fetching image digest", "digest", artifact.ImageDigest, "err", err)
			return false, err
		}
		for _, item := range imageScanResult {
			cveStores = append(cveStores, &item.CveStore)
		}
		_, span = otel.Tracer("orchestrator").Start(ctx, "cvePolicyRepository.GetBlockedCVEList")
		if cdPipeline.Environment.ClusterId == 0 {
			envDetails, err := impl.envRepository.FindById(cdPipeline.EnvironmentId)
			if err != nil {
				impl.logger.Errorw("error fetching cluster details by env, GetArtifactVulnerabilityStatus", "envId", cdPipeline.EnvironmentId, "err", err)
				return false, err
			}
			cdPipeline.Environment = *envDetails
		}
		blockCveList, err := impl.cvePolicyRepository.GetBlockedCVEList(cveStores, cdPipeline.Environment.ClusterId, cdPipeline.EnvironmentId, cdPipeline.AppId, false)
		span.End()
		if err != nil {
			impl.logger.Errorw("error while fetching env", "err", err)
			return false, err
		}
		if len(blockCveList) > 0 {
			isVulnerable = true
		}
	}
	return isVulnerable, nil
}

func (impl *WorkflowDagExecutorImpl) ManualCdTrigger(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context) (int, error) {
	//setting triggeredAt variable to have consistent data for various audit log places in db for deployment time
	triggeredAt := time.Now()
	releaseId := 0

	var err error
	_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineRepository.FindById")
	cdPipeline, err := impl.pipelineRepository.FindById(overrideRequest.PipelineId)
	span.End()
	if err != nil {
		impl.logger.Errorf("invalid req", "err", err, "req", overrideRequest)
		return 0, err
	}
	impl.SetPipelineFieldsInOverrideRequest(overrideRequest, cdPipeline)

	if overrideRequest.CdWorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
		_, span = otel.Tracer("orchestrator").Start(ctx, "ciArtifactRepository.Get")
		artifact, err := impl.ciArtifactRepository.Get(overrideRequest.CiArtifactId)
		span.End()
		if err != nil {
			impl.logger.Errorw("err", "err", err)
			return 0, err
		}
		_, span = otel.Tracer("orchestrator").Start(ctx, "TriggerPreStage")
		err = impl.TriggerPreStage(ctx, nil, artifact, cdPipeline, overrideRequest.UserId, false, 0)
		span.End()
		if err != nil {
			impl.logger.Errorw("err", "err", err)
			return 0, err
		}
	} else if overrideRequest.CdWorkflowType == bean.CD_WORKFLOW_TYPE_DEPLOY {
		if overrideRequest.DeploymentType == models.DEPLOYMENTTYPE_UNKNOWN {
			overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_DEPLOY
		}
		cdWf, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(ctx, overrideRequest.CdWorkflowId, bean.CD_WORKFLOW_TYPE_PRE)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("err", "err", err)
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
				impl.logger.Errorw("err", "err", err)
				return 0, err
			}
			cdWorkflowId = cdWf.Id
		}

		runner := &pipelineConfig.CdWorkflowRunner{
			Name:         cdPipeline.Name,
			WorkflowType: bean.CD_WORKFLOW_TYPE_DEPLOY,
			ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
			Status:       pipelineConfig.WorkflowInProgress,
			TriggeredBy:  overrideRequest.UserId,
			StartedOn:    triggeredAt,
			Namespace:    impl.config.GetDefaultNamespace(),
			CdWorkflowId: cdWorkflowId,
			AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
		}
		savedWfr, err := impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
		overrideRequest.WfrId = savedWfr.Id
		if err != nil {
			impl.logger.Errorw("err", "err", err)
			return 0, err
		}
		runner.CdWorkflow = &pipelineConfig.CdWorkflow{
			Pipeline: cdPipeline,
		}
		overrideRequest.CdWorkflowId = cdWorkflowId
		// creating cd pipeline status timeline for deployment initialisation
		timeline := impl.pipelineStatusTimelineService.GetTimelineDbObjectByTimelineStatusAndTimelineDescription(savedWfr.Id, pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_INITIATED, pipelineConfig.TIMELINE_DESCRIPTION_DEPLOYMENT_INITIATED, overrideRequest.UserId)
		_, span = otel.Tracer("orchestrator").Start(ctx, "cdPipelineStatusTimelineRepo.SaveTimelineForACDHelmApps")
		err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)

		span.End()
		if err != nil {
			impl.logger.Errorw("error in creating timeline status for deployment initiation", "err", err, "timeline", timeline)
		}

		//checking vulnerability for deploying image
		_, span = otel.Tracer("orchestrator").Start(ctx, "ciArtifactRepository.Get")
		artifact, err := impl.ciArtifactRepository.Get(overrideRequest.CiArtifactId)
		span.End()
		if err != nil {
			impl.logger.Errorw("err", "err", err)
			return 0, err
		}
		isVulnerable, err := impl.GetArtifactVulnerabilityStatus(artifact, cdPipeline, ctx)
		if err != nil {
			impl.logger.Errorw("error in getting Artifact vulnerability status, ManualCdTrigger", "err", err)
			return 0, err
		}
		if isVulnerable == true {
			// if image vulnerable, update timeline status and return
			runner.Status = pipelineConfig.WorkflowFailed
			runner.Message = "Found vulnerability on image"
			runner.FinishedOn = time.Now()
			runner.UpdatedOn = time.Now()
			runner.UpdatedBy = overrideRequest.UserId
			err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
			if err != nil {
				impl.logger.Errorw("error in updating wfr status due to vulnerable image", "err", err)
				return 0, err
			}
			runner.CdWorkflow = &pipelineConfig.CdWorkflow{
				Pipeline: cdPipeline,
			}
			cdMetrics := util4.CDMetrics{
				AppName:         runner.CdWorkflow.Pipeline.DeploymentAppName,
				Status:          runner.Status,
				DeploymentType:  runner.CdWorkflow.Pipeline.DeploymentAppType,
				EnvironmentName: runner.CdWorkflow.Pipeline.Environment.Name,
				Time:            time.Since(runner.StartedOn).Seconds() - time.Since(runner.FinishedOn).Seconds(),
			}
			util4.TriggerCDMetrics(cdMetrics, impl.config.ExposeCDMetrics)
			// creating cd pipeline status timeline for deployment failed
			timeline := impl.pipelineStatusTimelineService.GetTimelineDbObjectByTimelineStatusAndTimelineDescription(runner.Id, pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_FAILED, pipelineConfig.TIMELINE_DESCRIPTION_VULNERABLE_IMAGE, 1)

			_, span = otel.Tracer("orchestrator").Start(ctx, "cdPipelineStatusTimelineRepo.SaveTimelineForACDHelmApps")
			err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
			span.End()
			if err != nil {
				impl.logger.Errorw("error in creating timeline status for deployment fail - cve policy violation", "err", err, "timeline", timeline)
			}
			_, span = otel.Tracer("orchestrator").Start(ctx, "updatePreviousDeploymentStatus")
			err1 := impl.updatePreviousDeploymentStatus(runner, cdPipeline.Id, err, triggeredAt, overrideRequest.UserId)
			span.End()
			if err1 != nil {
				impl.logger.Errorw("error while update previous cd workflow runners", "err", err, "runner", runner, "pipelineId", cdPipeline.Id)
			}
			return 0, fmt.Errorf("found vulnerability for image digest %s", artifact.ImageDigest)
		}
		_, span = otel.Tracer("orchestrator").Start(ctx, "appService.TriggerRelease")
		releaseId, _, err = impl.TriggerRelease(overrideRequest, ctx, triggeredAt, overrideRequest.UserId)
		span.End()

		if overrideRequest.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_MANIFEST_DOWNLOAD {
			runner := &pipelineConfig.CdWorkflowRunner{
				Id:           runner.Id,
				Name:         cdPipeline.Name,
				WorkflowType: bean.CD_WORKFLOW_TYPE_DEPLOY,
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
				impl.logger.Errorw("error in updating runner for manifest_download type", "err", err)
				return 0, updateErr
			}
		}

		_, span = otel.Tracer("orchestrator").Start(ctx, "updatePreviousDeploymentStatus")
		err1 := impl.updatePreviousDeploymentStatus(runner, cdPipeline.Id, err, triggeredAt, overrideRequest.UserId)
		span.End()
		if err1 != nil || err != nil {
			impl.logger.Errorw("error while update previous cd workflow runners", "err", err, "runner", runner, "pipelineId", cdPipeline.Id)
			return 0, err
		}
	} else if overrideRequest.CdWorkflowType == bean.CD_WORKFLOW_TYPE_POST {
		cdWfRunner, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(ctx, overrideRequest.CdWorkflowId, bean.CD_WORKFLOW_TYPE_DEPLOY)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("err", "err", err)
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
				impl.logger.Errorw("err", "err", err)
				return 0, err
			}
		} else {
			_, span = otel.Tracer("orchestrator").Start(ctx, "cdWorkflowRepository.FindById")
			cdWf, err = impl.cdWorkflowRepository.FindById(overrideRequest.CdWorkflowId)
			span.End()
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("err", "err", err)
				return 0, err
			}
		}
		_, span = otel.Tracer("orchestrator").Start(ctx, "TriggerPostStage")
		err = impl.TriggerPostStage(cdWf, cdPipeline, overrideRequest.UserId, 0)
		span.End()
		if err != nil {
			impl.logger.Errorw("err", "err", err)
			return 0, err
		}
	}
	return releaseId, err
}

type BulkTriggerRequest struct {
	CiArtifactId int `sql:"ci_artifact_id"`
	PipelineId   int `sql:"pipeline_id"`
}

func (impl *WorkflowDagExecutorImpl) TriggerBulkDeploymentAsync(requests []*BulkTriggerRequest, UserId int32) (interface{}, error) {
	var cdWorkflows []*pipelineConfig.CdWorkflow
	for _, request := range requests {
		cdWf := &pipelineConfig.CdWorkflow{
			CiArtifactId:   request.CiArtifactId,
			PipelineId:     request.PipelineId,
			AuditLog:       sql.AuditLog{CreatedOn: time.Now(), CreatedBy: UserId, UpdatedOn: time.Now(), UpdatedBy: UserId},
			WorkflowStatus: pipelineConfig.REQUEST_ACCEPTED,
		}
		cdWorkflows = append(cdWorkflows, cdWf)
	}
	err := impl.cdWorkflowRepository.SaveWorkFlows(cdWorkflows...)
	if err != nil {
		impl.logger.Errorw("error in saving wfs", "req", requests, "err", err)
		return nil, err
	}
	impl.triggerNatsEventForBulkAction(cdWorkflows)
	return nil, nil
	//return
	//publish nats async
	//update status
	//consume message
}

type DeploymentGroupAppWithEnv struct {
	EnvironmentId     int         `json:"environmentId"`
	DeploymentGroupId int         `json:"deploymentGroupId"`
	AppId             int         `json:"appId"`
	Active            bool        `json:"active"`
	UserId            int32       `json:"userId"`
	RequestType       RequestType `json:"requestType" validate:"oneof=START STOP"`
}

func (impl *WorkflowDagExecutorImpl) TriggerBulkHibernateAsync(request StopDeploymentGroupRequest, ctx context.Context) (interface{}, error) {
	dg, err := impl.groupRepository.FindByIdWithApp(request.DeploymentGroupId)
	if err != nil {
		impl.logger.Errorw("error while fetching dg", "err", err)
		return nil, err
	}

	for _, app := range dg.DeploymentGroupApps {
		deploymentGroupAppWithEnv := &DeploymentGroupAppWithEnv{
			AppId:             app.AppId,
			EnvironmentId:     dg.EnvironmentId,
			DeploymentGroupId: dg.Id,
			Active:            dg.Active,
			UserId:            request.UserId,
			RequestType:       request.RequestType,
		}

		data, err := json.Marshal(deploymentGroupAppWithEnv)
		if err != nil {
			impl.logger.Errorw("error while writing app stop event to nats ", "app", app.AppId, "deploymentGroup", app.DeploymentGroupId, "err", err)
		} else {
			err = impl.pubsubClient.Publish(pubsub.BULK_HIBERNATE_TOPIC, string(data))
			if err != nil {
				impl.logger.Errorw("Error while publishing request", "topic", pubsub.BULK_HIBERNATE_TOPIC, "error", err)
			}
		}
	}
	return nil, nil
}

func (impl *WorkflowDagExecutorImpl) triggerNatsEventForBulkAction(cdWorkflows []*pipelineConfig.CdWorkflow) {
	for _, wf := range cdWorkflows {
		data, err := json.Marshal(wf)
		if err != nil {
			wf.WorkflowStatus = pipelineConfig.QUE_ERROR
		} else {
			err = impl.pubsubClient.Publish(pubsub.BULK_DEPLOY_TOPIC, string(data))
			if err != nil {
				wf.WorkflowStatus = pipelineConfig.QUE_ERROR
			} else {
				wf.WorkflowStatus = pipelineConfig.ENQUEUED
			}
		}
		err = impl.cdWorkflowRepository.UpdateWorkFlow(wf)
		if err != nil {
			impl.logger.Errorw("error in publishing wf msg", "wf", wf, "err", err)
		}
	}
}

func (impl *WorkflowDagExecutorImpl) subscribeTriggerBulkAction() error {
	callback := func(msg *pubsub.PubSubMsg) {
		impl.logger.Debug("subscribeTriggerBulkAction event received")
		//defer msg.Ack()
		cdWorkflow := new(pipelineConfig.CdWorkflow)
		err := json.Unmarshal([]byte(string(msg.Data)), cdWorkflow)
		if err != nil {
			impl.logger.Error("Error while unmarshalling cdWorkflow json object", "error", err)
			return
		}
		impl.logger.Debugw("subscribeTriggerBulkAction event:", "cdWorkflow", cdWorkflow)
		wf := &pipelineConfig.CdWorkflow{
			Id:           cdWorkflow.Id,
			CiArtifactId: cdWorkflow.CiArtifactId,
			PipelineId:   cdWorkflow.PipelineId,
			AuditLog: sql.AuditLog{
				UpdatedOn: time.Now(),
			},
		}
		latest, err := impl.cdWorkflowRepository.IsLatestWf(cdWorkflow.PipelineId, cdWorkflow.Id)
		if err != nil {
			impl.logger.Errorw("error in determining latest", "wf", cdWorkflow, "err", err)
			wf.WorkflowStatus = pipelineConfig.DEQUE_ERROR
			impl.cdWorkflowRepository.UpdateWorkFlow(wf)
			return
		}
		if !latest {
			wf.WorkflowStatus = pipelineConfig.DROPPED_STALE
			impl.cdWorkflowRepository.UpdateWorkFlow(wf)
			return
		}
		pipeline, err := impl.pipelineRepository.FindById(cdWorkflow.PipelineId)
		if err != nil {
			impl.logger.Errorw("error in fetching pipeline", "err", err)
			wf.WorkflowStatus = pipelineConfig.TRIGGER_ERROR
			impl.cdWorkflowRepository.UpdateWorkFlow(wf)
			return
		}
		artefact, err := impl.ciArtifactRepository.Get(cdWorkflow.CiArtifactId)
		if err != nil {
			impl.logger.Errorw("error in fetching artefact", "err", err)
			wf.WorkflowStatus = pipelineConfig.TRIGGER_ERROR
			impl.cdWorkflowRepository.UpdateWorkFlow(wf)
			return
		}
		err = impl.triggerStageForBulk(wf, pipeline, artefact, false, false, cdWorkflow.CreatedBy)
		if err != nil {
			impl.logger.Errorw("error in cd trigger ", "err", err)
			wf.WorkflowStatus = pipelineConfig.TRIGGER_ERROR
		} else {
			wf.WorkflowStatus = pipelineConfig.WF_STARTED
		}
		impl.cdWorkflowRepository.UpdateWorkFlow(wf)
	}
	err := impl.pubsubClient.Subscribe(pubsub.BULK_DEPLOY_TOPIC, callback)
	return err
}

func (impl *WorkflowDagExecutorImpl) subscribeHibernateBulkAction() error {
	callback := func(msg *pubsub.PubSubMsg) {
		impl.logger.Debug("subscribeHibernateBulkAction event received")
		//defer msg.Ack()
		deploymentGroupAppWithEnv := new(DeploymentGroupAppWithEnv)
		err := json.Unmarshal([]byte(string(msg.Data)), deploymentGroupAppWithEnv)
		if err != nil {
			impl.logger.Error("Error while unmarshalling deploymentGroupAppWithEnv json object", err)
			return
		}
		impl.logger.Debugw("subscribeHibernateBulkAction event:", "DeploymentGroupAppWithEnv", deploymentGroupAppWithEnv)

		stopAppRequest := &StopAppRequest{
			AppId:         deploymentGroupAppWithEnv.AppId,
			EnvironmentId: deploymentGroupAppWithEnv.EnvironmentId,
			UserId:        deploymentGroupAppWithEnv.UserId,
			RequestType:   deploymentGroupAppWithEnv.RequestType,
		}
		ctx, err := impl.buildACDContext()
		if err != nil {
			impl.logger.Errorw("error in creating acd synch context", "err", err)
			return
		}
		_, err = impl.StopStartApp(stopAppRequest, ctx)
		if err != nil {
			impl.logger.Errorw("error in stop app request", "err", err)
			return
		}
	}
	err := impl.pubsubClient.Subscribe(pubsub.BULK_HIBERNATE_TOPIC, callback)
	return err
}

func (impl *WorkflowDagExecutorImpl) buildACDContext() (acdContext context.Context, err error) {
	//this part only accessible for acd apps hibernation, if acd configured it will fetch latest acdToken, else it will return error
	acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		impl.logger.Errorw("error in getting acd token", "err", err)
		return nil, err
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "token", acdToken)
	return ctx, nil
}

func (impl *WorkflowDagExecutorImpl) TriggerRelease(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context, triggeredAt time.Time, deployedBy int32) (releaseNo int, manifest []byte, err error) {
	triggerEvent := impl.GetTriggerEvent(overrideRequest.DeploymentAppType, triggeredAt, deployedBy)
	releaseNo, manifest, err = impl.TriggerPipeline(overrideRequest, triggerEvent, ctx)
	if err != nil {
		return 0, manifest, err
	}
	return releaseNo, manifest, nil
}

func (impl *WorkflowDagExecutorImpl) TriggerCD(artifact *repository.CiArtifact, cdWorkflowId, wfrId int, pipeline *pipelineConfig.Pipeline, triggeredAt time.Time) error {
	impl.logger.Debugw("automatic pipeline trigger attempt async", "artifactId", artifact.Id)

	return impl.triggerReleaseAsync(artifact, cdWorkflowId, wfrId, pipeline, triggeredAt)
}

func (impl *WorkflowDagExecutorImpl) triggerReleaseAsync(artifact *repository.CiArtifact, cdWorkflowId, wfrId int, pipeline *pipelineConfig.Pipeline, triggeredAt time.Time) error {
	err := impl.validateAndTrigger(pipeline, artifact, cdWorkflowId, wfrId, triggeredAt)
	if err != nil {
		impl.logger.Errorw("error in trigger for pipeline", "pipelineId", strconv.Itoa(pipeline.Id))
	}
	impl.logger.Debugw("trigger attempted for all pipeline ", "artifactId", artifact.Id)
	return err
}

func (impl *WorkflowDagExecutorImpl) validateAndTrigger(p *pipelineConfig.Pipeline, artifact *repository.CiArtifact, cdWorkflowId, wfrId int, triggeredAt time.Time) error {
	object := impl.enforcerUtil.GetAppRBACNameByAppId(p.AppId)
	envApp := strings.Split(object, "/")
	if len(envApp) != 2 {
		impl.logger.Error("invalid req, app and env not found from rbac")
		return errors.New("invalid req, app and env not found from rbac")
	}
	err := impl.releasePipeline(p, artifact, cdWorkflowId, wfrId, triggeredAt)
	return err
}

func (impl *WorkflowDagExecutorImpl) releasePipeline(pipeline *pipelineConfig.Pipeline, artifact *repository.CiArtifact, cdWorkflowId, wfrId int, triggeredAt time.Time) error {
	impl.logger.Debugw("triggering release for ", "cdPipelineId", pipeline.Id, "artifactId", artifact.Id)

	pipeline, err := impl.pipelineRepository.FindById(pipeline.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline by pipelineId", "err", err)
		return err
	}

	request := &bean.ValuesOverrideRequest{
		PipelineId:           pipeline.Id,
		UserId:               artifact.CreatedBy,
		CiArtifactId:         artifact.Id,
		AppId:                pipeline.AppId,
		CdWorkflowId:         cdWorkflowId,
		ForceTrigger:         true,
		DeploymentWithConfig: bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED,
		WfrId:                wfrId,
	}
	impl.SetPipelineFieldsInOverrideRequest(request, pipeline)

	ctx, err := impl.buildACDContext()
	if err != nil {
		impl.logger.Errorw("error in creating acd synch context", "pipelineId", pipeline.Id, "artifactId", artifact.Id, "err", err)
		return err
	}
	//setting deployedBy as 1(system user) since case of auto trigger
	id, _, err := impl.TriggerRelease(request, ctx, triggeredAt, 1)
	if err != nil {
		impl.logger.Errorw("error in auto  cd pipeline trigger", "pipelineId", pipeline.Id, "artifactId", artifact.Id, "err", err)
	} else {
		impl.logger.Infow("pipeline successfully triggered ", "cdPipelineId", pipeline.Id, "artifactId", artifact.Id, "releaseId", id)
	}
	return err

}

func (impl *WorkflowDagExecutorImpl) SetPipelineFieldsInOverrideRequest(overrideRequest *bean.ValuesOverrideRequest, pipeline *pipelineConfig.Pipeline) {
	overrideRequest.PipelineId = pipeline.Id
	overrideRequest.PipelineName = pipeline.Name
	overrideRequest.EnvId = pipeline.EnvironmentId
	overrideRequest.EnvName = pipeline.Environment.Name
	overrideRequest.ClusterId = pipeline.Environment.ClusterId
	overrideRequest.AppId = pipeline.AppId
	overrideRequest.AppName = pipeline.App.AppName
	overrideRequest.DeploymentAppType = pipeline.DeploymentAppType
}

func (impl *WorkflowDagExecutorImpl) GetTriggerEvent(deploymentAppType string, triggeredAt time.Time, deployedBy int32) bean.TriggerEvent {
	// trigger event will decide whether to perform GitOps or deployment for a particular deployment app type
	triggerEvent := bean.TriggerEvent{
		TriggeredBy: deployedBy,
		TriggerdAt:  triggeredAt,
	}
	switch deploymentAppType {
	case bean2.ArgoCd:
		triggerEvent.PerformChartPush = true
		triggerEvent.PerformDeploymentOnCluster = true
		triggerEvent.GetManifestInResponse = false
		triggerEvent.DeploymentAppType = bean2.ArgoCd
		triggerEvent.ManifestStorageType = bean2.ManifestStorageGit
	case bean2.Helm:
		triggerEvent.PerformChartPush = false
		triggerEvent.PerformDeploymentOnCluster = true
		triggerEvent.GetManifestInResponse = false
		triggerEvent.DeploymentAppType = bean2.Helm
	}
	return triggerEvent
}

// write integration/unit test for each function
func (impl *WorkflowDagExecutorImpl) TriggerPipeline(overrideRequest *bean.ValuesOverrideRequest, triggerEvent bean.TriggerEvent, ctx context.Context) (releaseNo int, manifest []byte, err error) {

	isRequestValid, err := impl.ValidateTriggerEvent(triggerEvent)
	if !isRequestValid {
		return releaseNo, manifest, err
	}

	valuesOverrideResponse, builtChartPath, err := impl.BuildManifestForTrigger(overrideRequest, triggerEvent.TriggerdAt, ctx)
	_, span := otel.Tracer("orchestrator").Start(ctx, "CreateHistoriesForDeploymentTrigger")
	err1 := impl.CreateHistoriesForDeploymentTrigger(valuesOverrideResponse.Pipeline, valuesOverrideResponse.PipelineStrategy, valuesOverrideResponse.EnvOverride, triggerEvent.TriggerdAt, triggerEvent.TriggeredBy)
	if err1 != nil {
		impl.logger.Errorw("error in saving histories for trigger", "err", err1, "pipelineId", valuesOverrideResponse.Pipeline.Id, "wfrId", overrideRequest.WfrId)
	}
	span.End()
	if err != nil {
		return releaseNo, manifest, err
	}

	if triggerEvent.PerformChartPush {
		manifestPushTemplate, err := impl.BuildManifestPushTemplate(overrideRequest, valuesOverrideResponse, builtChartPath, &manifest)
		if err != nil {
			impl.logger.Errorw("error in building manifest push template", "err", err)
			return releaseNo, manifest, err
		}
		manifestPushService := impl.GetManifestPushService(triggerEvent)
		manifestPushResponse := manifestPushService.PushChart(manifestPushTemplate, ctx)
		if manifestPushResponse.Error != nil {
			impl.logger.Errorw("Error in pushing manifest to git", "err", err, "git_repo_url", manifestPushTemplate.RepoUrl)
			return releaseNo, manifest, err
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
		err = impl.DeployApp(overrideRequest, valuesOverrideResponse, triggerEvent.TriggerdAt, ctx)
		if err != nil {
			impl.logger.Errorw("error in deploying app", "err", err)
			return releaseNo, manifest, err
		}
	}

	go impl.WriteCDTriggerEvent(overrideRequest, valuesOverrideResponse.Artifact, valuesOverrideResponse.PipelineOverride.PipelineReleaseCounter, valuesOverrideResponse.PipelineOverride.Id)

	_, spann := otel.Tracer("orchestrator").Start(ctx, "MarkImageScanDeployed")
	_ = impl.MarkImageScanDeployed(overrideRequest.AppId, valuesOverrideResponse.EnvOverride.TargetEnvironment, valuesOverrideResponse.Artifact.ImageDigest, overrideRequest.ClusterId, valuesOverrideResponse.Artifact.ScanEnabled)
	spann.End()

	middleware.CdTriggerCounter.WithLabelValues(overrideRequest.AppName, overrideRequest.EnvName).Inc()

	return valuesOverrideResponse.PipelineOverride.PipelineReleaseCounter, manifest, nil

}

func (impl *WorkflowDagExecutorImpl) ValidateTriggerEvent(triggerEvent bean.TriggerEvent) (bool, error) {

	switch triggerEvent.DeploymentAppType {
	case bean2.ArgoCd:
		if !triggerEvent.PerformChartPush {
			return false, errors2.New("For deployment type ArgoCd, PerformChartPush flag expected value = true, got false")
		}
	case bean2.Helm:
		return true, nil
	case bean2.GitOpsWithoutDeployment:
		if triggerEvent.PerformDeploymentOnCluster {
			return false, errors2.New("For deployment type GitOpsWithoutDeployment, PerformDeploymentOnCluster flag expected value = false, got value = true")
		}
	case bean2.ManifestDownload:
		if triggerEvent.PerformChartPush {
			return false, errors3.New("For deployment type ManifestDownload,  PerformChartPush flag expected value = false, got true")
		}
		if triggerEvent.PerformDeploymentOnCluster {
			return false, errors3.New("For deployment type ManifestDownload,  PerformDeploymentOnCluster flag expected value = false, got true")
		}
	}
	return true, nil

}

func (impl *WorkflowDagExecutorImpl) BuildManifestForTrigger(overrideRequest *bean.ValuesOverrideRequest, triggeredAt time.Time, ctx context.Context) (valuesOverrideResponse *app.ValuesOverrideResponse, builtChartPath string, err error) {

	valuesOverrideResponse = &app.ValuesOverrideResponse{}
	valuesOverrideResponse, err = impl.GetValuesOverrideForTrigger(overrideRequest, triggeredAt, ctx)
	if err != nil {
		impl.logger.Errorw("error in fetching values for trigger", "err", err)
		return valuesOverrideResponse, "", err
	}
	builtChartPath, err = impl.appService.BuildChartAndGetPath(overrideRequest.AppName, valuesOverrideResponse.EnvOverride, ctx)
	if err != nil {
		impl.logger.Errorw("error in parsing reference chart", "err", err)
		return valuesOverrideResponse, "", err
	}
	return valuesOverrideResponse, builtChartPath, err
}

func (impl *WorkflowDagExecutorImpl) CreateHistoriesForDeploymentTrigger(pipeline *pipelineConfig.Pipeline, strategy *chartConfig.PipelineStrategy, envOverride *chartConfig.EnvConfigOverride, deployedOn time.Time, deployedBy int32) error {
	//creating history for deployment template
	deploymentTemplateHistory, err := impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryForDeploymentTrigger(pipeline, envOverride, envOverride.Chart.ImageDescriptorTemplate, deployedOn, deployedBy)
	if err != nil {
		impl.logger.Errorw("error in creating deployment template history for deployment trigger", "err", err)
		return err
	}
	err = impl.configMapHistoryService.CreateCMCSHistoryForDeploymentTrigger(pipeline, deployedOn, deployedBy)
	if err != nil {
		impl.logger.Errorw("error in creating CM/CS history for deployment trigger", "err", err)
		return err
	}
	if strategy != nil {
		err = impl.pipelineStrategyHistoryService.CreateStrategyHistoryForDeploymentTrigger(strategy, deployedOn, deployedBy, pipeline.TriggerType)
		if err != nil {
			impl.logger.Errorw("error in creating strategy history for deployment trigger", "err", err)
			return err
		}
	}
	//VARIABLE_SNAPSHOT_SAVE
	if envOverride.VariableSnapshot != nil && len(envOverride.VariableSnapshot) > 0 {
		variableMapBytes, _ := json.Marshal(envOverride.VariableSnapshot)
		variableSnapshotHistory := &repository5.VariableSnapshotHistoryBean{
			VariableSnapshot: variableMapBytes,
			HistoryReference: repository5.HistoryReference{
				HistoryReferenceId:   deploymentTemplateHistory.Id,
				HistoryReferenceType: repository5.HistoryReferenceTypeDeploymentTemplate,
			},
		}
		err = impl.variableSnapshotHistoryService.SaveVariableHistoriesForTrigger([]*repository5.VariableSnapshotHistoryBean{variableSnapshotHistory}, deployedBy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) BuildManifestPushTemplate(overrideRequest *bean.ValuesOverrideRequest, valuesOverrideResponse *app.ValuesOverrideResponse, builtChartPath string, manifest *[]byte) (*bean4.ManifestPushTemplate, error) {

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

func (impl *WorkflowDagExecutorImpl) GetManifestPushService(triggerEvent bean.TriggerEvent) app.ManifestPushService {
	var manifestPushService app.ManifestPushService
	if triggerEvent.ManifestStorageType == bean2.ManifestStorageGit {
		manifestPushService = impl.gitOpsManifestPushService
	}
	return manifestPushService
}

func (impl *WorkflowDagExecutorImpl) DeployApp(overrideRequest *bean.ValuesOverrideRequest, valuesOverrideResponse *app.ValuesOverrideResponse, triggeredAt time.Time, ctx context.Context) error {

	if util.IsAcdApp(overrideRequest.DeploymentAppType) {
		_, span := otel.Tracer("orchestrator").Start(ctx, "DeployArgocdApp")
		err := impl.DeployArgocdApp(overrideRequest, valuesOverrideResponse, ctx)
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

func (impl *WorkflowDagExecutorImpl) WriteCDTriggerEvent(overrideRequest *bean.ValuesOverrideRequest, artifact *repository.CiArtifact, releaseId, pipelineOverrideId int) {

	event := impl.eventFactory.Build(util2.Trigger, &overrideRequest.PipelineId, overrideRequest.AppId, &overrideRequest.EnvId, util2.CD)
	impl.logger.Debugw("event WriteCDTriggerEvent", "event", event)
	event = impl.eventFactory.BuildExtraCDData(event, nil, pipelineOverrideId, bean.CD_WORKFLOW_TYPE_DEPLOY)
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

func (impl *WorkflowDagExecutorImpl) MarkImageScanDeployed(appId int, envId int, imageDigest string, clusterId int, isScanEnabled bool) error {
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

func (impl *WorkflowDagExecutorImpl) GetValuesOverrideForTrigger(overrideRequest *bean.ValuesOverrideRequest, triggeredAt time.Time, ctx context.Context) (*app.ValuesOverrideResponse, error) {
	if overrideRequest.DeploymentType == models.DEPLOYMENTTYPE_UNKNOWN {
		overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_DEPLOY
	}
	if len(overrideRequest.DeploymentWithConfig) == 0 {
		overrideRequest.DeploymentWithConfig = bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED
	}
	valuesOverrideResponse := &app.ValuesOverrideResponse{}

	pipeline, err := impl.pipelineRepository.FindById(overrideRequest.PipelineId)
	valuesOverrideResponse.Pipeline = pipeline
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline by pipeline id", "err", err, "pipeline-id-", overrideRequest.PipelineId)
		return valuesOverrideResponse, err
	}

	_, span := otel.Tracer("orchestrator").Start(ctx, "ciArtifactRepository.Get")
	artifact, err := impl.ciArtifactRepository.Get(overrideRequest.CiArtifactId)
	valuesOverrideResponse.Artifact = artifact
	span.End()
	if err != nil {
		return valuesOverrideResponse, err
	}
	overrideRequest.Image = artifact.Image

	strategy, err := impl.GetDeploymentStrategyByTriggerType(overrideRequest, ctx)
	valuesOverrideResponse.PipelineStrategy = strategy
	if err != nil {
		impl.logger.Errorw("error in getting strategy by trigger type", "err", err)
		return valuesOverrideResponse, err
	}

	envOverride, err := impl.GetEnvOverrideByTriggerType(overrideRequest, triggeredAt, ctx)
	valuesOverrideResponse.EnvOverride = envOverride
	if err != nil {
		impl.logger.Errorw("error in getting env override by trigger type", "err", err)
		return valuesOverrideResponse, err
	}
	appMetrics, err := impl.GetAppMetricsByTriggerType(overrideRequest, ctx)
	valuesOverrideResponse.AppMetrics = appMetrics
	if err != nil {
		impl.logger.Errorw("error in getting app metrics by trigger type", "err", err)
		return valuesOverrideResponse, err
	}

	_, span = otel.Tracer("orchestrator").Start(ctx, "getDbMigrationOverride")
	//FIXME: how to determine rollback
	//we can't depend on ciArtifact ID because CI pipeline can be manually triggered in any order regardless of sourcecode status
	dbMigrationOverride, err := impl.getDbMigrationOverride(overrideRequest, artifact, false)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching db migration config", "req", overrideRequest, "err", err)
		return valuesOverrideResponse, err
	}
	chartVersion := envOverride.Chart.ChartVersion
	_, span = otel.Tracer("orchestrator").Start(ctx, "getConfigMapAndSecretJsonV2")
	configMapJson, err := impl.getConfigMapAndSecretJsonV2(overrideRequest.AppId, envOverride.TargetEnvironment, overrideRequest.PipelineId, chartVersion, overrideRequest.DeploymentWithConfig, overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching config map n secret ", "err", err)
		configMapJson = nil
	}
	_, span = otel.Tracer("orchestrator").Start(ctx, "appCrudOperationService.GetLabelsByAppIdForDeployment")
	appLabelJsonByte, err := impl.appCrudOperationService.GetLabelsByAppIdForDeployment(overrideRequest.AppId)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching app labels for gitOps commit", "err", err)
		appLabelJsonByte = nil
	}
	_, span = otel.Tracer("orchestrator").Start(ctx, "mergeAndSave")
	pipelineOverride, err := impl.savePipelineOverride(overrideRequest, envOverride.Id, triggeredAt)
	valuesOverrideResponse.PipelineOverride = pipelineOverride
	if err != nil {
		return valuesOverrideResponse, err
	}
	//TODO: check status and apply lock
	releaseOverrideJson, err := impl.getReleaseOverride(envOverride, overrideRequest, artifact, pipelineOverride, strategy, &appMetrics)
	valuesOverrideResponse.ReleaseOverrideJSON = releaseOverrideJson
	if err != nil {
		return valuesOverrideResponse, err
	}
	mergedValues, err := impl.mergeOverrideValues(envOverride, dbMigrationOverride, releaseOverrideJson, configMapJson, appLabelJsonByte, strategy)

	appName := fmt.Sprintf("%s-%s", overrideRequest.AppName, envOverride.Environment.Name)
	mergedValues = impl.autoscalingCheckBeforeTrigger(ctx, appName, envOverride.Namespace, mergedValues, overrideRequest)

	_, span = otel.Tracer("orchestrator").Start(ctx, "dockerRegistryIpsConfigService.HandleImagePullSecretOnApplicationDeployment")
	// handle image pull secret if access given
	mergedValues, err = impl.dockerRegistryIpsConfigService.HandleImagePullSecretOnApplicationDeployment(envOverride.Environment, pipeline.CiPipelineId, mergedValues)
	valuesOverrideResponse.MergedValues = string(mergedValues)
	span.End()
	if err != nil {
		return valuesOverrideResponse, err
	}
	pipelineOverride.PipelineMergedValues = string(mergedValues)
	err = impl.pipelineOverrideRepository.Update(pipelineOverride)
	if err != nil {
		return valuesOverrideResponse, err
	}
	return valuesOverrideResponse, err
}

func (impl *WorkflowDagExecutorImpl) DeployArgocdApp(overrideRequest *bean.ValuesOverrideRequest, valuesOverrideResponse *app.ValuesOverrideResponse, ctx context.Context) error {

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
	updateAppInArgocd, err := impl.updateArgoPipeline(overrideRequest.AppId, valuesOverrideResponse.Pipeline.Name, valuesOverrideResponse.EnvOverride, ctx)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in updating argocd app ", "err", err)
		return err
	}
	if updateAppInArgocd {
		impl.logger.Debug("argo-cd successfully updated")
	} else {
		impl.logger.Debug("argo-cd failed to update, ignoring it")
	}
	return nil
}
func (impl *WorkflowDagExecutorImpl) createArgoApplicationIfRequired(appId int, envConfigOverride *chartConfig.EnvConfigOverride, pipeline *pipelineConfig.Pipeline, userId int32) (string, error) {
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
			ValuesFile:      impl.getValuesFileForEnv(envModel.Id),
			RepoPath:        chart.ChartLocation,
			RepoUrl:         chart.GitRepoUrl,
		}

		argoAppName, err := impl.argoK8sClient.CreateAcdApp(appRequest, envModel.Cluster)
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

func (impl *WorkflowDagExecutorImpl) createHelmAppForCdPipeline(overrideRequest *bean.ValuesOverrideRequest, valuesOverrideResponse *app.ValuesOverrideResponse, triggeredAt time.Time, ctx context.Context) (bool, error) {

	pipeline := valuesOverrideResponse.Pipeline
	envOverride := valuesOverrideResponse.EnvOverride
	mergeAndSave := valuesOverrideResponse.MergedValues

	chartMetaData := &chart.Metadata{
		Name:    pipeline.App.AppName,
		Version: envOverride.Chart.ChartVersion,
	}
	referenceTemplatePath := path.Join(string(impl.refChartDir), envOverride.Chart.ReferenceTemplate)

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
		clusterConfig := &client2.ClusterConfig{
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
		releaseIdentifier := &client2.ReleaseIdentifier{
			ReleaseName:      releaseName,
			ReleaseNamespace: envOverride.Namespace,
			ClusterConfig:    clusterConfig,
		}

		if pipeline.DeploymentAppCreated {
			req := &client2.UpgradeReleaseRequest{
				ReleaseIdentifier: releaseIdentifier,
				ValuesYaml:        mergeAndSave,
				HistoryMax:        impl.helmAppService.GetRevisionHistoryMaxValue(client2.SOURCE_DEVTRON_APP),
				ChartContent:      &client2.ChartContent{Content: referenceChartByte},
			}

			updateApplicationResponse, err := impl.helmAppClient.UpdateApplication(ctx, req)

			// For cases where helm release was not found but db flag for deployment app created was true
			if err != nil && strings.Contains(err.Error(), "release: not found") {

				// retry install
				_, err = impl.helmInstallReleaseWithCustomChart(ctx, releaseIdentifier, referenceChartByte, mergeAndSave)

				// if retry failed, return
				if err != nil {
					impl.logger.Errorw("release not found, failed to re-install helm application", "err", err)
					return false, err
				}
			} else if err != nil {
				impl.logger.Errorw("error in updating helm application for cd pipeline", "err", err)
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
		cdWf, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(ctx, overrideRequest.CdWorkflowId, bean.CD_WORKFLOW_TYPE_DEPLOY)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("err on fetching cd workflow", "err", err)
			return false, err
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
				impl.logger.Errorw("err on updating cd workflow for status update", "err", err)
				return false, err
			}
			cdWorkflowId = cdWf.Id
			runner := &pipelineConfig.CdWorkflowRunner{
				Id:           cdWf.Id,
				Name:         pipeline.Name,
				WorkflowType: bean.CD_WORKFLOW_TYPE_DEPLOY,
				ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
				Status:       pipelineConfig.WorkflowInProgress,
				TriggeredBy:  overrideRequest.UserId,
				StartedOn:    triggeredAt,
				CdWorkflowId: cdWorkflowId,
				AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
			}
			_, err = impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
			if err != nil {
				impl.logger.Errorw("err on updating cd workflow runner for status update", "err", err)
				return false, err
			}
		} else {
			cdWf.Status = pipelineConfig.WorkflowInProgress
			cdWf.FinishedOn = time.Now()
			cdWf.UpdatedBy = overrideRequest.UserId
			cdWf.UpdatedOn = time.Now()
			err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(&cdWf)
			if err != nil {
				impl.logger.Errorw("error on update cd workflow runner", "cdWf", cdWf, "err", err)
				return false, err
			}
		}
	}
	return true, nil
}

func (impl *WorkflowDagExecutorImpl) GetDeploymentStrategyByTriggerType(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context) (*chartConfig.PipelineStrategy, error) {

	strategy := &chartConfig.PipelineStrategy{}
	var err error
	if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {
		_, span := otel.Tracer("orchestrator").Start(ctx, "strategyHistoryRepository.GetHistoryByPipelineIdAndWfrId")
		strategyHistory, err := impl.strategyHistoryRepository.GetHistoryByPipelineIdAndWfrId(overrideRequest.PipelineId, overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in getting deployed strategy history by pipleinId and wfrId", "err", err, "pipelineId", overrideRequest.PipelineId, "wfrId", overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
			return nil, err
		}
		strategy.Strategy = strategyHistory.Strategy
		strategy.Config = strategyHistory.Config
		strategy.PipelineId = overrideRequest.PipelineId
	} else if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
		if overrideRequest.ForceTrigger {
			_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineConfigRepository.GetDefaultStrategyByPipelineId")
			strategy, err = impl.pipelineConfigRepository.GetDefaultStrategyByPipelineId(overrideRequest.PipelineId)
			span.End()
		} else {
			var deploymentTemplate chartRepoRepository.DeploymentStrategy
			if overrideRequest.DeploymentTemplate == "ROLLING" {
				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_ROLLING
			} else if overrideRequest.DeploymentTemplate == "BLUE-GREEN" {
				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_BLUE_GREEN
			} else if overrideRequest.DeploymentTemplate == "CANARY" {
				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_CANARY
			} else if overrideRequest.DeploymentTemplate == "RECREATE" {
				deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_RECREATE
			}

			if len(deploymentTemplate) > 0 {
				_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineConfigRepository.FindByStrategyAndPipelineId")
				strategy, err = impl.pipelineConfigRepository.FindByStrategyAndPipelineId(deploymentTemplate, overrideRequest.PipelineId)
				span.End()
			} else {
				_, span := otel.Tracer("orchestrator").Start(ctx, "pipelineConfigRepository.GetDefaultStrategyByPipelineId")
				strategy, err = impl.pipelineConfigRepository.GetDefaultStrategyByPipelineId(overrideRequest.PipelineId)
				span.End()
			}
		}
		if err != nil && errors2.IsNotFound(err) == false {
			impl.logger.Errorf("invalid state", "err", err, "req", strategy)
			return nil, err
		}
	}
	return strategy, nil
}

func (impl *WorkflowDagExecutorImpl) GetEnvOverrideByTriggerType(overrideRequest *bean.ValuesOverrideRequest, triggeredAt time.Time, ctx context.Context) (*chartConfig.EnvConfigOverride, error) {

	envOverride := &chartConfig.EnvConfigOverride{}

	var err error
	if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {
		_, span := otel.Tracer("orchestrator").Start(ctx, "deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId")
		deploymentTemplateHistory, err := impl.deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId(overrideRequest.PipelineId, overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
		//VARIABLE_SNAPSHOT_GET and resolve

		span.End()
		if err != nil {
			impl.logger.Errorw("error in getting deployed deployment template history by pipelineId and wfrId", "err", err, "pipelineId", &overrideRequest, "wfrId", overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
			return nil, err
		}
		templateName := deploymentTemplateHistory.TemplateName
		templateVersion := deploymentTemplateHistory.TemplateVersion
		if templateName == "Rollout Deployment" {
			templateName = ""
		}
		//getting chart_ref by id
		_, span = otel.Tracer("orchestrator").Start(ctx, "chartRefRepository.FindByVersionAndName")
		chartRef, err := impl.chartRefRepository.FindByVersionAndName(templateName, templateVersion)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in getting chartRef by version and name", "err", err, "version", templateVersion, "name", templateName)
			return nil, err
		}
		//assuming that if a chartVersion is deployed then it's envConfigOverride will be available
		_, span = otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.GetByAppIdEnvIdAndChartRefId")
		envOverride, err = impl.environmentConfigRepository.GetByAppIdEnvIdAndChartRefId(overrideRequest.AppId, overrideRequest.EnvId, chartRef.Id)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in getting envConfigOverride for pipeline for specific chartVersion", "err", err, "appId", overrideRequest.AppId, "envId", overrideRequest.EnvId, "chartRefId", chartRef.Id)
			return nil, err
		}

		_, span = otel.Tracer("orchestrator").Start(ctx, "envRepository.FindById")
		env, err := impl.envRepository.FindById(envOverride.TargetEnvironment)
		span.End()
		if err != nil {
			impl.logger.Errorw("unable to find env", "err", err)
			return nil, err
		}
		envOverride.Environment = env

		//updating historical data in envConfigOverride and appMetrics flag
		envOverride.IsOverride = true
		envOverride.EnvOverrideValues = deploymentTemplateHistory.Template

		resolvedTemplate, variableMap, err := impl.getResolvedTemplateWithSnapshot(deploymentTemplateHistory.Id, envOverride.EnvOverrideValues)
		envOverride.ResolvedEnvOverrideValues = resolvedTemplate
		envOverride.VariableSnapshot = variableMap
		if err != nil {
			return envOverride, err
		}
	} else if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
		_, span := otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.ActiveEnvConfigOverride")
		envOverride, err = impl.environmentConfigRepository.ActiveEnvConfigOverride(overrideRequest.AppId, overrideRequest.EnvId)

		var chart *chartRepoRepository.Chart
		span.End()
		if err != nil {
			impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
			return nil, err
		}
		if envOverride.Id == 0 {
			_, span = otel.Tracer("orchestrator").Start(ctx, "chartRepository.FindLatestChartForAppByAppId")
			chart, err = impl.chartRepository.FindLatestChartForAppByAppId(overrideRequest.AppId)
			span.End()
			if err != nil {
				impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
				return nil, err
			}
			_, span = otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.FindChartByAppIdAndEnvIdAndChartRefId")
			envOverride, err = impl.environmentConfigRepository.FindChartByAppIdAndEnvIdAndChartRefId(overrideRequest.AppId, overrideRequest.EnvId, chart.ChartRefId)
			span.End()
			if err != nil && !errors2.IsNotFound(err) {
				impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
				return nil, err
			}

			//creating new env override config
			if errors2.IsNotFound(err) || envOverride == nil {
				_, span = otel.Tracer("orchestrator").Start(ctx, "envRepository.FindById")
				environment, err := impl.envRepository.FindById(overrideRequest.EnvId)
				span.End()
				if err != nil && !util.IsErrNoRows(err) {
					return nil, err
				}
				envOverride = &chartConfig.EnvConfigOverride{
					Active:            true,
					ManualReviewed:    true,
					Status:            models.CHARTSTATUS_SUCCESS,
					TargetEnvironment: overrideRequest.EnvId,
					ChartId:           chart.Id,
					AuditLog:          sql.AuditLog{UpdatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId},
					Namespace:         environment.Namespace,
					IsOverride:        false,
					EnvOverrideValues: "{}",
					Latest:            false,
					IsBasicViewLocked: chart.IsBasicViewLocked,
					CurrentViewEditor: chart.CurrentViewEditor,
				}
				_, span = otel.Tracer("orchestrator").Start(ctx, "environmentConfigRepository.Save")
				err = impl.environmentConfigRepository.Save(envOverride)
				span.End()
				if err != nil {
					impl.logger.Errorw("error in creating envconfig", "data", envOverride, "error", err)
					return nil, err
				}
			}
			envOverride.Chart = chart
		} else if envOverride.Id > 0 && !envOverride.IsOverride {
			_, span = otel.Tracer("orchestrator").Start(ctx, "chartRepository.FindLatestChartForAppByAppId")
			chart, err = impl.chartRepository.FindLatestChartForAppByAppId(overrideRequest.AppId)
			span.End()
			if err != nil {
				impl.logger.Errorw("invalid state", "err", err, "req", overrideRequest)
				return nil, err
			}
			envOverride.Chart = chart
		}

		_, span = otel.Tracer("orchestrator").Start(ctx, "envRepository.FindById")
		env, err := impl.envRepository.FindById(envOverride.TargetEnvironment)
		span.End()
		if err != nil {
			impl.logger.Errorw("unable to find env", "err", err)
			return nil, err
		}
		envOverride.Environment = env

		//VARIABLE different cases for variable resolution
		scope := resourceQualifiers.Scope{
			AppId:     overrideRequest.AppId,
			EnvId:     overrideRequest.EnvId,
			ClusterId: overrideRequest.ClusterId,
			SystemMetadata: &resourceQualifiers.SystemMetadata{
				EnvironmentName: env.Name,
				ClusterName:     env.Cluster.ClusterName,
				Namespace:       env.Namespace,
				AppName:         overrideRequest.AppName,
				Image:           overrideRequest.Image,
				ImageTag:        util3.GetImageTagFromImage(overrideRequest.Image),
			},
		}

		if envOverride.IsOverride {

			resolvedTemplate, variableMap, err := impl.extractVariablesAndResolveTemplate(scope, envOverride.EnvOverrideValues, repository5.Entity{
				EntityType: repository5.EntityTypeDeploymentTemplateEnvLevel,
				EntityId:   envOverride.Id,
			})
			envOverride.ResolvedEnvOverrideValues = resolvedTemplate
			envOverride.VariableSnapshot = variableMap
			if err != nil {
				return envOverride, err
			}

		} else {
			resolvedTemplate, variableMap, err := impl.extractVariablesAndResolveTemplate(scope, chart.GlobalOverride, repository5.Entity{
				EntityType: repository5.EntityTypeDeploymentTemplateAppLevel,
				EntityId:   chart.Id,
			})
			envOverride.Chart.ResolvedGlobalOverride = resolvedTemplate
			envOverride.VariableSnapshot = variableMap
			if err != nil {
				return envOverride, err
			}

		}
	}

	return envOverride, nil
}

func (impl *WorkflowDagExecutorImpl) GetAppMetricsByTriggerType(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context) (bool, error) {

	var appMetrics bool
	if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {
		_, span := otel.Tracer("orchestrator").Start(ctx, "deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId")
		deploymentTemplateHistory, err := impl.deploymentTemplateHistoryRepository.GetHistoryByPipelineIdAndWfrId(overrideRequest.PipelineId, overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in getting deployed deployment template history by pipelineId and wfrId", "err", err, "pipelineId", &overrideRequest, "wfrId", overrideRequest.WfrIdForDeploymentWithSpecificTrigger)
			return appMetrics, err
		}
		appMetrics = deploymentTemplateHistory.IsAppMetricsEnabled

	} else if overrideRequest.DeploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
		_, span := otel.Tracer("orchestrator").Start(ctx, "appLevelMetricsRepository.FindByAppId")
		appLevelMetrics, err := impl.appLevelMetricsRepository.FindByAppId(overrideRequest.AppId)
		span.End()
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("err", err)
			return appMetrics, &util.ApiError{InternalMessage: "unable to fetch app level metrics flag"}
		}
		appMetrics = appLevelMetrics.AppMetrics

		_, span = otel.Tracer("orchestrator").Start(ctx, "envLevelMetricsRepository.FindByAppIdAndEnvId")
		envLevelMetrics, err := impl.envLevelMetricsRepository.FindByAppIdAndEnvId(overrideRequest.AppId, overrideRequest.EnvId)
		span.End()
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("err", err)
			return appMetrics, &util.ApiError{InternalMessage: "unable to fetch env level metrics flag"}
		}
		if envLevelMetrics.Id != 0 && envLevelMetrics.AppMetrics != nil {
			appMetrics = *envLevelMetrics.AppMetrics
		}
	}
	return appMetrics, nil
}

func (impl *WorkflowDagExecutorImpl) getDbMigrationOverride(overrideRequest *bean.ValuesOverrideRequest, artifact *repository.CiArtifact, isRollback bool) (overrideJson []byte, err error) {
	if isRollback {
		return nil, fmt.Errorf("rollback not supported ye")
	}
	notConfigured := false
	config, err := impl.dbMigrationConfigRepository.FindByPipelineId(overrideRequest.PipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching pipeline override config", "req", overrideRequest, "err", err)
		return nil, err
	} else if util.IsErrNoRows(err) {
		notConfigured = true
	}
	envVal := &EnvironmentOverride{}
	if notConfigured {
		impl.logger.Warnw("no active db migration found", "pipeline", overrideRequest.PipelineId)
		envVal.Enabled = false
	} else {
		materialInfos, err := artifact.ParseMaterialInfo()
		if err != nil {
			return nil, err
		}

		hash, ok := materialInfos[config.GitMaterial.Url]
		if !ok {
			impl.logger.Errorf("wrong url map ", "map", materialInfos, "url", config.GitMaterial.Url)
			return nil, fmt.Errorf("configured url not found in material %s", config.GitMaterial.Url)
		}

		envVal.Enabled = true
		if config.GitMaterial.GitProvider.AuthMode != repository.AUTH_MODE_USERNAME_PASSWORD &&
			config.GitMaterial.GitProvider.AuthMode != repository.AUTH_MODE_ACCESS_TOKEN &&
			config.GitMaterial.GitProvider.AuthMode != repository.AUTH_MODE_ANONYMOUS {
			return nil, fmt.Errorf("auth mode %s not supported for migration", config.GitMaterial.GitProvider.AuthMode)
		}
		envVal.appendEnvironmentVariable("GIT_REPO_URL", config.GitMaterial.Url)
		envVal.appendEnvironmentVariable("GIT_USER", config.GitMaterial.GitProvider.UserName)
		var password string
		if config.GitMaterial.GitProvider.AuthMode == repository.AUTH_MODE_USERNAME_PASSWORD {
			password = config.GitMaterial.GitProvider.Password
		} else {
			password = config.GitMaterial.GitProvider.AccessToken
		}
		envVal.appendEnvironmentVariable("GIT_AUTH_TOKEN", password)
		// parse git-tag not required
		//envVal.appendEnvironmentVariable("GIT_TAG", "")
		envVal.appendEnvironmentVariable("GIT_HASH", hash)
		envVal.appendEnvironmentVariable("SCRIPT_LOCATION", config.ScriptSource)
		envVal.appendEnvironmentVariable("DB_TYPE", string(config.DbConfig.Type))
		envVal.appendEnvironmentVariable("DB_USER_NAME", config.DbConfig.UserName)
		envVal.appendEnvironmentVariable("DB_PASSWORD", config.DbConfig.Password)
		envVal.appendEnvironmentVariable("DB_HOST", config.DbConfig.Host)
		envVal.appendEnvironmentVariable("DB_PORT", config.DbConfig.Port)
		envVal.appendEnvironmentVariable("DB_NAME", config.DbConfig.DbName)
		//Will be used for rollback don't delete it
		//envVal.appendEnvironmentVariable("MIGRATE_TO_VERSION", strconv.Itoa(overrideRequest.TargetDbVersion))
	}
	dbMigrationConfig := map[string]interface{}{"dbMigrationConfig": envVal}
	confByte, err := json.Marshal(dbMigrationConfig)
	if err != nil {
		return nil, err
	}
	return confByte, nil
}

func (impl *WorkflowDagExecutorImpl) getConfigMapAndSecretJsonV2(appId int, envId int, pipelineId int, chartVersion string, deploymentWithConfig bean.DeploymentConfigurationType, wfrIdForDeploymentWithSpecificTrigger int) ([]byte, error) {

	var configMapJson string
	var secretDataJson string
	var configMapJsonApp string
	var secretDataJsonApp string
	var configMapJsonEnv string
	var secretDataJsonEnv string
	var err error
	//var configMapJsonPipeline string
	//var secretDataJsonPipeline string

	merged := []byte("{}")
	if deploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
		configMapA, err := impl.configMapRepository.GetByAppIdAppLevel(appId)
		if err != nil && pg.ErrNoRows != err {
			return []byte("{}"), err
		}
		if configMapA != nil && configMapA.Id > 0 {
			configMapJsonApp = configMapA.ConfigMapData
			secretDataJsonApp = configMapA.SecretData
		}
		configMapE, err := impl.configMapRepository.GetByAppIdAndEnvIdEnvLevel(appId, envId)
		if err != nil && pg.ErrNoRows != err {
			return []byte("{}"), err
		}
		if configMapE != nil && configMapE.Id > 0 {
			configMapJsonEnv = configMapE.ConfigMapData
			secretDataJsonEnv = configMapE.SecretData
		}
	} else if deploymentWithConfig == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {
		//fetching history and setting envLevelConfig and not appLevelConfig because history already contains merged appLevel and envLevel configs
		configMapHistory, err := impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(pipelineId, wfrIdForDeploymentWithSpecificTrigger, repository3.CONFIGMAP_TYPE)
		if err != nil {
			impl.logger.Errorw("error in getting config map history config by pipelineId and wfrId ", "err", err, "pipelineId", pipelineId, "wfrid", wfrIdForDeploymentWithSpecificTrigger)
			return []byte("{}"), err
		}
		configMapJsonEnv = configMapHistory.Data
		secretHistory, err := impl.configMapHistoryRepository.GetHistoryByPipelineIdAndWfrId(pipelineId, wfrIdForDeploymentWithSpecificTrigger, repository3.SECRET_TYPE)
		if err != nil {
			impl.logger.Errorw("error in getting config map history config by pipelineId and wfrId ", "err", err, "pipelineId", pipelineId, "wfrid", wfrIdForDeploymentWithSpecificTrigger)
			return []byte("{}"), err
		}
		secretDataJsonEnv = secretHistory.Data
	}
	configMapJson, err = impl.mergeUtil.ConfigMapMerge(configMapJsonApp, configMapJsonEnv)
	if err != nil {
		return []byte("{}"), err
	}
	chartMajorVersion, chartMinorVersion, err := util4.ExtractChartVersion(chartVersion)
	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return []byte("{}"), err
	}
	secretDataJson, err = impl.mergeUtil.ConfigSecretMerge(secretDataJsonApp, secretDataJsonEnv, chartMajorVersion, chartMinorVersion, false)
	if err != nil {
		return []byte("{}"), err
	}
	configResponseR := bean.ConfigMapRootJson{}
	configResponse := bean.ConfigMapJson{}
	if configMapJson != "" {
		err = json.Unmarshal([]byte(configMapJson), &configResponse)
		if err != nil {
			return []byte("{}"), err
		}
	}
	configResponseR.ConfigMapJson = configResponse
	secretResponseR := bean.ConfigSecretRootJson{}
	secretResponse := bean.ConfigSecretJson{}
	if configMapJson != "" {
		err = json.Unmarshal([]byte(secretDataJson), &secretResponse)
		if err != nil {
			return []byte("{}"), err
		}
	}
	secretResponseR.ConfigSecretJson = secretResponse

	configMapByte, err := json.Marshal(configResponseR)
	if err != nil {
		return []byte("{}"), err
	}
	secretDataByte, err := json.Marshal(secretResponseR)
	if err != nil {
		return []byte("{}"), err
	}

	merged, err = impl.mergeUtil.JsonPatch(configMapByte, secretDataByte)
	if err != nil {
		return []byte("{}"), err
	}
	return merged, nil
}

func (impl *WorkflowDagExecutorImpl) savePipelineOverride(overrideRequest *bean.ValuesOverrideRequest, envOverrideId int, triggeredAt time.Time) (override *chartConfig.PipelineOverride, err error) {
	currentReleaseNo, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(overrideRequest.PipelineId)
	if err != nil {
		return nil, err
	}
	po := &chartConfig.PipelineOverride{
		EnvConfigOverrideId:    envOverrideId,
		Status:                 models.CHARTSTATUS_NEW,
		PipelineId:             overrideRequest.PipelineId,
		CiArtifactId:           overrideRequest.CiArtifactId,
		PipelineReleaseCounter: currentReleaseNo + 1,
		CdWorkflowId:           overrideRequest.CdWorkflowId,
		AuditLog:               sql.AuditLog{CreatedBy: overrideRequest.UserId, CreatedOn: triggeredAt, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
		DeploymentType:         overrideRequest.DeploymentType,
	}

	err = impl.pipelineOverrideRepository.Save(po)
	if err != nil {
		return nil, err
	}
	err = impl.checkAndFixDuplicateReleaseNo(po)
	if err != nil {
		impl.logger.Errorw("error in checking release no duplicacy", "pipeline", po, "err", err)
		return nil, err
	}
	return po, nil
}

func (impl *WorkflowDagExecutorImpl) getReleaseOverride(envOverride *chartConfig.EnvConfigOverride, overrideRequest *bean.ValuesOverrideRequest, artifact *repository.CiArtifact, pipelineOverride *chartConfig.PipelineOverride, strategy *chartConfig.PipelineStrategy, appMetrics *bool) (releaseOverride string, err error) {

	artifactImage := artifact.Image
	imageTag := strings.Split(artifactImage, ":")

	imageTagLen := len(imageTag)

	imageName := ""

	for i := 0; i < imageTagLen-1; i++ {
		if i != imageTagLen-2 {
			imageName = imageName + imageTag[i] + ":"
		} else {
			imageName = imageName + imageTag[i]
		}
	}

	appId := strconv.Itoa(overrideRequest.AppId)
	envId := strconv.Itoa(overrideRequest.EnvId)

	deploymentStrategy := ""
	if strategy != nil {
		deploymentStrategy = string(strategy.Strategy)
	}
	releaseAttribute := app.ReleaseAttributes{
		Name:           imageName,
		Tag:            imageTag[imageTagLen-1],
		PipelineName:   overrideRequest.PipelineName,
		ReleaseVersion: strconv.Itoa(pipelineOverride.PipelineReleaseCounter),
		DeploymentType: deploymentStrategy,
		App:            appId,
		Env:            envId,
		AppMetrics:     appMetrics,
	}
	override, err := util4.Tprintf(envOverride.Chart.ImageDescriptorTemplate, releaseAttribute)
	if err != nil {
		return "", &util.ApiError{InternalMessage: "unable to render ImageDescriptorTemplate"}
	}
	if overrideRequest.AdditionalOverride != nil {
		userOverride, err := overrideRequest.AdditionalOverride.MarshalJSON()
		if err != nil {
			return "", err
		}
		data, err := impl.mergeUtil.JsonPatch(userOverride, []byte(override))
		if err != nil {
			return "", err
		}
		override = string(data)
	}
	return override, nil
}

func (impl *WorkflowDagExecutorImpl) mergeAndSave(envOverride *chartConfig.EnvConfigOverride,
	overrideRequest *bean.ValuesOverrideRequest,
	dbMigrationOverride []byte,
	artifact *repository.CiArtifact,
	pipeline *pipelineConfig.Pipeline, configMapJson, appLabelJsonByte []byte, strategy *chartConfig.PipelineStrategy, ctx context.Context,
	triggeredAt time.Time, deployedBy int32, appMetrics *bool) (releaseId int, overrideId int, mergedValues string, err error) {

	//register release , obtain release id TODO: populate releaseId to template
	override, err := impl.savePipelineOverride(overrideRequest, envOverride.Id, triggeredAt)
	if err != nil {
		return 0, 0, "", err
	}
	//TODO: check status and apply lock
	overrideJson, err := impl.getReleaseOverride(envOverride, overrideRequest, artifact, override, strategy, appMetrics)
	if err != nil {
		return 0, 0, "", err
	}

	//merge three values on the fly
	//ordering is important here
	//global < environment < db< release
	var merged []byte
	if !envOverride.IsOverride {
		merged, err = impl.mergeUtil.JsonPatch([]byte("{}"), []byte(envOverride.Chart.GlobalOverride))
		if err != nil {
			return 0, 0, "", err
		}
	} else {
		merged, err = impl.mergeUtil.JsonPatch([]byte("{}"), []byte(envOverride.EnvOverrideValues))
		if err != nil {
			return 0, 0, "", err
		}
	}

	//pipeline override here comes from pipeline strategy table
	if strategy != nil && len(strategy.Config) > 0 {
		merged, err = impl.mergeUtil.JsonPatch(merged, []byte(strategy.Config))
		if err != nil {
			return 0, 0, "", err
		}
	}
	merged, err = impl.mergeUtil.JsonPatch(merged, dbMigrationOverride)
	if err != nil {
		return 0, 0, "", err
	}
	merged, err = impl.mergeUtil.JsonPatch(merged, []byte(overrideJson))
	if err != nil {
		return 0, 0, "", err
	}

	if configMapJson != nil {
		merged, err = impl.mergeUtil.JsonPatch(merged, configMapJson)
		if err != nil {
			return 0, 0, "", err
		}
	}

	if appLabelJsonByte != nil {
		merged, err = impl.mergeUtil.JsonPatch(merged, appLabelJsonByte)
		if err != nil {
			return 0, 0, "", err
		}
	}

	appName := fmt.Sprintf("%s-%s", pipeline.App.AppName, envOverride.Environment.Name)
	merged = impl.autoscalingCheckBeforeTrigger(ctx, appName, envOverride.Namespace, merged, overrideRequest)

	_, span := otel.Tracer("orchestrator").Start(ctx, "dockerRegistryIpsConfigService.HandleImagePullSecretOnApplicationDeployment")
	// handle image pull secret if access given
	merged, err = impl.dockerRegistryIpsConfigService.HandleImagePullSecretOnApplicationDeployment(envOverride.Environment, pipeline.CiPipelineId, merged)
	span.End()
	if err != nil {
		return 0, 0, "", err
	}

	commitHash := ""
	commitTime := time.Time{}
	if util.IsAcdApp(pipeline.DeploymentAppType) {
		chartRepoName := impl.chartTemplateService.GetGitOpsRepoNameFromUrl(envOverride.Chart.GitRepoUrl)
		_, span = otel.Tracer("orchestrator").Start(ctx, "chartTemplateService.GetUserEmailIdAndNameForGitOpsCommit")
		//getting username & emailId for commit author data
		userEmailId, userName := impl.chartTemplateService.GetUserEmailIdAndNameForGitOpsCommit(overrideRequest.UserId)
		span.End()
		chartGitAttr := &util.ChartConfig{
			FileName:       fmt.Sprintf("_%d-values.yaml", envOverride.TargetEnvironment),
			FileContent:    string(merged),
			ChartName:      envOverride.Chart.ChartName,
			ChartLocation:  envOverride.Chart.ChartLocation,
			ChartRepoName:  chartRepoName,
			ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", override.Id, envOverride.TargetEnvironment),
			UserName:       userName,
			UserEmailId:    userEmailId,
		}
		gitOpsConfigBitbucket, err := impl.gitOpsConfigRepository.GetGitOpsConfigByProvider(util.BITBUCKET_PROVIDER)
		if err != nil {
			if err == pg.ErrNoRows {
				gitOpsConfigBitbucket.BitBucketWorkspaceId = ""
			} else {
				return 0, 0, "", err
			}
		}
		gitOpsConfig := &bean.GitOpsConfigDto{BitBucketWorkspaceId: gitOpsConfigBitbucket.BitBucketWorkspaceId}
		_, span = otel.Tracer("orchestrator").Start(ctx, "gitFactory.Client.CommitValues")
		commitHash, commitTime, err = impl.gitFactory.Client.CommitValues(chartGitAttr, gitOpsConfig)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in git commit", "err", err)
			return 0, 0, "", err
		}
	}
	if commitTime.IsZero() {
		commitTime = time.Now()
	}
	pipelineOverride := &chartConfig.PipelineOverride{
		Id:                     override.Id,
		GitHash:                commitHash,
		CommitTime:             commitTime,
		EnvConfigOverrideId:    envOverride.Id,
		PipelineOverrideValues: overrideJson,
		PipelineId:             overrideRequest.PipelineId,
		CiArtifactId:           overrideRequest.CiArtifactId,
		PipelineMergedValues:   string(merged),
		AuditLog:               sql.AuditLog{UpdatedOn: triggeredAt, UpdatedBy: deployedBy},
	}
	_, span = otel.Tracer("orchestrator").Start(ctx, "pipelineOverrideRepository.Update")
	err = impl.pipelineOverrideRepository.Update(pipelineOverride)
	span.End()
	if err != nil {
		return 0, 0, "", err
	}
	mergedValues = string(merged)
	return override.PipelineReleaseCounter, override.Id, mergedValues, nil
}

func (impl *WorkflowDagExecutorImpl) mergeOverrideValues(envOverride *chartConfig.EnvConfigOverride,
	dbMigrationOverride []byte,
	releaseOverrideJson string,
	configMapJson []byte,
	appLabelJsonByte []byte,
	strategy *chartConfig.PipelineStrategy,
) (mergedValues []byte, err error) {

	//merge three values on the fly
	//ordering is important here
	//global < environment < db< release
	var merged []byte
	if !envOverride.IsOverride {
		merged, err = impl.mergeUtil.JsonPatch([]byte("{}"), []byte(envOverride.Chart.ResolvedGlobalOverride))
		if err != nil {
			return nil, err
		}
	} else {
		merged, err = impl.mergeUtil.JsonPatch([]byte("{}"), []byte(envOverride.ResolvedEnvOverrideValues))
		if err != nil {
			return nil, err
		}
	}
	if strategy != nil && len(strategy.Config) > 0 {
		merged, err = impl.mergeUtil.JsonPatch(merged, []byte(strategy.Config))
		if err != nil {
			return nil, err
		}
	}
	merged, err = impl.mergeUtil.JsonPatch(merged, dbMigrationOverride)
	if err != nil {
		return nil, err
	}
	merged, err = impl.mergeUtil.JsonPatch(merged, []byte(releaseOverrideJson))
	if err != nil {
		return nil, err
	}
	if configMapJson != nil {
		merged, err = impl.mergeUtil.JsonPatch(merged, configMapJson)
		if err != nil {
			return nil, err
		}
	}
	if appLabelJsonByte != nil {
		merged, err = impl.mergeUtil.JsonPatch(merged, appLabelJsonByte)
		if err != nil {
			return nil, err
		}
	}
	return merged, nil
}

func (impl *WorkflowDagExecutorImpl) autoscalingCheckBeforeTrigger(ctx context.Context, appName string, namespace string, merged []byte, overrideRequest *bean.ValuesOverrideRequest) []byte {
	//pipeline := overrideRequest.Pipeline
	var appId = overrideRequest.AppId
	pipelineId := overrideRequest.PipelineId
	var appDeploymentType = overrideRequest.DeploymentAppType
	var clusterId = overrideRequest.ClusterId
	deploymentType := overrideRequest.DeploymentType
	templateMap := make(map[string]interface{})
	err := json.Unmarshal(merged, &templateMap)
	if err != nil {
		return merged
	}

	hpaResourceRequest := impl.getAutoScalingReplicaCount(templateMap, appName)
	impl.logger.Debugw("autoscalingCheckBeforeTrigger", "hpaResourceRequest", hpaResourceRequest)
	if hpaResourceRequest.IsEnable {
		resourceManifest := make(map[string]interface{})
		if util.IsAcdApp(appDeploymentType) {
			query := &application.ApplicationResourceRequest{
				Name:         &appName,
				Version:      &hpaResourceRequest.Version,
				Group:        &hpaResourceRequest.Group,
				Kind:         &hpaResourceRequest.Kind,
				ResourceName: &hpaResourceRequest.ResourceName,
				Namespace:    &namespace,
			}
			recv, err := impl.acdClient.GetResource(ctx, query)
			impl.logger.Debugw("resource manifest get replica count", "response", recv)
			if err != nil {
				impl.logger.Errorw("ACD Get Resource API Failed", "err", err)
				middleware.AcdGetResourceCounter.WithLabelValues(strconv.Itoa(appId), namespace, appName).Inc()
				return merged
			}
			if recv != nil && len(*recv.Manifest) > 0 {
				err := json.Unmarshal([]byte(*recv.Manifest), &resourceManifest)
				if err != nil {
					impl.logger.Errorw("unmarshal failed for hpa check", "err", err)
					return merged
				}
			}
		} else {
			version := "v2beta2"
			k8sResource, err := impl.k8sCommonService.GetResource(ctx, &k8s.ResourceRequestBean{ClusterId: clusterId,
				K8sRequest: &util5.K8sRequestBean{ResourceIdentifier: util5.ResourceIdentifier{Name: hpaResourceRequest.ResourceName,
					Namespace: namespace, GroupVersionKind: schema.GroupVersionKind{Group: hpaResourceRequest.Group, Kind: hpaResourceRequest.Kind, Version: version}}}})
			if err != nil {
				impl.logger.Errorw("error occurred while fetching resource for app", "resourceName", hpaResourceRequest.ResourceName, "err", err)
				return merged
			}
			resourceManifest = k8sResource.Manifest.Object
		}
		if len(resourceManifest) > 0 {
			statusMap := resourceManifest["status"].(map[string]interface{})
			currentReplicaVal := statusMap["currentReplicas"]
			currentReplicaCount, err := util4.ParseFloatNumber(currentReplicaVal)
			if err != nil {
				impl.logger.Errorw("error occurred while parsing replica count", "currentReplicas", currentReplicaVal, "err", err)
				return merged
			}

			reqReplicaCount := impl.fetchRequiredReplicaCount(currentReplicaCount, hpaResourceRequest.ReqMaxReplicas, hpaResourceRequest.ReqMinReplicas)
			templateMap["replicaCount"] = reqReplicaCount
			merged, err = json.Marshal(&templateMap)
			if err != nil {
				impl.logger.Errorw("marshaling failed for hpa check", "err", err)
				return merged
			}
		}
	} else {
		impl.logger.Errorw("autoscaling is not enabled", "pipelineId", pipelineId)
	}

	//check for custom chart support
	if autoscalingEnabledPath, ok := templateMap[bean2.CustomAutoScalingEnabledPathKey]; ok {
		if deploymentType == models.DEPLOYMENTTYPE_STOP {
			merged, err = impl.setScalingValues(templateMap, bean2.CustomAutoScalingEnabledPathKey, merged, false)
			if err != nil {
				return merged
			}
			merged, err = impl.setScalingValues(templateMap, bean2.CustomAutoscalingReplicaCountPathKey, merged, 0)
			if err != nil {
				return merged
			}
		} else {
			autoscalingEnabled := false
			autoscalingEnabledValue := gjson.Get(string(merged), autoscalingEnabledPath.(string)).Value()
			if val, ok := autoscalingEnabledValue.(bool); ok {
				autoscalingEnabled = val
			}
			if autoscalingEnabled {
				// extract replica count, min, max and check for required value
				replicaCount, err := impl.getReplicaCountFromCustomChart(templateMap, merged)
				if err != nil {
					return merged
				}
				merged, err = impl.setScalingValues(templateMap, bean2.CustomAutoscalingReplicaCountPathKey, merged, replicaCount)
				if err != nil {
					return merged
				}
			}
		}
	}

	return merged
}

func (impl *WorkflowDagExecutorImpl) updateArgoPipeline(appId int, pipelineName string, envOverride *chartConfig.EnvConfigOverride, ctx context.Context) (bool, error) {
	//repo has been registered while helm create
	if ctx == nil {
		impl.logger.Errorw("err in syncing ACD, ctx is NULL", "pipelineName", pipelineName)
		return false, nil
	}
	app, err := impl.appRepository.FindById(appId)
	if err != nil {
		impl.logger.Errorw("no app found ", "err", err)
		return false, err
	}
	envModel, err := impl.envRepository.FindById(envOverride.TargetEnvironment)
	if err != nil {
		return false, err
	}
	argoAppName := fmt.Sprintf("%s-%s", app.AppName, envModel.Name)
	impl.logger.Infow("received payload, updateArgoPipeline", "appId", appId, "pipelineName", pipelineName, "envId", envOverride.TargetEnvironment, "argoAppName", argoAppName, "context", ctx)
	application3, err := impl.acdClient.Get(ctx, &application.ApplicationQuery{Name: &argoAppName})
	if err != nil {
		impl.logger.Errorw("no argo app exists", "app", argoAppName, "pipeline", pipelineName)
		return false, err
	}
	//if status, ok:=status.FromError(err);ok{
	appStatus, _ := status2.FromError(err)

	if appStatus.Code() == codes.OK {
		impl.logger.Debugw("argo app exists", "app", argoAppName, "pipeline", pipelineName)
		if application3.Spec.Source.Path != envOverride.Chart.ChartLocation || application3.Spec.Source.TargetRevision != "master" {
			patchReq := v1alpha1.Application{Spec: v1alpha1.ApplicationSpec{Source: v1alpha1.ApplicationSource{Path: envOverride.Chart.ChartLocation, RepoURL: envOverride.Chart.GitRepoUrl, TargetRevision: "master"}}}
			reqbyte, err := json.Marshal(patchReq)
			if err != nil {
				impl.logger.Errorw("error in creating patch", "err", err)
			}
			reqString := string(reqbyte)
			patchType := "merge"
			_, err = impl.acdClient.Patch(ctx, &application.ApplicationPatchRequest{Patch: &reqString, Name: &argoAppName, PatchType: &patchType})
			if err != nil {
				impl.logger.Errorw("error in creating argo pipeline ", "name", pipelineName, "patch", string(reqbyte), "err", err)
				return false, err
			}
			impl.logger.Debugw("pipeline update req ", "res", patchReq)
		} else {
			impl.logger.Debug("pipeline no need to update ")
		}
		// Doing normal refresh to avoid the sync delay in argo-cd.
		err2 := impl.argoClientWrapperService.GetArgoAppWithNormalRefresh(ctx, argoAppName)
		if err2 != nil {
			impl.logger.Errorw("error in getting argo application with normal refresh", "argoAppName", argoAppName, "pipelineName", pipelineName)
		}
		return true, nil
	} else if appStatus.Code() == codes.NotFound {
		impl.logger.Errorw("argo app not found", "app", argoAppName, "pipeline", pipelineName)
		return false, nil
	} else {
		impl.logger.Errorw("err in checking application on gocd", "err", err, "pipeline", pipelineName)
		return false, err
	}
}

func (impl *WorkflowDagExecutorImpl) getValuesFileForEnv(environmentId int) string {
	return fmt.Sprintf("_%d-values.yaml", environmentId) //-{envId}-values.yaml
}

func (impl *WorkflowDagExecutorImpl) updatePipeline(pipeline *pipelineConfig.Pipeline, userId int32) (bool, error) {
	err := impl.pipelineRepository.SetDeploymentAppCreatedInPipeline(true, pipeline.Id, userId)
	if err != nil {
		impl.logger.Errorw("error on updating cd pipeline for setting deployment app created", "err", err)
		return false, err
	}
	return true, nil
}

// helmInstallReleaseWithCustomChart performs helm install with custom chart
func (impl *WorkflowDagExecutorImpl) helmInstallReleaseWithCustomChart(ctx context.Context, releaseIdentifier *client2.ReleaseIdentifier, referenceChartByte []byte, valuesYaml string) (*client2.HelmInstallCustomResponse, error) {

	helmInstallRequest := client2.HelmInstallCustomRequest{
		ValuesYaml:        valuesYaml,
		ChartContent:      &client2.ChartContent{Content: referenceChartByte},
		ReleaseIdentifier: releaseIdentifier,
	}

	// Request exec
	return impl.helmAppClient.InstallReleaseWithCustomChart(ctx, &helmInstallRequest)
}

func (impl *WorkflowDagExecutorImpl) getResolvedTemplateWithSnapshot(deploymentTemplateHistoryId int, template string) (string, map[string]string, error) {

	variableSnapshotMap := make(map[string]string)
	reference := repository5.HistoryReference{
		HistoryReferenceId:   deploymentTemplateHistoryId,
		HistoryReferenceType: repository5.HistoryReferenceTypeDeploymentTemplate,
	}
	variableSnapshot, err := impl.variableSnapshotHistoryService.GetVariableHistoryForReferences([]repository5.HistoryReference{reference})
	if err != nil {
		return template, variableSnapshotMap, err
	}

	if _, ok := variableSnapshot[reference]; !ok {
		return template, variableSnapshotMap, nil
	}

	err = json.Unmarshal(variableSnapshot[reference].VariableSnapshot, &variableSnapshotMap)
	if err != nil {
		return template, variableSnapshotMap, err
	}

	if len(variableSnapshotMap) == 0 {
		return template, variableSnapshotMap, nil
	}
	scopedVariableData := parsers.GetScopedVarData(variableSnapshotMap, make(map[string]bool), true)
	request := parsers.VariableParserRequest{Template: template, TemplateType: parsers.JsonVariableTemplate, Variables: scopedVariableData}
	parserResponse := impl.variableTemplateParser.ParseTemplate(request)
	err = parserResponse.Error
	if err != nil {
		return template, variableSnapshotMap, err
	}
	resolvedTemplate := parserResponse.ResolvedTemplate
	return resolvedTemplate, variableSnapshotMap, nil
}

func (impl *WorkflowDagExecutorImpl) extractVariablesAndResolveTemplate(scope resourceQualifiers.Scope, template string, entity repository5.Entity) (string, map[string]string, error) {

	variableMap := make(map[string]string)
	entityToVariables, err := impl.variableEntityMappingService.GetAllMappingsForEntities([]repository5.Entity{entity})
	if err != nil {
		return template, variableMap, err
	}

	if vars, ok := entityToVariables[entity]; !ok || len(vars) == 0 {
		return template, variableMap, nil
	}

	// pre-populating variable map with variable so that the variables which don't have any resolved data
	// is saved in snapshot
	for _, variable := range entityToVariables[entity] {
		variableMap[variable] = impl.scopedVariableService.GetFormattedVariableForName(variable)
	}

	scopedVariables, err := impl.scopedVariableService.GetScopedVariables(scope, entityToVariables[entity], true)
	if err != nil {
		return template, variableMap, err
	}

	for _, variable := range scopedVariables {
		variableMap[variable.VariableName] = variable.VariableValue.StringValue()
	}

	parserRequest := parsers.VariableParserRequest{Template: template, Variables: scopedVariables, TemplateType: parsers.JsonVariableTemplate}
	parserResponse := impl.variableTemplateParser.ParseTemplate(parserRequest)
	err = parserResponse.Error
	if err != nil {
		return template, variableMap, err
	}

	resolvedTemplate := parserResponse.ResolvedTemplate
	return resolvedTemplate, variableMap, nil
}

type EnvironmentOverride struct {
	Enabled   bool        `json:"enabled"`
	EnvValues []*KeyValue `json:"envValues"`
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (conf *EnvironmentOverride) appendEnvironmentVariable(key, value string) {
	item := &KeyValue{Key: key, Value: value}
	conf.EnvValues = append(conf.EnvValues, item)
}

func (impl *WorkflowDagExecutorImpl) checkAndFixDuplicateReleaseNo(override *chartConfig.PipelineOverride) error {

	uniqueVerified := false
	retryCount := 0

	for !uniqueVerified && retryCount < 5 {
		retryCount = retryCount + 1
		overrides, err := impl.pipelineOverrideRepository.GetByPipelineIdAndReleaseNo(override.PipelineId, override.PipelineReleaseCounter)
		if err != nil {
			return err
		}
		if overrides[0].Id == override.Id {
			uniqueVerified = true
		} else {
			//duplicate might be due to concurrency, lets fix it
			currentReleaseNo, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(override.PipelineId)
			if err != nil {
				return err
			}
			override.PipelineReleaseCounter = currentReleaseNo + 1
			err = impl.pipelineOverrideRepository.Save(override)
			if err != nil {
				return err
			}
		}
	}
	if !uniqueVerified {
		return fmt.Errorf("duplicate verification retry count exide max overrideId: %d ,count: %d", override.Id, retryCount)
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) getAutoScalingReplicaCount(templateMap map[string]interface{}, appName string) *util4.HpaResourceRequest {
	hasOverride := false
	if _, ok := templateMap[fullnameOverride]; ok {
		appNameOverride := templateMap[fullnameOverride].(string)
		if len(appNameOverride) > 0 {
			appName = appNameOverride
			hasOverride = true
		}
	}
	if !hasOverride {
		if _, ok := templateMap[nameOverride]; ok {
			nameOverride := templateMap[nameOverride].(string)
			if len(nameOverride) > 0 {
				appName = fmt.Sprintf("%s-%s", appName, nameOverride)
			}
		}
	}
	hpaResourceRequest := &util4.HpaResourceRequest{}
	hpaResourceRequest.Version = ""
	hpaResourceRequest.Group = autoscaling.ServiceName
	hpaResourceRequest.Kind = horizontalPodAutoscaler
	impl.logger.Infow("getAutoScalingReplicaCount", "hpaResourceRequest", hpaResourceRequest)
	if _, ok := templateMap[kedaAutoscaling]; ok {
		as := templateMap[kedaAutoscaling]
		asd := as.(map[string]interface{})
		if _, ok := asd[enabled]; ok {
			impl.logger.Infow("getAutoScalingReplicaCount", "hpaResourceRequest", hpaResourceRequest)
			enable := asd[enabled].(bool)
			if enable {
				hpaResourceRequest.IsEnable = enable
				hpaResourceRequest.ReqReplicaCount = templateMap[replicaCount].(float64)
				hpaResourceRequest.ReqMaxReplicas = asd["maxReplicaCount"].(float64)
				hpaResourceRequest.ReqMinReplicas = asd["minReplicaCount"].(float64)
				hpaResourceRequest.ResourceName = fmt.Sprintf("%s-%s-%s", "keda-hpa", appName, "keda")
				impl.logger.Infow("getAutoScalingReplicaCount", "hpaResourceRequest", hpaResourceRequest)
				return hpaResourceRequest
			}
		}
	}

	if _, ok := templateMap[autoscaling.ServiceName]; ok {
		as := templateMap[autoscaling.ServiceName]
		asd := as.(map[string]interface{})
		if _, ok := asd[enabled]; ok {
			enable := asd[enabled].(bool)
			if enable {
				hpaResourceRequest.IsEnable = asd[enabled].(bool)
				hpaResourceRequest.ReqReplicaCount = templateMap[replicaCount].(float64)
				hpaResourceRequest.ReqMaxReplicas = asd["MaxReplicas"].(float64)
				hpaResourceRequest.ReqMinReplicas = asd["MinReplicas"].(float64)
				hpaResourceRequest.ResourceName = fmt.Sprintf("%s-%s", appName, "hpa")
				return hpaResourceRequest
			}
		}
	}
	return hpaResourceRequest

}

func (impl *WorkflowDagExecutorImpl) fetchRequiredReplicaCount(currentReplicaCount float64, reqMaxReplicas float64, reqMinReplicas float64) float64 {
	var reqReplicaCount float64
	if currentReplicaCount <= reqMaxReplicas && currentReplicaCount >= reqMinReplicas {
		reqReplicaCount = currentReplicaCount
	} else if currentReplicaCount > reqMaxReplicas {
		reqReplicaCount = reqMaxReplicas
	} else if currentReplicaCount < reqMinReplicas {
		reqReplicaCount = reqMinReplicas
	}
	return reqReplicaCount
}

func (impl *WorkflowDagExecutorImpl) getReplicaCountFromCustomChart(templateMap map[string]interface{}, merged []byte) (float64, error) {
	autoscalingMinVal, err := impl.extractParamValue(templateMap, bean2.CustomAutoscalingMinPathKey, merged)
	if err != nil {
		return 0, err
	}
	autoscalingMaxVal, err := impl.extractParamValue(templateMap, bean2.CustomAutoscalingMaxPathKey, merged)
	if err != nil {
		return 0, err
	}
	autoscalingReplicaCountVal, err := impl.extractParamValue(templateMap, bean2.CustomAutoscalingReplicaCountPathKey, merged)
	if err != nil {
		return 0, err
	}
	return impl.fetchRequiredReplicaCount(autoscalingReplicaCountVal, autoscalingMaxVal, autoscalingMinVal), nil
}

func (impl *WorkflowDagExecutorImpl) setScalingValues(templateMap map[string]interface{}, customScalingKey string, merged []byte, value interface{}) ([]byte, error) {
	autoscalingJsonPath := templateMap[customScalingKey]
	autoscalingJsonPathKey := autoscalingJsonPath.(string)
	mergedRes, err := sjson.Set(string(merged), autoscalingJsonPathKey, value)
	if err != nil {
		impl.logger.Errorw("error occurred while setting autoscaling key", "JsonPathKey", autoscalingJsonPathKey, "err", err)
		return []byte{}, err
	}
	return []byte(mergedRes), nil
}

func (impl *WorkflowDagExecutorImpl) extractParamValue(inputMap map[string]interface{}, key string, merged []byte) (float64, error) {
	if _, ok := inputMap[key]; !ok {
		return 0, errors.New("empty-val-err")
	}
	floatNumber, err := util4.ParseFloatNumber(gjson.Get(string(merged), inputMap[key].(string)).Value())
	if err != nil {
		impl.logger.Errorw("error occurred while parsing float number", "key", key, "err", err)
	}
	return floatNumber, err
}
