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
	"context"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/helm-app/gRPC"
	client "github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"k8s.io/utils/strings/slices"
	"net/http"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/util"
	"go.uber.org/zap"
)

type DockerRegistryConfig interface {
	Create(bean *types.DockerArtifactStoreBean) (*types.DockerArtifactStoreBean, error)
	ListAllActive() ([]types.DockerArtifactStoreBean, error)
	FetchAllDockerAccounts() ([]types.DockerArtifactStoreBean, error)
	FetchOneDockerAccount(storeId string) (*types.DockerArtifactStoreBean, error)
	Update(bean *types.DockerArtifactStoreBean) (*types.DockerArtifactStoreBean, error)
	UpdateInactive(bean *types.DockerArtifactStoreBean) (*types.DockerArtifactStoreBean, error)
	Delete(storeId string) (string, error)
	DeleteReg(bean *types.DockerArtifactStoreBean) error
	CheckInActiveDockerAccount(storeId string) (bool, error)
	ValidateRegistryCredentials(bean *types.DockerArtifactStoreBean) bool
	ConfigureOCIRegistry(bean *types.DockerArtifactStoreBean, isUpdate bool, userId int32, tx *pg.Tx) error
	CreateOrUpdateOCIRegistryConfig(ociRegistryConfig *repository.OCIRegistryConfig, userId int32, tx *pg.Tx) error
	FilterOCIRegistryConfigForSpecificRepoType(ociRegistryConfigList []*repository.OCIRegistryConfig, repositoryType string) *repository.OCIRegistryConfig
	FilterRegistryBeanListBasedOnStorageTypeAndAction(bean []types.DockerArtifactStoreBean, storageType string, actionTypes ...string) []types.DockerArtifactStoreBean
	ValidateRegistryStorageType(registryId string, storageType string, storageActions ...string) bool
}

const (
	DISABLED_CONTAINER  types.DisabledFields = "CONTAINER"
	DISABLED_CHART_PULL types.DisabledFields = "CHART_PULL"
	DISABLED_CHART_PUSH types.DisabledFields = "CHART_PUSH"
)

type DockerRegistryConfigImpl struct {
	logger                            *zap.SugaredLogger
	helmAppService                    client.HelmAppService
	dockerArtifactStoreRepository     repository.DockerArtifactStoreRepository
	dockerRegistryIpsConfigRepository repository.DockerRegistryIpsConfigRepository
	ociRegistryConfigRepository       repository.OCIRegistryConfigRepository
}

func NewDockerRegistryConfigImpl(logger *zap.SugaredLogger, helmAppService client.HelmAppService, dockerArtifactStoreRepository repository.DockerArtifactStoreRepository,
	dockerRegistryIpsConfigRepository repository.DockerRegistryIpsConfigRepository, ociRegistryConfigRepository repository.OCIRegistryConfigRepository) *DockerRegistryConfigImpl {
	return &DockerRegistryConfigImpl{
		logger:                            logger,
		helmAppService:                    helmAppService,
		dockerArtifactStoreRepository:     dockerArtifactStoreRepository,
		dockerRegistryIpsConfigRepository: dockerRegistryIpsConfigRepository,
		ociRegistryConfigRepository:       ociRegistryConfigRepository,
	}
}

