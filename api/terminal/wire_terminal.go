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
