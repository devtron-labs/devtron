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

package restHandler

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	repository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/util"
	chartProviderService "github.com/devtron-labs/devtron/pkg/appStore/chartProvider"
	deleteService "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util2 "github.com/devtron-labs/devtron/util"
	"k8s.io/utils/strings/slices"
	"net/http"
	"strings"

	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

const (
	REG_DELETE_SUCCESS_RESP = "Container Registry deleted successfully."
	OCIScheme               = "oci://"
	secureWithCert          = "secure-with-cert"
)

type DockerRegRestHandler interface {
	SaveDockerRegistryConfig(w http.ResponseWriter, r *http.Request)
	ValidateDockerRegistryConfig(w http.ResponseWriter, r *http.Request)
	GetDockerArtifactStore(w http.ResponseWriter, r *http.Request)
	FetchAllDockerAccounts(w http.ResponseWriter, r *http.Request)
	FetchOneDockerAccounts(w http.ResponseWriter, r *http.Request)
	UpdateDockerRegistryConfig(w http.ResponseWriter, r *http.Request)
	FetchAllDockerRegistryForAutocomplete(w http.ResponseWriter, r *http.Request)
	IsDockerRegConfigured(w http.ResponseWriter, r *http.Request)
	DeleteDockerRegistryConfig(w http.ResponseWriter, r *http.Request)
}

type DockerRegRestHandlerExtendedImpl struct {
	deleteServiceFullMode deleteService.DeleteServiceFullMode
	*DockerRegRestHandlerImpl
}

type DockerRegRestHandlerImpl struct {
	dockerRegistryConfig pipeline.DockerRegistryConfig
	logger               *zap.SugaredLogger
	chartProviderService chartProviderService.ChartProviderService
	userAuthService      user.UserService
	validator            *validator.Validate
	enforcer             casbin.Enforcer
	teamService          team.TeamService
	deleteService        deleteService.DeleteService
}

func NewDockerRegRestHandlerExtendedImpl(
	dockerRegistryConfig pipeline.DockerRegistryConfig,
	logger *zap.SugaredLogger,
	chartProviderService chartProviderService.ChartProviderService,
	userAuthService user.UserService,
	validator *validator.Validate, enforcer casbin.Enforcer, teamService team.TeamService,
	deleteService deleteService.DeleteService,
	deleteServiceFullMode deleteService.DeleteServiceFullMode) *DockerRegRestHandlerExtendedImpl {
	return &DockerRegRestHandlerExtendedImpl{
		deleteServiceFullMode: deleteServiceFullMode,
		DockerRegRestHandlerImpl: &DockerRegRestHandlerImpl{
			dockerRegistryConfig: dockerRegistryConfig,
			logger:               logger,
			chartProviderService: chartProviderService,
			userAuthService:      userAuthService,
			validator:            validator,
			enforcer:             enforcer,
			teamService:          teamService,
			deleteService:        deleteService,
		},
	}
}

func NewDockerRegRestHandlerImpl(
	dockerRegistryConfig pipeline.DockerRegistryConfig,
	logger *zap.SugaredLogger,
	chartProviderService chartProviderService.ChartProviderService,
	userAuthService user.UserService,
	validator *validator.Validate, enforcer casbin.Enforcer, teamService team.TeamService,
	deleteService deleteService.DeleteService) *DockerRegRestHandlerImpl {
	return &DockerRegRestHandlerImpl{
		dockerRegistryConfig: dockerRegistryConfig,
		logger:               logger,
		chartProviderService: chartProviderService,
		userAuthService:      userAuthService,
		validator:            validator,
		enforcer:             enforcer,
		teamService:          teamService,
		deleteService:        deleteService,
	}
}

