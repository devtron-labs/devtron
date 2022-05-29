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
	"go.uber.org/zap"
)

type ApiTokenSecretService interface {
	GetApiTokenSecret() (string, error)
}

type ApiTokenSecretServiceImpl struct {
	logger                   *zap.SugaredLogger
	apiTokenSecretRepository ApiTokenSecretRepository
	apiTokenSecretStore      *ApiTokenSecretStore
}

func NewApiTokenSecretServiceImpl(logger *zap.SugaredLogger, apiTokenSecretRepository ApiTokenSecretRepository, apiTokenSecretStore *ApiTokenSecretStore) *ApiTokenSecretServiceImpl {
	return &ApiTokenSecretServiceImpl{
		logger:                   logger,
		apiTokenSecretRepository: apiTokenSecretRepository,
		apiTokenSecretStore:      apiTokenSecretStore,
	}
}

func (impl ApiTokenSecretServiceImpl) GetApiTokenSecret() (string, error) {
	impl.logger.Info("Getting api token secret")

	// return from local if found
	if len(impl.apiTokenSecretStore.Secret) > 0 {
		return impl.apiTokenSecretStore.Secret, nil
	}

	// get from db
	apiTokenSecret, err := impl.apiTokenSecretRepository.Get()
	if err != nil {
		impl.logger.Errorw("error while getting api token secret from DB", "error", err)
		return "", err
	}
	if apiTokenSecret == nil {
		error := "api token secret from DB found nil"
		impl.logger.Error(error)
		return "", errors.New(error)
	}

	// set locally
	secret := apiTokenSecret.Secret
	impl.apiTokenSecretStore.Secret = secret
	return secret, nil
}
