package user

import (
	"github.com/devtron-labs/devtron/api/logger"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type RbacRoleRouter interface {
	InitRbacRoleRouter(rbacRoleRouter *mux.Router)
}

type RbacRoleRouterImpl struct {
	logger              *zap.SugaredLogger
	validator           *validator.Validate
	rbacRoleRestHandler RbacRoleRestHandler
}

func NewRbacRoleRouterImpl(logger *zap.SugaredLogger,
	validator *validator.Validate, rbacRoleRestHandler RbacRoleRestHandler, userAuth logger.UserAuth) *RbacRoleRouterImpl {
	rbacRoleRouterImpl := &RbacRoleRouterImpl{
		logger:              logger,
		validator:           validator,
		rbacRoleRestHandler: rbacRoleRestHandler,
	}
	return rbacRoleRouterImpl
}

func (router RbacRoleRouterImpl) InitRbacRoleRouter(rbacRoleRouter *mux.Router) {
	rbacRoleRouter.Path("").
		HandlerFunc(router.rbacRoleRestHandler.GetAllDefaultRoles).Methods("GET")
}
