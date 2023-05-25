package team

import (
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/google/wire"
)

//depends on sql,
//TODO integrate user auth module

var TeamsWireSet = wire.NewSet(
	team.NewTeamRepositoryImpl,
	wire.Bind(new(team.TeamRepository), new(*team.TeamRepositoryImpl)),
	team.NewTeamServiceImpl,
	wire.Bind(new(team.TeamService), new(*team.TeamServiceImpl)),
	NewTeamRestHandlerImpl,
	wire.Bind(new(TeamRestHandler), new(*TeamRestHandlerImpl)),
	NewTeamRouterImpl,
	wire.Bind(new(TeamRouter), new(*TeamRouterImpl)),
)
