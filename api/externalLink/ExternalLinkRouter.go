package externalLink

import (
	"github.com/gorilla/mux"
)

type ExternalLinkRouter interface {
	InitExternalLinkRouter(gocdRouter *mux.Router)
}
type ExternalLinkRouterImpl struct {
	externalLinkRestHandler ExternalLinkRestHandler
}

func NewExternalLinkRouterImpl(externalLinkRestHandler ExternalLinkRestHandler) *ExternalLinkRouterImpl {
	return &ExternalLinkRouterImpl{externalLinkRestHandler: externalLinkRestHandler}
}

func (router *ExternalLinkRouterImpl) InitExternalLinkRouter(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(router.externalLinkRestHandler.CreateExternalLinks).Methods("POST")
	configRouter.Path("/tools").HandlerFunc(router.externalLinkRestHandler.GetExternalLinkMonitoringTools).Methods("GET")
	configRouter.Path("").HandlerFunc(router.externalLinkRestHandler.GetExternalLinks).Methods("GET")
	configRouter.Path("").HandlerFunc(router.externalLinkRestHandler.UpdateExternalLink).Methods("PUT")
	configRouter.Path("").HandlerFunc(router.externalLinkRestHandler.DeleteExternalLink).Queries("id", "{id}").Methods("DELETE")
}
