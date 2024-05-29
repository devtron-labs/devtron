/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package pipeline

import (
	"github.com/devtron-labs/devtron/api/restHandler/app/pipeline"
	"github.com/gorilla/mux"
)

type DevtronAppAutoCompleteRouter interface {
	InitDevtronAppAutoCompleteRouter(configRouter *mux.Router)
}
type DevtronAppAutoCompleteRouterImpl struct {
	devtronAppAutoCompleteRestHandler pipeline.DevtronAppAutoCompleteRestHandler
}

func NewDevtronAppAutoCompleteRouterImpl(
	devtronAppAutoCompleteRestHandler pipeline.DevtronAppAutoCompleteRestHandler,
) *DevtronAppAutoCompleteRouterImpl {
	return &DevtronAppAutoCompleteRouterImpl{
		devtronAppAutoCompleteRestHandler: devtronAppAutoCompleteRestHandler,
	}
}

func (router DevtronAppAutoCompleteRouterImpl) InitDevtronAppAutoCompleteRouter(configRouter *mux.Router) {
	configRouter.Path("/autocomplete").HandlerFunc(router.devtronAppAutoCompleteRestHandler.GetAppListForAutocomplete).Methods("GET")
	configRouter.Path("/autocomplete/all").HandlerFunc(router.devtronAppAutoCompleteRestHandler.GetAppListAllWithoutRBAC).Methods("GET")
	configRouter.Path("/{appId}/autocomplete/environment").HandlerFunc(router.devtronAppAutoCompleteRestHandler.EnvironmentListAutocomplete).Methods("GET")
	configRouter.Path("/{appId}/autocomplete/git").HandlerFunc(router.devtronAppAutoCompleteRestHandler.GitListAutocomplete).Methods("GET")
	configRouter.Path("/{appId}/autocomplete/docker").HandlerFunc(router.devtronAppAutoCompleteRestHandler.RegistriesListAutocomplete).Methods("GET")
	configRouter.Path("/{appId}/autocomplete/team").HandlerFunc(router.devtronAppAutoCompleteRestHandler.TeamListAutocomplete).Methods("GET")
}
