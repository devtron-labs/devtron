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

package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	repository4 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	"k8s.io/utils/strings/slices"
	"time"

	v1alpha12 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ApplicationStatusHandler interface {
	Subscribe() error
	SubscribeDeleteStatus() error
}

type ApplicationStatusHandlerImpl struct {
	logger                    *zap.SugaredLogger
	pubsubClient              *pubsub.PubSubClientServiceImpl
	appService                app.AppService
	workflowDagExecutor       pipeline.WorkflowDagExecutor
	installedAppService       service.InstalledAppService
	appStoreDeploymentService service.AppStoreDeploymentService
	pipelineBuilder           pipeline.PipelineBuilder
	pipelineRepository        pipelineConfig.PipelineRepository
	installedAppRepository    repository4.InstalledAppRepository
}

func NewApplicationStatusHandlerImpl(logger *zap.SugaredLogger, pubsubClient *pubsub.PubSubClientServiceImpl, appService app.AppService,
	workflowDagExecutor pipeline.WorkflowDagExecutor, installedAppService service.InstalledAppService,
	appStoreDeploymentService service.AppStoreDeploymentService, pipelineBuilder pipeline.PipelineBuilder,
	pipelineRepository pipelineConfig.PipelineRepository, installedAppRepository repository4.InstalledAppRepository) *ApplicationStatusHandlerImpl {
	appStatusUpdateHandlerImpl := &ApplicationStatusHandlerImpl{
		logger:                    logger,
		pubsubClient:              pubsubClient,
		appService:                appService,
		workflowDagExecutor:       workflowDagExecutor,
		installedAppService:       installedAppService,
		appStoreDeploymentService: appStoreDeploymentService,
		pipelineBuilder:           pipelineBuilder,
		pipelineRepository:        pipelineRepository,
		installedAppRepository:    installedAppRepository,
	}
	err := appStatusUpdateHandlerImpl.Subscribe()
	if err != nil {
		//logger.Error("err", err)
		return nil
	}
	err = appStatusUpdateHandlerImpl.SubscribeDeleteStatus()
	if err != nil {
		return nil
	}
	return appStatusUpdateHandlerImpl
}

type ApplicationDetail struct {
	Application *v1alpha12.Application `json:"application"`
	StatusTime  time.Time              `json:"statusTime"`
}

func (impl *ApplicationStatusHandlerImpl) Subscribe() error {
	callback := func(msg *pubsub.PubSubMsg) {
		impl.logger.Debugw("APP_STATUS_UPDATE_REQ", "stage", "raw", "data", msg.Data)
		applicationDetail := ApplicationDetail{}
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
		_, err = impl.pipelineRepository.GetArgoPipelineByArgoAppName(app.ObjectMeta.Name)
		if err != nil && err == pg.ErrNoRows {
			impl.logger.Infow("this app not found in pipeline table looking in installed_apps table", "appName", app.ObjectMeta.Name)
			//if not found in pipeline table then search in installed_apps table
			gitOpsDeployedAppNames, err := impl.installedAppRepository.GetAllGitOpsDeploymentAppName()
			if err != nil && err == pg.ErrNoRows {
				//no installed_apps found
				impl.logger.Errorw("no installed apps found", "err", err)
				return
			} else if err != nil {
				impl.logger.Errorw("error in getting all gitops deployment app names from installed_apps ", "err", err)
				return
			}
			var devtronGitOpsAppName string
			gitOpsRepoPrefix := impl.appService.GetGitOpsRepoPrefix()
			if len(gitOpsRepoPrefix) > 0 {
				devtronGitOpsAppName = fmt.Sprintf("%s-%s", gitOpsRepoPrefix, app.ObjectMeta.Name)
			} else {
				devtronGitOpsAppName = app.ObjectMeta.Name
			}
			if slices.Contains(gitOpsDeployedAppNames, devtronGitOpsAppName) {
				//app found in installed_apps table hence setting flag to true
				isAppStoreApplication = true
			} else {
				//app neither found in installed_apps nor in pipeline table hence returning
				return
			}
		}
		isSucceeded, err := impl.appService.UpdateDeploymentStatusAndCheckIsSucceeded(app, applicationDetail.StatusTime, isAppStoreApplication)
		if err != nil {
			impl.logger.Errorw("error on application status update", "err", err, "msg", string(msg.Data))
			//TODO - check update for charts - fix this call
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
			gitHash := ""
			if app != nil {
				gitHash = app.Status.Sync.Revision
			}
			err = impl.workflowDagExecutor.HandleDeploymentSuccessEvent(gitHash, 0)
			if err != nil {
				impl.logger.Errorw("deployment success event error", "gitHash", gitHash, "err", err)
				return
			}
		}
		impl.logger.Debugw("application status update completed", "app", app.Name)
	}

	err := impl.pubsubClient.Subscribe(pubsub.APPLICATION_STATUS_UPDATE_TOPIC, callback)
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}

