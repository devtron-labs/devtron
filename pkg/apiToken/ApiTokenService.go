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
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/devtron/api/bean"
	openapi "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/pkg/notifier"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/go-pg/pg"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ApiTokenService interface {
	GetAllActiveApiTokens() ([]*openapi.ApiToken, error)
	CreateApiToken(request *openapi.CreateApiTokenRequest, createdBy int32, managerAuth func(resource, token, object string) bool) (*openapi.CreateApiTokenResponse, error)
	UpdateApiToken(apiTokenId int, request *openapi.UpdateApiTokenRequest, updatedBy int32) (*openapi.UpdateApiTokenResponse, error)
	DeleteApiToken(apiTokenId int, deletedBy int32) (*openapi.ActionResponse, error)
	GetAllApiTokensForWebhook(projectName string, environmentName string, appName string, auth func(token string, projectObject string, envObject string) bool) ([]*openapi.ApiToken, error)
	CreateApiJwtTokenForNotification(draftApprovalRequest notifier.DraftApprovalRequest, ImageRequest *notifier.ImageApprovalRequest, expireAtInMs int64) (string, error)
}

type ApiTokenServiceImpl struct {
	logger                *zap.SugaredLogger
	apiTokenSecretService ApiTokenSecretService
	userService           user.UserService
	userAuditService      user.UserAuditService
	apiTokenRepository    ApiTokenRepository
}

func NewApiTokenServiceImpl(logger *zap.SugaredLogger, apiTokenSecretService ApiTokenSecretService, userService user.UserService, userAuditService user.UserAuditService,
	apiTokenRepository ApiTokenRepository) *ApiTokenServiceImpl {
	return &ApiTokenServiceImpl{
		logger:                logger,
		apiTokenSecretService: apiTokenSecretService,
		userService:           userService,
		userAuditService:      userAuditService,
		apiTokenRepository:    apiTokenRepository,
	}
}

const API_TOKEN_USER_EMAIL_PREFIX = "API-TOKEN:"

var invalidCharsInApiTokenName = regexp.MustCompile("[,\\s]")

type ApiTokenCustomClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}
type TokenCustomClaimsForNotification struct {
	DraftId           int    `json:"draftId"`
	DraftVersionId    int    `json:"draftVersionId"`
	ApprovalRequestId int    `json:"approvalRequestId"`
	ArtifactId        int    `json:"artifactId"`
	AppId             int    `json:"appId" validate:"required"`
	EnvId             int    `json:"envId"`
	ApprovalType      string `json:"approvalType"`
	ApiTokenCustomClaims
}

func (tokenCustomClaimsForNotification TokenCustomClaimsForNotification) generateToken(secretByteArr []byte) (string, error) {
	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenCustomClaimsForNotification)
	token, err := unsignedToken.SignedString(secretByteArr)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (impl ApiTokenServiceImpl) GetAllApiTokensForWebhook(projectName string, environmentName string, appName string, auth func(token string, projectObject string, envObject string) bool) ([]*openapi.ApiToken, error) {
	impl.logger.Info("Getting active api tokens")
	apiTokensFromDb, err := impl.apiTokenRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error while getting all active api tokens from DB", "error", err)
		return nil, err
	}

	apiTokens := make([]*openapi.ApiToken, 0)
	for _, apiTokenFromDb := range apiTokensFromDb {
		authPassed := true
		userId := apiTokenFromDb.User.Id
		//checking permission on each of the roles associated with this API Token
		environmentNames := strings.Split(environmentName, ",")
		for _, environment := range environmentNames {
			projectObject := fmt.Sprintf("%s/%s", projectName, appName)
			envObject := fmt.Sprintf("%s/%s", environment, appName)
			isValidAuth := auth(apiTokenFromDb.Token, projectObject, envObject)
			if !isValidAuth {
				impl.logger.Debugw("authentication for token failed", "apiTokenFromDb", apiTokenFromDb)
				authPassed = false
				continue
			}
		}

		if authPassed {
			apiTokenIdI32 := int32(apiTokenFromDb.Id)
			updatedAtStr := apiTokenFromDb.UpdatedOn.String()
			apiToken := &openapi.ApiToken{
				Id:             &apiTokenIdI32,
				UserId:         &userId,
				UserIdentifier: &apiTokenFromDb.User.EmailId,
				Name:           &apiTokenFromDb.Name,
				Description:    &apiTokenFromDb.Description,
				ExpireAtInMs:   &apiTokenFromDb.ExpireAtInMs,
				Token:          &apiTokenFromDb.Token,
				UpdatedAt:      &updatedAtStr,
			}
			apiTokens = append(apiTokens, apiToken)
		}
	}

	return apiTokens, nil
}

