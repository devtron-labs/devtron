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
	"github.com/go-pg/pg"
	"k8s.io/utils/strings/slices"
	"strings"
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
	ConfigureOCIRegistry(bean *DockerArtifactStoreBean, isUpdate bool, userId int32, tx *pg.Tx) error
	CreateOrUpdateOCIRegistryConfig(ociRegistryConfig *repository.OCIRegistryConfig, userId int32, tx *pg.Tx) error
	FilterOCIRegistryConfigForSpecificRepoType(ociRegistryConfigList []*repository.OCIRegistryConfig, repositoryType string) *repository.OCIRegistryConfig
	FilterRegistryBeanListBasedOnStorageTypeAndAction(bean []DockerArtifactStoreBean, storageType string, actionTypes ...string) []DockerArtifactStoreBean
	ValidateRegistryStorageType(registryId string, storageType string, storageActions ...string) bool
}

type ArtifactStoreType int

type DockerArtifactStoreBean struct {
	Id                      string                       `json:"id,omitempty" validate:"required"`
	PluginId                string                       `json:"pluginId,omitempty" validate:"required"`
	RegistryURL             string                       `json:"registryUrl,omitempty"`
	RegistryType            repository.RegistryType      `json:"registryType,omitempty" validate:"required"`
	IsOCICompliantRegistry  bool                         `json:"isOCICompliantRegistry"`
	OCIRegistryConfig       map[string]string            `json:"ociRegistryConfig,omitempty"`
	IsPublic                bool                         `json:"isPublic,omitempty"`
	RepositoryList          []string                     `json:"repositoryList,omitempty"`
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
	ociRegistryConfigRepository       repository.OCIRegistryConfigRepository
}

func NewDockerRegistryConfigImpl(logger *zap.SugaredLogger, dockerArtifactStoreRepository repository.DockerArtifactStoreRepository,
	dockerRegistryIpsConfigRepository repository.DockerRegistryIpsConfigRepository, ociRegistryConfigRepository repository.OCIRegistryConfigRepository) *DockerRegistryConfigImpl {
	return &DockerRegistryConfigImpl{
		logger:                            logger,
		dockerArtifactStoreRepository:     dockerArtifactStoreRepository,
		dockerRegistryIpsConfigRepository: dockerRegistryIpsConfigRepository,
		ociRegistryConfigRepository:       ociRegistryConfigRepository,
	}
}

func NewDockerArtifactStore(bean *DockerArtifactStoreBean, isActive bool, createdOn time.Time, updatedOn time.Time, createdBy int32, updateBy int32) *repository.DockerArtifactStore {
	return &repository.DockerArtifactStore{
		Id:                     bean.Id,
		PluginId:               bean.PluginId,
		RegistryURL:            bean.RegistryURL,
		RegistryType:           bean.RegistryType,
		IsOCICompliantRegistry: bean.IsOCICompliantRegistry,
		AWSAccessKeyId:         bean.AWSAccessKeyId,
		AWSSecretAccessKey:     bean.AWSSecretAccessKey,
		AWSRegion:              bean.AWSRegion,
		Username:               bean.Username,
		Password:               bean.Password,
		IsDefault:              bean.IsDefault,
		Connection:             bean.Connection,
		Cert:                   bean.Cert,
		Active:                 isActive,
		AuditLog:               sql.AuditLog{CreatedBy: createdBy, CreatedOn: createdOn, UpdatedOn: updatedOn, UpdatedBy: updateBy},
	}
}

/*
	ValidateRegistryStorageType

Parameters:
  - registryId,
  - storageType:
    it can be any from repository.OCI_REGISRTY_REPO_TYPE_LIST
  - list of storageActions:
    it can be PULL, PUSH, PULL/PUSH

Logic:
  - Fetch registry config for the given registryId and validate if the storageType has any of the given storageActions

Returns:
  - isValid bool
*/
func (impl DockerRegistryConfigImpl) ValidateRegistryStorageType(registryId string, storageType string, storageActions ...string) bool {
	isValid := false
	store, err := impl.dockerArtifactStoreRepository.FindOne(registryId)
	if err != nil {
		return false
	}
	if store.IsOCICompliantRegistry {
		for _, ociRegistryConfig := range store.OCIRegistryConfig {
			if ociRegistryConfig.RepositoryType == storageType && slices.Contains(storageActions, ociRegistryConfig.RepositoryAction) {
				isValid = true
			}
		}
	} else {
		return true
	}
	return isValid
}

