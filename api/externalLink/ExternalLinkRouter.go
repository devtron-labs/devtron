package externalLink

import (
	"github.com/devtron-labs/devtron/api/logger"
	"github.com/gorilla/mux"
)

type ExternalLinkRouter interface {
	InitExternalLinkRouter(gocdRouter *mux.Router)
}
type ExternalLinkRouterImpl struct {
	externalLinkRestHandler ExternalLinkRestHandler
	userAuth                logger.UserAuth
}

func NewExternalLinkRouterImpl(externalLinkRestHandler ExternalLinkRestHandler, userAuth logger.UserAuth) *ExternalLinkRouterImpl {
	return &ExternalLinkRouterImpl{externalLinkRestHandler: externalLinkRestHandler, userAuth: userAuth}
}

func (impl ExternalLinkRouterImpl) InitExternalLinkRouter(configRouter *mux.Router) {
	configRouter.Use(impl.userAuth.LoggingMiddleware)
	configRouter.Path("").HandlerFunc(impl.externalLinkRestHandler.CreateExternalLinks).Methods("POST")
	configRouter.Path("/tools").HandlerFunc(impl.externalLinkRestHandler.GetExternalLinkMonitoringTools).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.externalLinkRestHandler.GetExternalLinks).Methods("GET")
	configRouter.Path("/v2").HandlerFunc(impl.externalLinkRestHandler.GetExternalLinksV2).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.externalLinkRestHandler.UpdateExternalLink).Methods("PUT")
	configRouter.Path("").HandlerFunc(impl.externalLinkRestHandler.DeleteExternalLink).Queries("id", "{id}").Methods("DELETE")
}
