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

package user

import (
	"github.com/devtron-labs/devtron/pkg/auth/authentication"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	user2 "github.com/devtron-labs/devtron/pkg/auth/user"
	repository2 "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/google/wire"
)

//depends on sql,validate,logger

var UserWireSet = wire.NewSet(
	UserAuditWireSet,
	NewUserAuthRouterImpl,
	wire.Bind(new(UserAuthRouter), new(*UserAuthRouterImpl)),
	NewUserAuthHandlerImpl,
	wire.Bind(new(UserAuthHandler), new(*UserAuthHandlerImpl)),
	user2.NewUserAuthServiceImpl,
	wire.Bind(new(user2.UserAuthService), new(*user2.UserAuthServiceImpl)),
	repository2.NewUserAuthRepositoryImpl,
	wire.Bind(new(repository2.UserAuthRepository), new(*repository2.UserAuthRepositoryImpl)),
	repository2.NewDefaultAuthPolicyRepositoryImpl,
	wire.Bind(new(repository2.DefaultAuthPolicyRepository), new(*repository2.DefaultAuthPolicyRepositoryImpl)),
	repository2.NewDefaultAuthRoleRepositoryImpl,
	wire.Bind(new(repository2.DefaultAuthRoleRepository), new(*repository2.DefaultAuthRoleRepositoryImpl)),

	NewUserRouterImpl,
	wire.Bind(new(UserRouter), new(*UserRouterImpl)),
	NewUserRestHandlerImpl,
	wire.Bind(new(UserRestHandler), new(*UserRestHandlerImpl)),
	user2.NewUserServiceImpl,
	wire.Bind(new(user2.UserService), new(*user2.UserServiceImpl)),
	repository2.NewUserRepositoryImpl,
	wire.Bind(new(repository2.UserRepository), new(*repository2.UserRepositoryImpl)),
	user2.NewRoleGroupServiceImpl,
	wire.Bind(new(user2.RoleGroupService), new(*user2.RoleGroupServiceImpl)),
	repository2.NewRoleGroupRepositoryImpl,
	wire.Bind(new(repository2.RoleGroupRepository), new(*repository2.RoleGroupRepositoryImpl)),

	casbin.NewEnforcerImpl,
	wire.Bind(new(casbin.Enforcer), new(*casbin.EnforcerImpl)),
	casbin.Create,

	user2.NewUserCommonServiceImpl,
	wire.Bind(new(user2.UserCommonService), new(*user2.UserCommonServiceImpl)),

	authentication.NewUserAuthOidcHelperImpl,
	wire.Bind(new(authentication.UserAuthOidcHelper), new(*authentication.UserAuthOidcHelperImpl)),

	repository2.NewRbacPolicyDataRepositoryImpl,
	wire.Bind(new(repository2.RbacPolicyDataRepository), new(*repository2.RbacPolicyDataRepositoryImpl)),
	repository2.NewRbacRoleDataRepositoryImpl,
	wire.Bind(new(repository2.RbacRoleDataRepository), new(*repository2.RbacRoleDataRepositoryImpl)),
	repository2.NewRbacDataCacheFactoryImpl,
	wire.Bind(new(repository2.RbacDataCacheFactory), new(*repository2.RbacDataCacheFactoryImpl)),

	NewRbacRoleRouterImpl,
	wire.Bind(new(RbacRoleRouter), new(*RbacRoleRouterImpl)),
	NewRbacRoleHandlerImpl,
	wire.Bind(new(RbacRoleRestHandler), new(*RbacRoleRestHandlerImpl)),
	user2.NewRbacRoleServiceImpl,
	wire.Bind(new(user2.RbacRoleService), new(*user2.RbacRoleServiceImpl)),
)
