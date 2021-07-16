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

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type TeamRouter interface {
	InitTeamRouter(gocdRouter *mux.Router)
}
type TeamRouterImpl struct {
	teamRestHandler restHandler.TeamRestHandler
}

func NewTeamRouterImpl(teamRestHandler restHandler.TeamRestHandler) *TeamRouterImpl {
	return &TeamRouterImpl{teamRestHandler: teamRestHandler}
}

func (impl TeamRouterImpl) InitTeamRouter(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(impl.teamRestHandler.SaveTeam).Methods("POST")
	configRouter.Path("").HandlerFunc(impl.teamRestHandler.FetchAll).Methods("GET")
	//make sure autocomplete API, must add before FetchOne API
	configRouter.Path("/autocomplete").HandlerFunc(impl.teamRestHandler.FetchForAutocomplete).Methods("GET")
	configRouter.Path("/{id}").HandlerFunc(impl.teamRestHandler.FetchOne).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.teamRestHandler.UpdateTeam).Methods("PUT")

	configRouter.Path("/app/by-id/{appId}").HandlerFunc(impl.teamRestHandler.FindTeamByAppId).Methods("GET")
	configRouter.Path("/app/by-name/{appName}").HandlerFunc(impl.teamRestHandler.FindActiveTeamByAppName).Methods("GET")

	configRouter.Path("/app/user").HandlerFunc(impl.teamRestHandler.FetchForUser).Methods("GET")

}
