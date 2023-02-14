package terminal

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/terminal"
	"github.com/devtron-labs/devtron/pkg/clusterTerminalAccess"
	"github.com/google/wire"
)

var terminalWireBaseSet = wire.NewSet(
	NewUserTerminalAccessRouterImpl,
	wire.Bind(new(UserTerminalAccessRouter), new(*UserTerminalAccessRouterImpl)),
	NewUserTerminalAccessRestHandlerImpl,
	wire.Bind(new(UserTerminalAccessRestHandler), new(*UserTerminalAccessRestHandlerImpl)),
	clusterTerminalAccess.GetTerminalAccessConfig,
	clusterTerminalAccess.NewUserTerminalAccessServiceImpl,
	wire.Bind(new(clusterTerminalAccess.UserTerminalAccessService), new(*clusterTerminalAccess.UserTerminalAccessServiceImpl)),
)

var TerminalWireSet = wire.NewSet(
	terminalWireBaseSet,
	terminal.NewTerminalAccessRepositoryImpl,
	wire.Bind(new(terminal.TerminalAccessRepository), new(*terminal.TerminalAccessRepositoryImpl)),
)

var TerminalWireSetK8sClient = wire.NewSet(
	terminalWireBaseSet,
	terminal.NewTerminalAccessFileBasedRepository,
	wire.Bind(new(terminal.TerminalAccessRepository), new(*terminal.TerminalAccessFileBasedRepository)),
)
