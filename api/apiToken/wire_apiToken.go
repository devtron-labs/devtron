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

package apiToken

import (
	"github.com/devtron-labs/devtron/pkg/apiToken"
	"github.com/google/wire"
)

var ApiTokenWireSet = wire.NewSet(
	apiToken.NewApiTokenRepositoryImpl,
	wire.Bind(new(apiToken.ApiTokenRepository), new(*apiToken.ApiTokenRepositoryImpl)),
	apiToken.NewApiTokenServiceImpl,
	wire.Bind(new(apiToken.ApiTokenService), new(*apiToken.ApiTokenServiceImpl)),
	NewApiTokenRestHandlerImpl,
	wire.Bind(new(ApiTokenRestHandler), new(*ApiTokenRestHandlerImpl)),
	NewApiTokenRouterImpl,
	wire.Bind(new(ApiTokenRouter), new(*ApiTokenRouterImpl)),
)
