/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/helm-app/bean"
	service2 "github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	"github.com/devtron-labs/devtron/pkg/argoApplication"
	"github.com/devtron-labs/devtron/pkg/argoApplication/helper"
	clusterBean "github.com/devtron-labs/devtron/pkg/cluster/bean"
	clientErrors "github.com/devtron-labs/devtron/pkg/errors"
	"github.com/devtron-labs/devtron/pkg/fluxApplication"
	bean2 "github.com/devtron-labs/devtron/pkg/k8s/application/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devtron-labs/common-lib/utils/k8sObjectsUtil"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/cluster"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type HelmAppRestHandler interface {
	ListApplications(w http.ResponseWriter, r *http.Request)
	GetApplicationDetail(w http.ResponseWriter, r *http.Request)
	Hibernate(w http.ResponseWriter, r *http.Request)
	UnHibernate(w http.ResponseWriter, r *http.Request)
	GetReleaseInfo(w http.ResponseWriter, r *http.Request)
	GetDesiredManifest(w http.ResponseWriter, r *http.Request)
	DeleteApplication(w http.ResponseWriter, r *http.Request)
	UpdateApplication(w http.ResponseWriter, r *http.Request)
	TemplateChart(w http.ResponseWriter, r *http.Request)
	SaveHelmAppDetailsViewedTelemetryData(w http.ResponseWriter, r *http.Request)
	ListHelmApplicationsForEnvironment(w http.ResponseWriter, r *http.Request)
}

const HELM_APP_ACCESS_COUNTER = "HelmAppAccessCounter"
const HELM_APP_UPDATE_COUNTER = "HelmAppUpdateCounter"

type HelmAppRestHandlerImpl struct {
	logger                    *zap.SugaredLogger
	helmAppService            service2.HelmAppService
	enforcer                  casbin.Enforcer
	clusterService            cluster.ClusterService
	enforcerUtil              rbac.EnforcerUtilHelm
	appStoreDeploymentService service.AppStoreDeploymentService
	installedAppService       EAMode.InstalledAppDBService
	userAuthService           user.UserService
	attributesService         attributes.AttributesService
	serverEnvConfig           *serverEnvConfig.ServerEnvConfig
	fluxApplication           fluxApplication.FluxApplicationService
	argoApplication           argoApplication.ArgoApplicationService
}

func NewHelmAppRestHandlerImpl(logger *zap.SugaredLogger,
	helmAppService service2.HelmAppService, enforcer casbin.Enforcer,
	clusterService cluster.ClusterService, enforcerUtil rbac.EnforcerUtilHelm,
	appStoreDeploymentService service.AppStoreDeploymentService, installedAppService EAMode.InstalledAppDBService,
	userAuthService user.UserService, attributesService attributes.AttributesService, serverEnvConfig *serverEnvConfig.ServerEnvConfig, fluxApplication fluxApplication.FluxApplicationService, argoApplication argoApplication.ArgoApplicationService,
) *HelmAppRestHandlerImpl {
	return &HelmAppRestHandlerImpl{
		logger:                    logger,
		helmAppService:            helmAppService,
		enforcer:                  enforcer,
		clusterService:            clusterService,
		enforcerUtil:              enforcerUtil,
		appStoreDeploymentService: appStoreDeploymentService,
		installedAppService:       installedAppService,
		userAuthService:           userAuthService,
		attributesService:         attributesService,
		serverEnvConfig:           serverEnvConfig,
		fluxApplication:           fluxApplication,
		argoApplication:           argoApplication,
	}
}

