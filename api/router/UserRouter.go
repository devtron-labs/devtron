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
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/pkg/dex"
	"github.com/argoproj/argo-cd/util/settings"
	"github.com/gorilla/mux"
)

type UserRouter interface {
	initUserRouter(helmRouter *mux.Router)
}

type UserRouterImpl struct {
	userRestHandler restHandler.UserRestHandler
}

func NewUserRouterImpl(userRestHandler restHandler.UserRestHandler, dexCfg *dex.Config, cdCfg *argocdServer.Config, settings *settings.ArgoCDSettings) *UserRouterImpl {
	tlsConfig := settings.TLSConfig()
	if tlsConfig != nil {
		tlsConfig.InsecureSkipVerify = true
	}
	router := &UserRouterImpl{
		userRestHandler: userRestHandler,
	}
	return router
}

func (router UserRouterImpl) initUserRouter(userAuthRouter *mux.Router) {
	//User management
	userAuthRouter.Path("/{id}").
		HandlerFunc(router.userRestHandler.GetById).Methods("GET")
	userAuthRouter.Path("").
		HandlerFunc(router.userRestHandler.CreateUser).Methods("POST")
	userAuthRouter.Path("/email").
		Queries("email-id", "{email-id}").
		HandlerFunc(router.userRestHandler.GetUserByEmail).Methods("GET")
	userAuthRouter.Path("").
		HandlerFunc(router.userRestHandler.GetAll).Methods("GET")
	userAuthRouter.Path("/filter").
		Queries("size", "{size}").
		Queries("from", "{from}").
		HandlerFunc(router.userRestHandler.GetUsersByFilter).Methods("GET")
	userAuthRouter.Path("").
		HandlerFunc(router.userRestHandler.UpdateUser).Methods("PUT")
	userAuthRouter.Path("/{id}").
		HandlerFunc(router.userRestHandler.DeleteUser).Methods("DELETE")

	userAuthRouter.Path("/role/group/{id}").
		HandlerFunc(router.userRestHandler.FetchRoleGroupById).Methods("GET")
	userAuthRouter.Path("/role/group").
		HandlerFunc(router.userRestHandler.CreateRoleGroup).Methods("POST")
	userAuthRouter.Path("/role/group").
		HandlerFunc(router.userRestHandler.UpdateRoleGroup).Methods("PUT")
	userAuthRouter.Path("/role/group").
		HandlerFunc(router.userRestHandler.FetchRoleGroups).Methods("GET")
	userAuthRouter.Path("/role/group/search").
		Queries("name", "{name}").
		HandlerFunc(router.userRestHandler.FetchRoleGroupsByName).Methods("GET")
	userAuthRouter.Path("/role/group/{id}").
		HandlerFunc(router.userRestHandler.DeleteRoleGroup).Methods("DELETE")

	userAuthRouter.Path("/check/roles").
		HandlerFunc(router.userRestHandler.CheckUserRoles).Methods("GET")
	userAuthRouter.Path("/sync/orchestratortocasbin").
		HandlerFunc(router.userRestHandler.SyncOrchestratorToCasbin).Methods("GET")
}