func NewDockerArtifactStore(bean *types.DockerArtifactStoreBean, isActive bool, createdOn time.Time, updatedOn time.Time, createdBy int32, updateBy int32) *repository.DockerArtifactStore {
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
func (impl DockerRegistryConfigImpl) FilterRegistryBeanListBasedOnStorageTypeAndAction(bean []types.DockerArtifactStoreBean, storageType string, actionTypes ...string) []types.DockerArtifactStoreBean {
	var registryConfigs []types.DockerArtifactStoreBean
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
func (impl DockerRegistryConfigImpl) ConfigureOCIRegistry(bean *types.DockerArtifactStoreBean, isUpdate bool, userId int32, tx *pg.Tx) error {
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
		ociRegistryConfig.IsPublic = false
		switch storageActionType {
		case repository.STORAGE_ACTION_TYPE_PULL:
			ociRegistryConfig.RepositoryAction = repository.STORAGE_ACTION_TYPE_PULL
			ociRegistryConfig.RepositoryType = repositoryType
			if repositoryType == repository.OCI_REGISRTY_REPO_TYPE_CHART {
				ociRegistryConfig.RepositoryList = strings.Join(bean.RepositoryList, ",")
				ociRegistryConfig.IsPublic = bean.IsPublic
				ociRegistryConfig.IsChartPullActive = true
			}
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
			if repositoryType == repository.OCI_REGISRTY_REPO_TYPE_CHART {
				ociRegistryConfig.RepositoryList = strings.Join(bean.RepositoryList, ",")
				ociRegistryConfig.IsPublic = bean.IsPublic
				ociRegistryConfig.IsChartPullActive = true
			}
			err := impl.CreateOrUpdateOCIRegistryConfig(ociRegistryConfig, userId, tx)
			if err != nil {
				return err
			}
		case "":
			ociRegistryConfig.Deleted = true
			ociRegistryConfig.IsChartPullActive = false
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
func (impl DockerRegistryConfigImpl) Create(bean *types.DockerArtifactStoreBean) (*types.DockerArtifactStoreBean, error) {
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

	if !bean.IsPublic && bean.DockerRegistryIpsConfig != nil {
		// 4- insert imagePullSecretConfig for this docker registry
		dockerRegistryIpsConfig := bean.DockerRegistryIpsConfig
		ipsConfig := &repository.DockerRegistryIpsConfig{
			DockerArtifactStoreId: store.Id,
			CredentialType:        dockerRegistryIpsConfig.CredentialType,
			CredentialValue:       dockerRegistryIpsConfig.CredentialValue,
			AppliedClusterIdsCsv:  dockerRegistryIpsConfig.AppliedClusterIdsCsv,
			IgnoredClusterIdsCsv:  dockerRegistryIpsConfig.IgnoredClusterIdsCsv,
			Active:                true,
		}
		err = impl.createDockerIpConfig(tx, ipsConfig)
		if err != nil {
			return nil, err
		}
		dockerRegistryIpsConfig.Id = ipsConfig.Id
	}

	// 4- now commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return nil, err
	}

	return bean, nil
}

// ListAllActive Returns the list all active artifact stores
func (impl DockerRegistryConfigImpl) ListAllActive() ([]types.DockerArtifactStoreBean, error) {
	impl.logger.Debug("list docker repo request")
	stores, err := impl.dockerArtifactStoreRepository.FindAllActiveForAutocomplete()
	if err != nil {
		impl.logger.Errorw("error in listing artifact", "err", err)
		return nil, err
	}
	var storeBeans []types.DockerArtifactStoreBean
	for _, store := range stores {
		storeBean := types.DockerArtifactStoreBean{
			Id:                     store.Id,
			RegistryURL:            store.RegistryURL,
			IsDefault:              store.IsDefault,
			RegistryType:           store.RegistryType,
			IsOCICompliantRegistry: store.IsOCICompliantRegistry,
		}
		if store.IsOCICompliantRegistry {
			impl.PopulateOCIRegistryConfig(&store, &storeBean)
		}
		storeBeans = append(storeBeans, storeBean)
	}
	return storeBeans, err
}

// FetchAllDockerAccounts method used for getting all registry accounts with complete details
func (impl DockerRegistryConfigImpl) FetchAllDockerAccounts() ([]types.DockerArtifactStoreBean, error) {
	impl.logger.Debug("list docker repo request")
	stores, err := impl.dockerArtifactStoreRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("error in listing artifact", "err", err)
		return nil, err
	}
	var storeBeans []types.DockerArtifactStoreBean
	for _, store := range stores {
		ipsConfig := store.IpsConfig
		storeBean := types.DockerArtifactStoreBean{
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
		}
		if store.IsOCICompliantRegistry {
			impl.PopulateOCIRegistryConfig(&store, &storeBean)
		}
		if ipsConfig != nil {
			storeBean.DockerRegistryIpsConfig = &types.DockerRegistryIpsConfigBean{
				Id:                   ipsConfig.Id,
				CredentialType:       ipsConfig.CredentialType,
				CredentialValue:      ipsConfig.CredentialValue,
				AppliedClusterIdsCsv: ipsConfig.AppliedClusterIdsCsv,
				IgnoredClusterIdsCsv: ipsConfig.IgnoredClusterIdsCsv,
				Active:               ipsConfig.Active,
			}
		}
		storeBeans = append(storeBeans, storeBean)
	}

	return storeBeans, err
}

// PopulateOCIRegistryConfig Takes the DB docker_artifact_store response and generates
func (impl DockerRegistryConfigImpl) PopulateOCIRegistryConfig(store *repository.DockerArtifactStore, storeBean *types.DockerArtifactStoreBean) *types.DockerArtifactStoreBean {
	ociRegistryConfigs := map[string]string{}
	for _, ociRegistryConfig := range store.OCIRegistryConfig {
		ociRegistryConfigs[ociRegistryConfig.RepositoryType] = ociRegistryConfig.RepositoryAction
		if ociRegistryConfig.RepositoryType == repository.OCI_REGISRTY_REPO_TYPE_CHART {
			storeBean.RepositoryList = strings.Split(ociRegistryConfig.RepositoryList, ",")
			storeBean.IsPublic = ociRegistryConfig.IsPublic
		}
	}
	storeBean.OCIRegistryConfig = ociRegistryConfigs
	return storeBean
}

// FetchOneDockerAccount this method takes the docker account id and Returns DockerArtifactStoreBean and Error (if any)
func (impl DockerRegistryConfigImpl) FetchOneDockerAccount(storeId string) (*types.DockerArtifactStoreBean, error) {
	impl.logger.Debug("fetch docker account by id from db")
	store, err := impl.dockerArtifactStoreRepository.FindOne(storeId)
	if err != nil {
		impl.logger.Errorw("error in listing artifact", "err", err)
		return nil, err
	}

	ipsConfig := store.IpsConfig
	storeBean := &types.DockerArtifactStoreBean{
		Id:                     store.Id,
		PluginId:               store.PluginId,
		RegistryURL:            store.RegistryURL,
		RegistryType:           store.RegistryType,
		AWSAccessKeyId:         store.AWSAccessKeyId,
		AWSSecretAccessKey:     store.AWSSecretAccessKey,
		AWSRegion:              store.AWSRegion,
		Username:               store.Username,
		Password:               store.Password,
		IsDefault:              store.IsDefault,
		Connection:             store.Connection,
		Cert:                   store.Cert,
		Active:                 store.Active,
		IsOCICompliantRegistry: store.IsOCICompliantRegistry,
	}
	if store.IsOCICompliantRegistry {
		impl.PopulateOCIRegistryConfig(store, storeBean)
	}
	if ipsConfig != nil {
		storeBean.DockerRegistryIpsConfig = &types.DockerRegistryIpsConfigBean{
			Id:                   ipsConfig.Id,
			CredentialType:       ipsConfig.CredentialType,
			CredentialValue:      ipsConfig.CredentialValue,
			AppliedClusterIdsCsv: ipsConfig.AppliedClusterIdsCsv,
			IgnoredClusterIdsCsv: ipsConfig.IgnoredClusterIdsCsv,
			Active:               ipsConfig.Active,
		}
	}
	return storeBean, err
}

// Update will update the existing registry with the given DockerArtifactStoreBean
func (impl DockerRegistryConfigImpl) Update(bean *types.DockerArtifactStoreBean) (*types.DockerArtifactStoreBean, error) {
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

	existingRepositoryList := make([]string, 0)
	for _, ociRegistryConfig := range existingStore.OCIRegistryConfig {
		if ociRegistryConfig.RepositoryType == repository.OCI_REGISRTY_REPO_TYPE_CHART {
			existingRepositoryList = strings.Split(ociRegistryConfig.RepositoryList, ",")
		}
	}
	deployedChartList := make([]string, 0)
	for _, repository := range existingRepositoryList {
		if !slices.Contains(bean.RepositoryList, repository) {
			chartDeploymentCount, err := impl.dockerArtifactStoreRepository.FindOneWithChartDeploymentCount(bean.Id, repository)
			if err != nil && err != pg.ErrNoRows {
				return nil, err
			} else if chartDeploymentCount != nil && chartDeploymentCount.DeploymentCount > 0 {
				deployedChartList = append(deployedChartList, chartDeploymentCount.OCIChartName)
			}
		}
	}
	if len(deployedChartList) > 0 {
		err := &util.ApiError{
			HttpStatusCode:  http.StatusConflict,
			InternalMessage: fmt.Sprintf("%s chart(s) cannot be removed as they are being used by helm applications.", strings.Join(deployedChartList, ", ")),
			UserMessage:     fmt.Sprintf("%s chart(s) cannot be removed as they are being used by helm applications.", strings.Join(deployedChartList, ", "))}
		return nil, err
	}

	bean.PluginId = existingStore.PluginId

	store := NewDockerArtifactStore(bean, true, existingStore.CreatedOn, time.Now(), existingStore.CreatedBy, bean.User)
	if isValid := impl.ValidateRegistryCredentials(bean); !isValid {
		impl.logger.Errorw("registry credentials validation err, SaveDockerRegistryConfig", "err", err, "payload", bean)
		err = &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: "Invalid authentication credentials. Please verify.",
			UserMessage:     "Invalid authentication credentials. Please verify.",
		}
		return nil, err
	}
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
	existingIpsConfig, err := impl.dockerRegistryIpsConfigRepository.FindByDockerRegistryId(store.Id)
	if !bean.IsPublic && bean.DockerRegistryIpsConfig != nil {
		dockerRegistryIpsConfig := bean.DockerRegistryIpsConfig
		ipsConfig := &repository.DockerRegistryIpsConfig{
			DockerArtifactStoreId: store.Id,
			CredentialType:        dockerRegistryIpsConfig.CredentialType,
			CredentialValue:       dockerRegistryIpsConfig.CredentialValue,
			AppliedClusterIdsCsv:  dockerRegistryIpsConfig.AppliedClusterIdsCsv,
			IgnoredClusterIdsCsv:  dockerRegistryIpsConfig.IgnoredClusterIdsCsv,
			Active:                true,
		}
		if err != nil {
			impl.logger.Errorw("Error while getting docker registry ips config", "dockerRegistryId", store.Id, "err", err)
			// Create a new docker registry ips config
			if err == pg.ErrNoRows {
				err = impl.createDockerIpConfig(tx, ipsConfig)
				if err != nil {
					return nil, err
				}
			} else {
				// Throw error
				return nil, err
			}
		} else {
			// Update the docker registry ips config
			ipsConfig.Id = existingIpsConfig.Id
			err = impl.updateDockerIpConfig(tx, ipsConfig)
			if err != nil {
				return nil, err
			}
		}
		bean.DockerRegistryIpsConfig.Id = ipsConfig.Id
	} else {
		if err == nil {
			// Update the docker registry ips config to inactive
			existingIpsConfig.Active = false
			err = impl.updateDockerIpConfig(tx, existingIpsConfig)
			if err != nil {
				return nil, err
			}
		}
	}

	// 6- now commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return nil, err
	}

	return bean, nil
}

