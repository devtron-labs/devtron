/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package team

import (
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/team/read"
	"github.com/devtron-labs/devtron/pkg/team/repository"
	"github.com/google/wire"
)

//depends on sql,
//TODO integrate user auth module

var TeamsWireSet = wire.NewSet(
	repository.NewTeamRepositoryImpl,
	wire.Bind(new(repository.TeamRepository), new(*repository.TeamRepositoryImpl)),
	team.NewTeamServiceImpl,
	wire.Bind(new(team.TeamService), new(*team.TeamServiceImpl)),
	NewTeamRestHandlerImpl,
	wire.Bind(new(TeamRestHandler), new(*TeamRestHandlerImpl)),
	NewTeamRouterImpl,
	wire.Bind(new(TeamRouter), new(*TeamRouterImpl)),
	read.NewTeamReadService,
	wire.Bind(new(read.TeamReadService), new(*read.TeamReadServiceImpl)),
)
