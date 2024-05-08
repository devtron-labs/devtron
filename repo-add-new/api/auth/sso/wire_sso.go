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