func (impl DockerRegistryConfigImpl) updateDockerIpConfig(tx *pg.Tx, existingIpsConfig *repository.DockerRegistryIpsConfig) error {
	err := impl.dockerRegistryIpsConfigRepository.Update(existingIpsConfig, tx)
	if err != nil {
		impl.logger.Errorw("error in updating registry config ips", "ipsConfig", existingIpsConfig, "err", err)
		err = &util.ApiError{
			Code:            constants.DockerRegUpdateFailedInDb,
			InternalMessage: "docker registry ips config failed to update in db",
			UserMessage:     "docker registry ips config failed to update in db",
		}
		return err
	}
	impl.logger.Infow("updated ips config for this docker repository ", "ipsConfig", existingIpsConfig)
	return nil
}

func (impl DockerRegistryConfigImpl) createDockerIpConfig(tx *pg.Tx, ipsConfig *repository.DockerRegistryIpsConfig) error {
	existingIpsConfig, err := impl.dockerRegistryIpsConfigRepository.FindInActiveByDockerRegistryId(ipsConfig.DockerArtifactStoreId)
	if err != nil {
		if err == pg.ErrNoRows {
			err = impl.dockerRegistryIpsConfigRepository.Save(ipsConfig, tx)
			if err != nil {
				impl.logger.Errorw("error in saving registry config ips", "ipsConfig", ipsConfig, "err", err)
				err = &util.ApiError{
					Code:            constants.DockerRegCreateFailedInDb,
					InternalMessage: "docker registry ips config to create in db",
					UserMessage:     fmt.Sprintf("Container registry [%s] already exists.", ipsConfig.DockerArtifactStoreId),
				}
				return err
			}
			impl.logger.Infow("created ips config for this docker repository", "ipsConfig", ipsConfig)
			return nil
		} else {
			return err
		}
	}
	ipsConfig.Id = existingIpsConfig.Id
	err = impl.updateDockerIpConfig(tx, ipsConfig)
	if err != nil {
		return err
	}
	return nil
}

