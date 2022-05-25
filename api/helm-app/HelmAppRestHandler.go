package client

import (
	"context"
	"encoding/json"
	"errors"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/deployment/common"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/k8sObjectsUtil"
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
}

type HelmAppRestHandlerImpl struct {
	logger                          *zap.SugaredLogger
	helmAppService                  HelmAppService
	enforcer                        casbin.Enforcer
	clusterService                  cluster.ClusterService
	enforcerUtil                    rbac.EnforcerUtilHelm
	appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService
	userAuthService                 user.UserService
}

func NewHelmAppRestHandlerImpl(logger *zap.SugaredLogger,
	helmAppService HelmAppService, enforcer casbin.Enforcer,
	clusterService cluster.ClusterService, enforcerUtil rbac.EnforcerUtilHelm, appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService,
	userAuthService user.UserService) *HelmAppRestHandlerImpl {
	return &HelmAppRestHandlerImpl{
		logger:                          logger,
		helmAppService:                  helmAppService,
		enforcer:                        enforcer,
		clusterService:                  clusterService,
		enforcerUtil:                    enforcerUtil,
		appStoreDeploymentCommonService: appStoreDeploymentCommonService,
		userAuthService:                 userAuthService,
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
	handler.helmAppService.ListHelmApplications(clusterIds, w, token, handler.checkHelmAuth)
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
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject); !ok {
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
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	res, err := handler.helmAppService.HibernateApplication(context.Background(), appIdentifier, hibernateRequest)
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
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	res, err := handler.helmAppService.UnHibernateApplication(context.Background(), appIdentifier, hibernateRequest)
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
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	releaseInfo, err := handler.helmAppService.GetValuesYaml(context.Background(), appIdentifier)
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
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	res, err := handler.helmAppService.GetDesiredManifest(context.Background(), appIdentifier, desiredManifestRequest.Resource)
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
	vars := mux.Vars(r)
	appId := vars["appId"]
	appIdentifier, err := handler.helmAppService.DecodeAppId(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionDelete, rbacObject); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	res, err := handler.helmAppService.DeleteApplication(context.Background(), appIdentifier)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	request := &openapi.UpdateReleaseRequest{}
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
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	installedApp, err := handler.appStoreDeploymentCommonService.GetInstalledAppByClusterNamespaceAndName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	updateReleaseRequest := &InstallReleaseRequest{
		ValuesYaml: request.GetValuesYaml(),
		ReleaseIdentifier: &ReleaseIdentifier{
			ReleaseNamespace: appIdentifier.Namespace,
			ReleaseName:      appIdentifier.ReleaseName,
		},
	}

	var res *openapi.UpdateReleaseResponse

	if installedApp != nil {
		chartInfo := installedApp.InstallAppVersionChartDTO
		chartRepoInfo := chartInfo.InstallAppVersionChartRepoDTO
		updateReleaseRequest.ChartName = chartInfo.ChartName
		updateReleaseRequest.ChartVersion = chartInfo.ChartVersion
		updateReleaseRequest.ChartRepository = &ChartRepository{
			Name:     chartRepoInfo.RepoName,
			Url:      chartRepoInfo.RepoUrl,
			Username: chartRepoInfo.UserName,
			Password: chartRepoInfo.Password,
		}
		res, err = handler.helmAppService.UpdateApplicationWithChartInfo(context.Background(), appIdentifier.ClusterId, updateReleaseRequest)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	} else {
		res, err = handler.helmAppService.UpdateApplication(context.Background(), appIdentifier, request)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}

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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

func convertToInstalledAppInfo(installedApp *appStoreBean.InstallAppVersionDTO) *InstalledAppInfo {
	if installedApp == nil {
		return nil
	}

	return &InstalledAppInfo{
		AppId:                 installedApp.AppId,
		EnvironmentName:       installedApp.EnvironmentName,
		AppOfferingMode:       installedApp.AppOfferingMode,
		InstalledAppId:        installedApp.InstalledAppId,
		InstalledAppVersionId: installedApp.InstalledAppVersionId,
		AppStoreChartId:       installedApp.InstallAppVersionChartDTO.AppStoreChartId,
		ClusterId:             installedApp.ClusterId,
		EnvironmentId:         installedApp.EnvironmentId,
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
}
