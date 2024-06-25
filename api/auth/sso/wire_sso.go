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

package sso

import (
	sso2 "github.com/devtron-labs/devtron/pkg/auth/sso"
	"github.com/google/wire"
)

//depends on sql,user,K8sUtil, logger, enforcer,

var SsoConfigWireSet = wire.NewSet(
	sso2.NewSSOLoginServiceImpl,
	wire.Bind(new(sso2.SSOLoginService), new(*sso2.SSOLoginServiceImpl)),
	sso2.NewSSOLoginRepositoryImpl,
	wire.Bind(new(sso2.SSOLoginRepository), new(*sso2.SSOLoginRepositoryImpl)),

	NewSsoLoginRouterImpl,
	wire.Bind(new(SsoLoginRouter), new(*SsoLoginRouterImpl)),
	NewSsoLoginRestHandlerImpl,
	wire.Bind(new(SsoLoginRestHandler), new(*SsoLoginRestHandlerImpl)),
)