func (impl ApiTokenServiceImpl) GetAllActiveApiTokens() ([]*openapi.ApiToken, error) {
	impl.logger.Info("Getting all active api tokens")
	apiTokensFromDb, err := impl.apiTokenRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error while getting all active api tokens from DB", "error", err)
		return nil, err
	}

	var apiTokens []*openapi.ApiToken
	for _, apiTokenFromDb := range apiTokensFromDb {
		userId := apiTokenFromDb.User.Id
		latestAuditLog, err := impl.userAuditService.GetLatestByUserId(userId)
		if err != nil {
			impl.logger.Errorw("error while getting latest audit log", "error", err)
			return nil, err
		}

		apiTokenIdI32 := int32(apiTokenFromDb.Id)
		updatedAtStr := apiTokenFromDb.UpdatedOn.String()
		apiToken := &openapi.ApiToken{
			Id:             &apiTokenIdI32,
			UserId:         &userId,
			UserIdentifier: &apiTokenFromDb.User.EmailId,
			Name:           &apiTokenFromDb.Name,
			Description:    &apiTokenFromDb.Description,
			ExpireAtInMs:   &apiTokenFromDb.ExpireAtInMs,
			Token:          &apiTokenFromDb.Token,
			UpdatedAt:      &updatedAtStr,
		}
		if latestAuditLog != nil {
			lastUsedAtStr := latestAuditLog.CreatedOn.String()
			apiToken.LastUsedAt = &lastUsedAtStr
			apiToken.LastUsedByIp = &latestAuditLog.ClientIp
		}
		apiTokens = append(apiTokens, apiToken)
	}

	return apiTokens, nil
}

func (impl ApiTokenServiceImpl) CreateApiToken(request *openapi.CreateApiTokenRequest, createdBy int32, managerAuth func(resource, token string, object string) bool) (*openapi.CreateApiTokenResponse, error) {
	impl.logger.Infow("Creating API token", "request", request, "createdBy", createdBy)

	name := request.GetName()
	// check if name contains some characters which are not allowed
	if invalidCharsInApiTokenName.MatchString(name) {
		return nil, errors.New(fmt.Sprintf("name '%s' contains either white-space or comma, which is not allowed", name))
	}

	// step-1 - check if the name exists, if exists with active user - throw error
	apiToken, err := impl.apiTokenRepository.FindByName(name)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while getting api token by name", "name", name, "error", err)
		return nil, err
	}
	var apiTokenExists bool
	if apiToken != nil && apiToken.Id > 0 {
		apiTokenExists = true
		if apiToken.User.Active {
			return nil, errors.New(fmt.Sprintf("name '%s' is already used. please use another name", name))
		}
	}

	impl.logger.Info(fmt.Sprintf("apiTokenExists : %s", strconv.FormatBool(apiTokenExists)))

	// step-2 - Build email
	email := fmt.Sprintf("%s%s", API_TOKEN_USER_EMAIL_PREFIX, name)

	// step-3 - Build token
	token, err := impl.createApiJwtToken(email, *request.ExpireAtInMs)
	if err != nil {
		return nil, err
	}

	// step-4 - Create user using email
	createUserRequest := bean.UserInfo{
		UserId:   createdBy,
		EmailId:  email,
		UserType: bean.USER_TYPE_API_TOKEN,
	}
	createUserResponse, err := impl.userService.CreateUser(&createUserRequest, token, managerAuth)
	if err != nil {
		impl.logger.Errorw("error while creating user for api-token", "email", email, "error", err)
		return nil, err
	}
	createUserResponseLength := len(createUserResponse)
	if createUserResponseLength != 1 {
		return nil, errors.New(fmt.Sprintf("some error while creating user. length of createUserResponse expected 1. found %d", createUserResponseLength))
	}
	userId := createUserResponse[0].Id

	// step-5 - Save API token (update or save)
	apiTokenSaveRequest := &ApiToken{
		UserId:       userId,
		Name:         name,
		Description:  *request.Description,
		ExpireAtInMs: *request.ExpireAtInMs,
		Token:        token,
		AuditLog:     sql.AuditLog{UpdatedOn: time.Now()},
	}
	if apiTokenExists {
		apiTokenSaveRequest.Id = apiToken.Id
		apiTokenSaveRequest.CreatedBy = apiToken.CreatedBy
		apiTokenSaveRequest.CreatedOn = apiToken.CreatedOn
		apiTokenSaveRequest.UpdatedBy = createdBy
		err = impl.apiTokenRepository.Update(apiTokenSaveRequest)
	} else {
		apiTokenSaveRequest.CreatedBy = createdBy
		apiTokenSaveRequest.CreatedOn = time.Now()
		err = impl.apiTokenRepository.Save(apiTokenSaveRequest)
	}
	if err != nil {
		impl.logger.Errorw("error while saving api-token into DB", "error", err)
		return nil, err
	}

	success := true
	return &openapi.CreateApiTokenResponse{
		Success:        &success,
		Token:          &token,
		UserId:         &userId,
		UserIdentifier: &email,
	}, nil
}

