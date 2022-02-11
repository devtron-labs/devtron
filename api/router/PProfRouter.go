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

func (ppr PProfRouterImpl) initPProfRouter(router *mux.Router) {

	router.HandleFunc("/", ppr.pprofRestHandler.Index)
	router.HandleFunc("/cmdline", ppr.pprofRestHandler.Cmdline)
	router.HandleFunc("/profile", ppr.pprofRestHandler.Profile)
	router.HandleFunc("/symbol", ppr.pprofRestHandler.Symbol)
	router.HandleFunc("/trace", ppr.pprofRestHandler.Trace)
	router.HandleFunc("/goroutine", ppr.pprofRestHandler.Goroutine)
	router.HandleFunc("/threadcreate", ppr.pprofRestHandler.Threadcreate)
	router.HandleFunc("/heap", ppr.pprofRestHandler.Heap)
	router.HandleFunc("/block", ppr.pprofRestHandler.Block)
	router.HandleFunc("/mutex", ppr.pprofRestHandler.Mutex)
	router.HandleFunc("/allocs", ppr.pprofRestHandler.Allocs)
}

