package user

import (
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/google/wire"
)

//depends on sql,
//TODO integrate user auth module

var SelfRegistrationWireSet = wire.NewSet(
	repository.NewSelfRegistrationRolesRepositoryImpl,
	wire.Bind(new(repository.SelfRegistrationRolesRepository), new(*repository.SelfRegistrationRolesRepositoryImpl)),

	user.NewSelfRegistrationRolesServiceImpl,
	wire.Bind(new(user.SelfRegistrationRolesService), new(*user.SelfRegistrationRolesServiceImpl)),
)
