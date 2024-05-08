package client

import (
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/google/wire"
)

var HelmAppWireSet = wire.NewSet(
	gRPC.NewHelmAppClientImpl,
	wire.Bind(new(gRPC.HelmAppClient), new(*gRPC.HelmAppClientImpl)),
	service.GetHelmReleaseConfig,
	service.NewHelmAppServiceImpl,
	wire.Bind(new(service.HelmAppService), new(*service.HelmAppServiceImpl)),
	NewHelmAppRestHandlerImpl,
	wire.Bind(new(HelmAppRestHandler), new(*HelmAppRestHandlerImpl)),
	NewHelmAppRouterImpl,
	wire.Bind(new(HelmAppRouter), new(*HelmAppRouterImpl)),
	gRPC.GetConfig,
	rbac.NewEnforcerUtilHelmImpl,
	wire.Bind(new(rbac.EnforcerUtilHelm), new(*rbac.EnforcerUtilHelmImpl)),
)