func (handler *HelmAppRestHandlerImpl) ListApplications(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterIdString := vars["clusterIds"]
	clusterIdSlices := strings.Split(clusterIdString, ",")
	var clusterIds []int
	for _, is := range clusterIdSlices {
		if len(is) == 0 {
			continue
		}
		j, err := strconv.Atoi(is)
		if err != nil {
			handler.logger.Errorw("request err, CreateUser", "err", err, "payload", clusterIds)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		clusterIds = append(clusterIds, j)
	}
	token := r.Header.Get("token")
	handler.helmAppService.ListHelmApplications(r.Context(), clusterIds, w, token, handler.checkHelmAuth)
}

func (handler *HelmAppRestHandlerImpl) GetApplicationDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterIdString := vars["appId"]

	appIdentifier, err := handler.helmAppService.DecodeAppId(clusterIdString)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject, rbacObject2 := handler.enforcerUtil.GetHelmObjectByClusterIdNamespaceAndAppName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")

	ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject2)

	if !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	appdetail, err := handler.helmAppService.GetApplicationDetail(context.Background(), appIdentifier)
	if err != nil {

		if pipeline.CheckAppReleaseNotExist(err) {
			common.WriteJsonResp(w, err, nil, http.StatusNotFound)
			return
		}
		apiError := clientErrors.ConvertToApiError(err)
		if apiError != nil {
			err = apiError
		}
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	installedApp, err := handler.installedAppService.GetInstalledAppByClusterNamespaceAndName(appIdentifier)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	res := &bean.AppDetailAndInstalledAppInfo{
		AppDetail:        appdetail,
		InstalledAppInfo: bean.ConvertToInstalledAppInfo(installedApp),
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) Hibernate(w http.ResponseWriter, r *http.Request) {
	hibernateRequest, err := decodeHibernateRequest(r)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	var appType int
	appTypeString := vars["appType"] //to check the type of app
	if appTypeString == "" {
		appType = 1
	} else {
		appType, err = strconv.Atoi(appTypeString)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	token := r.Header.Get("token")
	var res []*openapi.HibernateStatus

	if appType == bean2.ArgoAppType {
		res, err = handler.handleArgoApplicationHibernate(r, token, hibernateRequest)
	} else if appType == bean2.HelmAppType {
		res, err = handler.handleHelmApplicationHibernate(r, token, hibernateRequest)
	} else if appType == bean2.FluxAppType {
		res, err = handler.handleFluxApplicationHibernate(r, token, hibernateRequest)
	}

	if err != nil {
		common.WriteJsonResp(w, err, nil, service2.GetStatusCode(err))
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func decodeHibernateRequest(r *http.Request) (*openapi.HibernateRequest, error) {
	hibernateRequest := &openapi.HibernateRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(hibernateRequest)
	if err != nil {
		return nil, err
	}
	return hibernateRequest, nil
}

func (handler *HelmAppRestHandlerImpl) handleFluxApplicationHibernate(r *http.Request, token string, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {
	appIdentifier, err := fluxApplication.DecodeFluxExternalAppId(*hibernateRequest.AppId)
	if err != nil {
		return nil, err
	}

	if !handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*") {
		return nil, errors.New("unauthorized")
	}

	return handler.fluxApplication.HibernateFluxApplication(r.Context(), appIdentifier, hibernateRequest)
}
func (handler *HelmAppRestHandlerImpl) handleArgoApplicationHibernate(r *http.Request, token string, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {
	appIdentifier, err := helper.DecodeExternalArgoAppId(*hibernateRequest.AppId)
	if err != nil {
		return nil, err
	}

	if !handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*") {
		return nil, errors.New("unauthorized")
	}

	return handler.argoApplication.HibernateArgoApplication(r.Context(), appIdentifier, hibernateRequest)
}

func (handler *HelmAppRestHandlerImpl) handleHelmApplicationHibernate(r *http.Request, token string, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {
	appIdentifier, err := handler.helmAppService.DecodeAppId(*hibernateRequest.AppId)
	if err != nil {
		return nil, err
	}
	rbacObject, rbacObject2 := handler.enforcerUtil.GetHelmObjectByClusterIdNamespaceAndAppName(
		appIdentifier.ClusterId,
		appIdentifier.Namespace,
		appIdentifier.ReleaseName,
	)

	ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject) ||
		handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject2)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	return handler.helmAppService.HibernateApplication(r.Context(), appIdentifier, hibernateRequest)
}
func (handler *HelmAppRestHandlerImpl) UnHibernate(w http.ResponseWriter, r *http.Request) {
	hibernateRequest, err := decodeHibernateRequest(r)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	var appType int
	appTypeString := vars["appType"] //to check the type of app currently handled the case for identifying of external app types
	if appTypeString == "" {
		appType = 1
	} else {
		appType, err = strconv.Atoi(appTypeString)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	token := r.Header.Get("token")
	var res []*openapi.HibernateStatus

	if appType == bean2.ArgoAppType {
		res, err = handler.handleArgoApplicationUnHibernate(r, token, hibernateRequest)
	} else if appType == bean2.HelmAppType {
		res, err = handler.handleHelmApplicationUnHibernate(r, token, hibernateRequest)
	} else if appType == bean2.FluxAppType {
		res, err = handler.handleFluxApplicationUnHibernate(r, token, hibernateRequest)
	}
	//if k8s.IsClusterStringContainsFluxField(*hibernateRequest.AppId) {
	//	res, err = handler.handleFluxApplicationUnHibernate(r, token, hibernateRequest)
	//} else {
	//	res, err = handler.handleHelmApplicationUnHibernate(r, token, hibernateRequest)
	//}

	if err != nil {
		common.WriteJsonResp(w, err, nil, service2.GetStatusCode(err))
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) handleFluxApplicationUnHibernate(r *http.Request, token string, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {
	appIdentifier, err := fluxApplication.DecodeFluxExternalAppId(*hibernateRequest.AppId)
	if err != nil {
		return nil, err
	}
	if !handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*") {
		return nil, errors.New("unauthorized")
	}
	return handler.fluxApplication.UnHibernateFluxApplication(r.Context(), appIdentifier, hibernateRequest)
}
func (handler *HelmAppRestHandlerImpl) handleArgoApplicationUnHibernate(r *http.Request, token string, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {
	appIdentifier, err := helper.DecodeExternalArgoAppId(*hibernateRequest.AppId)
	if err != nil {
		return nil, err
	}
	if !handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*") {
		return nil, errors.New("unauthorized")
	}
	return handler.argoApplication.UnHibernateArgoApplication(r.Context(), appIdentifier, hibernateRequest)
}

func (handler *HelmAppRestHandlerImpl) handleHelmApplicationUnHibernate(r *http.Request, token string, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {
	appIdentifier, err := handler.helmAppService.DecodeAppId(*hibernateRequest.AppId)
	if err != nil {
		return nil, err
	}

	rbacObject, rbacObject2 := handler.enforcerUtil.GetHelmObjectByClusterIdNamespaceAndAppName(
		appIdentifier.ClusterId,
		appIdentifier.Namespace,
		appIdentifier.ReleaseName,
	)

	ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject) ||
		handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject2)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	return handler.helmAppService.UnHibernateApplication(r.Context(), appIdentifier, hibernateRequest)
}

func (handler *HelmAppRestHandlerImpl) GetReleaseInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId := vars["appId"]
	appIdentifier, err := handler.helmAppService.DecodeAppId(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject, rbacObject2 := handler.enforcerUtil.GetHelmObjectByClusterIdNamespaceAndAppName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")

	ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject2)

	if !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	releaseInfo, err := handler.helmAppService.GetValuesYaml(r.Context(), appIdentifier)
	if err != nil {
		apiError := clientErrors.ConvertToApiError(err)
		if apiError != nil {
			err = apiError
		}
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	installedAppVersionDto, err := handler.installedAppService.GetReleaseInfo(appIdentifier)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	res := &bean.ReleaseAndInstalledAppInfo{
		ReleaseInfo:      releaseInfo,
		InstalledAppInfo: bean.ConvertToInstalledAppInfo(installedAppVersionDto),
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) GetDesiredManifest(w http.ResponseWriter, r *http.Request) {
	desiredManifestRequest := &openapi.DesiredManifestRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(desiredManifestRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	appIdentifier, err := handler.helmAppService.DecodeAppId(*desiredManifestRequest.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject, rbacObject2 := handler.enforcerUtil.GetHelmObjectByClusterIdNamespaceAndAppName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")

	ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject2)

	if !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	res, err := handler.helmAppService.GetDesiredManifest(r.Context(), appIdentifier, desiredManifestRequest.Resource)
	if err != nil {
		apiError := clientErrors.ConvertToApiError(err)
		if apiError != nil {
			err = apiError
		}
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// Obfuscate secret if user does not have edit access
	canUpdate := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject)
	if !canUpdate && res != nil && res.Manifest != nil {
		modifiedManifest, err := k8sObjectsUtil.HideValuesIfSecretForManifestStringInput(*res.Manifest, *desiredManifestRequest.Resource.Kind, *desiredManifestRequest.Resource.Group)
		if err != nil {
			handler.logger.Errorw("error in hiding secret values", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		res.Manifest = &modifiedManifest
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId := vars["appId"]
	appIdentifier, err := handler.helmAppService.DecodeAppId(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject, rbacObject2 := handler.enforcerUtil.GetHelmObjectByClusterIdNamespaceAndAppName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")

	ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionDelete, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionDelete, rbacObject2)

	if !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	// validate if the devtron-operator helm release, block that for deletion
	if appIdentifier.ReleaseName == handler.serverEnvConfig.DevtronHelmReleaseName &&
		appIdentifier.Namespace == handler.serverEnvConfig.DevtronHelmReleaseNamespace &&
		appIdentifier.ClusterId == clusterBean.DefaultClusterId {
		common.WriteJsonResp(w, errors.New("cannot delete this default helm app"), nil, http.StatusForbidden)
		return
	}

	res, err := handler.helmAppService.DeleteDBLinkedHelmApplication(r.Context(), appIdentifier, userId)
	if util.IsErrNoRows(err) {
		res, err = handler.helmAppService.DeleteApplication(r.Context(), appIdentifier)
	}
	if err != nil {
		apiError := clientErrors.ConvertToApiError(err)
		if apiError != nil {
			err = apiError
		}
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	request := &bean.UpdateApplicationRequestDto{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(request)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	appIdentifier, err := handler.helmAppService.DecodeAppId(*request.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject, rbacObject2 := handler.enforcerUtil.GetHelmObjectByClusterIdNamespaceAndAppName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")

	ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject2)

	if !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	request.SourceAppType = bean.SOURCE_EXTERNAL_HELM_APP
	// update application externally
	res, err := handler.helmAppService.UpdateApplication(r.Context(), appIdentifier, request)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	err = handler.attributesService.UpdateKeyValueByOne(HELM_APP_UPDATE_COUNTER)
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) TemplateChart(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	request := &openapi2.TemplateChartRequest{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(request)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//making this api rbac free

	// template chart starts
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	response, err := handler.helmAppService.TemplateChart(ctx, request)
	if err != nil {
		handler.logger.Errorw("Error in helm-template", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, response, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) checkHelmAuth(token string, object string) bool {
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, object); !ok {
		return false
	}
	return true
}

func (handler *HelmAppRestHandlerImpl) SaveHelmAppDetailsViewedTelemetryData(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	clusterIdString := vars["appId"]

	appIdentifier, err := handler.helmAppService.DecodeAppId(clusterIdString)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject, rbacObject2 := handler.enforcerUtil.GetHelmObjectByClusterIdNamespaceAndAppName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject2)
	if !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	// RBAC enforcer ends
	err = handler.attributesService.UpdateKeyValueByOne(HELM_APP_ACCESS_COUNTER)

	if err != nil {
		handler.logger.Errorw("error in saving external helm apps details visited count")
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, nil, http.StatusOK)

}

func (handler *HelmAppRestHandlerImpl) ListHelmApplicationsForEnvironment(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()

	clusterIdString := query.Get("clusterId")
	var (
		clusterId int
		envId     int
		err       error
	)

	if len(clusterIdString) != 0 {
		clusterId, err = strconv.Atoi(clusterIdString)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	envIdString := query.Get("envId")
	if len(envIdString) != 0 {
		envId, err = strconv.Atoi(envIdString)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	releaseList, err := handler.helmAppService.ListHelmApplicationsForClusterOrEnv(r.Context(), clusterId, envId)
	if err != nil {
		handler.logger.Errorw("error in fetching helm release for given env", "err", err)
		common.WriteJsonResp(w, err, "error in fetching helm release", http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, releaseList, http.StatusOK)
	return
}