/*
	FilterRegistriesBasedOnStorageTypeAndAction

Parameters:
  - List of DockerArtifactStoreBean,
  - storageType:
    it can be any from repository.OCI_REGISRTY_REPO_TYPE_LIST
  - list of actionTypes:
    it can be PULL, PUSH, PULL/PUSH

Returns:
  - List of DockerArtifactStoreBean
  - Error: if invalid storageType
*/
func (impl DockerRegistryConfigImpl) FilterRegistryBeanListBasedOnStorageTypeAndAction(bean []DockerArtifactStoreBean, storageType string, actionTypes ...string) []DockerArtifactStoreBean {
	var registryConfigs []DockerArtifactStoreBean
	for _, registryConfig := range bean {
		// For OCI registries
		if registryConfig.IsOCICompliantRegistry {
			// Appends the OCI registries for specific Repo type; CHARTS or CONTAINERS
			if slices.Contains(actionTypes, registryConfig.OCIRegistryConfig[storageType]) {
				registryConfigs = append(registryConfigs, registryConfig)
			}
		} else if storageType == repository.OCI_REGISRTY_REPO_TYPE_CONTAINER {
			// Appends the container registries (not OCI)
			registryConfigs = append(registryConfigs, registryConfig)
		}
	}
	return registryConfigs
}

// CreateOrUpdateOCIRegistryConfig Takes the OCIRegistryConfig to be saved/updated and the DB context. For update OCIRegistryConfig.Id should be the record id to be updated. Returns Error if any.
func (impl DockerRegistryConfigImpl) CreateOrUpdateOCIRegistryConfig(ociRegistryConfig *repository.OCIRegistryConfig, userId int32, tx *pg.Tx) error {
	if ociRegistryConfig.Id > 0 {
		ociRegistryConfig.UpdatedOn = time.Now()
		ociRegistryConfig.UpdatedBy = userId
		err := impl.ociRegistryConfigRepository.Update(ociRegistryConfig, tx)
		if err != nil {
			impl.logger.Errorw("error in updating OCI config db", "err", err)
			return err
		}
	} else {
		// Prevents from creating entry with deleted True
		if ociRegistryConfig.Deleted {
			return nil
		}
		ociRegistryConfig.CreatedOn = time.Now()
		ociRegistryConfig.CreatedBy = userId
		ociRegistryConfig.UpdatedOn = time.Now()
		ociRegistryConfig.UpdatedBy = userId
		err := impl.ociRegistryConfigRepository.Save(ociRegistryConfig, tx)
		if err != nil {
			impl.logger.Errorw("error in saving OCI config db", "err", err)
			return err
		}
	}
	return nil
}

// FilterOCIRegistryConfigForSpecificRepoType Takes the list of OCIRegistryConfigs and the RepositoryType to be filtered. Returns the first entry that matches the repositoryType.
func (impl DockerRegistryConfigImpl) FilterOCIRegistryConfigForSpecificRepoType(ociRegistryConfigList []*repository.OCIRegistryConfig, repositoryType string) *repository.OCIRegistryConfig {
	ociRegistryConfig := &repository.OCIRegistryConfig{}
	for _, registryConfig := range ociRegistryConfigList {
		if registryConfig.RepositoryType == repositoryType {
			ociRegistryConfig = registryConfig
			break
		}
	}
	return ociRegistryConfig
}

