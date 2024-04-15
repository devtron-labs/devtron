package scanningResultsParser

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/enterprise/pkg/scanningResultsParser"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type ScanningResultRestHandler interface {
	ScanResults(w http.ResponseWriter, r *http.Request)
}

type ScanningResultRestHandlerImpl struct {
	logger       *zap.SugaredLogger
	userService  user.UserService
	scanService  scanningResultsParser.Service
	enforcer     casbin.Enforcer
	enforcerUtil rbac.EnforcerUtil
	validator    *validator.Validate
}

func NewScanningResultRestHandlerImpl(
	logger *zap.SugaredLogger,
	userService user.UserService,
	scanService scanningResultsParser.Service,
	enforcer casbin.Enforcer,
	enforcerUtil rbac.EnforcerUtil,
	validator *validator.Validate,
) *ScanningResultRestHandlerImpl {
	return &ScanningResultRestHandlerImpl{
		logger:       logger,
		userService:  userService,
		scanService:  scanService,
		enforcer:     enforcer,
		enforcerUtil: enforcerUtil,
		validator:    validator,
	}
}
func (impl ScanningResultRestHandlerImpl) ScanResults(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	v := r.URL.Query()
	appId, err := strconv.Atoi(v.Get("appId"))
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(v.Get("envId"))
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//RBAC
	token := r.Header.Get("token")
	object := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	object = impl.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
	if ok := impl.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	resp, err := impl.scanService.GetScanResults(appId, envId)
	if err != nil {
		impl.logger.Errorw("service err, scan results", "err", err, "appId %d envId %d", appId, envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)

		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)

}
