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
	apiTokenAuth "github.com/devtron-labs/authenticator/apiToken"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"go.uber.org/zap"
)

type ApiTokenSecretService interface {
	GetApiTokenSecretByteArr() ([]byte, error)
}

type ApiTokenSecretServiceImpl struct {
	logger              *zap.SugaredLogger
	attributesService   attributes.AttributesService
	apiTokenSecretStore *apiTokenAuth.ApiTokenSecretStore
}

func NewApiTokenSecretServiceImpl(logger *zap.SugaredLogger, attributesService attributes.AttributesService, apiTokenSecretStore *apiTokenAuth.ApiTokenSecretStore) (*ApiTokenSecretServiceImpl, error) {
	impl := &ApiTokenSecretServiceImpl{
		logger:                   logger,
		attributesService: attributesService,
		apiTokenSecretStore:      apiTokenSecretStore,
	}

	// get secret from db and store
	secret, err := impl.getApiSecretFromDb()
	if err != nil {
		return nil, err
	}
	impl.apiTokenSecretStore.Secret = secret

	return impl, nil
}

func (impl ApiTokenSecretServiceImpl) GetApiTokenSecretByteArr() ([]byte, error) {
	impl.logger.Info("Getting api token secret")

	// return from local
	// if found empty, throw error
	if len(impl.apiTokenSecretStore.Secret) == 0 {
		errorMsg := "secret found empty"
		impl.logger.Error(errorMsg)
		return nil, errors.New(errorMsg)
	}

	return []byte(impl.apiTokenSecretStore.Secret), nil
}

func (impl ApiTokenSecretServiceImpl) getApiSecretFromDb() (string, error) {
	// get from db
	apiTokenSecret, err := impl.attributesService.GetByKey(attributes.API_SECRET_KEY)
	if err != nil {
		impl.logger.Errorw("error while getting api token secret from DB", "error", err)
		return "", err
	}
	if apiTokenSecret == nil {
		errorMsg := "api token secret from DB found nil"
		impl.logger.Error(errorMsg)
		return "", errors.New(errorMsg)
	}
	secret := apiTokenSecret.Value
	if len(secret) == 0 {
		errorMsg := "api token secret from DB found empty"
		impl.logger.Error(errorMsg)
		return "", errors.New(errorMsg)
	}

	return secret, nil
}
