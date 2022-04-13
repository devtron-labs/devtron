package externalLinkout

import (
	"github.com/gorilla/mux"
)

type ExternalLinkoutRouter interface {
	InitExternalLinkoutRouter(gocdRouter *mux.Router)
}
type ExternalLinkoutRouterImpl struct {
	externalLinkoutRestHandler ExternalLinkoutRestHandler
}

func NewExternalLinkoutRouterImpl(externalLinkoutRestHandler ExternalLinkoutRestHandler) *ExternalLinkoutRouterImpl {
	return &ExternalLinkoutRouterImpl{externalLinkoutRestHandler: externalLinkoutRestHandler}
}

func (impl ExternalLinkoutRouterImpl) InitExternalLinkoutRouter(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(impl.externalLinkoutRestHandler.CreateExternalLinks).Methods("POST")

	configRouter.Path("/tools").HandlerFunc(impl.externalLinkoutRestHandler.GetAllTools).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.externalLinkoutRestHandler.GetAllLinks).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.externalLinkoutRestHandler.Update).Methods("PUT")

	//configRouter.Path("").HandlerFunc(impl.externalLinkoutRestHandler.Delete).Methods("PUT")
}
