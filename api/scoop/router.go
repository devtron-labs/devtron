package scoop

import (
	"github.com/devtron-labs/devtron/api/restHandler/autoRemediation"
	"github.com/gorilla/mux"
)

type Router interface {
	InitScoopRouter(router *mux.Router)
}

type RouterImpl struct {
	restHandler        RestHandler
	watcherRestHandler autoRemediation.WatcherRestHandler
}

func NewRouterImpl(restHandler RestHandler, watcherRestHandler autoRemediation.WatcherRestHandler) *RouterImpl {
	return &RouterImpl{
		restHandler:        restHandler,
		watcherRestHandler: watcherRestHandler,
	}
}

func (impl RouterImpl) InitScoopRouter(router *mux.Router) {
	router.Path("/intercept-event").
		HandlerFunc(impl.restHandler.HandleInterceptedEvent).Methods("POST")

	router.Path("/k8s/watcher").HandlerFunc(impl.watcherRestHandler.SaveWatcher).Methods("POST")

	//router.Path("/k8s/watcher").Queries("search", "{search}").
	//	Queries("orderBy", "{orderBy}").
	//	Queries("order", "{order}").
	//	Queries("offset", "{offset}").
	//	Queries("size", "{size}").HandlerFunc(impl.watcherRestHandler.RetrieveWatchers).Methods("GET")

	router.Path("/k8s/watcher/{identifier}").HandlerFunc(impl.watcherRestHandler.GetWatcherById).Methods("GET")

	router.Path("/k8s/watcher/{identifier}").HandlerFunc(impl.watcherRestHandler.DeleteWatcherById).Methods("DELETE")

	// k8sAppRouter.Path("/watcher/events").HandlerFunc(impl.watcherRestHandler.RetrieveInterceptedEvents).Methods("GET")
	router.Path("/k8s/watcher/{identifier}").HandlerFunc(impl.watcherRestHandler.UpdateWatcherById).Methods("PUT")

	// k8sAppRouter.Path("").
	//	Queries("watchers", "{watchers}").
	//	Queries("clusters", "{clusters}").
	//	Queries("namespaces", "{namespaces}").
	//	Queries("executionStatuses", "{executionStatuses}").
	//	Queries("from", "{from}").
	//	Queries("to", "{to}").
	//	Queries("offset", "{offset}").
	//	Queries("size", "{size}").
	//	Queries("searchString", "{searchString}").
	//	HandlerFunc(impl.watcherRestHandler.RetrieveWatchers).
	//	Methods("GET")

}
