package user

import (
	"github.com/devtron-labs/devtron/pkg/auth"
	casbin2 "github.com/devtron-labs/devtron/pkg/enterprise/user/casbin"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/google/wire"
)

//depends on sql,validate,logger

var UserWireSet = wire.NewSet(
	UserAuditWireSet,

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
	user.NewCleanUpPoliciesServiceImpl,
	wire.Bind(new(user.CleanUpPoliciesService), new(*user.CleanUpPoliciesServiceImpl)),
	repository.NewPoliciesCleanUpRepositoryImpl,
	wire.Bind(new(repository.PoliciesCleanUpRepository), new(*repository.PoliciesCleanUpRepositoryImpl)),
	user.NewUserServiceImpl,
	wire.Bind(new(user.UserService), new(*user.UserServiceImpl)),
	repository.NewUserRepositoryImpl,
	wire.Bind(new(repository.UserRepository), new(*repository.UserRepositoryImpl)),
	user.NewRoleGroupServiceImpl,
	wire.Bind(new(user.RoleGroupService), new(*user.RoleGroupServiceImpl)),
	repository.NewRoleGroupRepositoryImpl,
	wire.Bind(new(repository.RoleGroupRepository), new(*repository.RoleGroupRepositoryImpl)),

	//casbin.NewEnforcerImpl,
	casbin2.NewEnterpriseEnforcerImpl,
	wire.Bind(new(casbin.Enforcer), new(*casbin2.EnterpriseEnforcerImpl)),
	casbin.Create,

	user.NewUserCommonServiceImpl,
	wire.Bind(new(user.UserCommonService), new(*user.UserCommonServiceImpl)),

	auth.NewUserAuthOidcHelperImpl,
	wire.Bind(new(auth.UserAuthOidcHelper), new(*auth.UserAuthOidcHelperImpl)),

	repository.NewRbacPolicyResourceDetailRepositoryImpl,
	wire.Bind(new(repository.RbacPolicyResourceDetailRepository), new(*repository.RbacPolicyResourceDetailRepositoryImpl)),
	repository.NewRbacRoleResourceDetailRepositoryImpl,
	wire.Bind(new(repository.RbacRoleResourceDetailRepository), new(*repository.RbacRoleResourceDetailRepositoryImpl)),
	repository.NewRbacPolicyDataRepositoryImpl,
	wire.Bind(new(repository.RbacPolicyDataRepository), new(*repository.RbacPolicyDataRepositoryImpl)),
	repository.NewRbacRoleDataRepositoryImpl,
	wire.Bind(new(repository.RbacRoleDataRepository), new(*repository.RbacRoleDataRepositoryImpl)),
	repository.NewRbacDataCacheFactoryImpl,
	wire.Bind(new(repository.RbacDataCacheFactory), new(*repository.RbacDataCacheFactoryImpl)),

	NewRbacRoleRouterImpl,
	wire.Bind(new(RbacRoleRouter), new(*RbacRoleRouterImpl)),
	NewRbacRoleHandlerImpl,
	wire.Bind(new(RbacRoleRestHandler), new(*RbacRoleRestHandlerImpl)),
	user.NewRbacRoleServiceImpl,
	wire.Bind(new(user.RbacRoleService), new(*user.RbacRoleServiceImpl)),
)
