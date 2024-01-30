package status

import (
	"fmt"
	"github.com/devtron-labs/devtron/client/cron"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"net/http"
	"strconv"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type PipelineStatusTimelineRestHandler interface {
	FetchTimelines(w http.ResponseWriter, r *http.Request)
	ManualSyncAcdPipelineDeploymentStatus(w http.ResponseWriter, r *http.Request)
}

type PipelineStatusTimelineRestHandlerImpl struct {
	logger                           *zap.SugaredLogger
	userService                      user.UserService
	pipelineStatusTimelineService    status.PipelineStatusTimelineService
	enforcerUtil                     rbac.EnforcerUtil
	enforcer                         casbin.Enforcer
	pipeline                         pipeline.PipelineBuilder
	cdApplicationStatusUpdateHandler cron.CdApplicationStatusUpdateHandler
}

func NewPipelineStatusTimelineRestHandlerImpl(logger *zap.SugaredLogger,
	pipelineStatusTimelineService status.PipelineStatusTimelineService, enforcerUtil rbac.EnforcerUtil,
	enforcer casbin.Enforcer) *PipelineStatusTimelineRestHandlerImpl {
	return &PipelineStatusTimelineRestHandlerImpl{
		logger:                        logger,
		pipelineStatusTimelineService: pipelineStatusTimelineService,
		enforcerUtil:                  enforcerUtil,
		enforcer:                      enforcer,
	}
}

func (handler *PipelineStatusTimelineRestHandlerImpl) FetchTimelines(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	wfrId := 0
	wfrIdParam := r.URL.Query().Get("wfrId")
	if len(wfrIdParam) != 0 {
		wfrId, err = strconv.Atoi(wfrIdParam)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	showTimeline := false
	showTimelineParam := r.URL.Query().Get("showTimeline")
	if len(showTimelineParam) > 0 {
		showTimeline, err = strconv.ParseBool(showTimelineParam)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	timelines, err := handler.pipelineStatusTimelineService.FetchTimelines(appId, envId, wfrId, showTimeline)
	if err != nil {
		handler.logger.Errorw("error in getting cd pipeline status timelines by wfrId", "err", err, "wfrId", wfrId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, timelines, http.StatusOK)
	return
}

func (handler *PipelineStatusTimelineRestHandlerImpl) ManualSyncAcdPipelineDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, ManualSyncAcdPipelineDeploymentStatus", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		handler.logger.Errorw("request err, ManualSyncAcdPipelineDeploymentStatus", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	app, err := handler.pipeline.GetApp(appId)
	if err != nil {
		handler.logger.Errorw("bad request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	object := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "unauthorized user", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	if app.AppType == helper.ChartStoreApp {
		err = handler.cdApplicationStatusUpdateHandler.ManualSyncPipelineStatus(appId, 0, userId)
	} else {
		err = handler.cdApplicationStatusUpdateHandler.ManualSyncPipelineStatus(appId, envId, userId)
	}

	if err != nil {
		handler.logger.Errorw("service err, ManualSyncAcdPipelineDeploymentStatus", "err", err, "appId", appId, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, "App synced successfully.", http.StatusOK)
}
