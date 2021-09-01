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
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/user"
	util2 "github.com/devtron-labs/devtron/util/event"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"github.com/nats-io/stan.go"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

type WorkflowDagExecutor interface {
	HandleCiSuccessEvent(artifact *repository.CiArtifact, applyAuth bool, async bool, triggeredBy int32) error
	HandlePreStageSuccessEvent(cdStageCompleteEvent CdStageCompleteEvent) error
	HandleDeploymentSuccessEvent(gitHash string) error
	Subscribe() error
	TriggerPostStage(cdWf *pipelineConfig.CdWorkflow, cdPipeline *pipelineConfig.Pipeline, triggeredBy int32) error
	TriggerDeployment(cdWf *pipelineConfig.CdWorkflow, artifact *repository.CiArtifact, pipeline *pipelineConfig.Pipeline, applyAuth bool, async bool, triggeredBy int32) error
	ManualCdTrigger(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context) (int, error)
	TriggerBulkDeploymentAsync(requests []*BulkTriggerRequest, UserId int32) (interface{}, error)
	StopStartApp(stopRequest *StopAppRequest, ctx context.Context) (int, error)
	TriggerBulkHibernateAsync(request StopDeploymentGroupRequest, ctx context.Context) (interface{}, error)
}

type WorkflowDagExecutorImpl struct {
	logger                     *zap.SugaredLogger
	pipelineRepository         pipelineConfig.PipelineRepository
	cdWorkflowRepository       pipelineConfig.CdWorkflowRepository
	pubsubClient               *pubsub.PubSubClient
	appService                 app.AppService
	cdWorkflowService          CdWorkflowService
	ciPipelineRepository       pipelineConfig.CiPipelineRepository
	materialRepository         pipelineConfig.MaterialRepository
	cdConfig                   *CdConfig
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository
	ciArtifactRepository       repository.CiArtifactRepository
	user                       user.UserService
	enforcer                   rbac.Enforcer
	enforcerUtil               rbac.EnforcerUtil
	groupRepository            repository.DeploymentGroupRepository
	tokenCache                 *user.TokenCache
	acdAuthConfig              *user.ACDAuthConfig
	envRepository              cluster.EnvironmentRepository
	eventFactory               client.EventFactory
	eventClient                client.EventClient
	cvePolicyRepository        security.CvePolicyRepository
	scanResultRepository       security.ImageScanResultRepository
}

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

const CD_STAGE_COMPLETE_TOPIC = "CI-RUNNER.CD-STAGE-COMPLETE"
const CD_COMPLETE_GROUP = "CI-RUNNER.CD-COMPLETE_GROUP-1"
const CD_COMPLETE_DURABLE = "CI-RUNNER.CD-COMPLETE_DURABLE-1"
const BULK_DEPLOY_TOPIC = "ORCHESTRATOR.CD.BULK"
const BULK_HIBERNATE_TOPIC = "ORCHESTRATOR.CD.BULK-HIBERNATE"
const BULK_DEPLOY_GROUP = "ORCHESTRATOR.CD.BULK.GROUP-1"
const BULK_HIBERNATE_GROUP = "ORCHESTRATOR.CD.BULK-HIBERNATE.GROUP-1"
const BULK_DEPLOY_DURABLE = "ORCHESTRATOR.CD.BULK.DURABLE-1"
const BULK_HIBERNATE_DURABLE = "ORCHESTRATOR.CD.BULK-HIBERNATE.DURABLE-1"

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

