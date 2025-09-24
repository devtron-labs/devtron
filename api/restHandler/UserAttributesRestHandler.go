/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package restHandler

import (
	"encoding/json"
	"net/http"

	"github.com/devtron-labs/devtron/pkg/attributes/bean"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"go.uber.org/zap"
)

type UserAttributesRestHandler interface {
	AddUserAttributes(w http.ResponseWriter, r *http.Request)
	UpdateUserAttributes(w http.ResponseWriter, r *http.Request)
	PatchUserAttributes(w http.ResponseWriter, r *http.Request)
	GetUserAttribute(w http.ResponseWriter, r *http.Request)
}

type UserAttributesRestHandlerImpl struct {
	logger                *zap.SugaredLogger
	enforcer              casbin.Enforcer
	userService           user.UserService
	userAttributesService attributes.UserAttributesService
}

func NewUserAttributesRestHandlerImpl(logger *zap.SugaredLogger, enforcer casbin.Enforcer,
	userService user.UserService, userAttributesService attributes.UserAttributesService) *UserAttributesRestHandlerImpl {
	userAuthHandler := &UserAttributesRestHandlerImpl{
		logger:                logger,
		enforcer:              enforcer,
		userService:           userService,
		userAttributesService: userAttributesService,
	}
	return userAuthHandler
}

func (handler *UserAttributesRestHandlerImpl) AddUserAttributes(w http.ResponseWriter, r *http.Request) {
	dto, success := handler.validateUserAttributesRequest(w, r, "AddUserAttributes")
	if !success {
		return
	}

	handler.logger.Infow("Adding user attributes",
		"operation", "add_user_attributes",
		"userId", dto.UserId,
		"key", dto.Key)

	resp, err := handler.userAttributesService.AddUserAttributes(dto)
	if err != nil {
		handler.logger.Errorw("Failed to add user attributes",
			"operation", "add_user_attributes",
			"userId", dto.UserId,
			"key", dto.Key,
			"err", err)

		// Use enhanced error response builder
		errBuilder := common.NewErrorResponseBuilder(w, r).
			WithOperation("user attributes creation").
			WithResource("user attribute", dto.Key)
		errBuilder.HandleError(err)
		return
	}

	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

// @Summary update user attributes
// @version 1.0
// @produce application/json
// @Param payload body attributes.UserAttributesDto true "Input key"
// @Success 200 {object} attributes.UserAttributesDto
// @Router /orchestrator/attributes/user/update [POST]
func (handler *UserAttributesRestHandlerImpl) UpdateUserAttributes(w http.ResponseWriter, r *http.Request) {
	dto, success := handler.validateUserAttributesRequest(w, r, "UpdateUserAttributes")
	if !success {
		return
	}

	handler.logger.Infow("Updating user attributes",
		"operation", "update_user_attributes",
		"userId", dto.UserId,
		"key", dto.Key)

	resp, err := handler.userAttributesService.UpdateUserAttributes(dto)
	if err != nil {
		handler.logger.Errorw("Failed to update user attributes",
			"operation", "update_user_attributes",
			"userId", dto.UserId,
			"key", dto.Key,
			"err", err)

		// Use enhanced error response builder
		errBuilder := common.NewErrorResponseBuilder(w, r).
			WithOperation("user attributes update").
			WithResource("user attribute", dto.Key)
		errBuilder.HandleError(err)
		return
	}

	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler *UserAttributesRestHandlerImpl) PatchUserAttributes(w http.ResponseWriter, r *http.Request) {
	dto, success := handler.validateUserAttributesRequest(w, r, "PatchUserAttributes")
	if !success {
		return
	}

	handler.logger.Infow("Patching user attributes",
		"operation", "patch_user_attributes",
		"userId", dto.UserId,
		"key", dto.Key)

	resp, err := handler.userAttributesService.PatchUserAttributes(dto)
	if err != nil {
		handler.logger.Errorw("Failed to patch user attributes",
			"operation", "patch_user_attributes",
			"userId", dto.UserId,
			"key", dto.Key,
			"err", err)

		// Use enhanced error response builder
		errBuilder := common.NewErrorResponseBuilder(w, r).
			WithOperation("user attributes patch").
			WithResource("user attribute", dto.Key)
		errBuilder.HandleError(err)
		return
	}

	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler *UserAttributesRestHandlerImpl) validateUserAttributesRequest(w http.ResponseWriter, r *http.Request, operation string) (*bean.UserAttributesDto, bool) {
	// 1. Authentication check using enhanced error handling
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return nil, false
	}

	// 2. Request body parsing with enhanced error handling
	decoder := json.NewDecoder(r.Body)
	var dto bean.UserAttributesDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("Request parsing error",
			"operation", operation,
			"err", err,
			"userId", userId)

		// Use enhanced error response builder for request parsing errors
		errBuilder := common.NewErrorResponseBuilder(w, r).
			WithOperation(operation).
			WithResource("user attributes", "")
		errBuilder.HandleError(err)
		return nil, false
	}

	dto.UserId = userId

	// 3. Get user email with enhanced error handling
	emailId, err := handler.userService.GetActiveEmailById(userId)
	if err != nil {
		handler.logger.Errorw("Failed to get user email",
			"operation", operation,
			"userId", userId,
			"err", err)

		// Use enhanced error response for forbidden access
		common.WriteForbiddenError(w, "access user attributes", "user")
		return nil, false
	}
	dto.EmailId = emailId

	return &dto, true
}

// @Summary get user attributes
// @version 1.0
// @produce application/json
// @Param name query string true "Input key"
// @Success 200 {object} attributes.UserAttributesDto
// @Router /orchestrator/attributes/user/get [GET]
func (handler *UserAttributesRestHandlerImpl) GetUserAttribute(w http.ResponseWriter, r *http.Request) {
	// 1. Authentication check using enhanced error handling
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return
	}

	// 2. Enhanced parameter extraction with automatic validation
	key, err := common.ExtractStringPathParamWithContext(w, r, "key")
	if err != nil {
		// Error already written by ExtractStringPathParamWithContext
		return
	}

	// 3. Get user email with enhanced error handling
	emailId, err := handler.userService.GetActiveEmailById(userId)
	if err != nil {
		handler.logger.Errorw("Failed to get user email",
			"operation", "get_user_attribute",
			"userId", userId,
			"key", key,
			"err", err)

		// Use enhanced error response for forbidden access
		common.WriteForbiddenError(w, "access user attributes", "user")
		return
	}

	// 4. Prepare DTO
	dto := bean.UserAttributesDto{
		UserId:  userId,
		EmailId: emailId,
		Key:     key,
	}

	// 5. Service call with enhanced error handling
	res, err := handler.userAttributesService.GetUserAttribute(&dto)
	if err != nil {
		handler.logger.Errorw("Failed to get user attribute",
			"operation", "get_user_attribute",
			"userId", userId,
			"key", key,
			"err", err)

		// Use enhanced error response builder
		errBuilder := common.NewErrorResponseBuilder(w, r).
			WithOperation("user attribute retrieval").
			WithResource("user attribute", key)
		errBuilder.HandleError(err)
		return
	}

	// 6. Success response
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}
