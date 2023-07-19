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
	"fmt"
	"github.com/argoproj/gitops-engine/pkg/health"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	gitSensorClient "github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/k8s"
	util4 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	util5 "github.com/devtron-labs/devtron/util/k8s"
	"go.opentelemetry.io/otel"
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
	HandleDeploymentSuccessEvent(gitHash string, pipelineOverrideId int) error
	HandlePostStageSuccessEvent(cdWorkflowId int, cdPipelineId int, triggeredBy int32) error
	Subscribe() error
	TriggerPostStage(cdWf *pipelineConfig.CdWorkflow, cdPipeline *pipelineConfig.Pipeline, triggeredBy int32) error
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
	cdWorkflowService             CdWorkflowService
	ciPipelineRepository          pipelineConfig.CiPipelineRepository
	materialRepository            pipelineConfig.MaterialRepository
	cdConfig                      *CdConfig
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
}

const (
	CD_PIPELINE_ENV_NAME_KEY     = "CD_PIPELINE_ENV_NAME"
	CD_PIPELINE_CLUSTER_NAME_KEY = "CD_PIPELINE_CLUSTER_NAME"
	GIT_COMMIT_HASH_PREFIX       = "GIT_COMMIT_HASH"
	GIT_SOURCE_TYPE_PREFIX       = "GIT_SOURCE_TYPE"
	GIT_SOURCE_VALUE_PREFIX      = "GIT_SOURCE_VALUE"
	GIT_SOURCE_COUNT             = "GIT_SOURCE_COUNT"
	APP_LABEL_KEY_PREFIX         = "APP_LABEL_KEY"
	APP_LABEL_VALUE_PREFIX       = "APP_LABEL_VALUE"
	APP_LABEL_COUNT              = "APP_LABEL_COUNT"
	CHILD_CD_ENV_NAME_PREFIX     = "CHILD_CD_ENV_NAME"
	CHILD_CD_CLUSTER_NAME_PREFIX = "CHILD_CD_CLUSTER_NAME"
	CHILD_CD_COUNT               = "CHILD_CD_COUNT"
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
	CiProjectDetails []CiProjectDetails           `json:"ciProjectDetails"`
	WorkflowId       int                          `json:"workflowId"`
	WorkflowRunnerId int                          `json:"workflowRunnerId"`
	CdPipelineId     int                          `json:"cdPipelineId"`
	TriggeredBy      int32                        `json:"triggeredBy"`
	StageYaml        string                       `json:"stageYaml"`
	ArtifactLocation string                       `json:"artifactLocation"`
	PipelineName     string                       `json:"pipelineName"`
	CiArtifactDTO    pipelineConfig.CiArtifactDTO `json:"ciArtifactDTO"`
}

