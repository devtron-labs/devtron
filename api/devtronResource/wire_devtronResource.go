package devtronResource

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/history/deployment/cdPipeline"
	"github.com/devtron-labs/devtron/pkg/devtronResource/read"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/google/wire"
)

var DevtronResourceWireSet = wire.NewSet(
	//old bindings, migrated from wire.go
	read.NewDevtronResourceSearchableKeyServiceImpl,
	wire.Bind(new(read.DevtronResourceSearchableKeyService), new(*read.DevtronResourceSearchableKeyServiceImpl)),
	repository.NewDevtronResourceSearchableKeyRepositoryImpl,
	wire.Bind(new(repository.DevtronResourceSearchableKeyRepository), new(*repository.DevtronResourceSearchableKeyRepositoryImpl)),

	NewDevtronResourceRouterImpl,
	wire.Bind(new(DevtronResourceRouter), new(*DevtronResourceRouterImpl)),

	NewHistoryRouterImpl,
	wire.Bind(new(HistoryRouter), new(*HistoryRouterImpl)),
	NewHistoryRestHandlerImpl,
	wire.Bind(new(HistoryRestHandler), new(*HistoryRestHandlerImpl)),
	cdPipeline.NewDeploymentHistoryServiceImpl,
	wire.Bind(new(cdPipeline.DeploymentHistoryService), new(*cdPipeline.DeploymentHistoryServiceImpl)),

	devtronResource.NewAPIReqDecoderServiceImpl,
	wire.Bind(new(devtronResource.APIReqDecoderService), new(*devtronResource.APIReqDecoderServiceImpl)),
)