func (impl *ApplicationStatusHandlerImpl) SubscribeDeleteStatus() error {
	callback := func(msg *pubsub.PubSubMsg) {
		impl.logger.Debug("received app delete event")

		impl.logger.Debugw("APP_STATUS_DELETE_REQ", "stage", "raw", "data", msg.Data)
		applicationDetail := ApplicationDetail{}
		err := json.Unmarshal([]byte(msg.Data), &applicationDetail)
		if err != nil {
			impl.logger.Errorw("unmarshal error on app delete status", "err", err)
			return
		}
		app := applicationDetail.Application
		if app == nil {
			return
		}
		err = impl.updateArgoAppDeleteStatus(app)
		if err != nil {
			impl.logger.Errorw("error in updating pipeline delete status", "err", err)
		}
	}
	err := impl.pubsubClient.Subscribe(pubsub.APPLICATION_STATUS_DELETE_TOPIC, callback)
	if err != nil {
		impl.logger.Errorw("error in subscribing to argo application status delete topic", "err", err)
		return err
	}
	return nil
}

func (impl *ApplicationStatusHandlerImpl) updateArgoAppDeleteStatus(app *v1alpha12.Application) error {
	pipeline, err := impl.pipelineRepository.GetArgoPipelineByArgoAppName(app.ObjectMeta.Name)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching pipeline from Pipeline Repository", "err", err)
		return err
	}
	if pipeline.Deleted == true {
		impl.logger.Errorw("invalid nats message, pipeline already deleted")
		return errors.New("invalid nats message, pipeline already deleted")
	}
	if err == pg.ErrNoRows {
		//Helm app deployed using argocd
		var gitHash string
		if app.Operation != nil && app.Operation.Sync != nil {
			gitHash = app.Operation.Sync.Revision
		} else if app.Status.OperationState != nil && app.Status.OperationState.Operation.Sync != nil {
			gitHash = app.Status.OperationState.Operation.Sync.Revision
		}
		model, err := impl.installedAppRepository.GetInstalledAppByGitHash(gitHash)
		if err != nil {
			impl.logger.Errorw("error in fetching installed app by git hash from installed app repository", "err", err)
			return err
		}
		deleteRequest := &appStoreBean.InstallAppVersionDTO{}
		deleteRequest.ForceDelete = false
		deleteRequest.AcdPartialDelete = false
		deleteRequest.InstalledAppId = model.InstalledAppId
		deleteRequest.AppId = model.AppId
		deleteRequest.AppName = model.AppName
		deleteRequest.Namespace = model.Namespace
		deleteRequest.ClusterId = model.ClusterId
		deleteRequest.EnvironmentId = model.EnvironmentId
		deleteRequest.AppOfferingMode = model.AppOfferingMode
		deleteRequest.UserId = 1
		_, err = impl.appStoreDeploymentService.DeleteInstalledApp(context.Background(), deleteRequest)
		if err != nil {
			impl.logger.Errorw("error in deleting installed app", "err", err)
			return err
		}
	} else {
		// devtron app
		err = impl.pipelineBuilder.DeleteCdPipeline(&pipeline, context.Background(), true, false, 0)
		if err != nil {
			impl.logger.Errorw("error in deleting cd pipeline", "err", err)
			return err
		}
	}
	return nil
}
