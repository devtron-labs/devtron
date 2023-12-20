package sso

import (
	"github.com/devtron-labs/devtron/pkg/sso"
	"github.com/google/wire"
)

//depends on sql,user,K8sUtil, logger, enforcer,

var SsoConfigWireSet = wire.NewSet(
	sso.NewSSOLoginServiceImpl,
	wire.Bind(new(sso.SSOLoginService), new(*sso.SSOLoginServiceImpl)),
	sso.NewSSOLoginRepositoryImpl,
	wire.Bind(new(sso.SSOLoginRepository), new(*sso.SSOLoginRepositoryImpl)),

	NewSsoLoginRouterImpl,
	wire.Bind(new(SsoLoginRouter), new(*SsoLoginRouterImpl)),
	NewSsoLoginRestHandlerImpl,
	wire.Bind(new(SsoLoginRestHandler), new(*SsoLoginRestHandlerImpl)),
)
