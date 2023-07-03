package drafts

import (
	"github.com/gorilla/mux"
)

type ConfigDraftRouter interface {
	InitConfigDraftRouter(configRouter *mux.Router)
}

type ConfigDraftRouterImpl struct {
	configDraftRestHandler ConfigDraftRestHandler
}

func NewConfigDraftRouterImpl(configDraftRestHandler ConfigDraftRestHandler) *ConfigDraftRouterImpl {
	return &ConfigDraftRouterImpl{configDraftRestHandler: configDraftRestHandler}
}

func (router *ConfigDraftRouterImpl) InitConfigDraftRouter(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(router.configDraftRestHandler.CreateDraft).Methods("POST")
	configRouter.Path("/version").HandlerFunc(router.configDraftRestHandler.AddDraftVersion).Methods("PUT")
}
