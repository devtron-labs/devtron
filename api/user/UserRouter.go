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

package user

import (
	"github.com/gorilla/mux"
)

type UserRouter interface {
	InitUserRouter(helmRouter *mux.Router)
}

type UserRouterImpl struct {
	userRestHandler UserRestHandler
}

func NewUserRouterImpl(userRestHandler UserRestHandler) *UserRouterImpl {
	router := &UserRouterImpl{
		userRestHandler: userRestHandler,
	}
	return router
}

func (router UserRouterImpl) InitUserRouter(userAuthRouter *mux.Router) {
	//User management
	userAuthRouter.Path("/{id}").
		HandlerFunc(router.userRestHandler.GetById).Methods("GET")
	userAuthRouter.Path("").
		HandlerFunc(router.userRestHandler.CreateUser).Methods("POST")
	userAuthRouter.Path("").
		HandlerFunc(router.userRestHandler.GetAll).Methods("GET")
	userAuthRouter.Path("").
		HandlerFunc(router.userRestHandler.UpdateUser).Methods("PUT")
	userAuthRouter.Path("/{id}").
		HandlerFunc(router.userRestHandler.DeleteUser).Methods("DELETE")

	userAuthRouter.Path("/detail/get").
		HandlerFunc(router.userRestHandler.GetAllDetailedUsers).Methods("GET")

	userAuthRouter.Path("/role/group/{id}").
		HandlerFunc(router.userRestHandler.FetchRoleGroupById).Methods("GET")
	userAuthRouter.Path("/role/group").
		HandlerFunc(router.userRestHandler.CreateRoleGroup).Methods("POST")
	userAuthRouter.Path("/role/group").
		HandlerFunc(router.userRestHandler.UpdateRoleGroup).Methods("PUT")
	userAuthRouter.Path("/role/group").
		HandlerFunc(router.userRestHandler.FetchRoleGroups).Methods("GET")
	userAuthRouter.Path("/role/group/detailed/get").
		HandlerFunc(router.userRestHandler.FetchDetailedRoleGroups).Methods("GET")
	userAuthRouter.Path("/role/group/search").
		Queries("name", "{name}").
		HandlerFunc(router.userRestHandler.FetchRoleGroupsByName).Methods("GET")
	userAuthRouter.Path("/role/group/{id}").
		HandlerFunc(router.userRestHandler.DeleteRoleGroup).Methods("DELETE")

	userAuthRouter.Path("/check/roles").
		HandlerFunc(router.userRestHandler.CheckUserRoles).Methods("GET")
	userAuthRouter.Path("/sync/orchestratortocasbin").
		HandlerFunc(router.userRestHandler.SyncOrchestratorToCasbin).Methods("GET")
	userAuthRouter.Path("/update/trigger/terminal").
		HandlerFunc(router.userRestHandler.UpdateTriggerPolicyForTerminalAccess).Methods("PUT")
	userAuthRouter.Path("/role/cache").
		HandlerFunc(router.userRestHandler.GetRoleCacheDump).Methods("GET")
	userAuthRouter.Path("/role/cache/invalidate").
		HandlerFunc(router.userRestHandler.InvalidateRoleCache).Methods("GET")
}