func NewWorkflowDagExecutorImpl(Logger *zap.SugaredLogger, pipelineRepository pipelineConfig.PipelineRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pubsubClient *pubsub.PubSubClient,
	appService app.AppService,
	cdWorkflowService CdWorkflowService,
	cdConfig *CdConfig,
	ciArtifactRepository repository.CiArtifactRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	materialRepository pipelineConfig.MaterialRepository,
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
	user user.UserService,
	groupRepository repository.DeploymentGroupRepository,
	envRepository cluster.EnvironmentRepository,
	enforcer rbac.Enforcer, enforcerUtil rbac.EnforcerUtil, tokenCache *user.TokenCache,
	acdAuthConfig *user.ACDAuthConfig, eventFactory client.EventFactory,
	eventClient client.EventClient, cvePolicyRepository security.CvePolicyRepository,
	scanResultRepository security.ImageScanResultRepository) *WorkflowDagExecutorImpl {
	wde := &WorkflowDagExecutorImpl{logger: Logger,
		pipelineRepository:         pipelineRepository,
		cdWorkflowRepository:       cdWorkflowRepository,
		pubsubClient:               pubsubClient,
		appService:                 appService,
		cdWorkflowService:          cdWorkflowService,
		ciPipelineRepository:       ciPipelineRepository,
		cdConfig:                   cdConfig,
		ciArtifactRepository:       ciArtifactRepository,
		materialRepository:         materialRepository,
		pipelineOverrideRepository: pipelineOverrideRepository,
		user:                       user,
		enforcer:                   enforcer,
		enforcerUtil:               enforcerUtil,
		groupRepository:            groupRepository,
		tokenCache:                 tokenCache,
		acdAuthConfig:              acdAuthConfig,
		envRepository:              envRepository,
		eventFactory:               eventFactory,
		eventClient:                eventClient,
		cvePolicyRepository:        cvePolicyRepository,
		scanResultRepository:       scanResultRepository,
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
	_, err := impl.pubsubClient.Conn.QueueSubscribe(CD_STAGE_COMPLETE_TOPIC, CD_COMPLETE_GROUP, func(msg *stan.Msg) {
		impl.logger.Debug("cd stage event received")
		defer msg.Ack()
		cdStageCompleteEvent := CdStageCompleteEvent{}
		err := json.Unmarshal([]byte(string(msg.Data)), &cdStageCompleteEvent)
		if err != nil {
			impl.logger.Errorw("error on cd stage complete", "err", err, "msg", string(msg.Data))
			return
		}
		impl.logger.Debugw("cd stage event:", "workflowRunnerId", cdStageCompleteEvent.WorkflowRunnerId)
		wf, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(cdStageCompleteEvent.WorkflowRunnerId)
		if err != nil {
			impl.logger.Errorw("could not get wf runner", "err", err)
			return
		}
		if wf.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
			_ = impl.HandlePreStageSuccessEvent(cdStageCompleteEvent)
		} else if wf.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
			impl.logger.Debugw("received post success event for workflow runner ", "wfId", strconv.Itoa(wf.Id))
		}
	}, stan.DurableName(CD_COMPLETE_DURABLE), stan.StartWithLastReceived(), stan.AckWait(time.Duration(impl.pubsubClient.AckDuration)*time.Second), stan.SetManualAckMode(), stan.MaxInflight(1))
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
	pipelines, err := impl.pipelineRepository.FindByCiPipelineId(artifact.PipelineId)
	if err != nil {
		impl.logger.Errorw("error in fetching cd pipeline", "pipelineId", artifact.PipelineId, "err", err)
		return err
	}
	for _, pipeline := range pipelines {
		err = impl.triggerStage(nil, pipeline, artifact, applyAuth, async, triggeredBy)
		if err != nil {
			impl.logger.Debugw("err", "err", err)
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) triggerStage(cdWf *pipelineConfig.CdWorkflow, pipeline *pipelineConfig.Pipeline, artifact *repository.CiArtifact, applyAuth bool, async bool, triggeredBy int32) error {
	var err error
	if len(pipeline.PreStageConfig) > 0 {
		//pre stage exists
		if pipeline.PreTriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC {
			impl.logger.Debugw("trigger pre stage for pipeline", "artifactId", artifact.Id, "pipelineId", pipeline.Id)
			err = impl.TriggerPreStage(cdWf, artifact, pipeline, artifact.UpdatedBy, applyAuth) //TODO handle error here
			return err
		}
	} else if pipeline.TriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC {
		// trigger deployment
		impl.logger.Debugw("trigger cd for pipeline", "artifactId", artifact.Id, "pipelineId", pipeline.Id)
		err = impl.TriggerDeployment(cdWf, artifact, pipeline, applyAuth, async, triggeredBy)
		return err
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) triggerStageForBulk(cdWf *pipelineConfig.CdWorkflow, pipeline *pipelineConfig.Pipeline, artifact *repository.CiArtifact, applyAuth bool, async bool, triggeredBy int32) error {
	var err error
	if len(pipeline.PreStageConfig) > 0 {
		//pre stage exists
		impl.logger.Debugw("trigger pre stage for pipeline", "artifactId", artifact.Id, "pipelineId", pipeline.Id)
		err = impl.TriggerPreStage(cdWf, artifact, pipeline, artifact.UpdatedBy, applyAuth) //TODO handle error here
		return err
	} else {
		// trigger deployment
		impl.logger.Debugw("trigger cd for pipeline", "artifactId", artifact.Id, "pipelineId", pipeline.Id)
		err = impl.TriggerDeployment(cdWf, artifact, pipeline, applyAuth, async, triggeredBy)
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

			applyAuth := false
			if cdStageCompleteEvent.TriggeredBy != 1 {
				applyAuth = true
			}
			err = impl.TriggerDeployment(cdWorkflow, ciArtifact, pipeline, applyAuth, false, cdStageCompleteEvent.TriggeredBy)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) TriggerPreStage(cdWf *pipelineConfig.CdWorkflow, artifact *repository.CiArtifact, pipeline *pipelineConfig.Pipeline, triggeredBy int32, applyAuth bool) error {

	//in case of pre stage manual trigger auth is already applied
	if applyAuth {
		user, err := impl.user.GetById(artifact.UpdatedBy)
		if err != nil {
			impl.logger.Errorw("error in fetching user for auto pipeline", "UpdatedBy", artifact.UpdatedBy)
			return nil
		}
		token := user.EmailId
		object := impl.enforcerUtil.GetAppRBACNameByAppId(pipeline.AppId)
		impl.logger.Debugw("Triggered Request (App Permission Checking):", "token", token, "object", object)
		if ok := impl.enforcer.EnforceByEmail(strings.ToLower(token), rbac.ResourceApplications, rbac.ActionTrigger, object); !ok {
			impl.logger.Warnw("unauthorized for pipeline ", "pipelineId", strconv.Itoa(pipeline.Id))
			return fmt.Errorf("unauthorized for pipeline " + strconv.Itoa(pipeline.Id))
		}
	}

	if cdWf == nil {
		cdWf = &pipelineConfig.CdWorkflow{
			CiArtifactId: artifact.Id,
			PipelineId:   pipeline.Id,
			AuditLog:     models.AuditLog{CreatedOn: time.Now(), CreatedBy: 1, UpdatedOn: time.Now(), UpdatedBy: 1},
		}
		err := impl.cdWorkflowRepository.SaveWorkFlow(cdWf)
		if err != nil {
			return err
		}
	}
	runner := &pipelineConfig.CdWorkflowRunner{
		Name:         pipeline.Name,
		WorkflowType: bean.CD_WORKFLOW_TYPE_PRE,
		ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
		Status:       WorkflowStarting, //starting
		TriggeredBy:  triggeredBy,
		StartedOn:    time.Now(),
		Namespace:    impl.cdConfig.DefaultNamespace,
		CdWorkflowId: cdWf.Id,
	}
	var env *cluster.Environment
	var err error
	if pipeline.RunPreStageInEnv {
		env, err = impl.envRepository.FindById(pipeline.EnvironmentId)
		if err != nil {
			impl.logger.Errorw(" unable to find env ", "err", err)
			return err
		}
		impl.logger.Debugw("env", "env", env)
		runner.Namespace = env.Namespace
	}
	err = impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
	if err != nil {
		return err
	}

	cdStageWorkflowRequest, err := impl.buildWFRequest(runner, cdWf, pipeline, triggeredBy)
	if err != nil {
		return err
	}
	cdStageWorkflowRequest.StageType = PRE
	_, err = impl.cdWorkflowService.SubmitWorkflow(cdStageWorkflowRequest, pipeline, env)

	wfr, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(cdWf.Id, bean.CD_WORKFLOW_TYPE_PRE)
	if err != nil {
		return err
	}

	event := impl.eventFactory.Build(util2.Trigger, &pipeline.Id, pipeline.AppId, &pipeline.EnvironmentId, util2.CD)
	impl.logger.Debugw("event PreStageTrigger", "event", event)
	event = impl.eventFactory.BuildExtraCDData(event, &wfr, 0, bean.CD_WORKFLOW_TYPE_PRE)
	_, evtErr := impl.eventClient.WriteEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("CD trigger event not sent", "error", evtErr)
	}
	return err

}

func convert(ts string) (*time.Time, error) {
	//layout := "2006-01-02T15:04:05Z"
	t, err := time.Parse(bean2.LayoutRFC3339, ts)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (impl *WorkflowDagExecutorImpl) TriggerPostStage(cdWf *pipelineConfig.CdWorkflow, cpipeline *pipelineConfig.Pipeline, triggeredBy int32) error {
	runner := &pipelineConfig.CdWorkflowRunner{
		Name:         cpipeline.Name,
		WorkflowType: bean.CD_WORKFLOW_TYPE_POST,
		ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
		Status:       WorkflowStarting,
		TriggeredBy:  triggeredBy,
		StartedOn:    time.Now(),
		Namespace:    impl.cdConfig.DefaultNamespace,
		CdWorkflowId: cdWf.Id,
	}
	var env *cluster.Environment
	var err error
	if cpipeline.RunPostStageInEnv {
		env, err = impl.envRepository.FindById(cpipeline.EnvironmentId)
		if err != nil {
			impl.logger.Errorw(" unable to find env ", "err", err)
			return err
		}
		runner.Namespace = env.Namespace
	}

	err = impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
	if err != nil {
		return err
	}
	cdStageWorkflowRequest, err := impl.buildWFRequest(runner, cdWf, cpipeline, triggeredBy)
	if err != nil {
		return err
	}
	cdStageWorkflowRequest.StageType = POST
	_, err = impl.cdWorkflowService.SubmitWorkflow(cdStageWorkflowRequest, cpipeline, env)
	if err != nil {
		return err
	}

	wfr, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(cdWf.Id, bean.CD_WORKFLOW_TYPE_POST)
	if err != nil {
		return err
	}

	event := impl.eventFactory.Build(util2.Trigger, &cpipeline.Id, cpipeline.AppId, &cpipeline.EnvironmentId, util2.CD)
	impl.logger.Debugw("event Cd Post Trigger", "event", event)
	event = impl.eventFactory.BuildExtraCDData(event, &wfr, 0, bean.CD_WORKFLOW_TYPE_POST)
	_, evtErr := impl.eventClient.WriteEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("CD trigger event not sent", "error", evtErr)
	}
	return err
}
func (impl *WorkflowDagExecutorImpl) buildArtifactLocation(cdWorkflowConfig *pipelineConfig.CdWorkflowConfig, cdWf *pipelineConfig.CdWorkflow, runner *pipelineConfig.CdWorkflowRunner) string {
	cdArtifactLocationFormat := cdWorkflowConfig.CdArtifactLocationFormat
	if cdArtifactLocationFormat == "" {
		cdArtifactLocationFormat = impl.cdConfig.CdArtifactLocationFormat
	}
	if cdWorkflowConfig.LogsBucket == "" {
		cdWorkflowConfig.LogsBucket = impl.cdConfig.DefaultBuildLogsBucket
	}
	ArtifactLocation := fmt.Sprintf("s3://%s/"+impl.cdConfig.DefaultArtifactKeyPrefix+"/"+cdArtifactLocationFormat, cdWorkflowConfig.LogsBucket, cdWf.Id, runner.Id)
	return ArtifactLocation
}

func (impl *WorkflowDagExecutorImpl) buildWFRequest(runner *pipelineConfig.CdWorkflowRunner, cdWf *pipelineConfig.CdWorkflow, cdPipeline *pipelineConfig.Pipeline, triggeredBy int32) (*CdWorkflowRequest, error) {
	cdWorkflowConfig, err := impl.cdWorkflowRepository.FindConfigByPipelineId(cdPipeline.Id)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}

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
	ciPipeline, err := impl.ciPipelineRepository.FindById(cdPipeline.CiPipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("cannot find ciPipeline", "err", err)
		return nil, err
	}

	for _, m := range ciPipeline.CiPipelineMaterials {
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

		commitTime, err := convert(ciMaterialCurrent.Modifications[0].ModifiedTime)
		if err != nil {
			return nil, err
		}
		ciProjectDetail := CiProjectDetails{
			GitRepository: ciMaterialCurrent.Material.GitConfiguration.URL,
			MaterialName:  gitMaterial.Name,
			CheckoutPath:  gitMaterial.CheckoutPath,
			CommitHash:    ciMaterialCurrent.Modifications[0].Revision,
			Author:        ciMaterialCurrent.Modifications[0].Author,
			SourceType:    m.Type,
			SourceValue:   m.Value,
			GitTag:        ciMaterialCurrent.Modifications[0].Tag,
			Message:       ciMaterialCurrent.Modifications[0].Message,
			Type:          string(m.Type),
			CommitTime:    *commitTime,
			GitOptions: GitOptions{
				UserName:    gitMaterial.GitProvider.UserName,
				Password:    gitMaterial.GitProvider.Password,
				SSHKey:      gitMaterial.GitProvider.SshKey,
				AccessToken: gitMaterial.GitProvider.AccessToken,
				AuthMode:    gitMaterial.GitProvider.AuthMode,
			},
		}

		// set webhook data
		if m.Type == pipelineConfig.SOURCE_TYPE_WEBHOOK {
			webhookData := ciMaterialCurrent.Modifications[0].WebhookData
			ciProjectDetail.WebhookData = pipelineConfig.WebhookData{
				Id:              webhookData.Id,
				EventActionType: webhookData.EventActionType,
				Data:            webhookData.Data,
			}
		}

		ciProjectDetails = append(ciProjectDetails, ciProjectDetail)
	}

	var stageYaml string
	if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
		stageYaml = cdPipeline.PreStageConfig
	} else if runner.WorkflowType == bean.CD_WORKFLOW_TYPE_POST {
		stageYaml = cdPipeline.PostStageConfig
	} else {
		return nil, fmt.Errorf("unsupported workflow triggerd")
	}
	extraEnvVariables := make(map[string]string)
	extraEnvVariables["APP_NAME"] = ciPipeline.App.AppName
	cdStageWorkflowRequest := &CdWorkflowRequest{
		EnvironmentId:         cdPipeline.EnvironmentId,
		AppId:                 cdPipeline.AppId,
		WorkflowId:            cdWf.Id,
		WorkflowRunnerId:      runner.Id,
		WorkflowNamePrefix:    strconv.Itoa(runner.Id) + "-" + runner.Name,
		CdImage:               impl.cdConfig.DefaultImage,
		CdPipelineId:          cdWf.PipelineId,
		TriggeredBy:           triggeredBy,
		StageYaml:             stageYaml,
		CiProjectDetails:      ciProjectDetails,
		Namespace:             runner.Namespace,
		ActiveDeadlineSeconds: impl.cdConfig.DefaultTimeout,
		DockerUsername:        ciPipeline.CiTemplate.DockerRegistry.Username,
		DockerPassword:        ciPipeline.CiTemplate.DockerRegistry.Password,
		AwsRegion:             ciPipeline.CiTemplate.DockerRegistry.AWSRegion,
		AccessKey:             ciPipeline.CiTemplate.DockerRegistry.AWSAccessKeyId,
		SecretKey:             ciPipeline.CiTemplate.DockerRegistry.AWSSecretAccessKey,
		DockerRegistryType:    string(ciPipeline.CiTemplate.DockerRegistry.RegistryType),
		DockerRegistryURL:     ciPipeline.CiTemplate.DockerRegistry.RegistryURL,
		CiArtifactDTO: CiArtifactDTO{
			Id:           artifact.Id,
			PipelineId:   artifact.PipelineId,
			Image:        artifact.Image,
			ImageDigest:  artifact.ImageDigest,
			MaterialInfo: artifact.MaterialInfo,
			DataSource:   artifact.DataSource,
			WorkflowId:   artifact.WorkflowId,
		},
		OrchestratorHost:          impl.cdConfig.OrchestratorHost,
		OrchestratorToken:         impl.cdConfig.OrchestratorToken,
		ExtraEnvironmentVariables: extraEnvVariables,
		CloudProvider:             impl.cdConfig.CloudProvider,
	}
	switch cdStageWorkflowRequest.CloudProvider {
	case BLOB_STORAGE_S3:
		//No AccessKey is used for uploading artifacts, instead IAM based auth is used
		cdStageWorkflowRequest.CdCacheRegion = cdWorkflowConfig.CdCacheRegion
		cdStageWorkflowRequest.CdCacheLocation = cdWorkflowConfig.CdCacheBucket
		cdStageWorkflowRequest.ArtifactLocation = impl.buildArtifactLocation(cdWorkflowConfig, cdWf, runner)
	case BLOB_STORAGE_AZURE:
		cdStageWorkflowRequest.AzureBlobConfig = &AzureBlobConfig{
			Enabled:              true,
			AccountName:          impl.cdConfig.AzureAccountName,
			BlobContainerCiCache: impl.cdConfig.AzureBlobContainerCiCache,
			AccountKey:           impl.cdConfig.AzureAccountKey,
		}
	case BLOB_STORAGE_MINIO:
		//For MINIO type blob storage, AccessKey & SecretAccessKey are injected through EnvVar
		cdStageWorkflowRequest.CdCacheLocation = cdWorkflowConfig.CdCacheBucket
		cdStageWorkflowRequest.ArtifactLocation = impl.buildArtifactLocation(cdWorkflowConfig, cdWf, runner)
		cdStageWorkflowRequest.MinioEndpoint = impl.cdConfig.MinioEndpoint
	default:
		return nil, fmt.Errorf("cloudprovider %s not supported", cdStageWorkflowRequest.CloudProvider)
	}
	return cdStageWorkflowRequest, nil
}

