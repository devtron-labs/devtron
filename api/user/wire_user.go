package user

import (
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/google/wire"
)

//depends on sql,validate,logger

var UserWireSet = wire.NewSet(
	NewUserAuthRouterImpl,
	wire.Bind(new(UserAuthRouter), new(*UserAuthRouterImpl)),
	NewUserAuthHandlerImpl,
	wire.Bind(new(UserAuthHandler), new(*UserAuthHandlerImpl)),
	user.NewUserAuthServiceImpl,
	wire.Bind(new(user.UserAuthService), new(*user.UserAuthServiceImpl)),
	repository.NewUserAuthRepositoryImpl,
	wire.Bind(new(repository.UserAuthRepository), new(*repository.UserAuthRepositoryImpl)),
	repository.NewDefaultAuthPolicyRepositoryImpl,
	wire.Bind(new(repository.DefaultAuthPolicyRepository), new(*repository.DefaultAuthPolicyRepositoryImpl)),
	repository.NewDefaultAuthRoleRepositoryImpl,
	wire.Bind(new(repository.DefaultAuthRoleRepository), new(*repository.DefaultAuthRoleRepositoryImpl)),

	NewUserRouterImpl,
	wire.Bind(new(UserRouter), new(*UserRouterImpl)),
	NewUserRestHandlerImpl,
	wire.Bind(new(UserRestHandler), new(*UserRestHandlerImpl)),
	user.NewUserServiceImpl,
	wire.Bind(new(user.UserService), new(*user.UserServiceImpl)),
	repository.NewUserRepositoryImpl,
	wire.Bind(new(repository.UserRepository), new(*repository.UserRepositoryImpl)),
	user.NewRoleGroupServiceImpl,
	wire.Bind(new(user.RoleGroupService), new(*user.RoleGroupServiceImpl)),
	repository.NewRoleGroupRepositoryImpl,
	wire.Bind(new(repository.RoleGroupRepository), new(*repository.RoleGroupRepositoryImpl)),

	casbin.NewEnforcerImpl,
	wire.Bind(new(casbin.Enforcer), new(*casbin.EnforcerImpl)),
	casbin.Create,

	user.NewUserCommonServiceImpl,
	wire.Bind(new(user.UserCommonService), new(*user.UserCommonServiceImpl)),
)
