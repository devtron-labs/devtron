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
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"

	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/util"
	"go.uber.org/zap"
)

type DockerRegistryConfig interface {
	Create(bean *DockerArtifactStoreBean) (*DockerArtifactStoreBean, error)
	ListAllActive() ([]DockerArtifactStoreBean, error)
	FetchAllDockerAccounts() ([]DockerArtifactStoreBean, error)
	FetchOneDockerAccount(storeId string) (*DockerArtifactStoreBean, error)
	Update(bean *DockerArtifactStoreBean) (*DockerArtifactStoreBean, error)
	UpdateInactive(bean *DockerArtifactStoreBean) (*DockerArtifactStoreBean, error)
	Delete(storeId string) (string, error)
	DeleteReg(bean *DockerArtifactStoreBean) error
	CheckInActiveDockerAccount(storeId string) (bool, error)
}

type DockerArtifactStoreBean struct {
	Id                      string                       `json:"id,omitempty" validate:"required"`
	PluginId                string                       `json:"pluginId,omitempty" validate:"required"`
	RegistryURL             string                       `json:"registryUrl,omitempty"`
	RegistryType            repository.RegistryType      `json:"registryType,omitempty" validate:"required"`
	AWSAccessKeyId          string                       `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey      string                       `json:"awsSecretAccessKey,omitempty"`
	AWSRegion               string                       `json:"awsRegion,omitempty"`
	Username                string                       `json:"username,omitempty"`
	Password                string                       `json:"password,omitempty"`
	IsDefault               bool                         `json:"isDefault"`
	Connection              string                       `json:"connection"`
	Cert                    string                       `json:"cert"`
	Active                  bool                         `json:"active"`
	User                    int32                        `json:"-"`
	DockerRegistryIpsConfig *DockerRegistryIpsConfigBean `json:"ipsConfig,notnull,omitempty" validate:"required"`
}

type DockerRegistryIpsConfigBean struct {
	Id                   int                                        `json:"id"`
	CredentialType       repository.DockerRegistryIpsCredentialType `json:"credentialType,omitempty" validate:"oneof=SAME_AS_REGISTRY NAME CUSTOM_CREDENTIAL"`
	CredentialValue      string                                     `json:"credentialValue,omitempty"`
	AppliedClusterIdsCsv string                                     `json:"appliedClusterIdsCsv,omitempty"`
	IgnoredClusterIdsCsv string                                     `json:"ignoredClusterIdsCsv,omitempty"`
}

type DockerRegistryConfigImpl struct {
	logger                            *zap.SugaredLogger
	dockerArtifactStoreRepository     repository.DockerArtifactStoreRepository
	dockerRegistryIpsConfigRepository repository.DockerRegistryIpsConfigRepository
}

func NewDockerRegistryConfigImpl(logger *zap.SugaredLogger, dockerArtifactStoreRepository repository.DockerArtifactStoreRepository,
	dockerRegistryIpsConfigRepository repository.DockerRegistryIpsConfigRepository) *DockerRegistryConfigImpl {
	return &DockerRegistryConfigImpl{
		logger:                            logger,
		dockerArtifactStoreRepository:     dockerArtifactStoreRepository,
		dockerRegistryIpsConfigRepository: dockerRegistryIpsConfigRepository,
	}
}

func NewDockerArtifactStore(bean *DockerArtifactStoreBean, isActive bool, createdOn time.Time, updatedOn time.Time, createdBy int32, updateBy int32) *repository.DockerArtifactStore {
	return &repository.DockerArtifactStore{
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
		Connection:         bean.Connection,
		Cert:               bean.Cert,
		Active:             isActive,
		AuditLog:           sql.AuditLog{CreatedBy: createdBy, CreatedOn: createdOn, UpdatedOn: updatedOn, UpdatedBy: updateBy},
	}
}

func (impl DockerRegistryConfigImpl) Create(bean *DockerArtifactStoreBean) (*DockerArtifactStoreBean, error) {
	impl.logger.Debugw("docker registry create request", "request", bean)

	// 1- initiate DB transaction
	dbConnection := impl.dockerArtifactStoreRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in initiating db tx", "err", err)
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// 2- insert docker_registry_config
	store := NewDockerArtifactStore(bean, true, time.Now(), time.Now(), bean.User, bean.User)
	err = impl.dockerArtifactStoreRepository.Save(store, tx)
	if err != nil {
		impl.logger.Errorw("error in saving registry config", "config", store, "err", err)
		err = &util.ApiError{
			Code:            constants.DockerRegCreateFailedInDb,
			InternalMessage: "docker registry failed to create in db",
			UserMessage:     fmt.Sprintf("Container registry [%s] already exists.", bean.Id),
		}
		return nil, err
	}
	impl.logger.Infow("created repository ", "repository", store)
	bean.Id = store.Id

	// 3- insert imagePullSecretConfig for this docker registry
	dockerRegistryIpsConfig := bean.DockerRegistryIpsConfig
	ipsConfig := &repository.DockerRegistryIpsConfig{
		DockerArtifactStoreId: store.Id,
		CredentialType:        dockerRegistryIpsConfig.CredentialType,
		CredentialValue:       dockerRegistryIpsConfig.CredentialValue,
		AppliedClusterIdsCsv:  dockerRegistryIpsConfig.AppliedClusterIdsCsv,
		IgnoredClusterIdsCsv:  dockerRegistryIpsConfig.IgnoredClusterIdsCsv,
	}
	err = impl.dockerRegistryIpsConfigRepository.Save(ipsConfig, tx)
	if err != nil {
		impl.logger.Errorw("error in saving registry config ips", "ipsConfig", ipsConfig, "err", err)
		err = &util.ApiError{
			Code:            constants.DockerRegCreateFailedInDb,
			InternalMessage: "docker registry ips config to create in db",
			UserMessage:     fmt.Sprintf("Container registry [%s] already exists.", bean.Id),
		}
		return nil, err
	}
	impl.logger.Infow("created ips config for this docker repository ", "ipsConfig", ipsConfig)
	dockerRegistryIpsConfig.Id = ipsConfig.Id

	// 4- now commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return nil, err
	}

	return bean, nil
}

// list all active artifact store
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
			Id:           store.Id,
			RegistryURL:  store.RegistryURL,
			IsDefault:    store.IsDefault,
			RegistryType: store.RegistryType,
		}
		storeBeans = append(storeBeans, storeBean)
	}
	return storeBeans, err
}

/*
*
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
		ipsConfig := store.IpsConfig
		storeBean := DockerArtifactStoreBean{
			Id:                 store.Id,
			PluginId:           store.PluginId,
			RegistryURL:        store.RegistryURL,
			RegistryType:       store.RegistryType,
			AWSAccessKeyId:     store.AWSAccessKeyId,
			AWSSecretAccessKey: "",
			AWSRegion:          store.AWSRegion,
			Username:           store.Username,
			Password:           "",
			IsDefault:          store.IsDefault,
			Connection:         store.Connection,
			Cert:               store.Cert,
			Active:             store.Active,
			DockerRegistryIpsConfig: &DockerRegistryIpsConfigBean{
				Id:                   ipsConfig.Id,
				CredentialType:       ipsConfig.CredentialType,
				CredentialValue:      ipsConfig.CredentialValue,
				AppliedClusterIdsCsv: ipsConfig.AppliedClusterIdsCsv,
				IgnoredClusterIdsCsv: ipsConfig.IgnoredClusterIdsCsv,
			},
		}
		storeBeans = append(storeBeans, storeBean)
	}

	return storeBeans, err
}

/*
*
this method used for getting all the docker account details
*/
func (impl DockerRegistryConfigImpl) FetchOneDockerAccount(storeId string) (*DockerArtifactStoreBean, error) {
	impl.logger.Debug("fetch docker account by id from db")
	store, err := impl.dockerArtifactStoreRepository.FindOne(storeId)
	if err != nil {
		impl.logger.Errorw("error in listing artifact", "err", err)
		return nil, err
	}

	ipsConfig := store.IpsConfig
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
		Connection:         store.Connection,
		Cert:               store.Cert,
		Active:             store.Active,
		DockerRegistryIpsConfig: &DockerRegistryIpsConfigBean{
			Id:                   ipsConfig.Id,
			CredentialType:       ipsConfig.CredentialType,
			CredentialValue:      ipsConfig.CredentialValue,
			AppliedClusterIdsCsv: ipsConfig.AppliedClusterIdsCsv,
			IgnoredClusterIdsCsv: ipsConfig.IgnoredClusterIdsCsv,
		},
	}

	return storeBean, err
}

func (impl DockerRegistryConfigImpl) Update(bean *DockerArtifactStoreBean) (*DockerArtifactStoreBean, error) {
	impl.logger.Debugw("docker registry update request", "request", bean)

	// 1- find by id, if err - return error
	existingStore, err0 := impl.dockerArtifactStoreRepository.FindOne(bean.Id)
	if err0 != nil {
		impl.logger.Errorw("no matching entry found of update ..", "err", err0)
		return nil, err0
	}

	// 2- initiate DB transaction
	dbConnection := impl.dockerArtifactStoreRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in initiating db tx", "err", err)
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// 3- update docker_registry_config
	if bean.Password == "" {
		bean.Password = existingStore.Password
	}

	if bean.AWSSecretAccessKey == "" {
		bean.AWSSecretAccessKey = existingStore.AWSSecretAccessKey
	}

	if bean.Cert == "" {
		bean.Cert = existingStore.Cert
	}

	if bean.Cert == "" {
		bean.Cert = existingStore.Cert
	}

	bean.PluginId = existingStore.PluginId

	store := NewDockerArtifactStore(bean, true, existingStore.CreatedOn, time.Now(), existingStore.CreatedBy, bean.User)

	err = impl.dockerArtifactStoreRepository.Update(store, tx)
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

	// 4- update imagePullSecretConfig for this docker registry
	dockerRegistryIpsConfig := bean.DockerRegistryIpsConfig
	ipsConfig := &repository.DockerRegistryIpsConfig{
		Id:                    dockerRegistryIpsConfig.Id,
		DockerArtifactStoreId: store.Id,
		CredentialType:        dockerRegistryIpsConfig.CredentialType,
		CredentialValue:       dockerRegistryIpsConfig.CredentialValue,
		AppliedClusterIdsCsv:  dockerRegistryIpsConfig.AppliedClusterIdsCsv,
		IgnoredClusterIdsCsv:  dockerRegistryIpsConfig.IgnoredClusterIdsCsv,
	}
	err = impl.dockerRegistryIpsConfigRepository.Update(ipsConfig, tx)
	if err != nil {
		impl.logger.Errorw("error in updating registry config ips", "ipsConfig", ipsConfig, "err", err)
		err = &util.ApiError{
			Code:            constants.DockerRegUpdateFailedInDb,
			InternalMessage: "docker registry ips config failed to update in db",
			UserMessage:     "docker registry ips config failed to update in db",
		}
		return nil, err
	}
	impl.logger.Infow("updated ips config for this docker repository ", "ipsConfig", ipsConfig)

	// 5- now commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return nil, err
	}

	return bean, nil
}

func (impl DockerRegistryConfigImpl) UpdateInactive(bean *DockerArtifactStoreBean) (*DockerArtifactStoreBean, error) {
	impl.logger.Debugw("docker registry update request", "request", bean)

	// 1- find by id, if err - return error
	existingStore, err0 := impl.dockerArtifactStoreRepository.FindOneInactive(bean.Id)
	if err0 != nil {
		impl.logger.Errorw("no matching entry found of update ..", "err", err0)
		return nil, err0
	}

	// 2- initiate DB transaction
	dbConnection := impl.dockerArtifactStoreRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in initiating db tx", "err", err)
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// 3- update docker_registry_config

	bean.PluginId = existingStore.PluginId

	store := NewDockerArtifactStore(bean, true, existingStore.CreatedOn, time.Now(), bean.User, bean.User)

	err = impl.dockerArtifactStoreRepository.Update(store, tx)
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

	// 4- update imagePullSecretConfig for this docker registry
	dockerRegistryIpsConfig := bean.DockerRegistryIpsConfig
	ipsConfig := &repository.DockerRegistryIpsConfig{
		Id:                    existingStore.IpsConfig.Id,
		DockerArtifactStoreId: store.Id,
		CredentialType:        dockerRegistryIpsConfig.CredentialType,
		CredentialValue:       dockerRegistryIpsConfig.CredentialValue,
		AppliedClusterIdsCsv:  dockerRegistryIpsConfig.AppliedClusterIdsCsv,
		IgnoredClusterIdsCsv:  dockerRegistryIpsConfig.IgnoredClusterIdsCsv,
	}
	err = impl.dockerRegistryIpsConfigRepository.Update(ipsConfig, tx)
	if err != nil {
		impl.logger.Errorw("error in updating registry config ips", "ipsConfig", ipsConfig, "err", err)
		err = &util.ApiError{
			Code:            constants.DockerRegUpdateFailedInDb,
			InternalMessage: "docker registry ips config failed to update in db",
			UserMessage:     "docker registry ips config failed to update in db",
		}
		return nil, err
	}
	impl.logger.Infow("updated ips config for this docker repository ", "ipsConfig", ipsConfig)

	// 5- now commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return nil, err
	}

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

func (impl DockerRegistryConfigImpl) DeleteReg(bean *DockerArtifactStoreBean) error {
	dockerReg, err := impl.dockerArtifactStoreRepository.FindOne(bean.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete.", "id", bean.Id, "err", err)
		return err
	}
	deleteReq := dockerReg
	deleteReq.UpdatedOn = time.Now()
	deleteReq.UpdatedBy = bean.User
	err = impl.dockerArtifactStoreRepository.MarkRegistryDeleted(deleteReq)
	if err != nil {
		impl.logger.Errorw("err in deleting docker registry", "id", bean.Id, "err", err)
		return err
	}
	return nil
}

func (impl DockerRegistryConfigImpl) CheckInActiveDockerAccount(storeId string) (bool, error) {
	exist, err := impl.dockerArtifactStoreRepository.FindInactive(storeId)
	if err != nil {
		impl.logger.Errorw("err in deleting docker registry", "id", storeId, "err", err)
		return false, err
	}
	return exist, nil
}
