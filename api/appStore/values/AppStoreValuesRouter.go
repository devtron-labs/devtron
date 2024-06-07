/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package appStoreValues

import (
	"github.com/gorilla/mux"
)

type AppStoreValuesRouter interface {
	Init(configRouter *mux.Router)
}

type AppStoreValuesRouterImpl struct {
	appStoreValuesRestHandler AppStoreValuesRestHandler
}

func NewAppStoreValuesRouterImpl(appStoreValuesRestHandler AppStoreValuesRestHandler) *AppStoreValuesRouterImpl {
	return &AppStoreValuesRouterImpl{
		appStoreValuesRestHandler: appStoreValuesRestHandler,
	}
}

func (router AppStoreValuesRouterImpl) Init(configRouter *mux.Router) {

	configRouter.Path("/template/values").
		HandlerFunc(router.appStoreValuesRestHandler.CreateAppStoreVersionValues).Methods("POST")
	configRouter.Path("/template/values").
		HandlerFunc(router.appStoreValuesRestHandler.UpdateAppStoreVersionValues).Methods("PUT")
	configRouter.Path("/template/values").Queries("referenceId", "{referenceId}", "kind", "{kind}").
		HandlerFunc(router.appStoreValuesRestHandler.FindValuesById).Methods("GET")
	configRouter.Path("/template/values/{appStoreValueId}").
		HandlerFunc(router.appStoreValuesRestHandler.DeleteAppStoreVersionValues).Methods("DELETE")

	//used for manage api listing, will return only saved(template) values
	configRouter.Path("/template/values/list/{appStoreId}").
		HandlerFunc(router.appStoreValuesRestHandler.FindValuesByAppStoreIdAndReferenceType).Methods("GET")

	//used for all types of values category wise
	configRouter.Path("/application/values/list/{appStoreId}").
		HandlerFunc(router.appStoreValuesRestHandler.FetchTemplateValuesByAppStoreId).Methods("GET")

	configRouter.Path("/chart/selected/metadata").
		HandlerFunc(router.appStoreValuesRestHandler.GetSelectedChartMetadata).Methods("POST")

}
