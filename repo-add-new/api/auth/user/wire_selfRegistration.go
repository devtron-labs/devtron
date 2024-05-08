package user

import (
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/google/wire"
)

//depends on sql,
//TODO integrate user auth module

var SelfRegistrationWireSet = wire.NewSet(
	repository.NewSelfRegistrationRolesRepositoryImpl,
	wire.Bind(new(repository.SelfRegistrationRolesRepository), new(*repository.SelfRegistrationRolesRepositoryImpl)),

	user.NewUserSelfRegistrationServiceImpl,
	wire.Bind(new(user.UserSelfRegistrationService), new(*user.UserSelfRegistrationServiceImpl)),
)
