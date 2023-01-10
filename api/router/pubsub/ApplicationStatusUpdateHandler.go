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
	"encoding/json"
	"time"

	v1alpha12 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ApplicationStatusUpdateHandler interface {
	Subscribe() error
}

type ApplicationStatusUpdateHandlerImpl struct {
	logger              *zap.SugaredLogger
	pubsubClient        *pubsub.PubSubClientServiceImpl
	appService          app.AppService
	workflowDagExecutor pipeline.WorkflowDagExecutor
	installedAppService service.InstalledAppService
}

func NewApplicationStatusUpdateHandlerImpl(logger *zap.SugaredLogger, pubsubClient *pubsub.PubSubClientServiceImpl, appService app.AppService,
	workflowDagExecutor pipeline.WorkflowDagExecutor, installedAppService service.InstalledAppService) *ApplicationStatusUpdateHandlerImpl {
	appStatusUpdateHandlerImpl := &ApplicationStatusUpdateHandlerImpl{
		logger:              logger,
		pubsubClient:        pubsubClient,
		appService:          appService,
		workflowDagExecutor: workflowDagExecutor,
		installedAppService: installedAppService,
	}
	err := appStatusUpdateHandlerImpl.Subscribe()
	if err != nil {
		//logger.Error("err", err)
		return nil
	}
	return appStatusUpdateHandlerImpl
}

type ApplicationDetail struct {
	Application *v1alpha12.Application `json:"application"`
	StatusTime  time.Time              `json:"statusTime"`
}

func (impl *ApplicationStatusUpdateHandlerImpl) Subscribe() error {
	callback := func(msg *pubsub.PubSubMsg) {
		impl.logger.Debug("received app update request")
		//defer msg.Ack()
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
		isSucceeded, err := impl.appService.UpdateDeploymentStatusAndCheckIsSucceeded(app, applicationDetail.StatusTime)
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
