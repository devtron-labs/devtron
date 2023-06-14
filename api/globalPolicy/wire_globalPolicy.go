package globalPolicy

import (
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/history"
	repository2 "github.com/devtron-labs/devtron/pkg/globalPolicy/history/repository"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/repository"
	"github.com/google/wire"
)

var GlobalPolicyWireSet = wire.NewSet(
	repository.NewGlobalPolicySearchableFieldRepositoryImpl,
	wire.Bind(new(repository.GlobalPolicySearchableFieldRepository), new(*repository.GlobalPolicySearchableFieldRepositoryImpl)),
	repository.NewGlobalPolicyRepositoryImpl,
	wire.Bind(new(repository.GlobalPolicyRepository), new(*repository.GlobalPolicyRepositoryImpl)),
	repository2.NewGlobalPolicyHistoryRepositoryImpl,
	wire.Bind(new(repository2.GlobalPolicyHistoryRepository), new(*repository2.GlobalPolicyHistoryRepositoryImpl)),

	globalPolicy.NewGlobalPolicyServiceImpl,
	wire.Bind(new(globalPolicy.GlobalPolicyService), new(*globalPolicy.GlobalPolicyServiceImpl)),
	history.NewGlobalPolicyHistoryServiceImpl,
	wire.Bind(new(history.GlobalPolicyHistoryService), new(*history.GlobalPolicyHistoryServiceImpl)),

	NewGlobalPolicyRestHandlerImpl,
	wire.Bind(new(GlobalPolicyRestHandler), new(*GlobalPolicyRestHandlerImpl)),

	NewGlobalPolicyRouterImpl,
	wire.Bind(new(GlobalPolicyRouter), new(*GlobalPolicyRouterImpl)),
)
