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
	openapi "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"go.uber.org/zap"
)

type ApiTokenService interface {
	GetAllActiveApiTokens() ([]*openapi.ApiToken, error)
	CreateApiToken(request *openapi.CreateApiTokenRequest, createdBy int32) (*openapi.ActionResponse, error)
	UpdateApiToken(apiTokenId int, request *openapi.UpdateApiTokenRequest, updatedBy int32) (*openapi.ActionResponse, error)
	DeleteApiToken(apiTokenId int, deletedBy int32) (*openapi.ActionResponse, error)
}

type ApiTokenServiceImpl struct {
	logger             *zap.SugaredLogger
	apiTokenRepository ApiTokenRepository
}

func NewApiTokenServiceImpl(logger *zap.SugaredLogger, apiTokenRepository ApiTokenRepository) *ApiTokenServiceImpl {
	return &ApiTokenServiceImpl{
		logger:             logger,
		apiTokenRepository: apiTokenRepository,
	}
}

func (impl ApiTokenServiceImpl) GetAllActiveApiTokens() ([]*openapi.ApiToken, error){
	return nil, nil
}

func (impl ApiTokenServiceImpl) CreateApiToken(request *openapi.CreateApiTokenRequest, createdBy int32) (*openapi.ActionResponse, error){
	return nil, nil
}

func (impl ApiTokenServiceImpl) UpdateApiToken(apiTokenId int, request *openapi.UpdateApiTokenRequest, updatedBy int32) (*openapi.ActionResponse, error){
	return nil, nil
}

func (impl ApiTokenServiceImpl) DeleteApiToken(apiTokenId int, deletedBy int32) (*openapi.ActionResponse, error){
	return nil, nil
}
