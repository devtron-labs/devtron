package externalLinks

import (
	"github.com/gorilla/mux"
)

type ExternalLinksRouter interface {
	InitExternalLinksRouter(gocdRouter *mux.Router)
}
type ExternalLinksRouterImpl struct {
	externalLinksRestHandler ExternalLinksRestHandler
}

func NewExternalLinksRouterImpl(externalLinksRestHandler ExternalLinksRestHandler) *ExternalLinksRouterImpl {
	return &ExternalLinksRouterImpl{externalLinksRestHandler: externalLinksRestHandler}
}

func (impl ExternalLinksRouterImpl) InitExternalLinksRouter(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(impl.externalLinksRestHandler.CreateExternalLinks).Methods("POST")
	configRouter.Path("/tools").HandlerFunc(impl.externalLinksRestHandler.GetExternalLinksTools).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.externalLinksRestHandler.GetExternalLinks).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.externalLinksRestHandler.UpdateExternalLinks).Methods("PUT")
	configRouter.Path("").HandlerFunc(impl.externalLinksRestHandler.DeleteExternalLinks).Queries("id", "{id}").Methods("DELETE")
}
