package user

import (
	"github.com/devtron-labs/devtron/pkg/auth/authentication"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	user2 "github.com/devtron-labs/devtron/pkg/auth/user"
	repository2 "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	helper "github.com/devtron-labs/devtron/pkg/auth/user/repository/helper"
	"github.com/google/wire"
)

//depends on sql,validate,logger

var UserWireSet = wire.NewSet(
	UserAuditWireSet,

	helper.NewUserRepositoryQueryBuilder,
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
	repository2.NewUserGroupMapRepositoryImpl,
	wire.Bind(new(repository2.UserGroupMapRepository), new(*repository2.UserGroupMapRepositoryImpl)),

	NewUserRouterImpl,
	wire.Bind(new(UserRouter), new(*UserRouterImpl)),
	NewUserRestHandlerImpl,
	wire.Bind(new(UserRestHandler), new(*UserRestHandlerImpl)),
	user2.NewCleanUpPoliciesServiceImpl,
	wire.Bind(new(user2.CleanUpPoliciesService), new(*user2.CleanUpPoliciesServiceImpl)),
	repository2.NewPoliciesCleanUpRepositoryImpl,
	wire.Bind(new(repository2.PoliciesCleanUpRepository), new(*repository2.PoliciesCleanUpRepositoryImpl)),
	user2.NewUserServiceImpl,
	wire.Bind(new(user2.UserService), new(*user2.UserServiceImpl)),
	repository2.NewUserRepositoryImpl,
	wire.Bind(new(repository2.UserRepository), new(*repository2.UserRepositoryImpl)),
	user2.NewRoleGroupServiceImpl,
	wire.Bind(new(user2.RoleGroupService), new(*user2.RoleGroupServiceImpl)),
	repository2.NewRoleGroupRepositoryImpl,
	wire.Bind(new(repository2.RoleGroupRepository), new(*repository2.RoleGroupRepositoryImpl)),

	//casbin.NewEnforcerImpl,
	casbin.NewEnterpriseEnforcerImpl,
	wire.Bind(new(casbin.Enforcer), new(*casbin.EnterpriseEnforcerImpl)),
	casbin.Create, casbin.CreateV2,

	user2.NewUserCommonServiceImpl,
	wire.Bind(new(user2.UserCommonService), new(*user2.UserCommonServiceImpl)),

	authentication.NewUserAuthOidcHelperImpl,
	wire.Bind(new(authentication.UserAuthOidcHelper), new(*authentication.UserAuthOidcHelperImpl)),

	repository2.NewRbacPolicyResourceDetailRepositoryImpl,
	wire.Bind(new(repository2.RbacPolicyResourceDetailRepository), new(*repository2.RbacPolicyResourceDetailRepositoryImpl)),
	repository2.NewRbacRoleResourceDetailRepositoryImpl,
	wire.Bind(new(repository2.RbacRoleResourceDetailRepository), new(*repository2.RbacRoleResourceDetailRepositoryImpl)),
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

	user2.NewDefaultRbacRoleServiceImpl,
	wire.Bind(new(user2.DefaultRbacRoleService), new(*user2.DefaultRbacRoleServiceImpl)),
	repository2.NewDefaultRbacRoleDataRepositoryImpl,
	wire.Bind(new(repository2.DefaultRbacRoleDataRepository), new(*repository2.DefaultRbacRoleDataRepositoryImpl)),
)
