/*
 * Copyright (c) 2024. Devtron Inc.
 */

package in

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/chartGroup"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"go.uber.org/zap"
)

type AppStoreAppsEventProcessorImpl struct {
	logger            *zap.SugaredLogger
	pubSubClient      *pubsub.PubSubClientServiceImpl
	chartGroupService chartGroup.ChartGroupService

	iavHistoryRepository repository.InstalledAppVersionHistoryRepository
}

func NewAppStoreAppsEventProcessorImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	chartGroupService chartGroup.ChartGroupService,
	iavHistoryRepository repository.InstalledAppVersionHistoryRepository) *AppStoreAppsEventProcessorImpl {
	return &AppStoreAppsEventProcessorImpl{
		logger:               logger,
		pubSubClient:         pubSubClient,
		chartGroupService:    chartGroupService,
		iavHistoryRepository: iavHistoryRepository,
	}
}

func (impl *AppStoreAppsEventProcessorImpl) SubscribeAppStoreAppsBulkDeployEvent() error {
	callback := func(msg *model.PubSubMsg) {
		deployPayload := &bean.BulkDeployPayload{}
		err := json.Unmarshal([]byte(msg.Data), &deployPayload)
		if err != nil {
			impl.logger.Error("Error while unmarshalling deployPayload json object", "error", err)
			return
		}
		impl.logger.Debugw("deployPayload:", "deployPayload", deployPayload)
		//using userId 1 - for system user
		_, err = impl.chartGroupService.PerformDeployStage(deployPayload.InstalledAppVersionId, deployPayload.InstalledAppVersionHistoryId, 1)
		if err != nil {
			impl.logger.Errorw("error in performing deploy stage", "deployPayload", deployPayload, "err", err)
		}
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		deployPayload := &bean.BulkDeployPayload{}
		err := json.Unmarshal([]byte(msg.Data), &deployPayload)
		if err != nil {
			return "error while unmarshalling deployPayload json object", []interface{}{"error", err}
		}
		return "got message for deploy app-store apps in bulk", []interface{}{"installedAppVersionId", deployPayload.InstalledAppVersionId, "installedAppVersionHistoryId", deployPayload.InstalledAppVersionHistoryId}
	}

	err := impl.pubSubClient.Subscribe(pubsub.BULK_APPSTORE_DEPLOY_TOPIC, callback, loggerFunc)
	if err != nil {
		impl.logger.Error("err", err)
		return err
	}
	return nil
}

func (impl *AppStoreAppsEventProcessorImpl) SubscribeHelmInstallStatusEvent() error {

	callback := func(msg *model.PubSubMsg) {

		helmInstallNatsMessage := &appStoreBean.HelmReleaseStatusConfig{}
		err := json.Unmarshal([]byte(msg.Data), helmInstallNatsMessage)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling helm install status nats message", "err", err)
			return
		}

		installedAppVersionHistory, err := impl.iavHistoryRepository.GetInstalledAppVersionHistory(helmInstallNatsMessage.InstallAppVersionHistoryId)
		if err != nil {
			impl.logger.Errorw("error in fetching installed app by installed app id in subscribe helm status callback", "err", err)
			return
		}
		if helmInstallNatsMessage.ErrorInInstallation {
			installedAppVersionHistory.Status = pipelineConfig.WorkflowFailed
		} else {
			installedAppVersionHistory.Status = pipelineConfig.WorkflowSucceeded
		}
		installedAppVersionHistory.HelmReleaseStatusConfig = msg.Data
		_, err = impl.iavHistoryRepository.UpdateInstalledAppVersionHistory(installedAppVersionHistory, nil)
		if err != nil {
			impl.logger.Errorw("error in updating helm release status data in installedAppVersionHistoryRepository", "err", err)
			return
		}
	}
	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		helmInstallNatsMessage := &appStoreBean.HelmReleaseStatusConfig{}
		err := json.Unmarshal([]byte(msg.Data), helmInstallNatsMessage)
		if err != nil {
			return "error in unmarshalling helm install status nats message", []interface{}{"err", err}
		}
		return "got nats msg for helm chart install status", []interface{}{"InstallAppVersionHistoryId", helmInstallNatsMessage.InstallAppVersionHistoryId, "ErrorInInstallation", helmInstallNatsMessage.ErrorInInstallation, "IsReleaseInstalled", helmInstallNatsMessage.IsReleaseInstalled}
	}

	err := impl.pubSubClient.Subscribe(pubsub.HELM_CHART_INSTALL_STATUS_TOPIC, callback, loggerFunc)
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}
