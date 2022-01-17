package k8s

import (
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/gorilla/mux"
)

type K8sApplicationRouter interface {
	InitK8sApplicationRouter(helmRouter *mux.Router)
}
type K8sApplicationRouterImpl struct {
	k8sApplicationRestHandler K8sApplicationRestHandler
}

func NewK8sApplicationRouterImpl(k8sApplicationRestHandler K8sApplicationRestHandler) *K8sApplicationRouterImpl {
	return &K8sApplicationRouterImpl{
		k8sApplicationRestHandler: k8sApplicationRestHandler,
	}
}

func (impl *K8sApplicationRouterImpl) InitK8sApplicationRouter(k8sAppRouter *mux.Router) {

	k8sAppRouter.Path("/resource").
		HandlerFunc(impl.k8sApplicationRestHandler.GetResource).Methods("POST")

	k8sAppRouter.Path("/resource/create").
		HandlerFunc(impl.k8sApplicationRestHandler.CreateResource).Methods("POST")

	k8sAppRouter.Path("/resource").
		HandlerFunc(impl.k8sApplicationRestHandler.UpdateResource).Methods("PUT")

	k8sAppRouter.Path("/resource/delete").
		HandlerFunc(impl.k8sApplicationRestHandler.DeleteResource).Methods("POST")

	k8sAppRouter.Path("/events").
		HandlerFunc(impl.k8sApplicationRestHandler.ListEvents).Methods("POST")

	k8sAppRouter.Path("/pods/logs/{podName}").
		Queries("containerName", "{containerName}", "appId", "{appId}").
		//Queries("sinceSeconds", "{sinceSeconds}").
		Queries("follow", "{follow}").
		Queries("tailLines", "{tailLines}").
		HandlerFunc(impl.k8sApplicationRestHandler.GetPodLogs).Methods("GET")

	k8sAppRouter.Path("/pod/exec/session/{applicationId}/{namespace}/{pod}/{shell}/{container}").
		HandlerFunc(impl.k8sApplicationRestHandler.GetTerminalSession).Methods("GET")
	k8sAppRouter.PathPrefix("/pod/exec/sockjs/ws").Handler(terminal.CreateAttachHandler("/pod/exec/sockjs/ws"))

	/*k8sAppRouter.Path("/pod/exec/sockjs/ws/").
		Handler(terminal.CreateAttachHandler("/api/v1/applications/pod/exec/sockjs/ws/"))*/
}