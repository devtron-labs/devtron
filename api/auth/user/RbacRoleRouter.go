/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package user

import (
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type RbacRoleRouter interface {
	InitRbacRoleRouter(rbacRoleRouter *mux.Router)
}

type RbacRoleRouterImpl struct {
	logger              *zap.SugaredLogger
	validator           *validator.Validate
	rbacRoleRestHandler RbacRoleRestHandler
}

func NewRbacRoleRouterImpl(logger *zap.SugaredLogger,
	validator *validator.Validate, rbacRoleRestHandler RbacRoleRestHandler) *RbacRoleRouterImpl {
	rbacRoleRouterImpl := &RbacRoleRouterImpl{
		logger:              logger,
		validator:           validator,
		rbacRoleRestHandler: rbacRoleRestHandler,
	}
	return rbacRoleRouterImpl
}

func (router RbacRoleRouterImpl) InitRbacRoleRouter(rbacRoleRouter *mux.Router) {
	rbacRoleRouter.Path("").
		HandlerFunc(router.rbacRoleRestHandler.GetAllDefaultRoles).Methods("GET")

	rbacRoleRouter.Path("").
		HandlerFunc(router.rbacRoleRestHandler.CreateDefaultRole).Methods("POST")

	rbacRoleRouter.Path("").
		HandlerFunc(router.rbacRoleRestHandler.UpdateDefaultRole).Methods("PUT")

	rbacRoleRouter.Path("/sync").
		HandlerFunc(router.rbacRoleRestHandler.SyncDefaultRoles).Methods("POST")

	rbacRoleRouter.Path("/{id}").
		HandlerFunc(router.rbacRoleRestHandler.GetDefaultRoleDetailById).Methods("GET")

	rbacRoleRouter.Path("/entity/{entity}").
		HandlerFunc(router.rbacRoleRestHandler.GetAllDefaultRolesByEntityAccessType).Methods("GET")

	rbacRoleRouter.Path("/policy/resource/list").
		HandlerFunc(router.rbacRoleRestHandler.GetRbacPolicyResourceListForAllEntityAccessTypes).Methods("GET")

	rbacRoleRouter.Path("/policy/resource/{entity}").
		HandlerFunc(router.rbacRoleRestHandler.GetRbacPolicyResourceListByEntityAndAccessType).Methods("GET")
}
