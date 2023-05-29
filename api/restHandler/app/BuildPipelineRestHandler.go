package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/util"
	appGroup2 "github.com/devtron-labs/devtron/pkg/appGroup"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	bean1 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const GIT_MATERIAL_DELETE_SUCCESS_RESP = "Git material deleted successfully."

type DevtronAppBuildRestHandler interface {
	CreateCiConfig(w http.ResponseWriter, r *http.Request)
	UpdateCiTemplate(w http.ResponseWriter, r *http.Request)

	GetCiPipeline(w http.ResponseWriter, r *http.Request)
	GetExternalCi(w http.ResponseWriter, r *http.Request)
	GetExternalCiById(w http.ResponseWriter, r *http.Request)
	PatchCiPipelines(w http.ResponseWriter, r *http.Request)
	TriggerCiPipeline(w http.ResponseWriter, r *http.Request)
	GetCiPipelineMin(w http.ResponseWriter, r *http.Request)
	GetCIPipelineById(w http.ResponseWriter, r *http.Request)
	HandleWorkflowWebhook(w http.ResponseWriter, r *http.Request)
	GetBuildLogs(w http.ResponseWriter, r *http.Request)
	FetchWorkflowDetails(w http.ResponseWriter, r *http.Request)
	// CancelWorkflow CancelBuild
	CancelWorkflow(w http.ResponseWriter, r *http.Request)

	UpdateBranchCiPipelinesWithRegex(w http.ResponseWriter, r *http.Request)
	GetCiPipelineByEnvironment(w http.ResponseWriter, r *http.Request)
	GetCiPipelineByEnvironmentMin(w http.ResponseWriter, r *http.Request)
	GetExternalCiByEnvironment(w http.ResponseWriter, r *http.Request)
}

type DevtronAppBuildMaterialRestHandler interface {
	CreateMaterial(w http.ResponseWriter, r *http.Request)
	UpdateMaterial(w http.ResponseWriter, r *http.Request)
	FetchMaterials(w http.ResponseWriter, r *http.Request)
	FetchMaterialsByMaterialId(w http.ResponseWriter, r *http.Request)
	RefreshMaterials(w http.ResponseWriter, r *http.Request)
	FetchMaterialInfo(w http.ResponseWriter, r *http.Request)
	FetchChanges(w http.ResponseWriter, r *http.Request)
	DeleteMaterial(w http.ResponseWriter, r *http.Request)
	GetCommitMetadataForPipelineMaterial(w http.ResponseWriter, r *http.Request)
}

type DevtronAppBuildHistoryRestHandler interface {
	GetHistoricBuildLogs(w http.ResponseWriter, r *http.Request)
	GetBuildHistory(w http.ResponseWriter, r *http.Request)
	DownloadCiWorkflowArtifacts(w http.ResponseWriter, r *http.Request)
}

type ImageTaggingRestHandler interface {
	CreateImageTagging(w http.ResponseWriter, r *http.Request)
	GetImageTaggingData(w http.ResponseWriter, r *http.Request)
}

