/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
	userAuthRouter.Path("/v2").
		HandlerFunc(router.userRestHandler.GetAllV2).Methods("GET")
	userAuthRouter.Path("/{id}").
		HandlerFunc(router.userRestHandler.GetById).Methods("GET")
	userAuthRouter.Path("").
		HandlerFunc(router.userRestHandler.CreateUser).Methods("POST")
	userAuthRouter.Path("").
		HandlerFunc(router.userRestHandler.GetAll).Methods("GET")
	userAuthRouter.Path("").
		HandlerFunc(router.userRestHandler.UpdateUser).Methods("PUT")
	userAuthRouter.Path("/cleanup").
		HandlerFunc(router.userRestHandler.CleanUpPolicies).Methods("DELETE")
	userAuthRouter.Path("/bulk").
		HandlerFunc(router.userRestHandler.BulkDeleteUsers).Methods("DELETE")
	userAuthRouter.Path("/{id}").
		HandlerFunc(router.userRestHandler.DeleteUser).Methods("DELETE")
	userAuthRouter.Path("/status").
		HandlerFunc(router.userRestHandler.BulkUpdateStatus).Methods("PUT")
	userAuthRouter.Path("/detail/get").
		HandlerFunc(router.userRestHandler.GetAllDetailedUsers).Methods("GET")

	userAuthRouter.Path("/role/group/v2").
		HandlerFunc(router.userRestHandler.FetchRoleGroupsV2).Methods("GET")
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
	userAuthRouter.Path("/role/group/bulk").
		HandlerFunc(router.userRestHandler.BulkDeleteRoleGroups).Methods("DELETE")
	userAuthRouter.Path("/role/group/{id}").
		HandlerFunc(router.userRestHandler.DeleteRoleGroup).Methods("DELETE")

	userAuthRouter.Path("/check/roles").
		HandlerFunc(router.userRestHandler.CheckUserRoles).Methods("GET")
	userAuthRouter.Path("/sync/orchestratortocasbin").
		HandlerFunc(router.userRestHandler.SyncOrchestratorToCasbin).Methods("GET")
	userAuthRouter.Path("/role/cache").
		HandlerFunc(router.userRestHandler.GetRoleCacheDump).Methods("GET")
	userAuthRouter.Path("/role/cache/invalidate").
		HandlerFunc(router.userRestHandler.InvalidateRoleCache).Methods("GET")
}
