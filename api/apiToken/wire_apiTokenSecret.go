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
	apiTokenAuth "github.com/devtron-labs/authenticator/apiToken"
	"github.com/devtron-labs/devtron/pkg/apiToken"
	"github.com/google/wire"
)

var ApiTokenSecretWireSet = wire.NewSet(
	apiTokenAuth.InitApiTokenSecretStore,
	apiToken.NewApiTokenSecretServiceImpl,
	wire.Bind(new(apiToken.ApiTokenSecretService), new(*apiToken.ApiTokenSecretServiceImpl)),
)