func (handler PipelineConfigRestHandlerImpl) CreateCiConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var createRequest bean.CiConfigRequest
	err = decoder.Decode(&createRequest)
	createRequest.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, create ci config", "err", err, "create request", createRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, create ci config", "create request", createRequest)
	err = handler.validator.Struct(createRequest)
	if err != nil {
		handler.Logger.Errorw("validation err, create ci config", "err", err, "create request", createRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(createRequest.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.pipelineBuilder.CreateCiPipeline(&createRequest)
	if err != nil {
		handler.Logger.Errorw("service err, create", "err", err, "create request", createRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) UpdateCiTemplate(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var configRequest bean.CiConfigRequest
	err = decoder.Decode(&configRequest)
	configRequest.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, UpdateCiTemplate", "err", err, "UpdateCiTemplate", configRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, update ci template", "UpdateCiTemplate", configRequest, "userId", userId)
	err = handler.validator.Struct(configRequest)
	if err != nil {
		handler.Logger.Errorw("validation err, UpdateCiTemplate", "err", err, "UpdateCiTemplate", configRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(configRequest.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.pipelineBuilder.UpdateCiTemplate(&configRequest)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateCiTemplate", "err", err, "UpdateCiTemplate", configRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) UpdateBranchCiPipelinesWithRegex(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var patchRequest bean.CiRegexPatchRequest
	err = decoder.Decode(&patchRequest)
	patchRequest.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, PatchCiPipelines", "err", err, "PatchCiPipelines", patchRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	handler.Logger.Debugw("update request ", "req", patchRequest)
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(patchRequest.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	var materialList []*bean.CiPipelineMaterial
	for _, material := range patchRequest.CiPipelineMaterial {
		if handler.ciPipelineMaterialRepository.CheckRegexExistsForMaterial(material.Id) {
			materialList = append(materialList, material)
		}
	}
	if len(materialList) == 0 {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	patchRequest.CiPipelineMaterial = materialList

	err = handler.pipelineBuilder.PatchRegexCiPipeline(&patchRequest)
	if err != nil {
		handler.Logger.Errorw("service err, PatchCiPipelines", "err", err, "PatchCiPipelines", patchRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//if include/exclude configured showAll will include excluded materials also in list, if not configured it will ignore this flag
	resp, err := handler.ciHandler.FetchMaterialsByPipelineId(patchRequest.Id, false)
	if err != nil {
		handler.Logger.Errorw("service err, FetchMaterials", "err", err, "pipelineId", patchRequest.Id)
		common.WriteJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) PatchCiPipelines(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isSuperAdmin, err := handler.userAuthService.IsSuperAdmin(int(userId))
	if err != nil {
		common.WriteJsonResp(w, err, "failed to check if user is super admin", http.StatusInternalServerError)
		return
	}
	var patchRequest bean.CiPatchRequest
	err = decoder.Decode(&patchRequest)
	patchRequest.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, PatchCiPipelines", "err", err, "PatchCiPipelines", patchRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, PatchCiPipelines", "PatchCiPipelines", patchRequest)
	err = handler.validator.Struct(patchRequest)
	if err != nil {
		handler.Logger.Errorw("validation err", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Debugw("update request ", "req", patchRequest)
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(patchRequest.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	var ok bool
	if app.AppType == helper.Job {
		ok = isSuperAdmin
	} else {
		ok = handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName)
	}
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	ciConf, err := handler.pipelineBuilder.GetCiPipeline(patchRequest.AppId)

	var emptyDockerRegistry string
	if app.AppType == helper.Job && ciConf == nil {
		ciConfigRequest := bean.CiConfigRequest{}
		ciConfigRequest.DockerRegistry = emptyDockerRegistry
		ciConfigRequest.AppId = patchRequest.AppId
		ciConfigRequest.CiBuildConfig = &bean1.CiBuildConfigBean{}
		ciConfigRequest.CiBuildConfig.CiBuildType = "skip-build"
		ciConfigRequest.UserId = patchRequest.UserId
		if patchRequest.CiPipeline == nil || patchRequest.CiPipeline.CiMaterial == nil {
			handler.Logger.Errorw("Invalid patch ci-pipeline request", "request", patchRequest, "err", "invalid CiPipeline data")
			common.WriteJsonResp(w, fmt.Errorf("invalid CiPipeline data"), nil, http.StatusBadRequest)
			return
		}
		ciConfigRequest.CiBuildConfig.GitMaterialId = patchRequest.CiPipeline.CiMaterial[0].GitMaterialId
		ciConfigRequest.IsJob = true
		_, err = handler.pipelineBuilder.CreateCiPipeline(&ciConfigRequest)
		if err != nil {
			handler.Logger.Errorw("error occurred in creating ci-pipeline for the Job", "payload", ciConfigRequest, "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}

	createResp, err := handler.pipelineBuilder.PatchCiPipeline(&patchRequest)
	if err != nil {
		handler.Logger.Errorw("service err, PatchCiPipelines", "err", err, "PatchCiPipelines", patchRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if createResp != nil && app != nil {
		createResp.AppName = app.AppName
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCiPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCiPipeline", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	ciConf, err := handler.pipelineBuilder.GetCiPipeline(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCiPipeline", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if ciConf == nil || ciConf.Id == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no data found"}
	}
	common.WriteJsonResp(w, err, ciConf, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetExternalCi(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetExternalCi", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	ciConf, err := handler.pipelineBuilder.GetExternalCi(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetExternalCi", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, ciConf, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetExternalCiById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	externalCiId, err := strconv.Atoi(vars["externalCiId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetExternalCiById", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	ciConf, err := handler.pipelineBuilder.GetExternalCiById(appId, externalCiId)
	if err != nil {
		handler.Logger.Errorw("service err, GetExternalCiById", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, ciConf, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) TriggerCiPipeline(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var ciTriggerRequest bean.CiTriggerRequest
	err = decoder.Decode(&ciTriggerRequest)
	if err != nil {
		handler.Logger.Errorw("request err, TriggerCiPipeline", "err", err, "payload", ciTriggerRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if !handler.validForMultiMaterial(ciTriggerRequest) {
		handler.Logger.Errorw("invalid req, commit hash not present for multi-git", "payload", ciTriggerRequest)
		common.WriteJsonResp(w, errors.New("invalid req, commit hash not present for multi-git"),
			nil, http.StatusBadRequest)
	}
	ciTriggerRequest.TriggeredBy = userId
	token := r.Header.Get("token")
	userEmailId, err := handler.userAuthService.GetEmailFromToken(token)
	if err != nil {
		handler.Logger.Errorw("error in getting user emailId from token", "userId", userId, "err", err)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	handler.Logger.Infow("request payload, TriggerCiPipeline", "payload", ciTriggerRequest)

	//RBAC STARTS
	//checking if user has trigger access on app, if not will be forbidden to trigger independent of number of cd cdPipelines
	ciPipeline, err := handler.ciPipelineRepository.FindById(ciTriggerRequest.PipelineId)
	if err != nil {
		handler.Logger.Errorw("err in finding ci pipeline, TriggerCiPipeline", "err", err, "ciPipelineId", ciTriggerRequest.PipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	appObject := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if appRbacOk := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, appObject); !appRbacOk {
		handler.Logger.Debug(fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//checking rbac for cd cdPipelines
	cdPipelines, err := handler.pipelineRepository.FindByCiPipelineId(ciTriggerRequest.PipelineId)
	if err != nil {
		handler.Logger.Errorw("error in finding ccd cdPipelines by ciPipelineId", "err", err, "ciPipelineId", ciTriggerRequest.PipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	cdPipelineRbacObjects := make([]string, len(cdPipelines))
	for i, cdPipeline := range cdPipelines {
		envObject := handler.enforcerUtil.GetAppRBACByAppIdAndPipelineId(cdPipeline.AppId, cdPipeline.Id)
		cdPipelineRbacObjects[i] = envObject
	}
	envRbacResultMap := handler.enforcer.EnforceByEmailInBatch(userEmailId, casbin.ResourceEnvironment, casbin.ActionTrigger, cdPipelineRbacObjects)
	i := 0
	for _, rbacResultOk := range envRbacResultMap {
		if !rbacResultOk {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
		i++
	}
	//RBAC ENDS
	response := make(map[string]string)
	resp, err := handler.ciHandler.HandleCIManual(ciTriggerRequest)
	if err != nil {
		handler.Logger.Errorw("service err, TriggerCiPipeline", "err", err, "payload", ciTriggerRequest)
		common.WriteJsonResp(w, err, response, http.StatusInternalServerError)
		return
	}
	response["apiResponse"] = strconv.Itoa(resp)
	common.WriteJsonResp(w, err, response, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) FetchMaterials(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	v := r.URL.Query()
	showAll := false
	show := v.Get("showAll")
	if len(show) > 0 {
		showAll, err = strconv.ParseBool(show)
		if err != nil {
			showAll = true
			err = nil
			//ignore error, apply rbac by default
		}
	}
	handler.Logger.Infow("request payload, FetchMaterials", "pipelineId", pipelineId)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateCiTemplate", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	resp, err := handler.ciHandler.FetchMaterialsByPipelineId(pipelineId, showAll)
	if err != nil {
		handler.Logger.Errorw("service err, FetchMaterials", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) FetchMaterialsByMaterialId(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	gitMaterialId, err := strconv.Atoi(vars["gitMaterialId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	v := r.URL.Query()
	showAll := false
	show := v.Get("showAll")
	if len(show) > 0 {
		showAll, err = strconv.ParseBool(show)
		if err != nil {
			showAll = true
			err = nil
			//ignore error, apply rbac by default
		}
	}
	handler.Logger.Infow("request payload, FetchMaterials", "pipelineId", pipelineId)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateCiTemplate", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	resp, err := handler.ciHandler.FetchMaterialsByPipelineIdAndGitMaterialId(pipelineId, gitMaterialId, showAll)
	if err != nil {
		handler.Logger.Errorw("service err, FetchMaterials", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) RefreshMaterials(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	gitMaterialId, err := strconv.Atoi(vars["gitMaterialId"])
	if err != nil {
		handler.Logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, RefreshMaterials", "gitMaterialId", gitMaterialId)
	material, err := handler.materialRepository.FindById(gitMaterialId)
	if err != nil {
		handler.Logger.Errorw("service err, RefreshMaterials", "err", err, "gitMaterialId", gitMaterialId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(material.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	resp, err := handler.ciHandler.RefreshMaterialByCiPipelineMaterialId(material.Id)
	if err != nil {
		handler.Logger.Errorw("service err, RefreshMaterials", "err", err, "gitMaterialId", gitMaterialId)
		common.WriteJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCiPipelineMin(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//RBAC
	handler.Logger.Infow("request payload, GetCiPipelineMin", "appId", appId)
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	ciPipelines, err := handler.pipelineBuilder.GetCiPipelineMin(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetCiPipelineMin", "err", err, "appId", appId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: http.StatusNotFound, UserMessage: "no data found"}
			common.WriteJsonResp(w, err, nil, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	common.WriteJsonResp(w, err, ciPipelines, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) DownloadCiWorkflowArtifacts(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	buildId, err := strconv.Atoi(vars["workflowId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, DownloadCiWorkflowArtifacts", "pipelineId", pipelineId, "buildId", buildId)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	file, err := handler.ciHandler.DownloadCiWorkflowArtifacts(pipelineId, buildId)
	defer file.Close()
	if err != nil {
		handler.Logger.Errorw("service err, DownloadCiWorkflowArtifacts", "err", err, "pipelineId", pipelineId, "buildId", buildId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no workflow found"}
			common.WriteJsonResp(w, err, nil, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Itoa(buildId)+".zip")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", r.Header.Get("Content-Length"))
	_, err = io.Copy(w, file)
	if err != nil {
		handler.Logger.Errorw("service err, DownloadCiWorkflowArtifacts", "err", err, "pipelineId", pipelineId, "buildId", buildId)
	}
}

func (handler PipelineConfigRestHandlerImpl) GetHistoricBuildLogs(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.Logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	workflowId, err := strconv.Atoi(vars["workflowId"])
	if err != nil {
		handler.Logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetHistoricBuildLogs", "pipelineId", pipelineId, "workflowId", workflowId)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, GetHistoricBuildLogs", "err", err, "pipelineId", pipelineId, "workflowId", workflowId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	resp, err := handler.ciHandler.GetHistoricBuildLogs(pipelineId, workflowId, nil)
	if err != nil {
		handler.Logger.Errorw("service err, GetHistoricBuildLogs", "err", err, "pipelineId", pipelineId, "workflowId", workflowId)
		common.WriteJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler *PipelineConfigRestHandlerImpl) GetBuildHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	offsetQueryParam := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetQueryParam)
	if offsetQueryParam == "" || err != nil {
		common.WriteJsonResp(w, err, "invalid offset", http.StatusBadRequest)
		return
	}
	sizeQueryParam := r.URL.Query().Get("size")
	limit, err := strconv.Atoi(sizeQueryParam)
	if sizeQueryParam == "" || err != nil {
		common.WriteJsonResp(w, err, "invalid size", http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetBuildHistory", "pipelineId", pipelineId, "offset", offset)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, GetBuildHistory", "err", err, "pipelineId", pipelineId, "offset", offset)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	resp, err := handler.ciHandler.GetBuildHistory(pipelineId, offset, limit)
	if err != nil {
		handler.Logger.Errorw("service err, GetBuildHistory", "err", err, "pipelineId", pipelineId, "offset", offset)
		common.WriteJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler *PipelineConfigRestHandlerImpl) GetBuildLogs(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	workflowId, err := strconv.Atoi(vars["workflowId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetBuildLogs", "pipelineId", pipelineId, "workflowId", workflowId)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	lastSeenMsgId := -1
	lastEventId := r.Header.Get("Last-Event-ID")
	if len(lastEventId) > 0 {
		lastSeenMsgId, err = strconv.Atoi(lastEventId)
		if err != nil {
			handler.Logger.Errorw("request err, GetBuildLogs", "err", err, "pipelineId", pipelineId, "workflowId", workflowId, "lastEventId", lastEventId)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	logsReader, cleanUp, err := handler.ciHandler.GetRunningWorkflowLogs(pipelineId, workflowId)
	if err != nil {
		handler.Logger.Errorw("service err, GetBuildLogs", "err", err, "pipelineId", pipelineId, "workflowId", workflowId, "lastEventId", lastEventId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	defer cancel()
	defer cleanUp()
	handler.streamOutput(w, logsReader, lastSeenMsgId)
}

func (handler PipelineConfigRestHandlerImpl) FetchMaterialInfo(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	ciArtifactId, err := strconv.Atoi(vars["ciArtifactId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchMaterialInfo", "err", err, "ciArtifactId", ciArtifactId)
	resp, err := handler.ciHandler.FetchMaterialInfoByArtifactId(ciArtifactId, envId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchMaterialInfo", "err", err, "ciArtifactId", ciArtifactId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: http.StatusNotFound, UserMessage: "no material info found"}
			common.WriteJsonResp(w, err, nil, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(resp.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCIPipelineById(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	handler.Logger.Infow("request payload, GetCIPipelineById", "err", err, "appId", appId, "pipelineId", pipelineId)

	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Infow("service error, GetCIPipelineById", "err", err, "appId", appId, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	pipelineData, err := handler.pipelineRepository.FindActiveByAppIdAndPipelineId(appId, pipelineId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	var environmentIds []int
	for _, pipeline := range pipelineData {
		environmentIds = append(environmentIds, pipeline.EnvironmentId)
	}
	if handler.appWorkflowService.CheckCdPipelineByCiPipelineId(pipelineId) {
		for _, envId := range environmentIds {
			envObject := handler.enforcerUtil.GetEnvRBACNameByCiPipelineIdAndEnvId(pipelineId, envId)
			if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, envObject); !ok {
				common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
				return
			}
		}
	}

	ciPipeline, err := handler.pipelineBuilder.GetCiPipelineById(pipelineId)
	if err != nil {
		handler.Logger.Infow("service error, GetCIPipelineById", "err", err, "appId", appId, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, ciPipeline, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) CreateMaterial(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var createMaterialDto bean.CreateMaterialDTO
	err = decoder.Decode(&createMaterialDto)
	createMaterialDto.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, CreateMaterial", "err", err, "CreateMaterial", createMaterialDto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, CreateMaterial", "CreateMaterial", createMaterialDto)
	err = handler.validator.Struct(createMaterialDto)
	if err != nil {
		handler.Logger.Errorw("validation err, CreateMaterial", "err", err, "CreateMaterial", createMaterialDto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	app, err := handler.pipelineBuilder.GetApp(createMaterialDto.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if app.AppType == helper.Job {
		isSuperAdmin, err := handler.userAuthService.IsSuperAdmin(int(userId))
		if !isSuperAdmin || err != nil {
			if err != nil {
				handler.Logger.Errorw("request err, CheckSuperAdmin", "err", isSuperAdmin, "isSuperAdmin", isSuperAdmin)
			}
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
	} else {
		resourceObject := handler.enforcerUtil.GetAppRBACNameByAppId(createMaterialDto.AppId)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceObject); !ok {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	for _, gitMaterial := range createMaterialDto.Material {
		validationResult, err := handler.ValidateGitMaterialUrl(gitMaterial.GitProviderId, gitMaterial.Url)
		if err != nil {
			handler.Logger.Errorw("service err, CreateMaterial", "err", err, "CreateMaterial", createMaterialDto)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		} else {
			if !validationResult {
				handler.Logger.Errorw("validation err, CreateMaterial : invalid git material url", "err", err, "gitMaterialUrl", gitMaterial.Url, "CreateMaterial", createMaterialDto)
				common.WriteJsonResp(w, fmt.Errorf("validation for url failed"), nil, http.StatusBadRequest)
				return
			}
		}
	}

	createResp, err := handler.pipelineBuilder.CreateMaterialsForApp(&createMaterialDto)
	if err != nil {
		handler.Logger.Errorw("service err, CreateMaterial", "err", err, "CreateMaterial", createMaterialDto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) UpdateMaterial(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var updateMaterialDto bean.UpdateMaterialDTO
	err = decoder.Decode(&updateMaterialDto)
	updateMaterialDto.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, UpdateMaterial", "err", err, "UpdateMaterial", updateMaterialDto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, UpdateMaterial", "UpdateMaterial", updateMaterialDto)
	err = handler.validator.Struct(updateMaterialDto)
	if err != nil {
		handler.Logger.Errorw("validation err, UpdateMaterial", "err", err, "UpdateMaterial", updateMaterialDto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	validationResult, err := handler.ValidateGitMaterialUrl(updateMaterialDto.Material.GitProviderId, updateMaterialDto.Material.Url)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateMaterial", "err", err, "UpdateMaterial", updateMaterialDto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	} else {
		if !validationResult {
			handler.Logger.Errorw("validation err, UpdateMaterial : invalid git material url", "err", err, "gitMaterialUrl", updateMaterialDto.Material.Url, "UpdateMaterial", updateMaterialDto)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	app, err := handler.pipelineBuilder.GetApp(updateMaterialDto.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if app.AppType == helper.Job {
		isSuperAdmin, err := handler.userAuthService.IsSuperAdmin(int(userId))
		if !isSuperAdmin || err != nil {
			if err != nil {
				handler.Logger.Errorw("request err, CheckSuperAdmin", "err", isSuperAdmin, "isSuperAdmin", isSuperAdmin)
			}
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
	} else {
		resourceObject := handler.enforcerUtil.GetAppRBACNameByAppId(updateMaterialDto.AppId)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceObject); !ok {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
	}

	createResp, err := handler.pipelineBuilder.UpdateMaterialsForApp(&updateMaterialDto)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateMaterial", "err", err, "UpdateMaterial", updateMaterialDto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) DeleteMaterial(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var deleteMaterial bean.UpdateMaterialDTO
	err = decoder.Decode(&deleteMaterial)
	deleteMaterial.UserId = userId
	if err != nil {
		handler.Logger.Errorw("request err, DeleteMaterial", "err", err, "DeleteMaterial", deleteMaterial)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, DeleteMaterial", "DeleteMaterial", deleteMaterial)
	err = handler.validator.Struct(deleteMaterial)
	if err != nil {
		handler.Logger.Errorw("validation err, DeleteMaterial", "err", err, "DeleteMaterial", deleteMaterial)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//rbac starts
	app, err := handler.pipelineBuilder.GetApp(deleteMaterial.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if app.AppType == helper.Job {
		isSuperAdmin, err := handler.userAuthService.IsSuperAdmin(int(userId))
		if !isSuperAdmin || err != nil {
			if err != nil {
				handler.Logger.Errorw("request err, CheckSuperAdmin", "err", isSuperAdmin, "isSuperAdmin", isSuperAdmin)
			}
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
	} else {
		resourceObject := handler.enforcerUtil.GetAppRBACNameByAppId(deleteMaterial.AppId)
		token := r.Header.Get("token")
		if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceObject); !ok {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	//rbac ends
	err = handler.pipelineBuilder.DeleteMaterial(&deleteMaterial)
	if err != nil {
		handler.Logger.Errorw("service err, DeleteMaterial", "err", err, "DeleteMaterial", deleteMaterial)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, GIT_MATERIAL_DELETE_SUCCESS_RESP, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) HandleWorkflowWebhook(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var wfUpdateReq v1alpha1.WorkflowStatus
	err := decoder.Decode(&wfUpdateReq)
	if err != nil {
		handler.Logger.Errorw("request err, HandleWorkflowWebhook", "err", err, "payload", wfUpdateReq)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, HandleWorkflowWebhook", "payload", wfUpdateReq)
	resp, err := handler.ciHandler.UpdateWorkflow(wfUpdateReq)
	if err != nil {
		handler.Logger.Errorw("service err, HandleWorkflowWebhook", "err", err, "payload", wfUpdateReq)
		common.WriteJsonResp(w, err, resp, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) validForMultiMaterial(ciTriggerRequest bean.CiTriggerRequest) bool {
	if len(ciTriggerRequest.CiPipelineMaterial) > 1 {
		for _, m := range ciTriggerRequest.CiPipelineMaterial {
			if m.GitCommit.Commit == "" {
				return false
			}
		}
	}
	return true
}

func (handler PipelineConfigRestHandlerImpl) ValidateGitMaterialUrl(gitProviderId int, url string) (bool, error) {
	gitProvider, err := handler.gitProviderRepo.FindOne(strconv.Itoa(gitProviderId))
	if err != nil {
		return false, err
	}
	if gitProvider.AuthMode == repository.AUTH_MODE_SSH {
		hasPrefixResult := strings.HasPrefix(url, SSH_URL_PREFIX)
		return hasPrefixResult, nil
	}
	hasPrefixResult := strings.HasPrefix(url, HTTPS_URL_PREFIX)
	return hasPrefixResult, nil
}

func (handler PipelineConfigRestHandlerImpl) CancelWorkflow(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	workflowId, err := strconv.Atoi(vars["workflowId"])
	if err != nil {
		handler.Logger.Errorw("request err, CancelWorkflow", "err", err, "workflowId", workflowId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.Logger.Errorw("request err, CancelWorkflow", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, CancelWorkflow", "workflowId", workflowId, "pipelineId", pipelineId)

	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, CancelWorkflow", "err", err, "workflowId", workflowId, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	if handler.appWorkflowService.CheckCdPipelineByCiPipelineId(pipelineId) {
		pipelineData, err := handler.pipelineRepository.FindActiveByAppIdAndPipelineId(ciPipeline.AppId, pipelineId)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		var environmentIds []int
		for _, pipeline := range pipelineData {
			environmentIds = append(environmentIds, pipeline.EnvironmentId)
		}
		if handler.appWorkflowService.CheckCdPipelineByCiPipelineId(pipelineId) {
			for _, envId := range environmentIds {
				envObject := handler.enforcerUtil.GetEnvRBACNameByCiPipelineIdAndEnvId(pipelineId, envId)
				if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, envObject); !ok {
					common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
					return
				}
			}
		}
	}

	//RBAC

	resp, err := handler.ciHandler.CancelBuild(workflowId)
	if err != nil {
		handler.Logger.Errorw("service err, CancelWorkflow", "err", err, "workflowId", workflowId, "pipelineId", pipelineId)
		if util.IsErrNoRows(err) {
			common.WriteJsonResp(w, err, nil, http.StatusNotFound)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

// FetchChanges FIXME check if deprecated
func (handler PipelineConfigRestHandlerImpl) FetchChanges(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	ciMaterialId, err := strconv.Atoi(vars["ciMaterialId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	showAll := false
	v := r.URL.Query()
	show := v.Get("showAll")
	if len(show) > 0 {
		showAll, err = strconv.ParseBool(show)
		if err != nil {
			showAll = true
			err = nil
			//ignore error, apply rbac by default
		}
	}
	handler.Logger.Infow("request payload, FetchChanges", "ciMaterialId", ciMaterialId, "pipelineId", pipelineId)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("request err, FetchChanges", "err", err, "ciMaterialId", ciMaterialId, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	changeRequest := &gitSensor.FetchScmChangesRequest{
		PipelineMaterialId: ciMaterialId,
		ShowAll:            showAll,
	}
	changes, err := handler.gitSensorClient.FetchChanges(context.Background(), changeRequest)
	if err != nil {
		handler.Logger.Errorw("service err, FetchChanges", "err", err, "ciMaterialId", ciMaterialId, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, changes.Commits, http.StatusCreated)
}

func (handler PipelineConfigRestHandlerImpl) GetCommitMetadataForPipelineMaterial(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	ciPipelineMaterialId, err := strconv.Atoi(vars["ciPipelineMaterialId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	gitHash := vars["gitHash"]
	handler.Logger.Infow("request payload, GetCommitMetadataForPipelineMaterial", "ciPipelineMaterialId", ciPipelineMaterialId, "gitHash", gitHash)

	// get ci-pipeline-material
	ciPipelineMaterial, err := handler.ciPipelineMaterialRepository.GetById(ciPipelineMaterialId)
	if err != nil {
		handler.Logger.Errorw("error while fetching ciPipelineMaterial", "err", err, "ciPipelineMaterialId", ciPipelineMaterialId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipelineMaterial.CiPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	commitMetadataRequest := &gitSensor.CommitMetadataRequest{
		PipelineMaterialId: ciPipelineMaterialId,
		GitHash:            gitHash,
	}
	commit, err := handler.gitSensorClient.GetCommitMetadataForPipelineMaterial(context.Background(), commitMetadataRequest)
	if err != nil {
		handler.Logger.Errorw("error while fetching commit metadata for pipeline material", "commitMetadataRequest", commitMetadataRequest, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, commit, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) FetchWorkflowDetails(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	buildId, err := strconv.Atoi(vars["workflowId"])
	if err != nil || buildId == 0 {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, FetchWorkflowDetails", "appId", appId, "pipelineId", pipelineId, "buildId", buildId, "buildId", buildId)
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	resp, err := handler.ciHandler.FetchWorkflowDetails(appId, pipelineId, buildId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchWorkflowDetails", "err", err, "appId", appId, "pipelineId", pipelineId, "buildId", buildId, "buildId", buildId)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{Code: "404", HttpStatusCode: http.StatusNotFound, UserMessage: "no workflow found"}
			common.WriteJsonResp(w, err, nil, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCiPipelineByEnvironment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	user, err := handler.userAuthService.GetById(userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	userEmailId := strings.ToLower(user.EmailId)
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetCdPipelines", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	v := r.URL.Query()
	appIdsString := v.Get("appIds")
	var appIds []int
	if len(appIdsString) > 0 {
		appIdsSlices := strings.Split(appIdsString, ",")
		for _, appId := range appIdsSlices {
			id, err := strconv.Atoi(appId)
			if err != nil {
				common.WriteJsonResp(w, err, "please provide valid appIds", http.StatusBadRequest)
				return
			}
			appIds = append(appIds, id)
		}
	}
	var appGroupId int
	appGroupIdStr := v.Get("appGroupId")
	if len(appGroupIdStr) > 0 {
		appGroupId, err = strconv.Atoi(appGroupIdStr)
		if err != nil {
			common.WriteJsonResp(w, err, "please provide valid appGroupId", http.StatusBadRequest)
			return
		}
	}
	request := appGroup2.AppGroupingRequest{
		EnvId:          envId,
		AppGroupId:     appGroupId,
		AppIds:         appIds,
		EmailId:        userEmailId,
		CheckAuthBatch: handler.checkAuthBatch,
		UserId:         userId,
		Ctx:            r.Context(),
	}
	_, span := otel.Tracer("orchestrator").Start(r.Context(), "ciHandler.FetchCiPipelinesForAppGrouping")
	ciConf, err := handler.pipelineBuilder.GetCiPipelineByEnvironment(request)
	span.End()
	if err != nil {
		handler.Logger.Errorw("service err, GetCiPipeline", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, ciConf, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetCiPipelineByEnvironmentMin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	user, err := handler.userAuthService.GetById(userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	userEmailId := strings.ToLower(user.EmailId)
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		handler.Logger.Errorw("request err, GetCdPipelines", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	v := r.URL.Query()
	appIdsString := v.Get("appIds")
	var appIds []int
	if len(appIdsString) > 0 {
		appIdsSlices := strings.Split(appIdsString, ",")
		for _, appId := range appIdsSlices {
			id, err := strconv.Atoi(appId)
			if err != nil {
				common.WriteJsonResp(w, err, "please provide valid appIds", http.StatusBadRequest)
				return
			}
			appIds = append(appIds, id)
		}
	}
	var appGroupId int
	appGroupIdStr := v.Get("appGroupId")
	if len(appGroupIdStr) > 0 {
		appGroupId, err = strconv.Atoi(appGroupIdStr)
		if err != nil {
			common.WriteJsonResp(w, err, "please provide valid appGroupId", http.StatusBadRequest)
			return
		}
	}
	request := appGroup2.AppGroupingRequest{
		EnvId:          envId,
		AppGroupId:     appGroupId,
		AppIds:         appIds,
		EmailId:        userEmailId,
		CheckAuthBatch: handler.checkAuthBatch,
		UserId:         userId,
		Ctx:            r.Context(),
	}
	_, span := otel.Tracer("orchestrator").Start(r.Context(), "ciHandler.FetchCiPipelinesForAppGrouping")
	results, err := handler.pipelineBuilder.GetCiPipelineByEnvironmentMin(request)
	span.End()
	if err != nil {
		handler.Logger.Errorw("service err, GetCiPipeline", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, results, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetExternalCiByEnvironment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	user, err := handler.userAuthService.GetById(userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	userEmailId := strings.ToLower(user.EmailId)
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	v := r.URL.Query()
	appIdsString := v.Get("appIds")
	var appIds []int
	if len(appIdsString) > 0 {
		appIdsSlices := strings.Split(appIdsString, ",")
		for _, appId := range appIdsSlices {
			id, err := strconv.Atoi(appId)
			if err != nil {
				common.WriteJsonResp(w, err, "please provide valid appIds", http.StatusBadRequest)
				return
			}
			appIds = append(appIds, id)
		}
	}

	var appGroupId int
	appGroupIdStr := v.Get("appGroupId")
	if len(appGroupIdStr) > 0 {
		appGroupId, err = strconv.Atoi(appGroupIdStr)
		if err != nil {
			common.WriteJsonResp(w, err, "please provide valid appGroupId", http.StatusBadRequest)
			return
		}
	}
	request := appGroup2.AppGroupingRequest{
		EnvId:          envId,
		AppGroupId:     appGroupId,
		AppIds:         appIds,
		EmailId:        userEmailId,
		CheckAuthBatch: handler.checkAuthBatch,
		UserId:         userId,
		Ctx:            r.Context(),
	}
	_, span := otel.Tracer("orchestrator").Start(r.Context(), "ciHandler.FetchExternalCiPipelinesForAppGrouping")
	ciConf, err := handler.pipelineBuilder.GetExternalCiByEnvironment(request)
	span.End()
	if err != nil {
		handler.Logger.Errorw("service err, GetExternalCi", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, ciConf, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) CreateUpdateImageTagging(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	user, err := handler.userAuthService.GetById(userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	isSuperAdmin, err := handler.userService.IsSuperAdmin(int(user.Id))
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	artifactId, err := strconv.Atoi(vars["artifactId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	pipelineId, err := strconv.Atoi(vars["ciPipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	decoder := json.NewDecoder(r.Body)
	req := &pipeline.ImageTaggingRequestDTO{}
	err = decoder.Decode(&req)
	if err != nil {
		handler.Logger.Errorw("request err, CreateUpdateImageTagging", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//RBAC
	if !!isSuperAdmin {
		object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
		if ok := handler.enforcer.EnforceByEmail(strings.ToLower(user.EmailId), casbin.ResourceApplications, casbin.ActionTrigger, object); !ok {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	//RBAC
	//check prod env exists
	prodEnvExists, err := handler.imageTaggingService.GetEnvFromParentAndLinkedWorkflow(ciPipeline.Id)
	if err != nil {
		handler.Logger.Errorw("error occured in checking existance prod prod environment ", "err", err, "ciPipelineId", ciPipeline.Id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//not allowed to perform edit/save if no cd exists in prod env in the app_workflow
	if !prodEnvExists {
		handler.Logger.Errorw("save or edit operation not possible for this artifact", "err", nil, "artifactId", artifactId, "ciPipelineId", ciPipeline.Id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
	}
	hardDeleteTags := req.HardDeleteTags
	if !isSuperAdmin && len(hardDeleteTags) > 0 {
		errMsg := errors.New("user dont have permission to delete the tags")
		handler.Logger.Errorw("request err, CreateUpdateImageTagging", "err", errMsg, "payload", req)
		common.WriteJsonResp(w, errMsg, nil, http.StatusBadRequest)
		return
	}

	//pass it to service layer
	resp, err := handler.imageTaggingService.CreateUpdateImageTagging(ciPipeline.Id, ciPipeline.AppId, artifactId, req)
	if err != nil {
		handler.Logger.Errorw("error occured in creating/updating image tagging data", "err", err, "ciPipelineId", ciPipeline.Id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GetImageTaggingData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	user, err := handler.userAuthService.GetById(userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	userEmailId := strings.ToLower(user.EmailId)
	artifactId, err := strconv.Atoi(vars["artifactId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["ciPipelineId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	if ok := handler.enforcer.EnforceByEmail(userEmailId, casbin.ResourceApplications, casbin.ActionTrigger, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	resp, err := handler.imageTaggingService.GetTagsData(ciPipeline.Id, ciPipeline.AppId, artifactId)
	if err != nil {
		//logg
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, resp, http.StatusOK)
}
