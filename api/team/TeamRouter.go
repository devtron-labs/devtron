/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package team

import (
	"github.com/gorilla/mux"
)

type TeamRouter interface {
	InitTeamRouter(gocdRouter *mux.Router)
}
type TeamRouterImpl struct {
	teamRestHandler TeamRestHandler
}

func NewTeamRouterImpl(teamRestHandler TeamRestHandler) *TeamRouterImpl {
	return &TeamRouterImpl{teamRestHandler: teamRestHandler}
}

func (impl TeamRouterImpl) InitTeamRouter(configRouter *mux.Router) {
	configRouter.Path("").HandlerFunc(impl.teamRestHandler.SaveTeam).Methods("POST")
	configRouter.Path("").HandlerFunc(impl.teamRestHandler.FetchAll).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.teamRestHandler.DeleteTeam).Methods("DELETE")
	//make sure autocomplete API, must add before FetchOne API
	configRouter.Path("/autocomplete").HandlerFunc(impl.teamRestHandler.FetchForAutocomplete).Methods("GET")
	configRouter.Path("/{id}").HandlerFunc(impl.teamRestHandler.FetchOne).Methods("GET")
	configRouter.Path("").HandlerFunc(impl.teamRestHandler.UpdateTeam).Methods("PUT")
}