func ValidateDockerArtifactStoreRequestBean(bean types.DockerArtifactStoreBean) error {
	// handling for EA mode registry setup
	if util2.IsBaseStack() {
		// For EA MODE, DockerRegistryIpsConfig should be nil
		bean.DockerRegistryIpsConfig = nil
		// For EA MODE, IsDefault should be FALSE
		bean.IsDefault = false

		// For EA MODE, IsOCICompliantRegistry should be TRUE
		if bean.RegistryType == repository.REGISTRYTYPE_GCR {
			return fmt.Errorf("Invalid payload! 'GCR' is not supported as an OCI registry.")
		}
		// For EA MODE, IsOCICompliantRegistry should be TRUE
		if !bean.IsOCICompliantRegistry {
			return fmt.Errorf("Invalid payload! 'isOCICompliantRegistry' is required.")
		}
		// For EA MODE, OCIRegistryConfig is mandatory
		if bean.OCIRegistryConfig == nil {
			return fmt.Errorf("Invalid payload! 'ociRegistryConfig' is required.")
		}

		// For EA MODE there should be no config for Container "PULL/PUSH"
		if _, containerStorageActionExists := bean.OCIRegistryConfig[repository.OCI_REGISRTY_REPO_TYPE_CONTAINER]; containerStorageActionExists {
			return fmt.Errorf("Invalid payload! 'ociRegistryConfig' has invalid field 'CONTAINER'.")
		}
		// For EA MODE, Charts storage action type should always be "PULL"
		chartStorageActionType, chartStorageActionExists := bean.OCIRegistryConfig[repository.OCI_REGISRTY_REPO_TYPE_CHART]
		if chartStorageActionExists && chartStorageActionType != repository.STORAGE_ACTION_TYPE_PULL {
			return fmt.Errorf("Invalid payload! 'ociRegistryConfig[CHART]' has invalid value '%s'.", chartStorageActionType)
		}
	}
	// validating secure connection configs
	if (bean.Connection == secureWithCert && bean.Cert == "") ||
		(bean.Connection != secureWithCert && bean.Cert != "") {
		return fmt.Errorf("Invalid payload! invalid value of 'cert' for the 'connection' type '%s'.", bean.Connection)
	}
	// validating OCI Registry configs
	if bean.IsOCICompliantRegistry {
		if bean.OCIRegistryConfig == nil {
			return fmt.Errorf("Invalid payload! 'ociRegistryConfig' is required")
		}
		// For Containers, storage action should be "PULL/PUSH"
		containerStorageActionType, containerStorageActionExists := bean.OCIRegistryConfig[repository.OCI_REGISRTY_REPO_TYPE_CONTAINER]
		if containerStorageActionExists && containerStorageActionType != repository.STORAGE_ACTION_TYPE_PULL_AND_PUSH && bean.DockerRegistryIpsConfig == nil {
			return fmt.Errorf("Invalid payload! 'ociRegistryConfig[CONTAINER]' has invalid value '%s'.", containerStorageActionType)
		}
		chartStorageActionType, chartStorageActionExists := bean.OCIRegistryConfig[repository.OCI_REGISRTY_REPO_TYPE_CHART]
		// For Charts with storage action type "PULL", default will always be false
		if chartStorageActionExists && chartStorageActionType == repository.STORAGE_ACTION_TYPE_PULL {
			bean.IsDefault = false
		}

		// For Charts with storage action type "PULL/PUSH" or "PULL", RepositoryList cannot be nil
		if chartStorageActionExists && (chartStorageActionType == repository.STORAGE_ACTION_TYPE_PULL_AND_PUSH || chartStorageActionType == repository.STORAGE_ACTION_TYPE_PULL) {
			if bean.RepositoryList == nil || len(bean.RepositoryList) == 0 || slices.Contains(bean.RepositoryList, "") {
				return fmt.Errorf("Invalid payload! invalid value for the required field 'repositoryList'.")
			}
		}
		// For public registry, URL prefix "oci://" should be trimmed, DockerRegistryIpsConfig should be nil and default should be false
		if bean.IsPublic {
			bean.IsDefault = false
			bean.DockerRegistryIpsConfig = nil
			bean.RegistryURL = strings.TrimPrefix(bean.RegistryURL, OCIScheme)
		} else if containerStorageActionExists && bean.DockerRegistryIpsConfig == nil {
			return fmt.Errorf("Invalid payload! 'ipsConfig' is required.")
		}
	} else if bean.OCIRegistryConfig != nil {
		return fmt.Errorf("Invalid payload! 'ociRegistryConfig' should be empty.")
	} else if bean.IsPublic {
		return fmt.Errorf("Invalid payload! 'isPublic' should be FALSE.")
	} else if bean.DockerRegistryIpsConfig == nil {
		return fmt.Errorf("Invalid payload! 'ipsConfig' is required.")
	}
	return nil
}