// ConfigureOCIRegistry Takes DockerArtifactStoreBean, IsUpdate flag and the DB context. It finally creates/updates the OCI config in the DB. Returns Error if any.
func (impl DockerRegistryConfigImpl) ConfigureOCIRegistry(bean *DockerArtifactStoreBean, isUpdate bool, userId int32, tx *pg.Tx) error {
	ociRegistryConfigList, err := impl.ociRegistryConfigRepository.FindByDockerRegistryId(bean.Id)
	if err != nil && (isUpdate || err != pg.ErrNoRows) {
		return err
	}

	// If the ociRegistryConfigBean doesn't have any repoType, then mark delete true.
	for _, repoType := range repository.OCI_REGISRTY_REPO_TYPE_LIST {
		if _, ok := bean.OCIRegistryConfig[repoType]; !ok {
			bean.OCIRegistryConfig[repoType] = ""
		}
	}

	for repositoryType, storageActionType := range bean.OCIRegistryConfig {
		if !slices.Contains(repository.OCI_REGISRTY_REPO_TYPE_LIST, repositoryType) {
			return fmt.Errorf("invalid repository type for OCI registry configuration")
		}
		var ociRegistryConfig *repository.OCIRegistryConfig
		if !isUpdate {
			ociRegistryConfig = &repository.OCIRegistryConfig{
				DockerArtifactStoreId: bean.Id,
				Deleted:               false,
			}
		} else {
			ociRegistryConfig = impl.FilterOCIRegistryConfigForSpecificRepoType(ociRegistryConfigList, repositoryType)
			if ociRegistryConfig.Id == 0 {
				ociRegistryConfig.DockerArtifactStoreId = bean.Id
				ociRegistryConfig.Deleted = false
			}
		}
		switch storageActionType {
		case repository.STORAGE_ACTION_TYPE_PULL:
			ociRegistryConfig.RepositoryAction = repository.STORAGE_ACTION_TYPE_PULL
			ociRegistryConfig.RepositoryType = repositoryType
			ociRegistryConfig.RepositoryList = strings.Join(bean.RepositoryList, ",")
			ociRegistryConfig.IsPublic = bean.IsPublic
			err := impl.CreateOrUpdateOCIRegistryConfig(ociRegistryConfig, userId, tx)
			if err != nil {
				return err
			}
		case repository.STORAGE_ACTION_TYPE_PUSH:
			ociRegistryConfig.RepositoryAction = repository.STORAGE_ACTION_TYPE_PUSH
			ociRegistryConfig.RepositoryType = repositoryType
			err := impl.CreateOrUpdateOCIRegistryConfig(ociRegistryConfig, userId, tx)
			if err != nil {
				return err
			}
		case repository.STORAGE_ACTION_TYPE_PULL_AND_PUSH:
			ociRegistryConfig.RepositoryAction = repository.STORAGE_ACTION_TYPE_PULL_AND_PUSH
			ociRegistryConfig.RepositoryType = repositoryType
			ociRegistryConfig.RepositoryList = strings.Join(bean.RepositoryList, ",")
			ociRegistryConfig.IsPublic = bean.IsPublic
			err := impl.CreateOrUpdateOCIRegistryConfig(ociRegistryConfig, userId, tx)
			if err != nil {
				return err
			}
		case "":
			ociRegistryConfig.Deleted = true
			err := impl.CreateOrUpdateOCIRegistryConfig(ociRegistryConfig, userId, tx)
			if err != nil {
				return err
			}
			delete(bean.OCIRegistryConfig, repositoryType)
		default:
			return fmt.Errorf("invalid repository action type for OCI registry configuration")
		}
	}
	return nil
}

