package appStore

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type AppStoreStatusTimelineRestHandler interface {
	FetchTimelinesForAppStore(w http.ResponseWriter, r *http.Request)
}

type AppStoreStatusTimelineRestHandlerImpl struct {
	logger                        *zap.SugaredLogger
	pipelineStatusTimelineService status.PipelineStatusTimelineService
	enforcerUtil                  rbac.EnforcerUtil
	enforcer                      casbin.Enforcer
}

func NewAppStoreStatusTimelineRestHandlerImpl(logger *zap.SugaredLogger,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	enforcerUtil rbac.EnforcerUtil,
	enforcer casbin.Enforcer) *AppStoreStatusTimelineRestHandlerImpl {
	return &AppStoreStatusTimelineRestHandlerImpl{
		logger:                        logger,
		pipelineStatusTimelineService: pipelineStatusTimelineService,
		enforcerUtil:                  enforcerUtil,
		enforcer:                      enforcer,
	}
}

func (handler AppStoreStatusTimelineRestHandlerImpl) FetchTimelinesForAppStore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	installedAppId, err := strconv.Atoi(vars["installedAppId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	installedAppVersionHistoryId := 0
	installedAppVersionHistoryIdParam := r.URL.Query().Get("installedAppVersionHistoryId")
	if len(installedAppVersionHistoryIdParam) != 0 {
		installedAppVersionHistoryId, err = strconv.Atoi(installedAppVersionHistoryIdParam)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(installedAppId)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	timelines, err := handler.pipelineStatusTimelineService.FetchTimelinesForAppStore(installedAppId, envId, installedAppVersionHistoryId)
	if err != nil {
		handler.logger.Errorw("error in getting cd pipeline status timelines by wfrId", "err", err, "wfrId", installedAppVersionHistoryId, "installedAppId", installedAppId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, timelines, http.StatusOK)
	return
}
