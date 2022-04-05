package restHandler

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	history2 "github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type PipelineHistoryRestHandler interface {
	FetchDeploymentDetailsForDeployedTemplatesHistory(w http.ResponseWriter, r *http.Request)
	FetchDeploymentDetailsForDeployedStrategyHistory(w http.ResponseWriter, r *http.Request)
	FetchDeploymentDetailsForDeployedCMHistory(w http.ResponseWriter, r *http.Request)
	FetchDeploymentDetailsForDeployedCSHistory(w http.ResponseWriter, r *http.Request)
	FetchDeployedTemplatesHistoryById(w http.ResponseWriter, r *http.Request)
	FetchDeployedStrategyHistoryById(w http.ResponseWriter, r *http.Request)
	FetchDeployedCMHistoryById(w http.ResponseWriter, r *http.Request)
	FetchDeployedCSHistoryById(w http.ResponseWriter, r *http.Request)
	FetchDeployedPrePostCdScriptHistory(w http.ResponseWriter, r *http.Request)
}

type PipelineHistoryRestHandlerImpl struct {
	logger                           *zap.SugaredLogger
	userAuthService                  user.UserService
	enforcer                         casbin.Enforcer
	strategyHistoryService           history2.PipelineStrategyHistoryService
	deploymentTemplateHistoryService history2.DeploymentTemplateHistoryService
	configMapHistoryService          history2.ConfigMapHistoryService
	prePostCiScriptHistoryService    history2.PrePostCiScriptHistoryService
	prePostCdScriptHistoryService    history2.PrePostCdScriptHistoryService
	pipelineBuilder                  pipeline.PipelineBuilder
	enforcerUtil                     rbac.EnforcerUtil
}

func NewPipelineHistoryRestHandlerImpl(logger *zap.SugaredLogger, userAuthService user.UserService,
	enforcer casbin.Enforcer, strategyHistoryService history2.PipelineStrategyHistoryService,
	deploymentTemplateHistoryService history2.DeploymentTemplateHistoryService,
	configMapHistoryService history2.ConfigMapHistoryService,
	prePostCiScriptHistoryService history2.PrePostCiScriptHistoryService,
	prePostCdScriptHistoryService history2.PrePostCdScriptHistoryService,
	pipelineBuilder pipeline.PipelineBuilder,
	enforcerUtil rbac.EnforcerUtil) *PipelineHistoryRestHandlerImpl {
	return &PipelineHistoryRestHandlerImpl{
		logger:                           logger,
		userAuthService:                  userAuthService,
		enforcer:                         enforcer,
		strategyHistoryService:           strategyHistoryService,
		deploymentTemplateHistoryService: deploymentTemplateHistoryService,
		configMapHistoryService:          configMapHistoryService,
		prePostCdScriptHistoryService:    prePostCdScriptHistoryService,
		prePostCiScriptHistoryService:    prePostCiScriptHistoryService,
		pipelineBuilder:                  pipelineBuilder,
		enforcerUtil:                     enforcerUtil,
	}
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeploymentDetailsForDeployedTemplatesHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeploymentDetailsForDeployedTemplatesHistory", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeploymentDetailsForDeployedTemplatesHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchDeploymentDetailsForDeployedTemplatesHistory", "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	offsetQueryParam := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(offsetQueryParam)
	if offsetQueryParam == "" || err != nil {
		handler.logger.Errorw("request err, FetchDeploymentDetailsForDeployedTemplatesHistory", "err", err, "offset", offset)
		common.WriteJsonResp(w, err, "invalid offset", http.StatusBadRequest)
		return
	}
	sizeQueryParam := r.URL.Query().Get("size")
	limit, err := strconv.Atoi(sizeQueryParam)
	if sizeQueryParam == "" || err != nil {
		handler.logger.Errorw("request err, FetchDeploymentDetailsForDeployedTemplatesHistory", "err", err, "limit", limit)
		common.WriteJsonResp(w, err, "invalid size", http.StatusBadRequest)
		return
	}

	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.deploymentTemplateHistoryService.GetDeploymentDetailsForDeployedTemplateHistory(pipelineId, offset, limit)
	if err != nil {
		handler.logger.Errorw("service err, GetDeploymentDetailsForDeployedTemplateHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeployedTemplatesHistoryById(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedTemplatesHistoryById", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedTemplatesHistoryById", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedTemplatesHistoryById", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchDeployedTemplatesHistoryById", "id", id, "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.deploymentTemplateHistoryService.GetHistoryForDeployedTemplatesById(id, pipelineId)
	if err != nil {
		handler.logger.Errorw("service err, GetHistoryForDeployedTemplatesById", "err", err, "id", id, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeploymentDetailsForDeployedStrategyHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeploymentDetailsForDeployedStrategyHistory", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeploymentDetailsForDeployedStrategyHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchDeploymentDetailsForDeployedStrategyHistory", "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.strategyHistoryService.GetDeploymentDetailsForDeployedStrategyHistory(pipelineId)
	if err != nil {
		handler.logger.Errorw("service err, GetDeploymentDetailsForDeployedStrategyHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeployedStrategyHistoryById(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedStrategyHistoryById", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedStrategyHistoryById", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedStrategyHistoryById", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchDeployedStrategyHistoryById", "id", id, "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.strategyHistoryService.GetHistoryForDeployedStrategyById(id, pipelineId)
	if err != nil {
		handler.logger.Errorw("service err, GetHistoryForDeployedStrategyById", "err", err, "id", id, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeploymentDetailsForDeployedCMHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeploymentDetailsForDeployedCMHistory", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeploymentDetailsForDeployedCMHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchDeploymentDetailsForDeployedCMHistory", "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapHistoryService.GetDeploymentDetailsForDeployedCMCSHistory(pipelineId, repository.CONFIGMAP_TYPE)
	if err != nil {
		handler.logger.Errorw("service err, GetDeploymentDetailsForDeployedCMCSHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeployedCMHistoryById(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedCMHistoryById", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedCMHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedCMHistoryById", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchDeployedCMHistoryById", "id", id, "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapHistoryService.GetHistoryForDeployedCMCSById(id, pipelineId, repository.CONFIGMAP_TYPE)
	if err != nil {
		handler.logger.Errorw("service err, GetHistoryForDeployedCMCSById", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeploymentDetailsForDeployedCSHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeploymentDetailsForDeployedCSHistory", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeploymentDetailsForDeployedCSHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchDeploymentDetailsForDeployedCSHistory", "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapHistoryService.GetDeploymentDetailsForDeployedCMCSHistory(pipelineId, repository.SECRET_TYPE)
	if err != nil {
		handler.logger.Errorw("service err, GetDeploymentDetailsForDeployedCMCSHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeployedCSHistoryById(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedCSHistoryById", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedCSHistoryById", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedCSHistoryById", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchDeployedCSHistoryById", "id", id, "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapHistoryService.GetHistoryForDeployedCMCSById(id, pipelineId, repository.SECRET_TYPE)
	if err != nil {
		handler.logger.Errorw("service err, GetHistoryForDeployedCMCSById", "err", err, "id", id, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeployedPrePostCdScriptHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedPrePostCdScriptHistory", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedPrePostCdScriptHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	stage := r.URL.Query().Get("stage")
	handler.logger.Debugw("request payload, FetchDeployedPrePostCdScriptHistory", "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.prePostCdScriptHistoryService.GetHistoryForDeployedPrePostCdScript(pipelineId, repository.CdStageType(stage))
	if err != nil {
		handler.logger.Errorw("service err, GetHistoryForDeployedPrePostCdScript", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
