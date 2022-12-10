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

func (router *HelmAppRouterImpl) InitAppListRouter(helmRouter *mux.Router) {
	helmRouter.Path("").Queries("clusterIds", "{clusterIds}").
		HandlerFunc(router.helmAppRestHandler.ListApplications).Methods("GET")
	helmRouter.Path("/app").Queries("appId", "{appId}").
		HandlerFunc(router.helmAppRestHandler.GetApplicationDetail).Methods("GET")

	helmRouter.Path("/app/save-telemetry").Queries("appId", "{appId}").
		HandlerFunc(router.helmAppRestHandler.SaveHelmAppDetailsViewedTelemetryData).Methods("GET")

	helmRouter.Path("/hibernate").HandlerFunc(router.helmAppRestHandler.Hibernate).Methods("POST")
	helmRouter.Path("/unhibernate").HandlerFunc(router.helmAppRestHandler.UnHibernate).Methods("POST")

	helmRouter.Path("/release-info").Queries("appId", "{appId}").
		HandlerFunc(router.helmAppRestHandler.GetReleaseInfo).Methods("GET")

	helmRouter.Path("/desired-manifest").HandlerFunc(router.helmAppRestHandler.GetDesiredManifest).Methods("POST")

	helmRouter.Path("/update").HandlerFunc(router.helmAppRestHandler.UpdateApplication).Methods("PUT")

	helmRouter.Path("/delete").Queries("appId", "{appId}").
		HandlerFunc(router.helmAppRestHandler.DeleteApplication).Methods("DELETE")

	helmRouter.Path("/template-chart").HandlerFunc(router.helmAppRestHandler.TemplateChart).Methods("POST")
}