// Create Takes the DockerArtifactStoreBean and creates the record in DB. Returns Error if any
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

	// 3- insert OCIRegistryConfig for this docker registry
	if store.IsOCICompliantRegistry {
		err = impl.ConfigureOCIRegistry(bean, false, bean.User, tx)
		if err != nil {
			impl.logger.Errorw("error in saving OCI registry config", "OCIRegistryConfig", bean.OCIRegistryConfig, "err", err)
			err = &util.ApiError{
				Code:            constants.DockerRegCreateFailedInDb,
				InternalMessage: err.Error(),
				UserMessage:     "Error in creating OCI registry config in db",
			}
			return nil, err
		}
		impl.logger.Infow("created OCI registry config successfully")
	}

	// 4- insert imagePullSecretConfig for this docker registry
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
	impl.logger.Infow("created ips config for this docker repository", "ipsConfig", ipsConfig)
	dockerRegistryIpsConfig.Id = ipsConfig.Id

	// 4- now commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return nil, err
	}

	return bean, nil
}

// ListAllActive Returns the list all active artifact stores
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
			Id:                     store.Id,
			RegistryURL:            store.RegistryURL,
			IsDefault:              store.IsDefault,
			RegistryType:           store.RegistryType,
			IsOCICompliantRegistry: store.IsOCICompliantRegistry,
		}
		if store.IsOCICompliantRegistry {
			storeBean.OCIRegistryConfig, storeBean.RepositoryList, storeBean.IsPublic = impl.PopulateOCIRegistryConfig(&store)
		}
		storeBeans = append(storeBeans, storeBean)
	}
	return storeBeans, err
}

// FetchAllDockerAccounts method used for getting all registry accounts with complete details
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
			Id:                     store.Id,
			PluginId:               store.PluginId,
			RegistryURL:            store.RegistryURL,
			RegistryType:           store.RegistryType,
			AWSAccessKeyId:         store.AWSAccessKeyId,
			AWSSecretAccessKey:     "",
			AWSRegion:              store.AWSRegion,
			Username:               store.Username,
			Password:               "",
			IsDefault:              store.IsDefault,
			Connection:             store.Connection,
			Cert:                   store.Cert,
			Active:                 store.Active,
			IsOCICompliantRegistry: store.IsOCICompliantRegistry,
			DockerRegistryIpsConfig: &DockerRegistryIpsConfigBean{
				Id:                   ipsConfig.Id,
				CredentialType:       ipsConfig.CredentialType,
				CredentialValue:      ipsConfig.CredentialValue,
				AppliedClusterIdsCsv: ipsConfig.AppliedClusterIdsCsv,
				IgnoredClusterIdsCsv: ipsConfig.IgnoredClusterIdsCsv,
			},
		}
		if store.IsOCICompliantRegistry {
			storeBean.OCIRegistryConfig, storeBean.RepositoryList, storeBean.IsPublic = impl.PopulateOCIRegistryConfig(&store)
		}
		storeBeans = append(storeBeans, storeBean)
	}

	return storeBeans, err
}

// PopulateOCIRegistryConfig Takes the DB docker_artifact_store response and generates
func (impl DockerRegistryConfigImpl) PopulateOCIRegistryConfig(store *repository.DockerArtifactStore) (map[string]string, []string, bool) {
	ociRegistryConfigs := map[string]string{}
	ociPullRepositryList := []string{}
	isPublic := false
	for _, ociRegistryConfig := range store.OCIRegistryConfig {
		ociRegistryConfigs[ociRegistryConfig.RepositoryType] = ociRegistryConfig.RepositoryAction
		if ociRegistryConfig.RepositoryAction == repository.OCI_REGISRTY_REPO_TYPE_CHART {
			ociPullRepositryList = strings.Split(ociRegistryConfig.RepositoryList, ",")
			isPublic = ociRegistryConfig.IsPublic
		}
	}
	return ociRegistryConfigs, ociPullRepositryList, isPublic
}

// FetchOneDockerAccount this method takes the docker account id and Returns DockerArtifactStoreBean and Error (if any)
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
	if store.IsOCICompliantRegistry {
		storeBean.OCIRegistryConfig, storeBean.RepositoryList, storeBean.IsPublic = impl.PopulateOCIRegistryConfig(store)
	}
	return storeBean, err
}

