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

func (impl ExternalLinkRouterImpl) InitExternalLinkRouter(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(impl.externalLinkRestHandler.CreateExternalLinks).Methods("POST")
	configRouter.Path("/tools").HandlerFunc(impl.externalLinkRestHandler.GetExternalLinkMonitoringTools).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.externalLinkRestHandler.GetExternalLinks).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.externalLinkRestHandler.UpdateExternalLink).Methods("PUT")
	configRouter.Path("").HandlerFunc(impl.externalLinkRestHandler.DeleteExternalLink).Queries("id", "{id}").Methods("DELETE")
}