func (impl ApiTokenServiceImpl) UpdateApiToken(apiTokenId int, request *openapi.UpdateApiTokenRequest, updatedBy int32) (*openapi.UpdateApiTokenResponse, error) {
	impl.logger.Infow("Updating API token", "request", request, "updatedBy", updatedBy, "apiTokenId", apiTokenId)

	// step-1 - check if the api-token exists, if not exists - throw error
	apiToken, err := impl.apiTokenRepository.FindActiveById(apiTokenId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while getting api token by id", "apiTokenId", apiTokenId, "error", err)
		return nil, err
	}
	if apiToken == nil || apiToken.Id == 0 {
		return nil, errors.New(fmt.Sprintf("api-token corresponds to apiTokenId '%d' is not found", apiTokenId))
	}

	// step-2 - If expires_at is not same, then token needs to be generated again
	if *request.ExpireAtInMs != apiToken.ExpireAtInMs {
		// regenerate token
		token, err := impl.createApiJwtToken(apiToken.User.EmailId, *request.ExpireAtInMs)
		if err != nil {
			return nil, err
		}
		apiToken.Token = token
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
	return &openapi.UpdateApiTokenResponse{
		Success: &success,
		Token:   &apiToken.Token,
	}, nil
}

func (impl ApiTokenServiceImpl) DeleteApiToken(apiTokenId int, deletedBy int32) (*openapi.ActionResponse, error) {
	impl.logger.Infow("Deleting API token", "deletedBy", deletedBy, "apiTokenId", apiTokenId)

	// step-1 - check if the api-token exists, if not exists - throw error
	apiToken, err := impl.apiTokenRepository.FindActiveById(apiTokenId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while getting api token by id", "apiTokenId", apiTokenId, "error", err)
		return nil, err
	}
	if apiToken == nil || apiToken.Id == 0 {
		return nil, errors.New(fmt.Sprintf("api-token corresponds to apiTokenId '%d' is not found", apiTokenId))
	}

	apiToken.ExpireAtInMs = time.Now().UnixMilli()
	err = impl.apiTokenRepository.Update(apiToken)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while getting api token by id", "apiTokenId", apiTokenId, "error", err)
		return nil, err
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
func (impl ApiTokenServiceImpl) CreateApiJwtTokenForNotification(draftRequest notifier.DraftApprovalRequest, ImageRequest *notifier.ImageApprovalRequest, expireAtInMs int64) (string, error) {
	registeredClaims, secretByteArr, err := impl.setRegisteredClaims(expireAtInMs)
	if err != nil {
		return "", err
	}
	claims := &TokenCustomClaimsForNotification{
		DraftId:        draftRequest.DraftId,
		DraftVersionId: draftRequest.DraftVersionId,
		AppId:          draftRequest.NotificationApprovalRequest.AppId,
		EnvId:          draftRequest.NotificationApprovalRequest.EnvId,
		ApiTokenCustomClaims: ApiTokenCustomClaims{
			Email:            draftRequest.NotificationApprovalRequest.EmailId,
			RegisteredClaims: registeredClaims,
		},
	}
	token, err := claims.generateToken(secretByteArr)
	if err != nil {
		//todo put log here
		return "", err
	}
	return token, nil

}

func (impl ApiTokenServiceImpl) createApiJwtToken(email string, expireAtInMs int64) (string, error) {
	registeredClaims, secretByteArr, err := impl.setRegisteredClaims(expireAtInMs)
	if err != nil {
		return "", err
	}
	claims := &ApiTokenCustomClaims{
		email,
		registeredClaims,
	}
	token, err := impl.generateToken(claims, secretByteArr)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (impl ApiTokenServiceImpl) setRegisteredClaims(expireAtInMs int64) (jwt.RegisteredClaims, []byte, error) {
	secretByteArr, err := impl.apiTokenSecretService.GetApiTokenSecretByteArr()
	if err != nil {
		impl.logger.Errorw("error while getting api token secret", "error", err)
		return jwt.RegisteredClaims{}, secretByteArr, err
	}

	registeredClaims := jwt.RegisteredClaims{
		Issuer: middleware.ApiTokenClaimIssuer,
	}

	if expireAtInMs > 0 {
		registeredClaims.ExpiresAt = jwt.NewNumericDate(time.Unix(expireAtInMs/1000, 0))
	}
	return registeredClaims, secretByteArr, nil
}

func (impl ApiTokenServiceImpl) generateToken(claims *ApiTokenCustomClaims, secretByteArr []byte) (string, error) {
	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := unsignedToken.SignedString(secretByteArr)
	if err != nil {
		impl.logger.Errorw("error while signing api-token", "error", err)
		return "", err
	}
	return token, nil
}
