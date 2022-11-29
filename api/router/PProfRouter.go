package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type PProfRouter interface {
	initPProfRouter(router *mux.Router)
}

type PProfRouterImpl struct {
	logger           *zap.SugaredLogger
	pprofRestHandler restHandler.PProfRestHandler
}

func NewPProfRouter(logger *zap.SugaredLogger,
	pprofRestHandler restHandler.PProfRestHandler) *PProfRouterImpl {
	return &PProfRouterImpl{
		logger:           logger,
		pprofRestHandler: pprofRestHandler,
	}
}

func (router *PProfRouterImpl) initPProfRouter(pprofRouter *mux.Router) {

	pprofRouter.HandleFunc("/", router.pprofRestHandler.Index)
	pprofRouter.HandleFunc("/cmdline", router.pprofRestHandler.Cmdline)
	pprofRouter.HandleFunc("/profile", router.pprofRestHandler.Profile)
	pprofRouter.HandleFunc("/symbol", router.pprofRestHandler.Symbol)
	pprofRouter.HandleFunc("/trace", router.pprofRestHandler.Trace)
	pprofRouter.HandleFunc("/goroutine", router.pprofRestHandler.Goroutine)
	pprofRouter.HandleFunc("/threadcreate", router.pprofRestHandler.Threadcreate)
	pprofRouter.HandleFunc("/heap", router.pprofRestHandler.Heap)
	pprofRouter.HandleFunc("/block", router.pprofRestHandler.Block)
	pprofRouter.HandleFunc("/mutex", router.pprofRestHandler.Mutex)
	pprofRouter.HandleFunc("/allocs", router.pprofRestHandler.Allocs)
}