func (impl DockerRegRestHandlerImpl) SaveDockerRegistryConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean types.DockerArtifactStoreBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, SaveDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.User = userId
	requestErr := ValidateDockerArtifactStoreRequestBean(bean)
	if requestErr != nil {
		err = fmt.Errorf("invalid payload, missing or incorrect values for required fields")
		impl.logger.Errorw("validation err, SaveDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, requestErr, nil, http.StatusBadRequest)
		return
	}

	impl.logger.Infow("request payload, SaveDockerRegistryConfig", "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, SaveDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	// valid registry credentials from kubelink
	if isValid := impl.dockerRegistryConfig.ValidateRegistryCredentials(&bean); !isValid {
		impl.logger.Errorw("registry credentials validation err, SaveDockerRegistryConfig", "err", err, "payload", bean)
		err = &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: "Invalid authentication credentials. Please verify.",
			UserMessage:     "Invalid authentication credentials. Please verify.",
		}
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	exist, err := impl.dockerRegistryConfig.CheckInActiveDockerAccount(bean.Id)
	if err != nil {
		impl.logger.Errorw("service err, SaveDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if exist {
		res, err := impl.dockerRegistryConfig.UpdateInactive(&bean)
		if err != nil {
			impl.logger.Errorw("service err, UpdateDockerRegistryConfig", "err", err, "payload", bean)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		err = impl.TriggerChartSync(bean)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		common.WriteJsonResp(w, err, res, http.StatusOK)
		return
	}

	res, err := impl.dockerRegistryConfig.Create(&bean)
	if err != nil {
		impl.logger.Errorw("service err, SaveDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// trigger a chart sync job
	err = impl.TriggerChartSync(bean)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) TriggerChartSync(bean types.DockerArtifactStoreBean) error {
	if bean.IsOCICompliantRegistry && len(bean.RepositoryList) != 0 {
		request := &chartProviderService.ChartProviderRequestDto{
			Id:            bean.Id,
			IsOCIRegistry: bean.IsOCICompliantRegistry,
		}
		err := impl.chartProviderService.SyncChartProvider(request)
		if err != nil {
			impl.logger.Errorw("service err, SaveDockerRegistryConfig", "err", err)
			return err
		}
	}
	return nil
}

func (impl DockerRegRestHandlerImpl) ValidateDockerRegistryConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean types.DockerArtifactStoreBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, ValidateDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.User = userId

	impl.logger.Infow("request payload, ValidateDockerRegistryConfig", "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, ValidateDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	existingStore, err := impl.dockerRegistryConfig.FetchOneDockerAccount(bean.Id)
	if err != nil {
		impl.logger.Errorw("no matching entry found of update ..", "err", err)
		return
	}
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
	// valid registry credentials from kubelink
	if isValid := impl.dockerRegistryConfig.ValidateRegistryCredentials(&bean); !isValid {
		impl.logger.Errorw("registry credentials validation err, SaveDockerRegistryConfig", "err", err, "payload", bean)
		err = &util.ApiError{
			HttpStatusCode:  http.StatusBadRequest,
			InternalMessage: "Invalid authentication credentials. Please verify.",
			UserMessage:     "Invalid authentication credentials. Please verify.",
		}
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) GetDockerArtifactStore(w http.ResponseWriter, r *http.Request) {
	res, err := impl.dockerRegistryConfig.ListAllActive()
	if err != nil {
		impl.logger.Errorw("service err, GetDockerArtifactStore", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	var result []types.DockerArtifactStoreBean
	for _, item := range res {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionGet, item.Id); ok {
			result = append(result, item)
		}
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) FetchAllDockerAccounts(w http.ResponseWriter, r *http.Request) {
	res, err := impl.dockerRegistryConfig.FetchAllDockerAccounts()
	if err != nil {
		impl.logger.Errorw("service err, FetchAllDockerAccounts", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	var result []types.DockerArtifactStoreBean
	for _, item := range res {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionGet, item.Id); ok {
			item.DisabledFields = make([]types.DisabledFields, 0)
			if !item.IsPublic {
				if isEditable := impl.deleteService.CanDeleteChartRegistryPullConfig(item.Id); !(isEditable || item.IsPublic) {
					item.DisabledFields = append(item.DisabledFields, pipeline.DISABLED_CHART_PULL)
				}
			}
			result = append(result, item)
		}
	}
	//RBAC enforcer Ends
	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (impl DockerRegRestHandlerExtendedImpl) FetchAllDockerAccounts(w http.ResponseWriter, r *http.Request) {
	res, err := impl.dockerRegistryConfig.FetchAllDockerAccounts()
	if err != nil {
		impl.logger.Errorw("service err, FetchAllDockerAccounts", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	var result []types.DockerArtifactStoreBean
	for _, item := range res {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionGet, item.Id); ok {
			item.DisabledFields = make([]types.DisabledFields, 0)
			if !item.IsPublic {
				if isContainerEditable := impl.deleteServiceFullMode.CanDeleteContainerRegistryConfig(item.Id); !(isContainerEditable || item.IsPublic) {
					item.DisabledFields = append(item.DisabledFields, pipeline.DISABLED_CONTAINER)
				}
				if isChartEditable := impl.DockerRegRestHandlerImpl.deleteService.CanDeleteChartRegistryPullConfig(item.Id); !(isChartEditable || item.IsPublic) {
					item.DisabledFields = append(item.DisabledFields, pipeline.DISABLED_CHART_PULL)
				}
			}
			result = append(result, item)
		}
	}
	//RBAC enforcer Ends
	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) FetchOneDockerAccounts(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	res, err := impl.dockerRegistryConfig.FetchOneDockerAccount(id)
	if err != nil {
		impl.logger.Errorw("service err, FetchOneDockerAccounts", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	res.DisabledFields = make([]types.DisabledFields, 0)
	if !res.IsPublic {
		if isChartEditable := impl.deleteService.CanDeleteChartRegistryPullConfig(res.Id); !(isChartEditable || res.IsPublic) {
			res.DisabledFields = append(res.DisabledFields, pipeline.DISABLED_CONTAINER)
		}
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionGet, res.Id); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl DockerRegRestHandlerExtendedImpl) FetchOneDockerAccounts(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	res, err := impl.dockerRegistryConfig.FetchOneDockerAccount(id)
	if err != nil {
		impl.logger.Errorw("service err, FetchOneDockerAccounts", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	res.DisabledFields = make([]types.DisabledFields, 0)
	if !res.IsPublic {
		if isContainerEditable := impl.deleteServiceFullMode.CanDeleteContainerRegistryConfig(res.Id); !(isContainerEditable || res.IsPublic) {
			res.DisabledFields = append(res.DisabledFields, pipeline.DISABLED_CONTAINER)
		}
		if isChartEditable := impl.DockerRegRestHandlerImpl.deleteService.CanDeleteChartRegistryPullConfig(res.Id); !(isChartEditable || res.IsPublic) {
			res.DisabledFields = append(res.DisabledFields, pipeline.DISABLED_CHART_PULL)
		}
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionGet, res.Id); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) UpdateDockerRegistryConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean types.DockerArtifactStoreBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, UpdateDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.User = userId
	requestErr := ValidateDockerArtifactStoreRequestBean(bean)
	if requestErr != nil {
		err = fmt.Errorf("invalid payload, missing or incorrect values for required fields")
		impl.logger.Errorw("validation err, SaveDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, requestErr, nil, http.StatusBadRequest)
		return
	}

	impl.logger.Infow("request payload, UpdateDockerRegistryConfig", "err", err, "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, UpdateDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionUpdate, bean.Id); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	res, err := impl.dockerRegistryConfig.Update(&bean)
	if err != nil {
		impl.logger.Errorw("service err, UpdateDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// trigger a chart sync job
	if res.IsOCICompliantRegistry && len(res.RepositoryList) != 0 {
		request := &chartProviderService.ChartProviderRequestDto{
			Id:            res.Id,
			IsOCIRegistry: res.IsOCICompliantRegistry,
		}
		err = impl.chartProviderService.SyncChartProvider(request)
		if err != nil {
			impl.logger.Errorw("service err, SaveDockerRegistryConfig", "err", err, "userId", userId)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) FetchAllDockerRegistryForAutocomplete(w http.ResponseWriter, r *http.Request) {
	res, err := impl.dockerRegistryConfig.ListAllActive()
	if err != nil {
		impl.logger.Errorw("service err, FetchAllDockerRegistryForAutocomplete", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) IsDockerRegConfigured(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	storageType := v.Get("storageType")
	if storageType == "" {
		storageType = repository.OCI_REGISRTY_REPO_TYPE_CONTAINER
	}
	if !slices.Contains(repository.OCI_REGISRTY_REPO_TYPE_LIST, storageType) {
		common.WriteJsonResp(w, fmt.Errorf("invalid query parameters"), nil, http.StatusBadRequest)
		return
	}
	storageAction := v.Get("storageAction")
	if storageAction == "" {
		storageAction = repository.STORAGE_ACTION_TYPE_PUSH
	}
	if !(storageAction == repository.STORAGE_ACTION_TYPE_PULL || storageAction == repository.STORAGE_ACTION_TYPE_PUSH) {
		common.WriteJsonResp(w, fmt.Errorf("invalid query parameters"), nil, http.StatusBadRequest)
		return
	}
	isConfigured := false
	registryConfigs, err := impl.dockerRegistryConfig.ListAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("service err, IsDockerRegConfigured", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if len(registryConfigs) > 0 {
		// Filter out all registries with CONTAINER push or pull/push access
		res := impl.dockerRegistryConfig.FilterRegistryBeanListBasedOnStorageTypeAndAction(registryConfigs, storageType, storageAction, repository.STORAGE_ACTION_TYPE_PULL_AND_PUSH)
		if len(res) > 0 {
			isConfigured = true
		}
	}

	common.WriteJsonResp(w, err, isConfigured, http.StatusOK)
}

func (impl DockerRegRestHandlerImpl) DeleteDockerRegistryConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean types.DockerArtifactStoreBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, DeleteDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.User = userId
	impl.logger.Infow("request payload, DeleteDockerRegistryConfig", "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, DeleteDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionCreate, bean.Id); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	err = impl.deleteService.DeleteDockerRegistryConfig(&bean)
	if err != nil {
		impl.logger.Errorw("service err, DeleteDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, REG_DELETE_SUCCESS_RESP, http.StatusOK)
}

func (impl DockerRegRestHandlerExtendedImpl) DeleteDockerRegistryConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean types.DockerArtifactStoreBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, DeleteDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.User = userId
	impl.logger.Infow("request payload, DeleteDockerRegistryConfig", "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, DeleteDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceDocker, casbin.ActionCreate, bean.Id); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	err = impl.deleteServiceFullMode.DeleteDockerRegistryConfig(&bean)
	if err != nil {
		impl.logger.Errorw("service err, DeleteDockerRegistryConfig", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, REG_DELETE_SUCCESS_RESP, http.StatusOK)
}
