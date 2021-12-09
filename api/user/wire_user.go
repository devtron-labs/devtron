package user

import (
	user2 "github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	repository4 "github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/google/wire"
)

//depends on sql,casbin,validate,logger

var UserWireSet = wire.NewSet(
	NewUserAuthRouterImpl,
	wire.Bind(new(UserAuthRouter), new(*UserAuthRouterImpl)),
	NewUserAuthHandlerImpl,
	wire.Bind(new(UserAuthHandler), new(*UserAuthHandlerImpl)),
	user2.NewUserAuthServiceImpl,
	wire.Bind(new(user2.UserAuthService), new(*user2.UserAuthServiceImpl)),
	repository4.NewUserAuthRepositoryImpl,
	wire.Bind(new(repository4.UserAuthRepository), new(*repository4.UserAuthRepositoryImpl)),

	NewUserRouterImpl,
	wire.Bind(new(UserRouter), new(*UserRouterImpl)),
	NewUserRestHandlerImpl,
	wire.Bind(new(UserRestHandler), new(*UserRestHandlerImpl)),
	user2.NewUserServiceImpl,
	wire.Bind(new(user2.UserService), new(*user2.UserServiceImpl)),
	repository4.NewUserRepositoryImpl,
	wire.Bind(new(repository4.UserRepository), new(*repository4.UserRepositoryImpl)),
	user2.NewRoleGroupServiceImpl,
	wire.Bind(new(user2.RoleGroupService), new(*user2.RoleGroupServiceImpl)),
	repository4.NewRoleGroupRepositoryImpl,
	wire.Bind(new(repository4.RoleGroupRepository), new(*repository4.RoleGroupRepositoryImpl)),


	casbin.NewEnforcerImpl,
	wire.Bind(new(casbin.Enforcer), new(*casbin.EnforcerImpl)),
	casbin.Create,

)
