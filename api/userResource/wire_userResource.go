package userResource

import (
	"github.com/devtron-labs/devtron/pkg/userResource"
	"github.com/google/wire"
)

var UserResourceWireSet = wire.NewSet(
	NewUserResourceRouterImpl,
	wire.Bind(new(Router), new(*RouterImpl)),

	NewUserResourceRestHandler,
	wire.Bind(new(RestHandler), new(*RestHandlerImpl)),

	userResource.NewUserResourceExtendedServiceImpl,
	wire.Bind(new(userResource.UserResourceService), new(*userResource.UserResourceExtendedServiceImpl)),
)

var UserResourceWireSetEA = wire.NewSet(
	NewUserResourceRouterImpl,
	wire.Bind(new(Router), new(*RouterImpl)),

	NewUserResourceRestHandler,
	wire.Bind(new(RestHandler), new(*RestHandlerImpl)),

	userResource.NewUserResourceServiceImpl,
	wire.Bind(new(userResource.UserResourceService), new(*userResource.UserResourceServiceImpl)),
)