// UpdateInactive will update the existing soft deleted registry with the given DockerArtifactStoreBean instead of creating one
func (impl DockerRegistryConfigImpl) UpdateInactive(bean *types.DockerArtifactStoreBean) (*types.DockerArtifactStoreBean, error) {
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

	if !bean.IsPublic && bean.DockerRegistryIpsConfig != nil {
		// 5- update imagePullSecretConfig for this docker registry
		dockerRegistryIpsConfig := bean.DockerRegistryIpsConfig
		ipsConfig := &repository.DockerRegistryIpsConfig{
			DockerArtifactStoreId: store.Id,
			CredentialType:        dockerRegistryIpsConfig.CredentialType,
			CredentialValue:       dockerRegistryIpsConfig.CredentialValue,
			AppliedClusterIdsCsv:  dockerRegistryIpsConfig.AppliedClusterIdsCsv,
			IgnoredClusterIdsCsv:  dockerRegistryIpsConfig.IgnoredClusterIdsCsv,
			Active:                true,
		}
		if existingStore.IpsConfig != nil {
			ipsConfig.Id = existingStore.IpsConfig.Id
			err = impl.updateDockerIpConfig(tx, ipsConfig)
			if err != nil {
				return nil, err
			}
		} else {
			err = impl.createDockerIpConfig(tx, ipsConfig)
			if err != nil {
				return nil, err
			}
		}
	}

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
func (impl DockerRegistryConfigImpl) DeleteReg(bean *types.DockerArtifactStoreBean) error {
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

func (impl DockerRegistryConfigImpl) ValidateRegistryCredentials(bean *types.DockerArtifactStoreBean) bool {
	if bean.IsPublic ||
		bean.RegistryType == repository.REGISTRYTYPE_GCR ||
		bean.RegistryType == repository.REGISTRYTYPE_ARTIFACT_REGISTRY ||
		bean.RegistryType == repository.REGISTRYTYPE_OTHER {
		return true
	}
	request := &bean2.RegistryCredential{
		RegistryUrl:  bean.RegistryURL,
		Username:     bean.Username,
		Password:     bean.Password,
		AwsRegion:    bean.AWSRegion,
		AccessKey:    bean.AWSAccessKeyId,
		SecretKey:    bean.AWSSecretAccessKey,
		RegistryType: string(bean.RegistryType),
		IsPublic:     bean.IsPublic,
	}
	return impl.helmAppService.ValidateOCIRegistry(context.Background(), request)
}
