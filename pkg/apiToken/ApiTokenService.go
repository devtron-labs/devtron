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

package apiToken

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	openapi "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/go-pg/pg"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strings"
	"time"
)

type ApiTokenService interface {
	GetAllActiveApiTokens() ([]*openapi.ApiToken, error)
	CreateApiToken(request *openapi.CreateApiTokenRequest, createdBy int32) (*openapi.ActionResponse, error)
	UpdateApiToken(apiTokenId int, request *openapi.UpdateApiTokenRequest, updatedBy int32) (*openapi.ActionResponse, error)
	DeleteApiToken(apiTokenId int, deletedBy int32) (*openapi.ActionResponse, error)
}

type ApiTokenServiceImpl struct {
	logger                *zap.SugaredLogger
	apiTokenSecretService ApiTokenSecretService
	userService           user.UserService
	apiTokenRepository    ApiTokenRepository
}

func NewApiTokenServiceImpl(logger *zap.SugaredLogger, apiTokenSecretService ApiTokenSecretService, userService user.UserService, apiTokenRepository ApiTokenRepository) *ApiTokenServiceImpl {
	return &ApiTokenServiceImpl{
		logger:                logger,
		apiTokenSecretService: apiTokenSecretService,
		userService:           userService,
		apiTokenRepository:    apiTokenRepository,
	}
}

const API_TOKEN_USER_EMAIL_PREFIX = "api-token:"

func (impl ApiTokenServiceImpl) GetAllActiveApiTokens() ([]*openapi.ApiToken, error) {
	impl.logger.Info("Getting all active api tokens")
	apiTokensFromDb, err := impl.apiTokenRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error while getting all active api tokens from DB", "error", err)
		return nil, err
	}

	var apiTokens []*openapi.ApiToken
	for _, apiTokenFromDb := range apiTokensFromDb {
		apiTokenIdI32 := int32(apiTokenFromDb.Id)
		lastUsedAtStr := apiTokenFromDb.LastUsedAt.String()
		updatedAtStr := apiTokenFromDb.UpdatedOn.String()
		apiToken := &openapi.ApiToken{
			Id:           &apiTokenIdI32,
			UserId:       &apiTokenFromDb.User.Id,
			Name:         &apiTokenFromDb.Name,
			Description:  &apiTokenFromDb.Description,
			ExpireAtInMs: &apiTokenFromDb.ExpireAtInMs,
			Token:        &apiTokenFromDb.Token,
			LastUsedAt:   &lastUsedAtStr,
			LastUsedByIp: &apiTokenFromDb.LastUsedByIp,
			UpdatedAt:    &updatedAtStr,
		}
		apiTokens = append(apiTokens, apiToken)
	}

	return apiTokens, nil
}

func (impl ApiTokenServiceImpl) CreateApiToken(request *openapi.CreateApiTokenRequest, createdBy int32) (*openapi.ActionResponse, error) {
	impl.logger.Infow("Creating API token", "request", request, "createdBy", createdBy)

	// step-1 - check if the name exists, if exists - throw error
	name := request.GetName()
	apiToken, err := impl.apiTokenRepository.FindByName(name)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while getting api token by name", "name", name, "error", err)
		return nil, err
	}
	if apiToken != nil && apiToken.Id > 0 {
		return nil, errors.New(fmt.Sprintf("name '%s' is already used. please use another name", name))
	}

	// step-2 - Create user using email
	// removing comma from email as user-service expects multiple emailIds separated from comma. since this is single email Id, hence removing comma
	emailSuffix := strings.ReplaceAll(name, ",", "")
	email := fmt.Sprintf("%s%s", API_TOKEN_USER_EMAIL_PREFIX, emailSuffix)
	createUserRequest := bean.UserInfo{
		EmailId: email,
	}

	createUserResponse, err := impl.userService.CreateUser(&createUserRequest)
	if err != nil {
		impl.logger.Errorw("error while creating user for api-token", "email", email, "error", err)
		return nil, err
	}
	createUserResponseLength := len(createUserResponse)
	if createUserResponseLength != 1 {
		return nil, errors.New(fmt.Sprintf("some error while creating user. length of createUserResponse expected 1. found %d", createUserResponseLength))
	}
	userId := createUserResponse[0].Id

	// step-3 - Create API token
	// create token using email
	token := uuid.New().String()
	apiTokenSaveRequest := &ApiToken{
		UserId:       userId,
		Name:         name,
		Description:  *request.Description,
		ExpireAtInMs: *request.ExpireAtInMs,
		Token:        token,
		AuditLog:     sql.AuditLog{CreatedOn: time.Now(), CreatedBy: createdBy, UpdatedOn: time.Now()},
	}
	err = impl.apiTokenRepository.Save(apiTokenSaveRequest)
	if err != nil {
		impl.logger.Errorw("error while saving api-token into DB", "error", err)
		return nil, err
	}

	success := true
	return &openapi.ActionResponse{
		Success: &success,
	}, nil
}

