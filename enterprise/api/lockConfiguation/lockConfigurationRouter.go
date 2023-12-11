package lockConfiguation

import "github.com/gorilla/mux"

type LockConfigurationRouter interface {
	InitLockConfigurationRouter(lockConfiguration *mux.Router)
}

type LockConfigurationRouterImpl struct {
	lockConfigRestHandler LockConfigRestHandler
}

func NewLockConfigurationRouterImpl(lockConfigRestHandler LockConfigRestHandler) *LockConfigurationRouterImpl {
	return &LockConfigurationRouterImpl{lockConfigRestHandler: lockConfigRestHandler}
}

func (router *LockConfigurationRouterImpl) InitLockConfigurationRouter(lockConfiguration *mux.Router) {
	lockConfiguration.Path("").HandlerFunc(router.lockConfigRestHandler.CreateLockConfig).
		Methods("POST")
	lockConfiguration.Path("").HandlerFunc(router.lockConfigRestHandler.GetLockConfig).
		Methods("GET")
	lockConfiguration.Path("").HandlerFunc(router.lockConfigRestHandler.DeleteLockConfig).
		Methods("DELETE")
}
