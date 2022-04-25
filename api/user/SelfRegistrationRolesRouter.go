package user

import (
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type SelfRegistrationRolesRouter interface {
	InitSelfRegistrationRolesRouter(router *mux.Router)
}

type SelfRegistrationRolesRouterImpl struct {
	logger                       *zap.SugaredLogger
	selfRegistrationRolesHandler SelfRegistrationRolesHandler
}

func NewSelfRegistrationRolesRouterImpl(logger *zap.SugaredLogger, selfRegistrationRolesHandler SelfRegistrationRolesHandler) *SelfRegistrationRolesRouterImpl {
	return &SelfRegistrationRolesRouterImpl{
		logger:                       logger,
		selfRegistrationRolesHandler: selfRegistrationRolesHandler,
	}
}

func (impl *SelfRegistrationRolesRouterImpl) InitSelfRegistrationRolesRouter(router *mux.Router) {
	router.Path("/register").
		HandlerFunc(impl.selfRegistrationRolesHandler.SelfRegister).Methods("POST")
}
