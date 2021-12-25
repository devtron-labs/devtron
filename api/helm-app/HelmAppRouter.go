package client

import "github.com/gorilla/mux"

type HelmAppRouter interface {
	InitAppListRouter(helmRouter *mux.Router)
}
type HelmAppRouterImpl struct {
	helmAppRestHandler HelmAppRestHandler
}

func NewHelmAppRouterImpl(helmAppRestHandler HelmAppRestHandler) *HelmAppRouterImpl {
	return &HelmAppRouterImpl{
		helmAppRestHandler: helmAppRestHandler,
	}
}

func (impl *HelmAppRouterImpl) InitAppListRouter(helmRouter *mux.Router) {
	helmRouter.Path("").Queries("clusterIds", "{clusterIds}").
		HandlerFunc(impl.helmAppRestHandler.ListApplications).Methods("GET")
	helmRouter.Path("/app").Queries("appId", "{appId}").
		HandlerFunc(impl.helmAppRestHandler.GetApplicationDetail).Methods("GET")
	helmRouter.Path("/hibernate").HandlerFunc(impl.helmAppRestHandler.Hibernate).Methods("POST")
	helmRouter.Path("/unhibernate").HandlerFunc(impl.helmAppRestHandler.UnHibernate).Methods("POST")

	helmRouter.Path("/{name}/resource").Queries("appId", "{appId}").
		HandlerFunc(impl.helmAppRestHandler.GetResource).Methods("POST")

	helmRouter.Path("/{name}/resource").Queries("appId", "{appId}").
		HandlerFunc(impl.helmAppRestHandler.UpdateResource).Methods("PUT")

	helmRouter.Path("/{name}/resource/delete").Queries("appId", "{appId}").
		HandlerFunc(impl.helmAppRestHandler.DeleteResource).Methods("POST")

	helmRouter.Path("/{name}/events").Queries("appId", "{appId}").
		HandlerFunc(impl.helmAppRestHandler.ListEvents).Methods("POST")
}
