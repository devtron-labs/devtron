package out

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"go.uber.org/zap"
	"os"
)

type ChartScanPublishService interface {
	PublishChartScanEvent(chartScanEventBean bean.ChartScanEventBean) error
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

func NewChartScanPublishServiceImplEA() *ChartScanPublishServiceImpl {
	return nil
}

func (impl ChartScanPublishServiceImpl) PublishChartScanEvent(chartScanEventBean bean.ChartScanEventBean) error {

	isV2Enabled := os.Getenv("SCAN_V2_ENABLED")
	if isV2Enabled != "true" {
		return nil
	}
	appVersionDto := chartScanEventBean.AppVersionDto
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
