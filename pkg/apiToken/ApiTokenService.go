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
	"go.uber.org/zap"
)

type ApiTokenService interface {
	CreateApiToken(apiTokenDto *ApiTokenDto, createdBy int32) error
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

func (impl ApiTokenServiceImpl) CreateApiToken(apiTokenDto *ApiTokenDto, createdBy int32) error {

	// step-1 - check if the name exists in the DB, if yes - throw error

	// step-2 - create a token (using dex logic)

	// step-3 - save entry in user

	// step-4 - save entry in api_token

	return nil
}
