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
	configRouter.Path("").HandlerFunc(router.configDraftRestHandler.UpdateDraftState).
		Queries("draftId", "{draftId}").
		Queries("draftVersionId", "{draftVersionId}").
		Queries("state", "{state}").
		Methods("PUT")
	configRouter.Path("/app").HandlerFunc(router.configDraftRestHandler.GetAppDrafts).
		Queries("appId", "{appId}").
		Queries("envId", "{envId}").
		Queries("resourceType", "{resourceType}").
		Methods("GET")
	configRouter.Path("/app/count").HandlerFunc(router.configDraftRestHandler.GetDraftsCount).
		Queries("appId", "{appId}").
		Queries("envIds", "{envIds}").
		Methods("GET")
	configRouter.Path("/{draftId}").HandlerFunc(router.configDraftRestHandler.GetDraftById).Methods("GET")
	configRouter.Path("").HandlerFunc(router.configDraftRestHandler.GetDraftByName).
		Queries("appId", "{appId}").
		Queries("envId", "{envId}").
		Queries("resourceName", "{resourceName}").
		Queries("resourceType", "{resourceType}").
		Methods("GET")
	configRouter.Path("/version").HandlerFunc(router.configDraftRestHandler.AddDraftVersion).Methods("PUT")
	configRouter.Path("/version/{draftId}").HandlerFunc(router.configDraftRestHandler.GetDraftVersionMetadata).Methods("GET")
	configRouter.Path("/version/comments/{draftId}").HandlerFunc(router.configDraftRestHandler.GetDraftComments).Methods("GET")
	configRouter.Path("/approve").HandlerFunc(router.configDraftRestHandler.ApproveDraft).
		Queries("draftId", "{draftId}").
		Queries("draftVersionId", "{draftVersionId}").
		Methods("POST")
	configRouter.Path("/version/comments").HandlerFunc(router.configDraftRestHandler.DeleteUserComment).
		Queries("draftId", "{draftId}").
		Queries("draftCommentId", "{draftCommentId}").
		Methods("DELETE")
}
