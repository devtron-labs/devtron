package chartProvider

import (
	chartProviderService "github.com/devtron-labs/devtron/pkg/appStore/chartProvider"
	"github.com/google/wire"
)

var AppStoreChartProviderWireSet = wire.NewSet(
	chartProviderService.NewChartProviderServiceImpl,
	wire.Bind(new(chartProviderService.ChartProviderService), new(*chartProviderService.ChartProviderServiceImpl)),
	NewChartProviderRestHandlerImpl,
	wire.Bind(new(ChartProviderRestHandler), new(*ChartProviderRestHandlerImpl)),
	NewChartProviderRouterImpl,
	wire.Bind(new(ChartProviderRouter), new(*ChartProviderRouterImpl)))
