package fluxApplication

import (
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/fluxApplication"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

type FluxApplicationRestHandler interface {
	ListApplications(w http.ResponseWriter, r *http.Request)
	//GetApplicationDetail(w http.ResponseWriter, r *http.Request)
}

type FluxApplicationRestHandlerImpl struct {
	fluxApplicationService fluxApplication.FluxApplicationService
	logger                 *zap.SugaredLogger
	enforcer               casbin.Enforcer
}

func NewFluxApplicationRestHandlerImpl(fluxApplicationService fluxApplication.FluxApplicationService,
	logger *zap.SugaredLogger, enforcer casbin.Enforcer) *FluxApplicationRestHandlerImpl {
	return &FluxApplicationRestHandlerImpl{
		fluxApplicationService: fluxApplicationService,
		logger:                 logger,
		enforcer:               enforcer,
	}

}

func (handler *FluxApplicationRestHandlerImpl) ListApplications(w http.ResponseWriter, r *http.Request) {
	//handle super-admin RBAC
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	v := r.URL.Query()
	clusterIdString := v.Get("clusterIds")
	var clusterIds []int
	if clusterIdString != "" {
		clusterIdSlices := strings.Split(clusterIdString, ",")
		for _, clusterId := range clusterIdSlices {
			id, err := strconv.Atoi(clusterId)
			if err != nil {
				handler.logger.Errorw("error in converting clusterId", "err", err, "clusterIdString", clusterIdString)
				common.WriteJsonResp(w, err, "please send valid cluster Ids", http.StatusBadRequest)
				return
			}
			clusterIds = append(clusterIds, id)
		}
	}
	resp, err := handler.fluxApplicationService.ListFluxApplications(r.Context(), clusterIds)
	if err != nil {
		handler.logger.Errorw("error in listing all argo applications", "err", err, "clusterIds", clusterIds)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}
