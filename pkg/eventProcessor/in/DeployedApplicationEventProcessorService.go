/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package in

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	installedAppReader "github.com/devtron-labs/devtron/pkg/appStore/installedApp/read"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/common"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	"github.com/devtron-labs/devtron/pkg/workflow/dag"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"k8s.io/utils/pointer"
	"time"
)

type DeployedApplicationEventProcessorImpl struct {
	logger                    *zap.SugaredLogger
	pubSubClient              *pubsub.PubSubClientServiceImpl
	appService                app.AppService
	gitOpsConfigReadService   config.GitOpsConfigReadService
	installedAppService       FullMode.InstalledAppDBExtendedService
	workflowDagExecutor       dag.WorkflowDagExecutor
	cdWorkflowCommonService   cd.CdWorkflowCommonService
	pipelineBuilder           pipeline.PipelineBuilder
	appStoreDeploymentService service.AppStoreDeploymentService
	pipelineRepository        pipelineConfig.PipelineRepository // TODO: should use cdPipelineReadService instead
	installedAppReadService   installedAppReader.InstalledAppReadService
	DeploymentConfigService   common.DeploymentConfigService
}

func NewDeployedApplicationEventProcessorImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	appService app.AppService,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	installedAppService FullMode.InstalledAppDBExtendedService,
	workflowDagExecutor dag.WorkflowDagExecutor,
	cdWorkflowCommonService cd.CdWorkflowCommonService,
	pipelineBuilder pipeline.PipelineBuilder,
	appStoreDeploymentService service.AppStoreDeploymentService,
	pipelineRepository pipelineConfig.PipelineRepository,
	installedAppReadService installedAppReader.InstalledAppReadService,
	DeploymentConfigService common.DeploymentConfigService) *DeployedApplicationEventProcessorImpl {
	deployedApplicationEventProcessorImpl := &DeployedApplicationEventProcessorImpl{
		logger:                    logger,
		pubSubClient:              pubSubClient,
		appService:                appService,
		gitOpsConfigReadService:   gitOpsConfigReadService,
		installedAppService:       installedAppService,
		workflowDagExecutor:       workflowDagExecutor,
		cdWorkflowCommonService:   cdWorkflowCommonService,
		pipelineBuilder:           pipelineBuilder,
		appStoreDeploymentService: appStoreDeploymentService,
		pipelineRepository:        pipelineRepository,
		installedAppReadService:   installedAppReadService,
		DeploymentConfigService:   DeploymentConfigService,
	}
	return deployedApplicationEventProcessorImpl
}

