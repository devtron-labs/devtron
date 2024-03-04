package appStore

import (
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service"
	"github.com/google/wire"
)

var AppStoreWireSet = wire.NewSet(

	service.NewDeletePostProcessorImpl,
	wire.Bind(new(service.DeletePostProcessor), new(*service.DeletePostProcessorImpl)),

	service.NewAppAppStoreValidatorImpl,
	wire.Bind(new(service.AppStoreValidator), new(*service.AppStoreValidatorImpl)),
)
