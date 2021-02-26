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

package pipeline

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"go.uber.org/zap"
	"time"
)

type DockerRegistryConfig interface {
	Create(bean *DockerArtifactStoreBean) (*DockerArtifactStoreBean, error)
	ListAllActive() ([]DockerArtifactStoreBean, error)
	FetchAllDockerAccounts() ([]DockerArtifactStoreBean, error)
	FetchOneDockerAccount(storeId string) (*DockerArtifactStoreBean, error)
	Update(bean *DockerArtifactStoreBean) (*DockerArtifactStoreBean, error)
	Delete(storeId string) (string, error)
}

type DockerArtifactStoreBean struct {
	Id                 string                  `json:"id,omitempty" validate:"required"`
	PluginId           string                  `json:"pluginId,omitempty" validate:"required"`
	RegistryURL        string                  `json:"registryUrl,omitempty"`
	RegistryType       repository.RegistryType `json:"registryType,omitempty" validate:"required"`
	AWSAccessKeyId     string                  `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey string                  `json:"awsSecretAccessKey,omitempty"`
	AWSRegion          string                  `json:"awsRegion,omitempty"`
	Username           string                  `json:"username,omitempty"`
	Password           string                  `json:"password,omitempty"`
	IsDefault          bool                    `json:"isDefault"`
	Active             bool                    `json:"active"`
	User               int32                   `json:"-"`
}

type DockerRegistryConfigImpl struct {
	dockerArtifactStoreRepository repository.DockerArtifactStoreRepository
	logger                        *zap.SugaredLogger
}

func NewDockerRegistryConfigImpl(dockerArtifactStoreRepository repository.DockerArtifactStoreRepository, logger *zap.SugaredLogger) *DockerRegistryConfigImpl {
	return &DockerRegistryConfigImpl{
		dockerArtifactStoreRepository: dockerArtifactStoreRepository,
		logger:                        logger,
	}
}

func (impl DockerRegistryConfigImpl) Create(bean *DockerArtifactStoreBean) (*DockerArtifactStoreBean, error) {
	impl.logger.Debugw("docker registry create request", "request", bean)
	store := &repository.DockerArtifactStore{
		Id:                 bean.Id,
		PluginId:           bean.PluginId,
		RegistryURL:        bean.RegistryURL,
		RegistryType:       bean.RegistryType,
		AWSAccessKeyId:     bean.AWSAccessKeyId,
		AWSSecretAccessKey: bean.AWSSecretAccessKey,
		AWSRegion:          bean.AWSRegion,
		Username:           bean.Username,
		Password:           bean.Password,
		IsDefault:          bean.IsDefault,
		Active:             true,
		AuditLog:           models.AuditLog{CreatedBy: bean.User, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: bean.User},
	}
	err := impl.dockerArtifactStoreRepository.Save(store)
	if err != nil {
		impl.logger.Errorw("error in saving registry config", "config", store, "err", err)
		err = &util.ApiError{
			Code:            constants.DockerRegCreateFailedInDb,
			InternalMessage: "docker registry failed to create in db",
			UserMessage:     fmt.Sprintf("requested by %d", bean.User),
		}
		return nil, err
	}
	impl.logger.Infow("created repository ", "repository", store)
	bean.Id = store.Id
	return bean, nil
}

//list all active artifact store
func (impl DockerRegistryConfigImpl) ListAllActive() ([]DockerArtifactStoreBean, error) {
	impl.logger.Debug("list docker repo request")
	stores, err := impl.dockerArtifactStoreRepository.FindAllActiveForAutocomplete()
	if err != nil {
		impl.logger.Errorw("error in listing artifact", "err", err)
		return nil, err
	}
	var storeBeans []DockerArtifactStoreBean
	for _, store := range stores {
		storeBean := DockerArtifactStoreBean{
			Id:          store.Id,
			RegistryURL: store.RegistryURL,
			IsDefault:   store.IsDefault,
		}
		storeBeans = append(storeBeans, storeBean)
	}
	return storeBeans, err
}

/**
this method used for getting all the docker account details
*/
func (impl DockerRegistryConfigImpl) FetchAllDockerAccounts() ([]DockerArtifactStoreBean, error) {
	impl.logger.Debug("list docker repo request")
	stores, err := impl.dockerArtifactStoreRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("error in listing artifact", "err", err)
		return nil, err
	}
	var storeBeans []DockerArtifactStoreBean
	for _, store := range stores {
		storeBean := DockerArtifactStoreBean{
			Id:                 store.Id,
			PluginId:           store.PluginId,
			RegistryURL:        store.RegistryURL,
			RegistryType:       store.RegistryType,
			AWSAccessKeyId:     store.AWSAccessKeyId,
			AWSSecretAccessKey: store.AWSSecretAccessKey,
			AWSRegion:          store.AWSRegion,
			Username:           store.Username,
			Password:           store.Password,
			IsDefault:          store.IsDefault,
			Active:             store.Active,
		}
		storeBeans = append(storeBeans, storeBean)
	}

	return storeBeans, err
}

/**
this method used for getting all the docker account details
*/
func (impl DockerRegistryConfigImpl) FetchOneDockerAccount(storeId string) (*DockerArtifactStoreBean, error) {
	impl.logger.Debug("fetch docker account by id from db")
	store, err := impl.dockerArtifactStoreRepository.FindOne(storeId)
	if err != nil {
		impl.logger.Errorw("error in listing artifact", "err", err)
		return nil, err
	}

	storeBean := &DockerArtifactStoreBean{
		Id:                 store.Id,
		PluginId:           store.PluginId,
		RegistryURL:        store.RegistryURL,
		RegistryType:       store.RegistryType,
		AWSAccessKeyId:     store.AWSAccessKeyId,
		AWSSecretAccessKey: store.AWSSecretAccessKey,
		AWSRegion:          store.AWSRegion,
		Username:           store.Username,
		Password:           store.Password,
		IsDefault:          store.IsDefault,
		Active:             store.Active,
	}

	return storeBean, err
}

func (impl DockerRegistryConfigImpl) Update(bean *DockerArtifactStoreBean) (*DockerArtifactStoreBean, error) {
	impl.logger.Debugw("docker registry update request", "request", bean)
	existingStore, err0 := impl.dockerArtifactStoreRepository.FindOne(bean.Id)
	if err0 != nil {
		impl.logger.Errorw("no matching entry found of update ..", "err", err0)
		return nil, err0
	}
	store := &repository.DockerArtifactStore{
		Id:                 bean.Id,
		PluginId:           existingStore.PluginId,
		RegistryURL:        bean.RegistryURL,
		RegistryType:       bean.RegistryType,
		AWSAccessKeyId:     bean.AWSAccessKeyId,
		AWSSecretAccessKey: bean.AWSSecretAccessKey,
		AWSRegion:          bean.AWSRegion,
		Username:           bean.Username,
		Password:           bean.Password,
		IsDefault:          bean.IsDefault,
		Active:             true, // later it will change
		AuditLog:           models.AuditLog{CreatedBy: existingStore.CreatedBy, CreatedOn: existingStore.CreatedOn, UpdatedOn: time.Now(), UpdatedBy: bean.User},
	}
	err := impl.dockerArtifactStoreRepository.Update(store)
	if err != nil {
		impl.logger.Errorw("error in updating registry config in db", "config", store, "err", err)
		err = &util.ApiError{
			Code:            constants.DockerRegUpdateFailedInDb,
			InternalMessage: "docker registry failed to update in db",
			UserMessage:     "docker registry failed to update in db",
		}
		return nil, err
	}

	impl.logger.Infow("updated repository ", "repository", store)
	bean.Id = store.Id
	return bean, nil
}

func (impl DockerRegistryConfigImpl) Delete(storeId string) (string, error) {
	impl.logger.Debugw("docker registry update request", "request", storeId)

	err := impl.dockerArtifactStoreRepository.Delete(storeId)
	if err != nil {
		impl.logger.Errorw("error in delete registry in db", "storeId", storeId, "err", err)
		err = &util.ApiError{
			Code:            constants.DockerRegDeleteFailedInDb,
			InternalMessage: "docker registry failed to delete in db",
			UserMessage:     err.Error(),
		}
		return "", err
	}

	impl.logger.Infow("delete docker registry ", "storeId", storeId)
	return storeId, nil
}
