package drafts

import (
	"github.com/devtron-labs/devtron/enterprise/pkg/drafts"
	"github.com/google/wire"
)

var DraftsWireSet = wire.NewSet(
	drafts.NewConfigDraftRepositoryImpl,
	wire.Bind(new(drafts.ConfigDraftRepository), new(*drafts.ConfigDraftRepositoryImpl)),
	drafts.NewConfigDraftServiceImpl,
	wire.Bind(new(drafts.ConfigDraftService), new(*drafts.ConfigDraftServiceImpl)),
	NewConfigDraftRestHandlerImpl,
	wire.Bind(new(ConfigDraftRestHandler), new(*ConfigDraftRestHandlerImpl)),
	NewConfigDraftRouterImpl,
	wire.Bind(new(ConfigDraftRouter), new(*ConfigDraftRouterImpl)),
)
