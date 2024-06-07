package devtronResource

import (
	"fmt"
	apiBean "github.com/devtron-labs/devtron/api/devtronResource/bean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/history/deployment/cdPipeline"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/schema"
	"go.uber.org/zap"
	"net/http"
)

type HistoryRestHandler interface {
	GetDeploymentHistory(w http.ResponseWriter, r *http.Request)
	GetDeploymentHistoryConfigList(w http.ResponseWriter, r *http.Request)
}

type HistoryRestHandlerImpl struct {
	logger                   *zap.SugaredLogger
	enforcer                 casbin.Enforcer
	deploymentHistoryService cdPipeline.DeploymentHistoryService
	apiReqDecoderService     devtronResource.APIReqDecoderService
	enforcerUtil             rbac.EnforcerUtil
}

func NewHistoryRestHandlerImpl(logger *zap.SugaredLogger,
	enforcer casbin.Enforcer,
	deploymentHistoryService cdPipeline.DeploymentHistoryService,
	apiReqDecoderService devtronResource.APIReqDecoderService,
	enforcerUtil rbac.EnforcerUtil) *HistoryRestHandlerImpl {
	return &HistoryRestHandlerImpl{
		logger:                   logger,
		enforcer:                 enforcer,
		deploymentHistoryService: deploymentHistoryService,
		apiReqDecoderService:     apiReqDecoderService,
		enforcerUtil:             enforcerUtil,
	}
}

func (handler *HistoryRestHandlerImpl) GetDeploymentHistory(w http.ResponseWriter, r *http.Request) {
	kind, _, _, caughtError := getKindSubKindVersion(w, r)
	if caughtError || kind != bean.DevtronResourceCdPipeline.ToString() {
		common.WriteJsonResp(w, fmt.Errorf(apiBean.RequestInvalidKindVersionErrMessage), nil, http.StatusBadRequest)
		return
	}
	v := r.URL.Query()
	var decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	queryParams := apiBean.GetHistoryQueryParams{}
	err := decoder.Decode(&queryParams, v)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	decodedReqBean, err := handler.apiReqDecoderService.GetFilterCriteriaParamsForDeploymentHistory(queryParams.FilterCriteria)
	if err != nil {
		handler.logger.Errorw("error in getting filter criteria params", "err", err, "filterCriteria", queryParams.FilterCriteria)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC START
	token := r.Header.Get("token")
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(decodedReqBean.AppId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	// RBAC END

	resp, err := handler.deploymentHistoryService.
		GetCdPipelineDeploymentHistory(adapter.GetCDDeploymentHistoryListReq(&queryParams, decodedReqBean))
	if err != nil {
		handler.logger.Errorw("service error, GetCdPipelineDeploymentHistory", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
	return
}

func (handler *HistoryRestHandlerImpl) GetDeploymentHistoryConfigList(w http.ResponseWriter, r *http.Request) {
	kind, _, _, caughtError := getKindSubKindVersion(w, r)
	if caughtError || kind != bean.DevtronResourceCdPipeline.ToString() {
		common.WriteJsonResp(w, fmt.Errorf(apiBean.RequestInvalidKindVersionErrMessage), nil, http.StatusBadRequest)
		return
	}
	v := r.URL.Query()
	var decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	queryParams := apiBean.GetHistoryConfigQueryParams{}
	err := decoder.Decode(&queryParams, v)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	decodedReqBean, err := handler.apiReqDecoderService.GetFilterCriteriaParamsForDeploymentHistory(queryParams.FilterCriteria)
	if err != nil {
		handler.logger.Errorw("error in getting filter criteria params", "err", err, "filterCriteria", queryParams.FilterCriteria)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//RBAC START
	token := r.Header.Get("token")
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(decodedReqBean.AppId)
	if isValidated := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !isValidated {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END
	resp, err := handler.deploymentHistoryService.
		GetCdPipelineDeploymentHistoryConfigList(adapter.GetCDDeploymentHistoryConfigListReq(&queryParams, decodedReqBean))
	if err != nil {
		handler.logger.Errorw("service error, GetCdPipelineDeploymentHistoryConfigList", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
	return
}