func (impl *DeployedApplicationEventProcessorImpl) SubscribeArgoAppUpdate() error {
	callback := func(msg *model.PubSubMsg) {
		applicationDetail := bean3.ApplicationDetail{}
		err := json.Unmarshal([]byte(msg.Data), &applicationDetail)
		if err != nil {
			impl.logger.Errorw("unmarshal error on app update status", "err", err)
			return
		}
		app := applicationDetail.Application
		if app == nil {
			return
		}
		if applicationDetail.StatusTime.IsZero() {
			applicationDetail.StatusTime = time.Now()
		}
		isAppStoreApplication := false
		pipelines, err := impl.pipelineRepository.GetArgoPipelineByArgoAppName(app.ObjectMeta.Name)
		if err != nil {
			impl.logger.Errorw("error in fetching pipeline from Pipeline Repository", "err", err, "appName", app.ObjectMeta.Name)
			return
		}
		if len(pipelines) == 0 {
			impl.logger.Infow("this app not found in pipeline table looking in installed_apps table", "appName", app.ObjectMeta.Name)
			// if not found in pipeline table then search in installed_apps table
			installedAppModel, err := impl.installedAppReadService.GetInstalledAppByGitOpsAppName(app.ObjectMeta.Name)
			if err == pg.ErrNoRows {
				// no installed_apps found
				impl.logger.Errorw("no installed apps found", "err", err)
				return
			}
			if err != nil {
				impl.logger.Errorw("error in getting all gitops deployment app names from installed_apps ", "err", err)
				return
			}
			if installedAppModel.Id > 0 {
				// app found in installed_apps table hence setting flag to true
				isAppStoreApplication = true
			} else {
				// app neither found in installed_apps nor in pipeline table hence returning
				return
			}
		}
		isSucceeded, _, pipelineOverride, err := impl.appService.UpdateDeploymentStatusForGitOpsPipelines(app, applicationDetail.ClusterId, applicationDetail.StatusTime, isAppStoreApplication)
		if err != nil {
			impl.logger.Errorw("error on application status update", "err", err, "msg", string(msg.Data))
			// TODO - check update for charts - fix this call
			if err == pg.ErrNoRows {
				// if not found in charts (which is for devtron apps) try to find in installed app (which is for devtron charts)
				_, err := impl.installedAppService.UpdateInstalledAppVersionStatus(app)
				if err != nil {
					impl.logger.Errorw("error on application status update", "err", err, "msg", string(msg.Data))
					return
				}
			}
			// return anyway whether updates or failure, no further processing for charts status update
			return
		}

		// invoke DagExecutor, for cd success which will trigger post stage if exist.
		if isSucceeded {
			impl.logger.Debugw("git hash history", "list", app.Status.History)
			triggerContext := bean2.TriggerContext{
				ReferenceId: pointer.String(msg.MsgId),
			}
			err = impl.workflowDagExecutor.HandleDeploymentSuccessEvent(triggerContext, pipelineOverride)
			if err != nil {
				impl.logger.Errorw("deployment success event error", "pipelineOverride", pipelineOverride, "err", err)
				return
			}
		}
		impl.logger.Debugw("application status update completed", "app", app.Name)
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		return "", nil
	}

	validations := impl.cdWorkflowCommonService.GetTriggerValidateFuncs()
	err := impl.pubSubClient.Subscribe(pubsub.APPLICATION_STATUS_UPDATE_TOPIC, callback, loggerFunc, validations...)
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}

func (impl *DeployedApplicationEventProcessorImpl) SubscribeArgoAppDeleteStatus() error {
	callback := func(msg *model.PubSubMsg) {

		applicationDetail := bean3.ApplicationDetail{}
		err := json.Unmarshal([]byte(msg.Data), &applicationDetail)
		if err != nil {
			impl.logger.Errorw("unmarshal error on app delete status", "err", err)
			return
		}
		app := applicationDetail.Application
		if app == nil {
			return
		}
		impl.logger.Infow("argo delete event received", "appName", app.Name, "namespace", app.Namespace, "deleteTimestamp", app.DeletionTimestamp)

		err = impl.updateArgoAppDeleteStatus(applicationDetail)
		if err != nil {
			impl.logger.Errorw("error in updating pipeline delete status", "err", err, "appName", app.Name)
		}
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		applicationDetail := bean3.ApplicationDetail{}
		err := json.Unmarshal([]byte(msg.Data), &applicationDetail)
		if err != nil {
			return "unmarshal error on app delete status", []interface{}{"err", err}
		}
		return "got message for application status delete", []interface{}{"appName", applicationDetail.Application.Name, "namespace", applicationDetail.Application.Namespace, "deleteTimestamp", applicationDetail.Application.DeletionTimestamp}
	}

	err := impl.pubSubClient.Subscribe(pubsub.APPLICATION_STATUS_DELETE_TOPIC, callback, loggerFunc)
	if err != nil {
		impl.logger.Errorw("error in subscribing to argo application status delete topic", "err", err)
		return err
	}
	return nil
}

