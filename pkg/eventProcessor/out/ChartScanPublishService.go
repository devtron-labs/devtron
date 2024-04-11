package out

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"go.uber.org/zap"
)

type ChartScanPublishService interface {
	PublishChartScanEvent(appVersionDto *appStoreBean.InstallAppVersionDTO) error
}

type ChartScanPublishServiceImpl struct {
	logger       *zap.SugaredLogger
	pubSubClient *pubsub.PubSubClientServiceImpl
}

func NewChartScanPublishServiceImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl) *ChartScanPublishServiceImpl {
	return &ChartScanPublishServiceImpl{
		logger:       logger,
		pubSubClient: pubSubClient,
	}
}

func (impl ChartScanPublishServiceImpl) PublishChartScanEvent(appVersionDto *appStoreBean.InstallAppVersionDTO) error {

	data, err := json.Marshal(appVersionDto)
	if err != nil {
		return err
	} else {
		err = impl.pubSubClient.Publish(pubsub.CHART_SCAN_TOPIC, string(data))
		if err != nil {
			impl.logger.Errorw("err while publishing msg for PublishChartScanEvent", "msg", data, "err", err)
			return err
		}
	}
	return nil
}