// Update will update the existing registry with the given DockerArtifactStoreBean
func (impl DockerRegistryConfigImpl) Update(bean *DockerArtifactStoreBean) (*DockerArtifactStoreBean, error) {
	impl.logger.Debugw("docker registry update request", "request", bean)

	// 1- find by id, if err - return error
	existingStore, err := impl.dockerArtifactStoreRepository.FindOne(bean.Id)
	if err != nil {
		impl.logger.Errorw("no matching entry found of update ..", "err", err)
		return nil, err
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

	// 4- update OCIRegistryConfig for this docker registry
	if store.IsOCICompliantRegistry {
		err = impl.ConfigureOCIRegistry(bean, true, bean.User, tx)
		if err != nil {
			impl.logger.Errorw("error in updating OCI registry config", "OCIRegistryConfig", bean.OCIRegistryConfig, "err", err)
			err = &util.ApiError{
				Code:            constants.DockerRegCreateFailedInDb,
				InternalMessage: err.Error(),
				UserMessage:     "Error in updating OCI registry config in db",
			}
			return nil, err
		}
		impl.logger.Infow("updated OCI registry config successfully")
	}

	// 5- update imagePullSecretConfig for this docker registry
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

	// 6- now commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return nil, err
	}

	return bean, nil
}

// UpdateInactive will update the existing soft deleted registry with the given DockerArtifactStoreBean instead of creating one
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

	// 4- update OCIRegistryConfig for this docker registry
	if store.IsOCICompliantRegistry {
		err = impl.ConfigureOCIRegistry(bean, true, bean.User, tx)
		if err != nil {
			impl.logger.Errorw("error in updating OCI registry config", "OCIRegistryConfig", bean.OCIRegistryConfig, "err", err)
			err = &util.ApiError{
				Code:            constants.DockerRegCreateFailedInDb,
				InternalMessage: err.Error(),
				UserMessage:     "Error in updating OCI registry config in db",
			}
			return nil, err
		}
		impl.logger.Infow("updated OCI registry config successfully")
	}

	// 5- update imagePullSecretConfig for this docker registry
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

	// 6- now commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return nil, err
	}

	return bean, nil
}

// Delete is a Deprecated function. It was used to hard delete the registry from DB.
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

// DeleteReg Takes DockerArtifactStoreBean and soft deletes the OCI configs (if exists), finally soft deletes the registry. Returns Error if any.
func (impl DockerRegistryConfigImpl) DeleteReg(bean *DockerArtifactStoreBean) error {
	// 1- fetching Artifact Registry
	dockerReg, err := impl.dockerArtifactStoreRepository.FindOne(bean.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete.", "id", bean.Id, "err", err)
		return err
	}
	// 2- initiate DB transaction
	dbConnection := impl.dockerArtifactStoreRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in initiating db tx", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// 3- fetching OCI config attached to the Artifact Registry
	ociRegistryConfigs, err := impl.ociRegistryConfigRepository.FindByDockerRegistryId(dockerReg.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete.", "id", bean.Id, "err", err)
		return err
	}

	// 4- marking deleted, OCI configs attached to the Artifact Registry
	for _, ociRegistryConfig := range ociRegistryConfigs {
		if !ociRegistryConfig.Deleted {
			ociRegistryConfig.Deleted = true
			ociRegistryConfig.UpdatedOn = time.Now()
			ociRegistryConfig.UpdatedBy = bean.User
			err = impl.ociRegistryConfigRepository.Update(ociRegistryConfig, tx)
			if err != nil {
				impl.logger.Errorw("err in deleting OCI configs for registry", "registryId", bean.Id, "err", err)
				return err
			}
		}
	}

	// 4- mark deleted, Artifact Registry
	deleteReq := dockerReg
	deleteReq.UpdatedOn = time.Now()
	deleteReq.UpdatedBy = bean.User
	err = impl.dockerArtifactStoreRepository.MarkRegistryDeleted(deleteReq, tx)
	if err != nil {
		impl.logger.Errorw("err in deleting docker registry", "id", bean.Id, "err", err)
		return err
	}

	// 6- now commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
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
