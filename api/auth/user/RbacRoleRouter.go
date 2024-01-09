package user

import (
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
	validator *validator.Validate, rbacRoleRestHandler RbacRoleRestHandler) *RbacRoleRouterImpl {
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

	rbacRoleRouter.Path("").
		HandlerFunc(router.rbacRoleRestHandler.CreateDefaultRole).Methods("POST")

	rbacRoleRouter.Path("").
		HandlerFunc(router.rbacRoleRestHandler.UpdateDefaultRole).Methods("PUT")

	rbacRoleRouter.Path("/sync").
		HandlerFunc(router.rbacRoleRestHandler.SyncDefaultRoles).Methods("POST")

	rbacRoleRouter.Path("/{id}").
		HandlerFunc(router.rbacRoleRestHandler.GetDefaultRoleDetailById).Methods("GET")

	rbacRoleRouter.Path("/entity/{entity}").
		HandlerFunc(router.rbacRoleRestHandler.GetAllDefaultRolesByEntityAccessType).Methods("GET")

	rbacRoleRouter.Path("/policy/resource/list").
		HandlerFunc(router.rbacRoleRestHandler.GetRbacPolicyResourceListForAllEntityAccessTypes).Methods("GET")

	rbacRoleRouter.Path("/policy/resource/{entity}").
		HandlerFunc(router.rbacRoleRestHandler.GetRbacPolicyResourceListByEntityAndAccessType).Methods("GET")
}
