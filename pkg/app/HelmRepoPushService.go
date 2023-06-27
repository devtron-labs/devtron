package app

import (
	"context"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app/bean"
	"go.uber.org/zap"
)

type HelmRepoPushService interface {
	ManifestPushService
}

type HelmRepoPushServiceImpl struct {
	logger               *zap.SugaredLogger
	chartTemplateService util.ChartTemplateService
}

func NewHelmRepoPushServiceImpl(
	logger *zap.SugaredLogger,
	chartTemplateService util.ChartTemplateService,
) *HelmRepoPushServiceImpl {
	return &HelmRepoPushServiceImpl{
		logger:               logger,
		chartTemplateService: chartTemplateService,
	}
}

func (impl *HelmRepoPushServiceImpl) PushChart(manifestPushConfig *bean.ManifestPushTemplate, ctx context.Context) bean.ManifestPushResponse {
	return bean.ManifestPushResponse{}
}