func NewWorkflowDagExecutorImpl(Logger *zap.SugaredLogger, pipelineRepository pipelineConfig.PipelineRepository, cdWorkflowRepository pipelineConfig.CdWorkflowRepository, pubsubClient *pubsub.PubSubClientServiceImpl, appService app.AppService, cdWorkflowService CdWorkflowService, cdConfig *CdConfig, ciArtifactRepository repository.CiArtifactRepository, ciPipelineRepository pipelineConfig.CiPipelineRepository, materialRepository pipelineConfig.MaterialRepository, pipelineOverrideRepository chartConfig.PipelineOverrideRepository, user user.UserService, groupRepository repository.DeploymentGroupRepository, envRepository repository2.EnvironmentRepository, enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil, tokenCache *util3.TokenCache, acdAuthConfig *util3.ACDAuthConfig, eventFactory client.EventFactory, eventClient client.EventClient, cvePolicyRepository security.CvePolicyRepository, scanResultRepository security.ImageScanResultRepository, appWorkflowRepository appWorkflow.AppWorkflowRepository, prePostCdScriptHistoryService history2.PrePostCdScriptHistoryService, argoUserService argo.ArgoUserService, cdPipelineStatusTimelineRepo pipelineConfig.PipelineStatusTimelineRepository, pipelineStatusTimelineService status.PipelineStatusTimelineService, CiTemplateRepository pipelineConfig.CiTemplateRepository, ciWorkflowRepository pipelineConfig.CiWorkflowRepository, appLabelRepository pipelineConfig.AppLabelRepository, gitSensorGrpcClient gitSensorClient.Client, k8sCommonService k8s.K8sCommonService) *WorkflowDagExecutorImpl {
	wde := &WorkflowDagExecutorImpl{logger: Logger,
		pipelineRepository:            pipelineRepository,
		cdWorkflowRepository:          cdWorkflowRepository,
		pubsubClient:                  pubsubClient,
		appService:                    appService,
		cdWorkflowService:             cdWorkflowService,
		ciPipelineRepository:          ciPipelineRepository,
		cdConfig:                      cdConfig,
		ciArtifactRepository:          ciArtifactRepository,
		materialRepository:            materialRepository,
		pipelineOverrideRepository:    pipelineOverrideRepository,
		user:                          user,
		enforcer:                      enforcer,
		enforcerUtil:                  enforcerUtil,
		groupRepository:               groupRepository,
		tokenCache:                    tokenCache,
		acdAuthConfig:                 acdAuthConfig,
		envRepository:                 envRepository,
		eventFactory:                  eventFactory,
		eventClient:                   eventClient,
		cvePolicyRepository:           cvePolicyRepository,
		scanResultRepository:          scanResultRepository,
		appWorkflowRepository:         appWorkflowRepository,
		prePostCdScriptHistoryService: prePostCdScriptHistoryService,
		argoUserService:               argoUserService,
		cdPipelineStatusTimelineRepo:  cdPipelineStatusTimelineRepo,
		pipelineStatusTimelineService: pipelineStatusTimelineService,
		CiTemplateRepository:          CiTemplateRepository,
		ciWorkflowRepository:          ciWorkflowRepository,
		appLabelRepository:            appLabelRepository,
		gitSensorGrpcClient:           gitSensorGrpcClient,
		k8sCommonService:              k8sCommonService,
	}
	err := wde.Subscribe()
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

func (impl *WorkflowDagExecutorImpl) triggerStage(cdWf *pipelineConfig.CdWorkflow, pipeline *pipelineConfig.Pipeline, artifact *repository.CiArtifact, applyAuth bool, triggeredBy int32) error {
	var err error
	if len(pipeline.PreStageConfig) > 0 {
		// pre stage exists
		if pipeline.PreTriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC {
			impl.logger.Debugw("trigger pre stage for pipeline", "artifactId", artifact.Id, "pipelineId", pipeline.Id)
			err = impl.TriggerPreStage(context.Background(), cdWf, artifact, pipeline, artifact.UpdatedBy, applyAuth) //TODO handle error here
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

func (impl *WorkflowDagExecutorImpl) triggerStageForBulk(cdWf *pipelineConfig.CdWorkflow, pipeline *pipelineConfig.Pipeline, artifact *repository.CiArtifact, applyAuth bool, async bool, triggeredBy int32) error {
	var err error
	if len(pipeline.PreStageConfig) > 0 {
		//pre stage exists
		impl.logger.Debugw("trigger pre stage for pipeline", "artifactId", artifact.Id, "pipelineId", pipeline.Id)
		err = impl.TriggerPreStage(context.Background(), cdWf, artifact, pipeline, artifact.UpdatedBy, applyAuth) //TODO handle error here
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

func (impl *WorkflowDagExecutorImpl) TriggerPreStage(ctx context.Context, cdWf *pipelineConfig.CdWorkflow, artifact *repository.CiArtifact, pipeline *pipelineConfig.Pipeline, triggeredBy int32, applyAuth bool) error {
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
	cdWorkflowExecutorType := impl.cdConfig.CdWorkflowExecutorType
	runner := &pipelineConfig.CdWorkflowRunner{
		Name:               pipeline.Name,
		WorkflowType:       bean.CD_WORKFLOW_TYPE_PRE,
		ExecutorType:       cdWorkflowExecutorType,
		Status:             pipelineConfig.WorkflowStarting, //starting
		TriggeredBy:        triggeredBy,
		StartedOn:          triggeredAt,
		Namespace:          impl.cdConfig.DefaultNamespace,
		BlobStorageEnabled: impl.cdConfig.BlobStorageEnabled,
		CdWorkflowId:       cdWf.Id,
		LogLocation:        fmt.Sprintf("%s/%s%s-%s/main.log", impl.cdConfig.DefaultBuildLogsKeyPrefix, strconv.Itoa(cdWf.Id), string(bean.CD_WORKFLOW_TYPE_PRE), pipeline.Name),
		AuditLog:           sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: 1, UpdatedOn: triggeredAt, UpdatedBy: 1},
	}
	var env *repository2.Environment
	var err error
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

	_, span = otel.Tracer("orchestrator").Start(ctx, "buildWFRequest")
	cdStageWorkflowRequest, err := impl.buildWFRequest(runner, cdWf, pipeline, triggeredBy)
	span.End()
	if err != nil {
		return err
	}
	cdStageWorkflowRequest.StageType = PRE
	_, span = otel.Tracer("orchestrator").Start(ctx, "cdWorkflowService.SubmitWorkflow")
	err = impl.cdWorkflowService.SubmitWorkflow(cdStageWorkflowRequest, pipeline, env)
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

func (impl *WorkflowDagExecutorImpl) TriggerPostStage(cdWf *pipelineConfig.CdWorkflow, pipeline *pipelineConfig.Pipeline, triggeredBy int32) error {
	//setting triggeredAt variable to have consistent data for various audit log places in db for deployment time
	triggeredAt := time.Now()

	runner := &pipelineConfig.CdWorkflowRunner{
		Name:               pipeline.Name,
		WorkflowType:       bean.CD_WORKFLOW_TYPE_POST,
		ExecutorType:       impl.cdConfig.CdWorkflowExecutorType,
		Status:             pipelineConfig.WorkflowStarting,
		TriggeredBy:        triggeredBy,
		StartedOn:          triggeredAt,
		Namespace:          impl.cdConfig.DefaultNamespace,
		BlobStorageEnabled: impl.cdConfig.BlobStorageEnabled,
		CdWorkflowId:       cdWf.Id,
		LogLocation:        fmt.Sprintf("%s/%s%s-%s/main.log", impl.cdConfig.DefaultBuildLogsKeyPrefix, strconv.Itoa(cdWf.Id), string(bean.CD_WORKFLOW_TYPE_POST), pipeline.Name),
		AuditLog:           sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: triggeredBy, UpdatedOn: triggeredAt, UpdatedBy: triggeredBy},
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
	cdStageWorkflowRequest, err := impl.buildWFRequest(runner, cdWf, pipeline, triggeredBy)
	if err != nil {
		impl.logger.Errorw("error in building wfRequest", "err", err, "runner", runner, "cdWf", cdWf, "pipeline", pipeline)
		return err
	}
	cdStageWorkflowRequest.StageType = POST
	err = impl.cdWorkflowService.SubmitWorkflow(cdStageWorkflowRequest, pipeline, env)
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
		cdArtifactLocationFormat = impl.cdConfig.CdArtifactLocationFormat
	}
	if cdWorkflowConfig.LogsBucket == "" {
		cdWorkflowConfig.LogsBucket = impl.cdConfig.DefaultBuildLogsBucket
	}
	ArtifactLocation := fmt.Sprintf("s3://%s/"+impl.cdConfig.DefaultArtifactKeyPrefix+"/"+cdArtifactLocationFormat, cdWorkflowConfig.LogsBucket, cdWf.Id, runner.Id)
	artifactFileName := fmt.Sprintf(impl.cdConfig.DefaultArtifactKeyPrefix+"/"+cdArtifactLocationFormat, cdWf.Id, runner.Id)
	return ArtifactLocation, cdWorkflowConfig.LogsBucket, artifactFileName
}

func (impl *WorkflowDagExecutorImpl) buildWFRequest(runner *pipelineConfig.CdWorkflowRunner, cdWf *pipelineConfig.CdWorkflow, cdPipeline *pipelineConfig.Pipeline, triggeredBy int32) (*CdWorkflowRequest, error) {
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

	var ciProjectDetails []CiProjectDetails
	var ciPipeline *pipelineConfig.CiPipeline
	if cdPipeline.CiPipelineId > 0 {
		ciPipeline, err = impl.ciPipelineRepository.FindById(cdPipeline.CiPipelineId)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("cannot find ciPipeline", "err", err)
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

			ciProjectDetail := CiProjectDetails{
				GitRepository:   ciMaterialCurrent.Material.GitConfiguration.URL,
				MaterialName:    gitMaterial.Name,
				CheckoutPath:    gitMaterial.CheckoutPath,
				FetchSubmodules: gitMaterial.FetchSubmodules,
				SourceType:      m.Type,
				SourceValue:     m.Value,
				Type:            string(m.Type),
				GitOptions: GitOptions{
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
	if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
		stageYaml = cdPipeline.PreStageConfig
	} else if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
		stageYaml = cdPipeline.PostStageConfig
		//getting deployment pipeline latest wfr by pipelineId
		pipelineId := cdPipeline.Id
		deployStageWfr, err = impl.cdWorkflowRepository.FindLastStatusByPipelineIdAndRunnerType(pipelineId, bean.CD_WORKFLOW_TYPE_DEPLOY)
		if err != nil {
			impl.logger.Errorw("error in getting latest status of deploy type wfr by pipelineId", "err", err, "pipelineId", pipelineId)
			return nil, err
		}
		deployStageTriggeredByUser, err = impl.user.GetById(deployStageWfr.TriggeredBy)
		if err != nil {
			impl.logger.Errorw("error in getting userDetails by id", "err", err, "userId", deployStageWfr.TriggeredBy)
			return nil, err
		}
		pipelineReleaseCounter, err = impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(pipelineId)
		if err != nil {
			impl.logger.Errorw("error occurred while fetching latest release counter for pipeline", "pipelineId", pipelineId, "err", err)
			return nil, err
		}

	} else {
		return nil, fmt.Errorf("unsupported workflow triggerd")
	}

	cdStageWorkflowRequest := &CdWorkflowRequest{
		EnvironmentId:         cdPipeline.EnvironmentId,
		AppId:                 cdPipeline.AppId,
		WorkflowId:            cdWf.Id,
		WorkflowRunnerId:      runner.Id,
		WorkflowNamePrefix:    strconv.Itoa(runner.Id) + "-" + runner.Name,
		WorkflowPrefixForLog:  strconv.Itoa(cdWf.Id) + string(runner.WorkflowType) + "-" + runner.Name,
		CdImage:               impl.cdConfig.DefaultImage,
		CdPipelineId:          cdWf.PipelineId,
		TriggeredBy:           triggeredBy,
		StageYaml:             stageYaml,
		CiProjectDetails:      ciProjectDetails,
		Namespace:             runner.Namespace,
		ActiveDeadlineSeconds: impl.cdConfig.DefaultTimeout,
		CiArtifactDTO: CiArtifactDTO{
			Id:           artifact.Id,
			PipelineId:   artifact.PipelineId,
			Image:        artifact.Image,
			ImageDigest:  artifact.ImageDigest,
			MaterialInfo: artifact.MaterialInfo,
			DataSource:   artifact.DataSource,
			WorkflowId:   artifact.WorkflowId,
		},
		OrchestratorHost:  impl.cdConfig.OrchestratorHost,
		OrchestratorToken: impl.cdConfig.OrchestratorToken,
		CloudProvider:     impl.cdConfig.CloudProvider,
		WorkflowExecutor:  workflowExecutor,
	}
	extraEnvVariables := make(map[string]string)
	env, err := impl.envRepository.FindById(cdPipeline.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in getting environment by id", "err", err)
		return nil, err
	}
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

	if ciWf != nil && ciWf.GitTriggers != nil {
		i := 1
		for ciPipelineMaterialId, gitTrigger := range ciWf.GitTriggers {
			extraEnvVariables[fmt.Sprintf("%s_%d", GIT_COMMIT_HASH_PREFIX, i)] = gitTrigger.Commit
			extraEnvVariables[fmt.Sprintf("%s_%d", GIT_SOURCE_TYPE_PREFIX, i)] = string(gitTrigger.CiConfigureSourceType)
			extraEnvVariables[fmt.Sprintf("%s_%d", GIT_SOURCE_VALUE_PREFIX, i)] = gitTrigger.CiConfigureSourceValue

			// CODE-BLOCK starts - store extra environment variables if webhook
			if gitTrigger.CiConfigureSourceType == pipelineConfig.SOURCE_TYPE_WEBHOOK {
				webhookDataId := gitTrigger.WebhookData.Id
				if webhookDataId > 0 {
					webhookDataRequest := &gitSensorClient.WebhookDataRequest{
						Id:                   webhookDataId,
						CiPipelineMaterialId: ciPipelineMaterialId,
					}
					webhookAndCiData, err := impl.gitSensorGrpcClient.GetWebhookData(context.Background(), webhookDataRequest)
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
		for i, childPipeline := range childPipelines {
			extraEnvVariables[fmt.Sprintf("%s_%d", CHILD_CD_ENV_NAME_PREFIX, i+1)] = childPipeline.Environment.Name
			extraEnvVariables[fmt.Sprintf("%s_%d", CHILD_CD_CLUSTER_NAME_PREFIX, i+1)] = childPipeline.Environment.Cluster.ClusterName
		}
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
		for i, appLabel := range appLabels {
			extraEnvVariables[fmt.Sprintf("%s_%d", APP_LABEL_KEY_PREFIX, i+1)] = appLabel.Key
			extraEnvVariables[fmt.Sprintf("%s_%d", APP_LABEL_VALUE_PREFIX, i+1)] = appLabel.Value
		}
		if len(appLabels) > 0 {
			extraEnvVariables[APP_LABEL_COUNT] = strconv.Itoa(len(appLabels))
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
		cdWorkflowConfig.CdCacheRegion = impl.cdConfig.DefaultCdLogsBucketRegion
	}
	cdStageWorkflowRequest.BlobStorageConfigured = runner.BlobStorageEnabled
	switch cdStageWorkflowRequest.CloudProvider {
	case BLOB_STORAGE_S3:
		//No AccessKey is used for uploading artifacts, instead IAM based auth is used
		cdStageWorkflowRequest.CdCacheRegion = cdWorkflowConfig.CdCacheRegion
		cdStageWorkflowRequest.CdCacheLocation = cdWorkflowConfig.CdCacheBucket
		cdStageWorkflowRequest.ArtifactLocation, cdStageWorkflowRequest.ArtifactBucket, cdStageWorkflowRequest.ArtifactFileName = impl.buildArtifactLocationForS3(cdWorkflowConfig, cdWf, runner)
		cdStageWorkflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			AccessKey:                  impl.cdConfig.BlobStorageS3AccessKey,
			Passkey:                    impl.cdConfig.BlobStorageS3SecretKey,
			EndpointUrl:                impl.cdConfig.BlobStorageS3Endpoint,
			IsInSecure:                 impl.cdConfig.BlobStorageS3EndpointInsecure,
			CiCacheBucketName:          cdWorkflowConfig.CdCacheBucket,
			CiCacheRegion:              cdWorkflowConfig.CdCacheRegion,
			CiCacheBucketVersioning:    impl.cdConfig.BlobStorageS3BucketVersioned,
			CiArtifactBucketName:       cdStageWorkflowRequest.ArtifactBucket,
			CiArtifactRegion:           cdWorkflowConfig.CdCacheRegion,
			CiArtifactBucketVersioning: impl.cdConfig.BlobStorageS3BucketVersioned,
			CiLogBucketName:            impl.cdConfig.DefaultBuildLogsBucket,
			CiLogRegion:                impl.cdConfig.DefaultCdLogsBucketRegion,
			CiLogBucketVersioning:      impl.cdConfig.BlobStorageS3BucketVersioned,
		}
	case BLOB_STORAGE_GCP:
		cdStageWorkflowRequest.GcpBlobConfig = &blob_storage.GcpBlobConfig{
			CredentialFileJsonData: impl.cdConfig.BlobStorageGcpCredentialJson,
			ArtifactBucketName:     impl.cdConfig.DefaultBuildLogsBucket,
			LogBucketName:          impl.cdConfig.DefaultBuildLogsBucket,
		}
		cdStageWorkflowRequest.ArtifactLocation = impl.buildDefaultArtifactLocation(cdWorkflowConfig, cdWf, runner)
		cdStageWorkflowRequest.ArtifactFileName = cdStageWorkflowRequest.ArtifactLocation
	case BLOB_STORAGE_AZURE:
		cdStageWorkflowRequest.AzureBlobConfig = &blob_storage.AzureBlobConfig{
			Enabled:               true,
			AccountName:           impl.cdConfig.AzureAccountName,
			BlobContainerCiCache:  impl.cdConfig.AzureBlobContainerCiCache,
			AccountKey:            impl.cdConfig.AzureAccountKey,
			BlobContainerCiLog:    impl.cdConfig.AzureBlobContainerCiLog,
			BlobContainerArtifact: impl.cdConfig.AzureBlobContainerCiLog,
		}
		cdStageWorkflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			EndpointUrl:     impl.cdConfig.AzureGatewayUrl,
			IsInSecure:      impl.cdConfig.AzureGatewayConnectionInsecure,
			CiLogBucketName: impl.cdConfig.AzureBlobContainerCiLog,
			CiLogRegion:     "",
			AccessKey:       impl.cdConfig.AzureAccountName,
		}
		cdStageWorkflowRequest.ArtifactLocation = impl.buildDefaultArtifactLocation(cdWorkflowConfig, cdWf, runner)
		cdStageWorkflowRequest.ArtifactFileName = cdStageWorkflowRequest.ArtifactLocation
	default:
		if impl.cdConfig.BlobStorageEnabled {
			return nil, fmt.Errorf("blob storage %s not supported", cdStageWorkflowRequest.CloudProvider)
		}
	}
	cdStageWorkflowRequest.DefaultAddressPoolBaseCidr = impl.cdConfig.DefaultAddressPoolBaseCidr
	cdStageWorkflowRequest.DefaultAddressPoolSize = impl.cdConfig.DefaultAddressPoolSize
	return cdStageWorkflowRequest, nil
}

func (impl *WorkflowDagExecutorImpl) buildDefaultArtifactLocation(cdWorkflowConfig *pipelineConfig.CdWorkflowConfig, savedWf *pipelineConfig.CdWorkflow, runner *pipelineConfig.CdWorkflowRunner) string {
	cdArtifactLocationFormat := cdWorkflowConfig.CdArtifactLocationFormat
	if cdArtifactLocationFormat == "" {
		cdArtifactLocationFormat = impl.cdConfig.CdArtifactLocationFormat
	}
	ArtifactLocation := fmt.Sprintf("%s/"+cdArtifactLocationFormat, impl.cdConfig.DefaultArtifactKeyPrefix, savedWf.Id, runner.Id)
	return ArtifactLocation
}

func (impl *WorkflowDagExecutorImpl) HandleDeploymentSuccessEvent(gitHash string, pipelineOverrideId int) error {
	var pipelineOverride *chartConfig.PipelineOverride
	var err error
	if len(gitHash) > 0 && pipelineOverrideId == 0 {
		pipelineOverride, err = impl.pipelineOverrideRepository.FindByPipelineTriggerGitHash(gitHash)
		if err != nil {
			impl.logger.Errorw("error in fetching pipeline trigger by hash", "gitHash", gitHash)
			return err
		}
	} else if len(gitHash) == 0 && pipelineOverrideId > 0 {
		pipelineOverride, err = impl.pipelineOverrideRepository.FindById(pipelineOverrideId)
		if err != nil {
			impl.logger.Errorw("error in fetching pipeline trigger by override id", "pipelineOverrideId", pipelineOverrideId)
			return err
		}
	} else {
		return fmt.Errorf("no release found")
	}
	cdWorkflow, err := impl.cdWorkflowRepository.FindById(pipelineOverride.CdWorkflowId)
	if err != nil {
		impl.logger.Errorw("error in fetching cd workflow by id", "pipelineOverride", pipelineOverride)
		return err
	}
	if len(pipelineOverride.Pipeline.PostStageConfig) > 0 {
		if pipelineOverride.Pipeline.PostTriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC &&
			pipelineOverride.DeploymentType != models.DEPLOYMENTTYPE_STOP &&
			pipelineOverride.DeploymentType != models.DEPLOYMENTTYPE_START {

			err = impl.TriggerPostStage(cdWorkflow, pipelineOverride.Pipeline, 1)
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
		Namespace:    impl.cdConfig.DefaultNamespace,
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
		util4.TriggerCDMetrics(cdMetrics, impl.cdConfig.ExposeCDMetrics)
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

	err = impl.appService.TriggerCD(artifact, cdWf.Id, savedWfr.Id, pipeline, triggeredAt)
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
		util4.TriggerCDMetrics(cdMetrics, impl.cdConfig.ExposeCDMetrics)
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
	impl.appService.SetPipelineFieldsInOverrideRequest(overrideRequest, cdPipeline)

	if overrideRequest.CdWorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
		_, span = otel.Tracer("orchestrator").Start(ctx, "ciArtifactRepository.Get")
		artifact, err := impl.ciArtifactRepository.Get(overrideRequest.CiArtifactId)
		span.End()
		if err != nil {
			impl.logger.Errorw("err", "err", err)
			return 0, err
		}
		_, span = otel.Tracer("orchestrator").Start(ctx, "TriggerPreStage")
		err = impl.TriggerPreStage(ctx, nil, artifact, cdPipeline, overrideRequest.UserId, false)
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
			Namespace:    impl.cdConfig.DefaultNamespace,
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
			return 0, err
		}
		if isVulnerable == true {
			// if image vulnerable, update timeline status and return
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
			util4.TriggerCDMetrics(cdMetrics, impl.cdConfig.ExposeCDMetrics)
			// creating cd pipeline status timeline for deployment failed
			timeline := impl.pipelineStatusTimelineService.GetTimelineDbObjectByTimelineStatusAndTimelineDescription(runner.Id, pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_FAILED, pipelineConfig.TIMELINE_DESCRIPTION_VULNERABLE_IMAGE, 1)

			_, span = otel.Tracer("orchestrator").Start(ctx, "cdPipelineStatusTimelineRepo.SaveTimelineForACDHelmApps")
			err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, false)
			span.End()
			if err != nil {
				impl.logger.Errorw("error in creating timeline status for deployment fail - cve policy violation", "err", err, "timeline", timeline)
			}
			return 0, fmt.Errorf("found vulnerability for image digest %s", artifact.ImageDigest)
		}
		_, span = otel.Tracer("orchestrator").Start(ctx, "appService.TriggerRelease")
		releaseId, _, err = impl.appService.TriggerRelease(overrideRequest, ctx, triggeredAt, overrideRequest.UserId)
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
				Namespace:    impl.cdConfig.DefaultNamespace,
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
		err = impl.TriggerPostStage(cdWf, cdPipeline, overrideRequest.UserId)
		span.End()
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
