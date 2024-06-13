package fluxApplication

import (
	"context"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	clientErrors "github.com/devtron-labs/devtron/pkg/errors"
	"github.com/devtron-labs/devtron/pkg/fluxApplication"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type FluxApplicationRestHandler interface {
	ListFluxApplications(w http.ResponseWriter, r *http.Request)
	GetApplicationDetail(w http.ResponseWriter, r *http.Request)
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

func (handler *FluxApplicationRestHandlerImpl) ListFluxApplications(w http.ResponseWriter, r *http.Request) {
	//handle super-admin RBAC
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	v := r.URL.Query()
	clusterIdString := v.Get("clusterIds")
	var clusterIds []int
	var err error
	//handling when the clusterIds string is not empty , if it is empty then it will fetch for all the active cluster
	if clusterIdString != "" {
		clusterIds, err = common.ExtractIntArrayQueryParam(w, r, "clusterIds")
		if err != nil {
			handler.logger.Errorw("error in getting cluster ids", "error", err, "clusterIds", clusterIds)
			return
		}
	}
	handler.fluxApplicationService.ListFluxApplications(r.Context(), clusterIds, w)
}
func (handler *FluxApplicationRestHandlerImpl) GetApplicationDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterIdString := vars["appId"]
	appIdentifier, err := handler.fluxApplicationService.DecodeFluxAppId(clusterIdString)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// handle super-admin RBAC
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	res, err := handler.fluxApplicationService.GetFluxAppDetail(context.Background(), appIdentifier)
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