func (impl ApiTokenServiceImpl) UpdateApiToken(apiTokenId int, request *openapi.UpdateApiTokenRequest, updatedBy int32) (*openapi.ActionResponse, error) {
	impl.logger.Infow("Updating API token", "request", request, "updatedBy", updatedBy, "apiTokenId", apiTokenId)

	// step-1 - check if the api-token exists, if not exists - throw error
	apiToken, err := impl.apiTokenRepository.FindActiveById(apiTokenId)
	if err != nil && err != pg.ErrNoRows{
		impl.logger.Errorw("error while getting api token by id", "apiTokenId", apiTokenId, "error", err)
		return nil, err
	}
	if apiToken == nil || apiToken.Id == 0 {
		return nil, errors.New(fmt.Sprintf("api-token corresponds to apiTokenId '%d' is not found", apiTokenId))
	}

	// step-2 - If expires_at is not same, then token needs to be generated again
	if *request.ExpireAtInMs != apiToken.ExpireAtInMs {
		// regenerate token
		apiToken.Token = uuid.New().String()
	}

	// step-3 - update in DB
	apiToken.Description = *request.Description
	apiToken.ExpireAtInMs = *request.ExpireAtInMs
	apiToken.UpdatedBy = updatedBy
	apiToken.UpdatedOn = time.Now()
	err = impl.apiTokenRepository.Update(apiToken)
	if err != nil {
		impl.logger.Errorw("error while updating api-token", "apiTokenId", apiTokenId, "error", err)
		return nil, err
	}

	success := true
	return &openapi.ActionResponse{
		Success: &success,
	}, nil
}

func (impl ApiTokenServiceImpl) DeleteApiToken(apiTokenId int, deletedBy int32) (*openapi.ActionResponse, error) {
	impl.logger.Infow("Deleting API token", "deletedBy", deletedBy, "apiTokenId", apiTokenId)

	// step-1 - check if the api-token exists, if not exists - throw error
	apiToken, err := impl.apiTokenRepository.FindActiveById(apiTokenId)
	if err != nil && err != pg.ErrNoRows{
		impl.logger.Errorw("error while getting api token by id", "apiTokenId", apiTokenId, "error", err)
		return nil, err
	}
	if apiToken == nil || apiToken.Id == 0 {
		return nil, errors.New(fmt.Sprintf("api-token corresponds to apiTokenId '%d' is not found", apiTokenId))
	}

	// step-2 inactivate user corresponds to this api-token
	deleteUserRequest := bean.UserInfo{
		Id:     apiToken.UserId,
		UserId: deletedBy,
	}
	success, err := impl.userService.DeleteUser(&deleteUserRequest)
	if err != nil {
		impl.logger.Errorw("error while inactivating user for", "apiTokenId", apiTokenId, "userId", apiToken.UserId, "error", err)
		return nil, err
	}
	if !success {
		return nil, errors.New(fmt.Sprintf("Couldn't in-activate user corresponds to apiTokenId '%d'", apiTokenId))
	}

	return &openapi.ActionResponse{
		Success: &success,
	}, nil

}
