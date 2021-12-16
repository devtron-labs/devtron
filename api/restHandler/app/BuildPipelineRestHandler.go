package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type DevtronAppBuildRestHandler interface {
	CreateCiConfig(w http.ResponseWriter, r *http.Request)
	UpdateCiTemplate(w http.ResponseWriter, r *http.Request)

	GetCiPipeline(w http.ResponseWriter, r *http.Request)
	PatchCiPipelines(w http.ResponseWriter, r *http.Request)
	TriggerCiPipeline(w http.ResponseWriter, r *http.Request)
	GetCiPipelineMin(w http.ResponseWriter, r *http.Request)
	GetCIPipelineById(w http.ResponseWriter, r *http.Request)
	HandleWorkflowWebhook(w http.ResponseWriter, r *http.Request)
	GetBuildLogs(w http.ResponseWriter, r *http.Request)
	FetchWorkflowDetails(w http.ResponseWriter, r *http.Request)
	// CancelWorkflow CancelBuild
	CancelWorkflow(w http.ResponseWriter, r *http.Request)
}

type DevtronAppBuildMaterialRestHandler interface {
	CreateMaterial(w http.ResponseWriter, r *http.Request)
	UpdateMaterial(w http.ResponseWriter, r *http.Request)
	FetchMaterials(w http.ResponseWriter, r *http.Request)
	RefreshMaterials(w http.ResponseWriter, r *http.Request)
	FetchMaterialInfo(w http.ResponseWriter, r *http.Request)
	FetchChanges(w http.ResponseWriter, r *http.Request)
}

type DevtronAppBuildHistoryRestHandler interface {
	GetHistoricBuildLogs(w http.ResponseWriter, r *http.Request)
	GetBuildHistory(w http.ResponseWriter, r *http.Request)
	DownloadCiWorkflowArtifacts(w http.ResponseWriter, r *http.Request)
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
	handler.Logger.Infow("request payload, update ci template", "UpdateCiTemplate", configRequest)
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

func (handler PipelineConfigRestHandlerImpl) PatchCiPipelines(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
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
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	createResp, err := handler.pipelineBuilder.PatchCiPipeline(&patchRequest)
	if err != nil {
		handler.Logger.Errorw("service err, PatchCiPipelines", "err", err, "PatchCiPipelines", patchRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
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
	handler.Logger.Infow("request payload, TriggerCiPipeline", "payload", ciTriggerRequest)

	//RBAC CHECK CD PIPELINE - FOR USER
	pipelines, err := handler.pipelineRepository.FindAutomaticByCiPipelineId(ciTriggerRequest.PipelineId)
	var authorizedPipelines []pipelineConfig.Pipeline
	var unauthorizedPipelines []pipelineConfig.Pipeline
	//fetching user only for getting token
	triggeredBy, err := handler.userAuthService.GetById(ciTriggerRequest.TriggeredBy)
	if err != nil {
		handler.Logger.Errorw("service err, TriggerCiPipeline", "err", err, "payload", ciTriggerRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := triggeredBy.AccessToken
	for _, p := range pipelines {
		pass := 0
		object := handler.enforcerUtil.GetAppRBACNameByAppId(p.AppId)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, object); !ok {
			handler.Logger.Debug(fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		} else {
			pass = 1
		}
		object = handler.enforcerUtil.GetAppRBACByAppIdAndPipelineId(p.AppId, p.Id)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, object); !ok {
			handler.Logger.Debug(fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		} else {
			pass = 2
		}
		if pass == 2 {
			authorizedPipelines = append(authorizedPipelines, *p)
		} else {
			unauthorizedPipelines = append(unauthorizedPipelines, *p)
		}
	}
	resMessage := "allowed for all pipelines"
	response := make(map[string]string)
	if len(unauthorizedPipelines) > 0 {
		resMessage = "not authorized for few pipelines, will not effected"
	}
	//RBAC CHECK CD PIPELINE - FOR USER

	resp, err := handler.ciHandler.HandleCIManual(ciTriggerRequest)
	if err != nil {
		handler.Logger.Errorw("service err, TriggerCiPipeline", "err", err, "payload", ciTriggerRequest)
		common.WriteJsonResp(w, err, response, http.StatusInternalServerError)
	}
	response["apiResponse"] = strconv.Itoa(resp)
	response["authStatus"] = resMessage

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
	resp, err := handler.ciHandler.FetchMaterialsByPipelineId(pipelineId)
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
	handler.Logger.Infow("request payload, FetchMaterialInfo", "err", err, "ciArtifactId", ciArtifactId)
	resp, err := handler.ciHandler.FetchMaterialInfoByArtifactId(ciArtifactId)
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
	resourceObject := handler.enforcerUtil.GetAppRBACNameByAppId(createMaterialDto.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceObject); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
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
	resourceObject := handler.enforcerUtil.GetAppRBACNameByAppId(updateMaterialDto.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceObject); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}

	createResp, err := handler.pipelineBuilder.UpdateMaterialsForApp(&updateMaterialDto)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateMaterial", "err", err, "UpdateMaterial", updateMaterialDto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, createResp, http.StatusOK)
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
	}
	changes, err := handler.gitSensorClient.FetchChanges(changeRequest)
	if err != nil {
		handler.Logger.Errorw("service err, FetchChanges", "err", err, "ciMaterialId", ciMaterialId, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, changes.Commits, http.StatusCreated)
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
