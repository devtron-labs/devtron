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
	configRouter.Path("/app").HandlerFunc(router.configDraftRestHandler.GetDrafts).
		Queries("appId", "{appId}").
		Queries("envId", "{envId}").
		Queries("resourceType", "{resourceType}").
		Methods("GET")
	configRouter.Path("/{draftId}").HandlerFunc(router.configDraftRestHandler.GetDraftById).Methods("GET")
	configRouter.Path("/version").HandlerFunc(router.configDraftRestHandler.AddDraftVersion).Methods("PUT")
	configRouter.Path("/version/{draftId}").HandlerFunc(router.configDraftRestHandler.GetDraftVersionMetadata).Methods("GET")
	configRouter.Path("/version/comments/{draftId}").HandlerFunc(router.configDraftRestHandler.GetDraftComments).Methods("GET")
}
