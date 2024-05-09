package scoop

import (
	"github.com/devtron-labs/devtron/api/restHandler/autoRemediation"
	"github.com/gorilla/mux"
)

type Router interface {
	InitScoopRouter(router *mux.Router)
}

type RouterImpl struct {
	interceptEventHandler RestHandler
	watcherRestHandler    autoRemediation.WatcherRestHandler
}

func NewRouterImpl(restHandler RestHandler, watcherRestHandler autoRemediation.WatcherRestHandler) *RouterImpl {
	return &RouterImpl{
		interceptEventHandler: restHandler,
		watcherRestHandler:    watcherRestHandler,
	}
}

func (impl RouterImpl) InitScoopRouter(router *mux.Router) {

	router.Path("/intercept-event/notify").
		HandlerFunc(impl.interceptEventHandler.HandleNotificationEvent).Methods("POST")

	router.Path("/intercept-event").
		HandlerFunc(impl.interceptEventHandler.HandleInterceptedEvent).Methods("POST")
	router.Path("/intercept-event/{identifier}").HandlerFunc(impl.watcherRestHandler.GetInterceptedEventById).Methods("GET")
	router.Path("/watchers/sync").
		Queries("clusterId", "{clusterId}").
		HandlerFunc(impl.watcherRestHandler.GetWatchersByClusterId).Methods("GET")

	router.Path("/k8s/watcher").HandlerFunc(impl.watcherRestHandler.SaveWatcher).Methods("POST")

	router.Path("/k8s/watcher").
		// Queries("search", "{search}").
		// Queries("orderBy", "{orderBy}").
		// Queries("order", "{order}").
		// Queries("offset", "{offset}").
		// Queries("size", "{size}").
		HandlerFunc(impl.watcherRestHandler.RetrieveWatchers).Methods("GET")
	router.Path("/k8s/intercept-events").
		// Queries("watchers", "{watchers}").
		// Queries("clusters", "{clusters}").
		// Queries("namespaces", "{namespaces}").
		// Queries("executionStatuses", "{executionStatuses}").
		// Queries("from", "{from}").
		// Queries("to", "{to}").
		// Queries("offset", "{offset}").
		// Queries("size", "{size}").
		// Queries("searchString", "{search}")
		HandlerFunc(impl.watcherRestHandler.RetrieveInterceptedEvents).
		Methods("GET")
	router.Path("/k8s/watcher/{identifier}").HandlerFunc(impl.watcherRestHandler.GetWatcherById).Methods("GET")

	router.Path("/k8s/watcher/{identifier}").HandlerFunc(impl.watcherRestHandler.DeleteWatcherById).Methods("DELETE")

	router.Path("/k8s/watcher/{identifier}").HandlerFunc(impl.watcherRestHandler.UpdateWatcherById).Methods("PUT")

}