func (impl *WorkflowDagExecutorImpl) HandleDeploymentSuccessEvent(gitHash string) error {
	pipelineOverride, err := impl.pipelineOverrideRepository.FindByPipelineTriggerGitHash(gitHash)
	if err != nil {
		return err
	}
	cdWorkflow, err := impl.cdWorkflowRepository.FindById(pipelineOverride.CdWorkflowId)
	if err != nil {
		return err
	}
	if len(pipelineOverride.Pipeline.PostStageConfig) > 0 {
		if pipelineOverride.Pipeline.PostTriggerType == pipelineConfig.TRIGGER_TYPE_AUTOMATIC &&
			pipelineOverride.DeploymentType != models.DEPLOYMENTTYPE_STOP &&
			pipelineOverride.DeploymentType != models.DEPLOYMENTTYPE_START {

			err := impl.TriggerPostStage(cdWorkflow, pipelineOverride.Pipeline, 1)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//Only used for auto trigger
func (impl *WorkflowDagExecutorImpl) TriggerDeployment(cdWf *pipelineConfig.CdWorkflow, artifact *repository.CiArtifact, pipeline *pipelineConfig.Pipeline, applyAuth bool, async bool, triggeredBy int32) error {
	//in case of manual ci RBAC need to apply, this method used for auto cd deployment
	if applyAuth {
		user, err := impl.user.GetById(triggeredBy)
		if err != nil {
			impl.logger.Errorw("error in fetching user for auto pipeline", "UpdatedBy", artifact.UpdatedBy)
			return nil
		}
		token := user.EmailId
		object := impl.enforcerUtil.GetAppRBACNameByAppId(pipeline.AppId)
		impl.logger.Debugw("Triggered Request (App Permission Checking):", "token", token, "object", object)
		if ok := impl.enforcer.EnforceByEmail(strings.ToLower(token), rbac.ResourceApplications, rbac.ActionTrigger, object); !ok {
			return fmt.Errorf("unauthorized for pipeline " + strconv.Itoa(pipeline.Id))
		}
	}

	if cdWf == nil {
		cdWf = &pipelineConfig.CdWorkflow{
			CiArtifactId: artifact.Id,
			PipelineId:   pipeline.Id,
			AuditLog:     models.AuditLog{CreatedOn: time.Now(), CreatedBy: 1, UpdatedOn: time.Now(), UpdatedBy: 1},
		}
		err := impl.cdWorkflowRepository.SaveWorkFlow(cdWf)
		if err != nil {
			return err
		}
	}

	runner := &pipelineConfig.CdWorkflowRunner{
		Name:         pipeline.Name,
		WorkflowType: bean.CD_WORKFLOW_TYPE_DEPLOY,
		ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM,
		Status:       WorkflowStarting, //starting
		TriggeredBy:  1,
		StartedOn:    time.Now(),
		Namespace:    impl.cdConfig.DefaultNamespace,
		CdWorkflowId: cdWf.Id,
	}
	err := impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
	if err != nil {
		return err
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
		if blockCveList != nil && len(blockCveList) > 0 {
			isVulnerable = true
		}
	}
	if isVulnerable == true {
		runner.Status = WorkflowFailed
		runner.Message = "Found vulnerability on image"
		runner.FinishedOn = time.Now()
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
		if err != nil {
			impl.logger.Errorw("error in updating status", "err", err)
			return err
		}
		err := impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
		if err != nil {
			return err
		}
		return nil
	}

	err = impl.appService.TriggerCD(artifact, cdWf.Id, pipeline, async)
	err1 := impl.updatePreviousDeploymentStatus(runner, pipeline.Id, err)
	if err1 != nil || err != nil {
		impl.logger.Errorw("error while update previous cd workflow runners", "err", err, "runner", runner, "pipelineId", pipeline.Id)
		return err
	}
	return nil
}

func (impl *WorkflowDagExecutorImpl) updatePreviousDeploymentStatus(currentRunner *pipelineConfig.CdWorkflowRunner, pipelineId int, err error) error {
	if err != nil {
		impl.logger.Errorw("error in triggering cd WF, setting wf status as fail ", "wfId", currentRunner.Id, "err", err)
		currentRunner.Status = WorkflowFailed
		currentRunner.Message = err.Error()
		currentRunner.FinishedOn = time.Now()
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(currentRunner)
		if err != nil {
			impl.logger.Errorw("error updating cd wf runner status", "err", err, "currentRunner", currentRunner)
			return err
		}
		return nil
		//update current WF with error status
	} else {
		//update n-1th  deploy status as aborted if not termainal(Healthy, Degraded)
		terminalStatus := []string{v1alpha1.HealthStatusHealthy, v1alpha1.HealthStatusDegraded, WorkflowAborted, WorkflowFailed}
		previousNonTerminalRunners, err := impl.cdWorkflowRepository.FindPreviousCdWfRunnerByStatus(pipelineId, currentRunner.Id, terminalStatus)
		if err != nil {
			impl.logger.Errorw("error fetching previous wf runner, updating cd wf runner status,", "err", err, "currentRunner", currentRunner)
			return err
		} else if len(previousNonTerminalRunners) == 0 {
			impl.logger.Errorw("no previous runner found in updating cd wf runner status,", "err", err, "currentRunner", currentRunner)
			return nil
		}
		for _, previousRunner := range previousNonTerminalRunners {
			if previousRunner.Status == v1alpha1.HealthStatusHealthy ||
				previousRunner.Status == v1alpha1.HealthStatusDegraded ||
				previousRunner.Status == WorkflowAborted ||
				previousRunner.Status == WorkflowFailed {
				//terminal status return
				impl.logger.Infow("skip updating cd wf runner status as previous runner status is", "status", previousRunner.Status)
				return nil
			}
			impl.logger.Infow("updating cd wf runner status as previous runner status is", "status", previousRunner.Status)
			previousRunner.FinishedOn = time.Now()
			previousRunner.Message = "triggered new deployment"
			previousRunner.Status = WorkflowAborted
		}

		err = impl.cdWorkflowRepository.UpdateWorkFlowRunners(previousNonTerminalRunners)
		if err != nil {
			impl.logger.Errorw("error updating cd wf runner status", "err", err, "previousNonTerminalRunners", previousNonTerminalRunners)
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

func (impl *WorkflowDagExecutorImpl) ManualCdTrigger(overrideRequest *bean.ValuesOverrideRequest, ctx context.Context) (int, error) {
	releaseId := 0
	var err error
	cdPipeline, err := impl.pipelineRepository.FindById(overrideRequest.PipelineId)
	if err != nil {
		impl.logger.Errorf("invalid req", "err", err, "req", overrideRequest)
		return 0, err
	}

	if overrideRequest.CdWorkflowType == bean.CD_WORKFLOW_TYPE_PRE {
		artifact, err := impl.ciArtifactRepository.Get(overrideRequest.CiArtifactId)
		if err != nil {
			return 0, err
		}
		err = impl.TriggerPreStage(nil, artifact, cdPipeline, overrideRequest.UserId, false)
		if err != nil {
			return 0, err
		}
	} else if overrideRequest.CdWorkflowType == bean.CD_WORKFLOW_TYPE_DEPLOY {
		if overrideRequest.DeploymentType == models.DEPLOYMENTTYPE_UNKNOWN {
			overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_DEPLOY
		}
		cdWf, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(overrideRequest.CdWorkflowId, bean.CD_WORKFLOW_TYPE_PRE)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("err", "err", err)
			return 0, nil
		}

		cdWorkflowId := cdWf.CdWorkflowId
		if cdWf.CdWorkflowId == 0 {
			cdWf := &pipelineConfig.CdWorkflow{
				CiArtifactId: overrideRequest.CiArtifactId,
				PipelineId:   overrideRequest.PipelineId,
				AuditLog:     models.AuditLog{CreatedOn: time.Now(), CreatedBy: overrideRequest.UserId, UpdatedOn: time.Now(), UpdatedBy: overrideRequest.UserId},
			}
			err := impl.cdWorkflowRepository.SaveWorkFlow(cdWf)
			if err != nil {
				return 0, err
			}
			cdWorkflowId = cdWf.Id
		}

		runner := &pipelineConfig.CdWorkflowRunner{
			Name:         cdPipeline.Name,
			WorkflowType: bean.CD_WORKFLOW_TYPE_DEPLOY,
			ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
			Status:       WorkflowStarting,
			TriggeredBy:  overrideRequest.UserId,
			StartedOn:    time.Now(),
			Namespace:    impl.cdConfig.DefaultNamespace,
			CdWorkflowId: cdWorkflowId,
		}
		err = impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
		if err != nil {
			return 0, err
		}
		overrideRequest.CdWorkflowId = cdWorkflowId

		//checking vulnerability for deploying image
		artifact, err := impl.ciArtifactRepository.Get(overrideRequest.CiArtifactId)
		if err != nil {
			return 0, err
		}
		isVulnerable := false
		if len(artifact.ImageDigest) > 0 {
			var cveStores []*security.CveStore
			imageScanResult, err := impl.scanResultRepository.FindByImageDigest(artifact.ImageDigest)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error fetching image digest", "digest", artifact.ImageDigest, "err", err)
				return 0, err
			}
			for _, item := range imageScanResult {
				cveStores = append(cveStores, &item.CveStore)
			}
			blockCveList, err := impl.cvePolicyRepository.GetBlockedCVEList(cveStores, cdPipeline.Environment.ClusterId, cdPipeline.EnvironmentId, cdPipeline.AppId, false)
			if err != nil {
				impl.logger.Errorw("error while fetching env", "err", err)
				return 0, err
			}
			if blockCveList != nil && len(blockCveList) > 0 {
				isVulnerable = true
			}
		}
		if isVulnerable == true {
			runner := &pipelineConfig.CdWorkflowRunner{
				Name:         cdPipeline.Name,
				WorkflowType: bean.CD_WORKFLOW_TYPE_DEPLOY,
				ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM,
				Status:       WorkflowFailed, //starting
				TriggeredBy:  1,
				StartedOn:    time.Now(),
				Namespace:    impl.cdConfig.DefaultNamespace,
				CdWorkflowId: cdWf.Id,
				Message:      "Found vulnerability on image",
			}
			err := impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
			if err != nil {
				return 0, err
			}
			return 0, fmt.Errorf("found vulnerability for image digest %s", artifact.ImageDigest)
		}

		releaseId, err = impl.appService.TriggerRelease(overrideRequest, ctx)
		//	return after error handling
		/*if err != nil {
			return 0, err
		}*/
		err1 := impl.updatePreviousDeploymentStatus(runner, cdPipeline.Id, err)
		if err1 != nil || err != nil {
			impl.logger.Errorw("error while update previous cd workflow runners", "err", err, "runner", runner, "pipelineId", cdPipeline.Id)
			return 0, err
		}
	} else if overrideRequest.CdWorkflowType == bean.CD_WORKFLOW_TYPE_POST {
		cdWfRunner, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(overrideRequest.CdWorkflowId, bean.CD_WORKFLOW_TYPE_DEPLOY)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("err", "err", err)
			return 0, nil
		}

		if cdWfRunner.CdWorkflowId == 0 {
			cdWf := &pipelineConfig.CdWorkflow{
				CiArtifactId: overrideRequest.CiArtifactId,
				PipelineId:   overrideRequest.PipelineId,
				AuditLog:     models.AuditLog{CreatedOn: time.Now(), CreatedBy: overrideRequest.UserId, UpdatedOn: time.Now(), UpdatedBy: overrideRequest.UserId},
			}
			err := impl.cdWorkflowRepository.SaveWorkFlow(cdWf)
			if err != nil {
				return 0, err
			}
			err = impl.TriggerPostStage(cdWf, cdPipeline, overrideRequest.UserId)
		} else {
			cdWf, err := impl.cdWorkflowRepository.FindById(overrideRequest.CdWorkflowId)
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("err", "err", err)
				return 0, nil
			}
			err = impl.TriggerPostStage(cdWf, cdPipeline, overrideRequest.UserId)
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
			AuditLog:       models.AuditLog{CreatedOn: time.Now(), CreatedBy: UserId, UpdatedOn: time.Now(), UpdatedBy: UserId},
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
			impl.pubsubClient.Conn.PublishAsync(BULK_HIBERNATE_TOPIC, data, func(s string, err error) {
				if err != nil {
					impl.logger.Errorw("error while writing app stop event to nats ", "app", app.AppId, "deploymentGroup", app.DeploymentGroupId, "err", err)
				}
			})
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
			impl.pubsubClient.Conn.PublishAsync(BULK_DEPLOY_TOPIC, data, func(s string, err error) {
				if err != nil {
					wf.WorkflowStatus = pipelineConfig.QUE_ERROR
				} else {
					wf.WorkflowStatus = pipelineConfig.ENQUEUED
				}
			})
		}
		err = impl.cdWorkflowRepository.UpdateWorkFlow(wf)
		if err != nil {
			impl.logger.Errorw("error in publishing wf msg", "wf", wf, "err", err)
		}
	}
}

func (impl *WorkflowDagExecutorImpl) subscribeTriggerBulkAction() error {
	_, err := impl.pubsubClient.Conn.QueueSubscribe(BULK_DEPLOY_TOPIC, BULK_DEPLOY_GROUP, func(msg *stan.Msg) {
		impl.logger.Debug("subscribeTriggerBulkAction event received")
		defer msg.Ack()
		cdWorkflow := new(pipelineConfig.CdWorkflow)
		err := json.Unmarshal([]byte(string(msg.Data)), cdWorkflow)
		if err != nil {
			impl.logger.Error("err", err)
			return
		}
		impl.logger.Debugw("subscribeTriggerBulkAction event:", "cdWorkflow", cdWorkflow)
		wf := &pipelineConfig.CdWorkflow{
			Id:           cdWorkflow.Id,
			CiArtifactId: cdWorkflow.CiArtifactId,
			PipelineId:   cdWorkflow.PipelineId,
			AuditLog: models.AuditLog{
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
	}, stan.DurableName(BULK_DEPLOY_DURABLE), stan.StartWithLastReceived(), stan.AckWait(time.Duration(180)*time.Second), stan.SetManualAckMode(), stan.MaxInflight(1))
	return err
}

func (impl *WorkflowDagExecutorImpl) subscribeHibernateBulkAction() error {
	_, err := impl.pubsubClient.Conn.QueueSubscribe(BULK_HIBERNATE_TOPIC, BULK_HIBERNATE_GROUP, func(msg *stan.Msg) {
		impl.logger.Debug("subscribeHibernateBulkAction event received")
		defer msg.Ack()
		deploymentGroupAppWithEnv := new(DeploymentGroupAppWithEnv)
		err := json.Unmarshal([]byte(string(msg.Data)), deploymentGroupAppWithEnv)
		if err != nil {
			impl.logger.Error("err", err)
			return
		}
		impl.logger.Debugw("subscribeHibernateBulkAction event:", "DeploymentGroupAppWithEnv", deploymentGroupAppWithEnv)

		stopAppRequest := &StopAppRequest{
			AppId:         deploymentGroupAppWithEnv.AppId,
			EnvironmentId: deploymentGroupAppWithEnv.EnvironmentId,
			UserId:        deploymentGroupAppWithEnv.UserId,
			RequestType:   deploymentGroupAppWithEnv.RequestType,
		}
		ctx, err := impl.buildACDSynchContext()
		if err != nil {
			impl.logger.Errorw("error in creating acd synch context", "err", err)
			return
		}
		_, err = impl.StopStartApp(stopAppRequest, ctx)
	}, stan.DurableName(BULK_HIBERNATE_DURABLE), stan.StartWithLastReceived(), stan.AckWait(time.Duration(180)*time.Second), stan.SetManualAckMode(), stan.MaxInflight(1))
	return err
}

func (impl *WorkflowDagExecutorImpl) buildACDSynchContext() (acdContext context.Context, err error) {
	return impl.tokenCache.BuildACDSynchContext()
}