func (impl *DeployedApplicationEventProcessorImpl) updateHelmAppArgoAppDeleteStatus(application *v1alpha1.Application) error {
	// Helm app deployed using argocd
	var gitHash string
	if application.Operation != nil && application.Operation.Sync != nil {
		gitHash = application.Operation.Sync.Revision
	} else if application.Status.OperationState != nil && application.Status.OperationState.Operation.Sync != nil {
		gitHash = application.Status.OperationState.Operation.Sync.Revision
	}
	installedAppDeleteReq, err := impl.installedAppReadService.GetInstalledAppByGitHash(gitHash)
	if err != nil {
		impl.logger.Errorw("error in fetching installed app by git hash from installed app repository", "err", err)
		return err
	}

	// Check to ensure that delete request for app was received
	installedApp, err := impl.installedAppService.GetInstalledAppById(installedAppDeleteReq.InstalledAppId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching app from installed app repository", "err", err, "installedAppId", installedAppDeleteReq.InstalledAppId)
		return err
	} else if errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("App not found in database", "installedAppId", installedAppDeleteReq.InstalledAppId, "err", err)
		return fmt.Errorf("app not found in database %s", err)
	}

	if installedApp.DeploymentAppDeleteRequest == false {
		// TODO 4465 remove app from log after final RCA
		impl.logger.Infow("Deployment delete not requested for app, not deleting app from DB", "appName", application.Name, "installedApp", installedApp)
		return nil
	}

	deleteRequest := &appStoreBean.InstallAppVersionDTO{}
	deleteRequest.ForceDelete = false
	deleteRequest.NonCascadeDelete = false
	deleteRequest.AcdPartialDelete = false
	deleteRequest.InstalledAppId = installedAppDeleteReq.InstalledAppId
	deleteRequest.AppId = installedAppDeleteReq.AppId
	deleteRequest.AppName = installedAppDeleteReq.AppName
	deleteRequest.Namespace = installedAppDeleteReq.Namespace
	deleteRequest.ClusterId = installedAppDeleteReq.ClusterId
	deleteRequest.EnvironmentId = installedAppDeleteReq.EnvironmentId
	deleteRequest.AppOfferingMode = installedAppDeleteReq.AppOfferingMode
	deleteRequest.UserId = 1
	_, err = impl.appStoreDeploymentService.DeleteInstalledApp(context.Background(), deleteRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting installed app", "err", err)
		return err
	}
	return nil
}

func (impl *DeployedApplicationEventProcessorImpl) updateDevtronAppArgoAppDeleteStatus(applicationDetail bean3.ApplicationDetail,
	pipelines []pipelineConfig.Pipeline) error {
	application := applicationDetail.Application
	pipelineModel, err := impl.DeploymentConfigService.FilterPipelinesByApplicationClusterIdAndNamespace(pipelines, applicationDetail.ClusterId, application.Namespace)
	if err != nil {
		impl.logger.Errorw("error in filtering pipeline by application cluster id and namespace", "err", err)
		return err
	}
	if pipelineModel.Deleted == true {
		impl.logger.Errorw("invalid nats message, pipeline already deleted")
		return errors.New("invalid nats message, pipeline already deleted")
	}
	// devtron app
	if pipelineModel.DeploymentAppDeleteRequest == false {
		impl.logger.Infow("Deployment delete not requested for app, not deleting app from DB", "appName", application.Name, "app", application)
		return nil
	}
	_, err = impl.pipelineBuilder.DeleteCdPipeline(&pipelineModel, context.Background(), bean.FORCE_DELETE, false, 1)
	if err != nil {
		impl.logger.Errorw("error in deleting cd pipeline", "err", err)
		return err
	}
	return nil
}

func (impl *DeployedApplicationEventProcessorImpl) updateArgoAppDeleteStatus(applicationDetail bean3.ApplicationDetail) error {
	application := applicationDetail.Application
	pipelines, err := impl.pipelineRepository.GetArgoPipelineByArgoAppName(application.ObjectMeta.Name)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline from Pipeline Repository", "err", err)
		return err
	}
	if len(pipelines) == 0 {
		return impl.updateHelmAppArgoAppDeleteStatus(application)
	}
	return impl.updateDevtronAppArgoAppDeleteStatus(applicationDetail, pipelines)
}
