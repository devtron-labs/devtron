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
	v1alpha12 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/go-pg/pg"
	"github.com/nats-io/stan.go"
	"go.uber.org/zap"
	"time"
)

type ApplicationStatusUpdateHandler interface {
	Subscribe() error
}

type ApplicationStatusUpdateHandlerImpl struct {
	logger              *zap.SugaredLogger
	pubsubClient        *pubsub.PubSubClient
	appService          app.AppService
	workflowDagExecutor pipeline.WorkflowDagExecutor
	installedAppService service.InstalledAppService
}

const applicationStatusUpdate = "APPLICATION_STATUS_UPDATE"
const applicationStatusUpdateGroup = "APPLICATION_STATUS_UPDATE_GROUP-1"
const applicationStatusUpdateDurable = "APPLICATION_STATUS_UPDATE_DURABLE-1"

func NewApplicationStatusUpdateHandlerImpl(logger *zap.SugaredLogger, pubsubClient *pubsub.PubSubClient, appService app.AppService,
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
		logger.Error("err", err)
		return nil
	}
	return appStatusUpdateHandlerImpl
}

func (impl *ApplicationStatusUpdateHandlerImpl) Subscribe() error {
	_, err := impl.pubsubClient.Conn.QueueSubscribe(applicationStatusUpdate, applicationStatusUpdateGroup, func(msg *stan.Msg) {
		impl.logger.Debug("received app update request")
		defer msg.Ack()
		application := v1alpha12.Application{}
		err := json.Unmarshal([]byte(string(msg.Data)), &application)
		if err != nil {
			impl.logger.Errorw("unmarshal error on app update status", "err", err)
			return
		}
		impl.logger.Infow("app update request", "application", application)
		isHealthy, err := impl.appService.UpdateApplicationStatusAndCheckIsHealthy(application)
		if err != nil {
			impl.logger.Errorw("error on application status update", "err", err, "msg", string(msg.Data))

			//TODO - check update for charts - fix this call
			if err == pg.ErrNoRows {
				// if not found in charts (which is for devtron apps) try to find in installed app (which is for devtron charts)
				_, err := impl.installedAppService.UpdateInstalledAppVersionStatus(application)
				if err != nil {
					impl.logger.Errorw("error on application status update", "err", err, "msg", string(msg.Data))
					return
				}
			}
			// return anyways weather updates or failure, no further processing for charts status update
			return
		}

		// invoke DagExecutor, for cd success which will trigger post stage if exist.
		if isHealthy == true {
			impl.logger.Debugw("git hash history", "list", application.Status.History)
			var gitHash string
			if application.Operation != nil && application.Operation.Sync != nil {
				gitHash = application.Operation.Sync.Revision
			} else if application.Status.OperationState != nil && application.Status.OperationState.Operation.Sync != nil {
				gitHash = application.Status.OperationState.Operation.Sync.Revision
			}
			err = impl.workflowDagExecutor.HandleDeploymentSuccessEvent(gitHash)
			if err != nil {
				impl.logger.Errorw("deployment success event error", "gitHash", gitHash, "err", err)
				return
			}
		}
		impl.logger.Debug("app" + application.Name + " updated")
	}, stan.DurableName(applicationStatusUpdateDurable), stan.StartWithLastReceived(), stan.AckWait(time.Duration(impl.pubsubClient.AckDuration)*time.Second), stan.SetManualAckMode(), stan.MaxInflight(1))

	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}
