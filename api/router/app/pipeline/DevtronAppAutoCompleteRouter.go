/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
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
	configRouter.Path("/{appId}/autocomplete/environment").HandlerFunc(router.devtronAppAutoCompleteRestHandler.EnvironmentListAutocomplete).Methods("GET")
	configRouter.Path("/{appId}/autocomplete/git").HandlerFunc(router.devtronAppAutoCompleteRestHandler.GitListAutocomplete).Methods("GET")
	configRouter.Path("/{appId}/autocomplete/docker").HandlerFunc(router.devtronAppAutoCompleteRestHandler.RegistriesListAutocomplete).Methods("GET")
	configRouter.Path("/{appId}/autocomplete/team").HandlerFunc(router.devtronAppAutoCompleteRestHandler.TeamListAutocomplete).Methods("GET")
}
