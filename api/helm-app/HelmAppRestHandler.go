package client

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/devtron-labs/common-lib-private/utils/k8sObjectsUtil"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/deployment/common"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/cluster"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
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
}

const HELM_APP_ACCESS_COUNTER = "HelmAppAccessCounter"
const HELM_APP_UPDATE_COUNTER = "HelmAppUpdateCounter"

type HelmAppRestHandlerImpl struct {
	logger                          *zap.SugaredLogger
	helmAppService                  HelmAppService
	enforcer                        casbin.Enforcer
	clusterService                  cluster.ClusterService
	enforcerUtil                    rbac.EnforcerUtilHelm
	appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService
	userAuthService                 user.UserService
	attributesService               attributes.AttributesService
	serverEnvConfig                 *serverEnvConfig.ServerEnvConfig
}

func NewHelmAppRestHandlerImpl(logger *zap.SugaredLogger,
	helmAppService HelmAppService, enforcer casbin.Enforcer,
	clusterService cluster.ClusterService, enforcerUtil rbac.EnforcerUtilHelm, appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService,
	userAuthService user.UserService, attributesService attributes.AttributesService, serverEnvConfig *serverEnvConfig.ServerEnvConfig) *HelmAppRestHandlerImpl {
	return &HelmAppRestHandlerImpl{
		logger:                          logger,
		helmAppService:                  helmAppService,
		enforcer:                        enforcer,
		clusterService:                  clusterService,
		enforcerUtil:                    enforcerUtil,
		appStoreDeploymentCommonService: appStoreDeploymentCommonService,
		userAuthService:                 userAuthService,
		attributesService:               attributesService,
		serverEnvConfig:                 serverEnvConfig,
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
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	installedApp, err := handler.appStoreDeploymentCommonService.GetInstalledAppByClusterNamespaceAndName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	res := &AppDetailAndInstalledAppInfo{
		AppDetail:        appdetail,
		InstalledAppInfo: convertToInstalledAppInfo(installedApp),
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) Hibernate(w http.ResponseWriter, r *http.Request) {
	hibernateRequest := &openapi.HibernateRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(hibernateRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	appIdentifier, err := handler.helmAppService.DecodeAppId(*hibernateRequest.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	rbacObject, rbacObject2 := handler.enforcerUtil.GetHelmObjectByClusterIdNamespaceAndAppName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject2)
	if !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	res, err := handler.helmAppService.HibernateApplication(r.Context(), appIdentifier, hibernateRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) UnHibernate(w http.ResponseWriter, r *http.Request) {
	hibernateRequest := &openapi.HibernateRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(hibernateRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	appIdentifier, err := handler.helmAppService.DecodeAppId(*hibernateRequest.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject, rbacObject2 := handler.enforcerUtil.GetHelmObjectByClusterIdNamespaceAndAppName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")

	ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject2)

	if !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	res, err := handler.helmAppService.UnHibernateApplication(r.Context(), appIdentifier, hibernateRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
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
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	installedApp, err := handler.appStoreDeploymentCommonService.GetInstalledAppByClusterNamespaceAndName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	res := &ReleaseAndInstalledAppInfo{
		ReleaseInfo:      releaseInfo,
		InstalledAppInfo: convertToInstalledAppInfo(installedApp),
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
		appIdentifier.ClusterId == DEFAULT_CLUSTER_ID {
		common.WriteJsonResp(w, errors.New("cannot delete this default helm app"), nil, http.StatusForbidden)
		return
	}

	res, err := handler.helmAppService.DeleteDBLinkedHelmApplication(r.Context(), appIdentifier, userId)
	if util.IsErrNoRows(err) {
		res, err = handler.helmAppService.DeleteApplication(r.Context(), appIdentifier)
	}
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	request := &UpdateApplicationRequestDto{}
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
	request.SourceAppType = SOURCE_EXTERNAL_HELM_APP
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
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, strings.ToLower(object)); !ok {
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

func convertToInstalledAppInfo(installedApp *appStoreBean.InstallAppVersionDTO) *InstalledAppInfo {
	if installedApp == nil {
		return nil
	}

	chartInfo := installedApp.InstallAppVersionChartDTO

	return &InstalledAppInfo{
		AppId:                 installedApp.AppId,
		EnvironmentName:       installedApp.EnvironmentName,
		AppOfferingMode:       installedApp.AppOfferingMode,
		InstalledAppId:        installedApp.InstalledAppId,
		InstalledAppVersionId: installedApp.InstalledAppVersionId,
		AppStoreChartId:       chartInfo.AppStoreChartId,
		ClusterId:             installedApp.ClusterId,
		EnvironmentId:         installedApp.EnvironmentId,
		AppStoreChartRepoName: chartInfo.InstallAppVersionChartRepoDTO.RepoName,
		AppStoreChartName:     chartInfo.ChartName,
		TeamId:                installedApp.TeamId,
		TeamName:              installedApp.TeamName,
	}
}

type AppDetailAndInstalledAppInfo struct {
	InstalledAppInfo *InstalledAppInfo `json:"installedAppInfo"`
	AppDetail        *AppDetail        `json:"appDetail"`
}

type ReleaseAndInstalledAppInfo struct {
	InstalledAppInfo *InstalledAppInfo `json:"installedAppInfo"`
	ReleaseInfo      *ReleaseInfo      `json:"releaseInfo"`
}

type DeploymentHistoryAndInstalledAppInfo struct {
	InstalledAppInfo  *InstalledAppInfo          `json:"installedAppInfo"`
	DeploymentHistory []*HelmAppDeploymentDetail `json:"deploymentHistory"`
}

type InstalledAppInfo struct {
	AppId                 int    `json:"appId"`
	InstalledAppId        int    `json:"installedAppId"`
	InstalledAppVersionId int    `json:"installedAppVersionId"`
	AppStoreChartId       int    `json:"appStoreChartId"`
	EnvironmentName       string `json:"environmentName"`
	AppOfferingMode       string `json:"appOfferingMode"`
	ClusterId             int    `json:"clusterId"`
	EnvironmentId         int    `json:"environmentId"`
	AppStoreChartRepoName string `json:"appStoreChartRepoName"`
	AppStoreChartName     string `json:"appStoreChartName"`
	TeamId                int    `json:"teamId"`
	TeamName              string `json:"teamName"`
	DeploymentType        string `json:"deploymentType"`
	HelmPackageName       string `json:"helmPackageName"`
}
