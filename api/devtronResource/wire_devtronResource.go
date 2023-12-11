package devtronResource

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/google/wire"
)

var DevtronResourceWireSet = wire.NewSet(
	//old bindings, migrated from wire.go
	devtronResource.NewDevtronResourceSearchableKeyServiceImpl,
	wire.Bind(new(devtronResource.DevtronResourceSearchableKeyService), new(*devtronResource.DevtronResourceSearchableKeyServiceImpl)),
	repository.NewDevtronResourceSearchableKeyRepositoryImpl,
	wire.Bind(new(repository.DevtronResourceSearchableKeyRepository), new(*repository.DevtronResourceSearchableKeyRepositoryImpl)),

	NewDevtronResourceRouterImpl,
	wire.Bind(new(DevtronResourceRouter), new(*DevtronResourceRouterImpl)),
	NewDevtronResourceRestHandlerImpl,
	wire.Bind(new(DevtronResourceRestHandler), new(*DevtronResourceRestHandlerImpl)),
	devtronResource.NewDevtronResourceServiceImpl,
	wire.Bind(new(devtronResource.DevtronResourceService), new(*devtronResource.DevtronResourceServiceImpl)),
	repository.NewDevtronResourceRepositoryImpl,
	wire.Bind(new(repository.DevtronResourceRepository), new(*repository.DevtronResourceRepositoryImpl)),
	repository.NewDevtronResourceSchemaRepositoryImpl,
	wire.Bind(new(repository.DevtronResourceSchemaRepository), new(*repository.DevtronResourceSchemaRepositoryImpl)),
	repository.NewDevtronResourceObjectRepositoryImpl,
	wire.Bind(new(repository.DevtronResourceObjectRepository), new(*repository.DevtronResourceObjectRepositoryImpl)),
	repository.NewDevtronResourceObjectAuditRepositoryImpl,
	wire.Bind(new(repository.DevtronResourceObjectAuditRepository), new(*repository.DevtronResourceObjectAuditRepositoryImpl)),
)

var DevtronResourceWireSetEA = wire.NewSet(
	devtronResource.NewDevtronResourceServiceImpl,
	wire.Bind(new(devtronResource.DevtronResourceService), new(*devtronResource.DevtronResourceServiceImpl)),
	repository.NewDevtronResourceRepositoryImpl,
	wire.Bind(new(repository.DevtronResourceRepository), new(*repository.DevtronResourceRepositoryImpl)),
	repository.NewDevtronResourceSchemaRepositoryImpl,
	wire.Bind(new(repository.DevtronResourceSchemaRepository), new(*repository.DevtronResourceSchemaRepositoryImpl)),
	repository.NewDevtronResourceObjectRepositoryImpl,
	wire.Bind(new(repository.DevtronResourceObjectRepository), new(*repository.DevtronResourceObjectRepositoryImpl)),
	repository.NewDevtronResourceObjectAuditRepositoryImpl,
	wire.Bind(new(repository.DevtronResourceObjectAuditRepository), new(*repository.DevtronResourceObjectAuditRepositoryImpl)),
)
