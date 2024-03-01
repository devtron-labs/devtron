package in

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/devtron/pkg/appStore/chartGroup"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"go.uber.org/zap"
)

type AppStoreAppsEventProcessorImpl struct {
	logger            *zap.SugaredLogger
	pubSubClient      *pubsub.PubSubClientServiceImpl
	chartGroupService chartGroup.ChartGroupService
}

func NewAppStoreAppsEventProcessorImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	chartGroupService chartGroup.ChartGroupService) *AppStoreAppsEventProcessorImpl {
	return &AppStoreAppsEventProcessorImpl{
		logger:            logger,
		pubSubClient:      pubSubClient,
		chartGroupService: chartGroupService,
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
